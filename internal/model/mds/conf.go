package mds

import (
	"mime/multipart"

	"gin-artweb/internal/model/common"
	"gin-artweb/pkg/fileutil"
)

type UploadMdsConfDTO struct {
	// 上传的mds配置文件
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type DownloadOrDeleteMdsConfRequest struct {
	// 集群号
	ColonyNum string `uri:"colony_num" binding:"required,max=2"`

	// 文件夹名称(all, host_01, host_02, host_03)
	DirName string `uri:"dir_name" binding:"required,oneof=all host_01 host_02 host_03"`

	// 配置文件名称
	Filename string `uri:"filename" form:"filename" binding:"required"`
}

type GetMdsConfDTO struct {
	// 集群号
	ColonyNum string `uri:"colony_num" binding:"required,max=2"`

	// 文件夹名称
	DirName string `uri:"dir_name" binding:"required,oneof=all host_01 host_02 host_03"`
}

type ListMdsConfDTO struct {
	// 集群号
	ColonyNum string `uri:"colony_num" binding:"required,max=2"`
}

// PagMdsConfResp 配置文件名列表结构
type PagMdsConfResp = common.APIResp[*fileutil.FileInfo]
