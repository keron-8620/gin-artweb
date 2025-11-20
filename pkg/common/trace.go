package common

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
