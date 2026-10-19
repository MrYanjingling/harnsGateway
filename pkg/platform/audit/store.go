package audit

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// Store persists and queries audit logs.
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

const auditColumns = `id, app, trace_id, user_id, user_name, operation_type, operation_desc,
 method, path, query, ip, user_agent, request_body, response_body,
 status, result_code, error_msg, duration_ms, created_time`

// Insert writes a single audit log entry.
func (s *Store) Insert(ctx context.Context, l *Log) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_logs (`+auditColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		l.ID, l.App, l.TraceID, l.UserID, l.UserName, l.OperationType, l.OperationDesc,
		l.Method, l.Path, l.Query, l.IP, l.UserAgent, l.RequestBody, l.ResponseBody,
		l.Status, l.ResultCode, l.ErrorMsg, l.DurationMs, l.CreatedTime)
	return err
}

// List returns audit logs filtered by the query.
func (s *Store) List(ctx context.Context, q ListQuery) ([]*Log, int64, error) {
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

	list := make([]*Log, 0, q.Limit())
	for rows.Next() {
		l, err := scanLog(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, l)
	}
	return list, total, rows.Err()
}

type scanner interface{ Scan(dest ...any) error }

func scanLog(sc scanner) (*Log, error) {
	l := &Log{}
	var requestBody, responseBody sql.NullString
	err := sc.Scan(&l.ID, &l.App, &l.TraceID, &l.UserID, &l.UserName, &l.OperationType, &l.OperationDesc,
		&l.Method, &l.Path, &l.Query, &l.IP, &l.UserAgent, &requestBody, &responseBody,
		&l.Status, &l.ResultCode, &l.ErrorMsg, &l.DurationMs, &l.CreatedTime)
	if err != nil {
		return nil, err
	}
	l.RequestBody = requestBody.String
	l.ResponseBody = responseBody.String
	return l, nil
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
