package crypto

import (
	"context"
	"io"
	"os"

	"emperror.dev/errors"
)

// FileEncryptor 文件加密器接口
type FileEncryptor interface {
	// EncryptFile 加密文件
	EncryptFile(ctx context.Context, srcPath, dstPath string) error
	// DecryptFile 解密文件
	DecryptFile(ctx context.Context, srcPath, dstPath string) error
}

// AESFileEncryptor AES文件加密器
type AESFileEncryptor struct {
	cipher Cipher
}

// NewAESFileEncryptor 创建AES文件加密器
func NewAESFileEncryptor(cipher Cipher) *AESFileEncryptor {
	return &AESFileEncryptor{
		cipher: cipher,
	}
}

// EncryptFile 加密文件
func (fe *AESFileEncryptor) EncryptFile(ctx context.Context, srcPath, dstPath string) error {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "上下文已取消")
	}

	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return errors.Wrap(err, "打开源文件错误")
	}
	defer srcFile.Close()

	// 读取文件内容
	content, err := io.ReadAll(srcFile)
	if err != nil {
		return errors.Wrap(err, "读取文件内容错误")
	}

	// 加密内容
	encryptedContent, err := fe.cipher.Encrypt(ctx, string(content))
	if err != nil {
		return errors.Wrap(err, "加密文件内容错误")
	}

	// 写入目标文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return errors.Wrap(err, "创建目标文件错误")
	}
	defer dstFile.Close()

	_, err = dstFile.WriteString(encryptedContent)
	if err != nil {
		return errors.Wrap(err, "写入加密内容错误")
	}

	return nil
}

// DecryptFile 解密文件
func (fe *AESFileEncryptor) DecryptFile(ctx context.Context, srcPath, dstPath string) error {
	// 检查context是否已取消
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), "上下文已取消")
	}

	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return errors.Wrap(err, "打开源文件错误")
	}
	defer srcFile.Close()

	// 读取文件内容
	content, err := io.ReadAll(srcFile)
	if err != nil {
		return errors.Wrap(err, "读取文件内容错误")
	}

	// 解密内容
	decryptedContent, err := fe.cipher.Decrypt(ctx, string(content))
	if err != nil {
		return errors.Wrap(err, "解密文件内容错误")
	}

	// 写入目标文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return errors.Wrap(err, "创建目标文件错误")
	}
	defer dstFile.Close()

	_, err = dstFile.WriteString(decryptedContent)
	if err != nil {
		return errors.Wrap(err, "写入解密内容错误")
	}

	return nil
}

// EncryptFile 加密文件的便捷函数
func EncryptFile(ctx context.Context, cipher Cipher, srcPath, dstPath string) error {
	encryptor := NewAESFileEncryptor(cipher)
	return encryptor.EncryptFile(ctx, srcPath, dstPath)
}

// DecryptFile 解密文件的便捷函数
func DecryptFile(ctx context.Context, cipher Cipher, srcPath, dstPath string) error {
	encryptor := NewAESFileEncryptor(cipher)
	return encryptor.DecryptFile(ctx, srcPath, dstPath)
}
