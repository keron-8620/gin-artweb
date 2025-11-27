package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

// ScryptHasher scrypt哈希实现
type ScryptHasher struct {
	saltLen   int
	n         int
	r         int
	p         int
	keyLen    int
	encodeFmt string // "hex" or "base64"
}

func NewScryptHasher() Hasher {
	return &ScryptHasher{
		saltLen:   16,
		n:         32768, // 2^15
		r:         8,
		p:         1,
		keyLen:    32,
		encodeFmt: "base64",
	}
}

func NewScryptHasherWithParams(saltLen, n, r, p, keyLen int, encodeFmt string) Hasher {
	return &ScryptHasher{
		saltLen:   saltLen,
		n:         n,
		r:         r,
		p:         p,
		keyLen:    keyLen,
		encodeFmt: encodeFmt,
	}
}

func (h *ScryptHasher) Hash(data string) (string, error) {
	// 生成随机盐值
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 生成哈希
	hash, err := scrypt.Key([]byte(data), salt, h.n, h.r, h.p, h.keyLen)
	if err != nil {
		return "", err
	}

	// 将盐值和哈希值组合
	result := append(salt, hash...)

	// 根据指定格式编码
	if h.encodeFmt == "hex" {
		return hex.EncodeToString(result), nil
	}
	return base64.StdEncoding.EncodeToString(result), nil
}

func (h *ScryptHasher) Verify(data, hash string) (bool, error) {
	// 解码哈希值
	var hashBytes []byte
	var err error

	if h.encodeFmt == "hex" {
		hashBytes, err = hex.DecodeString(hash)
	} else {
		hashBytes, err = base64.StdEncoding.DecodeString(hash)
	}

	if err != nil {
		return false, err
	}

	// 提取盐值和哈希部分
	if len(hashBytes) <= h.saltLen {
		return false, fmt.Errorf("invalid hash format")
	}

	salt := hashBytes[:h.saltLen]
	expectedHash := hashBytes[h.saltLen:]

	// 使用相同参数重新计算哈希
	computedHash, err := scrypt.Key([]byte(data), salt, h.n, h.r, h.p, h.keyLen)
	if err != nil {
		return false, err
	}

	// 比较哈希值
	return string(expectedHash) == string(computedHash), nil
}
