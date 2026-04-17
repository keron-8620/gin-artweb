package ctxutil

import (
	"context"
	"strings"
)

const (
	TraceIDKey     ContextKey = "request_id"
	defaultTraceID string     = "unknown-trace-id"
)

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return defaultTraceID
	}
	requestID, ok := ctx.Value(TraceIDKey).(string)
	if !ok || strings.TrimSpace(requestID) == "" {
		return defaultTraceID
	}
	return requestID
}

func SetTraceID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, requestID)
}
