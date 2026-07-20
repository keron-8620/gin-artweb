package resource

import (
	"context"
	"io"
	"sync"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	resomodel "gin-artweb/internal/model/resource"
	hostrepo "gin-artweb/internal/repo/resource"
	"gin-artweb/internal/shared/ctxutil"
)

type TerminalService struct {
	logger   *zap.Logger
	hostRepo *hostrepo.HostRepo
	auths    []ssh.AuthMethod
}

func NewTerminalService(
	logger *zap.Logger,
	hostRepo *hostrepo.HostRepo,
	auths []ssh.AuthMethod,
) *TerminalService {
	return &TerminalService{
		logger:   logger,
		hostRepo: hostRepo,
		auths:    auths,
	}
}

type TerminalSession struct {
	client    *ssh.Client
	session   *ssh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	stderr    io.Reader
	closeOnce sync.Once
	isClosed  bool
	mu        sync.Mutex
}

func (s *TerminalService) CreateSessionByHostID(
	ctx context.Context,
	hostID uint32,
	termType string,
	cols, rows int,
) (*TerminalSession, error) {
	log := ctxutil.NewLogger(s.logger, ctx)

	host, err := s.hostRepo.GetModel(ctx, hostID)
	if err != nil {
		log.Error("查询主机信息失败", zap.Error(err), zap.Uint32("host_id", hostID))
		return nil, errors.WithMessage(err, "查询主机信息失败")
	}

	return s.CreateSession(ctx, *host, termType, cols, rows)
}

func (s *TerminalService) CreateSession(
	ctx context.Context,
	host resomodel.HostModel,
	termType string,
	cols, rows int,
) (*TerminalSession, error) {
	log := ctxutil.NewLogger(s.logger, ctx)

	client, err := s.hostRepo.NewSSHClient(ctx, host.SSHIP, host.SSHPort, host.SSHUser, s.auths, 30*time.Second)
	if err != nil {
		log.Error("创建SSH客户端失败", zap.Error(err),
			zap.String("ssh_ip", host.SSHIP),
			zap.Uint16("ssh_port", host.SSHPort),
			zap.String("ssh_user", host.SSHUser))
		return nil, errors.WithMessage(err, "创建SSH客户端失败")
	}

	session, err := client.NewSession()
	if err != nil {
		_ = client.Close()
		log.Error("创建SSH会话失败", zap.Error(err))
		return nil, errors.WithMessage(err, "创建SSH会话失败")
	}

	if err = session.RequestPty(termType, rows, cols, ssh.TerminalModes{}); err != nil {
		_ = session.Close()
		_ = client.Close()
		log.Error("请求PTY失败", zap.Error(err))
		return nil, errors.WithMessage(err, "请求PTY失败")
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		_ = session.Close()
		_ = client.Close()
		log.Error("获取标准输入管道失败", zap.Error(err))
		return nil, errors.WithMessage(err, "获取标准输入管道失败")
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		_ = session.Close()
		_ = client.Close()
		log.Error("获取标准输出管道失败", zap.Error(err))
		return nil, errors.WithMessage(err, "获取标准输出管道失败")
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		_ = stdin.Close()
		_ = stdout.(io.Closer).Close()
		_ = session.Close()
		_ = client.Close()
		log.Error("获取标准错误管道失败", zap.Error(err))
		return nil, errors.WithMessage(err, "获取标准错误管道失败")
	}

	return &TerminalSession{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
	}, nil
}

func (ts *TerminalSession) StartShell() error {
	return ts.session.Shell()
}

func (ts *TerminalSession) Resize(cols, rows int) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.isClosed {
		return errors.New("会话已关闭")
	}

	return ts.session.WindowChange(rows, cols)
}

func (ts *TerminalSession) Write(data []byte) (int, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.isClosed {
		return 0, errors.New("会话已关闭")
	}

	return ts.stdin.Write(data)
}

func (ts *TerminalSession) Read() ([]byte, error) {
	buf := make([]byte, 4096)
	n, err := ts.stdout.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (ts *TerminalSession) ReadErr() ([]byte, error) {
	buf := make([]byte, 4096)
	n, err := ts.stderr.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (ts *TerminalSession) Close() {
	ts.closeOnce.Do(func() {
		ts.mu.Lock()
		ts.isClosed = true
		ts.mu.Unlock()

		if ts.stdin != nil {
			_ = ts.stdin.Close()
		}
		if ts.session != nil {
			_ = ts.session.Close()
		}
		if ts.client != nil {
			_ = ts.client.Close()
		}
	})
}
