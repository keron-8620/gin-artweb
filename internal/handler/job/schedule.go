package job

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	jobmodel "gin-artweb/internal/model/job"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

type ScheduleHandler struct {
	log         *zap.Logger
	scheduleSvc *jobsvc.ScheduleService
}

func NewScheduleHandler(
	logger *zap.Logger,
	svcSchedule *jobsvc.ScheduleService,
) *ScheduleHandler {
	return &ScheduleHandler{
		log:         logger,
		scheduleSvc: svcSchedule,
	}
}

// @Summary 创建计划任务
// @Description 本接口用于创建新的计划任务
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param request body jobmodel.ScheduleUpsertDTO true "创建计划任务请求"
// @Success 200 {object} jobmodel.ScheduleResp "成功返回计划任务信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule [post]
// @Security ApiKeyAuth
func (h *ScheduleHandler) CreateSchedule(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req jobmodel.ScheduleUpsertDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"创建计划任务：绑定请求参数失败") {
		return
	}

	log.Info("创建计划任务：开始执行")

	log.Debug(
		"创建计划任务：入参详情",
		zap.Object("schedule_upsert_dto", &req),
	)

	createStepStart := time.Now()
	m, rErr := h.scheduleSvc.CreateSchedule(ctx, req)
	createStepDuration := time.Since(createStepStart)
	if rErr != nil {
		log.Error(
			"创建计划任务：执行失败",
			zap.Error(rErr),
			zap.Object("schedule_upsert_dto", &req),
			zap.Duration("create_step_duration", createStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	log.Debug(
		"创建计划任务：创建的计划任务模型详情",
		zap.Object("schedule_model", m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	log.Info(
		"创建计划任务：执行成功",
		zap.Uint32("schedule_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &jobmodel.ScheduleResp{
		Code: http.StatusOK,
		Data: *jobmodel.ScheduleToDetailOut(*m),
	})
}

// @Summary 更新计划任务
// @Description 本接口用于更新指定ID的计划任务
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param id path uint true "计划任务编号"
// @Param request body jobmodel.ScheduleUpsertDTO true "更新计划任务请求"
// @Success 200 {object} jobmodel.ScheduleResp "成功返回计划任务信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "计划任务未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule/{id} [put]
// @Security ApiKeyAuth
func (h *ScheduleHandler) UpdateSchedule(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"更新计划任务：绑定更新计划任务ID参数失败") {
		return
	}

	var req jobmodel.ScheduleUpsertDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"更新计划任务：绑定更新计划任务请求参数失败") {
		return
	}

	log.Info(
		"更新计划任务：开始执行",
		zap.Uint32("schedule_id", uri.ID),
	)

	log.Debug(
		"更新计划任务：入参详情",
		zap.Uint32("schedule_id", uri.ID),
		zap.Object("schedule_upsert_dto", &req),
	)

	updateStepStart := time.Now()
	m, err := h.scheduleSvc.UpdateScheduleByID(ctx, uri.ID, req)
	updateStepDuration := time.Since(updateStepStart)
	if err != nil {
		log.Error(
			"更新计划任务：执行失败",
			zap.Error(err),
			zap.Uint32("schedule_id", uri.ID),
			zap.Object("schedule_upsert_dto", &req),
			zap.Duration("update_step_duration", updateStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Debug(
		"更新计划任务：更新后的计划任务模型详情",
		zap.Object("schedule_model", m),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	log.Info(
		"更新计划任务：执行成功",
		zap.Uint32("schedule_id", uri.ID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &jobmodel.ScheduleResp{
		Code: http.StatusOK,
		Data: *jobmodel.ScheduleToDetailOut(*m),
	})
}

// @Summary 删除计划任务
// @Description 本接口用于删除指定ID的计划任务
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param id path uint true "计划任务编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "计划任务未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule/{id} [delete]
// @Security ApiKeyAuth
func (h *ScheduleHandler) DeleteSchedule(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"删除计划任务：绑定删除计划任务ID参数失败") {
		return
	}

	log.Info(
		"删除计划任务：开始执行",
		zap.Uint32("schedule_id", uri.ID),
	)

	deleteStepStart := time.Now()
	err := h.scheduleSvc.DeleteScheduleByID(ctx, uri.ID)
	deleteStepDuration := time.Since(deleteStepStart)
	if err != nil {
		log.Error(
			"删除计划任务：执行失败",
			zap.Error(err),
			zap.Uint32("schedule_id", uri.ID),
			zap.Duration("delete_step_duration", deleteStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"删除计划任务：删除计划任务成功",
		zap.Uint32("schedule_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询计划任务详情
// @Description 本接口用于查询指定ID的计划任务详情
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param id path uint true "计划任务编号"
// @Success 200 {object} jobmodel.ScheduleResp "成功返回计划任务信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "计划任务未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule/{id} [get]
// @Security ApiKeyAuth
func (h *ScheduleHandler) GetSchedule(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"查询计划任务：绑定计划任务ID参数失败") {
		return
	}

	log.Info(
		"查询计划任务：开始执行",
		zap.Uint32("schedule_id", uri.ID),
	)

	findStepStart := time.Now()
	m, err := h.scheduleSvc.FindScheduleByID(ctx, []string{"Script"}, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询计划任务：执行失败",
			zap.Error(err),
			zap.Uint32("schedule_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"查询计划任务：查询到的计划任务模型详情",
		zap.Object("schedule_model", m),
		zap.Duration("find_step_duration", findStepDuration),
	)

	log.Info(
		"查询计划任务：执行成功",
		zap.Uint32("schedule_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := jobmodel.ScheduleToDetailOut(*m)
	ctx.JSON(http.StatusOK, &jobmodel.ScheduleResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询计划任务列表
// @Description 本接口用于查询计划任务列表
// @Tags 计划任务管理
// @Accept json
// @Produce json
// @Param request query jobmodel.ListScheduleDTO false "查询参数"
// @Success 200 {object} jobmodel.PagScheduleResp "成功返回计划任务列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/jobs/schedule [get]
// @Security ApiKeyAuth
func (h *ScheduleHandler) ListSchedule(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req jobmodel.ListScheduleDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"查询计划任务列表：绑定查询参数失败") {
		return
	}

	log.Info("查询计划任务列表：开始执行")

	log.Debug(
		"查询计划任务列表：参数详情",
		zap.Object("list_schedule_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.scheduleSvc.ListSchedule(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询计划任务列表：查询计划任务列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_schedule_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询计划任务列表：查询计划任务列表成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := jobmodel.ListScheduledToDetailOut(ms)
	ctx.JSON(http.StatusOK, &jobmodel.PagScheduleResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

func (h *ScheduleHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/schedule", h.CreateSchedule)
	r.PUT("/schedule/:id", h.UpdateSchedule)
	r.DELETE("/schedule/:id", h.DeleteSchedule)
	r.GET("/schedule/:id", h.GetSchedule)
	r.GET("/schedule", h.ListSchedule)
}
