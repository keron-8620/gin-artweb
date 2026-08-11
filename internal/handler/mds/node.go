package mds

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	mdsmodel "gin-artweb/internal/model/mds"
	mdssvc "gin-artweb/internal/service/mds"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

type MdsNodeHandler struct {
	log     *zap.Logger
	nodeSvc *mdssvc.MdsNodeService
}

func NewMdsNodeHandler(
	logger *zap.Logger,
	nodeSvc *mdssvc.MdsNodeService,
) *MdsNodeHandler {
	return &MdsNodeHandler{
		log:     logger,
		nodeSvc: nodeSvc,
	}
}

// @Summary 创建mds节点
// @Description 本接口用于创建新的mds节点
// @Tags mds节点管理
// @Accept json
// @Produce json
// @Param request body mdsmodel.MdsNodeUpsertDTO true "创建mds节点请求"
// @Success 200 {object} mdsmodel.MdsNodeResp "成功返回mds节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/node [post]
// @Security ApiKeyAuth
func (s *MdsNodeHandler) CreateMdsNode(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("创建mds节点:开始执行")

	var req mdsmodel.MdsNodeUpsertDTO
	if !common.ShouldBind(c, log, &req, "创建mds节点:绑定参数失败") {
		return
	}

	m, err := s.nodeSvc.CreateMdsNode(ctx, req)
	if err != nil {
		log.Error(
			"创建mds节点:执行失败",
			zap.Error(err),
			zap.Object("mds_node_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"创建mds节点:执行成功",
		zap.Uint32("mds_node_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &mdsmodel.MdsNodeResp{
		Code: http.StatusOK,
		Data: *mdsmodel.MdsNodeToDetailOut(*m),
	})
}

// @Summary 更新mds节点
// @Description 本接口用于更新指定ID的mds节点
// @Tags mds节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mds节点编号"
// @Param request body mdsmodel.MdsNodeUpsertDTO true "更新mds节点请求"
// @Success 200 {object} mdsmodel.MdsNodeResp "成功返回mds节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/node/{id} [put]
// @Security ApiKeyAuth
func (s *MdsNodeHandler) UpdateMdsNode(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("更新mds节点:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "更新mds节点:绑定更新mds节点ID参数失败") {
		return
	}

	var req mdsmodel.MdsNodeUpsertDTO
	if !common.ShouldBind(c, log, &req, "更新mds节点:绑定更新mds节点参数失败") {
		return
	}

	m, rErr := s.nodeSvc.UpdateMdsNodeByID(ctx, uri.ID, req)
	if rErr != nil {
		log.Error(
			"更新mds节点:执行失败",
			zap.Error(rErr),
			zap.Uint32("mds_node_id", uri.ID),
			zap.Object("mds_node_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	log.Info(
		"更新mds节点:执行成功",
		zap.Uint32("mds_node_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &mdsmodel.MdsNodeResp{
		Code: http.StatusOK,
		Data: *mdsmodel.MdsNodeToDetailOut(*m),
	})
}

// @Summary 删除mds节点
// @Description 本接口用于删除指定ID的mds节点
// @Tags mds节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mds节点编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/node/{id} [delete]
// @Security ApiKeyAuth
func (s *MdsNodeHandler) DeleteMdsNode(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("删除mds节点:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "删除mds节点:绑定删除mds节点ID参数失败") {
		return
	}

	err := s.nodeSvc.DeleteMdsNodeByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"删除mds节点:执行失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"删除mds节点:执行成功",
		zap.Uint32("mds_node_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询mds节点详情
// @Description 本接口用于查询指定ID的mds节点详情
// @Tags mds节点管理
// @Accept json
// @Produce json
// @Param id path uint true "mds节点编号"
// @Success 200 {object} mdsmodel.MdsNodeResp "成功返回mds节点信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds节点未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/node/{id} [get]
// @Security ApiKeyAuth
func (s *MdsNodeHandler) GetMdsNode(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询mds节点:绑定查询mds节点ID参数失败") {
		return
	}

	m, err := s.nodeSvc.FindMdsNodeByID(ctx, []string{"MdsColony", "Host"}, uri.ID)
	if err != nil {
		log.Error(
			"查询mds节点:详情查询失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mo := mdsmodel.MdsNodeToDetailOut(*m)
	c.JSON(http.StatusOK, &mdsmodel.MdsNodeResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询mds节点列表
// @Description 本接口用于查询mds节点列表
// @Tags mds节点管理
// @Accept json
// @Produce json
// @Param request query mdsmodel.ListMdsNodeDTO false "查询参数"
// @Success 200 {object} mdsmodel.PagMdsNodeResp "成功返回mds节点列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/node [get]
// @Security ApiKeyAuth
func (s *MdsNodeHandler) ListMdsNode(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var req mdsmodel.ListMdsNodeDTO
	if !common.ShouldBindQuery(c, log, &req, "查询mds节点列表:绑定查询参数失败") {
		return
	}

	page, size, total, ms, err := s.nodeSvc.ListMdsNode(ctx, &req)
	if err != nil {
		log.Error(
			"查询mds节点列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("mds_node_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mbs := mdsmodel.ListMdsNodeToDetailOut(ms)
	c.JSON(http.StatusOK, &mdsmodel.PagMdsNodeResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (s *MdsNodeHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/node", s.CreateMdsNode)
	r.PUT("/node/:id", s.UpdateMdsNode)
	r.DELETE("/node/:id", s.DeleteMdsNode)
	r.GET("/node/:id", s.GetMdsNode)
	r.GET("/node", s.ListMdsNode)
}
