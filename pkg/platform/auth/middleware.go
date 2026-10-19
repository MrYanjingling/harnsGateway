package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/logging"
	"harns-platform/pkg/platform/httpx"
	"harns-platform/pkg/platform/principal"
)

const sessionContextKey = "auth.session"

// Auth authenticates the request via JWT + Redis session, performs sliding
// expiry refresh and injects the principal into the request context.
func (s *Service) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := httpx.TraceID(c)
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path
		ip := c.ClientIP()

		logging.V(3).InfoS("auth enter",
			"traceId", traceID, "method", method, "path", path, "ip", ip)

		fail := func(reason string) {
			logging.V(2).InfoS("auth token validation failed",
				"traceId", traceID, "method", method, "path", path, "ip", ip,
				"reason", reason, "cost", time.Since(start))
			response.WriteError(c, http.StatusUnauthorized, response.ErrLoginExpired())
			c.Abort()
		}

		token := extractToken(c.GetHeader(s.cfg.TokenHeader))
		if token == "" {
			fail("missing token")
			return
		}

		sessionID, userID, err := s.jwt.Parse(token)
		if err != nil {
			fail("invalid jwt")
			return
		}

		sess, err := s.sessions.Get(c.Request.Context(), sessionID)
		if err != nil {
			fail("session store error")
			return
		}
		if sess == nil {
			fail("session not found")
			return
		}

		if userID == "" && sess.User != nil {
			userID = sess.User.ID
		}

		if err := s.sessions.RefreshIfNeeded(c.Request.Context(), sess); err != nil {
			fail("session refresh error")
			return
		}

		c.Set(sessionContextKey, sess)
		if sess.User != nil {
			ctx := principal.With(c.Request.Context(), principal.Principal{
				UserID:   sess.User.ID,
				UserName: sess.User.UserName,
			})
			c.Request = c.Request.WithContext(ctx)
		}
		logging.V(3).InfoS("auth token validated",
			"traceId", traceID, "userId", userID, "method", method, "path", path, "ip", ip,
			"cost", time.Since(start))
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
