package biz

import (
	"net/http"

	"gin-artweb/internal/shared/errors"
)

var (
	ErrExportOesColonyFailed = errors.New(
		http.StatusInternalServerError,
		"export_oes_colony_failed",
		"导出oes集群配置文件失败",
		nil,
	)
	ErrUntarGzOesPackage = errors.New(
		http.StatusInternalServerError,
		"untar_gz_oes_package_failed",
		"解压oes程序包失败",
		nil,
	)
	ErrUntarGzXCounterPackage = errors.New(
		http.StatusInternalServerError,
		"untar_gz_xcounter_package_failed",
		"解压xcounter程序包失败",
		nil,
	)
	ErrOesColonyListEmpty = errors.New(
		http.StatusNotFound,
		"oes_colony_list_empty",
		"oes集群列表为空",
		nil,
	)
	ErrOesColonySystemTypeInvalid = errors.New(
		http.StatusBadRequest,
		"oes_colony_system_type_invalid",
		"oes集群系统类型非法",
		nil,
	)
	ErrOesColonyHasTooManyFlags = errors.New(
		http.StatusBadRequest,
		"oes_colony_has_too_many_flags",
		"oes集群任务存在多个标识文件",
		nil,
	)
	ErrOesColonyInvalidFlag = errors.New(
		http.StatusBadRequest,
		"oes_colony_invalid_flag",
		"oes集群的标识文件非法",
		nil,
	)
)
