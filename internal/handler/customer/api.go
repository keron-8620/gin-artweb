package customer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	custmodel "gin-artweb/internal/model/customer"
	custsvc "gin-artweb/internal/service/customer"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type ApiHandler struct {
	log    *zap.Logger
	svcApi *custsvc.ApiService
}

func NewApiHandler(
	logger *zap.Logger,
	svcApi *custsvc.ApiService,
) *ApiHandler {
	return &ApiHandler{
		log:    logger,
		svcApi: svcApi,
	}
}

// @Summary 新增API
// @Description 本接口用于新增API
// @Tags API管理
// @Accept json
// @Produce json
// @Param request body custmodel.CreateApiRequest true "创建API请求"
// @Success 201 {object} custmodel.ApiReply "创建API成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Router /api/v1/customer/api [post]
// @Security ApiKeyAuth
func (h *ApiHandler) CreateApi(ctx *gin.Context) {
	var req custmodel.CreateApiRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定创建API请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始创建API",
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcApi.CreateApi(ctx, custmodel.ApiModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: req.ID},
		},
		URL:    req.URL,
		Method: req.Method,
		Label:  req.Label,
		Descr:  req.Descr,
	})
	if err != nil {
		h.log.Error(
			"创建API失败",
			zap.Error(err),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"创建API成功",
		zap.Uint32("api_id", m.ID),
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := custmodel.ApiModelToStandardOut(*m)
	ctx.JSON(http.StatusCreated, &custmodel.ApiReply{
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
// @Param request body custmodel.UpdateApiRequest true "更新API请求"
// @Success 200 {object} custmodel.ApiReply "更新API成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "API未找到"
// @Router /api/v1/customer/api/{id} [put]
// @Security ApiKeyAuth
func (h *ApiHandler) UpdateApi(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定APIID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req custmodel.UpdateApiRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定更新API请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始更新API",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcApi.UpdateApiByID(ctx, uri.ID, map[string]any{
		"url":    req.URL,
		"method": req.Method,
		"label":  req.Label,
		"descr":  req.Descr,
	})
	if err != nil {
		h.log.Error(
			"更新API失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"更新API成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := custmodel.ApiModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &custmodel.ApiReply{
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
// @Success 200 {object} commodel.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "API未找到"
// @Router /api/v1/customer/api/{id} [delete]
// @Security ApiKeyAuth
func (h *ApiHandler) DeleteApi(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定删除ApiID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始删除API",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcApi.DeleteApiByID(ctx, uri.ID); err != nil {
		h.log.Error(
			"删除API失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"删除API成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 查询API
// @Description 本接口用于查询指定ID的API
// @Tags API管理
// @Accept json
// @Produce json
// @Param id path uint true "API编号"
// @Success 200 {object} custmodel.ApiReply "获取API详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "API未找到"
// @Router /api/v1/customer/api/{id} [get]
// @Security ApiKeyAuth
func (h *ApiHandler) GetApi(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询ApiID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询API详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcApi.FindApiByID(ctx, uri.ID)
	if err != nil {
		h.log.Error(
			"查询API详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询API详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := custmodel.ApiModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &custmodel.ApiReply{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 查询API列表
// @Description 本接口用于查询API列表
// @Tags API管理
// @Accept json
// @Produce json
// @Param request query custmodel.ListApiRequest false "查询参数"
// @Success 200 {object} custmodel.PagApiReply "成功返回API列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "内部服务错误"
// @Router /api/v1/customer/api [get]
// @Security ApiKeyAuth
func (h *ApiHandler) ListApi(ctx *gin.Context) {
	var req custmodel.ListApiRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询API列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询API列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Size:    size,
		Page:    page,
		OrderBy: []string{"id ASC"},
		Query:   query,
	}
	total, ms, err := h.svcApi.ListApi(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询API列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询API列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := custmodel.ListApiModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &custmodel.PagApiReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (h *ApiHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/api", h.CreateApi)
	r.PUT("/api/:id", h.UpdateApi)
	r.DELETE("/api/:id", h.DeleteApi)
	r.GET("/api/:id", h.GetApi)
	r.GET("/api", h.ListApi)
}
