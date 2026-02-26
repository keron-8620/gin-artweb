package resource

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
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
	if m == nil {
		return nil
	}
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

// CreateOrUpdateHosrRequest 用于创建主机的请求结构体
//
// swagger:model CreateOrUpdateHosrRequest
type CreateOrUpdateHosrRequest struct {
	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 标签
	Label string `json:"label" form:"label" binding:"required,max=50"`

	// ip地址
	SSHIP string `json:"ssh_ip" form:"ssh_ip" binding:"required,max=108"`

	// 端口
	SSHPort uint16 `json:"ssh_port" form:"ssh_port" binding:"required,gt=0"`

	// 用户名
	SSHUser string `json:"ssh_user" form:"ssh_user" binding:"required,max=50"`

	// 密码
	SSHPassword string `json:"ssh_password" form:"ssh_password" binding:"required,max=150"`

	// python路径
	PyPath string `json:"py_path" form:"py_path" binding:"omitempty,max=254"`

	// 备注
	Remark string `json:"remark" form:"remark" binding:"max=254"`
}

func (req *CreateOrUpdateHosrRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", req.Name)
	enc.AddString("label", req.Label)
	enc.AddString("ssh_ip", req.SSHIP)
	enc.AddUint16("ssh_port", req.SSHPort)
	enc.AddString("ssh_user", req.SSHUser)
	enc.AddString("py_path", req.PyPath)
	enc.AddString("remark", req.Remark)
	return nil
}

// ListHostRequest 用于获取主机列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListHostRequest
type ListHostRequest struct {
	common.StandardModelQuery

	// 名称
	Name string `form:"name" binding:"omitempty,max=50"`

	// 标签
	Label string `form:"label" binding:"omitempty,max=50"`

	// ip地址
	SSHIP string `form:"ssh_ip" binding:"omitempty,max=108"`

	// 端口
	SSHPort *uint16 `form:"ssh_port" binding:"omitempty,gt=0"`

	// 用户名
	SSHUser string `form:"ssh_user" binding:"omitempty,max=50"`

	// python路径
	PyPath string `form:"py_path" binding:"omitempty,max=254"`

	// 备注
	Remark string `form:"remark" binding:"omitempty,max=254"`
}

func (req *ListHostRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(13)
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.Label != "" {
		query["label = ?"] = req.Label
	}
	if req.SSHIP != "" {
		query["ip_addr = ?"] = req.SSHIP
	}
	if req.SSHPort != nil {
		query["ssh_port = ?"] = *req.SSHPort
	}
	if req.SSHUser != "" {
		query["ssh_user like ?"] = "%" + req.SSHUser + "%"
	}
	if req.PyPath != "" {
		query["py_path like ?"] = "%" + req.PyPath + "%"
	}
	if req.Remark != "" {
		query["remark like ?"] = "%" + req.Remark + "%"
	}
	return page, size, query
}

type HostBaseOut struct {
	// 主机ID
	ID uint32 `json:"id" example:"1"`

	// 名称
	Name string `json:"name" example:"artweb主机"`

	// 标签
	Label string `json:"label" example:"artweb"`

	// IP地址
	SSHIP string `json:"ssh_ip" example:"192.168.1.1"`

	// 端口
	SSHPort uint16 `json:"ssh_port" example:"22"`

	// 用户名
	SSHUser string `json:"ssh_user" example:"root"`

	// Python路径
	PyPath string `json:"py_path" example:"/usr/bin/python3"`

	// 备注
	Remark string `json:"remark" example:"测试"`
}

// HostStandardOut 主机基础信息
type HostStandardOut struct {
	HostBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

// HostReply 主机响应结构
type HostReply = common.APIReply[HostStandardOut]

// PagHostReply 主机的分页响应结构
type PagHostReply = common.APIReply[*common.Pag[HostStandardOut]]

func HostModelToBaseOut(
	m HostModel,
) *HostBaseOut {
	return &HostBaseOut{
		ID:      m.ID,
		Name:    m.Name,
		Label:   m.Label,
		SSHIP:   m.SSHIP,
		SSHPort: m.SSHPort,
		SSHUser: m.SSHUser,
		PyPath:  m.PyPath,
		Remark:  m.Remark,
	}
}

func HostModelToStandardOut(
	m HostModel,
) *HostStandardOut {
	return &HostStandardOut{
		HostBaseOut: *HostModelToBaseOut(m),
		CreatedAt:   m.CreatedAt.Format(time.DateTime),
		UpdatedAt:   m.UpdatedAt.Format(time.DateTime),
	}
}

func ListHostModelToStandardOut(
	hms *[]HostModel,
) *[]HostStandardOut {
	if hms == nil {
		return &[]HostStandardOut{}
	}

	ms := *hms
	mso := make([]HostStandardOut, 0, len(ms))
	for _, m := range ms {
		mo := HostModelToStandardOut(m)
		mso = append(mso, *mo)
	}
	return &mso
}
