package errors

import (
	"net/http"
)

var ValidateError = New(
	http.StatusBadRequest,
	"validation",
	"参数验证错误",
	nil,
)
