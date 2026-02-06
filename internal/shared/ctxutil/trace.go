package ctxutil

import (
	"context"
)

const TraceIDKey = "request_id"

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	value := ctx.Value(TraceIDKey)
	if value == nil {
		return ""
	}

	if requestID, ok := value.(string); ok {
		return requestID
	}

	return ""
}

func SetTraceID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, requestID)
}
