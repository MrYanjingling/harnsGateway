package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtManager struct {
	secret []byte
}

func newJWTManager(secret string) (*jwtManager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("security.token-secret must be at least 32 bytes, got %d", len(secret))
	}
	return &jwtManager{secret: []byte(secret)}, nil
}

// Sign issues a HS256 JWT carrying the session id and user id.
func (m *jwtManager) Sign(sessionID, userID string) (string, error) {
	claims := jwt.MapClaims{
		"login_user": sessionID,
		"sub":        userID,
		"iat":        time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse verifies the token signature and extracts session id / user id.
func (m *jwtManager) Parse(tokenString string) (sessionID, userID string, err error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid claims")
	}
	sessionID, _ = claims["login_user"].(string)
	userID, _ = claims["sub"].(string)
	if sessionID == "" {
		return "", "", fmt.Errorf("missing login_user claim")
	}
	return sessionID, userID, nil
}
