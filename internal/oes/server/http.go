package server

import (
	"github.com/gin-gonic/gin"

	jobsServer "gin-artweb/internal/jobs/server"
	"gin-artweb/internal/oes/biz"
	"gin-artweb/internal/oes/data"
	"gin-artweb/internal/oes/service"
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
	colonyRepo := data.NewOesColonyRepo(loggers.Data, init.DB, init.DBTimeout)
	nodeRepo := data.NewOesNodeRepo(loggers.Data, init.DB, init.DBTimeout)

	colonyUsecase := biz.NewOesColonyUsecase(loggers.Biz, colonyRepo)
	nodeUsecase := biz.NewOesNodeUsecase(loggers.Biz, nodeRepo)
	recordUsecase := biz.NewRecordUsecase(loggers.Biz, jobsUsecase.Record)
	stkTaskUsecase := biz.NewStkTaskExecutionInfoUsecase(loggers.Biz, recordUsecase)
	crdaskUsecase := biz.NewCrdTaskExecutionInfoUsecase(loggers.Biz, recordUsecase)
	optTaskUsecase := biz.NewOptTaskExecutionInfoUsecase(loggers.Biz, recordUsecase)

	colonyService := service.NewOesColonyService(loggers.Service, colonyUsecase, stkTaskUsecase, crdaskUsecase, optTaskUsecase)
	nodeService := service.NewOesNodeService(loggers.Service, nodeUsecase)
	confService := service.NewOesConfService(loggers.Service, colonyUsecase, int64(init.Conf.Security.Upload.MaxConfSize)*1024*1024)

	appRouter := router.Group("/v1/oes")
	appRouter.Use(middleware.JWTAuthMiddleware(init.Conf.Security.Token.SecretKey, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	colonyService.LoadRouter(appRouter)
	nodeService.LoadRouter(appRouter)
	confService.LoadRouter(appRouter)
}
