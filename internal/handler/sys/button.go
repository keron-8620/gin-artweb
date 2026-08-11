package sys

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	sysmodel "gin-artweb/internal/model/sys"
	syssvc "gin-artweb/internal/service/sys"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

// ButtonHandler 处理按钮相关的请求
// 包含日志记录和按钮服务的引用
type ButtonHandler struct {
	log       *zap.Logger           // 日志记录器
	buttonSvc *syssvc.ButtonService // 按钮服务
}

func NewButtonHandler(
	logger *zap.Logger,
	svcButton *syssvc.ButtonService,
) *ButtonHandler {
	return &ButtonHandler{
		log:       logger,
		buttonSvc: svcButton,
	}
}

// @Summary 新增按钮
// @Description 本接口用于新增按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param request body sysmodel.CreateButtonDTO true "创建按钮请求"
// @Success 201 {object} sysmodel.ButtonResp "创建按钮成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button [post]
// @Security ApiKeyAuth
func (h *ButtonHandler) CreateButton(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("创建按钮:开始执行")

	var req sysmodel.CreateButtonDTO
	if !common.ShouldBind(c, log, &req, "创建按钮:绑定创建按钮请求参数失败") {
		return
	}

	m, err := h.buttonSvc.CreateButton(ctx, req)
	if err != nil {
		log.Error(
			"创建按钮:执行失败",
			zap.Error(err),
			zap.Object("create_button_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"创建按钮:执行成功",
		zap.Uint32("button_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusCreated, &sysmodel.ButtonResp{
		Code: http.StatusCreated,
		Data: sysmodel.ButtonModelToDetailOut(*m),
	})
}

// @Summary 更新按钮
// @Description 本接口用于更新指定ID的按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param id path uint true "按钮编号"
// @Param request body sysmodel.UpdateButtonDTO true "更新按钮请求"
// @Success 200 {object} sysmodel.ButtonResp "更新按钮成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "按钮未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button/{id} [put]
// @Security ApiKeyAuth
func (h *ButtonHandler) UpdateButton(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("更新按钮:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "更新按钮:绑定更新按钮ID参数失败") {
		return
	}

	var req sysmodel.UpdateButtonDTO
	if !common.ShouldBind(c, log, &req, "更新按钮:绑定更新按钮请求参数失败") {
		return
	}

	m, err := h.buttonSvc.UpdateButtonByID(ctx, uri.ID, req)
	if err != nil {
		log.Error(
			"更新按钮:执行失败",
			zap.Error(err),
			zap.Uint32("button_id", uri.ID),
			zap.Object("update_button_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"更新按钮:执行成功",
		zap.Uint32("button_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &sysmodel.ButtonResp{
		Code: http.StatusOK,
		Data: sysmodel.ButtonModelToDetailOut(*m),
	})
}

// @Summary 删除按钮
// @Description 本接口用于删除指定ID的按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param id path uint true "按钮编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "按钮未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button/{id} [delete]
// @Security ApiKeyAuth
func (h *ButtonHandler) DeleteButton(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("删除按钮:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "删除按钮:绑定删除按钮ID参数失败") {
		return
	}

	err := h.buttonSvc.DeleteButtonByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"删除按钮:执行失败",
			zap.Error(err),
			zap.Uint32("button_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"删除按钮:删除按钮成功",
		zap.Uint32("button_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询按钮
// @Description 本接口用于查询指定ID的按钮
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param id path uint true "按钮编号"
// @Success 200 {object} sysmodel.ButtonResp "获取按钮详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "按钮未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button/{id} [get]
// @Security ApiKeyAuth
func (h *ButtonHandler) GetButton(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询按钮:绑定查询按钮ID参数失败") {
		return
	}

	m, err := h.buttonSvc.FindButtonByID(ctx, []string{"Apis", "Menu"}, uri.ID)
	if err != nil {
		log.Error(
			"查询按钮:执行失败",
			zap.Error(err),
			zap.Uint32("button_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, &sysmodel.ButtonResp{
		Code: http.StatusOK,
		Data: sysmodel.ButtonModelToDetailOut(*m),
	})
}

// @Summary 查询按钮列表
// @Description 本接口用于查询按钮列表
// @Tags 按钮管理
// @Accept json
// @Produce json
// @Param request query sysmodel.ListButtonDTO false "查询参数"
// @Success 200 {object} sysmodel.PagButtonResp "成功返回按钮列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/button [get]
// @Security ApiKeyAuth
func (h *ButtonHandler) ListButton(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("查询按钮列表:开始执行")

	var req sysmodel.ListButtonDTO
	if !common.ShouldBindQuery(c, log, &req, "查询按钮列表:绑定查询按钮列表参数失败") {
		return
	}

	page, size, total, ms, err := h.buttonSvc.ListButton(ctx, req)
	if err != nil {
		log.Error(
			"查询按钮列表:查询按钮列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_button_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mbs := sysmodel.ListButtonModelToStandardOut(ms)
	c.JSON(http.StatusOK, &sysmodel.PagButtonResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// LoadRouter 注册按钮相关的路由
func (h *ButtonHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/button", h.CreateButton)
	r.PUT("/button/:id", h.UpdateButton)
	r.DELETE("/button/:id", h.DeleteButton)
	r.GET("/button/:id", h.GetButton)
	r.GET("/button", h.ListButton)
}
