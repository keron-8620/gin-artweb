// pkg/crypto/hash.go
package crypto

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"emperror.dev/errors"
)

// SHA256Hasher SHA-256哈希实现
type SHA256Hasher struct{}

func NewSHA256Hasher() Hasher {
	return &SHA256Hasher{}
}

func (h *SHA256Hasher) Hash(ctx context.Context, data string) (string, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return "", errors.Wrap(ctx.Err(), "上下文已取消")
	}

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

func (h *SHA256Hasher) Verify(ctx context.Context, data, hash string) (bool, error) {
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
