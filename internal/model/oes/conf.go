package oes

import (
	"mime/multipart"

	"gin-artweb/internal/model/common"
	"gin-artweb/pkg/fileutil"
)

// OesConfUriDTO oes配置文件路径参数
type OesConfUriDTO struct {
	// 集群号
	ColonyNum string `uri:"colony_num" binding:"required,max=2"`
}

type UploadOesConfDto struct {
	// 文件夹名称
	DirName string `form:"dir_name" binding:"required"`

	// 上传的oes配置文件
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// OesConfDirQueryDTO oes配置文件目录查询参数
type OesConfDirQueryDTO struct {
	// 文件夹名称
	DirName string `form:"dir_name" binding:"required"`
}

// OesConfFileQueryDTO oes配置文件目录与文件名查询参数
type OesConfFileQueryDTO struct {
	// 文件夹名称
	DirName string `form:"dir_name" binding:"required"`

	// 配置文件名称
	Filename string `form:"filename" binding:"required"`
}

// PagOesConfResp 配置文件名列表结构
type PagOesConfResp = common.APIResp[*fileutil.FileInfo]
