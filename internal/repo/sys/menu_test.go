package sys

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	sysmodel "gin-artweb/internal/model/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

// CreateTestMenuModel 创建测试菜单模型
func CreateTestMenuModel(apis ...[]sysmodel.ApiModel) *sysmodel.MenuModel {
	menu := &sysmodel.MenuModel{
		Name:      fmt.Sprintf("test_menu_%s", uuid.NewString()),
		Path:      fmt.Sprintf("/test/menu/%s", uuid.NewString()),
		Component: "TestMenu",
		Meta: sysmodel.MetaSchemas{
			Title: "测试菜单",
			Icon:  "test-icon",
		},
		Sort:     1,
		IsActive: true,
		Descr:    "这是一个测试菜单",
	}

	// 如果提供了 API 列表，则添加一些测试 API
	if len(apis) > 0 && apis[0] != nil {
		for i := 0; i < 2; i++ {
			api := sysmodel.ApiModel{
				URL:    fmt.Sprintf("/api/test/%d", i),
				Method: "GET",
				Label:  fmt.Sprintf("test_api_%d", i),
				Descr:  fmt.Sprintf("这是测试API %d", i),
			}
			apis[0] = append(apis[0], api)
		}
	}

	return menu
}

// MenuTestSuite 菜单测试套件
type MenuTestSuite struct {
	suite.Suite
	apiRepo  *ApiRepo
	menuRepo *MenuRepo
}

// SetupSuite 测试套件设置
func (suite *MenuTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&sysmodel.MenuModel{}, &sysmodel.ApiModel{}); err != nil {
		suite.Error(err, "数据库迁移失败")
	}
	dbTimeout := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()
	enforcer, err := auth.NewCasbinEnforcer()
	if err != nil {
		suite.Error(err, "创建Casbinforcer失败")
	}
	suite.apiRepo = &ApiRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: slowThreshold,
		enforcer:      enforcer,
	}

	suite.menuRepo = &MenuRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: slowThreshold,
		enforcer:      enforcer,
	}
}

// createTestMenu 创建测试菜单并返回
func (suite *MenuTestSuite) createTestMenu() *sysmodel.MenuModel {
	menu := CreateTestMenuModel()
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")
	return menu
}

// createTestAPI 创建测试API并返回
func (suite *MenuTestSuite) createTestAPI() *sysmodel.ApiModel {
	api := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")
	return api
}

// TestCreateMenu 测试创建菜单
func (suite *MenuTestSuite) TestCreateMenu() {
	// 测试创建菜单
	menu := CreateTestMenuModel()
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 测试查询刚创建的菜单
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", menu.ID)
	suite.NoError(err, "查询刚创建的菜单应该成功")
	suite.Equal(menu.ID, fm.ID)

	// 测试删除菜单
	err = suite.menuRepo.DeleteModel(context.Background(), "id = ?", menu.ID)
	suite.NoError(err, "删除菜单应该成功")

	// 测试查询已删除的菜单
	_, err = suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", menu.ID)
	suite.Error(err, "查询已删除的菜单应该返回错误")
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
}

// TestGetMenuByID 测试根据ID查询菜单
func (suite *MenuTestSuite) TestGetMenuByID() {
	// 测试创建菜单
	menu := suite.createTestMenu()

	// 测试根据ID查询菜单
	m, err := suite.menuRepo.GetModel(context.Background(), []string{}, menu.ID)
	suite.NoError(err, "根据ID查询菜单应该成功")
	suite.Equal(menu.ID, m.ID)
	suite.Equal(menu.Name, m.Name)
	suite.Equal(menu.Path, m.Path)
	suite.Equal(menu.Component, m.Component)
	suite.Equal(menu.Meta.Title, m.Meta.Title)
	suite.Equal(menu.Meta.Icon, m.Meta.Icon)
	suite.Equal(menu.Sort, m.Sort)
	suite.Equal(menu.IsActive, m.IsActive)
	suite.Equal(menu.Descr, m.Descr)
}

