package resource

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	resomodel "gin-artweb/internal/model/resource"
	resosvc "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type HostHandler struct {
	log     *zap.Logger
	svcHost *resosvc.HostService
}

func NewHostHandler(
	logger *zap.Logger,
	svcHost *resosvc.HostService,
) *HostHandler {
	return &HostHandler{
		log:     logger,
		svcHost: svcHost,
	}
}

// @Summary 创建主机
// @Description 本接口用于创建新的主机配置信息
// @Tags 主机管理
// @Accept json,application/x-www-form-urlencoded,multipart/form-data
// @Produce json
// @Param request body resomodel.CreateOrUpdateHosrRequest true "创建主机请求"
// @Success 201 {object} resomodel.HostReply "创建主机成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host [post]
// @Security ApiKeyAuth
func (h *HostHandler) CreateHost(ctx *gin.Context) {
	var req resomodel.CreateOrUpdateHosrRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定创建主机请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始创建主机",
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcHost.CreateHost(ctx, resomodel.HostModel{
		Name:    req.Name,
		Label:   req.Label,
		SSHIP:   req.SSHIP,
		SSHPort: req.SSHPort,
		SSHUser: req.SSHUser,
		PyPath:  req.PyPath,
		Remark:  req.Remark,
	}, req.SSHPassword)
	if err != nil {
		h.log.Error(
			"创建主机失败",
			zap.Error(err),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"创建主机成功",
		zap.Uint32("host_id", m.ID),
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := resomodel.HostModelToStandardOut(*m)
	ctx.JSON(http.StatusCreated, &resomodel.HostReply{
		Code: http.StatusCreated,
		Data: *mo,
	})
}

// @Summary 更新主机
// @Description 本接口用于更新指定ID的主机配置信息
// @Tags 主机管理
// @Accept json,application/x-www-form-urlencoded,multipart/form-data
// @Produce json
// @Param id path uint true "主机编号"
// @Param request body resomodel.CreateOrUpdateHosrRequest true "更新主机请求"
// @Success 200 {object} resomodel.HostReply "更新主机成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{id} [put]
// @Security ApiKeyAuth
func (h *HostHandler) UpdateHost(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定主机ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req resomodel.CreateOrUpdateHosrRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定更新主机请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始更新主机",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcHost.UpdateHostById(ctx, resomodel.HostModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: uri.ID},
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
		h.log.Error(
			"更新主机失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"更新主机成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := resomodel.HostModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &resomodel.HostReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 删除主机
// @Description 本接口用于删除指定ID的主机配置信息
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param id path uint true "主机编号"
// @Success 200 {object} commodel.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{id} [delete]
// @Security ApiKeyAuth
func (h *HostHandler) DeleteHost(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定删除主机ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始删除主机",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcHost.DeleteHostById(ctx, uri.ID); err != nil {
		h.log.Error(
			"删除主机失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"删除主机成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 查询主机详情
// @Description 本接口用于查询指定ID的主机详细信息
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param id path uint true "主机编号"
// @Success 200 {object} resomodel.HostReply "获取主机详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "主机未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host/{id} [get]
// @Security ApiKeyAuth
func (h *HostHandler) GetHost(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询主机ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询主机详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcHost.FindHostById(ctx, uri.ID)
	if err != nil {
		h.log.Error(
			"查询主机详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询主机详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := resomodel.HostModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &resomodel.HostReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询主机列表
// @Description 本接口用于查询主机配置信息列表
// @Tags 主机管理
// @Accept json
// @Produce json
// @Param request query resomodel.ListHostRequest false "查询参数"
// @Success 200 {object} resomodel.PagHostReply "成功返回主机列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/resource/host [get]
// @Security ApiKeyAuth
func (h *HostHandler) ListHost(ctx *gin.Context) {
	var req resomodel.ListHostRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询主机列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询主机列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Size:    size,
		Page:    page,
		OrderBy: []string{"id ASC"},
		Query:   query,
	}
	total, ms, err := h.svcHost.ListHost(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询主机列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询主机列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := resomodel.ListHostModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &resomodel.PagHostReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (h *HostHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/host", h.CreateHost)
	r.PUT("/host/:id", h.UpdateHost)
	r.DELETE("/host/:id", h.DeleteHost)
	r.GET("/host/:id", h.GetHost)
	r.GET("/host", h.ListHost)
}
