package data

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/customer/biz"
	"gin-artweb/internal/shared/database"
)

// CreateTestUserModel 创建测试用的用户模型
func CreateTestUserModel(pk uint32) *biz.UserModel {
	return &biz.UserModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: pk,
			},
		},
		Username: "test_user_" + string(rune('0'+pk)),
		Password: "test_password",
		IsActive: true,
		IsStaff:  false,
		RoleID:   1,
	}
}

type UserTestSuite struct {
	suite.Suite
	userRepo *userRepo
}

func (suite *UserTestSuite) SetupSuite() {
	suite.userRepo = NewTestUserRepo()
}

func (suite *UserTestSuite) TestCreateUser() {
	// 测试创建用户
	sm := CreateTestUserModel(1)
	err := suite.userRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建用户应该成功")

	// 测试查询刚创建的用户
	fm, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的用户应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(sm.Username, fm.Username)
	suite.Equal(sm.IsActive, fm.IsActive)
	suite.Equal(sm.IsStaff, fm.IsStaff)
	suite.Equal(sm.RoleID, fm.RoleID)
}

func (suite *UserTestSuite) TestUpdateUser() {
	// 测试创建用户
	sm := CreateTestUserModel(2)
	err := suite.userRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建用户应该成功")

	// 测试更新用户
	updatedIsActive := false
	updatedIsStaff := true
	updatedRoleID := uint32(2)

	err = suite.userRepo.UpdateModel(context.Background(), map[string]any{
		"is_active": updatedIsActive,
		"is_staff":  updatedIsStaff,
		"role_id":   updatedRoleID,
	}, "id = ?", sm.ID)
	suite.NoError(err, "更新用户应该成功")

	// 测试查询更新后的用户
	fm, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的用户应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(updatedIsActive, fm.IsActive)
	suite.Equal(updatedIsStaff, fm.IsStaff)
	suite.Equal(updatedRoleID, fm.RoleID)
	suite.Greater(fm.UpdatedAt, sm.UpdatedAt)
}

func (suite *UserTestSuite) TestDeleteUser() {
	// 测试创建用户
	sm := CreateTestUserModel(3)
	err := suite.userRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建用户应该成功")

	// 测试查询刚创建的用户
	fm, err := suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的用户应该成功")
	suite.Equal(sm.ID, fm.ID)

	// 测试删除用户
	err = suite.userRepo.DeleteModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "删除用户应该成功")

	// 测试查询已删除的用户
	_, err = suite.userRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *UserTestSuite) TestGetUserByID() {
	// 测试创建用户
	sm := CreateTestUserModel(4)
	err := suite.userRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建用户应该成功")

	// 测试根据ID查询用户
	m, err := suite.userRepo.GetModel(context.Background(), []string{}, sm.ID)
	suite.NoError(err, "根据ID查询用户应该成功")
	suite.Equal(sm.ID, m.ID)
	suite.Equal(sm.Username, m.Username)
	suite.Equal(sm.IsActive, m.IsActive)
	suite.Equal(sm.IsStaff, m.IsStaff)
	suite.Equal(sm.RoleID, m.RoleID)
}

func (suite *UserTestSuite) TestListUsers() {
	// 测试创建多个用户
	for i := 5; i < 15; i++ {
		sm := CreateTestUserModel(uint32(i))
		err := suite.userRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建用户应该成功")
	}

	// 测试查询用户列表
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	count, ms, err := suite.userRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询用户列表应该成功")
	suite.NotNil(ms, "用户列表不应该为nil")
	suite.GreaterOrEqual(count, int64(10), "用户总数应该至少有10条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:   5,
		Offset:  0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.userRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询用户列表应该成功")
	suite.NotNil(pMs, "分页用户列表不应该为nil")
	suite.Equal(5, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(5), "分页总数应该至少等于limit")
}

func (suite *UserTestSuite) TestCreateUserWithInvalidData() {
	// 测试创建用户时传入空数据
	err := suite.userRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建用户时传入nil应该返回错误")
}

func (suite *UserTestSuite) TestUpdateUserWithInvalidData() {
	// 测试更新用户时传入空数据
	err := suite.userRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", 1)
	suite.Error(err, "更新用户时传入空数据应该返回错误")
}

func (suite *UserTestSuite) TestFindNonExistentUser() {
	// 测试查找不存在的用户
	_, err := suite.userRepo.GetModel(context.Background(), []string{}, 999999)
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查找不存在的用户应该返回记录未找到错误")
}

// 每个测试文件都需要这个入口函数
func TestUserTestSuite(t *testing.T) {
	pts := &UserTestSuite{}
	suite.Run(t, pts)
}
