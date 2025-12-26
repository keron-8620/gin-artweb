package conf

import (
	"gin-artweb/api/common"
	"gin-artweb/pkg/fileutil"
)

// PagOesConfReply 配置文件名列表结构
type PagOesConfReply = common.APIReply[*fileutil.FileInfo]
