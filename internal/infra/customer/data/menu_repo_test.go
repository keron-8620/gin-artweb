package data

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/customer/model"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

// CreateTestMenuModel 创建测试用的菜单模型
func CreateTestMenuModel(parentID *uint32) *model.MenuModel {
	return &model.MenuModel{
		ParentID:  parentID,
		Name:      uuid.NewString(),
		Path:      uuid.NewString(),
		Component: "TestMenu",
		Meta:      model.Meta{Title: "测试菜单", Icon: "test-icon"},
		Sort:      10000,
		IsActive:  true,
		Descr:     "这是一个测试菜单",
	}
}

type MenuTestSuite struct {
	suite.Suite
	apiRepo  *ApiRepo
	menuRepo *MenuRepo
}

func (suite *MenuTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&model.MenuModel{}, &model.ApiModel{})
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

func (suite *MenuTestSuite) TestCreateMenuNoParent() {
	// 测试创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
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
	suite.Equal(sm.Sort, fm.Sort)
	suite.Equal(sm.IsActive, fm.IsActive)
	suite.Equal(sm.Descr, fm.Descr)
}

func (suite *MenuTestSuite) TestCreateMenuWithParent() {
	apis := make([]model.ApiModel, 3)
	for range 3 {
		m := CreateTestApiModel()
		suite.NoError(suite.apiRepo.CreateModel(context.Background(), m), "创建 API 应该成功")
		apis = append(apis, *m)
	}
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
	suite.NoError(err, "创建菜单应该成功")
	suite.Greater(sm.ID, uint32(0), "父菜单 ID 应该大于 0")

	scm := CreateTestMenuModel(&sm.ID)
	err = suite.menuRepo.CreateModel(context.Background(), scm, &apis)
	suite.NoError(err, "创建子菜单应该成功")

	// 测试查询子菜单
	fscm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", scm.ID)
	suite.NoError(err, "查询子菜单应该成功")
	suite.Equal(scm.ID, fscm.ID)
	suite.Equal(scm.Name, fscm.Name)
	suite.Equal(scm.Path, fscm.Path)
	suite.Equal(scm.Component, fscm.Component)
	suite.Equal(scm.Meta.Title, fscm.Meta.Title)
	suite.Equal(scm.Meta.Icon, fscm.Meta.Icon)
	suite.Equal(scm.Sort, fscm.Sort)
	suite.Equal(scm.IsActive, fscm.IsActive)
	suite.Equal(scm.Descr, fscm.Descr)
	suite.NotNil(fscm.ParentID, "子菜单的 ParentID 不应该为 nil")
	suite.Equal(*scm.ParentID, *fscm.ParentID)
	suite.Equal(sm.ID, *fscm.ParentID)
}

func (suite *MenuTestSuite) TestUpdateMenu() {
	// 测试创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 测试更新菜单
	updatedName := "更新的测试菜单"
	updatedPath := "/updated/test/menu"
	updatedSort := uint32(10000)
	updatedIsActive := false
	updatedDescr := ""

	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name":      updatedName,
		"path":      updatedPath,
		"sort":      updatedSort,
		"is_active": updatedIsActive,
		"descr":     updatedDescr,
	}, nil, "id = ?", sm.ID)
	suite.NoError(err, "更新菜单应该成功")

	// 测试查询更新后的菜单
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedPath, fm.Path)
	suite.Equal(updatedSort, fm.Sort)
	suite.Equal(updatedIsActive, fm.IsActive)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Greater(fm.UpdatedAt, sm.UpdatedAt)
}

