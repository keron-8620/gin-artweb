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

type ButtonHandler struct {
	log       *zap.Logger
	svcButton *custsvc.ButtonService
}

func NewButtonHandler(
	logger *zap.Logger,
	svcButton *custsvc.ButtonService,
) *ButtonHandler {
	return &ButtonHandler{
		log:       logger,
		svcButton: svcButton,
	}
}

// @Summary 新增按钮
// @Description 本接口用于新增按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param request body custmodel.CreateButtonRequest true "创建按钮请求"
// @Success 201 {object} custmodel.ButtonReply "成功返回按钮信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button [post]
// @Security ApiKeyAuth
func (h *ButtonHandler) CreateButton(ctx *gin.Context) {
	var req custmodel.CreateButtonRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定创建按钮请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始创建按钮",
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcButton.CreateButton(ctx, req.ApiIDs, custmodel.ButtonModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: req.ID},
		},
		Name:     req.Name,
		Sort:     req.Sort,
		IsActive: req.IsActive,
		Descr:    req.Descr,
		MenuID:   req.MenuID,
	})
	if err != nil {
		h.log.Error(
			"创建按钮失败",
			zap.Error(err),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"创建按钮成功",
		zap.Uint32(commodel.RequestIDKey, m.ID),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusCreated, &custmodel.ButtonReply{
		Code: http.StatusCreated,
		Data: custmodel.ButtonModelToDetailOut(*m),
	})
}

// @Summary 更新按钮
// @Description 本接口用于更新指定ID的按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param id path uint true "按钮编号"
// @Param request body custmodel.UpdateButtonRequest true "更新按钮请求"
// @Success 200 {object} custmodel.ButtonReply "成功返回按钮信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "按钮未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button/{id} [put]
// @Security ApiKeyAuth
func (h *ButtonHandler) UpdateButton(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定按钮ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req custmodel.UpdateButtonRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定更新按钮请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始更新按钮",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcButton.UpdateButtonByID(ctx, uri.ID, req.ApiIDs, map[string]any{
		"name":      req.Name,
		"sort":      req.Sort,
		"is_active": req.IsActive,
		"descr":     req.Descr,
		"menu_id":   req.MenuID,
	})
	if err != nil {
		h.log.Error(
			"更新按钮失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"更新按钮成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &custmodel.ButtonReply{
		Code: http.StatusOK,
		Data: custmodel.ButtonModelToDetailOut(*m),
	})
}

// @Summary 删除按钮
// @Description 本接口用于删除指定ID的按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param id path uint true "按钮编号"
// @Success 200 {object} commodel.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "按钮未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button/{id} [delete]
// @Security ApiKeyAuth
func (h *ButtonHandler) DeleteButton(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定删除按钮ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始删除按钮",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcButton.DeleteButtonByID(ctx, uri.ID); err != nil {
		h.log.Error(
			"删除按钮失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"删除按钮成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 查询按钮
// @Description 本接口用于查询指定ID的按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param id path uint true "按钮编号"
// @Success 200 {object} custmodel.ButtonReply "成功返回按钮信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "按钮未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button/{id} [get]
// @Security ApiKeyAuth
func (h *ButtonHandler) GetButton(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询按钮ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询按钮详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcButton.FindButtonByID(ctx, []string{"Apis", "Menu"}, uri.ID)
	if err != nil {
		h.log.Error(
			"查询按钮详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询按钮详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &custmodel.ButtonReply{
		Code: http.StatusOK,
		Data: custmodel.ButtonModelToDetailOut(*m),
	})
}

// @Summary 查询按钮列表
// @Description 本接口用于查询按钮列表
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param request query custmodel.ListButtonRequest false "查询参数"
// @Success 200 {object} custmodel.PagButtonReply "成功返回按钮列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button [get]
// @Security ApiKeyAuth
func (h *ButtonHandler) ListButton(ctx *gin.Context) {
	var req custmodel.ListButtonRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询按钮列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询按钮列表",
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
	total, ms, err := h.svcButton.ListButton(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询按钮列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询按钮列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := custmodel.ListButtonModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &custmodel.PagButtonReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (h *ButtonHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/button", h.CreateButton)
	r.PUT("/button/:id", h.UpdateButton)
	r.DELETE("/button/:id", h.DeleteButton)
	r.GET("/button/:id", h.GetButton)
	r.GET("/button", h.ListButton)
}
