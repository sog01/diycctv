package context

import (
	"context"

	"github.com/google/uuid"
)

// RequestIDKey is the context key type used for request IDs
type RequestIDKey string

// RequestIDCtxKey is the specific context key used to store the request ID
const RequestIDCtxKey RequestIDKey = "request_id"

// RequestIDHeader is the HTTP header used for request IDs
const RequestIDHeader = "X-Request-ID"

// WithRequestID returns a new context with the given request ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDCtxKey, requestID)
}

// GetRequestID extracts the request ID from the context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if id, ok := ctx.Value(RequestIDCtxKey).(string); ok {
		return id
	}
	return ""
}

// GenerateRequestID creates a new UUID for use as a request ID
func GenerateRequestID() string {
	return uuid.New().String()
}
