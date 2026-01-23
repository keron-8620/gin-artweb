package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbButton "gin-artweb/api/customer/button"
	pbRole "gin-artweb/api/customer/role"
	"gin-artweb/internal/customer/biz"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

type RoleService struct {
	log    *zap.Logger
	ucRole *biz.RoleUsecase
}

func NewRoleService(
	logger *zap.Logger,
	ucRole *biz.RoleUsecase,
) *RoleService {
	return &RoleService{
		log:    logger,
		ucRole: ucRole,
	}
}

// @Summary 新增角色
// @Description 本接口用于新增角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request body pbRole.CreateOrUpdateRoleRequest true "创建角色请求"
// @Success 201 {object} pbRole.RoleReply "成功返回角色信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role [post]
// @Security ApiKeyAuth
func (s *RoleService) CreateRole(ctx *gin.Context) {
	var req pbRole.CreateOrUpdateRoleRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建角色请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始创建角色",
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucRole.CreateRole(
		ctx,
		req.PermissionIDs,
		req.MenuIDs,
		req.ButtonIDs,
		biz.RoleModel{
			Name:  req.Name,
			Descr: req.Descr,
		},
	)
	if err != nil {
		s.log.Error(
			"创建角色失败",
			zap.Error(err),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"创建角色成功",
		zap.Uint32(pbComm.RequestIDKey, m.ID),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusCreated, &pbRole.RoleReply{
		Code: http.StatusCreated,
		Data: RoleModelToDetailOut(*m),
	})
}

// @Summary 更新角色
// @Description 本接口用于更新指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path uint true "角色编号"
// @Param request body pbRole.CreateOrUpdateRoleRequest true "更新角色请求"
// @Success 200 {object} pbRole.RoleReply "成功返回角色信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{id} [put]
// @Security ApiKeyAuth
func (s *RoleService) UpdateRole(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定角色ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}
	var req pbRole.CreateOrUpdateRoleRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新角色请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始更新角色",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucRole.UpdateRoleByID(
		ctx, uri.ID,
		req.PermissionIDs,
		req.MenuIDs,
		req.ButtonIDs,
		map[string]any{
			"name":  req.Name,
			"descr": req.Descr,
		},
	)
	if err != nil {
		s.log.Error(
			"更新角色失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"更新角色成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &pbRole.RoleReply{
		Code: http.StatusOK,
		Data: RoleModelToDetailOut(*m),
	})
}

