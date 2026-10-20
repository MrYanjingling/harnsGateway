package audit

import (
	"context"
	"database/sql"
)

// Reader queries audit logs. A reader created with a non-empty app is scoped
// to that app and overrides any app supplied by the caller.
type Reader interface {
	List(ctx context.Context, q Query) ([]*Event, int64, error)
}

type reader struct {
	store *store
	app   string
}

// NewReader creates a reader. Pass an empty app to query all applications.
func NewReader(db *sql.DB, app string) Reader {
	return &reader{store: newStore(db), app: normalizeApp(app)}
}

func (r *reader) List(ctx context.Context, q Query) ([]*Event, int64, error) {
	return r.store.List(ctx, r.scope(q))
}

func (r *reader) scope(q Query) Query {
	if r.app != "" {
		q.App = r.app
	}
	q.Normalize()
	return q
}
