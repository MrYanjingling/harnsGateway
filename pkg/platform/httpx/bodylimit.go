package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodyLimit caps the request body size. Requests exceeding the limit fail when
// the body is read (e.g. JSON binding returns an error).
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil && maxBytes > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
