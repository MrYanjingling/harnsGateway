package audit

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/apis"
	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/logging"
)

// List handles GET /audit-logs.
func (s *Service) List(c *gin.Context) {
	var q ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.WriteError(c, http.StatusBadRequest, response.ErrMalformedJSON)
		return
	}
	q.Normalize()

	list, total, err := s.store.List(c.Request.Context(), q)
	if err != nil {
		logging.V(2).InfoS("Failed to list audit logs", "err", err)
		response.WriteError(c, http.StatusInternalServerError, response.ErrInternal(err))
		return
	}
	response.WriteOK(c, apis.NewPageResp(list, total, q.PageQuery))
}
