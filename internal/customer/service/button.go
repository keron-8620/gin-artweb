package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "gin-artweb/api/customer/button"
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
// @Param request body pb.CreateButtonRequest true "创建按钮请求"
// @Success 200 {object} pb.ButtonReply "成功返回按钮信息"
// @Router /api/v1/customer/button [post]
func (s *ButtonService) CreateButton(ctx *gin.Context) {
	var req pb.CreateButtonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucButton.CreateButton(ctx, req.PermissionIds, biz.ButtonModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{Id: req.Id},
		},
		Name:         req.Name,
		ArrangeOrder: req.ArrangeOrder,
		IsActive:     req.IsActive,
		Descr:        req.Descr,
		MenuId:       req.MenuId,
	})
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusCreated, &pb.ButtonReply{
		Code: http.StatusCreated,
		Data: *ButtonModelToOut(*m),
	})
}

// @Summary 更新按钮
// @Description 本接口用于更新按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param pk path uint true "按钮编号"
// @Param request body pb.UpdateButtonRequest true "更新按钮请求"
// @Success 200 {object} pb.ButtonReply "成功返回按钮信息"
// @Router /api/v1/customer/button/{pk} [put]
func (s *ButtonService) UpdateButton(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pb.UpdateButtonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucButton.UpdateButtonById(ctx, uri.Pk, req.PermissionIds, map[string]any{
		"Name":         req.Name,
		"ArrangeOrder": req.ArrangeOrder,
		"IsActive":     req.IsActive,
		"Descr":        req.Descr,
	}); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	m, err := s.ucButton.FindButtonById(ctx, []string{"Permissions", "Menu"}, uri.Pk)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, &pb.ButtonReply{
		Code: http.StatusOK,
		Data: *ButtonModelToOut(*m),
	})
}

// @Summary 删除按钮
// @Description 本接口用于删除指定ID的按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param pk path uint true "按钮编号"
// @Router /api/v1/customer/button/{pk} [delete]
func (s *ButtonService) DeleteButton(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucButton.DeleteButtonById(ctx, uri.Pk); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, common.NoDataReply)
}

// @Summary 查询单个按钮
// @Description 本接口用于查询一个按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param pk path uint true "按钮编号"
// @Success 200 {object} pb.ButtonReply "成功返回按钮信息"
// @Router /api/v1/customer/button/{pk} [get]
func (s *ButtonService) GetButton(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucButton.FindButtonById(ctx, []string{"Permissions", "Menu"}, uri.Pk)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, &pb.ButtonReply{
		Code: http.StatusOK,
		Data: *ButtonModelToOut(*m),
	})
}

// @Summary 查询按钮列表
// @Description 本接口用于查询按钮列表
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Success 200 {object} pb.PagButtonBaseReply "成功返回按钮列表"
// @Router /api/v1/customer/button [get]
func (s *ButtonService) ListButton(ctx *gin.Context) {
	var req pb.ListButtonRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	page, size, query := req.Query()
	total, ms, err := s.ucButton.ListButton(ctx, page, size, query, []string{"id"}, true, nil)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mbs := ListButtonModelToOutBase(ms)
	ctx.JSON(http.StatusOK, &pb.PagButtonBaseReply{
		Code: http.StatusOK,
		Data: common.NewPag(page, size, total, mbs),
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
) *pb.ButtonOutBase {
	return &pb.ButtonOutBase{
		Id:           m.Id,
		CreatedAt:    m.CreatedAt.String(),
		UpdatedAt:    m.UpdatedAt.String(),
		Name:         m.Name,
		ArrangeOrder: m.ArrangeOrder,
		IsActive:     m.IsActive,
		Descr:        m.Descr,
	}
}

func ListButtonModelToOutBase(
	ms []biz.ButtonModel,
) []*pb.ButtonOutBase {
	mso := make([]*pb.ButtonOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ButtonModelToOutBase(m)
			mso = append(mso, mo)
		}
	}
	return mso
}

func ButtonModelToOut(
	m biz.ButtonModel,
) *pb.ButtonOut {
	return &pb.ButtonOut{
		ButtonOutBase: *ButtonModelToOutBase(m),
		Menu:          MenuModelToOutBase(m.Menu),
		Permissions:   ListPermModelToOut(m.Permissions),
	}
}
