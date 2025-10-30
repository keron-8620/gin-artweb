// Package database 提供数据库CRUD操作的通用方法
// 包括事务处理、关联关系更新、增删改查等常用数据库操作
package database

import (
	"context"
	"errors"
	"runtime"
	"strings"

	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

// DBPanic 处理数据库操作中的panic异常，自动回滚事务并记录错误日志
// ctx: 上下文
// tx: GORM事务对象
// 返回panic错误信息
func DBPanic(ctx context.Context, tx *gorm.DB) (err error) {
	defer func() {
		// 捕获panic异常
		if r := recover(); r != nil {
			// 发生panic时回滚事务
			tx.Rollback()

			// 获取调用栈信息
			buf := make([]byte, 64<<10)
			n := runtime.Stack(buf, false)
			buf = buf[:n]

			// 记录错误日志
			errMsg := "a panic error occurred during database operation"
			tx.Logger.Error(ctx, errMsg, r, buf)
			err = errors.New(errMsg)
		}
	}()
	return
}

// DBRollback 回滚数据库事务并记录错误日志
// ctx: 上下文
// tx: GORM事务对象
// 返回回滚操作可能产生的错误
func DBRollback(ctx context.Context, tx *gorm.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // 返回取消原因
	default:
	}
	// 执行回滚操作
	if err := tx.Rollback().Error; err != nil {
		// 回滚失败时记录错误日志
		tx.Logger.Error(ctx, "database rollback error", err)
		return err
	}
	return nil
}

// DBCommit 提交数据库事务，如果提交失败则自动回滚
// ctx: 上下文
// tx: GORM事务对象
// 返回提交操作可能产生的错误
func DBCommit(ctx context.Context, tx *gorm.DB) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // 返回取消原因
	default:
	}
	// 执行提交操作
	if err := tx.Commit().Error; err != nil {
		// 提交失败时记录错误日志并回滚事务
		tx.Logger.Error(ctx, "database commit error", err)
		DBRollback(ctx, tx)
		return err
	}
	return nil
}

// DBAssociate 更新模型的关联关系
// ctx: 上下文
// db: GORM数据库实例
// om: 目标模型对象
// upmap: 关联关系映射，key为关联字段名，value为关联数据
// 返回操作可能产生的错误
func DBAssociate(ctx context.Context, db *gorm.DB, om any, upmap map[string]any) error {
	// 遍历关联关系映射，逐个更新关联字段
	for k, v := range upmap {
		select {
		case <-ctx.Done():
			return ctx.Err() // 返回取消原因
		default:
		}
		// 使用Replace方法替换关联关系
		if err := db.Model(om).Association(k).Replace(v); err != nil {
			return err
		}
	}
	return nil
}

// DBCreate 创建数据库记录
// ctx: 上下文
// db: GORM数据库实例
// model: 目标模型
// value: 要创建的数据
// 返回操作可能产生的错误
func DBCreate(ctx context.Context, db *gorm.DB, model, value any) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // 返回取消原因
	default:
	}
	// 使用GORM的Create方法创建记录
	return db.Model(model).Create(value).Error
}

// DBUpdate 更新数据库记录，支持关联关系更新
// ctx: 上下文
// db: GORM数据库实例
// im: 要更新的数据
// om: 目标模型对象
// upmap: 关联关系映射
// conds: 查询条件
// 返回操作可能产生的错误
func DBUpdate(ctx context.Context, db *gorm.DB, m any, data map[string]any, upmap map[string]any, conds ...any) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // 返回取消原因
	default:
	}
	// 检查是否提供了查询条件
	if len(conds) == 0 {
		return gorm.ErrMissingWhereClause
	}

	// 如果没有关联关系更新，直接执行更新操作
	if len(upmap) == 0 {
		return db.Model(m).Where(conds[0], conds[1:]...).Updates(data).Error
	}

	// 开启事务处理
	tx := db.Begin()
	if tx.Error != nil {
		// 事务开启失败时记录错误日志
		tx.Logger.Error(ctx, "faild to begin transaction for database", tx.Error)
		return tx.Error
	}

	// 设置panic处理
	defer DBPanic(ctx, tx)

	// 更新主表数据
	if err := tx.Model(m).Where(conds[0], conds[1:]...).Updates(data).Error; err != nil {
		DBRollback(ctx, tx)
		return err
	}

	// 更新关联关系
	if err := DBAssociate(ctx, tx, m, upmap); err != nil {
		DBRollback(ctx, tx)
		return err
	}

	// 提交事务
	return DBCommit(ctx, tx)
}

