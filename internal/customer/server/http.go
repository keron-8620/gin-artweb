package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/casbin/casbin/v2"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/internal/customer/data"
	"gin-artweb/internal/customer/service"
	"gin-artweb/pkg/auth"
	"gin-artweb/pkg/config"
	"gin-artweb/pkg/crypto"
)

func NewServer(
	router *gin.RouterGroup,
	conf *config.SystemConf,
	db *gorm.DB,
	logger *zap.Logger,
) {
	if err := db.AutoMigrate(
		&biz.PermissionModel{},
		&biz.MenuModel{},
		&biz.ButtonModel{},
		&biz.RoleModel{},
		&biz.UserModel{},
		&biz.LoginRecordModel{},
	); err != nil {
		logger.Error("数据库自动迁移失败", zap.Error(err))
		panic(err)
	}

	ctx := context.Background()
	modelPath := filepath.Join(config.ConfigDir, "model.conf")
	loginURL := "/api/v1/customer/login"
	policyLine := fmt.Sprintf(`p, *, %s, %s`, loginURL, http.MethodPost)
	adapter := stringadapter.NewAdapter(policyLine)
	enf, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		logger.Error("创建casbin失败", zap.Error(err))
		panic(err)
	}
	enforcer := auth.NewAuthEnforcer(enf, conf.Security.SecretKey)
	hasher := crypto.NewBcryptHasher(12)

	permissionRepo := data.NewPermissionRepo(logger, db, enforcer)
	menuRepo := data.NewMenuRepo(logger, db, enforcer)
	buttonRepo := data.NewButtonRepo(logger, db, enforcer)
	roleRepo := data.NewRoleRepo(logger, db, enforcer)
	userRepo := data.NewUserRepo(logger, db)
	recordRepo := data.NewRecordRepo(logger, db,
		time.Duration(conf.Security.LoginFailLockMinutes)*time.Minute,
		time.Duration(conf.Security.TokenClearMinutes)*time.Minute,
		conf.Security.LoginFailMaxTimes,
	)

	permissionUsecase := biz.NewPermissionUsecase(logger, permissionRepo)
	menuUsecase := biz.NewMenuUsecase(logger, permissionRepo, menuRepo)
	buttonUsecase := biz.NewButtonUsecase(logger, permissionRepo, menuRepo, buttonRepo)
	roleUsecase := biz.NewRoleUsecase(logger, permissionRepo, menuRepo, buttonRepo, roleRepo)
	userUsecase := biz.NewUserUsecase(logger, roleRepo, userRepo, recordRepo, hasher, conf.Security)
	recordUsecase := biz.NewRecordUsecase(logger, recordRepo)

	permissionUsecase.LoadPermissionPolicy(ctx)
	menuUsecase.LoadMenuPolicy(ctx)
	buttonUsecase.LoadButtonPolicy(ctx)
	roleUsecase.LoadRolePolicy(ctx)
	router.Use(auth.AuthMiddleware(enforcer, logger, loginURL))

	permissionService := service.NewPermissionService(logger, permissionUsecase)
	menuService := service.NewMenuService(logger, menuUsecase)
	buttonService := service.NewButtonService(logger, buttonUsecase)
	roleService := service.NewRoleService(logger, roleUsecase)
	userService := service.NewUserService(logger, userUsecase, recordUsecase)
	recordService := service.NewRecordService(logger, recordUsecase)

	appRouter := router.Group("/v1/customer")
	permissionService.LoadRouter(appRouter)
	menuService.LoadRouter(appRouter)
	buttonService.LoadRouter(appRouter)
	roleService.LoadRouter(appRouter)
	userService.LoadRouter(appRouter)
	recordService.LoadRouter(appRouter)
}
