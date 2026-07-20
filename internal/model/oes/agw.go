package oes

import (
	"mime/multipart"
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
	"gin-artweb/pkg/fileutil"
)

type OesAgwModel struct {
	database.StandardModel
	Name       string                `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	DeployPath string                `gorm:"column:deploy_path;type:varchar(255);comment:部署路径" json:"deploy_path"`
	HostID     uint32                `gorm:"column:host_id;not null;comment:主机ID" json:"host_id"`
	Host       resource.HostModel    `gorm:"foreignKey:HostID;references:ID;constraint:OnDelete:CASCADE" json:"host"`
	PackageID  uint32                `gorm:"column:package_id;comment:程序包ID" json:"package_id"`
	Package    resource.PackageModel `gorm:"foreignKey:PackageID;references:ID;constraint:OnDelete:RESTRICT" json:"package"`
}

func (m *OesAgwModel) TableName() string {
	return "oes_agw"
}

func (m *OesAgwModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("deploy_path", m.DeployPath)
	enc.AddUint32("host_id", m.HostID)
	enc.AddUint32("package_id", m.PackageID)
	return nil
}

func ListAgwModelToUint32s(ms []OesAgwModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

type OesAgwVars struct {
	ID         uint32 `json:"id" yaml:"id"`
	Name       string `json:"name" yaml:"name"`
	DeployPath string `json:"slave_path_agw_home" yaml:"slave_path_agw_home"`
	HostID     uint32 `json:"host_id" yaml:"host_id"`
	PackageID  uint32 `json:"package_id" yaml:"package_id"`
}

func (vs *OesAgwVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", vs.ID)
	enc.AddString("name", vs.Name)
	enc.AddString("slave_path_agw_home", vs.DeployPath)
	enc.AddUint32("host_id", vs.HostID)
	enc.AddUint32("package_id", vs.PackageID)
	return nil
}

func AgwModelToAgwVars(m OesAgwModel) OesAgwVars {
	return OesAgwVars{
		ID:         m.ID,
		Name:       m.Name,
		DeployPath: m.DeployPath,
		HostID:     m.HostID,
		PackageID:  m.PackageID,
	}
}

type AgwUpsertDTO struct {
	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 部署路径
	DeployPath string `json:"deploy_path" form:"deploy_path" binding:"required"`

	// 主机ID
	HostID uint32 `json:"host_id" form:"host_id" binding:"required"`

	// 程序包ID
	PackageID uint32 `json:"package_id" form:"package_id" binding:"required"`
}

func (dto *AgwUpsertDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", dto.Name)
	enc.AddString("deploy_path", dto.DeployPath)
	enc.AddUint32("host_id", dto.HostID)
	enc.AddUint32("package_id", dto.PackageID)
	return nil
}

func (dto *AgwUpsertDTO) ToModel() OesAgwModel {
	return OesAgwModel{
		Name:       dto.Name,
		DeployPath: dto.DeployPath,
		HostID:     dto.HostID,
		PackageID:  dto.PackageID,
	}
}

func (dto *AgwUpsertDTO) ToUpdateMap() map[string]any {
	return map[string]any{
		"name":        dto.Name,
		"deploy_path": dto.DeployPath,
		"host_id":     dto.HostID,
		"package_id":  dto.PackageID,
	}
}

type ListAgwDTO struct {
	common.StandardModelQuery

	// 名称
	Name string `form:"name"`

	// 主机ID
	HostID uint32 `form:"host_id"`

	// 程序包ID
	PackageID uint32 `form:"package_id"`
}

func (dto *ListAgwDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := dto.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", dto.Name)
	enc.AddUint32("host_id", dto.HostID)
	enc.AddUint32("package_id", dto.PackageID)
	return nil
}

func (dto *ListAgwDTO) ToQueryMap() map[string]any {
	queryMap := dto.StandardModelQuery.ToQueryMap(11)
	if dto.Name != "" {
		queryMap["name like ?"] = "%" + dto.Name + "%"
	}
	if dto.HostID > 0 {
		queryMap["host_id = ?"] = dto.HostID
	}
	if dto.PackageID > 0 {
		queryMap["package_id = ?"] = dto.PackageID
	}
	return queryMap
}

type AgwBaseOut struct {
	// 计划任务ID
	ID uint32 `json:"id" example:"1"`

	// 名称
	Name string `json:"name" example:"test"`

	// 部署路径
	DeployPath string `json:"deploy_path" example:""`
}

type AgwStandardOut struct {
	AgwBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type AgwDetailOut struct {
	AgwStandardOut

	// 主机
	Host *resource.HostBaseOut `json:"host"`

	// 程序包
	Package *resource.PackageStandardOut `json:"package"`
}

// AgwResp 程序包响应结构
type AgwResp = common.APIResp[AgwDetailOut]

// PagAgwResp 程序包的分页响应结构
type PagAgwResp = common.APIResp[*common.Pag[AgwDetailOut]]

func AgwToBaseOut(
	m OesAgwModel,
) *AgwBaseOut {
	return &AgwBaseOut{
		ID:         m.ID,
		Name:       m.Name,
		DeployPath: m.DeployPath,
	}
}

func AgwToStandardOut(
	m OesAgwModel,
) *AgwStandardOut {
	return &AgwStandardOut{
		AgwBaseOut: *AgwToBaseOut(m),
		CreatedAt:  m.CreatedAt.Format(time.DateTime),
		UpdatedAt:  m.UpdatedAt.Format(time.DateTime),
	}
}

func AgwToDetailOut(
	m OesAgwModel,
) *AgwDetailOut {
	return &AgwDetailOut{
		AgwStandardOut: *AgwToStandardOut(m),
		Host:           resource.HostModelToBaseOut(m.Host),
		Package:        resource.PackageModelToBaseOut(m.Package),
	}
}

func ListAgwToDetailOut(
	ms []OesAgwModel,
) []AgwDetailOut {
	if len(ms) == 0 {
		return []AgwDetailOut{}
	}
	mso := make([]AgwDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := AgwToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return mso
}

type UploadAgwConfDto struct {
	// 上传的agw配置文件
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// AgwConfFileQueryDTO agw配置文件名查询参数
type AgwConfFileQueryDTO struct {
	// 配置文件名称
	Filename string `form:"filename" binding:"required"`
}

// PagAgwConfResp 配置文件名列表结构
type PagAgwConfResp = common.APIResp[*fileutil.FileInfo]
