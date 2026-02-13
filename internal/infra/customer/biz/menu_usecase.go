package biz

import (
	"context"

	"go.uber.org/zap"

	"gin-artweb/internal/infra/customer/data"
	"gin-artweb/internal/infra/customer/model"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type MenuUsecase struct {
	log      *zap.Logger
	apiRepo  *data.ApiRepo
	menuRepo *data.MenuRepo
}

func NewMenuUsecase(
	log *zap.Logger,
	apiRepo *data.ApiRepo,
	menuRepo *data.MenuRepo,
) *MenuUsecase {
	return &MenuUsecase{
		log:      log,
		apiRepo:  apiRepo,
		menuRepo: menuRepo,
	}
}

func (uc *MenuUsecase) GetParentMenu(
	ctx context.Context,
	parentID *uint32,
) (*model.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if parentID == nil || *parentID == 0 {
		return nil, nil
	}

	uc.log.Info(
		"开始查询父菜单",
		zap.Uint32("parent_id", *parentID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.menuRepo.GetModel(ctx, nil, *parentID)
	if err != nil {
		uc.log.Error(
			"查询父菜单失败",
			zap.Error(err),
			zap.Uint32("parent_id", *parentID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"parent_id": *parentID})
	}

	uc.log.Info(
		"查询父菜单成功",
		zap.Uint32("parent_id", *parentID),
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *MenuUsecase) GetApis(
	ctx context.Context,
	apiIDs []uint32,
) (*[]model.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(apiIDs) == 0 {
		return &[]model.ApiModel{}, nil
	}

	uc.log.Info(
		"开始查询菜单关联的权限列表",
		zap.Uint32s("api_ids", apiIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": apiIDs},
	}
	_, ms, err := uc.apiRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询菜单关联的权限列表失败",
			zap.Error(err),
			zap.Uint32s("api_ids", apiIDs),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询菜单关联的权限列表成功",
		zap.Uint32s("api_ids", apiIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (uc *MenuUsecase) CreateMenu(
	ctx context.Context,
	apiIDs []uint32,
	m model.MenuModel,
) (*model.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始创建菜单",
		zap.Uint32s("api_ids", apiIDs),
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var (
		menu *model.MenuModel
		apis *[]model.ApiModel
		rErr *errors.Error
	)

	menu, rErr = uc.GetParentMenu(ctx, m.ParentID)
	if rErr != nil {
		return nil, rErr
	}
	if menu != nil {
		m.Parent = menu
	}

	apis, rErr = uc.GetApis(ctx, apiIDs)
	if rErr != nil {
		return nil, rErr
	}

	if err := uc.menuRepo.CreateModel(ctx, &m, apis); err != nil {
		uc.log.Error(
			"创建菜单失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	if apis != nil && len(*apis) > 0 {
		m.Apis = *apis
	}

	if err := uc.menuRepo.AddGroupPolicy(ctx, &m); err != nil {
		uc.log.Error(
			"添加菜单组策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"创建菜单成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *MenuUsecase) UpdateMenuByID(
	ctx context.Context,
	menuID uint32,
	apiIDs []uint32,
	data map[string]any,
) (*model.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始更新菜单",
		zap.Uint32("menu_id", menuID),
		zap.Uint32s("api_ids", apiIDs),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var (
		apis *[]model.ApiModel
		rErr *errors.Error
	)

	apis, rErr = uc.GetApis(ctx, apiIDs)
	if rErr != nil {
		return nil, rErr
	}

	data["id"] = menuID
	if err := uc.menuRepo.UpdateModel(ctx, data, apis, "id = ?", menuID); err != nil {
		uc.log.Error(
			"更新菜单失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	var m *model.MenuModel
	m, rErr = uc.FindMenuByID(ctx, []string{"Parent", "Apis"}, menuID)
	if rErr != nil {
		return nil, rErr
	}
	if err := uc.menuRepo.RemoveGroupPolicy(ctx, m, false); err != nil {
		uc.log.Error(
			"移除旧菜单组策略失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	if err := uc.menuRepo.AddGroupPolicy(ctx, m); err != nil {
		uc.log.Error(
			"添加新菜单组策略失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"更新菜单成功",
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *MenuUsecase) DeleteMenuByID(
	ctx context.Context,
	menuID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始删除菜单",
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := uc.FindMenuByID(ctx, []string{"Parent", "Apis"}, menuID)
	if rErr != nil {
		return rErr
	}

	if err := uc.menuRepo.DeleteModel(ctx, menuID); err != nil {
		uc.log.Error(
			"删除菜单失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": menuID})
	}

	if err := uc.menuRepo.RemoveGroupPolicy(ctx, m, true); err != nil {
		uc.log.Error(
			"移除菜单组策略失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}

	uc.log.Info(
		"删除菜单成功",
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *MenuUsecase) FindMenuByID(
	ctx context.Context,
	preloads []string,
	menuID uint32,
) (*model.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询菜单",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.menuRepo.GetModel(ctx, preloads, menuID)
	if err != nil {
		uc.log.Error(
			"查询菜单失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": menuID})
	}

	uc.log.Info(
		"查询菜单成功",
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *MenuUsecase) ListMenu(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询菜单列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.menuRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询菜单列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询菜单列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *MenuUsecase) LoadMenuPolicy(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始加载菜单策略",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Preloads: []string{"Apis"},
		Columns:  []string{"id", "parent_id"},
	}
	_, mms, err := uc.ListMenu(ctx, qp)
	if err != nil {
		uc.log.Error(
			"加载菜单策略时查询菜单列表失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}

	var policyCount int
	if mms != nil {
		ms := *mms
		policyCount = len(ms)
		for i := range ms {
			if err := uc.menuRepo.AddGroupPolicy(ctx, &ms[i]); err != nil {
				uc.log.Error(
					"加载菜单策略失败",
					zap.Error(err),
					zap.Uint32("menu_id", ms[i].ID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return errors.FromError(err)
			}
		}
	}
	uc.log.Info(
		"加载菜单策略成功",
		zap.Int("policy_count", policyCount),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}
