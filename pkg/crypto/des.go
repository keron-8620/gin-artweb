package crypto

import (
	"context"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"

	"github.com/pkg/errors"
)

type desCipher struct {
	block cipher.Block
	key   []byte
	iv    []byte
	BaseCipher
}

// NewDESCipher 创建DES加密器实例
func NewDESCipher(key []byte, iv ...[]byte) (Cipher, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "DES创建加密器错误")
	}

	// 设置IV，默认生成随机IV
	actualIV := make([]byte, des.BlockSize)
	if len(iv) > 0 && len(iv[0]) == des.BlockSize {
		copy(actualIV, iv[0])
	} else {
		// 生成随机IV
		if _, err := rand.Read(actualIV); err != nil {
			return nil, errors.Wrap(err, "生成随机IV错误")
		}
	}

	return &desCipher{
		block: block,
		key:   key,
		iv:    actualIV,
	}, nil
}

// Encrypt 加密数据，接收字符串，返回加密后的字符串
func (d *desCipher) Encrypt(ctx context.Context, plaintext string) (string, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return "", errors.Wrap(ctx.Err(), "上下文已取消")
	}

	plainBytes := []byte(plaintext)

	// 使用PKCS7填充
	blockSize := d.block.BlockSize()
	plainBytes = pkcs7Padding(plainBytes, blockSize)

	// CBC模式加密
	ciphertext := make([]byte, len(plainBytes))

	// 创建加密器
	mode := cipher.NewCBCEncrypter(d.block, d.iv)
	mode.CryptBlocks(ciphertext, plainBytes)

	// 将加密结果编码为base64字符串
	return d.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据，接收加密字符串，返回解密后的字符串
func (d *desCipher) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return "", errors.Wrap(ctx.Err(), "上下文已取消")
	}

	// 解码base64字符串
	cipherBytes, err := d.DecodeString(ciphertext)
	if err != nil {
		return "", errors.Wrap(err, "DES解密解码错误")
	}

	// CBC模式解密
	blockSize := d.block.BlockSize()
	if len(cipherBytes)%blockSize != 0 {
		return "", errors.Errorf("密文不是块大小的倍数: %d", len(cipherBytes))
	}

	plainBytes := make([]byte, len(cipherBytes))

	mode := cipher.NewCBCDecrypter(d.block, d.iv)
	mode.CryptBlocks(plainBytes, cipherBytes)

	// 去除PKCS7填充
	plainBytes, err = pkcs7Unpadding(plainBytes)
	if err != nil {
		return "", errors.Wrap(err, "DES解密去填充错误")
	}

	return string(plainBytes), nil
}
