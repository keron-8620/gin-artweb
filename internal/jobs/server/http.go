package server

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"gin-artweb/internal/jobs/biz"
	"gin-artweb/internal/jobs/data"
	"gin-artweb/internal/jobs/service"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
)

func NewServer(
	router *gin.RouterGroup,
	conf *config.SystemConf,
	db *gorm.DB,
	dbTimeout *database.DBTimeout,
	loggers *log.Loggers,
	crontab *cron.Cron,
) {
	if err := dbAutoMigrate(db, loggers.Data); err != nil {
		panic(err)
	}

	scriptRepo := data.NewScriptRepo(loggers.Data, db, dbTimeout)
	recordRepo := data.NewRecordRepo(loggers.Data, db, dbTimeout)
	scheduleRepo := data.NewScheduleRepo(loggers.Data, db, dbTimeout)

	scriptUsecase := biz.NewScriptUsecase(loggers.Biz, scriptRepo)
	recordUsecase := biz.NewScriptRecordUsecase(loggers.Biz, scriptRepo, recordRepo)
	scheduleUsecase := biz.NewScheduleUsecase(loggers.Biz, scriptRepo, scheduleRepo, recordUsecase, crontab)

	scriptService := service.NewScriptService(loggers.Service, scriptUsecase, int64(conf.Security.Upload.MaxFileSize))
	recordService := service.NewScriptRecordService(loggers.Service, recordUsecase)
	scheduleService := service.NewScheduleService(loggers.Service, scheduleUsecase)

	appRouter := router.Group("/v1/jobs")
	scriptService.LoadRouter(appRouter)
	recordService.LoadRouter(appRouter)
	scheduleService.LoadRouter(appRouter)
}
