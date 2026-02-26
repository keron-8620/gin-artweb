package customer

import (
	"context"

	"go.uber.org/zap"

	custmodel "gin-artweb/internal/model/customer"
	custsvc "gin-artweb/internal/repository/customer"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type MenuService struct {
	log      *zap.Logger
	apiRepo  *custsvc.ApiRepo
	menuRepo *custsvc.MenuRepo
}

func NewMenuService(
	log *zap.Logger,
	apiRepo *custsvc.ApiRepo,
	menuRepo *custsvc.MenuRepo,
) *MenuService {
	return &MenuService{
		log:      log,
		apiRepo:  apiRepo,
		menuRepo: menuRepo,
	}
}

func (s *MenuService) GetParentMenu(
	ctx context.Context,
	parentID *uint32,
) (*custmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if parentID == nil || *parentID == 0 {
		return nil, nil
	}

	s.log.Info(
		"开始查询父菜单",
		zap.Uint32("parent_id", *parentID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.menuRepo.GetModel(ctx, nil, *parentID)
	if err != nil {
		s.log.Error(
			"查询父菜单失败",
			zap.Error(err),
			zap.Uint32("parent_id", *parentID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"parent_id": *parentID})
	}

	s.log.Info(
		"查询父菜单成功",
		zap.Uint32("parent_id", *parentID),
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *MenuService) GetApis(
	ctx context.Context,
	apiIDs []uint32,
) (*[]custmodel.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(apiIDs) == 0 {
		return &[]custmodel.ApiModel{}, nil
	}

	s.log.Info(
		"开始查询菜单关联的权限列表",
		zap.Uint32s("api_ids", apiIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": apiIDs},
	}
	_, ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询菜单关联的权限列表失败",
			zap.Error(err),
			zap.Uint32s("api_ids", apiIDs),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询菜单关联的权限列表成功",
		zap.Uint32s("api_ids", apiIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (s *MenuService) CreateMenu(
	ctx context.Context,
	apiIDs []uint32,
	m custmodel.MenuModel,
) (*custmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建菜单",
		zap.Uint32s("api_ids", apiIDs),
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var (
		menu *custmodel.MenuModel
		apis *[]custmodel.ApiModel
		rErr *errors.Error
	)

	menu, rErr = s.GetParentMenu(ctx, m.ParentID)
	if rErr != nil {
		return nil, rErr
	}
	if menu != nil {
		m.Parent = menu
	}

	apis, rErr = s.GetApis(ctx, apiIDs)
	if rErr != nil {
		return nil, rErr
	}

	if err := s.menuRepo.CreateModel(ctx, &m, apis); err != nil {
		s.log.Error(
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

	if err := s.menuRepo.AddGroupPolicy(ctx, &m); err != nil {
		s.log.Error(
			"添加菜单组策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	s.log.Info(
		"创建菜单成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (s *MenuService) UpdateMenuByID(
	ctx context.Context,
	menuID uint32,
	apiIDs []uint32,
	data map[string]any,
) (*custmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新菜单",
		zap.Uint32("menu_id", menuID),
		zap.Uint32s("api_ids", apiIDs),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var (
		apis *[]custmodel.ApiModel
		rErr *errors.Error
	)

	apis, rErr = s.GetApis(ctx, apiIDs)
	if rErr != nil {
		return nil, rErr
	}

	data["id"] = menuID
	if err := s.menuRepo.UpdateModel(ctx, data, apis, "id = ?", menuID); err != nil {
		s.log.Error(
			"更新菜单失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	var m *custmodel.MenuModel
	m, rErr = s.FindMenuByID(ctx, []string{"Parent", "Apis"}, menuID)
	if rErr != nil {
		return nil, rErr
	}
	if err := s.menuRepo.RemoveGroupPolicy(ctx, m, false); err != nil {
		s.log.Error(
			"移除旧菜单组策略失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	if err := s.menuRepo.AddGroupPolicy(ctx, m); err != nil {
		s.log.Error(
			"添加新菜单组策略失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	s.log.Info(
		"更新菜单成功",
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *MenuService) DeleteMenuByID(
	ctx context.Context,
	menuID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除菜单",
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := s.FindMenuByID(ctx, []string{"Parent", "Apis"}, menuID)
	if rErr != nil {
		return rErr
	}

	if err := s.menuRepo.DeleteModel(ctx, menuID); err != nil {
		s.log.Error(
			"删除菜单失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": menuID})
	}

	if err := s.menuRepo.RemoveGroupPolicy(ctx, m, true); err != nil {
		s.log.Error(
			"移除菜单组策略失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}

	s.log.Info(
		"删除菜单成功",
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *MenuService) FindMenuByID(
	ctx context.Context,
	preloads []string,
	menuID uint32,
) (*custmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询菜单",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.menuRepo.GetModel(ctx, preloads, menuID)
	if err != nil {
		s.log.Error(
			"查询菜单失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": menuID})
	}

	s.log.Info(
		"查询菜单成功",
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *MenuService) ListMenu(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]custmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询菜单列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.menuRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询菜单列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询菜单列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *MenuService) LoadMenuPolicy(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始加载菜单策略",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Preloads: []string{"Apis"},
		Columns:  []string{"id", "parent_id"},
	}
	_, mms, err := s.ListMenu(ctx, qp)
	if err != nil {
		s.log.Error(
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
			if err := s.menuRepo.AddGroupPolicy(ctx, &ms[i]); err != nil {
				s.log.Error(
					"加载菜单策略失败",
					zap.Error(err),
					zap.Uint32("menu_id", ms[i].ID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return errors.FromError(err)
			}
		}
	}
	s.log.Info(
		"加载菜单策略成功",
		zap.Int("policy_count", policyCount),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}
