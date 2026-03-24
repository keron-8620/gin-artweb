package routers

import (
	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/mon"
	monrepo "gin-artweb/internal/repo/mon"
	monsvc "gin-artweb/internal/service/mon"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/middleware"
)

func newMonRouter(
	router *gin.RouterGroup,
	init *common.Initialize,
	loggers *common.Loggers,
) {
	nodeRepo := monrepo.NewMonNodeRepo(loggers.Data, init.DB, init.DBTimeout)

	nodeService := monsvc.NewMonNodeService(loggers.Service, nodeRepo)

	nodeHandler := handler.NewNodeHandler(loggers.Handler, nodeService)

	appRouter := router.Group("/v1/mon")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Handler))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Handler))

	nodeHandler.LoadRouter(appRouter)
}
