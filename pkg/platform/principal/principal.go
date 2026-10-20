// Package principal carries the authenticated user through the request context
// so service layers can access the full user (roles/orgs) without depending on
// the transport layer.
package principal

import (
	"context"

	identityv1 "harns-platform/pkg/platform/identity/v1"
)

type contextKey struct{}

// With returns a context carrying the given user.
func With(ctx context.Context, user *identityv1.UserVo) context.Context {
	if user == nil {
		return ctx
	}
	return context.WithValue(ctx, contextKey{}, user)
}

// FromContext returns the authenticated user stored in ctx, if any.
func FromContext(ctx context.Context) (*identityv1.UserVo, bool) {
	user, ok := ctx.Value(contextKey{}).(*identityv1.UserVo)
	return user, ok
}

// UserID returns the authenticated user id, or an empty string.
func UserID(ctx context.Context) string {
	user, ok := FromContext(ctx)
	if !ok || user == nil {
		return ""
	}
	return user.ID
}

// UserName returns the authenticated user name, or an empty string.
func UserName(ctx context.Context) string {
	user, ok := FromContext(ctx)
	if !ok || user == nil {
		return ""
	}
	return user.UserName
}
