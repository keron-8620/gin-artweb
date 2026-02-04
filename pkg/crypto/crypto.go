// pkg/crypto/crypto.go
package crypto

import (
	"bytes"
	"context"
	"encoding/base64"

	"emperror.dev/errors"
)

// Hasher 定义哈希接口（用于单向加密）
type Hasher interface {
	// Hash 对数据进行哈希处理
	Hash(ctx context.Context, data string) (string, error)

	// Verify 验证数据与哈希值是否匹配
	Verify(ctx context.Context, data, hash string) (bool, error)
}

// Cipher 定义加密解密接口
type Cipher interface {
	// Encrypt 加密数据
	Encrypt(ctx context.Context, plaintext string) (string, error)

	// Decrypt 解密数据
	Decrypt(ctx context.Context, ciphertext string) (string, error)
}

// BaseCipher 提供基础的编解码功能
type BaseCipher struct{}

// EncodeToString 将字节数据编码为字符串
func (c *BaseCipher) EncodeToString(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeString 将字符串解码为字节数据
func (c *BaseCipher) DecodeString(data string) ([]byte, error) {
	dataBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, errors.Wrap(err, "Base64解码错误")
	}
	return dataBytes, nil
}

// pkcs7Padding PKCS7填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pkcs7Unpadding PKCS7去除填充
func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("无效的填充大小：数据为空")
	}

	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, errors.Errorf("无效的填充大小：%d 大于数据长度 %d", unpadding, length)
	}

	return data[:(length - unpadding)], nil
}
