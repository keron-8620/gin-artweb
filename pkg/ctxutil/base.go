package ctxutil

import (
	"context"

	"github.com/pkg/errors"
)

func CheckContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("ctx不能为nil")
	}
	select {
	case <-ctx.Done():
		return errors.WithMessage(ctx.Err(), "操作已取消/超时")
	default:
		return nil
	}
}
