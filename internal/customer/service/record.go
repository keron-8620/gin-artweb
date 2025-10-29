package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pb "gitee.com/keion8620/go-dango-gin/api/customer/record"
	"gitee.com/keion8620/go-dango-gin/internal/customer/biz"
	"gitee.com/keion8620/go-dango-gin/pkg/common"
	"gitee.com/keion8620/go-dango-gin/pkg/errors"
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
