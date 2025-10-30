package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "gin-artweb/api/customer/permission"
	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
)

type PermissionService struct {
	log    *zap.Logger
	ucPerm *biz.PermissionUsecase
}

func NewPermissionService(
	logger *zap.Logger,
	ucPerm *biz.PermissionUsecase,
) *PermissionService {
	return &PermissionService{
		log:    logger,
		ucPerm: ucPerm,
	}
}

// @Summary 新增权限
// @Description 本接口用于新增权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param request body pb.CreatePermissionRequest true "创建权限请求"
// @Success 200 {object} pb.PermissionReply "成功返回权限信息"
// @Router /api/v1/customer/permission [post]
func (s *PermissionService) CreatePermission(ctx *gin.Context) {
	var req pb.CreatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucPerm.CreatePermission(ctx, biz.PermissionModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{Id: req.Id},
		},
		Url:    req.Url,
		Method: req.Method,
		Descr:  req.Descr,
	})
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := PermModelToOutBase(*m)
	ctx.JSON(http.StatusCreated, &pb.PermissionReply{
		Code: http.StatusCreated,
		Data: *mo,
	})
}

// @Summary 更新权限
// @Description 本接口用于更新权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param pk path uint true "权限编号"
// @Param request body pb.UpdatePermissionRequest true "更新权限请求"
// @Success 200 {object} pb.PermissionReply "成功返回权限信息"
// @Router /api/v1/customer/permission/{pk} [put]
func (s *PermissionService) UpdatePermission(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pb.UpdatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucPerm.UpdatePermissionById(ctx, uri.Pk, map[string]any{
		"Url":    req.Url,
		"Method": req.Method,
		"Descr":  req.Descr,
	}); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	m, err := s.ucPerm.FindPermissionById(ctx, uri.Pk)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := PermModelToOutBase(*m)
	ctx.JSON(http.StatusOK, &pb.PermissionReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 删除权限
// @Description 本接口用于删除指定ID的权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param pk path uint true "权限编号"
// @Router /api/v1/customer/permission/{pk} [delete]
func (s *PermissionService) DeletePermission(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucPerm.DeletePermissionById(ctx, uri.Pk); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, common.NoDataReply)
}

// @Summary 查询单个权限
// @Description 本接口用于查询一个权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param pk path uint true "权限编号"
// @Success 200 {object} pb.PermissionReply "成功返回用户信息"
// @Router /api/v1/customer/permission/{pk} [get]
func (s *PermissionService) GetPermission(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucPerm.FindPermissionById(ctx, uri.Pk)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := PermModelToOutBase(*m)
	ctx.JSON(http.StatusOK, &pb.PermissionReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询权限列表
// @Description 本接口用于查询权限列表
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param pk query uint false "权限主键，可选参数，如果提供则必须大于0"
// @Param pks query string false "权限主键列表，可选参数，多个用,隔开，如1,2,3"
// @Param before_create_at query string false "创建时间之前的记录 (RFC3339格式)"
// @Param after_create_at query string false "创建时间之后的记录 (RFC3339格式)"
// @Param before_update_at query string false "更新时间之前的记录 (RFC3339格式)"
// @Param after_update_at query string false "更新时间之后的记录 (RFC3339格式)"
// @Param http_url query string false "HTTP路径"
// @Param method query string false "HTTP方法"
// @Success 200 {object} pb.PagPermissionReply "成功返回权限列表"
// @Router /api/v1/customer/permission [get]
func (s *PermissionService) ListPermission(ctx *gin.Context) {
	var req pb.ListPermissionRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	page, size, query := req.Query()
	total, ms, err := s.ucPerm.ListPermission(ctx, page, size, query, []string{"id"}, true)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mbs := ListPermModelToOut(ms)
	ctx.JSON(http.StatusOK, &pb.PagPermissionReply{
		Code: http.StatusOK,
		Data: common.NewPag(page, size, total, mbs),
	})
}

func (s *PermissionService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/permission", s.CreatePermission)
	r.PUT("/permission/:pk", s.UpdatePermission)
	r.DELETE("/permission/:pk", s.DeletePermission)
	r.GET("/permission/:pk", s.GetPermission)
	r.GET("/permission", s.ListPermission)
}

func PermModelToOutBase(
	m biz.PermissionModel,
) *pb.PermissionOutBase {
	return &pb.PermissionOutBase{
		Id:        m.Id,
		CreatedAt: m.CreatedAt.String(),
		UpdatedAt: m.UpdatedAt.String(),
		Url:       m.Url,
		Method:    m.Method,
		Label:     m.Label,
		Descr:     m.Descr,
	}
}

func ListPermModelToOut(
	ms []biz.PermissionModel,
) []*pb.PermissionOutBase {
	mso := make([]*pb.PermissionOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := PermModelToOutBase(m)
			mso = append(mso, mo)
		}
	}
	return mso
}
