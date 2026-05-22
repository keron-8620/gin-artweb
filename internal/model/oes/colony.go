package oes

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/model/job"
	"gin-artweb/internal/model/mon"
	"gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
)

type OesColonyModel struct {
	database.StandardModel
	SystemType    string                `gorm:"column:system_type;type:varchar(20);comment:系统类型" json:"system_type"`
	ColonyNum     string                `gorm:"column:colony_num;type:varchar(2);uniqueIndex;comment:集群号" json:"colony_num"`
	ExtractedName string                `gorm:"column:extracted_name;type:varchar(50);comment:解压后名称" json:"extracted_name"`
	IsEnable      bool                  `gorm:"column:is_enable;type:boolean;comment:是否启用" json:"is_enable"`
	PackageID     uint32                `gorm:"column:package_id;comment:程序包ID" json:"package_id"`
	Package       resource.PackageModel `gorm:"foreignKey:PackageID;references:ID;constraint:OnDelete:RESTRICT" json:"package"`
	XCounterID    uint32                `gorm:"column:xcounter_id;comment:xcounter包ID" json:"xcounter_id"`
	XCounter      resource.PackageModel `gorm:"foreignKey:XCounterID;references:ID;constraint:OnDelete:RESTRICT" json:"xcounter"`
	MonNodeID     uint32                `gorm:"column:mon_node_id;not null;comment:mon节点ID" json:"mon_node_id"`
	MonNode       mon.MonNodeModel      `gorm:"foreignKey:MonNodeID;references:ID;constraint:OnDelete:RESTRICT" json:"mon_node"`
}

func (m *OesColonyModel) TableName() string {
	return "oes_colony"
}

func (m *OesColonyModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("system_type", m.SystemType)
	enc.AddString("colony_num", m.ColonyNum)
	enc.AddString("extracted_name", m.ExtractedName)
	enc.AddBool("is_enable", m.IsEnable)
	enc.AddUint32("package_id", m.PackageID)
	enc.AddUint32("xcounter_id", m.XCounterID)
	enc.AddUint32("mon_node_id", m.MonNodeID)
	return nil
}

