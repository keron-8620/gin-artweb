package biz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/infra/customer/data"
	"gin-artweb/internal/infra/customer/model"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
	"gin-artweb/pkg/crypto"
)

// CreateTestUserModel 创建测试用的用户模型
func CreateTestUserModel(roleID uint32) *model.UserModel {
	return &model.UserModel{
		Username: uuid.NewString(),
		Password: "Test123!@#$%", // 强度足够的密码（长度12，包含大小写字母、数字和多个特殊字符）
		IsActive: true,
		IsStaff:  false,
		RoleID:   roleID,
	}
}

// CreateTestLoginRecordModel 创建测试用的登录记录模型
func CreateTestLoginRecordModel(ip string) *model.LoginRecordModel {
	return &model.LoginRecordModel{
		Username:  "test_user",
		IPAddress: ip,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		Status:    true,
	}
}

type UserTestSuite struct {
	suite.Suite
	uc *UserUsecase
}

func (suite *UserTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&model.MenuModel{},
		&model.ApiModel{},
		&model.ButtonModel{},
		&model.RoleModel{},
		&model.UserModel{},
		&model.LoginRecordModel{},
	)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()

	suite.uc = &UserUsecase{
		log: logger,
		roleRepo: data.NewRoleRepo(
			logger,
			db,
			dbTimeout,
			enforcer,
		),
		userRepo: data.NewUserRepo(
			logger,
			db,
			dbTimeout,
		),
		recordRepo: data.NewLoginRecordRepo(
			logger,
			db,
			dbTimeout,
			time.Duration(10)*time.Minute,
			time.Duration(10)*time.Minute,
			2,
		),
		hasher: crypto.NewBcryptHasher(12),
		jwt: auth.NewJWTConfig(
			time.Duration(10)*time.Second,
			time.Duration(10)*time.Minute,
			"HS256",
			"HS256",
			[]byte("test_access_secret"),
			[]byte("test_refresh_secret"),
		),
		sec: SecuritySettings{
			MaxFailedAttempts: 2,
			LockDuration:      time.Duration(5) * time.Second,
			PasswordStrength:  3,
		},
	}
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
	testUser := CreateTestUserModel(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), *testUser)
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
	testUser := CreateTestUserModel(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), *testUser)
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
		testUser := CreateTestUserModel(testRole.ID)
		_, err := suite.uc.CreateUser(context.Background(), *testUser)
		suite.Nil(err, "创建用户应该成功")
	}

	// 测试查询用户列表
	count, users, err := suite.uc.ListUser(context.Background(), database.QueryParams{})
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
	testUser := CreateTestUserModel(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), *testUser)
	suite.Nil(err, "创建用户应该成功")

	// 测试登录，生成登录记录
	_, _, err = suite.uc.Login(context.Background(), createdUser.Username, "Test123!@#$%", "127.0.0.1", "test_user_agent")
	suite.Nil(err, "登录应该成功")

	// 测试查询登录记录列表
	count, records, err := suite.uc.ListLoginRecord(context.Background(), database.QueryParams{})
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
	testUser := CreateTestUserModel(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), *testUser)
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
	testUser := CreateTestUserModel(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), *testUser)
	suite.Nil(err, "创建用户应该成功")

	// 更新用户
	updatedUsername := uuid.NewString()
	err = suite.uc.UpdateUserByID(context.Background(), createdUser.ID, map[string]any{
		"username":  updatedUsername,
		"is_active": false,
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
	testUser := CreateTestUserModel(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), *testUser)
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
	testUser := CreateTestUserModel(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), *testUser)
	suite.Nil(err, "创建用户应该成功")

	// 测试登录
	accessToken, refreshToken, err := suite.uc.Login(context.Background(), createdUser.Username, "Test123!@#$%", "127.0.0.1", "test_user_agent")
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
	testUser := CreateTestUserModel(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), *testUser)
	suite.Nil(err, "创建用户应该成功")

	// 测试登录（密码错误）
	_, _, err = suite.uc.Login(context.Background(), createdUser.Username, "wrong_password", "127.0.0.1", "test_user_agent")
	suite.NotNil(err, "登录应该失败")
}

// TestPatchPassword 测试修改密码
func (suite *UserTestSuite) TestPatchPassword() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	testUser := CreateTestUserModel(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), *testUser)
	suite.Nil(err, "创建用户应该成功")

	// 修改密码
	newPassword := "NewTest123!@#$%" // 强度足够的新密码
	err = suite.uc.PatchPassword(context.Background(), createdUser.ID, "Test123!@#$%", newPassword)
	suite.Nil(err, "修改密码应该成功")

	// 验证新密码可以登录
	_, _, err = suite.uc.Login(context.Background(), createdUser.Username, newPassword, "127.0.0.1", "test_user_agent")
	suite.Nil(err, "使用新密码登录应该成功")
}

// TestRefreshTokens 测试刷新令牌
func (suite *UserTestSuite) TestRefreshTokens() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	err := suite.uc.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil)
	suite.Nil(err, "创建角色应该成功")

	// 创建测试用户
	testUser := CreateTestUserModel(testRole.ID)
	createdUser, err := suite.uc.CreateUser(context.Background(), *testUser)
	suite.Nil(err, "创建用户应该成功")

	// 登录获取令牌
	_, refreshToken, err := suite.uc.Login(context.Background(), createdUser.Username, "Test123!@#$%", "127.0.0.1", "test_user_agent")
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
	testUser := CreateTestUserModel(1)
	_, err := suite.uc.CreateUser(ctx, *testUser)
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestUpdateUserByIDWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestUpdateUserByIDWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	err := suite.uc.UpdateUserByID(ctx, 1, map[string]any{})
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
	_, _, err := suite.uc.ListUser(ctx, database.QueryParams{})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestListLoginRecordWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestListLoginRecordWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, _, err := suite.uc.ListLoginRecord(ctx, database.QueryParams{})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestLoginWithContextError 测试上下文错误处理
func (suite *UserTestSuite) TestLoginWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, _, err := suite.uc.Login(ctx, "test", "test", "127.0.0.1", "test")
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
