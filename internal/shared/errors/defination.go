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
	ErrSetFilePermission = New(
		http.StatusInternalServerError,
		"set_file_permission",
		"设置文件权限错误",
		nil,
	)
		ErrNoAuthor = New(
		http.StatusUnauthorized,
		"no_authorization",
		"请求头中缺少授权令牌",
		nil,
	)
	ErrInvalidToken = New(
		http.StatusUnauthorized,
		"invalid_token",
		"无效或未知的授权令牌",
		nil,
	)
	ErrTokenExpired = New(
		http.StatusUnauthorized,
		"token_expired",
		"授权令牌已过期",
		nil,
	)
	ErrForbidden = New(
		http.StatusForbidden,
		"forbidden",
		"您没有访问该资源的权限",
		nil,
	)
	ErrGeneToken = New(
		http.StatusInternalServerError,
		"generate_token_failed",
		"生成token失败",
		nil,
	)
	ErrGetUserClaims = New(
		http.StatusInternalServerError,
		"get_user_claims_failed",
		"无法从上下文中提取有效的用户身份信息",
		nil,
	)
	ErrUserClaimsMissing = New(
		http.StatusInternalServerError,
		"user_claims_missing",
		"请求上下文中未找到用户身份信息",
		nil,
	)
	ErrSetClaims = New(
		http.StatusInternalServerError,
		"set_claims_failed",
		"设置用户信息失败",
		nil,
	)
)
