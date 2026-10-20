package audit

import (
	"context"
	"database/sql"
)

// Recorder writes audit events. Implementations must be safe for concurrent
// use and must not block the caller on database I/O.
type Recorder interface {
	Record(ctx context.Context, event Event) error
	RecordBatch(ctx context.Context, events []Event) error
	Close(ctx context.Context) error
}

// Option customizes a Module.
type Option func(*moduleOptions)

type moduleOptions struct {
	metrics Metrics
}

// WithMetrics injects an optional metrics implementation.
func WithMetrics(m Metrics) Option {
	return func(o *moduleOptions) {
		o.metrics = m
	}
}

// Module binds a Recorder and a scoped Reader to one application.
type Module struct {
	app      string
	recorder *recorder
	reader   *reader
}

// NewModule creates an audit module for the given application.
func NewModule(db *sql.DB, cfg Config, app string, opts ...Option) *Module {
	var options moduleOptions
	for _, opt := range opts {
		opt(&options)
	}
	st := newStore(db)
	m := &Module{
		app:      normalizeApp(app),
		recorder: newRecorder(st, cfg, app, options.metrics),
	}
	m.reader = &reader{store: st, app: m.app}
	return m
}

// App returns the application this module is bound to.
func (m *Module) App() string {
	return m.app
}

// Record writes one audit event.
func (m *Module) Record(ctx context.Context, event Event) error {
	return m.recorder.Record(ctx, event)
}

// RecordBatch writes multiple audit events.
func (m *Module) RecordBatch(ctx context.Context, events []Event) error {
	return m.recorder.RecordBatch(ctx, events)
}

// List queries audit events scoped to this module's application.
func (m *Module) List(ctx context.Context, q Query) ([]*Event, int64, error) {
	return m.reader.List(ctx, q)
}

// Close drains and stops the recorder.
func (m *Module) Close(ctx context.Context) error {
	m.recorder.Close(ctx)
	return nil
}
