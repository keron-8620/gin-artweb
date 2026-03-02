package resource

import (
	"mime/multipart"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
)

type PackageModel struct {
	database.BaseModel
	Label           string    `gorm:"column:label;type:varchar(50);uniqueIndex:idx_package_label_version;comment:标签" json:"label"`
	StorageFilename string    `gorm:"column:storage_filename;type:varchar(50);not null;uniqueIndex;comment:磁盘存储文件名" json:"storage_filename"`
	OriginFilename  string    `gorm:"column:origin_filename;type:varchar(255);comment:原始文件名" json:"origin_filename"`
	Version         string    `gorm:"column:version;type:varchar(50);uniqueIndex:idx_package_label_version;comment:版本号" json:"version"`
	UploadedAt      time.Time `gorm:"column:uploaded_at;autoCreateTime;comment:上传时间" json:"uploaded_at"`
}

func (m *PackageModel) TableName() string {
	return "resource_package"
}

func (m *PackageModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.BaseModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("label", m.Label)
	enc.AddString("storage_filename", m.StorageFilename)
	enc.AddString("origin_filename", m.OriginFilename)
	enc.AddString("version", m.Version)
	enc.AddTime("uploaded_at", m.UploadedAt)
	return nil
}

type UploadPackageRequest struct {
	// 标签
	Label string `form:"label" binding:"required,oneof=mds oes xcounter"`

	// 版本号
	Version string `form:"version" binding:"required"`

	// 上传的程序包文件
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func (req *UploadPackageRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("label", req.Label)
	enc.AddString("version", req.Version)
	return nil
}

type ListPackageRequest struct {
	common.BaseModelQuery

	// 文件名
	Filename string `form:"filename" binding:"omitempty,max=50"`

	// 标签
	Label string `form:"label" binding:"omitempty,max=50"`

	// 标签组(多个用,隔开)
	Labels string `form:"labels" binding:"omitempty,max=50"`

	// 版本号
	Version string `form:"version" binding:"omitempty" json:"version"`

	// 上传时间之前的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	BeforeUploadedAt string `form:"before_uploaded_at" binding:"omitempty"`

	// 上传时间之后的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	AfterUploadedAt string `form:"after_uploaded_at" binding:"omitempty"`
}

func (req *ListPackageRequest) Query() (int, int, map[string]any) {
	page, size, query := req.BaseModelQuery.QueryMap(7)
	if req.Filename != "" {
		query["origin_filename like ?"] = "%" + req.Filename + "%"
	}
	if req.Label != "" {
		query["label = ?"] = req.Label
	}
	if req.Labels != "" {
		rawLabels := strings.Split(req.Labels, ",")
		var labels []string
		for _, label := range rawLabels {
			trimmedLabel := strings.TrimSpace(label)
			if trimmedLabel != "" { // 过滤掉空标签
				labels = append(labels, trimmedLabel)
			}
		}
		if len(labels) > 0 {
			query["label in ?"] = labels
		}
	}
	if req.Version != "" {
		query["version like ?"] = "%" + req.Version + "%"
	}
	if req.BeforeUploadedAt != "" {
		bft, err := time.Parse(time.RFC3339, req.BeforeUploadedAt)
		if err == nil {
			query["uploaded_at < ?"] = bft
		}
	}
	if req.AfterUploadedAt != "" {
		act, err := time.Parse(time.RFC3339, req.AfterUploadedAt)
		if err == nil {
			query["uploaded_at > ?"] = act
		}
	}
	return page, size, query
}

// PackageStandardOut 程序包基础信息
type PackageStandardOut struct {
	// 主机ID
	ID uint32 `json:"id" example:"1"`

	// 名称
	Filename string `json:"filename" example:"oes.tar.gz"`

	// 标签
	Label string `json:"label" example:"artweb"`

	// IP地址
	Version string `json:"version" example:"0.17.0.0.1"`

	// 上传时间
	UploadedAt string `json:"uploaded_at" example:"2023-01-01 12:00:00"`
}

// PackageReply 程序包响应结构
type PackageReply = common.APIReply[PackageStandardOut]

// PagPackageReply程序包的分页响应结构
type PagPackageReply = common.APIReply[*common.Pag[PackageStandardOut]]

func PackageModelToOutBase(
	m PackageModel,
) *PackageStandardOut {
	return &PackageStandardOut{
		ID:         m.ID,
		Filename:   m.OriginFilename,
		Label:      m.Label,
		Version:    m.Version,
		UploadedAt: m.UploadedAt.Format(time.DateTime),
	}
}

func ListPkgModelToOut(
	pms *[]PackageModel,
) *[]PackageStandardOut {
	if pms == nil {
		return &[]PackageStandardOut{}
	}

	ms := *pms
	mso := make([]PackageStandardOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := PackageModelToOutBase(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
