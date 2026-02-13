package biz

import (
	"context"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/infra/customer/data"
	"gin-artweb/internal/infra/customer/model"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

// CreateTestRoleModel 创建测试用的角色模型
func CreateTestRoleModel() *model.RoleModel {
	return &model.RoleModel{
		Name:  uuid.NewString(),
		Descr: "这是一个测试角色",
	}
}

type RoleTestSuite struct {
	suite.Suite
	enforcer *casbin.Enforcer
	uc       *RoleUsecase
}

func (suite *RoleTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&model.MenuModel{},
		&model.ApiModel{},
		&model.ButtonModel{},
		&model.RoleModel{},
	)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	suite.enforcer = enforcer
	suite.uc = &RoleUsecase{
		log: logger,
		apiRepo: data.NewApiRepo(
			logger,
			db,
			dbTimeout,
			enforcer,
		),
		menuRepo: data.NewMenuRepo(
			logger,
			db,
			dbTimeout,
			enforcer,
		),
		buttonRepo: data.NewButtonRepo(
			logger,
			db,
			dbTimeout,
			enforcer,
		),
		roleRepo: data.NewRoleRepo(
			logger,
			db,
			dbTimeout,
			enforcer,
		),
	}
}

// 每个测试文件都需要这个入口函数
func TestRoleTestSuite(t *testing.T) {
	pts := &RoleTestSuite{}
	suite.Run(t, pts)
}

// TestGetApis 测试获取API列表
func (suite *RoleTestSuite) TestGetApis() {
	// 创建测试API
	apiCount := 2
	apiIDs := make([]uint32, 0, apiCount)
	for i := 0; i < apiCount; i++ {
		testApi := CreateTestApiModel()
		err := suite.uc.apiRepo.CreateModel(context.Background(), testApi)
		suite.Nil(err, "创建API应该成功")
		apiIDs = append(apiIDs, testApi.ID)
	}

	// 测试获取API列表
	apis, err := suite.uc.GetApis(context.Background(), apiIDs)
	suite.Nil(err, "获取API列表应该成功")
	suite.Len(*apis, apiCount, "API列表数量应该正确")
}

// TestGetMenus 测试获取菜单列表
func (suite *RoleTestSuite) TestGetMenus() {
	// 创建测试菜单
	menuCount := 2
	menuIDs := make([]uint32, 0, menuCount)
	for i := 0; i < menuCount; i++ {
		testMenu := CreateTestMenuModel(nil)
		err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
		suite.Nil(err, "创建菜单应该成功")
		menuIDs = append(menuIDs, testMenu.ID)
	}

	// 测试获取菜单列表
	menus, err := suite.uc.GetMenus(context.Background(), menuIDs)
	suite.Nil(err, "获取菜单列表应该成功")
	suite.Len(*menus, menuCount, "菜单列表数量应该正确")
}

// TestGetButtons 测试获取按钮列表
func (suite *RoleTestSuite) TestGetButtons() {
	// 创建测试菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")

	// 创建测试按钮
	buttonCount := 2
	buttonIDs := make([]uint32, 0, buttonCount)
	for i := 0; i < buttonCount; i++ {
		testButton := CreateTestButtonModel(testMenu.ID)
		err := suite.uc.buttonRepo.CreateModel(context.Background(), testButton, nil)
		suite.Nil(err, "创建按钮应该成功")
		buttonIDs = append(buttonIDs, testButton.ID)
	}

	// 测试获取按钮列表
	buttons, err := suite.uc.GetButtons(context.Background(), buttonIDs)
	suite.Nil(err, "获取按钮列表应该成功")
	suite.Len(*buttons, buttonCount, "按钮列表数量应该正确")
}

// TestFindRoleByID 测试根据ID查询角色
func (suite *RoleTestSuite) TestFindRoleByID() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	createdRole, err := suite.uc.CreateRole(context.Background(), []uint32{}, []uint32{}, []uint32{}, *testRole)
	suite.Nil(err, "创建角色应该成功")

	// 测试查询角色
	foundRole, err := suite.uc.FindRoleByID(context.Background(), []string{}, createdRole.ID)
	suite.Nil(err, "查询角色应该成功")
	suite.NotNil(foundRole, "角色不应该为空")
	suite.Equal(createdRole.ID, foundRole.ID, "角色ID应该匹配")
}

