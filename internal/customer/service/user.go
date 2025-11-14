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
	"gin-artweb/pkg/auth"
	"gin-artweb/pkg/errors"
)

type UserService struct {
	log      *zap.Logger
	ucUser   *biz.UserUsecase
	ucRecord *biz.RecordUsecase
}

func NewUserService(
	log *zap.Logger,
	ucUser *biz.UserUsecase,
	ucRecord *biz.RecordUsecase,
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucUser.CreateUser(ctx, biz.UserModel{
		Username: req.Username,
		Password: req.Password,
		IsActive: req.IsActive,
		IsStaff:  req.IsStaff,
		RoleID:   req.RoleID,
	})
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusCreated, &pbUser.UserReply{
		Code: http.StatusCreated,
		Data: *UserModelToOut(*m),
	})
}

// @Summary 更新用户
// @Description 本接口用于更新指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param pk path uint true "用户编号"
// @Param request body pbUser.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} pbUser.UserReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{pk} [put]
// @Security ApiKeyAuth
func (s *UserService) UpdateUser(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pbUser.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucUser.UpdateUserByID(ctx, uri.PK, map[string]any{
		"username":  req.Username,
		"is_active": req.IsActive,
		"is_staff":  req.IsStaff,
		"role_id":   req.RoleID,
	}); err != nil {
		s.log.Error(err.Error())
		ctx.JSON(err.Code, err.Reply())
		return
	}
	m, err := s.ucUser.FindUserByID(ctx, []string{"Role"}, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, &pbUser.UserReply{
		Code: http.StatusOK,
		Data: *UserModelToOut(*m),
	})
}

// @Summary 删除用户
// @Description 本接口用于删除指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param pk path uint true "用户编号"
// @Success 200 {object} pbComm.MapAPIReply "删除成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{pk} [delete]
// @Security ApiKeyAuth
func (s *UserService) DeleteUser(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucUser.DeleteUserByID(ctx, uri.PK); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(pbComm.NoDataReply.Code, pbComm.NoDataReply)
}

// @Summary 查询用户
// @Description 本接口用于查询指定ID的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param pk path uint true "用户编号"
// @Success 200 {object} pbUser.UserReply "成功返回用户信息"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/{pk} [get]
// @Security ApiKeyAuth
func (s *UserService) GetUser(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucUser.FindUserByID(ctx, []string{"Role"}, uri.PK)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, &pbUser.UserReply{
		Code: http.StatusOK,
		Data: *UserModelToOut(*m),
	})
}

// @Summary 查询用户列表
// @Description 本接口用于查询用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1)
// @Param size query int false "每页数量" minimum(1) maximum(100)
// @Param username query string false "用户名"
// @Param is_active query bool false "是否激活"
// @Param is_staff query bool false "是否是工作人员"
// @Param role_id query uint false "角色ID"
// @Success 200 {object} pbUser.PagUserReply "成功返回用户列表"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user [get]
// @Security ApiKeyAuth
func (s *UserService) ListUser(ctx *gin.Context) {
	var req pbUser.ListUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	page, size, query := req.Query()
	total, ms, err := s.ucUser.ListUser(ctx, page, size, query, []string{"id"}, true, []string{"Role"})
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mbs := ListUserModelToOut(ms)
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
// @Param pk path uint true "用户编号"
// @Param request body pbUser.ResetPasswordRequest true "重置用户密码请求"
// @Success 200 {object} pbComm.MapAPIReply "密码重置成功"
// @Failure 400 {object} errors.Error "请求参数错误"
// @Failure 404 {object} errors.Error "用户未找到"
// @Failure 500 {object} errors.Error "服务器内部错误"
// @Router /api/v1/customer/user/password/{pk} [put]
// @Security ApiKeyAuth
func (s *UserService) ResetPassword(ctx *gin.Context) {
	var uri pbComm.PKUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pbUser.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		ctx.JSON(errors.ValidateError.Code, errors.ValidateError.Reply())
		return
	}
	if err := s.ucUser.UpdateUserByID(ctx, uri.PK, map[string]any{
		"password": req.NewPassword,
	}); err != nil {
		s.log.Error(err.Error())
		ctx.JSON(err.Code, err.Reply())
		return
	}
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
// @Router /api/v1/customer/me/password [put]
// @Security ApiKeyAuth
func (s *UserService) PatchPassword(ctx *gin.Context) {
	var req pbUser.PatchPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		ctx.JSON(errors.ValidateError.Code, errors.ValidateError.Reply())
		return
	}
	claims := auth.GetGinUserClaims(ctx)
	if claims == nil {
		ctx.JSON(auth.ErrGetUserClaims.Code, auth.ErrGetUserClaims.Reply())
		return
	}
	m, rErr := s.ucUser.FindUserByID(ctx, []string{"Role"}, claims.UserID)
	if rErr != nil {
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	ok, rErr := s.ucUser.VerifyPassword(req.OldPassword, m.Password)
	if rErr != nil {
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if !ok {
		ctx.JSON(biz.ErrPasswordMismatch.Code, biz.ErrPasswordMismatch.Reply())
	}
	if err := s.ucUser.UpdateUserByID(ctx, claims.UserID, map[string]any{
		"password": req.NewPassword,
	}); err != nil {
		s.log.Error(err.Error())
		ctx.JSON(err.Code, err.Reply())
		return
	}
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
// @Router /api/v1/customer/login [post]
func (s *UserService) Login(ctx *gin.Context) {
	var req pbUser.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	lrm := biz.LoginRecordModel{
		Username:  req.Username,
		LoginAt:   time.Now(),
		IPAddress: ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
		Status:    false,
	}
	token, rErr := s.ucUser.Login(ctx, req.Username, req.Password, lrm.IPAddress)
	if rErr != nil {
		s.ucRecord.CreateLoginRecord(ctx, lrm)
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	lrm.Status = true
	s.ucRecord.CreateLoginRecord(ctx, lrm)
	reply := pbUser.LoginReply{
		Code: http.StatusOK,
		Data: pbUser.LoginOut{Token: token},
	}
	ctx.JSON(reply.Code, &reply)
}

func (s *UserService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/user", s.CreateUser)
	r.PUT("/user/:pk", s.UpdateUser)
	r.DELETE("/user/:pk", s.DeleteUser)
	r.GET("/user/:pk", s.GetUser)
	r.GET("/user", s.ListUser)
	r.PATCH("/user/password/:pk", s.ResetPassword)
	r.PATCH("/me/password", s.PatchPassword)
	r.POST("/login", s.Login)
}

func UserModelToOutBase(
	m biz.UserModel,
) *pbUser.UserOutBase {
	return &pbUser.UserOutBase{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.String(),
		UpdatedAt: m.UpdatedAt.String(),
		Username:  m.Username,
		IsActive:  m.IsActive,
		IsStaff:   m.IsStaff,
	}
}

func UserModelToOut(
	m biz.UserModel,
) *pbUser.UserOut {
	var role *pbRole.RoleOutBase
	if m.Role.ID != 0 {
		role = RoleModelToOutBase(m.Role)
	}
	return &pbUser.UserOut{
		UserOutBase: *UserModelToOutBase(m),
		Role:        role,
	}
}

func ListUserModelToOut(
	ms []biz.UserModel,
) []*pbUser.UserOut {
	mso := make([]*pbUser.UserOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := UserModelToOut(m)
			mso = append(mso, mo)
		}
	}
	return mso
}