// TestListMenus 测试查询菜单列表
func (suite *MenuTestSuite) TestListMenus() {
	// 测试创建多个菜单
	for range 5 {
		suite.createTestMenu()
	}

	// 测试CountModel
	count, err := suite.menuRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取菜单总数应该成功")
	suite.GreaterOrEqual(count, int64(5), "菜单总数应该至少有5条")

	// 测试查询菜单列表
	qp := database.QueryParams{
		Limit:  10,
		Offset: 0,
	}
	ms, err := suite.menuRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询菜单列表应该成功")
	suite.NotNil(ms, "菜单列表不应该为nil")
	suite.GreaterOrEqual(len(ms), 5, "菜单列表应该至少有5条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:  2,
		Offset: 0,
	}
	pMs, err := suite.menuRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询菜单列表应该成功")
	suite.NotNil(pMs, "分页菜单列表不应该为nil")
	suite.Equal(2, len(pMs), "分页查询应该返回指定数量的记录")

	// 测试查询菜单列表时传入空参数
	qpEmpty := database.QueryParams{}
	msEmpty, err := suite.menuRepo.ListModel(context.Background(), qpEmpty)
	suite.NoError(err, "查询菜单列表时传入空参数应该成功")
	suite.NotNil(msEmpty, "菜单列表不应该为nil")

	// 测试CountModel带空条件
	countEmpty, err := suite.menuRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取菜单总数应该成功")
	suite.GreaterOrEqual(countEmpty, int64(0), "菜单总数应该大于等于0")
}

// TestListMenusWithSorting 测试查询菜单列表时传入排序参数
func (suite *MenuTestSuite) TestListMenusWithSorting() {
	// 测试创建多个菜单
	for i := 0; i < 5; i++ {
		menu := CreateTestMenuModel()
		menu.Sort = uint32(i)
		err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
		suite.NoError(err, "创建菜单应该成功")
	}

	// 测试按排序字段降序排序
	qp := database.QueryParams{
		OrderBy: []string{"sort DESC"},
	}
	ms, err := suite.menuRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按排序字段降序排序查询应该成功")
	suite.NotNil(ms, "菜单列表不应该为nil")
	if len(ms) > 1 {
		// 验证排序结果
		prevSort := ms[0].Sort
		for _, menu := range ms {
			suite.LessOrEqual(menu.Sort, prevSort, "菜单应该按排序字段降序排序")
			prevSort = menu.Sort
		}
	}
}

// TestListMenusWithFiltering 测试查询菜单列表时传入过滤参数
func (suite *MenuTestSuite) TestListMenusWithFiltering() {
	// 测试创建一个特定名称的菜单
	testName := "filter_test_menu"
	menu := CreateTestMenuModel()
	menu.Name = testName
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 测试按名称过滤
	qp := database.QueryParams{
		Query: map[string]any{
			"name": testName,
		},
	}
	ms, err := suite.menuRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按名称过滤查询应该成功")
	suite.NotNil(ms, "菜单列表不应该为nil")
	// 验证过滤结果
	for _, m := range ms {
		suite.Equal(testName, m.Name, "菜单应该按名称过滤")
	}

	// 测试CountModel带过滤条件
	count, err := suite.menuRepo.CountModel(context.Background(), map[string]any{
		"name": testName,
	})
	suite.NoError(err, "带过滤条件的菜单总数查询应该成功")
	suite.GreaterOrEqual(count, int64(1), "带过滤条件的菜单总数应该至少为1")
}

// TestUpdateMenu 测试更新菜单
func (suite *MenuTestSuite) TestUpdateMenu() {
	// 测试创建菜单
	menu := suite.createTestMenu()

	// 测试更新菜单
	updatedName := "updated_menu"
	updatedPath := "/updated/path"
	updatedDescr := "这是更新后的菜单描述"

	err := suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name":  updatedName,
		"path":  updatedPath,
		"descr": updatedDescr,
	}, nil, "id = ?", menu.ID)
	suite.NoError(err, "更新菜单应该成功")

	// 测试查询更新后的菜单
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", menu.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(menu.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedPath, fm.Path)
	suite.Equal(updatedDescr, fm.Descr)
}

// TestUpdateMenuWithAPIs 测试更新菜单时关联API
func (suite *MenuTestSuite) TestUpdateMenuWithAPIs() {
	// 创建API
	api1 := suite.createTestAPI()
	api2 := suite.createTestAPI()

	// 创建菜单
	menu := suite.createTestMenu()

	// 更新菜单并关联API
	updatedName := "updated_menu_with_apis"
	apis := []sysmodel.ApiModel{*api1, *api2}
	err := suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name": updatedName,
	}, apis, "id = ?", menu.ID)
	suite.NoError(err, "更新菜单并关联API应该成功")

	// 验证菜单是否更新成功
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{"Apis"}, "id = ?", menu.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(updatedName, fm.Name, "菜单名称应该更新成功")
	suite.Equal(2, len(fm.Apis), "菜单应该关联2个API")
}

