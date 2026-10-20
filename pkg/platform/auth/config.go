package auth

// Config configures authentication (JWT + Redis session).
type Config struct {
	TokenHeader       string `json:"token-header"`
	TokenSecret       string `json:"token-secret"`
	ExpireMinutes     int    `json:"expire-minutes"`
	RefreshMinutes    int    `json:"refresh-minutes"`
	MaxLoginFailed    int    `json:"max-login-failed"`
	LockMinutes       int    `json:"lock-minutes"`
	RSAEnabled        bool   `json:"rsa-enabled"`
	CaptchaTTLSeconds int    `json:"captcha-ttl-seconds"`
	SuperRoleCode     string `json:"super-role-code"`
}

// DefaultConfig returns the default authentication configuration.
func DefaultConfig() Config {
	return Config{
		TokenHeader:       "Authorization",
		TokenSecret:       "please-change-me-32-bytes-or-more",
		ExpireMinutes:     60,
		RefreshMinutes:    10,
		MaxLoginFailed:    5,
		LockMinutes:       60,
		RSAEnabled:        true,
		CaptchaTTLSeconds: 300,
		SuperRoleCode:     "admin",
	}
}
