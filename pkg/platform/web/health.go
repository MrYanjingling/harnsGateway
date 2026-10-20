package web

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"harns-platform/cmd/platform/config"
	"harns-platform/pkg/apis/response"
)

const maxRequestBodyBytes int64 = 1 << 20

// readiness checks database and redis connectivity.
func readiness(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		checks := gin.H{}
		ok := true

		if cfg.DB != nil {
			if err := cfg.DB.PingContext(ctx); err != nil {
				checks["db"] = "down"
				ok = false
			} else {
				checks["db"] = "ok"
			}
		} else {
			checks["db"] = "disabled"
		}

		if cfg.Redis != nil {
			if err := cfg.Redis.Ping(ctx).Err(); err != nil {
				checks["redis"] = "down"
				ok = false
			} else {
				checks["redis"] = "ok"
			}
		} else {
			checks["redis"] = "disabled"
		}

		if ok {
			response.WriteOK(c, gin.H{"status": "ok", "checks": checks})
			return
		}
		c.JSON(http.StatusServiceUnavailable, response.Result{
			Code:    int(response.ErrCodeInternal),
			Message: "service unavailable",
			Data:    gin.H{"status": "unavailable", "checks": checks},
		})
	}
}
