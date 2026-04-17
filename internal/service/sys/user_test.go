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
	"gin-artweb/internal/shared/config"
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

// CreateTestRole 创建测试角色并返回ID
func (suite *UserTestSuite) CreateTestRole() uint32 {
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")
	return testRole.ID
}

// CreateTestUser 创建测试用户并返回用户对象
func (suite *UserTestSuite) CreateTestUser(roleID uint32) *sysmodel.UserModel {
	createdUser, err := suite.uc.CreateUser(context.Background(), CreateTestUserDTO(roleID))
	suite.Nil(err, "创建用户应该成功")
	return createdUser
}

type UserTestSuite struct {
	suite.Suite
	uc *UserService
}

func (suite *UserTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(
		&sysmodel.MenuModel{},
		&sysmodel.ApiModel{},
		&sysmodel.ButtonModel{},
		&sysmodel.RoleModel{},
		&sysmodel.UserModel{},
		&sysmodel.LoginRecordModel{},
	); err != nil {
		suite.Error(err, "数据库迁移失败")
	}
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, err := auth.NewCasbinEnforcer()
	if err != nil {
		suite.Error(err, "创建Casbinforcer失败")
	}
	suite.uc = NewUserService(
		logger,
		syssvc.NewRoleRepo(
			logger,
			db,
			dbTimeout,
			slowThreshold,
			enforcer,
		),
		syssvc.NewUserRepo(
			logger,
			db,
			dbTimeout,
			slowThreshold,
		),
		syssvc.NewLoginRecordRepo(
			logger,
			db,
			dbTimeout,
			slowThreshold,
			time.Duration(10)*time.Minute,
			time.Duration(10)*time.Minute,
			2,
		),
		crypto.NewBcryptHasher(12),
		config.NewJWTConfig(
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
	roleID := suite.CreateTestRole()

	// 测试获取角色
	role, err := suite.uc.GetRole(context.Background(), roleID)
	suite.Nil(err, "获取角色应该成功")
	suite.NotNil(role, "角色不应该为空")
	suite.Equal(roleID, role.ID, "角色ID应该匹配")
	suite.NotEmpty(role.Name, "角色名称不应该为空")
}

// TestFindUserByID 测试根据ID查询用户
func (suite *UserTestSuite) TestFindUserByID() {
	// 创建测试角色和用户
	roleID := suite.CreateTestRole()
	createdUser := suite.CreateTestUser(roleID)

	// 测试查询用户
	foundUser, err := suite.uc.FindUserByID(context.Background(), []string{"Role"}, createdUser.ID)
	suite.Nil(err, "查询用户应该成功")
	suite.NotNil(foundUser, "用户不应该为空")
	suite.Equal(createdUser.ID, foundUser.ID, "用户ID应该匹配")
	suite.Equal(createdUser.Username, foundUser.Username, "用户名应该匹配")
	suite.Equal(createdUser.RoleID, foundUser.RoleID, "角色ID应该匹配")
	suite.True(foundUser.IsActive, "用户应该是活跃状态")
}

// TestFindUserByName 测试根据用户名查询用户
func (suite *UserTestSuite) TestFindUserByName() {
	// 创建测试角色和用户
	roleID := suite.CreateTestRole()
	createdUser := suite.CreateTestUser(roleID)

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
	roleID := suite.CreateTestRole()

	// 创建测试用户
	userCount := 2
	for i := 0; i < userCount; i++ {
		suite.CreateTestUser(roleID)
	}

	// 测试查询用户列表
	count, users, err := suite.uc.ListUser(context.Background(), 1, 10, sysmodel.ListUserDTO{})
	suite.Nil(err, "查询用户列表应该成功")
	suite.GreaterOrEqual(int(count), userCount, "用户数量应该大于或等于创建的数量")
	suite.NotNil(users, "用户列表不应该为空")
	suite.Greater(len(users), 0, "用户列表长度应该大于0")
}

// TestListLoginRecord 测试查询登录记录列表
func (suite *UserTestSuite) TestListLoginRecord() {
	// 创建测试角色和用户
	roleID := suite.CreateTestRole()
	createdUser := suite.CreateTestUser(roleID)

	// 测试登录，生成登录记录
	_, _, err := suite.uc.Login(context.Background(), sysmodel.LoginDTO{
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
	suite.Greater(len(records), 0, "登录记录列表长度应该大于0")
}

// TestCreateUser 测试创建用户
func (suite *UserTestSuite) TestCreateUser() {
	// 创建测试角色
	roleID := suite.CreateTestRole()

	// 创建测试用户
	testUser := CreateTestUserDTO(roleID)
	createdUser, err := suite.uc.CreateUser(context.Background(), testUser)
	suite.Nil(err, "创建用户应该成功")
	suite.NotNil(createdUser, "用户不应该为空")
	suite.Equal(testUser.Username, createdUser.Username, "用户名应该匹配")
	suite.NotEqual(testUser.Password, createdUser.Password, "密码应该被哈希处理")
	suite.Equal(testUser.RoleID, createdUser.RoleID, "角色ID应该匹配")
	suite.True(createdUser.IsActive, "用户应该是活跃状态")
	suite.False(createdUser.IsStaff, "用户不应该是管理员")
}

// TestUpdateUserByID 测试更新用户
func (suite *UserTestSuite) TestUpdateUserByID() {
	// 创建测试角色和用户
	roleID := suite.CreateTestRole()
	createdUser := suite.CreateTestUser(roleID)

	// 更新用户
	updatedUsername := uuid.NewString()
	err := suite.uc.UpdateUserByID(context.Background(), createdUser.ID, sysmodel.UpdateUserDTO{
		Username: updatedUsername,
		IsActive: false,
	})
	suite.Nil(err, "更新用户应该成功")

	// 验证更新
	foundUser, err := suite.uc.FindUserByID(context.Background(), []string{}, createdUser.ID)
	suite.Nil(err, "查询用户应该成功")
	suite.Equal(updatedUsername, foundUser.Username, "用户名应该更新")
	suite.False(foundUser.IsActive, "用户状态应该更新为非活跃")
}

// TestDeleteUserByID 测试删除用户
func (suite *UserTestSuite) TestDeleteUserByID() {
	// 创建测试角色和用户
	roleID := suite.CreateTestRole()
	createdUser := suite.CreateTestUser(roleID)

	// 删除用户
	err := suite.uc.DeleteUserByID(context.Background(), createdUser.ID)
	suite.Nil(err, "删除用户应该成功")

	// 验证用户已删除
	_, err = suite.uc.FindUserByID(context.Background(), []string{}, createdUser.ID)
	suite.NotNil(err, "查询已删除的用户应该失败")
}

// TestPasswordManagement 测试密码管理功能
func (suite *UserTestSuite) TestPasswordManagement() {
	// 创建测试角色和用户
	roleID := suite.CreateTestRole()
	createdUser := suite.CreateTestUser(roleID)

	// 测试修改密码成功场景
	newPassword := "NewTest123!@#$%" // 强度足够的新密码
	err := suite.uc.PatchPassword(context.Background(), createdUser.ID, "Test123!@#$%", newPassword)
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

	// 测试修改密码失败场景（旧密码错误）
	err = suite.uc.PatchPassword(context.Background(), createdUser.ID, "wrong_old_password", "AnotherNewPassword123!@#")
	suite.NotNil(err, "旧密码错误应该返回错误")

	// 测试重置密码失败场景（密码强度不足）
	err = suite.uc.ResetPassword(context.Background(), createdUser.ID, "weak")
	suite.NotNil(err, "密码强度不足应该返回错误")
}

// TestPasswordFunctions 测试密码相关功能
func (suite *UserTestSuite) TestPasswordFunctions() {
	// 测试密码强度验证
	err := suite.uc.validatePasswordStrength("Test123!@#$%")
	suite.Nil(err, "密码强度足够应该返回 nil")

	err = suite.uc.validatePasswordStrength("weak")
	suite.NotNil(err, "密码强度不足应该返回错误")

	// 测试密码哈希
	hashedPassword, err := suite.uc.hashPassword(context.Background(), "Test123!@#$%")
	suite.Nil(err, "密码哈希应该成功")
	suite.NotEmpty(hashedPassword, "哈希密码不应该为空")
	suite.Greater(len(hashedPassword), 0, "哈希密码长度应该大于0")

	// 测试密码验证
	err = suite.uc.verifyPassword(context.Background(), "Test123!@#$%", hashedPassword)
	suite.Nil(err, "密码验证应该成功")

	err = suite.uc.verifyPassword(context.Background(), "wrong_password", hashedPassword)
	suite.NotNil(err, "密码验证失败应该返回错误")
}

// TestLoginFunctions 测试登录相关功能
func (suite *UserTestSuite) TestLoginFunctions() {
	// 创建测试角色和用户
	roleID := suite.CreateTestRole()
	createdUser := suite.CreateTestUser(roleID)

	// 测试登录成功场景
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
	suite.Greater(len(accessToken), 0, "访问令牌长度应该大于0")
	suite.Greater(len(refreshToken), 0, "刷新令牌长度应该大于0")

	// 测试登录失败场景（密码错误）
	_, _, err = suite.uc.Login(context.Background(), sysmodel.LoginDTO{
		Username: createdUser.Username,
		Password: "wrong_password",
	}, sysmodel.RequestContext{
		IP:        "127.0.0.1",
		UserAgent: "test_user_agent",
	})
	suite.NotNil(err, "登录应该失败")

	// 测试刷新令牌
	newAccessToken, newRefreshToken, err := suite.uc.RefreshTokens(context.Background(), refreshToken)
	suite.Nil(err, "刷新令牌应该成功")
	suite.NotEmpty(newAccessToken, "新访问令牌不应该为空")
	suite.NotEmpty(newRefreshToken, "新刷新令牌不应该为空")
	suite.Greater(len(newAccessToken), 0, "新访问令牌长度应该大于0")
	suite.Greater(len(newRefreshToken), 0, "新刷新令牌长度应该大于0")

	// 测试无效刷新令牌
	_, _, err = suite.uc.RefreshTokens(context.Background(), "invalid_token")
	suite.NotNil(err, "无效令牌应该返回错误")
}

// TestWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试 GetRole 上下文错误
	_, err := suite.uc.GetRole(ctx, 1)
	suite.NotNil(err, "GetRole 上下文错误应该返回错误")

	// 测试 CreateUser 上下文错误
	_, err = suite.uc.CreateUser(ctx, CreateTestUserDTO(1))
	suite.NotNil(err, "CreateUser 上下文错误应该返回错误")

	// 测试 UpdateUserByID 上下文错误
	err = suite.uc.UpdateUserByID(ctx, 1, sysmodel.UpdateUserDTO{})
	suite.NotNil(err, "UpdateUserByID 上下文错误应该返回错误")

	// 测试 DeleteUserByID 上下文错误
	err = suite.uc.DeleteUserByID(ctx, 1)
	suite.NotNil(err, "DeleteUserByID 上下文错误应该返回错误")

	// 测试 FindUserByID 上下文错误
	_, err = suite.uc.FindUserByID(ctx, []string{}, 1)
	suite.NotNil(err, "FindUserByID 上下文错误应该返回错误")

	// 测试 FindUserByName 上下文错误
	_, err = suite.uc.FindUserByName(ctx, []string{}, "test")
	suite.NotNil(err, "FindUserByName 上下文错误应该返回错误")

	// 测试 ListUser 上下文错误
	_, _, err = suite.uc.ListUser(ctx, 1, 10, sysmodel.ListUserDTO{})
	suite.NotNil(err, "ListUser 上下文错误应该返回错误")

	// 测试 ListLoginRecord 上下文错误
	_, _, err = suite.uc.ListLoginRecord(ctx, 1, 10, sysmodel.ListLoginRecordDTO{})
	suite.NotNil(err, "ListLoginRecord 上下文错误应该返回错误")

	// 测试 Login 上下文错误
	_, _, err = suite.uc.Login(ctx, sysmodel.LoginDTO{
		Username: "test",
		Password: "test",
	}, sysmodel.RequestContext{
		IP:        "127.0.0.1",
		UserAgent: "test",
	})
	suite.NotNil(err, "Login 上下文错误应该返回错误")

	// 测试 PatchPassword 上下文错误
	err = suite.uc.PatchPassword(ctx, 1, "old", "new")
	suite.NotNil(err, "PatchPassword 上下文错误应该返回错误")

	// 测试 RefreshTokens 上下文错误
	_, _, err = suite.uc.RefreshTokens(ctx, "test")
	suite.NotNil(err, "RefreshTokens 上下文错误应该返回错误")
}
