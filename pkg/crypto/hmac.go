package crypto

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"emperror.dev/errors"
)

// HMACType HMAC算法类型
type HMACType string

const (
	// HMACSHA256 HMAC-SHA256算法
	HMACSHA256 HMACType = "hmac-sha256"
	// HMACSHA512 HMAC-SHA512算法
	HMACSHA512 HMACType = "hmac-sha512"
)

// HMACHasher HMAC哈希实现
type HMACHasher struct {
	key      []byte
	hmacType HMACType
}

// NewHMACHasher 创建HMAC哈希器实例
func NewHMACHasher(key []byte, hmacType HMACType) *HMACHasher {
	if hmacType == "" {
		hmacType = HMACSHA256 // 默认使用SHA256
	}

	return &HMACHasher{
		key:      key,
		hmacType: hmacType,
	}
}

// Hash 对数据进行HMAC哈希处理
func (h *HMACHasher) Hash(ctx context.Context, data string) (string, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return "", errors.Wrap(ctx.Err(), "上下文已取消")
	}

	var mac []byte
	switch h.hmacType {
	case HMACSHA256:
		h := hmac.New(sha256.New, h.key)
		h.Write([]byte(data))
		mac = h.Sum(nil)
	case HMACSHA512:
		h := hmac.New(sha512.New, h.key)
		h.Write([]byte(data))
		mac = h.Sum(nil)
	default:
		return "", errors.Errorf("不支持的HMAC算法: %s", h.hmacType)
	}

	return hex.EncodeToString(mac), nil
}

// Verify 验证数据与HMAC值是否匹配
func (h *HMACHasher) Verify(ctx context.Context, data, hash string) (bool, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return false, errors.Wrap(ctx.Err(), "上下文已取消")
	}

	computedHash, err := h.Hash(ctx, data)
	if err != nil {
		return false, errors.Wrap(err, "验证HMAC错误")
	}

	return computedHash == hash, nil
}

// GenerateHMAC 生成HMAC值的便捷函数
func GenerateHMAC(data string, key []byte, hmacType HMACType) (string, error) {
	hasher := NewHMACHasher(key, hmacType)
	return hasher.Hash(nil, data)
}

// VerifyHMAC 验证HMAC值的便捷函数
func VerifyHMAC(data, hash string, key []byte, hmacType HMACType) (bool, error) {
	hasher := NewHMACHasher(key, hmacType)
	return hasher.Verify(nil, data, hash)
}
