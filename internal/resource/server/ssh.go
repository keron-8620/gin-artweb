package server

import (
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

func NewSigner(logger *zap.Logger, keyPath string, timeout time.Duration) (ssh.Signer, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		logger.Error(
			"读取私钥文件失败",
			zap.Error(err),
			zap.String("private_key", keyPath),
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
