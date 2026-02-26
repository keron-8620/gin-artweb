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

type ButtonService struct {
	log        *zap.Logger
	apiRepo    *custsvc.ApiRepo
	menuRepo   *custsvc.MenuRepo
	buttonRepo *custsvc.ButtonRepo
}

func NewButtonService(
	log *zap.Logger,
	apiRepo *custsvc.ApiRepo,
	menuRepo *custsvc.MenuRepo,
	buttonRepo *custsvc.ButtonRepo,
) *ButtonService {
	return &ButtonService{
		log:        log,
		apiRepo:    apiRepo,
		menuRepo:   menuRepo,
		buttonRepo: buttonRepo,
	}
}

func (s *ButtonService) GetMenu(
	ctx context.Context,
	menuID uint32,
) (*custmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询按钮关联的菜单",
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.menuRepo.GetModel(ctx, nil, menuID)
	if err != nil {
		s.log.Error(
			"查询按钮关联的菜单失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"menu_id": menuID})
	}

	s.log.Info(
		"查询按钮关联的菜单成功",
		zap.Uint32("menu_id", menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *ButtonService) GetApis(
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
		"开始查询按钮关联的API列表",
		zap.Uint32s("api_ids", apiIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": apiIDs},
	}
	_, ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询按钮关联的API列表失败",
			zap.Error(err),
			zap.Uint32s("api_ids", apiIDs),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询按钮关联的API列表成功",
		zap.Uint32s("api_ids", apiIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (s *ButtonService) CreateButton(
	ctx context.Context,
	apiIDs []uint32,
	m custmodel.ButtonModel,
) (*custmodel.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建按钮",
		zap.Uint32s("api_ids", apiIDs),
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var (
		menu *custmodel.MenuModel
		apis *[]custmodel.ApiModel
		rErr *errors.Error
	)

	menu, rErr = s.GetMenu(ctx, m.MenuID)
	if rErr != nil {
		return nil, rErr
	}
	m.Menu = *menu

	apis, rErr = s.GetApis(ctx, apiIDs)
	if rErr != nil {
		return nil, rErr
	}

	if err := s.buttonRepo.CreateModel(ctx, &m, apis); err != nil {
		s.log.Error(
			"创建按钮失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	if len(*apis) > 0 {
		m.Apis = *apis
	}

	if err := s.buttonRepo.AddGroupPolicy(ctx, &m); err != nil {
		s.log.Error(
			"添加按钮组策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	s.log.Info(
		"创建按钮成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (s *ButtonService) UpdateButtonByID(
	ctx context.Context,
	buttonID uint32,
	apiIDs []uint32,
	data map[string]any,
) (*custmodel.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新按钮",
		zap.Uint32("button_id", buttonID),
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

	data["id"] = buttonID
	if err := s.buttonRepo.UpdateModel(ctx, data, apis, "id = ?", buttonID); err != nil {
		s.log.Error(
			"更新按钮失败",
			zap.Error(err),
			zap.Uint32("button_id", buttonID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	var m *custmodel.ButtonModel
	m, rErr = s.FindButtonByID(ctx, []string{"Menu", "Apis"}, buttonID)
	if rErr != nil {
		return nil, rErr
	}

	if err := s.buttonRepo.RemoveGroupPolicy(ctx, m, false); err != nil {
		s.log.Error(
			"移除旧按钮组策略失败",
			zap.Error(err),
			zap.Uint32("button_id", buttonID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	if err := s.buttonRepo.AddGroupPolicy(ctx, m); err != nil {
		s.log.Error(
			"添加新按钮组策略失败",
			zap.Error(err),
			zap.Uint32("button_id", buttonID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	s.log.Info(
		"更新按钮成功",
		zap.Uint32("button_id", buttonID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *ButtonService) DeleteButtonByID(
	ctx context.Context,
	buttonID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除按钮",
		zap.Uint32("button_id", buttonID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := s.FindButtonByID(ctx, []string{"Menu", "Apis"}, buttonID)
	if rErr != nil {
		return rErr
	}

	if err := s.buttonRepo.DeleteModel(ctx, buttonID); err != nil {
		s.log.Error(
			"删除按钮失败",
			zap.Error(err),
			zap.Uint32("button_id", buttonID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": buttonID})
	}

	if err := s.buttonRepo.RemoveGroupPolicy(ctx, m, true); err != nil {
		s.log.Error(
			"移除按钮组策略失败",
			zap.Error(err),
			zap.Uint32("button_id", buttonID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}

	s.log.Info(
		"删除按钮成功",
		zap.Uint32("button_id", buttonID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *ButtonService) FindButtonByID(
	ctx context.Context,
	preloads []string,
	buttonID uint32,
) (*custmodel.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询按钮",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32("button_id", buttonID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.buttonRepo.GetModel(ctx, preloads, buttonID)
	if err != nil {
		s.log.Error(
			"查询按钮失败",
			zap.Error(err),
			zap.Uint32("button_id", buttonID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": buttonID})
	}

	s.log.Info(
		"查询按钮成功",
		zap.Uint32("button_id", buttonID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *ButtonService) ListButton(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]custmodel.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询按钮列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询按钮列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询按钮列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *ButtonService) LoadButtonPolicy(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始加载按钮策略",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Preloads: []string{"Apis"},
		Columns:  []string{"id", "menu_id"},
	}

	_, bms, rErr := s.ListButton(ctx, qp)
	if rErr != nil {
		s.log.Error(
			"加载按钮策略时查询按钮列表失败",
			zap.Error(rErr),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return rErr
	}

	var policyCount int
	if bms != nil {
		ms := *bms
		policyCount = len(ms)
		for i := range ms {
			if err := s.buttonRepo.AddGroupPolicy(ctx, &ms[i]); err != nil {
				s.log.Error(
					"加载按钮策略失败",
					zap.Error(err),
					zap.Uint32("menu_id", ms[i].ID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return errors.FromError(err)
			}
		}
	}

	s.log.Info(
		"加载按钮策略成功",
		zap.Int("policy_count", policyCount),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}
