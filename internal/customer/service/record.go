package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "gin-artweb/api/customer/record"
	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/auth"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/errors"
)

type RecordService struct {
	log      *zap.Logger
	ucRecord *biz.RecordUsecase
}

func NewRecordService(
	log *zap.Logger,
	ucRecord *biz.RecordUsecase,
) *RecordService {
	return &RecordService{
		log:      log,
		ucRecord: ucRecord,
	}
}

// @Summary 查询用户登录记录列表
// @Description 本接口用于查询用户登录记录列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param pk query uint false "用户主键，可选参数，如果提供则必须大于0"
// @Param pks query string false "用户主键列表，可选参数，多个用,隔开，如1,2,3"
// @Param username query string false "用户名"
// @Param ip_address query string false "ip地址"
// @Param before_login_at query string false "登录时间之前的记录 (RFC3339格式)"
// @Param after_login_at query string false "登录时间之后的记录 (RFC3339格式)"
// @Success 200 {object} pb.PagLoginRecordReply "成功返回用户登录记录列表"
// @Router /api/v1/customer/user/record/login [get]
func (s *RecordService) ListLoginRecord(ctx *gin.Context) {
	var req pb.ListLoginRecordRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	page, size, query := req.Query()
	total, ms, err := s.ucRecord.ListLoginRecord(ctx, page, size, query)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mbs := ListLoginRecordModelToOutBase(ms)
	ctx.JSON(http.StatusOK, &pb.PagLoginRecordReply{
		Code: http.StatusOK,
		Data: common.NewPag(page, size, total, mbs),
	})
}

// @Summary 查询个人登录记录列表
// @Description 本接口用于查询人登录记录列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param pk query uint false "用户主键，可选参数，如果提供则必须大于0"
// @Param pks query string false "用户主键列表，可选参数，多个用,隔开，如1,2,3"
// @Param ip_address query string false "ip地址"
// @Param before_login_at query string false "登录时间之前的记录 (RFC3339格式)"
// @Param after_login_at query string false "登录时间之后的记录 (RFC3339格式)"
// @Success 200 {object} pb.PagLoginRecordReply "成功返回用户登录记录列表"
// @Router /api/v1/customer/own/record/login [get]
func (s *RecordService) ListOwnLoginRecord(ctx *gin.Context) {
	var req pb.ListLoginRecordRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		rErr := errors.ValidateError.WithCause(err)
		s.log.Error(rErr.Error())
		ctx.JSON(rErr.Code, rErr.Reply())
		return
	}
	claims := auth.GetGinUserClaims(ctx)
	if claims == nil {
		ctx.JSON(auth.ErrGetUserClaims.Code, auth.ErrGetUserClaims.Reply())
		return
	}
	req.Username = claims.Subject
	page, size, query := req.Query()
	total, ms, err := s.ucRecord.ListLoginRecord(ctx, page, size, query)
	if err != nil {
		ctx.JSON(err.Code, err.Reply())
		return
	}
	mbs := ListLoginRecordModelToOutBase(ms)
	ctx.JSON(http.StatusOK, &pb.PagLoginRecordReply{
		Code: http.StatusOK,
		Data: common.NewPag(page, size, total, mbs),
	})
}

func (s *RecordService) LoadRouter(r *gin.RouterGroup) {
	r.GET("/user/record/login", s.ListLoginRecord)
	r.GET("/own/record/login", s.ListOwnLoginRecord)
}

func LoginRecordModelToOutBase(
	m biz.LoginRecordModel,
) *pb.LoginRecordOutBase {
	return &pb.LoginRecordOutBase{
		Id:        m.Id,
		Username:  m.Username,
		LoginAt:   m.LoginAt.String(),
		Status:    m.Status,
		IPAddress: m.IPAddress,
		UserAgent: m.UserAgent,
	}
}

func ListLoginRecordModelToOutBase(
	ms []biz.LoginRecordModel,
) []*pb.LoginRecordOutBase {
	mso := make([]*pb.LoginRecordOutBase, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := LoginRecordModelToOutBase(m)
			mso = append(mso, mo)
		}
	}
	return mso
}
