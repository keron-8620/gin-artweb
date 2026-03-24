package sys

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	sysmodel "gin-artweb/internal/model/sys"
	syssvc "gin-artweb/internal/service/sys"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

// MenuHandler 处理菜单相关的请求
// 包含日志记录和菜单服务的引用
type MenuHandler struct {
	log     *zap.Logger         // 日志记录器
	menuSvc *syssvc.MenuService // 菜单服务
}

func NewMenuHandler(
	logger *zap.Logger,
	svcMenu *syssvc.MenuService,
) *MenuHandler {
	return &MenuHandler{
		log:     logger,
		menuSvc: svcMenu,
	}
}

// @Summary 新增菜单
// @Description 本接口用于新增菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param request body sysmodel.CreateMenuDTO true "创建菜单请求"
// @Success 201 {object} sysmodel.MenuResp "创建菜单成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu [post]
// @Security ApiKeyAuth
func (h *MenuHandler) CreateMenu(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.CreateMenuDTO
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"创建菜单：绑定请求参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	if req.ParentID != nil && *req.ParentID == 0 {
		req.ParentID = nil
	}

	log.Info(
		"创建菜单：开始执行",
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"创建菜单：入参详情",
		zap.Object("create_menu_dto", &req),
	)

	createStepStart := time.Now()
	m, err := h.menuSvc.CreateMenu(ctx, req)
	createStepDuration := time.Since(createStepStart)
	if err != nil {
		log.Error(
			"创建菜单：菜单创建失败",
			zap.Error(err),
			zap.Object("create_menu_dto", &req),
			zap.Duration("create_step_duration", createStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"创建菜单：菜单创建成功",
		zap.Object("menu_model", m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	log.Info(
		"创建菜单：菜单创建成功",
		zap.Uint32("menu_id", m.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := sysmodel.MenuModelToDetailOut(*m)
	ctx.JSON(http.StatusCreated, &sysmodel.MenuResp{
		Code: http.StatusCreated,
		Data: mo,
	})
}

// @Summary 更新菜单
// @Description 本接口用于更新指定ID的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path uint true "菜单编号"
// @Param request body sysmodel.UpdateMenuDTO true "更新菜单请求"
// @Success 200 {object} sysmodel.MenuResp "更新菜单成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{id} [put]
// @Security ApiKeyAuth
func (h *MenuHandler) UpdateMenu(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"更新菜单：绑定菜单ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req sysmodel.UpdateMenuDTO
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"更新菜单：绑定更新菜单请求参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"更新菜单：开始执行",
		zap.Uint32("menu_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"更新菜单：入参详情",
		zap.Uint32("menu_id", uri.ID),
		zap.Object("update_menu_dto", &req),
	)

	updateStepStart := time.Now()
	m, err := h.menuSvc.UpdateMenuByID(ctx, uri.ID, req)
	updateStepDuration := time.Since(updateStepStart)
	if err != nil {
		log.Error(
			"更新菜单：更新菜单失败",
			zap.Error(err),
			zap.Uint32("menu_id", uri.ID),
			zap.Object("update_menu_dto", &req),
			zap.Duration("update_step_duration", updateStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"更新菜单：更新菜单成功",
		zap.Object("menu_model", m),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	log.Info(
		"更新菜单：更新菜单成功",
		zap.Uint32("menu_id", uri.ID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := sysmodel.MenuModelToDetailOut(*m)
	ctx.JSON(http.StatusOK, &sysmodel.MenuResp{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 删除菜单
// @Description 本接口用于删除指定ID的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path uint true "菜单编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{id} [delete]
// @Security ApiKeyAuth
func (h *MenuHandler) DeleteMenu(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"删除菜单：绑定删除菜单ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"删除菜单：开始执行",
		zap.Uint32("menu_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	deleteStepStart := time.Now()
	if err := h.menuSvc.DeleteMenuByID(ctx, uri.ID); err != nil {
		log.Error(
			"删除菜单：删除菜单失败",
			zap.Error(err),
			zap.Uint32("menu_id", uri.ID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除菜单：删除菜单成功",
		zap.Uint32("menu_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	log.Info(
		"删除菜单：删除菜单成功",
		zap.Uint32("menu_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询菜单
// @Description 本接口用于查询指定ID的菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param id path uint true "菜单编号"
// @Success 200 {object} sysmodel.MenuResp "获取菜单详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "菜单未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu/{id} [get]
// @Security ApiKeyAuth
func (h *MenuHandler) GetMenu(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"查询菜单：绑定查询菜单ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"查询菜单：开始执行",
		zap.Uint32("menu_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	findStepStart := time.Now()
	m, err := h.menuSvc.FindMenuByID(ctx, []string{"Parent", "Apis"}, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询菜单：查询失败",
			zap.Error(err),
			zap.Uint32("menu_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"查询菜单：查询菜单成功",
		zap.Uint32("menu_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	log.Info(
		"查询菜单：查询菜单成功",
		zap.Uint32("menu_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mo := sysmodel.MenuModelToDetailOut(*m)
	ctx.JSON(http.StatusOK, &sysmodel.MenuResp{
		Code: http.StatusOK,
		Data: mo,
	})
}

// @Summary 查询菜单列表
// @Description 本接口用于查询菜单列表
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Param request query sysmodel.ListMenuDTO false "查询参数"
// @Success 200 {object} sysmodel.PagMenuResp "成功返回菜单列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/menu [get]
// @Security ApiKeyAuth
func (h *MenuHandler) ListMenu(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.ListMenuDTO
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Error(
			"查询菜单列表：绑定查询菜单列表参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"查询菜单列表：开始执行",
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"查询菜单列表：参数详情",
		zap.Object("list_menu_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.menuSvc.ListMenu(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询菜单列表：查询菜单列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_menu_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询菜单列表：查询菜单列表成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := sysmodel.ListMenuModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &sysmodel.PagMenuResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// LoadRouter 注册菜单相关的路由
func (h *MenuHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/menu", h.CreateMenu)
	r.PUT("/menu/:id", h.UpdateMenu)
	r.DELETE("/menu/:id", h.DeleteMenu)
	r.GET("/menu/:id", h.GetMenu)
	r.GET("/menu", h.ListMenu)
}
