package jobs

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	jobsmodel "gin-artweb/internal/model/jobs"
	jobsrepo "gin-artweb/internal/repository/jobs"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type ScheduleService struct {
	log           *zap.Logger
	scriptRepo    *jobsrepo.ScriptRepo
	scheduleRepo  *jobsrepo.ScheduleRepo
	recordService *RecordService
	crontab       *cron.Cron
	entryMap      map[uint32]cron.EntryID
	mutex         sync.RWMutex
}

func NewScheduleService(
	log *zap.Logger,
	scriptRepo *jobsrepo.ScriptRepo,
	scheduleRepo *jobsrepo.ScheduleRepo,
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

func (s *ScheduleService) addJob(ctx context.Context, m *jobsmodel.ScheduleModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	entryID, err := s.crontab.AddJob(m.Specification, cron.FuncJob(func() {
		execReq := jobsmodel.ExecuteRequest{
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
				s.log.Info(
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
					s.log.Error(
						"计划任务执行失败，准备重试",
						zap.Uint32("schedule_id", m.ID),
						zap.Int("attempt", retryCount),
						zap.Int("max_attempts", maxRetryCount),
						zap.Time("next_execution", time.Now().Add(waitTime)),
					)
					time.Sleep(waitTime)
				} else {
					s.log.Error(
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
		s.log.Error(
			"添加计划任务到调度器中失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return errors.FromReason(errors.ReasonUnknown).WithCause(err)
	}

	s.entryMap[m.ID] = entryID
	return nil
}

func (s *ScheduleService) removeJob(ctx context.Context, scheduleID uint32) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.log.Info(
		"开始从调度器中移除计划任务",
		zap.Uint32("schedule_id", scheduleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	if entryID, exists := s.entryMap[scheduleID]; exists {
		s.crontab.Remove(entryID)
		delete(s.entryMap, scheduleID)

		s.log.Info(
			"计划任务从调度器中移除成功",
			zap.Uint32("schedule_id", scheduleID),
			zap.Int64("entry_id", int64(entryID)),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
	} else {
		s.log.Info(
			"计划任务在调度器中不存在, 无需移除",
			zap.Uint32("schedule_id", scheduleID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
	}
	return nil
}

func (s *ScheduleService) CreateSchedule(
	ctx context.Context,
	m jobsmodel.ScheduleModel,
) (*jobsmodel.ScheduleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建计划任务",
		zap.Object(database.ModelKey, &m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	script, err := s.scriptRepo.GetModel(ctx, "id = ?", m.ScriptID)
	if err != nil {
		s.log.Error(
			"查询脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", m.ScriptID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": m.ScriptID})
	}
	m.Script = *script

	if err := s.scheduleRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建计划任务失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	if m.IsEnabled {
		if err := s.addJob(ctx, &m); err != nil {
			s.removeJob(ctx, m.ID)
			return nil, err
		}
	}

	s.log.Info(
		"创建计划任务成功",
		zap.Object(database.ModelKey, &m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (s *ScheduleService) UpdateScheduleByID(
	ctx context.Context,
	scheduleID uint32,
	data map[string]any,
) (*jobsmodel.ScheduleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新计划任务",
		zap.Uint32("schedule_id", scheduleID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := s.scheduleRepo.UpdateModel(ctx, data, "id = ?", scheduleID); err != nil {
		s.log.Error(
			"更新计划任务失败",
			zap.Error(err),
			zap.Uint32("schedule_id", scheduleID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	m, rErr := s.FindScheduleByID(ctx, []string{"Script"}, scheduleID)
	if rErr != nil {
		return nil, rErr
	}

	// 获取原始计划任务id
	entryID, exists := s.entryMap[scheduleID]

	if err := s.removeJob(ctx, scheduleID); err != nil {
		return nil, err
	}

	// 添加新的计划任务
	if m.IsEnabled {
		if err := s.addJob(ctx, m); err != nil {
			return nil, err
		}
	}

	// 删除原有计划任务
	if exists {
		s.crontab.Remove(entryID)
	}

	s.log.Info(
		"更新计划任务成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *ScheduleService) DeleteScheduleByID(
	ctx context.Context,
	scheduleID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除计划任务",
		zap.Uint32("schedule_id", scheduleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := s.scheduleRepo.DeleteModel(ctx, scheduleID); err != nil {
		s.log.Error(
			"删除计划任务失败",
			zap.Error(err),
			zap.Uint32("schedule_id", scheduleID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": scheduleID})
	}

	if err := s.removeJob(ctx, scheduleID); err != nil {
		return err
	}

	s.log.Info(
		"计划任务删除成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *ScheduleService) FindScheduleByID(
	ctx context.Context,
	preloads []string,
	scheduleID uint32,
) (*jobsmodel.ScheduleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询计划任务",
		zap.Uint32("schedule_id", scheduleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := s.scheduleRepo.GetModel(ctx, preloads, scheduleID)
	if err != nil {
		s.log.Error(
			"查询计划任务失败",
			zap.Error(err),
			zap.Uint32("schedule_id", scheduleID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": scheduleID})
	}

	s.log.Info(
		"查询计划任务成功",
		zap.Uint32("schedule_id", scheduleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *ScheduleService) ListSchedule(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]jobsmodel.ScheduleModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询计划任务列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.scheduleRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询计划任务列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询计划任务列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *ScheduleService) ReloadScheduleJobs(ctx context.Context, query map[string]any) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	qp := database.QueryParams{
		Query:   query,
		IsCount: false,
	}
	_, ms, rErr := s.ListSchedule(ctx, qp)
	if rErr != nil {
		return rErr
	}
	if ms != nil {
		for _, m := range *ms {
			if err := s.removeJob(ctx, m.ID); err != nil {
				return err
			}
			if m.IsEnabled {
				if err := s.addJob(ctx, &m); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *ScheduleService) ListScheduleJob(
	ctx context.Context,
) (*[]jobsmodel.ScheduleJobInfo, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	// 获取 cron 调度器中的所有条目
	entries := s.crontab.Entries()

	// 准备返回结果
	jobs := make([]jobsmodel.ScheduleJobInfo, 0, len(entries))

	// 创建反向映射以便查找 schedule ID
	scheduleToEntry := make(map[cron.EntryID]uint32)
	for scheduleID, entryID := range s.entryMap {
		scheduleToEntry[entryID] = scheduleID
	}

	// 遍历所有条目
	for _, entry := range entries {
		scheduleID, exists := scheduleToEntry[entry.ID]
		if !exists {
			// 如果找不到对应的 schedule ID，使用 0
			scheduleID = 0
		}

		jobInfo := jobsmodel.ScheduleJobInfo{
			EntryID:    entry.ID,
			ScheduleID: scheduleID,
			NextRun:    entry.Next,
			PrevRun:    entry.Prev,
		}

		jobs = append(jobs, jobInfo)
	}

	s.log.Info(
		"获取调度器任务列表成功",
		zap.Int("job_count", len(jobs)),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	return &jobs, nil
}
