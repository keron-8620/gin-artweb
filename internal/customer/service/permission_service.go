package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbPerm "gin-artweb/api/customer/permission"
	"gin-artweb/internal/customer/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
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
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建权限请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始创建权限",
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

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
		s.log.Error(
			"创建权限失败",
			zap.Error(err),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"创建权限成功",
		zap.Uint32(biz.PermissionIDKey, m.ID),
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := PermModelToStandardOut(*m)
	ctx.JSON(http.StatusCreated, &pbPerm.PermissionReply{
		Code: http.StatusCreated,
		Data: mo,
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
		s.log.Error(
			"绑定权限ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbPerm.UpdatePermissionRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新权限请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始更新权限",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucPerm.UpdatePermissionByID(ctx, uri.PK, map[string]any{
		"url":    req.URL,
		"method": req.Method,
		"label":  req.Label,
		"descr":  req.Descr,
	})
	if err != nil {
		s.log.Error(
			"更新权限失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"更新权限成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := PermModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &pbPerm.PermissionReply{
		Code: http.StatusOK,
		Data: mo,
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
		s.log.Error(
			"绑定删除权限ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除权限",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucPerm.DeletePermissionByID(ctx, uri.PK); err != nil {
		s.log.Error(
			"删除权限失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除权限成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

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
		s.log.Error(
			"绑定查询权限ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询权限详情",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucPerm.FindPermissionByID(ctx, uri.PK)
	if err != nil {
		s.log.Error(
			"查询权限详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询权限详情成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := PermModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &pbPerm.PermissionReply{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 查询权限列表
// @Description 本接口用于查询权限列表
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param request query pbPerm.ListPermissionRequest false "查询参数"
// @Success 200 {object} pbPerm.PagPermissionReply "成功返回权限列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "内部服务错误"
// @Router /api/v1/customer/permission [get]
// @Security ApiKeyAuth
func (s *PermissionService) ListPermission(ctx *gin.Context) {
	var req pbPerm.ListPermissionRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询权限列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询权限列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Limit:   size,
		Offset:  page,
		OrderBy: []string{"id ASC"},
		Query:   query,
	}
	total, ms, err := s.ucPerm.ListPermission(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询权限列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询权限列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mbs := ListPermModelToStandardOut(ms)
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

func PermModelToStandardOut(
	m biz.PermissionModel,
) *pbPerm.PermissionStandardOut {
	return &pbPerm.PermissionStandardOut{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.String(),
		UpdatedAt: m.UpdatedAt.String(),
		URL:       m.URL,
		Method:    m.Method,
		Label:     m.Label,
		Descr:     m.Descr,
	}
}

func ListPermModelToStandardOut(
	pms *[]biz.PermissionModel,
) *[]pbPerm.PermissionStandardOut {
	if pms == nil {
		return &[]pbPerm.PermissionStandardOut{}
	}

	ms := *pms
	mso := make([]pbPerm.PermissionStandardOut, 0, len(ms))
	for _, m := range ms {
		mo := PermModelToStandardOut(m)
		mso = append(mso, *mo)
	}
	return &mso
}
