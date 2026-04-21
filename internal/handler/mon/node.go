package mon

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	monmodel "gin-artweb/internal/model/mon"
	monsvc "gin-artweb/internal/service/mon"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

// NodeHandler 处理mon节点相关的请求
// 包含日志记录和mon节点服务的引用
type NodeHandler struct {
	log     *zap.Logger            // 日志记录器
	nodeSvc *monsvc.MonNodeService // mon节点服务
}

func NewNodeHandler(
	logger *zap.Logger,
	svcNode *monsvc.MonNodeService,
) *NodeHandler {
	return &NodeHandler{
		log:     logger,
		nodeSvc: svcNode,
	}
}

// @Summary 新增mon节点
// @Description 本接口用于新增mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param request body monmodel.MonNodeUpsertDTO true "创建mon节点请求"
// @Success 201 {object} monmodel.MonNodeResp "创建mon节点成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node [post]
// @Security ApiKeyAuth
func (h *NodeHandler) CreateMonNode(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("新增mon节点:开始执行")

	var req monmodel.MonNodeUpsertDTO
	if !common.ShouldBind(c, log, &req, "新增mon节点:绑定创建mon节点参数失败") {
		return
	}

	m, rErr := h.nodeSvc.CreateMonNode(ctx, req)
	if rErr != nil {
		log.Error(
			"创建mon节点:执行失败",
			zap.Error(rErr),
			zap.Object("mon_node_upsert_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	log.Info(
		"创建mon节点:执行成功",
		zap.Uint32("node_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusCreated, &monmodel.MonNodeResp{
		Code: http.StatusCreated,
		Data: *monmodel.MonNodeToDetailOut(*m),
	})
}

// @Summary 更新mon节点
// @Description 本接口用于更新指定ID的mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mon节点编号"
// @Param request body monmodel.MonNodeUpsertDTO true "更新mon节点请求"
// @Success 200 {object} monmodel.MonNodeResp "更新mon节点成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{id} [put]
// @Security ApiKeyAuth
func (h *NodeHandler) UpdateMonNode(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("更新mon节点:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "更新mon节点:绑定更新mon节点ID参数失败") {
		return
	}

	var req monmodel.MonNodeUpsertDTO
	if !common.ShouldBind(c, log, &req, "更新mon节点:绑定更新mon节点参数失败") {
		return
	}

	m, rErr := h.nodeSvc.UpdateMonNodeByID(ctx, uri.ID, req)
	if rErr != nil {
		log.Error(
			"更新mon节点:执行失败",
			zap.Error(rErr),
			zap.Uint32("node_id", uri.ID),
			zap.Object("mon_node_upsert_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	log.Info(
		"更新mon节点:执行成功",
		zap.Uint32("node_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &monmodel.MonNodeResp{
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
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{id} [delete]
// @Security ApiKeyAuth
func (h *NodeHandler) DeleteMonNode(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("删除mon节点:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "删除mon节点:绑定删除mon节点ID参数失败") {
		return
	}

	err := h.nodeSvc.DeleteMonNodeByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"删除mon节点:执行失败",
			zap.Error(err),
			zap.Uint32("node_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"删除mon节点:执行成功",
		zap.Uint32("node_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询mon节点
// @Description 本接口用于查询指定ID的mon节点
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mon节点编号"
// @Success 200 {object} monmodel.MonNodeResp "获取mon节点详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mon节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node/{id} [get]
// @Security ApiKeyAuth
func (h *NodeHandler) GetMonNode(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询mon节点:绑定查询mon节点ID参数失败") {
		return
	}

	m, rErr := h.nodeSvc.FindMonNodeByID(ctx, []string{"Host"}, uri.ID)
	if rErr != nil {
		log.Error(
			"查询mon节点:执行失败",
			zap.Error(rErr),
			zap.Uint32("node_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	mo := monmodel.MonNodeToDetailOut(*m)
	c.JSON(http.StatusOK, &monmodel.MonNodeResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询mon节点列表
// @Description 本接口用于查询mon节点列表
// @Tags mon节点管理
// @Accept json
// @Produce json
// @Param request query monmodel.ListMonNodeDTO false "查询参数"
// @Success 200 {object} monmodel.PagMonNodeResp "成功返回mon节点列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mon/node [get]
// @Security ApiKeyAuth
func (h *NodeHandler) ListMonNode(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var req monmodel.ListMonNodeDTO
	if !common.ShouldBindQuery(c, log, &req, "查询mon节点列表:绑定查询参数失败") {
		return
	}

	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, rErr := h.nodeSvc.ListMonNode(ctx, page, size, req)
	if rErr != nil {
		log.Error(
			"查询mon节点列表:执行失败",
			zap.Error(rErr),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_mon_node_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	mbs := monmodel.ListMonNodeToDetailOut(ms)
	c.JSON(http.StatusOK, &monmodel.PagMonNodeResp{
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
