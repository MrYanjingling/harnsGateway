package audit

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeWriter struct {
	mu         sync.Mutex
	batches    [][]*Event
	rows       []*Event
	batchCalls int
	batchErr   error
	insertErr  error
	block      chan struct{}
}

func (f *fakeWriter) Insert(_ context.Context, e *Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rows = append(f.rows, e)
	return f.insertErr
}

func (f *fakeWriter) InsertBatch(_ context.Context, events []*Event) error {
	f.mu.Lock()
	f.batchCalls++
	block := f.block
	f.mu.Unlock()

	if block != nil {
		<-block
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.batchErr != nil {
		return f.batchErr
	}
	f.batches = append(f.batches, append([]*Event(nil), events...))
	return nil
}

func (f *fakeWriter) batchCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.batches)
}

func (f *fakeWriter) rowCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.rows)
}

func (f *fakeWriter) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.batchCalls
}

func testConfig(batchSize, flushMs int) Config {
	return Config{
		Enabled:         true,
		MaxBodyBytes:    16384,
		QueueSize:       64,
		BatchSize:       batchSize,
		FlushIntervalMs: flushMs,
	}
}

func testEvent(id string) Event {
	return Event{ID: id, Path: "/test"}
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func TestRecorderFlushesWhenBatchIsFull(t *testing.T) {
	f := &fakeWriter{}
	r := newRecorderWithWriter(f, testConfig(2, 10000), AppEMS, nil)
	defer r.Close(context.Background())

	_ = r.Record(context.Background(), testEvent("1"))
	_ = r.Record(context.Background(), testEvent("2"))

	waitFor(t, time.Second, func() bool { return f.batchCount() == 1 })
	if got := f.batchCount(); got != 1 {
		t.Fatalf("batches = %d, want 1", got)
	}
}

func TestRecorderFlushesOnInterval(t *testing.T) {
	f := &fakeWriter{}
	r := newRecorderWithWriter(f, testConfig(100, 20), AppEMS, nil)
	defer r.Close(context.Background())

	_ = r.Record(context.Background(), testEvent("1"))

	waitFor(t, time.Second, func() bool { return f.batchCount() == 1 })
}

func TestRecorderCloseDrainsRemainingEvents(t *testing.T) {
	f := &fakeWriter{}
	r := newRecorderWithWriter(f, testConfig(2, 10000), AppEMS, nil)

	_ = r.Record(context.Background(), testEvent("1"))
	_ = r.Record(context.Background(), testEvent("2"))
	_ = r.Record(context.Background(), testEvent("3"))

	r.Close(context.Background())

	total := 0
	for _, b := range f.batches {
		total += len(b)
	}
	if total != 3 {
		t.Fatalf("flushed rows = %d, want 3", total)
	}
}

func TestRecorderRetriesAndFallsBackToPerRow(t *testing.T) {
	f := &fakeWriter{batchErr: errors.New("batch insert failed")}
	r := newRecorderWithWriter(f, testConfig(2, 10000), AppEMS, nil)
	defer r.Close(context.Background())

	_ = r.Record(context.Background(), testEvent("1"))
	_ = r.Record(context.Background(), testEvent("2"))

	waitFor(t, 2*time.Second, func() bool { return f.rowCount() == 2 })
	if got := f.calls(); got < 2 {
		t.Fatalf("batch attempts = %d, want at least 2", got)
	}
}

func TestRecorderStampsAppAndSanitizes(t *testing.T) {
	f := &fakeWriter{}
	r := newRecorderWithWriter(f, testConfig(1, 10000), AppEMS, nil)
	defer r.Close(context.Background())

	err := r.Record(context.Background(), Event{
		App:         "EAM",
		Path:        "/test",
		RequestBody: `{"password":"secret","name":"x"}`,
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	waitFor(t, time.Second, func() bool { return f.batchCount() == 1 })
	e := f.batches[0][0]
	if e.App != AppEMS {
		t.Fatalf("app = %q, want %q", e.App, AppEMS)
	}
	if e.ID == "" {
		t.Fatal("id was not generated")
	}
	if e.CreatedTime.IsZero() {
		t.Fatal("createdTime was not generated")
	}
	if strings.Contains(e.RequestBody, "secret") || !strings.Contains(e.RequestBody, "***") {
		t.Fatalf("request body was not masked: %s", e.RequestBody)
	}
}

func TestRecorderOmitsNonJSONBodies(t *testing.T) {
	f := &fakeWriter{}
	r := newRecorderWithWriter(f, testConfig(1, 10000), AppEMS, nil)
	defer r.Close(context.Background())

	_ = r.Record(context.Background(), Event{
		Path:         "/test",
		RequestBody:  "plain text",
		ResponseBody: "<response body exceeds limit>",
	})
	waitFor(t, time.Second, func() bool { return f.batchCount() == 1 })
	e := f.batches[0][0]
	if e.RequestBody != "<non-json body omitted>" {
		t.Fatalf("request body = %q", e.RequestBody)
	}
	if e.ResponseBody != "<response body exceeds limit>" {
		t.Fatalf("response body marker = %q", e.ResponseBody)
	}
}

func TestRecorderReturnsQueueFull(t *testing.T) {
	f := &fakeWriter{block: make(chan struct{})}
	r := newRecorderWithWriter(f, Config{
		Enabled:         true,
		MaxBodyBytes:    16384,
		QueueSize:       1,
		BatchSize:       1,
		FlushIntervalMs: 10000,
	}, AppEMS, nil)

	if err := r.Record(context.Background(), testEvent("1")); err != nil {
		t.Fatalf("first Record() error = %v", err)
	}
	waitFor(t, time.Second, func() bool { return f.calls() == 1 })
	if err := r.Record(context.Background(), testEvent("2")); err != nil {
		t.Fatalf("second Record() error = %v", err)
	}
	if err := r.Record(context.Background(), testEvent("3")); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("third Record() error = %v, want ErrQueueFull", err)
	}
	close(f.block)
	r.Close(context.Background())
}

func TestRecorderRejectsAfterClose(t *testing.T) {
	f := &fakeWriter{}
	r := newRecorderWithWriter(f, testConfig(100, 10000), AppEMS, nil)
	r.Close(context.Background())

	if err := r.Record(context.Background(), testEvent("late")); !errors.Is(err, ErrClosed) {
		t.Fatalf("Record() error = %v, want ErrClosed", err)
	}
}
