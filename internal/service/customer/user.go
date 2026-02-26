package customer

import (
	"context"
	"time"

	"go.uber.org/zap"

	custmodel "gin-artweb/internal/model/customer"
	custsvc "gin-artweb/internal/repository/customer"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/crypto"
)

type SecuritySettings struct {
	MaxFailedAttempts int           `yaml:"max_failed_attempts"` // 最大登录失败次数
	LockDuration      time.Duration `yaml:"lock_minutes"`        // 锁定时长(分钟)
	PasswordStrength  int           `yaml:"password_strength"`   // 密码强度等级
}

type UserService struct {
	log        *zap.Logger
	roleRepo   *custsvc.RoleRepo
	userRepo   *custsvc.UserRepo
	recordRepo *custsvc.LoginRecordRepo
	hasher     crypto.Hasher
	jwt        *auth.JWTConfig
	sec        SecuritySettings
}

func NewUserService(
	log *zap.Logger,
	roleRepo *custsvc.RoleRepo,
	userRepo *custsvc.UserRepo,
	recordRepo *custsvc.LoginRecordRepo,
	hasher crypto.Hasher,
	jwt *auth.JWTConfig,
	sec SecuritySettings,
) *UserService {
	return &UserService{
		log:        log,
		roleRepo:   roleRepo,
		userRepo:   userRepo,
		recordRepo: recordRepo,
		hasher:     hasher,
		jwt:        jwt,
		sec:        sec,
	}
}

