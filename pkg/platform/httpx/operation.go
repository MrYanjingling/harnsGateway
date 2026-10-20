package httpx

import (
	"github.com/gin-gonic/gin"

	"harns-platform/pkg/audit"
	"harns-platform/pkg/audit/ginmw"
)

// Operation types recorded by the audit log.
const (
	OpLogin  = audit.OpLogin
	OpLogout = audit.OpLogout
	OpCreate = audit.OpCreate
	OpUpdate = audit.OpUpdate
	OpDelete = audit.OpDelete
	OpQuery  = audit.OpQuery
)

// Operation describes what a route does.
type Operation = audit.Operation

// WithOperation declares the operation metadata of a route explicitly.
func WithOperation(opType, desc string) gin.HandlerFunc {
	return ginmw.WithOperation(opType, desc)
}

// GetOperation returns the operation metadata set by WithOperation.
func GetOperation(c *gin.Context) audit.Operation {
	return ginmw.GetOperation(c)
}
