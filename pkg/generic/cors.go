package generic

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig configures the CORS middleware.
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed-origins"`
	AllowedMethods   []string `json:"allowed-methods"`
	AllowedHeaders   []string `json:"allowed-headers"`
	ExposeHeaders    []string `json:"expose-headers"`
	AllowCredentials bool     `json:"allow-credentials"`
	MaxAgeSeconds    int      `json:"max-age-seconds"`
}

// DefaultCORSConfig returns a permissive development-friendly CORS config.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders: []string{"Authorization", "Content-Type", "X-Trace-Id"},
		ExposeHeaders:  []string{"X-Trace-Id"},
		MaxAgeSeconds:  3600,
	}
}

// CORS applies CORS headers and handles preflight requests.
func CORS(cfg CORSConfig) gin.HandlerFunc {
	allowAll := containsString(cfg.AllowedOrigins, "*")
	allowedMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowedHeaders := strings.Join(cfg.AllowedHeaders, ", ")
	exposeHeaders := strings.Join(cfg.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(cfg.MaxAgeSeconds)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			switch {
			case allowAll && !cfg.AllowCredentials:
				c.Header("Access-Control-Allow-Origin", "*")
			case allowAll || containsString(cfg.AllowedOrigins, origin):
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			if exposeHeaders != "" {
				c.Header("Access-Control-Expose-Headers", exposeHeaders)
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", allowedMethods)
			if allowedHeaders != "" {
				c.Header("Access-Control-Allow-Headers", allowedHeaders)
			}
			if cfg.MaxAgeSeconds > 0 {
				c.Header("Access-Control-Max-Age", maxAge)
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
