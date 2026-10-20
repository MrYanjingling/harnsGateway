package httpx

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"harns-platform/pkg/logging"
)

const (
	traceContextKey = "httpx.traceId"
	// TraceHeader is the request/response header carrying the trace id.
	TraceHeader = "X-Trace-Id"
)

// Trace assigns (or reuses) a trace id for the request, echoes it back in the
// response header and records the request outcome.
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.Set(traceContextKey, traceID)
		c.Header(TraceHeader, traceID)

		start := time.Now()
		c.Next()

		logging.V(3).InfoS("http request",
			"traceId", traceID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"ip", c.ClientIP(),
			"cost", time.Since(start),
		)
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
