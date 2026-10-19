// Package principal carries the authenticated user through the request context
// so service layers can record creator/updater without depending on the
// transport layer.
package principal

import "context"

// Principal is the authenticated user of the current request.
type Principal struct {
	UserID   string
	UserName string
}

type contextKey struct{}

// With returns a context carrying the given principal.
func With(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

// FromContext returns the principal stored in ctx, if any.
func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(contextKey{}).(Principal)
	return p, ok
}

// UserID returns the authenticated user id, or an empty string.
func UserID(ctx context.Context) string {
	p, _ := FromContext(ctx)
	return p.UserID
}
