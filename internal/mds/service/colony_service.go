package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbColony "gin-artweb/api/mds/colony"
	"gin-artweb/internal/mds/biz"
	svMon "gin-artweb/internal/mon/service"
	svReso "gin-artweb/internal/resource/service"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type MdsColonyService struct {
	log      *zap.Logger
	ucColony *biz.MdsColonyUsecase
}

func NewMdsColonyService(
	logger *zap.Logger,
	ucColony *biz.MdsColonyUsecase,
) *MdsColonyService {
	return &MdsColonyService{
		log:      logger,
		ucColony: ucColony,
	}
}

// @Summary 创建mds集群
// @Description 本接口用于创建新的mds集群
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param request body pbColony.CreateOrUpdateMdsColonyRequest true "创建mds集群请求"
// @Success 200 {object} pbColony.MdsColonyReply "成功返回mds集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony [post]
// @Security ApiKeyAuth
func (s *MdsColonyService) CreateMdsColony(ctx *gin.Context) {
	var req pbColony.CreateOrUpdateMdsColonyRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建mds集群参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	colony := biz.MdsColonyModel{
		ColonyNum:     req.ColonyNum,
		ExtractedName: req.ExtractedName,
		MonNodeID:     req.MonNodeID,
		PackageID:     req.PackageID,
	}

	m, rErr := s.ucColony.CreateMdsColony(ctx, colony)
	if rErr != nil {
		s.log.Error(
			"创建mds集群失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbColony.MdsColonyReply{
		Code: http.StatusOK,
		Data: *MdsColonyToDetailOut(*m),
	})
}

// @Summary 更新mds集群
// @Description 本接口用于更新指定ID的mds集群
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param id path uint true "mds集群编号"
// @Param request body pbColony.CreateOrUpdateMdsColonyRequest true "更新mds集群请求"
// @Success 200 {object} pbColony.MdsColonyReply "成功返回mds集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony/{id} [put]
// @Security ApiKeyAuth
func (s *MdsColonyService) UpdateMdsColony(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定更新mds集群ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbColony.CreateOrUpdateMdsColonyRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新mds集群参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	data := map[string]any{
		"colony_num":     req.ColonyNum,
		"extracted_name": req.ExtractedName,
		"package_id":     req.PackageID,
		"mon_node_id":    req.MonNodeID,
	}

	m, err := s.ucColony.UpdateMdsColonyByID(ctx, uri.ID, data)
	if err != nil {
		s.log.Error(
			"更新mds集群失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbColony.MdsColonyReply{
		Code: http.StatusOK,
		Data: *MdsColonyToDetailOut(*m),
	})
}

