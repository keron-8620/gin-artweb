package oes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	oesmodel "gin-artweb/internal/model/oes"
	oessvc "gin-artweb/internal/service/oes"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

// AgwHandler 处理agw相关的请求
// 包含日志记录和agw服务的引用
type AgwHandler struct {
	log    *zap.Logger        // 日志记录器
	agwSvc *oessvc.AgwService // agw服务
}

func NewAgwHandler(
	logger *zap.Logger,
	agwSvc *oessvc.AgwService,
) *AgwHandler {
	return &AgwHandler{
		log:    logger,
		agwSvc: agwSvc,
	}
}

// @Summary 新增agw
// @Description 本接口用于新增agw
// @Tags agw管理
// @Accept json
// @Produce json
// @Param request body oesmodel.AgwUpsertDTO true "创建agw请求"
// @Success 201 {object} oesmodel.AgwResp "创建agw成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/agw [post]
// @Security ApiKeyAuth
func (h *AgwHandler) CreateAgw(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("新增agw:开始执行")

	var req oesmodel.AgwUpsertDTO
	if !common.ShouldBind(c, log, &req, "新增agw:绑定创建agw参数失败") {
		return
	}

	m, rErr := h.agwSvc.CreateAgw(ctx, req)
	if rErr != nil {
		log.Error(
			"创建agw:执行失败",
			zap.Error(rErr),
			zap.Object("agw_upsert_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	log.Info(
		"创建agw:执行成功",
		zap.Uint32("agw_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusCreated, &oesmodel.AgwResp{
		Code: http.StatusCreated,
		Data: *oesmodel.AgwToDetailOut(*m),
	})
}

// @Summary 更新agw
// @Description 本接口用于更新指定ID的agw
// @Tags agw管理
// @Accept json
// @Produce json
// @Param id path uint true "agw编号"
// @Param request body oesmodel.AgwUpsertDTO true "更新agw请求"
// @Success 200 {object} oesmodel.AgwResp "更新agw成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "agw未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/agw/{id} [put]
// @Security ApiKeyAuth
func (h *AgwHandler) UpdateAgw(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("更新agw:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "更新agw:绑定更新agwID参数失败") {
		return
	}

	var req oesmodel.AgwUpsertDTO
	if !common.ShouldBind(c, log, &req, "更新agw:绑定更新agw参数失败") {
		return
	}

	m, rErr := h.agwSvc.UpdateAgwByID(ctx, uri.ID, req)
	if rErr != nil {
		log.Error(
			"更新agw:执行失败",
			zap.Error(rErr),
			zap.Uint32("agw_id", uri.ID),
			zap.Object("agw_upsert_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	log.Info(
		"更新agw:执行成功",
		zap.Uint32("agw_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &oesmodel.AgwResp{
		Code: http.StatusOK,
		Data: *oesmodel.AgwToDetailOut(*m),
	})
}

// @Summary 删除agw
// @Description 本接口用于删除指定ID的agw
// @Tags agw管理
// @Accept json
// @Produce json
// @Param id path uint true "agw编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "agw未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/agw/{id} [delete]
// @Security ApiKeyAuth
func (h *AgwHandler) DeleteAgw(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("删除agw:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "删除agw:绑定删除agwID参数失败") {
		return
	}

	err := h.agwSvc.DeleteAgwByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"删除agw:执行失败",
			zap.Error(err),
			zap.Uint32("agw_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"删除agw:执行成功",
		zap.Uint32("agw_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询agw
// @Description 本接口用于查询指定ID的agw
// @Tags agw管理
// @Accept json
// @Produce json
// @Param id path uint true "agw编号"
// @Success 200 {object} oesmodel.AgwResp "获取agw详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "agw未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/agw/{id} [get]
// @Security ApiKeyAuth
func (h *AgwHandler) GetAgw(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询agw:绑定查询agwID参数失败") {
		return
	}

	m, rErr := h.agwSvc.FindAgwByID(ctx, []string{"Host", "Package"}, uri.ID)
	if rErr != nil {
		log.Error(
			"查询agw:执行失败",
			zap.Error(rErr),
			zap.Uint32("agw_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	mo := oesmodel.AgwToDetailOut(*m)
	c.JSON(http.StatusOK, &oesmodel.AgwResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询agw列表
// @Description 本接口用于查询agw列表
// @Tags agw管理
// @Accept json
// @Produce json
// @Param request query oesmodel.ListAgwDTO false "查询参数"
// @Success 200 {object} oesmodel.PagAgwResp "成功返回agw列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/oes/agw [get]
// @Security ApiKeyAuth
func (h *AgwHandler) ListAgw(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var req oesmodel.ListAgwDTO
	if !common.ShouldBindQuery(c, log, &req, "查询agw列表:绑定查询参数失败") {
		return
	}

	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, rErr := h.agwSvc.ListAgw(ctx, page, size, req)
	if rErr != nil {
		log.Error(
			"查询agw列表:执行失败",
			zap.Error(rErr),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_agw_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	mbs := oesmodel.ListAgwToDetailOut(ms)
	c.JSON(http.StatusOK, &oesmodel.PagAgwResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (h *AgwHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/agw", h.CreateAgw)
	r.PUT("/agw/:id", h.UpdateAgw)
	r.DELETE("/agw/:id", h.DeleteAgw)
	r.GET("/agw/:id", h.GetAgw)
	r.GET("/agw", h.ListAgw)
}
