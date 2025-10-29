package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "gitee.com/keion8620/go-dango-gin/api/customer/menu"
	"gitee.com/keion8620/go-dango-gin/internal/customer/biz"
	"gitee.com/keion8620/go-dango-gin/pkg/common"
	"gitee.com/keion8620/go-dango-gin/pkg/database"
	"gitee.com/keion8620/go-dango-gin/pkg/errors"
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
// @Param request body pb.CreateMenuRequest true "创建菜单请求"
// @Success 200 {object} pb.MenuReply "成功返回菜单信息"
// @Router /api/v1/customer/menu [post]
func (s *MenuService) CreateMenu(ctx *gin.Context) {
	var req pb.CreateMenuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if *req.ParentId == 0 {
		req.ParentId = nil
	}
	m, err := s.ucMenu.CreateMenu(ctx, req.PermissionIds, biz.MenuModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{Id: req.Id},
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
		ParentId:     req.ParentId,
	})
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := MenuModelToOut(*m)
	ctx.JSON(http.StatusCreated, &pb.MenuReply{
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
// @Param request body pb.UpdateMenuRequest true "更新菜单请求"
// @Success 200 {object} pb.MenuReply "成功返回菜单信息"
// @Router /api/v1/customer/menu/{pk} [put]
func (s *MenuService) UpdateMenu(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pb.UpdateMenuRequest
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
		"Path":         req.Path,
		"Component":    req.Component,
		"Name":         req.Name,
		"Meta":         meta.Json(),
		"ArrangeOrder": req.ArrangeOrder,
		"IsActive":     req.IsActive,
		"Descr":        req.Descr,
	}
	if req.ParentId != nil && *req.ParentId == 0 {
		data["ParentId"] = nil
	}
	if err := s.ucMenu.UpdateMenuById(ctx, uri.Pk, req.PermissionIds, data); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	m, err := s.ucMenu.FindMenuById(ctx, []string{"Parent", "Permissions"}, uri.Pk)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := MenuModelToOut(*m)
	ctx.JSON(http.StatusOK, &pb.MenuReply{
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
// @Success 200 {object} errors.Reply[map[string]string] "删除成功"
// @Router /api/v1/customer/menu/{pk} [delete]
func (s *MenuService) DeleteMenu(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucMenu.DeleteMenuById(ctx, uri.Pk); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, common.NoDataReply)
}

// @Summary 查询单个菜单
// @Description 本接口用于查询一个菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param pk path uint true "菜单编号"
// @Success 200 {object} pb.MenuReply "成功返回用户信息"
// @Router /api/v1/customer/menu/{pk} [get]
func (s *MenuService) GetMenu(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucMenu.FindMenuById(ctx, []string{"Parent", "Permissons"}, uri.Pk)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mo := MenuModelToOut(*m)
	ctx.JSON(http.StatusOK, &pb.MenuReply{
		Code: http.StatusOK,
		Data: *mo,
	})
}

// @Summary 查询菜单列表
// @Description 本接口用于查询菜单列表
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Success 200 {object} pb.PagMenuBaseReply "成功返回菜单列表"
// @Router /api/v1/customer/menu [get]
func (s *MenuService) ListMenu(ctx *gin.Context) {
	var req pb.ListMenuRequest
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
	ctx.JSON(http.StatusOK, &pb.PagMenuBaseReply{
		Code: http.StatusOK,
		Data: common.NewPag(page, size, total, mbs),
	})
}

func (s *MenuService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/menu", s.CreateMenu)
	r.PUT("/menu/:pk", s.UpdateMenu)
	r.DELETE("/menu/:pk", s.DeleteMenu)
	r.GET("/menu/:pk", s.GetMenu)
	r.GET("/menu", s.ListMenu)
}

func MenuModelToOutBase(
	m biz.MenuModel,
) *pb.MenuOutBase {
	return &pb.MenuOutBase{
		Id:        m.Id,
		CreatedAt:  m.CreatedAt.String(),
		UpdatedAt:  m.UpdatedAt.String(),
		Path:      m.Path,
		Component: m.Component,
		Name:      m.Name,
		Meta: pb.MetaSchemas{
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
) []*pb.MenuOutBase {
	mso := make([]*pb.MenuOutBase, 0, len(ms))
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
) *pb.MenuOut {
	var parent *pb.MenuOutBase
	if m.Parent != nil {
		parent = MenuModelToOutBase(*m.Parent)
	}
	return &pb.MenuOut{
		MenuOutBase: *MenuModelToOutBase(m),
		Parent:      parent,
		Permissions: ListPermModelToOut(m.Permissions),
	}
}