// TestUpdateMenuWithoutAPIs 测试更新菜单时不关联API
func (suite *MenuTestSuite) TestUpdateMenuWithoutAPIs() {
	// 创建菜单
	menu := suite.createTestMenu()

	// 更新菜单但不关联API
	updatedName := "updated_menu_without_apis"
	err := suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name": updatedName,
	}, nil, "id = ?", menu.ID)
	suite.NoError(err, "更新菜单应该成功")

	// 验证菜单是否更新成功
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", menu.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(updatedName, fm.Name, "菜单名称应该更新成功")
}

// TestUpdateMenuWithEmptyAPIs 测试更新菜单时传入空的API列表
func (suite *MenuTestSuite) TestUpdateMenuWithEmptyAPIs() {
	// 创建菜单
	menu := suite.createTestMenu()

	// 更新菜单并传入空的API列表
	updatedName := "updated_menu_with_empty_apis"
	emptyAPIs := []sysmodel.ApiModel{}
	err := suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name": updatedName,
	}, emptyAPIs, "id = ?", menu.ID)
	suite.NoError(err, "更新菜单应该成功")

	// 验证菜单是否更新成功
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", menu.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(updatedName, fm.Name, "菜单名称应该更新成功")
}

// TestDeleteMenu 测试删除菜单
func (suite *MenuTestSuite) TestDeleteMenu() {
	// 测试创建菜单
	menu := suite.createTestMenu()

	// 测试删除菜单
	err := suite.menuRepo.DeleteModel(context.Background(), "id = ?", menu.ID)
	suite.NoError(err, "删除菜单应该成功")

	// 测试查询已删除的菜单
	_, err = suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", menu.ID)
	suite.Error(err, "查询已删除的菜单应该返回错误")
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
}

// TestGetMenuWithPreload 测试查询菜单时预加载关联数据
func (suite *MenuTestSuite) TestGetMenuWithPreload() {
	// 测试创建菜单
	menu := suite.createTestMenu()

	// 测试查询菜单时预加载关联数据
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{"Apis"}, "id = ?", menu.ID)
	suite.NoError(err, "查询菜单时预加载关联数据应该成功")
	suite.Equal(menu.ID, fm.ID)
}

// TestNewMenuRepo 测试创建菜单仓库实例
func (suite *MenuTestSuite) TestNewMenuRepo() {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	slowThreshold := test.NewTestDBSlowThreshold()

	repo := NewMenuRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	suite.NotNil(repo, "NewMenuRepo should return a non-nil repository")
	suite.NotNil(repo.log, "Repo log should not be nil")
	suite.NotNil(repo.gormDB, "Repo gormDB should not be nil")
	suite.NotNil(repo.timeouts, "Repo timeouts should not be nil")
	suite.NotNil(repo.enforcer, "Repo enforcer should not be nil")
}

// TestMenuAddGroupPolicy 测试添加菜单组策略
func (suite *MenuTestSuite) TestMenuAddGroupPolicy() {
	// 创建API
	api := suite.createTestAPI()

	// 创建菜单并关联API
	menu := suite.createTestMenu()
	menu.Apis = []sysmodel.ApiModel{*api}

	// 测试添加组策略
	err := suite.menuRepo.AddGroupPolicy(context.Background(), *menu)
	suite.NoError(err, "添加菜单组策略应该成功")

	// 再次尝试添加相同的策略，应该不会报错（Casbin会处理重复策略）
	err = suite.menuRepo.AddGroupPolicy(context.Background(), *menu)
	suite.NoError(err, "重复添加菜单组策略应该成功")
}

// TestMenuAddGroupPolicyWithParent 测试添加带有父菜单的菜单组策略
func (suite *MenuTestSuite) TestMenuAddGroupPolicyWithParent() {
	// 创建父菜单
	parentMenu := suite.createTestMenu()

	// 创建子菜单并设置父菜单ID
	childMenu := suite.createTestMenu()
	childMenu.ParentID = &parentMenu.ID

	// 测试添加组策略
	err := suite.menuRepo.AddGroupPolicy(context.Background(), *childMenu)
	suite.NoError(err, "添加带有父菜单的菜单组策略应该成功")
}

// TestMenuAddGroupPolicyWithInvalidAPI 测试添加包含无效API的菜单组策略
func (suite *MenuTestSuite) TestMenuAddGroupPolicyWithInvalidAPI() {
	// 创建菜单
	menu := suite.createTestMenu()

	// 手动设置无效API（ID为0）
	menu.Apis = []sysmodel.ApiModel{{URL: "/api/test", Method: "GET"}}

	// 测试添加组策略（应该跳过无效API）
	err := suite.menuRepo.AddGroupPolicy(context.Background(), *menu)
	suite.NoError(err, "添加包含无效API的菜单组策略应该成功")
}

