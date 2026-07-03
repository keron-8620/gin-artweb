package mds

import (
	"mime/multipart"

	"gin-artweb/internal/model/common"
	"gin-artweb/pkg/fileutil"
)

// MdsConfUriDTO mds配置文件路径参数
type MdsConfUriDTO struct {
	// 集群号
	ColonyNum string `uri:"colony_num" binding:"required,max=2"`
}

type UploadMdsConfDTO struct {
	// 文件夹名称(允许任意文件夹,做路径穿透校验)
	DirName string `form:"dir_name" binding:"required,nopathtraversal"`

	// 上传的mds配置文件
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// MdsConfDirQueryDTO mds配置文件目录查询参数
type MdsConfDirQueryDTO struct {
	// 文件夹名称(允许任意文件夹,做路径穿透校验)
	DirName string `form:"dir_name" binding:"required,nopathtraversal"`
}

// MdsConfFileQueryDTO mds配置文件目录与文件名查询参数
type MdsConfFileQueryDTO struct {
	// 文件夹名称(允许任意文件夹,做路径穿透校验)
	DirName string `form:"dir_name" binding:"required,nopathtraversal"`

	// 配置文件名称(做路径穿透校验)
	Filename string `form:"filename" binding:"required,nopathtraversal"`
}

// PagMdsConfResp 配置文件名列表结构
type PagMdsConfResp = common.APIResp[*fileutil.FileInfo]
