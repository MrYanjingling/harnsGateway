package audit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"

	"harns-platform/pkg/platform/httpx"
	"harns-platform/pkg/platform/principal"
)

var excludedPathSuffixes = []string{
	"/auth/publicKey",
	"/auth/verificationCode",
	"/health",
}

// Middleware captures request/response and asynchronously persists an audit log
// for write operations.
func (s *Service) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.shouldAudit(c) {
			c.Next()
			return
		}

		start := time.Now()
		requestBody := s.captureRequest(c)

		cw := &captureWriter{ResponseWriter: c.Writer, limit: s.cfg.MaxBodyBytes}
		c.Writer = cw

		c.Next()

		duration := time.Since(start)
		op := httpx.GetOperation(c)
		app := httpx.AppFromPath(c.Request.URL.Path)
		if app == "" {
			app = "platform"
		}
		p, _ := principal.FromContext(c.Request.Context())

		resultCode, errorMsg := parseResult(cw.body.Bytes(), cw.truncated)

		responseBody := ""
		if cw.truncated {
			responseBody = "<response body exceeds limit>"
		} else if masked, ok := MaskJSON(cw.body.Bytes()); ok {
			responseBody = truncate(masked, s.cfg.MaxBodyBytes)
		} else {
			responseBody = "<non-json body omitted>"
		}

		s.enqueue(&Log{
			ID:            ulid.Make().String(),
			App:           app,
			TraceID:       httpx.TraceID(c),
			UserID:        p.UserID,
			UserName:      p.UserName,
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
		})
	}
}

func (s *Service) shouldAudit(c *gin.Context) bool {
	if !s.cfg.Enabled {
		return false
	}
	switch c.Request.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return false
	}
	path := c.Request.URL.Path
	for _, suffix := range excludedPathSuffixes {
		if strings.HasSuffix(path, suffix) {
			return false
		}
	}
	return true
}

func (s *Service) captureRequest(c *gin.Context) string {
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
	masked, ok := MaskJSON(raw)
	if !ok {
		return "<non-json body omitted>"
	}
	return truncate(masked, s.cfg.MaxBodyBytes)
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
	if w.truncated || w.limit <= 0 {
		return
	}
	remain := w.limit - w.body.Len()
	if len(b) > remain {
		w.body.Write(b[:remain])
		w.truncated = true
		return
	}
	w.body.Write(b)
}
