package biz

import (
	"context"
	"encoding/json"

	"go.uber.org/zap/zapcore"

	bizCustomer "gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/database"
)

type ScriptRecordModel struct {
	database.StandardModel
	TriggerType string                `gorm:"column:trigger_type;type:varchar(20);comment:触发类型(cron/api)" json:"trigger_type"`
	Status      int                   `gorm:"column:status;type:tinyint;not null;default:0;comment:执行状态(0-待执行,1-执行中,2-成功,3-失败,4-超时)" json:"status"`
	ExitCode    int                   `gorm:"column:exit_code;comment:退出码" json:"exit_code"`
	EnvVars     string                `gorm:"column:env_vars;type:json;comment:环境变量(JSON对象)" json:"env_vars"`
	CommandArgs string                `gorm:"column:command_args;type:varchar(254);comment:命令行参数(JSON数组)" json:"command_args"`
	WorkDir     string                `gorm:"column:work_dir;type:varchar(255);comment:工作目录" json:"work_dir"`
	Timeout     int                   `gorm:"column:timeout;type:int;not null;default:300;comment:超时时间(秒)" json:"timeout"`
	Command     string                `gorm:"column:command;type:text;comment:命令" json:"command"`
	LogPath     string                `gorm:"column:log_path;type:varchar(255);comment:日志文件路径" json:"log_path"`
	ScriptID    uint32                `gorm:"column:script_id;not null;index;comment:脚本ID" json:"script_id"`
	Script      ScriptModel           `gorm:"foreignKey:ScriptID;references:ID" json:"script"`
	UserID      uint32                `gorm:"column:user_id;not null;comment:执行用户ID" json:"user_id"`
	User        bizCustomer.UserModel `gorm:"foreignKey:UserID;references:ID" json:"user"`
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
	enc.AddString("command", m.Command)
	enc.AddString("log_path", m.LogPath)
	enc.AddUint32("script_id", m.ScriptID)
	enc.AddUint32("user_id", m.UserID)
	return nil
}

func (m *ScriptRecordModel) GetEnvVars() (map[string]string, error) {
	if m.EnvVars == "" {
		return make(map[string]string), nil
	}

	var envMap map[string]string
	err := json.Unmarshal([]byte(m.EnvVars), &envMap)
	return envMap, err
}

type RecordRepo interface {
	CreateModel(context.Context, *ScriptRecordModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, ...any) (*ScriptRecordModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]ScriptRecordModel, error)
}

// type ExecuteRequest struct {
// 	TriggerType string            `json:"trigger_type"` // cron/manual/api
// 	ScriptID    uint32            `json:"script_id"`
// 	CommandArgs []string          `json:"command_args"`
// 	EnvVars     map[string]string `json:"env_vars"`
// 	Timeout     int               `json:"timeout"`  // 超时时间(秒)，默认300秒
// 	WorkDir     string            `json:"work_dir"` // 工作目录，默认为临时目录
// 	UserID      uint32            `json:"user_id"`
// }

// type ExecuteResult struct {
// 	RecordID  uint32 `json:"record_id"`
// 	ExitCode  int    `json:"exit_code"`
// 	Output    string `json:"output"`
// 	Error     string `json:"error"`
// 	Duration  int64  `json:"duration"`
// 	StartTime int64  `json:"start_time"`
// 	EndTime   int64  `json:"end_time"`
// 	LogPath   string `json:"log_path"`
// }

// type RecordUsecase struct {
// 	log        *zap.Logger
// 	scriptRepo ScriptRepo
// 	recordRepo RecordRepo
// }

// func NewScriptRecordUsecase(
// 	log *zap.Logger,
// 	scriptRepo ScriptRepo,
// 	recordRepo RecordRepo,
// ) *RecordUsecase {
// 	return &RecordUsecase{
// 		log:        log,
// 		scriptRepo: scriptRepo,
// 		recordRepo: recordRepo,
// 	}
// }

// func (uc *RecordUsecase) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, *errors.Error) {
// 	if err := errors.CheckContext(ctx); err != nil {
// 		return nil, errors.FromError(err)
// 	}

// 	// 1. 获取脚本信息
// 	script, err := uc.scriptRepo.FindModel(ctx, req.ScriptID)
// 	if err != nil {
// 		uc.log.Error(
// 			"查询脚本失败",
// 			zap.Error(err),
// 			zap.Uint32(ScriptIDKey, req.ScriptID),
// 			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
// 		)
// 		return nil, database.NewGormError(err, map[string]any{"id": req.ScriptID})
// 	}

