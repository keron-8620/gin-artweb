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
	Name        string             `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	DeployPath  string             `gorm:"column:deploy_path;type:varchar(255);comment:部署路径" json:"deploy_path"`
	OutportPath string             `gorm:"column:outport_path;type:varchar(255);comment:导出路径" json:"outport_path"`
	JavaHome    string             `gorm:"column:java_home;type:varchar(255);comment:JAVA_HOME" json:"java_home"`
	URL         string             `gorm:"column:url;type:varchar(150);not null;uniqueIndex;comment:URL地址" json:"url"`
	HostID      uint32             `gorm:"column:host_id;not null;comment:主机ID" json:"host_id"`
	Host        resource.HostModel `gorm:"foreignKey:HostID;references:ID;constraint:OnDelete:CASCADE" json:"host"`
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
	return nil
}

type MonNodeVars struct {
	ID          uint32 `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	DeployPath  string `json:"slave_path_mon_home" yaml:"slave_path_mon_home"`
	OutportPath string `json:"slave_path_mon_outport" yaml:"slave_path_mon_outport"`
	JavaHome    string `json:"java_home" yaml:"java_home"`
	URL         string `json:"url" yaml:"url"`
	HostID      uint32 `json:"host_id" yaml:"host_id"`
}

func (vs *MonNodeVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", vs.ID)
	enc.AddString("name", vs.Name)
	enc.AddString("slave_path_mon_home", vs.DeployPath)
	enc.AddString("slave_path_mon_outport", vs.OutportPath)
	enc.AddString("java_home", vs.JavaHome)
	enc.AddString("url", vs.URL)
	enc.AddUint32("host_id", vs.HostID)
	return nil
}

// CreateOrUpdateMonNodeRequest 用于创建mon节点的请求结构体
//
// swagger:model CreateOrUpdateMonNodeRequest
type CreateOrUpdateMonNodeRequest struct {
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
}

// ListMonNodeRequest 用于获取mon节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMonNodeRequest
type ListMonNodeRequest struct {
	common.StandardModelQuery

	// 名称
	Name string `form:"name"`

	// 主机ID
	HostID uint32 `form:"host_id"`
}

func (req *ListMonNodeRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(10)
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.HostID > 0 {
		query["host_id = ?"] = req.HostID
	}
	return page, size, query
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
}

// MonNodeReply 程序包响应结构
type MonNodeReply = common.APIReply[MonNodeDetailOut]

// PagMonNodeReply 程序包的分页响应结构
type PagMonNodeReply = common.APIReply[*common.Pag[MonNodeDetailOut]]

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
	return &MonNodeDetailOut{
		MonNodeStandardOut: *MonNodeToStandardOut(m),
		Host:               resource.HostModelToBaseOut(m.Host),
	}
}

func ListMonNodeToDetailOut(
	rms *[]MonNodeModel,
) *[]MonNodeDetailOut {
	if rms == nil {
		return &[]MonNodeDetailOut{}
	}

	ms := *rms
	mso := make([]MonNodeDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := MonNodeToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
