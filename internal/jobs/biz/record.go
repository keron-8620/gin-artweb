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
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	bizCustomer "gin-artweb/internal/customer/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

const ScriptRecordIDKey = "script_record_id"

type ScriptRecordModel struct {
	database.StandardModel
	TriggerType  string                `gorm:"column:trigger_type;type:varchar(20);comment:触发类型(cron/api)" json:"trigger_type"`
	Status       int                   `gorm:"column:status;type:tinyint;not null;default:0;comment:执行状态(0-待执行,1-执行中,2-成功,3-失败,4-超时,5-崩溃)" json:"status"`
	ExitCode     int                   `gorm:"column:exit_code;comment:退出码" json:"exit_code"`
	EnvVars      string                `gorm:"column:env_vars;type:json;comment:环境变量(JSON对象)" json:"env_vars"`
	CommandArgs  string                `gorm:"column:command_args;type:varchar(254);comment:命令行参数(JSON数组)" json:"command_args"`
	WorkDir      string                `gorm:"column:work_dir;type:varchar(255);comment:工作目录" json:"work_dir"`
	Timeout      int                   `gorm:"column:timeout;type:int;not null;default:300;comment:超时时间(秒)" json:"timeout"`
	LogName      string                `gorm:"column:log_name;type:varchar(255);comment:日志文件路径" json:"log_name"`
	ErrorMessage string                `gorm:"column:error_message;type:text;comment:错误信息" json:"error_message"`
	ScriptID     uint32                `gorm:"column:script_id;not null;index;comment:脚本ID" json:"script_id"`
	Script       ScriptModel           `gorm:"foreignKey:ScriptID;references:ID" json:"script"`
	UserID       uint32                `gorm:"column:user_id;not null;comment:执行用户ID" json:"user_id"`
	User         bizCustomer.UserModel `gorm:"foreignKey:UserID;references:ID" json:"user"`
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
	enc.AddUint32("script_id", m.ScriptID)
	enc.AddUint32("user_id", m.UserID)
	return nil
}

func (m *ScriptRecordModel) InitEnv() []string {
	env := os.Environ()
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
	FindModel(context.Context, ...any) (*ScriptRecordModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]ScriptRecordModel, error)
}

type ExecuteRequest struct {
	TriggerType string `json:"trigger_type"`
	ScriptID    uint32 `json:"script_id"`
	CommandArgs string `json:"command_args"`
	EnvVars     string `json:"env_vars"`
	Timeout     int    `json:"timeout"`
	WorkDir     string `json:"work_dir"`
	UserID      uint32 `json:"user_id"`
}

