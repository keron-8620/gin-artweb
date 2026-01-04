package biz

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/errors"
)

type ColonyTasksInfo struct {
	ColonyNum string `json:"colony_num"`
	Mon       string `json:"mon"`
	Sse       string `json:"sse"`
	Szse      string `json:"szse"`
}

type MdsTaskInfoUsecase struct {
	log        *zap.Logger
	colonyRepo MdsColonyRepo
}

func NewMdsTaskInfoUsecase(
	log *zap.Logger,
	colonyRepo MdsColonyRepo,
) *MdsTaskInfoUsecase {
	return &MdsTaskInfoUsecase{
		log:        log,
		colonyRepo: colonyRepo,
	}
}

func (uc *MdsTaskInfoUsecase) GetMdsTaskInfo(ctx context.Context, colonyNum string) (*ColonyTasksInfo, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	colonyTaskInfo := &ColonyTasksInfo{
		ColonyNum: colonyNum,
		Mon:       "unstart",
		Sse:       "unstart",
		Szse:      "unstart",
	}

	flagDir := filepath.Join(config.StorageDir, "mds", "bin", colonyNum)
	flagBaseName := "_collector_" + time.Now().Format("20060102") + ".*"

	// 获取mon任务状态
	monStatus, err := uc.getTaskInfo(ctx, filepath.Join(flagDir, "mon"+flagBaseName))
	if err != nil {
		return nil, errors.FromError(err)
	}
	colonyTaskInfo.Mon = monStatus

	// 获取sse任务状态
	sseStatus, err := uc.getTaskInfo(ctx, filepath.Join(flagDir, "sse"+flagBaseName))
	if err != nil {
		return nil, errors.FromError(err)
	}
	colonyTaskInfo.Sse = sseStatus

	// 获取szse任务状态
	szseStatus, err := uc.getTaskInfo(ctx, filepath.Join(flagDir, "szse"+flagBaseName))
	if err != nil {
		return nil, errors.FromError(err)
	}
	colonyTaskInfo.Szse = szseStatus
	return colonyTaskInfo, nil
}

func (uc *MdsTaskInfoUsecase) getTaskInfo(ctx context.Context, pattern string) (string, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return "", errors.FromError(err)
	}

	flagFiles, err := filepath.Glob(pattern)
	if err != nil {
		uc.log.Error(
			"查询mds任务标识文件失败",
			zap.Error(err),
			zap.String("pattern", pattern),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return "", errors.FromError(err)
	}
	if len(flagFiles) == 1 {
		flagFile := flagFiles[0]
		slice := strings.Split(flagFile, ".")
		if len(slice) > 1 {
			return slice[len(slice)-1], nil
		}
	} else if len(flagFiles) > 1 {
		runningFlag := pattern[:len(pattern)-1] + "running"
		if _, err := os.Stat(runningFlag); err == nil {
			return "running", nil
		} else if os.IsNotExist(err) {
			return "failed", nil
		} else {
			uc.log.Error(
				"检查mds的running标志文件时出错",
				zap.Error(err),
				zap.String("running_flag", runningFlag),
				zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			)
			return "", errors.FromError(err)
		}
	}
	return "unstart", nil
}
