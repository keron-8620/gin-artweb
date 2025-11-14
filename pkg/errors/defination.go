package errors

import (
	"net/http"
)

var (
	ValidateError = New(
		http.StatusBadRequest,
		"validation",
		"参数验证错误",
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
)
