package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/mds/biz"
	"gin-artweb/internal/mds/data"
	"gin-artweb/internal/mds/service"
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
		loggers.Server.Fatal("数据库自动迁移mds模型失败", zap.Error(err))
	}

	colonyRepo := data.NewMdsColonyRepo(loggers.Data, init.DB, init.DBTimeout)
	nodeRepo := data.NewMdsNodeRepo(loggers.Data, init.DB, init.DBTimeout)

	colonyUsecase := biz.NewMdsColonyUsecase(loggers.Biz, colonyRepo)
	nodeUsecase := biz.NewMdsNodeUsecase(loggers.Biz, nodeRepo)

	colonyService := service.NewMdsColonyService(loggers.Service, colonyUsecase)
	nodeService := service.NewMdsNodeService(loggers.Service, nodeUsecase)
	confService := service.NewMdsConfService(loggers.Service, colonyUsecase, int64(init.Conf.Security.Upload.MaxConfSize)*1024*1024)

	appRouter := router.Group("/v1/mds")
	appRouter.Use(middleware.JWTAuthMiddleware(init.Conf.Security.Token.SecretKey, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	colonyService.LoadRouter(appRouter)
	nodeService.LoadRouter(appRouter)
	confService.LoadRouter(appRouter)
}
