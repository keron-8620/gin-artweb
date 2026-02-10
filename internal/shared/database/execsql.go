package database

import (
	"context"

	"emperror.dev/errors"
	"gorm.io/gorm"
)

func ExecSQL(ctx context.Context, db *gorm.DB, sql string, args ...any) error {
	// 开启事务
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "执行SQL语句时开启事务失败")
	}

	// 捕获异常，使用事务对象tx确保在panic时能正确回滚事务
	defer DBPanic(ctx, tx)

	// 使用事务对象执行SQL，确保所有操作都在同一事务中
	if err := tx.Exec(sql, args...).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(tx.Error, "执行SQL语句时执行失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.Wrap(tx.Error, "执行SQL语句时提交事务失败")
	}
	return nil
}
