package server

import (
	"github.com/gin-gonic/gin"

	"gin-artweb/internal/jobs/biz"
	"gin-artweb/internal/jobs/data"
	"gin-artweb/internal/jobs/service"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

func NewServer(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
) {
	if err := dbAutoMigrate(init.DB, loggers.Data); err != nil {
		panic(err)
	}

	scriptRepo := data.NewScriptRepo(loggers.Data, init.DB, init.DBTimeout)
	recordRepo := data.NewRecordRepo(loggers.Data, init.DB, init.DBTimeout)
	scheduleRepo := data.NewScheduleRepo(loggers.Data, init.DB, init.DBTimeout)

	scriptUsecase := biz.NewScriptUsecase(loggers.Biz, scriptRepo)
	recordUsecase := biz.NewScriptRecordUsecase(loggers.Biz, scriptRepo, recordRepo)
	scheduleUsecase := biz.NewScheduleUsecase(loggers.Biz, scriptRepo, scheduleRepo, recordUsecase, init.Crontab)

	scriptService := service.NewScriptService(loggers.Service, scriptUsecase, int64(init.Conf.Security.Upload.MaxFileSize))
	recordService := service.NewScriptRecordService(loggers.Service, recordUsecase)
	scheduleService := service.NewScheduleService(loggers.Service, scheduleUsecase)

	appRouter := router.Group("/v1/jobs")
	appRouter.Use(middleware.JWTAuthMiddleware(init.Conf.Security.Token.SecretKey, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	scriptService.LoadRouter(appRouter)
	recordService.LoadRouter(appRouter)
	scheduleService.LoadRouter(appRouter)
}
