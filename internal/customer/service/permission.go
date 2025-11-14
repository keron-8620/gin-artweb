package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbPerm "gin-artweb/api/customer/permission"
	"gin-artweb/internal/customer/biz"
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
// @Param request body pbPerm.CreatePermissionRequest true "创建权限请求"
// @Success 201 {object} pbPerm.PermissionReply "创建权限成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Router /api/v1/customer/permission [post]
// @Security ApiKeyAuth
func (s *PermissionService) CreatePermission(ctx *gin.Context) {
	var req pbPerm.CreatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucPerm.CreatePermission(ctx, biz.PermissionModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: req.ID},
		},
		URL:    req.URL,
		Method: req.Method,
		Label:  req.Label,
		Descr:  req.Descr,
	})
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := PermModelToOutBase(*m)
	ctx.JSON(http.StatusCreated, &pbPerm.PermissionReply{
		Code: http.StatusCreated,
		Data: *mo,
	})
}

// @Summary 更新权限
// @Description 本接口用于更新指定ID的权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param pk path uint true "权限编号"
// @Param request body pbPerm.UpdatePermissionRequest true "更新权限请求"
// @Success 200 {object} pbPerm.PermissionReply "更新权限成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "权限未找到"
// @Router /api/v1/customer/permission/{pk} [put]
// @Security ApiKeyAuth
func (s *PermissionService) UpdatePermission(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pbPerm.UpdatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucPerm.UpdatePermissionByID(ctx, uri.PK, map[string]any{
		"url":    req.URL,
		"method": req.Method,
		"descr":  req.Descr,
	}); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	m, err := s.ucPerm.FindPermissionByID(ctx, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := PermModelToOutBase(*m)
	ctx.JSON(http.StatusOK, &pbPerm.PermissionReply{
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
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "权限未找到"
// @Router /api/v1/customer/permission/{pk} [delete]
// @Security ApiKeyAuth
func (s *PermissionService) DeletePermission(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucPerm.DeletePermissionByID(ctx, uri.PK); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询权限
// @Description 本接口用于查询指定ID的权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param pk path uint true "权限编号"
// @Success 200 {object} pbPerm.PermissionReply "获取权限详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "权限未找到"
// @Router /api/v1/customer/permission/{pk} [get]
// @Security ApiKeyAuth
func (s *PermissionService) GetPermission(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucPerm.FindPermissionByID(ctx, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := PermModelToOutBase(*m)
	ctx.JSON(http.StatusOK, &pbPerm.PermissionReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询权限列表
// @Description 本接口用于查询权限列表
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param pk query uint false "权限主键，可选参数，如果提供则必须大于0"
// @Param pks query string false "权限主键列表，可选参数，多个用,隔开，如1,2,3"
// @Param before_create_at query string false "创建时间之前的记录 (RFC3339格式)"
// @Param after_create_at query string false "创建时间之后的记录 (RFC3339格式)"
// @Param before_update_at query string false "更新时间之前的记录 (RFC3339格式)"
// @Param after_update_at query string false "更新时间之后的记录 (RFC3339格式)"
// @Param url query string false "HTTP路径"
// @Param method query string false "HTTP方法" Enums(GET, POST, PUT, DELETE, PATCH)
// @Success 200 {object} pbPerm.PagPermissionReply "成功返回权限列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Router /api/v1/customer/permission [get]
// @Security ApiKeyAuth
func (s *PermissionService) ListPermission(ctx *gin.Context) {
	var req pbPerm.ListPermissionRequest
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
	ctx.JSON(http.StatusOK, &pbPerm.PagPermissionReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
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
) *pbPerm.PermissionOutBase {
	return &pbPerm.PermissionOutBase{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.String(),
		UpdatedAt: m.UpdatedAt.String(),
		URL:       m.URL,
		Method:    m.Method,
		Label:     m.Label,
		Descr:     m.Descr,
	}
}

func ListPermModelToOut(
	ms []*biz.PermissionModel,
) []*pbPerm.PermissionOutBase {
	mso := make([]*pbPerm.PermissionOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			if m != nil {
				mo := PermModelToOutBase(*m)
				mso = append(mso, mo)
			}
		}
	}
	return mso
}
