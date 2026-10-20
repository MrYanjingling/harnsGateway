package audit

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// store persists and queries audit logs. It is intentionally unexported so
// other modules depend on Recorder/Reader rather than the storage details.
type store struct {
	db *sql.DB
}

func newStore(db *sql.DB) *store {
	return &store{db: db}
}

const auditColumns = `id, app, trace_id, user_id, user_name, operation_type, operation_desc,
 method, path, query, ip, user_agent, request_body, response_body,
 status, result_code, error_msg, duration_ms, created_time`

const (
	auditInsertPrefix   = `INSERT INTO audit_logs (` + auditColumns + `) VALUES `
	auditRowPlaceholder = `(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	auditInsertSuffix   = ` ON DUPLICATE KEY UPDATE id = id`
	auditColumnCount    = 19
)

// Insert writes a single audit log entry.
func (s *store) Insert(ctx context.Context, e *Event) error {
	_, err := s.db.ExecContext(ctx, auditInsertPrefix+auditRowPlaceholder+auditInsertSuffix, eventValues(e)...)
	return err
}

// InsertBatch writes multiple audit log entries in one multi-row INSERT.
func (s *store) InsertBatch(ctx context.Context, events []*Event) error {
	if len(events) == 0 {
		return nil
	}
	rows := make([]*Event, 0, len(events))
	for _, e := range events {
		if e != nil {
			rows = append(rows, e)
		}
	}
	if len(rows) == 0 {
		return nil
	}
	query, args := buildInsertBatch(rows)
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func buildInsertBatch(events []*Event) (string, []any) {
	var b strings.Builder
	b.Grow(len(auditInsertPrefix) + len(events)*(len(auditRowPlaceholder)+1) + len(auditInsertSuffix))
	b.WriteString(auditInsertPrefix)
	args := make([]any, 0, len(events)*auditColumnCount)
	for i, e := range events {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(auditRowPlaceholder)
		args = append(args, eventValues(e)...)
	}
	b.WriteString(auditInsertSuffix)
	return b.String(), args
}

func eventValues(e *Event) []any {
	return []any{
		e.ID, e.App, e.TraceID, e.UserID, e.UserName, e.OperationType, e.OperationDesc,
		e.Method, e.Path, e.Query, e.IP, e.UserAgent, e.RequestBody, e.ResponseBody,
		e.Status, e.ResultCode, e.ErrorMsg, e.DurationMs, e.CreatedTime,
	}
}

// List returns audit logs filtered by the query.
func (s *store) List(ctx context.Context, q Query) ([]*Event, int64, error) {
	where, args := "1=1", []any{}
	if q.App != "" {
		where += " AND app = ?"
		args = append(args, q.App)
	}
	if q.UserID != "" {
		where += " AND user_id = ?"
		args = append(args, q.UserID)
	}
	if q.Path != "" {
		where += " AND path LIKE ?"
		args = append(args, "%"+q.Path+"%")
	}
	if q.Method != "" {
		where += " AND method = ?"
		args = append(args, strings.ToUpper(q.Method))
	}
	if q.ResultCode != nil {
		where += " AND result_code = ?"
		args = append(args, *q.ResultCode)
	}
	if t, ok := parseTime(q.Start); ok {
		where += " AND created_time >= ?"
		args = append(args, t)
	}
	if t, ok := parseTime(q.End); ok {
		where += " AND created_time <= ?"
		args = append(args, t)
	}

	var total int64
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM audit_logs WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT "+auditColumns+" FROM audit_logs WHERE "+where+" ORDER BY created_time DESC LIMIT ? OFFSET ?",
		append(args, q.Limit(), q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*Event, 0, q.Limit())
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}
	return list, total, rows.Err()
}

// DeleteBefore removes audit logs older than the given time in batches to
// avoid long locks on large tables. It returns the total number of rows deleted.
func (s *store) DeleteBefore(ctx context.Context, before time.Time, batchSize int) (int64, error) {
	if batchSize <= 0 {
		batchSize = 1000
	}
	var total int64
	for {
		res, err := s.db.ExecContext(ctx,
			`DELETE FROM audit_logs WHERE created_time < ? LIMIT ?`, before, batchSize)
		if err != nil {
			return total, err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return total, err
		}
		total += n
		if n < int64(batchSize) {
			return total, nil
		}
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}
	}
}

type scanner interface{ Scan(dest ...any) error }

func scanEvent(sc scanner) (*Event, error) {
	e := &Event{}
	var requestBody, responseBody sql.NullString
	err := sc.Scan(&e.ID, &e.App, &e.TraceID, &e.UserID, &e.UserName, &e.OperationType, &e.OperationDesc,
		&e.Method, &e.Path, &e.Query, &e.IP, &e.UserAgent, &requestBody, &responseBody,
		&e.Status, &e.ResultCode, &e.ErrorMsg, &e.DurationMs, &e.CreatedTime)
	if err != nil {
		return nil, err
	}
	e.RequestBody = requestBody.String
	e.ResponseBody = responseBody.String
	return e, nil
}

func parseTime(v string) (time.Time, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339} {
		if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
