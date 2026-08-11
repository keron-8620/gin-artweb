package sys

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	sysmodel "gin-artweb/internal/model/sys"
	syssvc "gin-artweb/internal/repo/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/test"
)

// CreateTestRoleModel 创建测试用的角色模型
func CreateTestRoleModel() *sysmodel.RoleModel {
	return &sysmodel.RoleModel{
		Name:  uuid.NewString(),
		Descr: "这是一个测试角色",
	}
}

// CreateTestRoleDTO 创建测试用的角色DTO
func CreateTestRoleDTO(apiIDs, menuIDs, buttonIDs []uint32) sysmodel.RoleUpsertDTO {
	return sysmodel.RoleUpsertDTO{
		Name:      uuid.NewString(),
		Descr:     "这是一个测试角色",
		ApiIDs:    apiIDs,
		MenuIDs:   menuIDs,
		ButtonIDs: buttonIDs,
	}
}

// CreateTestApis 创建多个测试API
func (suite *RoleTestSuite) CreateTestApis(count int) []uint32 {
	apiIDs := make([]uint32, 0, count)
	for i := 0; i < count; i++ {
		testApi := CreateTestApiModel()
		err := suite.roleservice.apiRepo.CreateModel(context.Background(), testApi)
		suite.Nil(err, "创建API应该成功")
		apiIDs = append(apiIDs, testApi.ID)
	}
	return apiIDs
}

// CreateTestMenus 创建多个测试菜单
func (suite *RoleTestSuite) CreateTestMenus(count int) []uint32 {
	menuIDs := make([]uint32, 0, count)
	for i := 0; i < count; i++ {
		testMenu := CreateTestMenuModel(nil)
		err := suite.roleservice.menuRepo.CreateModel(context.Background(), testMenu, nil)
		suite.Nil(err, "创建菜单应该成功")
		menuIDs = append(menuIDs, testMenu.ID)
	}
	return menuIDs
}

// CreateTestButtons 创建多个测试按钮
func (suite *RoleTestSuite) CreateTestButtons(menuID uint32, count int) []uint32 {
	buttonIDs := make([]uint32, 0, count)
	for i := 0; i < count; i++ {
		testButton := CreateTestButtonModel(menuID)
		err := suite.roleservice.buttonRepo.CreateModel(context.Background(), testButton, nil)
		suite.Nil(err, "创建按钮应该成功")
		buttonIDs = append(buttonIDs, testButton.ID)
	}
	return buttonIDs
}

type RoleTestSuite struct {
	suite.Suite
	roleservice *RoleService
}

