package routers

import (
	golog "log"
	"path/filepath"

	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/mds"
	mdsmodel "gin-artweb/internal/model/mds"
	mdsrepo "gin-artweb/internal/repo/mds"
	mdssvc "gin-artweb/internal/service/mds"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/middleware"
	"gin-artweb/pkg/serializer"
)

func newMdsRouter(
	router *gin.RouterGroup,
	init *config.SystemInit,
	loggers *config.Loggers,
	jobsvc *JobServices,
) {
	var cronConf map[string]mdsmodel.MdsCronTask
	cronConfPath := filepath.Join(config.ResourceDir, "mds", "config", "mds_cron.yaml")
	if _, err := serializer.ReadYAML(cronConfPath, &cronConf); err != nil {
		golog.Fatalf("加载mds_cron.yaml失败: %v", err)
	}

	colonyRepo := mdsrepo.NewMdsColonyRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)
	nodeRepo := mdsrepo.NewMdsNodeRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)
	cronRepo := mdsrepo.NewMdsCronRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)

	cronService := mdssvc.NewMdsCronService(loggers.Service, jobsvc.Schedule, cronRepo, cronConf)
	colonyService := mdssvc.NewMdsColonyService(loggers.Service, colonyRepo, cronService)
	nodeService := mdssvc.NewMdsNodeService(loggers.Service, nodeRepo)
	taskService := mdssvc.NewMdsTaskService(loggers.Service, jobsvc.Record, colonyRepo)

	colonyHandler := handler.NewMdsColonyHandler(loggers.Handler, colonyService, taskService)
	nodeHandler := handler.NewMdsNodeHandler(loggers.Handler, nodeService)
	confHandler := handler.NewMdsConfHandler(loggers.Handler, int64(init.Conf.Upload.MaxConfSize)*1024*1024)

	appRouter := router.Group("/v1/mds")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Handler))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Handler))

	colonyHandler.LoadRouter(appRouter)
	nodeHandler.LoadRouter(appRouter)
	confHandler.LoadRouter(appRouter)
}
