package routers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	handler "gin-artweb/internal/handler/sys"
	sysrepo "gin-artweb/internal/repo/sys"
	syssvc "gin-artweb/internal/service/sys"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/middleware"
	"gin-artweb/pkg/crypto"
)

func newSysRouter(
	router *gin.RouterGroup,
	init *config.SystemInit,
	loggers *config.Loggers,
) {
	secSettings := syssvc.SecuritySettings{
		MaxFailedAttempts: init.Conf.Security.Login.MaxFailedAttempts,
		LockDuration:      time.Duration(init.Conf.Security.Login.LockMinutes) * time.Minute,
		PasswordStrength:  init.Conf.Security.Password.StrengthLevel,
	}

	apiRepo := sysrepo.NewApiRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold, init.Enforcer)
	menuRepo := sysrepo.NewMenuRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold, init.Enforcer)
	buttonRepo := sysrepo.NewButtonRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold, init.Enforcer)
	roleRepo := sysrepo.NewRoleRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold, init.Enforcer)
	userRepo := sysrepo.NewUserRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)
	recordRepo := sysrepo.NewLoginRecordRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold,
		time.Duration(init.Conf.Security.Login.LockMinutes)*time.Minute,
		time.Duration(init.Conf.Security.Token.AccessMinutes*2)*time.Minute,
		init.Conf.Security.Login.MaxFailedAttempts,
	)

	apiService := syssvc.NewApiService(loggers.Service, apiRepo)
	menuService := syssvc.NewMenuService(loggers.Service, apiRepo, menuRepo)
	buttonService := syssvc.NewButtonService(loggers.Service, apiRepo, menuRepo, buttonRepo)
	roleService := syssvc.NewRoleService(loggers.Service, apiRepo, menuRepo, buttonRepo, roleRepo)
	userService := syssvc.NewUserService(
		loggers.Service,
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
	loggers.Handler.Debug("已加载所有p策略", zap.Any("pPolicies", pPolicies))

	gPolicies, gErr := init.Enforcer.GetGroupingPolicy()
	if gErr != nil {
		loggers.Server.Error("系统初始化查询g策略失败", zap.Error(gErr))
		panic(gErr)
	}
	loggers.Handler.Debug("已加载所有g策略", zap.Any("gPolicies", gPolicies))

	apiHandler := handler.NewApiHandler(loggers.Handler, apiService)
	menuHandler := handler.NewMenuHandler(loggers.Handler, menuService)
	buttonHandler := handler.NewButtonHandler(loggers.Handler, buttonService)
	roleHandler := handler.NewRoleHandler(loggers.Handler, roleService)
	userHandler := handler.NewUserHandler(loggers.Handler, userService)

	router.POST("/v1/login", userHandler.Login)
	router.POST("/v1/refresh/token", userHandler.RefreshToken)
	appRouter := router.Group("/v1/customer")

	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Handler))
	appRouter.GET("/me/menu/tree", roleHandler.GetRoleMenuTree)
	appRouter.PATCH("/me/password", userHandler.PatchPassword)

	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Handler))
	apiHandler.LoadRouter(appRouter)
	menuHandler.LoadRouter(appRouter)
	buttonHandler.LoadRouter(appRouter)
	roleHandler.LoadRouter(appRouter)
	userHandler.LoadRouter(appRouter)
}
