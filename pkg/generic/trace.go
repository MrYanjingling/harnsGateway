package generic

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// TraceHeader is the request/response header carrying the trace id.
	TraceHeader = "X-Trace-Id"

	traceContextKey = "harns.traceId"
)

type traceContextKeyType struct{}

var traceCtxKey traceContextKeyType

// Trace assigns (or reuses) a trace id for the request, stores it in the
// request context and echoes it back in the response header.
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.Set(traceContextKey, traceID)
		c.Header(TraceHeader, traceID)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), traceCtxKey, traceID))
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

// TraceIDFromContext returns the trace id stored in the request context.
func TraceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(traceCtxKey).(string); ok {
		return id
	}
	return ""
}

// SetTraceHeader propagates the trace id from ctx to an outbound request.
func SetTraceHeader(ctx context.Context, req *http.Request) {
	if req == nil {
		return
	}
	if id := TraceIDFromContext(ctx); id != "" {
		req.Header.Set(TraceHeader, id)
	}
}
