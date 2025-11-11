package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pnComm "gin-artweb/api/common"
	pbMenu "gin-artweb/api/customer/menu"
	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
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
func (s *MenuService) CreateMenu(ctx *gin.Context) {
	var req pbMenu.CreateMenuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if req.ParentID != nil && *req.ParentID == 0 {
		req.ParentID = nil
	}
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
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := MenuModelToOut(*m)
	ctx.JSON(http.StatusCreated, &pbMenu.MenuReply{
		Code: http.StatusCreated,
		Data: *mo,
	})
}

// @Summary 更新菜单
// @Description 本接口用于更新菜单
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
func (s *MenuService) UpdateMenu(ctx *gin.Context) {
	var uri pnComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pbMenu.UpdateMenuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	meta := biz.Meta{
		Icon:  req.Meta.Icon,
		Title: req.Meta.Title,
	}
	data := map[string]any{
		"path":          req.Path,
		"component":     req.Component,
		"name":          req.Name,
		"meta":          meta.Json(),
		"arrange_order": req.ArrangeOrder,
		"is_active":     req.IsActive,
		"descr":         req.Descr,
	}
	if req.ParentID != nil && *req.ParentID == 0 {
		data["ParentId"] = nil
	}
	if err := s.ucMenu.UpdateMenuByID(ctx, uri.PK, req.PermissionIDs, data); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	m, err := s.ucMenu.FindMenuByID(ctx, []string{"Parent", "Permissions"}, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := MenuModelToOut(*m)
	ctx.JSON(http.StatusOK, &pbMenu.MenuReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 删除菜单
// @Description 本接口用于删除指定ID的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param pk path uint true "菜单编号"
// @Success 200 {object} common.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{pk} [delete]
func (s *MenuService) DeleteMenu(ctx *gin.Context) {
	var uri pnComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucMenu.DeleteMenuByID(ctx, uri.PK); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(common.NoDataReply.Code, common.NoDataReply)
}

// @Summary 查询单个菜单
// @Description 本接口用于查询一个菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param pk path uint true "菜单编号"
// @Success 200 {object} pbMenu.MenuReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{pk} [get]
func (s *MenuService) GetMenu(ctx *gin.Context) {
	var uri pnComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucMenu.FindMenuByID(ctx, []string{"Parent", "Permissions"}, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := MenuModelToOut(*m)
	ctx.JSON(http.StatusOK, &pbMenu.MenuReply{
		Code: http.StatusOK,
		Data: *mo,
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
// @Success 200 {object} pbMenu.PagMenuBaseReply "成功返回菜单列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu [get]
func (s *MenuService) ListMenu(ctx *gin.Context) {
	var req pbMenu.ListMenuRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	page, size, query := req.Query()
	total, ms, err := s.ucMenu.ListMenu(ctx, page, size, query, []string{"id"}, true, nil)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mbs := ListMenuModelToOutBase(ms)
	ctx.JSON(http.StatusOK, &pbMenu.PagMenuBaseReply{
		Code: http.StatusOK,
		Data: common.NewPag(page, size, total, mbs),
	})
}

func (s *MenuService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/menuinfo", s.CreateMenu)
	r.PUT("/menuinfo/:pk", s.UpdateMenu)
	r.DELETE("/menuinfo/:pk", s.DeleteMenu)
	r.GET("/menuinfo/:pk", s.GetMenu)
	r.GET("/menuinfo", s.ListMenu)
}

func MenuModelToOutBase(
	m biz.MenuModel,
) *pbMenu.MenuOutBase {
	return &pbMenu.MenuOutBase{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.String(),
		UpdatedAt: m.UpdatedAt.String(),
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

func ListMenuModelToOutBase(
	ms []biz.MenuModel,
) []*pbMenu.MenuOutBase {
	mso := make([]*pbMenu.MenuOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := MenuModelToOutBase(m)
			mso = append(mso, mo)
		}
	}
	return mso
}

func MenuModelToOut(
	m biz.MenuModel,
) *pbMenu.MenuOut {
	var parent *pbMenu.MenuOutBase
	if m.Parent != nil {
		parent = MenuModelToOutBase(*m.Parent)
	}
	return &pbMenu.MenuOut{
		MenuOutBase: *MenuModelToOutBase(m),
		Parent:      parent,
		Permissions: ListPermModelToOut(m.Permissions),
	}
}
