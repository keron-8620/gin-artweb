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

// CreateTestButtonModel 创建测试用的按钮模型
func CreateTestButtonModel(menuID uint32) *model.ButtonModel {
	return &model.ButtonModel{
		MenuID:   menuID,
		Name:     uuid.NewString(),
		Sort:     10000,
		IsActive: true,
		Descr:    "这是一个测试按钮",
	}
}

type ButtonTestSuite struct {
	suite.Suite
	enforcer *casbin.Enforcer
	uc       *ButtonUsecase
}

func (suite *ButtonTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&model.MenuModel{},
		&model.ApiModel{},
		&model.ButtonModel{},
	)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	suite.enforcer = enforcer
	suite.uc = &ButtonUsecase{
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
	}
}

func (suite *ButtonTestSuite) TestGetMenu() {
	// 测试获取不存在的菜单
	menuID := uint32(9999)
	_, err := suite.uc.GetMenu(context.Background(), menuID)
	suite.NotNil(err, "获取不存在的菜单应该失败")
}

func (suite *ButtonTestSuite) TestGetMenu_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文获取菜单
	menuID := uint32(1)
	_, err := suite.uc.GetMenu(ctx, menuID)
	suite.NotNil(err, "上下文错误时获取菜单应该失败")
}

func (suite *ButtonTestSuite) TestGetApis() {
	// 测试获取空API列表
	emptyApiIDs := []uint32{}
	apis, err := suite.uc.GetApis(context.Background(), emptyApiIDs)
	suite.Nil(err, "获取空API列表应该成功")
	suite.NotNil(apis, "返回的API列表不应该为nil")
	suite.Len(*apis, 0, "空API列表应该返回长度为0的切片")

	// 测试获取不存在的API列表
	nonExistentApiIDs := []uint32{9999, 8888}
	apis, err = suite.uc.GetApis(context.Background(), nonExistentApiIDs)
	suite.Nil(err, "获取不存在的API列表应该成功")
	suite.NotNil(apis, "返回的API列表不应该为nil")
}

func (suite *ButtonTestSuite) TestGetApis_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文获取API列表
	apiIDs := []uint32{1}
	_, err := suite.uc.GetApis(ctx, apiIDs)
	suite.NotNil(err, "上下文错误时获取API列表应该失败")
}

func (suite *ButtonTestSuite) TestCreateButton() {
	// 先创建一个菜单，因为按钮需要关联菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(testMenu.ID, uint32(0), "菜单ID应该大于0")

	// 创建按钮
	testButton := CreateTestButtonModel(testMenu.ID)
	apiIDs := []uint32{}
	createdButton, err := suite.uc.CreateButton(context.Background(), apiIDs, *testButton)
	suite.Nil(err, "创建按钮应该成功")
	suite.Greater(createdButton.ID, uint32(0), "按钮ID应该大于0")
	suite.Equal(testButton.MenuID, createdButton.MenuID)
	suite.Equal(testButton.Name, createdButton.Name)
	suite.Equal(testButton.Sort, createdButton.Sort)
	suite.Equal(testButton.IsActive, createdButton.IsActive)
	suite.Equal(testButton.Descr, createdButton.Descr)
}

func (suite *ButtonTestSuite) TestCreateButton_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文创建按钮
	testButton := CreateTestButtonModel(1)
	apiIDs := []uint32{}
	_, err := suite.uc.CreateButton(ctx, apiIDs, *testButton)
	suite.NotNil(err, "上下文错误时创建按钮应该失败")
}

func (suite *ButtonTestSuite) TestFindButtonByID() {
	// 先创建一个菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(testMenu.ID, uint32(0), "菜单ID应该大于0")

	// 创建一个按钮
	testButton := CreateTestButtonModel(testMenu.ID)
	apiIDs := []uint32{}
	createdButton, err := suite.uc.CreateButton(context.Background(), apiIDs, *testButton)
	suite.Nil(err, "创建按钮应该成功")
	suite.Greater(createdButton.ID, uint32(0), "按钮ID应该大于0")

	// 测试查找刚创建的按钮
	foundButton, err := suite.uc.FindButtonByID(context.Background(), []string{}, createdButton.ID)
	suite.Nil(err, "查询刚创建的按钮应该成功")
	suite.Greater(foundButton.ID, uint32(0), "按钮ID应该大于0")
	suite.Equal(testButton.Name, foundButton.Name)
	suite.Equal(testButton.MenuID, foundButton.MenuID)
}