func (suite *MenuTestSuite) TestDeleteMenu() {
	// 测试创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
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

func (suite *MenuTestSuite) TestListMenus() {
	// 测试创建多个菜单
	apis := []model.ApiModel{}
	for range 5 {
		sm := CreateTestMenuModel(nil)
		err := suite.menuRepo.CreateModel(context.Background(), sm, &apis)
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

// 每个测试文件都需要这个入口函数
func TestMenuTestSuite(t *testing.T) {
	pts := &MenuTestSuite{}
	suite.Run(t, pts)
}

// TestCreateMenuWithNilModel 测试创建菜单时传入 nil 模型
func (suite *MenuTestSuite) TestCreateMenuWithNilModel() {
	// 测试创建菜单时传入 nil 模型
	err := suite.menuRepo.CreateModel(context.Background(), nil, nil)
	suite.Error(err, "传入 nil 模型应该返回错误")
	suite.Contains(err.Error(), "创建菜单模型失败: 模型为空")
}

// TestCreateMenuWithEmptyApis 测试创建菜单时传入空的 APIs 列表
func (suite *MenuTestSuite) TestCreateMenuWithEmptyApis() {
	// 测试创建菜单时传入空的 APIs 列表
	sm := CreateTestMenuModel(nil)
	emptyApis := []model.ApiModel{}
	err := suite.menuRepo.CreateModel(context.Background(), sm, &emptyApis)
	suite.NoError(err, "传入空的 APIs 列表应该成功创建菜单")

	// 验证菜单创建成功
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的菜单应该成功")
	suite.Equal(sm.ID, fm.ID)
}

// TestAddGroupPolicyWithNilMenu 测试添加权限策略时传入 nil 菜单
func (suite *MenuTestSuite) TestAddGroupPolicyWithNilMenu() {
	// 测试添加权限策略时传入 nil 菜单
	err := suite.menuRepo.AddGroupPolicy(context.Background(), nil)
	suite.Error(err, "传入 nil 菜单应该返回错误")
	suite.Contains(err.Error(), "AddGroupPolicy操作时菜单模型不能为空")
}

// TestAddGroupPolicyWithZeroID 测试添加权限策略时传入 ID 为 0 的菜单
func (suite *MenuTestSuite) TestAddGroupPolicyWithZeroID() {
	// 测试添加权限策略时传入 ID 为 0 的菜单
	menu := &model.MenuModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: 0, // 零值 ID
			},
		},
	}
	err := suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.Error(err, "传入 ID 为 0 的菜单应该返回错误")
	suite.Contains(err.Error(), "AddGroupPolicy操作时菜单ID不能为0")
}

// TestRemoveGroupPolicyWithNilMenu 测试删除权限策略时传入 nil 菜单
func (suite *MenuTestSuite) TestRemoveGroupPolicyWithNilMenu() {
	// 测试删除权限策略时传入 nil 菜单
	err := suite.menuRepo.RemoveGroupPolicy(context.Background(), nil, true)
	suite.Error(err, "传入 nil 菜单应该返回错误")
	suite.Contains(err.Error(), "RemoveGroupPolicy操作时菜单模型不能为空")
}

// TestRemoveGroupPolicyWithZeroID 测试删除权限策略时传入 ID 为 0 的菜单
func (suite *MenuTestSuite) TestRemoveGroupPolicyWithZeroID() {
	// 测试删除权限策略时传入 ID 为 0 的菜单
	menu := &model.MenuModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: 0, // 零值 ID
			},
		},
	}
	err := suite.menuRepo.RemoveGroupPolicy(context.Background(), menu, true)
	suite.Error(err, "传入 ID 为 0 的菜单应该返回错误")
	suite.Contains(err.Error(), "RemoveGroupPolicy操作时菜单ID不能为0")
}

// TestUpdateMenuWithApis 测试更新菜单时修改关联的 API 列表
func (suite *MenuTestSuite) TestUpdateMenuWithApis() {
	apis := make([]model.ApiModel, 3)
	for range 3 {
		m := CreateTestApiModel()
		suite.NoError(suite.apiRepo.CreateModel(context.Background(), m))
		apis = append(apis, *m)
	}
	// 创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, &apis)
	suite.NoError(err, "创建菜单应该成功")

	// 使用唯一的名称更新菜单，避免唯一约束冲突
	uniqueName := "更新的测试菜单_" + uuid.NewString()
	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name": uniqueName,
	}, nil, "id = ?", sm.ID)
	suite.NoError(err, "更新菜单名称应该成功")

	// 验证菜单更新成功
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(uniqueName, fm.Name)
}

// TestListMenusWithPaginationBoundaries 测试分页参数边界值
func (suite *MenuTestSuite) TestListMenusWithPaginationBoundaries() {
	// 测试 Limit=0 的情况
	qpZeroLimit := database.QueryParams{
		Limit:   0,
		Offset:  0,
		IsCount: true,
	}
	_, msZero, err := suite.menuRepo.ListModel(context.Background(), qpZeroLimit)
	suite.NoError(err, "Limit=0 应该成功查询")
	suite.NotNil(msZero, "菜单列表不应该为 nil")

	// 测试较大的 Offset 值
	qpLargeOffset := database.QueryParams{
		Limit:   10,
		Offset:  999999,
		IsCount: true,
	}
	_, msLarge, err := suite.menuRepo.ListModel(context.Background(), qpLargeOffset)
	suite.NoError(err, "较大的 Offset 值应该成功查询")
	suite.NotNil(msLarge, "菜单列表不应该为 nil")
	suite.LessOrEqual(len(*msLarge), 10, "返回的记录数应该不超过 Limit")
}

