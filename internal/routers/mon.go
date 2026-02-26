package routers

import (
	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/mon"
	monrepo "gin-artweb/internal/repository/mon"
	monsvc "gin-artweb/internal/service/mon"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/log"
	"gin-artweb/internal/shared/middleware"
)

func newMonRouter(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *log.Loggers,
) {
	nodeRepo := monrepo.NewMonNodeRepo(loggers.Data, init.DB, init.DBTimeout)

	nodeService := monsvc.NewMonNodeService(loggers.Biz, nodeRepo)

	nodeHandler := handler.NewNodeHandler(loggers.Service, nodeService)

	appRouter := router.Group("/v1/mon")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	nodeHandler.LoadRouter(appRouter)
}
