package auditapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/apis"
	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/audit"
	"harns-platform/pkg/logging"
	"harns-platform/pkg/platform/httpx"
)

// List returns the audit log query handler.
func List(reader audit.Reader) gin.HandlerFunc {
	return func(c *gin.Context) {
		var q audit.Query
		if !httpx.BindQuery(c, &q) {
			return
		}
		q.Normalize()

		list, total, err := reader.List(c.Request.Context(), q)
		if err != nil {
			logging.V(2).InfoS("Failed to list audit logs", "err", err)
			response.WriteError(c, http.StatusInternalServerError, response.ErrInternal(err))
			return
		}
		response.WriteOK(c, apis.NewPageResp(list, total, q.PageQuery))
	}
}
