package mon

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
)

type MonNodeModel struct {
	database.StandardModel
	Name        string                 `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	DeployPath  string                 `gorm:"column:deploy_path;type:varchar(255);comment:部署路径" json:"deploy_path"`
	OutportPath string                 `gorm:"column:outport_path;type:varchar(255);comment:导出路径" json:"outport_path"`
	JavaHome    string                 `gorm:"column:java_home;type:varchar(255);comment:JAVA_HOME" json:"java_home"`
	URL         string                 `gorm:"column:url;type:varchar(150);not null;uniqueIndex;comment:URL地址" json:"url"`
	HostID      uint32                 `gorm:"column:host_id;not null;comment:主机ID" json:"host_id"`
	Host        resource.HostModel     `gorm:"foreignKey:HostID;references:ID;constraint:OnDelete:CASCADE" json:"host"`
	PackageID   *uint32                `gorm:"column:package_id;comment:程序包ID" json:"package_id"`
	Package     *resource.PackageModel `gorm:"foreignKey:PackageID;references:ID;constraint:OnDelete:RESTRICT" json:"package"`
	JdkID       *uint32                `gorm:"column:jdk_id;comment:JDK包ID" json:"jdk_id"`
	Jdk         *resource.PackageModel `gorm:"foreignKey:JdkID;references:ID;constraint:OnDelete:RESTRICT" json:"jdk"`
}

func (m *MonNodeModel) TableName() string {
	return "mon_node"
}

func (m *MonNodeModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("deploy_path", m.DeployPath)
	enc.AddString("outport_path", m.OutportPath)
	enc.AddString("java_home", m.JavaHome)
	enc.AddString("url", m.URL)
	enc.AddUint32("host_id", m.HostID)
	if m.PackageID != nil {
		enc.AddUint32("package_id", *m.PackageID)
	}
	if m.JdkID != nil {
		enc.AddUint32("jdk_id", *m.JdkID)
	}
	return nil
}

func ListMonNodeModelToUint32s(ms []MonNodeModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

type MonNodeVars struct {
	ID          uint32 `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	DeployPath  string `json:"slave_path_mon_home" yaml:"slave_path_mon_home"`
	OutportPath string `json:"slave_path_mon_outport" yaml:"slave_path_mon_outport"`
	JavaHome    string `json:"java_home" yaml:"java_home"`
	URL         string `json:"url" yaml:"url"`
	HostID      uint32 `json:"host_id" yaml:"host_id"`
	PackageID   uint32 `json:"package_id" yaml:"package_id"`
	JdkID       uint32 `json:"jdk_id" yaml:"jdk_id"`
	JdkName     string `json:"jdk_name" yaml:"jdk_name"`
}

func (vs *MonNodeVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", vs.ID)
	enc.AddString("name", vs.Name)
	enc.AddString("slave_path_mon_home", vs.DeployPath)
	enc.AddString("slave_path_mon_outport", vs.OutportPath)
	enc.AddString("java_home", vs.JavaHome)
	enc.AddString("url", vs.URL)
	enc.AddUint32("host_id", vs.HostID)
	enc.AddUint32("package_id", vs.PackageID)
	enc.AddUint32("jdk_id", vs.JdkID)
	return nil
}

func MonNodeModelToNodeVars(m MonNodeModel) MonNodeVars {
	var (
		pkgID   uint32 = 0
		jdkID   uint32 = 0
		jdkName string = ""
	)

	if m.PackageID != nil {
		pkgID = *m.PackageID
	}
	if m.JdkID != nil {
		jdkID = *m.JdkID
	}
	if m.Jdk != nil {
		jdkName = m.Jdk.StorageFilename
	}
	return MonNodeVars{
		ID:          m.ID,
		Name:        m.Name,
		DeployPath:  m.DeployPath,
		OutportPath: m.OutportPath,
		JavaHome:    m.JavaHome,
		URL:         m.URL,
		HostID:      m.HostID,
		PackageID:   pkgID,
		JdkID:       jdkID,
		JdkName:     jdkName,
	}
}

// MonNodeUpsertDTO 用于创建mon节点的请求结构体
//
// swagger:model MonNodeUpsertDTO
type MonNodeUpsertDTO struct {
	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 部署路径
	DeployPath string `json:"deploy_path" form:"deploy_path" binding:"required"`

	// 导出路径
	OutportPath string `json:"outport_path" form:"outport_path" binding:"required"`

	// JAVA_HOME
	JavaHome string `json:"java_home" form:"java_home" bunding:"required"`

	// URL地址
	URL string `json:"url" form:"url" bunding:"required"`

	// 主机ID
	HostID uint32 `json:"host_id" form:"host_id" binding:"required"`

	// 程序包ID
	PackageID uint32 `json:"package_id" form:"package_id" binding:"required"`

	// jdk包ID
	JdkID uint32 `json:"jdk_id" form:"jdk_id" binding:"required"`
}

