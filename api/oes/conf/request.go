package conf

import "mime/multipart"

type UploadOesConfRequest struct {
	// 上传的oes配置文件
	File *multipart.FileHeader `form:"file" binding:"required"`

	// 文件夹名称
	// required: true
	// example: "all"
	DirName string `json:"dir_name" form:"dir_name" binding:"required,oneof=all host_01 host_02 host_03"`

	// oes集群ID
	// required: true
	// example: 1
	OesColonyID uint32 `json:"oes_colony_id" form:"oes_colony_id" binding:"required"`
}

type DownloadOrDeleteOesConfRequest struct {
	// oes集群ID
	// required: true
	// example: 1
	OesColonyID uint32 `json:"oes_colony_id" form:"oes_colony_id" binding:"required"`

	// 文件夹名称
	// required: true
	// example: "all"
	DirName string `json:"dir_name" form:"dir_name" binding:"required,oneof=all host_01 host_02 host_03"`

	// 配置文件名称
	// required: true
	// example: "oes.conf"
	Filename string `json:"filename" form:"filename" binding:"required"`
}

type ListOesConfRequest struct {
	// oes集群ID
	// required: true
	// example: 1
	OesColonyID uint32 `json:"oes_colony_id" form:"oes_colony_id" binding:"required"`

	// 文件夹名称
	// required: true
	// example: "all"
	DirName string `json:"dir_name" form:"dir_name" binding:"required,oneof=all host_01 host_02 host_03"`
}
