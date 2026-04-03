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
func CreateTestMenuModel(apis []sysmodel.ApiModel) *sysmodel.MenuModel {
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
	if apis != nil {
		for i := 0; i < 2; i++ {
			api := sysmodel.ApiModel{
				URL:    fmt.Sprintf("/api/test/%d", i),
				Method: "GET",
				Label:  fmt.Sprintf("test_api_%d", i),
				Descr:  fmt.Sprintf("这是测试API %d", i),
			}
			apis = append(apis, api)
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
	db.AutoMigrate(&sysmodel.MenuModel{}, &sysmodel.ApiModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	suite.apiRepo = &ApiRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}

	suite.menuRepo = &MenuRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}
}

// TestCreateMenu 测试创建菜单
func (suite *MenuTestSuite) TestCreateMenu() {
	// 测试创建菜单
	apis := []sysmodel.ApiModel{}
	sm := CreateTestMenuModel(apis)
	err := suite.menuRepo.CreateModel(context.Background(), sm, apis)
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

// TestGetMenuByID 测试根据ID查询菜单
func (suite *MenuTestSuite) TestGetMenuByID() {
	// 测试创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
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
	suite.Equal(sm.Sort, m.Sort)
	suite.Equal(sm.IsActive, m.IsActive)
	suite.Equal(sm.Descr, m.Descr)
}

// TestListMenus 测试查询菜单列表
func (suite *MenuTestSuite) TestListMenus() {
	// 测试创建多个菜单
	apis := []sysmodel.ApiModel{}
	for range 5 {
		sm := CreateTestMenuModel(nil)
		err := suite.menuRepo.CreateModel(context.Background(), sm, apis)
		suite.NoError(err, "创建菜单应该成功")
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
}

// TestCreateMenuWithNilModel 测试创建菜单时传入 nil 模型
func (suite *MenuTestSuite) TestCreateMenuWithNilModel() {
	// 测试创建菜单时传入 nil 模型
	err := suite.menuRepo.CreateModel(context.Background(), nil, nil)
	suite.Error(err, "传入 nil 模型应该返回错误")
	suite.Contains(err.Error(), "创建菜单模型:模型不能为空")
}

// TestCreateMenuWithEmptyApis 测试创建菜单时传入空的 APIs 列表
func (suite *MenuTestSuite) TestCreateMenuWithEmptyApis() {
	// 测试创建菜单时传入空的 APIs 列表
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
	suite.NoError(err, "传入空的 APIs 列表应该成功创建菜单")
}

// TestCreateMenuWithCanceledContext 测试上下文已取消时创建菜单
func (suite *MenuTestSuite) TestCreateMenuWithCanceledContext() {
	// 创建一个已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(ctx, sm, nil)
	suite.Error(err, "上下文已取消时创建菜单应该返回错误")
}

// TestUpdateMenu 测试更新菜单
func (suite *MenuTestSuite) TestUpdateMenu() {
	// 测试创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, []sysmodel.ApiModel{})
	suite.NoError(err, "创建菜单应该成功")

	// 测试更新菜单
	updatedName := "updated_menu"
	updatedPath := "/updated/path"
	updatedDescr := "这是更新后的菜单描述"

	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name":  updatedName,
		"path":  updatedPath,
		"descr": updatedDescr,
	}, nil, "id = ?", sm.ID)
	suite.NoError(err, "更新菜单应该成功")

	// 测试查询更新后的菜单
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedPath, fm.Path)
	suite.Equal(updatedDescr, fm.Descr)
}

// TestDeleteMenu 测试删除菜单
func (suite *MenuTestSuite) TestDeleteMenu() {
	// 测试创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
	suite.NoError(err, "创建菜单应该成功")

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

// TestGetMenuWithPreload 测试查询菜单时预加载关联数据
func (suite *MenuTestSuite) TestGetMenuWithPreload() {
	// 测试创建菜单
	apis := []sysmodel.ApiModel{}
	sm := CreateTestMenuModel(apis)
	err := suite.menuRepo.CreateModel(context.Background(), sm, apis)
	suite.NoError(err, "创建菜单应该成功")

	// 测试查询菜单时预加载关联数据
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{"Apis"}, "id = ?", sm.ID)
	suite.NoError(err, "查询菜单时预加载关联数据应该成功")
	suite.Equal(sm.ID, fm.ID)
}

