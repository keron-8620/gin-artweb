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

type MenuHandler struct {
	log     *zap.Logger
	svcMenu *custsvc.MenuService
}

func NewMenuHandler(
	logger *zap.Logger,
	svcMenu *custsvc.MenuService,
) *MenuHandler {
	return &MenuHandler{
		log:     logger,
		svcMenu: svcMenu,
	}
}

// @Summary 新增菜单
// @Description 本接口用于新增菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param request body custmodel.CreateMenuRequest true "创建菜单请求"
// @Success 200 {object} custmodel.MenuReply "成功返回菜单信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu [post]
// @Security ApiKeyAuth
func (h *MenuHandler) CreateMenu(ctx *gin.Context) {
	var req custmodel.CreateMenuRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定创建菜单请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	if req.ParentID != nil && *req.ParentID == 0 {
		req.ParentID = nil
	}

	h.log.Info(
		"开始创建菜单",
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcMenu.CreateMenu(ctx, req.ApiIDs, custmodel.MenuModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: req.ID},
		},
		Path:      req.Path,
		Component: req.Component,
		Name:      req.Name,
		Meta: custmodel.MetaSchemas{
			Icon:  req.Meta.Icon,
			Title: req.Meta.Title,
		},
		Sort:     req.Sort,
		IsActive: req.IsActive,
		Descr:    req.Descr,
		ParentID: req.ParentID,
	})
	if err != nil {
		h.log.Error(
			"创建菜单失败",
			zap.Error(err),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"创建菜单成功",
		zap.Uint32(commodel.RequestIDKey, m.ID),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := custmodel.MenuModelToDetailOut(*m)
	ctx.JSON(http.StatusCreated, &custmodel.MenuReply{
		Code: http.StatusCreated,
		Data: mo,
	})
}

// @Summary 更新菜单
// @Description 本接口用于更新指定ID的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path uint true "菜单编号"
// @Param request body custmodel.UpdateMenuRequest true "更新菜单请求"
// @Success 200 {object} custmodel.MenuReply "成功返回菜单信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{id} [put]
// @Security ApiKeyAuth
func (h *MenuHandler) UpdateMenu(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定菜单ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req custmodel.UpdateMenuRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定更新菜单请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始更新菜单",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	data := map[string]any{
		"path":      req.Path,
		"component": req.Component,
		"name":      req.Name,
		"meta":      req.Meta.Json(),
		"sort":      req.Sort,
		"is_active": req.IsActive,
		"descr":     req.Descr,
	}
	if req.ParentID != nil && *req.ParentID != 0 {
		data["parent_id"] = req.ParentID
	}

	m, err := h.svcMenu.UpdateMenuByID(ctx, uri.ID, req.ApiIDs, data)
	if err != nil {
		h.log.Error(
			"更新菜单失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"更新菜单成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := custmodel.MenuModelToDetailOut(*m)
	ctx.JSON(http.StatusOK, &custmodel.MenuReply{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 删除菜单
// @Description 本接口用于删除指定ID的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path uint true "菜单编号"
// @Success 200 {object} commodel.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{id} [delete]
// @Security ApiKeyAuth
func (h *MenuHandler) DeleteMenu(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定删除菜单ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始删除菜单",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcMenu.DeleteMenuByID(ctx, uri.ID); err != nil {
		h.log.Error(
			"删除菜单失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"删除菜单成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 查询菜单
// @Description 本接口用于查询指定ID的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path uint true "菜单编号"
// @Success 200 {object} custmodel.MenuReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{id} [get]
// @Security ApiKeyAuth
func (h *MenuHandler) GetMenu(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询菜单ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询菜单详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcMenu.FindMenuByID(ctx, []string{"Parent", "Apis"}, uri.ID)
	if err != nil {
		h.log.Error(
			"查询菜单详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询菜单详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := custmodel.MenuModelToDetailOut(*m)
	ctx.JSON(http.StatusOK, &custmodel.MenuReply{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 查询菜单列表
// @Description 本接口用于查询菜单列表
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param request query custmodel.ListMenuRequest false "查询参数"
// @Success 200 {object} custmodel.PagMenuReply "成功返回菜单列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu [get]
// @Security ApiKeyAuth
func (h *MenuHandler) ListMenu(ctx *gin.Context) {
	var req custmodel.ListMenuRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询菜单列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询菜单列表",
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
	total, ms, err := h.svcMenu.ListMenu(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询菜单列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询菜单列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := custmodel.ListMenuModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &custmodel.PagMenuReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (h *MenuHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/menu", h.CreateMenu)
	r.PUT("/menu/:id", h.UpdateMenu)
	r.DELETE("/menu/:id", h.DeleteMenu)
	r.GET("/menu/:id", h.GetMenu)
	r.GET("/menu", h.ListMenu)
}
