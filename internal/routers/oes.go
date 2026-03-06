package routers

import (
	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/oes"
	oesrepo "gin-artweb/internal/repository/oes"
	oessvc "gin-artweb/internal/service/oes"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

func newOesRouter(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
	jobsvc *JobsRouter,
) {
	colonyRepo := oesrepo.NewOesColonyRepo(loggers.Data, init.DB, init.DBTimeout)
	nodeRepo := oesrepo.NewOesNodeRepo(loggers.Data, init.DB, init.DBTimeout)

	colonyService := oessvc.NewOesColonyService(loggers.Biz, colonyRepo)
	nodeService := oessvc.NewOesNodeService(loggers.Biz, nodeRepo)
	recordService := oessvc.NewRecordService(loggers.Biz, jobsvc.Script, jobsvc.Record, jobsvc.Schedule)
	stkTaskUsecase := oessvc.NewStkTaskExecutionInfoUsecase(loggers.Biz, recordService)
	crdaskUsecase := oessvc.NewCrdTaskExecutionInfoUsecase(loggers.Biz, recordService)
	optTaskUsecase := oessvc.NewOptTaskExecutionInfoUsecase(loggers.Biz, recordService)

	colonyHandler := handler.NewOesColonyService(loggers.Service, colonyService, stkTaskUsecase, crdaskUsecase, optTaskUsecase)
	nodeHandler := handler.NewOesNodeService(loggers.Service, nodeService)
	confHandler := handler.NewOesConfService(loggers.Service, int64(init.Conf.Upload.MaxConfSize)*1024*1024)

	appRouter := router.Group("/v1/oes")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	colonyHandler.LoadRouter(appRouter)
	nodeHandler.LoadRouter(appRouter)
	confHandler.LoadRouter(appRouter)
}
