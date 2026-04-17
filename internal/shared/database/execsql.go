package database

import (
	"context"
	"os"
	"strings"

	"emperror.dev/errors"
	"gorm.io/gorm"
)

// ExecSQLFile 执行 SQL 脚本文件，支持所有数据库 + 所有注释类型
func ExecSQLFile(ctx context.Context, db *gorm.DB, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 核心：清理所有注释（// -- # /* */）
	sqlClean := cleanSQLComments(string(content))
	// 按分号拆分语句
	sqlList := splitSQL(sqlClean)

	// 开启事务
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "执行SQL语句时开启事务失败")
	}

	// 捕获异常，使用事务对象tx确保在panic时能正确回滚事务
	defer func() {
		_ = DBPanic(ctx, tx)
	}()

	// 逐条执行
	for _, sql := range sqlList {
		if sql == "" {
			continue
		}
		// 使用事务对象执行SQL，确保所有操作都在同一事务中
		if err := tx.Exec(sql).Error; err != nil {
			tx.Rollback()
			return errors.Wrap(err, "执行SQL语句时执行失败")
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "执行SQL语句时提交事务失败")
	}
	return nil
}

// cleanSQLComments 清理：//  --  #  /* */ 全部注释
func cleanSQLComments(sql string) string {
	// 1. 去掉 多行注释 /* ... */
	clean := removeMultiLineComments(sql)
	// 2. 去掉 单行注释 -- 、 # 、 //
	clean = removeSingleLineComments(clean)
	return clean
}

// 移除 /* ... */ 多行注释
func removeMultiLineComments(s string) string {
	for {
		start := strings.Index(s, "/*")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "*/")
		if end == -1 {
			break
		}
		end += start
		s = s[:start] + s[end+2:]
	}
	return s
}

// 移除 --、#、// 单行注释
func removeSingleLineComments(s string) string {
	lines := strings.Split(s, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// 去掉 -- 注释
		if idx := strings.Index(line, "--"); idx != -1 {
			line = line[:idx]
		}
		// 去掉 # 注释
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}
		// 去掉 // 注释
		if idx := strings.Index(line, "//"); idx != -1 {
			line = line[:idx]
		}

		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return strings.Join(result, " ")
}

// splitSQL 按 ; 分割 SQL
func splitSQL(sql string) []string {
	var statements []string
	for _, stmt := range strings.Split(sql, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}
	return statements
}
