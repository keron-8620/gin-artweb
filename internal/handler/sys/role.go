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

// RoleHandler 处理角色相关的请求
// 包含日志记录和角色服务的引用
type RoleHandler struct {
	log     *zap.Logger         // 日志记录器
	roleSvc *syssvc.RoleService // 角色服务
}

func NewRoleHandler(
	logger *zap.Logger,
	svcRole *syssvc.RoleService,
) *RoleHandler {
	return &RoleHandler{
		log:     logger,
		roleSvc: svcRole,
	}
}

// @Summary 新增角色
// @Description 本接口用于新增角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request body sysmodel.RoleUpsertDTO true "创建角色请求"
// @Success 201 {object} sysmodel.RoleResp "创建角色成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role [post]
// @Security ApiKeyAuth
func (h *RoleHandler) CreateRole(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.RoleUpsertDTO
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"新增角色：绑定参数失败",
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
		"创建角色：开始执行",
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"创建角色：入参详情",
		zap.Object("role_upsert_dto", &req),
	)

	createStepStart := time.Now()
	m, err := h.roleSvc.CreateRole(ctx, req)
	createStepDuration := time.Since(createStepStart)
	if err != nil {
		log.Error(
			"创建角色：创建角色失败",
			zap.Error(err),
			zap.Object("role_upsert_dto", &req),
			zap.Duration("create_step_duration", createStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"创建角色：创建角色成功",
		zap.Object("role_model", m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	log.Info(
		"创建角色：创建角色成功",
		zap.Uint32("role_id", m.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusCreated, &sysmodel.RoleResp{
		Code: http.StatusCreated,
		Data: sysmodel.RoleModelToDetailOut(*m),
	})
}

// @Summary 更新角色
// @Description 本接口用于更新指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path uint true "角色编号"
// @Param request body sysmodel.RoleUpsertDTO true "更新角色请求"
// @Success 200 {object} sysmodel.RoleResp "更新角色成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{id} [put]
// @Security ApiKeyAuth
func (h *RoleHandler) UpdateRole(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"更新角色：绑定角色ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req sysmodel.RoleUpsertDTO
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"更新角色：绑定请求参数失败",
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
		"更新角色：开始执行",
		zap.Uint32("role_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"更新角色：入参详情",
		zap.Uint32("role_id", uri.ID),
		zap.Object("role_upsert_dto", &req),
	)

	updateStepStart := time.Now()
	m, err := h.roleSvc.UpdateRoleByID(ctx, uri.ID, req)
	updateStepDuration := time.Since(updateStepStart)
	if err != nil {
		log.Error(
			"更新角色：更新角色失败",
			zap.Error(err),
			zap.Uint32("role_id", uri.ID),
			zap.Object("role_upsert_dto", &req),
			zap.Duration("update_step_duration", updateStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"更新角色：更新角色成功",
		zap.Object("role_model", m),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	log.Info(
		"更新角色：更新角色成功",
		zap.Uint32("role_id", uri.ID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &sysmodel.RoleResp{
		Code: http.StatusOK,
		Data: sysmodel.RoleModelToDetailOut(*m),
	})
}

// @Summary 删除角色
// @Description 本接口用于删除指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path uint true "角色编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{id} [delete]
// @Security ApiKeyAuth
func (h *RoleHandler) DeleteRole(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"删除角色：绑定角色ID参数失败",
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
		"删除角色：开始执行",
		zap.Uint32("role_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	deleteStepStart := time.Now()
	if err := h.roleSvc.DeleteRoleByID(ctx, uri.ID); err != nil {
		log.Error(
			"删除角色：删除角色失败",
			zap.Error(err),
			zap.Uint32("role_id", uri.ID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除角色：删除角色成功",
		zap.Uint32("role_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	log.Info(
		"删除角色：删除角色成功",
		zap.Uint32("role_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询角色
// @Description 本接口用于查询指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path uint true "角色编号"
// @Success 200 {object} sysmodel.RoleResp "获取角色详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{id} [get]
// @Security ApiKeyAuth
func (h *RoleHandler) GetRole(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"查询角色：绑定角色ID参数失败",
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
		"查询角色：开始执行",
		zap.Uint32("role_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	findStepStart := time.Now()
	m, err := h.roleSvc.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询角色：查询失败",
			zap.Error(err),
			zap.Uint32("role_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"查询角色：查询角色成功",
		zap.Uint32("role_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &sysmodel.RoleResp{
		Code: http.StatusOK,
		Data: sysmodel.RoleModelToDetailOut(*m),
	})
}

// @Summary 查询角色列表
// @Description 本接口用于查询角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request query sysmodel.ListRoleDTO false "查询参数"
// @Success 200 {object} sysmodel.PagRoleResp "成功返回角色列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role [get]
// @Security ApiKeyAuth
func (h *RoleHandler) ListRole(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.ListRoleDTO
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Error(
			"查询角色列表：绑定查询参数失败",
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
		"查询角色列表：开始执行",
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"查询角色列表：参数详情",
		zap.Object("list_role_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.roleSvc.ListRole(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询角色列表：查询角色列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_role_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Info(
		"查询角色列表：查询角色列表成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := sysmodel.ListRoleModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &sysmodel.PagRoleResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 获取当前用户菜单树
// @Description 本接口用于获取当前登录用户的菜单权限树
// @Tags 角色管理
// @Accept json
// @Produce json
// @Success 200 {object} sysmodel.RoleMenuTreeResp "成功返回菜单权限树"
// @Failure 401 {object} errors.Error "用户未认证"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/me/menu/tree [get]
// @Security ApiKeyAuth
func (h *RoleHandler) GetRoleMenuTree(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	claims := ctxutil.MustGetJwtClaims(ctx)

	log.Info(
		"查询角色菜单树：开始执行",
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	treeStepStart := time.Now()
	menuTrees, err := h.roleSvc.GetRoleMenuTree(ctx, claims.RoleID)
	treeStepDuration := time.Since(treeStepStart)
	if err != nil {
		log.Error(
			"查询角色菜单树：获取当前用户菜单树失败",
			zap.Error(err),
			zap.Uint32("role_id", claims.RoleID),
			zap.Uint32(ctxutil.UserIDKey, claims.UserID),
			zap.Duration("tree_step_duration", treeStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"查询角色菜单树：获取当前用户菜单树成功",
		zap.Uint32("role_id", claims.RoleID),
		zap.Uint32(ctxutil.UserIDKey, claims.UserID),
		zap.Duration("tree_step_duration", treeStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	log.Info(
		"查询角色菜单树：查询角色菜单树成功",
		zap.Uint32(ctxutil.UserIDKey, claims.UserID),
		zap.Uint32("role_id", claims.RoleID),
		zap.Duration("tree_step_duration", treeStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &sysmodel.RoleMenuTreeResp{
		Code: http.StatusOK,
		Data: menuTrees,
	})
}

// LoadRouter 注册角色相关的路由
func (h *RoleHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/role", h.CreateRole)
	r.PUT("/role/:id", h.UpdateRole)
	r.DELETE("/role/:id", h.DeleteRole)
	r.GET("/role/:id", h.GetRole)
	r.GET("/role", h.ListRole)
	r.GET("/me/menu/tree", h.GetRoleMenuTree)
}
