package biz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh"

	"gin-artweb/internal/infra/resource/data"
	"gin-artweb/internal/infra/resource/model"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/serializer"
)

type HostUsecase struct {
	log        *zap.Logger
	hostRepo   *data.HostRepo
	sshTimeout time.Duration
	authMethod ssh.AuthMethod
	pubKeyB64s []string
}

func NewHostUsecase(
	log *zap.Logger,
	hostRepo *data.HostRepo,
	sshTimeout time.Duration,
	authMethod ssh.AuthMethod,
	pubKeyB64s []string,
) *HostUsecase {
	return &HostUsecase{
		log:        log,
		hostRepo:   hostRepo,
		sshTimeout: sshTimeout,
		authMethod: authMethod,
		pubKeyB64s: pubKeyB64s,
	}
}

func (uc *HostUsecase) CreateHost(
	ctx context.Context,
	m model.HostModel,
	password string,
) (*model.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始创建主机",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.TestSSHConnection(ctx, m.SSHIP, m.SSHPort, m.SSHUser, password); err != nil {
		return nil, err
	}

	if err := uc.hostRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建主机失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	if err := uc.ExportHost(ctx, m); err != nil {
		return nil, err
	}

	uc.log.Info(
		"主机创建成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *HostUsecase) UpdateHostById(
	ctx context.Context,
	m model.HostModel,
	password string,
) (*model.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始更新主机",
		zap.Uint32("host_id", m.ID),
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.TestSSHConnection(ctx, m.SSHIP, m.SSHPort, m.SSHUser, password); err != nil {
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
	if err := uc.hostRepo.UpdateModel(ctx, data, "id = ?", m.ID); err != nil {
		uc.log.Error(
			"更新主机失败",
			zap.Error(err),
			zap.Uint32("host_id", m.ID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	if err := uc.ExportHost(ctx, m); err != nil {
		return nil, err
	}

	uc.log.Info(
		"更新主机成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return uc.FindHostById(ctx, m.ID)
}

func (uc *HostUsecase) DeleteHostById(
	ctx context.Context,
	hostId uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始删除主机",
		zap.Uint32("host_id", hostId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.hostRepo.DeleteModel(ctx, hostId); err != nil {
		uc.log.Error(
			"删除主机失败",
			zap.Error(err),
			zap.Uint32("host_id", hostId),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": hostId})
	}

	path := HostVarsStoragePath(hostId)
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		uc.log.Error(
			"删除ansible主机变量文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Uint32("host_id", hostId),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrDeleteCacheFileFailed.WithCause(err)
	}

	uc.log.Info(
		"删除主机成功",
		zap.Uint32("host_id", hostId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *HostUsecase) FindHostById(
	ctx context.Context,
	hostId uint32,
) (*model.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询主机",
		zap.Uint32("host_id", hostId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.hostRepo.GetModel(ctx, nil, hostId)
	if err != nil {
		uc.log.Error(
			"查询主机失败",
			zap.Error(err),
			zap.Uint32("host_id", hostId),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": hostId})
	}

	uc.log.Info(
		"查询主机成功",
		zap.Uint32("host_id", hostId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *HostUsecase) ListHost(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.HostModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询主机列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.hostRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询主机列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询主机列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *HostUsecase) TestSSHConnection(
	ctx context.Context,
	sshIP string,
	sshPort uint16,
	sshUser, sshPassword string,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始测试ssh连接",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	cli, err := uc.hostRepo.NewSSHClient(ctx, sshIP, sshPort, sshUser, []ssh.AuthMethod{uc.authMethod}, uc.sshTimeout)
	if err == nil {
		uc.log.Info(
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

	client, err := uc.hostRepo.NewSSHClient(ctx, sshIP, sshPort, sshUser, sshAuths, uc.sshTimeout)
	if err != nil {
		uc.log.Error(
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
		uc.log.Error(
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

	for _, pubKeyB64 := range uc.pubKeyB64s {
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
		if err := uc.hostRepo.ExecuteCommand(ctx, session, script); err != nil {
			uc.log.Error(
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

	uc.log.Info(
		"测试ssh连接通过",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *HostUsecase) ExportHost(ctx context.Context, m model.HostModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始导出ansible主机变量文件",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ansibleHost := AnsibleHostVars{
		HostID:                   m.ID,
		AnsibleHost:              m.SSHIP,
		AnsiblePort:              m.SSHPort,
		AnsibleUser:              m.SSHUser,
		AnsiblePythonInterpreter: m.PyPath,
	}

	path := HostVarsStoragePath(m.ID)
	if _, err := serializer.WriteYAML(path, ansibleHost); err != nil {
		uc.log.Error(
			"导出ansible主机变量文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Object("ansible_host", &ansibleHost),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	uc.log.Info(
		"导出ansible主机变量文件成功",
		zap.String("path", path),
		zap.Object("ansible_host", &ansibleHost),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

type AnsibleHostVars struct {
	HostID                   uint32 `json:"host_id" yaml:"host_id"`
	AnsibleHost              string `json:"ansible_host" yaml:"ansible_host"`
	AnsiblePort              uint16 `json:"ansible_port" yaml:"ansible_port"`
	AnsibleUser              string `json:"ansible_user" yaml:"ansible_user"`
	AnsiblePythonInterpreter string `json:"ansible_python_interpreter" yaml:"ansible_python_interpreter"`
}

func (vs *AnsibleHostVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("host_id", vs.HostID)
	enc.AddString("ansible_host", vs.AnsibleHost)
	enc.AddUint16("ansible_port", vs.AnsiblePort)
	enc.AddString("ansible_user", vs.AnsibleUser)
	enc.AddString("ansible_python_interpreter", vs.AnsiblePythonInterpreter)
	return nil
}

func HostVarsStoragePath(pk uint32) string {
	filename := fmt.Sprintf("host_%d.yaml", pk)
	return filepath.Join(config.StorageDir, "host_vars", filename)
}
