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
	apiRepo    *ApiRepo
	menuRepo   *MenuRepo
	buttonRepo *ButtonRepo
}

func (suite *ButtonTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&model.ApiModel{},
		&model.MenuModel{},
		&model.ButtonModel{},
	)
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
	suite.buttonRepo = &ButtonRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}
}

func (suite *ButtonTestSuite) TestCreateButton() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 测试创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 测试查询刚创建的按钮
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询刚创建的按钮应该成功")
	suite.Equal(button.ID, fm.ID)
	suite.Equal(button.MenuID, fm.MenuID)
	suite.Equal(button.Name, fm.Name)
	suite.Equal(button.Sort, fm.Sort)
	suite.Equal(button.IsActive, fm.IsActive)
	suite.Equal(button.Descr, fm.Descr)
}

func (suite *ButtonTestSuite) TestUpdateButton() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 测试更新按钮
	updatedName := "更新的测试按钮"
	updatedSort := uint32(20000)
	updatedIsActive := false
	updatedDescr := "更新后的按钮描述"

	err = suite.buttonRepo.UpdateModel(context.Background(), map[string]any{
		"name":      updatedName,
		"sort":      updatedSort,
		"is_active": updatedIsActive,
		"descr":     updatedDescr,
	}, nil, "id = ?", button.ID)
	suite.NoError(err, "更新按钮应该成功")

	// 测试查询更新后的按钮
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询更新后的按钮应该成功")
	suite.Equal(button.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedSort, fm.Sort)
	suite.Equal(updatedIsActive, fm.IsActive)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Greater(fm.UpdatedAt, button.UpdatedAt)
}

func (suite *ButtonTestSuite) TestDeleteButton() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 测试查询刚创建的按钮
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询刚创建的按钮应该成功")
	suite.Equal(button.ID, fm.ID)

	// 测试删除按钮
	err = suite.buttonRepo.DeleteModel(context.Background(), "id = ?", button.ID)
	suite.NoError(err, "删除按钮应该成功")

	// 测试查询已删除的按钮
	_, err = suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *ButtonTestSuite) TestGetButton() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 测试根据ID查询按钮
	m, err := suite.buttonRepo.GetModel(context.Background(), []string{}, button.ID)
	suite.NoError(err, "根据ID查询按钮应该成功")
	suite.Equal(button.ID, m.ID)
	suite.Equal(button.MenuID, m.MenuID)
	suite.Equal(button.Name, m.Name)
	suite.Equal(button.Sort, m.Sort)
	suite.Equal(button.IsActive, m.IsActive)
	suite.Equal(button.Descr, m.Descr)
}

func (suite *ButtonTestSuite) TestListButton() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 测试创建多个按钮
	for range 5 {
		button := CreateTestButtonModel(menu.ID)
		err := suite.buttonRepo.CreateModel(context.Background(), button, nil)
		suite.NoError(err, "创建按钮应该成功")
	}

	// 测试查询按钮列表
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	count, ms, err := suite.buttonRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询按钮列表应该成功")
	suite.NotNil(ms, "按钮列表不应该为nil")
	suite.GreaterOrEqual(count, int64(5), "按钮总数应该至少有5条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:   2,
		Offset:  0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.buttonRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询按钮列表应该成功")
	suite.NotNil(pMs, "分页按钮列表不应该为nil")
	suite.Equal(2, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(2), "分页总数应该至少等于limit")
}

func (suite *ButtonTestSuite) TestAddGroupPolicy() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建API用于测试
	apis := make([]model.ApiModel, 2)
	for i := range 2 {
		api := CreateTestApiModel()
		err := suite.apiRepo.CreateModel(context.Background(), api)
		suite.NoError(err, "创建API应该成功")
		apis[i] = *api
	}

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, &apis)
	suite.NoError(err, "创建按钮应该成功")

	// 测试添加权限策略
	err = suite.buttonRepo.AddGroupPolicy(context.Background(), button)
	suite.NoError(err, "添加按钮权限策略应该成功")
}

func (suite *ButtonTestSuite) TestRemoveGroupPolicy() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建API用于测试
	api := CreateTestApiModel()
	err = suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	apis := []model.ApiModel{*api}
	err = suite.buttonRepo.CreateModel(context.Background(), button, &apis)
	suite.NoError(err, "创建按钮应该成功")

	// 添加权限策略
	err = suite.buttonRepo.AddGroupPolicy(context.Background(), button)
	suite.NoError(err, "添加按钮权限策略应该成功")

	// 测试删除权限策略
	err = suite.buttonRepo.RemoveGroupPolicy(context.Background(), button, true)
	suite.NoError(err, "删除按钮权限策略应该成功")
}

