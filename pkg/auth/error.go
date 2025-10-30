package auth

import (
	"net/http"

	"gin-artweb/pkg/errors"
)

var (
	ErrNoAuthor = errors.New(
		http.StatusUnauthorized,
		"no_authorization",
		"请求头中缺少授权令牌",
		nil,
	)
	ErrInvalidToken = errors.New(
		http.StatusUnauthorized,
		"invalid_token",
		"无效或未知的授权令牌",
		nil,
	)
	ErrTokenExpired = errors.New(
		http.StatusUnauthorized,
		"token_expired",
		"授权令牌已过期",
		nil,
	)
	ErrForbidden = errors.New(
		http.StatusForbidden,
		"forbidden",
		"您没有访问该资源的权限",
		nil,
	)
	ErrGeneToken = errors.New(
		http.StatusInternalServerError,
		"generate_token_failed",
		"生成token失败",
		nil,
	)
	ErrGetUserClaims = errors.New(
		http.StatusInternalServerError,
		"get_user_claims_failed",
		"无法从上下文中提取有效的用户身份信息",
		nil,
	)
	ErrUserClaimsMissing = errors.New(
		http.StatusInternalServerError,
		"user_claims_missing",
		"请求上下文中未找到用户身份信息",
		nil,
	)
	ErrSetClaims = errors.New(
		http.StatusInternalServerError,
		"set_claims_failed",
		"设置用户信息失败",
		nil,
	)
)