// 	if !script.Status {
// 		return nil, ErrScriptDisabled.WithData(map[string]any{"id": req.ScriptID})
// 	}

// 	// 2. 创建执行记录
// 	record := &ScriptRecordModel{
// 		ScriptID:    req.ScriptID,
// 		UserID:      req.UserID,
// 		TriggerType: req.TriggerType,
// 		Status:      1,
// 	}

// 	// 序列化参数
// 	if argsBytes, err := json.Marshal(req.CommandArgs); err == nil {
// 		record.CommandArgs = string(argsBytes)
// 	}

// 	if envBytes, err := json.Marshal(req.EnvVars); err == nil {
// 		record.EnvVars = string(envBytes)
// 	}

// 	// 保存初始记录
// 	if err := uc.db.Create(record).Error; err != nil {
// 		return nil, fmt.Errorf("创建执行记录失败: %w", err)
// 	}

// 	// 3. 执行脚本
// 	result := &ExecuteResult{
// 		RecordID:  record.ID,
// 		StartTime: time.Now().Unix(),
// 	}

// 	// 更新记录开始时间
// 	uc.db.Model(record).Updates(map[string]interface{}{
// 		"start_time": result.StartTime,
// 	})

// 	defer func() {
// 		// 更新执行结果
// 		result.EndTime = time.Now().Unix()
// 		result.Duration = (result.EndTime - result.StartTime) * 1000 // 毫秒

// 		updates := map[string]interface{}{
// 			"end_time":  result.EndTime,
// 			"duration":  result.Duration,
// 			"exit_code": result.ExitCode,
// 			"status":    getStatusFromExitCode(result.ExitCode),
// 		}

// 		if result.LogPath != "" {
// 			updates["log_path"] = result.LogPath
// 		}

// 		uc.db.Model(record).Updates(updates)
// 	}()

// 	// 4. 准备执行环境
// 	workDir, logFile, err := uc.prepareLocalExecutionEnvironment(record.ID, req.WorkDir)
// 	if err != nil {
// 		result.ExitCode = -1
// 		result.Error = fmt.Sprintf("准备执行环境失败: %v", err)
// 		return result, err
// 	}
// 	defer os.RemoveAll(workDir) // 清理工作目录

// 	result.LogPath = logFile.Name()

// 	// 5. 构建命令
// 	scriptFile, err := uc.createLocalScriptFile(script, workDir)
// 	if err != nil {
// 		result.ExitCode = -1
// 		result.Error = fmt.Sprintf("创建脚本文件失败: %v", err)
// 		return result, err
// 	}

// 	// 6. 构建执行命令
// 	cmdArgs := append([]string{scriptFile}, req.CommandArgs...)
// 	var cmd *exec.Cmd

// 	switch strings.ToLower(script.Language) {
// 	case "bash", "sh", "":
// 		cmd = exec.Command("/bin/bash", cmdArgs...)
// 	case "python":
// 		cmd = exec.Command("python3", cmdArgs...)
// 	case "node", "javascript":
// 		cmd = exec.Command("node", cmdArgs...)
// 	default:
// 		// 默认直接执行脚本文件
// 		cmd = exec.Command(scriptFile, req.CommandArgs...)
// 	}

// 	// 7. 设置工作目录
// 	cmd.Dir = workDir

// 	// 8. 设置环境变量
// 	env := os.Environ()
// 	for k, v := range req.EnvVars {
// 		env = append(env, fmt.Sprintf("%s=%s", k, v))
// 	}
// 	cmd.Env = env

// 	// 9. 设置超时
// 	timeout := time.Duration(req.Timeout) * time.Second
// 	if timeout <= 0 {
// 		timeout = 5 * time.Minute // 默认5分钟
// 	}
// 	ctx, cancel := context.WithTimeout(ctx, timeout)
// 	defer cancel()
// 	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)

// 	// 10. 重定向输出到日志文件
// 	cmd.Stdout = logFile
// 	cmd.Stderr = logFile

// 	// 11. 执行命令
// 	start := time.Now()
// 	err = cmd.Run()
// 	executionTime := time.Since(start)

