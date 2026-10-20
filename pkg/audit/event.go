package audit

import (
	"time"

	"harns-platform/pkg/apis"
)

// Event is a persisted audit log entry.
type Event struct {
	ID            string    `json:"id"`
	App           string    `json:"app"`
	TraceID       string    `json:"traceId"`
	UserID        string    `json:"userId"`
	UserName      string    `json:"userName"`
	OperationType string    `json:"operationType"`
	OperationDesc string    `json:"operationDesc"`
	Method        string    `json:"method"`
	Path          string    `json:"path"`
	Query         string    `json:"query"`
	IP            string    `json:"ip"`
	UserAgent     string    `json:"userAgent"`
	RequestBody   string    `json:"requestBody,omitempty"`
	ResponseBody  string    `json:"responseBody,omitempty"`
	Status        int       `json:"status"`
	ResultCode    int       `json:"resultCode"`
	ErrorMsg      string    `json:"errorMsg"`
	DurationMs    int64     `json:"durationMs"`
	CreatedTime   time.Time `json:"createdTime"`
}

// Query is the audit log query.
type Query struct {
	apis.PageQuery
	App        string `json:"app" form:"app"`
	UserID     string `json:"userId" form:"userId"`
	Path       string `json:"path" form:"path"`
	Method     string `json:"method" form:"method"`
	ResultCode *int   `json:"resultCode" form:"resultCode"`
	Start      string `json:"start" form:"start"`
	End        string `json:"end" form:"end"`
}

// Normalize applies paging defaults and canonicalizes the app value.
func (q *Query) Normalize() {
	q.PageQuery.Normalize()
	q.App = normalizeApp(q.App)
}
