package conf

import "mime/multipart"

type UploadOesConfRequest struct {
	// 上传的oes配置文件
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type DownloadOrDeleteOesConfRequest struct {
	// 集群号
	ColonyNum string `uri:"colony_num" binding:"required,max=2"`

	// 文件夹名称
	DirName string `uri:"dir_name" binding:"required,oneof=all host_01 host_02 host_03"`

	// 配置文件名称
	Filename string `uri:"filename" form:"filename" binding:"required"`
}

type GetOesConfRequest struct {
	// 集群号
	ColonyNum string `uri:"colony_num" binding:"required,max=2"`

	// 文件夹名称
	DirName string `uri:"dir_name" binding:"required,oneof=all host_01 host_02 host_03"`
}

type ListOesConfRequest struct {
	// 集群号
	ColonyNum string `uri:"colony_num" binding:"required,max=2"`
}
