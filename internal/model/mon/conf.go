package mon

import (
	"mime/multipart"

	"gin-artweb/internal/model/common"
	"gin-artweb/pkg/fileutil"
)

// UploadMonConfDTO 上传mon配置文件请求。
type UploadMonConfDTO struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// MonConfFileQueryDTO mon配置文件查询参数。
type MonConfFileQueryDTO struct {
	Filename string `form:"filename" binding:"required"`
}

// PagMonConfResp mon配置文件列表响应。
type PagMonConfResp = common.APIResp[*fileutil.FileInfo]
