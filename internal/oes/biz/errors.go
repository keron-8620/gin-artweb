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
)
