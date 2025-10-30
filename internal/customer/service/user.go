package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "gin-artweb/api/customer/user"
	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/auth"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/errors"
)

type UserService struct {
	log      *zap.Logger
	ucUser   *biz.UserUsecase
	ucRecord *biz.RecordUsecase
}

func NewUserService(
	log *zap.Logger,
	usUser *biz.UserUsecase,
	ucRecord *biz.RecordUsecase,
) *UserService {
	return &UserService{
		log:      log,
		ucUser:   usUser,
		ucRecord: ucRecord,
	}
}

// @Summary 新增用户
// @Description 本接口用于新增用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body pb.CreateUserRequest true "创建用户请求"
// @Success 200 {object} pb.UserReply "成功返回用户信息"
// @Router /api/v1/customer/user [post]
func (s *UserService) CreateUser(ctx *gin.Context) {
	var req pb.CreateUserRequest
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
		RoleId:   req.RoleId,
	})
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusCreated, &pb.UserReply{
		Code: http.StatusCreated,
		Data: *UserModelToOut(*m),
	})
}

// @Summary 更新用户
// @Description 本接口用于更新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param pk path uint true "用户编号"
// @Param request body pb.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} pb.UserReply "成功返回用户信息"
// @Router /api/v1/customer/user/{pk} [put]
func (s *UserService) UpdateUser(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pb.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucUser.UpdateUserById(ctx, uri.Pk, map[string]any{
		"username": req.Username,
		"IsActive": req.IsActive,
		"IsStaff":  req.IsStaff,
		"RoleId":   req.RoleId,
	}); err != nil {
		s.log.Error(err.Error())
		ctx.JSON(err.Code, err.Reply())
		return
	}
	m, err := s.ucUser.FindUserById(ctx, []string{"Role"}, uri.Pk)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, &pb.UserReply{
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
// @Router /api/v1/customer/user/{pk} [delete]
func (s *UserService) DeleteUser(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	if err := s.ucUser.DeleteUserById(ctx, uri.Pk); err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(common.NoDataReply.Code, common.NoDataReply)
}

// @Summary 查询单个用户
// @Description 本接口用于查询一个用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param pk path uint true "用户编号"
// @Success 200 {object} pb.UserReply "成功返回用户信息"
// @Router /api/v1/customer/user/{pk} [get]
func (s *UserService) GetUser(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	m, err := s.ucUser.FindUserById(ctx, []string{"Role"}, uri.Pk)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(http.StatusOK, &pb.UserReply{
		Code: http.StatusOK,
		Data: *UserModelToOut(*m),
	})
}

// @Summary 查询用户列表
// @Description 本接口用于查询用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param pk query uint false "用户主键，可选参数，如果提供则必须大于0"
// @Param pks query string false "用户主键列表，可选参数，多个用,隔开，如1,2,3"
// @Param before_create_at query string false "创建时间之前的记录 (RFC3339格式)"
// @Param after_create_at query string false "创建时间之后的记录 (RFC3339格式)"
// @Param before_update_at query string false "更新时间之前的记录 (RFC3339格式)"
// @Param after_update_at query string false "更新时间之后的记录 (RFC3339格式)"
// @Param username query string false "用户名"
// @Param is_active query bool false "是否激活"
// @Param is_staff query bool false "是否是工作人员"
// @Param role_id query uint false "角色ID"
// @Success 200 {object} pb.PagUserReply "成功返回用户列表"
// @Router /api/v1/customer/user [get]
func (s *UserService) ListUser(ctx *gin.Context) {
	var req pb.ListUserRequest
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
	ctx.JSON(http.StatusOK, &pb.PagUserReply{
		Code: http.StatusOK,
		Data: common.NewPag(page, size, total, mbs),
	})
}

// @Summary 重置用户密码
// @Description 本接口用于重置用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param pk path uint true "用户编号"
// @Param request body pb.ResetPasswordRequest true "重置用户密码请求"
// @Router /api/v1/customer/user/password/{pk} [put]
func (s *UserService) ResetPassword(ctx *gin.Context) {
	var uri PkUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	var req pb.ResetPasswordRequest
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
	if err := s.ucUser.UpdateUserById(ctx, uri.Pk, map[string]any{
		"password": req.NewPassword,
	}); err != nil {
		s.log.Error(err.Error())
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(common.NoDataReply.Code, common.NoDataReply)
}

// @Summary 修改用户密码
// @Description 本接口用于修改用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param pk path uint true "用户编号"
// @Param request body pb.PatchPasswordRequest true "修改用户密码请求"
// @Router /api/v1/customer/own/password [put]
func (s *UserService) PatchPassword(ctx *gin.Context) {
	var req pb.PatchPasswordRequest
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
	m, rErr := s.ucUser.FindUserById(ctx, []string{"Role"}, claims.UserId)
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
	if err := s.ucUser.UpdateUserById(ctx, claims.UserId, map[string]any{
		"password": req.NewPassword,
	}); err != nil {
		s.log.Error(err.Error())
		ctx.JSON(err.Code, err.Reply())
		return
	}
	ctx.JSON(common.NoDataReply.Code, common.NoDataReply)
}

// @Summary 登陆接口
// @Description 本接口用于登陆
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body pb.LoginRequest true "登陆请求参数"
// @Success 200 {object} pb.LoginReply "成功返回用户信息"
// @Router /api/v1/customer/own/login [put]
func (s *UserService) Login(ctx *gin.Context) {
	var req pb.LoginRequest
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
		ctx.JSON(rErr.Code, rErr.Reply())
	} else {
		lrm.Status = true
		reply := pb.LoginReply{
			Code: http.StatusOK,
			Data: pb.LoginOut{Token: token},
		}
		ctx.JSON(reply.Code, &reply)
	}
	s.ucRecord.CreateLoginRecord(ctx, lrm)
}

func (s *UserService) LoadRouter(r *gin.RouterGroup) {
	r.POST("/user", s.CreateUser)
	r.PUT("/user/:pk", s.UpdateUser)
	r.DELETE("/user/:pk", s.DeleteUser)
	r.GET("/user/:pk", s.GetUser)
	r.GET("/user", s.ListUser)
	r.PATCH("/user/password/:pk", s.ResetPassword)
	r.PATCH("/own/password", s.PatchPassword)
	r.POST("/own/login", s.Login)
}

func UserModelToOutBase(
	m biz.UserModel,
) *pb.UserOutBase {
	return &pb.UserOutBase{
		Id:        m.Id,
		CreatedAt: m.CreatedAt.String(),
		UpdatedAt: m.UpdatedAt.String(),
		Username:  m.Username,
		IsActive:  m.IsActive,
		IsStaff:   m.IsStaff,
	}
}

func UserModelToOut(
	m biz.UserModel,
) *pb.UserOut {
	return &pb.UserOut{
		UserOutBase: *UserModelToOutBase(m),
		Role:        RoleModelToOutBase(m.Role),
	}
}

func ListUserModelToOut(
	ms []biz.UserModel,
) []*pb.UserOut {
	mso := make([]*pb.UserOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := UserModelToOut(m)
			mso = append(mso, mo)
		}
	}
	return mso
}
