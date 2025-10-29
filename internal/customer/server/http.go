package server

import (
	"context"
	"path/filepath"
	"time"

	"github.com/casbin/casbin/v2"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gitee.com/keion8620/go-dango-gin/internal/customer/biz"
	"gitee.com/keion8620/go-dango-gin/internal/customer/data"
	"gitee.com/keion8620/go-dango-gin/internal/customer/service"
	"gitee.com/keion8620/go-dango-gin/pkg/auth"
	"gitee.com/keion8620/go-dango-gin/pkg/common"
	"gitee.com/keion8620/go-dango-gin/pkg/config"
	"gitee.com/keion8620/go-dango-gin/pkg/crypto"
)

func NewServer(
	router *gin.RouterGroup,
	conf *config.SystemConf,
	db *gorm.DB,
	logger *zap.Logger,
) {
	ctx := context.Background()
	modelPath := filepath.Join(common.ConfigDir, "model.conf")
	adapter := stringadapter.NewAdapter(`p, admin, *, *`)
	enf, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		panic(err)
	}
	enforcer := auth.NewAuthEnforcer(enf, logger, conf.Security.SecretKey)
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

	permissionService := service.NewPermissionService(logger, permissionUsecase)
	menuService := service.NewMenuService(logger, menuUsecase)
	buttonService := service.NewButtonService(logger, buttonUsecase)
	roleService := service.NewRoleService(logger, roleUsecase)
	userService := service.NewUserService(logger, userUsecase, recordUsecase)

	appRouter := router.Group("/v1/customer")
	permissionService.LoadRouter(appRouter)
	menuService.LoadRouter(appRouter)
	buttonService.LoadRouter(appRouter)
	roleService.LoadRouter(appRouter)
	userService.LoadRouter(appRouter)

	permissionUsecase.LoadPermissionPolicy(ctx)
	menuUsecase.LoadMenuPolicy(ctx)
	buttonUsecase.LoadButtonPolicy(ctx)
	roleUsecase.LoadRolePolicy(ctx)
	router.Use(auth.AuthMiddleware(enforcer, "/api/v1/customer/own/login"))
}
