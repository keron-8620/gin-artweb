package shell

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// DefaultKeyPaths 是默认的SSH私钥文件路径列表
var DefaultKeyPaths = []string{
	"~/.ssh/id_rsa",
	"~/.ssh/id_ecdsa",
	"~/.ssh/id_ed25519",
	"~/.ssh/id_dsa",
}

// ExpandHomeDir 将路径中的 ~ 替换为用户主目录
func ExpandHomeDir(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}

	return filepath.Join(homeDir, path[1:]), nil
}

// FindAllValidKeys 尝试查找所有存在的有效SSH私钥文件
func FindAllValidKeys() []string {
	var validKeys []string
	for _, keyPath := range DefaultKeyPaths {
		expandedPath, err := ExpandHomeDir(keyPath)
		if err != nil {
			continue
		}

		// 检查文件是否存在且有效
		if err := validateKeyFile(expandedPath); err == nil {
			validKeys = append(validKeys, expandedPath)
		}
	}

	return validKeys
}

// validateKeyFile 验证SSH私钥文件的有效性
func validateKeyFile(filePath string) error {
	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("SSH私钥文件不存在: %s", filePath)
		}
		return fmt.Errorf("检查SSH私钥文件失败: %w", err)
	}

	// 检查是否为普通文件
	if !info.Mode().IsRegular() {
		return fmt.Errorf("SSH私钥路径不是一个普通文件: %s", filePath)
	}

	// 检查文件权限
	if info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("SSH私钥文件权限过于宽松 (%s)", filePath)
	}

	return nil
}

// ParsePrivateKey 解析SSH私钥文件
func ParsePrivateKey(filePath string) (ssh.Signer, error) {
	// 验证文件
	if err := validateKeyFile(filePath); err != nil {
		return nil, err
	}

	// 读取私钥文件
	key, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取SSH私钥文件失败: %w", err)
	}

	// 解析私钥
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("解析SSH私钥失败: %w", err)
	}

	return signer, nil
}

// GetSignersFromDefaultKeys 获取所有默认密钥的signer
func GetSignersFromDefaultKeys() ([]ssh.Signer, error) {
	keyPaths := FindAllValidKeys()
	var signers []ssh.Signer

	for _, keyPath := range keyPaths {
		signer, err := ParsePrivateKey(keyPath)
		if err != nil {
			return nil, err
		}
		signers = append(signers, signer)
	}

	if len(signers) == 0 {
		return nil, fmt.Errorf("未能从任何默认密钥文件创建SSH Signer")
	}

	return signers, nil
}

// GetPublicKeysFromSigners 从signers中获取公钥
func GetPublicKeysFromSigners(signers []ssh.Signer) []ssh.PublicKey {
	var publicKeys []ssh.PublicKey
	for _, signer := range signers {
		publicKeys = append(publicKeys, signer.PublicKey())
	}
	return publicKeys
}

// GetPublicKeyFromSigner 从单个signer中获取公钥
func GetPublicKeyFromSigner(signer ssh.Signer) ssh.PublicKey {
	return signer.PublicKey()
}

// GetPublicKeyBytesFromSigner 从signer中获取公钥字节
func GetPublicKeyBytesFromSigner(signer ssh.Signer) []byte {
	return ssh.MarshalAuthorizedKey(signer.PublicKey())
}

// GetPublicKeyStringFromSigner 从signer中获取公钥字符串
func GetPublicKeyStringFromSigner(signer ssh.Signer) string {
	return string(ssh.MarshalAuthorizedKey(signer.PublicKey()))
}
