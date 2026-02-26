package mds

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/model/mon"
	"gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
)

type MdsColonyModel struct {
	database.StandardModel
	ColonyNum     string                `gorm:"column:colony_num;type:varchar(2);uniqueIndex;comment:集群号" json:"colony_num"`
	ExtractedName string                `gorm:"column:extracted_name;type:varchar(50);comment:解压后名称" json:"extracted_name"`
	IsEnable      bool                  `gorm:"column:is_enable;type:boolean;comment:是否启用" json:"is_enable"`
	PackageID     uint32                `gorm:"column:package_id;comment:程序包ID" json:"package_id"`
	Package       resource.PackageModel `gorm:"foreignKey:PackageID;references:ID;constraint:OnDelete:CASCADE" json:"package"`
	MonNodeID     uint32                `gorm:"column:mon_node_id;not null;comment:mon节点ID" json:"mon_node_id"`
	MonNode       mon.MonNodeModel      `gorm:"foreignKey:MonNodeID;references:ID;constraint:OnDelete:CASCADE" json:"mon_node"`
}

func (m *MdsColonyModel) TableName() string {
	return "mds_colony"
}

func (m *MdsColonyModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("colony_num", m.ColonyNum)
	enc.AddString("extracted_name", m.ExtractedName)
	enc.AddBool("is_enable", m.IsEnable)
	enc.AddUint32("package_id", m.PackageID)
	enc.AddUint32("mon_node_id", m.MonNodeID)
	return nil
}

type MdsColonyVars struct {
	ID        uint32 `json:"id" yaml:"id"`
	ColonyNum string `json:"colony_num" yaml:"colony_num"`
	PkgName   string `json:"pkg_name" yaml:"pkg_name"`
	PackageID uint32 `json:"package_id" yaml:"package_id"`
	Version   string `json:"version" yaml:"version"`
	MonNodeID uint32 `json:"mon_node_id" yaml:"mon_node_id"`
	IsEnable  bool   `json:"is_enable" yaml:"is_enable"`
}

func (vs *MdsColonyVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("mds_colony_id", vs.ID)
	enc.AddString("colony_num", vs.ColonyNum)
	enc.AddString("pkg_name", vs.PkgName)
	enc.AddUint32("package_id", vs.PackageID)
	enc.AddString("version", vs.Version)
	enc.AddUint32("mon_node_id", vs.MonNodeID)
	enc.AddBool("is_enable", vs.IsEnable)
	return nil
}

// CreateOrUpdateMdsColonyRequest 用于创建mon节点的请求结构体
//
// swagger:model CreateOrUpdateMdsColonyRequest
type CreateOrUpdateMdsColonyRequest struct {
	// 集群号
	ColonyNum string `json:"colony_num" form:"colony_num" binding:"required,max=2"`

	// 解压后名称
	ExtractedName string `json:"extracted_name" form:"extracted_name" binding:"required,max=50"`

	// 是否启用
	IsEnable bool `json:"is_enable" form:"is_enable" binding:"required"`

	// 程序包ID
	PackageID uint32 `json:"package_id" form:"package_id" binding:"required"`

	// mon节点ID
	MonNodeID uint32 `json:"mon_node_id" form:"mon_node_id" binding:"required"`
}

// ListMdsColonyRequest 用于获取mon节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMdsColonyRequest
type ListMdsColonyRequest struct {
	common.StandardModelQuery

	// 集群号
	ColonyNum string `form:"colony_num"`

	// 解压后名称
	ExtractedName string `form:"extracted_name"`

	// 是否启用
	IsEnable *bool `form:"is_enable"`

	// 程序包ID
	PackageID uint32 `form:"package_id"`

	// mon节点ID
	MonNodeID uint32 `form:"mon_node_id"`
}

func (req *ListMdsColonyRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(12)
	if req.ColonyNum != "" {
		query["colony_num = ?"] = req.ColonyNum
	}
	if req.ExtractedName != "" {
		query["extracted_name = ?"] = "%" + req.ExtractedName + "%"
	}
	if req.IsEnable != nil {
		query["is_enable = ?"] = *req.IsEnable
	}
	if req.PackageID > 0 {
		query["package_id = ?"] = req.PackageID
	}
	if req.MonNodeID > 0 {
		query["mon_node_id = ?"] = req.MonNodeID
	}
	return page, size, query
}

type MdsColonyBaseOut struct {
	// ID
	ID uint32 `json:"id" example:"1"`

	// 集群号
	ColonyNum string `json:"colony_num" example:"01"`

	// 解压后名称
	ExtractedName string `json:"extracted_name" example:"mds"`

	// 是否启用
	IsEnable bool `json:"is_enable" example:"true"`
}

type MdsColonyStandardOut struct {
	MdsColonyBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type MdsColonyDetailOut struct {
	MdsColonyStandardOut

	// mds程序包
	Package *resource.PackageStandardOut `json:"package"`

	// mon节点
	MonNode *mon.MonNodeBaseOut `json:"mon_node"`
}

// MdsColonyReply mds集群配置的响应结构
type MdsColonyReply = common.APIReply[MdsColonyDetailOut]

// PagMdsColonyReply mds集群配置的分页响应结构
type PagMdsColonyReply = common.APIReply[*common.Pag[MdsColonyDetailOut]]

// mds 任务状态
type MdsColonyTaskInfo struct {
	// 集群号
	ColonyNum string `json:"colony_num" example:"01"`

	// 任务状态
	Tasks []common.TaskInfo `json:"tasks"`
}

// ListMdsTasksInfoReply 多个mds集群的任务状态响应结构
type ListMdsTasksInfoReply = common.APIReply[[]MdsColonyTaskInfo]

func MdsColonyToBaseOut(
	m MdsColonyModel,
) *MdsColonyBaseOut {
	return &MdsColonyBaseOut{
		ID:            m.ID,
		ColonyNum:     m.ColonyNum,
		ExtractedName: m.ExtractedName,
		IsEnable:      m.IsEnable,
	}
}

func MdsColonyToStandardOut(
	m MdsColonyModel,
) *MdsColonyStandardOut {
	return &MdsColonyStandardOut{
		MdsColonyBaseOut: *MdsColonyToBaseOut(m),
		CreatedAt:        m.CreatedAt.Format(time.DateTime),
		UpdatedAt:        m.UpdatedAt.Format(time.DateTime),
	}
}

func MdsColonyToDetailOut(
	m MdsColonyModel,
) *MdsColonyDetailOut {
	return &MdsColonyDetailOut{
		MdsColonyStandardOut: *MdsColonyToStandardOut(m),
		Package:              resource.PackageModelToOutBase(m.Package),
		MonNode:              mon.MonNodeToBaseOut(m.MonNode),
	}
}

func ListMdsColonyToDetailOut(
	rms *[]MdsColonyModel,
) *[]MdsColonyDetailOut {
	if rms == nil {
		return &[]MdsColonyDetailOut{}
	}

	ms := *rms
	mso := make([]MdsColonyDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := MdsColonyToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
