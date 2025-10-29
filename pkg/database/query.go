package database

import (
	"strconv"
	"strings"
	"time"
)

type ModelQuerier interface {
	Query() map[string]any
}

var (
	DefaultPage int = 1
	DefaultSize int = 10
)

type BaseModelQuery struct {
	// 页码必须大于0
	// Minimum: 1
	Page int `form:"page" binding:"omitempty,gt=0"`

	// 每页大小必须大于等于0
	// Minimum: 0
	Size int `form:"size" binding:"omitempty,gte=0"`

	// 权限主键，可选参数，如果提供则必须大于0
	// Minimum: 1
	Pk uint `form:"pk" binding:"omitempty,gt=0"`

	// "权限主键列表，可选参数，多个用,隔开，如1,2,3"
	// Max length: 100
	Pks string `form:"pks" binding:"omitempty,max=100"`
}

func (q *BaseModelQuery) QueryMap(l int) (int, int, map[string]any) {
	var (
		page int = DefaultPage
		size int = DefaultSize
	)
	if q.Page > 1 {
		page = q.Page
	}
	if q.Size > 0 {
		size = q.Size
	}
	query := make(map[string]any, l)
	if q.Pk > 0 {
		query["id = ?"] = q.Pk
	}
	if q.Pks != "" {
		pks := StringToListUint(q.Pks)
		if len(pks) > 1 {
			query["id in ?"] = pks
		}
	}
	return page, size, query
}

type StandardModelQuery struct {
	BaseModelQuery

	// 创建时间之前的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	BeforeCreateAt string `form:"before_create_at" json:"before_create_at"`

	// 创建时间之后的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	AfterCreateAt string `form:"after_create_at" json:"after_create_at"`

	// 更新时间之前的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	BeforeUpdateAt string `form:"before_update_at" json:"before_update_at"`

	// 更新时间之后的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	AfterUpdateAt string `form:"after_update_at" json:"after_update_at"`
}

func (q *StandardModelQuery) QueryMap(l int) (int, int, map[string]any) {
	page, size, query := q.BaseModelQuery.QueryMap(l)
	if q.BeforeCreateAt != "" {
		bft, err := time.Parse(time.RFC3339, q.BeforeCreateAt)
		if err == nil {
			query["created_at < ?"] = bft
		}
	}
	if q.AfterCreateAt != "" {
		act, err := time.Parse(time.RFC3339, q.AfterCreateAt)
		if err == nil {
			query["created_at > ?"] = act
		}
	}
	if q.BeforeUpdateAt != "" {
		but, err := time.Parse(time.RFC3339, q.BeforeUpdateAt)
		if err == nil {
			query["update_at < ?"] = but
		}
	}
	if q.AfterUpdateAt != "" {
		aut, err := time.Parse(time.RFC3339, q.AfterUpdateAt)
		if err == nil {
			query["update_at > ?"] = aut
		}
	}
	return page, size, query
}

func StringToListUint(pks string) []uint {
	pks = strings.TrimSpace(pks)
	if pks == "" {
		return make([]uint, 0)
	}
	pkList := strings.Split(pks, ",")
	var ids []uint
	for _, pk := range pkList {
		pk = strings.TrimSpace(pk)
		if pk == "" {
			continue
		}
		value, err := strconv.ParseUint(pk, 10, 32)
		if err != nil {
			continue
		}
		ids = append(ids, uint(value))
	}
	return ids
}
