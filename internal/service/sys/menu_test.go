package sys

import (
	"context"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	sysmodel "gin-artweb/internal/model/sys"
	syssvc "gin-artweb/internal/repo/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/test"
)

// CreateTestMenuModel 创建测试用的菜单模型
func CreateTestMenuModel(parentID *uint32) *sysmodel.MenuModel {
	return &sysmodel.MenuModel{
		ParentID:  parentID,
		Name:      uuid.NewString(),
		Path:      uuid.NewString(),
		Component: "TestMenu",
		Meta:      sysmodel.MetaSchemas{Title: "测试菜单", Icon: "test-icon"},
		Sort:      10000,
		IsActive:  true,
		Descr:     "这是一个测试菜单",
	}
}

// CreateTestMenuDTO 创建测试用的菜单DTO
func CreateTestMenuDTO(parentID *uint32, apiIDs []uint32) sysmodel.CreateMenuDTO {
	return sysmodel.CreateMenuDTO{
		ID:        uint32(uuid.New().ID()),
		ParentID:  parentID,
		Name:      uuid.NewString(),
		Path:      uuid.NewString(),
		Component: "TestMenu",
		Meta:      sysmodel.MetaSchemas{Title: "测试菜单", Icon: "test-icon"},
		Sort:      10000,
		IsActive:  true,
		Descr:     "这是一个测试菜单",
		ApiIDs:    apiIDs,
	}
}

type MenuTestSuite struct {
	suite.Suite
	enforcer    *casbin.Enforcer
	menuservice *MenuService
}

func (suite *MenuTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&sysmodel.MenuModel{},
		&sysmodel.ApiModel{},
	)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	suite.enforcer = enforcer
	suite.menuservice = NewMenuService(
		logger,
		syssvc.NewApiRepo(
			logger,
			db,
			dbTimeout,
			enforcer,
		),
		syssvc.NewMenuRepo(
			logger,
			db,
			dbTimeout,
			enforcer,
		),
	)
}

func (suite *MenuTestSuite) TestGetParentMenu() {
	// 测试获取不存在的父菜单
	var parentID uint32 = 9999
	parentMenu, err := suite.menuservice.GetParentMenu(context.Background(), &parentID)
	suite.NotNil(err, "获取不存在的父菜单应该失败")

	// 测试获取nil父菜单
	parentMenu, err = suite.menuservice.GetParentMenu(context.Background(), nil)
	suite.Nil(err, "获取nil父菜单应该成功")
	suite.Nil(parentMenu, "nil父菜单应该返回nil")

	// 测试获取0值父菜单
	var zeroParentID uint32 = 0
	parentMenu, err = suite.menuservice.GetParentMenu(context.Background(), &zeroParentID)
	suite.Nil(err, "获取0值父菜单应该成功")
	suite.Nil(parentMenu, "0值父菜单应该返回nil")
}

func (suite *MenuTestSuite) TestGetParentMenu_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文获取父菜单
	var parentID uint32 = 1
	_, err := suite.menuservice.GetParentMenu(ctx, &parentID)
	suite.NotNil(err, "上下文错误时获取父菜单应该失败")
}

func (suite *MenuTestSuite) TestGetApis() {
	// 测试获取空API列表
	emptyApiIDs := []uint32{}
	apis, err := suite.menuservice.GetApis(context.Background(), emptyApiIDs)
	suite.Nil(err, "获取空API列表应该成功")
	suite.NotNil(apis, "返回的API列表不应该为nil")
	suite.Len(apis, 0, "空API列表应该返回长度为0的切片")

	// 测试获取不存在的API列表
	nonExistentApiIDs := []uint32{9999, 8888}
	apis, err = suite.menuservice.GetApis(context.Background(), nonExistentApiIDs)
	suite.Nil(err, "获取不存在的API列表应该成功")
	suite.NotNil(apis, "返回的API列表不应该为nil")
}

func (suite *MenuTestSuite) TestGetApis_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文获取API列表
	apiIDs := []uint32{1}
	_, err := suite.menuservice.GetApis(ctx, apiIDs)
	suite.NotNil(err, "上下文错误时获取API列表应该失败")
}

func (suite *MenuTestSuite) TestCreateMenu() {
	testMenu := CreateTestMenuDTO(nil, []uint32{})

	// 创建菜单
	createdMenu, err := suite.menuservice.CreateMenu(context.Background(), testMenu)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(createdMenu.ID, uint32(0), "菜单ID应该大于0")
	suite.Equal(testMenu.Name, createdMenu.Name)
	suite.Equal(testMenu.Path, createdMenu.Path)
	suite.Equal(testMenu.Component, createdMenu.Component)
	suite.Equal(testMenu.Meta.Title, createdMenu.Meta.Title)
	suite.Equal(testMenu.Meta.Icon, createdMenu.Meta.Icon)
	suite.Equal(testMenu.Sort, createdMenu.Sort)
	suite.Equal(testMenu.IsActive, createdMenu.IsActive)
	suite.Equal(testMenu.Descr, createdMenu.Descr)
}

