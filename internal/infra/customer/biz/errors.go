package biz

// import (
// 	"net/http"

// 	"gin-artweb/internal/shared/errors"
// )

// var (
// 	ErrAddPolicy = errors.New(
// 		http.StatusInternalServerError,
// 		"add_policies_failed",
// 		"添加策略失败",
// 		nil,
// 	)
// 	ErrRemovePolicy = errors.New(
// 		http.StatusInternalServerError,
// 		"remove_policies_failed",
// 		"删除策略失败",
// 		nil,
// 	)
// 	ErrAddGroupPolicy = errors.New(
// 		http.StatusInternalServerError,
// 		"add_group_policies_failed",
// 		"添加组策略失败",
// 		nil,
// 	)
// 	ErrRemoveGroupPolicy = errors.New(
// 		http.StatusInternalServerError,
// 		"remove_group_policies_failed",
// 		"删除组策略失败",
// 		nil,
// 	)
// 	ErrPasswordStrengthFailed = errors.New(
// 		http.StatusBadRequest,
// 		"password_strength_failed",
// 		"密码强度不够",
// 		nil,
// 	)
// 	ErrInvalidCredentials = errors.New(
// 		http.StatusUnauthorized,
// 		"invalid_credentials",
// 		"用户名或密码错误",
// 		nil,
// 	)
// 	ErrPasswordMismatch = errors.New(
// 		http.StatusUnauthorized,
// 		"password_mismatch",
// 		"密码错误",
// 		nil,
// 	)
// 	ErrUserInActive = errors.New(
// 		http.StatusUnauthorized,
// 		"user_inactive",
// 		"用户未激活",
// 		nil,
// 	)
// 	ErrAccessLock = errors.New(
// 		http.StatusUnauthorized,
// 		"account_locked_too_many_attempts",
// 		"因登录失败次数过多，账户已被锁定",
// 		nil,
// 	)

// 	ErrPermissionExists = errors.New(
// 		http.StatusNotFound,
// 		"perm_exists",
// 		"用户不存在",
// 		nil,
// 	)

// )
