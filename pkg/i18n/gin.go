package i18n

import "github.com/gin-gonic/gin"

const (
	contextKey   = "harns-lang"
	langQueryKey = "lang"
)

// Middleware resolves the request language from the "lang" query parameter
// (which overrides) or the Accept-Language header, and stores it in the gin
// context for later use by Lang.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Query(langQueryKey)
		if lang == "" {
			lang = c.GetHeader("Accept-Language")
		}
		c.Set(contextKey, Match(lang))
		c.Next()
	}
}

// Lang returns the resolved language for the current request, falling back to
// the default language when the middleware has not run.
func Lang(c *gin.Context) string {
	if v, ok := c.Get(contextKey); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return defaultLang
}
