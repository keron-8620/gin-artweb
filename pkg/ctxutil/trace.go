package ctxutil

import (
	"context"
)

const TraceIDKey ContextKey = "request_id"

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	value := ctx.Value(TraceIDKey)
	if value == nil {
		return ""
	}

	if requestID, ok := value.(ContextKey); ok {
		return string(requestID)
	}

	return ""
}
