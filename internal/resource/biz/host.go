package biz

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh"

	"gin-artweb/pkg/common"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
	"gin-artweb/pkg/file"
)

const (
	HostIDKey  = "host_id"
	HostIDsKey = "host_ids"
)

type AnsibleHostVars struct {
	ID                       uint32 `json:"id"`
	AnsibleHost              string `json:"ansible_host"`
	AnsiblePort              uint16 `json:"ansible_port"`
	AnsibleUser              string `json:"ansible_user"`
	AnsiblePythonInterpreter string `json:"ansible_python_interpreter,omitempty"`
}

func (h *AnsibleHostVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", h.ID)
	enc.AddString("ansible_host", h.AnsibleHost)
	enc.AddUint16("ansible_port", h.AnsiblePort)
	enc.AddString("ansible_user", h.AnsibleUser)
	enc.AddString("ansible_python_interpreter", h.AnsiblePythonInterpreter)
	return nil
}

type HostModel struct {
	database.StandardModel
	Name     string `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Label    string `gorm:"column:label;type:varchar(50);index:idx_host_label;comment:标签" json:"label"`
	IPAddr   string `gorm:"column:ip_addr;type:varchar(108);comment:IP地址" json:"ip_addr"`
	Port     uint16 `gorm:"column:port;type:smallint;comment:端口" json:"port"`
	Username string `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	PyPath   string `gorm:"column:py_path;type:varchar(254);comment:python路径" json:"py_path"`
	Remark   string `gorm:"column:remark;type:varchar(254);comment:备注" json:"remark"`
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
	enc.AddString("ip_addr", m.IPAddr)
	enc.AddUint16("port", m.Port)
	enc.AddString("username", m.Username)
	enc.AddString("py_path", m.PyPath)
	enc.AddString("remark", m.Remark)
	return nil
}

func (m *HostModel) ExportAnsibleHostVars() AnsibleHostVars {
	return AnsibleHostVars{
		AnsibleHost:              m.IPAddr,
		AnsiblePort:              m.Port,
		AnsibleUser:              m.Username,
		AnsiblePythonInterpreter: m.PyPath,
	}
}

type HostRepo interface {
	CreateModel(context.Context, *HostModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*HostModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]HostModel, error)
	NewSSHClient(context.Context, string, ssh.ClientConfig) (*ssh.Client, error)
	DeployPublicKey(context.Context, *ssh.Client, ssh.PublicKey) error
}

type HostUsecase struct {
	log      *zap.Logger
	hostRepo HostRepo
	signer   ssh.Signer
	timeout  time.Duration
	dir      string
}

func NewHostUsecase(
	log *zap.Logger,
	hostRepo HostRepo,
	signer ssh.Signer,
	timeout time.Duration,
	dir string,
) *HostUsecase {
	return &HostUsecase{
		log:      log,
		hostRepo: hostRepo,
		signer:   signer,
		timeout:  timeout,
		dir:      dir,
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

	if err := uc.TestSSHConnection(ctx, m.IPAddr, m.Port, m.Username, password); err != nil {
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
	hostId uint32,
	m HostModel,
	password string,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始更新主机",
		zap.Uint32(HostIDKey, hostId),
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.TestSSHConnection(ctx, m.IPAddr, m.Port, m.Username, password); err != nil {
		return err
	}

	data := map[string]any{
		"name":     m.Name,
		"label":    m.Label,
		"ip_addr":  m.IPAddr,
		"port":     m.Port,
		"username": m.Username,
		"py_path":  m.PyPath,
		"remark":   m.Remark,
	}
	if err := uc.hostRepo.UpdateModel(ctx, data, "id = ?", hostId); err != nil {
		uc.log.Error(
			"更新主机失败",
			zap.Error(err),
			zap.Uint32(HostIDKey, hostId),
			zap.Any(database.UpdateDataKey, data),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, data)
	}

	if err := uc.ExportHost(ctx, m); err != nil {
		return err
	}

	uc.log.Info(
		"更新主机成功",
		zap.Uint32(HostIDKey, hostId),
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
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

	path := uc.HostPath(hostId)
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		uc.log.Error(
			"删除ansible主机变量文件失败",
			zap.String("path", path),
			zap.Uint32("host_id", hostId),
			zap.Error(err),
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
		zap.String("ip_addr", ip),
		zap.Uint16("port", port),
		zap.String("username", user),
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
			zap.String("ip_addr", ip),
			zap.Uint16("port", port),
			zap.String("username", user),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrSSHConnect.WithCause(err)
	}
	defer client.Close()
	if err := uc.hostRepo.DeployPublicKey(ctx, client, uc.signer.PublicKey()); err != nil {
		uc.log.Error(
			"部署ssh公钥失败",
			zap.Error(err),
			zap.String("ip_addr", ip),
			zap.Uint16("port", port),
			zap.String("username", user),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrSSHKeyDeployment.WithCause(err)
	}
	uc.log.Info(
		"测试ssh连接通过",
		zap.String("ip_addr", ip),
		zap.Uint16("port", port),
		zap.String("username", user),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *HostUsecase) HostPath(pk uint32) string {
	filename := fmt.Sprintf("host_%d.json", pk)
	return filepath.Join(uc.dir, filename)
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

	path := uc.HostPath(m.ID)
	ansibleHost := AnsibleHostVars{
		ID:                       m.ID,
		AnsibleHost:              m.IPAddr,
		AnsiblePort:              m.Port,
		AnsibleUser:              m.Username,
		AnsiblePythonInterpreter: m.PyPath,
	}

	if err := file.WriteJSON(path, ansibleHost, 4); err != nil {
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
