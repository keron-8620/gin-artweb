package job

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	jobmodel "gin-artweb/internal/model/job"
	jobrepo "gin-artweb/internal/repo/job"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type ScheduleService struct {
	log           *zap.Logger
	scriptRepo    *jobrepo.ScriptRepo
	scheduleRepo  *jobrepo.ScheduleRepo
	recordService *RecordService
	crontab       *cron.Cron
	entryMap      map[uint32]cron.EntryID
	mutex         sync.RWMutex
}

func NewScheduleService(
	log *zap.Logger,
	scriptRepo *jobrepo.ScriptRepo,
	scheduleRepo *jobrepo.ScheduleRepo,
	recordService *RecordService,
	crontab *cron.Cron,
) *ScheduleService {
	return &ScheduleService{
		log:           log,
		scriptRepo:    scriptRepo,
		scheduleRepo:  scheduleRepo,
		recordService: recordService,
		crontab:       crontab,
		entryMap:      make(map[uint32]cron.EntryID),
	}
}

func (s *ScheduleService) AddJob(
	ctx context.Context,
	m *jobmodel.ScheduleModel,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"添加计划任务：开始执行",
		zap.Uint32("schedule_id", m.ID),
		zap.String("specification", m.Specification),
	)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	entryID, err := s.crontab.AddJob(m.Specification, cron.FuncJob(func() {
		execReq := jobmodel.ExecuteBIZ{
			CommandArgs: m.CommandArgs,
			EnvVars:     m.EnvVars,
			ScriptID:    m.ScriptID,
			Timeout:     m.Timeout,
			TriggerType: "cron",
			WorkDir:     m.WorkDir,
			Username:    m.Username,
		}

		var retryCount int
		maxRetryCount := 1
		if m.IsRetry && m.MaxRetries > 0 {
			maxRetryCount = m.MaxRetries + 1 // 总尝试次数 = 初始执行 + 重试次数
		}

		for retryCount < maxRetryCount {
			taskinfo, err := s.recordService.SyncExecuteScript(context.Background(), execReq)
			if err == nil && taskinfo.Status == 2 {
				log.Info(
					"计划任务执行成功",
					zap.Uint32("schedule_id", m.ID),
					zap.Int("attempt", retryCount+1),
					zap.Object("taskinfo", taskinfo),
				)
				break
			} else {
				retryCount++
				if retryCount < maxRetryCount {
					waitTime := time.Duration(m.RetryInterval) * time.Second
					log.Error(
						"计划任务执行失败，准备重试",
						zap.Uint32("schedule_id", m.ID),
						zap.Int("attempt", retryCount),
						zap.Int("max_attempts", maxRetryCount),
						zap.Time("next_execution", time.Now().Add(waitTime)),
					)
					time.Sleep(waitTime)
				} else {
					log.Error(
						"计划任务最终执行失败，已达到最大重试次数",
						zap.Uint32("schedule_id", m.ID),
						zap.Int("attempt", retryCount),
						zap.Int("max_attempts", maxRetryCount),
					)
				}
			}
		}
	}))
	if err != nil {
		log.Error(
			"添加计划任务到调度器中失败",
			zap.Error(err),
			zap.Object("schedule_model", m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromReason(errors.ReasonUnknown).WithCause(err)
	}

	s.entryMap[m.ID] = entryID

	log.Debug(
		"添加计划任务：执行成功",
		zap.Uint32("schedule_id", m.ID),
		zap.Int64("entry_id", int64(entryID)),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ScheduleService) RemoveJob(ctx context.Context, scheduleID uint32) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"移除计划任务：开始执行",
		zap.Uint32("schedule_id", scheduleID),
	)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if entryID, exists := s.entryMap[scheduleID]; exists {
		s.crontab.Remove(entryID)
		delete(s.entryMap, scheduleID)

		log.Info(
			"移除计划任务：执行成功",
			zap.Uint32("schedule_id", scheduleID),
			zap.Int64("entry_id", int64(entryID)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
	} else {
		log.Info(
			"移除计划任务：计划任务在调度器中不存在, 无需移除",
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
	}
	return nil
}

func (s *ScheduleService) CreateSchedule(
	ctx context.Context,
	dto jobmodel.ScheduleUpsertDTO,
) (*jobmodel.ScheduleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建计划任务：开始执行")

	log.Debug(
		"创建计划任务：输入参数",
		zap.Object("create_schedule_dto", &dto),
	)

	claims := ctxutil.MustGetJwtClaims(ctx)
	m := jobmodel.ScheduleModel{
		Name:          dto.Name,
		Specification: dto.Specification,
		IsEnabled:     dto.IsEnabled,
		EnvVars:       dto.EnvVars,
		CommandArgs:   dto.CommandArgs,
		WorkDir:       dto.WorkDir,
		Timeout:       dto.Timeout,
		IsRetry:       dto.IsRetry,
		RetryInterval: dto.RetryInterval,
		MaxRetries:    dto.MaxRetries,
		Username:      claims.Subject,
		ScriptID:      dto.ScriptID,
	}

	script, err := s.scriptRepo.GetModel(ctx, "id = ?", dto.ScriptID)
	if err != nil {
		log.Error(
			"创建计划任务：查询脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", m.ScriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": m.ScriptID})
	}
	m.Script = *script

	createStepStart := time.Now()
	log.Debug(
		"创建计划任务：开始创建数据库模型",
		zap.Object("schedule_model", &m),
	)
	if err := s.scheduleRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建计划任务：创建数据库模型失败",
			zap.Error(err),
			zap.Object("schedule_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建计划任务：创建数据库模型成功",
		zap.Object("schedule_model", &m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	if m.IsEnabled {
		addJobStepStart := time.Now()
		log.Debug(
			"创建计划任务：开始添加计划任务到调度器",
			zap.Uint32("schedule_id", m.ID),
		)
		if err := s.AddJob(ctx, &m); err != nil {
			s.RemoveJob(ctx, m.ID)
			log.Error(
				"创建计划任务：添加计划任务到调度器失败",
				zap.Error(err),
				zap.Uint32("schedule_id", m.ID),
				zap.Duration("add_job_step_duration", time.Since(addJobStepStart)),
			)
			return nil, err
		}
		addJobStepDuration := time.Since(addJobStepStart)
		log.Debug(
			"创建计划任务：添加计划任务到调度器成功",
			zap.Uint32("schedule_id", m.ID),
			zap.Duration("add_job_step_duration", addJobStepDuration),
		)
	}

	log.Info(
		"创建计划任务：执行成功",
		zap.Uint32("schedule_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *ScheduleService) UpdateScheduleByID(
	ctx context.Context,
	scheduleID uint32,
	dto jobmodel.ScheduleUpsertDTO,
) (*jobmodel.ScheduleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新计划任务：开始执行",
		zap.Uint32("schedule_id", scheduleID),
	)

	log.Debug(
		"更新计划任务：输入参数",
		zap.Uint32("schedule_id", scheduleID),
		zap.Object("update_schedule_dto", &dto),
	)

	claims := ctxutil.MustGetJwtClaims(ctx)
	updateData := map[string]any{
		"name":           dto.Name,
		"specification":  dto.Specification,
		"is_enabled":     dto.IsEnabled,
		"env_vars":       dto.EnvVars,
		"command_args":   dto.CommandArgs,
		"work_dir":       dto.WorkDir,
		"timeout":        dto.Timeout,
		"is_retry":       dto.IsRetry,
		"retry_interval": dto.RetryInterval,
		"max_retries":    dto.MaxRetries,
		"username":       claims.Subject,
		"script_id":      dto.ScriptID,
	}

	updateStepStart := time.Now()
	log.Debug(
		"更新计划任务：开始更新数据库模型",
		zap.Any("update_data", updateData),
		zap.Uint32("schedule_id", scheduleID),
	)
	if err := s.scheduleRepo.UpdateModel(ctx, updateData, "id = ?", scheduleID); err != nil {
		log.Error(
			"更新计划任务：更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("schedule_id", scheduleID),
			zap.Any("update_data", updateData),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
		)
		return nil, errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新计划任务：更新数据库模型成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	m, rErr := s.FindScheduleByID(ctx, []string{"Script"}, scheduleID)
	if rErr != nil {
		log.Error(
			"更新计划任务：查询更新后的计划任务详情失败",
			zap.Error(rErr),
			zap.Uint32("schedule_id", scheduleID),
		)
		return nil, rErr
	}

	removeJobStepStart := time.Now()
	log.Debug(
		"更新计划任务：开始移除旧计划任务",
		zap.Uint32("schedule_id", scheduleID),
	)
	if err := s.RemoveJob(ctx, scheduleID); err != nil {
		log.Error(
			"更新计划任务：移除旧计划任务失败",
			zap.Error(err),
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("remove_job_step_duration", time.Since(removeJobStepStart)),
		)
		return nil, err
	}
	removeJobStepDuration := time.Since(removeJobStepStart)
	log.Debug(
		"更新计划任务：移除旧计划任务成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.Duration("remove_job_step_duration", removeJobStepDuration),
	)

	// 添加新的计划任务
	if m.IsEnabled {
		addJobStepStart := time.Now()
		log.Debug(
			"更新计划任务：开始添加新计划任务",
			zap.Uint32("schedule_id", scheduleID),
		)
		if err := s.AddJob(ctx, m); err != nil {
			log.Error(
				"更新计划任务：添加新计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", scheduleID),
				zap.Duration("add_job_step_duration", time.Since(addJobStepStart)),
			)
			return nil, err
		}
		addJobStepDuration := time.Since(addJobStepStart)
		log.Debug(
			"更新计划任务：添加新计划任务成功",
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("add_job_step_duration", addJobStepDuration),
		)
	}

	log.Info(
		"更新计划任务：执行成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("remove_job_step_duration", removeJobStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ScheduleService) UpdateScheduleByIDs(
	ctx context.Context,
	scheduleIDs []uint32,
	updateData map[string]any,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"批量更新计划任务：开始执行",
		zap.Uint32s("schedule_ids", scheduleIDs),
	)

	claims := ctxutil.MustGetJwtClaims(ctx)
	updateData["username"] = claims.Username

	updateStepStart := time.Now()
	log.Debug(
		"批量更新计划任务：开始更新数据库模型",
		zap.Any("update_data", updateData),
		zap.Uint32s("schedule_ids", scheduleIDs),
	)
	if err := s.scheduleRepo.UpdateModel(ctx, updateData, "id IN ?", scheduleIDs); err != nil {
		log.Error(
			"批量更新计划任务：更新数据库模型失败",
			zap.Error(err),
			zap.Uint32s("schedule_ids", scheduleIDs),
			zap.Any("update_data", updateData),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
		)
		return errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"批量更新计划任务：更新数据库模型成功",
		zap.Uint32s("schedule_ids", scheduleIDs),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	listStepStart := time.Now()
	log.Debug(
		"批量更新计划任务：开始查询更新后的计划任务详情",
		zap.Uint32s("schedule_ids", scheduleIDs),
	)
	ms, err := s.scheduleRepo.ListModel(ctx, database.QueryParams{
		Query: map[string]any{"id in ?": scheduleIDs},
	})
	if err != nil {
		log.Error(
			"批量更新计划任务：查询更新后的计划任务详情失败",
			zap.Error(err),
			zap.Uint32s("schedule_ids", scheduleIDs),
		)
		return errors.NewGormError(err, nil)
	}
	log.Debug(
		"批量更新计划任务：查询更新后的计划任务详情成功",
		zap.Uint32s("schedule_ids", scheduleIDs),
		zap.Duration("list_step_duration", time.Since(listStepStart)),
	)

	resetJobStepStart := time.Now()
	log.Debug(
		"批量更新计划任务：开始重置计划任务",
		zap.Uint32s("schedule_ids", scheduleIDs),
	)
	for _, m := range ms {
		removeJobStepStart := time.Now()
		if err := s.RemoveJob(ctx, m.ID); err != nil {
			log.Error(
				"批量更新计划任务：移除旧计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", m.ID),
				zap.Duration("remove_job_step_duration", time.Since(removeJobStepStart)),
			)
			return err
		}
		removeJobStepDuration := time.Since(removeJobStepStart)
		log.Debug(
			"批量更新计划任务：移除旧计划任务成功",
			zap.Uint32("schedule_id", m.ID),
			zap.Duration("remove_job_step_duration", removeJobStepDuration),
		)

		// 添加新的计划任务
		if m.IsEnabled {
			addJobStepStart := time.Now()
			log.Debug(
				"批量更新计划任务：开始添加新计划任务",
				zap.Uint32("schedule_id", m.ID),
			)
			if err := s.AddJob(ctx, &m); err != nil {
				log.Error(
					"批量更新计划任务：添加新计划任务失败",
					zap.Error(err),
					zap.Uint32("schedule_id", m.ID),
					zap.Duration("add_job_step_duration", time.Since(addJobStepStart)),
				)
				return err
			}
			addJobStepDuration := time.Since(addJobStepStart)
			log.Debug(
				"批量更新计划任务：添加新计划任务成功",
				zap.Uint32("schedule_id", m.ID),
				zap.Duration("add_job_step_duration", addJobStepDuration),
			)
		}
	}
	resetJobStepDuration := time.Since(resetJobStepStart)
	log.Debug(
		"批量更新计划任务：重置计划任务成功",
		zap.Uint32s("schedule_ids", scheduleIDs),
		zap.Duration("reset_job_step_duration", resetJobStepDuration),
	)

	log.Info(
		"批量更新计划任务：执行成功",
		zap.Uint32s("schedule_ids", scheduleIDs),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("reset_job_step_duration", resetJobStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ScheduleService) DeleteScheduleByID(
	ctx context.Context,
	scheduleID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除计划任务：开始执行",
		zap.Uint32("schedule_id", scheduleID),
	)

	deleteStepStart := time.Now()
	log.Debug(
		"删除计划任务：开始删除数据库模型",
		zap.Uint32("schedule_id", scheduleID),
	)
	if err := s.scheduleRepo.DeleteModel(ctx, scheduleID); err != nil {
		log.Error(
			"删除计划任务：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"id": scheduleID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除计划任务：删除数据库模型成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	removeJobStepStart := time.Now()
	log.Debug(
		"删除计划任务：开始从调度器中移除计划任务",
		zap.Uint32("schedule_id", scheduleID),
	)
	if err := s.RemoveJob(ctx, scheduleID); err != nil {
		log.Error(
			"删除计划任务：从调度器中移除计划任务失败",
			zap.Error(err),
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("remove_job_step_duration", time.Since(removeJobStepStart)),
		)
		return err
	}
	removeJobStepDuration := time.Since(removeJobStepStart)
	log.Debug(
		"删除计划任务：从调度器中移除计划任务成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.Duration("remove_job_step_duration", removeJobStepDuration),
	)

	log.Info(
		"删除计划任务：执行成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("remove_job_step_duration", removeJobStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ScheduleService) DeleteScheduleByIDs(
	ctx context.Context,
	scheduleIDs []uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info(
		"批量删除计划任务：开始执行",
		zap.Uint32s("schedule_ids", scheduleIDs),
	)

	deleteStepStart := time.Now()
	log.Debug(
		"批量删除计划任务：开始删除数据库模型",
		zap.Uint32s("schedule_ids", scheduleIDs),
	)
	if err := s.scheduleRepo.DeleteModel(ctx, scheduleIDs); err != nil {
		log.Error(
			"批量删除计划任务：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32s("schedule_ids", scheduleIDs),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"ids": scheduleIDs})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"批量删除计划任务：删除数据库模型成功",
		zap.Uint32s("schedule_ids", scheduleIDs),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	// 移除计划任务
	removeStepStart := time.Now()
	log.Debug(
		"批量删除计划任务：开始移除计划任务",
		zap.Uint32s("schedule_ids", scheduleIDs),
	)
	for _, id := range scheduleIDs {
		if err := s.RemoveJob(ctx, id); err != nil {
			log.Error(
				"批量删除计划任务：移除计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", id),
			)
			return err
		}
	}
	removeStepDuration := time.Since(removeStepStart)
	log.Debug(
		"批量删除计划任务：移除计划任务成功",
		zap.Uint32s("schedule_ids", scheduleIDs),
		zap.Duration("remove_step_duration", removeStepDuration),
	)

	log.Info(
		"批量删除计划任务：执行成功",
		zap.Uint32s("schedule_ids", scheduleIDs),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("remove_step_duration", removeStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ScheduleService) FindScheduleByID(
	ctx context.Context,
	preloads []string,
	scheduleID uint32,
) (*jobmodel.ScheduleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询计划任务：开始执行",
		zap.Strings("preloads", preloads),
		zap.Uint32("schedule_id", scheduleID),
	)

	m, err := s.scheduleRepo.GetModel(ctx, preloads, scheduleID)
	if err != nil {
		log.Error(
			"查询计划任务：查询数据库模型失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": scheduleID})
	}
	log.Debug(
		"查询计划任务：查询到的数据库模型详情",
		zap.Object("schedule_model", m),
	)

	log.Info(
		"查询计划任务：执行成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ScheduleService) ListSchedule(
	ctx context.Context,
	page, size int,
	dto jobmodel.ListScheduleDTO,
) (int64, []jobmodel.ScheduleModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询计划任务列表：开始执行")

	log.Debug(
		"查询计划任务列表：参数详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_schedule_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Script"},
		Limit:    limit,
		Offset:   offset,
		OrderBy:  []string{"id DESC"},
		Query:    dto.ToQueryMap(),
	}

	log.Debug(
		"查询计划任务列表：查询数据库模型参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询计划任务列表：开始查询数据库模型总数",
		zap.Object("query_params", &qp),
	)
	count, err := s.scheduleRepo.CountModel(ctx, qp.Query)
	countStepDuration := time.Since(countStepStart)
	if err != nil {
		log.Error(
			"查询计划任务列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("count_step_duration", countStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询计划任务列表：查询数据库模型总数成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		log.Warn(
			"查询计划任务列表：数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return count, nil, nil
	}

	ms, err := s.scheduleRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询计划任务列表：查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	log.Info(
		"查询计划任务列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *ScheduleService) ListScheduleJob(
	ctx context.Context,
) ([]jobmodel.ScheduleJobInfo, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("获取调度器任务列表：开始执行")

	// 获取 cron 调度器中的所有条目
	entries := s.crontab.Entries()

	// 准备返回结果
	job := make([]jobmodel.ScheduleJobInfo, 0, len(entries))

	// 创建反向映射以便查找 schedule ID
	s.mutex.RLock()
	scheduleToEntry := make(map[cron.EntryID]uint32)
	for scheduleID, entryID := range s.entryMap {
		scheduleToEntry[entryID] = scheduleID
	}
	s.mutex.RUnlock()

	// 遍历所有条目
	for _, entry := range entries {
		scheduleID, exists := scheduleToEntry[entry.ID]
		if !exists {
			// 如果找不到对应的 schedule ID，使用 0
			scheduleID = 0
		}

		jobInfo := jobmodel.ScheduleJobInfo{
			EntryID:    entry.ID,
			ScheduleID: scheduleID,
			NextRun:    entry.Next,
			PrevRun:    entry.Prev,
		}

		job = append(job, jobInfo)
	}

	log.Info(
		"获取调度器任务列表：执行成功",
		zap.Int("job_count", len(job)),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	return job, nil
}

func (s *ScheduleService) ReLoadSchedule(
	ctx context.Context,
	query map[string]any,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("重新加载计划任务：开始执行")

	log.Debug(
		"重新加载计划任务：参数详情",
		zap.Any("query", query),
	)

	qp := database.QueryParams{
		Query: query,
	}

	listStepStart := time.Now()
	log.Debug(
		"重新加载计划任务：开始查询数据库模型列表",
		zap.Object("query_params", &qp),
	)
	ms, err := s.scheduleRepo.ListModel(ctx, qp)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"重新加载计划任务：查询计划任务失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, query)
	}

	log.Debug(
		"重新加载计划任务：查询到计划任务数量",
		zap.Int("schedule_count", len(ms)),
		zap.Duration("list_step_duration", listStepDuration),
	)

	if len(ms) > 0 {
		processStepStart := time.Now()
		log.Debug(
			"重新加载计划任务：开始处理计划任务",
			zap.Int("schedule_count", len(ms)),
		)
		for _, m := range ms {
			log.Debug(
				"重新加载计划任务：处理计划任务",
				zap.Uint32("schedule_id", m.ID),
				zap.Bool("is_enabled", m.IsEnabled),
			)

			if err := s.RemoveJob(ctx, m.ID); err != nil {
				log.Error(
					"重新加载计划任务：移除计划任务失败",
					zap.Error(err),
					zap.Uint32("schedule_id", m.ID),
					zap.Duration("total_duration", time.Since(startTime)),
				)
				return err
			}

			if m.IsEnabled {
				if err := s.AddJob(ctx, &m); err != nil {
					log.Error(
						"重新加载计划任务：添加计划任务失败",
						zap.Error(err),
						zap.Uint32("schedule_id", m.ID),
						zap.Duration("total_duration", time.Since(startTime)),
					)
					return err
				}
			}
		}
		processStepDuration := time.Since(processStepStart)
		log.Debug(
			"重新加载计划任务：处理计划任务成功",
			zap.Int("schedule_count", len(ms)),
			zap.Duration("process_step_duration", processStepDuration),
		)
	}

	log.Info(
		"重新加载计划任务：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
