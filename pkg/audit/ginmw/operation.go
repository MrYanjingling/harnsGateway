package ginmw

import (
	"github.com/gin-gonic/gin"

	"harns-platform/pkg/audit"
)

const operationContextKey = "audit.operation"

// WithOperation declares the business operation of a route.
func WithOperation(opType, desc string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(operationContextKey, audit.Operation{Type: opType, Desc: desc})
		c.Next()
	}
}

// GetOperation returns the operation declared by WithOperation.
func GetOperation(c *gin.Context) audit.Operation {
	if v, ok := c.Get(operationContextKey); ok {
		if op, ok := v.(audit.Operation); ok {
			return op
		}
	}
	return audit.Operation{}
}
