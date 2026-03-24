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
func (h *ButtonHandler) CreateButton(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.CreateButtonDTO
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"新增按钮：绑定创建按钮请求参数失败",
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
		"创建按钮：开始执行",
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"创建按钮：入参详情",
		zap.Object("create_button_dto", &req),
	)

	createStepStart := time.Now()
	m, err := h.buttonSvc.CreateButton(ctx, req)
	createStepDuration := time.Since(createStepStart)
	if err != nil {
		log.Error(
			"创建按钮：创建按钮失败",
			zap.Error(err),
			zap.Object("create_button_dto", &req),
			zap.Duration("create_step_duration", createStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"创建按钮：创建按钮成功",
		zap.Object("button_model", m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	log.Info(
		"创建按钮：创建按钮成功",
		zap.Uint32("button_id", m.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusCreated, &sysmodel.ButtonResp{
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
func (h *ButtonHandler) UpdateButton(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"更新按钮：绑定按钮ID参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req sysmodel.UpdateButtonDTO
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(
			"更新按钮：绑定更新按钮请求参数失败",
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"更新按钮：开始执行",
		zap.Uint32("button_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"更新按钮：入参详情",
		zap.Uint32("button_id", uri.ID),
		zap.Object("update_button_dto", &req),
	)

	updateStepStart := time.Now()
	m, err := h.buttonSvc.UpdateButtonByID(ctx, uri.ID, req)
	updateStepDuration := time.Since(updateStepStart)
	if err != nil {
		log.Error(
			"更新按钮：更新按钮失败",
			zap.Error(err),
			zap.Uint32("button_id", uri.ID),
			zap.Object("update_button_dto", &req),
			zap.Duration("update_step_duration", updateStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"更新按钮：更新按钮成功",
		zap.Object("button_model", m),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	log.Info(
		"更新按钮：更新按钮成功",
		zap.Uint32("button_id", uri.ID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &sysmodel.ButtonResp{
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
func (h *ButtonHandler) DeleteButton(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"删除按钮：绑定按钮ID参数失败",
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
		"删除按钮：开始执行",
		zap.Uint32("button_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	deleteStepStart := time.Now()
	if err := h.buttonSvc.DeleteButtonByID(ctx, uri.ID); err != nil {
		log.Error(
			"删除按钮：删除按钮失败",
			zap.Error(err),
			zap.Uint32("button_id", uri.ID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除按钮：删除按钮成功",
		zap.Uint32("button_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	log.Info(
		"删除按钮：删除按钮成功",
		zap.Uint32("button_id", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
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
func (h *ButtonHandler) GetButton(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Error(
			"查询按钮：绑定按钮ID参数失败",
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
		"查询按钮：开始执行",
		zap.Uint32("button_id", uri.ID),
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	findStepStart := time.Now()
	m, err := h.buttonSvc.FindButtonByID(ctx, []string{"Apis", "Menu"}, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询按钮：查询失败",
			zap.Error(err),
			zap.Uint32("button_id", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"查询按钮：查询按钮成功",
		zap.Uint32("button_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	log.Info(
		"查询按钮：查询按钮成功",
		zap.Uint32("button_id", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &sysmodel.ButtonResp{
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
func (h *ButtonHandler) ListButton(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.ListButtonDTO
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Error(
			"查询按钮列表：绑定参数失败",
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
		"查询按钮列表：开始执行",
		zap.String("request_uri", ctx.Request.RequestURI),
		zap.String("request_method", ctx.Request.Method),
	)

	log.Debug(
		"查询按钮列表：参数详情",
		zap.Object("list_button_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.buttonSvc.ListButton(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询按钮列表：查询按钮列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_button_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询按钮列表：查询按钮列表成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := sysmodel.ListButtonModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &sysmodel.PagButtonResp{
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