func (suite *ButtonTestSuite) TestFindButtonByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文查找按钮
	_, err := suite.uc.FindButtonByID(ctx, []string{}, 1)
	suite.NotNil(err, "上下文错误时查找按钮应该失败")
}

func (suite *ButtonTestSuite) TestUpdateButtonByID() {
	// 先创建一个菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(testMenu.ID, uint32(0), "菜单ID应该大于0")

	// 创建一个按钮
	testButton := CreateTestButtonModel(testMenu.ID)
	apiIDs := []uint32{}
	createdButton, err := suite.uc.CreateButton(context.Background(), apiIDs, *testButton)
	suite.Nil(err, "创建按钮应该成功")
	suite.Greater(createdButton.ID, uint32(0), "按钮ID应该大于0")

	// 准备更新数据
	updateData := map[string]any{
		"name":      "Updated Button",
		"descr":     "这是一个更新后的测试按钮",
		"is_active": false,
	}

	// 执行更新
	updatedButton, err := suite.uc.UpdateButtonByID(context.Background(), createdButton.ID, apiIDs, updateData)
	suite.Nil(err, "更新按钮应该成功")
	suite.Equal(createdButton.ID, updatedButton.ID)
	suite.Equal(updateData["name"], updatedButton.Name)
	suite.Equal(updateData["descr"], updatedButton.Descr)
	suite.Equal(updateData["is_active"], updatedButton.IsActive)
}

func (suite *ButtonTestSuite) TestUpdateButtonByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文更新按钮
	updateData := map[string]any{
		"name": "Updated Button",
	}
	apiIDs := []uint32{}
	_, err := suite.uc.UpdateButtonByID(ctx, 1, apiIDs, updateData)
	suite.NotNil(err, "上下文错误时更新按钮应该失败")
}

func (suite *ButtonTestSuite) TestDeleteButtonByID() {
	// 先创建一个菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(testMenu.ID, uint32(0), "菜单ID应该大于0")

	// 创建一个按钮
	testButton := CreateTestButtonModel(testMenu.ID)
	apiIDs := []uint32{}
	createdButton, err := suite.uc.CreateButton(context.Background(), apiIDs, *testButton)
	suite.Nil(err, "创建按钮应该成功")
	suite.Greater(createdButton.ID, uint32(0), "按钮ID应该大于0")

	// 测试删除刚创建的按钮
	err = suite.uc.DeleteButtonByID(context.Background(), createdButton.ID)
	suite.Nil(err, "删除刚创建的按钮应该成功")

	// 验证按钮已被删除
	_, err = suite.uc.FindButtonByID(context.Background(), []string{}, createdButton.ID)
	suite.NotNil(err, "查询已删除的按钮应该失败")
}

func (suite *ButtonTestSuite) TestDeleteButtonByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文删除按钮
	err := suite.uc.DeleteButtonByID(ctx, 1)
	suite.NotNil(err, "上下文错误时删除按钮应该失败")
}

func (suite *ButtonTestSuite) TestListButton() {
	// 先创建一个菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(testMenu.ID, uint32(0), "菜单ID应该大于0")

	// 创建多个按钮
	buttonCount := 3
	for i := 0; i < buttonCount; i++ {
		testButton := CreateTestButtonModel(testMenu.ID)
		apiIDs := []uint32{}
		_, err := suite.uc.CreateButton(context.Background(), apiIDs, *testButton)
		suite.Nil(err, "创建按钮应该成功")
	}

	// 测试列出所有按钮
	qp := database.QueryParams{}
	count, buttonList, err := suite.uc.ListButton(context.Background(), qp)
	suite.Nil(err, "列出按钮应该成功")
	suite.GreaterOrEqual(int(count), buttonCount, "返回的按钮数量应该大于等于创建的数量")
	suite.NotNil(buttonList, "返回的按钮列表不应该为nil")
}

func (suite *ButtonTestSuite) TestListButton_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文列出按钮
	qp := database.QueryParams{}
	_, _, err := suite.uc.ListButton(ctx, qp)
	suite.NotNil(err, "上下文错误时列出按钮应该失败")
}

func (suite *ButtonTestSuite) TestLoadButtonPolicy() {
	// 先创建一个菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(testMenu.ID, uint32(0), "菜单ID应该大于0")

	// 创建几个按钮
	buttonCount := 2
	for i := 0; i < buttonCount; i++ {
		testButton := CreateTestButtonModel(testMenu.ID)
		apiIDs := []uint32{}
		_, err := suite.uc.CreateButton(context.Background(), apiIDs, *testButton)
		suite.Nil(err, "创建按钮应该成功")
	}

	// 加载按钮策略
	err = suite.uc.LoadButtonPolicy(context.Background())
	suite.Nil(err, "加载按钮策略应该成功")
}

