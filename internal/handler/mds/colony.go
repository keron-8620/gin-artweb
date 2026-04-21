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
func (s *MdsColonyHandler) CreateMdsColony(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("创建mds集群:开始执行")

	var req mdsmodel.MdsColonyUpsertDTO
	if !common.ShouldBind(c, log, &req, "创建mds集群:绑定创建mds集群参数失败") {
		return
	}

	m, rErr := s.colonySvc.CreateMdsColony(ctx, &req)
	if rErr != nil {
		log.Error(
			"创建mds集群:执行失败",
			zap.Error(rErr),
			zap.Object("mds_colony_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	log.Info(
		"创建mds集群:执行成功",
		zap.Uint32("mds_colony_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &mdsmodel.MdsColonyResp{
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
func (s *MdsColonyHandler) UpdateMdsColony(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("更新mds集群:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "更新mds集群:绑定更新mds集群ID参数失败") {
		return
	}

	var req mdsmodel.MdsColonyUpsertDTO
	if !common.ShouldBind(c, log, &req, "更新mds集群:绑定更新mds集群参数失败") {
		return
	}

	m, err := s.colonySvc.UpdateMdsColonyByID(ctx, uri.ID, req)
	if err != nil {
		log.Error(
			"更新mds集群:执行失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", uri.ID),
			zap.Object("mds_colony_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"更新mds集群:执行成功",
		zap.Uint32("mds_colony_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &mdsmodel.MdsColonyResp{
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
func (s *MdsColonyHandler) DeleteMdsColony(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("删除mds集群:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "删除mds集群:绑定删除mds集群ID参数失败") {
		return
	}

	err := s.colonySvc.DeleteMdsColonyByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"删除mds集群:执行失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"删除mds集群:执行成功",
		zap.Uint32("mds_colony_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
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
func (s *MdsColonyHandler) GetMdsColony(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询mds集群详情:绑定查询mds集群ID参数失败") {
		return
	}

	m, err := s.colonySvc.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, uri.ID)
	if err != nil {
		log.Error(
			"查询mds集群:执行失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mo := mdsmodel.MdsColonyToDetailOut(*m)
	c.JSON(http.StatusOK, &mdsmodel.MdsColonyResp{
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
func (s *MdsColonyHandler) ListMdsColony(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var req mdsmodel.ListMdsColonyDTO
	if !common.ShouldBindQuery(c, log, &req, "查询mds集群列表:绑定查询mds集群列表参数失败") {
		return
	}

	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := s.colonySvc.ListMdsColony(ctx, page, size, req)
	if err != nil {
		log.Error(
			"查询mds集群列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("mds_colony_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mbs := mdsmodel.ListMdsColonyToDetailOut(ms)
	c.JSON(http.StatusOK, &mdsmodel.PagMdsColonyResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 查询mds集群计划任务列表
// @Description 本接口用于查询指定ID的mds集群计划任务列表
// @Tags mds集群管理
// @Accept json
// @Produce json
// @Param id path uint true "mds集群编号"
// @Success 200 {object} jobmodel.PagScheduleResp "成功返回mds计划任务列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "mds集群未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/mds/colony/{id}/schedule [get]
// @Security ApiKeyAuth
func (s *MdsColonyHandler) ListMdsSchedules(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询mds计划任务:绑定查询mds集群ID参数失败") {
		return
	}

	schedules, err := s.colonySvc.ListMdsSchedules(ctx, uri.ID)
	if err != nil {
		log.Error(
			"查询mds计划任务:执行失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"查询mds计划任务:执行成功",
		zap.Uint32("mds_colony_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	num := len(schedules)
	mos := jobmodel.ListScheduledToDetailOut(schedules)
	c.JSON(http.StatusOK, &jobmodel.PagScheduleResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(1, num, int64(num), mos),
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
func (s *MdsColonyHandler) ListMdsTaskStatus(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(s.log, ctx)

	var req mdsmodel.ListMdsColonyDTO
	if !common.ShouldBindQuery(c, log, &req, "查询mds集群任务状态:绑定查询mds集群任务状态参数失败") {
		return
	}

	tasks, rErr := s.mdsTaskSvc.BuildTaskExecutionInfos(ctx, req)
	if rErr != nil {
		log.Error(
			"构建mds集群任务信息失败",
			zap.Error(rErr),
			zap.Object("mds_colony_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	if len(tasks) == 0 {
		c.JSON(http.StatusOK, &mdsmodel.ListMdsTasksInfoResp{
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

	c.JSON(http.StatusOK, &mdsmodel.ListMdsTasksInfoResp{
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
	r.GET("/colony/:id/schedule", s.ListMdsSchedules)
	r.GET("/colony/status", s.ListMdsTaskStatus)
}
