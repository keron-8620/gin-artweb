package resource

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	resomodel "gin-artweb/internal/model/resource"
	resorepo "gin-artweb/internal/repository/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/serializer"
)

type HostService struct {
	log        *zap.Logger
	hostRepo   *resorepo.HostRepo
	sshTimeout time.Duration
	authMethod ssh.AuthMethod
	pubKeyB64s []string
}

func NewHostService(
	log *zap.Logger,
	hostRepo *resorepo.HostRepo,
	sshTimeout time.Duration,
	authMethod ssh.AuthMethod,
	pubKeyB64s []string,
) *HostService {
	return &HostService{
		log:        log,
		hostRepo:   hostRepo,
		sshTimeout: sshTimeout,
		authMethod: authMethod,
		pubKeyB64s: pubKeyB64s,
	}
}

func (s *HostService) CreateHost(
	ctx context.Context,
	m resomodel.HostModel,
	password string,
) (*resomodel.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建主机",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.TestSSHConnection(ctx, m.SSHIP, m.SSHPort, m.SSHUser, password); err != nil {
		return nil, err
	}

	if err := s.hostRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建主机失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	if err := s.ExportHost(ctx, m); err != nil {
		return nil, err
	}

	s.log.Info(
		"主机创建成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (s *HostService) UpdateHostById(
	ctx context.Context,
	m resomodel.HostModel,
	password string,
) (*resomodel.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新主机",
		zap.Uint32("host_id", m.ID),
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.TestSSHConnection(ctx, m.SSHIP, m.SSHPort, m.SSHUser, password); err != nil {
		return nil, err
	}

	data := map[string]any{
		"name":     m.Name,
		"label":    m.Label,
		"ssh_ip":   m.SSHIP,
		"ssh_port": m.SSHPort,
		"ssh_user": m.SSHUser,
		"py_path":  m.PyPath,
		"remark":   m.Remark,
	}
	if err := s.hostRepo.UpdateModel(ctx, data, "id = ?", m.ID); err != nil {
		s.log.Error(
			"更新主机失败",
			zap.Error(err),
			zap.Uint32("host_id", m.ID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	if err := s.ExportHost(ctx, m); err != nil {
		return nil, err
	}

	s.log.Info(
		"更新主机成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return s.FindHostById(ctx, m.ID)
}

func (s *HostService) DeleteHostById(
	ctx context.Context,
	hostId uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除主机",
		zap.Uint32("host_id", hostId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.hostRepo.DeleteModel(ctx, hostId); err != nil {
		s.log.Error(
			"删除主机失败",
			zap.Error(err),
			zap.Uint32("host_id", hostId),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": hostId})
	}

	path := common.GetHostVarsExportPath(hostId)
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		s.log.Error(
			"删除ansible主机变量文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Uint32("host_id", hostId),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrDeleteCacheFileFailed.WithCause(err)
	}

	s.log.Info(
		"删除主机成功",
		zap.Uint32("host_id", hostId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *HostService) FindHostById(
	ctx context.Context,
	hostId uint32,
) (*resomodel.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询主机",
		zap.Uint32("host_id", hostId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.hostRepo.GetModel(ctx, nil, hostId)
	if err != nil {
		s.log.Error(
			"查询主机失败",
			zap.Error(err),
			zap.Uint32("host_id", hostId),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": hostId})
	}

	s.log.Info(
		"查询主机成功",
		zap.Uint32("host_id", hostId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *HostService) ListHost(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]resomodel.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询主机列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.hostRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询主机列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询主机列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *HostService) TestSSHConnection(
	ctx context.Context,
	sshIP string,
	sshPort uint16,
	sshUser, sshPassword string,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始测试ssh连接",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	cli, err := s.hostRepo.NewSSHClient(ctx, sshIP, sshPort, sshUser, []ssh.AuthMethod{s.authMethod}, s.sshTimeout)
	if err == nil {
		s.log.Info(
			"主机已认证并部署密钥的主机",
			zap.String("ssh_ip", sshIP),
			zap.Uint16("ssh_port", sshPort),
			zap.String("ssh_user", sshUser),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		cli.Close()
		return nil
	}

	sshAuths := []ssh.AuthMethod{
		ssh.Password(sshPassword),
	}

	client, err := s.hostRepo.NewSSHClient(ctx, sshIP, sshPort, sshUser, sshAuths, s.sshTimeout)
	if err != nil {
		s.log.Error(
			"创建ssh连接失败",
			zap.Error(err),
			zap.String("ssh_ip", sshIP),
			zap.Uint16("ssh_port", sshPort),
			zap.String("ssh_user", sshUser),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrSSHConnectionFailed.WithCause(err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		s.log.Error(
			"创建ssh session失败",
			zap.Error(err),
			zap.String("ssh_ip", sshIP),
			zap.Uint16("ssh_port", sshPort),
			zap.String("ssh_user", sshUser),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrSSHConnectionFailed.WithCause(err)
	}
	defer session.Close()

	for _, pubKeyB64 := range s.pubKeyB64s {
		script := `
			mkdir -p ~/.ssh
			tmp_key=$(mktemp)
			echo '` + pubKeyB64 + `' | base64 -d > "$tmp_key"
			if ! grep -Fq "$(cat "$tmp_key")" ~/.ssh/authorized_keys 2>/dev/null; then
				cat "$tmp_key" >> ~/.ssh/authorized_keys
			fi
			rm -f "$tmp_key"
			chmod 700 ~/.ssh
			chmod 600 ~/.ssh/authorized_keys
		`
		if err := s.hostRepo.ExecuteCommand(ctx, session, script); err != nil {
			s.log.Error(
				"部署ssh公钥失败",
				zap.Error(err),
				zap.String("ssh_ip", sshIP),
				zap.Uint16("ssh_port", sshPort),
				zap.String("ssh_user", sshUser),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			return errors.ErrSSHKeyDeployFailed.WithCause(err)
		}
	}

	s.log.Info(
		"测试ssh连接通过",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *HostService) ExportHost(ctx context.Context, m resomodel.HostModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始导出ansible主机变量文件",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ansibleHost := resomodel.AnsibleHostVars{
		HostID:                   m.ID,
		AnsibleHost:              m.SSHIP,
		AnsiblePort:              m.SSHPort,
		AnsibleUser:              m.SSHUser,
		AnsiblePythonInterpreter: m.PyPath,
	}

	path := common.GetHostVarsExportPath(m.ID)
	if _, err := serializer.WriteYAML(path, ansibleHost); err != nil {
		s.log.Error(
			"导出ansible主机变量文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Object("ansible_host", &ansibleHost),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	s.log.Info(
		"导出ansible主机变量文件成功",
		zap.String("path", path),
		zap.Object("ansible_host", &ansibleHost),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}
