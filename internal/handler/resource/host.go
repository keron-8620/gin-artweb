package resource

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	resomodel "gin-artweb/internal/model/resource"
	resosvc "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

// HostHandler 处理主机相关的请求
// 包含日志记录和主机服务的引用
type HostHandler struct {
	log     *zap.Logger          // 日志记录器
	hostSvc *resosvc.HostService // 主机服务
}

func NewHostHandler(
	logger *zap.Logger,
	svcHost *resosvc.HostService,
) *HostHandler {
	return &HostHandler{
		log:     logger,
		hostSvc: svcHost,
	}
}

// @Summary 新增主机
// @Description 本接口用于新增主机
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param request body resomodel.HostUpsertDTO true "创建主机请求"
// @Success 201 {object} resomodel.HostResp "创建主机成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host [post]
// @Security ApiKeyAuth
func (h *HostHandler) CreateHost(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req resomodel.HostUpsertDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"新增主机:绑定创建主机请求参数失败") {
		return
	}

	log.Info("创建主机:开始执行")

	log.Debug(
		"创建主机:入参详情",
		zap.Object("host_upsert_dto", &req),
	)

	createStepStart := time.Now()
	m, err := h.hostSvc.CreateHost(ctx, req)
	createStepDuration := time.Since(createStepStart)
	if err != nil {
		log.Error(
			"创建主机:执行失败",
			zap.Error(err),
			zap.Object("host_upsert_dto", &req),
			zap.Duration("create_step_duration", createStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"创建主机:创建主机模型详情",
		zap.Object("host_model", m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	log.Info(
		"创建主机:执行成功",
		zap.Uint32("host_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := resomodel.HostModelToStandardOut(*m)
	ctx.JSON(http.StatusCreated, &resomodel.HostResp{
		Code: http.StatusCreated,
		Data: *mo,
	})
}

// @Summary 更新主机
// @Description 本接口用于更新指定ID的主机
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param id path uint true "主机编号"
// @Param request body resomodel.HostUpsertDTO true "更新主机请求"
// @Success 200 {object} resomodel.HostResp "更新主机成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{id} [put]
// @Security ApiKeyAuth
func (h *HostHandler) UpdateHost(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBindUri(
		ctx, log, &uri,
		"更新主机:绑定更新主机ID参数失败") {
		return
	}

	var req resomodel.HostUpsertDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"更新主机:绑定更新主机请求参数失败") {
		return
	}

	log.Info(
		"更新主机:开始执行",
		zap.Uint32("host_id", uri.ID),
	)

	log.Debug(
		"更新主机:入参详情",
		zap.Uint32("host_id", uri.ID),
		zap.Object("host_upsert_dto", &req),
	)

	updateStepStart := time.Now()
	m, err := h.hostSvc.UpdateHostById(ctx, uri.ID, req)
	updateStepDuration := time.Since(updateStepStart)
	if err != nil {
		log.Error(
			"更新主机:执行失败",
			zap.Error(err),
			zap.Uint32("host_id", uri.ID),
			zap.Object("host_upsert_dto", &req),
			zap.Duration("update_step_duration", updateStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"更新主机:更新后的主机模型详情",
		zap.Object("host_model", m),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	log.Info(
		"更新主机:执行成功",
		zap.Uint32("host_id", uri.ID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := resomodel.HostModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &resomodel.HostResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 删除主机
// @Description 本接口用于删除指定ID的主机
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param id path uint true "主机编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{id} [delete]
// @Security ApiKeyAuth
func (h *HostHandler) DeleteHost(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBindUri(
		ctx, log, &uri,
		"删除主机:绑定删除主机ID参数失败") {
		return
	}

	log.Info(
		"删除主机:开始执行",
		zap.Uint32("host_id", uri.ID),
	)

	deleteStepStart := time.Now()
	err := h.hostSvc.DeleteHostById(ctx, uri.ID)
	deleteStepDuration := time.Since(deleteStepStart)
	if err != nil {
		log.Error(
			"删除主机:执行失败",
			zap.Error(err),
			zap.Uint32("host_id", uri.ID),
			zap.Duration("delete_step_duration", deleteStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"删除主机:执行成功",
		zap.Uint32("host_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询主机
// @Description 本接口用于查询指定ID的主机
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param id path uint true "主机编号"
// @Success 200 {object} resomodel.HostResp "获取主机详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{id} [get]
// @Security ApiKeyAuth
func (h *HostHandler) GetHost(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBindUri(
		ctx, log, &uri,
		"查询主机:绑定查询主机ID参数失败") {
		return
	}

	log.Info(
		"查询主机:开始执行",
		zap.Uint32("host_id", uri.ID),
	)

	findStepStart := time.Now()
	m, err := h.hostSvc.FindHostById(ctx, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询主机:执行失败",
			zap.Error(err),
			zap.Uint32("host_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"查询主机:查询到的主机模型详情",
		zap.Object("host_model", m),
		zap.Duration("find_step_duration", findStepDuration),
	)

	log.Info(
		"查询主机:执行成功",
		zap.Uint32("host_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := resomodel.HostModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &resomodel.HostResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询主机列表
// @Description 本接口用于查询主机列表
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param request query resomodel.ListHostDTO false "查询参数"
// @Success 200 {object} resomodel.PagHostResp "成功返回主机列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host [get]
// @Security ApiKeyAuth
func (h *HostHandler) ListHost(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req resomodel.ListHostDTO
	if !common.ShouldBindQuery(
		ctx, log, &req,
		"查询主机列表:绑定查询主机列表参数失败") {
		return
	}

	log.Info("查询主机列表:开始执行")

	log.Debug(
		"查询主机列表:参数详情",
		zap.Object("list_host_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.hostSvc.ListHost(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询主机列表:执行失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_host_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询主机列表:执行成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := resomodel.ListHostModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &resomodel.PagHostResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (h *HostHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/host", h.CreateHost)
	r.PUT("/host/:id", h.UpdateHost)
	r.DELETE("/host/:id", h.DeleteHost)
	r.GET("/host/:id", h.GetHost)
	r.GET("/host", h.ListHost)
}