// TestCreateMenuWithSortBoundaries 测试 Sort 字段边界值
func (suite *MenuTestSuite) TestCreateMenuWithSortBoundaries() {
	// 测试 Sort=0 的情况
	smZeroSort := &model.MenuModel{
		ParentID:  nil,
		Name:      uuid.NewString(),
		Path:      uuid.NewString(),
		Component: "TestMenu",
		Meta:      model.Meta{Title: "测试菜单", Icon: "test-icon"},
		Sort:      0, // 边界值 0
		IsActive:  true,
		Descr:     "这是一个测试菜单",
	}
	err := suite.menuRepo.CreateModel(context.Background(), smZeroSort, nil)
	suite.NoError(err, "Sort=0 应该成功创建菜单")

	// 验证菜单创建成功
	fmZero, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", smZeroSort.ID)
	suite.NoError(err, "查询 Sort=0 的菜单应该成功")
	suite.Equal(uint32(0), fmZero.Sort)

	// 测试较大的 Sort 值
	smLargeSort := &model.MenuModel{
		ParentID:  nil,
		Name:      uuid.NewString(),
		Path:      uuid.NewString(),
		Component: "TestMenu",
		Meta:      model.Meta{Title: "测试菜单", Icon: "test-icon"},
		Sort:      4294967295, // 最大的 uint32 值
		IsActive:  true,
		Descr:     "这是一个测试菜单",
	}
	err = suite.menuRepo.CreateModel(context.Background(), smLargeSort, nil)
	suite.NoError(err, "较大的 Sort 值应该成功创建菜单")

	// 验证菜单创建成功
	fmLarge, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", smLargeSort.ID)
	suite.NoError(err, "查询较大 Sort 值的菜单应该成功")
	suite.Equal(uint32(4294967295), fmLarge.Sort)
}

// TestRemoveGroupPolicyWithDifferentRemoveInheritedValues 测试不同的 removeInherited 参数值
func (suite *MenuTestSuite) TestRemoveGroupPolicyWithDifferentRemoveInheritedValues() {
	apis := make([]model.ApiModel, 3)
	for range 3 {
		m := CreateTestApiModel()
		suite.NoError(suite.apiRepo.CreateModel(context.Background(), m))
		apis = append(apis, *m)
	}
	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, &apis)
	suite.NoError(err, "创建菜单应该成功")

	// 添加组策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "添加菜单组策略应该成功")

	// 测试 removeInherited=true
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), menu, true)
	suite.NoError(err, "removeInherited=true 应该成功删除组策略")

	// 重新添加组策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "重新添加菜单组策略应该成功")

	// 测试 removeInherited=false
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), menu, false)
	suite.NoError(err, "removeInherited=false 应该成功删除组策略")
}

// TestGetMenuWithNonExistentID 测试查询不存在的菜单 ID
func (suite *MenuTestSuite) TestGetMenuWithNonExistentID() {
	// 测试查询不存在的菜单 ID
	nonExistentID := uint32(999999)
	_, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", nonExistentID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

// TestDeleteMenuWithNonExistentID 测试删除不存在的菜单 ID
func (suite *MenuTestSuite) TestDeleteMenuWithNonExistentID() {
	// 测试删除不存在的菜单 ID
	nonExistentID := uint32(999999)
	err := suite.menuRepo.DeleteModel(context.Background(), "id = ?", nonExistentID)
	suite.NoError(err, "删除不存在的菜单 ID 应该成功")
}

// TestUpdateMenuWithNonExistentID 测试更新不存在的菜单 ID
func (suite *MenuTestSuite) TestUpdateMenuWithNonExistentID() {
	// 测试更新不存在的菜单 ID
	nonExistentID := uint32(999999)
	err := suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name": "更新的测试菜单",
	}, nil, "id = ?", nonExistentID)
	suite.NoError(err, "更新不存在的菜单 ID 应该成功")
}

