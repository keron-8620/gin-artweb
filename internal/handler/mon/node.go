package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	monmodel "gin-artweb/internal/model/mon"
	monsvc "gin-artweb/internal/service/mon"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type NodeHandler struct {
	log     *zap.Logger
	svcNode *monsvc.MonNodeService
}

func NewNodeHandler(
	logger *zap.Logger,
	svcNode *monsvc.MonNodeService,
) *NodeHandler {
	return &NodeHandler{
		log:     logger,
		svcNode: svcNode,
	}
}

// @Summary 创建mon节点
// @Description 本接口用于创建新的mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param request body monmodel.CreateOrUpdateMonNodeRequest true "创建mon节点请求"
// @Success 200 {object} monmodel.MonNodeReply "成功返回mon节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node [post]
// @Security ApiKeyAuth
func (h *NodeHandler) CreateMonNode(ctx *gin.Context) {
	var req monmodel.CreateOrUpdateMonNodeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定创建mon节点参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	node := monmodel.MonNodeModel{
		Name:        req.Name,
		DeployPath:  req.DeployPath,
		OutportPath: req.OutportPath,
		JavaHome:    req.JavaHome,
		URL:         req.URL,
		HostID:      req.HostID,
	}

	m, rErr := h.svcNode.CreateMonNode(ctx, node)
	if rErr != nil {
		h.log.Error(
			"创建mon节点失败",
			zap.Error(rErr),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &monmodel.MonNodeReply{
		Code: http.StatusOK,
		Data: *monmodel.MonNodeToDetailOut(*m),
	})
}

// @Summary 更新mon节点
// @Description 本接口用于更新指定ID的mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mon节点编号"
// @Param request body monmodel.CreateOrUpdateMonNodeRequest true "更新mon节点请求"
// @Success 200 {object} monmodel.MonNodeReply "成功返回mon节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{id} [put]
// @Security ApiKeyAuth
func (h *NodeHandler) UpdateMonNode(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定更新mon节点ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req monmodel.CreateOrUpdateMonNodeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定更新mon节点参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	data := map[string]any{
		"name":         req.Name,
		"deploy_path":  req.DeployPath,
		"outport_path": req.OutportPath,
		"java_home":    req.JavaHome,
		"url":          req.URL,
		"host_id":      req.HostID,
	}

	m, rErr := h.svcNode.UpdateMonNodeByID(ctx, uri.ID, data)
	if rErr != nil {
		h.log.Error(
			"更新mon节点失败",
			zap.Error(rErr),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &monmodel.MonNodeReply{
		Code: http.StatusOK,
		Data: *monmodel.MonNodeToDetailOut(*m),
	})
}

// @Summary 删除mon节点
// @Description 本接口用于删除指定ID的mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mon节点编号"
// @Success 200 {object} commodel.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{id} [delete]
// @Security ApiKeyAuth
func (h *NodeHandler) DeleteMonNode(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定删除mon节点ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始删除mon节点",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if rErr := h.svcNode.DeleteMonNodeByID(ctx, uri.ID); rErr != nil {
		h.log.Error(
			"删除mon节点失败",
			zap.Error(rErr),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"删除mon节点成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 查询mon节点详情
// @Description 本接口用于查询指定ID的mon节点详情
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mon节点编号"
// @Success 200 {object} monmodel.MonNodeReply "成功返回mon节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{id} [get]
// @Security ApiKeyAuth
func (h *NodeHandler) GetMonNode(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询mon节点ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询mon节点详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := h.svcNode.FindMonNodeByID(ctx, []string{"Host"}, uri.ID)
	if rErr != nil {
		h.log.Error(
			"查询mon节点详情失败",
			zap.Error(rErr),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"查询mon节点详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := monmodel.MonNodeToDetailOut(*m)
	ctx.JSON(http.StatusOK, &monmodel.MonNodeReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询mon节点列表
// @Description 本接口用于查询mon节点列表
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param request query monmodel.ListMonNodeRequest false "查询参数"
// @Success 200 {object} monmodel.PagMonNodeReply "成功返回mon节点列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node [get]
// @Security ApiKeyAuth
func (h *NodeHandler) ListMonNode(ctx *gin.Context) {
	var req monmodel.ListMonNodeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询mon节点列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询mon节点列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"Host"},
		IsCount:  true,
		Size:     size,
		Page:     page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, rErr := h.svcNode.ListMonNode(ctx, qp)
	if rErr != nil {
		h.log.Error(
			"查询mon节点列表失败",
			zap.Error(rErr),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"查询mon节点列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := monmodel.ListMonNodeToDetailOut(ms)
	ctx.JSON(http.StatusOK, &monmodel.PagMonNodeReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (h *NodeHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/node", h.CreateMonNode)
	r.PUT("/node/:id", h.UpdateMonNode)
	r.DELETE("/node/:id", h.DeleteMonNode)
	r.GET("/node/:id", h.GetMonNode)
	r.GET("/node", h.ListMonNode)
}
