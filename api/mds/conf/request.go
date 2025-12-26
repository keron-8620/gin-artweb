package conf

import "mime/multipart"

type UploadMdsConfRequest struct {
	// 上传的mds配置文件
	File *multipart.FileHeader `form:"file" binding:"required"`

	// 文件夹名称
	// required: true
	// example: "all"
	DirName string `json:"dir_name" form:"dir_name" binding:"required,oneof=all host_01 host_02 host_03"`

	// mds集群ID
	// required: true
	// example: 1
	MdsColonyID uint32 `json:"mds_colony_id" form:"mds_colony_id" binding:"required"`
}

type DownloadOrDeleteMdsConfRequest struct {
	// mds集群ID
	// required: true
	// example: 1
	MdsColonyID uint32 `json:"mds_colony_id" form:"mds_colony_id" binding:"required"`

	// 文件夹名称
	// required: true
	// example: "all"
	DirName string `json:"dir_name" form:"dir_name" binding:"required,oneof=all host_01 host_02 host_03"`

	// 配置文件名称
	// required: true
	// example: "mds.conf"
	Filename string `json:"filename" form:"filename" binding:"required"`
}

type ListMdsConfRequest struct {
	// mds集群ID
	// required: true
	// example: 1
	MdsColonyID uint32 `json:"mds_colony_id" form:"mds_colony_id" binding:"required"`

	// 文件夹名称
	// required: true
	// example: "all"
	DirName string `json:"dir_name" form:"dir_name" binding:"required,oneof=all host_01 host_02 host_03"`
}
