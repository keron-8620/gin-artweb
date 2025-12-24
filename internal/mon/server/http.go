package server

import (
	"github.com/gin-gonic/gin"

	"gin-artweb/internal/mon/biz"
	"gin-artweb/internal/mon/data"
	"gin-artweb/internal/mon/service"
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

	nodeRepo := data.NewMonNodeRepo(loggers.Data, init.DB, init.DBTimeout)

	nodeUsecase := biz.NewMonNodeUsecase(loggers.Biz, nodeRepo)

	nodeService := service.NewNodeService(loggers.Service, nodeUsecase)

	appRouter := router.Group("/v1/mon")
	appRouter.Use(middleware.JWTAuthMiddleware(init.Conf.Security.Token.SecretKey, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))
	nodeService.LoadRouter(appRouter)
}
