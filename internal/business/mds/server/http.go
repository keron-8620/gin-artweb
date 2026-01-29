package server

import (
	"github.com/gin-gonic/gin"

	"gin-artweb/internal/business/mds/biz"
	"gin-artweb/internal/business/mds/data"
	"gin-artweb/internal/business/mds/service"
	jobsServer "gin-artweb/internal/infra/jobs/server"
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
	appRouter.Use(middleware.JWTAuthMiddleware(init.Conf.Security.Token.SecretKey, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	colonyService.LoadRouter(appRouter)
	nodeService.LoadRouter(appRouter)
	confService.LoadRouter(appRouter)
}
