package biz

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"

	jobsmodel "gin-artweb/internal/model/jobs"
	jobsvc "gin-artweb/internal/service/jobs"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type JobsService struct {
	log        *zap.Logger
	ucScript   *jobsvc.ScriptService
	ucRecord   *jobsvc.RecordService
	ucSchedule *jobsvc.ScheduleService
}

func NewRecordService(
	log *zap.Logger,
	ucScript *jobsvc.ScriptService,
	ucRecord *jobsvc.RecordService,
	ucSchedule *jobsvc.ScheduleService,
) *JobsService {
	return &JobsService{
		log:        log,
		ucScript:   ucScript,
		ucRecord:   ucRecord,
		ucSchedule: ucSchedule,
	}
}

func (uc *JobsService) FindRecordsByIDs(
	ctx context.Context,
	recordIDs []uint32,
) (map[uint32]jobsmodel.ScriptRecordModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(recordIDs) == 0 {
		return map[uint32]jobsmodel.ScriptRecordModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": recordIDs},
	}
	_, ms, rErr := uc.ucRecord.ListcriptRecord(ctx, qp)
	if rErr != nil {
		return nil, rErr
	}
	if ms == nil || len(*ms) == 0 {
		return map[uint32]jobsmodel.ScriptRecordModel{}, nil
	}
	rms := *ms
	result := make(map[uint32]jobsmodel.ScriptRecordModel, len(rms))
	for _, m := range rms {
		result[m.ID] = m
	}
	return result, nil
}

func (uc *JobsService) FindRecordsByMap(
	ctx context.Context,
	cache map[uint32]jobsmodel.ScriptRecordModel,
	recordID uint32,
	colonyNum, taskName string,
) *jobsmodel.ScriptRecordModel {
	task, exists := cache[recordID]
	if !exists {
		uc.log.Debug(
			"未找到oes的任务状态",
			zap.String("colony_num", colonyNum),
			zap.String("task_name", taskName),
			zap.Uint32("script_record_id", recordID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil
	}
	uc.log.Debug(
		"获取oes的任务状态成功",
		zap.String("colony_num", colonyNum),
		zap.String("task_name", taskName),
		zap.Uint32("script_record_id", recordID),
		zap.Object("task", &task),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &task
}

func (uc *JobsService) InitCron(
	ctx context.Context,
	colonyNum string,
	tasks map[string]string,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	claims, cErr := ctxutil.GetUserClaims(ctx)
	if cErr != nil {
		uc.log.Error(
			"获取用户信息失败",
			zap.Error(cErr),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return cErr
	}

	_, ms, rErr := uc.ucScript.ListScript(ctx, database.QueryParams{
		Query: map[string]any{
			"is_builtin = ?": true,
			"project = ?":    "oes",
			"label = ?":      "cmd",
			"name in ?":      tasks,
		},
		Columns: []string{"id", "name"},
	})
	if rErr != nil {
		uc.log.Error(
			"获取mds的任务脚本失败",
			zap.Error(rErr),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return rErr
	}
	if ms == nil || len(*ms) == 0 {
		return nil
	}
	for _, m := range *ms {
		_, err := uc.ucSchedule.CreateSchedule(ctx, jobsmodel.ScheduleModel{
			ScriptID:      m.ID,
			Name:          filepath.Base(m.Name),
			Specification: tasks[m.Name],
			IsEnabled:     true,
			EnvVars:       "{}",
			CommandArgs:   colonyNum,
			WorkDir:       "",
			Timeout:       3600,
			IsRetry:       false,
			RetryInterval: 300,
			MaxRetries:    3,
			Username:      claims.Username,
		})
		if err != nil {
			uc.log.Error(
				"创建mds的任务失败",
				zap.Error(err),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			return err
		}
	}
	return nil
}
