package sys

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	sysmodel "gin-artweb/internal/model/sys"
	syssvc "gin-artweb/internal/repo/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/test"
	"gin-artweb/pkg/crypto"
)

// CreateTestUserModel 创建测试用的用户模型
func CreateTestUserModel(roleID uint32) *sysmodel.UserModel {
	return &sysmodel.UserModel{
		Username: uuid.NewString(),
		Password: "Test123!@#$%",
		IsActive: true,
		IsStaff:  false,
		RoleID:   roleID,
	}
}

// CreateTestUserDTO 创建测试用的用户DTO
func CreateTestUserDTO(roleID uint32) sysmodel.CreateUserDTO {
	return sysmodel.CreateUserDTO{
		Username: uuid.NewString(),
		Password: "Test123!@#$%",
		IsActive: true,
		IsStaff:  false,
		RoleID:   roleID,
	}
}

// CreateTestLoginRecordModel 创建测试用的登录记录模型
func CreateTestLoginRecordModel(ip string) *sysmodel.LoginRecordModel {
	return &sysmodel.LoginRecordModel{
		Username:  "test_user",
		IPAddress: ip,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		Status:    true,
	}
}

type UserTestSuite struct {
	suite.Suite
	uc *UserService
}

func (suite *UserTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&sysmodel.MenuModel{},
		&sysmodel.ApiModel{},
		&sysmodel.ButtonModel{},
		&sysmodel.RoleModel{},
		&sysmodel.UserModel{},
		&sysmodel.LoginRecordModel{},
	)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()

	suite.uc = NewUserService(
		logger,
		syssvc.NewRoleRepo(
			logger,
			db,
			dbTimeout,
			enforcer,
		),
		syssvc.NewUserRepo(
			logger,
			db,
			dbTimeout,
		),
		syssvc.NewLoginRecordRepo(
			logger,
			db,
			dbTimeout,
			time.Duration(10)*time.Minute,
			time.Duration(10)*time.Minute,
			2,
		),
		crypto.NewBcryptHasher(12),
		auth.NewJWTConfig(
			time.Duration(10)*time.Second,
			time.Duration(10)*time.Minute,
			"HS256",
			"HS256",
			[]byte("test_access_secret"),
			[]byte("test_refresh_secret"),
		),
		SecuritySettings{
			MaxFailedAttempts: 2,
			LockDuration:      time.Duration(5) * time.Second,
			PasswordStrength:  3,
		},
	)
}

// 每个测试文件都需要这个入口函数
func TestUserTestSuite(t *testing.T) {
	pts := &UserTestSuite{}
	suite.Run(t, pts)
}

// TestGetRole 测试获取用户关联的角色
func (suite *UserTestSuite) TestGetRole() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 测试获取角色
	role, err := suite.uc.GetRole(context.Background(), testRole.ID)
	suite.Nil(err, "获取角色应该成功")
	suite.NotNil(role, "角色不应该为空")
	suite.Equal(testRole.ID, role.ID, "角色ID应该匹配")
	suite.Equal(testRole.Name, role.Name, "角色名称应该匹配")
}

// TestFindUserByID 测试根据ID查询用户
func (suite *UserTestSuite) TestFindUserByID() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 测试查询用户
	foundUser, err := suite.uc.FindUserByID(context.Background(), []string{"Role"}, createdUser.ID)
	suite.Nil(err, "查询用户应该成功")
	suite.NotNil(foundUser, "用户不应该为空")
	suite.Equal(createdUser.ID, foundUser.ID, "用户ID应该匹配")
	suite.Equal(createdUser.Username, foundUser.Username, "用户名应该匹配")
	suite.Equal(createdUser.RoleID, foundUser.RoleID, "角色ID应该匹配")
}

// TestFindUserByName 测试根据用户名查询用户
func (suite *UserTestSuite) TestFindUserByName() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 测试查询用户
	foundUser, err := suite.uc.FindUserByName(context.Background(), []string{"Role"}, createdUser.Username)
	suite.Nil(err, "查询用户应该成功")
	suite.NotNil(foundUser, "用户不应该为空")
	suite.Equal(createdUser.Username, foundUser.Username, "用户名应该匹配")
	suite.Equal(createdUser.RoleID, foundUser.RoleID, "角色ID应该匹配")
}

// TestListUser 测试查询用户列表
func (suite *UserTestSuite) TestListUser() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	userCount := 2
	for i := 0; i < userCount; i++ {
		_, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
		suite.Nil(err, "创建用户应该成功")
	}

	// 测试查询用户列表
	count, users, err := suite.uc.ListUser(context.Background(), 1, 10, sysmodel.ListUserDTO{})
	suite.Nil(err, "查询用户列表应该成功")
	suite.GreaterOrEqual(int(count), userCount, "用户数量应该大于或等于创建的数量")
	suite.NotNil(users, "用户列表不应该为空")
}

