package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbHost "gin-artweb/api/resource/host"
	"gin-artweb/internal/resource/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type HostService struct {
	log    *zap.Logger
	ucHost *biz.HostUsecase
}

func NewHostService(
	logger *zap.Logger,
	ucHost *biz.HostUsecase,
) *HostService {
	return &HostService{
		log:    logger,
		ucHost: ucHost,
	}
}

// @Summary 创建主机
// @Description 本接口用于创建新的主机配置信息
// @Tags 主机管理
// @Accept json,application/x-www-form-urlencoded,multipart/form-data
// @Produce json
// @Param request body pbHost.CreateOrUpdateHosrRequest true "创建主机请求"
// @Success 201 {object} pbHost.HostReply "创建主机成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host [post]
// @Security ApiKeyAuth
func (s *HostService) CreateHost(ctx *gin.Context) {
	var req pbHost.CreateOrUpdateHosrRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建主机请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始创建主机",
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucHost.CreateHost(ctx, biz.HostModel{
		Name:    req.Name,
		Label:   req.Label,
		SSHIP:   req.SSHIP,
		SSHPort: req.SSHPort,
		SSHUser: req.SSHUser,
		PyPath:  req.PyPath,
		Remark:  req.Remark,
	}, req.SSHPassword)
	if err != nil {
		s.log.Error(
			"创建主机失败",
			zap.Error(err),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"创建主机成功",
		zap.Uint32(biz.HostIDKey, m.ID),
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := HostModelToStandardOut(*m)
	ctx.JSON(http.StatusCreated, &pbHost.HostReply{
		Code: http.StatusCreated,
		Data: *mo,
	})
}

// @Summary 更新主机
// @Description 本接口用于更新指定ID的主机配置信息
// @Tags 主机管理
// @Accept json,application/x-www-form-urlencoded,multipart/form-data
// @Produce json
// @Param pk path uint true "主机编号"
// @Param request body pbHost.CreateOrUpdateHosrRequest true "更新主机请求"
// @Success 200 {object} pbHost.HostReply "更新主机成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{pk} [put]
// @Security ApiKeyAuth
func (s *HostService) UpdateHost(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定主机ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbHost.CreateOrUpdateHosrRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新主机请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始更新主机",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucHost.UpdateHostById(ctx, biz.HostModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: uri.PK},
		},
		Name:    req.Name,
		Label:   req.Label,
		SSHIP:   req.SSHIP,
		SSHPort: req.SSHPort,
		SSHUser: req.SSHUser,
		PyPath:  req.PyPath,
		Remark:  req.Remark,
	}, req.SSHPassword)
	if err != nil {
		s.log.Error(
			"更新主机失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"更新主机成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := HostModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &pbHost.HostReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 删除主机
// @Description 本接口用于删除指定ID的主机配置信息
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param pk path uint true "主机编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{pk} [delete]
// @Security ApiKeyAuth
func (s *HostService) DeleteHost(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除主机ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除主机",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucHost.DeleteHostById(ctx, uri.PK); err != nil {
		s.log.Error(
			"删除主机失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除主机成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询主机详情
// @Description 本接口用于查询指定ID的主机详细信息
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param pk path uint true "主机编号"
// @Success 200 {object} pbHost.HostReply "获取主机详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{pk} [get]
// @Security ApiKeyAuth
func (s *HostService) GetHost(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询主机ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询主机详情",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucHost.FindHostById(ctx, uri.PK)
	if err != nil {
		s.log.Error(
			"查询主机详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询主机详情成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := HostModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &pbHost.HostReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询主机列表
// @Description 本接口用于查询主机配置信息列表
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param name query string false "主机名称"
// @Param label query string false "主机标签"
// @Param ip_addr query string false "IP地址"
// @Success 200 {object} pbHost.PagHostReply "成功返回主机列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host [get]
// @Security ApiKeyAuth
func (s *HostService) ListHost(ctx *gin.Context) {
	var req pbHost.ListHostRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询主机列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询主机列表",
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
	total, ms, err := s.ucHost.ListHost(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询主机列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询主机列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mbs := ListHostModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &pbHost.PagHostReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *HostService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/host", s.CreateHost)
	r.PUT("/host/:pk", s.UpdateHost)
	r.DELETE("/host/:pk", s.DeleteHost)
	r.GET("/host/:pk", s.GetHost)
	r.GET("/host", s.ListHost)
}

func HostModelToBaseOut(
	m biz.HostModel,
) *pbHost.HostBaseOut {
	return &pbHost.HostBaseOut{
		ID:      m.ID,
		Name:    m.Name,
		Label:   m.Label,
		SSHIP:   m.SSHIP,
		SSHPort: m.SSHPort,
		SSHUser: m.SSHUser,
		PyPath:  m.PyPath,
		Remark:  m.Remark,
	}
}

func HostModelToStandardOut(
	m biz.HostModel,
) *pbHost.HostStandardOut {
	return &pbHost.HostStandardOut{
		HostBaseOut: *HostModelToBaseOut(m),
		CreatedAt:   m.CreatedAt.String(),
		UpdatedAt:   m.UpdatedAt.String(),
	}
}

func ListHostModelToStandardOut(
	hms *[]biz.HostModel,
) *[]pbHost.HostStandardOut {
	if hms == nil {
		return &[]pbHost.HostStandardOut{}
	}

	ms := *hms
	mso := make([]pbHost.HostStandardOut, 0, len(ms))
	for _, m := range ms {
		mo := HostModelToStandardOut(m)
		mso = append(mso, *mo)
	}
	return &mso
}
