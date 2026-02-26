package jobs

import (
	"context"
	"os"

	"go.uber.org/zap"

	jobsmodel "gin-artweb/internal/model/jobs"
	jobsrepo "gin-artweb/internal/repository/jobs"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type ScriptService struct {
	log        *zap.Logger
	scriptRepo *jobsrepo.ScriptRepo
}

func NewScriptService(
	log *zap.Logger,
	scriptRepo *jobsrepo.ScriptRepo,
) *ScriptService {
	return &ScriptService{
		log:        log,
		scriptRepo: scriptRepo,
	}
}

func (s *ScriptService) CreateScript(
	ctx context.Context,
	m jobsmodel.ScriptModel,
) (*jobsmodel.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建脚本",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.scriptRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建脚本失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"创建脚本成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (s *ScriptService) UpdateScriptByID(
	ctx context.Context,
	scriptID uint32,
	data map[string]any,
) (*jobsmodel.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	om, rErr := s.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		return nil, rErr
	}
	if om.IsBuiltin {
		s.log.Error(
			"内置脚本不能修改",
			zap.Uint32("script_id", scriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromReason(errors.ReasonScriptIsBuiltin).WithField("script_id", scriptID)
	}

	s.log.Info(
		"开始更新脚本",
		zap.Uint32("script_id", scriptID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.scriptRepo.UpdateModel(ctx, data, "id = ?", scriptID); err != nil {
		s.log.Error(
			"更新脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	s.log.Info(
		"更新脚本成功",
		zap.Uint32("script_id", scriptID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return s.FindScriptByID(ctx, scriptID)
}

func (s *ScriptService) DeleteScriptByID(
	ctx context.Context,
	scriptID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	m, rErr := s.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		return rErr
	}
	if m.IsBuiltin {
		s.log.Error(
			"内置脚本不能删除",
			zap.Uint32("script_id", scriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromReason(errors.ReasonScriptIsBuiltin).WithField("script_id", scriptID)
	}

	s.log.Info(
		"开始删除脚本",
		zap.Uint32("script_id", scriptID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.scriptRepo.DeleteModel(ctx, scriptID); err != nil {
		s.log.Error(
			"删除脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": scriptID})
	}

	if rErr := s.RemoveScript(ctx, *m); rErr != nil {
		return rErr
	}

	s.log.Info(
		"删除脚本成功",
		zap.Uint32("script_id", scriptID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *ScriptService) FindScriptByID(
	ctx context.Context,
	scriptID uint32,
) (*jobsmodel.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询脚本",
		zap.Uint32("script_id", scriptID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.scriptRepo.GetModel(ctx, scriptID)
	if err != nil {
		s.log.Error(
			"查询脚本失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": scriptID})
	}

	s.log.Info(
		"查询脚本成功",
		zap.Uint32("script_id", scriptID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *ScriptService) ListScript(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]jobsmodel.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询脚本列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.scriptRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询脚本列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询脚本列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *ScriptService) RemoveScript(ctx context.Context, m jobsmodel.ScriptModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	savePath := common.GetScriptStoragePath(m.Project, m.Label, m.Name, m.IsBuiltin)

	s.log.Info(
		"开始删除脚本文件",
		zap.String("path", savePath),
		zap.Uint32("script_id", m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查文件是否存在
	if _, statErr := os.Stat(savePath); os.IsNotExist(statErr) {
		// 文件不存在，视为删除成功
		s.log.Warn(
			"脚本文件不存在，无需删除",
			zap.String("path", savePath),
			zap.Uint32("script_id", m.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil
	} else if statErr != nil {
		// 其他 stat 错误
		s.log.Error(
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
		s.log.Error(
			"删除脚本文件失败",
			zap.Error(rmErr),
			zap.String("path", savePath),
			zap.Uint32("script_id", m.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(rmErr)
	}

	s.log.Info(
		"删除脚本文件成功",
		zap.String("path", savePath),
		zap.Uint32("script_id", m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *ScriptService) ListProjects(ctx context.Context, query map[string]any) ([]string, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	projects, err := s.scriptRepo.ListProjects(ctx, query)
	if err != nil {
		s.log.Error(
			"查询项目名称失败",
			zap.Error(err),
			zap.Any(database.QueryParamsKey, query),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	return projects, nil
}

func (s *ScriptService) ListLabels(ctx context.Context, query map[string]any) ([]string, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	labels, err := s.scriptRepo.ListLabels(ctx, query)
	if err != nil {
		s.log.Error(
			"查询标签名称失败",
			zap.Error(err),
			zap.Any(database.QueryParamsKey, query),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	return labels, nil
}
