package crypto

import (
	"context"
	"crypto/sha512"
	"encoding/hex"

	"github.com/pkg/errors"
)

// SHA512Hasher SHA-512哈希实现
type SHA512Hasher struct{}

func NewSHA512Hasher() Hasher {
	return &SHA512Hasher{}
}

func (h *SHA512Hasher) Hash(ctx context.Context, data string) (string, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return "", errors.Wrap(ctx.Err(), "上下文已取消")
	}

	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

func (h *SHA512Hasher) Verify(ctx context.Context, data, hash string) (bool, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return false, errors.Wrap(ctx.Err(), "上下文已取消")
	}

	computedHash, err := h.Hash(ctx, data)
	if err != nil {
		return false, errors.Wrap(err, "验证哈希错误")
	}
	return computedHash == hash, nil
}
