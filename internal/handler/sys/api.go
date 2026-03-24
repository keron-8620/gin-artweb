package sys

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	sysmodel "gin-artweb/internal/model/sys"
	syssvc "gin-artweb/internal/service/sys"
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
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"创建API：绑定请求参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"创建API：开始执行",
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"创建API：入参详情",
		zap.Object("create_api_dto", &req),
	)

	createStepStart := time.Now()
	m, err := h.apiSvc.CreateApi(ctx, req)
	createStepDuration := time.Since(createStepStart)
	if err != nil {
		log.Error(
			"创建API：创建API失败",
			zap.Error(err),
			zap.Object("create_api_dto", &req),
			zap.Duration("create_step_duration", createStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"创建API：创建API成功",
		zap.Object("api_model", m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	log.Info(
		"创建API：创建API成功",
		zap.Uint32("api_id", m.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
		zap.Duration("create_step_duration", createStepDuration),
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
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"更新API：绑定APIID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req sysmodel.UpdateApiDTO
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"更新API：绑定请求参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"更新API：开始执行",
		zap.Uint32("api_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"更新API：入参详情",
		zap.Uint32("api_id", uri.ID),
		zap.Object("update_api_dto", &req),
	)

	updateStepStart := time.Now()
	m, err := h.apiSvc.UpdateApiByID(ctx, uri.ID, req)
	updateStepDuration := time.Since(updateStepStart)
	if err != nil {
		log.Error(
			"更新API：更新API失败",
			zap.Error(err),
			zap.Uint32("api_id", uri.ID),
			zap.Object("update_api_dto", &req),
			zap.Duration("update_step_duration", updateStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"更新API：更新API成功",
		zap.Object("api_model", m),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	log.Info(
		"更新API：更新API成功",
		zap.Uint32("api_id", uri.ID),
		zap.Duration("update_step_duration", updateStepDuration),
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
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"删除API：绑定APIID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"删除API：开始执行",
		zap.Uint32("api_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	deleteStepStart := time.Now()
	if err := h.apiSvc.DeleteApiByID(ctx, uri.ID); err != nil {
		log.Error(
			"删除API：删除API失败",
			zap.Error(err),
			zap.Uint32("api_id", uri.ID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除API：删除API成功",
		zap.Uint32("api_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	log.Info(
		"删除API：删除API成功",
		zap.Uint32("api_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

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
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"查询API：绑定APIID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"查询API：开始执行",
		zap.Uint32("api_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	findStepStart := time.Now()
	m, err := h.apiSvc.FindApiByID(ctx, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询API：查询失败",
			zap.Error(err),
			zap.Uint32("api_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"查询API：查询API成功",
		zap.Uint32("api_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
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
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Error(
			"查询API列表：绑定查询参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"查询API列表：开始执行",
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"查询API列表：参数详情",
		zap.Object("list_api_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.apiSvc.ListApi(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询API列表：查询API列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_api_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询API列表：查询API列表成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
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
