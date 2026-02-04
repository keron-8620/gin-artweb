package crypto

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher bcrypt哈希实现
type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) Hasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(ctx context.Context, data string) (string, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return "", errors.Wrap(ctx.Err(), "上下文已取消")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(data), h.cost)
	if err != nil {
		return "", errors.Wrap(err, "Bcrypt生成哈希错误")
	}
	return string(hashed), nil
}

func (h *BcryptHasher) Verify(ctx context.Context, data, hash string) (bool, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return false, errors.Wrap(ctx.Err(), "上下文已取消")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(data))
	if err != nil {
		return false, nil
	}
	return true, nil
}
