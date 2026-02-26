package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	oesmodel "gin-artweb/internal/model/oes"
	oessvc "gin-artweb/internal/service/oes"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type OesNodeService struct {
	log    *zap.Logger
	ucNode *oessvc.OesNodeService
}

func NewOesNodeService(
	logger *zap.Logger,
	ucNode *oessvc.OesNodeService,
) *OesNodeService {
	return &OesNodeService{
		log:    logger,
		ucNode: ucNode,
	}
}

// @Summary 创建oes节点
// @Description 本接口用于创建新的oes节点
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param request body oesmodel.CreateOrUpdateOesNodeRequest true "创建oes节点请求"
// @Success 200 {object} oesmodel.OesNodeReply "成功返回oes节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node [post]
// @Security ApiKeyAuth
func (s *OesNodeService) CreateOesNode(ctx *gin.Context) {
	var req oesmodel.CreateOrUpdateOesNodeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建oes节点参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	colony := oesmodel.OesNodeModel{
		NodeRole:    req.NodeRole,
		IsEnable:    req.IsEnable,
		HostID:      req.HostID,
		OesColonyID: req.OesColonyID,
	}

	m, rErr := s.ucNode.CreateOesNode(ctx, colony)
	if rErr != nil {
		s.log.Error(
			"创建oes节点失败",
			zap.Error(rErr),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &oesmodel.OesNodeReply{
		Code: http.StatusOK,
		Data: *oesmodel.OesNodeToDetailOut(*m),
	})
}

// @Summary 更新oes节点
// @Description 本接口用于更新指定ID的oes节点
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param id path uint true "oes节点编号"
// @Param request body oesmodel.CreateOrUpdateOesNodeRequest true "更新oes节点请求"
// @Success 200 {object} oesmodel.OesNodeReply "成功返回oes节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node/{id} [put]
// @Security ApiKeyAuth
func (s *OesNodeService) UpdateOesNode(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定更新oes节点ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req oesmodel.CreateOrUpdateOesNodeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新oes节点参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	data := map[string]any{
		"node_role":     req.NodeRole,
		"is_enable":     req.IsEnable,
		"host_id":       req.HostID,
		"oes_colony_id": req.OesColonyID,
	}

	m, rErr := s.ucNode.UpdateOesNodeByID(ctx, uri.ID, data)
	if rErr != nil {
		s.log.Error(
			"更新oes节点失败",
			zap.Error(rErr),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &oesmodel.OesNodeReply{
		Code: http.StatusOK,
		Data: *oesmodel.OesNodeToDetailOut(*m),
	})
}

// @Summary 删除oes节点
// @Description 本接口用于删除指定ID的oes节点
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param id path uint true "oes节点编号"
// @Success 200 {object} commodel.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node/{id} [delete]
// @Security ApiKeyAuth
func (s *OesNodeService) DeleteOesNode(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除oes节点ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始删除oes节点",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	rErr := s.ucNode.DeleteOesNodeByID(ctx, uri.ID)
	if rErr != nil {
		s.log.Error(
			"删除oes节点失败",
			zap.Error(rErr),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"删除oes节点成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 查询oes节点详情
// @Description 本接口用于查询指定ID的oes节点详情
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param id path uint true "oes节点编号"
// @Success 200 {object} oesmodel.OesNodeReply "成功返回oes节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node/{id} [get]
// @Security ApiKeyAuth
func (s *OesNodeService) GetOesNode(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询oes节点ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始查询oes节点详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := s.ucNode.FindOesNodeByID(ctx, []string{"OesColony", "Host"}, uri.ID)
	if rErr != nil {
		s.log.Error(
			"查询oes节点详情失败",
			zap.Error(rErr),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"查询oes节点详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := oesmodel.OesNodeToDetailOut(*m)
	ctx.JSON(http.StatusOK, &oesmodel.OesNodeReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询oes节点列表
// @Description 本接口用于查询oes节点列表
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param request query oesmodel.ListOesNodeRequest false "查询参数"
// @Success 200 {object} oesmodel.PagOesNodeReply "成功返回oes节点列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node [get]
// @Security ApiKeyAuth
func (s *OesNodeService) ListOesNode(ctx *gin.Context) {
	var req oesmodel.ListOesNodeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询oes节点列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始查询oes节点列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"OesColony", "Host"},
		IsCount:  true,
		Size:     size,
		Page:     page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, rErr := s.ucNode.ListOesNode(ctx, qp)
	if rErr != nil {
		s.log.Error(
			"查询oes节点列表失败",
			zap.Error(rErr),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"查询oes节点列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := oesmodel.ListOesNodeToDetailOut(ms)
	ctx.JSON(http.StatusOK, &oesmodel.PagOesNodeReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (s *OesNodeService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/node", s.CreateOesNode)
	r.PUT("/node/:id", s.UpdateOesNode)
	r.DELETE("/node/:id", s.DeleteOesNode)
	r.GET("/node/:id", s.GetOesNode)
	r.GET("/node", s.ListOesNode)
}
