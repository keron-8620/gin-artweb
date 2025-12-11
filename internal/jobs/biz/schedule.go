package biz

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

const ScheduleIDKey = "schedule_id"

type ScheduleModel struct {
	database.StandardModel
	Name          string      `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Specification string      `gorm:"column:specification;type:text;comment:条件" json:"specification"`
	IsEnabled     bool        `gorm:"column:is_enabled;type:boolean;comment:是否启用" json:"is_enabled"`
	EnvVars       string      `gorm:"column:env_vars;type:json;comment:环境变量(JSON对象)" json:"env_vars"`
	CommandArgs   string      `gorm:"column:command_args;type:varchar(254);comment:命令行参数" json:"command_args"`
	WorkDir       string      `gorm:"column:work_dir;type:varchar(255);comment:工作目录" json:"work_dir"`
	Timeout       int         `gorm:"column:timeout;type:int;not null;default:300;comment:超时时间(秒)" json:"timeout"`
	IsRetry       bool        `gorm:"column:is_retry;type:boolean;default:false;comment:是否启用重试" json:"is_retry"`
	RetryInterval int         `gorm:"column:retry_interval;type:int;default:60;comment:重试间隔(秒)" json:"retry_interval"`
	MaxRetries    int         `gorm:"column:max_retries;type:int;default:3;comment:最大重试次数" json:"max_retries"`
	Username      string      `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	ScriptID      uint32      `gorm:"column:script_id;not null;index;comment:计划任务ID" json:"script_id"`
	Script        ScriptModel `gorm:"foreignKey:ScriptID;references:ID" json:"script"`
}

func (m *ScheduleModel) TableName() string {
	return "jobs_schedule"
}

func (m *ScheduleModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", m.Name)
	enc.AddString("specification", m.Specification)
	enc.AddBool("is_enabled", m.IsEnabled)
	enc.AddString("env_vars", m.EnvVars)
	enc.AddString("command_args", m.CommandArgs)
	enc.AddString("work_dir", m.WorkDir)
	enc.AddInt("timeout", m.Timeout)
	enc.AddString("username", m.Username)
	enc.AddUint32("script_id", m.ScriptID)
	return nil
}

func (m *ScheduleModel) ToExecuteRequest() ExecuteRequest {
	return ExecuteRequest{
		CommandArgs: m.CommandArgs,
		EnvVars:     m.EnvVars,
		ScriptID:    m.ScriptID,
		Timeout:     m.Timeout,
		TriggerType: "cron",
		WorkDir:     m.WorkDir,
		Username:    m.Username,
	}
}

type ScheduleRepo interface {
	CreateModel(context.Context, *ScheduleModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*ScheduleModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]ScheduleModel, error)
}

type ScheduleUsecase struct {
	log           *zap.Logger
	scriptRepo    ScriptRepo
	scheduleRepo  ScheduleRepo
	recordUsecase *RecordUsecase
	crontab       *cron.Cron
	entryMap      map[uint32]cron.EntryID
	mutex         sync.RWMutex
}

func NewScheduleUsecase(
	log *zap.Logger,
	scriptRepo ScriptRepo,
	scheduleRepo ScheduleRepo,
	recordUsecase *RecordUsecase,
	crontab *cron.Cron,
) *ScheduleUsecase {
	return &ScheduleUsecase{
		log:           log,
		scriptRepo:    scriptRepo,
		scheduleRepo:  scheduleRepo,
		recordUsecase: recordUsecase,
		crontab:       crontab,
		entryMap:      make(map[uint32]cron.EntryID),
	}
}

