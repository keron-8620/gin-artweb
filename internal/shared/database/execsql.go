package database

import (
	"context"

	"gorm.io/gorm"
)

func ExecSQL(ctx context.Context, db *gorm.DB, sql string, args ...any) error {
	// 开启事务
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 捕获异常，使用事务对象tx确保在panic时能正确回滚事务
	defer DBPanic(ctx, tx)

	// 使用事务对象执行SQL，确保所有操作都在同一事务中
	if err := tx.Exec(sql, args...).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
