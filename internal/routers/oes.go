package routers

import (
	"context"
	golog "log"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/oes"
	jobmodel "gin-artweb/internal/model/job"
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

	stkCronConf := newOesCronConf(jobsvc, "stk_cron.yaml")
	crdCronConf := newOesCronConf(jobsvc, "crd_cron.yaml")
	optCronConf := newOesCronConf(jobsvc, "opt_cron.yaml")

	colonyRepo := oesrepo.NewOesColonyRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)
	nodeRepo := oesrepo.NewOesNodeRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)
	cronRepo := oesrepo.NewOesCronRepo(loggers.Repo, init.DB, init.DBTimeout, init.DBSlowThreshold)

	cronService := oessvc.NewOesCronService(loggers.Service, jobsvc.Schedule, cronRepo, stkCronConf, crdCronConf, optCronConf)
	colonyService := oessvc.NewOesColonyService(loggers.Service, colonyRepo, cronService)
	nodeService := oessvc.NewOesNodeService(loggers.Service, nodeRepo)
	stkService := oessvc.NewStkTaskService(loggers.Service, jobsvc.Record, colonyRepo)
	crdService := oessvc.NewCrdTaskService(loggers.Service, jobsvc.Record, colonyRepo)
	optService := oessvc.NewOptTaskService(loggers.Service, jobsvc.Record, colonyRepo)

	colonyHandler := handler.NewOesColonyHandler(loggers.Handler, colonyService, stkService, crdService, optService)
	nodeHandler := handler.NewOesNodeHandler(loggers.Handler, nodeService)
	confHandler := handler.NewOesConfHandler(loggers.Handler, int64(init.Conf.Upload.MaxConfSize)*1024*1024)

	appRouter := router.Group("/v1/oes")
	appRouter.Use(middleware.JWTAuthMiddleware(init.JwtConf, loggers.Handler))
	appRouter.Use(middleware.CasbinAuthMiddleware(init.Enforcer, loggers.Handler))

	colonyHandler.LoadRouter(appRouter)
	nodeHandler.LoadRouter(appRouter)
	confHandler.LoadRouter(appRouter)
}

func newOesCronConf(jobsvc *JobServices, cronConfName string) map[string]oesmodel.OesCronTask {
	cronConfPath := filepath.Join(config.ResourceDir, "oes", "config", cronConfName)
	yamlConf := make(map[string]oesmodel.OesCronConf)
	if _, err := serializer.ReadYAML(cronConfPath, &yamlConf); err != nil {
		golog.Fatalf("加载 %s 失败: %v", cronConfPath, err)
	}
	labelSet := make(map[string]struct{})
	nameSet := make(map[string]struct{})
	for _, cronConf := range yamlConf {
		labelSet[cronConf.ScriptLabel] = struct{}{}
		nameSet[cronConf.ScriptName] = struct{}{}
	}
	scriptLabels := make([]string, 0, len(labelSet))
	scriptNames := make([]string, 0, len(nameSet))
	for label := range labelSet {
		scriptLabels = append(scriptLabels, label)
	}
	for name := range nameSet {
		scriptNames = append(scriptNames, name)
	}

	scriptStatus := true
	scriptBuiltin := true
	_, scripts, err := jobsvc.Script.ListScript(
		context.Background(), 1, 1000, jobmodel.ListScriptDTO{
			Project:   "oes",
			Names:     strings.Join(scriptNames, ","),
			Labels:    strings.Join(scriptLabels, ","),
			IsBuiltin: &scriptBuiltin,
			Status:    &scriptStatus,
		})
	if err != nil {
		golog.Fatalf("获取脚本列表失败: %v", err)
	}
	scriptMap := make(map[string]jobmodel.ScriptModel)
	for _, m := range scripts {
		key := m.Project + "_" + m.Label + "_" + m.Name
		scriptMap[key] = m
	}
	cronTasks := make(map[string]oesmodel.OesCronTask, len(yamlConf))
	for taskName, cronConf := range yamlConf {
		key := "oes_" + cronConf.ScriptLabel + "_" + cronConf.ScriptName
		script, ok := scriptMap[key]
		if !ok {
			golog.Fatalf("脚本 %s 不存在", key)
		}
		cronTasks[taskName] = oesmodel.OesCronTask{
			ScriptID:      script.ID,
			Specification: cronConf.Specification,
		}
	}
	return cronTasks
}
