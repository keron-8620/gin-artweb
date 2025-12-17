package host

import (
	"gin-artweb/api/common"

	"go.uber.org/zap/zapcore"
)

// CreateOrUpdateHosrRequest 用于创建主机的请求结构体
//
// swagger:model CreateOrUpdateHosrRequest
type CreateOrUpdateHosrRequest struct {
	// 名称，最大长度50
	// Required: true
	// Max length: 50
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 标签，最大长度50
	// Required: true
	// Max length: 50
	Label string `json:"label" form:"label" binding:"required,max=50"`

	// ip地址，最大长度108
	// Required: true
	// Max length: 108
	SSHIP string `json:"ssh_ip" form:"ssh_ip" binding:"required,max=108"`

	// 端口，必填
	// Required: true
	SSHPort uint16 `json:"ssh_port" form:"ssh_port" binding:"required,gt=0"`

	// 用户名，最大长度50
	// Required: true
	// Max length: 50
	SSHUser string `json:"ssh_user" form:"ssh_user" binding:"required,max=50"`

	// 密码，最大长度150
	// Required: true
	// Max length: 150
	SSHPassword string `json:"ssh_password" form:"ssh_password" binding:"required,max=150"`

	// python路径，最大长度254
	// Required: true
	// Max length: 254
	PyPath string `json:"py_path" form:"py_path" binding:"omitempty,max=254"`

	// 备注，最大长度254
	// Max length: 254
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

	// 名称，最大长度50
	// Required: true
	// Max length: 50
	Name string `form:"name" binding:"omitempty,max=50"`

	// 标签，最大长度50
	// Required: true
	// Max length: 50
	Label string `form:"label" binding:"omitempty,max=50"`

	// ip地址，最大长度108
	// Required: true
	// Max length: 108
	SSHIP string `form:"ssh_ip" binding:"omitempty,max=108"`

	// 端口，必填
	// Required: true
	SSHPort *uint16 `form:"ssh_port" binding:"omitempty,gt=0"`

	// 用户名，最大长度50
	// Required: true
	// Max length: 50
	SSHUser string `form:"ssh_user" binding:"omitempty,max=50"`

	// python路径，最大长度254
	// Required: true
	// Max length: 254
	PyPath string `form:"py_path" binding:"omitempty,max=254"`

	// 备注，最大长度254
	// Max length: 254
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
		query["username list ?"] = "%" + req.SSHUser + "%"
	}
	if req.PyPath != "" {
		query["py_path lisk ?"] = "%" + req.PyPath + "%"
	}
	if req.Remark != "" {
		query["remark like ?"] = "%" + req.Remark + "%"
	}
	return page, size, query
}