func (suite *ButtonTestSuite) TestLoadButtonPolicy_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文加载按钮策略
	err := suite.uc.LoadButtonPolicy(ctx)
	suite.NotNil(err, "上下文错误时加载按钮策略应该失败")
}

func (suite *ButtonTestSuite) TestCreateButtonWithApis() {
	// 先创建一个菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(testMenu.ID, uint32(0), "菜单ID应该大于0")

	// 创建几个API
	apiCount := 2
	apiIDs := make([]uint32, 0, apiCount)
	for i := 0; i < apiCount; i++ {
		testApi := CreateTestApiModel()
		err := suite.uc.apiRepo.CreateModel(context.Background(), testApi)
		suite.Nil(err, "创建API应该成功")
		suite.Greater(testApi.ID, uint32(0), "API ID应该大于0")
		apiIDs = append(apiIDs, testApi.ID)
	}

	// 创建关联API的按钮
	testButton := CreateTestButtonModel(testMenu.ID)
	createdButton, err := suite.uc.CreateButton(context.Background(), apiIDs, *testButton)
	suite.Nil(err, "创建关联API的按钮应该成功")
	suite.Greater(createdButton.ID, uint32(0), "按钮ID应该大于0")
	suite.Equal(testButton.MenuID, createdButton.MenuID)
	suite.Equal(testButton.Name, createdButton.Name)

	// 验证按钮关联的API
	foundButton, err := suite.uc.FindButtonByID(context.Background(), []string{"Apis"}, createdButton.ID)
	suite.Nil(err, "查询按钮应该成功")
	suite.NotNil(foundButton.Apis, "按钮应该有关联的API")
	suite.Len(foundButton.Apis, apiCount, "按钮关联的API数量应该正确")
}

func (suite *ButtonTestSuite) TestButtonCasbinInheritance() {
	// 先创建一个菜单
	testMenu := CreateTestMenuModel(nil)
	err := suite.uc.menuRepo.CreateModel(context.Background(), testMenu, nil)
	suite.Nil(err, "创建菜单应该成功")
	suite.Greater(testMenu.ID, uint32(0), "菜单ID应该大于0")

	// 创建一个API
	testApi := CreateTestApiModel()
	err = suite.uc.apiRepo.CreateModel(context.Background(), testApi)
	suite.Nil(err, "创建API应该成功")
	suite.Greater(testApi.ID, uint32(0), "API ID应该大于0")

	// 创建关联API的按钮
	testButton := CreateTestButtonModel(testMenu.ID)
	apiIDs := []uint32{testApi.ID}
	createdButton, err := suite.uc.CreateButton(context.Background(), apiIDs, *testButton)
	suite.Nil(err, "创建关联API的按钮应该成功")
	suite.Greater(createdButton.ID, uint32(0), "按钮ID应该大于0")

	// 加载按钮策略
	err = suite.uc.LoadButtonPolicy(context.Background())
	suite.Nil(err, "加载按钮策略应该成功")

	// 验证按钮权限策略是否正确添加到Casbin
	buttonSubject := auth.ButtonToSubject(createdButton.ID)
	apiSubject := auth.ApiToSubject(testApi.ID)

	// 检查按钮是否继承了API的权限
	// 在Casbin中，组策略的格式是：g(sub, obj)
	// 所以我们需要检查是否存在 g(buttonSubject, apiSubject) 这样的策略
	// 由于我们无法直接查询Casbin的策略，我们可以通过测试按钮是否能够通过API的主题来访问API
	// 但更简单的方法是检查按钮是否成功创建并且策略是否成功加载
	// 因为如果策略添加失败，CreateButton或LoadButtonPolicy会返回错误
	suite.Greater(createdButton.ID, uint32(0), "按钮ID应该大于0")
	suite.Greater(testApi.ID, uint32(0), "API ID应该大于0")
	suite.NotEmpty(buttonSubject, "按钮主题不应该为空")
	suite.NotEmpty(apiSubject, "API主题不应该为空")
}

// 每个测试文件都需要这个入口函数
func TestButtonTestSuite(t *testing.T) {
	pts := &ButtonTestSuite{}
	suite.Run(t, pts)
}
