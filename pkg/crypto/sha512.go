package crypto

import (
	"crypto/sha512"
	"encoding/hex"
)

// SHA512Hasher SHA-512哈希实现
type SHA512Hasher struct{}

func NewSHA512Hasher() Hasher {
	return &SHA512Hasher{}
}

func (h *SHA512Hasher) Hash(data string) (string, error) {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

func (h *SHA512Hasher) Verify(data, hash string) (bool, error) {
	computedHash, err := h.Hash(data)
	if err != nil {
		return false, err
	}
	return computedHash == hash, nil
}

