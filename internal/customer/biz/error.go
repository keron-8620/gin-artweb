package biz

import (
	"net/http"

	"gitee.com/keion8620/go-dango-gin/pkg/errors"
)

var (
	RsonAddPolicy = errors.ReasonEnum{
		Reason: "add_policies_failed",
		Msg:    "添加策略失败",
	}
	RsonRemovePolicy = errors.ReasonEnum{
		Reason: "remove_policies_failed",
		Msg:    "删除策略失败",
	}
	RsonAddGroupPolicy = errors.ReasonEnum{
		Reason: "add_group_policies_failed",
		Msg:    "添加组策略失败",
	}
	RsonRemoveGroupPolicy = errors.ReasonEnum{
		Reason: "remove_group_policies_failed",
		Msg:    "删除组策略失败",
	}
	RsonPasswordStrengthFailed = errors.ReasonEnum{
		Reason: "password_strength_failed",
		Msg:    "密码强度不够",
	}
	RsonInvalidCredentials = errors.ReasonEnum{
		Reason: "invalid_credentials",
		Msg:    "用户名或密码错误",
	}
	RsonPasswordMismatch = errors.ReasonEnum{
		Reason: "password_mismatch",
		Msg:    "密码错误",
	}
	RsonUserInActive = errors.ReasonEnum{
		Reason: "user_inactive",
		Msg:    "用户未激活",
	}
)

var (
	ErrAddPolicy = errors.New(
		http.StatusInternalServerError,
		RsonAddPolicy.Reason,
		RsonAddPolicy.Msg,
		nil,
	)
	ErrRemovePolicy = errors.New(
		http.StatusInternalServerError,
		RsonRemovePolicy.Reason,
		RsonRemovePolicy.Msg,
		nil,
	)
	ErrAddGroupPolicy = errors.New(
		http.StatusInternalServerError,
		RsonAddGroupPolicy.Reason,
		RsonAddGroupPolicy.Msg,
		nil,
	)
	ErrRemoveGroupPolicy = errors.New(
		http.StatusInternalServerError,
		RsonRemoveGroupPolicy.Reason,
		RsonRemoveGroupPolicy.Msg,
		nil,
	)
	ErrPasswordStrengthFailed = errors.New(
		http.StatusBadRequest,
		RsonPasswordStrengthFailed.Reason,
		RsonPasswordStrengthFailed.Msg,
		nil,
	)
	ErrInvalidCredentials = errors.New(
		http.StatusUnauthorized,
		RsonInvalidCredentials.Reason,
		RsonInvalidCredentials.Msg,
		nil,
	)
	ErrPasswordMismatch = errors.New(
		http.StatusUnauthorized,
		RsonPasswordMismatch.Reason,
		RsonPasswordMismatch.Msg,
		nil,
	)
	ErrUserInActive = errors.New(
		http.StatusUnauthorized,
		RsonUserInActive.Reason,
		RsonUserInActive.Msg,
		nil,
	)
)