// @Summary 删除角色
// @Description 本接口用于删除指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path uint true "角色编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{id} [delete]
// @Security ApiKeyAuth
func (s *RoleService) DeleteRole(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除角色ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除角色",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := s.ucRole.DeleteRoleByID(ctx, uri.ID); err != nil {
		s.log.Error(
			"删除角色失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除角色成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询角色
// @Description 本接口用于查询指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path uint true "角色编号"
// @Success 200 {object} pbRole.RoleReply "成功返回角色信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{id} [get]
// @Security ApiKeyAuth
func (s *RoleService) GetRole(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询角色ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询角色详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucRole.FindRoleByID(ctx, []string{"Permissions", "Menus", "Buttons"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询角色详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询角色详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &pbRole.RoleReply{
		Code: http.StatusOK,
		Data: RoleModelToDetailOut(*m),
	})
}

// @Summary 查询角色列表
// @Description 本接口用于查询角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request query pbRole.ListRoleRequest false "查询参数"
// @Success 200 {object} pbRole.PagRoleReply "成功返回角色列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role [get]
// @Security ApiKeyAuth
func (s *RoleService) ListRole(ctx *gin.Context) {
	var req pbRole.ListRoleRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询角色列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询角色列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Limit:   size,
		Offset:  page,
		OrderBy: []string{"id ASC"},
		Query:   query,
	}
	total, ms, err := s.ucRole.ListRole(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询角色列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询角色列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	mbs := ListRoleModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &pbRole.PagRoleReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

// @Summary 获取当前用户菜单树
// @Description 本接口用于获取当前登录用户的菜单权限树
// @Tags 角色管理
// @Accept json
// @Produce json
// @Success 200 {object} pbRole.RoleMenuTreeReply "成功返回菜单权限树"
// @Failure 401 {object} errors.Error "用户未认证"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/me/menu/tree [get]
// @Security ApiKeyAuth
func (s *RoleService) GetRoleMenuTree(ctx *gin.Context) {
	claims := auth.GetUserClaims(ctx)
	if claims == nil {
		s.log.Error(
			"获取个人登录信息失败",
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(errors.ErrGetUserClaims.Code, errors.ErrGetUserClaims.ToMap())
		return
	}
	s.log.Info(
		"开始获取当前用户菜单树",
		zap.Uint32(auth.UserIDKey, claims.UserID),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	menuTree, err := s.ucRole.GetRoleMenuTree(ctx, claims.RoleID)
	if err != nil {
		s.log.Error(
			"获取当前用户菜单树失败",
			zap.Error(err),
			zap.Uint32(auth.UserIDKey, claims.UserID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}
	s.log.Info(
		"当前用户菜单树获取成功",
		zap.Uint32(auth.UserIDKey, claims.UserID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	roleMenuPerms := make([]pbRole.RoleMenuPerm, 0)
	for _, node := range menuTree {
		perm := RoleMenuTreeToOut(node)
		if perm != nil {
			roleMenuPerms = append(roleMenuPerms, *perm)
		}
	}
	ctx.JSON(http.StatusOK, &pbRole.RoleMenuTreeReply{
		Code: http.StatusOK,
		Data: &roleMenuPerms,
	})
}

func (s *RoleService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/role", s.CreateRole)
	r.PUT("/role/:id", s.UpdateRole)
	r.DELETE("/role/:id", s.DeleteRole)
	r.GET("/role/:id", s.GetRole)
	r.GET("/role", s.ListRole)
}

func RoleModelToBaseOut(
	m biz.RoleModel,
) *pbRole.RoleBaseOut {
	return &pbRole.RoleBaseOut{
		ID:    m.ID,
		Name:  m.Name,
		Descr: m.Descr,
	}
}

func RoleModelToStandardOut(
	m biz.RoleModel,
) *pbRole.RoleStandardOut {
	return &pbRole.RoleStandardOut{
		RoleBaseOut: *RoleModelToBaseOut(m),
		CreatedAt:   m.CreatedAt.Format(time.DateTime),
		UpdatedAt:   m.UpdatedAt.Format(time.DateTime),
	}
}

func RoleModelToDetailOut(
	m biz.RoleModel,
) *pbRole.RoleDetailOut {
	var permissionIDs = []uint32{}
	if len(m.Permissions) > 0 {
		permissionIDs = make([]uint32, len(m.Permissions))
		for i, p := range m.Permissions {
			permissionIDs[i] = p.ID
		}
	}

	var menuIDs = []uint32{}
	if len(m.Menus) > 0 {
		menuIDs = make([]uint32, len(m.Menus))
		for i, p := range m.Menus {
			menuIDs[i] = p.ID
		}
	}

	var buttonIDs = []uint32{}
	if len(m.Buttons) > 0 {
		buttonIDs = make([]uint32, len(m.Buttons))
		for i, p := range m.Buttons {
			buttonIDs[i] = p.ID
		}
	}
	return &pbRole.RoleDetailOut{
		RoleStandardOut: *RoleModelToStandardOut(m),
		PermissionIDs:   permissionIDs,
		MenuIDs:         menuIDs,
		ButtonIDs:       buttonIDs,
	}
}

func ListRoleModelToStandardOut(
	rms *[]biz.RoleModel,
) *[]pbRole.RoleStandardOut {
	if rms == nil {
		return &[]pbRole.RoleStandardOut{}
	}
	ms := *rms
	mso := make([]pbRole.RoleStandardOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := RoleModelToStandardOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}

// RoleMenuTreeToOut 将菜单树节点转换为RoleMenuPerm输出对象
func RoleMenuTreeToOut(
	mt *biz.MenuTreeNode,
) *pbRole.RoleMenuPerm {
	if mt == nil {
		return nil
	}

	// 转换子节点
	children := make([]pbRole.RoleMenuPerm, 0, len(mt.Children))
	for _, child := range mt.Children {
		childOut := RoleMenuTreeToOut(child)
		if childOut != nil {
			children = append(children, *childOut)
		}
	}

	// 转换按钮
	buttons := make([]pbButton.ButtonBaseOut, 0, len(mt.Buttons))
	for _, button := range mt.Buttons {
		buttons = append(buttons, *ButtonModelToBaseOut(button))
	}

	// 转换菜单基本信息
	menuBase := MenuModelToBaseOut(mt.MenuModel)
	if menuBase == nil {
		return nil
	}
	return &pbRole.RoleMenuPerm{
		MenuBaseOut: *menuBase,
		Buttons:     buttons,
		Children:    children,
	}
}
