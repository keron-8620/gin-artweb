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

// 使用 menu_test.go 中的 CreateTestMenuModel 函数

func CreateTestButtonModel(menuID uint32) *sysmodel.ButtonModel {
	return &sysmodel.ButtonModel{
		MenuID:   menuID,
		Name:     fmt.Sprintf("test_button_%s", uuid.NewString()),
		Sort:     1,
		IsActive: true,
		Descr:    "这是一个测试按钮",
	}
}

type ButtonTestSuite struct {
	suite.Suite
	apiRepo    *ApiRepo
	buttonRepo *ButtonRepo
	menuRepo   *MenuRepo
}

// 通用测试数据准备函数
func (suite *ButtonTestSuite) prepareTestMenuAndButton() (*sysmodel.MenuModel, *sysmodel.ButtonModel) {
	apis := []sysmodel.ApiModel{}
	menu := CreateTestMenuModel(apis)
	err := suite.menuRepo.CreateModel(context.Background(), menu, apis)
	suite.NoError(err, "创建菜单应该成功")

	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	return menu, button
}

// 通用测试数据准备函数（带API）
func (suite *ButtonTestSuite) prepareTestMenuButtonAndApi() (*sysmodel.MenuModel, *sysmodel.ButtonModel, *sysmodel.ApiModel) {
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	api := CreateTestApiModel()
	err = suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")

	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	return menu, button, api
}

func (suite *ButtonTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&sysmodel.MenuModel{}, &sysmodel.ButtonModel{}, &sysmodel.ApiModel{}); err != nil {
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
	suite.buttonRepo = &ButtonRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: slowThreshold,
		enforcer:      enforcer,
	}
}

func (suite *ButtonTestSuite) TestCreateButton() {
	// 准备测试数据
	_, button := suite.prepareTestMenuAndButton()

	// 验证按钮是否创建成功
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, button.ID)
	suite.NoError(err, "查询刚创建的按钮应该成功")
	suite.Equal(button.ID, fm.ID, "按钮ID应该匹配")
	suite.Equal(button.MenuID, fm.MenuID, "菜单ID应该匹配")
	suite.Equal(button.Name, fm.Name, "按钮名称应该匹配")
	suite.Equal(button.Sort, fm.Sort, "排序值应该匹配")
	suite.Equal(button.IsActive, fm.IsActive, "激活状态应该匹配")
	suite.Equal(button.Descr, fm.Descr, "描述应该匹配")
}

func (suite *ButtonTestSuite) TestUpdateButton() {
	// 准备测试数据
	_, button := suite.prepareTestMenuAndButton()

	// 更新按钮
	updatedName := "updated_button"
	updatedSort := uint32(2)
	updatedDescr := "这是更新的测试按钮"
	err := suite.buttonRepo.UpdateModel(context.Background(), map[string]any{
		"name":      updatedName,
		"sort":      updatedSort,
		"descr":     updatedDescr,
		"is_active": false,
	}, nil, "id = ?", button.ID)
	suite.NoError(err, "更新按钮应该成功")

	// 验证按钮是否更新成功
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询更新后的按钮应该成功")
	suite.Equal(button.ID, fm.ID, "按钮ID应该保持不变")
	suite.Equal(updatedName, fm.Name, "按钮名称应该更新")
	suite.Equal(updatedSort, fm.Sort, "排序值应该更新")
	suite.Equal(false, fm.IsActive, "激活状态应该更新")
	suite.Equal(updatedDescr, fm.Descr, "描述应该更新")
	suite.Greater(fm.UpdatedAt, button.UpdatedAt, "更新时间应该晚于创建时间")
}

func (suite *ButtonTestSuite) TestDeleteButton() {
	// 准备测试数据
	_, button := suite.prepareTestMenuAndButton()

	// 验证按钮是否创建成功
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询刚创建的按钮应该成功")
	suite.Equal(button.ID, fm.ID, "按钮ID应该匹配")

	// 删除按钮
	err = suite.buttonRepo.DeleteModel(context.Background(), "id = ?", button.ID)
	suite.NoError(err, "删除按钮应该成功")

	// 验证按钮是否删除成功
	_, err = suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.Error(err, "删除后查询按钮应该返回错误")
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
}

