package httpx

import "github.com/gin-gonic/gin"

// Operation types recorded by the audit log.
const (
	OpLogin  = "LOGIN"
	OpLogout = "LOGOUT"
	OpCreate = "CREATE"
	OpUpdate = "UPDATE"
	OpDelete = "DELETE"
	OpQuery  = "QUERY"
)

const operationContextKey = "httpx.operation"

// Operation describes what a route does.
type Operation struct {
	Type string
	Desc string
}

// WithOperation declares the operation metadata of a route explicitly.
func WithOperation(opType, desc string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(operationContextKey, Operation{Type: opType, Desc: desc})
		c.Next()
	}
}

// GetOperation returns the operation metadata set by WithOperation.
func GetOperation(c *gin.Context) Operation {
	if v, ok := c.Get(operationContextKey); ok {
		if op, ok := v.(Operation); ok {
			return op
		}
	}
	return Operation{}
}
