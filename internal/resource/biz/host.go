package biz

import (
	"context"
	"net"
	"strconv"

	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh"
)

type AnsibleHostVars struct {
	ID                       uint32 `json:"id"`
	AnsibleHost              string `json:"ansible_host"`
	AnsiblePort              uint16 `json:"ansible_port"`
	AnsibleUser              string `json:"ansible_user"`
	AnsiblePythonInterpreter string `json:"ansible_python_interpreter,omitempty"`
}

type HostModel struct {
	database.StandardModel
	Name     string `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Label    string `gorm:"column:label;type:varchar(50);index:idx_member;comment:标签" json:"label"`
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
	ListModel(context.Context, database.QueryParams) (int64, []HostModel, error)
	NewSSHClient(context.Context, string, ssh.ClientConfig) (*ssh.Client, error)
	DeployPublicKey(*ssh.Client, ssh.PublicKey) error
}

type HostUsecase struct {
	log      *zap.Logger
	hostRepo HostRepo
	signer   ssh.Signer
}

func NewHostUsecase(
	log *zap.Logger,
	hostRepo HostRepo,
	signer ssh.Signer,
) *HostUsecase {
	return &HostUsecase{
		log:      log,
		hostRepo: hostRepo,
		signer:   signer,
	}
}

func (uc *HostUsecase) TestSSHConnection(ctx context.Context, ip string, port uint16, user, password string) *errors.Error {
	sshConfig := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(uc.signer),
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := net.JoinHostPort(ip, strconv.FormatUint(uint64(port), 10))
	client, err := uc.hostRepo.NewSSHClient(ctx, addr, sshConfig)
	if err != nil {
		return ErrSSHConnect.WithCause(err)
	}
	if err := uc.hostRepo.DeployPublicKey(client, uc.signer.PublicKey()); err != nil {
		return ErrSSHKeyDeployment.WithCause(err)
	}
	return nil
}

func (uc *HostUsecase) CreateHost(
	ctx context.Context,
	m HostModel,
	password string,
) (*HostModel, *errors.Error) {
	if err := uc.TestSSHConnection(ctx, m.IPAddr, m.Port, m.Username, password); err != nil {
		return nil, err
	}
	if err := uc.hostRepo.CreateModel(ctx, &m); err != nil {
		return nil, database.NewGormError(err, nil)
	}
	return &m, nil
}

func (uc *HostUsecase) UpdateHostById(
	ctx context.Context,
	hostId uint32,
	m HostModel,
	password string,
) *errors.Error {
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
		return database.NewGormError(err, data)
	}
	return nil
}

func (uc *HostUsecase) DeleteHostById(
	ctx context.Context,
	hostId uint32,
) *errors.Error {
	if err := uc.hostRepo.DeleteModel(ctx, hostId); err != nil {
		return database.NewGormError(err, map[string]any{"id": hostId})
	}
	return nil
}

func (uc *HostUsecase) FindHostById(
	ctx context.Context,
	hostId uint32,
) (*HostModel, *errors.Error) {
	m, err := uc.hostRepo.FindModel(ctx, nil, hostId)
	if err != nil {
		return nil, database.NewGormError(err, map[string]any{"id": hostId})
	}
	return m, nil
}

func (uc *HostUsecase) ListHost(
	ctx context.Context,
	page, size int,
	query map[string]any,
	orderBy []string,
	isCount bool,
) (int64, []HostModel, *errors.Error) {
	qp := database.QueryParams{
		Preloads: []string{},
		Query:    query,
		OrderBy:  orderBy,
		Limit:    max(size, 0),
		Offset:   max(page-1, 0),
		IsCount:  isCount,
	}
	count, ms, err := uc.hostRepo.ListModel(ctx, qp)
	if err != nil {
		return 0, nil, database.NewGormError(err, nil)
	}
	return count, ms, nil
}
