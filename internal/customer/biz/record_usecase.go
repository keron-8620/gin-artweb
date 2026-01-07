package biz

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

const LoginRecordTableName = "customer_login_record"

type LoginRecordModel struct {
	database.BaseModel
	Username  string    `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	LoginAt   time.Time `gorm:"column:login_at;autoCreateTime;comment:登录时间" json:"login_at"`
	IPAddress string    `gorm:"column:ip_address;type:varchar(108);comment:ip地址" json:"ip_address"`
	UserAgent string    `gorm:"column:user_agent;type:varchar(254);comment:客户端信息" json:"user_agent"`
	Status    bool      `gorm:"column:status;type:boolean;comment:是否登录成功" json:"status"`
}

func (m *LoginRecordModel) TableName() string {
	return LoginRecordTableName
}

func (m *LoginRecordModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return database.GormModelIsNil(LoginRecordTableName)
	}
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

type LoginRecordRepo interface {
	CreateModel(context.Context, *LoginRecordModel) error
	ListModel(context.Context, database.QueryParams) (int64, *[]LoginRecordModel, error)
	GetLoginFailNum(context.Context, string) (int, error)
	SetLoginFailNum(context.Context, string, int) error
}

type LoginRecordUsecase struct {
	log        *zap.Logger
	recordRepo LoginRecordRepo
}

func NewLoginRecordUsecase(
	log *zap.Logger,
	recordRepo LoginRecordRepo,
) *LoginRecordUsecase {
	return &LoginRecordUsecase{
		log:        log,
		recordRepo: recordRepo,
	}
}

func (uc *LoginRecordUsecase) CreateLoginRecord(
	ctx context.Context,
	m LoginRecordModel,
) (*LoginRecordModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建用户登录记录",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.recordRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建用户登录记录失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"创建用户登录记录成功",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *LoginRecordUsecase) ListLoginRecord(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]LoginRecordModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询用户登录记录列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	count, ms, err := uc.recordRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询用户登录记录列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询用户登录记录列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return count, ms, nil
}