// @Summary 删除mds集群
// @Description 本接口用于删除指定ID的mds集群
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param id path uint true "mds集群编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony/{id} [delete]
// @Security ApiKeyAuth
func (s *MdsColonyService) DeleteMdsColony(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除mds集群ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除mds集群",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucColony.DeleteMdsColonyByID(ctx, uri.ID); err != nil {
		s.log.Error(
			"删除mds集群失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除mds集群成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询mds集群详情
// @Description 本接口用于查询指定ID的mds集群详情
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param id path uint true "mds集群编号"
// @Success 200 {object} pbColony.MdsColonyReply "成功返回mds集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony/{id} [get]
// @Security ApiKeyAuth
func (s *MdsColonyService) GetMdsColony(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询mds集群ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询mds集群详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucColony.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询mds集群详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	if m == nil {
		s.log.Error(
			"查询mds集群详情失败",
			zap.Error(biz.ErrMdsColonyNotFound),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(biz.ErrMdsColonyNotFound.Code, biz.ErrMdsColonyNotFound.ToMap())
		return
	}

	s.log.Info(
		"查询mds集群详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := MdsColonyToDetailOut(*m)
	ctx.JSON(http.StatusOK, &pbColony.MdsColonyReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询mds集群列表
// @Description 本接口用于查询mds集群列表
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param name query string false "mds集群名称"
// @Param is_enabled query bool false "是否启用"
// @Param username query string false "创建用户名"
// @Success 200 {object} pbColony.PagMdsColonyReply "成功返回mds集群列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony [get]
// @Security ApiKeyAuth
func (s *MdsColonyService) ListMdsColony(ctx *gin.Context) {
	var req pbColony.ListMdsColonyRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询mds集群列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询mds集群列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"Package", "MonNode"},
		IsCount:  true,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, err := s.ucColony.ListMdsColony(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询mds集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询mds集群列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mbs := ListMdsColonyToDetailOut(ms)
	ctx.JSON(http.StatusOK, &pbColony.PagMdsColonyReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

// @Summary 查询mds集群列表的任务状态
// @Description 本接口用于查询mds集群列表的任务状态
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param request query pbColony.ListMdsColonyRequest false "查询参数"
// @Success 200 {object} pbColony.ListMdsTaskStatusReply "成功返回mds集群列表的任务状态"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony/status [get]
// @Security ApiKeyAuth
func (s *MdsColonyService) ListMdsTaskStatus(ctx *gin.Context) {
	var req pbColony.ListMdsColonyRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询mds集群列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
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

	_, ms, err := s.ucColony.ListMdsColony(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询mds集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	if ms == nil {
		s.log.Warn(
			"查询mds集群列表为nil",
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(biz.ErrMdsColonyListEmpty.Code, biz.ErrMdsColonyListEmpty.ToMap())
		return
	}

	data := make(map[string]pbColony.MdsTaskStatus, len(*ms))
	for _, m := range *ms {
		taskStatus, rErr := s.ucColony.GetMdsTaskStatus(ctx, m.ColonyNum)
		if rErr != nil {
			s.log.Error(
				"查询mds集群任务状态失败",
				zap.Error(rErr),
				zap.Uint32(pbComm.RequestIDKey, m.ID),
				zap.String("colony_num", m.ColonyNum),
				zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			)
		}
		data[m.ColonyNum] = *taskStatus
	}

	ctx.JSON(http.StatusOK, &pbColony.ListMdsTaskStatusReply{
		Code: http.StatusOK,
		Data: data,
	})
}

func (s *MdsColonyService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/colony", s.CreateMdsColony)
	r.PUT("/colony/:id", s.UpdateMdsColony)
	r.DELETE("/colony/:id", s.DeleteMdsColony)
	r.GET("/colony/:id", s.GetMdsColony)
	r.GET("/colony", s.ListMdsColony)
	r.GET("/colony/status", s.ListMdsTaskStatus)
}

func MdsColonyToBaseOut(
	m biz.MdsColonyModel,
) *pbColony.MdsColonyBaseOut {
	return &pbColony.MdsColonyBaseOut{
		ID:            m.ID,
		ColonyNum:     m.ColonyNum,
		ExtractedName: m.ExtractedName,
	}
}

func MdsColonyToStandardOut(
	m biz.MdsColonyModel,
) *pbColony.MdsColonyStandardOut {
	return &pbColony.MdsColonyStandardOut{
		MdsColonyBaseOut: *MdsColonyToBaseOut(m),
		CreatedAt:        m.CreatedAt.String(),
		UpdatedAt:        m.UpdatedAt.String(),
	}
}

func MdsColonyToDetailOut(
	m biz.MdsColonyModel,
) *pbColony.MdsColonyDetailOut {
	return &pbColony.MdsColonyDetailOut{
		MdsColonyStandardOut: *MdsColonyToStandardOut(m),
		Package:              svReso.PackageModelToOutBase(m.Package),
		MonNode:              svMon.MonNodeToBaseOut(m.MonNode),
	}
}

func ListMdsColonyToDetailOut(
	rms *[]biz.MdsColonyModel,
) *[]pbColony.MdsColonyDetailOut {
	if rms == nil {
		return &[]pbColony.MdsColonyDetailOut{}
	}

	ms := *rms
	mso := make([]pbColony.MdsColonyDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := MdsColonyToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
