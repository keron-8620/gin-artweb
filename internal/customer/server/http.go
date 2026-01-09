package server

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/internal/customer/data"
	"gin-artweb/internal/customer/service"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
	"gin-artweb/pkg/crypto"
)

func NewServer(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
) {
	if err := dbAutoMigrate(init.DB); err != nil {
		loggers.Server.Fatal("数据库自动迁移customer模型失败", zap.Error(err))
	}

	permissionRepo := data.NewPermissionRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	menuRepo := data.NewMenuRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	buttonRepo := data.NewButtonRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	roleRepo := data.NewRoleRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	userRepo := data.NewUserRepo(loggers.Data, init.DB, init.DBTimeout)
	recordRepo := data.NewLoginRecordRepo(loggers.Data, init.DB, init.DBTimeout,
		time.Duration(init.Conf.Security.Login.LockMinutes)*time.Minute,
		time.Duration(init.Conf.Security.Token.ClearMinutes)*time.Minute,
		init.Conf.Security.Login.MaxFailedAttempts,
	)

	permissionUsecase := biz.NewPermissionUsecase(loggers.Biz, permissionRepo)
	menuUsecase := biz.NewMenuUsecase(loggers.Biz, permissionRepo, menuRepo)
	buttonUsecase := biz.NewButtonUsecase(loggers.Biz, permissionRepo, menuRepo, buttonRepo)
	roleUsecase := biz.NewRoleUsecase(loggers.Biz, permissionRepo, menuRepo, buttonRepo, roleRepo)
	userUsecase := biz.NewUserUsecase(loggers.Biz, roleRepo, userRepo, recordRepo, crypto.NewBcryptHasher(12), init.Conf.Security)
	recordUsecase := biz.NewLoginRecordUsecase(loggers.Biz, recordRepo)

	ctx := context.Background()
	if pErr := permissionUsecase.LoadPermissionPolicy(ctx); pErr != nil {
		loggers.Server.Fatal("系统初始化加载权限策略时失败", zap.Error(pErr))
	}
	if pErr := menuUsecase.LoadMenuPolicy(ctx); pErr != nil {
		loggers.Server.Fatal("系统初始化加载菜单策略时失败", zap.Error(pErr))
	}
	if pErr := buttonUsecase.LoadButtonPolicy(ctx); pErr != nil {
		loggers.Server.Fatal("系统初始化加载按钮策略时失败", zap.Error(pErr))
	}
	if pErr := roleUsecase.LoadRolePolicy(ctx); pErr != nil {
		loggers.Server.Fatal("系统初始化加载角色策略时失败", zap.Error(pErr))
	}

	pPolicies, pErr := init.Enforcer.GetPolicy()
	if pErr != nil {
		loggers.Server.Fatal("系统初始化查询p策略失败", zap.Error(pErr))
	}
	loggers.Data.Debug("已加载所有p策略", zap.Any("pPolicies", pPolicies))

	gPolicies, gErr := init.Enforcer.GetGroupingPolicy()
	if gErr != nil {
		loggers.Server.Fatal("系统初始化查询g策略失败", zap.Error(gErr))
	}
	loggers.Data.Debug("已加载所有g策略", zap.Any("gPolicies", gPolicies))

	permissionService := service.NewPermissionService(loggers.Service, permissionUsecase)
	menuService := service.NewMenuService(loggers.Service, menuUsecase)
	buttonService := service.NewButtonService(loggers.Service, buttonUsecase)
	roleService := service.NewRoleService(loggers.Service, roleUsecase)
	userService := service.NewUserService(loggers.Service, userUsecase, recordUsecase)
	recordService := service.NewRecordService(loggers.Service, recordUsecase)

	router.POST("/v1/login", userService.Login)
	appRouter := router.Group("/v1/customer")

	appRouter.Use(middleware.JWTAuthMiddleware(init.Conf.Security.Token.SecretKey, loggers.Service))
	appRouter.GET("/me/menu/tree", roleService.GetRoleMenuTree)
	appRouter.PATCH("/me/password", userService.PatchPassword)

	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))
	permissionService.LoadRouter(appRouter)
	menuService.LoadRouter(appRouter)
	buttonService.LoadRouter(appRouter)
	roleService.LoadRouter(appRouter)
	userService.LoadRouter(appRouter)
	recordService.LoadRouter(appRouter)
}
