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

func (s *ScheduleService) CreateSchedule(
	ctx context.Context,
	dto jobmodel.ScheduleUpsertDTO,
) (*jobmodel.ScheduleModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建计划任务:开始执行",
		zap.Object("create_schedule_dto", &dto),
	)

	claims := ctxutil.MustGetJwtClaims(ctx)
	m := dto.ToModel(claims.Username)

	script, err := s.scriptRepo.GetModel(ctx, "id = ?", dto.ScriptID)
	if err != nil {
		log.Error(
			"创建计划任务:查询脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", m.ScriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": m.ScriptID})
	}

	if err := s.scheduleRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建计划任务:创建数据库模型失败",
			zap.Error(err),
			zap.Object("schedule_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	rollback := func() {
		if err := s.DeleteScheduleByID(ctx, m.ID); err != nil {
			log.Error(
				"创建计划任务:回滚计划任务失败,请手动清理脏数据",
				zap.Error(err),
				zap.Uint32("schedule_id", m.ID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	m.Script = *script
	if m.IsEnabled {
		if err := s.AddJob(ctx, m); err != nil {
			log.Error(
				"创建计划任务:添加计划任务到调度器失败",
				zap.Error(err),
				zap.Uint32("schedule_id", m.ID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			rollback()
			return nil, err
		}
	}

	log.Info(
		"创建计划任务:执行成功",
		zap.Uint32("schedule_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *ScheduleService) UpdateScheduleByID(
	ctx context.Context,
	scheduleID uint32,
	dto jobmodel.ScheduleUpsertDTO,
) (*jobmodel.ScheduleModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新计划任务:开始执行",
		zap.Uint32("schedule_id", scheduleID),
		zap.Object("update_schedule_dto", &dto),
	)

	claims := ctxutil.MustGetJwtClaims(ctx)
	updateData := dto.ToUpdateMap(claims.Username)

	om, rErr := s.FindScheduleByID(ctx, []string{"Script"}, scheduleID)
	if rErr != nil {
		log.Error(
			"更新计划任务:查询更新前的计划任务失败",
			zap.Error(rErr),
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if err := s.scheduleRepo.UpdateModel(ctx, updateData, "id = ?", scheduleID); err != nil {
		log.Error(
			"更新计划任务:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("schedule_id", scheduleID),
			zap.Any("update_data", updateData),
			zap.Duration("update_step_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, updateData)
	}

	rollback := func() {
		upData := om.ToUpdateMap()
		if err := s.scheduleRepo.UpdateModel(ctx, upData, "id = ?", scheduleID); err != nil {
			log.Error(
				"更新计划任务:回滚计划任务失败,请手动清理脏数据",
				zap.Error(err),
				zap.Uint32("schedule_id", scheduleID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	nm, rErr := s.FindScheduleByID(ctx, []string{"Script"}, scheduleID)
	if rErr != nil {
		log.Error(
			"更新计划任务:查询更新后的计划任务失败",
			zap.Error(rErr),
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		return nil, rErr
	}

	if om.IsEnabled {
		if err := s.RemoveJob(ctx, scheduleID); err != nil {
			log.Error(
				"更新计划任务:移除旧计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", scheduleID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			rollback()
			return nil, err
		}
	}

	// 添加新的计划任务
	if nm.IsEnabled {
		if err := s.AddJob(ctx, *nm); err != nil {
			log.Error(
				"更新计划任务:添加新计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", scheduleID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			rollback()
			if om.IsEnabled {
				if addErr := s.AddJob(ctx, *om); addErr != nil {
					log.Error(
						"更新计划任务:回退旧计划任务失败, 请手动处理",
						zap.Error(addErr),
						zap.Uint32("schedule_id", scheduleID),
						zap.Duration("total_duration", time.Since(startTime)),
					)
				}
			}
			return nil, err
		}
	}

	log.Info(
		"更新计划任务:执行成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nm, nil
}

func (s *ScheduleService) DeleteScheduleByID(
	ctx context.Context,
	scheduleID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除计划任务:开始执行",
		zap.Uint32("schedule_id", scheduleID),
	)

	m, err := s.scheduleRepo.GetModel(ctx, []string{"Script"}, scheduleID)
	if err != nil {
		log.Error(
			"删除计划任务:查询计划任务失败",
			zap.Error(err),
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": scheduleID})
	}

	if m.IsEnabled {
		if err := s.RemoveJob(ctx, scheduleID); err != nil {
			log.Error(
				"删除计划任务:从调度器中移除计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", scheduleID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return err
		}
	}

	if err := s.scheduleRepo.DeleteModel(ctx, scheduleID); err != nil {
		log.Error(
			"删除计划任务:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		if m.IsEnabled {
			if err := s.AddJob(ctx, *m); err != nil {
				log.Error(
					"删除计划任务:添加计划任务到调度器失败, 请手动处理",
					zap.Error(err),
					zap.Uint32("schedule_id", scheduleID),
					zap.Duration("total_duration", time.Since(startTime)),
				)
			}
		}
		return errors.NewGormError(err, map[string]any{"id": scheduleID})
	}

	log.Info(
		"删除计划任务:执行成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ScheduleService) FindScheduleByID(
	ctx context.Context,
	preloads []string,
	scheduleID uint32,
) (*jobmodel.ScheduleModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.scheduleRepo.GetModel(ctx, preloads, scheduleID)
	if err != nil {
		log.Error(
			"查询计划任务:查询数据库模型失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": scheduleID})
	}
	log.Debug(
		"查询计划任务:查询到的数据库模型详情",
		zap.Object("schedule_model", m),
	)
	return m, nil
}

func (s *ScheduleService) ListSchedule(
	ctx context.Context,
	page, size int,
	dto jobmodel.ListScheduleDTO,
) (int64, []jobmodel.ScheduleModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询计划任务列表:参数详情",
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

	count, err := s.scheduleRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询计划任务列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询计划任务列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.scheduleRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询计划任务列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *ScheduleService) AddJob(
	ctx context.Context,
	m jobmodel.ScheduleModel,
) *errors.Error {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"添加计划任务:开始执行",
		zap.Object("schedule_model", &m),
	)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	entryID, err := s.crontab.AddJob(m.Specification, cron.FuncJob(func() {
		execReq := jobmodel.ExecuteScriptDTO{
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
				break
			} else {
				retryCount++
				if retryCount < maxRetryCount {
					waitTime := time.Duration(m.RetryInterval) * time.Second
					time.Sleep(waitTime)
				} else {
					log.Debug(
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
			zap.Object("schedule_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromReason(errors.ReasonUnknown).WithCause(err)
	}

	s.entryMap[m.ID] = entryID

	log.Info(
		"添加计划任务:执行成功",
		zap.Uint32("schedule_id", m.ID),
		zap.Int64("entry_id", int64(entryID)),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ScheduleService) RemoveJob(
	ctx context.Context,
	scheduleID uint32,
) *errors.Error {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"移除计划任务:入参详情",
		zap.Uint32("schedule_id", scheduleID),
	)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	entryID, exists := s.entryMap[scheduleID]
	if !exists {
		log.Debug(
			"移除计划任务:计划任务在调度器中不存在, 无需移除",
			zap.Uint32("schedule_id", scheduleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil
	}

	s.crontab.Remove(entryID)
	delete(s.entryMap, scheduleID)
	return nil
}

func (s *ScheduleService) ListJob(
	ctx context.Context,
) ([]jobmodel.ScheduleJobInfo, *errors.Error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

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

	log.Debug(
		"获取调度器任务列表:查询到的计划任务数量",
		zap.Int("job_count", len(job)),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	return job, nil
}

func (s *ScheduleService) LoadSchedule(
	ctx context.Context,
) *errors.Error {
	startTime := time.Now()

	s.log.Debug(
		"加载计划任务:开始执行",
		zap.Duration("total_duration", time.Since(startTime)),
	)

	qp := database.QueryParams{
		Preloads: []string{"Script"},
		Query:    map[string]any{"is_enabled = ?": true},
	}

	ms, err := s.scheduleRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"加载计划任务:查询计划任务失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, nil)
	}

	if len(ms) > 0 {
		for _, m := range ms {
			if err := s.AddJob(ctx, m); err != nil {
				s.log.Error(
					"加载计划任务:添加计划任务失败",
					zap.Error(err),
					zap.Object("schedule_model", &m),
					zap.Duration("total_duration", time.Since(startTime)),
				)
				return err
			}
		}
	}

	s.log.Debug(
		"加载计划任务:执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ScheduleService) UpdateScheduleByIDs(
	ctx context.Context,
	scheduleIDs []uint32,
	updateData map[string]any,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	claims := ctxutil.MustGetJwtClaims(ctx)
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"批量更新计划任务:开始执行",
		zap.Uint32s("schedule_ids", scheduleIDs),
	)

	updateData["username"] = claims.Username
	if err := s.scheduleRepo.UpdateModel(ctx, updateData, "id IN ?", scheduleIDs); err != nil {
		log.Error(
			"批量更新计划任务:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32s("schedule_ids", scheduleIDs),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, updateData)
	}

	ms, err := s.scheduleRepo.ListModel(ctx, database.QueryParams{
		Query: map[string]any{"id in ?": scheduleIDs},
	})
	if err != nil {
		log.Error(
			"批量更新计划任务:查询更新后的计划任务详情失败",
			zap.Error(err),
			zap.Uint32s("schedule_ids", scheduleIDs),
		)
		return errors.NewGormError(err, nil)
	}

	for _, m := range ms {
		removeJobStepStart := time.Now()
		if err := s.RemoveJob(ctx, m.ID); err != nil {
			log.Error(
				"批量更新计划任务:移除旧计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", m.ID),
				zap.Duration("remove_job_step_duration", time.Since(removeJobStepStart)),
			)
			return err
		}

		// 添加新的计划任务
		if m.IsEnabled {
			if err := s.AddJob(ctx, m); err != nil {
				log.Error(
					"批量更新计划任务:添加新计划任务失败",
					zap.Error(err),
					zap.Uint32("schedule_id", m.ID),
					zap.Duration("total_duration", time.Since(startTime)),
				)
				return err
			}
		}
	}

	log.Info(
		"批量更新计划任务:执行成功",
		zap.Uint32s("schedule_ids", scheduleIDs),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ScheduleService) DeleteScheduleByIDs(
	ctx context.Context,
	scheduleIDs []uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"批量删除计划任务:开始执行",
		zap.Uint32s("schedule_ids", scheduleIDs),
	)

	if err := s.scheduleRepo.DeleteModel(ctx, scheduleIDs); err != nil {
		log.Error(
			"批量删除计划任务:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32s("schedule_ids", scheduleIDs),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"ids": scheduleIDs})
	}

	for _, id := range scheduleIDs {
		if err := s.RemoveJob(ctx, id); err != nil {
			log.Error(
				"批量删除计划任务:移除计划任务失败",
				zap.Error(err),
				zap.Uint32("schedule_id", id),
			)
			return err
		}
	}

	log.Info(
		"批量删除计划任务:执行成功",
		zap.Uint32s("schedule_ids", scheduleIDs),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
