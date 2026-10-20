package httpx

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/logging"
)

// Recovery recovers from panics, logs them with the trace id and returns a
// structured JSON 500 response.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%v", r)
				logging.ErrorS(err, "panic recovered",
					"traceId", TraceID(c),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"stack", string(debug.Stack()),
				)
				if !c.Writer.Written() {
					response.WriteError(c, http.StatusInternalServerError, response.ErrInternal(err))
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