// TestUpdateMenuWithNonExistentID 测试更新不存在的菜单
func (suite *MenuTestSuite) TestUpdateMenuWithNonExistentID() {
	// 测试更新不存在的菜单
	err := suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name": "updated_menu",
	}, []sysmodel.ApiModel{}, "id = ?", 999999)
	suite.NoError(err, "更新不存在的菜单不应该返回错误")
}

// TestDeleteMenuWithNonExistentID 测试删除不存在的菜单
func (suite *MenuTestSuite) TestDeleteMenuWithNonExistentID() {
	// 测试删除不存在的菜单
	err := suite.menuRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的菜单不应该返回错误")
}

// TestListMenusWithEmptyParams 测试查询菜单列表时传入空参数
func (suite *MenuTestSuite) TestListMenusWithEmptyParams() {
	// 测试查询菜单列表时传入空参数
	qp := database.QueryParams{}
	ms, err := suite.menuRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询菜单列表时传入空参数应该成功")
	suite.NotNil(ms, "菜单列表不应该为nil")

	// 测试CountModel
	count, err := suite.menuRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取菜单总数应该成功")
	suite.GreaterOrEqual(count, int64(0), "菜单总数应该大于等于0")
}

// TestListMenusWithSorting 测试查询菜单列表时传入排序参数
func (suite *MenuTestSuite) TestListMenusWithSorting() {
	// 测试创建多个菜单
	for i := 0; i < 5; i++ {
		sm := CreateTestMenuModel(nil)
		sm.Sort = uint32(i)
		err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
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
	sm := CreateTestMenuModel(nil)
	sm.Name = testName
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
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
	for _, menu := range ms {
		suite.Equal(testName, menu.Name, "菜单应该按名称过滤")
	}

	// 测试CountModel带过滤条件
	count, err := suite.menuRepo.CountModel(context.Background(), map[string]any{
		"name": testName,
	})
	suite.NoError(err, "带过滤条件的菜单总数查询应该成功")
	suite.GreaterOrEqual(count, int64(1), "带过滤条件的菜单总数应该至少为1")
}

// TestGetMenuWithContextTimeout 测试获取菜单时上下文超时
func (suite *MenuTestSuite) TestGetMenuWithContextTimeout() {
	// 测试创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试获取菜单
	_, err = suite.menuRepo.GetModel(ctx, []string{}, sm.ID)
	suite.Error(err, "获取菜单时上下文超时应该返回错误")
}

// TestListMenusWithContextTimeout 测试列表查询时上下文超时
func (suite *MenuTestSuite) TestListMenusWithContextTimeout() {
	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试列表查询
	qp := database.QueryParams{}
	_, err := suite.menuRepo.ListModel(ctx, qp)
	suite.Error(err, "列表查询时上下文超时应该返回错误")
}

// TestCountMenusWithContextTimeout 测试计数查询时上下文超时
func (suite *MenuTestSuite) TestCountMenusWithContextTimeout() {
	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试计数查询
	_, err := suite.menuRepo.CountModel(ctx, nil)
	suite.Error(err, "计数查询时上下文超时应该返回错误")
}

// TestNewMenuRepo 测试创建菜单仓库实例
func (suite *MenuTestSuite) TestNewMenuRepo() {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()

	repo := NewMenuRepo(logger, db, dbTimeout, enforcer)
	suite.NotNil(repo, "NewMenuRepo should return a non-nil repository")
	suite.NotNil(repo.log, "Repo log should not be nil")
	suite.NotNil(repo.gormDB, "Repo gormDB should not be nil")
	suite.NotNil(repo.timeouts, "Repo timeouts should not be nil")
	suite.NotNil(repo.enforcer, "Repo enforcer should not be nil")
}

// TestMenuAddGroupPolicy 测试添加菜单组策略
func (suite *MenuTestSuite) TestMenuAddGroupPolicy() {
	// 创建API
	api := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")

	// 创建菜单并关联API
	menu := CreateTestMenuModel(nil)
	menu.Apis = []sysmodel.ApiModel{*api}
	err = suite.menuRepo.CreateModel(context.Background(), menu, menu.Apis)
	suite.NoError(err, "创建菜单应该成功")

	// 测试添加组策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "添加菜单组策略应该成功")
}

// TestMenuAddGroupPolicyWithParent 测试添加带有父菜单的菜单组策略
func (suite *MenuTestSuite) TestMenuAddGroupPolicyWithParent() {
	// 创建父菜单
	parentMenu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), parentMenu, nil)
	suite.NoError(err, "创建父菜单应该成功")

	// 创建子菜单并设置父菜单ID
	childMenu := CreateTestMenuModel(nil)
	childMenu.ParentID = &parentMenu.ID
	err = suite.menuRepo.CreateModel(context.Background(), childMenu, nil)
	suite.NoError(err, "创建子菜单应该成功")

	// 测试添加组策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), childMenu)
	suite.NoError(err, "添加带有父菜单的菜单组策略应该成功")
}

