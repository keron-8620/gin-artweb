package biz

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/crypto"
	"gin-artweb/pkg/ctxutil"
)

const (
	UserTableName = "customer_user"
	UserIDKey     = "user_id"
	UsernameKey   = "username"
)

type UserModel struct {
	database.StandardModel
	Username string    `gorm:"column:username;type:varchar(50);not null;uniqueIndex;comment:用户名" json:"username"`
	Password string    `gorm:"column:password;type:varchar(150);not null;comment:密码" json:"password"`
	IsActive bool      `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	IsStaff  bool      `gorm:"column:is_staff;type:boolean;comment:是否是工作人员" json:"is_staff"`
	RoleID   uint32    `gorm:"column:role_id;not null;comment:角色ID" json:"role_id"`
	Role     RoleModel `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:CASCADE" json:"role"`
}

func (m *UserModel) TableName() string {
	return UserTableName
}

func (m *UserModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return errors.GormModelIsNil(UserTableName)
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("username", m.Username)
	enc.AddBool("is_active", m.IsActive)
	enc.AddBool("is_staff", m.IsStaff)
	enc.AddUint32("role_id", m.RoleID)
	return nil
}

type UserRepo interface {
	CreateModel(context.Context, *UserModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*UserModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]UserModel, error)
}

type UserUsecase struct {
	log        *zap.Logger
	roleRepo   RoleRepo
	userRepo   UserRepo
	recordRepo LoginRecordRepo
	hasher     crypto.Hasher
	conf       *config.SecurityConfig
}

