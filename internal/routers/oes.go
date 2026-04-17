package routers

import (
	golog "log"
	"path/filepath"

	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/oes"
	oesmodel "gin-artweb/internal/model/oes"
	oesrepo "gin-artweb/internal/repo/oes"
	oessvc "gin-artweb/internal/service/oes"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/middleware"
	"gin-artweb/pkg/serializer"
)

func newOesRouter(
	router *gin.RouterGroup,
	init *config.SystemInit,
	loggers *config.Loggers,
	jobsvc *JobServices,
) {
	var (
		stkCronConf map[string]oesmodel.OesCronTask
		crdCronConf map[string]oesmodel.OesCronTask
		optCronConf map[string]oesmodel.OesCronTask
	)
	cronConfDir := filepath.Join(config.ResourceDir, "oes", "config")
	if _, err := serializer.ReadYAML(filepath.Join(cronConfDir, "stk_cron.yaml"), &stkCronConf); err != nil {
		golog.Fatalf("加载stk_cron.yaml失败: %v", err)
	}
	if _, err := serializer.ReadYAML(filepath.Join(cronConfDir, "crd_cron.yaml"), &crdCronConf); err != nil {
		golog.Fatalf("加载crd_cron.yaml失败: %v", err)
	}
	if _, err := serializer.ReadYAML(filepath.Join(cronConfDir, "opt_cron.yaml"), &optCronConf); err != nil {
		golog.Fatalf("加载opt_cron.yaml失败: %v", err)
	}

	colonyRepo := oesrepo.NewOesColonyRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)
	nodeRepo := oesrepo.NewOesNodeRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)
	cronRepo := oesrepo.NewOesCronRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)

	cronService := oessvc.NewOesCronService(loggers.Service, jobsvc.Schedule, cronRepo, stkCronConf, crdCronConf, optCronConf)
	colonyService := oessvc.NewOesColonyService(loggers.Service, colonyRepo, cronService)
	nodeService := oessvc.NewOesNodeService(loggers.Service, nodeRepo)
	stkService := oessvc.NewStkTaskService(loggers.Service, jobsvc.Record, colonyRepo)
	crdService := oessvc.NewCrdTaskService(loggers.Service, jobsvc.Record, colonyRepo)
	optService := oessvc.NewOptTaskService(loggers.Service, jobsvc.Record, colonyRepo)

	colonyHandler := handler.NewOesColonyHandler(loggers.Service, colonyService, stkService, crdService, optService)
	nodeHandler := handler.NewOesNodeHandler(loggers.Service, nodeService)
	confHandler := handler.NewOesConfHandler(loggers.Service, int64(init.Conf.Upload.MaxConfSize)*1024*1024)

	appRouter := router.Group("/v1/oes")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Service))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Service))

	colonyHandler.LoadRouter(appRouter)
	nodeHandler.LoadRouter(appRouter)
	confHandler.LoadRouter(appRouter)
}
