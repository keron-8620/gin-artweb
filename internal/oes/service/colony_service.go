package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbColony "gin-artweb/api/oes/colony"
	svMon "gin-artweb/internal/mon/service"
	"gin-artweb/internal/oes/biz"
	svReso "gin-artweb/internal/resource/service"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

type OesColonyService struct {
	log      *zap.Logger
	ucColony *biz.OesColonyUsecase
	ucTask   *biz.OesTaskInfoUsecase
}

func NewOesColonyService(
	logger *zap.Logger,
	ucColony *biz.OesColonyUsecase,
	ucTask *biz.OesTaskInfoUsecase,
) *OesColonyService {
	return &OesColonyService{
		log:      logger,
		ucColony: ucColony,
		ucTask:   ucTask,
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	colony := biz.OesColonyModel{
		SystemType:    req.SystemType,
		ColonyNum:     req.ColonyNum,
		ExtractedName: req.ExtractedName,
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbColony.CreateOrUpdateOesColonyRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新oes集群参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	data := map[string]any{
		"system_type":    req.SystemType,
		"colony_num":     req.ColonyNum,
		"extracted_name": req.ExtractedName,
		"package_id":     req.PackageID,
		"xcounter_id":    req.XCounterID,
		"mon_node_id":    req.MonNodeID,
	}

	m, err := s.ucColony.UpdateOesColonyByID(ctx, uri.ID, data)
	if err != nil {
		s.log.Error(
			"更新oes集群失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除oes集群",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := s.ucColony.DeleteOesColonyByID(ctx, uri.ID); err != nil {
		s.log.Error(
			"删除oes集群失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除oes集群成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询oes集群详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucColony.FindOesColonyByID(ctx, []string{"Package", "XCounter", "MonNode"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询oes集群详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询oes集群详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询oes集群列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
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
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询oes集群列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	mbs := ListOesColonyToDetailOut(ms)
	ctx.JSON(http.StatusOK, &pbColony.PagOesColonyReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

// @Summary 查询oes集群列表的任务状态
// @Description 本接口用于查询oes集群列表的任务状态
// @Tags oes集群管理
// @Accept json
// @Produce json
// @Param request query pbColony.ListOesColonyRequest false "查询参数"
// @Success 200 {object} pbColony.ListOesTasksInfoReply "成功返回oes集群列表的任务状态"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/colony/status [get]
// @Security ApiKeyAuth
func (s *OesColonyService) ListOesTaskStatus(ctx *gin.Context) {
	var req pbColony.ListOesColonyRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询oes集群列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	page, size, query := req.Query()
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
			"查询oes集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	if ms == nil {
		s.log.Warn(
			"查询oes集群列表为nil",
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.JSON(biz.ErrOesColonyListEmpty.Code, biz.ErrOesColonyListEmpty.ToMap())
		return
	}

	oesModels := *ms
	tasks := make([]pbColony.OesColonyTaskInfo, len(oesModels))
	for i, m := range oesModels {
		taskInfo, err := s.ucTask.GetColonyTaskInfo(ctx, m.ColonyNum, m.SystemType)
		if err != nil {
			s.log.Error(
				"查询oes集群任务状态失败",
				zap.Error(err),
				zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
				zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
			)
			ctx.AbortWithStatusJSON(err.Code, err.ToMap())
			return
		}
		tasks[i] = *taskInfo
	}

	ctx.JSON(http.StatusOK, &pbColony.ListOesTasksInfoReply{
		Code: http.StatusOK,
		Data: tasks,
	})
}

func (s *OesColonyService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/colony", s.CreateOesColony)
	r.PUT("/colony/:id", s.UpdateOesColony)
	r.DELETE("/colony/:id", s.DeleteOesColony)
	r.GET("/colony/:id", s.GetOesColony)
	r.GET("/colony", s.ListOesColony)
	r.GET("/colony/status", s.ListOesTaskStatus)
}

func OesColonyToBaseOut(
	m biz.OesColonyModel,
) *pbColony.OesColonyBaseOut {
	return &pbColony.OesColonyBaseOut{
		ID:            m.ID,
		SystemType:    m.SystemType,
		ColonyNum:     m.ColonyNum,
		ExtractedName: m.ExtractedName,
	}
}

func OesColonyToStandardOut(
	m biz.OesColonyModel,
) *pbColony.OesColonyStandardOut {
	return &pbColony.OesColonyStandardOut{
		OesColonyBaseOut: *OesColonyToBaseOut(m),
		CreatedAt:        m.CreatedAt.String(),
		UpdatedAt:        m.UpdatedAt.String(),
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
