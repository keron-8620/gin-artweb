package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbApi "gin-artweb/api/customer/api"
	"gin-artweb/internal/infra/customer/biz"
	"gin-artweb/internal/infra/customer/model"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type ApiService struct {
	log   *zap.Logger
	ucApi *biz.ApiUsecase
}

func NewApiService(
	logger *zap.Logger,
	ucApi *biz.ApiUsecase,
) *ApiService {
	return &ApiService{
		log:   logger,
		ucApi: ucApi,
	}
}

// @Summary 新增API
// @Description 本接口用于新增API
// @Tags API管理
// @Accept json
// @Produce json
// @Param request body pbApi.CreateApiRequest true "创建API请求"
// @Success 201 {object} pbApi.ApiReply "创建API成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Router /api/v1/customer/api [post]
// @Security ApiKeyAuth
func (s *ApiService) CreateApi(ctx *gin.Context) {
	var req pbApi.CreateApiRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建API请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始创建API",
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucApi.CreateApi(ctx, model.ApiModel{
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
			"创建API失败",
			zap.Error(err),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	s.log.Info(
		"创建API成功",
		zap.Uint32("api_id", m.ID),
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := ApiModelToStandardOut(*m)
	ctx.JSON(http.StatusCreated, &pbApi.ApiReply{
		Code: http.StatusCreated,
		Data: mo,
	})
}

// @Summary 更新API
// @Description 本接口用于更新指定ID的API
// @Tags API管理
// @Accept json
// @Produce json
// @Param id path uint true "API编号"
// @Param request body pbApi.UpdateApiRequest true "更新API请求"
// @Success 200 {object} pbApi.ApiReply "更新API成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "API未找到"
// @Router /api/v1/customer/api/{id} [put]
// @Security ApiKeyAuth
func (s *ApiService) UpdateApi(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定APIID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req pbApi.UpdateApiRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新API请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始更新API",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucApi.UpdateApiByID(ctx, uri.ID, map[string]any{
		"url":    req.URL,
		"method": req.Method,
		"label":  req.Label,
		"descr":  req.Descr,
	})
	if err != nil {
		s.log.Error(
			"更新API失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	s.log.Info(
		"更新API成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := ApiModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &pbApi.ApiReply{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 删除API
// @Description 本接口用于删除指定ID的API
// @Tags API管理
// @Accept json
// @Produce json
// @Param id path uint true "API编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "API未找到"
// @Router /api/v1/customer/api/{id} [delete]
// @Security ApiKeyAuth
func (s *ApiService) DeleteApi(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除ApiID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始删除API",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.ucApi.DeleteApiByID(ctx, uri.ID); err != nil {
		s.log.Error(
			"删除API失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	s.log.Info(
		"删除API成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询API
// @Description 本接口用于查询指定ID的API
// @Tags API管理
// @Accept json
// @Produce json
// @Param id path uint true "API编号"
// @Success 200 {object} pbApi.ApiReply "获取API详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "API未找到"
// @Router /api/v1/customer/api/{id} [get]
// @Security ApiKeyAuth
func (s *ApiService) GetApi(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询ApiID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始查询API详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucApi.FindApiByID(ctx, uri.ID)
	if err != nil {
		s.log.Error(
			"查询API详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	s.log.Info(
		"查询API详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := ApiModelToStandardOut(*m)
	ctx.JSON(http.StatusOK, &pbApi.ApiReply{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 查询API列表
// @Description 本接口用于查询API列表
// @Tags API管理
// @Accept json
// @Produce json
// @Param request query pbApi.ListApiRequest false "查询参数"
// @Success 200 {object} pbApi.PagApiReply "成功返回API列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "内部服务错误"
// @Router /api/v1/customer/api [get]
// @Security ApiKeyAuth
func (s *ApiService) ListApi(ctx *gin.Context) {
	var req pbApi.ListApiRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询API列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	s.log.Info(
		"开始查询API列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
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
	total, ms, err := s.ucApi.ListApi(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询API列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	s.log.Info(
		"查询API列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := ListApiModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &pbApi.PagApiReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *ApiService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/api", s.CreateApi)
	r.PUT("/api/:id", s.UpdateApi)
	r.DELETE("/api/:id", s.DeleteApi)
	r.GET("/api/:id", s.GetApi)
	r.GET("/api", s.ListApi)
}

func ApiModelToStandardOut(
	m model.ApiModel,
) *pbApi.ApiStandardOut {
	return &pbApi.ApiStandardOut{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.Format(time.DateTime),
		UpdatedAt: m.UpdatedAt.Format(time.DateTime),
		URL:       m.URL,
		Method:    m.Method,
		Label:     m.Label,
		Descr:     m.Descr,
	}
}

func ListApiModelToStandardOut(
	pms *[]model.ApiModel,
) *[]pbApi.ApiStandardOut {
	if pms == nil {
		return &[]pbApi.ApiStandardOut{}
	}

	ms := *pms
	mso := make([]pbApi.ApiStandardOut, 0, len(ms))
	for _, m := range ms {
		mo := ApiModelToStandardOut(m)
		mso = append(mso, *mo)
	}
	return &mso
}
