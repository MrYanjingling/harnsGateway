package gateway

import (
	"fmt"

	"github.com/MrYanjingling/harns-platform/pkg/apis"
	"github.com/MrYanjingling/harns-platform/pkg/apis/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

func InstallHandler(group *gin.RouterGroup, mgr *Manager) {
	group.GET("/gatewayMeta", getGatewayMeta(mgr))
	group.GET("/gatewayCpu", getGatewayCpu(mgr))
	group.GET("/gatewayMem", getGatewayMem(mgr))
	group.GET("/gatewayDisk", getGatewayDisk(mgr))
}

func getGatewayMeta(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		g, _ := mgr.GetGatewayMeta()
		c.Header(apis.ETag, fmt.Sprintf("%s", g.GetVersion()))
		response.WriteOK(c, g)
	}
}

func getGatewayCpu(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cpu, err := mgr.getGatewayCpu()
		if err != nil {
			response.WriteError(c, http.StatusInternalServerError, response.ErrInternal(err))
			return
		}
		response.WriteOK(c, ResponseModel{Cpus: cpu})
	}
}

func getGatewayMem(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		mem, err := mgr.getGatewayMem()
		if err != nil {
			response.WriteError(c, http.StatusInternalServerError, response.ErrInternal(err))
			return
		}
		response.WriteOK(c, ResponseModel{Mem: mem})
	}
}

func getGatewayDisk(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		disks, err := mgr.getGatewayDisk()
		if err != nil {
			response.WriteError(c, http.StatusInternalServerError, response.ErrInternal(err))
			return
		}
		response.WriteOK(c, ResponseModel{Disks: disks})
	}
}
