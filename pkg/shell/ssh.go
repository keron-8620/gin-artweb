package shell

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func NewSSHClient(
	sshIP string,
	sshPort uint16,
	sshUser string,
	sshAuths []ssh.AuthMethod,
	useKnowHosts bool,
	timeout time.Duration,
) (*ssh.Client, error) {
	if sshIP == "" {
		return nil, fmt.Errorf("缺少远程主机的ip地址")
	}
	if sshUser == "" {
		return nil, fmt.Errorf("缺少远程主机的用户名")
	}
	if len(sshAuths) == 0 {
		return nil, fmt.Errorf("缺少远程主机的认证信息")
	}
	if sshPort == 0 {
		sshPort = 22
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	// 设置主机密钥验证回调
	hostKeyCallback, err := getHostKeyCallback(useKnowHosts)
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
	client, err := ssh.Dial("tcp", addr, &sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH连接失败 (%s@%s:%d): %w",
			sshUser, sshIP, sshPort, err)
	}

	return client, nil
}

// getHostKeyCallback 获取主机密钥验证回调函数
func getHostKeyCallback(useKnownHosts bool) (ssh.HostKeyCallback, error) {
	if !useKnownHosts {
		// 不使用known_hosts，返回不安全的回调（仅用于开发环境）
		return ssh.InsecureIgnoreHostKey(), nil
	}

	// 尝试使用默认的known_hosts路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 如果无法获取用户目录，回退到不安全的回调
		return nil, errors.New("无法通过获取当前用户的家目录来找到默认的known_hosts文件")
	}

	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")

	// 检查known_hosts文件是否存在
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		return nil, errors.New("缺少主机的known_hosts文件")
	}

	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, errors.New("创建known_hosts回调失败")
	}

	return callback, nil
}
