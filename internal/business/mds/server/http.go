package server

import (
	"time"

	"github.com/gin-gonic/gin"

	"gin-artweb/internal/business/mds/biz"
	"gin-artweb/internal/business/mds/data"
	"gin-artweb/internal/business/mds/service"
	jobsServer "gin-artweb/internal/infra/jobs/server"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

func NewServer(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
	jobsUsecase *jobsServer.JobsUsecase,

) {
	jwtConf := auth.NewJWTConfig(
		time.Duration(init.Conf.Security.Token.AccessMinutes)*time.Minute,
		time.Duration(init.Conf.Security.Token.RefreshMinutes)*time.Minute,
		init.Conf.Security.Token.AccessMethod,
		init.Conf.Security.Token.RefreshMethod,
	)

	colonyRepo := data.NewMdsColonyRepo(loggers.Data, init.DB, init.DBTimeout)
	nodeRepo := data.NewMdsNodeRepo(loggers.Data, init.DB, init.DBTimeout)

	colonyUsecase := biz.NewMdsColonyUsecase(loggers.Biz, colonyRepo)
	nodeUsecase := biz.NewMdsNodeUsecase(loggers.Biz, nodeRepo)
	recordUsecase := biz.NewRecordUsecase(loggers.Biz, jobsUsecase.Record)
	taskUsecase := biz.NewMdsTaskExecutionInfoUsecase(loggers.Biz, recordUsecase)

	colonyService := service.NewMdsColonyService(loggers.Service, colonyUsecase, taskUsecase)
	nodeService := service.NewMdsNodeService(loggers.Service, nodeUsecase)
	confService := service.NewMdsConfService(loggers.Service, colonyUsecase, int64(init.Conf.Security.Upload.MaxConfSize)*1024*1024)

	appRouter := router.Group("/v1/mds")
	appRouter.Use(middleware.JWTAuthMiddleware(jwtConf, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	colonyService.LoadRouter(appRouter)
	nodeService.LoadRouter(appRouter)
	confService.LoadRouter(appRouter)
}
