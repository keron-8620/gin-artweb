package routers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	handler "gin-artweb/internal/handler/customer"
	custrepo "gin-artweb/internal/repository/customer"
	custsvc "gin-artweb/internal/service/customer"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
	"gin-artweb/pkg/crypto"
)

func newCustomerRouter(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
) {
	secSettings := custsvc.SecuritySettings{
		MaxFailedAttempts: init.Conf.Security.Login.MaxFailedAttempts,
		LockDuration:      time.Duration(init.Conf.Security.Login.LockMinutes) * time.Minute,
		PasswordStrength:  init.Conf.Security.Password.StrengthLevel,
	}

	apiRepo := custrepo.NewApiRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	menuRepo := custrepo.NewMenuRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	buttonRepo := custrepo.NewButtonRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	roleRepo := custrepo.NewRoleRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	userRepo := custrepo.NewUserRepo(loggers.Data, init.DB, init.DBTimeout)
	recordRepo := custrepo.NewLoginRecordRepo(loggers.Data, init.DB, init.DBTimeout,
		time.Duration(init.Conf.Security.Login.LockMinutes)*time.Minute,
		time.Duration(init.Conf.Security.Token.AccessMinutes*2)*time.Minute,
		init.Conf.Security.Login.MaxFailedAttempts,
	)

	apiService := custsvc.NewApiService(loggers.Biz, apiRepo)
	menuService := custsvc.NewMenuService(loggers.Biz, apiRepo, menuRepo)
	buttonService := custsvc.NewButtonService(loggers.Biz, apiRepo, menuRepo, buttonRepo)
	roleService := custsvc.NewRoleService(loggers.Biz, apiRepo, menuRepo, buttonRepo, roleRepo)
	userService := custsvc.NewUserService(
		loggers.Biz,
		roleRepo, userRepo,
		recordRepo,
		crypto.NewBcryptHasher(12), init.JwtConf, secSettings)

	ctx := context.Background()
	if pErr := apiService.LoadApiPolicy(ctx); pErr != nil {
		loggers.Server.Error("系统初始化加载API策略时失败", zap.Error(pErr))
		panic(pErr)
	}
	if pErr := menuService.LoadMenuPolicy(ctx); pErr != nil {
		loggers.Server.Error("系统初始化加载菜单策略时失败", zap.Error(pErr))
		panic(pErr)
	}
	if pErr := buttonService.LoadButtonPolicy(ctx); pErr != nil {
		loggers.Server.Error("系统初始化加载按钮策略时失败", zap.Error(pErr))
		panic(pErr)
	}
	if pErr := roleService.LoadRolePolicy(ctx); pErr != nil {
		loggers.Server.Error("系统初始化加载角色策略时失败", zap.Error(pErr))
		panic(pErr)
	}

	pPolicies, pErr := init.Enforcer.GetPolicy()
	if pErr != nil {
		loggers.Server.Error("系统初始化查询p策略失败", zap.Error(pErr))
		panic(pErr)
	}
	loggers.Service.Debug("已加载所有p策略", zap.Any("pPolicies", pPolicies))

	gPolicies, gErr := init.Enforcer.GetGroupingPolicy()
	if gErr != nil {
		loggers.Server.Error("系统初始化查询g策略失败", zap.Error(gErr))
		panic(gErr)
	}
	loggers.Service.Debug("已加载所有g策略", zap.Any("gPolicies", gPolicies))

	apiHandler := handler.NewApiHandler(loggers.Service, apiService)
	menuHandler := handler.NewMenuHandler(loggers.Service, menuService)
	buttonHandler := handler.NewButtonHandler(loggers.Service, buttonService)
	roleHandler := handler.NewRoleHandler(loggers.Service, roleService)
	userHandler := handler.NewUserHandler(loggers.Service, userService)

	router.POST("/v1/login", userHandler.Login)
	router.POST("/v1/refresh/token", userHandler.RefreshToken)
	appRouter := router.Group("/v1/customer")

	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Service))
	appRouter.GET("/me/menu/tree", roleHandler.GetRoleMenuTree)
	appRouter.PATCH("/me/password", userHandler.PatchPassword)

	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))
	apiHandler.LoadRouter(appRouter)
	menuHandler.LoadRouter(appRouter)
	buttonHandler.LoadRouter(appRouter)
	roleHandler.LoadRouter(appRouter)
	userHandler.LoadRouter(appRouter)
}
