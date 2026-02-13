package biz

import (
	"context"
	"os"

	"go.uber.org/zap"

	"gin-artweb/internal/infra/jobs/data"
	"gin-artweb/internal/infra/jobs/model"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type ScriptUsecase struct {
	log        *zap.Logger
	scriptRepo *data.ScriptRepo
}

func NewScriptUsecase(
	log *zap.Logger,
	scriptRepo *data.ScriptRepo,
) *ScriptUsecase {
	return &ScriptUsecase{
		log:        log,
		scriptRepo: scriptRepo,
	}
}

func (uc *ScriptUsecase) CreateScript(
	ctx context.Context,
	m model.ScriptModel,
) (*model.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始创建脚本",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.scriptRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建脚本失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"创建脚本成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *ScriptUsecase) UpdateScriptByID(
	ctx context.Context,
	scriptID uint32,
	data map[string]any,
) (*model.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	om, rErr := uc.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		return nil, rErr
	}
	if om.IsBuiltin {
		uc.log.Error(
			"内置脚本不能修改",
			zap.Uint32("script_id", scriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromReason(errors.ReasonScriptIsBuiltin).WithField("script_id", scriptID)
	}

	uc.log.Info(
		"开始更新脚本",
		zap.Uint32("script_id", scriptID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.scriptRepo.UpdateModel(ctx, data, "id = ?", scriptID); err != nil {
		uc.log.Error(
			"更新脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	uc.log.Info(
		"更新脚本成功",
		zap.Uint32("script_id", scriptID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return uc.FindScriptByID(ctx, scriptID)
}

func (uc *ScriptUsecase) DeleteScriptByID(
	ctx context.Context,
	scriptID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	m, rErr := uc.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		return rErr
	}
	if m.IsBuiltin {
		uc.log.Error(
			"内置脚本不能删除",
			zap.Uint32("script_id", scriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromReason(errors.ReasonScriptIsBuiltin).WithField("script_id", scriptID)
	}

	uc.log.Info(
		"开始删除脚本",
		zap.Uint32("script_id", scriptID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.scriptRepo.DeleteModel(ctx, scriptID); err != nil {
		uc.log.Error(
			"删除脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": scriptID})
	}

	if rErr := uc.RemoveScript(ctx, *m); rErr != nil {
		return rErr
	}

	uc.log.Info(
		"删除脚本成功",
		zap.Uint32("script_id", scriptID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *ScriptUsecase) FindScriptByID(
	ctx context.Context,
	scriptID uint32,
) (*model.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询脚本",
		zap.Uint32("script_id", scriptID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.scriptRepo.GetModel(ctx, scriptID)
	if err != nil {
		uc.log.Error(
			"查询脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": scriptID})
	}

	uc.log.Info(
		"查询脚本成功",
		zap.Uint32("script_id", scriptID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *ScriptUsecase) ListScript(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询脚本列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.scriptRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询脚本列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询脚本列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *ScriptUsecase) RemoveScript(ctx context.Context, m model.ScriptModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	savePath := GetScriptPath(m)

	uc.log.Info(
		"开始删除脚本文件",
		zap.String("path", savePath),
		zap.Uint32("script_id", m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查文件是否存在
	if _, statErr := os.Stat(savePath); os.IsNotExist(statErr) {
		// 文件不存在，视为删除成功
		uc.log.Warn(
			"脚本文件不存在，无需删除",
			zap.String("path", savePath),
			zap.Uint32("script_id", m.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil
	} else if statErr != nil {
		// 其他 stat 错误
		uc.log.Error(
			"检查脚本文件状态失败",
			zap.Error(statErr),
			zap.String("path", savePath),
			zap.Uint32("script_id", m.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(statErr)
	}

	// 执行删除操作
	if rmErr := os.Remove(savePath); rmErr != nil {
		uc.log.Error(
			"删除脚本文件失败",
			zap.Error(rmErr),
			zap.String("path", savePath),
			zap.Uint32("script_id", m.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(rmErr)
	}

	uc.log.Info(
		"删除脚本文件成功",
		zap.String("path", savePath),
		zap.Uint32("script_id", m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *ScriptUsecase) ListProjects(ctx context.Context, query map[string]any) ([]string, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	projects, err := uc.scriptRepo.ListProjects(ctx, query)
	if err != nil {
		uc.log.Error(
			"查询项目名称失败",
			zap.Error(err),
			zap.Any(database.QueryParamsKey, query),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	return projects, nil
}

func (uc *ScriptUsecase) ListLabels(ctx context.Context, query map[string]any) ([]string, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	labels, err := uc.scriptRepo.ListLabels(ctx, query)
	if err != nil {
		uc.log.Error(
			"查询标签名称失败",
			zap.Error(err),
			zap.Any(database.QueryParamsKey, query),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	return labels, nil
}