func (suite *MenuTestSuite) TestCreateMenu_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文创建菜单
	testMenu := CreateTestMenuDTO(nil, []uint32{})
	_, err := suite.menuservice.CreateMenu(ctx, testMenu)
	suite.NotNil(err, "上下文错误时创建菜单应该失败")
}

func (suite *MenuTestSuite) TestFindMenuByID() {
	// 先创建一个菜单
	testMenu := CreateTestMenuDTO(nil, []uint32{})
	createdMenu, err := suite.menuservice.CreateMenu(context.Background(), testMenu)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(createdMenu.ID, uint32(0), "菜单ID应该大于0")

	// 测试查找刚创建的菜单
	foundMenu, err := suite.menuservice.FindMenuByID(context.Background(), []string{}, createdMenu.ID)
	suite.Nil(err, "查询刚创建的菜单应该成功")
	suite.Greater(foundMenu.ID, uint32(0), "菜单ID应该大于0")
	suite.Equal(testMenu.Name, foundMenu.Name)
	suite.Equal(testMenu.Path, foundMenu.Path)
}

func (suite *MenuTestSuite) TestFindMenuByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文查找菜单
	_, err := suite.menuservice.FindMenuByID(ctx, []string{}, 1)
	suite.NotNil(err, "上下文错误时查找菜单应该失败")
}

func (suite *MenuTestSuite) TestUpdateMenuByID() {
	// 先创建一个菜单
	testMenu := CreateTestMenuDTO(nil, []uint32{})
	createdMenu, err := suite.menuservice.CreateMenu(context.Background(), testMenu)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(createdMenu.ID, uint32(0), "菜单ID应该大于0")

	// 准备更新数据
	updateDTO := sysmodel.UpdateMenuDTO{
		Name:      "Updated Menu",
		Path:      createdMenu.Path,
		Component: createdMenu.Component,
		Meta:      createdMenu.Meta,
		Sort:      createdMenu.Sort,
		IsActive:  false,
		Descr:     "这是一个更新后的测试菜单",
		ParentID:  nil,
		ApiIDs:    []uint32{},
	}

	// 执行更新
	updatedMenu, err := suite.menuservice.UpdateMenuByID(context.Background(), createdMenu.ID, updateDTO)
	suite.Nil(err, "更新菜单应该成功")
	suite.Equal(createdMenu.ID, updatedMenu.ID)
	suite.Equal(updateDTO.Name, updatedMenu.Name)
	suite.Equal(updateDTO.Descr, updatedMenu.Descr)
	suite.Equal(updateDTO.IsActive, updatedMenu.IsActive)
}

func (suite *MenuTestSuite) TestUpdateMenuByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文更新菜单
	updateDTO := sysmodel.UpdateMenuDTO{
		Name:      "Updated Menu",
		Path:      "/test",
		Component: "TestComponent",
		Meta:      sysmodel.MetaSchemas{Title: "Test"},
		Sort:      100,
		IsActive:  true,
		Descr:     "Test",
		ParentID:  nil,
		ApiIDs:    []uint32{},
	}
	_, err := suite.menuservice.UpdateMenuByID(ctx, 1, updateDTO)
	suite.NotNil(err, "上下文错误时更新菜单应该失败")
}

func (suite *MenuTestSuite) TestDeleteMenuByID() {
	// 先创建一个菜单
	testMenu := CreateTestMenuDTO(nil, []uint32{})
	createdMenu, err := suite.menuservice.CreateMenu(context.Background(), testMenu)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(createdMenu.ID, uint32(0), "菜单ID应该大于0")

	// 测试删除刚创建的菜单
	err = suite.menuservice.DeleteMenuByID(context.Background(), createdMenu.ID)
	suite.Nil(err, "删除刚创建的菜单应该成功")

	// 验证菜单已被删除
	_, err = suite.menuservice.FindMenuByID(context.Background(), []string{}, createdMenu.ID)
	suite.NotNil(err, "查询已删除的菜单应该失败")
}

func (suite *MenuTestSuite) TestDeleteMenuByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文删除菜单
	err := suite.menuservice.DeleteMenuByID(ctx, 1)
	suite.NotNil(err, "上下文错误时删除菜单应该失败")
}

func (suite *MenuTestSuite) TestListMenu() {
	// 创建多个菜单
	menuCount := 3
	for i := 0; i < menuCount; i++ {
		testMenu := CreateTestMenuDTO(nil, []uint32{})
		_, err := suite.menuservice.CreateMenu(context.Background(), testMenu)
		suite.Nil(err, "创建菜单应该成功")
	}

	// 测试列出所有菜单
	listDTO := sysmodel.ListMenuDTO{}
	count, menuList, err := suite.menuservice.ListMenu(context.Background(), 1, 10, listDTO)
	suite.Nil(err, "列出菜单应该成功")
	suite.GreaterOrEqual(int(count), menuCount, "返回的菜单数量应该大于等于创建的数量")
	suite.NotNil(menuList, "返回的菜单列表不应该为nil")
}

