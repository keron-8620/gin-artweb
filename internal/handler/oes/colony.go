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
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"绑定创建oes集群参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, rErr := s.colonySvc.CreateOesColony(ctx, req)
	if rErr != nil {
		log.Error(
			"创建oes集群失败",
			zap.Error(rErr),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

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
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"绑定更新oes集群ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req oesmodel.OesColonyUpsertDTO
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"绑定更新oes集群参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, rErr := s.colonySvc.UpdateOesColonyByID(ctx, uri.ID, req)
	if rErr != nil {
		log.Error(
			"更新oes集群失败",
			zap.Error(rErr),
			zap.Uint32("oes_colony_id", uri.ID),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

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
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"绑定删除oes集群ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"开始删除oes集群",
		zap.Uint32("oes_colony_id", uri.ID),
	)

	rErr := s.colonySvc.DeleteOesColonyByID(ctx, uri.ID)
	if rErr != nil {
		log.Error(
			"删除oes集群失败",
			zap.Error(rErr),
			zap.Uint32("oes_colony_id", uri.ID),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"删除oes集群成功",
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
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"绑定查询oes集群ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"开始查询oes集群详情",
		zap.Uint32("oes_colony_id", uri.ID),
	)

	preloads := []string{"Package", "XCounter", "MonNode"}
	m, rErr := s.colonySvc.FindOesColonyByID(ctx, preloads, uri.ID)
	if rErr != nil {
		log.Error(
			"查询oes集群详情失败",
			zap.Error(rErr),
			zap.Strings("preloads", preloads),
			zap.Uint32("oes_colony_id", uri.ID),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"查询oes集群详情成功",
		zap.Uint32("oes_colony_id", uri.ID),
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
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Error(
			"绑定查询oes集群列表参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info("开始查询oes集群列表")

	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := s.colonySvc.ListOesColony(ctx, page, size, req)
	if err != nil {
		log.Error(
			"查询oes集群列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询oes集群列表成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
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
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Error(
			"绑定查询oes集群列表参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	tasks, rErr := s.stkTaskSvc.BuildTaskExecutionInfos(ctx, req)
	if rErr != nil {
		log.Error(
			"构建oes现货集群任务信息失败",
			zap.Error(rErr),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

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
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Error(
			"绑定查询oes集群列表参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	tasks, rErr := s.crdTaskSvc.BuildTaskExecutionInfos(ctx, req)
	if rErr != nil {
		log.Error(
			"构建oes两融集群任务信息失败",
			zap.Error(rErr),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

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
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Error(
			"绑定查询oes集群列表参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	tasks, rErr := s.optTaskSvc.BuildTaskExecutionInfos(ctx, req)
	if rErr != nil {
		log.Error(
			"构建oes期权集群任务信息失败",
			zap.Error(rErr),
			zap.Object("oes_colony_dto", &req),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
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
