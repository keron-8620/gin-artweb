package ctxutil

import (
	"context"
	"strings"
)

const (
	TraceIDKey     ContextKey = "trace_id"
	DefaultTraceID string     = "unknown-trace-id"
)

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return DefaultTraceID
	}

	traceID, ok := ctx.Value(TraceIDKey).(string)
	if ok && strings.TrimSpace(traceID) != "" {
		return traceID
	}

	return DefaultTraceID
}

func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}
