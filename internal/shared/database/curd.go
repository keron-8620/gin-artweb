// Package database 提供数据库CRUD操作的通用方法
// 包括事务处理、关联关系更新、增删改查等常用数据库操作
package database

import (
	"context"
	"runtime/debug"
	"strings"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

func contextError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		return context.DeadlineExceeded
	}
	return nil
}

// DBPanic 通用 GORM 数据库操作 panic 捕获函数（必须配合 defer 使用）
// 功能:1. 捕获 panic 并转为标准错误 2. 事务回滚（仅事务场景）3. 记录详细日志（含堆栈、SQL 上下文）
// 参数:
//   - ctx:上下文，用于日志记录和 GORM 操作
//   - db:GORM 数据库实例（事务场景传入 tx 实例，非事务场景传入普通 db 实例）
//
// 返回值:
//   - error:panic 时返回封装了「错误信息+堆栈」的标准错误，无 panic 时返回 nil
func DBPanic(ctx context.Context, db *gorm.DB, bizErr error) error {
	// 1. 捕获 panic
	r := recover()
	if r == nil {
		// 无 panic，返回业务原本的错误
		return bizErr
	}

	// 2. 安全回滚事务（防御式判断，避免空指针 panic）
	if db != nil && db.Statement != nil && db.Statement.ConnPool != nil {
		// 回滚失败只打日志，不影响主流程
		if rollbackErr := db.Rollback().Error; rollbackErr != nil && db.Logger != nil {
			db.Logger.Error(ctx, "事务回滚失败", "panic", r, "rollback_err", rollbackErr)
		}
	}

	// 3. 构建错误信息
	stackInfo := string(debug.Stack())
	errMsg := "数据库操作发生 panic"

	// 4. 打印完整日志（SQL、参数、堆栈、panic 内容）
	if db != nil && db.Logger != nil {
		logFields := []any{
			"panic", r,
			"stack", stackInfo,
		}
		// 安全追加 SQL 上下文
		if db.Statement != nil {
			logFields = append(logFields, "sql", db.Statement.SQL.String())
			logFields = append(logFields, "vars", db.Statement.Vars)
		}
		db.Logger.Error(ctx, errMsg, logFields...)
	}

	return errors.NewWithDetails("recovered database panic", "panic", r)
}

// DBCreate 创建数据库记录
// ctx: 上下文
// db: GORM数据库实例
// model: 目标模型
// value: 要创建的数据
// 返回操作可能产生的错误
func DBCreate(ctx context.Context, db *gorm.DB, model, value any) error {
	if err := contextError(ctx); err != nil {
		return err
	}

	if err := db.WithContext(ctx).Model(model).Create(value).Error; err != nil {
		return errors.WrapIf(err, "创建数据库记录失败")
	}
	return nil
}

// withTransaction 统一处理事务提交、回滚和 panic 恢复，确保 panic 不会被吞掉。
func withTransaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) (err error) {
	if db == nil {
		return errors.New("数据库实例为空")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := contextError(ctx); err != nil {
		return err
	}

	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return errors.WrapIf(tx.Error, "数据库事务开启失败")
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			rollbackErr := tx.Rollback().Error
			err = errors.NewWithDetails(
				"数据库事务发生 panic",
				"panic", recovered,
				"stack", string(debug.Stack()),
			)
			if rollbackErr != nil {
				err = errors.WrapWithDetails(err, "数据库事务回滚失败", "rollback_error", rollbackErr)
			}
			return
		}

		if err != nil {
			if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
				err = errors.WrapWithDetails(err, "数据库事务回滚失败", "rollback_error", rollbackErr)
			}
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit().Error; err != nil {
		return errors.WrapIf(err, "数据库事务提交失败")
	}
	return nil
}

func DBCreateTX(ctx context.Context, db *gorm.DB, model, value any) error {
	return withTransaction(ctx, db, func(tx *gorm.DB) error {
		if err := tx.Model(model).Create(value).Error; err != nil {
			return errors.WrapIf(err, "创建数据库记录失败")
		}
		return nil
	})
}

func DBCreateRelationTx(ctx context.Context, db *gorm.DB, model, value any, upmap map[string]any) error {
	return withTransaction(ctx, db, func(tx *gorm.DB) error {
		if err := tx.Model(model).Create(value).Error; err != nil {
			return errors.WrapIf(err, "创建数据库记录失败")
		}
		for k, v := range upmap {
			if err := tx.Model(value).Association(k).Append(v); err != nil {
				return errors.WrapIf(err, "更新关联关系失败")
			}
		}
		return nil
	})
}

// DBUpdate 更新数据库记录，支持关联关系更新
// ctx: 上下文
// db: GORM数据库实例
// m: 目标模型
// data: 主表更新数据 (可为nil或空map)
// upmap: 关联关系映射 (可为nil或空map)
// conds: 查询条件
// 返回操作可能产生的错误
func DBUpdate(ctx context.Context, db *gorm.DB, m any, data map[string]any, conds ...any) error {
	if err := contextError(ctx); err != nil {
		return err
	}

	// 如果没有需要更新的内容，直接返回
	if len(data) == 0 {
		return nil
	}

	// 检查是否提供了查询条件
	if len(conds) == 0 {
		return errors.WithStack(gorm.ErrMissingWhereClause)
	}

	if err := db.WithContext(ctx).Model(m).Where(conds[0], conds[1:]...).Updates(data).Error; err != nil {
		return errors.WrapIf(err, "更新数据库记录失败")
	}
	return nil
}

