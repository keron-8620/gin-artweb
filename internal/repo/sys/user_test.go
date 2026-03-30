package sys

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	sysmodel "gin-artweb/internal/model/sys"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

// CreateTestUserModel 创建测试用户模型
func CreateTestUserModel(roleID uint32) *sysmodel.UserModel {
	return &sysmodel.UserModel{
		Username: fmt.Sprintf("test_user_%s", uuid.NewString()),
		Password: "password123",
		RoleID:   roleID,
		IsActive: true,
		IsStaff:  false,
	}
}

// UserTestSuite 用户测试套件
type UserTestSuite struct {
	suite.Suite
	userRepo *UserRepo
}

// SetupSuite 测试套件设置
func (suite *UserTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&sysmodel.UserModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.userRepo = &UserRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

// TestCreateUser 测试创建用户
func (suite *UserTestSuite) TestCreateUser() {
	// 创建用户
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 测试查询刚创建的用户
	fm, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", user.ID)
	suite.NoError(err, "查询刚创建的用户应该成功")
	suite.Equal(user.ID, fm.ID)
}

// TestUpdateUser 测试更新用户
func (suite *UserTestSuite) TestUpdateUser() {
	// 创建用户
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 测试更新用户
	updatedUsername := "updated_username"
	err = suite.userRepo.UpdateModel(context.Background(), map[string]any{
		"username": updatedUsername,
	}, "id = ?", user.ID)
	suite.NoError(err, "更新用户应该成功")

	// 测试查询更新后的用户
	fm, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", user.ID)
	suite.NoError(err, "查询更新后的用户应该成功")
	suite.Equal(user.ID, fm.ID)
	suite.Equal(updatedUsername, fm.Username)
}

// TestDeleteUser 测试删除用户
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

// TestListUsers 测试查询用户列表
func (suite *UserTestSuite) TestListUsers() {
	// 测试创建多个用户
	for range 5 {
		user := CreateTestUserModel(0)
		err := suite.userRepo.CreateModel(context.Background(), user)
		suite.NoError(err, "创建用户应该成功")
	}

	// 测试CountModel
	count, err := suite.userRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取用户总数应该成功")
	suite.GreaterOrEqual(count, int64(5), "用户总数应该至少有5条")

	// 测试查询用户列表
	qp := database.QueryParams{
		Limit:  10,
		Offset: 0,
	}
	ms, err := suite.userRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询用户列表应该成功")
	suite.NotNil(ms, "用户列表不应该为nil")
	suite.GreaterOrEqual(len(ms), 5, "用户列表应该至少有5条")
}

// TestListUserWithPagination 测试分页查询用户列表
func (suite *UserTestSuite) TestListUserWithPagination() {
	// 测试创建多个用户
	for range 5 {
		user := CreateTestUserModel(0)
		err := suite.userRepo.CreateModel(context.Background(), user)
		suite.NoError(err, "创建用户应该成功")
	}

	// 测试CountModel
	count, err := suite.userRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取用户总数应该成功")
	suite.GreaterOrEqual(count, int64(5), "用户总数应该至少有5条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:  2,
		Offset: 0,
	}
	pMs, err := suite.userRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询用户列表应该成功")
	suite.NotNil(pMs, "分页用户列表不应该为nil")
	suite.Equal(2, len(pMs), "分页查询应该返回指定数量的记录")
}

// TestListUserWithPaginationBoundaries 测试分页边界情况
func (suite *UserTestSuite) TestListUserWithPaginationBoundaries() {
	// 测试Limit=0的情况
	qpZeroLimit := database.QueryParams{
		Limit:  0,
		Offset: 0,
	}
	msZero, err := suite.userRepo.ListModel(context.Background(), qpZeroLimit)
	suite.NoError(err, "Limit=0应该成功查询")
	suite.NotNil(msZero, "用户列表不应该为nil")

	// 测试较大的Offset值
	qpLargeOffset := database.QueryParams{
		Limit:  10,
		Offset: 999999,
	}
	msLarge, err := suite.userRepo.ListModel(context.Background(), qpLargeOffset)
	suite.NoError(err, "较大的Offset值应该成功查询")
	suite.NotNil(msLarge, "用户列表不应该为nil")
	suite.LessOrEqual(len(msLarge), 10, "返回的记录数应该不超过Limit")
}

// TestListUserWithNoRecords 测试查询无记录的情况
func (suite *UserTestSuite) TestListUserWithNoRecords() {
	// 测试查询不存在的条件
	qp := database.QueryParams{
		Limit:  10,
		Offset: 0,
		Query:  map[string]any{"id": uint32(999999)},
	}
	ms, err := suite.userRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询无记录的用户列表应该成功")
	suite.NotNil(ms, "用户列表不应该为nil")
	suite.Len(ms, 0, "无记录时用户列表长度应该为0")

	// 测试CountModel带过滤条件
	count, err := suite.userRepo.CountModel(context.Background(), map[string]any{"id": uint32(999999)})
	suite.NoError(err, "带过滤条件的用户总数查询应该成功")
	suite.Equal(int64(0), count, "无记录时计数应该为0")
}

// TestContextTimeout 测试上下文超时
func (suite *UserTestSuite) TestContextTimeout() {
	// 创建一个会立即超时的上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, 1*time.Nanosecond)
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
	_, err = suite.userRepo.GetModel(ctx, []string{}, user.ID)
	suite.Error(err, "上下文超时后获取用户应该返回错误")

	// 测试ListModel方法
	qp := database.QueryParams{}
	_, err = suite.userRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文超时后查询用户列表应该返回错误")

	// 测试CountModel方法
	_, err = suite.userRepo.CountModel(ctx, nil)
	suite.Error(err, "上下文超时后计数查询应该返回错误")
}

// TestGetUserWithPreload 测试预加载关联数据
func (suite *UserTestSuite) TestGetUserWithPreload() {
	// 创建用户
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 测试预加载
	fm, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", user.ID)
	suite.NoError(err, "预加载查询应该成功")
	suite.Equal(user.ID, fm.ID)
}

// TestUpdateUserWithNonExistentID 测试更新不存在的用户ID
func (suite *UserTestSuite) TestUpdateUserWithNonExistentID() {
	// 测试更新不存在的用户ID
	err := suite.userRepo.UpdateModel(context.Background(), map[string]any{
		"username": "test_user",
	}, "id = ?", uint32(999999))
	suite.NoError(err, "更新不存在的用户ID应该成功")
}

// TestDeleteUserWithNonExistentID 测试删除不存在的用户ID
func (suite *UserTestSuite) TestDeleteUserWithNonExistentID() {
	// 测试删除不存在的用户ID
	err := suite.userRepo.DeleteModel(context.Background(), "id = ?", uint32(999999))
	suite.NoError(err, "删除不存在的用户ID应该成功")
}

// TestGetUser 测试根据ID查询用户
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

// TestListUsersWithEmptyParams 测试查询用户列表时传入空参数
func (suite *UserTestSuite) TestListUsersWithEmptyParams() {
	// 测试查询用户列表时传入空参数
	qp := database.QueryParams{}
	ms, err := suite.userRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询用户列表时传入空参数应该成功")
	suite.NotNil(ms, "用户列表不应该为nil")

	// 测试CountModel
	count, err := suite.userRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取用户总数应该成功")
	suite.GreaterOrEqual(count, int64(0), "用户总数应该大于等于0")
}

// TestListUsersWithSorting 测试查询用户列表时传入排序参数
func (suite *UserTestSuite) TestListUsersWithSorting() {
	// 测试创建多个用户
	for i := 0; i < 5; i++ {
		user := CreateTestUserModel(0)
		err := suite.userRepo.CreateModel(context.Background(), user)
		suite.NoError(err, "创建用户应该成功")
	}

	// 测试按ID降序排序
	qp := database.QueryParams{
		OrderBy: []string{"id DESC"},
	}
	ms, err := suite.userRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按ID降序排序查询应该成功")
	suite.NotNil(ms, "用户列表不应该为nil")
	if len(ms) > 1 {
		// 验证排序结果
		prevID := ms[0].ID
		for _, user := range ms {
			suite.LessOrEqual(user.ID, prevID, "用户应该按ID降序排序")
			prevID = user.ID
		}
	}
}

// TestListUsersWithFiltering 测试查询用户列表时传入过滤参数
func (suite *UserTestSuite) TestListUsersWithFiltering() {
	// 测试创建一个特定用户名的用户
	testUsername := "filter_test_user"
	user := CreateTestUserModel(0)
	user.Username = testUsername
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	// 测试按用户名过滤
	qp := database.QueryParams{
		Query: map[string]any{
			"username": testUsername,
		},
	}
	ms, err := suite.userRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按用户名过滤查询应该成功")
	suite.NotNil(ms, "用户列表不应该为nil")
	// 验证过滤结果
	for _, user := range ms {
		suite.Equal(testUsername, user.Username, "用户应该按用户名过滤")
	}

	// 测试CountModel带过滤条件
	count, err := suite.userRepo.CountModel(context.Background(), map[string]any{
		"username": testUsername,
	})
	suite.NoError(err, "带过滤条件的用户总数查询应该成功")
	suite.GreaterOrEqual(count, int64(1), "带过滤条件的用户总数应该至少为1")
}

func (suite *UserTestSuite) TestCreateUserWithNilModel() {
	err := suite.userRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建用户时传入nil模型应该返回错误")
	suite.True(strings.Contains(err.Error(), "创建用户模型: 模型不能为空"), "错误信息应该包含'创建用户模型: 模型不能为空'")
}

func (suite *UserTestSuite) TestUpdateUserWithEmptyData() {
	err := suite.userRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", 1)
	suite.Error(err, "更新用户时传入空数据应该返回错误")
	suite.True(strings.Contains(err.Error(), "更新用户模型：更新数据不能为空"), "错误信息应该包含'更新用户模型：更新数据不能为空'")
}

func (suite *UserTestSuite) TestCreateUserWithContextTimeout() {
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	time.Sleep(time.Millisecond * 5)

	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(ctx, user)
	suite.Error(err, "创建用户时上下文超时应该返回错误")
}

func (suite *UserTestSuite) TestUpdateUserWithContextTimeout() {
	user := CreateTestUserModel(0)
	err := suite.userRepo.CreateModel(context.Background(), user)
	suite.NoError(err, "创建用户应该成功")

	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	time.Sleep(time.Millisecond * 5)

	err = suite.userRepo.UpdateModel(ctx, map[string]any{
		"username": "updated_user",
	}, "id = ?", user.ID)
	suite.Error(err, "更新用户时上下文超时应该返回错误")
}

// 每个测试文件都需要这个入口函数
func TestUserTestSuite(t *testing.T) {
	pts := &UserTestSuite{}
	suite.Run(t, pts)
}
