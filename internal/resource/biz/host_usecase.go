package biz

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pkg/sftp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh"

	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/serializer"
)

const (
	HostIDKey = "host_id"
)

type HostModel struct {
	database.StandardModel
	Name    string `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Label   string `gorm:"column:label;type:varchar(50);index:idx_host_label;comment:标签" json:"label"`
	SSHIP   string `gorm:"column:ssh_ip;type:varchar(108);uniqueIndex:idx_host_ip_port_user;comment:IP地址" json:"ssh_ip"`
	SSHPort uint16 `gorm:"column:ssh_port;type:smallint;uniqueIndex:idx_host_ip_port_user;comment:端口" json:"ssh_port"`
	SSHUser string `gorm:"column:ssh_user;type:varchar(50);uniqueIndex:idx_host_ip_port_user;comment:用户名" json:"ssh_user"`
	PyPath  string `gorm:"column:py_path;type:varchar(254);comment:python路径" json:"py_path"`
	Remark  string `gorm:"column:remark;type:varchar(254);comment:备注" json:"remark"`
}

func (m *HostModel) TableName() string {
	return "resource_host"
}

func (m *HostModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("label", m.Label)
	enc.AddString("ssh_ip", m.SSHIP)
	enc.AddUint16("ssh_port", m.SSHPort)
	enc.AddString("ssh_user", m.SSHUser)
	enc.AddString("py_path", m.PyPath)
	enc.AddString("remark", m.Remark)
	return nil
}

type HostRepo interface {
	CreateModel(context.Context, *HostModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*HostModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]HostModel, error)
	NewSSHClient(context.Context, string, ssh.ClientConfig) (*ssh.Client, error)
	NewSession(context.Context, *ssh.Client) (*ssh.Session, error)
	ExecuteCommand(context.Context, *ssh.Session, string) error
	NewSFTPClient(context.Context, *ssh.Client) (*sftp.Client, error)
	UploadFile(context.Context, *sftp.Client, string, string) error
	DownloadFile(context.Context, *sftp.Client, string, string) error
}

type HostUsecase struct {
	log      *zap.Logger
	hostRepo HostRepo
	signer   ssh.Signer
	timeout  time.Duration
}

func NewHostUsecase(
	log *zap.Logger,
	hostRepo HostRepo,
	signer ssh.Signer,
	timeout time.Duration,
) *HostUsecase {
	return &HostUsecase{
		log:      log,
		hostRepo: hostRepo,
		signer:   signer,
		timeout:  timeout,
	}
}

func (uc *HostUsecase) CreateHost(
	ctx context.Context,
	m HostModel,
	password string,
) (*HostModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建主机",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.TestSSHConnection(ctx, m.SSHIP, m.SSHPort, m.SSHUser, password); err != nil {
		return nil, err
	}

	if err := uc.hostRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建主机失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	if err := uc.ExportHost(ctx, m); err != nil {
		return nil, err
	}

	uc.log.Info(
		"主机创建成功",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *HostUsecase) UpdateHostById(
	ctx context.Context,
	m HostModel,
	password string,
) (*HostModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始更新主机",
		zap.Uint32(HostIDKey, m.ID),
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
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
			zap.Uint32(HostIDKey, m.ID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, data)
	}

	if err := uc.ExportHost(ctx, m); err != nil {
		return nil, err
	}

	uc.log.Info(
		"更新主机成功",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return uc.FindHostById(ctx, m.ID)
}

func (uc *HostUsecase) DeleteHostById(
	ctx context.Context,
	hostId uint32,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除主机",
		zap.Uint32(HostIDKey, hostId),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.hostRepo.DeleteModel(ctx, hostId); err != nil {
		uc.log.Error(
			"删除主机失败",
			zap.Error(err),
			zap.Uint32(HostIDKey, hostId),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, map[string]any{"id": hostId})
	}

	path := HostVarsStoragePath(hostId)
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		uc.log.Error(
			"删除ansible主机变量文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Uint32("host_id", hostId),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrDeleteHostFileFailed.WithCause(err)
	}

	uc.log.Info(
		"删除主机成功",
		zap.Uint32(HostIDKey, hostId),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *HostUsecase) FindHostById(
	ctx context.Context,
	hostId uint32,
) (*HostModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询主机",
		zap.Uint32(HostIDKey, hostId),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := uc.hostRepo.FindModel(ctx, nil, hostId)
	if err != nil {
		uc.log.Error(
			"查询主机失败",
			zap.Error(err),
			zap.Uint32(HostIDKey, hostId),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": hostId})
	}

	uc.log.Info(
		"查询主机成功",
		zap.Uint32(HostIDKey, hostId),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *HostUsecase) ListHost(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]HostModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询主机列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	count, ms, err := uc.hostRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询主机列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询主机列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *HostUsecase) TestSSHConnection(
	ctx context.Context,
	ip string,
	port uint16,
	user, password string,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始测试ssh连接",
		zap.String("ssh_ip", ip),
		zap.Uint16("ssh_port", port),
		zap.String("ssh_user", user),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	sshConfig := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(uc.signer),
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         uc.timeout,
	}

	addr := net.JoinHostPort(ip, strconv.FormatUint(uint64(port), 10))
	client, err := uc.hostRepo.NewSSHClient(ctx, addr, sshConfig)
	if err != nil {
		uc.log.Error(
			"创建ssh连接失败",
			zap.Error(err),
			zap.String("ssh_ip", ip),
			zap.Uint16("ssh_port", port),
			zap.String("ssh_user", user),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrSSHConnect.WithCause(err)
	}
	defer client.Close()

	session, err := uc.hostRepo.NewSession(ctx, client)
	if err != nil {
		uc.log.Error(
			"创建ssh session失败",
			zap.Error(err),
			zap.String("ssh_ip", ip),
			zap.Uint16("ssh_port", port),
			zap.String("ssh_user", user),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrSSHConnect.WithCause(err)
	}
	defer session.Close()

	pubKeyBytes := ssh.MarshalAuthorizedKey(uc.signer.PublicKey())
	pubKeyB64 := base64.StdEncoding.EncodeToString(pubKeyBytes)
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
			zap.String("ssh_ip", ip),
			zap.Uint16("ssh_port", port),
			zap.String("ssh_user", user),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrSSHKeyDeployment.WithCause(err)
	}
	uc.log.Info(
		"测试ssh连接通过",
		zap.String("ssh_ip", ip),
		zap.Uint16("ssh_port", port),
		zap.String("ssh_user", user),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *HostUsecase) ExportHost(ctx context.Context, m HostModel) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始导出ansible主机变量文件",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
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
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrExportHostFailed.WithCause(err)
	}

	uc.log.Info(
		"导出ansible主机变量文件成功",
		zap.String("path", path),
		zap.Object("ansible_host", &ansibleHost),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
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