type ExecuteResult struct {
	RecordID  uint32 `json:"record_id"`
	ExitCode  int    `json:"exit_code"`
	Output    string `json:"output"`
	Error     string `json:"error"`
	Duration  int64  `json:"duration"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	LogPath   string `json:"log_path"`
}

type RecordUsecase struct {
	log        *zap.Logger
	scriptRepo ScriptRepo
	recordRepo ScriptRecordRepo
	contexts   sync.Map
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
	}
}

func (uc *RecordUsecase) UpdateScriptRecordByID(
	ctx context.Context,
	recordID uint32,
	data map[string]any,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始更新脚本执行记录",
		zap.Uint32(ScriptRecordIDKey, recordID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.recordRepo.UpdateModel(ctx, data, "id = ?", recordID); err != nil {
		uc.log.Error(
			"更新脚本执行记录失败",
			zap.Error(err),
			zap.Uint32(ScriptRecordIDKey, recordID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, data)
	}

	uc.log.Info(
		"更新脚本执行记录成功",
		zap.Uint32(ScriptRecordIDKey, recordID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *RecordUsecase) FindScriptRecordByID(
	ctx context.Context,
	recordID uint32,
) (*ScriptRecordModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询脚本执行记录",
		zap.Uint32(ScriptRecordIDKey, recordID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := uc.recordRepo.FindModel(ctx, recordID)
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
	if err := errors.CheckContext(ctx); err != nil {
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

func (uc *RecordUsecase) ExecuteScript(
	ctx context.Context,
	req ExecuteRequest,
) (*ScriptRecordModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	// 1. 获取脚本信息
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

	// 2. 创建执行记录
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
		ScriptID:     req.ScriptID,
		UserID:       req.UserID,
	}

	// 3. 保存初始记录
	if err := uc.recordRepo.CreateModel(ctx, record); err != nil {
		uc.log.Error(
			"创建执行记录失败",
			zap.Error(err),
			zap.Uint32(ScriptIDKey, req.ScriptID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	// 4. 创建上下文并存储
	execCtx, cancel := context.WithCancel(context.Background())
	uc.contexts.Store(record.ID, cancel)

	// 5. 异步执行脚本
	record.Script = *script
	go uc.executeScriptAsync(execCtx, record)

	// 6. 立即返回，不等待执行完成
	return record, nil
}

func (uc *RecordUsecase) ExecuteSchedule(
	ctx context.Context,
	m *ScheduleModel,
) (*ScriptRecordModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	// 构造请求结构体
	excReq := ExecuteRequest{
		TriggerType: "cron",
		ScriptID:    m.ScriptID,
		CommandArgs: m.CommandArgs,
		EnvVars:     m.EnvVars,
		Timeout:     m.Timeout,
		WorkDir:     m.WorkDir,
		UserID:      m.UserID,
	}

	// 执行计划任务的脚本
	return uc.ExecuteScript(ctx, excReq)
}

func (uc *RecordUsecase) executeScriptAsync(
	ctx context.Context,
	record *ScriptRecordModel,
) {
	// 处理执行结果
	var (
		exitCode int    = -1
		status   int    = 3
		errMsg   string = ""
		err      error
		logFile  *os.File
	)

	// panic 恢复保护
	defer func() {
		if r := recover(); r != nil {
			// 记录panic信息和堆栈跟踪
			stack := debug.Stack()

			// 构造错误响应
			var errMsg string
			switch v := r.(type) {
			case error:
				errMsg = v.Error()
			case string:
				errMsg = v
			default:
				errMsg = fmt.Sprintf("%v", v)
			}

			uc.log.Error("脚本执行panic",
				zap.String("error", errMsg),
				zap.Any("panic", r),
				zap.String("stack", string(stack)),
				zap.Uint32(ScriptRecordIDKey, record.ID),
			)

			if logFile != nil {
				format := time.Now().Format(time.RFC3339)
				fmt.Fprintf(logFile, "[%s] [PANIC] 脚本执行发生严重错误: %s\n", format, errMsg)
				fmt.Fprintf(logFile, "[%s] [STACK] %s\n", format, stack)
			}

			// 设置脚本状态为崩溃
			status = 5
		}
		// 更新记录状态为失败
		uc.UpdateScriptRecordByID(
			context.Background(),
			record.ID,
			map[string]any{
				"status":        status,
				"exit_code":     exitCode,
				"error_message": errMsg,
			},
		)

		// 关闭日志文件句柄
		if logFile != nil {
			logFile.Close()
		}

		// 清理执行完成的上下文
		uc.contexts.Delete(record.ID)
	}()

	// 3. 生成日志路径并创建日志目录
	logDir := filepath.Join(config.LogDir, time.Now().Format(time.DateOnly))
	if err = os.MkdirAll(logDir, 0755); err != nil {
		status = 5
		errMsg = fmt.Sprintf("创建日志目录失败: %v", err)
		uc.log.Error(
			"创建日志目录失败",
			zap.Error(err),
			zap.String("path", logDir),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return
	}

	// 5. 创建并打开日志文件
	logPath := record.LogPath()
	logFile, err = os.Create(logPath)
	if err != nil {
		status = 5
		errMsg = fmt.Sprintf("创建日志文件失败: %v", err)
		uc.log.Error(
			"创建日志文件失败",
			zap.Error(err),
			zap.String("path", logPath),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return
	}
	defer logFile.Close()

	// 写入开始执行日志
	startTime := time.Now()
	fmt.Fprintf(logFile, "[%s] 开始执行脚本 (ID: %d, ScriptID: %d)\n",
		startTime.Format(time.RFC3339), record.ID, record.ScriptID)

	// 创建带超时的上下文
	timeout := time.Duration(record.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Minute // 默认超时时间
	}
	ctxExe, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 解析命令参数
	var cmdArgs []string
	if record.CommandArgs != "" {
		cmdArgs = strings.Fields(record.CommandArgs)
	}

	// 创建命令
	scriptPath := record.Script.ScriptPath()
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Fprintf(logFile, "脚本文件不存在: %s\n", scriptPath)
		return
	}
	cmd := exec.CommandContext(ctxExe, scriptPath, cmdArgs...)

	// 设置工作目录
	if record.WorkDir != "" {
		if _, err := os.Stat(record.WorkDir); os.IsNotExist(err) {
			fmt.Fprintf(logFile, "工作目录不存在，尝试创建: %s\n", record.WorkDir)
			if err := os.MkdirAll(record.WorkDir, 0755); err != nil {
				fmt.Fprintf(logFile, "创建工作目录失败: %s\n", err)
				return
			}
		}
		cmd.Dir = record.WorkDir
	}

	// 设置环境变量
	if record.EnvVars != "" {
		cmd.Env = record.InitEnv()
	}

	// 重定向输出到日志文件
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// 执行命令
	fmt.Fprintf(logFile, "执行命令: %s %s\n", scriptPath, strings.Join(cmdArgs, " "))
	err = cmd.Run()
	endTime := time.Now()

	// 计算执行时长
	duration := endTime.Sub(startTime).Milliseconds()

	select {
	case <-ctxExe.Done():
		// 检查是否是超时或手动取消
		if ctxExe.Err() == context.DeadlineExceeded {
			exitCode = 124 // 标准超时退出码
			status = 4     // 超时状态
			fmt.Fprintf(logFile, "[%s] 脚本执行超时 (耗时: %dms)\n",
				endTime.Format(time.RFC3339), duration)
		} else {
			exitCode = -1 // 手动取消
			status = 3    // 失败状态
			fmt.Fprintf(logFile, "[%s] 脚本执行被取消 (耗时: %dms)\n",
				endTime.Format(time.RFC3339), duration)
		}
	default:
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = -1
			}
			status = 3 // 失败状态
			fmt.Fprintf(logFile, "[%s] 脚本执行失败 (退出码: %d, 耗时: %dms): %s\n",
				endTime.Format(time.RFC3339), exitCode, duration, err)
		} else {
			exitCode = 0
			status = 2 // 成功状态
			fmt.Fprintf(logFile, "[%s] 脚本执行成功 (耗时: %dms)\n",
				endTime.Format(time.RFC3339), duration)
		}
	}
}

func (uc *RecordUsecase) Cancel(ctx context.Context, recordID uint32) {
	if cancel, ok := uc.contexts.Load(recordID); ok {
		uc.log.Info(
			"取消脚本执行",
			zap.Uint32(ScriptRecordIDKey, recordID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		cancel.(context.CancelFunc)()
		uc.contexts.Delete(recordID)
		uc.log.Info(
			"取消脚本执行成功",
			zap.Uint32(ScriptRecordIDKey, recordID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
	}
}
