package customer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	commodel "gin-artweb/internal/model/common"
	custmodel "gin-artweb/internal/model/customer"
	custsvc "gin-artweb/internal/service/customer"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type UserHandler struct {
	log     *zap.Logger
	svcUser *custsvc.UserService
}

func NewUserHandler(
	log *zap.Logger,
	svcUser *custsvc.UserService,
) *UserHandler {
	return &UserHandler{
		log:     log,
		svcUser: svcUser,
	}
}

// @Summary 新增用户
// @Description 本接口用于新增用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body custmodel.CreateUserRequest true "创建用户请求"
// @Success 201 {object} custmodel.UserReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user [post]
// @Security ApiKeyAuth
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var req custmodel.CreateUserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定创建用户请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始创建用户",
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcUser.CreateUser(ctx, custmodel.UserModel{
		Username: req.Username,
		Password: req.Password,
		IsActive: req.IsActive,
		IsStaff:  req.IsStaff,
		RoleID:   req.RoleID,
	})
	if err != nil {
		h.log.Error(
			"创建用户失败",
			zap.Error(err),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"创建用户成功",
		zap.Uint32(commodel.RequestIDKey, m.ID),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusCreated, &custmodel.UserReply{
		Code: http.StatusCreated,
		Data: custmodel.UserModelToDetailOut(*m),
	})
}

