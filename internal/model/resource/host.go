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

func (m *HostModel) ToUpdateMap() map[string]any {
	return map[string]any{
		"name":     m.Name,
		"label":    m.Label,
		"ssh_ip":   m.SSHIP,
		"ssh_port": m.SSHPort,
		"ssh_user": m.SSHUser,
		"py_path":  m.PyPath,
		"remark":   m.Remark,
	}
}

func ListHostModelToUint32s(ms []HostModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
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

func HostModelToAnsibleHostVars(m HostModel) AnsibleHostVars {
	return AnsibleHostVars{
		HostID:                   m.ID,
		AnsibleHost:              m.SSHIP,
		AnsiblePort:              m.SSHPort,
		AnsibleUser:              m.SSHUser,
		AnsiblePythonInterpreter: m.PyPath,
	}
}

// HostUpsertDTO 用于创建主机的请求结构体
//
// swagger:model HostUpsertDTO
type HostUpsertDTO struct {
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

func (dto *HostUpsertDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", dto.Name)
	enc.AddString("label", dto.Label)
	enc.AddString("ssh_ip", dto.SSHIP)
	enc.AddUint16("ssh_port", dto.SSHPort)
	enc.AddString("ssh_user", dto.SSHUser)
	enc.AddString("py_path", dto.PyPath)
	enc.AddString("remark", dto.Remark)
	return nil
}

func (dto *HostUpsertDTO) ToModel() HostModel {
	return HostModel{
		Name:    dto.Name,
		Label:   dto.Label,
		SSHIP:   dto.SSHIP,
		SSHPort: dto.SSHPort,
		SSHUser: dto.SSHUser,
		PyPath:  dto.PyPath,
		Remark:  dto.Remark,
	}
}

func (dto *HostUpsertDTO) ToUpdateMap() map[string]any {
	return map[string]any{
		"name":     dto.Name,
		"label":    dto.Label,
		"ssh_ip":   dto.SSHIP,
		"ssh_port": dto.SSHPort,
		"ssh_user": dto.SSHUser,
		"py_path":  dto.PyPath,
		"remark":   dto.Remark,
	}
}

// ListHostDTO 用于获取主机列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListHostDTO
type ListHostDTO struct {
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

func (req *ListHostDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", req.Name)
	enc.AddString("label", req.Label)
	enc.AddString("ssh_ip", req.SSHIP)
	if req.SSHPort != nil {
		enc.AddUint16("ssh_port", *req.SSHPort)
	} else {
		enc.AddUint16("ssh_port", 0)
	}
	enc.AddString("ssh_user", req.SSHUser)
	enc.AddString("py_path", req.PyPath)
	enc.AddString("remark", req.Remark)
	return nil
}

func (req *ListHostDTO) ToQueryMap() map[string]any {
	queryMap := req.StandardModelQuery.ToQueryMap(13)
	if req.Name != "" {
		queryMap["name like ?"] = "%" + req.Name + "%"
	}
	if req.Label != "" {
		queryMap["label = ?"] = req.Label
	}
	if req.SSHIP != "" {
		queryMap["ssh_ip = ?"] = req.SSHIP
	}
	if req.SSHPort != nil {
		queryMap["ssh_port = ?"] = *req.SSHPort
	}
	if req.SSHUser != "" {
		queryMap["ssh_user like ?"] = "%" + req.SSHUser + "%"
	}
	if req.PyPath != "" {
		queryMap["py_path like ?"] = "%" + req.PyPath + "%"
	}
	if req.Remark != "" {
		queryMap["remark like ?"] = "%" + req.Remark + "%"
	}
	return queryMap
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

// HostResp 主机响应结构
type HostResp = common.APIResp[HostStandardOut]

// PagHostResp 主机的分页响应结构
type PagHostResp = common.APIResp[*common.Pag[HostStandardOut]]

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
	ms []HostModel,
) []HostStandardOut {
	if len(ms) == 0 {
		return []HostStandardOut{}
	}
	mso := make([]HostStandardOut, 0, len(ms))
	for _, m := range ms {
		mo := HostModelToStandardOut(m)
		mso = append(mso, *mo)
	}
	return mso
}

// HostSSHDTO 主机SSH连接参数
type HostSSHDTO struct {
	HostID  uint32 `json:"host_id" form:"host_id" binding:"required"`
	Columns int    `json:"columns" form:"columns" binding:"omitempty,gt=0"`
	Rows    int    `json:"rows" form:"rows" binding:"omitempty,gt=0"`
}
