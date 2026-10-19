package auth

import (
	"context"
	"image/color"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

const captchaKeyPrefix = "platform:captcha:"

// redisCaptchaStore is a base64Captcha.Store backed by Redis.
type redisCaptchaStore struct {
	rdb redis.UniversalClient
	ttl time.Duration
}

func newRedisCaptchaStore(rdb redis.UniversalClient, ttlSeconds int) *redisCaptchaStore {
	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}
	return &redisCaptchaStore{
		rdb: rdb,
		ttl: time.Duration(ttlSeconds) * time.Second,
	}
}

func captchaKey(id string) string {
	return captchaKeyPrefix + id
}

func (s *redisCaptchaStore) Set(id, value string) error {
	return s.rdb.Set(context.Background(), captchaKey(id), value, s.ttl).Err()
}

func (s *redisCaptchaStore) Get(id string, clear bool) string {
	ctx := context.Background()
	value, err := s.rdb.Get(ctx, captchaKey(id)).Result()
	if err != nil {
		return ""
	}
	if clear {
		_ = s.rdb.Del(ctx, captchaKey(id)).Err()
	}
	return value
}

func (s *redisCaptchaStore) Verify(id, answer string, clear bool) bool {
	stored := strings.TrimSpace(s.Get(id, clear))
	return stored != "" && strings.EqualFold(stored, strings.TrimSpace(answer))
}

func newCaptchaManager(rdb redis.UniversalClient, ttlSeconds int) *base64Captcha.Captcha {
	driver := base64Captcha.NewDriverMath(
		50, 120, 0, base64Captcha.OptionShowSlimeLine,
		&color.RGBA{R: 255, G: 255, B: 255, A: 255}, nil, nil,
	)
	return base64Captcha.NewCaptcha(driver, newRedisCaptchaStore(rdb, ttlSeconds))
}