func (s *UserService) GetRole(
	ctx context.Context,
	roleID uint32,
) (*custmodel.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询用户关联的角色",
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.roleRepo.GetModel(ctx, nil, roleID)
	if err != nil {
		s.log.Error(
			"查询用户关联的角色失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"role_id": roleID})
	}

	s.log.Info(
		"查询用户关联的角色成功",
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *UserService) CreateUser(
	ctx context.Context,
	m custmodel.UserModel,
) (*custmodel.UserModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建用户",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查密码强度
	if err := s.validatePasswordStrength(ctx, m.Password); err != nil {
		return nil, err
	}

	// 密码哈希
	if password, err := s.hashPassword(ctx, m.Password); err != nil {
		return nil, err
	} else {
		m.Password = password
	}

	// 获取角色信息
	if rm, err := s.GetRole(ctx, m.RoleID); err != nil {
		return nil, err
	} else {
		m.Role = *rm
	}

	// 创建用户
	if err := s.userRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建用户失败",
			zap.Error(err),
			zap.String("username", m.Username),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"创建用户成功",
		zap.String("username", m.Username),
		zap.Uint32("user_id", m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (s *UserService) UpdateUserByID(
	ctx context.Context,
	userID uint32,
	data map[string]any,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新用户",
		zap.Uint32("user_id", userID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 处理密码更新
	if password, exists := data["password"]; exists {
		if pwdStr, ok := password.(string); ok {
			s.log.Info(
				"检测到密码更新，开始验证密码强度",
				zap.Uint32("user_id", userID),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)

			if err := s.validatePasswordStrength(ctx, pwdStr); err != nil {
				s.log.Warn(
					"密码强度不足",
					zap.Uint32("user_id", userID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return err
			}

			hashed, err := s.hashPassword(ctx, pwdStr)
			if err != nil {
				s.log.Error(
					"密码哈希失败",
					zap.Error(err),
					zap.Uint32("user_id", userID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return err
			}
			data["password"] = hashed

			s.log.Info(
				"密码哈希处理完成",
				zap.Uint32("user_id", userID),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
		} else {
			s.log.Warn(
				"密码不是字符串类型，已删除",
				zap.Any("password", password),
				zap.Uint32("user_id", userID),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			delete(data, "password")
		}
	}

	// 更新用户信息
	if err := s.userRepo.UpdateModel(ctx, data, "id = ?", userID); err != nil {
		s.log.Error(
			"更新用户失败",
			zap.Error(err),
			zap.Uint32("user_id", userID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, data)
	}

	s.log.Info(
		"更新用户成功",
		zap.Uint32("user_id", userID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *UserService) DeleteUserByID(
	ctx context.Context,
	userID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除用户",
		zap.Uint32("user_id", userID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.userRepo.DeleteModel(ctx, userID); err != nil {
		s.log.Error(
			"删除用户失败",
			zap.Error(err),
			zap.Uint32("user_id", userID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": userID})
	}

	s.log.Info(
		"删除用户成功",
		zap.Uint32("user_id", userID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *UserService) FindUserByID(
	ctx context.Context,
	preloads []string,
	userID uint32,
) (*custmodel.UserModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始根据ID查询用户",
		zap.Uint32("user_id", userID),
		zap.Strings(database.PreloadKey, preloads),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.userRepo.GetModel(ctx, preloads, userID)
	if err != nil {
		s.log.Error(
			"根据ID查询用户失败",
			zap.Error(err),
			zap.Uint32("user_id", userID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": userID})
	}

	s.log.Info(
		"根据ID查询用户成功",
		zap.Uint32("user_id", userID),
		zap.String("username", m.Username),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *UserService) FindUserByName(
	ctx context.Context,
	preloads []string,
	username string,
) (*custmodel.UserModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始根据用户名查询用户",
		zap.String("username", username),
		zap.Strings(database.PreloadKey, preloads),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.userRepo.GetModel(ctx, preloads, map[string]any{"username": username})
	if err != nil {
		s.log.Error(
			"根据用户名查询用户失败",
			zap.Error(err),
			zap.String("username", username),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"username": username})
	}

	s.log.Info(
		"根据用户名查询用户成功",
		zap.String("username", username),
		zap.Uint32("user_id", m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *UserService) ListUser(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]custmodel.UserModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询用户列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.userRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询用户列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询用户列表成功",
		zap.Int64("total_count", count),
		zap.Int("result_count", len(*ms)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *UserService) ListLoginRecord(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]custmodel.LoginRecordModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询用户登录记录列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.recordRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询用户登录记录列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询用户登录记录列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *UserService) createLoginRecord(
	ctx context.Context,
	m custmodel.LoginRecordModel,
) (*custmodel.LoginRecordModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建用户登录记录",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.recordRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建用户登录记录失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"创建用户登录记录成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (s *UserService) Login(
	ctx context.Context,
	username string,
	password string,
	ipAddress string,
	userAgent string,
) (string, string, *errors.Error) {
	if ctx.Err() != nil {
		return "", "", errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始刷新令牌",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	s.log.Info(
		"用户登录请求",
		zap.String("username", username),
		zap.String("ip_address", ipAddress),
		zap.String("user_agent", userAgent),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	lrm := custmodel.LoginRecordModel{
		Username:  username,
		LoginAt:   time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Status:    false,
	}

	// 验证登录信息
	m, rErr := s.validateLogin(ctx, username, password, ipAddress)
	if rErr != nil {
		s.createLoginRecord(ctx, lrm)
		return "", "", rErr
	}

	// 登录认证成功
	lrm.Status = true
	if _, err := s.createLoginRecord(ctx, lrm); err != nil {
		return "", "", err
	}
	s.setLoginFailNum(ctx, ipAddress, s.sec.MaxFailedAttempts)

	userinfo := auth.UserInfo{
		Username: username,
		UserID:   m.ID,
		RoleID:   m.RoleID,
		IsStaff:  m.IsStaff,
	}

	// 生成JWT token
	accessToken, rErr := s.newAccessJWT(ctx, userinfo)
	if rErr != nil {
		return "", "", rErr
	}

	refreshToken, rErr := s.newRefreshJWT(ctx, userinfo)
	if rErr != nil {
		return "", "", rErr
	}

	s.log.Info(
		"用户登录成功",
		zap.String("username", username),
		zap.Uint32("user_id", m.ID),
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return accessToken, refreshToken, nil
}

func (s *UserService) validateLogin(
	ctx context.Context,
	username string,
	password string,
	ipAddress string,
) (*custmodel.UserModel, *errors.Error) {
	s.log.Info(
		"开始验证用户登录信息",
		zap.String("username", username),
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查登录失败次数
	s.log.Debug(
		"检查登录失败次数",
		zap.String("username", username),
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	num, rErr := s.getLoginFailNum(ctx, username, ipAddress)
	if rErr != nil {
		s.log.Error(
			"获取登录失败次数失败",
			zap.Error(rErr),
			zap.String("username", username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, rErr
	}

	s.log.Debug(
		"获取登录失败次数成功",
		zap.String("username", username),
		zap.String("ip_address", ipAddress),
		zap.Int("remaining_attempts", num),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if num == 0 {
		s.log.Warn(
			"登录尝试次数用尽，账户被锁定",
			zap.String("username", username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.ErrAccountLocked
	}

	// 查找用户
	s.log.Debug(
		"开始查找用户",
		zap.String("username", username),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := s.FindUserByName(ctx, []string{"Role"}, username)
	if rErr != nil {
		s.log.Warn(
			"用户不存在或查找失败",
			zap.Error(rErr),
			zap.String("username", username),
			zap.String("ip_address", ipAddress),
			zap.Int("remaining_attempts", num-1),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		// 更新失败次数
		s.setLoginFailNum(ctx, ipAddress, num-1)
		return nil, errors.ErrAuthFailed
	}

	s.log.Debug(
		"用户查找成功",
		zap.String("username", username),
		zap.Uint32("user_id", m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查用户状态
	if !m.IsActive {
		s.log.Warn(
			"用户账户被锁定",
			zap.String("username", username),
			zap.Uint32("user_id", m.ID),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.ErrAccountLocked
	}

	// 验证密码
	s.log.Debug(
		"开始验证用户密码",
		zap.String("username", username),
		zap.Uint32("user_id", m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if rErr = s.verifyPassword(ctx, password, m.Password); rErr != nil {
		s.log.Warn(
			"用户密码验证失败",
			zap.Error(rErr),
			zap.String("username", username),
			zap.Uint32("user_id", m.ID),
			zap.String("ip_address", ipAddress),
			zap.Int("remaining_attempts", num-1),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		s.setLoginFailNum(ctx, ipAddress, num-1)
		return nil, rErr.WithField("remaining_attempts", num-1)
	}

	s.log.Info(
		"用户登录验证成功",
		zap.String("username", username),
		zap.Uint32("user_id", m.ID),
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	return m, nil
}

func (s *UserService) getLoginFailNum(ctx context.Context, username string, ipAddress string) (int, *errors.Error) {
	if ctx.Err() != nil {
		return 0, errors.FromError(ctx.Err())
	}
	num, err := s.recordRepo.GetLoginFailNum(ctx, ipAddress)
	if err != nil {
		s.log.Error(
			"获取登录失败次数失败",
			zap.Error(err),
			zap.String("username", username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, errors.FromError(err)
	}

	if num <= 0 {
		s.log.Warn(
			"登录失败次数超限，账户被锁定",
			zap.String("username", username),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil
	}
	return num, nil
}

func (s *UserService) setLoginFailNum(
	ctx context.Context,
	ipAddress string,
	num int,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	if err := s.recordRepo.SetLoginFailNum(ctx, ipAddress, num); err != nil {
		s.log.Warn(
			"重置登录失败次数失败",
			zap.Error(err),
			zap.String("ip_address", ipAddress),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}
	s.log.Info(
		"登录失败次数已重置",
		zap.String("ip_address", ipAddress),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *UserService) newAccessJWT(ctx context.Context, ui auth.UserInfo) (string, *errors.Error) {
	if ctx.Err() != nil {
		return "", errors.FromError(ctx.Err())
	}
	token, err := auth.NewAccessJWT(ctx, s.jwt, ui)
	if err != nil {
		s.log.Error(
			"生成JWT token失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", errors.FromError(err)
	}
	return token, nil
}

func (s *UserService) newRefreshJWT(ctx context.Context, ui auth.UserInfo) (string, *errors.Error) {
	if ctx.Err() != nil {
		return "", errors.FromError(ctx.Err())
	}
	token, err := auth.NewRefreshJWT(ctx, s.jwt, ui)
	if err != nil {
		s.log.Error(
			"生成JWT token失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", errors.FromError(err)
	}
	return token, nil
}

func (s *UserService) verifyPassword(ctx context.Context, pwd, hash string) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始密码验证",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	verified, err := s.hasher.Verify(ctx, pwd, hash)
	if err != nil {
		s.log.Error(
			"密码验证过程中发生错误",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrAuthFailed
	}

	if !verified {
		s.log.Warn(
			"密码验证失败",
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrAuthFailed
	}

	s.log.Info(
		"密码验证通过",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *UserService) hashPassword(ctx context.Context, pwd string) (string, *errors.Error) {
	if ctx.Err() != nil {
		return "", errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始密码哈希处理",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	verified, err := s.hasher.Hash(ctx, pwd)
	if err != nil {
		s.log.Error(
			"密码哈希失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", errors.FromError(err)
	}

	s.log.Info(
		"密码哈希处理完成",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return verified, nil
}

func (s *UserService) validatePasswordStrength(ctx context.Context, pwd string) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始检查密码强度",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	strength := GetPasswordStrength(pwd)
	if strength < s.sec.PasswordStrength {
		s.log.Warn(
			"密码强度不足",
			zap.Int("password_strength", strength),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrPasswordStrengthFailed
	}

	s.log.Info(
		"密码强度检查通过",
		zap.Int("password_strength", strength),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *UserService) PatchPassword(
	ctx context.Context,
	userID uint32,
	oldPassword string,
	newPassword string,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	// 检查旧密码是否正确
	m, rErr := s.FindUserByID(ctx, []string{"Role"}, userID)
	if rErr != nil {
		s.log.Error(
			"获取用户信息失败",
			zap.Error(rErr),
			zap.Uint32("user_id", userID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return rErr
	}
	if rErr = s.verifyPassword(ctx, oldPassword, m.Password); rErr != nil {
		s.log.Error(
			"旧密码验证失败",
			zap.Error(rErr),
			zap.Uint32("user_id", userID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return rErr
	}
	return s.UpdateUserByID(ctx, userID, map[string]any{"password": newPassword})
}

func (s *UserService) RefreshTokens(
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

	claims, err := auth.ParseRefreshToken(ctx, s.jwt, refresh)
	if err != nil {
		s.log.Error(
			"解析刷新令牌失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return "", "", errors.ErrTokenInvalid
	}
	accessToken, rErr = s.newAccessJWT(ctx, claims.UserInfo)
	if rErr != nil {
		return "", "", rErr
	}
	refreshToken, rErr = s.newRefreshJWT(ctx, claims.UserInfo)
	if rErr != nil {
		return "", "", rErr
	}
	return accessToken, refreshToken, nil
}