func (suite *RoleTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(
		&sysmodel.MenuModel{},
		&sysmodel.ApiModel{},
		&sysmodel.ButtonModel{},
		&sysmodel.RoleModel{},
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
	suite.roleservice = NewRoleService(
		logger,
		syssvc.NewApiRepo(
			logger,
			db,
			dbTimeout,
			slowThreshold,
			enforcer,
		),
		syssvc.NewMenuRepo(
			logger,
			db,
			dbTimeout,
			slowThreshold,
			enforcer,
		),
		syssvc.NewButtonRepo(
			logger,
			db,
			dbTimeout,
			slowThreshold,
			enforcer,
		),
		syssvc.NewRoleRepo(
			logger,
			db,
			dbTimeout,
			slowThreshold,
			enforcer,
		),
	)
}

// 每个测试文件都需要这个入口函数
func TestRoleTestSuite(t *testing.T) {
	pts := &RoleTestSuite{}
	suite.Run(t, pts)
}

// TestGetApis 测试获取API列表
func (suite *RoleTestSuite) TestGetApis() {
	// 创建测试API
	apiIDs := suite.CreateTestApis(2)

	// 测试获取API列表
	apis, err := suite.roleservice.GetApis(context.Background(), apiIDs)
	suite.Nil(err, "获取API列表应该成功")
	suite.Len(apis, 2, "API列表数量应该正确")
}

// TestGetMenus 测试获取菜单列表
func (suite *RoleTestSuite) TestGetMenus() {
	// 创建测试菜单
	menuIDs := suite.CreateTestMenus(2)

	// 测试获取菜单列表
	menus, err := suite.roleservice.GetMenus(context.Background(), menuIDs)
	suite.Nil(err, "获取菜单列表应该成功")
	suite.Len(menus, 2, "菜单列表数量应该正确")
}

// TestGetButtons 测试获取按钮列表
func (suite *RoleTestSuite) TestGetButtons() {
	// 创建测试菜单
	menuIDs := suite.CreateTestMenus(1)

	// 创建测试按钮
	buttonIDs := suite.CreateTestButtons(menuIDs[0], 2)

	// 测试获取按钮列表
	buttons, err := suite.roleservice.GetButtons(context.Background(), buttonIDs)
	suite.Nil(err, "获取按钮列表应该成功")
	suite.Len(buttons, 2, "按钮列表数量应该正确")
}

// TestFindRoleByID 测试根据ID查询角色
func (suite *RoleTestSuite) TestFindRoleByID() {
	// 创建测试角色
	createdRole, err := suite.roleservice.CreateRole(context.Background(), CreateTestRoleDTO([]uint32{}, []uint32{}, []uint32{}))
	suite.Nil(err, "创建角色应该成功")

	// 测试查询角色
	foundRole, err := suite.roleservice.FindRoleByID(context.Background(), []string{}, createdRole.ID)
	suite.Nil(err, "查询角色应该成功")
	suite.NotNil(foundRole, "角色不应该为空")
	suite.Equal(createdRole.ID, foundRole.ID, "角色ID应该匹配")
}

// TestListRole 测试查询角色列表
func (suite *RoleTestSuite) TestListRole() {
	// 创建测试角色
	roleCount := 2
	for i := 0; i < roleCount; i++ {
		_, err := suite.roleservice.CreateRole(context.Background(), CreateTestRoleDTO([]uint32{}, []uint32{}, []uint32{}))
		suite.Nil(err, "创建角色应该成功")
	}

	// 测试查询角色列表
	listDTO := sysmodel.ListRoleDTO{}
	_, _, count, roles, err := suite.roleservice.ListRole(context.Background(), listDTO)
	suite.Nil(err, "查询角色列表应该成功")
	suite.GreaterOrEqual(int(count), roleCount, "角色数量应该大于或等于创建的数量")
	suite.NotNil(roles, "角色列表不应该为空")
}

// TestLoadRolePolicy 测试加载角色策略
func (suite *RoleTestSuite) TestLoadRolePolicy() {
	// 创建测试角色
	_, rErr := suite.roleservice.CreateRole(context.Background(), CreateTestRoleDTO([]uint32{}, []uint32{}, []uint32{}))
	suite.Nil(rErr, "创建角色应该成功")

	// 测试加载角色策略
	loadErr := suite.roleservice.LoadRolePolicy(context.Background())
	suite.Nil(loadErr, "加载角色策略应该成功")
}

// TestGetRoleMenuTree 测试获取角色菜单树
func (suite *RoleTestSuite) TestGetRoleMenuTree() {
	// 创建测试菜单
	menuIDs := suite.CreateTestMenus(1)

	// 创建测试角色并关联菜单
	createdRole, err := suite.roleservice.CreateRole(context.Background(), CreateTestRoleDTO([]uint32{}, menuIDs, []uint32{}))
	suite.Nil(err, "创建角色应该成功")

	// 测试获取角色菜单树
	menuTree, err := suite.roleservice.GetRoleMenuTree(context.Background(), createdRole.ID)
	suite.Nil(err, "获取角色菜单树应该成功")
	suite.NotNil(menuTree, "菜单树不应该为空")
}

// TestGetRoleMenuTreeWithNestedMenus 测试获取角色菜单树（包含嵌套菜单和按钮）
func (suite *RoleTestSuite) TestGetRoleMenuTreeWithNestedMenus() {
	// 创建父菜单
	parentMenu := CreateTestMenuModel(nil)
	err := suite.roleservice.menuRepo.CreateModel(context.Background(), parentMenu, nil)
	suite.Nil(err, "创建父菜单应该成功")

	// 创建子菜单，关联到父菜单
	childMenu := CreateTestMenuModel(&parentMenu.ID)
	err = suite.roleservice.menuRepo.CreateModel(context.Background(), childMenu, nil)
	suite.Nil(err, "创建子菜单应该成功")

	// 创建按钮，关联到子菜单
	buttonIDs := suite.CreateTestButtons(childMenu.ID, 1)

	// 创建测试角色并关联所有菜单和按钮
	createdRole, err := suite.roleservice.CreateRole(context.Background(),
		CreateTestRoleDTO(
			[]uint32{},
			[]uint32{parentMenu.ID, childMenu.ID},
			buttonIDs,
		),
	)
	suite.Nil(err, "创建角色应该成功")

	// 测试获取角色菜单树
	menuTree, err := suite.roleservice.GetRoleMenuTree(context.Background(), createdRole.ID)
	suite.Nil(err, "获取角色菜单树应该成功")
	suite.NotNil(menuTree, "菜单树不应该为空")
	suite.Greater(len(menuTree), 0, "菜单树应该包含至少一个菜单")

	// 验证父菜单包含子菜单
	parentTreeNode := menuTree[0]
	suite.Greater(len(parentTreeNode.Children), 0, "父菜单应该包含子菜单")

	// 验证子菜单包含按钮
	childTreeNode := parentTreeNode.Children[0]
	suite.Greater(len(childTreeNode.Buttons), 0, "子菜单应该包含按钮")
}

// TestCreateRole 测试创建角色（无关联）
func (suite *RoleTestSuite) TestCreateRole() {
	// 创建测试角色
	testRole := CreateTestRoleDTO([]uint32{}, []uint32{}, []uint32{})
	createdRole, err := suite.roleservice.CreateRole(context.Background(), testRole)
	suite.Nil(err, "创建角色应该成功")
	suite.NotNil(createdRole, "角色不应该为空")
	suite.Equal(testRole.Name, createdRole.Name, "角色名称应该匹配")
	suite.Equal(testRole.Descr, createdRole.Descr, "角色描述应该匹配")
}

// TestCreateRoleWithRelations 测试创建角色（关联 API、菜单、按钮）
func (suite *RoleTestSuite) TestCreateRoleWithRelations() {
	// 创建测试资源
	apiIDs := suite.CreateTestApis(1)
	menuIDs := suite.CreateTestMenus(1)
	buttonIDs := suite.CreateTestButtons(menuIDs[0], 1)

	// 创建测试角色并关联API、菜单、按钮
	testRole := CreateTestRoleDTO(apiIDs, menuIDs, buttonIDs)
	createdRole, err := suite.roleservice.CreateRole(context.Background(), testRole)
	suite.Nil(err, "创建角色应该成功")
	suite.NotNil(createdRole, "角色不应该为空")

	// 验证关联关系
	foundRole, err := suite.roleservice.FindRoleByID(context.Background(), []string{"Apis", "Menus", "Buttons"}, createdRole.ID)
	suite.Nil(err, "查询角色应该成功")
	suite.Len(foundRole.Apis, 1, "角色关联的API数量应该正确")
	suite.Len(foundRole.Menus, 1, "角色关联的菜单数量应该正确")
	suite.Len(foundRole.Buttons, 1, "角色关联的按钮数量应该正确")
}

// TestUpdateRoleByID 测试更新角色
func (suite *RoleTestSuite) TestUpdateRoleByID() {
	// 创建测试角色
	createdRole, err := suite.roleservice.CreateRole(context.Background(), CreateTestRoleDTO([]uint32{}, []uint32{}, []uint32{}))
	suite.Nil(err, "创建角色应该成功")

	// 更新角色
	updatedName := uuid.NewString()
	updatedDescr := "更新后的角色描述"
	updateDTO := sysmodel.RoleUpsertDTO{
		Name:      updatedName,
		Descr:     updatedDescr,
		ApiIDs:    []uint32{},
		MenuIDs:   []uint32{},
		ButtonIDs: []uint32{},
	}
	updatedRole, err := suite.roleservice.UpdateRoleByID(context.Background(), createdRole.ID, updateDTO)
	suite.Nil(err, "更新角色应该成功")
	suite.NotNil(updatedRole, "角色不应该为空")
	suite.Equal(updatedName, updatedRole.Name, "角色名称应该更新")
	suite.Equal(updatedDescr, updatedRole.Descr, "角色描述应该更新")
}

// TestDeleteRoleByID 测试删除角色
func (suite *RoleTestSuite) TestDeleteRoleByID() {
	// 创建测试角色
	createdRole, err := suite.roleservice.CreateRole(context.Background(), CreateTestRoleDTO([]uint32{}, []uint32{}, []uint32{}))
	suite.Nil(err, "创建角色应该成功")

	// 删除角色
	err = suite.roleservice.DeleteRoleByID(context.Background(), createdRole.ID)
	suite.Nil(err, "删除角色应该成功")

	// 验证角色已删除
	_, err = suite.roleservice.FindRoleByID(context.Background(), []string{}, createdRole.ID)
	suite.NotNil(err, "查询已删除的角色应该失败")
}

// TestWithContextError 测试上下文错误处理
func (suite *RoleTestSuite) TestWithContextError() {
	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试GetApis上下文错误
	_, err := suite.roleservice.GetApis(ctx, []uint32{1})
	suite.NotNil(err, "GetApis上下文错误应该返回错误")

	// 测试GetMenus上下文错误
	_, err = suite.roleservice.GetMenus(ctx, []uint32{1})
	suite.NotNil(err, "GetMenus上下文错误应该返回错误")

	// 测试GetButtons上下文错误
	_, err = suite.roleservice.GetButtons(ctx, []uint32{1})
	suite.NotNil(err, "GetButtons上下文错误应该返回错误")

	// 测试CreateRole上下文错误
	_, err = suite.roleservice.CreateRole(ctx, CreateTestRoleDTO([]uint32{}, []uint32{}, []uint32{}))
	suite.NotNil(err, "CreateRole上下文错误应该返回错误")

	// 测试UpdateRoleByID上下文错误
	_, err = suite.roleservice.UpdateRoleByID(ctx, 1, sysmodel.RoleUpsertDTO{})
	suite.NotNil(err, "UpdateRoleByID上下文错误应该返回错误")

	// 测试DeleteRoleByID上下文错误
	err = suite.roleservice.DeleteRoleByID(ctx, 1)
	suite.NotNil(err, "DeleteRoleByID上下文错误应该返回错误")

	// 测试FindRoleByID上下文错误
	_, err = suite.roleservice.FindRoleByID(ctx, []string{}, 1)
	suite.NotNil(err, "FindRoleByID上下文错误应该返回错误")

	// 测试ListRole上下文错误
	_, _, _, _, err = suite.roleservice.ListRole(ctx, sysmodel.ListRoleDTO{})
	suite.NotNil(err, "ListRole上下文错误应该返回错误")

	// 测试LoadRolePolicy上下文错误
	loadErr := suite.roleservice.LoadRolePolicy(ctx)
	suite.NotNil(loadErr, "LoadRolePolicy上下文错误应该返回错误")

	// 测试GetRoleMenuTree上下文错误
	_, err = suite.roleservice.GetRoleMenuTree(ctx, 1)
	suite.NotNil(err, "GetRoleMenuTree上下文错误应该返回错误")
}

// TestRoleResourceRelations 测试角色与资源的多对多关系
func (suite *RoleTestSuite) TestRoleResourceRelations() {
	// 创建测试资源
	apiIDs := suite.CreateTestApis(3)
	menuIDs := suite.CreateTestMenus(3)
	buttonIDs := suite.CreateTestButtons(menuIDs[0], 3)

	// 创建角色并关联所有资源
	createdRole, err := suite.roleservice.CreateRole(context.Background(), CreateTestRoleDTO(apiIDs, menuIDs, buttonIDs))
	suite.Nil(err, "创建角色应该成功")

	// 验证关联关系
	foundRole, err := suite.roleservice.FindRoleByID(context.Background(), []string{"Apis", "Menus", "Buttons"}, createdRole.ID)
	suite.Nil(err, "查询角色应该成功")
	suite.Len(foundRole.Apis, 3, "角色关联的API数量应该正确")
	suite.Len(foundRole.Menus, 3, "角色关联的菜单数量应该正确")
	suite.Len(foundRole.Buttons, 3, "角色关联的按钮数量应该正确")
}

// TestRoleCasbinInheritance 测试角色对 API、菜单、按钮的权限继承
func (suite *RoleTestSuite) TestRoleCasbinInheritance() {
	// 创建测试资源
	apiIDs := suite.CreateTestApis(1)
	menuIDs := suite.CreateTestMenus(1)
	buttonIDs := suite.CreateTestButtons(menuIDs[0], 1)

	// 创建角色并关联API、菜单、按钮
	createdRole, err := suite.roleservice.CreateRole(context.Background(), CreateTestRoleDTO(apiIDs, menuIDs, buttonIDs))
	suite.Nil(err, "创建角色应该成功")

	// 加载角色策略
	loadErr := suite.roleservice.LoadRolePolicy(context.Background())
	suite.Nil(loadErr, "加载角色策略应该成功")

	// 验证策略添加成功（通过查询角色及其关联资源）
	foundRole, err := suite.roleservice.FindRoleByID(context.Background(), []string{"Apis", "Menus", "Buttons"}, createdRole.ID)
	suite.Nil(err, "查询角色应该成功")
	suite.Len(foundRole.Apis, 1, "角色应该关联API")
	suite.Len(foundRole.Menus, 1, "角色应该关联菜单")
	suite.Len(foundRole.Buttons, 1, "角色应该关联按钮")
}
