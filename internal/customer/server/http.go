package server

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/internal/customer/data"
	"gin-artweb/internal/customer/service"
	"gin-artweb/pkg/auth"
	"gin-artweb/pkg/config"
	"gin-artweb/pkg/crypto"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/log"
)

func NewServer(
	router *gin.RouterGroup,
	conf *config.SystemConf,
	db *gorm.DB,
	dbTimeout *database.DBTimeout,
	loggers *log.Loggers,
) {
	if err := dbAutoMigrate(db, loggers.Data); err != nil {
		panic(err)
	}

	enforcer, err := NewCasbinEnforcer(loggers.Service, conf.Security.Token.SecretKey)
	if err != nil {
		panic(err)
	}

	permissionRepo := data.NewPermissionRepo(loggers.Data, db, dbTimeout, enforcer)
	menuRepo := data.NewMenuRepo(loggers.Data, db, dbTimeout, enforcer)
	buttonRepo := data.NewButtonRepo(loggers.Data, db, dbTimeout, enforcer)
	roleRepo := data.NewRoleRepo(loggers.Data, db, dbTimeout, enforcer)
	userRepo := data.NewUserRepo(loggers.Data, db, dbTimeout)
	recordRepo := data.NewRecordRepo(loggers.Data, db, dbTimeout,
		time.Duration(conf.Security.Login.LockMinutes)*time.Minute,
		time.Duration(conf.Security.Token.ClearMinutes)*time.Minute,
		conf.Security.Login.MaxFailedAttempts,
	)

	permissionUsecase := biz.NewPermissionUsecase(loggers.Biz, permissionRepo)
	menuUsecase := biz.NewMenuUsecase(loggers.Biz, permissionRepo, menuRepo)
	buttonUsecase := biz.NewButtonUsecase(loggers.Biz, permissionRepo, menuRepo, buttonRepo)
	roleUsecase := biz.NewRoleUsecase(loggers.Biz, permissionRepo, menuRepo, buttonRepo, roleRepo)
	userUsecase := biz.NewUserUsecase(loggers.Biz, roleRepo, userRepo, recordRepo, crypto.NewBcryptHasher(12), conf.Security)
	recordUsecase := biz.NewRecordUsecase(loggers.Biz, recordRepo)

	ctx := context.Background()
	if pErr := permissionUsecase.LoadPermissionPolicy(ctx); pErr != nil {
		panic(pErr.Error())
	}
	if pErr := menuUsecase.LoadMenuPolicy(ctx); pErr != nil {
		panic(pErr.Error())
	}
	if pErr := buttonUsecase.LoadButtonPolicy(ctx); pErr != nil {
		panic(pErr.Error())
	}
	if pErr := roleUsecase.LoadRolePolicy(ctx); pErr != nil {
		panic(pErr.Error())
	}
	router.Use(auth.AuthMiddleware(enforcer, loggers.Service, "/api/v1/customer/login"))

	permissionService := service.NewPermissionService(loggers.Service, permissionUsecase)
	menuService := service.NewMenuService(loggers.Service, menuUsecase)
	buttonService := service.NewButtonService(loggers.Service, buttonUsecase)
	roleService := service.NewRoleService(loggers.Service, roleUsecase)
	userService := service.NewUserService(loggers.Service, userUsecase, recordUsecase)
	recordService := service.NewRecordService(loggers.Service, recordUsecase)

	appRouter := router.Group("/v1/customer")
	permissionService.LoadRouter(appRouter)
	menuService.LoadRouter(appRouter)
	buttonService.LoadRouter(appRouter)
	roleService.LoadRouter(appRouter)
	userService.LoadRouter(appRouter)
	recordService.LoadRouter(appRouter)
}
