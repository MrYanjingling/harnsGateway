package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/platform/identity"
	identityv1 "harns-platform/pkg/platform/identity/v1"
)

// Service implements login, logout, session and current-user operations.
type Service struct {
	users    *identity.Service
	sessions *sessionStore
	jwt      *jwtManager
	rsa      *rsaManager
	captcha  *base64Captcha.Captcha
	cfg      Config
}

func NewService(users *identity.Service, rdb redis.UniversalClient, mc *minio.Client, bucket string, cfg Config) (*Service, error) {
	if cfg.TokenHeader == "" {
		cfg.TokenHeader = "Authorization"
	}
	if cfg.ExpireMinutes <= 0 {
		cfg.ExpireMinutes = 60
	}
	if cfg.RefreshMinutes < 0 {
		cfg.RefreshMinutes = 0
	}
	if cfg.MaxLoginFailed <= 0 {
		cfg.MaxLoginFailed = 5
	}
	if cfg.LockMinutes <= 0 {
		cfg.LockMinutes = 60
	}

	jwtMgr, err := newJWTManager(cfg.TokenSecret)
	if err != nil {
		return nil, err
	}

	var rsaMgr *rsaManager
	if cfg.RSAEnabled {
		rsaMgr, err = newRSAManager(context.Background(), mc, bucket)
		if err != nil {
			return nil, err
		}
	}

	return &Service{
		users:    users,
		sessions: newSessionStore(rdb, cfg.RefreshMinutes),
		jwt:      jwtMgr,
		rsa:      rsaMgr,
		captcha:  newCaptchaManager(rdb, cfg.CaptchaTTLSeconds),
		cfg:      cfg,
	}, nil
}

// login validates credentials and returns a signed JWT.
func (s *Service) login(ctx context.Context, username, password, code, key, ip, userAgent string) (string, *identityv1.UserVo, error) {
	// 验证码校验（暂不启用）：启用时取消下面一行注释
	// if err := s.verifyCaptcha(key, code); err != nil {
	// 	return "", nil, err
	// }

	plain := password
	if s.rsa != nil {
		if dec, err := s.rsa.Decrypt(password); err == nil {
			plain = dec
		}
	}

	u, err := s.users.GetUserByUserName(ctx, username)
	if err != nil {
		return "", nil, err
	}

	if u.Status != identity.StatusEnabled {
		if u.LockedUntil != nil {
			if time.Now().Before(*u.LockedUntil) {
				return "", nil, response.ErrAccountLocked(*u.LockedUntil)
			}
			if err := s.users.UnlockUser(ctx, u.ID); err != nil {
				return "", nil, err
			}
			u.Status = identity.StatusEnabled
			u.LockedUntil = nil
			u.LoginFailureTimes = 0
		} else {
			return "", nil, response.ErrAccountDisabled()
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plain)); err != nil {
		_, lockedUntil, ferr := s.users.RecordLoginFailure(ctx, u.ID, s.cfg.MaxLoginFailed, s.cfg.LockMinutes)
		if ferr != nil {
			return "", nil, ferr
		}
		if lockedUntil != nil {
			return "", nil, response.ErrAccountLocked(*lockedUntil)
		}
		return "", nil, response.ErrUsernamePasswordError()
	}

	now := time.Now()
	if err := s.users.MarkLoginSuccess(ctx, u.ID, now); err != nil {
		return "", nil, err
	}
	u, err = s.users.GetUser(ctx, u.ID)
	if err != nil {
		return "", nil, err
	}

	userVo := identity.ToUserVo(u)
	ttl := s.sessionTTL(u)
	token := uuid.NewString()
	sess := &Session{
		Token:      token,
		User:       userVo,
		LoginTime:  now.UnixMilli(),
		ExpireTime: expireTimeMs(now, ttl),
		TTLSeconds: int64(ttl / time.Second),
		IP:         ip,
		UserAgent:  userAgent,
	}
	if err := s.sessions.Save(ctx, sess, ttl); err != nil {
		return "", nil, response.ErrLoginFailed(err)
	}
	jwtToken, err := s.jwt.Sign(token, u.ID)
	if err != nil {
		return "", nil, response.ErrLoginFailed(err)
	}
	return jwtToken, userVo, nil
}

// logout deletes the session referenced by the bearer token (best effort).
func (s *Service) logout(ctx context.Context, rawHeader string) {
	sessionID, _, err := s.jwt.Parse(extractToken(rawHeader))
	if err != nil {
		return
	}
	_ = s.sessions.Delete(ctx, sessionID)
}

// index returns the session user; with refresh it reloads from DB and writes
// the fresh snapshot back to Redis.
func (s *Service) index(ctx context.Context, sess *Session, refresh bool) (*identityv1.UserVo, error) {
	if !refresh {
		return sess.User, nil
	}
	u, err := s.users.GetUser(ctx, sess.User.ID)
	if err != nil {
		return nil, err
	}
	vo := identity.ToUserVo(u)
	ttl := s.sessionTTL(u)
	sess.User = vo
	sess.TTLSeconds = int64(ttl / time.Second)
	sess.ExpireTime = expireTimeMs(time.Now(), ttl)
	if err := s.sessions.Save(ctx, sess, ttl); err != nil {
		return nil, response.ErrInternal(err)
	}
	return vo, nil
}

func (s *Service) publicKey() (string, error) {
	if s.rsa == nil {
		return "", response.ErrInternal(fmt.Errorf("rsa is disabled"))
	}
	return s.rsa.PublicKeyBase64(), nil
}

// generateCaptcha generates a new captcha image; the answer is stored in Redis.
func (s *Service) generateCaptcha() (string, string, error) {
	id, b64s, _, err := s.captcha.Generate()
	if err != nil {
		return "", "", response.ErrInternal(err)
	}
	return id, b64s, nil
}

// verifyCaptcha validates the captcha answer and deletes it on success.
// NOTE: not enabled in login yet (see the commented call in login).
func (s *Service) verifyCaptcha(key, answer string) error {
	if key == "" {
		return response.ErrCaptchaEmpty()
	}
	stored := s.captcha.Store.Get(key, false)
	if strings.TrimSpace(stored) == "" {
		return response.ErrCaptchaExpired()
	}
	if !strings.EqualFold(strings.TrimSpace(stored), strings.TrimSpace(answer)) {
		return response.ErrCaptchaError()
	}
	_ = s.captcha.Store.Get(key, true)
	return nil
}

// sessionTTL resolves the session TTL: user expireTimes overrides the default;
// negative means never expire.
func (s *Service) sessionTTL(u *identity.User) time.Duration {
	minutes := s.cfg.ExpireMinutes
	if u.ExpireTimes != 0 {
		minutes = u.ExpireTimes
	}
	if minutes < 0 {
		return 0
	}
	if minutes == 0 {
		minutes = s.cfg.ExpireMinutes
	}
	return time.Duration(minutes) * time.Minute
}

func expireTimeMs(now time.Time, ttl time.Duration) int64 {
	if ttl <= 0 {
		return 0
	}
	return now.Add(ttl).UnixMilli()
}
