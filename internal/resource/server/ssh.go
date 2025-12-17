package server

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

func resolveSSHKeyPath(keyPath string) (string, error) {
	// 如果是绝对路径，直接返回
	if filepath.IsAbs(keyPath) {
		return keyPath, nil
	}

	// 如果是相对路径或者是文件名
	if strings.Contains(keyPath, "/") || strings.Contains(keyPath, "\\") {
		// 相对路径，基于当前工作目录
		absPath, err := filepath.Abs(keyPath)
		if err != nil {
			return "", err
		}
		return absPath, nil
	} else {
		// 简单的文件名，放在 ~/.ssh/ 下
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".ssh", keyPath), nil
	}
}

func NewSigner(logger *zap.Logger, keyPath string, timeout time.Duration) (ssh.Signer, error) {
	// 如果没有提供 keyPath，使用默认值
	if keyPath == "" {
		keyPath = "id_rsa"
	}

	// 解析实际路径
	actualKeyPath, err := resolveSSHKeyPath(keyPath)
	if err != nil {
		logger.Error("解析 SSH 密钥路径失败", zap.Error(err), zap.String("input_path", keyPath))
		return nil, err
	}

	key, err := os.ReadFile(actualKeyPath)
	if err != nil {
		logger.Error(
			"读取私钥文件失败",
			zap.Error(err),
			zap.String("private_key", actualKeyPath),
		)
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		logger.Error(
			"解析私钥失败",
			zap.Error(err),
		)
		return nil, err
	}
	return signer, nil
}
