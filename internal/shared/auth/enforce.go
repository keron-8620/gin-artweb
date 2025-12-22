package auth

import (
	"context"
	goerrors "errors"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/golang-jwt/jwt/v5"

	"gin-artweb/internal/shared/errors"
)

const (
	SubKey = "sub"
	ObjKey = "obj"
	ActKey = "act"

	GroupSubKey = "group_sub" // 组策略主体键
	GroupObjKey = "group_obj" // 组策略对象键
)

// AuthEnforcer 管理身份验证令牌和授权权限
// 它提供线程安全的操作来存储和检索用户声明、角色权限和URL访问控制映射
type AuthEnforcer struct {
	// enforcer 用于访问控制
	Enforcer *casbin.Enforcer

	// jwt的密钥
	key []byte
}

// NewAuthEnforcer 创建一个新的认证缓存实例
// 返回初始化后的AuthCache指针
func NewAuthEnforcer(enforcer *casbin.Enforcer, key string) *AuthEnforcer {
	return &AuthEnforcer{
		Enforcer: enforcer,
		key:      []byte(key),
	}
}

// Authentication 验证给定令牌的有效性并返回相应的用户认证信息
// token：待验证的JWT令牌字符串
// 返回用户认证信息和可能发生的错误（如无效令牌、已过期等）
func (c *AuthEnforcer) Authentication(ctx context.Context, token string) (*UserClaims, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	// 解析token
	parsedToken, err := jwt.ParseWithClaims(token, &UserClaims{}, func(token *jwt.Token) (any, error) {
		return c.key, nil
	})
	if err != nil {
		if goerrors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken.WithCause(err)
	}

	// 验证token有效性
	if !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	// 类型断言获取claims
	claims, ok := parsedToken.Claims.(*UserClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// Authorization 检查特定角色是否具有对某个HTTP方法和URL路径组合的访问权限
// roleId：请求访问的角色
// url：请求的目标URL路径
// method：HTTP请求方法（GET/POST等）
// 返回是否有访问权限的布尔结果
func (c *AuthEnforcer) Authorization(ctx context.Context, role, url, method string) (bool, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return false, errors.FromError(err)
	}
	ok, err := c.Enforcer.Enforce(role, url, method)
	if err != nil {
		return false, errors.FromError(err)
	}
	return ok, nil
}

// AddPolicies 批量添加授权策略规则
// rules: 要添加的策略规则列表，每个规则是一个字符串切片
// 返回值: 如果添加成功返回nil，否则返回相应的错误信息
func (c *AuthEnforcer) AddPolicy(ctx context.Context, sub, obj, act string) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}

	// 参数校验
	if sub == "" || obj == "" || act == "" {
		return fmt.Errorf("添加策略失败: 参数不能为空")
	}

	// 添加策略
	if _, err := c.Enforcer.AddPolicy(sub, obj, act); err != nil {
		return fmt.Errorf("添加策略失败: %w", err)
	}
	return nil
}

// // RemovePolicies 批量移除授权策略规则
// // rules: 要移除的策略规则列表，每个规则是一个字符串切片
// // 返回值: 如果移除成功返回nil，否则返回相应的错误信息
func (c *AuthEnforcer) RemovePolicy(ctx context.Context, sub, obj, act string) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}

	// 参数校验
	if sub == "" || obj == "" || act == "" {
		return fmt.Errorf("移除策略失败: 参数不能为空")
	}

	// 移除策略
	if _, err := c.Enforcer.RemovePolicy(sub, obj, act); err != nil {
		return fmt.Errorf("移除策略失败: %w", err)
	}
	return nil
}

// AddGroupPolicies 批量添加用户组策略规则
// 返回值: 如果添加成功返回nil，否则返回相应的错误信息
func (c *AuthEnforcer) AddGroupPolicy(ctx context.Context, sub, obj string) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}

	// 参数校验
	if sub == "" || obj == "" {
		return fmt.Errorf("添加组策略失败: 参数不能为空")
	}

	// 添加组策略
	if _, err := c.Enforcer.AddGroupingPolicy(sub, obj); err != nil {
		return fmt.Errorf("添加组策略失败: %w", err)
	}
	return nil
}

// RemoveGroupPolicies 批量移除用户组策略规则
// 返回值: 如果移除成功返回nil，否则返回相应的错误信息
func (c *AuthEnforcer) RemoveGroupPolicy(ctx context.Context, index int, value string) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}

	// 参数校验
	if value == "" {
		return fmt.Errorf("移除组策略失败: 值不能为空")
	}
	if index < 0 || index > 1 {
		return fmt.Errorf("移除组策略失败: 索引必须在0-1之间")
	}

	// 移除组策略
	if _, err := c.Enforcer.RemoveFilteredGroupingPolicy(index, value); err != nil {
		return fmt.Errorf("删除组策略失败: %w", err)
	}
	return nil
}
