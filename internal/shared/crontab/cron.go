package crontab

import (
	"time"

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