func (suite *ButtonTestSuite) TestAddGroupPolicyWithNilButton() {
	// 测试添加权限策略时传入nil按钮
	err := suite.buttonRepo.AddGroupPolicy(context.Background(), nil)
	suite.Error(err, "传入nil按钮应该返回错误")
	suite.Contains(err.Error(), "AddGroupPolicy操作失败: 按钮模型不能为空")
}

func (suite *ButtonTestSuite) TestAddGroupPolicyWithZeroID() {
	// 测试添加权限策略时传入ID为0的按钮
	button := &model.ButtonModel{
		MenuID: 1,
	}
	err := suite.buttonRepo.AddGroupPolicy(context.Background(), button)
	suite.Error(err, "传入ID为0的按钮应该返回错误")
	suite.Contains(err.Error(), "AddGroupPolicy操作失败: 按钮ID不能为0")
}

func (suite *ButtonTestSuite) TestRemoveGroupPolicyWithNilButton() {
	// 测试删除权限策略时传入nil按钮
	err := suite.buttonRepo.RemoveGroupPolicy(context.Background(), nil, true)
	suite.Error(err, "传入nil按钮应该返回错误")
	suite.Contains(err.Error(), "RemoveGroupPolicy操作失败: 按钮模型不能为空")
}

func (suite *ButtonTestSuite) TestRemoveGroupPolicyWithZeroID() {
	// 测试删除权限策略时传入ID为0的按钮
	button := &model.ButtonModel{
		MenuID: 1,
	}
	err := suite.buttonRepo.RemoveGroupPolicy(context.Background(), button, true)
	suite.Error(err, "传入ID为0的按钮应该返回错误")
	suite.Contains(err.Error(), "RemoveGroupPolicy操作失败: 按钮ID不能为0")
}

func (suite *ButtonTestSuite) TestRemoveGroupPolicyWithRemoveInherited() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 添加权限策略
	err = suite.buttonRepo.AddGroupPolicy(context.Background(), button)
	suite.NoError(err, "添加按钮权限策略应该成功")

	// 测试删除权限策略并删除继承的组策略
	err = suite.buttonRepo.RemoveGroupPolicy(context.Background(), button, true)
	suite.NoError(err, "删除按钮权限策略并删除继承的组策略应该成功")
}

func (suite *ButtonTestSuite) TestRemoveGroupPolicyWithoutRemoveInherited() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 添加权限策略
	err = suite.buttonRepo.AddGroupPolicy(context.Background(), button)
	suite.NoError(err, "添加按钮权限策略应该成功")

	// 测试删除权限策略但不删除继承的组策略
	err = suite.buttonRepo.RemoveGroupPolicy(context.Background(), button, false)
	suite.NoError(err, "删除按钮权限策略但不删除继承的组策略应该成功")
}

func (suite *ButtonTestSuite) TestCreateButtonWithNilModel() {
	// 测试创建按钮时传入nil模型
	err := suite.buttonRepo.CreateModel(context.Background(), nil, nil)
	suite.Error(err, "传入nil模型应该返回错误")
	suite.Contains(err.Error(), "创建按钮模型失败: 模型为空")
}

func (suite *ButtonTestSuite) TestCreateButtonWithEmptyApis() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 测试创建按钮时传入空的APIs列表
	button := CreateTestButtonModel(menu.ID)
	emptyApis := []model.ApiModel{}
	err = suite.buttonRepo.CreateModel(context.Background(), button, &emptyApis)
	suite.NoError(err, "传入空的APIs列表应该成功创建按钮")

	// 验证按钮创建成功
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询刚创建的按钮应该成功")
	suite.Equal(button.ID, fm.ID)
}

func (suite *ButtonTestSuite) TestUpdateButtonWithEmptyData() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 传入空的data映射
	err = suite.buttonRepo.UpdateModel(context.Background(), map[string]any{}, nil, "id = ?", button.ID)
	suite.Error(err, "传入空data更新按钮应该返回错误")
	suite.Contains(err.Error(), "更新按钮模型失败: 更新数据为空")
}

