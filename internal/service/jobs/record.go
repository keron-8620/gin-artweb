package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	jobsmodel "gin-artweb/internal/model/jobs"
	jobsrepo "gin-artweb/internal/repository/jobs"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type RecordService struct {
	log        *zap.Logger
	scriptRepo *jobsrepo.ScriptRepo
	recordRepo *jobsrepo.RecordRepo
	contexts   map[uint32]context.CancelFunc
	mutex      sync.RWMutex
}

func NewScriptRecordService(
	log *zap.Logger,
	scriptRepo *jobsrepo.ScriptRepo,
	recordRepo *jobsrepo.RecordRepo,
) *RecordService {
	return &RecordService{
		log:        log,
		scriptRepo: scriptRepo,
		recordRepo: recordRepo,
		contexts:   make(map[uint32]context.CancelFunc),
	}
}

// 存储上下文
func (s *RecordService) StoreCancel(id uint32, cancel context.CancelFunc) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.contexts[id] = cancel
}

// 删除上下文
func (s *RecordService) DeleteCancel(id uint32) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.contexts, id)
}

// 获取上下文
func (s *RecordService) GetCancel(id uint32) context.CancelFunc {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.contexts[id]
}

func (s *RecordService) Execute(record *jobsmodel.ScriptRecordModel) *jobsmodel.TaskInfo {
	s.log.Debug(
		"开始执行脚本",
		zap.Object("script_record_id", record),
	)
	// 初始化执行任务
	ctx, cancel := context.WithCancel(context.Background())
	s.StoreCancel(record.ID, cancel)

	// 创建带超时的上下文
	timeout := time.Duration(record.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Minute // 默认超时时间
	}
	var timeoutCancel context.CancelFunc
	ctx, timeoutCancel = context.WithTimeout(ctx, timeout)
	defer timeoutCancel()

	taskinfo := &jobsmodel.TaskInfo{
		ExitCode: -1,
		Status:   3,
		ErrMSG:   "",
		Error:    nil,
		LogFile:  nil,
	}

	defer func() {
		// panic 恢复保护
		if r := recover(); r != nil {
			// 记录panic信息和堆栈跟踪
			stack := debug.Stack()

			// 构造错误响应
			switch v := r.(type) {
			case error:
				taskinfo.ErrMSG = v.Error()
			case string:
				taskinfo.ErrMSG = v
			default:
				taskinfo.ErrMSG = fmt.Sprintf("%v", v)
			}

			s.log.Error("脚本执行panic",
				zap.String("error", taskinfo.ErrMSG),
				zap.Any("panic", r),
				zap.String("stack", string(stack)),
				zap.Uint32("script_record_id", record.ID),
			)

			if taskinfo.LogFile != nil {
				format := time.Now().Format(time.RFC3339)
				fmt.Fprintf(taskinfo.LogFile, "[%s] [PANIC] 脚本执行发生严重错误: %s\n", format, taskinfo.ErrMSG)
				fmt.Fprintf(taskinfo.LogFile, "[%s] [STACK] %s\n", format, stack)
			}

			// 设置脚本状态为崩溃
			taskinfo.Status = 5
		}

		// 更新记录状态
		if err := s.UpdateScriptRecord(context.Background(), record.ID, taskinfo); err != nil {
			if taskinfo.LogFile != nil {
				fmt.Fprintf(taskinfo.LogFile, "[%s] [ERROR] 更新脚本记录状态失败: %s\n", time.Now().Format(time.RFC3339), err.Error())
			}
		}

		// 关闭日志文件句柄
		if taskinfo.LogFile != nil {
			taskinfo.LogFile.Close()
		}

		// 清理执行完成的上下文
		s.DeleteCancel(record.ID)

		// 输出日志
		s.log.Debug(
			"脚本执行完成",
			zap.Object("script_record_id", record),
		)
	}()

	// 生成日志路径并创建日志目录
	logPath := common.GetScriptLogStoragePath(record.CreatedAt.Format(time.DateOnly), record.LogName)
	logDir := filepath.Dir(logPath)
	if taskinfo.Error = os.MkdirAll(logDir, 0755); taskinfo.Error != nil {
		taskinfo.Status = 5
		taskinfo.ErrMSG = fmt.Sprintf("创建日志目录失败: %v", taskinfo.Error)
		s.log.Error(
			"创建日志目录失败",
			zap.Error(taskinfo.Error),
			zap.String("path", logDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return taskinfo
	}

	// 打开日志文件
	taskinfo.LogFile, taskinfo.Error = os.Create(logPath)
	if taskinfo.Error != nil {
		taskinfo.Status = 5
		taskinfo.ErrMSG = fmt.Sprintf("创建日志文件失败: %v", taskinfo.Error)
		s.log.Error(
			"创建日志文件失败",
			zap.Error(taskinfo.Error),
			zap.String("path", logPath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return taskinfo
	}

	// 写入开始执行日志
	startTime := time.Now()
	fmt.Fprintf(taskinfo.LogFile, "[%s] 开始执行脚本 (ID: %d, ScriptID: %d)\n",
		startTime.Format(time.RFC3339), record.ID, record.ScriptID)

	// 交验脚本是否存在
	scriptPath := common.GetScriptStoragePath(record.Script.Project, record.Script.Label, record.Script.Name, record.Script.IsBuiltin)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Fprintf(taskinfo.LogFile, "脚本文件不存在: %s\n", scriptPath)
		return taskinfo
	}

	// 解析命令参数
	var cmdArgs []string
	if record.CommandArgs != "" {
		cmdArgs = strings.Fields(record.CommandArgs)
	}

	cmd := exec.CommandContext(ctx, scriptPath, cmdArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}

		// 获取进程组ID
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			return syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			return cmd.Process.Kill()
		}
	}

	// 设置工作目录
	if record.WorkDir != "" {
		if _, err := os.Stat(record.WorkDir); os.IsNotExist(err) {
			fmt.Fprintf(taskinfo.LogFile, "工作目录不存在，尝试创建: %s\n", record.WorkDir)
			if err := os.MkdirAll(record.WorkDir, 0755); err != nil {
				fmt.Fprintf(taskinfo.LogFile, "创建工作目录失败: %s\n", err)
				return taskinfo
			}
		}
		cmd.Dir = record.WorkDir
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("JOBS_RECORD_ID=%d", record.ID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("JOBS_LOG_PATH=%s", logPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("JOBS_BASE_DIR=%s", config.BaseDir))
	if record.EnvVars != "" {
		var envMap map[string]string
		if err := json.Unmarshal([]byte(record.EnvVars), &envMap); err == nil {
			for k, v := range envMap {
				if k != "" {
					cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
				}
			}
		}
	}

	// 重定向输出到日志文件
	cmd.Stdout = taskinfo.LogFile
	cmd.Stderr = taskinfo.LogFile

	// 执行命令
	fmt.Fprintf(taskinfo.LogFile, "执行命令: %s %s\n", scriptPath, strings.Join(cmdArgs, " "))
	taskinfo.Error = cmd.Run()
	endTime := time.Now()
	duration := endTime.Sub(startTime).Seconds()

	// 检查执行结果
	if taskinfo.Error != nil {
		taskinfo.Status = 3 // 失败状态
		if exitError, ok := taskinfo.Error.(*exec.ExitError); ok {
			taskinfo.ExitCode = exitError.ExitCode()
		}
		fmt.Fprintf(taskinfo.LogFile, "[%s] 脚本执行失败 (退出码: %d, 耗时: %.3fs): %s\n",
			endTime.Format(time.RFC3339), taskinfo.ExitCode, duration, taskinfo.Error)
	} else {
		taskinfo.ExitCode = 0
		taskinfo.Status = 2 // 成功状态
		fmt.Fprintf(taskinfo.LogFile, "[%s] 脚本执行成功 (耗时: %.3fs)\n",
			endTime.Format(time.RFC3339), duration)
	}
	return taskinfo
}

func (s *RecordService) Cancel(ctx context.Context, recordID uint32) {
	cancel := s.GetCancel(recordID)
	if cancel == nil {
		s.log.Warn(
			"未找到要取消的脚本任务",
			zap.Uint32("script_record_id", recordID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
	} else {
		s.log.Info(
			"开始取消脚本执行",
			zap.Uint32("script_record_id", recordID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		cancel()
		s.DeleteCancel(recordID)
		s.log.Info(
			"取消脚本执行成功",
			zap.Uint32("script_record_id", recordID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
	}
}

func (s *RecordService) CreateScriptRecord(
	ctx context.Context,
	req jobsmodel.ExecuteRequest,
) (*jobsmodel.ScriptRecordModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	script, err := s.scriptRepo.GetModel(ctx, req.ScriptID)
	if err != nil {
		s.log.Error(
			"查询脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", req.ScriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": req.ScriptID})
	}

	if !script.Status {
		s.log.Error(
			"脚本已禁用",
			zap.Uint32("script_id", req.ScriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromReason(errors.ReasonScriptIsDisabled).WithField("script_id", req.ScriptID)
	}

	now := time.Now()
	record := &jobsmodel.ScriptRecordModel{
		StandardModel: database.StandardModel{
			CreatedAt: now,
			UpdatedAt: now,
		},
		TriggerType:  req.TriggerType,
		Status:       1, // 执行中
		ExitCode:     -1,
		EnvVars:      req.EnvVars,
		CommandArgs:  req.CommandArgs,
		WorkDir:      req.WorkDir,
		Timeout:      req.Timeout,
		LogName:      fmt.Sprintf("%s.log", uuid.NewString()),
		ErrorMessage: "",
		Username:     req.Username,
		ScriptID:     req.ScriptID,
	}

	if err := s.recordRepo.CreateModel(ctx, record); err != nil {
		s.log.Error(
			"创建执行记录失败",
			zap.Error(err),
			zap.Uint32("script_id", req.ScriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	record.Script = *script
	return record, nil
}

func (s *RecordService) UpdateScriptRecord(
	ctx context.Context,
	recordID uint32,
	taskinfo *jobsmodel.TaskInfo,
) *errors.Error {
	s.log.Debug(
		"开始更新执行记录状态",
	)
	if err := s.recordRepo.UpdateModel(ctx, taskinfo.ToMap(), "id = ?", recordID); err != nil {
		s.log.Error(
			"更新脚本记录失败",
			zap.Error(err),
			zap.Uint32("script_record_id", recordID),
			zap.Object("taskinfo", taskinfo),
		)
		return errors.NewGormError(err, taskinfo.ToMap())
	}
	return nil
}

func (s *RecordService) FindScriptRecordByID(
	ctx context.Context,
	preloads []string,
	recordID uint32,
) (*jobsmodel.ScriptRecordModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询脚本执行记录",
		zap.Uint32("script_record_id", recordID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.recordRepo.GetModel(ctx, preloads, recordID)
	if err != nil {
		s.log.Error(
			"查询脚本执行记录失败",
			zap.Error(err),
			zap.Uint32("script_record_id", recordID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": recordID})

	}

	s.log.Info(
		"查询脚本执行记录成功",
		zap.Uint32("script_record_id", recordID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *RecordService) ListcriptRecord(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]jobsmodel.ScriptRecordModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询脚本执行记录列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.recordRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询脚本执行记录列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询脚本执行记录列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *RecordService) AsyncExecuteScript(
	ctx context.Context,
	req jobsmodel.ExecuteRequest,
) (*jobsmodel.ScriptRecordModel, *errors.Error) {
	record, err := s.CreateScriptRecord(ctx, req)
	if err != nil {
		return nil, err
	}
	go s.Execute(record)
	return record, nil
}

func (s *RecordService) SyncExecuteScript(
	ctx context.Context,
	req jobsmodel.ExecuteRequest,
) (*jobsmodel.TaskInfo, *errors.Error) {
	record, err := s.CreateScriptRecord(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.Execute(record), nil
}
