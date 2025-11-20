package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbButton "gin-artweb/api/customer/button"
	pbMenu "gin-artweb/api/customer/menu"
	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
)

type ButtonService struct {
	log      *zap.Logger
	ucButton *biz.ButtonUsecase
}

func NewButtonService(
	logger *zap.Logger,
	ucButton *biz.ButtonUsecase,
) *ButtonService {
	return &ButtonService{
		log:      logger,
		ucButton: ucButton,
	}
}

// @Summary 新增按钮
// @Description 本接口用于新增按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param request body pbButton.CreateButtonRequest true "创建按钮请求"
// @Success 201 {object} pbButton.ButtonReply "成功返回按钮信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button [post]
// @Security ApiKeyAuth
func (s *ButtonService) CreateButton(ctx *gin.Context) {
	var req pbButton.CreateButtonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.log.Error(
			"绑定创建按钮请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始创建按钮",
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucButton.CreateButton(ctx, req.PermissionIDs, biz.ButtonModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: req.ID},
		},
		Name:         req.Name,
		ArrangeOrder: req.ArrangeOrder,
		IsActive:     req.IsActive,
		Descr:        req.Descr,
		MenuID:       req.MenuID,
	})
	if err != nil {
		s.log.Error(
			"创建按钮失败",
			zap.Error(err),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"创建按钮成功",
		zap.Uint32(pbComm.RequestPKKey, m.ID),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusCreated, &pbButton.ButtonReply{
		Code: http.StatusCreated,
		Data: ButtonModelToOut(*m),
	})
}

// @Summary 更新按钮
// @Description 本接口用于更新指定ID的按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param pk path uint true "按钮编号"
// @Param request body pbButton.UpdateButtonRequest true "更新按钮请求"
// @Success 200 {object} pbButton.ButtonReply "成功返回按钮信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "按钮未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button/{pk} [put]
// @Security ApiKeyAuth
func (s *ButtonService) UpdateButton(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定按钮ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	var req pbButton.UpdateButtonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.log.Error(
			"绑定更新按钮请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始更新按钮",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucButton.UpdateButtonByID(ctx, uri.PK, req.PermissionIDs, map[string]any{
		"name":          req.Name,
		"arrange_order": req.ArrangeOrder,
		"is_active":     req.IsActive,
		"descr":         req.Descr,
	}); err != nil {
		s.log.Error(
			"更新按钮失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"更新按钮成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucButton.FindButtonByID(ctx, []string{"Permissions", "Menu"}, uri.PK)
	if err != nil {
		s.log.Error(
			"查询更新后的按钮信息失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(err.Code, err.Reply())
		return
	}

	ctx.JSON(http.StatusOK, &pbButton.ButtonReply{
		Code: http.StatusOK,
		Data: ButtonModelToOut(*m),
	})
}

// @Summary 删除按钮
// @Description 本接口用于删除指定ID的按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param pk path uint true "按钮编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "按钮未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button/{pk} [delete]
// @Security ApiKeyAuth
func (s *ButtonService) DeleteButton(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除按钮ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始删除按钮",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := s.ucButton.DeleteButtonByID(ctx, uri.PK); err != nil {
		s.log.Error(
			"删除按钮失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"删除按钮成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询按钮
// @Description 本接口用于查询指定ID的按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param pk path uint true "按钮编号"
// @Success 200 {object} pbButton.ButtonReply "成功返回按钮信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "按钮未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button/{pk} [get]
// @Security ApiKeyAuth
func (s *ButtonService) GetButton(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询按钮ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始查询按钮详情",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := s.ucButton.FindButtonByID(ctx, []string{"Permissions", "Menu"}, uri.PK)
	if err != nil {
		s.log.Error(
			"查询按钮详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestPKKey, uri.PK),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"查询按钮详情成功",
		zap.Uint32(pbComm.RequestPKKey, uri.PK),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &pbButton.ButtonReply{
		Code: http.StatusOK,
		Data: ButtonModelToOut(*m),
	})
}

// @Summary 查询按钮列表
// @Description 本接口用于查询按钮列表
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param name query string false "按钮名称"
// @Param menu_id query uint false "菜单ID"
// @Param is_active query bool false "是否激活"
// @Param descr query string false "按钮描述"
// @Success 200 {object} pbButton.PagButtonBaseReply "成功返回按钮列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button [get]
// @Security ApiKeyAuth
func (s *ButtonService) ListButton(ctx *gin.Context) {
	var req pbButton.ListButtonRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询按钮列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}

	s.log.Info(
		"开始查询按钮列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	total, ms, err := s.ucButton.ListButton(ctx, page, size, query, []string{"id"}, true, nil)
	if err != nil {
		s.log.Error(
			"查询按钮列表失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		ctx.JSON(err.Code, err.Reply())
		return
	}

	s.log.Info(
		"查询按钮列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	mbs := ListButtonModelToOutBase(ms)
	ctx.JSON(http.StatusOK, &pbButton.PagButtonBaseReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

func (s *ButtonService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/button", s.CreateButton)
	r.PUT("/button/:pk", s.UpdateButton)
	r.DELETE("/button/:pk", s.DeleteButton)
	r.GET("/button/:pk", s.GetButton)
	r.GET("/button", s.ListButton)
}

func ButtonModelToOutBase(
	m biz.ButtonModel,
) *pbButton.ButtonOutBase {
	return &pbButton.ButtonOutBase{
		ID:           m.ID,
		CreatedAt:    m.CreatedAt.String(),
		UpdatedAt:    m.UpdatedAt.String(),
		Name:         m.Name,
		ArrangeOrder: m.ArrangeOrder,
		IsActive:     m.IsActive,
		Descr:        m.Descr,
	}
}

func ListButtonModelToOutBase(
	bms *[]biz.ButtonModel,
) *[]pbButton.ButtonOutBase {
	if bms == nil {
		return &[]pbButton.ButtonOutBase{}
	}
	ms := *bms
	mso := make([]pbButton.ButtonOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ButtonModelToOutBase(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}

func ButtonModelToOut(
	m biz.ButtonModel,
) *pbButton.ButtonOut {
	var menu *pbMenu.MenuOutBase
	if m.Menu.ID != 0 { // 或其他合适的判断条件
		menu = MenuModelToOutBase(m.Menu)
	}
	return &pbButton.ButtonOut{
		ButtonOutBase: *ButtonModelToOutBase(m),
		Menu:          menu,
		Permissions:   ListPermModelToOut(&m.Permissions),
	}
}