func (suite *ButtonTestSuite) TestListButton() {
	// 准备测试数据
	_, menu := suite.prepareTestMenuAndButton()

	// 测试创建多个按钮
	for range 5 {
		button := CreateTestButtonModel(menu.ID)
		err := suite.buttonRepo.CreateModel(context.Background(), button, nil)
		suite.NoError(err, "创建按钮应该成功")
	}

	// 测试CountModel
	count, err := suite.buttonRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取按钮总数应该成功")
	suite.GreaterOrEqual(count, int64(5), "按钮总数应该至少有5条")

	// 测试查询按钮列表
	qp := database.QueryParams{
		Limit:  10,
		Offset: 0,
	}
	ms, err := suite.buttonRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询按钮列表应该成功")
	suite.NotNil(ms, "按钮列表不应该为nil")
	suite.GreaterOrEqual(len(ms), 5, "按钮列表应该至少有5条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:  2,
		Offset: 0,
	}
	pMs, err := suite.buttonRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询按钮列表应该成功")
	suite.NotNil(pMs, "分页按钮列表不应该为nil")
	suite.Equal(2, len(pMs), "分页查询应该返回指定数量的记录")
}

func (suite *ButtonTestSuite) TestGetButtonWithEmptyConditions() {
	// 测试查询时传入空条件
	result, err := suite.buttonRepo.GetModel(context.Background(), []string{})
	// 当传入空条件时，GetModel方法会尝试获取数据库中的第一条记录
	// 如果数据库为空，会返回record not found错误
	// 如果数据库不为空，会返回第一条记录
	if err != nil {
		// 如果返回错误，应该是record not found
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询时传入空条件应该返回记录未找到错误")
	} else {
		// 如果返回结果，应该是一个有效的按钮模型
		suite.NotNil(result, "查询时传入空条件应该返回有效的按钮模型")
		suite.Greater(result.ID, uint32(0), "返回的按钮模型ID应该大于0")
	}
}

func (suite *ButtonTestSuite) TestCreateButtonWithInvalidData() {
	// 测试创建按钮时传入空数据
	err := suite.buttonRepo.CreateModel(context.Background(), nil, nil)
	suite.Error(err, "创建按钮时传入nil应该返回错误")
}

func (suite *ButtonTestSuite) TestFindNonExistentButton() {
	// 测试查找不存在的按钮
	_, err := suite.buttonRepo.GetModel(context.Background(), []string{}, 999999)
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查找不存在的按钮应该返回记录未找到错误")
}

func (suite *ButtonTestSuite) TestUpdateButtonWithEmptyData() {
	// 测试更新时传入空数据
	err := suite.buttonRepo.UpdateModel(context.Background(), map[string]any{}, nil, "id = ?", 1)
	suite.Error(err, "更新按钮时传入空数据应该返回错误")
}

func (suite *ButtonTestSuite) TestUpdateNonExistentButton() {
	// 测试更新不存在的按钮
	err := suite.buttonRepo.UpdateModel(context.Background(), map[string]any{
		"name": "updated_button",
	}, nil, "id = ?", 999999)
	suite.Error(err, "更新不存在的按钮应该返回错误")
}

func (suite *ButtonTestSuite) TestDeleteButtonWithEmptyConditions() {
	// 测试删除时传入空条件
	err := suite.buttonRepo.DeleteModel(context.Background())
	suite.Error(err, "删除时传入空条件应该返回错误")
}

func (suite *ButtonTestSuite) TestDeleteNonExistentButton() {
	// 测试删除不存在的按钮
	err := suite.buttonRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的按钮不应该返回错误")
}

// TestButtonWithContextCancellation 测试上下文取消和超时情况
func (suite *ButtonTestSuite) TestButtonWithContextCancellation() {
	// 准备测试数据
	_, button := suite.prepareTestMenuAndButton()

	// 测试场景1: 上下文已取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试创建按钮
	button2 := CreateTestButtonModel(button.MenuID)
	err := suite.buttonRepo.CreateModel(ctx, button2, nil)
	suite.Error(err, "创建按钮时上下文已取消应该返回错误")

	// 测试场景2: 上下文超时
	testCtx := context.Background()
	ctx, cancel = context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	time.Sleep(time.Millisecond * 5)

	// 测试更新按钮
	err = suite.buttonRepo.UpdateModel(ctx, map[string]any{
		"name": "updated_button",
	}, nil, "id = ?", button.ID)
	suite.Error(err, "更新按钮时上下文超时应该返回错误")

	// 测试删除按钮
	err = suite.buttonRepo.DeleteModel(ctx, "id = ?", button.ID)
	suite.Error(err, "删除按钮时上下文超时应该返回错误")

	// 测试获取按钮
	_, err = suite.buttonRepo.GetModel(ctx, []string{}, button.ID)
	suite.Error(err, "获取按钮时上下文超时应该返回错误")

	// 测试列表查询
	qp := database.QueryParams{}
	_, err = suite.buttonRepo.ListModel(ctx, qp)
	suite.Error(err, "列表查询时上下文超时应该返回错误")

	// 测试计数查询
	_, err = suite.buttonRepo.CountModel(ctx, nil)
	suite.Error(err, "计数查询时上下文超时应该返回错误")
}

