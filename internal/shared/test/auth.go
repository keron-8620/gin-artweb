package test

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"

	"gin-artweb/internal/shared/auth"
)

func NewTestEnforcer(jwtKey string) (*casbin.Enforcer, error) {
	base := auth.RoleToSubject(0)
	policies := []string{
		fmt.Sprintf("p, %s, /api/v1/customer/me/password, PATCH", base),
		fmt.Sprintf("p, %s, /api/v1/customer/me/menu/tree, GET", base),
	}
	baseProlicies := strings.Join(policies, "\n")

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
	adapter := stringadapter.NewAdapter(baseProlicies)
	return casbin.NewEnforcer(cm, adapter)
}
