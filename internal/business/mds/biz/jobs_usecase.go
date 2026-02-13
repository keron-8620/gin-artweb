package biz

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"

	bizJobs "gin-artweb/internal/infra/jobs/biz"
	jobsModel "gin-artweb/internal/infra/jobs/model"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type JobsUsecase struct {
	log        *zap.Logger
	ucScript   *bizJobs.ScriptUsecase
	ucRecord   *bizJobs.RecordUsecase
	ucSchedule *bizJobs.ScheduleUsecase
}

func NewRecordUsecase(
	log *zap.Logger,
	ucScript *bizJobs.ScriptUsecase,
	ucRecord *bizJobs.RecordUsecase,
	ucSchedule *bizJobs.ScheduleUsecase,
) *JobsUsecase {
	return &JobsUsecase{
		log:        log,
		ucScript:   ucScript,
		ucRecord:   ucRecord,
		ucSchedule: ucSchedule,
	}
}

// readUint32FromFile 从指定文件读取单个数字并转换为uint32
func (uc *JobsUsecase) ReadUint32FromFile(filePath string) uint32 {
	if _, err := os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			uc.log.Error(
				"获取文件状态失败",
				zap.Error(err),
				zap.String("filepath", filePath),
			)
		}
		return 0
	}
	// 读取文件内容
	file, err := os.Open(filePath)
	if err != nil {
		uc.log.Error(
			"打开文件失败",
			zap.Error(err),
			zap.String("filepath", filePath),
		)
		return 0
	}
	defer file.Close() // 确保文件句柄关闭

	// 读取第一行内容（因为文件只有一个数字）
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		uc.log.Error(
			"文件为空或读取失败",
			zap.Error(err),
			zap.String("filepath", filePath),
		)
		return 0
	}
	content := scanner.Text()

	// 去除空白字符（防止文件有换行/空格）
	numberStr := strings.TrimSpace(content)

	// 将字符串转换为uint64（先转uint64避免溢出判断），再转为uint32
	numberUint64, err := strconv.ParseUint(numberStr, 10, 32)
	if err != nil {
		uc.log.Error(
			"转换为uint32失败",
			zap.Error(err),
			zap.String("filepath", filePath),
		)
		return 0
	}

	// 转为uint32
	return uint32(numberUint64)
}

func (uc *JobsUsecase) FindRecordsByIDs(
	ctx context.Context,
	recordIDs []uint32,
) (map[uint32]jobsModel.ScriptRecordModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(recordIDs) == 0 {
		return map[uint32]jobsModel.ScriptRecordModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": recordIDs},
	}
	_, ms, rErr := uc.ucRecord.ListcriptRecord(ctx, qp)
	if rErr != nil {
		return nil, rErr
	}
	if ms == nil || len(*ms) == 0 {
		return map[uint32]jobsModel.ScriptRecordModel{}, nil
	}
	rms := *ms
	result := make(map[uint32]jobsModel.ScriptRecordModel, len(rms))
	for _, m := range rms {
		result[m.ID] = m
	}
	return result, nil
}

func (uc *JobsUsecase) FindRecordsByMap(
	ctx context.Context,
	cache map[uint32]jobsModel.ScriptRecordModel,
	recordID uint32,
	colonyNum, taskName string,
) *jobsModel.ScriptRecordModel {
	task, exists := cache[recordID]
	if !exists {
		uc.log.Debug(
			"未找到mds的任务状态",
			zap.String("colony_num", colonyNum),
			zap.String("task_name", taskName),
			zap.Uint32("script_record_id", recordID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil
	}
	uc.log.Debug(
		"获取mds的任务状态成功",
		zap.String("colony_num", colonyNum),
		zap.String("task_name", taskName),
		zap.Uint32("script_record_id", recordID),
		zap.Object("task", &task),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &task
}

func (uc *JobsUsecase) FindScriptIDs(
	ctx context.Context,
	tasks []string,
) (map[string]uint32, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	result := make(map[string]uint32, len(tasks))
	_, ms, rErr := uc.ucScript.ListScript(ctx, database.QueryParams{
		Query: map[string]any{
			"is_builtin = ?": true,
			"project = ?":    "mds",
			"label = ?":      "cmd",
			"name in ?":      tasks,
		},
		Columns: []string{"id", "name"},
	})
	if rErr != nil {
		return nil, rErr
	}
	if ms == nil || len(*ms) == 0 {
		return result, nil
	}
	for _, m := range *ms {
		uc.ucSchedule.CreateSchedule(ctx, jobsModel.ScheduleModel{
			ScriptID: m.ID,
			Name:     m.Name,
		})
	}
	return result, nil
}
