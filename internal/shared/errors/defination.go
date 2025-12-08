package errors

import (
	"net/http"
)

var (
	ErrHostNotAllowed = New(
		http.StatusForbidden,
		"host_not_allowed",
		"主机不允许访问",
		nil,
	)
	ValidateError = New(
		http.StatusBadRequest,
		"validation",
		"参数验证错误",
		nil,
	)
	ClientClosedRequest = New(
		http.StatusBadRequest,
		"client_cancelled",
		"客户端关闭了请求",
		nil,
	)
	ErrNoUploadedFileFound = New(
		http.StatusBadRequest,
		"no_uploaded_file_found",
		"未找到上传的文件",
		nil,
	)
	ErrFileTooLarge = New(
		http.StatusBadRequest,
		"file_too_large",
		"上传文件超出大小限制",
		nil,
	)
	ErrFileStatusCheckFailed = New(
		http.StatusInternalServerError,
		"file_status_check_failed",
		"文件状态检查失败",
		nil,
	)
	ErrFileNotFound = New(
		http.StatusBadRequest,
		"file_not_found",
		"未找到文件",
		nil,
	)
	ErrUploadFile = New(
		http.StatusInternalServerError,
		"upload_file",
		"上传文件错误",
		nil,
	)
	ErrRemoveFile = New(
		http.StatusInternalServerError,
		"remove_file",
		"删除文件错误",
		nil,
	)
)
