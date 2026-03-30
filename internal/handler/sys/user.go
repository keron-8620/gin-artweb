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

// UserHandler 处理用户相关的请求
// 包含日志记录和用户服务的引用
type UserHandler struct {
	log     *zap.Logger         // 日志记录器
	userSvc *syssvc.UserService // 用户服务
}

func NewUserHandler(
	log *zap.Logger,
	svcUser *syssvc.UserService,
) *UserHandler {
	return &UserHandler{
		log:     log,
		userSvc: svcUser,
	}
}

// @Summary 新增用户
// @Description 本接口用于新增用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body sysmodel.CreateUserDTO true "创建用户请求"
// @Success 201 {object} sysmodel.UserResp "创建用户成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user [post]
// @Security ApiKeyAuth
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.CreateUserDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"新增用户：绑定创建用户请求参数失败") {
		return
	}

	log.Info("创建用户：开始执行")

	log.Debug(
		"创建用户：入参详情",
		zap.Object("create_user_dto", &req),
	)

	createStepStart := time.Now()
	m, err := h.userSvc.CreateUser(ctx, req)
	createStepDuration := time.Since(createStepStart)
	if err != nil {
		log.Error(
			"创建用户：执行失败",
			zap.Error(err),
			zap.Object("create_user_dto", &req),
			zap.Duration("create_step_duration", createStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"创建用户：创建的用户模型详情",
		zap.Object("user_model", m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	log.Info(
		"创建用户：执行成功",
		zap.Uint32("uid", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusCreated, &sysmodel.UserResp{
		Code: http.StatusCreated,
		Data: sysmodel.UserModelToDetailOut(*m),
	})
}

// @Summary 更新用户
// @Description 本接口用于更新指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Param request body sysmodel.UpdateUserDTO true "更新用户请求"
// @Success 200 {object} sysmodel.UserResp "更新用户成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{id} [put]
// @Security ApiKeyAuth
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"更新用户：绑定用户ID参数失败") {
		return
	}
	var req sysmodel.UpdateUserDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"更新用户：绑定更新用户请求参数失败") {
		return
	}

	log.Info(
		"更新用户：开始执行",
		zap.Uint32("uid", uri.ID),
	)

	log.Debug(
		"更新用户：入参详情",
		zap.Uint32("uid", uri.ID),
		zap.Object("update_user_dto", &req),
	)

	updateStepStart := time.Now()
	err := h.userSvc.UpdateUserByID(ctx, uri.ID, req)
	updateStepDuration := time.Since(updateStepStart)
	if err != nil {
		log.Error(
			"更新用户：更新用户失败",
			zap.Error(err),
			zap.Uint32("uid", uri.ID),
			zap.Object("update_user_dto", &req),
			zap.Duration("update_step_duration", updateStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Debug(
		"更新用户：更新用户数据成功",
		zap.Uint32("uid", uri.ID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	findStepStart := time.Now()
	log.Debug(
		"更新用户：查询更新后的用户模型详情",
		zap.Uint32("uid", uri.ID),
	)
	m, err := h.userSvc.FindUserByID(ctx, []string{"Role"}, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"更新用户：查询更新后的用户模型详情失败",
			zap.Error(err),
			zap.Uint32("uid", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"更新用户：查询更新后的用户模型详情成功",
		zap.Uint32("uid", uri.ID),
		zap.Object("user_model", m),
		zap.Duration("find_step_duration", findStepDuration),
	)

	log.Info(
		"更新用户：执行成功",
		zap.Uint32("uid", uri.ID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &sysmodel.UserResp{
		Code: http.StatusOK,
		Data: sysmodel.UserModelToDetailOut(*m),
	})
}

// @Summary 删除用户
// @Description 本接口用于删除指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Success 200 {object} commodel.MapAPIResp "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{id} [delete]
// @Security ApiKeyAuth
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"删除用户：绑定用户ID参数失败") {
		return
	}

	log.Info(
		"删除用户：开始执行",
		zap.Uint32("uid", uri.ID),
	)

	deleteStepStart := time.Now()
	err := h.userSvc.DeleteUserByID(ctx, uri.ID)
	deleteStepDuration := time.Since(deleteStepStart)
	if err != nil {
		log.Error(
			"删除用户：执行失败",
			zap.Error(err),
			zap.Uint32("uid", uri.ID),
			zap.Duration("delete_step_duration", deleteStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"删除用户：执行成功",
		zap.Uint32("uid", uri.ID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 查询用户
// @Description 本接口用于查询指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Success 200 {object} sysmodel.UserResp "获取用户详情成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{id} [get]
// @Security ApiKeyAuth
func (h *UserHandler) GetUser(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"查询用户详情：绑定用户ID参数失败") {
		return
	}

	log.Info(
		"查询用户：开始执行",
		zap.Uint32("uid", uri.ID),
	)

	findStepStart := time.Now()
	m, err := h.userSvc.FindUserByID(ctx, []string{"Role"}, uri.ID)
	findStepDuration := time.Since(findStepStart)
	if err != nil {
		log.Error(
			"查询用户：执行失败",
			zap.Error(err),
			zap.Uint32("uid", uri.ID),
			zap.Duration("find_step_duration", findStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	log.Debug(
		"查询用户：查询到的用户模型详情",
		zap.Object("user_model", m),
		zap.Duration("find_step_duration", findStepDuration),
	)

	log.Info(
		"查询用户：执行成功",
		zap.Uint32("uid", uri.ID),
		zap.Duration("find_step_duration", findStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &sysmodel.UserResp{
		Code: http.StatusOK,
		Data: sysmodel.UserModelToDetailOut(*m),
	})
}

// @Summary 查询用户列表
// @Description 本接口用于查询用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request query sysmodel.ListUserDTO false "查询参数"
// @Success 200 {object} sysmodel.PagUserResp "成功返回用户列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user [get]
// @Security ApiKeyAuth
func (h *UserHandler) ListUser(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.ListUserDTO
	if !common.ShouldBindQuery(
		ctx, log, &req,
		"查询用户列表：绑定查询用户列表请求参数失败") {
		return
	}

	log.Info("查询用户列表：开始执行")

	log.Debug(
		"查询用户列表：参数详情",
		zap.Object("list_user_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.userSvc.ListUser(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询用户列表：执行失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_user_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询用户列表：执行成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := sysmodel.ListUserModelToDetailOut(ms)
	ctx.JSON(http.StatusOK, &sysmodel.PagUserResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 重置用户密码
// @Description 本接口用于重置指定ID的用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Param request body sysmodel.ResetPasswordDTO true "重置用户密码请求"
// @Success 200 {object} commodel.MapAPIResp "密码重置成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/password/{id} [patch]
// @Security ApiKeyAuth
func (h *UserHandler) ResetPassword(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var uri commodel.IDUri
	if !common.ShouldBind(
		ctx, log, &uri,
		"重置用户密码：绑定用户ID参数失败") {
		return
	}

	var req sysmodel.ResetPasswordDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"重置用户密码：绑定重置用户密码请求参数失败") {
		return
	}

	log.Info(
		"重置用户密码：开始执行",
		zap.Uint32("uid", uri.ID),
	)

	log.Debug(
		"重置用户密码：入参详情",
		zap.Uint32("uid", uri.ID),
	)

	resetStepStart := time.Now()
	err := h.userSvc.ResetPassword(ctx, uri.ID, req.NewPassword)
	resetStepDuration := time.Since(resetStepStart)
	if err != nil {
		log.Error(
			"重置用户密码：重置用户密码失败",
			zap.Error(err),
			zap.Uint32("uid", uri.ID),
			zap.Duration("reset_step_duration", resetStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"重置用户密码：执行成功",
		zap.Uint32("uid", uri.ID),
		zap.Duration("reset_step_duration", resetStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 修改当前用户密码
// @Description 本接口用于修改当前登录用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body sysmodel.PatchPasswordDTO true "修改用户密码请求"
// @Success 200 {object} commodel.MapAPIResp "密码修改成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 401 {object} errors.Error "认证失败"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/me/password [patch]
// @Security ApiKeyAuth
func (h *UserHandler) PatchPassword(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.PatchPasswordDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"修改个人密码：绑定修改个人密码请求参数失败") {
		return
	}

	claims := ctxutil.MustGetJwtClaims(ctx)
	log.Info(
		"修改个人密码：开始执行",
		zap.Uint32("uid", claims.UserID),
	)

	log.Debug(
		"修改个人密码：入参详情",
		zap.Uint32("uid", claims.UserID),
	)

	patchStepStart := time.Now()
	err := h.userSvc.PatchPassword(ctx, claims.UserID, req.OldPassword, req.NewPassword)
	patchStepDuration := time.Since(patchStepStart)
	if err != nil {
		log.Error(
			"修改个人密码：执行失败",
			zap.Error(err),
			zap.Uint32("uid", claims.UserID),
			zap.Duration("patch_step_duration", patchStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"修改个人密码：执行成功",
		zap.Uint32("uid", claims.UserID),
		zap.Duration("patch_step_duration", patchStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	ctx.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
}

// @Summary 登陆接口
// @Description 本接口用于登陆
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body sysmodel.LoginDTO true "登陆请求参数"
// @Success 200 {object} sysmodel.LoginResp "登录成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 401 {object} errors.Error "用户名或密码错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/login [post]
func (h *UserHandler) Login(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.LoginDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"用户登录：绑定用户登录请求参数失败") {
		return
	}

	log.Info(
		"用户登录：开始执行",
		zap.String("username", req.Username),
	)

	reqCtx := sysmodel.RequestContext{
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	}

	loginStart := time.Now()
	accessToken, refreshToken, err := h.userSvc.Login(ctx, req, reqCtx)
	loginStepDuration := time.Since(loginStart)
	if err != nil {
		log.Error(
			"用户登录：执行失败",
			zap.Error(err),
			zap.String("username", req.Username),
			zap.Duration("login_step_duration", loginStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"用户登录：执行成功",
		zap.String("username", req.Username),
		zap.Duration("login_step_duration", loginStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &sysmodel.LoginResp{
		Code: http.StatusOK,
		Data: sysmodel.LoginOut{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

// @Summary 刷新令牌接口
// @Description 本接口用于刷新令牌
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body sysmodel.RefreshTokenDTO true "刷新令牌请求参数"
// @Success 200 {object} sysmodel.LoginResp "刷新令牌成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 401 {object} errors.Error "用户名或密码错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/refresh/token [post]
func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.RefreshTokenDTO
	if !common.ShouldBind(
		ctx, log, &req,
		"刷新令牌：绑定刷新令牌请求参数失败") {
		return
	}

	refreshStart := time.Now()
	accessToken, refreshToken, rErr := h.userSvc.RefreshTokens(ctx, req.RefreshToken)
	refreshStepDuration := time.Since(refreshStart)
	if rErr != nil {
		log.Error(
			"刷新令牌：刷新令牌失败",
			zap.Error(rErr),
			zap.Duration("refresh_step_duration", refreshStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	log.Info(
		"刷新令牌：刷新令牌成功",
		zap.Duration("refresh_step_duration", refreshStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	ctx.JSON(http.StatusOK, &sysmodel.LoginResp{
		Code: http.StatusOK,
		Data: sysmodel.LoginOut{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

// @Summary 查询用户的登录记录列表
// @Description 本接口用于查询用户登录记录列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request query sysmodel.ListLoginRecordDTO false "查询参数"
// @Success 200 {object} sysmodel.PagLoginRecordResp "成功返回用户登录记录列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/record/login [get]
// @Security ApiKeyAuth
func (h *UserHandler) ListLoginRecord(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.ListLoginRecordDTO
	if !common.ShouldBindQuery(
		ctx, log, &req,
		"查询用户登录记录列表：绑定查询用户登录记录列表请求参数失败") {
		return
	}

	log.Info("查询用户登录记录列表：开始执行")

	log.Debug(
		"查询用户登录记录列表：入参详情",
		zap.Object("list_login_record_dto", &req),
	)

	listStepStart := time.Now()
	page, size := req.BaseModelQuery.GetPageParam()
	total, ms, err := h.userSvc.ListLoginRecord(ctx, page, size, req)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询用户登录记录列表：查询用户登录记录列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_login_record_dto", &req),
			zap.Duration("list_step_duration", listStepDuration),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询用户登录记录列表：执行成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := sysmodel.ListLoginRecordModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &sysmodel.PagLoginRecordResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 查询当前用户的登录记录列表
// @Description 本接口用于查询当前登录用户的登录记录列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request query sysmodel.ListLoginRecordDTO false "查询参数"
// @Success 200 {object} sysmodel.PagLoginRecordResp "成功返回用户登录记录列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 401 {object} errors.Error "未授权访问"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/me/record/login [get]
// @Security ApiKeyAuth
func (h *UserHandler) ListMeLoginRecord(ctx *gin.Context) {
	startTime := time.Now()
	log := ctxutil.NewLogger(h.log, ctx)
	var req sysmodel.ListLoginRecordDTO
	if !common.ShouldBindQuery(
		ctx, log, &req,
		"查询个人登录记录列表：绑定查询个人登录记录列表请求参数失败") {
		return
	}

	log.Info("查询个人登录记录列表：开始执行")

	claims := ctxutil.MustGetJwtClaims(ctx)
	req.Username = claims.Subject

	page, size := req.BaseModelQuery.GetPageParam()
	total, ms, err := h.userSvc.ListLoginRecord(ctx, page, size, req)
	if err != nil {
		log.Error(
			"查询个人登录记录列表：查询个人登录记录列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_login_record_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	log.Info(
		"查询个人登录记录列表：执行成功",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Int64("total", total),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	mbs := sysmodel.ListLoginRecordModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &sysmodel.PagLoginRecordResp{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// LoadRouter 注册用户相关的路由
func (h *UserHandler) LoadRouter(r *gin.RouterGroup) {
	r.POST("/user", h.CreateUser)
	r.PUT("/user/:id", h.UpdateUser)
	r.DELETE("/user/:id", h.DeleteUser)
	r.GET("/user/:id", h.GetUser)
	r.GET("/user", h.ListUser)
	r.PATCH("/user/password/:id", h.ResetPassword)
	r.GET("/user/record/login", h.ListLoginRecord)
	r.GET("/me/record/login", h.ListMeLoginRecord)
}
