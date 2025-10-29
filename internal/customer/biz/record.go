package biz

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gitee.com/keion8620/go-dango-gin/pkg/database"
	"gitee.com/keion8620/go-dango-gin/pkg/errors"
)

var (
	ErrAccessLock = errors.New(
		http.StatusUnauthorized,
		"account_locked_too_many_attempts",
		"因登录失败次数过多，账户已被锁定",
		nil,
	)
)

type LoginRecordModel struct {
	database.BaseModel
	Username  string    `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	LoginAt   time.Time `gorm:"column:login_at;autoCreateTime;comment:登录时间" json:"login_at"`
	IPAddress string    `gorm:"column:ip_address;type:varchar(108);comment:ip地址" json:"ip_address"`
	UserAgent string    `gorm:"column:user_agent;type:varchar(254);comment:客户端信息" json:"user_agent"`
	Status    bool      `gorm:"column:status;type:boolean;comment:是否登录成功" json:"status"`
}

func (m *LoginRecordModel) TableName() string {
	return "customer_login_record"
}

func (m *LoginRecordModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.BaseModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("username", m.Username)
	enc.AddTime("login_at", m.LoginAt)
	enc.AddString("ip_address", m.IPAddress)
	enc.AddString("user_agent", m.UserAgent)
	enc.AddBool("status", m.Status)
	return nil
}

type RecordRepo interface {
	CreateModel(context.Context, *LoginRecordModel) error
	ListModel(context.Context, database.QueryParams) (int64, []LoginRecordModel, error)
	GetLoginFailNum(context.Context, string) (int, error)
	SetLoginFailNum(context.Context, string, int) error
}

type RecordUsecase struct {
	log        *zap.Logger
	recordRepo RecordRepo
}

func NewRecordUsecase(
	log *zap.Logger,
	recordRepo RecordRepo,
) *RecordUsecase {
	return &RecordUsecase{
		log:        log,
		recordRepo: recordRepo,
	}
}

func (uc *RecordUsecase) CreateLoginRecord(
	ctx context.Context,
	m LoginRecordModel,
) (*LoginRecordModel, *errors.Error) {
	if err := uc.recordRepo.CreateModel(ctx, &m); err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return &m, nil
}

func (uc *RecordUsecase) ListLoginRecord(
	ctx context.Context,
	page, size int,
	query map[string]any,
) (int64, []LoginRecordModel, *errors.Error) {
	qp := database.QueryParams{
		Preloads: []string{},
		Query:    query,
		OrderBy:  []string{"id"},
		Limit:    max(size, 0),
		Offset:   max(page-1, 0),
		IsCount:  true,
	}
	count, ms, err := uc.recordRepo.ListModel(ctx, qp)
	if err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return 0, nil, rErr
	}
	return count, ms, nil
}
