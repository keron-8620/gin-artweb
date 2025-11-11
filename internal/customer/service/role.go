package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbButton "gin-artweb/api/customer/button"
	pbMenu "gin-artweb/api/customer/menu"
	pbPerm "gin-artweb/api/customer/permission"
	pbRole "gin-artweb/api/customer/role"
	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/auth"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/errors"
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
// @Param request body pbRole.CreateRoleRequest true "创建角色请求"
// @Success 201 {object} pbRole.RoleReply "成功返回角色信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role [post]
func (s *RoleService) CreateRole(ctx *gin.Context) {
	var req pbRole.CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
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
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusCreated, &pbRole.RoleReply{
		Code: http.StatusCreated,
		Data: *RoleModelToOut(*m),
	})
}

// @Summary 更新角色
// @Description 本接口用于更新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param pk path uint true "角色编号"
// @Param request body pbRole.UpdateRoleRequest true "更新角色请求"
// @Success 200 {object} pbRole.RoleReply "成功返回角色信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{pk} [put]
func (s *RoleService) UpdateRole(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pbRole.UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucRole.UpdateRoleByID(
		ctx, uri.PK,
		req.PermissionIDs,
		req.MenuIDs,
		req.ButtonIDs,
		map[string]any{
			"name":  req.Name,
			"descr": req.Descr,
		},
	); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	m, err := s.ucRole.FindRoleByID(ctx, []string{"Permissions", "Menus", "Buttons"}, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, &pbRole.RoleReply{
		Code: http.StatusOK,
		Data: *RoleModelToOut(*m),
	})
}

// @Summary 删除角色
// @Description 本接口用于删除指定ID的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param pk path uint true "角色编号"
// @Success 200 {object} common.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{pk} [delete]
func (s *RoleService) DeleteRole(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucRole.DeleteRoleByID(ctx, uri.PK); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(common.NoDataReply.Code, common.NoDataReply)
}

// @Summary 查询单个角色
// @Description 本接口用于查询一个角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param pk path uint true "角色编号"
// @Success 200 {object} pbRole.RoleReply "成功返回角色信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/{pk} [get]
func (s *RoleService) GetRole(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucRole.FindRoleByID(ctx, []string{"Permissions", "Menus", "Buttons"}, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, &pbRole.RoleReply{
		Code: http.StatusOK,
		Data: *RoleModelToOut(*m),
	})
}

// @Summary 查询角色列表
// @Description 本接口用于查询角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param name query string false "角色名称"
// @Success 200 {object} pbRole.PagRoleBaseReply "成功返回角色列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role [get]
func (s *RoleService) ListRole(ctx *gin.Context) {
	var req pbRole.ListRoleRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	page, size, query := req.Query()
	total, ms, err := s.ucRole.ListRole(ctx, page, size, query, []string{"id"}, true, nil)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mbs := ListRoleModelToOutBase(ms)
	ctx.JSON(http.StatusOK, &pbRole.PagRoleBaseReply{
		Code: http.StatusOK,
		Data: common.NewPag(page, size, total, mbs),
	})
}

// @Summary 获取角色权限树
// @Description 本接口用于获取指定角色的权限树结构
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param pk path uint true "角色编号"
// @Success 200 {object} pbRole.RoleMenuTreeReply "成功返回权限树"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "角色未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/role/menutree [get]
func (s *RoleService) GetRoleMenuTree(ctx *gin.Context) {
	claims := auth.GetGinUserClaims(ctx)
	if claims == nil {
		ctx.JSON(auth.ErrGetUserClaims.Code, auth.ErrGetUserClaims.Reply())
		return
	}
	menuTree, err := s.ucRole.GetRoleMenuTree(ctx, claims.RoleID)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	roleMenuPerms := make([]*pbRole.RoleMenuPerm, 0)
	for _, node := range menuTree {
		perm := RoleMenuTreeToOut(node)
		if perm != nil {
			roleMenuPerms = append(roleMenuPerms, perm)
		}
	}
	ctx.JSON(http.StatusOK, &pbRole.RoleMenuTreeReply{
		Code: http.StatusOK,
		Data: roleMenuPerms,
	})
}

func (s *RoleService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/roleinfo", s.CreateRole)
	r.PUT("/roleinfo/:pk", s.UpdateRole)
	r.DELETE("/roleinfo/:pk", s.DeleteRole)
	r.GET("/roleinfo/:pk", s.GetRole)
	r.GET("/roleinfo", s.ListRole)
}

func RoleModelToOutBase(
	m biz.RoleModel,
) *pbRole.RoleOutBase {
	return &pbRole.RoleOutBase{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.String(),
		UpdatedAt: m.UpdatedAt.String(),
		Name:      m.Name,
		Descr:     m.Descr,
	}
}

func ListRoleModelToOutBase(
	ms []biz.RoleModel,
) []*pbRole.RoleOutBase {
	mso := make([]*pbRole.RoleOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := RoleModelToOutBase(m)
			mso = append(mso, mo)
		}
	}
	return mso
}

func RoleModelToOut(
	m biz.RoleModel,
) *pbRole.RoleOut {
	permissions := make([]*pbPerm.PermissionOutBase, 0)
	menus := make([]*pbMenu.MenuOutBase, 0)
	buttons := make([]*pbButton.ButtonOutBase, 0)

	if m.Permissions != nil {
		permissions = ListPermModelToOut(m.Permissions)
	}

	if m.Menus != nil {
		menus = ListMenuModelToOutBase(m.Menus)
	}

	if m.Buttons != nil {
		buttons = ListButtonModelToOutBase(m.Buttons)
	}
	return &pbRole.RoleOut{
		RoleOutBase: *RoleModelToOutBase(m),
		Permissions: permissions,
		Menus:       menus,
		Buttons:     buttons,
	}
}

// RoleMenuTreeToOut 将菜单树节点转换为 RoleMenuPerm 输出对象
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
	buttons := make([]pbButton.ButtonOutBase, 0, len(mt.Buttons))
	for _, button := range mt.Buttons {
		buttons = append(buttons, *ButtonModelToOutBase(button))
	}

	// 转换菜单基本信息
	menuBase := MenuModelToOutBase(mt.MenuModel)
	if menuBase == nil {
		return nil
	}
	return &pbRole.RoleMenuPerm{
		MenuOutBase: *menuBase,
		Buttons:     buttons,
		Children:    children,
	}
}
