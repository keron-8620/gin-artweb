package errors

import (
	"net/http"
)

func FromReason(reason ErrorReason) *Error {
	message, ok := defaultErrorMessages[reason]
	if !ok {
		message = "未知错误"
	}
	return New(reason, message, nil)
}

// GetHTTPStatus 根据错误原因获取对应的HTTP状态码
func GetHTTPStatus(reason ErrorReason) int {
	if status, ok := reasonToStatus[reason]; ok {
		return status
	}
	// 默认返回500内部服务器错误
	return http.StatusInternalServerError
}

// ErrorResponse 生成错误响应，自动根据错误原因获取HTTP状态码
func ErrorResponse(err *Error) map[string]any {
	if err == nil {
		return nil
	}
	status := GetHTTPStatus(err.Reason)
	response := err.Fields()
	response["code"] = status
	return response
}

// ErrorResponseWithCode 生成错误响应，使用指定的HTTP状态码
func ErrorResponseWithCode(code int, err *Error) map[string]any {
	if err == nil {
		return nil
	}
	response := err.Fields()
	response["code"] = code
	return response
}

// RespondWithError 在Gin等框架中直接使用，返回错误响应
func RespondWithError(c interface {
	AbortWithStatusJSON(code int, obj any)
}, err *Error) {
	if err == nil {
		c.AbortWithStatusJSON(http.StatusOK, nil)
		return
	}
	status := GetHTTPStatus(err.Reason)
	c.AbortWithStatusJSON(status, ErrorResponse(err))
}
