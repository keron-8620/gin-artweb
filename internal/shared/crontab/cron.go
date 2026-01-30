package crontab

import (
	"time"

	"emperror.dev/errors"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func NewCron(logger *zap.Logger) *cron.Cron {
	// 创建cron logger适配器
	cronLogger := &cronLog{logger: logger}

	// 初始化cron实例
	return cron.New(
		// 设置自定义logger
		cron.WithLogger(cronLogger),

		// 设置本地时区
		cron.WithLocation(time.Local),

		// 添加panic恢复
		cron.WithChain(
			cron.Recover(cronLogger),
		),
	)
}

// ValidateCronExpression 校验cron表达式是否合法
// 参数 expr: 待校验的cron表达式
// 参数 withSeconds: 是否支持秒级（标准cron是5位，秒级是6位）
// 返回值: 合法返回true，不合法返回false及错误信息
func ValidateCronExpression(expr string, withSeconds bool) (bool, error) {
	// 创建解析器，根据是否支持秒级选择不同的解析规则
	var parser cron.Parser
	if withSeconds {
		// 支持秒级（6位：秒 分 时 日 月 周），兼容标准cron和秒级cron
		parser = cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
		)
	} else {
		// 标准cron（5位：分 时 日 月 周）
		parser = cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
		)
	}

	// 解析表达式，解析失败会返回错误
	_, err := parser.Parse(expr)
	if err != nil {
		return false, errors.Wrap(err, "cron表达式不合法")
	}
	return true, nil
}
