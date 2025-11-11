package common

import "net/http"

// APIReply 通用响应结构体
// 用于封装API返回的数据格式
//
// swagger:model APIReply
type APIReply[T any] struct {
	// 状态码
	// Example: 200
	Code int `json:"code"`
	// 信息
	// Example: "success"
	Msg string `json:"msg"`
	// 数据
	// 可以是任意类型的数据
	Data T `json:"data,omitempty"`
}

// Pag 分页响应结构体
// 用于封装分页查询的返回数据格式
//
// swagger:model Pag
type Pag[T any] struct {
	// 当前页码
	// Example: 1
	Page int `json:"page" example:"1"`
	// 每页数量
	// Example: 10
	Size int `json:"size" example:"10"`
	// 总记录数
	// Example: 100
	Total int64 `json:"total" example:"100"`
	// 总页数
	// Example: 10
	Pages int64 `json:"pages" example:"10"`
	// 对象数组
	Items []T `json:"items"`
}

func NewPag[T any](page, size int, total int64, items []T) *Pag[T] {
	var pages int64
	if total == 0 || size <= 0 {
		pages = 0
	} else {
		s := int64(size)
		pages = (total + s - 1) / s
	}
	return &Pag[T]{
		Page:  page,
		Size:  size,
		Total: total,
		Pages: pages,
		Items: items,
	}
}

type MapAPIReply struct {
	// 状态码
	// Example: 200
	Code int `json:"code"`
	// 信息
	// Example: "success"
	Msg string `json:"msg"`
	// 数据
	// 可以是任意类型的数据
	Data map[string]any `json:"data,omitempty"`
}

var NoDataReply = MapAPIReply{
	Code: http.StatusOK,
	Msg:  "",
	Data: map[string]any{},
}
