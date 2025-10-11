package cim

import (
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	"net/http"
)

func InstallHandler(group *gin.RouterGroup, mgr *Manager) {
	// group.POST("/tagNames", importTagNames(mgr))
	group.GET("/data", getTagNames(mgr))
}

// func importTagNames(mgr *Manager) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		file, err1 := c.FormFile("file")
// 		if err1 != nil {
// 			c.JSON(http.StatusBadRequest, response.NewMultiError(response.ErrMalformedJSON))
// 		}
// 		if err := mgr.ImportTagNames(file); err != nil {
// 			c.Status(http.StatusInternalServerError)
// 			return
// 		}
// 		c.Status(http.StatusOK)
// 		return
// 	}
// }

func getTagNames(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Request.URL.Query()
		timeType := query.Get("type")
		time := query.Get("time")
		klog.V(2).InfoS("request data", "time", time, "type", timeType)
		go mgr.Data(timeType, time)
		c.Status(http.StatusOK)
		return
	}
}
