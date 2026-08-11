package routers

import (
	"context"
	golog "log"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	handler "gin-artweb/internal/handler/mds"
	commodel "gin-artweb/internal/model/common"
	jobmodel "gin-artweb/internal/model/job"
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
	cronConf := newMdsCronConf(jobsvc)

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

func newMdsCronConf(jobsvc *JobServices) map[string]mdsmodel.MdsCronTask {
	yamlConf := make(map[string]mdsmodel.MdsCronConf)
	cronConfPath := filepath.Join(config.ResourceDir, "mds", "config", "mds_cron.yaml")
	if _, err := serializer.ReadYAML(cronConfPath, &yamlConf); err != nil {
		golog.Fatalf("加载mds_cron.yaml失败: %v", err)
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
	_, _, _, scripts, err := jobsvc.Script.ListScript(
		context.Background(), jobmodel.ListScriptDTO{
			StandardModelQuery: commodel.StandardModelQuery{
				BaseModelQuery: commodel.BaseModelQuery{
					Page: 1,
					Size: 1000,
				},
			},
			Project:   "mds",
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
	cronTasks := make(map[string]mdsmodel.MdsCronTask, len(yamlConf))
	for taskName, cronConf := range yamlConf {
		key := "mds_" + cronConf.ScriptLabel + "_" + cronConf.ScriptName
		script, ok := scriptMap[key]
		if !ok {
			golog.Fatalf("脚本 %s 不存在", key)
		}
		cronTasks[taskName] = mdsmodel.MdsCronTask{
			ScriptID:      script.ID,
			Specification: cronConf.Specification,
		}
	}
	return cronTasks
}
