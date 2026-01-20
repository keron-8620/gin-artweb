package common

import (
	"strconv"
	"strings"
	"time"
)

type IDUri struct {
	// 唯一标识
	ID uint32 `uri:"id" binding:"required,gt=0"`
}

type ModelQuerier interface {
	Query() map[string]any
}

var (
	DefaultPage int = 0
	DefaultSize int = 10
)

type BaseModelQuery struct {
	// 分页页码
	Page int `form:"page" binding:"omitempty,gt=0"`

	// 分页大小
	Size int `form:"size" binding:"omitempty,gte=0"`

	// 唯一标识
	Pk uint32 `form:"pk" binding:"omitempty,gt=0"`

	// "唯一标识列表(多个用,隔开)"
	Pks string `form:"pks" binding:"omitempty,max=100"`
}

func (q *BaseModelQuery) QueryMap(l int) (int, int, map[string]any) {
	var (
		page int = DefaultPage
		size int = DefaultSize
	)
	if q.Page > 1 {
		page = q.Page - 1
	}
	if q.Size > 0 {
		size = q.Size
	}
	query := make(map[string]any, l)
	if q.Pk > 0 {
		query["id = ?"] = q.Pk
	}
	if q.Pks != "" {
		pks := stringToListUint32(q.Pks)
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
	BeforeCreateAt string `form:"before_created_at"`

	// 创建时间之后的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	AfterCreateAt string `form:"after_created_at"`

	// 更新时间之前的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	BeforeUpdateAt string `form:"before_updated_at"`

	// 更新时间之后的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	AfterUpdateAt string `form:"after_updated_at"`
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

func stringToListUint32(pks string) []uint32 {
	pks = strings.TrimSpace(pks)
	if pks == "" {
		return make([]uint32, 0)
	}
	pkList := strings.Split(pks, ",")
	var ids []uint32
	for _, pk := range pkList {
		pk = strings.TrimSpace(pk)
		if pk == "" {
			continue
		}
		value, err := strconv.ParseUint(pk, 10, 32)
		if err != nil {
			continue
		}
		ids = append(ids, uint32(value))
	}
	return ids
}
