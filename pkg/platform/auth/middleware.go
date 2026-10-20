package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/platform/principal"
)

const sessionContextKey = "auth.session"

// Auth authenticates the request via JWT + Redis session, performs sliding
// expiry refresh and injects the authenticated user into the request context.
func (s *Service) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c.GetHeader(s.cfg.TokenHeader))
		if token == "" {
			response.WriteError(c, http.StatusUnauthorized, response.ErrLoginExpired())
			c.Abort()
			return
		}

		sessionID, _, err := s.jwt.Parse(token)
		if err != nil {
			response.WriteError(c, http.StatusUnauthorized, response.ErrLoginExpired())
			c.Abort()
			return
		}

		sess, err := s.sessions.Get(c.Request.Context(), sessionID)
		if err != nil || sess == nil {
			response.WriteError(c, http.StatusUnauthorized, response.ErrLoginExpired())
			c.Abort()
			return
		}

		if err := s.sessions.RefreshIfNeeded(c.Request.Context(), sess); err != nil {
			response.WriteError(c, http.StatusUnauthorized, response.ErrLoginExpired())
			c.Abort()
			return
		}

		c.Set(sessionContextKey, sess)
		if sess.User != nil {
			c.Request = c.Request.WithContext(principal.With(c.Request.Context(), sess.User))
		}
		c.Next()
	}
}

func sessionFromContext(c *gin.Context) *Session {
	if v, ok := c.Get(sessionContextKey); ok {
		if sess, ok := v.(*Session); ok {
			return sess
		}
	}
	return nil
}

func extractToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	if len(header) > 7 && strings.EqualFold(header[:7], "Bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return header
}
