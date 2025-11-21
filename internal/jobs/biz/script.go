package biz

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	bizCustomer "gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
)

const ScriptIDKey = "script_id"

type ScriptModel struct {
	database.StandardModel
	Name      string                `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Descr     string                `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	Project   string                `gorm:"column:project;type:varchar(50);comment:项目" json:"project"`
	Label     string                `gorm:"column:label;type:varchar(50);index:idx_script_label;comment:标签" json:"label"`
	Language  string                `gorm:"column:language;type:varchar(50);comment:脚本语言" json:"language"`
	Status    bool                  `gorm:"column:status;type:boolean;comment:是否启用" json:"status"`
	IsBuiltin bool                  `gorm:"column:is_builtin;type:boolean;comment:是否是内置脚本" json:"is_builtin"`
	UserID    uint32                `gorm:"column:user_id;not null;comment:用户ID" json:"user_id"`
	User      bizCustomer.UserModel `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"user"`
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
	enc.AddUint32("user_id", m.UserID)
	return nil
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
		"脚本创建成功",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *ScriptUsecase) UpdateScriptByID(
	ctx context.Context,
	scriptID uint32,
	data map[string]any,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
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
		return database.NewGormError(err, data)
	}

	uc.log.Info(
		"脚本更新成功",
		zap.Uint32(ScriptIDKey, scriptID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *ScriptUsecase) DeleteScriptByID(
	ctx context.Context,
	scriptID uint32,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
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

	uc.log.Info(
		"脚本删除成功",
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
