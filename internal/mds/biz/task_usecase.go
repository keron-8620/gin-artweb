package biz

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbColony "gin-artweb/api/mds/colony"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

type MdsTaskInfoUsecase struct {
	log   *zap.Logger
	flags map[string]string
	tasks []string
}

func NewMdsTaskInfoUsecase(
	log *zap.Logger,
) *MdsTaskInfoUsecase {
	return &MdsTaskInfoUsecase{
		log: log,
		flags: map[string]string{
			"mon":  "mon_collector",
			"bse":  "bse_collector",
			"sse":  "sse_collector",
			"szse": "szse_collector",
		},
		tasks: []string{"mon", "bse", "sse", "szse"},
	}
}

func (uc *MdsTaskInfoUsecase) GetColonyTaskInfo(
	ctx context.Context,
	colonyNum string,
) (*pbColony.MdsColonyTaskInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	taskList := make([]pbComm.TaskInfo, len(uc.tasks))
	for i, taskName := range uc.tasks {
		taskInfo, err := uc.GetTaskInfo(ctx, colonyNum, taskName)
		if err != nil {
			return nil, err
		}
		taskList[i] = *taskInfo
	}
	return &pbColony.MdsColonyTaskInfo{
		ColonyNum: colonyNum,
		Tasks:     taskList,
	}, nil
}

func (uc *MdsTaskInfoUsecase) GetTaskInfo(
	ctx context.Context,
	colonyNum string,
	taskName string,
) (*pbComm.TaskInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	flagBaseName, exists := uc.flags[taskName]
	if !exists {
		return nil, ErrMdsColonyTaskUnKnown.WithData(map[string]any{
			"colony_num": colonyNum,
			"task_name":  taskName,
		})
	}
	flagName := fmt.Sprintf("%s_%s.*", flagBaseName, time.Now().Format("20060102"))
	flagPath := filepath.Join(config.StorageDir, "mds", "flags", colonyNum, flagName)
	flagFiles, err := filepath.Glob(flagPath)
	if err != nil {
		uc.log.Error(
			"查询mds任务标识文件失败",
			zap.Error(err),
			zap.String("colony_num", colonyNum),
			zap.String("task_name", taskName),
			zap.String("pattern", flagPath),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}
	if len(flagFiles) > 1 {
		return nil, ErrMdsColonyHasTooManyFlags.WithData(map[string]any{
			"colony_num": colonyNum,
			"task_name":  taskName,
		})
	}
	if len(flagFiles) == 0 {
		return &pbComm.TaskInfo{
			TaskName:   taskName,
			TaskTastus: "--",
			RecordID:   0,
		}, nil
	}
	flagFile := flagFiles[0]
	slice := strings.Split(flagFile, ".")
	if len(slice) == 1 {
		return nil, ErrMdsColonyInvalidFlag.WithData(map[string]any{
			"colony_num": colonyNum,
			"task_name":  taskName,
			"flagpath":   flagFile,
		})
	}
	recordID, err := readUint32FromFile(flagFile)
	if err != nil {
		return nil, errors.FromError(err)
	}
	return &pbComm.TaskInfo{
		TaskName:   taskName,
		TaskTastus: slice[len(slice)-1],
		RecordID:   recordID,
	}, nil
}

// readUint32FromFile 从指定文件读取单个数字并转换为uint32
func readUint32FromFile(filePath string) (uint32, error) {
	// 读取文件内容
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close() // 确保文件句柄关闭

	// 读取第一行内容（因为文件只有一个数字）
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, fmt.Errorf("文件为空或读取失败")
	}
	content := scanner.Text()

	// 去除空白字符（防止文件有换行/空格）
	numberStr := strings.TrimSpace(content)

	// 将字符串转换为uint64（先转uint64避免溢出判断），再转为uint32
	numberUint64, err := strconv.ParseUint(numberStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("转换为uint32失败: %v", err)
	}

	// 转为uint32（因为ParseUint指定了32位，这里不会有精度丢失）
	return uint32(numberUint64), nil
}
