package ctxutil

import (
	"context"
)

type ContextKey string

func CheckContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
