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

func ListOesNodeModelToUint32s(ms []OesNodeModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
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

// OesNodeUpsertDTO 用于创建oes节点的请求结构体
//
// swagger:model OesNodeUpsertDTO
type OesNodeUpsertDTO struct {
	// 节点角色
	NodeRole string `json:"node_role" form:"node_role" binding:"required,oneof=master follow arbiter"`

	// 是否启用
	IsEnable bool `json:"is_enable" form:"is_enable"`

	// oes集群ID
	OesColonyID uint32 `json:"oes_colony_id" form:"oes_colony_id" binding:"required"`

	// 主机ID
	HostID uint32 `json:"host_id" form:"host_id" binding:"required"`
}

func (dto OesNodeUpsertDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("node_role", dto.NodeRole)
	enc.AddBool("is_enable", dto.IsEnable)
	enc.AddUint32("oes_colony_id", dto.OesColonyID)
	enc.AddUint32("host_id", dto.HostID)
	return nil
}

func (dto OesNodeUpsertDTO) ToModel() OesNodeModel {
	return OesNodeModel{
		NodeRole:    dto.NodeRole,
		IsEnable:    dto.IsEnable,
		OesColonyID: dto.OesColonyID,
		HostID:      dto.HostID,
	}
}

func (dto OesNodeUpsertDTO) ToUpdateMap() map[string]any {
	return map[string]any{
		"node_role":     dto.NodeRole,
		"is_enable":     dto.IsEnable,
		"oes_colony_id": dto.OesColonyID,
		"host_id":       dto.HostID,
	}
}

// ListOesNodeDTO 用于获取oes节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListOesNodeDTO
type ListOesNodeDTO struct {
	common.StandardModelQuery

	// 节点角色
	NodeRole string `form:"node_role"`

	// 是否启用
	IsEnable *bool `form:"is_enable"`

	// oes集群ID
	OesColonyID uint32 `form:"oes_colony_id"`

	// 主机ID
	HostID uint32 `form:"host_id"`
}

func (dto *ListOesNodeDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := dto.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("node_role", dto.NodeRole)
	if dto.IsEnable != nil {
		enc.AddBool("is_enable", *dto.IsEnable)
	}
	enc.AddUint32("oes_colony_id", dto.OesColonyID)
	enc.AddUint32("host_id", dto.HostID)
	return nil
}

func (dto *ListOesNodeDTO) ToQueryMap() map[string]any {
	queryMap := dto.StandardModelQuery.ToQueryMap(12)
	if dto.NodeRole != "" {
		queryMap["node_role = ?"] = dto.NodeRole
	}
	if dto.IsEnable != nil {
		queryMap["is_enable = ?"] = *dto.IsEnable
	}
	if dto.OesColonyID > 0 {
		queryMap["oes_colony_id = ?"] = dto.OesColonyID
	}
	if dto.HostID > 0 {
		queryMap["host_id = ?"] = dto.HostID
	}
	return queryMap
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

// OesNodeResp 程序包响应结构
type OesNodeResp = common.APIResp[OesNodeDetailOut]

// PagOesNodeResp 程序包的分页响应结构
type PagOesNodeResp = common.APIResp[*common.Pag[OesNodeDetailOut]]

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
	ms []OesNodeModel,
) []OesNodeDetailOut {
	if len(ms) == 0 {
		return []OesNodeDetailOut{}
	}
	mso := make([]OesNodeDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := OesNodeToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return mso
}
