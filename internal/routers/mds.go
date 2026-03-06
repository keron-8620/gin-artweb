package routers

import (
	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/mds"
	mdsrepo "gin-artweb/internal/repository/mds"
	mdssvc "gin-artweb/internal/service/mds"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

func newMdsRouter(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
	jobsvc *JobsRouter,
) {
	colonyRepo := mdsrepo.NewMdsColonyRepo(loggers.Data, init.DB, init.DBTimeout)
	nodeRepo := mdsrepo.NewMdsNodeRepo(loggers.Data, init.DB, init.DBTimeout)

	colonyService := mdssvc.NewMdsColonyService(loggers.Biz, colonyRepo)
	nodeService := mdssvc.NewMdsNodeService(loggers.Biz, nodeRepo)
	recordService := mdssvc.NewJobsService(loggers.Biz, jobsvc.Script, jobsvc.Record, jobsvc.Schedule)
	taskService := mdssvc.NewMdsTaskExecutionInfoUsecase(loggers.Biz, recordService)

	colonyHandler := handler.NewMdsColonyService(loggers.Service, colonyService, taskService)
	nodeHandler := handler.NewMdsNodeService(loggers.Service, nodeService)
	confHandler := handler.NewMdsConfService(loggers.Service, int64(init.Conf.Upload.MaxConfSize)*1024*1024)

	appRouter := router.Group("/v1/mds")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	colonyHandler.LoadRouter(appRouter)
	nodeHandler.LoadRouter(appRouter)
	confHandler.LoadRouter(appRouter)
}
