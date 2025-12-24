package server

import (
	"github.com/gin-gonic/gin"

	"gin-artweb/internal/mds/biz"
	"gin-artweb/internal/mds/data"
	"gin-artweb/internal/mds/service"
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

	colonyRepo := data.NewMdsColonyRepo(loggers.Data, init.DB, init.DBTimeout)
	nodeRepo := data.NewMdsNodeRepo(loggers.Data, init.DB, init.DBTimeout)

	colonyUsecase := biz.NewMdsColonyUsecase(loggers.Biz, colonyRepo)
	nodeUsecase := biz.NewMdsNodeUsecase(loggers.Biz, nodeRepo)

	colonyService := service.NewMdsColonyService(loggers.Service, colonyUsecase)
	nodeService := service.NewMdsNodeService(loggers.Service, nodeUsecase)

	appRouter := router.Group("/v1/mds")
	appRouter.Use(middleware.JWTAuthMiddleware(init.Conf.Security.Token.SecretKey, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))
	colonyService.LoadRouter(appRouter)
	nodeService.LoadRouter(appRouter)
}
