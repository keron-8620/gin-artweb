package pkg

import (
	"mime/multipart"
	"time"

	"gin-artweb/api/common"

	"go.uber.org/zap/zapcore"
)

type UploadPackageRequest struct {
	// 标签，最大长度50
	// Required: true
	// Max length: 50
	Label string `form:"label" binding:"required"`

	// 版本号，最大长度50
	// Required: true
	// Max length: 50
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

	// 名称，最大长度50
	// Required: true
	// Max length: 50
	Name string `form:"name" binding:"omitempty,max=50"`

	// 标签，最大长度50
	// Required: true
	// Max length: 50
	Label string `form:"label" binding:"omitempty,max=50"`

	// 版本号，最大长度50
	// Required: true
	// Max length: 50
	Version string `form:"version" binding:"omitempty" json:"version"`

	// 上传时间之前的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	BeforeUploadedAt string `form:"before_uploaded_at" binding:"omitempty"`

	// 上传时间之后的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	AfterUploadedAt string `form:"after_uploaded_at" binding:"omitempty"`
}

func (req *ListPackageRequest) Query() (int, int, map[string]any) {
	page, size, query := req.BaseModelQuery.QueryMap(13)
	if req.Label != "" {
		query["label = ?"] = req.Label
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
