package sys

import (
	"context"
	"time"

	"go.uber.org/zap"

	sysmodel "gin-artweb/internal/model/sys"
	sysrepo "gin-artweb/internal/repo/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
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
	roleRepo   *sysrepo.RoleRepo
	userRepo   *sysrepo.UserRepo
	recordRepo *sysrepo.LoginRecordRepo
	hasher     crypto.Hasher
	jwt        *config.JWTConfig
	sec        SecuritySettings
}

func NewUserService(
	log *zap.Logger,
	roleRepo *sysrepo.RoleRepo,
	userRepo *sysrepo.UserRepo,
	recordRepo *sysrepo.LoginRecordRepo,
	hasher crypto.Hasher,
	jwt *config.JWTConfig,
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
) (*sysmodel.RoleModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.roleRepo.GetModel(ctx, nil, roleID)
	if err != nil {
		log.Error(
			"查询用户关联的角色:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"role_id": roleID})
	}
	return m, nil
}

func (s *UserService) CreateUser(
	ctx context.Context,
	dto sysmodel.CreateUserDTO,
) (*sysmodel.UserModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建用户:开始执行",
		zap.Object("create_user_dto", &dto),
	)

	// 检查密码强度
	if err := s.validatePasswordStrength(dto.Password); err != nil {
		log.Error(
			"创建用户:密码强度校验失败",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	m := dto.ToModel()

	// 密码哈希
	if password, err := s.hashPassword(ctx, dto.Password); err != nil {
		log.Error(
			"创建用户:密码哈希失败",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	} else {
		m.Password = password
	}

	// 获取角色信息
	role, rErr := s.GetRole(ctx, m.RoleID)
	if rErr != nil {
		log.Error(
			"创建用户:查询关联角色的数据库模型失败",
			zap.Error(rErr),
			zap.Uint32("role_id", m.RoleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}
	m.Role = *role

	if err := s.userRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建用户:创建数据库模型失败",
			zap.Error(err),
			zap.String("username", m.Username),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	log.Info(
		"创建用户:执行成功",
		zap.Uint32("user_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *UserService) UpdateUserByID(
	ctx context.Context,
	userID uint32,
	dto sysmodel.UpdateUserDTO,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新用户:开始执行",
		zap.Uint32("user_id", userID),
		zap.Object("update_user_dto", &dto),
	)

	updateData := dto.ToUpdateMap()
	if err := s.userRepo.UpdateModel(ctx, updateData, "id = ?", userID); err != nil {
		log.Error(
			"更新用户:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("user_id", userID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, updateData)
	}

	log.Info(
		"更新用户:执行成功",
		zap.Uint32("user_id", userID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *UserService) DeleteUserByID(
	ctx context.Context,
	userID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除用户:执行开始",
		zap.Uint32("user_id", userID),
	)

	if err := s.userRepo.DeleteModel(ctx, userID); err != nil {
		log.Error(
			"删除用户:数据库删除失败",
			zap.Error(err),
			zap.Uint32("user_id", userID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": userID})
	}

	log.Info(
		"删除用户:执行成功",
		zap.Uint32("user_id", userID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *UserService) FindUserByID(
	ctx context.Context,
	preloads []string,
	userID uint32,
) (*sysmodel.UserModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.userRepo.GetModel(ctx, preloads, userID)
	if err != nil {
		log.Error(
			"查询用户:数据库查询失败",
			zap.Error(err),
			zap.Uint32("user_id", userID),
			zap.Strings("preloads", preloads),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": userID})
	}

	log.Debug(
		"查询用户:查询到的用户详情",
		zap.Object("user_model", m),
	)
	return m, nil
}

func (s *UserService) FindUserByName(
	ctx context.Context,
	preloads []string,
	username string,
) (*sysmodel.UserModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.userRepo.GetModel(ctx, preloads, "username = ?", username)
	if err != nil {
		log.Error(
			"查询用户名:数据库查询失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.String("username", username),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"username": username})
	}

	log.Debug(
		"查询用户:查询到的用户详情",
		zap.Object("user_model", m),
	)
	return m, nil
}

func (s *UserService) ListUser(
	ctx context.Context,
	page, size int,
	dto sysmodel.ListUserDTO,
) (int64, []sysmodel.UserModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询用户列表:入参详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_user_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:    limit,
		Offset:   offset,
		OrderBy:  []string{"id ASC"},
		Query:    dto.ToQueryMap(),
		Preloads: []string{"Role"},
	}

	count, err := s.userRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询用户列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询用户列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.userRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询用户列表:数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *UserService) ListLoginRecord(
	ctx context.Context,
	page, size int,
	dto sysmodel.ListLoginRecordDTO,
) (int64, []sysmodel.LoginRecordModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询用户登录记录列表:入参详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_login_record_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:    limit,
		Offset:   offset,
		OrderBy:  []string{"id DESC"},
		Query:    dto.ToQueryMap(),
		Preloads: []string{},
	}

	count, err := s.recordRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询用户登录记录列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询用户登录记录列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return count, nil, nil
	}

	ms, err := s.recordRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询用户登录记录列表:数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *UserService) Login(
	ctx context.Context,
	dto sysmodel.LoginDTO,
	reqCtx sysmodel.RequestContext,
) (string, string, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return "", "", errors.FromError(ctx.Err())
	}
	log := s.log.With(zap.String("trace_id", ctxutil.GetTraceID(ctx)))

	log.Debug(
		"用户登录:请求参数详情",
		zap.String("username", dto.Username),
		zap.String("ip_address", reqCtx.IP),
		zap.String("user_agent", reqCtx.UserAgent),
	)

	num, err := s.recordRepo.GetLoginFailNum(ctx, reqCtx.IP)
	if err != nil {
		log.Error(
			"用户登录:获取登录失败次数失败",
			zap.Error(err),
			zap.String("username", dto.Username),
			zap.String("ip_address", reqCtx.IP),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return "", "", errors.FromError(err)
	}

	if num <= 0 {
		log.Warn(
			"用户登录:登录尝试次数用尽，账户已锁定",
			zap.String("username", dto.Username),
			zap.String("ip_address", reqCtx.IP),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return "", "", errors.ErrAccountLocked
	}

	lrm := sysmodel.LoginRecordModel{
		Username:  dto.Username,
		LoginAt:   time.Now(),
		IPAddress: reqCtx.IP,
		UserAgent: reqCtx.UserAgent,
		Status:    false,
	}

	defer func() {
		if dErr := s.recordRepo.CreateModel(ctx, &lrm); dErr != nil {
			log.Error(
				"创建用户登录记录失败",
				zap.Error(dErr),
				zap.Object("login_record_model", &lrm),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
		if lrm.Status {
			num = s.sec.MaxFailedAttempts
		} else {
			if num > 0 {
				num--
			}
		}
		if fErr := s.recordRepo.SetLoginFailNum(ctx, reqCtx.IP, num); fErr != nil {
			log.Error(
				"用户登录:设置登录失败次数失败",
				zap.Error(fErr),
				zap.String("ip_address", reqCtx.IP),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}()

	m, rErr := s.FindUserByName(ctx, []string{"Role"}, dto.Username)
	if rErr != nil {
		log.Warn(
			"用户不存在或查找失败",
			zap.Error(rErr),
			zap.String("username", dto.Username),
			zap.String("ip_address", reqCtx.IP),
			zap.Int("remaining_attempts", num-1),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		// 更新失败次数
		return "", "", errors.ErrAuthFailed
	}

	// 检查用户状态
	if !m.IsActive {
		log.Warn(
			"用户账户被锁定",
			zap.String("username", dto.Username),
			zap.Uint32("user_id", m.ID),
			zap.String("ip_address", reqCtx.IP),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return "", "", errors.ErrAccountLocked
	}

	if rErr = s.verifyPassword(ctx, dto.Password, m.Password); rErr != nil {
		log.Warn(
			"用户密码验证失败",
			zap.Error(rErr),
			zap.String("username", dto.Username),
			zap.Uint32("user_id", m.ID),
			zap.String("ip_address", reqCtx.IP),
			zap.Int("remaining_attempts", num-1),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return "", "", rErr.WithField("remaining_attempts", num-1)
	}

	// 登录认证成功
	log.Debug(
		"用户登录: 验证登录信息成功",
		zap.String("username", dto.Username),
	)
	lrm.Status = true

	userinfo := auth.UserInfo{
		Username: dto.Username,
		UserID:   m.ID,
		RoleID:   m.RoleID,
		IsStaff:  m.IsStaff,
	}

	// 生成JWT token
	accessToken, err := auth.NewAccessJWT(ctx, s.jwt, userinfo)
	if err != nil {
		log.Error(
			"用户登录: 生成访问令牌失败",
			zap.Error(err),
			zap.String("username", dto.Username),
			zap.Uint32("user_id", m.ID),
			zap.String("ip_address", reqCtx.IP),
		)
		return "", "", errors.FromError(err)
	}

	refreshToken, err := auth.NewRefreshJWT(ctx, s.jwt, userinfo)
	if err != nil {
		log.Error(
			"用户登录: 生成刷新令牌失败",
			zap.Error(err),
			zap.String("username", dto.Username),
			zap.Uint32("user_id", m.ID),
			zap.String("ip_address", reqCtx.IP),
		)
		return "", "", errors.FromError(err)
	}
	return accessToken, refreshToken, nil
}

func (s *UserService) verifyPassword(
	ctx context.Context,
	pwd, hash string,
) *errors.Error {
	verified, err := s.hasher.Verify(ctx, pwd, hash)
	if err != nil {
		return errors.ErrAuthFailed.WithCause(err)
	}

	if !verified {
		return errors.ErrAuthFailed
	}

	return nil
}

func (s *UserService) hashPassword(
	ctx context.Context,
	pwd string,
) (string, *errors.Error) {
	hashedPassword, err := s.hasher.Hash(ctx, pwd)
	if err != nil {
		return "", errors.FromError(err)
	}
	return hashedPassword, nil
}

func (s *UserService) validatePasswordStrength(
	pwd string,
) *errors.Error {
	strength := GetPasswordStrength(pwd)
	if strength < s.sec.PasswordStrength {
		return errors.ErrPasswordStrengthFailed
	}
	return nil
}

func (s *UserService) ResetPassword(
	ctx context.Context,
	userID uint32,
	password string,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"重置用户密码:开始执行",
		zap.Uint32("user_id", userID),
	)

	if err := s.validatePasswordStrength(password); err != nil {
		log.Error(
			"重置用户密码:密码强度校验失败",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	// 哈希新密码
	hashedPassword, err := s.hashPassword(ctx, password)
	if err != nil {
		log.Error(
			"重置用户密码:密码哈希失败",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	// 更新用户密码
	if err := s.userRepo.UpdateModel(ctx, map[string]any{"password": hashedPassword}, "id = ?", userID); err != nil {
		log.Error(
			"重置用户密码:更新数据库密码失败",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": userID})
	}

	log.Info(
		"重置用户密码:执行成功",
		zap.Uint32("user_id", userID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *UserService) PatchPassword(
	ctx context.Context,
	userID uint32,
	oldPassword string,
	newPassword string,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	// 检查旧密码是否正确
	m, err := s.FindUserByID(ctx, nil, userID)
	if err != nil {
		log.Error(
			"获取用户信息失败",
			zap.Error(err),
			zap.Uint32("user_id", userID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	if err = s.verifyPassword(ctx, oldPassword, m.Password); err != nil {
		log.Error(
			"旧密码验证失败",
			zap.Error(err),
			zap.Uint32("user_id", userID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	if err = s.ResetPassword(ctx, userID, newPassword); err != nil {
		log.Error(
			"重置用户密码失败",
			zap.Error(err),
			zap.Uint32("user_id", userID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	return nil
}

func (s *UserService) RefreshTokens(
	ctx context.Context,
	refresh string,
) (string, string, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return "", "", errors.FromError(ctx.Err())
	}
	log := s.log.With(zap.String("trace_id", ctxutil.GetTraceID(ctx)))

	claims, rErr := auth.ParseRefreshToken(ctx, s.jwt, refresh)
	if rErr != nil {
		log.Error(
			"刷新令牌:解析刷新令牌失败",
			zap.Error(rErr),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return "", "", errors.ErrTokenInvalid
	}
	accessToken, err := auth.NewAccessJWT(ctx, s.jwt, claims.UserInfo)
	if err != nil {
		log.Error(
			"刷新令牌:生成访问令牌失败",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return "", "", errors.FromError(err)
	}
	refreshToken, err := auth.NewRefreshJWT(ctx, s.jwt, claims.UserInfo)
	if err != nil {
		log.Error(
			"刷新令牌:生成刷新令牌失败",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return "", "", errors.FromError(err)
	}
	return accessToken, refreshToken, nil
}
