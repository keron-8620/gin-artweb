package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/jobs/biz"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"

	pbComm "gin-artweb/api/common"
	pbSchedule "gin-artweb/api/jobs/schedule"
	pbScript "gin-artweb/api/jobs/script"
)

type ScheduleService struct {
	log        *zap.Logger
	ucSchedule *biz.ScheduleUsecase
}

func NewScheduleService(
	logger *zap.Logger,
	ucSchedule *biz.ScheduleUsecase,
) *ScheduleService {
	return &ScheduleService{
		log:        logger,
		ucSchedule: ucSchedule,
	}
}

// @Summary 创建计划任务
// @Description 本接口用于创建新的计划任务
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param request body pbSchedule.CreateScheduleRequest true "创建计划任务请求"
// @Success 200 {object} pbSchedule.ScheduleReply "成功返回计划任务信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule [post]
// @Security ApiKeyAuth
func (s *ScheduleService) CreateSchedule(ctx *gin.Context) {
	var req pbSchedule.CreateScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.log.Error(
			"绑定创建计划任务参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	uc := auth.GetUserClaims(ctx)
	schedule := biz.ScheduleModel{
		Name:          req.Name,
		Specification: req.Specification,
		IsEnabled:     req.IsEnabled,
		EnvVars:       req.EnvVars,
		CommandArgs:   req.CommandArgs,
		WorkDir:       req.WorkDir,
		Timeout:       req.Timeout,
		Username:      uc.Subject,
		ScriptID:      req.ScriptID,
	}

	m, rErr := s.ucSchedule.CreateSchedule(ctx, schedule)
	if rErr != nil {
		s.log.Error(
			"创建计划任务失败",
			zap.Error(rErr),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbSchedule.ScheduleReply{
		Code: http.StatusOK,
		Data: *ScheduleToDetailOut(*m),
	})
}

// @Summary 更新计划任务
// @Description 本接口用于更新指定ID的计划任务
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param id path uint true "计划任务编号"
// @Param request body pbSchedule.UpdateScheduleRequest true "更新计划任务请求"
// @Success 200 {object} pbSchedule.ScheduleReply "成功返回计划任务信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "计划任务未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule/{id} [put]
// @Security ApiKeyAuth
func (s *ScheduleService) UpdateSchedule(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定更新计划任务ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbSchedule.UpdateScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.log.Error(
			"绑定更新计划任务参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	uc := auth.GetUserClaims(ctx)
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
		"username":       uc.Subject,
		"script_id":      req.ScriptID,
	}

	m, err := s.ucSchedule.UpdateScheduleByID(ctx, uri.ID, data)
	if err != nil {
		s.log.Error(
			"更新计划任务失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	ctx.JSON(http.StatusOK, &pbSchedule.ScheduleReply{
		Code: http.StatusOK,
		Data: *ScheduleToDetailOut(*m),
	})
}

// @Summary 删除计划任务
// @Description 本接口用于删除指定ID的计划任务
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param id path uint true "计划任务编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "计划任务未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule/{id} [delete]
// @Security ApiKeyAuth
func (s *ScheduleService) DeleteSchedule(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除计划任务ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除计划任务",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucSchedule.DeleteScheduleByID(ctx, uri.ID); err != nil {
		s.log.Error(
			"删除计划任务失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除计划任务成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询计划任务详情
// @Description 本接口用于查询指定ID的计划任务详情
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param id path uint true "计划任务编号"
// @Success 200 {object} pbSchedule.ScheduleReply "成功返回计划任务信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "计划任务未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule/{id} [get]
// @Security ApiKeyAuth
func (s *ScheduleService) GetSchedule(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询计划任务ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询计划任务详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucSchedule.FindScheduleByID(ctx, []string{"Script"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询计划任务详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询计划任务详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := ScheduleToDetailOut(*m)
	ctx.JSON(http.StatusOK, &pbSchedule.ScheduleReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询计划任务列表
// @Description 本接口用于查询计划任务列表
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param request query pbSchedule.ListScheduleRequest false "查询参数"
// @Success 200 {object} pbSchedule.PagScheduleReply "成功返回计划任务列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule [get]
// @Security ApiKeyAuth
func (s *ScheduleService) ListSchedule(ctx *gin.Context) {
	var req pbSchedule.ListScheduleRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询计划任务列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询计划任务列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		Preloads: []string{"Script"},
		IsCount:  true,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"id DESC"},
		Query:    query,
	}
	total, ms, err := s.ucSchedule.ListSchedule(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询计划任务列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询计划任务列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mbs := ListScheduledToDetailOut(ms)
	ctx.JSON(http.StatusOK, &pbSchedule.PagScheduleReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *ScheduleService) ListScheduleJobs(ctx *gin.Context) {

}

func (s *ScheduleService) ReoloadScheduleJobs(ctx *gin.Context) {

}


func (s *ScheduleService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/schedule", s.CreateSchedule)
	r.PUT("/schedule/:id", s.UpdateSchedule)
	r.DELETE("/schedule/:id", s.DeleteSchedule)
	r.GET("/schedule/:id", s.GetSchedule)
	r.GET("/schedule", s.ListSchedule)
	r.GET("/schedulejob", s.ListScheduleJobs)
	r.POST("/schedule/reload", s.ReoloadScheduleJobs)
}

func ScheduleToStandardOut(
	m biz.ScheduleModel,
) *pbSchedule.ScheduleStandardOut {
	return &pbSchedule.ScheduleStandardOut{
		ID:            m.ID,
		CreatedAt:     m.CreatedAt.String(),
		UpdatedAt:     m.UpdatedAt.String(),
		Name:          m.Name,
		Specification: m.Specification,
		IsEnabled:     m.IsEnabled,
		EnvVars:       m.EnvVars,
		CommandArgs:   m.CommandArgs,
		WorkDir:       m.WorkDir,
		Timeout:       m.Timeout,
		Username:      m.Username,
	}
}

func ScheduleToDetailOut(
	m biz.ScheduleModel,
) *pbSchedule.ScheduleDetailOut {
	var script *pbScript.ScriptStandardOut
	if m.Script.ID != 0 {
		script = ScriptModelToStandardOut(m.Script)
	}
	return &pbSchedule.ScheduleDetailOut{
		ScheduleStandardOut: *ScheduleToStandardOut(m),
		Script:              script,
	}
}

func ListScheduledToDetailOut(
	rms *[]biz.ScheduleModel,
) *[]pbSchedule.ScheduleDetailOut {
	if rms == nil {
		return &[]pbSchedule.ScheduleDetailOut{}
	}

	ms := *rms
	mso := make([]pbSchedule.ScheduleDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ScheduleToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
