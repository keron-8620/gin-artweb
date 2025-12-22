package biz

import (
	"net/http"

	"gin-artweb/internal/shared/errors"
)


var (
	ErrScriptDisabled = errors.New(
		http.StatusBadRequest,
		"script_disabled",
		"脚本已禁用",
		nil,
	)
	ErrNoSuchScriptFile = errors.New(
		http.StatusBadRequest,
		"no_such_script_file",
		"没有找到脚本文件",
		nil,
	)
	ErrScriptEnvInvalid = errors.New(
		http.StatusBadRequest,
		"script_env_invalid",
		"脚本环境变量无效",
		nil,
	)
	ErrCreateLogFailed = errors.New(
		http.StatusInternalServerError,
		"create_log_failed",
		"创建日志文件或目录失败",
		nil,
	)
	ErrAddScheduleFailed = errors.New(
		http.StatusInternalServerError,
		"add_schedule_failed",
		"添加计划任务到调度器中失败",
		nil,
	)
	ErrScriptIsBuiltin = errors.New(
		http.StatusBadRequest,
		"script_is_builtin",
		"内置脚本不允许修改或删除",
		nil,
	)
)


// 2. 获取环境变量
// var env string
// if len(req.EnvVars) > 0 {
// 	envBytes, err := json.Marshal(req.EnvVars)
// 	if err != nil {
// 		uc.log.Error(
// 			"序列化环境变量失败, 忽略默认环境变量",
// 			zap.Error(err),
// 			zap.Any("env", req.EnvVars),
// 			zap.Uint32(ScriptIDKey, req.ScriptID),
// 			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
// 		)
// 		return nil, ErrScriptEnvInvalid
// 	}
// 	env = string(envBytes)
// }