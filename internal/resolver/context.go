// Package resolver provides context helpers for resolver request metadata.
package resolver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type contextKey string

const (
	requestIDKey contextKey = "resolver.request_id"
	channelKey   contextKey = "resolver.channel"
)

// WithRequestID stores resolver request ID into context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext retrieves resolver request ID from context.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(requestIDKey).(string)
	if !ok || v == "" {
		return "", false
	}
	return v, true
}

// EnsureRequestID ensures resolver request ID exists in context.
func EnsureRequestID(ctx context.Context) (context.Context, string) {
	if requestID, ok := RequestIDFromContext(ctx); ok {
		return ctx, requestID
	}
	requestID := newRequestID()
	return WithRequestID(ctx, requestID), requestID
}

// WithChannel stores input channel into context.
func WithChannel(ctx context.Context, channel string) context.Context {
	return context.WithValue(ctx, channelKey, channel)
}

// ChannelFromContext retrieves input channel from context.
func ChannelFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(channelKey).(string)
	if !ok || v == "" {
		return "", false
	}
	return v, true
}

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "resolver-request-id-fallback"
	}
	return hex.EncodeToString(b)
}
