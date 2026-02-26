package routers

import (
	"context"

	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/jobs"
	jobsrepo "gin-artweb/internal/repository/jobs"
	jobsvc "gin-artweb/internal/service/jobs"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

type JobsRouter struct {
	Script   *jobsvc.ScriptService
	Record   *jobsvc.RecordService
	Schedule *jobsvc.ScheduleService
}

func NewJobsRouter(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
) *JobsRouter {
	scriptRepo := jobsrepo.NewScriptRepo(loggers.Data, init.DB, init.DBTimeout)
	recordRepo := jobsrepo.NewRecordRepo(loggers.Data, init.DB, init.DBTimeout)
	scheduleRepo := jobsrepo.NewScheduleRepo(loggers.Data, init.DB, init.DBTimeout)

	scriptService := jobsvc.NewScriptService(loggers.Biz, scriptRepo)
	recordService := jobsvc.NewScriptRecordService(loggers.Biz, scriptRepo, recordRepo)
	scheduleService := jobsvc.NewScheduleService(loggers.Biz, scriptRepo, scheduleRepo, recordService, init.Crontab)

	// 加载计划任务
	scheduleService.ReloadScheduleJobs(context.Background(), nil)

	scriptHandler := handler.NewScriptHandler(loggers.Service, scriptService, int64(init.Conf.Security.Upload.MaxScriptSize)*1024*1024)
	recordHandler := handler.NewScriptRecordHandler(loggers.Service, recordService)
	scheduleHandler := handler.NewScheduleHandler(loggers.Service, scheduleService)

	appRouter := router.Group("/v1/jobs")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	scriptHandler.LoadRouter(appRouter)
	recordHandler.LoadRouter(appRouter)
	scheduleHandler.LoadRouter(appRouter)

	return &JobsRouter{
		Script:   scriptService,
		Record:   recordService,
		Schedule: scheduleService,
	}
}