func (uc *ScheduleUsecase) addJob(ctx context.Context, m *ScheduleModel) *errors.Error {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()

	entryID, err := uc.crontab.AddJob(m.Specification, cron.FuncJob(func() {
		execReq := m.ToExecuteRequest()

		var retryCount int
		maxRetryCount := 1
		if m.IsRetry && m.MaxRetries > 0 {
			maxRetryCount = m.MaxRetries + 1 // 总尝试次数 = 初始执行 + 重试次数
		}

		for retryCount < maxRetryCount {
			taskinfo, err := uc.recordUsecase.SyncExecuteScript(context.Background(), execReq)
			if err == nil && taskinfo.Status == 2 {
				uc.log.Info(
					"计划任务执行成功",
					zap.Uint32(ScheduleIDKey, m.ID),
					zap.Int("attempt", retryCount+1),
					zap.Object("taskinfo", taskinfo),
				)
				break
			} else {
				retryCount++
				if retryCount < maxRetryCount {
					waitTime := time.Duration(m.RetryInterval) * time.Second
					uc.log.Error(
						"计划任务执行失败，准备重试",
						zap.Uint32(ScheduleIDKey, m.ID),
						zap.Int("attempt", retryCount),
						zap.Int("max_attempts", maxRetryCount),
						zap.Time("next_execution", time.Now().Add(waitTime)),
					)
					time.Sleep(waitTime)
				} else {
					uc.log.Error(
						"计划任务最终执行失败，已达到最大重试次数",
						zap.Uint32(ScheduleIDKey, m.ID),
						zap.Int("attempt", retryCount),
						zap.Int("max_attempts", maxRetryCount),
					)
				}
			}
		}
	}))
	if err != nil {
		uc.log.Error(
			"添加计划任务到调度器中失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrAddScheduleFailed.WithCause(err)
	}

	uc.entryMap[m.ID] = entryID
	return nil
}

func (uc *ScheduleUsecase) removeJob(ctx context.Context, scheduleID uint32) {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()

	uc.log.Info(
		"开始从调度器中移除计划任务",
		zap.Uint32(ScheduleIDKey, scheduleID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	if entryID, exists := uc.entryMap[scheduleID]; exists {
		uc.crontab.Remove(entryID)
		delete(uc.entryMap, scheduleID)

		uc.log.Info(
			"计划任务从调度器中移除成功",
			zap.Uint32(ScheduleIDKey, scheduleID),
			zap.Int64("entry_id", int64(entryID)),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
	} else {
		uc.log.Info(
			"计划任务在调度器中不存在, 无需移除",
			zap.Uint32(ScheduleIDKey, scheduleID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
	}
}

func (uc *ScheduleUsecase) CreateSchedule(
	ctx context.Context,
	m ScheduleModel,
) (*ScheduleModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建计划任务",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	script, err := uc.scriptRepo.FindModel(ctx, "id = ?", m.ScriptID)
	if err != nil {
		uc.log.Error(
			"查询脚本失败",
			zap.Error(err),
			zap.Uint32(ScriptIDKey, m.ScriptID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": m.ScriptID})
	}
	m.Script = *script

	if err := uc.scheduleRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建计划任务失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	if m.IsEnabled {
		if err := uc.addJob(ctx, &m); err != nil {
			return nil, err
		}
	}

	uc.log.Info(
		"创建计划任务成功",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *ScheduleUsecase) UpdateScheduleByID(
	ctx context.Context,
	scheduleID uint32,
	data map[string]any,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始更新计划任务",
		zap.Uint32(ScheduleIDKey, scheduleID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.scheduleRepo.UpdateModel(ctx, data, "id = ?", scheduleID); err != nil {
		uc.log.Error(
			"更新计划任务失败",
			zap.Error(err),
			zap.Uint32(ScheduleIDKey, scheduleID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, data)
	}

	m, rErr := uc.FindScheduleByID(ctx, []string{"Script"}, scheduleID)
	if rErr != nil {
		return rErr
	}

	uc.removeJob(ctx, scheduleID)
	if rErr := uc.addJob(ctx, m); rErr != nil {
		return rErr
	}

	uc.log.Info(
		"更新计划任务成功",
		zap.Uint32(ScheduleIDKey, scheduleID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *ScheduleUsecase) DeleteScheduleByID(
	ctx context.Context,
	scheduleID uint32,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除计划任务",
		zap.Uint32(ScheduleIDKey, scheduleID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.scheduleRepo.DeleteModel(ctx, scheduleID); err != nil {
		uc.log.Error(
			"删除计划任务失败",
			zap.Error(err),
			zap.Uint32(ScheduleIDKey, scheduleID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, map[string]any{"id": scheduleID})
	}

	uc.removeJob(ctx, scheduleID)

	uc.log.Info(
		"计划任务删除成功",
		zap.Uint32(ScheduleIDKey, scheduleID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *ScheduleUsecase) FindScheduleByID(
	ctx context.Context,
	preloads []string,
	scheduleID uint32,
) (*ScheduleModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询计划任务",
		zap.Uint32(ScheduleIDKey, scheduleID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := uc.scheduleRepo.FindModel(ctx, preloads, scheduleID)
	if err != nil {
		uc.log.Error(
			"查询计划任务失败",
			zap.Error(err),
			zap.Uint32(ScheduleIDKey, scheduleID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": scheduleID})
	}

	uc.log.Info(
		"查询计划任务成功",
		zap.Uint32(ScheduleIDKey, scheduleID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *ScheduleUsecase) ListSchedule(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]ScheduleModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询计划任务列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	count, ms, err := uc.scheduleRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询计划任务列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询计划任务列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *ScheduleUsecase) ReloadScheduleJob(
	ctx context.Context,
	scheduleID uint32,
) *errors.Error {
	schedule, rErr := uc.FindScheduleByID(ctx, []string{""}, scheduleID)
	if rErr != nil {
		return rErr
	}
	uc.removeJob(ctx, scheduleID)
	return uc.addJob(ctx, schedule)
}

func (uc *ScheduleUsecase) ListScheduleJob(
	ctx context.Context,
) (*[]ScheduleJobInfo, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	// 获取 cron 调度器中的所有条目
	entries := uc.crontab.Entries()

	// 准备返回结果
	jobs := make([]ScheduleJobInfo, 0, len(entries))

	// 创建反向映射以便查找 schedule ID
	scheduleToEntry := make(map[cron.EntryID]uint32)
	for scheduleID, entryID := range uc.entryMap {
		scheduleToEntry[entryID] = scheduleID
	}

	// 遍历所有条目
	for _, entry := range entries {
		scheduleID, exists := scheduleToEntry[entry.ID]
		if !exists {
			// 如果找不到对应的 schedule ID，使用 0
			scheduleID = 0
		}

		jobInfo := ScheduleJobInfo{
			EntryID:    entry.ID,
			ScheduleID: scheduleID,
			NextRun:    entry.Next,
			PrevRun:    entry.Prev,
			Spec:       "", // 这里需要额外存储规格信息才能获取
		}

		jobs = append(jobs, jobInfo)
	}

	uc.log.Info(
		"获取调度器任务列表成功",
		zap.Int("job_count", len(jobs)),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	return &jobs, nil

}

type ScheduleJobInfo struct {
	EntryID    cron.EntryID `json:"entry_id"`
	ScheduleID uint32       `json:"schedule_id"`
	NextRun    time.Time    `json:"next_run"`
	PrevRun    time.Time    `json:"prev_run"`
	Spec       string       `json:"specification"`
}
