package routers

import (
	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/mon"
	monrepo "gin-artweb/internal/repo/mon"
	monsvc "gin-artweb/internal/service/mon"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/middleware"
)

func newMonRouter(
	router *gin.RouterGroup,
	init *config.SystemInit,
	loggers *config.Loggers,
) {
	nodeRepo := monrepo.NewMonNodeRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)

	nodeService := monsvc.NewMonNodeService(loggers.Service, nodeRepo)

	nodeHandler := handler.NewNodeHandler(loggers.Handler, nodeService)

	appRouter := router.Group("/v1/mon")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Handler))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Handler))

	nodeHandler.LoadRouter(appRouter)
}