func NewUserUsecase(
	log *zap.Logger,
	roleRepo RoleRepo,
	userRepo UserRepo,
	recordRepo LoginRecordRepo,
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
	roleID uint32,
) (*RoleModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询用户关联的角色",
		zap.Uint32(RoleIDKey, roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.roleRepo.FindModel(ctx, nil, roleID)
	if err != nil {
		uc.log.Error(
			"查询用户关联的角色失败",
			zap.Error(err),
			zap.Uint32(RoleIDKey, roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"role_id": roleID})
	}

	uc.log.Info(
		"查询用户关联的角色成功",
		zap.Uint32(RoleIDKey, roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *UserUsecase) CreateUser(
	ctx context.Context,
	m UserModel,
) (*UserModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建用户",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查密码强度
	if err := uc.CheckPasswordStrength(ctx, m.Password); err != nil {
		return nil, err
	}

	// 密码哈希
	password, err := uc.HashPassword(ctx, m.Password)
	if err != nil {
		return nil, err
	}
	m.Password = password

	// 获取角色信息
	rm, err := uc.GetRole(ctx, m.RoleID)
	if err != nil {
		return nil, err
	}
	m.Role = *rm

	// 创建用户
	if err := uc.userRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建用户失败",
			zap.Error(err),
			zap.String(UsernameKey, m.Username),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"创建用户成功",
		zap.String(UsernameKey, m.Username),
		zap.Uint32(UserIDKey, m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *UserUsecase) UpdateUserByID(
	ctx context.Context,
	userID uint32,
	data map[string]any,
) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始更新用户",
		zap.Uint32(UserIDKey, userID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 处理密码更新
	if password, exists := data["password"]; exists {
		if pwdStr, ok := password.(string); ok {
			uc.log.Info(
				"检测到密码更新，开始验证密码强度",
				zap.Uint32(UserIDKey, userID),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)

			if err := uc.CheckPasswordStrength(ctx, pwdStr); err != nil {
				uc.log.Warn(
					"密码强度不足",
					zap.Uint32(UserIDKey, userID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return err
			}

			hashed, err := uc.HashPassword(ctx, pwdStr)
			if err != nil {
				uc.log.Error(
					"密码哈希失败",
					zap.Error(err),
					zap.Uint32(UserIDKey, userID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return err
			}
			data["password"] = hashed

			uc.log.Info(
				"密码哈希处理完成",
				zap.Uint32(UserIDKey, userID),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
		} else {
			uc.log.Error(
				"密码格式错误",
				zap.Uint32(UserIDKey, userID),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			return ErrPasswordStrengthFailed
		}
	}

	// 更新用户信息
	if err := uc.userRepo.UpdateModel(ctx, data, "id = ?", userID); err != nil {
		uc.log.Error(
			"更新用户失败",
			zap.Error(err),
			zap.Uint32(UserIDKey, userID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, data)
	}

	uc.log.Info(
		"更新用户成功",
		zap.Uint32(UserIDKey, userID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *UserUsecase) DeleteUserByID(
	ctx context.Context,
	userID uint32,
) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除用户",
		zap.Uint32(UserIDKey, userID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.userRepo.DeleteModel(ctx, userID); err != nil {
		uc.log.Error(
			"删除用户失败",
			zap.Error(err),
			zap.Uint32(UserIDKey, userID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": userID})
	}

	uc.log.Info(
		"删除用户成功",
		zap.Uint32(UserIDKey, userID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *UserUsecase) FindUserByID(
	ctx context.Context,
	preloads []string,
	userID uint32,
) (*UserModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始根据ID查询用户",
		zap.Uint32(UserIDKey, userID),
		zap.Strings(database.PreloadKey, preloads),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.userRepo.FindModel(ctx, preloads, userID)
	if err != nil {
		uc.log.Error(
			"根据ID查询用户失败",
			zap.Error(err),
			zap.Uint32(UserIDKey, userID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": userID})
	}

	uc.log.Info(
		"根据ID查询用户成功",
		zap.Uint32(UserIDKey, userID),
		zap.String(UsernameKey, m.Username),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *UserUsecase) FindUserByName(
	ctx context.Context,
	preloads []string,
	username string,
) (*UserModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始根据用户名查询用户",
		zap.String(UsernameKey, username),
		zap.Strings(database.PreloadKey, preloads),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.userRepo.FindModel(ctx, preloads, map[string]any{"username": username})
	if err != nil {
		uc.log.Error(
			"根据用户名查询用户失败",
			zap.Error(err),
			zap.String(UsernameKey, username),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"username": username})
	}

	uc.log.Info(
		"根据用户名查询用户成功",
		zap.String(UsernameKey, username),
		zap.Uint32(UserIDKey, m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *UserUsecase) ListUser(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]UserModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询用户列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.userRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询用户列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询用户列表成功",
		zap.Int64("total_count", count),
		zap.Int("result_count", len(*ms)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *UserUsecase) CheckPasswordStrength(ctx context.Context, pwd string) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始检查密码强度",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	strength := GetPasswordStrength(pwd)
	if strength < uc.conf.Password.StrengthLevel {
		uc.log.Warn(
			"密码强度不足",
			zap.Int("password_strength", strength),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return ErrPasswordStrengthFailed
	}

	uc.log.Info(
		"密码强度检查通过",
		zap.Int("password_strength", strength),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *UserUsecase) HashPassword(ctx context.Context, pwd string) (string, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return "", errors.FromError(err)
	}

	uc.log.Info(
		"开始密码哈希处理",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	verified, err := uc.hasher.Hash(pwd)
	if err != nil {
		uc.log.Error(
			"密码哈希失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", errors.FromError(err)
	}

	uc.log.Info(
		"密码哈希处理完成",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return verified, nil
}

func (uc *UserUsecase) Login(
	ctx context.Context,
	username string,
	password string,
	ipAddress string,
) (string, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return "", errors.FromError(err)
	}

	uc.log.Info(
		"用户登录请求",
		zap.String(UsernameKey, username),
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查登录失败次数
	num, err := uc.recordRepo.GetLoginFailNum(ctx, ipAddress)
	if err != nil {
		uc.log.Error(
			"获取登录失败次数失败",
			zap.Error(err),
			zap.String(UsernameKey, username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", errors.FromError(err)
	}

	if num <= 0 {
		uc.log.Warn(
			"登录失败次数超限，账户被锁定",
			zap.String(UsernameKey, username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", ErrAccessLock
	}

	// 查找用户
	m, rErr := uc.FindUserByName(ctx, []string{"Role"}, username)
	if rErr != nil {
		uc.log.Warn(
			"用户不存在",
			zap.String(UsernameKey, username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)

		// 更新失败次数
		newNum := num - 1
		if err := uc.recordRepo.SetLoginFailNum(ctx, ipAddress, newNum); err != nil {
			uc.log.Error(
				"更新登录失败次数失败",
				zap.Error(err),
				zap.String(UsernameKey, username),
				zap.String("ip_address", ipAddress),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
		} else {
			uc.log.Info(
				"登录失败次数已更新",
				zap.String(UsernameKey, username),
				zap.String("ip_address", ipAddress),
				zap.Int("remaining_attempts", newNum),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
		}

		return "", ErrInvalidCredentials
	}

	// 检查用户状态
	if !m.IsActive {
		uc.log.Warn(
			"用户账户未激活",
			zap.String(UsernameKey, username),
			zap.Uint32(UserIDKey, m.ID),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", ErrUserInActive
	}

	// 验证密码
	verified, rErr := uc.VerifyPassword(ctx, password, m.Password)
	if rErr != nil {
		uc.log.Error(
			"密码验证过程出错",
			zap.Error(rErr),
			zap.String(UsernameKey, username),
			zap.Uint32(UserIDKey, m.ID),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", rErr
	}

	if !verified {
		uc.log.Warn(
			"密码验证失败",
			zap.String(UsernameKey, username),
			zap.Uint32(UserIDKey, m.ID),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)

		// 更新失败次数
		newNum := num - 1
		if err := uc.recordRepo.SetLoginFailNum(ctx, ipAddress, newNum); err != nil {
			uc.log.Error(
				"更新登录失败次数失败",
				zap.Error(err),
				zap.String(UsernameKey, username),
				zap.Uint32(UserIDKey, m.ID),
				zap.String("ip_address", ipAddress),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
		} else {
			uc.log.Info(
				"登录失败次数已更新",
				zap.String(UsernameKey, username),
				zap.Uint32(UserIDKey, m.ID),
				zap.String("ip_address", ipAddress),
				zap.Int("remaining_attempts", newNum),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
		}

		return "", ErrInvalidCredentials
	}

	// 重置失败次数
	if err := uc.recordRepo.SetLoginFailNum(ctx, ipAddress, uc.conf.Login.MaxFailedAttempts); err != nil {
		uc.log.Warn(
			"重置登录失败次数失败",
			zap.Error(err),
			zap.String(UsernameKey, username),
			zap.Uint32(UserIDKey, m.ID),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
	} else {
		uc.log.Info(
			"登录失败次数已重置",
			zap.String(UsernameKey, username),
			zap.Uint32(UserIDKey, m.ID),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
	}

	// 生成JWT token
	claims, rErr := uc.NewUserClaims(ctx, *m)
	if rErr != nil {
		return "", rErr
	}

	token, rErr := uc.UserClaimsToToken(ctx, *claims)
	if rErr != nil {
		uc.log.Error(
			"生成JWT token失败",
			zap.Error(rErr),
			zap.String(UsernameKey, username),
			zap.Uint32(UserIDKey, m.ID),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", rErr
	}

	uc.log.Info(
		"用户登录成功",
		zap.String(UsernameKey, username),
		zap.Uint32(UserIDKey, m.ID),
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return token, nil
}

func (uc *UserUsecase) VerifyPassword(ctx context.Context, pwd, hash string) (bool, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return false, errors.FromError(err)
	}

	uc.log.Info(
		"开始密码验证",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	verified, err := uc.hasher.Verify(pwd, hash)
	if err != nil {
		uc.log.Error(
			"密码验证失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return false, ErrPasswordMismatch
	}

	uc.log.Info(
		"密码验证完成",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return verified, nil
}

func (uc *UserUsecase) NewUserClaims(
	ctx context.Context,
	m UserModel,
) (*auth.UserClaims, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始生成用户JWT声明",
		zap.String(UsernameKey, m.Username),
		zap.Uint32(UserIDKey, m.ID),
		zap.Uint32(RoleIDKey, m.RoleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	claims := &auth.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   m.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(uc.conf.Token.ExpireMinutes) * time.Minute)),
		},
		IsStaff: m.IsStaff,
		UserID:  m.ID,
		RoleID:  m.RoleID,
	}

	uc.log.Info(
		"用户JWT声明生成完成",
		zap.String(UsernameKey, m.Username),
		zap.Uint32(UserIDKey, m.ID),
		zap.Uint32(RoleIDKey, m.RoleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return claims, nil
}

func (uc *UserUsecase) UserClaimsToToken(ctx context.Context, claims auth.UserClaims) (string, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return "", errors.FromError(err)
	}

	uc.log.Info(
		"生成JWT token",
		zap.String(UsernameKey, claims.Subject),
		zap.Uint32(UserIDKey, claims.UserID),
		zap.Uint32(RoleIDKey, claims.RoleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	token, err := auth.NewJWT([]byte(uc.conf.Token.SecretKey), claims)
	if err != nil {
		uc.log.Error(
			"生成JWT失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", errors.FromError(err)
	}

	uc.log.Info(
		"JWT token生成成功",
		zap.String(UsernameKey, claims.Subject),
		zap.Uint32(UserIDKey, claims.UserID),
		zap.Uint32(RoleIDKey, claims.RoleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return token, nil
}
