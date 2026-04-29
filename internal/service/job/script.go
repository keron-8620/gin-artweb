package job

import (
	"context"
	"path/filepath"
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

type ScriptService struct {
	log        *zap.Logger
	scriptRepo *jobrepo.ScriptRepo
}

func NewScriptService(
	log *zap.Logger,
	scriptRepo *jobrepo.ScriptRepo,
) *ScriptService {
	return &ScriptService{
		log:        log,
		scriptRepo: scriptRepo,
	}
}

func (s *ScriptService) CreateScript(
	ctx context.Context,
	dto jobmodel.ScriptUpsertDTO,
) (*jobmodel.ScriptModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	claims := ctxutil.MustGetJwtClaims(ctx)
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建脚本:开始执行",
		zap.Object("script_upsert_dto", &dto),
	)

	scriptPath := GetScriptStoragePath(dto.Project, dto.Label, dto.Filename, false)
	if err := s.scriptRepo.SaveScriptFile(ctx, dto.File, scriptPath, false); err != nil {
		log.Error(
			"创建脚本:脚本文件创建失败",
			zap.Error(err),
			zap.String("script_path", scriptPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.ErrScriptSaveFailed.WithField("script_path", scriptPath)
	}

	clearScript := func() {
		if err := s.scriptRepo.RemoveScriptFile(ctx, scriptPath); err != nil {
			log.Error(
				"创建脚本:删除缓存的脚本文件失败,请手动清理",
				zap.Error(err),
				zap.String("script_path", scriptPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	m := dto.ToModel(claims.Username)
	if err := s.scriptRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建脚本:创建数据库模型失败",
			zap.Error(err),
			zap.Object("script_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		clearScript()
		return nil, errors.NewGormError(err, nil)
	}

	log.Info(
		"创建脚本:执行成功",
		zap.Uint32("script_id", m.ID),
		zap.String("script_path", scriptPath),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *ScriptService) UpdateScriptByID(
	ctx context.Context,
	scriptID uint32,
	dto jobmodel.ScriptUpsertDTO,
) (*jobmodel.ScriptModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	claims := ctxutil.MustGetJwtClaims(ctx)
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新脚本:开始执行",
		zap.Uint32("script_id", scriptID),
		zap.Object("script_upsert_dto", &dto),
	)

	om, rErr := s.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		log.Error(
			"更新脚本:查询脚本失败",
			zap.Error(rErr),
			zap.Uint32("script_id", scriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if om.IsBuiltin {
		log.Error(
			"更新脚本:内置脚本不能修改",
			zap.Uint32("script_id", scriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.FromReason(errors.ReasonScriptIsBuiltin).WithField("script_id", scriptID)
	}

	oldScriptPath := GetScriptStoragePath(om.Project, om.Label, om.Name, false)
	tmpScriptPath := filepath.Join(config.TmpDir, uuid.NewString())
	if err := s.scriptRepo.MoveScriptFile(ctx, oldScriptPath, tmpScriptPath); err != nil {
		log.Error(
			"更新脚本:备份旧脚本文件失败",
			zap.Error(err),
			zap.String("script_path", oldScriptPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.FromError(err)
	}

	mvOldScript := func() {
		if err := s.scriptRepo.MoveScriptFile(ctx, tmpScriptPath, oldScriptPath); err != nil {
			log.Error(
				"更新脚本:恢复旧脚本文件失败",
				zap.Error(err),
				zap.String("script_path", tmpScriptPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	newScriptPath := GetScriptStoragePath(dto.Project, dto.Label, dto.Filename, false)
	if err := s.scriptRepo.SaveScriptFile(ctx, dto.File, newScriptPath, true); err != nil {
		log.Error(
			"更新脚本:新脚本写入失败",
			zap.Error(err),
			zap.String("script_path", newScriptPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		mvOldScript()
		return nil, errors.ErrScriptSaveFailed.WithField("script_path", newScriptPath)
	}

	removeNewScript := func() {
		if err := s.scriptRepo.RemoveScriptFile(ctx, newScriptPath); err != nil {
			log.Error(
				"更新脚本:删除缓存的脚本文件失败,请手动清理",
				zap.Error(err),
				zap.String("script_path", newScriptPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	updateData := dto.ToUpdateMap(claims.Username)
	if err := s.scriptRepo.UpdateModel(ctx, updateData, "id = ?", scriptID); err != nil {
		log.Error(
			"更新脚本:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		removeNewScript()
		mvOldScript()
		return nil, errors.NewGormError(err, updateData)
	}

	m, rErr := s.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		log.Error(
			"更新脚本:查询更新后的脚本失败",
			zap.Error(rErr),
			zap.Uint32("script_id", scriptID),
		)
		return nil, rErr
	}

	log.Info(
		"更新脚本:执行成功",
		zap.Uint32("script_id", scriptID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ScriptService) DeleteScriptByID(
	ctx context.Context,
	scriptID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("删除脚本:开始执行", zap.Uint32("script_id", scriptID))

	m, rErr := s.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		log.Error(
			"删除脚本:查询脚本失败",
			zap.Uint32("script_id", scriptID),
			zap.Error(rErr),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return rErr
	}

	if m.IsBuiltin {
		log.Error(
			"删除脚本:内置脚本不能删除",
			zap.Uint32("script_id", scriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromReason(errors.ReasonScriptIsBuiltin).WithField("script_id", scriptID)
	}

	scriptPath := GetScriptStoragePath(m.Project, m.Label, m.Name, false)
	tmpScriptPath := filepath.Join(config.TmpDir, uuid.NewString())

	if err := s.scriptRepo.MoveScriptFile(ctx, scriptPath, tmpScriptPath); err != nil {
		log.Error(
			"删除脚本:备份脚本文件失败",
			zap.Error(err),
			zap.String("script_path", scriptPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	// 回滚函数：数据库删除失败时恢复文件
	mvOldScript := func() {
		if err := s.scriptRepo.MoveScriptFile(ctx, tmpScriptPath, scriptPath); err != nil {
			log.Error(
				"删除脚本:恢复脚本文件失败",
				zap.Error(err),
				zap.String("script_path", tmpScriptPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	if err := s.scriptRepo.DeleteModel(ctx, scriptID); err != nil {
		log.Error(
			"删除脚本:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		mvOldScript()
		return errors.NewGormError(err, map[string]any{"id": scriptID})
	}

	log.Info(
		"删除脚本:执行成功",
		zap.Uint32("script_id", scriptID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ScriptService) FindScriptByID(
	ctx context.Context,
	scriptID uint32,
) (*jobmodel.ScriptModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.scriptRepo.GetModel(ctx, scriptID)
	if err != nil {
		log.Error(
			"查询脚本:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": scriptID})
	}

	log.Debug(
		"查询脚本:查询到的数据库模型详情",
		zap.Object("script_model", m),
	)
	return m, nil
}

func (s *ScriptService) ListScript(
	ctx context.Context,
	page, size int,
	dto jobmodel.ListScriptDTO,
) (int64, []jobmodel.ScriptModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询脚本列表:参数详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_script_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:   limit,
		Offset:  offset,
		OrderBy: []string{"id ASC"},
		Query:   dto.ToQueryMap(),
	}

	count, err := s.scriptRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询脚本列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询脚本列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.scriptRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询脚本列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *ScriptService) ListScriptsByIDs(
	ctx context.Context,
	scriptIDs []uint32,
) ([]jobmodel.ScriptModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	if len(scriptIDs) == 0 {
		log.Warn(
			"查询指定的脚本列表:脚本 ID 列表为空",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return []jobmodel.ScriptModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": scriptIDs},
	}

	log.Debug(
		"查询指定的脚本列表:查询数据库模型参数",
		zap.Object("query_params", &qp),
	)

	ms, err := s.scriptRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询指定的脚本列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	return ms, nil
}

func (s *ScriptService) ListProjects(
	ctx context.Context,
	dto jobmodel.ListScriptDTO,
) ([]string, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询项目名称:参数详情",
		zap.Object("list_script_dto", &dto),
	)

	query := dto.ToQueryMap()
	projects, err := s.scriptRepo.ListProjects(ctx, query)
	if err != nil {
		log.Error(
			"查询项目名称:数据库查询失败",
			zap.Error(err),
			zap.Any("query_params", query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	log.Debug(
		"查询项目名称:查询到的项目名称列表",
		zap.Strings("projects", projects),
	)
	return projects, nil
}

func (s *ScriptService) ListLabels(
	ctx context.Context,
	dto jobmodel.ListScriptDTO,
) ([]string, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询标签名称:参数详情",
		zap.Object("list_script_dto", &dto),
	)

	query := dto.ToQueryMap()
	labels, err := s.scriptRepo.ListLabels(ctx, query)
	if err != nil {
		log.Error(
			"查询标签名称:数据库查询失败",
			zap.Error(err),
			zap.Any("query_params", query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	log.Debug(
		"查询标签名称:查询到的标签名称列表",
		zap.Strings("labels", labels),
	)
	return labels, nil
}

func GetScriptStoragePath(project, label, name string, isBuiltin bool) string {
	if isBuiltin {
		return filepath.Join(config.ResourceDir, project, "script", label, name)
	}
	return filepath.Join(config.StorageDir, "script", project, label, name)
}
