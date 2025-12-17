package server

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-artweb/internal/mds/biz"
	"gin-artweb/internal/mds/data"
	"gin-artweb/internal/mds/service"
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

	colonyRepo := data.NewMdsColonyRepo(loggers.Data, db, dbTimeout)
	nodeRepo := data.NewMdsNodeRepo(loggers.Data, db, dbTimeout)

	colonyUsecase := biz.NewMdsColonyUsecase(loggers.Biz, colonyRepo)
	nodeUsecase := biz.NewMdsNodeUsecase(loggers.Biz, nodeRepo)

	colonyService := service.NewMdsColonyService(loggers.Service, colonyUsecase)
	nodeService := service.NewMdsNodeService(loggers.Service, nodeUsecase)

	appRouter := router.Group("/v1/mds")
	colonyService.LoadRouter(appRouter)
	nodeService.LoadRouter(appRouter)
}
