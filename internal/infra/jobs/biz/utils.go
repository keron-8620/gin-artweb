package biz

import (
	"os"
	"path/filepath"
	"time"

	"gin-artweb/internal/infra/jobs/model"
	"gin-artweb/internal/shared/config"

	"go.uber.org/zap/zapcore"
)

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

func GetScriptPath(m model.ScriptModel) string {
	if m.IsBuiltin {
		return filepath.Join(config.ResourceDir, m.Project, "script", m.Label, m.Name)
	}
	return filepath.Join(config.StorageDir, "script", m.Project, m.Label, m.Name)
}

func GetScriptLogPath(m model.ScriptRecordModel) string {
	return filepath.Join(config.StorageDir, "logs", m.CreatedAt.Format(time.DateOnly), m.LogName)
}
