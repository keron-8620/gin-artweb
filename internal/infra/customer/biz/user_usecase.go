package biz

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/crypto"
)

const (
	UserTableName = "customer_user"
	UserIDKey     = "user_id"
	UsernameKey   = "username"

	LoginRecordTableName = "customer_login_record"
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
		return nil
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
	GetModel(context.Context, []string, ...any) (*UserModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]UserModel, error)
}

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
		return nil
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

type SecuritySettings struct {
	MaxFailedAttempts int `yaml:"max_failed_attempts"` // 最大登录失败次数
	LockMinutes       int `yaml:"lock_minutes"`        // 锁定时长(分钟)
	PasswordStrength  int `yaml:"password_strength"`   // 密码强度等级
}

type UserUsecase struct {
	log        *zap.Logger
	roleRepo   RoleRepo
	userRepo   UserRepo
	recordRepo LoginRecordRepo
	hasher     crypto.Hasher
	jwt        *auth.JWTConfig
	sec        SecuritySettings
}

func NewUserUsecase(
	log *zap.Logger,
	roleRepo RoleRepo,
	userRepo UserRepo,
	recordRepo LoginRecordRepo,
	hasher crypto.Hasher,
	jwt *auth.JWTConfig,
	sec SecuritySettings,
) *UserUsecase {
	return &UserUsecase{
		log:        log,
		roleRepo:   roleRepo,
		userRepo:   userRepo,
		recordRepo: recordRepo,
		hasher:     hasher,
		jwt:        jwt,
		sec:        sec,
	}
}

