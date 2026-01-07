package biz

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

const ScriptIDKey = "script_id"

type ScriptModel struct {
	database.StandardModel
	Name      string `gorm:"column:name;type:varchar(50);not null;index:idx_script_project_label_name;comment:名称" json:"name"`
	Descr     string `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	Project   string `gorm:"column:project;type:varchar(50);index:idx_script_project_label_name;comment:项目" json:"project"`
	Label     string `gorm:"column:label;type:varchar(50);index:idx_script_project_label_name;;comment:标签" json:"label"`
	Language  string `gorm:"column:language;type:varchar(50);comment:脚本语言" json:"language"`
	Status    bool   `gorm:"column:status;type:boolean;comment:是否启用" json:"status"`
	IsBuiltin bool   `gorm:"column:is_builtin;type:boolean;comment:是否是内置脚本" json:"is_builtin"`
	Username  string `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
}

func (m *ScriptModel) TableName() string {
	return "jobs_script"
}

func (m *ScriptModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("descr", m.Descr)
	enc.AddString("project", m.Project)
	enc.AddString("label", m.Label)
	enc.AddString("language", m.Language)
	enc.AddBool("status", m.Status)
	enc.AddBool("is_builtin", m.IsBuiltin)
	enc.AddString("username", m.Username)
	return nil
}

func (m *ScriptModel) ScriptPath() string {
	if m.IsBuiltin {
		return filepath.Join(config.ResourceDir, m.Project, "script", m.Label, m.Name)
	}
	return filepath.Join(config.StorageDir, m.Project, "script", m.Label, m.Name)
}

type ScriptRepo interface {
	CreateModel(context.Context, *ScriptModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, ...any) (*ScriptModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]ScriptModel, error)
}

type ScriptUsecase struct {
	log        *zap.Logger
	scriptRepo ScriptRepo
}

func NewScriptUsecase(
	log *zap.Logger,
	scriptRepo ScriptRepo,
) *ScriptUsecase {
	return &ScriptUsecase{
		log:        log,
		scriptRepo: scriptRepo,
	}
}

func (uc *ScriptUsecase) CreateScript(
	ctx context.Context,
	m ScriptModel,
) (*ScriptModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建脚本",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.scriptRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建脚本失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"创建脚本成功",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *ScriptUsecase) UpdateScriptByID(
	ctx context.Context,
	scriptID uint32,
	data map[string]any,
) (*ScriptModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	om, rErr := uc.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		return nil, rErr
	}
	if om.IsBuiltin {
		uc.log.Error(
			"内置脚本不能修改",
			zap.Uint32(ScriptIDKey, scriptID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, ErrScriptIsBuiltin.WithData(map[string]any{ScriptIDKey: scriptID})
	}

	uc.log.Info(
		"开始更新脚本",
		zap.Uint32(ScriptIDKey, scriptID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.scriptRepo.UpdateModel(ctx, data, "id = ?", scriptID); err != nil {
		uc.log.Error(
			"更新脚本失败",
			zap.Error(err),
			zap.Uint32(ScriptIDKey, scriptID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, data)
	}

	uc.log.Info(
		"更新脚本成功",
		zap.Uint32(ScriptIDKey, scriptID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return uc.FindScriptByID(ctx, scriptID)
}

func (uc *ScriptUsecase) DeleteScriptByID(
	ctx context.Context,
	scriptID uint32,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	m, rErr := uc.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		return rErr
	}
	if m.IsBuiltin {
		uc.log.Error(
			"内置脚本不能删除",
			zap.Uint32(ScriptIDKey, scriptID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrScriptIsBuiltin.WithData(map[string]any{ScriptIDKey: scriptID})
	}

	uc.log.Info(
		"开始删除脚本",
		zap.Uint32(ScriptIDKey, scriptID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.scriptRepo.DeleteModel(ctx, scriptID); err != nil {
		uc.log.Error(
			"删除脚本失败",
			zap.Error(err),
			zap.Uint32(ScriptIDKey, scriptID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, map[string]any{"id": scriptID})
	}

	if rErr := uc.RemoveScript(ctx, *m); rErr != nil {
		return rErr
	}

	uc.log.Info(
		"删除脚本成功",
		zap.Uint32(ScriptIDKey, scriptID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *ScriptUsecase) FindScriptByID(
	ctx context.Context,
	scriptID uint32,
) (*ScriptModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询脚本",
		zap.Uint32(ScriptIDKey, scriptID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := uc.scriptRepo.FindModel(ctx, scriptID)
	if err != nil {
		uc.log.Error(
			"查询脚本失败",
			zap.Error(err),
			zap.Uint32(ScriptIDKey, scriptID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": scriptID})
	}

	uc.log.Info(
		"查询脚本成功",
		zap.Uint32(ScriptIDKey, scriptID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *ScriptUsecase) ListScript(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]ScriptModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询脚本列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	count, ms, err := uc.scriptRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询脚本列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询脚本列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *ScriptUsecase) RemoveScript(ctx context.Context, m ScriptModel) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	savePath := m.ScriptPath()

	uc.log.Info(
		"开始删除脚本文件",
		zap.String("path", savePath),
		zap.Uint32(ScriptIDKey, m.ID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	// 检查文件是否存在
	if _, statErr := os.Stat(savePath); os.IsNotExist(statErr) {
		// 文件不存在，视为删除成功
		uc.log.Warn(
			"脚本文件不存在，无需删除",
			zap.String("path", savePath),
			zap.Uint32(ScriptIDKey, m.ID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil
	} else if statErr != nil {
		// 其他 stat 错误
		uc.log.Error(
			"检查脚本文件状态失败",
			zap.Error(statErr),
			zap.String("path", savePath),
			zap.Uint32(ScriptIDKey, m.ID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return errors.FromError(statErr)
	}

	// 执行删除操作
	if rmErr := os.Remove(savePath); rmErr != nil {
		uc.log.Error(
			"删除脚本文件失败",
			zap.Error(rmErr),
			zap.String("path", savePath),
			zap.Uint32(ScriptIDKey, m.ID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return errors.FromError(rmErr)
	}

	uc.log.Info(
		"删除脚本文件成功",
		zap.String("path", savePath),
		zap.Uint32(ScriptIDKey, m.ID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}
