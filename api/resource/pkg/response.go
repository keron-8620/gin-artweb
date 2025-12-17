package pkg

import (
	"gin-artweb/api/common"
)

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
