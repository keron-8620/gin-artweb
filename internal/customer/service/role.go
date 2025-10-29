package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "gitee.com/keion8620/go-dango-gin/api/customer/role"
	"gitee.com/keion8620/go-dango-gin/internal/customer/biz"
	"gitee.com/keion8620/go-dango-gin/pkg/common"
	"gitee.com/keion8620/go-dango-gin/pkg/errors"
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