// TestNewButtonRepo 测试创建按钮仓库实例
func (suite *ButtonTestSuite) TestNewButtonRepo() {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, _ := auth.NewCasbinEnforcer()

	repo := NewButtonRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	suite.NotNil(repo, "NewButtonRepo should return a non-nil repository")
	suite.NotNil(repo.log, "Repo log should not be nil")
	suite.NotNil(repo.gormDB, "Repo gormDB should not be nil")
	suite.NotNil(repo.timeouts, "Repo timeouts should not be nil")
	suite.NotNil(repo.slowThreshold, "Repo slowThreshold should not be nil")
	suite.NotNil(repo.enforcer, "Repo enforcer should not be nil")
}

// TestAddGroupPolicy 测试添加按钮组策略
func (suite *ButtonTestSuite) TestAddGroupPolicy() {
	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建API
	api := CreateTestApiModel()
	err = suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")

	// 创建按钮并关联API
	button := CreateTestButtonModel(menu.ID)
	button.Apis = []sysmodel.ApiModel{*api}
	err = suite.buttonRepo.CreateModel(context.Background(), button, button.Apis)
	suite.NoError(err, "创建按钮应该成功")

	// 测试添加组策略
	err = suite.buttonRepo.AddGroupPolicy(context.Background(), *button)
	suite.NoError(err, "添加按钮组策略应该成功")
}

// TestButtonAddGroupPolicy 测试添加按钮组策略
func (suite *ButtonTestSuite) TestButtonAddGroupPolicy() {
	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 测试添加组策略
	err = suite.buttonRepo.AddGroupPolicy(context.Background(), *button)
	suite.NoError(err, "添加按钮组策略应该成功")
}

// TestButtonAddGroupPolicyWithInvalidAPI 测试添加包含无效API的按钮组策略
func (suite *ButtonTestSuite) TestButtonAddGroupPolicyWithInvalidAPI() {
	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 手动设置无效API（ID为0）
	button.Apis = []sysmodel.ApiModel{{URL: "/api/test", Method: "GET"}}

	// 测试添加组策略（应该跳过无效API）
	err = suite.buttonRepo.AddGroupPolicy(context.Background(), *button)
	suite.NoError(err, "添加包含无效API的按钮组策略应该成功")
}

// TestButtonAddGroupPolicyWithZeroMenuID 测试菜单ID为0时添加按钮组策略
func (suite *ButtonTestSuite) TestButtonAddGroupPolicyWithZeroMenuID() {
	// 创建按钮
	button := &sysmodel.ButtonModel{}
	button.ID = 1
	button.MenuID = 0

	// 测试添加组策略
	err := suite.buttonRepo.AddGroupPolicy(context.Background(), *button)
	suite.Error(err, "菜单ID为0时添加按钮组策略应该返回错误")
}

// TestButtonAddGroupPolicyWithCanceledContext 测试上下文已取消时添加按钮组策略
func (suite *ButtonTestSuite) TestButtonAddGroupPolicyWithCanceledContext() {
	// 创建一个已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 创建按钮
	button := &sysmodel.ButtonModel{}
	button.ID = 1
	button.MenuID = 1

	// 测试添加组策略
	err := suite.buttonRepo.AddGroupPolicy(ctx, *button)
	suite.Error(err, "上下文已取消时添加按钮组策略应该返回错误")
}

// TestButtonAddGroupPolicyWithZeroID 测试按钮ID为0时添加组策略
func (suite *ButtonTestSuite) TestButtonAddGroupPolicyWithZeroID() {
	// 创建按钮
	button := &sysmodel.ButtonModel{}
	button.ID = 0
	button.MenuID = 1

	// 测试添加组策略
	err := suite.buttonRepo.AddGroupPolicy(context.Background(), *button)
	suite.Error(err, "按钮ID为0时添加组策略应该返回错误")
}

