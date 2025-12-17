package server

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-artweb/internal/mon/biz"
	"gin-artweb/internal/mon/data"
	"gin-artweb/internal/mon/service"
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

	nodeRepo := data.NewMonNodeRepo(loggers.Data, db, dbTimeout)

	nodeUsecase := biz.NewMonNodeUsecase(loggers.Biz, nodeRepo)

	nodeService := service.NewNodeService(loggers.Service, nodeUsecase)

	appRouter := router.Group("/v1/mon")
	nodeService.LoadRouter(appRouter)
}
