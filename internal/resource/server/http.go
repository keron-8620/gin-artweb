package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/resource/biz"
	"gin-artweb/internal/resource/data"
	"gin-artweb/internal/resource/service"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

func NewServer(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
) {
	if err := dbAutoMigrate(init.DB); err != nil {
		loggers.Server.Error("数据库自动迁移resource模型失败", zap.Error(err))
		panic(err)
	}

	sshTimeout := time.Duration(init.Conf.SSH.Timeout) * time.Second
	signer, err := NewSigner(loggers.Data, init.Conf.SSH.Private, sshTimeout)
	if err != nil {
		loggers.Server.Error("创建SSH密钥签名失败", zap.Error(err))
		panic(err)
	}
	hostRepo := data.NewHostRepo(loggers.Data, init.DB, init.DBTimeout)
	pkgRepo := data.NewpackageRepo(loggers.Data, init.DB, init.DBTimeout)

	hostUsecase := biz.NewHostUsecase(loggers.Biz, hostRepo, signer, sshTimeout)
	pkgUsecase := biz.NewPackageUsecase(loggers.Biz, pkgRepo)

	hostService := service.NewHostService(loggers.Service, hostUsecase)
	pkgService := service.NewPackageService(loggers.Service, pkgUsecase, int64(init.Conf.Security.Upload.MaxPkgSize)*1024*1024)

	appRouter := router.Group("/v1/resource")
	appRouter.Use(middleware.JWTAuthMiddleware(init.Conf.Security.Token.SecretKey, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	hostService.LoadRouter(appRouter)
	pkgService.LoadRouter(appRouter)
}