// TestListRole 测试查询角色列表
func (suite *RoleTestSuite) TestListRole() {
	// 创建测试角色
	roleCount := 2
	for i := 0; i < roleCount; i++ {
		testRole := CreateTestRoleModel()
		_, err := suite.uc.CreateRole(context.Background(), []uint32{}, []uint32{}, []uint32{}, *testRole)
		suite.Nil(err, "创建角色应该成功")
	}

	// 测试查询角色列表
	count, roles, err := suite.uc.ListRole(context.Background(), database.QueryParams{})
	suite.Nil(err, "查询角色列表应该成功")
	suite.GreaterOrEqual(int(count), roleCount, "角色数量应该大于或等于创建的数量")
	suite.NotNil(roles, "角色列表不应该为空")
}

// TestLoadRolePolicy 测试加载角色策略
func (suite *RoleTestSuite) TestLoadRolePolicy() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	_, err := suite.uc.CreateRole(context.Background(), []uint32{}, []uint32{}, []uint32{}, *testRole)
	suite.Nil(err, "创建角色应该成功")

	// 测试加载角色策略
	err = suite.uc.LoadRolePolicy(context.Background())
	suite.Nil(err, "加载角色策略应该成功")
}

// TestGetRoleMenuTree 测试获取角色菜单树
func (suite *RoleTestSuite) TestGetRoleMenuTree() {
	// 创建测试菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")

	// 创建测试角色并关联菜单
	testRole := CreateTestRoleModel()
	createdRole, err := suite.uc.CreateRole(context.Background(), []uint32{}, []uint32{testMenu.ID}, []uint32{}, *testRole)
	suite.Nil(err, "创建角色应该成功")

	// 测试获取角色菜单树
	menuTree, err := suite.uc.GetRoleMenuTree(context.Background(), createdRole.ID)
	suite.Nil(err, "获取角色菜单树应该成功")
	suite.NotNil(menuTree, "菜单树不应该为空")
}

// TestCreateRole 测试创建角色（无关联）
func (suite *RoleTestSuite) TestCreateRole() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	createdRole, err := suite.uc.CreateRole(context.Background(), []uint32{}, []uint32{}, []uint32{}, *testRole)
	suite.Nil(err, "创建角色应该成功")
	suite.NotNil(createdRole, "角色不应该为空")
	suite.Equal(testRole.Name, createdRole.Name, "角色名称应该匹配")
	suite.Equal(testRole.Descr, createdRole.Descr, "角色描述应该匹配")
}

// TestCreateRoleWithRelations 测试创建角色（关联 API、菜单、按钮）
func (suite *RoleTestSuite) TestCreateRoleWithRelations() {
	// 创建测试API
	testApi := CreateTestApiModel()
	err := suite.uc.apiRepo.CreateModel(context.Background(), testApi)
	suite.Nil(err, "创建API应该成功")

	// 创建测试菜单
	testMenu := CreateTestMenuModel(nil)
	err = suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")

	// 创建测试按钮
	testButton := CreateTestButtonModel(testMenu.ID)
	err = suite.uc.buttonRepo.CreateModel(context.Background(), testButton, nil)
	suite.Nil(err, "创建按钮应该成功")

	// 创建测试角色并关联API、菜单、按钮
	testRole := CreateTestRoleModel()
	createdRole, err := suite.uc.CreateRole(context.Background(), []uint32{testApi.ID}, []uint32{testMenu.ID}, []uint32{testButton.ID}, *testRole)
	suite.Nil(err, "创建角色应该成功")
	suite.NotNil(createdRole, "角色不应该为空")

	// 验证关联关系
	foundRole, err := suite.uc.FindRoleByID(context.Background(), []string{"Apis", "Menus", "Buttons"}, createdRole.ID)
	suite.Nil(err, "查询角色应该成功")
	suite.Len(foundRole.Apis, 1, "角色关联的API数量应该正确")
	suite.Len(foundRole.Menus, 1, "角色关联的菜单数量应该正确")
	suite.Len(foundRole.Buttons, 1, "角色关联的按钮数量应该正确")
}

