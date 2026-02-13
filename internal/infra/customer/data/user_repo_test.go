package data

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/customer/model"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

// CreateTestUserModel 创建测试用的用户模型
func CreateTestUserModel(roleID uint32) *model.UserModel {
	return &model.UserModel{
		Username: uuid.NewString(),
		Password: "test_password",
		IsActive: true,
		IsStaff:  false,
		RoleID:   roleID,
	}
}

type UserTestSuite struct {
	suite.Suite
	roleRepo *RoleRepo
	userRepo *UserRepo
}

func (suite *UserTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&model.ApiModel{},
		&model.MenuModel{},
		&model.ButtonModel{},
		&model.RoleModel{},
		&model.UserModel{},
	)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	suite.roleRepo = &RoleRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}
	suite.userRepo = &UserRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *UserTestSuite) TestCreateUser() {
	// 测试创建用户
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 测试查询刚创建的用户
	fm, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", user.ID)
	suite.NoError(err, "查询刚创建的用户应该成功")
	suite.Equal(user.ID, fm.ID)
	suite.Equal(user.Username, fm.Username)
	suite.Equal(user.Password, fm.Password)
	suite.Equal(user.IsActive, fm.IsActive)
	suite.Equal(user.IsStaff, fm.IsStaff)
	suite.Equal(user.RoleID, fm.RoleID)
}

func (suite *UserTestSuite) TestCreateUserWithNilModel() {
	// 测试创建用户时传入nil模型
	err := suite.userRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "传入nil模型应该返回错误")
	suite.Contains(err.Error(), "创建用户模型失败: 模型为空")
}

func (suite *UserTestSuite) TestCreateUserWithRole() {
	// 先创建一个角色用于测试
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 创建与角色关联的用户
	user := CreateTestUserModel(role.ID)
	err = suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建与角色关联的用户应该成功")

	// 验证用户创建成功且关联了正确的角色
	fm, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", user.ID)
	suite.NoError(err, "查询用户应该成功")
	suite.Equal(user.ID, fm.ID)
	suite.Equal(role.ID, fm.RoleID, "用户应该与正确的角色关联")
}

func (suite *UserTestSuite) TestUpdateUser() {
	// 创建用户
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 测试更新用户
	updatedUsername := "updated_" + uuid.NewString()
	updatedPassword := "updated_password"
	updatedIsActive := false
	updatedIsStaff := true

	err = suite.userRepo.UpdateModel(context.Background(), map[string]any{
		"username":  updatedUsername,
		"password":  updatedPassword,
		"is_active": updatedIsActive,
		"is_staff":  updatedIsStaff,
	}, "id = ?", user.ID)
	suite.NoError(err, "更新用户应该成功")

	// 测试查询更新后的用户
	fm, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", user.ID)
	suite.NoError(err, "查询更新后的用户应该成功")
	suite.Equal(user.ID, fm.ID)
	suite.Equal(updatedUsername, fm.Username)
	suite.Equal(updatedPassword, fm.Password)
	suite.Equal(updatedIsActive, fm.IsActive)
	suite.Equal(updatedIsStaff, fm.IsStaff)
}

func (suite *UserTestSuite) TestUpdateUserWithEmptyData() {
	// 创建用户
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 传入空的data映射
	err = suite.userRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", user.ID)
	suite.Error(err, "传入空data更新用户应该返回错误")
	suite.Contains(err.Error(), "更新用户模型失败: 更新数据为空")
}

func (suite *UserTestSuite) TestUpdateUserWithNonExistentID() {
	// 测试更新不存在的用户ID
	err := suite.userRepo.UpdateModel(context.Background(), map[string]any{
		"username": "test_user",
	}, "id = ?", uint32(999999))
	suite.NoError(err, "更新不存在的用户ID应该成功")
}

