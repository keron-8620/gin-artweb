package job

import (
	"context"
	"encoding/json"
	stdErrors "errors"
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

	jobmodel "gin-artweb/internal/model/job"
	jobrepo "gin-artweb/internal/repo/job"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type RecordService struct {
	log        *zap.Logger
	scriptRepo *jobrepo.ScriptRepo
	recordRepo *jobrepo.RecordRepo
	contexts   map[uint32]context.CancelFunc
	mutex      sync.RWMutex
}

func NewScriptRecordService(
	log *zap.Logger,
	scriptRepo *jobrepo.ScriptRepo,
	recordRepo *jobrepo.RecordRepo,
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

func (s *RecordService) Execute(
	ctx context.Context,
	record *jobmodel.ScriptRecordModel,
) *jobmodel.TaskInfo {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"脚本执行:参数详情",
		zap.Object("record_model", record),
	)

	// 初始化执行任务
	exCtx, cancel := context.WithCancel(ctx)
	s.StoreCancel(record.ID, cancel)

	// 创建带超时的上下文
	timeout := time.Duration(record.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 1 * time.Hour // 默认超时时间
	}
	var timeoutCancel context.CancelFunc
	exCtx, timeoutCancel = context.WithTimeout(exCtx, timeout)
	defer timeoutCancel()

	taskinfo := &jobmodel.TaskInfo{
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

			log.Error(
				"脚本执行:发生panic",
				zap.String("error", taskinfo.ErrMSG),
				zap.Any("panic", r),
				zap.String("stack", string(stack)),
				zap.Uint32("script_record_id", record.ID),
				zap.Duration("total_duration", time.Since(startTime)),
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
			log.Error(
				"脚本执行:更新脚本记录状态失败",
				zap.Error(err),
				zap.Uint32("script_record_id", record.ID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}

		// 关闭日志文件句柄
		if taskinfo.LogFile != nil {
			_ = taskinfo.LogFile.Close()
		}

		// 清理执行完成的上下文
		s.DeleteCancel(record.ID)

		// 输出日志
		log.Debug(
			"脚本执行:执行完成",
			zap.Int("status", taskinfo.Status),
			zap.Int("exit_code", taskinfo.ExitCode),
			zap.String("error_message", taskinfo.ErrMSG),
			zap.Duration("total_duration", time.Since(startTime)),
		)
	}()

	// 生成日志路径并创建日志目录
	logPath := s.GenerateScriptLogPath(record.CreatedAt, record.LogName)
	logDir := filepath.Dir(logPath)
	if taskinfo.Error = os.MkdirAll(logDir, 0750); taskinfo.Error != nil {
		taskinfo.Status = 5
		taskinfo.ErrMSG = fmt.Sprintf("创建日志目录失败: %v", taskinfo.Error)
		log.Error(
			"脚本执行:创建日志目录失败",
			zap.Error(taskinfo.Error),
			zap.String("path", logDir),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return taskinfo
	}

	// 打开日志文件
	taskinfo.LogFile, taskinfo.Error = os.Create(logPath)
	if taskinfo.Error != nil {
		taskinfo.Status = 5
		taskinfo.ErrMSG = fmt.Sprintf("创建日志文件失败: %v", taskinfo.Error)
		log.Error(
			"脚本执行:创建日志文件失败",
			zap.Error(taskinfo.Error),
			zap.String("path", logPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return taskinfo
	}

	// 写入开始执行日志
	execStartTime := time.Now()
	fmt.Fprintf(taskinfo.LogFile, "[%s] 开始执行脚本 (ID: %d, ScriptID: %d)\n",
		execStartTime.Format(time.RFC3339), record.ID, record.ScriptID)

	// 交验脚本是否存在
	scriptPath := GetScriptStoragePath(
		record.Script.Project,
		record.Script.Label,
		record.Script.Name,
		record.Script.IsBuiltin,
	)

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		taskinfo.Status = 5
		taskinfo.ErrMSG = fmt.Sprintf("脚本文件不存在: %s", scriptPath)
		fmt.Fprintf(taskinfo.LogFile, "脚本文件不存在: %s\n", scriptPath)
		log.Error(
			"脚本执行:脚本文件不存在",
			zap.String("script_path", scriptPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return taskinfo
	}
	log.Debug("脚本执行:脚本文件存在", zap.String("script_path", scriptPath))

	// 解析命令参数
	var cmdArgs []string
	if record.CommandArgs != "" {
		cmdArgs = strings.Fields(record.CommandArgs)
	}
	log.Debug(
		"脚本执行:解析命令参数",
		zap.String("command_args", record.CommandArgs),
		zap.Strings("parsed_args", cmdArgs),
	)

	// #nosec G204 -- 创建命令执行上下文
	cmd := exec.CommandContext(exCtx, scriptPath, cmdArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}

		// 先尝试优雅终止（SIGTERM）, 再强制终止（SIGKILL）
		process := cmd.Process
		pid := process.Pid

		// 尝试获取进程组ID
		pgid, err := syscall.Getpgid(pid)
		if err != nil {
			log.Debug(
				"脚本执行:获取进程组ID失败, 将直接终止单个进程",
				zap.Error(err),
				zap.Int("pid", pid),
			)

			// 先尝试优雅终止单个进程
			if err := process.Signal(syscall.SIGTERM); err != nil {
				log.Debug(
					"脚本执行:发送SIGTERM信号失败, 尝试强制终止",
					zap.Error(err),
					zap.Int("pid", pid),
				)
				return process.Kill()
			}

			// 等待进程终止
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				if err := process.Signal(syscall.Signal(0)); err != nil {
					// 进程已终止
					log.Debug(
						"脚本执行:进程已成功终止",
						zap.Int("pid", pid),
					)
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}

			// 超时后强制终止
			log.Debug(
				"脚本执行:进程终止超时, 尝试强制终止",
				zap.Int("pid", pid),
			)
			return process.Kill()
		}

		// 向整个进程组发送SIGTERM信号
		log.Debug(
			"脚本执行:向进程组发送SIGTERM信号",
			zap.Int("pgid", pgid),
			zap.Int("pid", pid),
		)

		if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
			log.Debug(
				"脚本执行:发送SIGTERM信号到进程组失败, 尝试发送SIGKILL",
				zap.Error(err),
				zap.Int("pgid", pgid),
			)
			return syscall.Kill(-pgid, syscall.SIGKILL)
		}

		// 等待进程组终止
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if err := process.Signal(syscall.Signal(0)); err != nil {
				// 进程已终止
				log.Debug(
					"脚本执行:进程组已成功终止",
					zap.Int("pgid", pgid),
					zap.Int("pid", pid),
				)
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}

		// 超时后强制终止整个进程组
		log.Debug(
			"脚本执行:进程组终止超时, 尝试强制终止",
			zap.Int("pgid", pgid),
			zap.Int("pid", pid),
		)
		return syscall.Kill(-pgid, syscall.SIGKILL)
	}

	// 设置工作目录
	if record.WorkDir != "" {
		log.Debug(
			"脚本执行:设置工作目录",
			zap.String("work_dir", record.WorkDir),
		)
		if _, err := os.Stat(record.WorkDir); os.IsNotExist(err) {
			fmt.Fprintf(taskinfo.LogFile, "工作目录不存在, 尝试创建: %s\n", record.WorkDir)
			log.Debug(
				"脚本执行:工作目录不存在, 尝试创建",
				zap.String("work_dir", record.WorkDir),
			)
			if err := os.MkdirAll(record.WorkDir, 0750); err != nil {
				taskinfo.Status = 5
				taskinfo.ErrMSG = fmt.Sprintf("创建工作目录失败: %v", err)
				fmt.Fprintf(taskinfo.LogFile, "创建工作目录失败: %s\n", err)
				log.Error(
					"脚本执行:创建工作目录失败",
					zap.Error(err),
					zap.String("work_dir", record.WorkDir),
					zap.Duration("total_duration", time.Since(startTime)),
				)
				return taskinfo
			}
			log.Debug(
				"脚本执行:工作目录创建成功",
				zap.String("work_dir", record.WorkDir),
			)
		}
		cmd.Dir = record.WorkDir
		log.Debug(
			"脚本执行:工作目录设置成功",
			zap.String("work_dir", record.WorkDir),
		)
	}

	// 设置环境变量
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("JOB_RECORD_ID=%d", record.ID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("JOB_LOG_PATH=%s", logPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("JOB_BASE_DIR=%s", config.BaseDir))
	if record.EnvVars != "" {
		var envMap map[string]string
		if err := json.Unmarshal([]byte(record.EnvVars), &envMap); err == nil {
			for k, v := range envMap {
				if k != "" {
					cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
				}
			}
			log.Debug("脚本执行:环境变量设置成功", zap.Int("env_var_count", len(envMap)))
		} else {
			log.Warn(
				"脚本执行:解析环境变量失败",
				zap.Error(err),
				zap.String("env_vars", record.EnvVars),
			)
		}
	}

	// 重定向输出到日志文件
	cmd.Stdout = taskinfo.LogFile
	cmd.Stderr = taskinfo.LogFile

	// 执行命令
	log.Info(
		"脚本执行:开始执行命令",
		zap.String("script_path", scriptPath),
		zap.Strings("args", cmdArgs),
	)
	fmt.Fprintf(taskinfo.LogFile, "执行命令: %s %s\n", scriptPath, strings.Join(cmdArgs, " "))
	taskinfo.Error = cmd.Run()
	execEndTime := time.Now()
	execDuration := execEndTime.Sub(execStartTime).Seconds()

	// 检查执行结果
	if taskinfo.Error != nil {
		// 检查是否是超时错误
		if stdErrors.Is(taskinfo.Error, context.DeadlineExceeded) {
			taskinfo.Status = 4 // 超时状态
			taskinfo.ErrMSG = "脚本执行超时"
			fmt.Fprintf(taskinfo.LogFile, "[%s] 脚本执行超时 (耗时: %.3fs)\n",
				execEndTime.Format(time.RFC3339), execDuration)
			log.Error(
				"脚本执行:脚本执行超时",
				zap.Duration("execution_duration", execEndTime.Sub(execStartTime)),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		} else {
			taskinfo.Status = 3 // 失败状态
			if exitError, ok := taskinfo.Error.(*exec.ExitError); ok {
				taskinfo.ExitCode = exitError.ExitCode()
			}
			fmt.Fprintf(taskinfo.LogFile, "[%s] 脚本执行失败 (退出码: %d, 耗时: %.3fs): %s\n",
				execEndTime.Format(time.RFC3339), taskinfo.ExitCode, execDuration, taskinfo.Error)
			log.Error(
				"脚本执行:脚本执行失败",
				zap.Error(taskinfo.Error),
				zap.Int("exit_code", taskinfo.ExitCode),
				zap.Duration("execution_duration", execEndTime.Sub(execStartTime)),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	} else {
		taskinfo.ExitCode = 0
		taskinfo.Status = 2 // 成功状态
		fmt.Fprintf(taskinfo.LogFile, "[%s] 脚本执行成功 (耗时: %.3fs)\n",
			execEndTime.Format(time.RFC3339), execDuration)
		log.Info(
			"脚本执行:脚本执行成功",
			zap.Int("exit_code", taskinfo.ExitCode),
			zap.Duration("execution_duration", execEndTime.Sub(execStartTime)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
	}
	return taskinfo
}

func (s *RecordService) Cancel(ctx context.Context, recordID uint32) {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	cancel := s.GetCancel(recordID)
	if cancel == nil {
		log.Warn(
			"未找到要取消的脚本任务",
			zap.Uint32("script_record_id", recordID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
	} else {
		log.Info(
			"开始取消脚本执行",
			zap.Uint32("script_record_id", recordID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		cancel()
		log.Info(
			"取消脚本执行成功",
			zap.Uint32("script_record_id", recordID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
	}
}

func (s *RecordService) GenerateScriptLogPath(data time.Time, logName string) string {
	return filepath.Join(config.StorageDir, "logs", data.Format(time.DateOnly), logName)
}

func (s *RecordService) CreateScriptRecord(
	ctx context.Context,
	dto jobmodel.ExecuteScriptDTO,
) (*jobmodel.ScriptRecordModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建脚本执行记录:开始执行",
		zap.Object("execute_script_dto", &dto),
	)

	script, err := s.scriptRepo.GetModel(ctx, dto.ScriptID)
	if err != nil {
		log.Error(
			"创建脚本执行记录:查询脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", dto.ScriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": dto.ScriptID})
	}

	if !script.Status {
		log.Error(
			"创建脚本执行记录:脚本已禁用",
			zap.Uint32("script_id", dto.ScriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.FromReason(errors.ReasonScriptIsDisabled).WithField("script_id", dto.ScriptID)
	}

	record := jobmodel.ExecuteScriptDTOToModel(dto, fmt.Sprintf("%s.log", uuid.NewString()))
	if err := s.recordRepo.CreateModel(ctx, &record); err != nil {
		log.Error(
			"创建脚本执行记录:创建执行记录失败",
			zap.Error(err),
			zap.Uint32("script_id", dto.ScriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	record.Script = *script

	log.Info(
		"创建脚本执行记录:执行成功",
		zap.Uint32("script_record_id", record.ID),
		zap.Uint32("script_id", dto.ScriptID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &record, nil
}

func (s *RecordService) UpdateScriptRecord(
	ctx context.Context,
	recordID uint32,
	taskinfo *jobmodel.TaskInfo,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新脚本执行记录:开始执行",
		zap.Uint32("script_record_id", recordID),
		zap.Object("taskinfo", taskinfo),
	)

	updateData := taskinfo.ToUpdateMap()
	if err := s.recordRepo.UpdateModel(ctx, updateData, "id = ?", recordID); err != nil {
		log.Error(
			"更新脚本执行记录:更新脚本记录失败",
			zap.Error(err),
			zap.Uint32("script_record_id", recordID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, updateData)
	}

	log.Info(
		"更新脚本执行记录:执行成功",
		zap.Uint32("script_record_id", recordID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *RecordService) FindScriptRecordByID(
	ctx context.Context,
	preloads []string,
	recordID uint32,
) (*jobmodel.ScriptRecordModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.recordRepo.GetModel(ctx, preloads, recordID)
	if err != nil {
		log.Error(
			"查询脚本执行记录:查询脚本执行记录失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("script_record_id", recordID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": recordID})
	}

	log.Debug(
		"查询脚本执行记录:查询脚本执行记录详情",
		zap.Object("record_model", m),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *RecordService) ListScriptRecord(
	ctx context.Context,
	page, size int,
	dto *jobmodel.ListScriptRecordDTO,
) (int64, []jobmodel.ScriptRecordModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询脚本执行记录列表:参数详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("dto", dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Script"},
		Limit:    limit,
		Offset:   offset,
		OrderBy:  []string{"id DESC"},
		Query:    dto.ToQueryMap(),
	}

	count, err := s.recordRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询脚本执行记录列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询脚本执行记录列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.recordRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询脚本执行记录列表:数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *RecordService) AsyncExecuteScript(
	ctx context.Context,
	req jobmodel.ExecuteScriptDTO,
) (*jobmodel.ScriptRecordModel, *errors.Error) {
	record, err := s.CreateScriptRecord(ctx, req)
	if err != nil {
		return nil, err
	}

	traceID := ctxutil.GetTraceID(ctx)
	bgCtx := context.WithValue(context.Background(), ctxutil.TraceIDKey, traceID)
	go s.Execute(bgCtx, record)
	return record, nil
}

func (s *RecordService) SyncExecuteScript(
	ctx context.Context,
	req jobmodel.ExecuteScriptDTO,
) (*jobmodel.TaskInfo, *errors.Error) {
	record, err := s.CreateScriptRecord(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.Execute(ctx, record), nil
}

func (s *RecordService) ListScriptRecordByIDs(
	ctx context.Context,
	preloads []string,
	recordIDs []uint32,
) ([]jobmodel.ScriptRecordModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询指定的脚本执行记录列表:参数详情",
		zap.Strings("preloads", preloads),
		zap.Uint32s("script_record_ids", recordIDs),
	)

	qp := database.QueryParams{
		Preloads: preloads,
		Query:    map[string]any{"id in ?": recordIDs},
	}

	ms, err := s.recordRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询指定的脚本执行记录列表:数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	return ms, nil
}

func GetScriptLogStoragePath(data time.Time, logname string) string {
	return filepath.Join(config.StorageDir, "logs", data.Format(time.DateOnly), logname)
}

func GetRecordIDByMap(
	cache map[uint32]jobmodel.ScriptRecordModel,
	recordID uint32,
) *jobmodel.ScriptRecordModel {
	task, exists := cache[recordID]
	if !exists {
		return nil
	}
	return &task
}

func BuildTaskInfoFromScriptRecord(
	taskName string,
	m *jobmodel.ScriptRecordModel,
) jobmodel.BizTaskInfo {
	result := jobmodel.BizTaskInfo{
		TaskName: taskName,
	}
	if m != nil {
		result.RecordID = m.ID
		result.Status = m.Status
		result.StartTime = m.CreatedAt.Format(time.DateTime)
		result.EndTime = m.UpdatedAt.Format(time.DateTime)
		result.TriggerType = m.TriggerType
	}
	return result
}
