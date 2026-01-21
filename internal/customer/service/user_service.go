package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbComm "gin-artweb/api/common"
	pbRole "gin-artweb/api/customer/role"
	pbUser "gin-artweb/api/customer/user"
	"gin-artweb/internal/customer/biz"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

type UserService struct {
	log      *zap.Logger
	ucUser   *biz.UserUsecase
	ucRecord *biz.LoginRecordUsecase
}

func NewUserService(
	log *zap.Logger,
	ucUser *biz.UserUsecase,
	ucRecord *biz.LoginRecordUsecase,
) *UserService {
	return &UserService{
		log:      log,
		ucUser:   ucUser,
		ucRecord: ucRecord,
	}
}

// @Summary 新增用户
// @Description 本接口用于新增用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body pbUser.CreateUserRequest true "创建用户请求"
// @Success 201 {object} pbUser.UserReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user [post]
// @Security ApiKeyAuth
func (s *UserService) CreateUser(ctx *gin.Context) {
	var req pbUser.CreateUserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定创建用户请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始创建用户",
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucUser.CreateUser(ctx, biz.UserModel{
		Username: req.Username,
		Password: req.Password,
		IsActive: req.IsActive,
		IsStaff:  req.IsStaff,
		RoleID:   req.RoleID,
	})
	if err != nil {
		s.log.Error(
			"创建用户失败",
			zap.Error(err),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"创建用户成功",
		zap.Uint32(pbComm.RequestIDKey, m.ID),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusCreated, &pbUser.UserReply{
		Code: http.StatusCreated,
		Data: UserModelToDetailOut(*m),
	})
}

// @Summary 更新用户
// @Description 本接口用于更新指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Param request body pbUser.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} pbUser.UserReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{id} [put]
// @Security ApiKeyAuth
func (s *UserService) UpdateUser(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定用户ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	var req pbUser.UpdateUserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新用户请求参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始更新用户",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.Object(pbComm.RequestModelKey, &req),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := s.ucUser.UpdateUserByID(ctx, uri.ID, map[string]any{
		"username":  req.Username,
		"is_active": req.IsActive,
		"is_staff":  req.IsStaff,
		"role_id":   req.RoleID,
	}); err != nil {
		s.log.Error(
			"更新用户失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.Object(pbComm.RequestModelKey, &req),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"更新用户成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucUser.FindUserByID(ctx, []string{"Role"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询更新后的用户信息失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}
	ctx.JSON(http.StatusOK, &pbUser.UserReply{
		Code: http.StatusOK,
		Data: UserModelToDetailOut(*m),
	})
}

// @Summary 删除用户
// @Description 本接口用于删除指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{id} [delete]
// @Security ApiKeyAuth
func (s *UserService) DeleteUser(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定删除用户ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始删除用户",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := s.ucUser.DeleteUserByID(ctx, uri.ID); err != nil {
		s.log.Error(
			"删除用户失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"删除用户成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询用户
// @Description 本接口用于查询指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Success 200 {object} pbUser.UserReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{id} [get]
// @Security ApiKeyAuth
func (s *UserService) GetUser(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询用户ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询用户详情",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := s.ucUser.FindUserByID(ctx, []string{"Role"}, uri.ID)
	if err != nil {
		s.log.Error(
			"查询用户详情失败",
			zap.Error(err),
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询用户详情成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	ctx.JSON(http.StatusOK, &pbUser.UserReply{
		Code: http.StatusOK,
		Data: UserModelToDetailOut(*m),
	})
}

// @Summary 查询用户列表
// @Description 本接口用于查询用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request query pbUser.ListUserRequest false "查询参数"
// @Success 200 {object} pbUser.PagUserReply "成功返回用户列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user [get]
// @Security ApiKeyAuth
func (s *UserService) ListUser(ctx *gin.Context) {
	var req pbUser.ListUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		s.log.Error(
			"绑定查询用户列表参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始查询用户列表",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	page, size, query := req.Query()
	qp := database.QueryParams{
		IsCount:  true,
		Limit:    size,
		Offset:   page,
		OrderBy:  []string{"id ASC"},
		Query:    query,
		Preloads: []string{"Role"},
	}
	total, ms, err := s.ucUser.ListUser(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询用户列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}

	s.log.Info(
		"查询用户列表成功",
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	mbs := ListUserModelToDetailOut(ms)
	ctx.JSON(http.StatusOK, &pbUser.PagUserReply{
		Code: http.StatusOK,
		Data: pbComm.NewPag(page, size, total, mbs),
	})
}

// @Summary 重置用户密码
// @Description 本接口用于重置指定ID的用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户编号"
// @Param request body pbUser.ResetPasswordRequest true "重置用户密码请求"
// @Success 200 {object} pbComm.MapAPIReply "密码重置成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/password/{id} [patch]
// @Security ApiKeyAuth
func (s *UserService) ResetPassword(ctx *gin.Context) {
	var uri pbComm.IDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		s.log.Error(
			"绑定查询用户ID参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}
	var req pbUser.ResetPasswordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新重置用户密码参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始重置用户密码",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := s.ucUser.UpdateUserByID(ctx, uri.ID, map[string]any{
		"password": req.NewPassword,
	}); err != nil {
		s.log.Error(
			"重置用户密码失败",
			zap.Uint32(pbComm.RequestIDKey, uri.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}
	s.log.Info(
		"重置用户密码成功",
		zap.Uint32(pbComm.RequestIDKey, uri.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 修改当前用户密码
// @Description 本接口用于修改当前登录用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body pbUser.PatchPasswordRequest true "修改用户密码请求"
// @Success 200 {object} pbComm.MapAPIReply "密码修改成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 401 {object} errors.Error "认证失败"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/me/password [patch]
// @Security ApiKeyAuth
func (s *UserService) PatchPassword(ctx *gin.Context) {
	var req pbUser.PatchPasswordRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定更新个人密码参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	claims := auth.GetUserClaims(ctx)
	if claims == nil {
		s.log.Error(
			"获取个人登录信息失败",
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(errors.ErrGetUserClaims.Code, errors.ErrGetUserClaims.ToMap())
		return
	}

	m, rErr := s.ucUser.FindUserByID(ctx, []string{"Role"}, claims.UserID)
	if rErr != nil {
		s.log.Error(
			"获取个人登录信息失败",
			zap.Error(rErr),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}
	ok, rErr := s.ucUser.VerifyPassword(ctx, req.OldPassword, m.Password)
	if rErr != nil {
		s.log.Error(
			"验证密码失败",
			zap.Error(rErr),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}
	if !ok {
		s.log.Error(
			"旧密码错误",
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(biz.ErrPasswordMismatch.Code, biz.ErrPasswordMismatch.ToMap())
	}
	if err := s.ucUser.UpdateUserByID(ctx, claims.UserID, map[string]any{
		"password": req.NewPassword,
	}); err != nil {
		s.log.Error(
			"重置用户密码失败",
			zap.Uint32(auth.UserIDKey, claims.UserID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		ctx.AbortWithStatusJSON(err.Code, err.ToMap())
		return
	}
	s.log.Info(
		"修改用户密码成功",
		zap.Uint32(auth.UserIDKey, claims.UserID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 登陆接口
// @Description 本接口用于登陆
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body pbUser.LoginRequest true "登陆请求参数"
// @Success 200 {object} pbUser.LoginReply "登录成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 401 {object} errors.Error "用户名或密码错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/login [post]
func (s *UserService) Login(ctx *gin.Context) {
	var req pbUser.LoginRequest
	if err := ctx.ShouldBind(&req); err != nil {
		s.log.Error(
			"绑定用户登录参数失败",
			zap.Error(err),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		rErr := errors.ValidateError.WithCause(err)
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	s.log.Info(
		"开始用户登录验证",
		zap.String("username", req.Username),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	lrm := biz.LoginRecordModel{
		Username:  req.Username,
		LoginAt:   time.Now(),
		IPAddress: ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
		Status:    false,
	}
	token, rErr := s.ucUser.Login(ctx, req.Username, req.Password, lrm.IPAddress)

	if rErr != nil {
		s.log.Warn(
			"用户登录验证失败",
			zap.Error(rErr),
			zap.String("username", req.Username),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		if _, lrErr := s.ucRecord.CreateLoginRecord(ctx, lrm); lrErr != nil {
			s.log.Error(
				"创建用户登录记录失败",
				zap.Error(lrErr),
				zap.Object("login_record", &lrm),
				zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
				zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
			)
		}
		ctx.AbortWithStatusJSON(rErr.Code, rErr.ToMap())
		return
	}

	lrm.Status = true
	if _, lrErr := s.ucRecord.CreateLoginRecord(ctx, lrm); lrErr != nil {
		s.log.Error(
			"创建用户登录记录失败",
			zap.Error(lrErr),
			zap.Object("login_record", &lrm),
			zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
	}

	s.log.Info(
		"用户登录成功",
		zap.String("username", req.Username),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	reply := pbUser.LoginReply{
		Code: http.StatusOK,
		Data: pbUser.LoginOut{Token: token},
	}

	s.log.Info(
		"登录响应已发送",
		zap.String("username", req.Username),
		zap.String(pbComm.RequestURIKey, ctx.Request.RequestURI),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	ctx.JSON(reply.Code, &reply)
}

func (s *UserService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/user", s.CreateUser)
	r.PUT("/user/:id", s.UpdateUser)
	r.DELETE("/user/:id", s.DeleteUser)
	r.GET("/user/:id", s.GetUser)
	r.GET("/user", s.ListUser)
	r.PATCH("/user/password/:id", s.ResetPassword)
}

func UserModelToBaseOut(
	m biz.UserModel,
) *pbUser.UserBaseOut {
	return &pbUser.UserBaseOut{
		ID:       m.ID,
		Username: m.Username,
		IsActive: m.IsActive,
		IsStaff:  m.IsStaff,
	}
}

func UserModelToStandardOut(
	m biz.UserModel,
) *pbUser.UserStandardOut {
	return &pbUser.UserStandardOut{
		UserBaseOut: *UserModelToBaseOut(m),
		CreatedAt:   m.CreatedAt.String(),
		UpdatedAt:   m.UpdatedAt.String(),
	}
}

func UserModelToDetailOut(
	m biz.UserModel,
) *pbUser.UserDetailOut {
	var role *pbRole.RoleBaseOut
	if m.Role.ID != 0 {
		role = RoleModelToBaseOut(m.Role)
	}
	return &pbUser.UserDetailOut{
		UserStandardOut: *UserModelToStandardOut(m),
		Role:            role,
	}
}

func ListUserModelToDetailOut(
	ums *[]biz.UserModel,
) *[]pbUser.UserDetailOut {
	if ums == nil {
		return &[]pbUser.UserDetailOut{}
	}
	ms := *ums
	mso := make([]pbUser.UserDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := UserModelToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