func DBUpdateTx(ctx context.Context, db *gorm.DB, m any, data map[string]any, conds ...any) error {
	if err := contextError(ctx); err != nil {
		return err
	}

	// 如果没有需要更新的内容，直接返回
	if len(data) == 0 {
		return nil
	}

	// 检查是否提供了查询条件
	if len(conds) == 0 {
		return errors.WithStack(gorm.ErrMissingWhereClause)
	}

	return withTransaction(ctx, db, func(tx *gorm.DB) error {
		if err := tx.Model(m).Where(conds[0], conds[1:]...).Updates(data).Error; err != nil {
			return errors.WrapIf(err, "更新数据库记录失败")
		}
		return nil
	})
}

func DBUpdateRelationTx(ctx context.Context, db *gorm.DB, m any, data map[string]any, upmap map[string]any, conds ...any) error {
	// 检查是否提供了查询条件
	if len(conds) == 0 {
		return errors.WithStack(gorm.ErrMissingWhereClause)
	}

	return withTransaction(ctx, db, func(tx *gorm.DB) error {
		if len(data) > 0 {
			if err := tx.Model(m).Where(conds[0], conds[1:]...).Updates(data).Error; err != nil {
				return errors.WrapIf(err, "更新数据库记录失败")
			}
		}

		if err := tx.Where(conds[0], conds[1:]...).First(m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.WrapIf(err, "记录不存在")
			}
			return errors.WrapIf(err, "查询记录失败")
		}

		for k, v := range upmap {
			if err := tx.Model(m).Association(k).Replace(v); err != nil {
				return errors.WrapIf(err, "更新关联关系失败")
			}
		}
		return nil
	})
}

// DBDeleteTx 删除数据库记录
// ctx: 上下文
// db: GORM数据库实例
// model: 目标模型
// conds: 查询条件
// 返回操作可能产生的错误
func DBDeleteTx(ctx context.Context, db *gorm.DB, model any, conds ...any) error {
	if err := contextError(ctx); err != nil {
		return err
	}

	// 检查是否提供了查询条件
	if len(conds) == 0 {
		return gorm.ErrMissingWhereClause
	}

	return withTransaction(ctx, db, func(tx *gorm.DB) error {
		if err := tx.Delete(model, conds...).Error; err != nil {
			return errors.WrapIf(err, "删除数据库记录失败")
		}
		return nil
	})
}

// DBGet 查询单条数据库记录，支持预加载关联关系
// ctx: 上下文
// db: GORM数据库实例
// preloads: 需要预加载的关联关系列表
// m: 查询结果存储对象
// conds: 查询条件
// 返回操作可能产生的错误
func DBGet(ctx context.Context, db *gorm.DB, preloads []string, m any, conds ...any) error {
	if err := contextError(ctx); err != nil {
		return err
	}

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
func DBList(ctx context.Context, db *gorm.DB, model, value any, query QueryParams) error {
	if err := contextError(ctx); err != nil {
		return err
	}

	// 初始化查询构建器
	mdb := db.WithContext(ctx).Model(model)

	// 添加查询条件
	for k, v := range query.Query {
		mdb = mdb.Where(k, v)
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
	if query.Limit > 0 {
		mdb = mdb.Limit(query.Limit)
	}
	if query.Offset > 0 {
		mdb = mdb.Offset(query.Offset)
	}

	// 预加载关联关系
	for _, preload := range query.Preloads {
		mdb = mdb.Preload(preload)
	}

	// 执行查询
	result := mdb.Find(value)
	if result.Error != nil {
		return errors.WrapIf(result.Error, "查询数据库记录失败")
	}

	return nil
}

func DBCount(ctx context.Context, db *gorm.DB, model any, query map[string]any) (int64, error) {
	if err := contextError(ctx); err != nil {
		return 0, err
	}

	// 初始化查询构建器
	mdb := db.WithContext(ctx).Model(model)

	// 添加查询条件
	for k, v := range query {
		if k == "" {
			continue
		}
		mdb = mdb.Where(k, v)
	}

	// 查询总数
	var count int64 = 0
	if err := mdb.Count(&count).Error; err != nil {
		return 0, errors.WrapIf(err, "查询数据库记录总数失败")
	}

	return count, nil
}

// QueryParams 查询参数结构体，用于配置列表查询的各种参数
type QueryParams struct {
	Preloads []string       // 需要预加载的关联关系列表
	Query    map[string]any // 查询条件映射
	OrderBy  []string       // 排序字段列表
	Limit    int            // 分页大小
	Offset   int            // 分页偏移量
	Omit     []string       // 需要忽略的字段列表
	Columns  []string       // 查询字段列表
}

func (q *QueryParams) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	// 记录预加载字段
	if len(q.Preloads) > 0 {
		enc.AddString("preloads", strings.Join(q.Preloads, ","))
	} else {
		enc.AddString("preloads", "")
	}

	// 记录查询条件
	if err := enc.AddReflected("query", q.Query); err != nil {
		return err
	}

	// 记录排序字段
	if len(q.OrderBy) > 0 {
		enc.AddString("order_by", strings.Join(q.OrderBy, ","))
	} else {
		enc.AddString("order_by", "")
	}

	// 记录分页参数
	enc.AddInt("limit", q.Limit)
	enc.AddInt("offset", q.Offset)

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
