package record

import "gitee.com/keion8620/go-dango-gin/pkg/common"

// LoginRecordOutBase登陆记录信息
type LoginRecordOutBase struct {
	// 用户ID
	Id uint `json:"id" example:"1"`
	// 名称
	Username string `json:"username" example:"judgement"`
	// 登录时间
	LoginAt string `json:"login_at" example:"2023-01-01 12:00:00"`
	// IP地址
	IPAddress string `json:"ip_address" example:"192.168.1.1"`
	// 用户浏览器信息
	UserAgent string `json:"user_agent" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"`
	// 登录状态
	Status bool `json:"is_active" example:"true"`
}

// PagUserReply 用户的分页响应结构
type PagLoginRecordReply = common.APIReply[*common.Pag[*LoginRecordOutBase]]