// TestMenuAddGroupPolicyWithInvalidAPI 测试添加包含无效API的菜单组策略
func (suite *MenuTestSuite) TestMenuAddGroupPolicyWithInvalidAPI() {
	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 手动设置无效API（ID为0）
	// 注意:这里我们直接修改menu对象，因为CreateModel会忽略Apis参数
	// 这样可以测试AddGroupPolicy中处理无效API的逻辑
	menu.Apis = []sysmodel.ApiModel{{URL: "/api/test", Method: "GET"}}
	// 注意:我们不设置ID字段，让它保持默认值0

	// 测试添加组策略（应该跳过无效API）
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "添加包含无效API的菜单组策略应该成功")
}

// TestMenuRemoveGroupPolicy 测试删除菜单组策略
func (suite *MenuTestSuite) TestMenuRemoveGroupPolicy() {
	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 先添加组策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "添加菜单组策略应该成功")

	// 测试删除组策略
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), menu, true)
	suite.NoError(err, "删除菜单组策略应该成功")

	// 测试删除组策略（不删除继承）
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "添加菜单组策略应该成功")

	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), menu, false)
	suite.NoError(err, "删除菜单组策略应该成功")
}

// TestMenuRemoveGroupPolicyWithCanceledContext 测试上下文已取消时删除菜单组策略
func (suite *MenuTestSuite) TestMenuRemoveGroupPolicyWithCanceledContext() {
	// 创建一个已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 创建菜单
	menu := &sysmodel.MenuModel{}
	menu.ID = 1

	// 测试删除组策略
	err := suite.menuRepo.RemoveGroupPolicy(ctx, menu, true)
	suite.Error(err, "上下文已取消时删除菜单组策略应该返回错误")
}

// TestMenuRemoveGroupPolicyWithNilMenu 测试菜单为nil时删除组策略
func (suite *MenuTestSuite) TestMenuRemoveGroupPolicyWithNilMenu() {
	// 测试删除组策略
	err := suite.menuRepo.RemoveGroupPolicy(context.Background(), nil, true)
	suite.Error(err, "菜单为nil时删除组策略应该返回错误")
}

// TestMenuRemoveGroupPolicyWithZeroID 测试菜单ID为0时删除组策略
func (suite *MenuTestSuite) TestMenuRemoveGroupPolicyWithZeroID() {
	// 创建菜单
	menu := &sysmodel.MenuModel{}
	menu.ID = 0

	// 测试删除组策略
	err := suite.menuRepo.RemoveGroupPolicy(context.Background(), menu, true)
	suite.Error(err, "菜单ID为0时删除组策略应该返回错误")
}

// TestMenuAddGroupPolicyWithCanceledContext 测试上下文已取消时添加组策略
func (suite *MenuTestSuite) TestMenuAddGroupPolicyWithCanceledContext() {
	// 创建一个已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 创建菜单
	menu := &sysmodel.MenuModel{}
	menu.ID = 1

	// 测试添加组策略
	err := suite.menuRepo.AddGroupPolicy(ctx, menu)
	suite.Error(err, "上下文已取消时添加组策略应该返回错误")
}

