package server

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/casbin/casbin/v2"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
	"go.uber.org/zap"

	"gin-artweb/pkg/auth"
	"gin-artweb/pkg/common"
)

func baseRoleMeProlicys() string {
	base := auth.RoleToSubject(0)
	policies := []string{
		fmt.Sprintf("p, %s, /api/v1/customer/me/password, PATCH", base),
		fmt.Sprintf("p, %s, /api/v1/customer/me/menu/tree, GET", base),
	}
	return strings.Join(policies, "\n")
}

func NewCasbinEnforcer(logger *zap.Logger, jwtKey string) (*auth.AuthEnforcer, error) {
	modelPath := filepath.Join(common.ConfigDir, "model.conf")
	adapter := stringadapter.NewAdapter(baseRoleMeProlicys())
	enf, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		logger.Error("创建casbin失败", zap.Error(err))
		return nil, err
	}
	return auth.NewAuthEnforcer(enf, jwtKey), nil
}
