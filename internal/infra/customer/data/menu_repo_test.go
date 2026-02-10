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

// CreateTestMenuModel 创建测试用的菜单模型
func CreateTestMenuModel(pk uint32, parentID *uint32) *biz.MenuModel {
	return &biz.MenuModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: pk,
			},
		},
		ParentID:     parentID,
		Name:         "test_menu_" + string(rune('0'+pk)),
		Path:         "/test/menu/" + string(rune('0'+pk)),
		Component:    "TestMenu",
		Meta:         biz.Meta{Title: "测试菜单" + string(rune('0'+pk)), Icon: "test-icon"},
		ArrangeOrder: pk,
		IsActive:     true,
		Descr:        "这是一个测试菜单" + string(rune('0'+pk)),
	}
}

type MenuTestSuite struct {
	suite.Suite
	menuRepo *menuRepo
}

func (suite *MenuTestSuite) SetupSuite() {
	suite.menuRepo = NewTestMenuRepo()
}

func (suite *MenuTestSuite) TestCreateMenu() {
	// 测试创建菜单
	parentID := uint32(0)
	sm := CreateTestMenuModel(1, &parentID)
	perms := []biz.PermissionModel{}
	err := suite.menuRepo.CreateModel(context.Background(), sm, &perms)
	suite.NoError(err, "创建菜单应该成功")

	// 测试查询刚创建的菜单
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的菜单应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(sm.Name, fm.Name)
	suite.Equal(sm.Path, fm.Path)
	suite.Equal(sm.Component, fm.Component)
	suite.Equal(sm.Meta.Title, fm.Meta.Title)
	suite.Equal(sm.Meta.Icon, fm.Meta.Icon)
	suite.Equal(sm.ArrangeOrder, fm.ArrangeOrder)
	suite.Equal(sm.IsActive, fm.IsActive)
	suite.Equal(sm.Descr, fm.Descr)
}

func (suite *MenuTestSuite) TestUpdateMenu() {
	// 测试创建菜单
	parentID := uint32(0)
	sm := CreateTestMenuModel(2, &parentID)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 测试更新菜单
	updatedName := "更新的测试菜单"
	updatedPath := "/updated/test/menu"
	updatedArrangeOrder := uint32(10)
	updatedIsActive := false
	updatedDescr := "更新的菜单描述"

	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name":          updatedName,
		"path":          updatedPath,
		"arrange_order": updatedArrangeOrder,
		"is_active":     updatedIsActive,
		"descr":         updatedDescr,
	}, nil, "id = ?", sm.ID)
	suite.NoError(err, "更新菜单应该成功")

	// 测试查询更新后的菜单
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedPath, fm.Path)
	suite.Equal(updatedArrangeOrder, fm.ArrangeOrder)
	suite.Equal(updatedIsActive, fm.IsActive)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Greater(fm.UpdatedAt, sm.UpdatedAt)
}

func (suite *MenuTestSuite) TestDeleteMenu() {
	// 测试创建菜单
	parentID := uint32(0)
	sm := CreateTestMenuModel(3, &parentID)
	perms := []biz.PermissionModel{}
	err := suite.menuRepo.CreateModel(context.Background(), sm, &perms)
	suite.NoError(err, "创建菜单应该成功")

	// 测试查询刚创建的菜单
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的菜单应该成功")
	suite.Equal(sm.ID, fm.ID)

	// 测试删除菜单
	err = suite.menuRepo.DeleteModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "删除菜单应该成功")

	// 测试查询已删除的菜单
	_, err = suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *MenuTestSuite) TestGetMenuByID() {
	// 测试创建菜单
	parentID := uint32(0)
	sm := CreateTestMenuModel(4, &parentID)
	perms := []biz.PermissionModel{}
	err := suite.menuRepo.CreateModel(context.Background(), sm, &perms)
	suite.NoError(err, "创建菜单应该成功")

	// 测试根据ID查询菜单
	m, err := suite.menuRepo.GetModel(context.Background(), []string{}, sm.ID)
	suite.NoError(err, "根据ID查询菜单应该成功")
	suite.Equal(sm.ID, m.ID)
	suite.Equal(sm.Name, m.Name)
	suite.Equal(sm.Path, m.Path)
	suite.Equal(sm.Component, m.Component)
	suite.Equal(sm.Meta.Title, m.Meta.Title)
	suite.Equal(sm.Meta.Icon, m.Meta.Icon)
	suite.Equal(sm.ArrangeOrder, m.ArrangeOrder)
	suite.Equal(sm.IsActive, m.IsActive)
	suite.Equal(sm.Descr, m.Descr)
}

func (suite *MenuTestSuite) TestListMenus() {
	// 测试创建多个菜单
	parentID := uint32(0)
	for i := 5; i < 10; i++ {
		sm := CreateTestMenuModel(uint32(i), &parentID)
		perms := []biz.PermissionModel{}
		err := suite.menuRepo.CreateModel(context.Background(), sm, &perms)
		suite.NoError(err, "创建菜单应该成功")
	}

	// 测试查询菜单列表
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	count, ms, err := suite.menuRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询菜单列表应该成功")
	suite.NotNil(ms, "菜单列表不应该为nil")
	suite.GreaterOrEqual(count, int64(5), "菜单总数应该至少有5条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:   2,
		Offset:  0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.menuRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询菜单列表应该成功")
	suite.NotNil(pMs, "分页菜单列表不应该为nil")
	suite.Equal(2, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(2), "分页总数应该至少等于limit")
}

func (suite *MenuTestSuite) TestAddGroupPolicy() {
	// 测试创建菜单
	parentID := uint32(0)
	menu := CreateTestMenuModel(10, &parentID)
	
	// 创建权限模型
	perms := []biz.PermissionModel{
		{
			StandardModel: database.StandardModel{
				BaseModel: database.BaseModel{
					ID: 1,
				},
			},
			URL:    "/test/permission/1",
			Method: "GET",
			Label:  "test_perm_1",
			Descr:  "测试权限1",
		},
		{
			StandardModel: database.StandardModel{
				BaseModel: database.BaseModel{
					ID: 2,
				},
			},
			URL:    "/test/permission/2",
			Method: "POST",
			Label:  "test_perm_2",
			Descr:  "测试权限2",
		},
	}
	
	// 创建菜单并关联权限
	err := suite.menuRepo.CreateModel(context.Background(), menu, &perms)
	suite.NoError(err, "创建菜单并关联权限应该成功")
	
	// 测试添加组策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "添加菜单组策略应该成功")
}

func (suite *MenuTestSuite) TestRemoveGroupPolicy() {
	// 测试创建菜单
	parentID := uint32(0)
	menu := CreateTestMenuModel(11, &parentID)
	
	// 创建权限模型
	perms := []biz.PermissionModel{
		{
			StandardModel: database.StandardModel{
				BaseModel: database.BaseModel{
					ID: 3,
				},
			},
			URL:    "/test/permission/3",
			Method: "GET",
			Label:  "test_perm_3",
			Descr:  "测试权限3",
		},
	}
	
	// 创建菜单并关联权限
	err := suite.menuRepo.CreateModel(context.Background(), menu, &perms)
	suite.NoError(err, "创建菜单并关联权限应该成功")
	
	// 测试添加组策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "添加菜单组策略应该成功")
	
	// 测试删除组策略
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), menu, true)
	suite.NoError(err, "删除菜单组策略应该成功")
}

// 每个测试文件都需要这个入口函数
func TestMenuTestSuite(t *testing.T) {
	pts := &MenuTestSuite{}
	suite.Run(t, pts)
}