func (dto *MonNodeUpsertDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", dto.Name)
	enc.AddString("deploy_path", dto.DeployPath)
	enc.AddString("outport_path", dto.OutportPath)
	enc.AddString("java_home", dto.JavaHome)
	enc.AddString("url", dto.URL)
	enc.AddUint32("host_id", dto.HostID)
	enc.AddUint32("package_id", dto.PackageID)
	enc.AddUint32("jdk_id", dto.JdkID)
	return nil
}

func (dto *MonNodeUpsertDTO) ToModel() MonNodeModel {
	return MonNodeModel{
		Name:        dto.Name,
		DeployPath:  dto.DeployPath,
		OutportPath: dto.OutportPath,
		JavaHome:    dto.JavaHome,
		URL:         dto.URL,
		HostID:      dto.HostID,
		PackageID:   &dto.PackageID,
		JdkID:       &dto.JdkID,
	}
}

func (dto *MonNodeUpsertDTO) ToUpdateMap() map[string]any {
	return map[string]any{
		"name":         dto.Name,
		"deploy_path":  dto.DeployPath,
		"outport_path": dto.OutportPath,
		"java_home":    dto.JavaHome,
		"url":          dto.URL,
		"host_id":      dto.HostID,
		"package_id":   dto.PackageID,
		"jdk_id":       dto.JdkID,
	}
}

// ListMonNodeDTO 用于获取mon节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMonNodeDTO
type ListMonNodeDTO struct {
	common.StandardModelQuery

	// 名称
	Name string `form:"name"`

	// 主机ID
	HostID uint32 `form:"host_id"`

	// 程序包ID
	PackageID uint32 `form:"package_id"`

	// jdk包ID
	JdkID uint32 `form:"jdk_id"`
}

func (dto *ListMonNodeDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := dto.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", dto.Name)
	enc.AddUint32("host_id", dto.HostID)
	enc.AddUint32("package_id", dto.PackageID)
	enc.AddUint32("jdk_id", dto.JdkID)
	return nil
}

func (dto *ListMonNodeDTO) ToQueryMap() map[string]any {
	queryMap := dto.StandardModelQuery.ToQueryMap(10)
	if dto.Name != "" {
		queryMap["name like ?"] = "%" + dto.Name + "%"
	}
	if dto.HostID > 0 {
		queryMap["host_id = ?"] = dto.HostID
	}
	if dto.PackageID > 0 {
		queryMap["package_id = ?"] = dto.PackageID
	}
	if dto.JdkID > 0 {
		queryMap["jdk_id = ?"] = dto.JdkID
	}
	return queryMap
}

type MonNodeBaseOut struct {
	// 计划任务ID
	ID uint32 `json:"id" example:"1"`

	// 名称
	Name string `json:"name" example:"test"`

	// 部署路径
	DeployPath string `json:"deploy_path" example:""`

	// 导出路径
	OutportPath string `json:"outport_path" example:""`

	// JAVA_HOME
	JavaHome string `json:"java_home" example:""`

	// URL地址
	URL string `json:"url" example:"http://192.168.11.189:8080"`
}

type MonNodeStandardOut struct {
	MonNodeBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type MonNodeDetailOut struct {
	MonNodeStandardOut

	// 主机
	Host *resource.HostBaseOut `json:"host"`

	// 程序包
	Package *resource.PackageStandardOut

	// jdk包
	Jdk *resource.PackageStandardOut
}

// MonNodeResp 程序包响应结构
type MonNodeResp = common.APIResp[MonNodeDetailOut]

// PagMonNodeResp 程序包的分页响应结构
type PagMonNodeResp = common.APIResp[*common.Pag[MonNodeDetailOut]]

func MonNodeToBaseOut(
	m MonNodeModel,
) *MonNodeBaseOut {
	return &MonNodeBaseOut{
		ID:          m.ID,
		Name:        m.Name,
		DeployPath:  m.DeployPath,
		OutportPath: m.OutportPath,
		JavaHome:    m.JavaHome,
		URL:         m.URL,
	}
}

func MonNodeToStandardOut(
	m MonNodeModel,
) *MonNodeStandardOut {
	return &MonNodeStandardOut{
		MonNodeBaseOut: *MonNodeToBaseOut(m),
		CreatedAt:      m.CreatedAt.Format(time.DateTime),
		UpdatedAt:      m.UpdatedAt.Format(time.DateTime),
	}
}

func MonNodeToDetailOut(
	m MonNodeModel,
) *MonNodeDetailOut {
	out := &MonNodeDetailOut{
		MonNodeStandardOut: *MonNodeToStandardOut(m),
		Host:               resource.HostModelToBaseOut(m.Host),
	}
	if m.Package != nil {
		out.Package = resource.PackageModelToBaseOut(*m.Package)
	}
	if m.Jdk != nil {
		out.Jdk = resource.PackageModelToBaseOut(*m.Jdk)
	}
	return out
}

func ListMonNodeToDetailOut(
	ms []MonNodeModel,
) []MonNodeDetailOut {
	if len(ms) == 0 {
		return []MonNodeDetailOut{}
	}
	mso := make([]MonNodeDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := MonNodeToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return mso
}