func (suite *UserTestSuite) TestDeleteUser() {
	// 创建用户
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 测试查询刚创建的用户
	fm, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", user.ID)
	suite.NoError(err, "查询刚创建的用户应该成功")
	suite.Equal(user.ID, fm.ID)

	// 测试删除用户
	err = suite.userRepo.DeleteModel(context.Background(), "id = ?", user.ID)
	suite.NoError(err, "删除用户应该成功")

	// 测试查询已删除的用户
	_, err = suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", user.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *UserTestSuite) TestDeleteUserWithNonExistentID() {
	// 测试删除不存在的用户ID
	err := suite.userRepo.DeleteModel(context.Background(), "id = ?", uint32(999999))
	suite.NoError(err, "删除不存在的用户ID应该成功")
}

func (suite *UserTestSuite) TestGetUser() {
	// 创建用户
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 测试根据ID查询用户
	m, err := suite.userRepo.GetModel(context.Background(), []string{}, user.ID)
	suite.NoError(err, "根据ID查询用户应该成功")
	suite.Equal(user.ID, m.ID)
	suite.Equal(user.Username, m.Username)
	suite.Equal(user.Password, m.Password)
	suite.Equal(user.IsActive, m.IsActive)
	suite.Equal(user.IsStaff, m.IsStaff)
	suite.Equal(user.RoleID, m.RoleID)
}

func (suite *UserTestSuite) TestGetUserWithNonExistentID() {
	// 测试查询不存在的用户ID
	_, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", uint32(999999))
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询不存在的用户ID应该返回记录未找到错误")
}

func (suite *UserTestSuite) TestGetUserWithEmptyConditions() {
	// 测试查询时传入空条件
	_, err := suite.userRepo.GetModel(context.Background(), []string{})
	// 当传入空条件时，GetModel方法会尝试获取数据库中的第一条记录
	// 如果数据库为空，会返回record not found错误
	// 如果数据库不为空，会返回第一条记录
	if err != nil {
		// 如果返回错误，应该是record not found
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询时传入空条件应该返回记录未找到错误")
	} else {
		// 如果返回结果，应该是一个有效的用户模型
		// 这里不做断言，因为可能没有数据
	}
}

func (suite *UserTestSuite) TestGetUserWithPreloadRole() {
	// 先创建一个角色用于测试
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 创建与角色关联的用户
	user := CreateTestUserModel(role.ID)
	err = suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建与角色关联的用户应该成功")

	// 不预加载角色查询用户
	userWithoutRole, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", user.ID)
	suite.NoError(err, "不预加载角色查询用户应该成功")

	// 预加载角色查询用户
	userWithRole, err := suite.userRepo.GetModel(context.Background(), []string{"Role"}, "id = ?", user.ID)
	suite.NoError(err, "预加载角色查询用户应该成功")
	suite.Equal(userWithoutRole.ID, userWithRole.ID)
	suite.NotNil(userWithRole.Role, "预加载后应该包含角色信息")
	suite.Equal(role.ID, userWithRole.Role.ID)
	suite.Equal(role.Name, userWithRole.Role.Name)
}

