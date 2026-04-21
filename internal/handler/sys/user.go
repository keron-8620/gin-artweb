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
func (h *UserHandler) CreateUser(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("新增用户:开始执行")

	var req sysmodel.CreateUserDTO
	if !common.ShouldBind(c, log, &req, "新增用户:绑定创建用户请求参数失败") {
		return
	}

	m, err := h.userSvc.CreateUser(ctx, req)
	if err != nil {
		log.Error(
			"创建用户:执行失败",
			zap.Error(err),
			zap.Object("create_user_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"创建用户:执行成功",
		zap.Uint32("user_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusCreated, &sysmodel.UserResp{
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
func (h *UserHandler) UpdateUser(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("更新用户:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "更新用户:绑定用户ID参数失败") {
		return
	}

	var req sysmodel.UpdateUserDTO
	if !common.ShouldBind(c, log, &req, "更新用户:绑定更新用户请求参数失败") {
		return
	}

	err := h.userSvc.UpdateUserByID(ctx, uri.ID, req)
	if err != nil {
		log.Error(
			"更新用户:更新用户失败",
			zap.Error(err),
			zap.Uint32("user_id", uri.ID),
			zap.Object("update_user_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	m, err := h.userSvc.FindUserByID(ctx, []string{"Role"}, uri.ID)
	if err != nil {
		log.Error(
			"更新用户:查询更新后的用户模型详情失败",
			zap.Error(err),
			zap.Uint32("user_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"更新用户:执行成功",
		zap.Uint32("user_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &sysmodel.UserResp{
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
func (h *UserHandler) DeleteUser(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("删除用户:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "删除用户:绑定用户ID参数失败") {
		return
	}

	err := h.userSvc.DeleteUserByID(ctx, uri.ID)
	if err != nil {
		log.Error(
			"删除用户:执行失败",
			zap.Error(err),
			zap.Uint32("user_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"删除用户:执行成功",
		zap.Uint32("user_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
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
func (h *UserHandler) GetUser(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "查询用户详情:绑定用户ID参数失败") {
		return
	}

	m, err := h.userSvc.FindUserByID(ctx, []string{"Role"}, uri.ID)
	if err != nil {
		log.Error(
			"查询用户:执行失败",
			zap.Error(err),
			zap.Uint32("user_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, &sysmodel.UserResp{
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
func (h *UserHandler) ListUser(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var req sysmodel.ListUserDTO
	if !common.ShouldBindQuery(c, log, &req, "查询用户列表:绑定查询用户列表请求参数失败") {
		return
	}

	page, size := req.StandardModelQuery.GetPageParam()
	total, ms, err := h.userSvc.ListUser(ctx, page, size, req)
	if err != nil {
		log.Error(
			"查询用户列表:执行失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_user_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mbs := sysmodel.ListUserModelToDetailOut(ms)
	c.JSON(http.StatusOK, &sysmodel.PagUserResp{
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
func (h *UserHandler) ResetPassword(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("重置用户密码:开始执行")

	var uri commodel.IDUri
	if !common.ShouldBindUri(c, log, &uri, "重置用户密码:绑定用户ID参数失败") {
		return
	}

	var req sysmodel.ResetPasswordDTO
	if !common.ShouldBind(c, log, &req, "重置用户密码:绑定重置用户密码请求参数失败") {
		return
	}

	err := h.userSvc.ResetPassword(ctx, uri.ID, req.NewPassword)
	if err != nil {
		log.Error(
			"重置用户密码:重置用户密码失败",
			zap.Error(err),
			zap.Uint32("user_id", uri.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"重置用户密码:执行成功",
		zap.Uint32("user_id", uri.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
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
func (h *UserHandler) PatchPassword(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("修改个人密码:开始执行")

	var req sysmodel.PatchPasswordDTO
	if !common.ShouldBind(c, log, &req, "修改个人密码:绑定修改个人密码请求参数失败") {
		return
	}

	claims := ctxutil.MustGetJwtClaims(ctx)
	err := h.userSvc.PatchPassword(ctx, claims.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		log.Error(
			"修改个人密码:执行失败",
			zap.Error(err),
			zap.Uint32("user_id", claims.UserID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"修改个人密码:执行成功",
		zap.Uint32("user_id", claims.UserID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	c.JSON(commodel.NoDataResp.Code, commodel.NoDataResp)
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
func (h *UserHandler) Login(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("用户登录:开始执行")

	var req sysmodel.LoginDTO
	if !common.ShouldBind(c, log, &req, "用户登录:绑定用户登录请求参数失败") {
		return
	}

	reqCtx := sysmodel.RequestContext{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	accessToken, refreshToken, err := h.userSvc.Login(ctx, req, reqCtx)
	if err != nil {
		log.Error(
			"用户登录:执行失败",
			zap.Error(err),
			zap.String("username", req.Username),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	log.Info(
		"用户登录:执行成功",
		zap.String("username", req.Username),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &sysmodel.LoginResp{
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
func (h *UserHandler) RefreshToken(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)
	log.Info("刷新令牌:开始执行")

	var req sysmodel.RefreshTokenDTO
	if !common.ShouldBind(c, log, &req, "刷新令牌:绑定刷新令牌请求参数失败") {
		return
	}

	accessToken, refreshToken, rErr := h.userSvc.RefreshTokens(ctx, req.RefreshToken)
	if rErr != nil {
		log.Error(
			"刷新令牌:刷新令牌失败",
			zap.Error(rErr),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, rErr)
		return
	}

	log.Info(
		"刷新令牌:刷新令牌成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, &sysmodel.LoginResp{
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
func (h *UserHandler) ListLoginRecord(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var req sysmodel.ListLoginRecordDTO
	if !common.ShouldBindQuery(c, log, &req, "查询用户登录记录列表:绑定查询用户登录记录列表请求参数失败") {
		return
	}

	page, size := req.BaseModelQuery.GetPageParam()
	total, ms, err := h.userSvc.ListLoginRecord(ctx, page, size, req)
	if err != nil {
		log.Error(
			"查询用户登录记录列表:查询用户登录记录列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_login_record_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mbs := sysmodel.ListLoginRecordModelToStandardOut(ms)
	c.JSON(http.StatusOK, &sysmodel.PagLoginRecordResp{
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
func (h *UserHandler) ListMeLoginRecord(c *gin.Context) {
	startTime := time.Now()
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.log, ctx)

	var req sysmodel.ListLoginRecordDTO
	if !common.ShouldBindQuery(c, log, &req, "查询个人登录记录列表:绑定查询个人登录记录列表请求参数失败") {
		return
	}

	claims := ctxutil.MustGetJwtClaims(ctx)
	req.Username = claims.Username
	page, size := req.BaseModelQuery.GetPageParam()
	total, ms, err := h.userSvc.ListLoginRecord(ctx, page, size, req)
	if err != nil {
		log.Error(
			"查询个人登录记录列表:查询个人登录记录列表失败",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.Object("list_login_record_dto", &req),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		errors.RespondWithError(c, err)
		return
	}

	mbs := sysmodel.ListLoginRecordModelToStandardOut(ms)
	c.JSON(http.StatusOK, &sysmodel.PagLoginRecordResp{
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
