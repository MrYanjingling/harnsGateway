package auditapi

import (
	"github.com/gin-gonic/gin"

	"harns-platform/pkg/audit"
	"harns-platform/pkg/audit/ginmw"
	"harns-platform/pkg/generic"
	"harns-platform/pkg/platform/principal"
)

var platformExcludePaths = []string{
	"/health",
	"/auth/publicKey",
	"/auth/verificationCode",
}

// Middleware builds the platform audit middleware bound to AppIOT.
func Middleware(recorder audit.Recorder, cfg audit.Config) gin.HandlerFunc {
	return ginmw.Middleware(ginmw.Options{
		Recorder:     recorder,
		Enabled:      cfg.Enabled,
		MaxBodyBytes: cfg.MaxBodyBytes,
		ExcludePaths: platformExcludePaths,
		User: func(c *gin.Context) (string, string) {
			user, _ := principal.FromContext(c.Request.Context())
			if user == nil {
				return "", ""
			}
			return user.ID, user.UserName
		},
		TraceID: generic.TraceID,
	})
}