func (suite *ButtonTestSuite) TestUpdateButtonWithNilApis() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建API用于测试
	api := CreateTestApiModel()
	err = suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")

	// 创建按钮并关联API
	button := CreateTestButtonModel(menu.ID)
	apis := []model.ApiModel{*api}
	err = suite.buttonRepo.CreateModel(context.Background(), button, &apis)
	suite.NoError(err, "创建按钮并关联API应该成功")

	// 更新按钮，传入nil apis
	updatedName := "更新的测试按钮_" + uuid.NewString()
	err = suite.buttonRepo.UpdateModel(context.Background(), map[string]any{
		"name": updatedName,
	}, nil, "id = ?", button.ID)
	suite.NoError(err, "传入nil apis更新按钮应该成功")

	// 验证按钮名称已更新
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询更新后的按钮应该成功")
	suite.Equal(updatedName, fm.Name, "按钮名称应该已更新")
}

func (suite *ButtonTestSuite) TestGetButtonWithEmptyConditions() {
	// 测试查询时传入空条件
	_, err := suite.buttonRepo.GetModel(context.Background(), []string{})
	// 当传入空条件时，GetModel方法会尝试获取数据库中的第一条记录
	// 如果数据库为空，会返回record not found错误
	// 如果数据库不为空，会返回第一条记录
	if err != nil {
		// 如果返回错误，应该是record not found
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询时传入空条件应该返回记录未找到错误")
	} else {
		// 如果返回结果，应该是一个有效的按钮模型
		// 这里不做断言，因为可能没有数据
	}
}

func (suite *ButtonTestSuite) TestGetButtonWithNonExistentID() {
	// 测试查询不存在的按钮ID
	_, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", uint32(999999))
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询不存在的按钮ID应该返回记录未找到错误")
}

func (suite *ButtonTestSuite) TestDeleteButtonWithNonExistentID() {
	// 测试删除不存在的按钮ID
	err := suite.buttonRepo.DeleteModel(context.Background(), "id = ?", uint32(999999))
	suite.NoError(err, "删除不存在的按钮ID应该成功")
}

func (suite *ButtonTestSuite) TestUpdateButtonWithNonExistentID() {
	// 测试更新不存在的按钮ID
	err := suite.buttonRepo.UpdateModel(context.Background(), map[string]any{
		"name": "更新的测试按钮",
	}, nil, "id = ?", uint32(999999))
	suite.NoError(err, "更新不存在的按钮ID应该成功")
}

func (suite *ButtonTestSuite) TestListButtonWithPaginationBoundaries() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 测试Limit=0的情况
	qpZeroLimit := database.QueryParams{
		Limit:   0,
		Offset:  0,
		IsCount: true,
	}
	_, msZero, err := suite.buttonRepo.ListModel(context.Background(), qpZeroLimit)
	suite.NoError(err, "Limit=0应该成功查询")
	suite.NotNil(msZero, "按钮列表不应该为nil")

	// 测试较大的Offset值
	qpLargeOffset := database.QueryParams{
		Limit:   10,
		Offset:  999999,
		IsCount: true,
	}
	_, msLarge, err := suite.buttonRepo.ListModel(context.Background(), qpLargeOffset)
	suite.NoError(err, "较大的Offset值应该成功查询")
	suite.NotNil(msLarge, "按钮列表不应该为nil")
	suite.LessOrEqual(len(*msLarge), 10, "返回的记录数应该不超过Limit")
}

func (suite *ButtonTestSuite) TestListButtonWithNoRecords() {
	// 测试查询不存在的条件
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
		Query:   map[string]any{"id": uint32(999999)},
	}
	count, ms, err := suite.buttonRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询无记录的按钮列表应该成功")
	suite.NotNil(ms, "按钮列表不应该为nil")
	suite.Equal(int64(0), count, "无记录时计数应该为0")
	suite.Len(*ms, 0, "无记录时按钮列表长度应该为0")
}

func (suite *ButtonTestSuite) TestContextTimeout() {
	// 创建一个会立即超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// 等待上下文超时
	time.Sleep(10 * time.Nanosecond)

	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮用于测试
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 测试CreateModel方法
	sm := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(ctx, sm, nil)
	suite.Error(err, "上下文超时后创建按钮应该返回错误")

	// 测试UpdateModel方法
	err = suite.buttonRepo.UpdateModel(ctx, map[string]any{
		"name": "测试按钮",
	}, nil, "id = ?", button.ID)
	suite.Error(err, "上下文超时后更新按钮应该返回错误")

	// 测试DeleteModel方法
	err = suite.buttonRepo.DeleteModel(ctx, "id = ?", button.ID)
	suite.Error(err, "上下文超时后删除按钮应该返回错误")

	// 测试GetModel方法
	_, err = suite.buttonRepo.GetModel(ctx, []string{}, "id = ?", button.ID)
	suite.Error(err, "上下文超时后获取按钮应该返回错误")

	// 测试ListModel方法
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	_, _, err = suite.buttonRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文超时后列出按钮应该返回错误")

	// 测试AddGroupPolicy方法
	err = suite.buttonRepo.AddGroupPolicy(ctx, button)
	suite.Error(err, "上下文超时后添加权限策略应该返回错误")

	// 测试RemoveGroupPolicy方法
	err = suite.buttonRepo.RemoveGroupPolicy(ctx, button, true)
	suite.Error(err, "上下文超时后删除权限策略应该返回错误")
}