// TestListLoginRecord 测试查询登录记录列表
func (suite *UserTestSuite) TestListLoginRecord() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 测试登录，生成登录记录
	_, _, err = suite.uc.Login(context.Background(), sysmodel.LoginDTO{
		Username: createdUser.Username,
		Password: "Test123!@#$%",
	}, sysmodel.RequestContext{
		IP:        "127.0.0.1",
		UserAgent: "test_user_agent",
	})
	suite.Nil(err, "登录应该成功")

	// 测试查询登录记录列表
	count, records, err := suite.uc.ListLoginRecord(context.Background(), 1, 10, sysmodel.ListLoginRecordDTO{})
	suite.Nil(err, "查询登录记录列表应该成功")
	suite.GreaterOrEqual(int(count), 1, "登录记录数量应该大于或等于1")
	suite.NotNil(records, "登录记录列表不应该为空")
}

// TestCreateUser 测试创建用户
func (suite *UserTestSuite) TestCreateUser() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	testUser := CreateTestUserDTO(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), testUser)
	suite.Nil(err, "创建用户应该成功")
	suite.NotNil(createdUser, "用户不应该为空")
	suite.Equal(testUser.Username, createdUser.Username, "用户名应该匹配")
	suite.NotEqual(testUser.Password, createdUser.Password, "密码应该被哈希处理")
	suite.Equal(testUser.RoleID, createdUser.RoleID, "角色ID应该匹配")
}

// TestUpdateUserByID 测试更新用户
func (suite *UserTestSuite) TestUpdateUserByID() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 更新用户
	updatedUsername := uuid.NewString()
	err = suite.uc.UpdateUserByID(context.Background(), createdUser.ID, sysmodel.UpdateUserDTO{
		Username: updatedUsername,
		IsActive: false,
	})
	suite.Nil(err, "更新用户应该成功")

	// 验证更新
	foundUser, err := suite.uc.FindUserByID(context.Background(), []string{}, createdUser.ID)
	suite.Nil(err, "查询用户应该成功")
	suite.Equal(updatedUsername, foundUser.Username, "用户名应该更新")
	suite.False(foundUser.IsActive, "用户状态应该更新")
}

// TestDeleteUserByID 测试删除用户
func (suite *UserTestSuite) TestDeleteUserByID() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 删除用户
	err = suite.uc.DeleteUserByID(context.Background(), createdUser.ID)
	suite.Nil(err, "删除用户应该成功")

	// 验证用户已删除
	_, err = suite.uc.FindUserByID(context.Background(), []string{}, createdUser.ID)
	suite.NotNil(err, "查询已删除的用户应该失败")
}

// TestLogin 测试用户登录（成功场景）
func (suite *UserTestSuite) TestLogin() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 测试登录
	accessToken, refreshToken, err := suite.uc.Login(context.Background(), sysmodel.LoginDTO{
		Username: createdUser.Username,
		Password: "Test123!@#$%",
	}, sysmodel.RequestContext{
		IP:        "127.0.0.1",
		UserAgent: "test_user_agent",
	})
	suite.Nil(err, "登录应该成功")
	suite.NotEmpty(accessToken, "访问令牌不应该为空")
	suite.NotEmpty(refreshToken, "刷新令牌不应该为空")
}

// TestLoginWithFailedPassword 测试用户登录（密码失败场景）
func (suite *UserTestSuite) TestLoginWithFailedPassword() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 测试登录（密码错误）
	_, _, err = suite.uc.Login(context.Background(), sysmodel.LoginDTO{
		Username: createdUser.Username,
		Password: "wrong_password",
	}, sysmodel.RequestContext{
		IP:        "127.0.0.1",
		UserAgent: "test_user_agent",
	})
	suite.NotNil(err, "登录应该失败")
}

// TestPatchPassword 测试修改密码
func (suite *UserTestSuite) TestPatchPassword() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 修改密码
	newPassword := "NewTest123!@#$%" // 强度足够的新密码
	err = suite.uc.PatchPassword(context.Background(), createdUser.ID, "Test123!@#$%", newPassword)
	suite.Nil(err, "修改密码应该成功")

	// 验证新密码可以登录
	_, _, err = suite.uc.Login(context.Background(), sysmodel.LoginDTO{
		Username: createdUser.Username,
		Password: newPassword,
	}, sysmodel.RequestContext{
		IP:        "127.0.0.1",
		UserAgent: "test_user_agent",
	})
	suite.Nil(err, "使用新密码登录应该成功")
}

// TestPatchPasswordWithInvalidOldPassword 测试修改密码（旧密码错误场景）
func (suite *UserTestSuite) TestPatchPasswordWithInvalidOldPassword() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 修改密码（旧密码错误）
	err = suite.uc.PatchPassword(context.Background(), createdUser.ID, "wrong_old_password", "NewTest123!@#$%")
	suite.NotNil(err, "旧密码错误应该返回错误")
}

// TestResetPasswordWithWeakPassword 测试重置密码（密码强度不足场景）
func (suite *UserTestSuite) TestResetPasswordWithWeakPassword() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 重置密码（密码强度不足）
	err = suite.uc.ResetPassword(context.Background(), createdUser.ID, "weak")
	suite.NotNil(err, "密码强度不足应该返回错误")
}

