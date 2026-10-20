package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	identityv1 "harns-platform/pkg/platform/identity/v1"
)

const (
	sessionKeyPrefix      = "platform:login_tokens:"
	userSessionsKeyPrefix = "platform:user_sessions:"
)

// Session is the server-side login state stored in Redis.
type Session struct {
	Token      string             `json:"token"`
	User       *identityv1.UserVo `json:"user"`
	LoginTime  int64              `json:"loginTime"`
	ExpireTime int64              `json:"expireTime"`
	TTLSeconds int64              `json:"ttlSeconds"`
	IP         string             `json:"ip"`
	UserAgent  string             `json:"userAgent"`
}

type sessionStore struct {
	rdb     redis.UniversalClient
	refresh time.Duration
}

func newSessionStore(rdb redis.UniversalClient, refreshMinutes int) *sessionStore {
	return &sessionStore{
		rdb:     rdb,
		refresh: time.Duration(refreshMinutes) * time.Minute,
	}
}

func sessionKey(token string) string {
	return sessionKeyPrefix + token
}

func userSessionsKey(userID string) string {
	return userSessionsKeyPrefix + userID
}

func (s *sessionStore) Save(ctx context.Context, sess *Session, ttl time.Duration) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	if ttl < 0 {
		ttl = 0
	}
	if err := s.rdb.Set(ctx, sessionKey(sess.Token), data, ttl).Err(); err != nil {
		return err
	}
	if sess.User != nil {
		key := userSessionsKey(sess.User.ID)
		if err := s.rdb.SAdd(ctx, key, sess.Token).Err(); err != nil {
			return err
		}
		if ttl > 0 {
			_ = s.rdb.Expire(ctx, key, ttl).Err()
		}
	}
	return nil
}

func (s *sessionStore) Get(ctx context.Context, token string) (*Session, error) {
	data, err := s.rdb.Get(ctx, sessionKey(token)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sess := &Session{}
	if err := json.Unmarshal(data, sess); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *sessionStore) Delete(ctx context.Context, token string) error {
	if sess, _ := s.Get(ctx, token); sess != nil && sess.User != nil {
		_ = s.rdb.SRem(ctx, userSessionsKey(sess.User.ID), token).Err()
	}
	return s.rdb.Del(ctx, sessionKey(token)).Err()
}

// DeleteByUser removes all sessions of a user.
func (s *sessionStore) DeleteByUser(ctx context.Context, userID string) error {
	key := userSessionsKey(userID)
	tokens, err := s.rdb.SMembers(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	if len(tokens) > 0 {
		keys := make([]string, 0, len(tokens))
		for _, token := range tokens {
			keys = append(keys, sessionKey(token))
		}
		if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}
	return s.rdb.Del(ctx, key).Err()
}

// RefreshIfNeeded extends the session TTL when the remaining time is below the
// configured refresh threshold.
func (s *sessionStore) RefreshIfNeeded(ctx context.Context, sess *Session) error {
	if sess.TTLSeconds <= 0 || s.refresh <= 0 {
		return nil
	}
	ttl, err := s.rdb.TTL(ctx, sessionKey(sess.Token)).Result()
	if err != nil {
		return err
	}
	if ttl > 0 && ttl < s.refresh {
		expire := time.Duration(sess.TTLSeconds) * time.Second
		if err := s.rdb.Expire(ctx, sessionKey(sess.Token), expire).Err(); err != nil {
			return err
		}
		if sess.User != nil {
			_ = s.rdb.Expire(ctx, userSessionsKey(sess.User.ID), expire).Err()
		}
	}
	return nil
}