// TestButtonRemoveGroupPolicy 测试删除按钮组策略
func (suite *ButtonTestSuite) TestButtonRemoveGroupPolicy() {
	// 准备测试数据
	_, button := suite.prepareTestMenuAndButton()

	// 先添加组策略
	err := suite.buttonRepo.AddGroupPolicy(context.Background(), *button)
	suite.NoError(err, "添加按钮组策略应该成功")

	// 测试场景1: 删除组策略（删除继承）
	err = suite.buttonRepo.RemoveGroupPolicy(context.Background(), *button, true)
	suite.NoError(err, "删除按钮组策略（删除继承）应该成功")

	// 重新添加组策略
	err = suite.buttonRepo.AddGroupPolicy(context.Background(), *button)
	suite.NoError(err, "重新添加按钮组策略应该成功")

	// 测试场景2: 删除组策略（不删除继承）
	err = suite.buttonRepo.RemoveGroupPolicy(context.Background(), *button, false)
	suite.NoError(err, "删除按钮组策略（不删除继承）应该成功")

	// 测试场景3: 删除不存在的组策略
	err = suite.buttonRepo.RemoveGroupPolicy(context.Background(), *button, false)
	suite.NoError(err, "删除不存在的组策略应该成功")
}

// TestButtonRemoveGroupPolicyEdgeCases 测试删除按钮组策略的边界情况
func (suite *ButtonTestSuite) TestButtonRemoveGroupPolicyEdgeCases() {
	// 测试场景1: 按钮ID为0
	button := sysmodel.ButtonModel{}
	err := suite.buttonRepo.RemoveGroupPolicy(context.Background(), button, true)
	suite.Error(err, "删除按钮组策略时ID为0应该返回错误")

	// 测试场景2: 上下文已取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	button = sysmodel.ButtonModel{StandardModel: database.StandardModel{BaseModel: database.BaseModel{ID: 1}}}
	err = suite.buttonRepo.RemoveGroupPolicy(ctx, button, true)
	suite.Error(err, "上下文已取消时删除按钮组策略应该返回错误")

	// 测试场景3: 上下文超时
	testCtx := context.Background()
	ctx, cancel = context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	time.Sleep(time.Millisecond * 5)
	_, button2 := suite.prepareTestMenuAndButton()
	err = suite.buttonRepo.AddGroupPolicy(context.Background(), *button2)
	suite.NoError(err, "添加按钮组策略应该成功")
	err = suite.buttonRepo.RemoveGroupPolicy(ctx, *button2, true)
	suite.Error(err, "删除按钮组策略时上下文超时应该返回错误")
}

// TestButtonUpdateModel 测试更新按钮
func (suite *ButtonTestSuite) TestButtonUpdateModel() {
	// 测试场景1: 更新按钮并关联API
	_, button, api := suite.prepareTestMenuButtonAndApi()
	api2 := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), api2)
	suite.NoError(err, "创建API2应该成功")

	updatedName := "updated_button_with_apis"
	apis := []sysmodel.ApiModel{*api, *api2}
	err = suite.buttonRepo.UpdateModel(context.Background(), map[string]any{
		"name": updatedName,
	}, apis, "id = ?", button.ID)
	suite.NoError(err, "更新按钮并关联API应该成功")

	// 验证按钮是否更新成功
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{"Apis"}, "id = ?", button.ID)
	suite.NoError(err, "查询更新后的按钮应该成功")
	suite.Equal(updatedName, fm.Name, "按钮名称更新失败")
	suite.Equal(2, len(fm.Apis), "按钮关联的API数量错误")

	// 测试场景2: 更新按钮但不关联API
	_, button2 := suite.prepareTestMenuAndButton()
	updatedName2 := "updated_button_without_apis"
	err = suite.buttonRepo.UpdateModel(context.Background(), map[string]any{
		"name": updatedName2,
	}, nil, "id = ?", button2.ID)
	suite.NoError(err, "更新按钮应该成功")

	// 验证按钮是否更新成功
	fm2, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button2.ID)
	suite.NoError(err, "查询更新后的按钮应该成功")
	suite.Equal(updatedName2, fm2.Name, "按钮名称更新失败")
}

// 每个测试文件都需要这个入口函数
func TestButtonTestSuite(t *testing.T) {
	pts := &ButtonTestSuite{}
	suite.Run(t, pts)
}
