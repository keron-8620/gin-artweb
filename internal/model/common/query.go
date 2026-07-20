package common

import (
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/common"
)

type IDUri struct {
	// 唯一标识
	ID uint32 `uri:"id" binding:"required,gt=0"`
}

type ModelQuerier interface {
	Query() map[string]any
}

type BaseModelQuery struct {
	// 分页页码
	Page int `form:"page" binding:"omitempty,gte=1"`

	// 分页大小
	Size int `form:"size" binding:"omitempty,gte=1"`

	// 唯一标识
	ID uint32 `form:"id" binding:"omitempty,gt=0"`

	// "唯一标识列表(多个用,隔开)"
	IDs string `form:"ids" binding:"omitempty,max=100"`
}

func (dto *BaseModelQuery) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	enc.AddInt("page", dto.Page)
	enc.AddInt("size", dto.Size)
	enc.AddUint32("id", dto.ID)
	enc.AddString("ids", dto.IDs)
	return nil
}

func (q *BaseModelQuery) GetPageParam() (int, int) {
	page := q.Page
	if page < 1 {
		page = common.DefaultPage
	}
	size := q.Size
	if size < 1 {
		size = common.DefaultSize
	}
	if size > common.MaxSize {
		size = common.MaxSize
	}
	return page, size
}

func (q *BaseModelQuery) ToQueryMap(l int) map[string]any {
	queryMap := make(map[string]any, l)
	if q.ID > 0 {
		queryMap["id = ?"] = q.ID
	}
	if q.IDs != "" {
		pks := stringToListUint32(q.IDs)
		if len(pks) > 1 {
			queryMap["id in ?"] = pks
		}
	}
	return queryMap
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

func (dto *StandardModelQuery) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	if err := dto.BaseModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("before_created_at", dto.BeforeCreateAt)
	enc.AddString("after_created_at", dto.AfterCreateAt)
	enc.AddString("before_updated_at", dto.BeforeUpdateAt)
	enc.AddString("after_updated_at", dto.AfterUpdateAt)
	return nil
}

func (q *StandardModelQuery) ToQueryMap(l int) map[string]any {
	query := q.BaseModelQuery.ToQueryMap(l)
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
			query["updated_at < ?"] = but
		}
	}
	if q.AfterUpdateAt != "" {
		aut, err := time.Parse(time.RFC3339, q.AfterUpdateAt)
		if err == nil {
			query["updated_at > ?"] = aut
		}
	}
	return query
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
