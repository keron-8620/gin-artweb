package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbColony "gin-artweb/api/oes/colony"
	svMon "gin-artweb/internal/business/mon/service"
	"gin-artweb/internal/business/oes/biz"
	jobsModel "gin-artweb/internal/infra/jobs/model"
	svReso "gin-artweb/internal/infra/resource/service"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type OesColonyService struct {
	log      *zap.Logger
	ucColony *biz.OesColonyUsecase
	ucStk    *biz.StkTaskExecutionInfoUsecase
	ucCrd    *biz.CrdTaskExecutionInfoUsecase
	ucOpt    *biz.OptTaskExecutionInfoUsecase
}

func NewOesColonyService(
	logger *zap.Logger,
	ucColony *biz.OesColonyUsecase,
	ucStk *biz.StkTaskExecutionInfoUsecase,
	ucCrd *biz.CrdTaskExecutionInfoUsecase,
	ucOpt *biz.OptTaskExecutionInfoUsecase,
) *OesColonyService {
	return &OesColonyService{
		log:      logger,
		ucColony: ucColony,
		ucStk:    ucStk,
		ucCrd:    ucCrd,
		ucOpt:    ucOpt,
	}
}

// @Summary 创建oes集群
// @Description 本接口用于创建新的oes集群
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request body pbColony.CreateOrUpdateOesColonyRequest true "创建oes集群请求"
// @Success 200 {object} pbColony.OesColonyReply "成功返回oes集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony [post]
// @Security ApiKeyAuth
func (s *OesColonyService) CreateOesColony(ctx *gin.Context) {
	var req pbColony.CreateOrUpdateOesColonyRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建oes集群参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	colony := biz.OesColonyModel{
		SystemType:    req.SystemType,
		ColonyNum:     req.ColonyNum,
		ExtractedName: req.ExtractedName,
		IsEnable:      req.IsEnable,
		MonNodeID:     req.MonNodeID,
		PackageID:     req.PackageID,
		XCounterID:    req.XCounterID,
	}

	m, rErr := s.ucColony.CreateOesColony(ctx, colony)
	if rErr != nil {
		s.log.Error(
			"创建oes集群失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &pbColony.OesColonyReply{
		Code: http.StatusOK,
		Data: *OesColonyToDetailOut(*m),
	})
}