// @Summary 更新用户
// @Description 本接口用于更新指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Param request body custmodel.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} custmodel.UserReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{id} [put]
// @Security ApiKeyAuth
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定用户ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	var req custmodel.UpdateUserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定更新用户请求参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始更新用户",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.Object(commodel.RequestModelKey, &req),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcUser.UpdateUserByID(ctx, uri.ID, map[string]any{
		"username":  req.Username,
		"is_active": req.IsActive,
		"is_staff":  req.IsStaff,
		"role_id":   req.RoleID,
	}); err != nil {
		h.log.Error(
			"更新用户失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.Object(commodel.RequestModelKey, &req),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"更新用户成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcUser.FindUserByID(ctx, []string{"Role"}, uri.ID)
	if err != nil {
		h.log.Error(
			"查询更新后的用户信息失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, &custmodel.UserReply{
		Code: http.StatusOK,
		Data: custmodel.UserModelToDetailOut(*m),
	})
}

// @Summary 删除用户
// @Description 本接口用于删除指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Success 200 {object} commodel.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{id} [delete]
// @Security ApiKeyAuth
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定删除用户ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始删除用户",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcUser.DeleteUserByID(ctx, uri.ID); err != nil {
		h.log.Error(
			"删除用户失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"删除用户成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 查询用户
// @Description 本接口用于查询指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Success 200 {object} custmodel.UserReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{id} [get]
// @Security ApiKeyAuth
func (h *UserHandler) GetUser(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询用户ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询用户详情",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := h.svcUser.FindUserByID(ctx, []string{"Role"}, uri.ID)
	if err != nil {
		h.log.Error(
			"查询用户详情失败",
			zap.Error(err),
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询用户详情成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &custmodel.UserReply{
		Code: http.StatusOK,
		Data: custmodel.UserModelToDetailOut(*m),
	})
}

// @Summary 查询用户列表
// @Description 本接口用于查询用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request query custmodel.ListUserRequest false "查询参数"
// @Success 200 {object} custmodel.PagUserReply "成功返回用户列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user [get]
// @Security ApiKeyAuth
func (h *UserHandler) ListUser(ctx *gin.Context) {
	var req custmodel.ListUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询用户列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询用户列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount:  true,
		Size:     size,
		Page:     page,
		OrderBy:  []string{"id ASC"},
		Query:    query,
		Preloads: []string{"Role"},
	}
	total, ms, err := h.svcUser.ListUser(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询用户列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询用户列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := custmodel.ListUserModelToDetailOut(ms)
	ctx.JSON(http.StatusOK, &custmodel.PagUserReply{
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
// @Param request body custmodel.ResetPasswordRequest true "重置用户密码请求"
// @Success 200 {object} commodel.MapAPIReply "密码重置成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/password/{id} [patch]
// @Security ApiKeyAuth
func (h *UserHandler) ResetPassword(ctx *gin.Context) {
	var uri commodel.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		h.log.Error(
			"绑定查询用户ID参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}
	var req custmodel.ResetPasswordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定更新重置用户密码参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始重置用户密码",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := h.svcUser.UpdateUserByID(ctx, uri.ID, map[string]any{
		"password": req.NewPassword,
	}); err != nil {
		h.log.Error(
			"重置用户密码失败",
			zap.Uint32(commodel.RequestIDKey, uri.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}
	h.log.Info(
		"重置用户密码成功",
		zap.Uint32(commodel.RequestIDKey, uri.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 修改当前用户密码
// @Description 本接口用于修改当前登录用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body custmodel.PatchPasswordRequest true "修改用户密码请求"
// @Success 200 {object} commodel.MapAPIReply "密码修改成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 401 {object} errors.Error "认证失败"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/me/password [patch]
// @Security ApiKeyAuth
func (h *UserHandler) PatchPassword(ctx *gin.Context) {
	var req custmodel.PatchPasswordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定更新个人密码参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	claims, rErr := ctxutil.GetUserClaims(ctx)
	if rErr != nil {
		h.log.Error(
			"获取个人登录信息失败",
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	if rErr = h.svcUser.PatchPassword(ctx, claims.UserID, req.OldPassword, req.NewPassword); rErr != nil {
		h.log.Error(
			"修改用户密码失败",
			zap.Uint32(ctxutil.UserIDKey, claims.UserID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"修改用户密码成功",
		zap.Uint32(ctxutil.UserIDKey, claims.UserID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	ctx.JSON(commodel.NoDataReply.Code, commodel.NoDataReply)
}

// @Summary 登陆接口
// @Description 本接口用于登陆
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body custmodel.LoginRequest true "登陆请求参数"
// @Success 200 {object} custmodel.LoginReply "登录成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 401 {object} errors.Error "用户名或密码错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/login [post]
func (h *UserHandler) Login(ctx *gin.Context) {
	var req custmodel.LoginRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定用户登录参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始用户登录验证",
		zap.String("username", req.Username),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	accessToken, refreshToken, rErr := h.svcUser.Login(
		ctx,
		req.Username,
		req.Password,
		ctx.ClientIP(),
		ctx.Request.UserAgent(),
	)

	if rErr != nil {
		h.log.Error(
			"用户登录验证失败",
			zap.Error(rErr),
			zap.String("username", req.Username),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"用户登录成功",
		zap.String("username", req.Username),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &custmodel.LoginReply{
		Code: http.StatusOK,
		Data: custmodel.LoginOut{
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
// @Param request body custmodel.RefreshTokenRequest true "刷新令牌请求参数"
// @Success 200 {object} custmodel.LoginReply "刷新令牌成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 401 {object} errors.Error "用户名或密码错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/refresh/token [post]
func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	var req custmodel.RefreshTokenRequest
	if err := ctx.ShouldBind(&req); err != nil {
		h.log.Error(
			"绑定刷新令牌参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}
	accessToken, refreshToken, rErr := h.svcUser.RefreshTokens(ctx, req.RefreshToken)
	if rErr != nil {
		h.log.Error(
			"刷新令牌失败",
			zap.Error(rErr),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	ctx.JSON(http.StatusOK, &custmodel.LoginReply{
		Code: http.StatusOK,
		Data: custmodel.LoginOut{
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
// @Param request query custmodel.ListLoginRecordRequest false "查询参数"
// @Success 200 {object} custmodel.PagLoginRecordReply "成功返回用户登录记录列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/record/login [get]
// @Security ApiKeyAuth
func (h *UserHandler) ListLoginRecord(ctx *gin.Context) {
	var req custmodel.ListLoginRecordRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询用户登录记录列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"开始查询用户登录记录列表",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Size:    size,
		Page:    page,
		OrderBy: []string{"id DESC"},
		Query:   query,
	}
	total, ms, rErr := h.svcUser.ListLoginRecord(ctx, qp)
	if rErr != nil {
		h.log.Error(
			"查询用户登录记录列表失败",
			zap.Error(rErr),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}

	h.log.Info(
		"查询用户登录记录列表成功",
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := custmodel.ListLoginRecordModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &custmodel.PagLoginRecordReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

// @Summary 查询当前用户的登录记录列表
// @Description 本接口用于查询当前登录用户的登录记录列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request query custmodel.ListLoginRecordRequest false "查询参数"
// @Success 200 {object} custmodel.PagLoginRecordReply "成功返回用户登录记录列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 401 {object} errors.Error "未授权访问"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/me/record/login [get]
// @Security ApiKeyAuth
func (h *UserHandler) ListMeLoginRecord(ctx *gin.Context) {
	var req custmodel.ListLoginRecordRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		h.log.Error(
			"绑定查询个人登录记录列表参数失败",
			zap.Error(err),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return
	}

	claims, rErr := ctxutil.GetUserClaims(ctx)
	if rErr != nil {
		h.log.Error(
			"获取个人登录信息失败",
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, rErr)
		return
	}
	req.Username = claims.Subject

	h.log.Info(
		"开始查询个人登录记录列表",
		zap.Uint32(ctxutil.UserIDKey, claims.UserID),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount: true,
		Size:    size,
		Page:    page,
		OrderBy: []string{"id DESC"},
		Query:   query,
	}
	total, ms, err := h.svcUser.ListLoginRecord(ctx, qp)
	if err != nil {
		h.log.Error(
			"查询个人登录记录列表失败",
			zap.Error(err),
			zap.Uint32(ctxutil.UserIDKey, claims.UserID),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		errors.RespondWithError(ctx, err)
		return
	}

	h.log.Info(
		"查询个人登录记录列表成功",
		zap.Uint32(ctxutil.UserIDKey, claims.UserID),
		zap.String(commodel.RequestURIKey, ctx.Request.RequestURI),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	mbs := custmodel.ListLoginRecordModelToStandardOut(ms)
	ctx.JSON(http.StatusOK, &custmodel.PagLoginRecordReply{
		Code: http.StatusOK,
		Data: commodel.NewPag(page, size, total, mbs),
	})
}

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
