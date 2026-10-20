package audit

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/oklog/ulid/v2"

	"harns-platform/pkg/logging"
)

const (
	writeTimeout    = 5 * time.Second
	fallbackTimeout = 10 * time.Second
)

var (
	// ErrQueueFull is returned when the audit queue cannot accept more events.
	ErrQueueFull = errors.New("audit queue is full")
	// ErrClosed is returned after the recorder has been closed.
	ErrClosed = errors.New("audit recorder is closed")
)

type writer interface {
	Insert(ctx context.Context, e *Event) error
	InsertBatch(ctx context.Context, events []*Event) error
}

type recorder struct {
	writer  writer
	cfg     Config
	metrics Metrics
	app     string

	ch   chan *Event
	done chan struct{}

	mu          sync.RWMutex
	closed      bool
	shutdownCtx context.Context

	wg sync.WaitGroup
}

func newRecorder(st *store, cfg Config, app string, m Metrics) *recorder {
	return newRecorderWithWriter(st, cfg, app, m)
}

func newRecorderWithWriter(w writer, cfg Config, app string, m Metrics) *recorder {
	cfg = normalizeConfig(cfg)
	if m == nil {
		m = noopMetrics{}
	}
	r := &recorder{
		writer:  w,
		cfg:     cfg,
		metrics: m,
		app:     normalizeApp(app),
		ch:      make(chan *Event, cfg.QueueSize),
		done:    make(chan struct{}),
	}
	r.wg.Add(1)
	go r.worker()
	return r
}

func normalizeConfig(cfg Config) Config {
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 16384
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1024
	}
	if cfg.CleanupBatchSize <= 0 {
		cfg.CleanupBatchSize = 1000
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = DefaultBatchSize
	}
	if cfg.BatchSize > MaxBatchSize {
		cfg.BatchSize = MaxBatchSize
	}
	if cfg.FlushIntervalMs <= 0 {
		cfg.FlushIntervalMs = DefaultFlushIntervalMs
	}
	return cfg
}

func (r *recorder) flushInterval() time.Duration {
	return time.Duration(r.cfg.FlushIntervalMs) * time.Millisecond
}

func (r *recorder) worker() {
	defer r.wg.Done()

	batch := make([]*Event, 0, r.cfg.BatchSize)
	timer := time.NewTimer(r.flushInterval())
	defer timer.Stop()

	flush := func(ctx context.Context) {
		if len(batch) == 0 {
			return
		}
		r.writeBatch(ctx, batch)
		clear(batch)
		batch = batch[:0]
	}
	resetTimer := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(r.flushInterval())
	}

	for {
		select {
		case e := <-r.ch:
			batch = append(batch, e)
			if len(batch) >= r.cfg.BatchSize {
				flush(context.Background())
				resetTimer()
			}
		case <-timer.C:
			flush(context.Background())
			timer.Reset(r.flushInterval())
		case <-r.done:
			ctx := r.shutdownContext()
			for {
				select {
				case e := <-r.ch:
					batch = append(batch, e)
					if len(batch) >= r.cfg.BatchSize {
						flush(ctx)
					}
				default:
					flush(ctx)
					return
				}
			}
		}
	}
}

func (r *recorder) shutdownContext() context.Context {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.shutdownCtx != nil {
		return r.shutdownCtx
	}
	return context.Background()
}

func (r *recorder) Record(ctx context.Context, e Event) error {
	_ = ctx
	if !r.cfg.Enabled {
		return nil
	}
	e.App = r.app
	sanitizeEvent(&e, r.cfg.MaxBodyBytes)
	return r.enqueue(&e)
}

