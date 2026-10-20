package audit

import (
	"context"
	"strings"
	"testing"
)

func TestBuildInsertBatch(t *testing.T) {
	events := []*Event{{ID: "1"}, {ID: "2"}, {ID: "3"}}

	query, args := buildInsertBatch(events)

	if !strings.HasPrefix(query, auditInsertPrefix) {
		t.Fatalf("query does not start with INSERT prefix: %s", query)
	}
	if !strings.HasSuffix(query, auditInsertSuffix) {
		t.Fatalf("query does not end with idempotency suffix: %s", query)
	}
	if got, want := strings.Count(query, "?"), len(events)*auditColumnCount; got != want {
		t.Fatalf("placeholder count = %d, want %d", got, want)
	}
	if got, want := len(args), len(events)*auditColumnCount; got != want {
		t.Fatalf("arg count = %d, want %d", got, want)
	}
	if got, want := strings.Count(query, "),("), len(events)-1; got != want {
		t.Fatalf("row separator count = %d, want %d", got, want)
	}
}

func TestBuildInsertBatchEmpty(t *testing.T) {
	query, args := buildInsertBatch(nil)
	if query != auditInsertPrefix+auditInsertSuffix {
		t.Fatalf("query = %q", query)
	}
	if len(args) != 0 {
		t.Fatalf("args = %d, want 0", len(args))
	}
}

func TestInsertBatchEmptyDoesNotTouchDB(t *testing.T) {
	st := &store{}
	if err := st.InsertBatch(context.Background(), nil); err != nil {
		t.Fatalf("InsertBatch(nil) error = %v", err)
	}
	if err := st.InsertBatch(context.Background(), []*Event{nil, nil}); err != nil {
		t.Fatalf("InsertBatch(all nil) error = %v", err)
	}
}

func TestEventValuesCount(t *testing.T) {
	if got := len(eventValues(&Event{})); got != auditColumnCount {
		t.Fatalf("eventValues count = %d, want %d", got, auditColumnCount)
	}
}