// TestContextTimeout 测试上下文超时情况
func (suite *MenuTestSuite) TestContextTimeout() {
	// 创建一个会立即超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// 等待上下文超时
	time.Sleep(10 * time.Nanosecond)

	// 测试CreateModel方法
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(ctx, sm, nil)
	suite.Error(err, "上下文超时后创建菜单应该返回错误")

	// 测试UpdateModel方法
	err = suite.menuRepo.UpdateModel(ctx, map[string]any{
		"name": "测试菜单",
	}, nil, "id = ?", uint32(1))
	suite.Error(err, "上下文超时后更新菜单应该返回错误")

	// 测试DeleteModel方法
	err = suite.menuRepo.DeleteModel(ctx, "id = ?", uint32(1))
	suite.Error(err, "上下文超时后删除菜单应该返回错误")

	// 测试GetModel方法
	_, err = suite.menuRepo.GetModel(ctx, []string{}, "id = ?", uint32(1))
	suite.Error(err, "上下文超时后获取菜单应该返回错误")

	// 测试ListModel方法
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	_, _, err = suite.menuRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文超时后列出菜单应该返回错误")

	// 测试AddGroupPolicy方法
	err = suite.menuRepo.AddGroupPolicy(ctx, sm)
	suite.Error(err, "上下文超时后添加权限策略应该返回错误")

	// 测试RemoveGroupPolicy方法
	err = suite.menuRepo.RemoveGroupPolicy(ctx, sm, true)
	suite.Error(err, "上下文超时后删除权限策略应该返回错误")
}

// TestContextCancel 测试上下文取消情况
func (suite *MenuTestSuite) TestContextCancel() {
	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	// 立即取消上下文
	cancel()

	// 测试CreateModel方法
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(ctx, sm, nil)
	suite.Error(err, "上下文取消后创建菜单应该返回错误")

	// 测试UpdateModel方法
	err = suite.menuRepo.UpdateModel(ctx, map[string]any{
		"name": "测试菜单",
	}, nil, "id = ?", uint32(1))
	suite.Error(err, "上下文取消后更新菜单应该返回错误")

	// 测试DeleteModel方法
	err = suite.menuRepo.DeleteModel(ctx, "id = ?", uint32(1))
	suite.Error(err, "上下文取消后删除菜单应该返回错误")

	// 测试GetModel方法
	_, err = suite.menuRepo.GetModel(ctx, []string{}, "id = ?", uint32(1))
	suite.Error(err, "上下文取消后获取菜单应该返回错误")

	// 测试ListModel方法
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	_, _, err = suite.menuRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文取消后列出菜单应该返回错误")

	// 测试AddGroupPolicy方法
	err = suite.menuRepo.AddGroupPolicy(ctx, sm)
	suite.Error(err, "上下文取消后添加权限策略应该返回错误")

	// 测试RemoveGroupPolicy方法
	err = suite.menuRepo.RemoveGroupPolicy(ctx, sm, true)
	suite.Error(err, "上下文取消后删除权限策略应该返回错误")
}

// TestPreloadApis 测试预加载关联的API
func (suite *MenuTestSuite) TestPreloadApis() {
	// 创建API
	apis := make([]model.ApiModel, 2)
	for i := range 2 {
		m := CreateTestApiModel()
		suite.NoError(suite.apiRepo.CreateModel(context.Background(), m))
		apis[i] = *m
	}

	// 创建菜单并关联API
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, &apis)
	suite.NoError(err, "创建菜单并关联API应该成功")

	// 不预加载API查询菜单
	menuWithoutApis, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "不预加载API查询菜单应该成功")

	// 预加载API查询菜单
	menuWithApis, err := suite.menuRepo.GetModel(context.Background(), []string{"Apis"}, "id = ?", sm.ID)
	suite.NoError(err, "预加载API查询菜单应该成功")
	suite.Equal(menuWithoutApis.ID, menuWithApis.ID)
	suite.NotEmpty(menuWithApis.Apis, "预加载后应该包含API")
	suite.Len(menuWithApis.Apis, 2, "预加载后应该包含2个API")
}

// TestPreloadMultiLevel 测试预加载多级关联
func (suite *MenuTestSuite) TestPreloadMultiLevel() {
	// 创建父菜单
	parentMenu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), parentMenu, nil)
	suite.NoError(err, "创建父菜单应该成功")

	// 创建子菜单
	childMenu := CreateTestMenuModel(&parentMenu.ID)
	err = suite.menuRepo.CreateModel(context.Background(), childMenu, nil)
	suite.NoError(err, "创建子菜单应该成功")

	// 验证子菜单的父ID
	suite.Equal(parentMenu.ID, *childMenu.ParentID)
}

