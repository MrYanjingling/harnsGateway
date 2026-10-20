package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/platform/principal"
)

// Permission codes.
const (
	PermUserRead  = "user:read"
	PermUserWrite = "user:write"
	PermRoleRead  = "role:read"
	PermRoleWrite = "role:write"
	PermOrgRead   = "org:read"
	PermOrgWrite  = "org:write"
	PermAuditRead = "audit:read"
	PermAll       = "*"
)

// RequirePermission allows the request only when the authenticated user has the
// given permission (or the wildcard "*").
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := principal.FromContext(c.Request.Context())
		if !ok || user == nil {
			response.WriteError(c, http.StatusUnauthorized, response.ErrUnauthorized())
			c.Abort()
			return
		}
		for _, p := range user.Permissions {
			if p == PermAll || p == perm {
				c.Next()
				return
			}
		}
		response.WriteError(c, http.StatusForbidden, response.ErrForbidden())
		c.Abort()
	}
}