// 	// 12. 处理执行结果
// 	if ctx.Err() == context.DeadlineExceeded {
// 		result.ExitCode = 124 // 标准超时退出码
// 		result.Error = "脚本执行超时"
// 	} else if err != nil {
// 		if exitError, ok := err.(*exec.ExitError); ok {
// 			result.ExitCode = exitError.ExitCode()
// 		} else {
// 			result.ExitCode = -1
// 		}
// 		result.Error = err.Error()
// 	} else {
// 		result.ExitCode = 0
// 	}

// 	// 13. 读取部分日志作为输出摘要
// 	outputSummary, err := uc.readLogSummary(logFile)
// 	if err == nil {
// 		result.Output = outputSummary
// 	}

// 	result.Duration = executionTime.Milliseconds()
// 	return result, nil
// }

// // prepareLocalExecutionEnvironment 准备本地执行环境
// func (se *ScriptExecutor) prepareLocalExecutionEnvironment(recordID uint32, workDir string) (string, *os.File, error) {
// 	// 确定工作目录
// 	if workDir == "" {
// 		workDir = filepath.Join(os.TempDir(), fmt.Sprintf("script_exec_%d_%d", recordID, time.Now().Unix()))
// 	}

// 	if err := os.MkdirAll(workDir, 0755); err != nil {
// 		return "", nil, fmt.Errorf("创建工作目录失败: %w", err)
// 	}

// 	// 创建日志文件
// 	logFileName := filepath.Join(workDir, "execution.log")
// 	logFile, err := os.Create(logFileName)
// 	if err != nil {
// 		os.RemoveAll(workDir)
// 		return "", nil, fmt.Errorf("创建日志文件失败: %w", err)
// 	}

// 	return workDir, logFile, nil
// }

// // createLocalScriptFile 创建本地脚本文件
// func (se *ScriptExecutor) createLocalScriptFile(script ScriptModel, workDir string) (string, error) {
// 	// 根据脚本语言确定文件扩展名
// 	scriptFileName := ""
// 	switch strings.ToLower(script.Language) {
// 	case "bash", "sh":
// 		scriptFileName = filepath.Join(workDir, "script.sh")
// 	case "python":
// 		scriptFileName = filepath.Join(workDir, "script.py")
// 	case "node", "javascript":
// 		scriptFileName = filepath.Join(workDir, "script.js")
// 	case "perl":
// 		scriptFileName = filepath.Join(workDir, "script.pl")
// 	default:
// 		scriptFileName = filepath.Join(workDir, "script")
// 	}

// 	// 写入脚本内容
// 	if err := os.WriteFile(scriptFileName, []byte(script.Content), 0755); err != nil {
// 		return "", fmt.Errorf("写入脚本文件失败: %w", err)
// 	}

// 	return scriptFileName, nil
// }

// // readLogSummary 读取日志摘要
// func (se *ScriptExecutor) readLogSummary(logFile *os.File) (string, error) {
// 	// 回到文件开始位置
// 	if _, err := logFile.Seek(0, 0); err != nil {
// 		return "", err
// 	}

// 	// 读取前4KB作为摘要
// 	buf := make([]byte, 4096)
// 	n, err := logFile.Read(buf)
// 	if err != nil && err != os.EOF {
// 		return "", err
// 	}

// 	return string(buf[:n]), nil
// }

// // getStatusFromExitCode 根据退出码确定状态
// func getStatusFromExitCode(exitCode int) int {
// 	switch {
// 	case exitCode == 0:
// 		return 2 // 成功
// 	case exitCode == 124:
// 		return 4 // 超时
// 	default:
// 		return 3 // 失败
// 	}
// }

// // GetScriptExecutionRecords 获取脚本执行记录
// func (se *ScriptExecutor) GetScriptExecutionRecords(ctx context.Context, scriptID uint32, page, pageSize int) ([]*ScriptRecordModel, int64, error) {
// 	var records []*ScriptRecordModel
// 	var total int64

// 	query := se.db.Model(&ScriptRecordModel{}).Where("script_id = ?", scriptID)

// 	// 获取总数
// 	if err := query.Count(&total).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	// 获取分页数据
// 	offset := (page - 1) * pageSize
// 	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	return records, total, nil
// }

// // GetScriptExecutionRecord 获取单条执行记录详情
// func (se *ScriptExecutor) GetScriptExecutionRecord(ctx context.Context, recordID uint32) (*ScriptRecordModel, error) {
// 	var record ScriptRecordModel
// 	if err := se.db.Preload("Script").Preload("User").First(&record, recordID).Error; err != nil {
// 		return nil, err
// 	}
// 	return &record, nil
// }
