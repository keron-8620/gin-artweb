package shell

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// TestExpandHomeDir 测试 ExpandHomeDir 函数
func TestExpandHomeDir(t *testing.T) {
	// 测试非 ~ 开头的路径
	path := "/test/path"
	expanded, err := ExpandHomeDir(path)
	if err != nil {
		t.Errorf("ExpandHomeDir(%s) 失败: %v", path, err)
	}
	if expanded != path {
		t.Errorf("ExpandHomeDir(%s) 期望 %s, 实际 %s", path, path, expanded)
	}

	// 测试空路径
	path = ""
	expanded, err = ExpandHomeDir(path)
	if err != nil {
		t.Errorf("ExpandHomeDir(%s) 失败: %v", path, err)
	}
	if expanded != path {
		t.Errorf("ExpandHomeDir(%s) 期望 %s, 实际 %s", path, path, expanded)
	}

	// 测试 ~ 开头的路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("无法获取用户主目录: %v", err)
	}

	path = "~/test"
	expected := filepath.Join(homeDir, "test")
	expanded, err = ExpandHomeDir(path)
	if err != nil {
		t.Errorf("ExpandHomeDir(%s) 失败: %v", path, err)
	}
	if expanded != expected {
		t.Errorf("ExpandHomeDir(%s) 期望 %s, 实际 %s", path, expected, expanded)
	}
}

// TestValidateKeyFile 测试 validateKeyFile 函数
func TestValidateKeyFile(t *testing.T) {
	// 测试空路径
	err := validateKeyFile("")
	if err == nil {
		t.Error("validateKeyFile(\"\") 期望返回错误, 实际返回 nil")
	}

	// 测试不存在的文件
	nonExistentPath := "/nonexistent/path/to/key"
	err = validateKeyFile(nonExistentPath)
	if err == nil {
		t.Errorf("validateKeyFile(%s) 期望返回错误, 实际返回 nil", nonExistentPath)
	}

	// 测试当前目录（不是文件）
	currentDir, _ := os.Getwd()
	err = validateKeyFile(currentDir)
	if err == nil {
		t.Errorf("validateKeyFile(%s) 期望返回错误, 实际返回 nil", currentDir)
	}
}

// TestGetPublicKeysFromSigners 测试 GetPublicKeysFromSigners 函数
func TestGetPublicKeysFromSigners(t *testing.T) {
	// 测试空切片
	signers := []ssh.Signer{}
	publicKeys := GetPublicKeysFromSigners(signers)
	if len(publicKeys) != 0 {
		t.Errorf("GetPublicKeysFromSigners([]) 期望返回空切片, 实际返回长度为 %d 的切片", len(publicKeys))
	}

	// 测试包含 nil 的切片
	signers = []ssh.Signer{nil}
	publicKeys = GetPublicKeysFromSigners(signers)
	if len(publicKeys) != 0 {
		t.Errorf("GetPublicKeysFromSigners([nil]) 期望返回空切片, 实际返回长度为 %d 的切片", len(publicKeys))
	}
}

// TestGetPublicKeyFromSigner 测试 GetPublicKeyFromSigner 函数
func TestGetPublicKeyFromSigner(t *testing.T) {
	// 测试 nil signer
	var signer ssh.Signer
	publicKey := GetPublicKeyFromSigner(signer)
	if publicKey != nil {
		t.Error("GetPublicKeyFromSigner(nil) 期望返回 nil, 实际返回非 nil")
	}
}

// TestGetPublicKeyBytesFromSigner 测试 GetPublicKeyBytesFromSigner 函数
func TestGetPublicKeyBytesFromSigner(t *testing.T) {
	// 测试 nil signer
	var signer ssh.Signer
	publicKeyBytes := GetPublicKeyBytesFromSigner(signer)
	if publicKeyBytes != nil {
		t.Error("GetPublicKeyBytesFromSigner(nil) 期望返回 nil, 实际返回非 nil")
	}
}

// TestGetPublicKeyStringFromSigner 测试 GetPublicKeyStringFromSigner 函数
func TestGetPublicKeyStringFromSigner(t *testing.T) {
	// 测试 nil signer
	var signer ssh.Signer
	publicKeyString := GetPublicKeyStringFromSigner(signer)
	if publicKeyString != "" {
		t.Errorf("GetPublicKeyStringFromSigner(nil) 期望返回空字符串, 实际返回 %s", publicKeyString)
	}
}

// TestFileExists 测试 FileExists 函数
func TestFileExists(t *testing.T) {
	// 测试 nil client
	var sftpClient *sftp.Client
	if _, err := FileExists(sftpClient, ""); err == nil {
		t.Error("FileExists(nil, \"\") 期望返回错误, 实际返回 nil")
	}

	// 测试空路径
	sftpClient = &sftp.Client{}
	if _, err := FileExists(sftpClient, ""); err == nil {
		t.Error("FileExists(client, \"\") 期望返回错误, 实际返回 nil")
	}
}

// TestNewSSHClient 测试 NewSSHClient 函数
func TestNewSSHClient(t *testing.T) {
	ctx := context.Background()

	// 测试空 IP
	_, err := NewSSHClient(ctx, "", 22, "user", []ssh.AuthMethod{}, false, 5*time.Second)
	if err == nil {
		t.Error("NewSSHClient 期望返回错误 (空 IP), 实际返回 nil")
	}

	// 测试空用户
	_, err = NewSSHClient(ctx, "127.0.0.1", 22, "", []ssh.AuthMethod{}, false, 5*time.Second)
	if err == nil {
		t.Error("NewSSHClient 期望返回错误 (空用户), 实际返回 nil")
	}

	// 测试空认证方法
	_, err = NewSSHClient(ctx, "127.0.0.1", 22, "user", []ssh.AuthMethod{}, false, 5*time.Second)
	if err == nil {
		t.Error("NewSSHClient 期望返回错误 (空认证方法), 实际返回 nil")
	}

	// 测试上下文取消
	ctx, cancel := context.WithCancel(ctx)
	cancel() // 立即取消上下文
	_, err = NewSSHClient(ctx, "127.0.0.1", 22, "user", []ssh.AuthMethod{}, false, 5*time.Second)
	if err == nil {
		t.Error("NewSSHClient 期望返回错误 (上下文已取消), 实际返回 nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("NewSSHClient 期望返回 context.Canceled 错误, 实际返回 %v", err)
	}
}

// TestCopyWithContext 测试 copyWithContext 函数
func TestCopyWithContext(t *testing.T) {
	ctx := context.Background()

	// 测试上下文取消
	ctx, cancel := context.WithCancel(ctx)
	cancel() // 立即取消上下文

	// 创建一个简单的 reader 和 writer
	reader := &mockReader{}
	writer := &mockWriter{}

	_, err := copyWithContext(ctx, reader, writer)
	if err == nil {
		t.Error("copyWithContext 期望返回错误 (上下文已取消), 实际返回 nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("copyWithContext 期望返回 context.Canceled 错误, 实际返回 %v", err)
	}
}

// mockReader 是一个模拟的 io.Reader
type mockReader struct{}

func (m *mockReader) Read(p []byte) (n int, err error) {
	// 模拟读取操作
	return 0, nil
}

// mockWriter 是一个模拟的 io.Writer
type mockWriter struct{}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	// 模拟写入操作
	return len(p), nil
}
