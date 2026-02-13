package server

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/infra/customer/biz"
	"gin-artweb/internal/infra/customer/data"
	"gin-artweb/internal/infra/customer/service"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
	"gin-artweb/pkg/crypto"
)

type CustomerUsecase struct {
	Api    *biz.ApiUsecase
	Menu   *biz.MenuUsecase
	Button *biz.ButtonUsecase
	Role   *biz.RoleUsecase
	User   *biz.UserUsecase
}

func NewServer(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
) *CustomerUsecase {
	secSettings := biz.SecuritySettings{
		MaxFailedAttempts: init.Conf.Security.Login.MaxFailedAttempts,
		LockDuration:      time.Duration(init.Conf.Security.Login.LockMinutes) * time.Minute,
		PasswordStrength:  init.Conf.Security.Password.StrengthLevel,
	}

	apiRepo := data.NewApiRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	menuRepo := data.NewMenuRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	buttonRepo := data.NewButtonRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	roleRepo := data.NewRoleRepo(loggers.Data, init.DB, init.DBTimeout, init.Enforcer)
	userRepo := data.NewUserRepo(loggers.Data, init.DB, init.DBTimeout)
	recordRepo := data.NewLoginRecordRepo(loggers.Data, init.DB, init.DBTimeout,
		time.Duration(init.Conf.Security.Login.LockMinutes)*time.Minute,
		time.Duration(init.Conf.Security.Token.AccessMinutes*2)*time.Minute,
		init.Conf.Security.Login.MaxFailedAttempts,
	)

	apiUsecase := biz.NewApiUsecase(loggers.Biz, apiRepo)
	menuUsecase := biz.NewMenuUsecase(loggers.Biz, apiRepo, menuRepo)
	buttonUsecase := biz.NewButtonUsecase(loggers.Biz, apiRepo, menuRepo, buttonRepo)
	roleUsecase := biz.NewRoleUsecase(loggers.Biz, apiRepo, menuRepo, buttonRepo, roleRepo)
	userUsecase := biz.NewUserUsecase(
		loggers.Biz,
		roleRepo, userRepo,
		recordRepo,
		crypto.NewBcryptHasher(12), init.JwtConf, secSettings)

	ctx := context.Background()
	if pErr := apiUsecase.LoadApiPolicy(ctx); pErr != nil {
		loggers.Server.Error("系统初始化加载API策略时失败", zap.Error(pErr))
		panic(pErr)
	}
	if pErr := menuUsecase.LoadMenuPolicy(ctx); pErr != nil {
		loggers.Server.Error("系统初始化加载菜单策略时失败", zap.Error(pErr))
		panic(pErr)
	}
	if pErr := buttonUsecase.LoadButtonPolicy(ctx); pErr != nil {
		loggers.Server.Error("系统初始化加载按钮策略时失败", zap.Error(pErr))
		panic(pErr)
	}
	if pErr := roleUsecase.LoadRolePolicy(ctx); pErr != nil {
		loggers.Server.Error("系统初始化加载角色策略时失败", zap.Error(pErr))
		panic(pErr)
	}

	pPolicies, pErr := init.Enforcer.GetPolicy()
	if pErr != nil {
		loggers.Server.Error("系统初始化查询p策略失败", zap.Error(pErr))
		panic(pErr)
	}
	loggers.Data.Debug("已加载所有p策略", zap.Any("pPolicies", pPolicies))

	gPolicies, gErr := init.Enforcer.GetGroupingPolicy()
	if gErr != nil {
		loggers.Server.Error("系统初始化查询g策略失败", zap.Error(gErr))
		panic(gErr)
	}
	loggers.Data.Debug("已加载所有g策略", zap.Any("gPolicies", gPolicies))

	apiService := service.NewApiService(loggers.Service, apiUsecase)
	menuService := service.NewMenuService(loggers.Service, menuUsecase)
	buttonService := service.NewButtonService(loggers.Service, buttonUsecase)
	roleService := service.NewRoleService(loggers.Service, roleUsecase)
	userService := service.NewUserService(loggers.Service, userUsecase)

	router.POST("/v1/login", userService.Login)
	router.POST("/v1/refresh/token", userService.RefreshToken)
	appRouter := router.Group("/v1/customer")

	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Service))
	appRouter.GET("/me/menu/tree", roleService.GetRoleMenuTree)
	appRouter.PATCH("/me/password", userService.PatchPassword)

	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))
	apiService.LoadRouter(appRouter)
	menuService.LoadRouter(appRouter)
	buttonService.LoadRouter(appRouter)
	roleService.LoadRouter(appRouter)
	userService.LoadRouter(appRouter)

	return &CustomerUsecase{
		Api:    apiUsecase,
		Menu:   menuUsecase,
		Button: buttonUsecase,
		Role:   roleUsecase,
		User:   userUsecase,
	}
}