func (r *recorder) RecordBatch(ctx context.Context, events []Event) error {
	for i := range events {
		if err := r.Record(ctx, events[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *recorder) enqueue(e *Event) error {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		r.metrics.Dropped()
		logging.V(1).InfoS("Audit recorder is closing, dropping log", "path", e.Path)
		return ErrClosed
	}
	select {
	case r.ch <- e:
		r.mu.RUnlock()
		return nil
	case <-r.done:
		r.mu.RUnlock()
		r.metrics.Dropped()
		return ErrClosed
	default:
		r.mu.RUnlock()
		r.metrics.Dropped()
		logging.V(1).InfoS("Audit queue is full, dropping log", "path", e.Path)
		return ErrQueueFull
	}
}

func (r *recorder) writeBatch(ctx context.Context, batch []*Event) {
	if len(batch) == 0 {
		return
	}
	start := time.Now()
	rows := len(batch)

	if err := r.execBatch(ctx, batch); err == nil {
		r.metrics.BatchFlushed("ok", rows, time.Since(start))
		return
	} else {
		logging.V(1).InfoS("Failed to insert audit batch, retrying", "err", err, "rows", rows)
	}

	if ctx.Err() != nil {
		r.metrics.BatchFlushed("dropped", rows, time.Since(start))
		r.metrics.BatchFailed(rows)
		logging.V(1).InfoS("Audit batch dropped during shutdown", "rows", rows, "err", ctx.Err())
		return
	}

	if err := r.execBatch(ctx, batch); err == nil {
		r.metrics.BatchFlushed("retry_ok", rows, time.Since(start))
		return
	}

	r.metrics.BatchFallback()
	failed := r.fallbackInsert(ctx, batch)
	r.metrics.BatchFailed(failed)
	r.metrics.BatchFlushed("fallback", rows, time.Since(start))
	if failed > 0 {
		logging.V(1).InfoS("Audit batch fallback completed with failures", "rows", rows, "failed", failed)
	}
}

func (r *recorder) execBatch(ctx context.Context, batch []*Event) error {
	wctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	return r.writer.InsertBatch(wctx, batch)
}

// fallbackInsert salvages a failed batch by inserting rows one by one. The
// whole fallback is bounded so a database outage cannot block the worker for
// rows*timeout.
func (r *recorder) fallbackInsert(ctx context.Context, batch []*Event) int {
	fctx, cancel := context.WithTimeout(ctx, fallbackTimeout)
	defer cancel()

	failed := 0
	for _, e := range batch {
		if fctx.Err() != nil {
			failed++
			continue
		}
		if err := r.writer.Insert(fctx, e); err != nil {
			failed++
		}
	}
	return failed
}

// Close stops the worker after draining pending logs. The context bounds how
// long shutdown may spend flushing; remaining logs are dropped on timeout.
func (r *recorder) Close(ctx context.Context) {
	r.mu.Lock()
	if !r.closed {
		r.closed = true
		r.shutdownCtx = ctx
		close(r.done)
	}
	r.mu.Unlock()

	waited := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(waited)
	}()
	select {
	case <-waited:
	case <-ctx.Done():
		logging.Warnf("Audit shutdown timed out; some logs may be dropped")
	}
}

func sanitizeEvent(e *Event, maxBody int) {
	if e.ID == "" {
		e.ID = ulid.Make().String()
	}
	if e.CreatedTime.IsZero() {
		e.CreatedTime = time.Now()
	}
	e.App = clamp(e.App, 32)
	e.TraceID = clamp(e.TraceID, 64)
	e.UserID = clamp(e.UserID, 32)
	e.UserName = clamp(e.UserName, 64)
	e.OperationType = clamp(e.OperationType, 16)
	e.OperationDesc = clamp(e.OperationDesc, 128)
	e.Method = clamp(e.Method, 16)
	e.Path = clamp(e.Path, 255)
	e.Query = clamp(e.Query, 1024)
	e.IP = clamp(e.IP, 64)
	e.UserAgent = clamp(e.UserAgent, 512)
	e.ErrorMsg = clamp(e.ErrorMsg, 512)
	e.RequestBody = sanitizeBody(e.RequestBody, maxBody)
	e.ResponseBody = sanitizeBody(e.ResponseBody, maxBody)
}

func sanitizeBody(body string, limit int) string {
	if body == "" {
		return ""
	}
	if strings.HasPrefix(body, "<") && strings.HasSuffix(body, ">") {
		// Keep capture markers such as "<non-json body omitted>".
		return truncate(body, limit)
	}
	masked, ok := MaskJSON([]byte(body))
	if !ok {
		return "<non-json body omitted>"
	}
	return truncate(masked, limit)
}

func clamp(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max])
}
