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
	// 读取SQL文件
	content, err := os.ReadFile(filePath) // #nosec G304
	if err != nil {
		return errors.WithMessagef(err, "读取SQL文件失败, 路径: %s", filePath)
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

// 安全版本：只删除 单引号外部 的注释，不会破坏字符串内容
func cleanSQLComments(sql string) string {
	var result strings.Builder
	n := len(sql)
	inQuote := false // 是否在 ' 字符串内部

	for i := 0; i < n; {
		if sql[i] == '\'' {
			inQuote = !inQuote
			result.WriteByte(sql[i])
			i++
			continue
		}

		// 不在字符串内，才处理注释
		if !inQuote {
			// 多行注释 /*
			if i+1 < n && sql[i] == '/' && sql[i+1] == '*' {
				// 跳过直到 */
				for i += 2; i+1 < n; i++ {
					if sql[i] == '*' && sql[i+1] == '/' {
						i += 2
						break
					}
				}
				continue
			}

			// 单行注释 --  //  #
			if i+1 < n && ((sql[i] == '-' && sql[i+1] == '-') || (sql[i] == '/' && sql[i+1] == '/')) {
				// 跳过本行剩余
				for i < n && sql[i] != '\n' {
					i++
				}
				continue
			}
			if sql[i] == '#' {
				for i < n && sql[i] != '\n' {
					i++
				}
				continue
			}
		}

		// 正常字符
		result.WriteByte(sql[i])
		i++
	}

	return strings.TrimSpace(result.String())
}

// splitSQL 按 ; 分割 SQL
func splitSQL(sql string) []string {
	var statements []string
	var current strings.Builder
	n := len(sql)
	inQuote := false

	for i := 0; i < n; i++ {
		c := sql[i]
		if c == '\'' {
			inQuote = !inQuote
			current.WriteByte(c)
		} else if c == ';' && !inQuote {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}

	// 最后一段
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}
