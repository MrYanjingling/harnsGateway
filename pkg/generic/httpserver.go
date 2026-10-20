package generic

import (
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"harns-platform/pkg/i18n"
	"harns-platform/pkg/logging"
	"harns-platform/pkg/metrics"
)

func Default() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(i18n.Middleware(), Trace(), accessLogger(), securityHeaders(), metrics.Middleware())
	engine.Use(gzip.Gzip(gzip.DefaultCompression))
	engine.GET("/metrics", metrics.Handler())
	engine.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
	return engine
}

func accessLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logging.V(2).InfoS("http request",
			"traceId", c.Writer.Header().Get("X-Trace-Id"),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"ip", c.ClientIP(),
			"cost", time.Since(start),
		)
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}
