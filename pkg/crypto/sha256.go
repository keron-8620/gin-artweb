// pkg/crypto/hash.go
package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Hasher SHA-256哈希实现
type SHA256Hasher struct{}

func NewSHA256Hasher() Hasher {
	return &SHA256Hasher{}
}

func (h *SHA256Hasher) Hash(data string) (string, error) {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

func (h *SHA256Hasher) Verify(data, hash string) (bool, error) {
	computedHash, err := h.Hash(data)
	if err != nil {
		return false, err
	}
	return computedHash == hash, nil
}
