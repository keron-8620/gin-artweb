// pkg/crypto/crypto.go
package crypto

import (
	"bytes"
	"encoding/base64"
	"fmt"
)

// Hasher 定义哈希接口（用于单向加密）
type Hasher interface {
	// Hash 对数据进行哈希处理
	Hash(data string) (string, error)

	// Verify 验证数据与哈希值是否匹配
	Verify(data, hash string) (bool, error)
}

// Cipher 定义加密解密接口
type Cipher interface {
	// Encrypt 加密数据
	Encrypt(plaintext string) (string, error)

	// Decrypt 解密数据
	Decrypt(ciphertext string) (string, error)
}

// BaseCipher 提供基础的编解码功能
type BaseCipher struct{}

// EncodeToString 将字节数据编码为字符串
func (c *BaseCipher) EncodeToString(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeString 将字符串解码为字节数据
func (c *BaseCipher) DecodeString(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
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
		return nil, fmt.Errorf("invalid padding size")
	}

	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, fmt.Errorf("invalid padding size")
	}

	return data[:(length - unpadding)], nil
}