// TestRefreshTokensWithInvalidToken 测试刷新令牌（无效令牌场景）
func (suite *UserTestSuite) TestRefreshTokensWithInvalidToken() {
	// 测试无效刷新令牌
	_, _, err := suite.uc.RefreshTokens(context.Background(), "invalid_token")
	suite.NotNil(err, "无效令牌应该返回错误")
}

// TestValidatePasswordStrength 测试密码强度验证
func (suite *UserTestSuite) TestValidatePasswordStrength() {
	// 测试密码强度足够
	err := suite.uc.validatePasswordStrength(context.Background(), "Test123!@#$%")
	suite.Nil(err, "密码强度足够应该返回 nil")

	// 测试密码强度不足
	err = suite.uc.validatePasswordStrength(context.Background(), "weak")
	suite.NotNil(err, "密码强度不足应该返回错误")
}

// TestValidatePasswordStrengthWithContextError 测试密码强度验证（上下文错误场景）
func (suite *UserTestSuite) TestValidatePasswordStrengthWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	err := suite.uc.validatePasswordStrength(ctx, "Test123!@#$%")
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestHashPassword 测试密码哈希
func (suite *UserTestSuite) TestHashPassword() {
	// 测试正常密码哈希
	hashedPassword, err := suite.uc.hashPassword(context.Background(), "Test123!@#$%")
	suite.Nil(err, "密码哈希应该成功")
	suite.NotEmpty(hashedPassword, "哈希密码不应该为空")
}

// TestHashPasswordWithContextError 测试密码哈希（上下文错误场景）
func (suite *UserTestSuite) TestHashPasswordWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.hashPassword(ctx, "Test123!@#$%")
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestVerifyPassword 测试密码验证
func (suite *UserTestSuite) TestVerifyPassword() {
	// 先哈希一个密码
	hashedPassword, err := suite.uc.hashPassword(context.Background(), "Test123!@#$%")
	suite.Nil(err, "密码哈希应该成功")

	// 测试密码验证成功
	err = suite.uc.verifyPassword(context.Background(), "Test123!@#$%", hashedPassword)
	suite.Nil(err, "密码验证应该成功")

	// 测试密码验证失败
	err = suite.uc.verifyPassword(context.Background(), "wrong_password", hashedPassword)
	suite.NotNil(err, "密码验证失败应该返回错误")
}

// TestVerifyPasswordWithContextError 测试密码验证（上下文错误场景）
func (suite *UserTestSuite) TestVerifyPasswordWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	err := suite.uc.verifyPassword(ctx, "test", "test")
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestRefreshTokens 测试刷新令牌
func (suite *UserTestSuite) TestRefreshTokens() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(testRole.ID))
	suite.Nil(err, "创建用户应该成功")

	// 登录获取令牌
	_, refreshToken, err := suite.uc.Login(context.Background(), sysmodel.LoginDTO{
		Username: createdUser.Username,
		Password: "Test123!@#$%",
	}, sysmodel.RequestContext{
		IP:        "127.0.0.1",
		UserAgent: "test_user_agent",
	})
	suite.Nil(err, "登录应该成功")

	// 刷新令牌
	newAccessToken, newRefreshToken, err := suite.uc.RefreshTokens(context.Background(), refreshToken)
	suite.Nil(err, "刷新令牌应该成功")
	suite.NotEmpty(newAccessToken, "新访问令牌不应该为空")
	suite.NotEmpty(newRefreshToken, "新刷新令牌不应该为空")
}

// TestGetRoleWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestGetRoleWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.GetRole(ctx, 1)
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestCreateUserWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestCreateUserWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.CreateUser(ctx, CreateTestUserDTO(1))
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestUpdateUserByIDWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestUpdateUserByIDWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	err := suite.uc.UpdateUserByID(ctx, 1, sysmodel.UpdateUserDTO{})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestDeleteUserByIDWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestDeleteUserByIDWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	err := suite.uc.DeleteUserByID(ctx, 1)
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestFindUserByIDWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestFindUserByIDWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.FindUserByID(ctx, []string{}, 1)
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestFindUserByNameWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestFindUserByNameWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.FindUserByName(ctx, []string{}, "test")
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestListUserWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestListUserWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, _, err := suite.uc.ListUser(ctx, 1, 10, sysmodel.ListUserDTO{})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestListLoginRecordWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestListLoginRecordWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, _, err := suite.uc.ListLoginRecord(ctx, 1, 10, sysmodel.ListLoginRecordDTO{})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestLoginWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestLoginWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, _, err := suite.uc.Login(ctx, sysmodel.LoginDTO{
		Username: "test",
		Password: "test",
	}, sysmodel.RequestContext{
		IP:        "127.0.0.1",
		UserAgent: "test",
	})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestPatchPasswordWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestPatchPasswordWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	err := suite.uc.PatchPassword(ctx, 1, "old", "new")
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestRefreshTokensWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestRefreshTokensWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, _, err := suite.uc.RefreshTokens(ctx, "test")
	suite.NotNil(err, "上下文错误应该返回错误")
}
