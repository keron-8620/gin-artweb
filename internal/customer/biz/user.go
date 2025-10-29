package biz

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gitee.com/keion8620/go-dango-gin/pkg/auth"
	"gitee.com/keion8620/go-dango-gin/pkg/config"
	"gitee.com/keion8620/go-dango-gin/pkg/crypto"
	"gitee.com/keion8620/go-dango-gin/pkg/database"
	"gitee.com/keion8620/go-dango-gin/pkg/errors"
	"gitee.com/keion8620/go-dango-gin/pkg/utils"
)

type UserModel struct {
	database.StandardModel
	Username string    `gorm:"column:username;type:varchar(50);not null;uniqueIndex;comment:用户名" json:"username"`
	Password string    `gorm:"column:password;type:varchar(150);not null;comment:密码" json:"password"`
	IsActive bool      `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	IsStaff  bool      `gorm:"column:is_staff;type:boolean;comment:是否是工作人员" json:"is_staff"`
	RoleId   uint      `gorm:"column:role_id;foreignKey:RoleId;references:Id;not null;constraint:OnDelete:CASCADE;comment:角色" json:"role"`
	Role     RoleModel `gorm:"foreignKey:RoleId;constraint:OnDelete:CASCADE"`
}

func (m *UserModel) TableName() string {
	return "customer_user"
}

func (m *UserModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("username", m.Username)
	enc.AddBool("is_active", m.IsActive)
	enc.AddBool("is_staff", m.IsStaff)
	enc.AddUint("role_id", m.RoleId)
	return nil
}

type UserRepo interface {
	CreateModel(context.Context, *UserModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*UserModel, error)
	ListModel(context.Context, database.QueryParams) (int64, []UserModel, error)
}

type UserUsecase struct {
	log        *zap.Logger
	roleRepo   RoleRepo
	userRepo   UserRepo
	recordRepo RecordRepo
	hasher     crypto.Hasher
	conf       *config.SecurityConfig
}

func NewUserUsecase(
	log *zap.Logger,
	roleRepo RoleRepo,
	userRepo UserRepo,
	recordRepo RecordRepo,
	hasher crypto.Hasher,
	conf *config.SecurityConfig,
) *UserUsecase {
	return &UserUsecase{
		log:        log,
		roleRepo:   roleRepo,
		userRepo:   userRepo,
		recordRepo: recordRepo,
		hasher:     hasher,
		conf:       conf,
	}
}

func (uc *UserUsecase) GetRole(
	ctx context.Context,
	roleId uint,
) (*RoleModel, *errors.Error) {
	m, err := uc.roleRepo.FindModel(ctx, nil, roleId)
	if err != nil {
		rErr := database.NewGormError(err, map[string]any{"role_id": roleId})
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return m, nil
}

func (uc *UserUsecase) CreateUser(
	ctx context.Context,
	m UserModel,
) (*UserModel, *errors.Error) {
	if err := uc.CheckPasswordStrength(m.Password); err != nil {
		return nil, err
	}
	password, err := uc.HashPassword(m.Password)
	if err != nil {
		return nil, err
	}
	m.Password = password
	rm, err := uc.GetRole(ctx, m.RoleId)
	if err != nil {
		return nil, err
	}
	m.Role = *rm
	if err := uc.userRepo.CreateModel(ctx, &m); err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return &m, nil
}

func (uc *UserUsecase) UpdateUserById(
	ctx context.Context,
	userId uint,
	data map[string]any,
) *errors.Error {
	if password, exists := data["password"]; exists {
		if pwdStr, ok := password.(string); ok {
			if err := uc.CheckPasswordStrength(pwdStr); err != nil {
				return err
			}
			hashed, err := uc.HashPassword(pwdStr)
			if err != nil {
				return err
			}
			data["password"] = hashed
		} else {
			return ErrPasswordStrengthFailed
		}
	}
	if err := uc.userRepo.UpdateModel(ctx, data, "id = ?", userId); err != nil {
		rErr := database.NewGormError(err, data)
		uc.log.Error(rErr.Error())
		return rErr
	}
	return nil
}

func (uc *UserUsecase) DeleteUserById(
	ctx context.Context,
	userId uint,
) *errors.Error {
	if err := uc.userRepo.DeleteModel(ctx, userId); err != nil {
		rErr := database.NewGormError(err, map[string]any{"id": userId})
		uc.log.Error(rErr.Error())
		return rErr
	}
	return nil
}

func (uc *UserUsecase) FindUserById(
	ctx context.Context,
	preloads []string,
	userId uint,
) (*UserModel, *errors.Error) {
	m, err := uc.userRepo.FindModel(ctx, preloads, userId)
	if err != nil {
		rErr := database.NewGormError(err, map[string]any{"id": userId})
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return m, nil
}

func (uc *UserUsecase) FindUserByName(
	ctx context.Context,
	preloads []string,
	username string,
) (*UserModel, *errors.Error) {
	m, err := uc.userRepo.FindModel(ctx, preloads, map[string]any{"username": username})
	if err != nil {
		rErr := database.NewGormError(err, map[string]any{"username": username})
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return m, nil
}

func (uc *UserUsecase) ListUser(
	ctx context.Context,
	page, size int,
	query map[string]any,
	orderBy []string,
	isCount bool,
	preloads []string,
) (int64, []UserModel, *errors.Error) {
	qp := database.QueryParams{
		Preloads: preloads,
		Query:    query,
		OrderBy:  orderBy,
		Limit:    max(size, 0),
		Offset:   max(page-1, 0),
		IsCount:  isCount,
	}
	count, ms, err := uc.userRepo.ListModel(ctx, qp)
	if err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return 0, nil, rErr
	}
	return count, ms, nil
}

func (uc *UserUsecase) CheckPasswordStrength(pwd string) *errors.Error {
	if t := utils.GetPasswordStrength(pwd); t < utils.StrengthStrong {
		return ErrPasswordStrengthFailed
	}
	return nil
}

func (uc *UserUsecase) HashPassword(pwd string) (string, *errors.Error) {
	verified, err := uc.hasher.Hash(pwd)
	if err != nil {
		rErr := errors.FromError(err)
		uc.log.Error(rErr.Error())
		return "", rErr
	}
	return verified, nil
}

func (uc *UserUsecase) Login(
	ctx context.Context,
	username string,
	password string,
	ipAddress string,
) (string, *errors.Error) {
	num, err := uc.recordRepo.GetLoginFailNum(ctx, ipAddress)
	if err != nil {
		nErr := errors.FromCtxError(err)
		uc.log.Error(nErr.Error())
		return "", nErr
	}
	if num <= 0 {
		return "", ErrAccessLock
	}
	m, rErr := uc.FindUserByName(ctx, []string{"Role"}, username)
	if rErr != nil {
		return "", ErrInvalidCredentials
	}
	if !m.IsActive {
		return "", ErrUserInActive
	}
	verified, rErr := uc.VerifyPassword(password, m.Password)
	if rErr != nil {
		return "", rErr
	}
	if !verified {
		return "", ErrInvalidCredentials
	}
	claims := uc.NewUserClaims(*m)
	return uc.UserClaimsToToken(*claims)
}

func (uc *UserUsecase) VerifyPassword(pwd, hash string) (bool, *errors.Error) {
	verified, err := uc.hasher.Verify(pwd, hash)
	if err != nil {
		uc.log.Error(
			"password verification error",
			zap.Error(err),
		)
		return false, errors.FromError(err)
	}
	return verified, nil
}

func (uc *UserUsecase) NewUserClaims(m UserModel) *auth.UserClaims {
	return &auth.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   m.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(uc.conf.TokenExpireMinutes) * time.Minute)),
		},
		IsStaff: m.IsStaff,
		UserId:  m.Id,
		Role:    uc.roleRepo.RoleModelToSub(m.Role),
	}
}

func (uc *UserUsecase) UserClaimsToToken(claims auth.UserClaims) (string, *errors.Error) {
	token, err := auth.NewJWT([]byte(uc.conf.SecretKey), claims)
	if err != nil {
		rErr := errors.FromError(err)
		uc.log.Error(rErr.Error())
		return "", rErr
	}
	return token, nil
}