func (uc *UserUsecase) GetRole(
	ctx context.Context,
	roleID uint32,
) (*RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询用户关联的角色",
		zap.Uint32(RoleIDKey, roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.roleRepo.GetModel(ctx, nil, roleID)
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
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始创建用户",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查密码强度
	if err := uc.validatePasswordStrength(ctx, m.Password); err != nil {
		return nil, err
	}

	// 密码哈希
	password, err := uc.hashPassword(ctx, m.Password)
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
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
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

			if err := uc.validatePasswordStrength(ctx, pwdStr); err != nil {
				uc.log.Warn(
					"密码强度不足",
					zap.Uint32(UserIDKey, userID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return err
			}

			hashed, err := uc.hashPassword(ctx, pwdStr)
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
			uc.log.Warn(
				"密码不是字符串类型，已删除",
				zap.Any("password", password),
				zap.Uint32(UserIDKey, userID),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			delete(data, "password")
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
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
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
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始根据ID查询用户",
		zap.Uint32(UserIDKey, userID),
		zap.Strings(database.PreloadKey, preloads),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.userRepo.GetModel(ctx, preloads, userID)
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
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始根据用户名查询用户",
		zap.String(UsernameKey, username),
		zap.Strings(database.PreloadKey, preloads),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.userRepo.GetModel(ctx, preloads, map[string]any{"username": username})
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
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
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

func (uc *UserUsecase) ListLoginRecord(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]LoginRecordModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询用户登录记录列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.recordRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询用户登录记录列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询用户登录记录列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *UserUsecase) createLoginRecord(
	ctx context.Context,
	m LoginRecordModel,
) (*LoginRecordModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始创建用户登录记录",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.recordRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建用户登录记录失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"创建用户登录记录成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *UserUsecase) Login(
	ctx context.Context,
	username string,
	password string,
	ipAddress string,
	userAgent string,
) (string, string, *errors.Error) {
	if ctx.Err() != nil {
		return "", "", errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始刷新令牌",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	uc.log.Info(
		"用户登录请求",
		zap.String(UsernameKey, username),
		zap.String("ip_address", ipAddress),
		zap.String("user_agent", userAgent),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	lrm := LoginRecordModel{
		Username:  username,
		LoginAt:   time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Status:    false,
	}

	// 验证登录信息
	m, rErr := uc.validateLogin(ctx, username, password, ipAddress)
	if rErr != nil {
		uc.createLoginRecord(ctx, lrm)
		return "", "", rErr
	}

	// 登录认证成功
	lrm.Status = true
	if _, err := uc.createLoginRecord(ctx, lrm); err != nil {
		return "", "", err
	}
	uc.setLoginFailNum(ctx, ipAddress, uc.sec.MaxFailedAttempts)

	userinfo := auth.UserInfo{
		Username: username,
		UserID:   m.ID,
		RoleID:   m.RoleID,
		IsStaff:  m.IsStaff,
	}

	// 生成JWT token
	accessToken, rErr := uc.newAccessJWT(ctx, userinfo)
	if rErr != nil {
		return "", "", rErr
	}

	refreshToken, rErr := uc.newRefreshJWT(ctx, userinfo)
	if rErr != nil {
		return "", "", rErr
	}

	uc.log.Info(
		"用户登录成功",
		zap.String(UsernameKey, username),
		zap.Uint32(UserIDKey, m.ID),
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return accessToken, refreshToken, nil
}

func (uc *UserUsecase) validateLogin(
	ctx context.Context,
	username string,
	password string,
	ipAddress string,
) (*UserModel, *errors.Error) {
	uc.log.Info(
		"开始验证用户登录信息",
		zap.String(UsernameKey, username),
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查登录失败次数
	uc.log.Debug(
		"检查登录失败次数",
		zap.String(UsernameKey, username),
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	num, rErr := uc.getLoginFailNum(ctx, username, ipAddress)
	if rErr != nil {
		uc.log.Error(
			"获取登录失败次数失败",
			zap.Error(rErr),
			zap.String(UsernameKey, username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, rErr
	}

	uc.log.Debug(
		"获取登录失败次数成功",
		zap.String(UsernameKey, username),
		zap.String("ip_address", ipAddress),
		zap.Int("remaining_attempts", num),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if num == 0 {
		uc.log.Warn(
			"登录尝试次数用尽，账户被锁定",
			zap.String(UsernameKey, username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.ErrAccountLocked
	}

	// 查找用户
	uc.log.Debug(
		"开始查找用户",
		zap.String(UsernameKey, username),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := uc.FindUserByName(ctx, []string{"Role"}, username)
	if rErr != nil {
		uc.log.Warn(
			"用户不存在或查找失败",
			zap.Error(rErr),
			zap.String(UsernameKey, username),
			zap.String("ip_address", ipAddress),
			zap.Int("remaining_attempts", num-1),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		// 更新失败次数
		uc.setLoginFailNum(ctx, ipAddress, num-1)
		return nil, errors.ErrAuthFailed
	}

	uc.log.Debug(
		"用户查找成功",
		zap.String(UsernameKey, username),
		zap.Uint32(UserIDKey, m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查用户状态
	if !m.IsActive {
		uc.log.Warn(
			"用户账户被锁定",
			zap.String(UsernameKey, username),
			zap.Uint32(UserIDKey, m.ID),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.ErrAccountLocked
	}

	// 验证密码
	uc.log.Debug(
		"开始验证用户密码",
		zap.String(UsernameKey, username),
		zap.Uint32(UserIDKey, m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if rErr = uc.verifyPassword(ctx, password, m.Password); rErr != nil {
		uc.log.Warn(
			"用户密码验证失败",
			zap.Error(rErr),
			zap.String(UsernameKey, username),
			zap.Uint32(UserIDKey, m.ID),
			zap.String("ip_address", ipAddress),
			zap.Int("remaining_attempts", num-1),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		uc.setLoginFailNum(ctx, ipAddress, num-1)
		return nil, rErr.WithField("remaining_attempts", num-1)
	}

	uc.log.Info(
		"用户登录验证成功",
		zap.String(UsernameKey, username),
		zap.Uint32(UserIDKey, m.ID),
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	return m, nil
}

func (uc *UserUsecase) getLoginFailNum(ctx context.Context, username string, ipAddress string) (int, *errors.Error) {
	if ctx.Err() != nil {
		return 0, errors.FromError(ctx.Err())
	}
	num, err := uc.recordRepo.GetLoginFailNum(ctx, ipAddress)
	if err != nil {
		uc.log.Error(
			"获取登录失败次数失败",
			zap.Error(err),
			zap.String(UsernameKey, username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, errors.FromError(err)
	}

	if num <= 0 {
		uc.log.Warn(
			"登录失败次数超限，账户被锁定",
			zap.String(UsernameKey, username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil
	}
	return num, nil
}

func (uc *UserUsecase) setLoginFailNum(
	ctx context.Context,
	ipAddress string,
	num int,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	if err := uc.recordRepo.SetLoginFailNum(ctx, ipAddress, num); err != nil {
		uc.log.Warn(
			"重置登录失败次数失败",
			zap.Error(err),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}
	uc.log.Info(
		"登录失败次数已重置",
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *UserUsecase) newAccessJWT(ctx context.Context, ui auth.UserInfo) (string, *errors.Error) {
	if ctx.Err() != nil {
		return "", errors.FromError(ctx.Err())
	}
	token, err := auth.NewAccessJWT(ctx, uc.jwt, ui)
	if err != nil {
		uc.log.Error(
			"生成JWT token失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", errors.FromError(err)
	}
	return token, nil
}

func (uc *UserUsecase) newRefreshJWT(ctx context.Context, ui auth.UserInfo) (string, *errors.Error) {
	if ctx.Err() != nil {
		return "", errors.FromError(ctx.Err())
	}
	token, err := auth.NewRefreshJWT(ctx, uc.jwt, ui)
	if err != nil {
		uc.log.Error(
			"生成JWT token失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", errors.FromError(err)
	}
	return token, nil
}

func (uc *UserUsecase) verifyPassword(ctx context.Context, pwd, hash string) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始密码验证",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	verified, err := uc.hasher.Verify(ctx, pwd, hash)
	if err != nil {
		uc.log.Error(
			"密码验证过程中发生错误",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrAuthFailed
	}

	if !verified {
		uc.log.Warn(
			"密码验证失败",
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrAuthFailed
	}

	uc.log.Info(
		"密码验证通过",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *UserUsecase) hashPassword(ctx context.Context, pwd string) (string, *errors.Error) {
	if ctx.Err() != nil {
		return "", errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始密码哈希处理",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	verified, err := uc.hasher.Hash(ctx, pwd)
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

func (uc *UserUsecase) validatePasswordStrength(ctx context.Context, pwd string) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始检查密码强度",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	strength := GetPasswordStrength(pwd)
	if strength < uc.sec.PasswordStrength {
		uc.log.Warn(
			"密码强度不足",
			zap.Int("password_strength", strength),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrPasswordStrengthFailed
	}

	uc.log.Info(
		"密码强度检查通过",
		zap.Int("password_strength", strength),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *UserUsecase) PatchPassword(
	ctx context.Context,
	userID uint32,
	oldPassword string,
	newPassword string,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	// 检查旧密码是否正确
	m, rErr := uc.FindUserByID(ctx, []string{"Role"}, userID)
	if rErr != nil {
		uc.log.Error(
			"获取用户信息失败",
			zap.Error(rErr),
			zap.Uint32(UserIDKey, userID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return rErr
	}
	if rErr = uc.verifyPassword(ctx, oldPassword, m.Password); rErr != nil {
		uc.log.Error(
			"旧密码验证失败",
			zap.Error(rErr),
			zap.Uint32(UserIDKey, userID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return rErr
	}
	return uc.UpdateUserByID(ctx, userID, map[string]any{"password": newPassword})
}

func (uc *UserUsecase) RefreshTokens(
	ctx context.Context,
	refresh string,
) (string, string, *errors.Error) {
	if ctx.Err() != nil {
		return "", "", errors.FromError(ctx.Err())
	}

	var (
		accessToken  string
		refreshToken string
		rErr         *errors.Error
	)

	claims, err := auth.ParseRefreshToken(ctx, uc.jwt, refresh)
	if err != nil {
		uc.log.Error(
			"解析刷新令牌失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", "", errors.ErrTokenInvalid.WithCause(err)
	}
	accessToken, rErr = uc.newAccessJWT(ctx, claims.UserInfo)
	if rErr != nil {
		return "", "", rErr
	}
	refreshToken, rErr = uc.newRefreshJWT(ctx, claims.UserInfo)
	if rErr != nil {
		return "", "", rErr
	}
	return accessToken, refreshToken, nil
}
