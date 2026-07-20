package shell

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"emperror.dev/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// NewSSHClient 创建SSH客户端连接
func NewSSHClient(
	ctx context.Context,
	sshIP string,
	sshPort uint16,
	sshUser string,
	sshAuths []ssh.AuthMethod,
	useKnowHosts bool,
	timeout time.Duration,
) (*ssh.Client, error) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return nil, errors.WithMessage(ctx.Err(), "上下文已取消")
	default:
	}

	if sshIP == "" {
		return nil, errors.New("缺少远程主机的IP地址")
	}
	if sshUser == "" {
		return nil, errors.New("缺少远程主机的用户名")
	}
	if len(sshAuths) == 0 {
		return nil, errors.New("缺少远程主机的认证信息")
	}
	if sshPort == 0 {
		sshPort = 22
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	// 设置主机密钥验证回调
	hostKeyCallback, err := getHostKeyCallback(ctx, useKnowHosts)
	if err != nil {
		return nil, err
	}

	addr := net.JoinHostPort(sshIP, strconv.FormatUint(uint64(sshPort), 10))
	sshConfig := ssh.ClientConfig{
		User:            sshUser,
		Auth:            sshAuths,
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}

	// 创建带上下文的连接
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, errors.WithMessagef(err, "TCP连接失败 (%s@%s:%d)", sshUser, sshIP, sshPort)
	}

	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		_ = conn.Close()
		return nil, errors.WithMessage(ctx.Err(), "上下文已取消")
	default:
	}

	// 协商SSH连接
	clientConn, chans, reqs, err := ssh.NewClientConn(conn, addr, &sshConfig)
	if err != nil {
		_ = conn.Close()
		return nil, errors.WithMessagef(err, "SSH协商失败 (%s@%s:%d)", sshUser, sshIP, sshPort)
	}

	return ssh.NewClient(clientConn, chans, reqs), nil
}

func getHostKeyCallback(ctx context.Context, useKnownHosts bool) (ssh.HostKeyCallback, error) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return nil, errors.WithMessage(ctx.Err(), "上下文已取消")
	default:
	}

	if !useKnownHosts {
		// #nosec G106 -- 不使用known_hosts，返回不安全的回调（仅用于开发环境）
		return ssh.InsecureIgnoreHostKey(), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.WithMessage(err, "获取用户主目录失败")
	}
	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")
	if _, err := os.Stat(knownHostsPath); err != nil {
		return nil, errors.WithMessage(err, "known_hosts 文件不可用，生产模式拒绝首次自动信任")
	}

	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, errors.WithMessage(err, "加载 known_hosts 失败")
	}
	return callback, nil
}
