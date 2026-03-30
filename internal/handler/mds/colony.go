package mds

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	jobmodel "gin-artweb/internal/model/job"
	mdsmodel "gin-artweb/internal/model/mds"
	jobsvc "gin-artweb/internal/service/job"
	mdssvc "gin-artweb/internal/service/mds"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

type MdsColonyHandler struct {
	log        *zap.Logger
	colonySvc  *mdssvc.MdsColonyService
	mdsTaskSvc *mdssvc.MdsTaskService
}

func NewMdsColonyHandler(
	logger *zap.Logger,
	colony *mdssvc.MdsColonyService,
	mdsTaskSvc *mdssvc.MdsTaskService,
) *MdsColonyHandler {
	return &MdsColonyHandler{
		log:        logger,
		colonySvc:  colony,
		mdsTaskSvc: mdsTaskSvc,
	}
}

// @Summary 创建mds集群
// @Description 本接口用于创建新的mds集群
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param request body mdsmodel.MdsColonyUpsertDTO true "创建mds集群请求"
// @Success 200 {object} mdsmodel.MdsColonyResp "成功返回mds集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony [post]
// @Security ApiKeyAuth
func (s *MdsColonyHandler) CreateMdsColony(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var req mdsmodel.MdsColonyUpsertDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"创建mds集群：绑定创建mds集群参数失败") {
		return
	}

	createStepStart := time.Now()
	m, rErr := s.colonySvc.CreateMdsColony(ctx, &req)
	createStepDuration := time.Since(createStepStart)
	if rErr != nil {
		log.Error(
			"创建mds集群：执行失败",
			zap.Error(rErr),
			zap.Object("mds_colony_dto", &req),
			zap.Duration("create_step_duration", createStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	log.Debug(
		"创建mds集群：创建后的mds集群模型详情",
		zap.Object("mds_colony_model", m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	log.Info(
		"创建mds集群：执行成功",
		zap.Uint32("mds_colony_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &mdsmodel.MdsColonyResp{
		Code: http.StatusOK,
		Data: *mdsmodel.MdsColonyToDetailOut(*m),
	})
}

// @Summary 更新mds集群
// @Description 本接口用于更新指定ID的mds集群
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param id path uint true "mds集群编号"
// @Param request body mdsmodel.MdsColonyUpsertDTO true "更新mds集群请求"
// @Success 200 {object} mdsmodel.MdsColonyResp "成功返回mds集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony/{id} [put]
// @Security ApiKeyAuth
func (s *MdsColonyHandler) UpdateMdsColony(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"更新mds集群：绑定更新mds集群ID参数失败") {
		return
	}

	var req mdsmodel.MdsColonyUpsertDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"更新mds集群：绑定更新mds集群参数失败") {
		return
	}

	updateStepStart := time.Now()
	m, err := s.colonySvc.UpdateMdsColonyByID(ctx, uri.ID, req)
	updateStepDuration := time.Since(updateStepStart)
	if err != nil {
		log.Error(
			"更新mds集群：执行失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", uri.ID),
			zap.Object("mds_colony_dto", &req),
			zap.Duration("update_step_duration", updateStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"更新mds集群：更新后的mds集群模型详情",
		zap.Object("mds_colony_model", m),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	log.Info(
		"更新mds集群：执行成功",
		zap.Uint32("mds_colony_id", uri.ID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &mdsmodel.MdsColonyResp{
		Code: http.StatusOK,
		Data: *mdsmodel.MdsColonyToDetailOut(*m),
	})
}

// @Summary 删除mds集群
// @Description 本接口用于删除指定ID的mds集群
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param id path uint true "mds集群编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony/{id} [delete]
// @Security ApiKeyAuth
func (s *MdsColonyHandler) DeleteMdsColony(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"删除mds集群：绑定删除mds集群ID参数失败") {
		return
	}

	log.Info(
		"删除mds集群：开始执行",
		zap.Uint32("mds_colony_id", uri.ID),
	)

	deleteStepStart := time.Now()
	err := s.colonySvc.DeleteMdsColonyByID(ctx, uri.ID)
	deleteStepDuration := time.Since(deleteStepStart)
	if err != nil {
		log.Error(
			"删除mds集群：执行失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", uri.ID),
			zap.Duration("delete_step_duration", deleteStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"删除mds集群：执行成功",
		zap.Uint32("mds_colony_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询mds集群详情
// @Description 本接口用于查询指定ID的mds集群详情
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param id path uint true "mds集群编号"
// @Success 200 {object} mdsmodel.MdsColonyResp "成功返回mds集群信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony/{id} [get]
// @Security ApiKeyAuth
func (s *MdsColonyHandler) GetMdsColony(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"查询mds集群详情：绑定查询mds集群ID参数失败") {
		return
	}

	log.Info(
		"查询mds集群：开始执行",
		zap.Uint32("mds_colony_id", uri.ID),
	)

	findStepStart := time.Now()
	m, err := s.colonySvc.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询mds集群：执行失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"查询mds集群：查询到的mds集群模型详情",
		zap.Object("mds_colony_model", m),
		zap.Duration("find_step_duration", findStepDuration),
	)

	log.Info(
		"查询mds集群：执行成功",
		zap.Uint32("mds_colony_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := mdsmodel.MdsColonyToDetailOut(*m)
	ctx.JSON(http.StatusOK, &mdsmodel.MdsColonyResp{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询mds集群列表
// @Description 本接口用于查询mds集群列表
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param name query string false "mds集群名称"
// @Param is_enabled query bool false "是否启用"
// @Param username query string false "创建用户名"
// @Success 200 {object} mdsmodel.PagMdsColonyResp "成功返回mds集群列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony [get]
// @Security ApiKeyAuth
func (s *MdsColonyHandler) ListMdsColony(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var req mdsmodel.ListMdsColonyDTO
	if !common.ShouldBindQuery(
		ctx, log, &req,
		"查询mds集群列表：绑定查询mds集群列表参数失败") {
		return
	}

	log.Info("开始查询mds集群列表")

	log.Debug(
		"查询mds集群列表：查询参数",
		zap.Object("mds_colony_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := s.colonySvc.ListMdsColony(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询mds集群列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("mds_colony_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询mds集群列表成功",
		zap.Int64("total", total),
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := mdsmodel.ListMdsColonyToDetailOut(ms)
	ctx.JSON(http.StatusOK, &mdsmodel.PagMdsColonyResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 查询mds集群列表的任务状态
// @Description 本接口用于查询mds集群列表的任务状态
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param request query mdsmodel.ListMdsColonyDTO false "查询参数"
// @Success 200 {object} mdsmodel.ListMdsTasksInfoResp "成功返回mds集群列表的任务状态"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony/status [get]
// @Security ApiKeyAuth
func (s *MdsColonyHandler) ListMdsTaskStatus(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	var req mdsmodel.ListMdsColonyDTO
	if !common.ShouldBindQuery(
		ctx, log, &req,
		"查询mds集群任务状态：绑定查询mds集群任务状态参数失败") {
		return
	}

	log.Info("查询mds集群任务状态：开始执行")

	log.Debug(
		"查询mds集群任务状态：入参详情",
		zap.Object("mds_colony_dto", &req),
	)

	buildStepStart := time.Now()
	tasks, rErr := s.mdsTaskSvc.BuildTaskExecutionInfos(ctx, req)
	buildStepDuration := time.Since(buildStepStart)
	if rErr != nil {
		log.Error(
			"构建mds集群任务信息失败",
			zap.Error(rErr),
			zap.Object("mds_colony_dto", &req),
			zap.Duration("build_step_duration", buildStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	if len(tasks) == 0 {
		ctx.JSON(http.StatusOK, &mdsmodel.ListMdsTasksInfoResp{
			Code: http.StatusOK,
			Data: []mdsmodel.MdsColonyTaskInfo{},
		})
		return
	}

	results := make([]mdsmodel.MdsColonyTaskInfo, len(tasks))
	for i, info := range tasks {
		// results[i] = s.svcTask.BuildTaskMdsTaskInfo(info)
		mon := jobsvc.BuildTaskInfoFromScriptRecord("mon", info.Mon)
		bse := jobsvc.BuildTaskInfoFromScriptRecord("bse", info.Bse)
		sse := jobsvc.BuildTaskInfoFromScriptRecord("sse", info.Sse)
		szse := jobsvc.BuildTaskInfoFromScriptRecord("szse", info.Szse)
		results[i] = mdsmodel.MdsColonyTaskInfo{
			ColonyNum: info.ColonyNum,
			Tasks:     []jobmodel.BizTaskInfo{mon, bse, sse, szse},
		}
	}

	log.Info(
		"查询mds集群任务状态：执行成功",
		zap.Duration("build_step_duration", buildStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &mdsmodel.ListMdsTasksInfoResp{
		Code: http.StatusOK,
		Data: results,
	})
}

func (s *MdsColonyHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/colony", s.CreateMdsColony)
	r.PUT("/colony/:id", s.UpdateMdsColony)
	r.DELETE("/colony/:id", s.DeleteMdsColony)
	r.GET("/colony/:id", s.GetMdsColony)
	r.GET("/colony", s.ListMdsColony)
	r.GET("/colony/status", s.ListMdsTaskStatus)
}
