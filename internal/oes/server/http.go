package server

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-artweb/internal/oes/biz"
	"gin-artweb/internal/oes/data"
	"gin-artweb/internal/oes/service"
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
) {
	if err := dbAutoMigrate(db, loggers.Data); err != nil {
		panic(err)
	}

	colonyRepo := data.NewOesColonyRepo(loggers.Data, db, dbTimeout)
	nodeRepo := data.NewOesNodeRepo(loggers.Data, db, dbTimeout)

	colonyUsecase := biz.NewOesColonyUsecase(loggers.Biz, colonyRepo)
	nodeUsecase := biz.NewOesNodeUsecase(loggers.Biz, nodeRepo)

	colonyService := service.NewOesColonyService(loggers.Service, colonyUsecase)
	nodeService := service.NewOesNodeService(loggers.Service, nodeUsecase)

	appRouter := router.Group("/v1/oes")
	colonyService.LoadRouter(appRouter)
	nodeService.LoadRouter(appRouter)
}
