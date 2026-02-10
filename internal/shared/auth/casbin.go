package auth

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
)

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
		return nil, errors.WrapIf(err, "创建Casbin模型失败")
	}
	adapter := stringadapter.NewAdapter("p, perm_0, /api/v1/login, POST")
	enforcer, eErr := casbin.NewEnforcer(cm, adapter)
	if eErr != nil {
		return nil, errors.WrapIf(eErr, "创建Casbin enforce失败")
	}
	return enforcer, nil
}

// AddPolicies 批量添加授权策略规则
// rules: 要添加的策略规则列表，每个规则是一个字符串切片
// 返回值: 如果添加成功返回nil，否则返回相应的错误信息
func AddPolicies(ctx context.Context, enf *casbin.Enforcer, rules [][]string) error {
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "添加Casbin策略: 上下文已取消")
	}

	// 添加策略
	for _, rule := range rules {
		if len(rule) != 3 {
			return errors.NewWithDetails(
				"添加Casbin策略失败: 每个策略规则必须包含3个元素",
				"rule", rule,
			)
		}
		if _, err := enf.AddPolicy(rule); err != nil {
			return errors.WrapIfWithDetails(
				err, "添加Casbin策略失败",
				"rule", rule,
			)
		}
	}
	// if _, err := enf.AddPolicies(rules); err != nil {
	// 	return errors.WrapIfWithDetails(
	// 		err, "添加Casbin策略失败",
	// 		"rules", rules,
	// 	)
	// }
	return nil
}

// RemovePolicies 批量移除授权策略规则
// rules: 要移除的策略规则列表，每个规则是一个字符串切片
// 返回值: 如果移除成功返回nil，否则返回相应的错误信息
func RemovePolicies(ctx context.Context, enf *casbin.Enforcer, rules [][]string) error {
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "移除Casbin策略: 上下文已取消")
	}

	// 移除策略
	for _, rule := range rules {
		if len(rule) != 3 {
			return errors.NewWithDetails(
				"移除Casbin策略失败: 每个策略规则必须包含3个元素",
				"rule", rule,
			)
		}
		if _, err := enf.RemovePolicy(rule); err != nil {
			return errors.WrapIfWithDetails(
				err, "移除Casbin策略失败",
				"rule", rule,
			)
		}
	}
	// if _, err := enf.RemovePolicies(rules); err != nil {
	// 	return errors.WrapIfWithDetails(
	// 		err, "移除Casbin策略失败",
	// 		"rules", rules,
	// 	)
	// }
	return nil
}

// AddGroupPolicies 批量添加用户组策略规则
// 返回值: 如果添加成功返回nil，否则返回相应的错误信息
func AddGroupPolicies(ctx context.Context, enf *casbin.Enforcer, rules [][]string) error {
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "添加Casbin组策略: 上下文已取消")
	}

	// 添加组策略
	for _, rule := range rules {
		if len(rule) != 2 {
			return errors.NewWithDetails(
				"添加Casbin组策略失败: 每个组策略规则必须包含2个元素",
				"rule", rule,
			)
		}
		if _, err := enf.AddGroupingPolicy(rule); err != nil {
			return errors.WrapWithDetails(
				err, "添加Casbin组策略失败",
				"rule", rule,
			)
		}
	}
	// if _, err := enf.AddGroupingPolicies(rules); err != nil {
	// 	return errors.WrapIfWithDetails(
	// 		err, "添加Casbin组策略失败",
	// 		"rules", rules,
	// 	)
	// }
	return nil
}

// RemoveFilteredGroupingPolicy 批量移除用户组策略规则
// 返回值: 如果移除成功返回nil，否则返回相应的错误信息
func RemoveFilteredGroupingPolicy(ctx context.Context, enf *casbin.Enforcer, index int, value string) error {
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "移除Casbin组策略: 上下文已取消")
	}
	// 参数校验
	if value == "" || index < 0 || index > 1 {
		return errors.NewWithDetails(
			"移除Casbin组策略失败: 值不能为空且索引必须在0-1之间",
			"index", index,
			"value", value,
		)
	}

	// 移除组策略
	if _, err := enf.RemoveFilteredGroupingPolicy(index, value); err != nil {
		return errors.WrapIfWithDetails(
			err, "添加Casbin组策略失败",
			"index", index,
			"value", value,
		)
	}
	return nil
}
