package audit

import (
	"time"

	"harns-platform/pkg/apis"
)

// Log is a persisted audit log entry.
type Log struct {
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

// ListQuery is the audit log query.
type ListQuery struct {
	apis.PageQuery
	App        string `form:"app"`
	UserID     string `form:"userId"`
	Path       string `form:"path"`
	Method     string `form:"method"`
	ResultCode *int   `form:"resultCode"`
	Start      string `form:"start"`
	End        string `form:"end"`
}
