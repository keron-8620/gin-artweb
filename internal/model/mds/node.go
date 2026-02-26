package mds

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
)

type MdsNodeModel struct {
	database.StandardModel
	NodeRole    string             `gorm:"column:node_role;type:varchar(50);comment:节点角色" json:"role"`
	IsEnable    bool               `gorm:"column:is_enable;type:boolean;comment:是否启用" json:"is_enable"`
	MdsColonyID uint32             `gorm:"column:mds_colony_id;not null;comment:mds集群ID" json:"mds_colony_id"`
	MdsColony   MdsColonyModel     `gorm:"foreignKey:MdsColonyID;references:ID;constraint:OnDelete:CASCADE" json:"mds_colony"`
	HostID      uint32             `gorm:"column:host_id;not null;comment:主机ID" json:"host_id"`
	Host        resource.HostModel `gorm:"foreignKey:HostID;references:ID;constraint:OnDelete:CASCADE" json:"host"`
}

func (m *MdsNodeModel) TableName() string {
	return "mds_node"
}

func (m *MdsNodeModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("role", m.NodeRole)
	enc.AddBool("is_enable", m.IsEnable)
	enc.AddUint32("mds_colony_id", m.MdsColonyID)
	enc.AddUint32("host_id", m.HostID)
	return nil
}

type MdsNodeVars struct {
	ID       uint32 `json:"id" yaml:"id"`
	NodeRole string `json:"node_role" yaml:"node_role"`
	Specdir  string `json:"specdir" yaml:"specdir"`
	HostID   uint32 `json:"host_id" yaml:"host_id"`
	IsEnable bool   `json:"is_enable" yaml:"is_enable"`
}

func (vs *MdsNodeVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", vs.ID)
	enc.AddString("node_role", vs.NodeRole)
	enc.AddString("specdir", vs.Specdir)
	enc.AddUint32("host_id", vs.HostID)
	enc.AddBool("is_enable", vs.IsEnable)
	return nil
}

// CreateOrUpdateMdsNodeRequest 用于创建mds节点的请求结构体
//
// swagger:model CreateOrUpdateMdsNodeRequest
type CreateOrUpdateMdsNodeRequest struct {
	// 节点角色
	NodeRole string `json:"node_role" form:"node_role" binding:"required,oneof=master follow arbiter"`

	// 是否启用
	IsEnable bool `json:"is_enable" form:"is_enable"`

	// mds集群ID
	MdsColonyID uint32 `json:"mds_colony_id" form:"mds_colony_id" binding:"required"`

	// 主机ID
	HostID uint32 `json:"host_id" form:"host_id" binding:"required"`
}

// ListMdsNodeRequest 用于获取mds节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMdsNodeRequest
type ListMdsNodeRequest struct {
	common.StandardModelQuery

	// 节点角色
	NodeRole string `form:"node_role"`

	// 是否启用
	IsEnable *bool `form:"is_enable"`

	// mds集群ID
	MdsColonyID uint32 `form:"mds_colony_id"`

	// 主机ID
	HostID uint32 `form:"host_id"`
}

func (req *ListMdsNodeRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(12)
	if req.NodeRole != "" {
		query["NodeRole = ?"] = req.NodeRole
	}
	if req.IsEnable != nil {
		query["is_enable = ?"] = *req.IsEnable
	}
	if req.MdsColonyID > 0 {
		query["mds_colony_id = ?"] = req.MdsColonyID
	}
	if req.HostID > 0 {
		query["host_id = ?"] = req.HostID
	}
	return page, size, query
}

type MdsNodeBaseOut struct {
	// ID
	ID uint32 `json:"id" example:"1"`

	// 节点角色
	NodeRole string `json:"node_role" example:"master"`

	// 是否启用
	IsEnable bool `json:"is_enable"`
}

type MdsNodeStandardOut struct {
	MdsNodeBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type MdsNodeDetailOut struct {
	MdsNodeStandardOut

	// mds集群
	MdsColony *MdsColonyBaseOut `json:"mds_colony"`

	// 主机
	Host *resource.HostBaseOut `json:"host"`
}

// MdsNodeReply mds节点配置的响应结构
type MdsNodeReply = common.APIReply[MdsNodeDetailOut]

// PagMdsNodeReply mds节点配置的分页响应结构
type PagMdsNodeReply = common.APIReply[*common.Pag[MdsNodeDetailOut]]

func MdsNodeToBaseOut(
	m MdsNodeModel,
) *MdsNodeBaseOut {
	return &MdsNodeBaseOut{
		ID:       m.ID,
		NodeRole: m.NodeRole,
		IsEnable: m.IsEnable,
	}
}

func MdsNodeToStandardOut(
	m MdsNodeModel,
) *MdsNodeStandardOut {
	return &MdsNodeStandardOut{
		MdsNodeBaseOut: *MdsNodeToBaseOut(m),
		CreatedAt:      m.CreatedAt.Format(time.DateTime),
		UpdatedAt:      m.UpdatedAt.Format(time.DateTime),
	}
}

func MdsNodeToDetailOut(
	m MdsNodeModel,
) *MdsNodeDetailOut {
	return &MdsNodeDetailOut{
		MdsNodeStandardOut: *MdsNodeToStandardOut(m),
		MdsColony:          MdsColonyToBaseOut(m.MdsColony),
		Host:               resource.HostModelToBaseOut(m.Host),
	}
}

func ListMdsNodeToDetailOut(
	rms *[]MdsNodeModel,
) *[]MdsNodeDetailOut {
	if rms == nil {
		return &[]MdsNodeDetailOut{}
	}

	ms := *rms
	mso := make([]MdsNodeDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := MdsNodeToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