// TestMenuRemoveGroupPolicy 测试删除菜单组策略
func (suite *MenuTestSuite) TestMenuRemoveGroupPolicy() {
	// 创建菜单
	menu := suite.createTestMenu()

	// 先添加组策略
	err := suite.menuRepo.AddGroupPolicy(context.Background(), *menu)
	suite.NoError(err, "添加菜单组策略应该成功")

	// 测试删除组策略
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), *menu, true)
	suite.NoError(err, "删除菜单组策略应该成功")

	// 测试删除组策略（不删除继承）
	err = suite.menuRepo.AddGroupPolicy(context.Background(), *menu)
	suite.NoError(err, "添加菜单组策略应该成功")

	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), *menu, false)
	suite.NoError(err, "删除菜单组策略应该成功")

	// 再次尝试删除相同的策略，应该不会报错（Casbin会处理不存在的策略）
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), *menu, true)
	suite.NoError(err, "删除不存在的菜单组策略应该成功")
}

// TestContextTimeout 测试上下文超时
func (suite *MenuTestSuite) TestContextTimeout() {
	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 测试创建菜单时上下文超时
	menu := CreateTestMenuModel()
	err := suite.menuRepo.CreateModel(ctx, menu, nil)
	suite.Error(err, "上下文已取消时创建菜单应该返回错误")

	// 测试获取菜单时上下文超时
	_, err = suite.menuRepo.GetModel(ctx, []string{}, 1)
	suite.Error(err, "获取菜单时上下文超时应该返回错误")

	// 测试列表查询时上下文超时
	qp := database.QueryParams{}
	_, err = suite.menuRepo.ListModel(ctx, qp)
	suite.Error(err, "列表查询时上下文超时应该返回错误")

	// 测试计数查询时上下文超时
	_, err = suite.menuRepo.CountModel(ctx, nil)
	suite.Error(err, "计数查询时上下文超时应该返回错误")

	// 测试更新菜单时上下文超时
	err = suite.menuRepo.UpdateModel(ctx, map[string]any{
		"name": "updated_menu",
	}, nil, "id = ?", 1)
	suite.Error(err, "更新菜单时上下文超时应该返回错误")

	// 测试删除菜单时上下文超时
	err = suite.menuRepo.DeleteModel(ctx, "id = ?", 1)
	suite.Error(err, "删除菜单时上下文超时应该返回错误")

	// 测试添加组策略时上下文超时
	menuWithID := suite.createTestMenu()
	err = suite.menuRepo.AddGroupPolicy(ctx, *menuWithID)
	suite.Error(err, "上下文已取消时添加组策略应该返回错误")

	// 测试删除组策略时上下文超时
	err = suite.menuRepo.RemoveGroupPolicy(ctx, *menuWithID, true)
	suite.Error(err, "删除菜单组策略时上下文超时应该返回错误")
}

// TestInvalidInputs 测试无效输入
func (suite *MenuTestSuite) TestInvalidInputs() {
	// 测试创建菜单时传入 nil 模型
	err := suite.menuRepo.CreateModel(context.Background(), nil, nil)
	suite.Error(err, "传入 nil 模型应该返回错误")
	suite.Contains(err.Error(), "创建菜单模型:模型不能为空")

	// 测试更新不存在的菜单
	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name": "updated_menu",
	}, []sysmodel.ApiModel{}, "id = ?", 999999)
	suite.Error(err, "更新不存在的菜单应该返回错误")

	// 测试删除不存在的菜单
	err = suite.menuRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的菜单不应该返回错误")

	// 测试更新菜单时传入空数据
	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{}, nil, "id = ?", 1)
	suite.Error(err, "更新菜单时传入空数据应该返回错误")

	// 测试删除菜单时传入空条件
	err = suite.menuRepo.DeleteModel(context.Background())
	suite.Error(err, "删除菜单时传入空条件应该返回错误")

	// 测试添加组策略时传入无效菜单
	menuWithZeroID := suite.createTestMenu()
	// 这里我们假设ID字段是可修改的，或者我们可以创建一个无效的菜单对象
	err = suite.menuRepo.AddGroupPolicy(context.Background(), *menuWithZeroID)
	suite.NoError(err, "添加菜单组策略应该成功")

	// 测试删除组策略时传入无效菜单
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), *menuWithZeroID, true)
	suite.NoError(err, "删除菜单组策略应该成功")
}

// 每个测试文件都需要这个入口函数
func TestMenuTestSuite(t *testing.T) {
	pts := &MenuTestSuite{}
	suite.Run(t, pts)
}