func (suite *ButtonTestSuite) TestContextCancel() {
	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	// 立即取消上下文
	cancel()

	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮用于测试
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 测试CreateModel方法
	sm := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(ctx, sm, nil)
	suite.Error(err, "上下文取消后创建按钮应该返回错误")

	// 测试UpdateModel方法
	err = suite.buttonRepo.UpdateModel(ctx, map[string]any{
		"name": "测试按钮",
	}, nil, "id = ?", button.ID)
	suite.Error(err, "上下文取消后更新按钮应该返回错误")

	// 测试DeleteModel方法
	err = suite.buttonRepo.DeleteModel(ctx, "id = ?", button.ID)
	suite.Error(err, "上下文取消后删除按钮应该返回错误")

	// 测试GetModel方法
	_, err = suite.buttonRepo.GetModel(ctx, []string{}, "id = ?", button.ID)
	suite.Error(err, "上下文取消后获取按钮应该返回错误")

	// 测试ListModel方法
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	_, _, err = suite.buttonRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文取消后列出按钮应该返回错误")

	// 测试AddGroupPolicy方法
	err = suite.buttonRepo.AddGroupPolicy(ctx, button)
	suite.Error(err, "上下文取消后添加权限策略应该返回错误")

	// 测试RemoveGroupPolicy方法
	err = suite.buttonRepo.RemoveGroupPolicy(ctx, button, true)
	suite.Error(err, "上下文取消后删除权限策略应该返回错误")
}

func (suite *ButtonTestSuite) TestCreateButtonWithMenu() {
	// 测试创建与菜单关联的按钮
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮并关联到菜单
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建与菜单关联的按钮应该成功")

	// 验证按钮与菜单的关联
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询按钮应该成功")
	suite.Equal(menu.ID, fm.MenuID, "按钮应该与正确的菜单关联")
}

func (suite *ButtonTestSuite) TestCreateButtonWithApis() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建API用于测试
	apis := make([]model.ApiModel, 3)
	for i := range 3 {
		api := CreateTestApiModel()
		err := suite.apiRepo.CreateModel(context.Background(), api)
		suite.NoError(err, "创建API应该成功")
		apis[i] = *api
	}

	// 创建按钮并关联API
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, &apis)
	suite.NoError(err, "创建与API关联的按钮应该成功")

	// 验证按钮创建成功
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询按钮应该成功")
	suite.Equal(button.ID, fm.ID)
}

func (suite *ButtonTestSuite) TestPreloadApis() {
	// 先创建一个菜单用于测试
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建API用于测试
	apis := make([]model.ApiModel, 2)
	for i := range 2 {
		api := CreateTestApiModel()
		err := suite.apiRepo.CreateModel(context.Background(), api)
		suite.NoError(err, "创建API应该成功")
		apis[i] = *api
	}

	// 创建按钮并关联API
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, &apis)
	suite.NoError(err, "创建与API关联的按钮应该成功")

	// 不预加载API查询按钮
	buttonWithoutApis, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "不预加载API查询按钮应该成功")

	// 预加载API查询按钮
	buttonWithApis, err := suite.buttonRepo.GetModel(context.Background(), []string{"Apis"}, "id = ?", button.ID)
	suite.NoError(err, "预加载API查询按钮应该成功")
	suite.Equal(buttonWithoutApis.ID, buttonWithApis.ID)
	suite.NotEmpty(buttonWithApis.Apis, "预加载后应该包含API")
	suite.Len(buttonWithApis.Apis, 2, "预加载后应该包含2个API")
}

// 每个测试文件都需要这个入口函数
func TestButtonTestSuite(t *testing.T) {
	pts := &ButtonTestSuite{}
	suite.Run(t, pts)
}
