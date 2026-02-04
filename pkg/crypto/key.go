package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"emperror.dev/errors"
	"golang.org/x/crypto/scrypt"
)

// KeyManager 密钥管理器
type KeyManager struct {
	saltLen int
	n       int
	r       int
	p       int
	keyLen  int
}

// NewKeyManager 创建密钥管理器实例
func NewKeyManager() *KeyManager {
	return &KeyManager{
		saltLen: 16,
		n:       32768, // 2^15
		r:       8,
		p:       1,
		keyLen:  32, // 默认生成32字节密钥（AES-256）
	}
}

// NewKeyManagerWithParams 使用自定义参数创建密钥管理器
func NewKeyManagerWithParams(saltLen, n, r, p, keyLen int) *KeyManager {
	return &KeyManager{
		saltLen: saltLen,
		n:       n,
		r:       r,
		p:       p,
		keyLen:  keyLen,
	}
}

// DeriveKey 从密码派生密钥
func (km *KeyManager) DeriveKey(password string, salt []byte) ([]byte, error) {
	if len(salt) == 0 {
		salt = make([]byte, km.saltLen)
		if _, err := rand.Read(salt); err != nil {
			return nil, errors.Wrap(err, "生成盐值错误")
		}
	}

	key, err := scrypt.Key([]byte(password), salt, km.n, km.r, km.p, km.keyLen)
	if err != nil {
		return nil, errors.Wrap(err, "密钥派生错误")
	}

	return key, nil
}

// GenerateRandomKey 生成随机密钥
func (km *KeyManager) GenerateRandomKey() ([]byte, error) {
	key := make([]byte, km.keyLen)
	if _, err := rand.Read(key); err != nil {
		return nil, errors.Wrap(err, "生成随机密钥错误")
	}
	return key, nil
}

// GenerateRandomKeyWithSize 生成指定大小的随机密钥
func GenerateRandomKeyWithSize(size int) ([]byte, error) {
	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, errors.Wrap(err, "生成随机密钥错误")
	}
	return key, nil
}

// KeyToBase64 将密钥转换为Base64字符串
func KeyToBase64(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

// KeyFromBase64 从Base64字符串解析密钥
func KeyFromBase64(keyStr string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(keyStr)
}

// KeyToHex 将密钥转换为十六进制字符串
func KeyToHex(key []byte) string {
	return hex.EncodeToString(key)
}

// KeyFromHex 从十六进制字符串解析密钥
func KeyFromHex(keyStr string) ([]byte, error) {
	return hex.DecodeString(keyStr)
}

// ValidateKeySize 验证密钥大小是否适合指定算法
func ValidateKeySize(key []byte, algorithm string) error {
	size := len(key)
	switch algorithm {
	case "aes":
		if size != 16 && size != 24 && size != 32 {
			return errors.Errorf("AES密钥大小必须是16、24或32字节，当前大小: %d", size)
		}
	case "des":
		if size != 8 {
			return errors.Errorf("DES密钥大小必须是8字节，当前大小: %d", size)
		}
	case "sha256":
		if size != 32 {
			return errors.Errorf("SHA-256密钥大小必须是32字节，当前大小: %d", size)
		}
	case "sha512":
		if size != 64 {
			return errors.Errorf("SHA-512密钥大小必须是64字节，当前大小: %d", size)
		}
	default:
		return errors.Errorf("不支持的算法: %s", algorithm)
	}
	return nil
}
