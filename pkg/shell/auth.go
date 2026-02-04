package shell

import (
	"os"
	"path/filepath"

	"emperror.dev/errors"
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
		return "", errors.WithMessage(err, "获取用户主目录失败")
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
	if filePath == "" {
		return errors.New("SSH私钥文件路径不能为空")
	}

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.WithMessagef(err, "SSH私钥文件不存在，路径: %s", filePath)
		}
		return errors.WithMessagef(err, "检查SSH私钥文件状态失败，路径: %s", filePath)
	}

	// 检查是否为普通文件
	if !info.Mode().IsRegular() {
		return errors.Errorf("SSH私钥路径不是一个普通文件，路径: %s", filePath)
	}

	// 检查文件权限
	if info.Mode().Perm()&0077 != 0 {
		return errors.Errorf("SSH私钥文件权限过于宽松，路径: %s", filePath)
	}

	return nil
}

// ParsePrivateKey 解析SSH私钥文件
func ParsePrivateKey(filePath string) (ssh.Signer, error) {
	if filePath == "" {
		return nil, errors.New("SSH私钥文件路径不能为空")
	}

	// 验证文件
	if err := validateKeyFile(filePath); err != nil {
		return nil, err
	}

	// 读取私钥文件
	key, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.WithMessagef(err, "读取SSH私钥文件失败，路径: %s", filePath)
	}

	// 检查文件内容是否为空
	if len(key) == 0 {
		return nil, errors.Errorf("SSH私钥文件内容为空，路径: %s", filePath)
	}

	// 解析私钥
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, errors.WithMessagef(err, "解析SSH私钥失败，路径: %s", filePath)
	}

	return signer, nil
}

// GetSignersFromDefaultKeys 获取所有默认密钥的signer
func GetSignersFromDefaultKeys() ([]ssh.Signer, error) {
	keyPaths := FindAllValidKeys()
	if len(keyPaths) == 0 {
		return nil, errors.New("未找到任何有效的SSH私钥文件")
	}

	var signers []ssh.Signer
	for _, keyPath := range keyPaths {
		signer, err := ParsePrivateKey(keyPath)
		if err != nil {
			return nil, err
		}
		signers = append(signers, signer)
	}

	if len(signers) == 0 {
		return nil, errors.New("未能从任何默认密钥文件创建SSH Signer")
	}

	return signers, nil
}

// GetPublicKeysFromSigners 从signers中获取公钥
func GetPublicKeysFromSigners(signers []ssh.Signer) []ssh.PublicKey {
	var publicKeys []ssh.PublicKey
	for _, signer := range signers {
		if signer != nil {
			publicKeys = append(publicKeys, signer.PublicKey())
		}
	}
	return publicKeys
}

// GetPublicKeyFromSigner 从单个signer中获取公钥
func GetPublicKeyFromSigner(signer ssh.Signer) ssh.PublicKey {
	if signer == nil {
		return nil
	}
	return signer.PublicKey()
}

// GetPublicKeyBytesFromSigner 从signer中获取公钥字节
func GetPublicKeyBytesFromSigner(signer ssh.Signer) []byte {
	if signer == nil {
		return nil
	}
	return ssh.MarshalAuthorizedKey(signer.PublicKey())
}

// GetPublicKeyStringFromSigner 从signer中获取公钥字符串
func GetPublicKeyStringFromSigner(signer ssh.Signer) string {
	if signer == nil {
		return ""
	}
	return string(ssh.MarshalAuthorizedKey(signer.PublicKey()))
}
