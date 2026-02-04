package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"math/big"

	"github.com/pkg/errors"
)

// GenerateRandomBytes 生成指定长度的随机字节
func GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return nil, errors.Wrap(err, "生成随机字节错误")
	}
	return b, nil
}

// GenerateRandomString 生成指定长度的随机字符串（base64编码）
func GenerateRandomString(length int) (string, error) {
	b, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateRandomHex 生成指定长度的随机十六进制字符串
func GenerateRandomHex(length int) (string, error) {
	b, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateRandomInt 生成指定范围内的随机整数 [min, max]
func GenerateRandomInt(min, max int64) (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return 0, errors.Wrap(err, "生成随机整数错误")
	}
	return n.Int64() + min, nil
}

// EncodeBase64 将字节数据编码为Base64字符串
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 将Base64字符串解码为字节数据
func DecodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// EncodeHex 将字节数据编码为十六进制字符串
func EncodeHex(data []byte) string {
	return hex.EncodeToString(data)
}

// DecodeHex 将十六进制字符串解码为字节数据
func DecodeHex(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

// IsValidBase64 检查字符串是否为有效的Base64编码
func IsValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// IsValidHex 检查字符串是否为有效的十六进制编码
func IsValidHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// PadRight 在字符串右侧填充指定字符到指定长度
func PadRight(s string, pad byte, length int) string {
	if len(s) >= length {
		return s
	}
	padding := make([]byte, length-len(s))
	for i := range padding {
		padding[i] = pad
	}
	return s + string(padding)
}

// PadLeft 在字符串左侧填充指定字符到指定长度
func PadLeft(s string, pad byte, length int) string {
	if len(s) >= length {
		return s
	}
	padding := make([]byte, length-len(s))
	for i := range padding {
		padding[i] = pad
	}
	return string(padding) + s
}

// TruncateString 截断字符串到指定长度
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}

// SafeEqual 安全比较两个字符串（防止时间攻击）
func SafeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i] ^ b[i])
	}
	return result == 0
}
