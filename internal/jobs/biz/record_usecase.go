package biz

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
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

const ScriptRecordIDKey = "script_record_id"

type ScriptRecordModel struct {
	database.StandardModel
	TriggerType  string      `gorm:"column:trigger_type;type:varchar(20);comment:触发类型(cron/api)" json:"trigger_type"`
	Status       int         `gorm:"column:status;type:tinyint;not null;default:0;comment:执行状态(0-待执行,1-执行中,2-成功,3-失败,4-超时,5-崩溃)" json:"status"`
	ExitCode     int         `gorm:"column:exit_code;comment:退出码" json:"exit_code"`
	EnvVars      string      `gorm:"column:env_vars;type:json;comment:环境变量(JSON对象)" json:"env_vars"`
	CommandArgs  string      `gorm:"column:command_args;type:varchar(254);comment:命令行参数(JSON数组)" json:"command_args"`
	WorkDir      string      `gorm:"column:work_dir;type:varchar(255);comment:工作目录" json:"work_dir"`
	Timeout      int         `gorm:"column:timeout;type:int;not null;default:300;comment:超时时间(秒)" json:"timeout"`
	LogName      string      `gorm:"column:log_name;type:varchar(255);comment:日志文件路径" json:"log_name"`
	ErrorMessage string      `gorm:"column:error_message;type:text;comment:错误信息" json:"error_message"`
	Username     string      `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	ScriptID     uint32      `gorm:"column:script_id;not null;index;comment:脚本ID" json:"script_id"`
	Script       ScriptModel `gorm:"foreignKey:ScriptID;references:ID" json:"script"`
}

func (m *ScriptRecordModel) TableName() string {
	return "jobs_script_record"
}

func (m *ScriptRecordModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("trigger_type", m.TriggerType)
	enc.AddInt("status", m.Status)
	enc.AddInt("exit_code", m.ExitCode)
	enc.AddString("env_vars", m.EnvVars)
	enc.AddString("command_args", m.CommandArgs)
	enc.AddString("work_dir", m.WorkDir)
	enc.AddString("log_path", m.LogName)
	enc.AddString("username", m.Username)
	enc.AddUint32("script_id", m.ScriptID)
	return nil
}

func (m *ScriptRecordModel) InitEnv() []string {
	env := os.Environ()
	env = append(env, fmt.Sprintf("JOB_LOG_PATH=%s", m.LogPath()))
	env = append(env, fmt.Sprintf("JOB_BASE_DIR=%s", config.BaseDir))
	if m.EnvVars != "" {
		var envMap map[string]string
		if err := json.Unmarshal([]byte(m.EnvVars), &envMap); err == nil {
			for k, v := range envMap {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}
	return env
}

func (m *ScriptRecordModel) LogPath() string {
	return filepath.Join(config.LogDir, m.CreatedAt.Format(time.DateOnly), m.LogName)
}

type ScriptRecordRepo interface {
	CreateModel(context.Context, *ScriptRecordModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*ScriptRecordModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]ScriptRecordModel, error)
}

type TaskInfo struct {
	Status   int
	ExitCode int
	ErrMSG   string
	Error    error
	LogFile  *os.File
}

func (t *TaskInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("status", t.Status)
	enc.AddInt("exit_code", t.ExitCode)
	enc.AddString("error_message", t.ErrMSG)
	return nil
}

func (t *TaskInfo) ToMap() map[string]any {
	return map[string]any{
		"status":        t.Status,
		"exit_code":     t.ExitCode,
		"error_message": t.ErrMSG,
	}
}

type ExecuteRequest struct {
	TriggerType string `json:"trigger_type"`
	ScriptID    uint32 `json:"script_id"`
	CommandArgs string `json:"command_args"`
	EnvVars     string `json:"env_vars"`
	Timeout     int    `json:"timeout"`
	WorkDir     string `json:"work_dir"`
	Username    string `json:"username"`
}

type RecordUsecase struct {
	log        *zap.Logger
	scriptRepo ScriptRepo
	recordRepo ScriptRecordRepo
	contexts   map[uint32]context.CancelFunc
	mutex      sync.RWMutex
}

func NewScriptRecordUsecase(
	log *zap.Logger,
	scriptRepo ScriptRepo,
	recordRepo ScriptRecordRepo,
) *RecordUsecase {
	return &RecordUsecase{
		log:        log,
		scriptRepo: scriptRepo,
		recordRepo: recordRepo,
		contexts:   make(map[uint32]context.CancelFunc),
	}
}

// 存储上下文
func (uc *RecordUsecase) StoreCancel(id uint32, cancel context.CancelFunc) {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()
	uc.contexts[id] = cancel
}

// 删除上下文
func (uc *RecordUsecase) DeleteCancel(id uint32) {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()
	delete(uc.contexts, id)
}

// 获取上下文
func (uc *RecordUsecase) GetCancel(id uint32) context.CancelFunc {
	uc.mutex.RLock()
	defer uc.mutex.RUnlock()
	return uc.contexts[id]
}

func (uc *RecordUsecase) Execute(record *ScriptRecordModel) *TaskInfo {
	uc.log.Debug(
		"开始执行脚本",
		zap.Object(ScriptRecordIDKey, record),
	)
	// 初始化执行任务
	ctx, cancel := context.WithCancel(context.Background())
	uc.StoreCancel(record.ID, cancel)

	// 创建带超时的上下文
	timeout := time.Duration(record.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Minute // 默认超时时间
	}
	var timeoutCancel context.CancelFunc
	ctx, timeoutCancel = context.WithTimeout(ctx, timeout)
	defer timeoutCancel()

	taskinfo := &TaskInfo{
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

			uc.log.Error("脚本执行panic",
				zap.String("error", taskinfo.ErrMSG),
				zap.Any("panic", r),
				zap.String("stack", string(stack)),
				zap.Uint32(ScriptRecordIDKey, record.ID),
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
		if err := uc.UpdateScriptRecord(context.Background(), record.ID, taskinfo); err != nil {
			if taskinfo.LogFile != nil {
				fmt.Fprintf(taskinfo.LogFile, "[%s] [ERROR] 更新脚本记录状态失败: %s\n", time.Now().Format(time.RFC3339), err.Error())
			}
		}

		// 关闭日志文件句柄
		if taskinfo.LogFile != nil {
			taskinfo.LogFile.Close()
		}

		// 清理执行完成的上下文
		uc.DeleteCancel(record.ID)

		// 输出日志
		uc.log.Debug(
			"脚本执行完成",
			zap.Object(ScriptRecordIDKey, record),
		)
	}()

	// 生成日志路径并创建日志目录
	logDir := filepath.Join(config.LogDir, time.Now().Format(time.DateOnly))
	if taskinfo.Error = os.MkdirAll(logDir, 0755); taskinfo.Error != nil {
		taskinfo.Status = 5
		taskinfo.ErrMSG = fmt.Sprintf("创建日志目录失败: %v", taskinfo.Error)
		uc.log.Error(
			"创建日志目录失败",
			zap.Error(taskinfo.Error),
			zap.String("path", logDir),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return taskinfo
	}

	// 创建并打开日志文件
	logPath := record.LogPath()
	taskinfo.LogFile, taskinfo.Error = os.Create(logPath)
	if taskinfo.Error != nil {
		taskinfo.Status = 5
		taskinfo.ErrMSG = fmt.Sprintf("创建日志文件失败: %v", taskinfo.Error)
		uc.log.Error(
			"创建日志文件失败",
			zap.Error(taskinfo.Error),
			zap.String("path", logPath),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return taskinfo
	}

	// 写入开始执行日志
	startTime := time.Now()
	fmt.Fprintf(taskinfo.LogFile, "[%s] 开始执行脚本 (ID: %d, ScriptID: %d)\n",
		startTime.Format(time.RFC3339), record.ID, record.ScriptID)

	// 交验脚本是否存在
	scriptPath := record.Script.ScriptPath()
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

	// 设置环境变量
	cmd.Env = record.InitEnv()

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

func (uc *RecordUsecase) Cancel(ctx context.Context, recordID uint32) {
	cancel := uc.GetCancel(recordID)
	if cancel == nil {
		uc.log.Warn(
			"未找到要取消的脚本任务",
			zap.Uint32(ScriptRecordIDKey, recordID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
	} else {
		uc.log.Info(
			"取消脚本执行",
			zap.Uint32(ScriptRecordIDKey, recordID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		cancel()
		uc.DeleteCancel(recordID)
		uc.log.Info(
			"取消脚本执行成功",
			zap.Uint32(ScriptRecordIDKey, recordID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
	}
}

func (uc *RecordUsecase) CreateScriptRecord(
	ctx context.Context,
	req ExecuteRequest,
) (*ScriptRecordModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	script, err := uc.scriptRepo.FindModel(ctx, req.ScriptID)
	if err != nil {
		uc.log.Error(
			"查询脚本失败",
			zap.Error(err),
			zap.Uint32(ScriptIDKey, req.ScriptID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": req.ScriptID})
	}

	if !script.Status {
		uc.log.Error(
			"脚本已禁用",
			zap.Uint32(ScriptIDKey, req.ScriptID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, ErrScriptDisabled
	}

	now := time.Now()
	record := &ScriptRecordModel{
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

	if err := uc.recordRepo.CreateModel(ctx, record); err != nil {
		uc.log.Error(
			"创建执行记录失败",
			zap.Error(err),
			zap.Uint32(ScriptIDKey, req.ScriptID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	record.Script = *script
	return record, nil
}

func (uc *RecordUsecase) UpdateScriptRecord(
	ctx context.Context,
	recordID uint32,
	taskinfo *TaskInfo,
) *errors.Error {
	uc.log.Debug(
		"开始更新执行记录状态",
	)
	if err := uc.recordRepo.UpdateModel(ctx, taskinfo.ToMap(), "id = ?", recordID); err != nil {
		uc.log.Error(
			"更新脚本记录失败",
			zap.Error(err),
			zap.Uint32(ScriptRecordIDKey, recordID),
			zap.Object("taskinfo", taskinfo),
		)
		return database.NewGormError(err, taskinfo.ToMap())
	}
	return nil
}

func (uc *RecordUsecase) FindScriptRecordByID(
	ctx context.Context,
	preloads []string,
	recordID uint32,
) (*ScriptRecordModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询脚本执行记录",
		zap.Uint32(ScriptRecordIDKey, recordID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := uc.recordRepo.FindModel(ctx, preloads, recordID)
	if err != nil {
		uc.log.Error(
			"查询脚本执行记录失败",
			zap.Error(err),
			zap.Uint32(ScriptRecordIDKey, recordID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": recordID})

	}

	uc.log.Info(
		"查询脚本执行记录成功",
		zap.Uint32(ScriptRecordIDKey, recordID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *RecordUsecase) ListcriptRecord(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]ScriptRecordModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询脚本执行记录列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	count, ms, err := uc.recordRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询脚本执行记录列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询脚本执行记录列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *RecordUsecase) AsyncExecuteScript(
	ctx context.Context,
	req ExecuteRequest,
) (*ScriptRecordModel, *errors.Error) {
	record, err := uc.CreateScriptRecord(ctx, req)
	if err != nil {
		return nil, err
	}
	go uc.Execute(record)
	return record, nil
}

func (uc *RecordUsecase) SyncExecuteScript(
	ctx context.Context,
	req ExecuteRequest,
) (*TaskInfo, *errors.Error) {
	record, err := uc.CreateScriptRecord(ctx, req)
	if err != nil {
		return nil, err
	}
	return uc.Execute(record), nil
}
