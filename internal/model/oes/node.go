package oes

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
)

type OesNodeModel struct {
	database.StandardModel
	NodeRole    string             `gorm:"column:node_role;type:varchar(50);comment:节点角色" json:"role"`
	IsEnable    bool               `gorm:"column:is_enable;type:boolean;comment:是否启用" json:"is_enable"`
	OesColonyID uint32             `gorm:"column:oes_colony_id;not null;comment:oes集群ID" json:"oes_colony_id"`
	OesColony   OesColonyModel     `gorm:"foreignKey:OesColonyID;references:ID;constraint:OnDelete:CASCADE" json:"oes_colony"`
	HostID      uint32             `gorm:"column:host_id;not null;comment:主机ID" json:"host_id"`
	Host        resource.HostModel `gorm:"foreignKey:HostID;references:ID;constraint:OnDelete:CASCADE" json:"host"`
}

func (m *OesNodeModel) TableName() string {
	return "oes_node"
}

func (m *OesNodeModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("role", m.NodeRole)
	enc.AddBool("is_enable", m.IsEnable)
	enc.AddUint32("oes_colony_id", m.OesColonyID)
	enc.AddUint32("host_id", m.HostID)
	return nil
}

type OesNodeVars struct {
	ID       uint32 `json:"id" yaml:"id"`
	NodeRole string `json:"node_role" yaml:"node_role"`
	Specdir  string `json:"specdir" yaml:"specdir"`
	HostID   uint32 `json:"host_id" yaml:"host_id"`
	IsEnable bool   `json:"is_enable" yaml:"is_enable"`
}

func (vs *OesNodeVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", vs.ID)
	enc.AddString("node_role", vs.NodeRole)
	enc.AddString("specdir", vs.Specdir)
	enc.AddUint32("host_id", vs.HostID)
	enc.AddBool("is_enable", vs.IsEnable)
	return nil
}

// CreateOrUpdateOesNodeRequest 用于创建oes节点的请求结构体
//
// swagger:model CreateOrUpdateOesNodeRequest
type CreateOrUpdateOesNodeRequest struct {
	// 节点角色
	// required: true
	// example: "01"
	NodeRole string `json:"node_role" form:"node_role" binding:"required,oneof=master follow arbiter"`

	// 是否启用
	// required: true
	// example: true
	IsEnable bool `json:"is_enable" form:"is_enable"`

	// oes集群ID
	// required: true
	// example: 1
	OesColonyID uint32 `json:"oes_colony_id" form:"oes_colony_id" binding:"required"`

	// 主机ID
	// required: true
	// example: 1
	HostID uint32 `json:"host_id" form:"host_id" binding:"required"`
}

// ListOesNodeRequest 用于获取oes节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListOesNodeRequest
type ListOesNodeRequest struct {
	common.StandardModelQuery

	// 节点角色
	// example: "master"
	NodeRole string `form:"node_role"`

	// 是否启用
	// required: false
	// example: true
	IsEnable *bool `form:"is_enable"`

	// oes集群ID
	// required: false
	// example: 1
	OesColonyID uint32 `form:"oes_colony_id"`

	// 主机ID
	// required: false
	// example: 1
	HostID uint32 `form:"host_id"`
}

func (req *ListOesNodeRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(12)
	if req.NodeRole != "" {
		query["NodeRole = ?"] = req.NodeRole
	}
	if req.IsEnable != nil {
		query["is_enable = ?"] = *req.IsEnable
	}
	if req.OesColonyID > 0 {
		query["oes_colony_id = ?"] = req.OesColonyID
	}
	if req.HostID > 0 {
		query["host_id = ?"] = req.HostID
	}
	return page, size, query
}

type OesNodeBaseOut struct {
	// ID
	ID uint32 `json:"id" example:"1"`
	// 节点角色
	NodeRole string `json:"node_role" example:"master"`
	// 是否启用
	IsEnable bool `json:"is_enable"`
}

type OesNodeStandardOut struct {
	OesNodeBaseOut
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type OesNodeDetailOut struct {
	OesNodeStandardOut
	OesColony *OesColonyBaseOut     `json:"oes_colony"`
	Host      *resource.HostBaseOut `json:"host"`
}

// OesNodeReply 程序包响应结构
type OesNodeReply = common.APIReply[OesNodeDetailOut]

// PagOesNodeReply 程序包的分页响应结构
type PagOesNodeReply = common.APIReply[*common.Pag[OesNodeDetailOut]]

func OesNodeToBaseOut(
	m OesNodeModel,
) *OesNodeBaseOut {
	return &OesNodeBaseOut{
		ID:       m.ID,
		NodeRole: m.NodeRole,
		IsEnable: m.IsEnable,
	}
}

func OesNodeToStandardOut(
	m OesNodeModel,
) *OesNodeStandardOut {
	return &OesNodeStandardOut{
		OesNodeBaseOut: *OesNodeToBaseOut(m),
		CreatedAt:      m.CreatedAt.Format(time.DateTime),
		UpdatedAt:      m.UpdatedAt.Format(time.DateTime),
	}
}

func OesNodeToDetailOut(
	m OesNodeModel,
) *OesNodeDetailOut {
	return &OesNodeDetailOut{
		OesNodeStandardOut: *OesNodeToStandardOut(m),
		OesColony:          OesColonyToBaseOut(m.OesColony),
		Host:               resource.HostModelToBaseOut(m.Host),
	}
}

func ListOesNodeToDetailOut(
	rms *[]OesNodeModel,
) *[]OesNodeDetailOut {
	if rms == nil {
		return &[]OesNodeDetailOut{}
	}

	ms := *rms
	mso := make([]OesNodeDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := OesNodeToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
