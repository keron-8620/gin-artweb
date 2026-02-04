package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"github.com/pkg/errors"
)

// AESGCMCipher AES-GCM模式加密器
type AESGCMCipher struct {
	block cipher.Block
	key   []byte
	BaseCipher
}

// NewAESGCMCipher 创建AES-GCM加密器实例
func NewAESGCMCipher(key []byte) (Cipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "AES创建加密器错误")
	}

	return &AESGCMCipher{
		block: block,
		key:   key,
	}, nil
}

// Encrypt 加密数据，使用GCM模式
func (a *AESGCMCipher) Encrypt(ctx context.Context, plaintext string) (string, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return "", errors.Wrap(ctx.Err(), "上下文已取消")
	}

	plainBytes := []byte(plaintext)

	// 创建GCM模式
	gcm, err := cipher.NewGCM(a.block)
	if err != nil {
		return "", errors.Wrap(err, "创建GCM模式错误")
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", errors.Wrap(err, "生成随机nonce错误")
	}

	// 加密并添加认证
	ciphertext := gcm.Seal(nonce, nonce, plainBytes, nil)

	// 将加密结果编码为base64字符串
	return a.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据，使用GCM模式
func (a *AESGCMCipher) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return "", errors.Wrap(ctx.Err(), "上下文已取消")
	}

	// 解码base64字符串
	cipherBytes, err := a.DecodeString(ciphertext)
	if err != nil {
		return "", errors.Wrap(err, "AES-GCM解密解码错误")
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(a.block)
	if err != nil {
		return "", errors.Wrap(err, "创建GCM模式错误")
	}

	// 检查密文长度
	if len(cipherBytes) < gcm.NonceSize() {
		return "", errors.New("密文长度不足")
	}

	// 提取nonce和密文
	nonce, ciphertextBytes := cipherBytes[:gcm.NonceSize()], cipherBytes[gcm.NonceSize():]

	// 解密并验证
	plainBytes, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", errors.Wrap(err, "GCM解密或验证错误")
	}

	return string(plainBytes), nil
}
