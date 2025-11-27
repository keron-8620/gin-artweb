package biz

import (
	"net/http"

	"gin-artweb/internal/shared/errors"
)

var (
	ErrSSHConnect = errors.New(
		http.StatusBadRequest,
		"ssh_connection_failed",
		"SSH连接失败",
		nil,
	)
	
	ErrSSHKeyDeployment = errors.New(
		http.StatusBadRequest,
		"ssh_key_deployment_failed",
		"SSH密钥部署失败",
		nil,
	)

	ErrExportHostFailed = errors.New(
        http.StatusInternalServerError,
        "export_host_failed",
        "导出主机配置文件失败",
        nil,
    )

	ErrDeleteHostFileFailed = errors.New(
        http.StatusInternalServerError,
        "delete_host_file_failed",
        "删除主机配置文件失败",
        nil,
    )
	
	ErrRemovePakage = errors.New(
		http.StatusBadRequest,
		"remove_package_failed",
		"删除软件包失败",
		nil,
	)
)