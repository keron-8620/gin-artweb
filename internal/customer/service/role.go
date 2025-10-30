package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "gin-artweb/api/customer/role"
	"gin-artweb/internal/customer/biz"
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
// @Param request body pb.CreateRoleRequest true "创建角色请求"
// @Success 200 {object} pb.RoleReply "成功返回角色信息"
// @Router /api/v1/customer/role [post]
func (s *RoleService) CreateRole(ctx *gin.Context) {
	var req pb.CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucRole.CreateRole(
		ctx,
		req.PermissionIds,
		req.MenuIds,
		req.ButtonIds,
		biz.RoleModel{
			Name:  req.Name,
			Descr: req.Descr,
		},
	)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusCreated, &pb.RoleReply{
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
// @Param request body pb.UpdateRoleRequest true "更新角色请求"
// @Success 200 {object} pb.RoleReply "成功返回角色信息"
// @Router /api/v1/customer/role/{pk} [put]
func (s *RoleService) UpdateRole(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pb.UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucRole.UpdateRoleById(
		ctx, uri.Pk,
		req.PermissionIds,
		req.MenuIds,
		req.ButtonIds,
		map[string]any{
			"Name":  req.Name,
			"Descr": req.Descr,
		},
	); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	m, err := s.ucRole.FindRoleById(ctx, []string{"Permissions", "Menus", "Buttons"}, uri.Pk)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, &pb.RoleReply{
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
// @Router /api/v1/customer/role/{pk} [delete]
func (s *RoleService) DeleteRole(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucRole.DeleteRoleById(ctx, uri.Pk); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, common.NoDataReply)
}

// @Summary 查询单个角色
// @Description 本接口用于查询一个角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param pk path uint true "角色编号"
// @Success 200 {object} pb.RoleReply "成功返回角色信息"
// @Router /api/v1/customer/role/{pk} [get]
func (s *RoleService) GetRole(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucRole.FindRoleById(ctx, []string{"Permissions", "Menus", "Buttons"}, uri.Pk)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, &pb.RoleReply{
		Code: http.StatusOK,
		Data: *RoleModelToOut(*m),
	})
}

// @Summary 查询角色列表
// @Description 本接口用于查询角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Success 200 {object} pb.PagRoleBaseReply "成功返回角色列表"
// @Router /api/v1/customer/button [get]
func (s *RoleService) ListRole(ctx *gin.Context) {
	var req pb.ListRoleRequest
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
	ctx.JSON(http.StatusOK, &pb.PagRoleBaseReply{
		Code: http.StatusOK,
		Data: common.NewPag(page, size, total, mbs),
	})
}

func (s *RoleService) TreeRolePermissions(ctx *gin.Context) {
}

func (s *RoleService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/role", s.CreateRole)
	r.PUT("/role/:pk", s.UpdateRole)
	r.DELETE("/role/:pk", s.DeleteRole)
	r.GET("/role/:pk", s.GetRole)
	r.GET("/role", s.ListRole)
	r.GET("/role/permission/:pk", s.TreeRolePermissions)
}

func RoleModelToOutBase(
	m biz.RoleModel,
) *pb.RoleOutBase {
	return &pb.RoleOutBase{
		Id:       m.Id,
		CreatedAt: m.CreatedAt.String(),
		UpdatedAt: m.UpdatedAt.String(),
		Name:     m.Name,
		Descr:    m.Descr,
	}
}

func ListRoleModelToOutBase(
	ms []biz.RoleModel,
) []*pb.RoleOutBase {
	mso := make([]*pb.RoleOutBase, 0, len(ms))
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
) *pb.RoleOut {
	return &pb.RoleOut{
		RoleOutBase: *RoleModelToOutBase(m),
		Permissions: ListPermModelToOut(m.Permissions),
		Menus:       ListMenuModelToOutBase(m.Menus),
		Buttons:     ListButtonModelToOutBase(m.Buttons),
	}
}
