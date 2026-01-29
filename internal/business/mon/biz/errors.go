package biz

import (
	"net/http"

	"gin-artweb/internal/shared/errors"
)

var (
	ErrExportMonNodeFailed = errors.New(
		http.StatusInternalServerError,
		"export_mon_node_file_failed",
		"导出mon节点文件失败",
		nil,
	)
	ErrDeleteMonNodeFileFailed = errors.New(
		http.StatusInternalServerError,
		"delete_mon_node_file_failed",
		"删除mon节点文件失败",
		nil,
	)
)
