package job

import (
	"context"
	"path/filepath"
	"time"

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

// func (s *ScriptService) GenerateScriptFilePath(
// 	ctx context.Context,
// 	project, label, name string,
// 	isBuiltin bool,
// ) string {
// 	if isBuiltin {
// 		return filepath.Join(config.ResourceDir, project, "script", label, name)
// 	}
// 	return filepath.Join(config.StorageDir, "script", project, label, name)
// }

func (s *ScriptService) CreateScript(
	ctx context.Context,
	dto jobmodel.UploadScriptBiz,
) (*jobmodel.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建脚本：开始执行")

	log.Debug(
		"创建脚本：输入参数",
		zap.Object("upload_script_biz", &dto),
	)

	claims := ctxutil.MustGetJwtClaims(ctx)
	m := jobmodel.ScriptModel{
		Name:      dto.Filename,
		Descr:     dto.Descr,
		Project:   dto.Project,
		Label:     dto.Label,
		Language:  dto.Language,
		Status:    dto.Status,
		IsBuiltin: false,
		Username:  claims.Username,
	}

	createStepStart := time.Now()
	log.Debug(
		"创建脚本：开始创建数据库模型",
		zap.Object("script_model", &m),
	)
	if err := s.scriptRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建脚本：创建数据库模型失败",
			zap.Error(err),
			zap.Object("script_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建脚本：创建数据库模型成功",
		zap.Object("script_model", &m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	saveStepStart := time.Now()
	log.Debug("创建脚本：开始保存脚本文件")
	scriptPath := GetScriptStoragePath(dto.Project, dto.Label, dto.Filename, false)
	if err := s.scriptRepo.SaveScriptFile(ctx, dto.File, scriptPath, false); err != nil {
		log.Error(
			"创建脚本：脚本文件创建失败",
			zap.Error(err),
			zap.String("script_path", scriptPath),
			zap.Duration("save_step_duration", time.Since(saveStepStart)),
		)
		return nil, errors.ErrScriptSaveFailed.WithField("script_path", scriptPath)
	}
	saveStepDuration := time.Since(saveStepStart)
	log.Debug(
		"创建脚本：脚本文件创建成功",
		zap.String("script_path", scriptPath),
		zap.Duration("save_step_duration", saveStepDuration),
	)

	log.Info(
		"创建脚本：执行成功",
		zap.Uint32("script_id", m.ID),
		zap.String("script_path", scriptPath),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("save_step_duration", saveStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *ScriptService) UpdateScriptByID(
	ctx context.Context,
	scriptID uint32,
	dto jobmodel.UploadScriptBiz,
) (*jobmodel.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	claims := ctxutil.MustGetJwtClaims(ctx)

	log.Info(
		"更新脚本：开始执行",
		zap.Uint32("script_id", scriptID),
	)

	log.Debug(
		"更新脚本：输入参数",
		zap.Uint32("script_id", scriptID),
		zap.Object("upload_script_biz", &dto),
	)

	om, rErr := s.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		log.Error(
			"更新脚本：查询脚本失败",
			zap.Uint32("script_id", scriptID),
			zap.Error(rErr),
		)
		return nil, rErr
	}
	log.Debug(
		"更新脚本：查询到的脚本详情",
		zap.Object("script_model", om),
	)
	if om.IsBuiltin {
		log.Error(
			"更新脚本：内置脚本不能修改",
			zap.Uint32("script_id", scriptID),
		)
		return nil, errors.FromReason(errors.ReasonScriptIsBuiltin).WithField("script_id", scriptID)
	}

	updateData := dto.ToUpdateMap()
	updateData["username"] = claims.Username
	updateStepStart := time.Now()
	log.Debug(
		"更新脚本：开始更新数据库模型",
		zap.Any("update_data", updateData),
		zap.Uint32("script_id", scriptID),
	)
	if err := s.scriptRepo.UpdateModel(ctx, updateData, "id = ?", scriptID); err != nil {
		log.Error(
			"更新脚本：更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.Any("update_data", updateData),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
		)
		return nil, errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新脚本：更新数据库模型成功",
		zap.Uint32("script_id", scriptID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	saveStepStart := time.Now()
	log.Debug("更新脚本：开始保存脚本文件")
	oldScriptPath := GetScriptStoragePath(om.Project, om.Label, om.Name, false)
	newScriptPath := GetScriptStoragePath(dto.Project, dto.Label, dto.Filename, false)
	if err := s.scriptRepo.SaveScriptFile(ctx, dto.File, newScriptPath, true); err != nil {
		log.Error(
			"更新脚本：新脚本写入失败",
			zap.Error(err),
			zap.String("script_path", newScriptPath),
			zap.Duration("save_step_duration", time.Since(saveStepStart)),
		)
		return nil, errors.ErrScriptSaveFailed.WithField("script_path", newScriptPath)
	}
	saveStepDuration := time.Since(saveStepStart)
	if oldScriptPath != newScriptPath {
		removeStepStart := time.Now()
		log.Debug(
			"更新脚本：开始删除旧脚本文件",
			zap.Uint32("script_id", scriptID),
			zap.String("old_script_path", oldScriptPath),
		)
		if err := s.scriptRepo.RemoveScriptFile(ctx, oldScriptPath); err != nil {
			log.Error(
				"更新脚本：删除旧脚本文件失败",
				zap.Error(err),
				zap.Uint32("script_id", scriptID),
				zap.String("script_path", oldScriptPath),
				zap.Duration("remove_step_duration", time.Since(removeStepStart)),
			)
			return nil, errors.ErrScriptRemoveFailed.WithField("script_path", oldScriptPath)
		}
		removeStepDuration := time.Since(removeStepStart)
		log.Debug(
			"更新脚本：删除旧脚本文件成功",
			zap.Uint32("script_id", scriptID),
			zap.String("script_path", oldScriptPath),
			zap.Duration("remove_step_duration", removeStepDuration),
		)
	}

	m, rErr := s.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		log.Error(
			"更新脚本：查询更新后的脚本详情失败",
			zap.Error(rErr),
			zap.Uint32("script_id", scriptID),
		)
		return nil, rErr
	}

	log.Info(
		"更新脚本：执行成功",
		zap.Uint32("script_id", scriptID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("save_step_duration", saveStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ScriptService) DeleteScriptByID(
	ctx context.Context,
	scriptID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info(
		"删除脚本：开始执行",
		zap.Uint32("script_id", scriptID),
	)

	m, rErr := s.FindScriptByID(ctx, scriptID)
	if rErr != nil {
		log.Error(
			"删除脚本：查询脚本详情失败",
			zap.Uint32("script_id", scriptID),
			zap.Error(rErr),
		)
		return rErr
	}
	if m.IsBuiltin {
		log.Error(
			"删除脚本：内置脚本不能删除",
			zap.Uint32("script_id", scriptID),
		)
		return errors.FromReason(errors.ReasonScriptIsBuiltin).WithField("script_id", scriptID)
	}

	scriptPath := GetScriptStoragePath(m.Project, m.Label, m.Name, false)
	deleteStepStart := time.Now()
	log.Debug(
		"删除脚本：开始删除数据库模型",
		zap.Uint32("script_id", scriptID),
	)
	if err := s.scriptRepo.DeleteModel(ctx, scriptID); err != nil {
		log.Error(
			"删除脚本：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"id": scriptID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除脚本：删除数据库模型成功",
		zap.Uint32("script_id", scriptID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	removeStepStart := time.Now()
	log.Debug(
		"删除脚本：开始删除脚本文件",
		zap.Uint32("script_id", scriptID),
		zap.String("script_path", scriptPath),
	)
	if err := s.scriptRepo.RemoveScriptFile(ctx, scriptPath); err != nil {
		log.Error(
			"删除脚本：脚本文件删除失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.String("script_path", scriptPath),
			zap.Duration("remove_step_duration", time.Since(removeStepStart)),
		)
		return errors.ErrScriptRemoveFailed.WithField("script_path", scriptPath)
	}
	removeStepDuration := time.Since(removeStepStart)
	log.Debug(
		"删除脚本：脚本文件删除成功",
		zap.Uint32("script_id", scriptID),
		zap.String("script_path", scriptPath),
		zap.Duration("remove_step_duration", removeStepDuration),
	)

	log.Info(
		"删除脚本：执行成功",
		zap.Uint32("script_id", scriptID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("remove_step_duration", removeStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ScriptService) FindScriptByID(
	ctx context.Context,
	scriptID uint32,
) (*jobmodel.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info(
		"查询脚本：开始执行",
		zap.Uint32("script_id", scriptID),
	)

	m, err := s.scriptRepo.GetModel(ctx, scriptID)
	if err != nil {
		log.Error(
			"查询脚本：查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("script_id", scriptID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": scriptID})
	}
	log.Debug(
		"查询脚本：查询到的数据库模型详情",
		zap.Object("script_model", m),
	)

	log.Info(
		"查询脚本：执行成功",
		zap.Uint32("script_id", scriptID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ScriptService) ListScript(
	ctx context.Context,
	page, size int,
	dto jobmodel.ListScriptDTO,
) (int64, []jobmodel.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("查询脚本列表：开始执行")

	log.Debug(
		"查询脚本列表：参数详情",
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

	log.Debug(
		"查询脚本列表：查询数据库模型参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询脚本列表：开始查询数据库模型总数",
		zap.Object("query_params", &qp),
	)
	count, err := s.scriptRepo.CountModel(ctx, qp.Query)
	countStepDuration := time.Since(countStepStart)
	if err != nil {
		log.Error(
			"查询脚本列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("count_step_duration", countStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询脚本列表：查询数据库模型总数成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		log.Warn(
			"查询脚本列表：数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return count, nil, nil
	}

	ms, err := s.scriptRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询脚本列表：查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	log.Info(
		"查询脚本列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *ScriptService) ListScriptsByIDs(
	ctx context.Context,
	scriptIDs []uint32,
) ([]jobmodel.ScriptModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info(
		"查询指定的脚本列表：开始执行",
		zap.Uint32s("script_ids", scriptIDs),
	)

	if len(scriptIDs) == 0 {
		log.Info(
			"查询指定的脚本列表：脚本 ID 列表为空",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return []jobmodel.ScriptModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": scriptIDs},
	}

	log.Debug(
		"查询指定的脚本列表：查询数据库模型参数",
		zap.Object("query_params", &qp),
	)

	ms, err := s.scriptRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询指定的脚本列表：查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询指定的脚本列表：查询数据库成功",
		zap.Uint32s("script_ids", scriptIDs),
	)

	log.Info(
		"查询指定的脚本列表：执行成功",
		zap.Uint32s("script_ids", scriptIDs),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (s *ScriptService) ListProjects(
	ctx context.Context,
	dto jobmodel.ListScriptDTO,
) ([]string, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("查询项目名称：开始执行")

	log.Debug(
		"查询项目名称：参数详情",
		zap.Object("list_script_dto", &dto),
	)

	query := dto.ToQueryMap()
	projects, err := s.scriptRepo.ListProjects(ctx, query)
	if err != nil {
		log.Error(
			"查询项目名称：数据库查询失败",
			zap.Error(err),
			zap.Any("query_params", query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	log.Info(
		"查询项目名称：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return projects, nil
}

func (s *ScriptService) ListLabels(
	ctx context.Context,
	dto jobmodel.ListScriptDTO,
) ([]string, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)
	log.Info("查询标签名称：开始执行")

	log.Debug(
		"查询标签名称：参数详情",
		zap.Object("list_script_dto", &dto),
	)

	query := dto.ToQueryMap()
	labels, err := s.scriptRepo.ListLabels(ctx, query)
	if err != nil {
		log.Error(
			"查询标签名称：数据库查询失败",
			zap.Error(err),
			zap.Any("query_params", query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	log.Info(
		"查询标签名称：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return labels, nil
}

func GetScriptStoragePath(project, label, name string, isBuiltin bool) string {
	if isBuiltin {
		return filepath.Join(config.ResourceDir, project, "script", label, name)
	}
	return filepath.Join(config.StorageDir, "script", project, label, name)
}
