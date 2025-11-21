package biz

import (
	"net/http"

	"gin-artweb/pkg/errors"
)


var (
	ErrScriptDisabled = errors.New(
		http.StatusBadRequest,
		"script_disabled",
		"脚本已禁用",
		nil,
	)
)
