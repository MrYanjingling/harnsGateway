package data

import (
	"github.com/gin-gonic/gin"
	"harnsgateway/pkg/apis/response"
	"net/http"
)

func InstallHandler(group *gin.RouterGroup, mgr *Manager) {
	group.POST("/tagNames", importTagNames(mgr))
	group.GET("/tagNames", getTagNames(mgr))
}

func importTagNames(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err1 := c.FormFile("file")

		if err1 != nil {
			c.JSON(http.StatusBadRequest, response.NewMultiError(response.ErrMalformedJSON))
		}
		if err := mgr.ImportTagNames(file); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
		return
	}
}

func getTagNames(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tagNames := mgr.ListTagNames()
		c.JSON(http.StatusOK, ResponseModel{TagNames: tagNames})
	}
}
