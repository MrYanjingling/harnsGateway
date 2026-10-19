package httpx

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	traceContextKey = "httpx.traceId"
	// TraceHeader is the request/response header carrying the trace id.
	TraceHeader = "X-Trace-Id"
)

// Trace assigns (or reuses) a trace id for the request and echoes it back in
// the response header.
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.Set(traceContextKey, traceID)
		c.Header(TraceHeader, traceID)
		c.Next()
	}
}

// TraceID returns the trace id of the current request.
func TraceID(c *gin.Context) string {
	if v, ok := c.Get(traceContextKey); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
