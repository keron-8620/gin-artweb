package oes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	oesmodel "gin-artweb/internal/model/oes"
	oessvc "gin-artweb/internal/service/oes"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

type OesNodeHandler struct {
	log     *zap.Logger
	nodeSvc *oessvc.OesNodeService
}

func NewOesNodeHandler(
	logger *zap.Logger,
	nodeSvc *oessvc.OesNodeService,
) *OesNodeHandler {
	return &OesNodeHandler{
		log:     logger,
		nodeSvc: nodeSvc,
	}
}

// @Summary 创建oes节点
// @Description 本接口用于创建新的oes节点
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param request body oesmodel.OesNodeUpsertDTO true "创建oes节点请求"
// @Success 200 {object} oesmodel.OesNodeResp "成功返回oes节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node [post]
// @Security ApiKeyAuth
func (s *OesNodeHandler) CreateOesNode(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	var req oesmodel.OesNodeUpsertDTO
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"绑定创建oes节点参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, rErr := s.nodeSvc.CreateOesNode(ctx, req)
	if rErr != nil {
		log.Error(
			"创建oes节点失败",
			zap.Error(rErr),
			zap.Object("oes_node_dto", &req),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &oesmodel.OesNodeResp{
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
// @Param request body oesmodel.OesNodeUpsertDTO true "更新oes节点请求"
// @Success 200 {object} oesmodel.OesNodeResp "成功返回oes节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node/{id} [put]
// @Security ApiKeyAuth
func (s *OesNodeHandler) UpdateOesNode(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"绑定更新oes节点ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req oesmodel.OesNodeUpsertDTO
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"绑定更新oes节点参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	m, rErr := s.nodeSvc.UpdateOesNodeByID(ctx, uri.ID, req)
	if rErr != nil {
		log.Error(
			"更新oes节点失败",
			zap.Error(rErr),
			zap.Uint32("oes_node_id", uri.ID),
			zap.Object("oes_node_dto", &req),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &oesmodel.OesNodeResp{
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
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node/{id} [delete]
// @Security ApiKeyAuth
func (s *OesNodeHandler) DeleteOesNode(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"绑定删除oes节点ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"开始删除oes节点",
		zap.Uint32("oes_node_id", uri.ID),
	)

	rErr := s.nodeSvc.DeleteOesNodeByID(ctx, uri.ID)
	if rErr != nil {
		log.Error(
			"删除oes节点失败",
			zap.Error(rErr),
			zap.Uint32("oes_node_id", uri.ID),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"删除oes节点成功",
		zap.Uint32("oes_node_id", uri.ID),
		zap.Duration("total_time", time.Since(startTime)),
	)

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询oes节点详情
// @Description 本接口用于查询指定ID的oes节点详情
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param id path uint true "oes节点编号"
// @Success 200 {object} oesmodel.OesNodeResp "成功返回oes节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "oes节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node/{id} [get]
// @Security ApiKeyAuth
func (s *OesNodeHandler) GetOesNode(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"绑定查询oes节点ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"开始查询oes节点详情",
		zap.Uint32("oes_node_id", uri.ID),
	)

	m, rErr := s.nodeSvc.FindOesNodeByID(ctx, []string{"OesColony", "Host"}, uri.ID)
	if rErr != nil {
		log.Error(
			"查询oes节点详情失败",
			zap.Error(rErr),
			zap.Uint32("oes_node_id", uri.ID),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"查询oes节点详情成功",
		zap.Uint32("oes_node_id", uri.ID),
		zap.Duration("total_time", time.Since(startTime)),
	)

	mo := oesmodel.OesNodeToDetailOut(*m)
	ctx.JSON(http.StatusOK, &oesmodel.OesNodeResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询oes节点列表
// @Description 本接口用于查询oes节点列表
// @Tags oes节点管理
// @Accept json
// @Produce json
// @Param request query oesmodel.ListOesNodeDTO false "查询参数"
// @Success 200 {object} oesmodel.PagOesNodeResp "成功返回oes节点列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/node [get]
// @Security ApiKeyAuth
func (s *OesNodeHandler) ListOesNode(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	var req oesmodel.ListOesNodeDTO
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Error(
			"绑定查询oes节点列表参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info("开始查询oes节点列表")

	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, rErr := s.nodeSvc.ListOesNode(ctx, page, size, req)
	if rErr != nil {
		log.Error(
			"查询oes节点列表失败",
			zap.Error(rErr),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("oes_node_dto", &req),
			zap.Duration("total_time", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"查询oes节点列表成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("total_time", time.Since(startTime)),
	)

	mbs := oesmodel.ListOesNodeToDetailOut(ms)
	ctx.JSON(http.StatusOK, &oesmodel.PagOesNodeResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (s *OesNodeHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/node", s.CreateOesNode)
	r.PUT("/node/:id", s.UpdateOesNode)
	r.DELETE("/node/:id", s.DeleteOesNode)
	r.GET("/node/:id", s.GetOesNode)
	r.GET("/node", s.ListOesNode)
}