// TestUpdateRoleByID 测试更新角色
func (suite *RoleTestSuite) TestUpdateRoleByID() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	createdRole, err := suite.uc.CreateRole(context.Background(), []uint32{}, []uint32{}, []uint32{}, *testRole)
	suite.Nil(err, "创建角色应该成功")

	// 更新角色
	updatedName := uuid.NewString()
	updatedDescr := "更新后的角色描述"
	updatedRole, err := suite.uc.UpdateRoleByID(context.Background(), createdRole.ID, []uint32{}, []uint32{}, []uint32{}, map[string]any{
		"name":  updatedName,
		"descr": updatedDescr,
	})
	suite.Nil(err, "更新角色应该成功")
	suite.NotNil(updatedRole, "角色不应该为空")
	suite.Equal(updatedName, updatedRole.Name, "角色名称应该更新")
	suite.Equal(updatedDescr, updatedRole.Descr, "角色描述应该更新")
}

// TestDeleteRoleByID 测试删除角色
func (suite *RoleTestSuite) TestDeleteRoleByID() {
	// 创建测试角色
	testRole := CreateTestRoleModel()
	createdRole, err := suite.uc.CreateRole(context.Background(), []uint32{}, []uint32{}, []uint32{}, *testRole)
	suite.Nil(err, "创建角色应该成功")

	// 删除角色
	err = suite.uc.DeleteRoleByID(context.Background(), createdRole.ID)
	suite.Nil(err, "删除角色应该成功")

	// 验证角色已删除
	_, err = suite.uc.FindRoleByID(context.Background(), []string{}, createdRole.ID)
	suite.NotNil(err, "查询已删除的角色应该失败")
}

// TestGetApisWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestGetApisWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.GetApis(ctx, []uint32{1})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestGetMenusWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestGetMenusWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.GetMenus(ctx, []uint32{1})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestGetButtonsWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestGetButtonsWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.GetButtons(ctx, []uint32{1})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestCreateRoleWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestCreateRoleWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	testRole := CreateTestRoleModel()
	_, err := suite.uc.CreateRole(ctx, []uint32{}, []uint32{}, []uint32{}, *testRole)
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestUpdateRoleByIDWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestUpdateRoleByIDWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.UpdateRoleByID(ctx, 1, []uint32{}, []uint32{}, []uint32{}, map[string]any{})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestDeleteRoleByIDWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestDeleteRoleByIDWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	err := suite.uc.DeleteRoleByID(ctx, 1)
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestFindRoleByIDWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestFindRoleByIDWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.FindRoleByID(ctx, []string{}, 1)
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestListRoleWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestListRoleWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, _, err := suite.uc.ListRole(ctx, database.QueryParams{})
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestLoadRolePolicyWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestLoadRolePolicyWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	err := suite.uc.LoadRolePolicy(ctx)
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestGetRoleMenuTreeWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestGetRoleMenuTreeWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文错误
	_, err := suite.uc.GetRoleMenuTree(ctx, 1)
	suite.NotNil(err, "上下文错误应该返回错误")
}

// TestRoleApiRelations 测试角色与API的多对多关系
func (suite *RoleTestSuite) TestRoleApiRelations() {
	// 创建多个测试API
	apiCount := 3
	apiIDs := make([]uint32, 0, apiCount)
	for i := 0; i < apiCount; i++ {
		testApi := CreateTestApiModel()
		err := suite.uc.apiRepo.CreateModel(context.Background(), testApi)
		suite.Nil(err, "创建API应该成功")
		apiIDs = append(apiIDs, testApi.ID)
	}

	// 创建角色并关联API
	testRole := CreateTestRoleModel()
	createdRole, err := suite.uc.CreateRole(context.Background(), apiIDs, []uint32{}, []uint32{}, *testRole)
	suite.Nil(err, "创建角色应该成功")

	// 验证关联关系
	foundRole, err := suite.uc.FindRoleByID(context.Background(), []string{"Apis"}, createdRole.ID)
	suite.Nil(err, "查询角色应该成功")
	suite.Len(foundRole.Apis, apiCount, "角色关联的API数量应该正确")
}