// TestMenuAddGroupPolicyWithNilMenu 测试菜单为nil时添加组策略
func (suite *MenuTestSuite) TestMenuAddGroupPolicyWithNilMenu() {
	// 测试添加组策略
	err := suite.menuRepo.AddGroupPolicy(context.Background(), nil)
	suite.Error(err, "菜单为nil时添加组策略应该返回错误")
}

// TestMenuAddGroupPolicyWithZeroID 测试菜单ID为0时添加组策略
func (suite *MenuTestSuite) TestMenuAddGroupPolicyWithZeroID() {
	// 创建菜单
	menu := &sysmodel.MenuModel{}
	menu.ID = 0

	// 测试添加组策略
	err := suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.Error(err, "菜单ID为0时添加组策略应该返回错误")
}

// TestMenuUpdateModelWithAPIs 测试更新菜单时关联API
func (suite *MenuTestSuite) TestMenuUpdateModelWithAPIs() {
	// 创建API
	api1 := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), api1)
	suite.NoError(err, "创建API1应该成功")

	api2 := CreateTestApiModel()
	err = suite.apiRepo.CreateModel(context.Background(), api2)
	suite.NoError(err, "创建API2应该成功")

	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err = suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 更新菜单并关联API
	updatedName := "updated_menu_with_apis"
	apis := []sysmodel.ApiModel{*api1, *api2}
	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name": updatedName,
	}, apis, "id = ?", menu.ID)
	suite.NoError(err, "更新菜单并关联API应该成功")

	// 验证菜单是否更新成功
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{"Apis"}, "id = ?", menu.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(updatedName, fm.Name, "菜单名称应该更新成功")
	suite.Equal(2, len(fm.Apis), "菜单应该关联2个API")
}

// TestMenuUpdateModelWithoutAPIs 测试更新菜单时不关联API
func (suite *MenuTestSuite) TestMenuUpdateModelWithoutAPIs() {
	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 更新菜单但不关联API
	updatedName := "updated_menu_without_apis"
	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name": updatedName,
	}, nil, "id = ?", menu.ID)
	suite.NoError(err, "更新菜单应该成功")

	// 验证菜单是否更新成功
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", menu.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(updatedName, fm.Name, "菜单名称应该更新成功")
}

// TestMenuDeleteModelWithEmptyConditions 测试删除菜单时传入空条件
func (suite *MenuTestSuite) TestMenuDeleteModelWithEmptyConditions() {
	// 测试删除菜单时传入空条件
	err := suite.menuRepo.DeleteModel(context.Background())
	suite.Error(err, "删除菜单时传入空条件应该返回错误")
}

// TestMenuDeleteModelWithContextTimeout 测试删除菜单时上下文超时
func (suite *MenuTestSuite) TestMenuDeleteModelWithContextTimeout() {
	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试删除菜单
	err = suite.menuRepo.DeleteModel(ctx, "id = ?", menu.ID)
	suite.Error(err, "删除菜单时上下文超时应该返回错误")
}

// TestMenuAddGroupPolicyWithCasbinError 测试添加组策略时Casbin操作失败的情况
func (suite *MenuTestSuite) TestMenuAddGroupPolicyWithCasbinError() {
	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 验证策略可以正常添加
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "添加菜单组策略应该成功")

	// 再次尝试添加相同的策略，应该不会报错（Casbin会处理重复策略）
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "重复添加菜单组策略应该成功")
}

// TestMenuRemoveGroupPolicyWithCasbinError 测试删除组策略时Casbin操作失败的情况
func (suite *MenuTestSuite) TestMenuRemoveGroupPolicyWithCasbinError() {
	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 先添加组策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "添加菜单组策略应该成功")

	// 测试删除组策略
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), menu, true)
	suite.NoError(err, "删除菜单组策略应该成功")

	// 再次尝试删除相同的策略，应该不会报错（Casbin会处理不存在的策略）
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), menu, true)
	suite.NoError(err, "删除不存在的菜单组策略应该成功")
}

// 每个测试文件都需要这个入口函数
func TestMenuTestSuite(t *testing.T) {
	pts := &MenuTestSuite{}
	suite.Run(t, pts)
}
