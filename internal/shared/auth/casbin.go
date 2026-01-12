package auth

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"

	"gin-artweb/pkg/ctxutil"
)

func NewCasbinEnforcer() (*casbin.Enforcer, error) {
	cm, err := model.NewModelFromString(`
		[request_definition]
		r = sub, obj, act

		[policy_definition]
		p = sub, obj, act

		[role_definition]
		g = _, _

		[policy_effect]
		e = some(where (p.eft == allow))

		[matchers]
		m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
	`)
	if err != nil {
		return nil, err
	}
	adapter := stringadapter.NewAdapter("p, perm_0, /api/v1/login, POST")
	return casbin.NewEnforcer(cm, adapter)
}

const (
	SubKey = "sub"
	ObjKey = "obj"
	ActKey = "act"

	GroupSubKey = "group_parent" // 组策略主体键
	GroupObjKey = "group_child"  // 组策略对象键
)

const (
	permissionSubjectFormat = "perm_%d"
	menuSubjectFormat       = "menu_%d"
	buttonSubjectFormat     = "button_%d"
	roleSubjectFormat       = "role_%d"
)

// PermissionToSubject 将权限ID转换为对应的Casbin主体
func PermissionToSubject(pk uint32) string {
	return fmt.Sprintf(permissionSubjectFormat, pk)
}

// MenuToSubject 将菜单ID转换为对应的Casbin主体
func MenuToSubject(pk uint32) string {
	return fmt.Sprintf(menuSubjectFormat, pk)
}

// ButtonToSubject 将按钮ID转换为对应的Casbin主体
func ButtonToSubject(pk uint32) string {
	return fmt.Sprintf(buttonSubjectFormat, pk)
}

// RoleToSubject 将角色ID转换为对应的Casbin主体
func RoleToSubject(pk uint32) string {
	return fmt.Sprintf(roleSubjectFormat, pk)
}

// AddPolicies 批量添加授权策略规则
// rules: 要添加的策略规则列表，每个规则是一个字符串切片
// 返回值: 如果添加成功返回nil，否则返回相应的错误信息
func AddPolicy(ctx context.Context, enf *casbin.Enforcer, sub, obj, act string) error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return err
	}
	// 参数校验
	if sub == "" || obj == "" || act == "" {
		return fmt.Errorf("添加策略失败: 参数不能为空")
	}

	// 添加策略
	if _, err := enf.AddPolicy(sub, obj, act); err != nil {
		return fmt.Errorf("添加策略失败: %w", err)
	}
	return nil
}

// // RemovePolicies 批量移除授权策略规则
// // rules: 要移除的策略规则列表，每个规则是一个字符串切片
// // 返回值: 如果移除成功返回nil，否则返回相应的错误信息
func RemovePolicy(ctx context.Context, enf *casbin.Enforcer, sub, obj, act string) error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return err
	}
	// 参数校验
	if sub == "" || obj == "" || act == "" {
		return fmt.Errorf("移除策略失败: 参数不能为空")
	}

	// 移除策略
	if _, err := enf.RemovePolicy(sub, obj, act); err != nil {
		return fmt.Errorf("移除策略失败: %w", err)
	}
	return nil
}

// AddGroupPolicies 批量添加用户组策略规则
// 返回值: 如果添加成功返回nil，否则返回相应的错误信息
func AddGroupPolicy(ctx context.Context, enf *casbin.Enforcer, sub, obj string) error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return err
	}
	// 参数校验
	if sub == "" || obj == "" {
		return fmt.Errorf("添加组策略失败: 参数不能为空")
	}

	// 添加组策略
	if _, err := enf.AddGroupingPolicy(sub, obj); err != nil {
		return fmt.Errorf("添加组策略失败: %w", err)
	}
	return nil
}

// RemoveGroupPolicies 批量移除用户组策略规则
// 返回值: 如果移除成功返回nil，否则返回相应的错误信息
func RemoveGroupPolicy(ctx context.Context, enf *casbin.Enforcer, index int, value string) error {
	if err := ctxutil.CheckContext(ctx); err != nil {
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
	if _, err := enf.RemoveFilteredGroupingPolicy(index, value); err != nil {
		return fmt.Errorf("删除组策略失败: %w", err)
	}
	return nil
}
