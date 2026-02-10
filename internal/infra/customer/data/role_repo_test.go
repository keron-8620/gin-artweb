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

// CreateTestRoleModel 创建测试用的角色模型
func CreateTestRoleModel(pk uint32) *biz.RoleModel {
	return &biz.RoleModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: pk,
			},
		},
		Name:  "test_role_" + string(rune('0'+pk)),
		Descr: "这是一个测试角色" + string(rune('0'+pk)),
	}
}

type RoleTestSuite struct {
	suite.Suite
	roleRepo *roleRepo
}

func (suite *RoleTestSuite) SetupSuite() {
	suite.roleRepo = NewTestRoleRepo()
}

func (suite *RoleTestSuite) TestCreateRole() {
	// 测试创建角色
	sm := CreateTestRoleModel(1)
	perms := []biz.PermissionModel{}
	menus := []biz.MenuModel{}
	buttons := []biz.ButtonModel{}
	err := suite.roleRepo.CreateModel(context.Background(), sm, &perms, &menus, &buttons)
	suite.NoError(err, "创建角色应该成功")

	// 测试查询刚创建的角色
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的角色应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(sm.Name, fm.Name)
	suite.Equal(sm.Descr, fm.Descr)
}

func (suite *RoleTestSuite) TestUpdateRole() {
	// 测试创建角色
	sm := CreateTestRoleModel(2)
	err := suite.roleRepo.CreateModel(context.Background(), sm, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 测试更新角色
	updatedName := "更新的测试角色"
	updatedDescr := "这是一个更新的测试角色"

	err = suite.roleRepo.UpdateModel(context.Background(), map[string]any{
		"name":  updatedName,
		"descr": updatedDescr,
	}, nil, nil, nil, "id = ?", sm.ID)
	suite.NoError(err, "更新角色应该成功")

	// 测试查询更新后的角色
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的角色应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Greater(fm.UpdatedAt, sm.UpdatedAt)
}

func (suite *RoleTestSuite) TestDeleteRole() {
	// 测试创建角色
	sm := CreateTestRoleModel(3)
	perms := []biz.PermissionModel{}
	menus := []biz.MenuModel{}
	buttons := []biz.ButtonModel{}
	err := suite.roleRepo.CreateModel(context.Background(), sm, &perms, &menus, &buttons)
	suite.NoError(err, "创建角色应该成功")

	// 测试查询刚创建的角色
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的角色应该成功")
	suite.Equal(sm.ID, fm.ID)

	// 测试删除角色
	err = suite.roleRepo.DeleteModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "删除角色应该成功")

	// 测试查询已删除的角色
	_, err = suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *RoleTestSuite) TestGetRoleByID() {
	// 测试创建角色
	sm := CreateTestRoleModel(4)
	perms := []biz.PermissionModel{}
	menus := []biz.MenuModel{}
	buttons := []biz.ButtonModel{}
	err := suite.roleRepo.CreateModel(context.Background(), sm, &perms, &menus, &buttons)
	suite.NoError(err, "创建角色应该成功")

	// 测试根据ID查询角色
	m, err := suite.roleRepo.GetModel(context.Background(), []string{}, sm.ID)
	suite.NoError(err, "根据ID查询角色应该成功")
	suite.Equal(sm.ID, m.ID)
	suite.Equal(sm.Name, m.Name)
	suite.Equal(sm.Descr, m.Descr)
}

func (suite *RoleTestSuite) TestListRoles() {
	// 测试创建多个角色
	for i := 5; i < 10; i++ {
		sm := CreateTestRoleModel(uint32(i))
		perms := []biz.PermissionModel{}
		menus := []biz.MenuModel{}
		buttons := []biz.ButtonModel{}
		err := suite.roleRepo.CreateModel(context.Background(), sm, &perms, &menus, &buttons)
		suite.NoError(err, "创建角色应该成功")
	}

	// 测试查询角色列表
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	count, ms, err := suite.roleRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询角色列表应该成功")
	suite.NotNil(ms, "角色列表不应该为nil")
	suite.GreaterOrEqual(count, int64(5), "角色总数应该至少有5条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:   2,
		Offset:  0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.roleRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询角色列表应该成功")
	suite.NotNil(pMs, "分页角色列表不应该为nil")
	suite.Equal(2, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(2), "分页总数应该至少等于limit")
}

func (suite *RoleTestSuite) TestAddGroupPolicy() {
	// 测试创建角色
	role := CreateTestRoleModel(10)

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
	}

	// 创建菜单模型
	menus := []biz.MenuModel{
		{
			StandardModel: database.StandardModel{
				BaseModel: database.BaseModel{
					ID: 1,
				},
			},
			Name:  "test_menu_1",
			Descr: "测试菜单1",
		},
	}

	// 创建按钮模型
	buttons := []biz.ButtonModel{
		{
			StandardModel: database.StandardModel{
				BaseModel: database.BaseModel{
					ID: 1,
				},
			},
			MenuID: 1,
			Name:   "test_button_1",
			Descr:  "测试按钮1",
		},
	}

	// 创建角色并关联权限、菜单和按钮
	err := suite.roleRepo.CreateModel(context.Background(), role, &perms, &menus, &buttons)
	suite.NoError(err, "创建角色并关联权限、菜单和按钮应该成功")

	// 测试添加组策略
	err = suite.roleRepo.AddGroupPolicy(context.Background(), role)
	suite.NoError(err, "添加角色组策略应该成功")
}

func (suite *RoleTestSuite) TestRemoveGroupPolicy() {
	// 测试创建角色
	role := CreateTestRoleModel(11)

	// 创建权限模型
	perms := []biz.PermissionModel{
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

	// 创建菜单模型
	menus := []biz.MenuModel{
		{
			StandardModel: database.StandardModel{
				BaseModel: database.BaseModel{
					ID: 2,
				},
			},
			Name:  "test_menu_2",
			Descr: "测试菜单2",
		},
	}

	// 创建按钮模型
	buttons := []biz.ButtonModel{
		{
			StandardModel: database.StandardModel{
				BaseModel: database.BaseModel{
					ID: 2,
				},
			},
			MenuID: 2,
			Descr:  "测试按钮2",
		},
	}

	// 创建角色并关联权限、菜单和按钮
	err := suite.roleRepo.CreateModel(context.Background(), role, &perms, &menus, &buttons)
	suite.NoError(err, "创建角色并关联权限、菜单和按钮应该成功")

	// 测试添加组策略
	err = suite.roleRepo.AddGroupPolicy(context.Background(), role)
	suite.NoError(err, "添加角色组策略应该成功")

	// 测试删除组策略
	err = suite.roleRepo.RemoveGroupPolicy(context.Background(), role)
	suite.NoError(err, "删除角色组策略应该成功")
}

// 每个测试文件都需要这个入口函数
func TestRoleTestSuite(t *testing.T) {
	pts := &RoleTestSuite{}
	suite.Run(t, pts)
}