// TestRoleMenuRelations 测试角色与菜单的多对多关系
func (suite *RoleTestSuite) TestRoleMenuRelations() {
	// 创建多个测试菜单
	menuCount := 3
	menuIDs := make([]uint32, 0, menuCount)
	for i := 0; i < menuCount; i++ {
		testMenu := CreateTestMenuModel(nil)
		err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
		suite.Nil(err, "创建菜单应该成功")
		menuIDs = append(menuIDs, testMenu.ID)
	}

	// 创建角色并关联菜单
	testRole := CreateTestRoleModel()
	createdRole, err := suite.uc.CreateRole(context.Background(), []uint32{}, menuIDs, []uint32{}, *testRole)
	suite.Nil(err, "创建角色应该成功")

	// 验证关联关系
	foundRole, err := suite.uc.FindRoleByID(context.Background(), []string{"Menus"}, createdRole.ID)
	suite.Nil(err, "查询角色应该成功")
	suite.Len(foundRole.Menus, menuCount, "角色关联的菜单数量应该正确")
}

// TestRoleButtonRelations 测试角色与按钮的多对多关系
func (suite *RoleTestSuite) TestRoleButtonRelations() {
	// 创建测试菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")

	// 创建多个测试按钮
	buttonCount := 3
	buttonIDs := make([]uint32, 0, buttonCount)
	for i := 0; i < buttonCount; i++ {
		testButton := CreateTestButtonModel(testMenu.ID)
		err := suite.uc.buttonRepo.CreateModel(context.Background(), testButton, nil)
		suite.Nil(err, "创建按钮应该成功")
		buttonIDs = append(buttonIDs, testButton.ID)
	}

	// 创建角色并关联按钮
	testRole := CreateTestRoleModel()
	createdRole, err := suite.uc.CreateRole(context.Background(), []uint32{}, []uint32{testMenu.ID}, buttonIDs, *testRole)
	suite.Nil(err, "创建角色应该成功")

	// 验证关联关系
	foundRole, err := suite.uc.FindRoleByID(context.Background(), []string{"Buttons"}, createdRole.ID)
	suite.Nil(err, "查询角色应该成功")
	suite.Len(foundRole.Buttons, buttonCount, "角色关联的按钮数量应该正确")
}

// TestRoleCasbinInheritance 测试角色对 API、菜单、按钮的权限继承
func (suite *RoleTestSuite) TestRoleCasbinInheritance() {
	// 创建测试API
	testApi := CreateTestApiModel()
	err := suite.uc.apiRepo.CreateModel(context.Background(), testApi)
	suite.Nil(err, "创建API应该成功")

	// 创建测试菜单
	testMenu := CreateTestMenuModel(nil)
	err = suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")

	// 创建测试按钮
	testButton := CreateTestButtonModel(testMenu.ID)
	err = suite.uc.buttonRepo.CreateModel(context.Background(), testButton, nil)
	suite.Nil(err, "创建按钮应该成功")

	// 创建角色并关联API、菜单、按钮
	testRole := CreateTestRoleModel()
	createdRole, err := suite.uc.CreateRole(context.Background(), []uint32{testApi.ID}, []uint32{testMenu.ID}, []uint32{testButton.ID}, *testRole)
	suite.Nil(err, "创建角色应该成功")

	// 加载角色策略
	err = suite.uc.LoadRolePolicy(context.Background())
	suite.Nil(err, "加载角色策略应该成功")

	// 验证策略添加成功（通过查询角色及其关联资源）
	foundRole, err := suite.uc.FindRoleByID(context.Background(), []string{"Apis", "Menus", "Buttons"}, createdRole.ID)
	suite.Nil(err, "查询角色应该成功")
	suite.Len(foundRole.Apis, 1, "角色应该关联API")
	suite.Len(foundRole.Menus, 1, "角色应该关联菜单")
	suite.Len(foundRole.Buttons, 1, "角色应该关联按钮")
}
