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
func (s *OesColonyHandler) CreateOesColony(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("创建oes集群:开始执行")

	var req oesmodel.OesColonyUpsertDTO
	if !common.ShouldBind(c, log, &req, "创建oes集群:绑定创建oes集群参数失败") {
		return
	}

	m, rErr := s.colonySvc.CreateOesColony(ctx, req)
	if rErr != nil {
		log.Error(
			"创建oes集群:执行失败",
			zap.Error(rErr),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	log.Info(
		"创建oes集群:执行成功",
		zap.Uint32("oes_colony_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &oesmodel.OesColonyResp{
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
func (s *OesColonyHandler) UpdateOesColony(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("更新oes集群:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "更新oes集群:绑定更新oes集群ID参数失败") {
		return
	}

	var req oesmodel.OesColonyUpsertDTO
	if !common.ShouldBind(c, log, &req, "更新oes集群:绑定更新oes集群参数失败") {
		return
	}

	m, err := s.colonySvc.UpdateOesColonyByID(ctx, uri.ID, req)
	if err != nil {
		log.Error(
			"更新oes集群:执行失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", uri.ID),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"更新oes集群:执行成功",
		zap.Uint32("oes_colony_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &oesmodel.OesColonyResp{
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
func (s *OesColonyHandler) DeleteOesColony(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("删除oes集群:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "删除oes集群:绑定删除oes集群ID参数失败") {
		return
	}

	err := s.colonySvc.DeleteOesColonyByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"删除oes集群:执行失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"删除oes集群:执行成功",
		zap.Uint32("oes_colony_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
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
func (s *OesColonyHandler) GetOesColony(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询oes集群详情:绑定查询oes集群ID参数失败") {
		return
	}

	preloads := []string{"Package", "XCounter", "MonNode"}
	m, err := s.colonySvc.FindOesColonyByID(ctx, preloads, uri.ID)
	if err != nil {
		log.Error(
			"查询oes集群详情:执行失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("oes_colony_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mo := oesmodel.OesColonyToDetailOut(*m)
	c.JSON(http.StatusOK, &oesmodel.OesColonyResp{
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
func (s *OesColonyHandler) ListOesColony(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var req oesmodel.ListOesColonyDTO
	if !common.ShouldBindQuery(c, log, &req, "查询oes集群列表:绑定查询oes集群列表参数失败") {
		return
	}

	page, size, total, ms, err := s.colonySvc.ListOesColony(ctx, req)
	if err != nil {
		log.Error(
			"查询oes集群列表:执行失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mbs := oesmodel.ListOesColonyToDetailOut(ms)
	c.JSON(http.StatusOK, &oesmodel.PagOesColonyResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 查询mds集群计划任务列表
// @Description 本接口用于查询指定ID的mds集群计划任务列表
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param id path uint true "mds集群编号"
// @Success 200 {object} jobmodel.PagScheduleResp "成功返回mds计划任务列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/{id}/schedule [get]
// @Security ApiKeyAuth
func (s *OesColonyHandler) ListOesSchedules(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询oes计划任务:绑定查询oes集群ID参数失败") {
		return
	}

	schedules, err := s.colonySvc.ListOesSchedules(ctx, uri.ID)
	if err != nil {
		log.Error(
			"查询oes计划任务:执行失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	num := len(schedules)
	mos := jobmodel.ListScheduledToDetailOut(schedules)
	c.JSON(http.StatusOK, &jobmodel.PagScheduleResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(1, num, int64(num), mos),
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
func (s *OesColonyHandler) ListStkTaskStatus(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var req oesmodel.ListOesColonyDTO
	if !common.ShouldBindQuery(c, log, &req, "查询oes现货集群列表的任务状态:绑定查询oes现货集群列表参数失败") {
		return
	}
	req.SystemType = "STK"

	tasks, err := s.stkTaskSvc.BuildTaskExecutionInfos(ctx, req)
	if err != nil {
		log.Error(
			"查询oes现货集群列表的任务状态:执行失败",
			zap.Error(err),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	results := make([]oesmodel.OesColonyTaskInfo, len(tasks))
	for i, task := range tasks {
		results[i] = BuildStkColonyTaskInfo(task)
	}
	c.JSON(http.StatusOK, &oesmodel.ListOesTasksInfoResp{
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
func (s *OesColonyHandler) ListCrdTaskStatus(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var req oesmodel.ListOesColonyDTO
	if !common.ShouldBindQuery(c, log, &req, "查询oes两融集群列表的任务状态:绑定查询oes两融集群列表参数失败") {
		return
	}
	req.SystemType = "CRD"

	tasks, err := s.crdTaskSvc.BuildTaskExecutionInfos(ctx, req)
	if err != nil {
		log.Error(
			"查询oes两融集群列表的任务状态:执行失败",
			zap.Error(err),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	results := make([]oesmodel.OesColonyTaskInfo, len(tasks))
	for i, task := range tasks {
		results[i] = BuildCrdColonyTaskInfo(task)
	}

	c.JSON(http.StatusOK, &oesmodel.ListOesTasksInfoResp{
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
func (s *OesColonyHandler) ListOptTaskStatus(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var req oesmodel.ListOesColonyDTO
	if !common.ShouldBindQuery(c, log, &req, "查询oes期权集群列表的任务状态:绑定查询oes期权集群列表参数失败") {
		return
	}
	req.SystemType = "OPT"

	tasks, err := s.optTaskSvc.BuildTaskExecutionInfos(ctx, req)
	if err != nil {
		log.Error(
			"查询oes期权集群列表的任务状态:执行失败",
			zap.Error(err),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	results := make([]oesmodel.OesColonyTaskInfo, len(tasks))
	for i, task := range tasks {
		results[i] = BuildOptColonyTaskInfo(task)
	}

	c.JSON(http.StatusOK, &oesmodel.ListOesTasksInfoResp{
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
	r.GET("/colony/:id/schedule", s.ListOesSchedules)
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
