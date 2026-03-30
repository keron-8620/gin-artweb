package oes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	jobmodel "gin-artweb/internal/model/job"
	oesmodel "gin-artweb/internal/model/oes"
	jobsvc "gin-artweb/internal/service/job"
	oessvc "gin-artweb/internal/service/oes"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

type OesColonyHandler struct {
	log        *zap.Logger
	colonySvc  *oessvc.OesColonyService
	stkTaskSvc *oessvc.StkTaskService
	crdTaskSvc *oessvc.CrdTaskService
	optTaskSvc *oessvc.OptTaskService
}

func NewOesColonyHandler(
	logger *zap.Logger,
	ucColony *oessvc.OesColonyService,
	stkTaskSvc *oessvc.StkTaskService,
	crdTaskSvc *oessvc.CrdTaskService,
	optTaskSvc *oessvc.OptTaskService,
) *OesColonyHandler {
	return &OesColonyHandler{
		log:        logger,
		colonySvc:  ucColony,
		stkTaskSvc: stkTaskSvc,
		crdTaskSvc: crdTaskSvc,
		optTaskSvc: optTaskSvc,
	}
}

// @Summary 创建oes集群
// @Description 本接口用于创建新的oes集群
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request body oesmodel.OesColonyUpsertDTO true "创建oes集群请求"
// @Success 200 {object} oesmodel.OesColonyResp "成功返回oes集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony [post]
// @Security ApiKeyAuth
func (s *OesColonyHandler) CreateOesColony(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var req oesmodel.OesColonyUpsertDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"创建oes集群：绑定创建oes集群参数失败") {
		return
	}

	log.Info("创建oes集群：开始执行")

	log.Debug(
		"创建oes集群：入参详情",
		zap.Object("oes_colony_dto", &req),
	)

	createStepStart := time.Now()
	createStepDuration := time.Since(createStepStart)
	m, rErr := s.colonySvc.CreateOesColony(ctx, req)
	createStepDuration = time.Since(createStepStart)
	if rErr != nil {
		log.Error(
			"创建oes集群：执行失败",
			zap.Error(rErr),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("create_step_duration", createStepDuration),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	log.Debug(
		"创建oes集群：创建后的oes集群模型详情",
		zap.Object("oes_colony_model", m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	log.Info(
		"创建oes集群：执行成功",
		zap.Uint32("oes_colony_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("total_time", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &oesmodel.OesColonyResp{
		Code: http.StatusOK,
		Data: *oesmodel.OesColonyToDetailOut(*m),
	})
}

// @Summary 更新oes集群
// @Description 本接口用于更新指定ID的oes集群
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param id path uint true "oes集群编号"
// @Param request body oesmodel.OesColonyUpsertDTO true "更新oes集群请求"
// @Success 200 {object} oesmodel.OesColonyResp "成功返回oes集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/{id} [put]
// @Security ApiKeyAuth
func (s *OesColonyHandler) UpdateOesColony(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"更新oes集群：绑定更新oes集群ID参数失败") {
		return
	}

	var req oesmodel.OesColonyUpsertDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"更新oes集群：绑定更新oes集群参数失败") {
		return
	}

	updateStepStart := time.Now()
	m, err := s.colonySvc.UpdateOesColonyByID(ctx, uri.ID, req)
	updateStepDuration := time.Since(updateStepStart)
	if err != nil {
		log.Error(
			"更新oes集群：执行失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", uri.ID),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("update_step_duration", updateStepDuration),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"更新oes集群：更新后的oes集群模型详情",
		zap.Object("oes_colony_model", m),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	log.Info(
		"更新oes集群：执行成功",
		zap.Uint32("oes_colony_id", uri.ID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("total_time", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &oesmodel.OesColonyResp{
		Code: http.StatusOK,
		Data: *oesmodel.OesColonyToDetailOut(*m),
	})
}

// @Summary 删除oes集群
// @Description 本接口用于删除指定ID的oes集群
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param id path uint true "oes集群编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/{id} [delete]
// @Security ApiKeyAuth
func (s *OesColonyHandler) DeleteOesColony(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"删除oes集群：绑定删除oes集群ID参数失败") {
		return
	}

	log.Info(
		"删除oes集群：开始执行",
		zap.Uint32("oes_colony_id", uri.ID),
	)
	deleteStepStart := time.Now()
	err := s.colonySvc.DeleteOesColonyByID(ctx, uri.ID)
	deleteStepDuration := time.Since(deleteStepStart)
	if err != nil {
		log.Error(
			"删除oes集群：执行失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", uri.ID),
			zap.Duration("delete_step_duration", deleteStepDuration),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"删除oes集群：执行成功",
		zap.Uint32("oes_colony_id", uri.ID),
		zap.Duration("total_time", time.Since(startTime)),
	)

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询oes集群详情
// @Description 本接口用于查询指定ID的oes集群详情
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param id path uint true "oes集群编号"
// @Success 200 {object} oesmodel.OesColonyResp "成功返回oes集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/{id} [get]
// @Security ApiKeyAuth
func (s *OesColonyHandler) GetOesColony(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"查询oes集群详情：绑定查询oes集群ID参数失败") {
		return
	}

	log.Info(
		"查询oes集群详情：开始执行",
		zap.Uint32("oes_colony_id", uri.ID),
	)

	findStepStart := time.Now()
	preloads := []string{"Package", "XCounter", "MonNode"}
	m, err := s.colonySvc.FindOesColonyByID(ctx, preloads, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询oes集群详情：执行失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("oes_colony_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询oes集群详情：执行成功",
		zap.Uint32("oes_colony_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_time", time.Since(startTime)),
	)

	mo := oesmodel.OesColonyToDetailOut(*m)
	ctx.JSON(http.StatusOK, &oesmodel.OesColonyResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询oes集群列表
// @Description 本接口用于查询oes集群列表
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request query oesmodel.ListOesColonyDTO false "查询参数"
// @Success 200 {object} oesmodel.PagOesColonyResp "成功返回oes集群列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony [get]
// @Security ApiKeyAuth
func (s *OesColonyHandler) ListOesColony(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var req oesmodel.ListOesColonyDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"查询oes集群列表：绑定查询oes集群列表参数失败") {
		return
	}

	log.Info("查询oes集群列表：开始执行")

	log.Debug(
		"查询oes集群列表：入参详情",
		zap.Object("oes_colony_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := s.colonySvc.ListOesColony(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询oes集群列表：执行失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Info(
		"查询oes集群列表：执行成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_time", time.Since(startTime)),
	)

	mbs := oesmodel.ListOesColonyToDetailOut(ms)
	ctx.JSON(http.StatusOK, &oesmodel.PagOesColonyResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 查询oes现货集群列表的任务状态
// @Description 本接口用于查询oes现货集群列表的任务状态
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request query oesmodel.ListOesColonyDTO false "查询参数"
// @Success 200 {object} oesmodel.ListOesTasksInfoResp "成功返回oes现货集群列表的任务状态"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/status/stk [get]
// @Security ApiKeyAuth
func (s *OesColonyHandler) ListStkTaskStatus(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var req oesmodel.ListOesColonyDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"查询oes现货集群列表的任务状态：绑定查询oes现货集群列表参数失败") {
		return
	}

	log.Info("查询oes现货集群列表的任务状态：开始执行")

	log.Debug(
		"查询oes现货集群列表的任务状态：入参详情",
		zap.Object("oes_colony_dto", &req),
	)

	buildStepStart := time.Now()
	tasks, err := s.stkTaskSvc.BuildTaskExecutionInfos(ctx, req)
	buildStepDuration := time.Since(buildStepStart)
	if err != nil {
		log.Error(
			"查询oes现货集群列表的任务状态：执行失败",
			zap.Error(err),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("build_step_duration", buildStepDuration),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询oes现货集群列表的任务状态：执行成功",
		zap.Duration("build_step_duration", buildStepDuration),
		zap.Duration("total_time", time.Since(startTime)),
	)

	results := make([]oesmodel.OesColonyTaskInfo, len(tasks))
	for i, task := range tasks {
		results[i] = BuildStkColonyTaskInfo(task)
	}

	ctx.JSON(http.StatusOK, &oesmodel.ListOesTasksInfoResp{
		Code: http.StatusOK,
		Data: results,
	})
}

// @Summary 查询oes两融集群列表的任务状态
// @Description 本接口用于查询oes两融集群列表的任务状态
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request query oesmodel.ListOesColonyDTO false "查询参数"
// @Success 200 {object} oesmodel.ListOesTasksInfoResp "成功返回oes两融集群列表的任务状态"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/status/crd [get]
// @Security ApiKeyAuth
func (s *OesColonyHandler) ListCrdTaskStatus(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var req oesmodel.ListOesColonyDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"查询oes两融集群列表的任务状态：绑定查询oes两融集群列表参数失败") {
		return
	}

	log.Info("查询oes两融集群列表的任务状态：开始执行")

	log.Debug(
		"查询oes两融集群列表的任务状态：入参详情",
		zap.Object("oes_colony_dto", &req),
	)

	buildStepStart := time.Now()
	tasks, err := s.crdTaskSvc.BuildTaskExecutionInfos(ctx, req)
	buildStepDuration := time.Since(buildStepStart)
	if err != nil {
		log.Error(
			"查询oes两融集群列表的任务状态：执行失败",
			zap.Error(err),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("build_step_duration", buildStepDuration),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询oes两融集群列表的任务状态：执行成功",
		zap.Duration("build_step_duration", buildStepDuration),
		zap.Duration("total_time", time.Since(startTime)),
	)

	results := make([]oesmodel.OesColonyTaskInfo, len(tasks))
	for i, task := range tasks {
		results[i] = BuildCrdColonyTaskInfo(task)
	}

	ctx.JSON(http.StatusOK, &oesmodel.ListOesTasksInfoResp{
		Code: http.StatusOK,
		Data: results,
	})
}

// @Summary 查询oes期权集群列表的任务状态
// @Description 本接口用于查询oes期权集群列表的任务状态
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request query oesmodel.ListOesColonyDTO false "查询参数"
// @Success 200 {object} oesmodel.ListOesTasksInfoResp "成功返回oes期权集群列表的任务状态"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/status/opt [get]
// @Security ApiKeyAuth
func (s *OesColonyHandler) ListOptTaskStatus(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var req oesmodel.ListOesColonyDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"查询oes期权集群列表的任务状态：绑定查询oes期权集群列表参数失败") {
		return
	}

	log.Info("查询oes期权集群列表的任务状态：开始执行")

	log.Debug(
		"查询oes期权集群列表的任务状态：入参详情",
		zap.Object("oes_colony_dto", &req),
	)

	buildStepStart := time.Now()
	tasks, err := s.optTaskSvc.BuildTaskExecutionInfos(ctx, req)
	buildStepDuration := time.Since(buildStepStart)
	if err != nil {
		log.Error(
			"查询oes期权集群列表的任务状态：执行失败",
			zap.Error(err),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("build_step_duration", buildStepDuration),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询oes期权集群列表的任务状态：执行成功",
		zap.Duration("build_step_duration", buildStepDuration),
		zap.Duration("total_time", time.Since(startTime)),
	)

	results := make([]oesmodel.OesColonyTaskInfo, len(tasks))
	for i, task := range tasks {
		results[i] = BuildOptColonyTaskInfo(task)
	}

	ctx.JSON(http.StatusOK, &oesmodel.ListOesTasksInfoResp{
		Code: http.StatusOK,
		Data: results,
	})
}

func (s *OesColonyHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/colony", s.CreateOesColony)
	r.PUT("/colony/:id", s.UpdateOesColony)
	r.DELETE("/colony/:id", s.DeleteOesColony)
	r.GET("/colony/:id", s.GetOesColony)
	r.GET("/colony", s.ListOesColony)
	r.GET("/colony/status/stk", s.ListStkTaskStatus)
	r.GET("/colony/status/crd", s.ListCrdTaskStatus)
	r.GET("/colony/status/opt", s.ListOptTaskStatus)
}

func BuildStkColonyTaskInfo(t oesmodel.StkColonyTaskExecutionInfo) oesmodel.OesColonyTaskInfo {
	mon := jobsvc.BuildTaskInfoFromScriptRecord("mon", t.Mon)
	conterFetch := jobsvc.BuildTaskInfoFromScriptRecord("counter_fetch", t.CounterFetch)
	counterDistribute := jobsvc.BuildTaskInfoFromScriptRecord("counter_distribute", t.CounterDistribute)
	bse := jobsvc.BuildTaskInfoFromScriptRecord("bse", t.Bse)
	sse := jobsvc.BuildTaskInfoFromScriptRecord("sse", t.Sse)
	szse := jobsvc.BuildTaskInfoFromScriptRecord("szse", t.Szse)
	csdc := jobsvc.BuildTaskInfoFromScriptRecord("csdc", t.Csdc)
	return oesmodel.OesColonyTaskInfo{
		ColonyNum: t.ColonyNum,
		Tasks:     []jobmodel.BizTaskInfo{mon, conterFetch, counterDistribute, bse, sse, szse, csdc},
	}
}

func BuildCrdColonyTaskInfo(t oesmodel.CrdColonyTaskExecutionInfo) oesmodel.OesColonyTaskInfo {
	mon := jobsvc.BuildTaskInfoFromScriptRecord("mon", t.Mon)
	conterFetch := jobsvc.BuildTaskInfoFromScriptRecord("counter_fetch", t.CounterFetch)
	counterDistribute := jobsvc.BuildTaskInfoFromScriptRecord("counter_distribute", t.CounterDistribute)
	sse := jobsvc.BuildTaskInfoFromScriptRecord("sse", t.Sse)
	szse := jobsvc.BuildTaskInfoFromScriptRecord("szse", t.Szse)
	csdc := jobsvc.BuildTaskInfoFromScriptRecord("csdc", t.Csdc)
	sseLate := jobsvc.BuildTaskInfoFromScriptRecord("sse_late", t.SseLate)
	szseLate := jobsvc.BuildTaskInfoFromScriptRecord("szse_late", t.SzseLate)
	return oesmodel.OesColonyTaskInfo{
		ColonyNum: t.ColonyNum,
		Tasks:     []jobmodel.BizTaskInfo{mon, conterFetch, counterDistribute, sse, szse, csdc, sseLate, szseLate},
	}
}

func BuildOptColonyTaskInfo(t oesmodel.OptColonyTaskExecutionInfo) oesmodel.OesColonyTaskInfo {
	mon := jobsvc.BuildTaskInfoFromScriptRecord("mon", t.Mon)
	conterFetch := jobsvc.BuildTaskInfoFromScriptRecord("counter_fetch", t.CounterFetch)
	counterDistribute := jobsvc.BuildTaskInfoFromScriptRecord("counter_distribute", t.CounterDistribute)
	sse := jobsvc.BuildTaskInfoFromScriptRecord("sse", t.Sse)
	szse := jobsvc.BuildTaskInfoFromScriptRecord("szse", t.Szse)
	return oesmodel.OesColonyTaskInfo{
		ColonyNum: t.ColonyNum,
		Tasks:     []jobmodel.BizTaskInfo{mon, conterFetch, counterDistribute, sse, szse},
	}
}
