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

type RecordUsecase struct {
	log      *zap.Logger
	ucRecord *bizJobs.RecordUsecase
}

func NewRecordUsecase(
	log *zap.Logger,
	ucRecord *bizJobs.RecordUsecase,
) *RecordUsecase {
	return &RecordUsecase{
		log:      log,
		ucRecord: ucRecord,
	}
}

// readUint32FromFile 从指定文件读取单个数字并转换为uint32
func (uc *RecordUsecase) ReadUint32FromFile(filePath string) uint32 {
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

func (uc *RecordUsecase) FindRecordsByIDs(
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

func (uc *RecordUsecase) FindRecordsByMap(
	ctx context.Context,
	cache map[uint32]jobsModel.ScriptRecordModel,
	recordID uint32,
	colonyNum, taskName string,
) *jobsModel.ScriptRecordModel {
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