// DBDelete 删除数据库记录
// ctx: 上下文
// db: GORM数据库实例
// model: 目标模型
// conds: 查询条件
// 返回操作可能产生的错误
func DBDelete(ctx context.Context, db *gorm.DB, model any, conds ...any) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // 返回取消原因
	default:
	}
	// 检查是否提供了查询条件
	if len(conds) == 0 {
		return gorm.ErrMissingWhereClause
	}

	// 执行删除操作
	return db.Delete(model, conds...).Error
}

// DBFind 查询单条数据库记录，支持预加载关联关系
// ctx: 上下文
// db: GORM数据库实例
// preloads: 需要预加载的关联关系列表
// m: 查询结果存储对象
// conds: 查询条件
// 返回操作可能产生的错误
func DBFind(ctx context.Context, db *gorm.DB, preloads []string, m any, conds ...any) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // 返回取消原因
	default:
	}
	// 预加载关联关系
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	// 查询第一条匹配的记录
	return db.First(m, conds...).Error
}

// QueryParams 查询参数结构体，用于配置列表查询的各种参数
type QueryParams struct {
	Preloads []string       // 需要预加载的关联关系列表
	Query    map[string]any // 查询条件映射
	OrderBy  []string       // 排序字段列表
	Limit    int            // 限制返回记录数
	Offset   int            // 偏移量
	IsCount  bool           // 是否只查询总数
}

func NewPksQueryParams(pks []uint32) QueryParams {
	return QueryParams{
		Preloads: []string{},
		Query:    map[string]any{"id in ?": pks},
		OrderBy:  []string{"id"},
		IsCount:  false,
		Limit:    0,
		Offset:   0,
	}
}

func (q *QueryParams) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	// 记录预加载字段
	if len(q.Preloads) > 0 {
		enc.AddString("preloads", strings.Join(q.Preloads, ","))
	} else {
		enc.AddString("preloads", "")
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
	enc.AddInt("limit", q.Limit)
	enc.AddInt("offset", q.Offset)
	// 记录是否查询总数
	enc.AddBool("is_count", q.IsCount)
	return nil
}

// DBList 查询数据库记录列表，支持分页、排序、条件查询等功能
// ctx: 上下文
// db: GORM数据库实例
// model: 目标模型
// value: 查询结果存储对象
// query: 查询参数
// 返回记录总数和操作可能产生的错误
func DBList(ctx context.Context, db *gorm.DB, model, value any, query QueryParams) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err() // 返回取消原因
	default:
	}
	// 初始化查询构建器
	mdb := db.Model(model)

	// 预加载关联关系
	for _, preload := range query.Preloads {
		mdb = mdb.Preload(preload)
	}

	// 添加查询条件
	for k, v := range query.Query {
		mdb = mdb.Where(k, v)
	}

	// 查询总数
	var count int64 = 0
	if query.IsCount {
		mdb = mdb.Count(&count)
	}

	// 添加排序条件
	orderByStr := strings.Join(query.OrderBy, ",")
	if orderByStr != "" {
		mdb = mdb.Order(orderByStr)
	}

	// 添加分页条件
	if query.Limit > 0 {
		mdb = mdb.Limit(query.Limit)
	}
	if query.Offset > 0 {
		mdb = mdb.Offset(query.Offset)
	}

	// 执行查询
	result := mdb.Find(value)
	if result.Error != nil {
		return 0, result.Error
	}

	// 如果没有查询总数，则使用影响行数作为总数
	if !query.IsCount {
		count = result.RowsAffected
	}

	return count, nil
}
