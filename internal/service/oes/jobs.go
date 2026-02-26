package biz

import (
	"context"

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

// func (uc *JobsService) FindScriptIDs(
// 	ctx context.Context,
// 	tasks []string,
// ) (map[string]uint32, *errors.Error) {
// 	if ctx.Err() != nil {
// 		return nil, errors.FromError(ctx.Err())
// 	}

// 	result := make(map[string]uint32, len(tasks))
// 	_, ms, rErr := uc.ucScript.ListScript(ctx, database.QueryParams{
// 		Query: map[string]any{
// 			"is_builtin = ?": true,
// 			"project = ?":    "oes",
// 			"label = ?":      "cmd",
// 			"name in ?":      tasks,
// 		},
// 		Columns: []string{"id", "name"},
// 	})
// 	if rErr != nil {
// 		return nil, rErr
// 	}
// 	if ms == nil || len(*ms) == 0 {
// 		return result, nil
// 	}
// 	for _, m := range *ms {
// 		uc.ucSchedule.CreateSchedule(ctx, jobsmodel.ScheduleModel{
// 			ScriptID: m.ID,
// 			Name:     m.Name,
// 		})
// 	}
// 	return result, nil
// }