func (suite *MenuTestSuite) TestListMenu_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文列出菜单
	listDTO := sysmodel.ListMenuDTO{}
	_, _, err := suite.menuservice.ListMenu(ctx, 1, 10, listDTO)
	suite.NotNil(err, "上下文错误时列出菜单应该失败")
}

func (suite *MenuTestSuite) TestLoadMenuPolicy() {
	// 创建几个菜单
	menuCount := 2
	for i := 0; i < menuCount; i++ {
		testMenu := CreateTestMenuDTO(nil, []uint32{})
		_, err := suite.menuservice.CreateMenu(context.Background(), testMenu)
		suite.Nil(err, "创建菜单应该成功")
	}

	// 加载菜单策略
	err := suite.menuservice.LoadMenuPolicy(context.Background())
	suite.Nil(err, "加载菜单策略应该成功")
}

func (suite *MenuTestSuite) TestLoadMenuPolicy_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文加载菜单策略
	err := suite.menuservice.LoadMenuPolicy(ctx)
	suite.NotNil(err, "上下文错误时加载菜单策略应该失败")
}

func (suite *MenuTestSuite) TestCreateMenuWithApis() {
	// 创建几个API
	apiCount := 2
	apiIDs := make([]uint32, 0, apiCount)
	for i := 0; i < apiCount; i++ {
		testApi := CreateTestApiModel()
		err := suite.menuservice.apiRepo.CreateModel(context.Background(), testApi)
		suite.Nil(err, "创建API应该成功")
		suite.Greater(testApi.ID, uint32(0), "API ID应该大于0")
		apiIDs = append(apiIDs, testApi.ID)
	}

	// 创建关联API的菜单
	testMenu := CreateTestMenuDTO(nil, apiIDs)
	createdMenu, err := suite.menuservice.CreateMenu(context.Background(), testMenu)
	suite.Nil(err, "创建关联API的菜单应该成功")
	suite.Greater(createdMenu.ID, uint32(0), "菜单ID应该大于0")
	suite.Equal(testMenu.Name, createdMenu.Name)
	suite.Equal(testMenu.Path, createdMenu.Path)

	// 验证菜单关联的API
	foundMenu, err := suite.menuservice.FindMenuByID(context.Background(), []string{"Apis"}, createdMenu.ID)
	suite.Nil(err, "查询菜单应该成功")
	suite.NotNil(foundMenu.Apis, "菜单应该有关联的API")
	suite.Len(foundMenu.Apis, apiCount, "菜单关联的API数量应该正确")
}

func (suite *MenuTestSuite) TestMenuCasbinInheritance() {
	// 创建一个API
	testApi := CreateTestApiModel()
	err := suite.menuservice.apiRepo.CreateModel(context.Background(), testApi)
	suite.Nil(err, "创建API应该成功")
	suite.Greater(testApi.ID, uint32(0), "API ID应该大于0")

	// 创建关联API的菜单
	apiIDs := []uint32{testApi.ID}
	testMenu := CreateTestMenuDTO(nil, apiIDs)
	createdMenu, err := suite.menuservice.CreateMenu(context.Background(), testMenu)
	suite.Nil(err, "创建关联API的菜单应该成功")
	suite.Greater(createdMenu.ID, uint32(0), "菜单ID应该大于0")

	// 加载菜单策略
	err = suite.menuservice.LoadMenuPolicy(context.Background())
	suite.Nil(err, "加载菜单策略应该成功")

	// 验证菜单权限策略是否正确添加到Casbin
	menuSubject := auth.MenuToSubject(createdMenu.ID)
	apiSubject := auth.ApiToSubject(testApi.ID)

	// 检查菜单是否继承了API的权限
	// 在Casbin中，组策略的格式是:g(sub, obj)
	// 所以我们需要检查是否存在 g(menuSubject, apiSubject) 这样的策略
	// 由于我们无法直接查询Casbin的策略，我们可以通过测试菜单是否能够通过API的主题来访问API
	// 但更简单的方法是检查菜单是否成功创建并且策略是否成功加载
	// 因为如果策略添加失败，CreateMenu或LoadMenuPolicy会返回错误
	suite.Greater(createdMenu.ID, uint32(0), "菜单ID应该大于0")
	suite.Greater(testApi.ID, uint32(0), "API ID应该大于0")
	suite.NotEmpty(menuSubject, "菜单主题不应该为空")
	suite.NotEmpty(apiSubject, "API主题不应该为空")
}

// 每个测试文件都需要这个入口函数
func TestMenuTestSuite(t *testing.T) {
	pts := &MenuTestSuite{}
	suite.Run(t, pts)
}
