// Package database 提供数据库CRUD操作的通用方法
// 包括事务处理、关联关系更新、增删改查等常用数据库操作
package database

import (
	"context"
	"runtime/debug"
	"strings"

	"emperror.dev/errors"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

// DBPanic 通用 GORM 数据库操作 panic 捕获函数（必须配合 defer 使用）
// 功能：1. 捕获 panic 并转为标准错误 2. 事务回滚（仅事务场景）3. 记录详细日志（含堆栈、SQL 上下文）
// 参数：
//   - ctx：上下文，用于日志记录和 GORM 操作
//   - db：GORM 数据库实例（事务场景传入 tx 实例，非事务场景传入普通 db 实例）
//
// 返回值：
//   - error：panic 时返回封装了「错误信息+堆栈」的标准错误，无 panic 时返回 nil
func DBPanic(ctx context.Context, db *gorm.DB) error {
	if r := recover(); r != nil {
		// 回滚事务，捕获回滚错误
		if db != nil {
			if db.Statement != nil && db.Statement.ConnPool != nil {
				// 回滚事务,并记录报错
				if err := db.Rollback().Error; err != nil && db.Logger != nil {
					db.Logger.Error(ctx, "数据库事务回滚失败", "回滚错误", err)
				}
			}
		}

		// 构建基础错误信息
		baseErr := errors.NewWithDetails("数据库操作发生panic, 已捕获: %v", r)

		// 获取完整堆栈信息
		stackInfo := string(debug.Stack())

		// 记录详细日志
		if db != nil && db.Logger != nil {
			logFields := []any{
				"panic", r,
				"error", baseErr.Error(),
				"stack", stackInfo,
			}
			// 补充SQL上下文
			if db.Statement != nil {
				logFields = append(logFields, "sql", db.Statement.SQL.String())
				logFields = append(logFields, "sql参数", db.Statement.Vars)
			}
			db.Logger.Error(ctx, "数据库操作发生panic", logFields...)
		}

		// 封装错误并向上传递
		return errors.WithStack(baseErr)
	}

	// 无 panic 时返回 nil
	return nil
}

// DBCreate 创建数据库记录
// ctx: 上下文
// db: GORM数据库实例
// model: 目标模型
// value: 要创建的数据
// 返回操作可能产生的错误
func DBCreate(ctx context.Context, db *gorm.DB, model, value any, upmap map[string]any) error {
	// 使用GORM的Create方法创建记录
	if len(upmap) == 0 {
		err := db.WithContext(ctx).Model(model).Create(value).Error
		return errors.WrapIf(err, "创建数据库记录失败")
	}

	// 开启事务处理
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		// 事务开启失败时记录错误日志
		return errors.WrapIf(tx.Error, "数据库事务开启失败")
	}

	// 设置panic处理
	defer DBPanic(ctx, tx)

	// 创建主表数据
	if err := tx.Model(model).Create(value).Error; err != nil {
		tx.Rollback()
		return errors.WrapIf(err, "创建数据库记录失败")
	}

	// 遍历关联关系映射，逐个更新关联字段
	for k, v := range upmap {
		if err := tx.Model(value).Association(k).Append(v); err != nil {
			tx.Rollback()
			return errors.WrapIf(err, "更新关联关系失败")
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.WrapIf(err, "数据库事务提交失败")
	}
	return nil
}

// DBUpdate 更新数据库记录，支持关联关系更新
// ctx: 上下文
// db: GORM数据库实例
// m: 目标模型
// data: 主表更新数据 (可为nil或空map)
// upmap: 关联关系映射 (可为nil或空map)
// conds: 查询条件
// 返回操作可能产生的错误
func DBUpdate(ctx context.Context, db *gorm.DB, m any, data map[string]any, upmap map[string]any, conds ...any) error {
	// 如果没有需要更新的内容，直接返回
	if len(data) == 0 && len(upmap) == 0 {
		return nil
	}

	// 检查是否提供了查询条件
	if len(conds) == 0 {
		return errors.WithStack(gorm.ErrMissingWhereClause)
	}

	// 如果没有关联关系更新，直接执行更新操作（无需事务）
	if len(upmap) == 0 {
		err := db.WithContext(ctx).Model(m).Where(conds[0], conds[1:]...).Updates(data).Error
		return errors.WrapIf(err, "更新数据库记录失败")
	}

	// 开启事务处理（有关联关系更新时必须使用事务）
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return errors.WrapIf(tx.Error, "数据库事务开启失败")
	}

	// 设置panic处理
	defer DBPanic(ctx, tx)

	// 更新主表数据
	if len(data) > 0 {
		if err := tx.Model(m).Where(conds[0], conds[1:]...).Updates(data).Error; err != nil {
			tx.Rollback()
			return errors.WrapIf(err, "更新数据库记录失败")
		}
	}

	// 遍历关联关系映射，逐个更新关联字段
	for k, v := range upmap {
		if err := tx.Model(m).Association(k).Replace(v); err != nil {
			tx.Rollback()
			return errors.WrapIf(err, "更新关联关系失败")
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		// 提交失败时回滚并返回提交错误
		tx.Rollback()
		return errors.WrapIf(err, "数据库事务提交失败")
	}

	return nil
}

// DBDelete 删除数据库记录
// ctx: 上下文
// db: GORM数据库实例
// model: 目标模型
// conds: 查询条件
// 返回操作可能产生的错误
func DBDelete(ctx context.Context, db *gorm.DB, model any, conds ...any) error {
	// 检查是否提供了查询条件
	if len(conds) == 0 {
		return gorm.ErrMissingWhereClause
	}

	// 执行删除操作
	err := db.WithContext(ctx).Delete(model, conds...).Error
	return errors.WrapIf(err, "删除数据库记录失败")
}

// DBGet 查询单条数据库记录，支持预加载关联关系
// ctx: 上下文
// db: GORM数据库实例
// preloads: 需要预加载的关联关系列表
// m: 查询结果存储对象
// conds: 查询条件
// 返回操作可能产生的错误
func DBGet(ctx context.Context, db *gorm.DB, preloads []string, m any, conds ...any) error {
	dbCtx := db.WithContext(ctx)

	// 预加载关联关系
	for _, preload := range preloads {
		dbCtx = dbCtx.Preload(preload)
	}

	// 查询第一条匹配的记录
	err := dbCtx.First(m, conds...).Error
	return errors.WrapIf(err, "查询数据库记录失败")
}

// DBList 查询数据库记录列表，支持分页、排序、条件查询等功能
// ctx: 上下文
// db: GORM数据库实例
// model: 目标模型
// value: 查询结果存储对象
// query: 查询参数
// 返回记录总数和操作可能产生的错误
func DBList(ctx context.Context, db *gorm.DB, model, value any, query QueryParams) (int64, error) {
	// 初始化查询构建器
	mdb := db.WithContext(ctx).Model(model)

	// 添加查询条件
	for k, v := range query.Query {
		mdb = mdb.Where(k, v)
	}

	// 查询总数
	var count int64 = 0
	if query.IsCount {
		if err := mdb.Count(&count).Error; err != nil {
			return 0, errors.WrapIf(err, "查询数据库记录总数失败")
		}
	}

	// 指定查询字段和忽略字段（如果有同时指定了Columns和Omit，Columns优先）
	if len(query.Columns) > 0 {
		mdb = mdb.Select(query.Columns)
	} else {
		if len(query.Omit) > 0 {
			mdb = mdb.Omit(query.Omit...)
		}
	}

	// 添加排序条件
	orderByStr := strings.Join(query.OrderBy, ",")
	if orderByStr != "" {
		mdb = mdb.Order(orderByStr)
	}

	// 添加分页条件
	if query.Size > 0 {
		mdb = mdb.Limit(query.Size)
		if query.Page > 0 {
			offset := (query.Page - 1) * query.Size
			mdb = mdb.Offset(offset)
		}
	}

	// 预加载关联关系
	for _, preload := range query.Preloads {
		mdb = mdb.Preload(preload)
	}

	// 执行查询
	result := mdb.Find(value)
	if result.Error != nil {
		return 0, errors.WrapIf(result.Error, "查询数据库记录失败")
	}

	// 如果没有查询总数，则使用影响行数作为总数
	if !query.IsCount {
		count = result.RowsAffected
	}

	return count, nil
}

// zap日志中数据库相关常用key
const (
	UpdateDataKey  = "data"         // 更新数据字段
	ConditionsKey  = "conds"        // 查询条件参数
	PreloadKey     = "preloads"     // 预加载关联关系
	ModelKey       = "model"        // 数据模型
	QueryParamsKey = "query_params" // 查询参数
)

// QueryParams 查询参数结构体，用于配置列表查询的各种参数
type QueryParams struct {
	Preloads []string       // 需要预加载的关联关系列表
	Query    map[string]any // 查询条件映射
	OrderBy  []string       // 排序字段列表
	Size     int            // 分页大小
	Page     int            // 分页页码
	IsCount  bool           // 是否查询总数
	Omit     []string       // 需要忽略的字段列表
	Columns  []string       // 查询字段列表
}

func (q *QueryParams) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	// 记录预加载字段
	if len(q.Preloads) > 0 {
		enc.AddString(PreloadKey, strings.Join(q.Preloads, ","))
	} else {
		enc.AddString(PreloadKey, "")
	}

	// 记录查询条件
	enc.AddReflected("query", q.Query)

	// 记录排序字段
	if len(q.OrderBy) > 0 {
		enc.AddString("order_by", strings.Join(q.OrderBy, ","))
	} else {
		enc.AddString("order_by", "")
	}

	// 记录分页参数
	enc.AddInt("size", q.Size)
	enc.AddInt("page", q.Page)

	// 记录是否查询总数
	enc.AddBool("is_count", q.IsCount)

	// 忽略字段
	if len(q.Omit) > 0 {
		enc.AddString("omit", strings.Join(q.Omit, ","))
	} else {
		enc.AddString("omit", "")
	}

	// 查询字段
	if len(q.Columns) > 0 {
		enc.AddString("columns", strings.Join(q.Columns, ","))
	} else {
		enc.AddString("columns", "")
	}
	return nil
}