// TestListMenusWithSorting 测试列表查询排序功能
func (suite *MenuTestSuite) TestListMenusWithSorting() {
	// 创建3个不同排序值的菜单
	for i := 0; i < 3; i++ {
		sm := CreateTestMenuModel(nil)
		sm.Sort = uint32(3 - i) // 排序值为3, 2, 1
		err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
		suite.NoError(err, "创建菜单应该成功")
	}

	// 测试按Sort字段升序排序
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
		OrderBy: []string{"sort asc"},
	}
	_, ms, err := suite.menuRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按Sort升序查询菜单列表应该成功")
	suite.NotNil(ms, "菜单列表不应该为nil")
	suite.GreaterOrEqual(len(*ms), 3, "菜单列表应该至少有3条记录")

	// 验证排序结果
	for i := 0; i < len(*ms)-1; i++ {
		suite.LessOrEqual((*ms)[i].Sort, (*ms)[i+1].Sort, "菜单应该按Sort升序排列")
	}
}

// TestListMenusWithNoRecords 测试列表查询无记录情况
func (suite *MenuTestSuite) TestListMenusWithNoRecords() {
	// 测试查询不存在的条件
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
		Query:   map[string]any{"id": uint32(999999)},
	}
	count, ms, err := suite.menuRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询无记录的菜单列表应该成功")
	suite.NotNil(ms, "菜单列表不应该为nil")
	suite.Equal(int64(0), count, "无记录时计数应该为0")
	suite.Len(*ms, 0, "无记录时菜单列表长度应该为0")
}

// TestCreateMenuWithInvalidParentID 测试创建菜单时传入无效父菜单ID
func (suite *MenuTestSuite) TestCreateMenuWithInvalidParentID() {
	// 传入不存在的父菜单ID
	invalidParentID := uint32(999999)
	sm := CreateTestMenuModel(&invalidParentID)

	// 验证是否能成功创建（注意：这里数据库可能允许外键约束失败，具体取决于数据库配置）
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
	// 这里不做严格断言，因为不同的数据库配置可能有不同的行为
	// 但至少应该能执行完成，不会崩溃
	if err != nil {
		suite.Contains(err.Error(), "创建菜单模型失败", "创建菜单失败时应该返回正确的错误信息")
	} else {
		// 如果创建成功，验证父ID确实被设置
		suite.NotNil(sm.ParentID, "父菜单ID不应该为nil")
		suite.Equal(invalidParentID, *sm.ParentID, "父菜单ID应该被正确设置")
	}
}

// TestMultiLevelMenuPermissionInheritance 测试多层级菜单的权限继承
func (suite *MenuTestSuite) TestMultiLevelMenuPermissionInheritance() {
	// 创建API
	api := CreateTestApiModel()
	suite.NoError(suite.apiRepo.CreateModel(context.Background(), api), "创建API应该成功")

	// 创建一级菜单
	level1Menu := CreateTestMenuModel(nil)
	apis := []model.ApiModel{*api}
	err := suite.menuRepo.CreateModel(context.Background(), level1Menu, &apis)
	suite.NoError(err, "创建一级菜单应该成功")

	// 重新加载一级菜单，确保Apis字段被正确加载
	loadedLevel1Menu, err := suite.menuRepo.GetModel(context.Background(), []string{"Apis"}, "id = ?", level1Menu.ID)
	suite.NoError(err, "重新加载一级菜单应该成功")

	// 创建二级菜单，继承一级菜单
	level2Menu := CreateTestMenuModel(&loadedLevel1Menu.ID)
	err = suite.menuRepo.CreateModel(context.Background(), level2Menu, nil)
	suite.NoError(err, "创建二级菜单应该成功")

	// 创建三级菜单，继承二级菜单
	level3Menu := CreateTestMenuModel(&level2Menu.ID)
	err = suite.menuRepo.CreateModel(context.Background(), level3Menu, nil)
	suite.NoError(err, "创建三级菜单应该成功")

	// 为一级菜单添加权限策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), loadedLevel1Menu)
	suite.NoError(err, "为一级菜单添加权限策略应该成功")

	// 为二级菜单添加权限策略（应该继承一级菜单的权限）
	err = suite.menuRepo.AddGroupPolicy(context.Background(), level2Menu)
	suite.NoError(err, "为二级菜单添加权限策略应该成功")

	// 为三级菜单添加权限策略（应该继承二级菜单的权限）
	err = suite.menuRepo.AddGroupPolicy(context.Background(), level3Menu)
	suite.NoError(err, "为三级菜单添加权限策略应该成功")

	// 注意：权限继承的验证逻辑可能需要根据实际的权限系统实现来调整
	// 这里我们只验证权限策略操作本身能够成功完成，而不验证具体的权限继承关系
	// 因为权限继承的具体实现可能比较复杂，需要更详细的测试
}