func ListOesColonyModelToUint32s(ms []OesColonyModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

type OesColonyVars struct {
	ID         uint32 `json:"id" yaml:"id"`
	SystemType string `json:"system_type" yaml:"system_type"`
	ColonyNum  string `json:"colony_num" yaml:"colony_num"`
	PkgName    string `json:"pkg_name" yaml:"pkg_name"`
	PackageID  uint32 `json:"package_id" yaml:"package_id"`
	Version    string `json:"version" yaml:"version"`
	MonNodeID  uint32 `json:"mon_node_id" yaml:"mon_node_id"`
	IsEnable   bool   `json:"is_enable" yaml:"is_enable"`
}

func (vs *OesColonyVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", vs.ID)
	enc.AddString("colony_num", vs.ColonyNum)
	enc.AddString("pkg_name", vs.PkgName)
	enc.AddUint32("package_id", vs.PackageID)
	enc.AddString("version", vs.Version)
	enc.AddUint32("mon_node_id", vs.MonNodeID)
	enc.AddBool("is_enable", vs.IsEnable)
	return nil
}

// OesColonyUpsertDTO 用于创建mon节点的请求结构体
//
// swagger:model OesColonyUpsertDTO
type OesColonyUpsertDTO struct {
	// 系统类型
	SystemType string `json:"system_type" form:"system_type" binding:"required,oneof=STK CRD OPT"`

	// 集群号
	ColonyNum string `json:"colony_num" form:"colony_num" binding:"required,max=2"`

	// 解压后名称
	ExtractedName string `json:"extracted_name" form:"extracted_name" binding:"required,max=50"`

	// 是否启用
	IsEnable bool `json:"is_enable" form:"is_enable"`

	// 程序包ID
	PackageID uint32 `json:"package_id" form:"package_id" binding:"required"`

	// xcounter包ID
	XCounterID uint32 `json:"xcounter_id" form:"xcounter_id" binding:"required"`

	// mon节点ID
	MonNodeID uint32 `json:"mon_node_id" form:"mon_node_id" binding:"required"`
}

func (dto *OesColonyUpsertDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	enc.AddString("system_type", dto.SystemType)
	enc.AddString("colony_num", dto.ColonyNum)
	enc.AddString("extracted_name", dto.ExtractedName)
	enc.AddBool("is_enable", dto.IsEnable)
	enc.AddUint32("package_id", dto.PackageID)
	enc.AddUint32("xcounter_id", dto.XCounterID)
	enc.AddUint32("mon_node_id", dto.MonNodeID)
	return nil
}

func (dto *OesColonyUpsertDTO) ToModel() OesColonyModel {
	return OesColonyModel{
		SystemType:    dto.SystemType,
		ColonyNum:     dto.ColonyNum,
		ExtractedName: dto.ExtractedName,
		IsEnable:      dto.IsEnable,
		PackageID:     dto.PackageID,
		XCounterID:    dto.XCounterID,
		MonNodeID:     dto.MonNodeID,
	}
}

func (dto *OesColonyUpsertDTO) ToUpdateMap() map[string]any {
	return map[string]any{
		"system_type":    dto.SystemType,
		"colony_num":     dto.ColonyNum,
		"extracted_name": dto.ExtractedName,
		"is_enable":      dto.IsEnable,
		"package_id":     dto.PackageID,
		"xcounter_id":    dto.XCounterID,
		"mon_node_id":    dto.MonNodeID,
	}
}

// ListOesColonyDTO 用于获取mon节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListOesColonyDTO
type ListOesColonyDTO struct {
	common.StandardModelQuery

	// 系统类型
	SystemType string `form:"system_type"`

	// 集群号
	ColonyNum string `form:"colony_num"`

	// 解压后名称
	ExtractedName string `form:"extracted_name"`

	// 是否启用
	IsEnable *bool `form:"is_enable"`

	// 程序包ID
	PackageID uint32 `form:"package_id"`

	// xcounter包ID
	XCounterID uint32 `form:"xcounter_id"`

	// mon节点ID
	MonNodeID uint32 `form:"mon_node_id"`
}

func (dto *ListOesColonyDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := dto.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("system_type", dto.SystemType)
	enc.AddString("colony_num", dto.ColonyNum)
	enc.AddString("extracted_name", dto.ExtractedName)
	if dto.IsEnable != nil {
		enc.AddBool("is_enable", *dto.IsEnable)
	}
	enc.AddUint32("package_id", dto.PackageID)
	enc.AddUint32("xcounter_id", dto.XCounterID)
	enc.AddUint32("mon_node_id", dto.MonNodeID)
	return nil
}

func (dto *ListOesColonyDTO) ToQueryMap() map[string]any {
	queryMap := dto.StandardModelQuery.ToQueryMap(12)
	if dto.SystemType != "" {
		queryMap["system_type = ?"] = dto.SystemType
	}
	if dto.ColonyNum != "" {
		queryMap["colony_num = ?"] = dto.ColonyNum
	}
	if dto.ExtractedName != "" {
		queryMap["extracted_name = ?"] = "%" + dto.ExtractedName + "%"
	}
	if dto.IsEnable != nil {
		queryMap["is_enable = ?"] = *dto.IsEnable
	}
	if dto.PackageID > 0 {
		queryMap["package_id = ?"] = dto.PackageID
	}
	if dto.XCounterID > 0 {
		queryMap["xcounter_id"] = dto.XCounterID
	}
	if dto.MonNodeID > 0 {
		queryMap["mon_node_id = ?"] = dto.MonNodeID
	}
	return queryMap
}

type OesColonyBaseOut struct {
	// ID
	ID uint32 `json:"id" example:"1"`

	// 系统类型
	SystemType string `json:"system_type" example:"STK"`

	// 集群号
	ColonyNum string `json:"colony_num" example:"01"`

	// 解压后名称
	ExtractedName string `json:"extracted_name" example:"oes"`

	// 是否启用
	IsEnable bool `json:"is_enable" example:"true"`
}

type OesColonyStandardOut struct {
	OesColonyBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type OesColonyDetailOut struct {
	OesColonyStandardOut

	// oes程序包
	Package *resource.PackageStandardOut `json:"package"`

	// xcounter程序包
	XCounter *resource.PackageStandardOut `json:"xcounter"`

	// mon节点
	MonNode *mon.MonNodeBaseOut `json:"mon_node"`
}

// OesColonyResp 程序包响应结构
type OesColonyResp = common.APIResp[OesColonyDetailOut]

// PagOesColonyResp 程序包的分页响应结构
type PagOesColonyResp = common.APIResp[*common.Pag[OesColonyDetailOut]]

// oes 任务状态
type OesColonyTaskInfo struct {
	// 集群号
	ColonyNum string `json:"colony_num" example:"01"`

	// 任务状态
	Tasks []job.BizTaskInfo `json:"tasks"`
}

// ListOesTasksInfoResp 多个oes集群的任务状态响应结构
type ListOesTasksInfoResp = common.APIResp[[]OesColonyTaskInfo]

func OesColonyToBaseOut(
	m OesColonyModel,
) *OesColonyBaseOut {
	return &OesColonyBaseOut{
		ID:            m.ID,
		SystemType:    m.SystemType,
		ColonyNum:     m.ColonyNum,
		ExtractedName: m.ExtractedName,
		IsEnable:      m.IsEnable,
	}
}

func OesColonyToStandardOut(
	m OesColonyModel,
) *OesColonyStandardOut {
	return &OesColonyStandardOut{
		OesColonyBaseOut: *OesColonyToBaseOut(m),
		CreatedAt:        m.CreatedAt.Format(time.DateTime),
		UpdatedAt:        m.UpdatedAt.Format(time.DateTime),
	}
}

func OesColonyToDetailOut(
	m OesColonyModel,
) *OesColonyDetailOut {
	return &OesColonyDetailOut{
		OesColonyStandardOut: *OesColonyToStandardOut(m),
		Package:              resource.PackageModelToOutBase(m.Package),
		XCounter:             resource.PackageModelToOutBase(m.XCounter),
		MonNode:              mon.MonNodeToBaseOut(m.MonNode),
	}
}

func ListOesColonyToDetailOut(
	ms []OesColonyModel,
) []OesColonyDetailOut {
	if len(ms) == 0 {
		return []OesColonyDetailOut{}
	}
	mso := make([]OesColonyDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := OesColonyToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return mso
}
