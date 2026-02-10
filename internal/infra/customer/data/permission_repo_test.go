package data

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/customer/biz"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/database"
)

func CreateTestPermModel(pk uint32) *biz.PermissionModel {
	return &biz.PermissionModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: pk,
			},
		},
		URL:    fmt.Sprintf("/permission/%d", pk),
		Method: "GET",
		Label:  fmt.Sprintf("test_%d", pk),
		Descr:  fmt.Sprintf("这是一个测试接口_%d", pk),
	}
}

type PermissionTestSuite struct {
	suite.Suite
	permRepo *permissionRepo
}

func (suite *PermissionTestSuite) SetupSuite() {
	suite.permRepo = NewTestPermissionRepo()
}

func (suite *PermissionTestSuite) TestCreatePermission() {
	sm := CreateTestPermModel(1)
	err := suite.permRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建权限应该成功")

	fm, err := suite.permRepo.GetModel(context.Background(), sm.ID)
	suite.NoError(err, "查询刚创建的权限应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(sm.URL, fm.URL)
	suite.Equal(sm.Method, fm.Method)
	suite.Equal(sm.Label, fm.Label)
	suite.Equal(sm.Descr, fm.Descr)
}

func (suite *PermissionTestSuite) TestUpdatePermission() {
	sm := CreateTestPermModel(2)
	err := suite.permRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建权限应该成功")

	updatedURL := fmt.Sprintf("/permission/%d", sm.ID)
	updatedMethod := "POST"
	updatedLabel := "updated_test"
	updatedDescr := "这是更新的测试接口"

	err = suite.permRepo.UpdateModel(context.Background(), map[string]any{
		"url":    updatedURL,
		"method": updatedMethod,
		"label":  updatedLabel,
		"descr":  updatedDescr,
	}, "id = ?", sm.ID)
	suite.NoError(err, "更新权限应该成功")

	fm, err := suite.permRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的权限应该成功")
	suite.Equal(fm.ID, sm.ID)
	suite.Equal(updatedURL, fm.URL)
	suite.Equal(updatedMethod, fm.Method)
	suite.Equal(updatedLabel, fm.Label)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Greater(fm.UpdatedAt, sm.UpdatedAt)
}

func (suite *PermissionTestSuite) TestDeletePermission() {
	sm := CreateTestPermModel(3)
	err := suite.permRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建权限应该成功")

	fm, err := suite.permRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的权限应该成功")
	suite.Equal(sm.ID, fm.ID)

	err = suite.permRepo.DeleteModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "删除权限应该成功")

	_, err = suite.permRepo.GetModel(context.Background(), "id = ?", sm.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *PermissionTestSuite) TestFindPermissionByID() {
	// 创建一个权限用于测试
	sm := CreateTestPermModel(4)
	err := suite.permRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建权限应该成功")

	m, err := suite.permRepo.GetModel(context.Background(), sm.ID)
	suite.NoError(err, "查询权限应该成功")
	suite.Equal(sm.URL, m.URL)
	suite.Equal(sm.Method, m.Method)
	suite.Equal(sm.Label, m.Label)
	suite.Equal(sm.Descr, m.Descr)
}

func (suite *PermissionTestSuite) TestListPermission() {
	// 清理可能存在的数据并创建测试数据
	for i := 10; i < 20; i++ {
		sm := CreateTestPermModel(uint32(i))
		err := suite.permRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建权限应该成功")
	}

	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	total, ms, err := suite.permRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列出权限应该成功")
	suite.NotNil(ms, "权限列表不应该为nil")
	suite.GreaterOrEqual(total, int64(10), "权限总数应该至少有10条")

	qpPaginated := database.QueryParams{
		Limit:   5,
		Offset:  0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.permRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页列出权限应该成功")
	suite.NotNil(pMs, "分页权限列表不应该为nil")
	suite.Equal(5, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(5), "分页总数应该至少等于limit")
}

func (suite *PermissionTestSuite) TestAddPolicy() {
	// 创建一个权限用于测试策略添加
	sm := CreateTestPermModel(6)
	err := suite.permRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建权限应该成功")

	err = suite.permRepo.AddPolicy(context.Background(), *sm)
	suite.NoError(err, "添加策略应该成功")

	sub := auth.PermissionToSubject(sm.ID)
	ok, err := suite.permRepo.enforcer.Enforce(sub, sm.URL, sm.Method)
	suite.NoError(err, "检查授权应该成功")
	suite.True(ok, "添加策略后应该有权限")
}

func (suite *PermissionTestSuite) TestRemovePolicy() {
	// 创建一个权限用于测试策略移除
	sm := CreateTestPermModel(7)
	err := suite.permRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建权限应该成功")

	// 先添加策略
	err = suite.permRepo.AddPolicy(context.Background(), *sm)
	suite.NoError(err, "添加策略应该成功")

	// 验证策略存在
	sub := auth.PermissionToSubject(sm.ID)
	ok, err := suite.permRepo.enforcer.Enforce(sub, sm.URL, sm.Method)
	suite.NoError(err, "检查授权应该成功")
	suite.True(ok, "添加策略后应该有权限")

	// 移除策略
	err = suite.permRepo.RemovePolicy(context.Background(), *sm, true)
	suite.NoError(err, "移除策略应该成功")

	// 验证策略已移除
	ok, cErr := suite.permRepo.enforcer.Enforce(sub, sm.URL, sm.Method)
	suite.NoError(cErr, "检查授权应该成功")
	suite.False(ok, "移除策略后不应该有权限")
}

func (suite *PermissionTestSuite) TestCreatePermissionWithInvalidData() {
	// 测试创建权限时传入空数据
	err := suite.permRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建权限时传入nil应该返回错误")
}

func (suite *PermissionTestSuite) TestFindNonExistentPermission() {
	// 测试查找不存在的权限
	_, err := suite.permRepo.GetModel(context.Background(), 999999)
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查找不存在的权限应该返回记录未找到错误")
}

func (suite *PermissionTestSuite) TestAuthorizationAfterRemovePolicy() {
	// 创建一个权限用于测试
	sm := CreateTestPermModel(8)
	err := suite.permRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建权限应该成功")

	// 添加策略
	err = suite.permRepo.AddPolicy(context.Background(), *sm)
	suite.NoError(err, "添加策略应该成功")

	// 验证策略存在
	sub := auth.PermissionToSubject(sm.ID)
	ok, err := suite.permRepo.enforcer.Enforce(sub, sm.URL, sm.Method)
	suite.NoError(err, "检查授权应该成功")
	suite.True(ok, "添加策略后应该有权限")

	// 移除策略
	err = suite.permRepo.RemovePolicy(context.Background(), *sm, true)
	suite.NoError(err, "移除策略应该成功")

	// 验证策略已移除
	ok, cErr := suite.permRepo.enforcer.Enforce(sub, sm.URL, sm.Method)
	suite.NoError(cErr, "检查授权应该成功")
	suite.False(ok, "移除策略后不应该有权限")
}

// 每个测试文件都需要这个入口函数
func TestPermissionTestSuite(t *testing.T) {
	pts := &PermissionTestSuite{}
	suite.Run(t, pts)
}
