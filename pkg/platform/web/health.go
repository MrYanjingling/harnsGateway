package web

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"harns-platform/cmd/platform/config"
	"harns-platform/pkg/apis/response"
)

// readiness checks database, redis and minio connectivity.
func readiness(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		checkCtx := func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(c.Request.Context(), 2*time.Second)
		}

		checks := gin.H{}
		ok := true

		if cfg.DB != nil {
			ctx, cancel := checkCtx()
			err := cfg.DB.PingContext(ctx)
			cancel()
			if err != nil {
				checks["db"] = "down"
				ok = false
			} else {
				checks["db"] = "ok"
			}
		} else {
			checks["db"] = "disabled"
		}

		if cfg.Redis != nil {
			ctx, cancel := checkCtx()
			err := cfg.Redis.Ping(ctx).Err()
			cancel()
			if err != nil {
				checks["redis"] = "down"
				ok = false
			} else {
				checks["redis"] = "ok"
			}
		} else {
			checks["redis"] = "disabled"
		}

		if cfg.MinIO != nil && cfg.Security.RSAEnabled {
			ctx, cancel := checkCtx()
			_, err := cfg.MinIO.BucketExists(ctx, cfg.MinIOBucket)
			cancel()
			if err != nil {
				checks["minio"] = "down"
				ok = false
			} else {
				checks["minio"] = "ok"
			}
		} else {
			checks["minio"] = "disabled"
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
