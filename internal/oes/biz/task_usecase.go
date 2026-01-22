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
	pbColony "gin-artweb/api/oes/colony"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

type OesTaskInfoUsecase struct {
	log      *zap.Logger
	stkTasks []string
	crdTasks []string
	optTasks []string
}

func NewOesTaskInfoUsecase(
	log *zap.Logger,
) *OesTaskInfoUsecase {
	return &OesTaskInfoUsecase{
		log:      log,
		stkTasks: []string{"mon_collector", "counter_fetch", "counter_distribute", "bse_collector", "sse_collector", "szse_collector", "csdc_collector"},
		crdTasks: []string{"mon_collector", "counter_fetch", "counter_distribute", "sse_collector", "szse_collector", "csdc_collector", "sse_late_collector", "szse_late_collector"},
		optTasks: []string{"mon_collector", "counter_fetch", "counter_distribute", "sse_collector", "szse_collector", "csdc_collector"},
	}
}

func (uc *OesTaskInfoUsecase) GetColonyTaskInfo(
	ctx context.Context,
	colonyNum string,
	systemType string,
) (*pbColony.OesColonyTaskInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	var tasks []string
	switch systemType {
	case "STK":
		tasks = uc.stkTasks
	case "CRD":
		tasks = uc.crdTasks
	case "OPT":
		tasks = uc.optTasks
	default:
		return nil, ErrOesColonySystemTypeInvalid.WithData(map[string]any{
			"colony_num":  colonyNum,
			"system_type": systemType,
		})
	}
	taskList := make([]pbComm.TaskInfo, len(tasks))
	for i, taskName := range tasks {
		taskInfo, err := uc.GetTaskInfo(ctx, colonyNum, taskName)
		if err != nil {
			return nil, err
		}
		taskList[i] = *taskInfo
	}
	return &pbColony.OesColonyTaskInfo{
		ColonyNum: colonyNum,
		Tasks:     taskList,
	}, nil
}

func (uc *OesTaskInfoUsecase) GetTaskInfo(
	ctx context.Context,
	colonyNum string,
	taskName string,
) (*pbComm.TaskInfo, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}
	flagBaseName := fmt.Sprintf("%s_%s.*", taskName, time.Now().Format("20060102"))
	flagPath := filepath.Join(config.StorageDir, "oes", "flags", colonyNum, flagBaseName)
	flagFiles, err := filepath.Glob(flagPath)
	if err != nil {
		uc.log.Error(
			"查询oes任务标识文件失败",
			zap.Error(err),
			zap.String("colony_num", colonyNum),
			zap.String("task_name", taskName),
			zap.String("pattern", flagPath),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}
	if len(flagFiles) > 1 {
		return nil, ErrOesColonyHasTooManyFlags.WithData(map[string]any{
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
		return nil, ErrOesColonyInvalidFlag.WithData(map[string]any{
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
