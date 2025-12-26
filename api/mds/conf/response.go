package conf

import (
	"gin-artweb/api/common"
	"gin-artweb/pkg/fileutil"
)

// PagMdsConfReply 配置文件名列表结构
type PagMdsConfReply = common.APIReply[*fileutil.FileInfo]
