package correlationID

import (
	"context"
)

type contextKey int

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// requestIdKey is the context key for the correlation id.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const requestIDKey key = 0

// NewContext returns a new Context carrying the correlation ID.
func NewContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// FromContext extracts the correlation ID from ctx, if present.
func FromContext(ctx context.Context) (string, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the correlationID type assertion returns ok=false for nil.
	userIP, ok := ctx.Value(requestIDKey).(string)
	return userIP, ok
}
