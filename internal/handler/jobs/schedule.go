package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	jobsmodel "gin-artweb/internal/model/jobs"
	jobsvc "gin-artweb/internal/service/jobs"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type ScheduleHandler struct {
	log         *zap.Logger
	svcSchedule *jobsvc.ScheduleService
}

func NewScheduleHandler(
	logger *zap.Logger,
	svcSchedule *jobsvc.ScheduleService,
) *ScheduleHandler {
	return &ScheduleHandler{
		log:         logger,
		svcSchedule: svcSchedule,
	}
}

// @Summary 创建计划任务
// @Description 本接口用于创建新的计划任务
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param request body jobsmodel.CreateScheduleRequest true "创建计划任务请求"
// @Success 200 {object} jobsmodel.ScheduleReply "成功返回计划任务信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule [post]
// @Security ApiKeyAuth
func (h *ScheduleHandler) CreateSchedule(ctx *gin.Context) {
	var req jobsmodel.CreateScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error(
			"绑定创建计划任务参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	claims, rErr := ctxutil.GetUserClaims(ctx)
	if rErr != nil {
		h.log.Error(
			"获取个人登录信息失败",
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	schedule := jobsmodel.ScheduleModel{
		Name:          req.Name,
		Specification: req.Specification,
		IsEnabled:     req.IsEnabled,
		EnvVars:       req.EnvVars,
		CommandArgs:   req.CommandArgs,
		WorkDir:       req.WorkDir,
		Timeout:       req.Timeout,
		IsRetry:       req.IsRetry,
		RetryInterval: req.RetryInterval,
		MaxRetries:    req.MaxRetries,
		Username:      claims.Subject,
		ScriptID:      req.ScriptID,
	}

	m, rErr := h.svcSchedule.CreateSchedule(ctx, schedule)
	if rErr != nil {
		h.log.Error(
			"创建计划任务失败",
			zap.Error(rErr),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	ctx.JSON(http.StatusOK, &jobsmodel.ScheduleReply{
		Code: http.StatusOK,
		Data: *jobsmodel.ScheduleToDetailOut(*m),
	})
}

// @Summary 更新计划任务
// @Description 本接口用于更新指定ID的计划任务
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param id path uint true "计划任务编号"
// @Param request body jobsmodel.UpdateScheduleRequest true "更新计划任务请求"
// @Success 200 {object} jobsmodel.ScheduleReply "成功返回计划任务信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "计划任务未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule/{id} [put]
// @Security ApiKeyAuth
func (h *ScheduleHandler) UpdateSchedule(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定更新计划任务ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req jobsmodel.UpdateScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Error(
			"绑定更新计划任务参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	claims, rErr := ctxutil.GetUserClaims(ctx)
	if rErr != nil {
		h.log.Error(
			"获取个人登录信息失败",
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	data := map[string]any{
		"name":           req.Name,
		"specification":  req.Specification,
		"is_enabled":     req.IsEnabled,
		"env_vars":       req.EnvVars,
		"command_args":   req.CommandArgs,
		"work_dir":       req.WorkDir,
		"timeout":        req.Timeout,
		"is_retry":       req.IsRetry,
		"retry_interval": req.RetryInterval,
		"max_retries":    req.MaxRetries,
		"username":       claims.Subject,
		"script_id":      req.ScriptID,
	}

	m, err := h.svcSchedule.UpdateScheduleByID(ctx, uri.ID, data)
	if err != nil {
		h.log.Error(
			"更新计划任务失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, &jobsmodel.ScheduleReply{
		Code: http.StatusOK,
		Data: *jobsmodel.ScheduleToDetailOut(*m),
	})
}

// @Summary 删除计划任务
// @Description 本接口用于删除指定ID的计划任务
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param id path uint true "计划任务编号"
// @Success 200 {object} commodel.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "计划任务未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule/{id} [delete]
// @Security ApiKeyAuth
func (h *ScheduleHandler) DeleteSchedule(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定删除计划任务ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始删除计划任务",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcSchedule.DeleteScheduleByID(ctx, uri.ID); err != nil {
		h.log.Error(
			"删除计划任务失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"删除计划任务成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 查询计划任务详情
// @Description 本接口用于查询指定ID的计划任务详情
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param id path uint true "计划任务编号"
// @Success 200 {object} jobsmodel.ScheduleReply "成功返回计划任务信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "计划任务未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule/{id} [get]
// @Security ApiKeyAuth
func (h *ScheduleHandler) GetSchedule(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询计划任务ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询计划任务详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcSchedule.FindScheduleByID(ctx, []string{"Script"}, uri.ID)
	if err != nil {
		h.log.Error(
			"查询计划任务详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询计划任务详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mo := jobsmodel.ScheduleToDetailOut(*m)
	ctx.JSON(http.StatusOK, &jobsmodel.ScheduleReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询计划任务列表
// @Description 本接口用于查询计划任务列表
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param request query jobsmodel.ListScheduleRequest false "查询参数"
// @Success 200 {object} jobsmodel.PagScheduleReply "成功返回计划任务列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule [get]
// @Security ApiKeyAuth
func (h *ScheduleHandler) ListSchedule(ctx *gin.Context) {
	var req jobsmodel.ListScheduleRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询计划任务列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询计划任务列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"Script"},
		IsCount:  true,
		Size:     size,
		Page:     page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, err := h.svcSchedule.ListSchedule(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询计划任务列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询计划任务列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := jobsmodel.ListScheduledToDetailOut(ms)
	ctx.JSON(http.StatusOK, &jobsmodel.PagScheduleReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// func (s *ScheduleService) ListScheduleJobs(ctx *gin.Context) {
// }

// func (s *ScheduleService) ReoloadScheduleJobs(ctx *gin.Context) {
// }

func (h *ScheduleHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/schedule", h.CreateSchedule)
	r.PUT("/schedule/:id", h.UpdateSchedule)
	r.DELETE("/schedule/:id", h.DeleteSchedule)
	r.GET("/schedule/:id", h.GetSchedule)
	r.GET("/schedule", h.ListSchedule)
	// r.GET("/schedulejob", s.ListScheduleJobs)
	// r.POST("/schedule/reload", s.ReoloadScheduleJobs)
}
