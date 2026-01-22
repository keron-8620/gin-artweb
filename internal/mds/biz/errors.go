package biz

import (
	"net/http"

	"gin-artweb/internal/shared/errors"
)

var (
	ErrExportMdsColonyFailed = errors.New(
		http.StatusInternalServerError,
		"export_mds_colony_failed",
		"导出mds集群配置文件失败",
		nil,
	)
	ErrUntarGzMdsPackage = errors.New(
		http.StatusInternalServerError,
		"untar_gz_mds_package_failed",
		"解压mds程序包失败",
		nil,
	)
	ErrMdsColonyNotFound = errors.New(
		http.StatusNotFound,
		"mds_colony_not_found",
		"mds集群不存在",
		nil,
	)
	ErrMdsColonyListEmpty = errors.New(
		http.StatusNotFound,
		"mds_colony_list_empty",
		"mds集群列表为空",
		nil,
	)
	ErrMdsColonyHasTooManyFlags = errors.New(
		http.StatusBadRequest,
		"mds_colony_has_too_many_flags",
		"mds集群任务存在多个标识文件",
		nil,
	)
	ErrMdsColonyInvalidFlag = errors.New(
		http.StatusBadRequest,
		"mds_colony_invalid_flag",
		"mds集群的标识文件非法",
		nil,
	)
)
