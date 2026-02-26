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

type RoleHandler struct {
	log     *zap.Logger
	svcRole *custsvc.RoleService
}

func NewRoleHandler(
	logger *zap.Logger,
	svcRole *custsvc.RoleService,
) *RoleHandler {
	return &RoleHandler{
		log:     logger,
		svcRole: svcRole,
	}
}

// @Summary 新增角色
// @Description 本接口用于新增角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request body custmodel.CreateOrUpdateRoleRequest true "创建角色请求"
// @Success 201 {object} custmodel.RoleReply "成功返回角色信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role [post]
// @Security ApiKeyAuth
func (h *RoleHandler) CreateRole(ctx *gin.Context) {
	var req custmodel.CreateOrUpdateRoleRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定创建角色请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始创建角色",
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcRole.CreateRole(
		ctx,
		req.ApiIDs,
		req.MenuIDs,
		req.ButtonIDs,
		custmodel.RoleModel{
			Name:  req.Name,
			Descr: req.Descr,
		},
	)
	if err != nil {
		h.log.Error(
			"创建角色失败",
			zap.Error(err),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"创建角色成功",
		zap.Uint32(commodel.RequestIDKey, m.ID),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusCreated, &custmodel.RoleReply{
		Code: http.StatusCreated,
		Data: custmodel.RoleModelToDetailOut(*m),
	})
}

// @Summary 更新角色
// @Description 本接口用于更新指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path uint true "角色编号"
// @Param request body custmodel.CreateOrUpdateRoleRequest true "更新角色请求"
// @Success 200 {object} custmodel.RoleReply "成功返回角色信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{id} [put]
// @Security ApiKeyAuth
func (h *RoleHandler) UpdateRole(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定角色ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}
	var req custmodel.CreateOrUpdateRoleRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定更新角色请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始更新角色",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcRole.UpdateRoleByID(
		ctx, uri.ID,
		req.ApiIDs,
		req.MenuIDs,
		req.ButtonIDs,
		map[string]any{
			"name":  req.Name,
			"descr": req.Descr,
		},
	)
	if err != nil {
		h.log.Error(
			"更新角色失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"更新角色成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &custmodel.RoleReply{
		Code: http.StatusOK,
		Data: custmodel.RoleModelToDetailOut(*m),
	})
}

// @Summary 删除角色
// @Description 本接口用于删除指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path uint true "角色编号"
// @Success 200 {object} commodel.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{id} [delete]
// @Security ApiKeyAuth
func (h *RoleHandler) DeleteRole(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定删除角色ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始删除角色",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcRole.DeleteRoleByID(ctx, uri.ID); err != nil {
		h.log.Error(
			"删除角色失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"删除角色成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 查询角色
// @Description 本接口用于查询指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path uint true "角色编号"
// @Success 200 {object} custmodel.RoleReply "成功返回角色信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{id} [get]
// @Security ApiKeyAuth
func (h *RoleHandler) GetRole(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询角色ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询角色详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcRole.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, uri.ID)
	if err != nil {
		h.log.Error(
			"查询角色详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询角色详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &custmodel.RoleReply{
		Code: http.StatusOK,
		Data: custmodel.RoleModelToDetailOut(*m),
	})
}

// @Summary 查询角色列表
// @Description 本接口用于查询角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request query custmodel.ListRoleRequest false "查询参数"
// @Success 200 {object} custmodel.PagRoleReply "成功返回角色列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role [get]
// @Security ApiKeyAuth
func (h *RoleHandler) ListRole(ctx *gin.Context) {
	var req custmodel.ListRoleRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询角色列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询角色列表",
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
	total, ms, err := h.svcRole.ListRole(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询角色列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询角色列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := custmodel.ListRoleModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &custmodel.PagRoleReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 获取当前用户菜单树
// @Description 本接口用于获取当前登录用户的菜单权限树
// @Tags 角色管理
// @Accept json
// @Produce json
// @Success 200 {object} custmodel.RoleMenuTreeReply "成功返回菜单权限树"
// @Failure 401 {object} errors.Error "用户未认证"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/me/menu/tree [get]
// @Security ApiKeyAuth
func (h *RoleHandler) GetRoleMenuTree(ctx *gin.Context) {
	claims, rErr := ctxutil.GetUserClaims(ctx)
	if rErr != nil {
		h.log.Error(
			"获取个人登录信息失败",
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	h.log.Info(
		"开始获取当前用户菜单树",
		zap.Uint32(ctxutil.UserIDKey, claims.UserID),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	menuTrees, err := h.svcRole.GetRoleMenuTree(ctx, claims.RoleID)
	if err != nil {
		h.log.Error(
			"获取当前用户菜单树失败",
			zap.Error(err),
			zap.Uint32(ctxutil.UserIDKey, claims.UserID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	h.log.Info(
		"当前用户菜单树获取成功",
		zap.Uint32(ctxutil.UserIDKey, claims.UserID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	ctx.JSON(http.StatusOK, &custmodel.RoleMenuTreeReply{
		Code: http.StatusOK,
		Data: &menuTrees,
	})
}

func (h *RoleHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/role", h.CreateRole)
	r.PUT("/role/:id", h.UpdateRole)
	r.DELETE("/role/:id", h.DeleteRole)
	r.GET("/role/:id", h.GetRole)
	r.GET("/role", h.ListRole)
}
