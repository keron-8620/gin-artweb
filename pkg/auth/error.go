package auth

import (
	"net/http"

	"gitee.com/keion8620/go-dango-gin/pkg/errors"
)

const (
	ErrAddPolicyMSG         = "failed to add policy"
	ErrRemovePolicyMSG      = "failed to remove policy"
	ErrAddGroupPolicyMSG    = "failed to add group policy"
	ErrRemoveGroupPolicyMSG = "failed to remove group policy"
)

var (
	RsonNoAuthorization = errors.ReasonEnum{
		Reason: "no_authorization",
		Msg:    "请求头中缺少授权令牌",
	}
	RsonInvalidToken = errors.ReasonEnum{
		Reason: "invalid_token",
		Msg:    "无效或未知的授权令牌",
	}
	RsonTokenExpired = errors.ReasonEnum{
		Reason: "token_expired",
		Msg:    "授权令牌已过期",
	}
	RsonForbidden = errors.ReasonEnum{
		Reason: "forbidden",
		Msg:    "您没有访问该资源的权限",
	}
	RsonCtxUserNotFound = errors.ReasonEnum{
		Reason: "ctx_user_not_found",
		Msg:    "无法从上下文中获取用户信息",
	}
	RsonGeneToken = errors.ReasonEnum{
		Reason: "generate_token_failed",
		Msg:    "生成token失败",
	}
)

var (
	ErrNoAuthor = errors.New(
		http.StatusUnauthorized,
		RsonNoAuthorization.Reason,
		RsonNoAuthorization.Msg,
		nil,
	)
	ErrInvalidToken = errors.New(
		http.StatusUnauthorized,
		RsonInvalidToken.Reason,
		RsonInvalidToken.Msg,
		nil,
	)
	ErrTokenExpired = errors.New(
		http.StatusUnauthorized,
		RsonTokenExpired.Reason,
		RsonTokenExpired.Msg,
		nil,
	)
	ErrForbidden = errors.New(
		http.StatusForbidden,
		RsonForbidden.Reason,
		RsonForbidden.Msg,
		nil,
	)
	ErrCtxUserNotFound = errors.New(
		http.StatusNotFound,
		RsonCtxUserNotFound.Reason,
		RsonCtxUserNotFound.Msg,
		nil,
	)
	ErrGeneToken = errors.New(
		http.StatusInternalServerError,
		RsonGeneToken.Reason,
		RsonGeneToken.Msg,
		nil,
	)
)
