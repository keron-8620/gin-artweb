package routers

import (
	"context"

	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/job"
	jobrepo "gin-artweb/internal/repo/job"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/middleware"
)

type JobServices struct {
	Script   *jobsvc.ScriptService
	Record   *jobsvc.RecordService
	Schedule *jobsvc.ScheduleService
}

func NewJobRouter(
	router *gin.RouterGroup,
	init *config.SystemInit,
	loggers *config.Loggers,
) *JobServices {
	scriptRepo := jobrepo.NewScriptRepo(loggers.Data, init.DB, init.DBTimeout)
	recordRepo := jobrepo.NewRecordRepo(loggers.Data, init.DB, init.DBTimeout)
	scheduleRepo := jobrepo.NewScheduleRepo(loggers.Data, init.DB, init.DBTimeout)

	scriptService := jobsvc.NewScriptService(loggers.Service, scriptRepo)
	recordService := jobsvc.NewScriptRecordService(loggers.Service, scriptRepo, recordRepo)
	scheduleService := jobsvc.NewScheduleService(loggers.Service, scriptRepo, scheduleRepo, recordService, init.Crontab)

	// 加载计划任务
	scheduleService.ReLoadSchedule(context.Background(), map[string]any{"IsEnabled": true})

	scriptHandler := handler.NewScriptHandler(loggers.Handler, scriptService, int64(init.Conf.Upload.MaxScriptSize)*1024*1024)
	recordHandler := handler.NewScriptRecordHandler(loggers.Handler, recordService)
	scheduleHandler := handler.NewScheduleHandler(loggers.Handler, scheduleService)

	appRouter := router.Group("/v1/jobs")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Handler))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Handler))

	scriptHandler.LoadRouter(appRouter)
	recordHandler.LoadRouter(appRouter)
	scheduleHandler.LoadRouter(appRouter)

	return &JobServices{
		Script:   scriptService,
		Record:   recordService,
		Schedule: scheduleService,
	}
}
