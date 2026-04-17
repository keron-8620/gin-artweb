package sys

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	sysmodel "gin-artweb/internal/model/sys"
	syssvc "gin-artweb/internal/service/sys"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

// ApiHandler 处理API相关的请求
// 包含日志记录和API服务的引用
type ApiHandler struct {
	log    *zap.Logger        // 日志记录器
	apiSvc *syssvc.ApiService // API服务
}

func NewApiHandler(
	logger *zap.Logger,
	svcApi *syssvc.ApiService,
) *ApiHandler {
	return &ApiHandler{
		log:    logger,
		apiSvc: svcApi,
	}
}

// @Summary 新增API
// @Description 本接口用于新增API
// @Tags API管理
// @Accept json
// @Produce json
// @Param request body sysmodel.CreateApiDTO true "创建API请求"
// @Success 201 {object} sysmodel.ApiResp "创建API成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Router /api/v1/customer/api [post]
// @Security ApiKeyAuth
func (h *ApiHandler) CreateApi(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)

	var req sysmodel.CreateApiDTO
	if !common.ShouldBind(ctx, log, &req, "创建API:绑定创建API请求参数失败") {
		return
	}

	m, err := h.apiSvc.CreateApi(ctx, req)
	if err != nil {
		log.Error(
			"创建API:执行失败",
			zap.Error(err),
			zap.Object("create_api_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"创建API:执行成功",
		zap.Uint32("api_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := sysmodel.ApiModelToStandardOut(*m)
	ctx.JSON(http.StatusCreated, &sysmodel.ApiResp{
		Code: http.StatusCreated,
		Data: mo,
	})
}

// @Summary 更新API
// @Description 本接口用于更新指定ID的API
// @Tags API管理
// @Accept json
// @Produce json
// @Param id path uint true "API编号"
// @Param request body sysmodel.UpdateApiDTO true "更新API请求"
// @Success 200 {object} sysmodel.ApiResp "更新API成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "API未找到"
// @Router /api/v1/customer/api/{id} [put]
// @Security ApiKeyAuth
func (h *ApiHandler) UpdateApi(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(ctx, log, &uri, "更新API:绑定更新APIID参数失败") {
		return
	}

	var req sysmodel.UpdateApiDTO
	if !common.ShouldBind(ctx, log, &req, "更新API:绑定更新API请求参数失败") {
		return
	}

	m, err := h.apiSvc.UpdateApiByID(ctx, uri.ID, req)
	if err != nil {
		log.Error(
			"更新API:执行失败",
			zap.Error(err),
			zap.Uint32("api_id", uri.ID),
			zap.Object("update_api_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"更新API:执行成功",
		zap.Uint32("api_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := sysmodel.ApiModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &sysmodel.ApiResp{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 删除API
// @Description 本接口用于删除指定ID的API
// @Tags API管理
// @Accept json
// @Produce json
// @Param id path uint true "API编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "API未找到"
// @Router /api/v1/customer/api/{id} [delete]
// @Security ApiKeyAuth
func (h *ApiHandler) DeleteApi(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(ctx, log, &uri, "删除API:绑定删除APIID参数失败") {
		return
	}

	if err := h.apiSvc.DeleteApiByID(ctx, uri.ID); err != nil {
		log.Error(
			"删除API:执行失败",
			zap.Error(err),
			zap.Uint32("api_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询API
// @Description 本接口用于查询指定ID的API
// @Tags API管理
// @Accept json
// @Produce json
// @Param id path uint true "API编号"
// @Success 200 {object} sysmodel.ApiResp "获取API详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "API未找到"
// @Router /api/v1/customer/api/{id} [get]
// @Security ApiKeyAuth
func (h *ApiHandler) GetApi(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(ctx, log, &uri, "查询API:绑定查询APIID参数失败") {
		return
	}

	m, err := h.apiSvc.FindApiByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"查询API:执行失败",
			zap.Error(err),
			zap.Uint32("api_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询API:执行成功",
		zap.Uint32("api_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := sysmodel.ApiModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &sysmodel.ApiResp{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 查询API列表
// @Description 本接口用于查询API列表
// @Tags API管理
// @Accept json
// @Produce json
// @Param request query sysmodel.ListApiDTO false "查询参数"
// @Success 200 {object} sysmodel.PagApiResp "成功返回API列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "内部服务错误"
// @Router /api/v1/customer/api [get]
// @Security ApiKeyAuth
func (h *ApiHandler) ListApi(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)

	var req sysmodel.ListApiDTO
	if !common.ShouldBindQuery(ctx, log, &req, "查询API列表:绑定查询API列表请求参数失败") {
		return
	}

	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.apiSvc.ListApi(ctx, page, size, req)
	if err != nil {
		log.Error(
			"查询API列表:执行失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_api_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询API列表:执行成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := sysmodel.ListApiModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &sysmodel.PagApiResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// LoadRouter 注册API相关的路由
func (h *ApiHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/api", h.CreateApi)
	r.PUT("/api/:id", h.UpdateApi)
	r.DELETE("/api/:id", h.DeleteApi)
	r.GET("/api/:id", h.GetApi)
	r.GET("/api", h.ListApi)
}