// @Summary 更新oes集群
// @Description 本接口用于更新指定ID的oes集群
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param id path uint true "oes集群编号"
// @Param request body pbColony.CreateOrUpdateOesColonyRequest true "更新oes集群请求"
// @Success 200 {object} pbColony.OesColonyReply "成功返回oes集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/{id} [put]
// @Security ApiKeyAuth
func (s *OesColonyService) UpdateOesColony(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定更新oes集群ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req pbColony.CreateOrUpdateOesColonyRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新oes集群参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	data := map[string]any{
		"system_type":    req.SystemType,
		"colony_num":     req.ColonyNum,
		"extracted_name": req.ExtractedName,
		"is_enable":      req.IsEnable,
		"package_id":     req.PackageID,
		"xcounter_id":    req.XCounterID,
		"mon_node_id":    req.MonNodeID,
	}

	m, rErr := s.ucColony.UpdateOesColonyByID(ctx, uri.ID, data)
	if rErr != nil {
		s.log.Error(
			"更新oes集群失败",
			zap.Error(rErr),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &pbColony.OesColonyReply{
		Code: http.StatusOK,
		Data: *OesColonyToDetailOut(*m),
	})
}

// @Summary 删除oes集群
// @Description 本接口用于删除指定ID的oes集群
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param id path uint true "oes集群编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/{id} [delete]
// @Security ApiKeyAuth
func (s *OesColonyService) DeleteOesColony(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除oes集群ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始删除oes集群",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	rErr := s.ucColony.DeleteOesColonyByID(ctx, uri.ID)
	if rErr != nil {
		s.log.Error(
			"删除oes集群失败",
			zap.Error(rErr),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"删除oes集群成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询oes集群详情
// @Description 本接口用于查询指定ID的oes集群详情
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param id path uint true "oes集群编号"
// @Success 200 {object} pbColony.OesColonyReply "成功返回oes集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/{id} [get]
// @Security ApiKeyAuth
func (s *OesColonyService) GetOesColony(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询oes集群ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始查询oes集群详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := s.ucColony.FindOesColonyByID(ctx, []string{"Package", "XCounter", "MonNode"}, uri.ID)
	if rErr != nil {
		s.log.Error(
			"查询oes集群详情失败",
			zap.Error(rErr),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"查询oes集群详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := OesColonyToDetailOut(*m)
	ctx.JSON(http.StatusOK, &pbColony.OesColonyReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询oes集群列表
// @Description 本接口用于查询oes集群列表
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request query pbColony.ListOesColonyRequest false "查询参数"
// @Success 200 {object} pbColony.PagOesColonyReply "成功返回oes集群列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony [get]
// @Security ApiKeyAuth
func (s *OesColonyService) ListOesColony(ctx *gin.Context) {
	var req pbColony.ListOesColonyRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询oes集群列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始查询oes集群列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"Package", "XCounter", "MonNode"},
		IsCount:  true,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, err := s.ucColony.ListOesColony(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询oes集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	s.log.Info(
		"查询oes集群列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := ListOesColonyToDetailOut(ms)
	ctx.JSON(http.StatusOK, &pbColony.PagOesColonyReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

// @Summary 查询oes现货集群列表的任务状态
// @Description 本接口用于查询oes现货集群列表的任务状态
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request query pbColony.ListOesColonyRequest false "查询参数"
// @Success 200 {object} pbColony.ListOesTasksInfoReply "成功返回oes现货集群列表的任务状态"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/status/stk [get]
// @Security ApiKeyAuth
func (s *OesColonyService) ListStkTaskStatus(ctx *gin.Context) {
	var req pbColony.ListOesColonyRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询oes集群列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	page, size, query := req.Query()
	query["system_type"] = "STK"
	query["is_enable = ?"] = true
	qp := database.QueryParams{
		Preloads: nil,
		IsCount:  false,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"colony_num ASC"},
		Query:    query,
	}

	_, ms, err := s.ucColony.ListOesColony(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询oes现货集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	tasks, rErr := s.ucStk.BuildTaskExecutionInfos(ctx, *ms)
	if rErr != nil {
		s.log.Error(
			"构建oes现货集群任务信息失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	if tasks == nil || len(*tasks) == 0 {
		ctx.JSON(http.StatusOK, &pbColony.ListOesTasksInfoReply{
			Code: http.StatusOK,
			Data: []pbColony.OesColonyTaskInfo{},
		})
		return
	}

	infos := *tasks
	results := make([]pbColony.OesColonyTaskInfo, len(infos))
	for i, info := range infos {
		results[i] = BuildStkColonyTaskInfo(info)
	}

	ctx.JSON(http.StatusOK, &pbColony.ListOesTasksInfoReply{
		Code: http.StatusOK,
		Data: results,
	})
}

// @Summary 查询oes两融集群列表的任务状态
// @Description 本接口用于查询oes两融集群列表的任务状态
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request query pbColony.ListOesColonyRequest false "查询参数"
// @Success 200 {object} pbColony.ListOesTasksInfoReply "成功返回oes两融集群列表的任务状态"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/status/crd [get]
// @Security ApiKeyAuth
func (s *OesColonyService) ListCrdTaskStatus(ctx *gin.Context) {
	var req pbColony.ListOesColonyRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询oes集群列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	page, size, query := req.Query()
	query["system_type"] = "CRD"
	query["is_enable = ?"] = true
	qp := database.QueryParams{
		Preloads: nil,
		IsCount:  false,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"colony_num ASC"},
		Query:    query,
	}

	_, ms, err := s.ucColony.ListOesColony(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询oes两融集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	tasks, rErr := s.ucCrd.BuildTaskExecutionInfos(ctx, *ms)
	if rErr != nil {
		s.log.Error(
			"构建oes两融集群任务信息失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	if tasks == nil || len(*tasks) == 0 {
		ctx.JSON(http.StatusOK, &pbColony.ListOesTasksInfoReply{
			Code: http.StatusOK,
			Data: []pbColony.OesColonyTaskInfo{},
		})
		return
	}

	infos := *tasks
	results := make([]pbColony.OesColonyTaskInfo, len(infos))
	for i, info := range infos {
		results[i] = BuildCrdColonyTaskInfo(info)
	}

	ctx.JSON(http.StatusOK, &pbColony.ListOesTasksInfoReply{
		Code: http.StatusOK,
		Data: results,
	})
}

// @Summary 查询oes期权集群列表的任务状态
// @Description 本接口用于查询oes期权集群列表的任务状态
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request query pbColony.ListOesColonyRequest false "查询参数"
// @Success 200 {object} pbColony.ListOesTasksInfoReply "成功返回oes期权集群列表的任务状态"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/status/opt [get]
// @Security ApiKeyAuth
func (s *OesColonyService) ListOptTaskStatus(ctx *gin.Context) {
	var req pbColony.ListOesColonyRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询oes集群列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	page, size, query := req.Query()
	query["system_type"] = "OPT"
	query["is_enable = ?"] = true
	qp := database.QueryParams{
		Preloads: nil,
		IsCount:  false,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"colony_num ASC"},
		Query:    query,
	}

	_, ms, err := s.ucColony.ListOesColony(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询oes期权集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	tasks, rErr := s.ucOpt.BuildTaskExecutionInfos(ctx, *ms)
	if rErr != nil {
		s.log.Error(
			"构建oes期权集群任务信息失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	if tasks == nil || len(*tasks) == 0 {
		ctx.JSON(http.StatusOK, &pbColony.ListOesTasksInfoReply{
			Code: http.StatusOK,
			Data: []pbColony.OesColonyTaskInfo{},
		})
		return
	}

	infos := *tasks
	results := make([]pbColony.OesColonyTaskInfo, len(infos))
	for i, info := range infos {
		results[i] = BuildOptColonyTaskInfo(info)
	}

	ctx.JSON(http.StatusOK, &pbColony.ListOesTasksInfoReply{
		Code: http.StatusOK,
		Data: results,
	})
}

func (s *OesColonyService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/colony", s.CreateOesColony)
	r.PUT("/colony/:id", s.UpdateOesColony)
	r.DELETE("/colony/:id", s.DeleteOesColony)
	r.GET("/colony/:id", s.GetOesColony)
	r.GET("/colony", s.ListOesColony)
	r.GET("/colony/status/stk", s.ListStkTaskStatus)
	r.GET("/colony/status/crd", s.ListCrdTaskStatus)
	r.GET("/colony/status/opt", s.ListOptTaskStatus)
}

func OesColonyToBaseOut(
	m biz.OesColonyModel,
) *pbColony.OesColonyBaseOut {
	return &pbColony.OesColonyBaseOut{
		ID:            m.ID,
		SystemType:    m.SystemType,
		ColonyNum:     m.ColonyNum,
		ExtractedName: m.ExtractedName,
		IsEnable:      m.IsEnable,
	}
}

func OesColonyToStandardOut(
	m biz.OesColonyModel,
) *pbColony.OesColonyStandardOut {
	return &pbColony.OesColonyStandardOut{
		OesColonyBaseOut: *OesColonyToBaseOut(m),
		CreatedAt:        m.CreatedAt.Format(time.DateTime),
		UpdatedAt:        m.UpdatedAt.Format(time.DateTime),
	}
}

func OesColonyToDetailOut(
	m biz.OesColonyModel,
) *pbColony.OesColonyDetailOut {
	return &pbColony.OesColonyDetailOut{
		OesColonyStandardOut: *OesColonyToStandardOut(m),
		Package:              svReso.PackageModelToOutBase(m.Package),
		XCounter:             svReso.PackageModelToOutBase(m.XCounter),
		MonNode:              svMon.MonNodeToBaseOut(m.MonNode),
	}
}

func ListOesColonyToDetailOut(
	rms *[]biz.OesColonyModel,
) *[]pbColony.OesColonyDetailOut {
	if rms == nil {
		return &[]pbColony.OesColonyDetailOut{}
	}

	ms := *rms
	mso := make([]pbColony.OesColonyDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := OesColonyToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}

func BuildStkColonyTaskInfo(t biz.StkTaskExecutionInfo) pbColony.OesColonyTaskInfo {
	mon := BuildTaskInfoFromScriptRecord("mon", t.Mon)
	conterFetch := BuildTaskInfoFromScriptRecord("counter_fetch", t.CounterFetch)
	counterDistribute := BuildTaskInfoFromScriptRecord("counter_distribute", t.CounterDistribute)
	bse := BuildTaskInfoFromScriptRecord("bse", t.Bse)
	sse := BuildTaskInfoFromScriptRecord("sse", t.Sse)
	szse := BuildTaskInfoFromScriptRecord("szse", t.Szse)
	csdc := BuildTaskInfoFromScriptRecord("csdc", t.Csdc)
	return pbColony.OesColonyTaskInfo{
		ColonyNum: t.ColonyNum,
		Tasks:     []pbComm.TaskInfo{mon, conterFetch, counterDistribute, bse, sse, szse, csdc},
	}
}

func BuildCrdColonyTaskInfo(t biz.CrdTaskExecutionInfo) pbColony.OesColonyTaskInfo {
	mon := BuildTaskInfoFromScriptRecord("mon", t.Mon)
	conterFetch := BuildTaskInfoFromScriptRecord("counter_fetch", t.CounterFetch)
	counterDistribute := BuildTaskInfoFromScriptRecord("counter_distribute", t.CounterDistribute)
	sse := BuildTaskInfoFromScriptRecord("sse", t.Sse)
	szse := BuildTaskInfoFromScriptRecord("szse", t.Szse)
	csdc := BuildTaskInfoFromScriptRecord("csdc", t.Csdc)
	sseLate := BuildTaskInfoFromScriptRecord("sse_late", t.SseLate)
	szseLate := BuildTaskInfoFromScriptRecord("szse_late", t.SzseLate)
	return pbColony.OesColonyTaskInfo{
		ColonyNum: t.ColonyNum,
		Tasks:     []pbComm.TaskInfo{mon, conterFetch, counterDistribute, sse, szse, csdc, sseLate, szseLate},
	}
}

func BuildOptColonyTaskInfo(t biz.OptTaskExecutionInfo) pbColony.OesColonyTaskInfo {
	mon := BuildTaskInfoFromScriptRecord("mon", t.Mon)
	conterFetch := BuildTaskInfoFromScriptRecord("counter_fetch", t.CounterFetch)
	counterDistribute := BuildTaskInfoFromScriptRecord("counter_distribute", t.CounterDistribute)
	sse := BuildTaskInfoFromScriptRecord("sse", t.Sse)
	szse := BuildTaskInfoFromScriptRecord("szse", t.Szse)
	return pbColony.OesColonyTaskInfo{
		ColonyNum: t.ColonyNum,
		Tasks:     []pbComm.TaskInfo{mon, conterFetch, counterDistribute, sse, szse},
	}
}

func BuildTaskInfoFromScriptRecord(taskName string, m *jobsModel.ScriptRecordModel) pbComm.TaskInfo {
	result := pbComm.TaskInfo{
		TaskName: taskName,
	}
	if m != nil {
		result.RecordID = m.ID
		result.Status = m.Status
		result.StartTime = m.CreatedAt.Format(time.DateTime)
		result.EndTime = m.UpdatedAt.Format(time.DateTime)
		result.TriggerType = m.TriggerType
	}
	return result
}
