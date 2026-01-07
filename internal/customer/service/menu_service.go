package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbMenu "gin-artweb/api/customer/menu"
	"gin-artweb/internal/customer/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type MenuService struct {
	log    *zap.Logger
	ucMenu *biz.MenuUsecase
}

func NewMenuService(
	logger *zap.Logger,
	ucMenu *biz.MenuUsecase,
) *MenuService {
	return &MenuService{
		log:    logger,
		ucMenu: ucMenu,
	}
}

// @Summary 新增菜单
// @Description 本接口用于新增菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param request body pbMenu.CreateMenuRequest true "创建菜单请求"
// @Success 200 {object} pbMenu.MenuReply "成功返回菜单信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu [post]
// @Security ApiKeyAuth
func (s *MenuService) CreateMenu(ctx *gin.Context) {
	var req pbMenu.CreateMenuRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建菜单请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	if req.ParentID != nil && *req.ParentID == 0 {
		req.ParentID = nil
	}

	s.log.Info(
		"开始创建菜单",
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucMenu.CreateMenu(ctx, req.PermissionIDs, biz.MenuModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: req.ID},
		},
		Path:      req.Path,
		Component: req.Component,
		Name:      req.Name,
		Meta: biz.Meta{
			Icon:  req.Meta.Icon,
			Title: req.Meta.Title,
		},
		ArrangeOrder: req.ArrangeOrder,
		IsActive:     req.IsActive,
		Descr:        req.Descr,
		ParentID:     req.ParentID,
	})
	if err != nil {
		s.log.Error(
			"创建菜单失败",
			zap.Error(err),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"创建菜单成功",
		zap.Uint32(pbComm.RequestPKKey, m.ID),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := MenuModelToDetailOut(*m)
	ctx.JSON(http.StatusCreated, &pbMenu.MenuReply{
		Code: http.StatusCreated,
		Data: mo,
	})
}

// @Summary 更新菜单
// @Description 本接口用于更新指定ID的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param pk path uint true "菜单编号"
// @Param request body pbMenu.UpdateMenuRequest true "更新菜单请求"
// @Success 200 {object} pbMenu.MenuReply "成功返回菜单信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{pk} [put]
// @Security ApiKeyAuth
func (s *MenuService) UpdateMenu(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定菜单ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbMenu.UpdateMenuRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新菜单请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始更新菜单",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	data := map[string]any{
		"path":          req.Path,
		"component":     req.Component,
		"name":          req.Name,
		"meta":          req.Meta.Json(),
		"arrange_order": req.ArrangeOrder,
		"is_active":     req.IsActive,
		"descr":         req.Descr,
	}
	if req.ParentID != nil && *req.ParentID != 0 {
		data["parent_id"] = req.ParentID
	}

	m, err := s.ucMenu.UpdateMenuByID(ctx, uri.PK, req.PermissionIDs, data)
	if err != nil {
		s.log.Error(
			"更新菜单失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"更新菜单成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := MenuModelToDetailOut(*m)
	ctx.JSON(http.StatusOK, &pbMenu.MenuReply{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 删除菜单
// @Description 本接口用于删除指定ID的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param pk path uint true "菜单编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{pk} [delete]
// @Security ApiKeyAuth
func (s *MenuService) DeleteMenu(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除菜单ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除菜单",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucMenu.DeleteMenuByID(ctx, uri.PK); err != nil {
		s.log.Error(
			"删除菜单失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除菜单成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询菜单
// @Description 本接口用于查询指定ID的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param pk path uint true "菜单编号"
// @Success 200 {object} pbMenu.MenuReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{pk} [get]
// @Security ApiKeyAuth
func (s *MenuService) GetMenu(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询菜单ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询菜单详情",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucMenu.FindMenuByID(ctx, []string{"Parent", "Permissions"}, uri.PK)
	if err != nil {
		s.log.Error(
			"查询菜单详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询菜单详情成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mo := MenuModelToDetailOut(*m)
	ctx.JSON(http.StatusOK, &pbMenu.MenuReply{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 查询菜单列表
// @Description 本接口用于查询菜单列表
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param name query string false "菜单名称"
// @Param path query string false "菜单路径"
// @Param is_active query bool false "是否激活"
// @Param parent_id query int false "父级菜单ID"
// @Success 200 {object} pbMenu.PagMenuReply "成功返回菜单列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu [get]
// @Security ApiKeyAuth
func (s *MenuService) ListMenu(ctx *gin.Context) {
	var req pbMenu.ListMenuRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询菜单列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询菜单列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Limit:   size,
		Offset:  page,
		OrderBy: []string{"id ASC"},
		Query:   query,
	}
	total, ms, err := s.ucMenu.ListMenu(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询菜单列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询菜单列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mbs := ListMenuModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &pbMenu.PagMenuReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *MenuService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/menu", s.CreateMenu)
	r.PUT("/menu/:pk", s.UpdateMenu)
	r.DELETE("/menu/:pk", s.DeleteMenu)
	r.GET("/menu/:pk", s.GetMenu)
	r.GET("/menu", s.ListMenu)
}

func MenuModelToBaseOut(
	m biz.MenuModel,
) *pbMenu.MenuBaseOut {
	return &pbMenu.MenuBaseOut{
		ID:        m.ID,
		Path:      m.Path,
		Component: m.Component,
		Name:      m.Name,
		Meta: pbMenu.MetaSchemas{
			Title: m.Meta.Title,
			Icon:  m.Meta.Icon,
		},
		ArrangeOrder: m.ArrangeOrder,
		IsActive:     m.IsActive,
		Descr:        m.Descr,
	}
}

func MenuModelToStandardOut(
	m biz.MenuModel,
) *pbMenu.MenuStandardOut {
	return &pbMenu.MenuStandardOut{
		MenuBaseOut: *MenuModelToBaseOut(m),
		CreatedAt:   m.CreatedAt.String(),
		UpdatedAt:   m.UpdatedAt.String(),
	}
}

func MenuModelToDetailOut(
	m biz.MenuModel,
) *pbMenu.MenuDetailOut {
	var parent *pbMenu.MenuStandardOut
	if m.Parent != nil {
		parent = MenuModelToStandardOut(*m.Parent)
	}
	var permissionIDs []uint32
	if len(m.Permissions) > 0 {
		permissionIDs = make([]uint32, len(m.Permissions))
		for i, p := range m.Permissions {
			permissionIDs[i] = p.ID
		}
	}
	return &pbMenu.MenuDetailOut{
		MenuStandardOut: *MenuModelToStandardOut(m),
		Parent:          parent,
		Permissions:     permissionIDs,
	}
}

func ListMenuModelToStandardOut(
	mms *[]biz.MenuModel,
) *[]pbMenu.MenuStandardOut {
	if mms == nil {
		return &[]pbMenu.MenuStandardOut{}
	}
	ms := *mms
	mso := make([]pbMenu.MenuStandardOut, 0, len(ms))
	for _, m := range ms {
		mo := MenuModelToStandardOut(m)
		mso = append(mso, *mo)
	}
	return &mso
}
