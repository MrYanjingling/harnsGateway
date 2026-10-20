package ginmw

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/audit"
	"harns-platform/pkg/logging"
)

var defaultMethods = []string{
	http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete,
}

// Options configures the audit middleware.
type Options struct {
	Recorder     audit.Recorder
	Enabled      bool
	MaxBodyBytes int
	Methods      []string
	ExcludePaths []string

	// User extracts the current user. Optional.
	User func(*gin.Context) (id, name string)
	// TraceID extracts the trace id. Defaults to the X-Trace-Id header.
	TraceID func(*gin.Context) string
	// Operation extracts the business operation. Defaults to GetOperation.
	Operation func(*gin.Context) audit.Operation
}

// Middleware captures write requests and records them asynchronously.
func Middleware(opts Options) gin.HandlerFunc {
	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = 16384
	}
	methods := opts.Methods
	if len(methods) == 0 {
		methods = defaultMethods
	}
	if opts.TraceID == nil {
		opts.TraceID = func(c *gin.Context) string {
			if id := c.GetHeader("X-Trace-Id"); id != "" {
				return id
			}
			return c.Writer.Header().Get("X-Trace-Id")
		}
	}
	if opts.Operation == nil {
		opts.Operation = GetOperation
	}
	methodSet := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		methodSet[strings.ToUpper(m)] = struct{}{}
	}

	return func(c *gin.Context) {
		if opts.Recorder == nil || !opts.Enabled || !shouldAudit(c, methodSet, opts.ExcludePaths) {
			c.Next()
			return
		}

		start := time.Now()
		requestBody := captureRequest(c)

		cw := &captureWriter{ResponseWriter: c.Writer, limit: opts.MaxBodyBytes}
		c.Writer = cw

		c.Next()

		duration := time.Since(start)
		responseBody := ""
		if cw.truncated {
			responseBody = "<response body exceeds limit>"
		} else {
			responseBody = cw.body.String()
		}

		resultCode, errorMsg := parseResult([]byte(responseBody), cw.truncated)

		userID, userName := "", ""
		if opts.User != nil {
			userID, userName = opts.User(c)
		}

		op := opts.Operation(c)
		if op.Type == "" {
			op.Type = fallbackOperation(c.Request.Method)
		}

		event := audit.Event{
			TraceID:       opts.TraceID(c),
			UserID:        userID,
			UserName:      userName,
			OperationType: op.Type,
			OperationDesc: op.Desc,
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			Query:         c.Request.URL.RawQuery,
			IP:            c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
			RequestBody:   requestBody,
			ResponseBody:  responseBody,
			Status:        c.Writer.Status(),
			ResultCode:    resultCode,
			ErrorMsg:      errorMsg,
			DurationMs:    duration.Milliseconds(),
			CreatedTime:   start,
		}
		if err := opts.Recorder.Record(c.Request.Context(), event); err != nil {
			logging.V(2).InfoS("Failed to record audit log", "err", err, "path", event.Path)
		}
	}
}

func shouldAudit(c *gin.Context, methods map[string]struct{}, excludes []string) bool {
	if _, ok := methods[strings.ToUpper(c.Request.Method)]; !ok {
		return false
	}
	path := c.Request.URL.Path
	for _, suffix := range excludes {
		if suffix != "" && strings.HasSuffix(path, suffix) {
			return false
		}
	}
	return true
}

func fallbackOperation(method string) string {
	switch method {
	case http.MethodPost:
		return audit.OpCreate
	case http.MethodPut, http.MethodPatch:
		return audit.OpUpdate
	case http.MethodDelete:
		return audit.OpDelete
	default:
		return audit.OpQuery
	}
}

func captureRequest(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}
	if ct := strings.ToLower(c.GetHeader("Content-Type")); !strings.Contains(ct, "application/json") {
		return "<non-json body omitted>"
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	if len(raw) == 0 {
		return ""
	}
	return string(raw)
}

func parseResult(body []byte, truncated bool) (int, string) {
	if truncated || len(body) == 0 {
		return 0, ""
	}
	var r struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return 0, ""
	}
	if r.Code == 200 {
		return r.Code, ""
	}
	return r.Code, r.Message
}

type captureWriter struct {
	gin.ResponseWriter
	body      bytes.Buffer
	limit     int
	truncated bool
}

func (w *captureWriter) Write(b []byte) (int, error) {
	w.capture(b)
	return w.ResponseWriter.Write(b)
}

func (w *captureWriter) WriteString(s string) (int, error) {
	w.capture([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

func (w *captureWriter) capture(b []byte) {
	if w.truncated {
		return
	}
	remaining := w.limit - w.body.Len()
	if remaining <= 0 {
		w.truncated = true
		return
	}
	if len(b) > remaining {
		w.body.Write(b[:remaining])
		w.truncated = true
		return
	}
	w.body.Write(b)
}
