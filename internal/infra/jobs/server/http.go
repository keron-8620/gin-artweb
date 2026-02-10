package server

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"gin-artweb/internal/infra/jobs/biz"
	"gin-artweb/internal/infra/jobs/data"
	"gin-artweb/internal/infra/jobs/service"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

type JobsUsecase struct {
	Script   *biz.ScriptUsecase
	Record   *biz.RecordUsecase
	Schedule *biz.ScheduleUsecase
}

func NewServer(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
) *JobsUsecase {
	jwtConf := auth.NewJWTConfig(
		time.Duration(init.Conf.Security.Token.AccessMinutes)*time.Minute,
		time.Duration(init.Conf.Security.Token.RefreshMinutes)*time.Minute,
		init.Conf.Security.Token.AccessMethod,
		init.Conf.Security.Token.RefreshMethod,
	)

	scriptRepo := data.NewScriptRepo(loggers.Data, init.DB, init.DBTimeout)
	recordRepo := data.NewRecordRepo(loggers.Data, init.DB, init.DBTimeout)
	scheduleRepo := data.NewScheduleRepo(loggers.Data, init.DB, init.DBTimeout)

	scriptUsecase := biz.NewScriptUsecase(loggers.Biz, scriptRepo)
	recordUsecase := biz.NewScriptRecordUsecase(loggers.Biz, scriptRepo, recordRepo)
	scheduleUsecase := biz.NewScheduleUsecase(loggers.Biz, scriptRepo, scheduleRepo, recordUsecase, init.Crontab)

	// 加载计划任务
	scheduleUsecase.ReloadScheduleJobs(context.Background(), nil)

	scriptService := service.NewScriptService(loggers.Service, scriptUsecase, int64(init.Conf.Security.Upload.MaxScriptSize)*1024*1024)
	recordService := service.NewScriptRecordService(loggers.Service, recordUsecase)
	scheduleService := service.NewScheduleService(loggers.Service, scheduleUsecase)

	appRouter := router.Group("/v1/jobs")
	appRouter.Use(middleware.JWTAuthMiddleware(jwtConf, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	scriptService.LoadRouter(appRouter)
	recordService.LoadRouter(appRouter)
	scheduleService.LoadRouter(appRouter)

	return &JobsUsecase{
		Script:   scriptUsecase,
		Record:   recordUsecase,
		Schedule: scheduleUsecase,
	}
}