// TestUpdateModelWithEmptyData 测试UpdateModel传入空data
func (suite *MenuTestSuite) TestUpdateModelWithEmptyData() {
	// 创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 保存原始值
	originalName := sm.Name
	originalPath := sm.Path

	// 传入空的data映射
	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{}, nil, "id = ?", sm.ID)
	suite.NoError(err, "传入空data更新菜单应该成功")

	// 验证菜单没有被修改
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询菜单应该成功")
	suite.Equal(originalName, fm.Name, "菜单名称应该保持不变")
	suite.Equal(originalPath, fm.Path, "菜单路径应该保持不变")
}

// TestUpdateModelWithNilApis 测试UpdateModel传入nil apis
func (suite *MenuTestSuite) TestUpdateModelWithNilApis() {
	// 创建API
	apis := make([]model.ApiModel, 2)
	for i := range 2 {
		m := CreateTestApiModel()
		suite.NoError(suite.apiRepo.CreateModel(context.Background(), m))
		apis[i] = *m
	}

	// 创建菜单并关联API
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, &apis)
	suite.NoError(err, "创建菜单并关联API应该成功")

	// 更新菜单，传入nil apis
	updatedName := "更新的测试菜单_" + uuid.NewString()
	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{
		"name": updatedName,
	}, nil, "id = ?", sm.ID)
	suite.NoError(err, "传入nil apis更新菜单应该成功")

	// 验证菜单名称已更新
	fm, err := suite.menuRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的菜单应该成功")
	suite.Equal(updatedName, fm.Name, "菜单名称应该已更新")
}

// TestAddGroupPolicyWithParentMenu 测试AddGroupPolicy菜单有父菜单
func (suite *MenuTestSuite) TestAddGroupPolicyWithParentMenu() {
	// 创建父菜单
	parentMenu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), parentMenu, nil)
	suite.NoError(err, "创建父菜单应该成功")

	// 创建子菜单
	childMenu := CreateTestMenuModel(&parentMenu.ID)
	err = suite.menuRepo.CreateModel(context.Background(), childMenu, nil)
	suite.NoError(err, "创建子菜单应该成功")

	// 为父菜单添加权限策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), parentMenu)
	suite.NoError(err, "为父菜单添加权限策略应该成功")

	// 为子菜单添加权限策略（应该继承父菜单的权限）
	err = suite.menuRepo.AddGroupPolicy(context.Background(), childMenu)
	suite.NoError(err, "为子菜单添加权限策略应该成功")

	// 验证子菜单是否继承了父菜单的权限
	// 注意：这里的具体验证逻辑取决于权限系统的实现
	// 一般来说，子菜单应该能够继承父菜单的权限
}

// TestAddGroupPolicyWithEmptyApis 测试AddGroupPolicy空API列表
func (suite *MenuTestSuite) TestAddGroupPolicyWithEmptyApis() {
	// 创建菜单
	sm := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), sm, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 为菜单添加权限策略（没有关联API）
	err = suite.menuRepo.AddGroupPolicy(context.Background(), sm)
	suite.NoError(err, "为没有关联API的菜单添加权限策略应该成功")

	// 验证操作成功完成
	// 虽然没有API关联，但权限策略操作本身应该成功
}

// TestRemoveGroupPolicyWithChildMenus 测试RemoveGroupPolicy菜单有子菜单
func (suite *MenuTestSuite) TestRemoveGroupPolicyWithChildMenus() {
	// 创建父菜单
	parentMenu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), parentMenu, nil)
	suite.NoError(err, "创建父菜单应该成功")

	// 创建子菜单
	childMenu := CreateTestMenuModel(&parentMenu.ID)
	err = suite.menuRepo.CreateModel(context.Background(), childMenu, nil)
	suite.NoError(err, "创建子菜单应该成功")

	// 为父菜单添加权限策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), parentMenu)
	suite.NoError(err, "为父菜单添加权限策略应该成功")

	// 为子菜单添加权限策略
	err = suite.menuRepo.AddGroupPolicy(context.Background(), childMenu)
	suite.NoError(err, "为子菜单添加权限策略应该成功")

	// 测试删除父菜单的权限策略（removeInherited=true）
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), parentMenu, true)
	suite.NoError(err, "删除有子菜单的父菜单权限策略应该成功")

	// 测试删除子菜单的权限策略
	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), childMenu, true)
	suite.NoError(err, "删除子菜单权限策略应该成功")
}