func (suite *UserTestSuite) TestListUser() {
	// 测试创建多个用户
	for range 5 {
		user := CreateTestUserModel(0)
		err := suite.userRepo.CreateModel(context.Background(), user)
		suite.NoError(err, "创建用户应该成功")
	}

	// 测试查询用户列表
	qp := database.QueryParams{
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	count, ms, err := suite.userRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询用户列表应该成功")
	suite.NotNil(ms, "用户列表不应该为nil")
	suite.GreaterOrEqual(count, int64(5), "用户总数应该至少有5条")
}

func (suite *UserTestSuite) TestListUserWithPagination() {
	// 测试创建多个用户
	for range 5 {
		user := CreateTestUserModel(0)
		err := suite.userRepo.CreateModel(context.Background(), user)
		suite.NoError(err, "创建用户应该成功")
	}

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Size:    2,
		Page:    0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.userRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询用户列表应该成功")
	suite.NotNil(pMs, "分页用户列表不应该为nil")
	suite.Equal(2, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(2), "分页总数应该至少等于limit")
}

func (suite *UserTestSuite) TestListUserWithPaginationBoundaries() {
	// 测试Limit=0的情况
	qpZeroLimit := database.QueryParams{
		Size:    0,
		Page:    0,
		IsCount: true,
	}
	_, msZero, err := suite.userRepo.ListModel(context.Background(), qpZeroLimit)
	suite.NoError(err, "Limit=0应该成功查询")
	suite.NotNil(msZero, "用户列表不应该为nil")

	// 测试较大的Offset值
	qpLargeOffset := database.QueryParams{
		Size:    10,
		Page:    999999,
		IsCount: true,
	}
	_, msLarge, err := suite.userRepo.ListModel(context.Background(), qpLargeOffset)
	suite.NoError(err, "较大的Offset值应该成功查询")
	suite.NotNil(msLarge, "用户列表不应该为nil")
	suite.LessOrEqual(len(*msLarge), 10, "返回的记录数应该不超过Limit")
}

func (suite *UserTestSuite) TestListUserWithNoRecords() {
	// 测试查询不存在的条件
	qp := database.QueryParams{
		Size:    10,
		Page:    0,
		IsCount: true,
		Query:   map[string]any{"id": uint32(999999)},
	}
	count, ms, err := suite.userRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询无记录的用户列表应该成功")
	suite.NotNil(ms, "用户列表不应该为nil")
	suite.Equal(int64(0), count, "无记录时计数应该为0")
	suite.Len(*ms, 0, "无记录时用户列表长度应该为0")
}

func (suite *UserTestSuite) TestContextTimeout() {
	// 创建一个会立即超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// 等待上下文超时
	time.Sleep(10 * time.Nanosecond)

	// 创建用户用于测试
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 测试CreateModel方法
	sm := CreateTestUserModel(0)
	err = suite.userRepo.CreateModel(ctx, sm)
	suite.Error(err, "上下文超时后创建用户应该返回错误")

	// 测试UpdateModel方法
	err = suite.userRepo.UpdateModel(ctx, map[string]any{
		"username": "test_user",
	}, "id = ?", user.ID)
	suite.Error(err, "上下文超时后更新用户应该返回错误")

	// 测试DeleteModel方法
	err = suite.userRepo.DeleteModel(ctx, "id = ?", user.ID)
	suite.Error(err, "上下文超时后删除用户应该返回错误")

	// 测试GetModel方法
	_, err = suite.userRepo.GetModel(ctx, []string{}, "id = ?", user.ID)
	suite.Error(err, "上下文超时后获取用户应该返回错误")

	// 测试ListModel方法
	qp := database.QueryParams{
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	_, _, err = suite.userRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文超时后列出用户应该返回错误")
}

func (suite *UserTestSuite) TestContextCancel() {
	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	// 立即取消上下文
	cancel()

	// 创建用户用于测试
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 测试CreateModel方法
	sm := CreateTestUserModel(0)
	err = suite.userRepo.CreateModel(ctx, sm)
	suite.Error(err, "上下文取消后创建用户应该返回错误")

	// 测试UpdateModel方法
	err = suite.userRepo.UpdateModel(ctx, map[string]any{
		"username": "test_user",
	}, "id = ?", user.ID)
	suite.Error(err, "上下文取消后更新用户应该返回错误")

	// 测试DeleteModel方法
	err = suite.userRepo.DeleteModel(ctx, "id = ?", user.ID)
	suite.Error(err, "上下文取消后删除用户应该返回错误")

	// 测试GetModel方法
	_, err = suite.userRepo.GetModel(ctx, []string{}, "id = ?", user.ID)
	suite.Error(err, "上下文取消后获取用户应该返回错误")

	// 测试ListModel方法
	qp := database.QueryParams{
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	_, _, err = suite.userRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文取消后列出用户应该返回错误")
}

// 每个测试文件都需要这个入口函数
func TestUserTestSuite(t *testing.T) {
	pts := &UserTestSuite{}
	suite.Run(t, pts)
}
