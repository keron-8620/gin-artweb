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
func (h *RoleHandler) CreateRole(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("新增角色:开始执行")

	var req sysmodel.RoleUpsertDTO
	if !common.ShouldBind(c, log, &req, "新增角色:绑定新增角色请求参数失败") {
		return
	}

	m, err := h.roleSvc.CreateRole(ctx, req)
	if err != nil {
		log.Error(
			"创建角色:执行失败",
			zap.Error(err),
			zap.Object("role_upsert_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"创建角色:执行成功",
		zap.Uint32("role_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusCreated, &sysmodel.RoleResp{
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
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("更新角色:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "更新角色:绑定角色ID参数失败") {
		return
	}

	var req sysmodel.RoleUpsertDTO
	if !common.ShouldBind(c, log, &req, "更新角色:绑定更新角色请求参数失败") {
		return
	}

	m, err := h.roleSvc.UpdateRoleByID(ctx, uri.ID, req)
	if err != nil {
		log.Error(
			"更新角色:执行失败",
			zap.Error(err),
			zap.Uint32("role_id", uri.ID),
			zap.Object("role_upsert_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"更新角色:执行成功",
		zap.Uint32("role_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	c.JSON(http.StatusOK, &sysmodel.RoleResp{
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
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("删除角色:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "删除角色:绑定角色ID参数失败") {
		return
	}

	if err := h.roleSvc.DeleteRoleByID(ctx, uri.ID); err != nil {
		log.Error(
			"删除角色:执行失败",
			zap.Error(err),
			zap.Uint32("role_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"删除角色:执行成功",
		zap.Uint32("role_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
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
func (h *RoleHandler) GetRole(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询角色:绑定角色ID参数失败") {
		return
	}

	m, err := h.roleSvc.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, uri.ID)
	if err != nil {
		log.Error(
			"查询角色:执行失败",
			zap.Error(err),
			zap.Uint32("role_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, &sysmodel.RoleResp{
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
func (h *RoleHandler) ListRole(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("查询角色列表:开始执行")

	var req sysmodel.ListRoleDTO
	if !common.ShouldBindQuery(c, log, &req, "查询角色列表:绑定查询参数失败") {
		return
	}

	page, size, total, ms, err := h.roleSvc.ListRole(ctx, req)
	if err != nil {
		log.Error(
			"查询角色列表:执行失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_role_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mbs := sysmodel.ListRoleModelToStandardOut(ms)
	c.JSON(http.StatusOK, &sysmodel.PagRoleResp{
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
func (h *RoleHandler) GetRoleMenuTree(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("查询个人角色菜单树:开始执行")

	claims := ctxutil.MustGetJwtClaims(ctx)
	menuTrees, err := h.roleSvc.GetRoleMenuTree(ctx, claims.RoleID)
	if err != nil {
		log.Error(
			"查询个人角色菜单树:执行失败",
			zap.Error(err),
			zap.Uint32("role_id", claims.RoleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, &sysmodel.RoleMenuTreeResp{
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
}
