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
	buttonRepo *ButtonRepo
	menuRepo   *MenuRepo
}

func (suite *ButtonTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&sysmodel.MenuModel{}, &sysmodel.ButtonModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.buttonRepo = &ButtonRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
	suite.menuRepo = &MenuRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *ButtonTestSuite) TestCreateButton() {
	// 先创建一个菜单
	apis := []sysmodel.ApiModel{}
	menu := CreateTestMenuModel(apis)
	err := suite.menuRepo.CreateModel(context.Background(), menu, apis)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 验证按钮是否创建成功
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, button.ID)
	suite.NoError(err, "查询刚创建的按钮应该成功")
	suite.Equal(button.ID, fm.ID)
	suite.Equal(button.MenuID, fm.MenuID)
	suite.Equal(button.Name, fm.Name)
	suite.Equal(button.Sort, fm.Sort)
	suite.Equal(button.IsActive, fm.IsActive)
	suite.Equal(button.Descr, fm.Descr)
}

func (suite *ButtonTestSuite) TestUpdateButton() {
	// 先创建一个菜单
	apis := []sysmodel.ApiModel{}
	menu := CreateTestMenuModel(apis)
	err := suite.menuRepo.CreateModel(context.Background(), menu, apis)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 更新按钮
	updatedName := "updated_button"
	updatedSort := uint32(2)
	updatedDescr := "这是更新的测试按钮"
	// 更新按钮
	err = suite.buttonRepo.UpdateModel(context.Background(), map[string]any{
		"name":      updatedName,
		"sort":      updatedSort,
		"descr":     updatedDescr,
		"is_active": false,
	}, nil, "id = ?", button.ID)
	suite.NoError(err, "更新按钮应该成功")

	// 验证按钮是否更新成功
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询更新后的按钮应该成功")
	suite.Equal(button.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedSort, fm.Sort)
	suite.Equal(false, fm.IsActive)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Greater(fm.UpdatedAt, button.UpdatedAt)
}

func (suite *ButtonTestSuite) TestDeleteButton() {
	// 先创建一个菜单
	apis := []sysmodel.ApiModel{}
	menu := CreateTestMenuModel(apis)
	err := suite.menuRepo.CreateModel(context.Background(), menu, apis)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 验证按钮是否创建成功
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	suite.NoError(err, "查询刚创建的按钮应该成功")
	suite.Equal(button.ID, fm.ID)

	// 删除按钮
	err = suite.buttonRepo.DeleteModel(context.Background(), "id = ?", button.ID)
	suite.NoError(err, "删除按钮应该成功")

	// 验证按钮是否删除成功
	_, err = suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", button.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *ButtonTestSuite) TestListButton() {
	// 先创建一个菜单
	apis := []sysmodel.ApiModel{}
	menu := CreateTestMenuModel(apis)
	err := suite.menuRepo.CreateModel(context.Background(), menu, apis)
	suite.NoError(err, "创建菜单应该成功")

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
	suite.NoError(err, "更新不存在的按钮不应该返回错误")
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

func (suite *ButtonTestSuite) TestCreateButtonWithCanceledContext() {
	// 测试创建按钮时上下文已取消
	testCtx := context.Background()
	ctx, cancel := context.WithCancel(testCtx)
	cancel()

	// 先创建一个菜单
	apis := []sysmodel.ApiModel{}
	menu := CreateTestMenuModel(apis)
	err := suite.menuRepo.CreateModel(context.Background(), menu, apis)
	suite.NoError(err, "创建菜单应该成功")

	// 尝试创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(ctx, button, nil)
	suite.Error(err, "创建按钮时上下文已取消应该返回错误")
}

func (suite *ButtonTestSuite) TestUpdateButtonWithContextTimeout() {
	// 测试更新按钮时上下文超时
	// 先创建一个菜单
	apis := []sysmodel.ApiModel{}
	menu := CreateTestMenuModel(apis)
	err := suite.menuRepo.CreateModel(context.Background(), menu, apis)
	suite.NoError(err, "创建菜单应该成功")

	// 创建一个按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试更新按钮
	err = suite.buttonRepo.UpdateModel(ctx, map[string]any{
		"name": "updated_button",
	}, nil, "id = ?", button.ID)
	suite.Error(err, "更新按钮时上下文超时应该返回错误")
}

func (suite *ButtonTestSuite) TestDeleteButtonWithContextTimeout() {
	// 测试删除按钮时上下文超时
	// 先创建一个菜单
	apis := []sysmodel.ApiModel{}
	menu := CreateTestMenuModel(apis)
	err := suite.menuRepo.CreateModel(context.Background(), menu, apis)
	suite.NoError(err, "创建菜单应该成功")

	// 创建一个按钮
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试删除按钮
	err = suite.buttonRepo.DeleteModel(ctx, "id = ?", button.ID)
	suite.Error(err, "删除按钮时上下文超时应该返回错误")
}

func (suite *ButtonTestSuite) TestGetButtonWithContextTimeout() {
	// 测试获取按钮时上下文超时
	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试获取按钮
	_, err := suite.buttonRepo.GetModel(ctx, []string{}, 1)
	suite.Error(err, "获取按钮时上下文超时应该返回错误")
}

func (suite *ButtonTestSuite) TestListButtonWithContextTimeout() {
	// 测试列表查询时上下文超时
	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试列表查询
	qp := database.QueryParams{}
	_, err := suite.buttonRepo.ListModel(ctx, qp)
	suite.Error(err, "列表查询时上下文超时应该返回错误")
}

func (suite *ButtonTestSuite) TestCountButtonWithContextTimeout() {
	// 测试计数查询时上下文超时
	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试计数查询
	_, err := suite.buttonRepo.CountModel(ctx, nil)
	suite.Error(err, "计数查询时上下文超时应该返回错误")
}

// 每个测试文件都需要这个入口函数
func TestButtonTestSuite(t *testing.T) {
	pts := &ButtonTestSuite{}
	suite.Run(t, pts)
}

// TestNewButtonRepo 测试创建按钮仓库实例
func TestNewButtonRepo(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()

	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)
	if repo == nil {
		t.Fatal("NewButtonRepo should return a non-nil repository")
	}
	if repo.log == nil {
		t.Fatal("Repo log should not be nil")
	}
	if repo.gormDB == nil {
		t.Fatal("Repo gormDB should not be nil")
	}
	if repo.timeouts == nil {
		t.Fatal("Repo timeouts should not be nil")
	}
	if repo.enforcer == nil {
		t.Fatal("Repo enforcer should not be nil")
	}
}

// TestAddGroupPolicy 测试添加按钮组策略
func TestAddGroupPolicy(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&sysmodel.MenuModel{}, &sysmodel.ButtonModel{}, &sysmodel.ApiModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)
	menuRepo := NewMenuRepo(logger, db, dbTimeout, enforcer)
	apiRepo := NewApiRepo(logger, db, dbTimeout, enforcer)

	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := menuRepo.CreateModel(context.Background(), menu, nil)
	if err != nil {
		t.Fatalf("创建菜单失败: %v", err)
	}

	// 创建API
	api := CreateTestApiModel()
	err = apiRepo.CreateModel(context.Background(), api)
	if err != nil {
		t.Fatalf("创建API失败: %v", err)
	}

	// 创建按钮并关联API
	button := CreateTestButtonModel(menu.ID)
	button.Apis = []sysmodel.ApiModel{*api}
	err = repo.CreateModel(context.Background(), button, button.Apis)
	if err != nil {
		t.Fatalf("创建按钮失败: %v", err)
	}

	// 测试添加组策略
	err = repo.AddGroupPolicy(context.Background(), button)
	if err != nil {
		t.Fatalf("添加按钮组策略失败: %v", err)
	}
}

// TestButtonAddGroupPolicy 测试添加按钮组策略
func TestButtonAddGroupPolicy(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&sysmodel.MenuModel{}, &sysmodel.ButtonModel{}, &sysmodel.ApiModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)
	menuRepo := NewMenuRepo(logger, db, dbTimeout, enforcer)

	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := menuRepo.CreateModel(context.Background(), menu, nil)
	if err != nil {
		t.Fatalf("创建菜单失败: %v", err)
	}

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = repo.CreateModel(context.Background(), button, nil)
	if err != nil {
		t.Fatalf("创建按钮失败: %v", err)
	}

	// 测试添加组策略
	err = repo.AddGroupPolicy(context.Background(), button)
	if err != nil {
		t.Fatalf("添加按钮组策略失败: %v", err)
	}
}

// TestButtonAddGroupPolicyWithInvalidAPI 测试添加包含无效API的按钮组策略
func TestButtonAddGroupPolicyWithInvalidAPI(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&sysmodel.MenuModel{}, &sysmodel.ButtonModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)
	menuRepo := NewMenuRepo(logger, db, dbTimeout, enforcer)

	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := menuRepo.CreateModel(context.Background(), menu, nil)
	if err != nil {
		t.Fatalf("创建菜单失败: %v", err)
	}

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = repo.CreateModel(context.Background(), button, nil)
	if err != nil {
		t.Fatalf("创建按钮失败: %v", err)
	}

	// 手动设置无效API（ID为0）
	button.Apis = []sysmodel.ApiModel{{URL: "/api/test", Method: "GET"}}

	// 测试添加组策略（应该跳过无效API）
	err = repo.AddGroupPolicy(context.Background(), button)
	if err != nil {
		t.Fatalf("添加包含无效API的按钮组策略失败: %v", err)
	}
}

// TestButtonAddGroupPolicyWithZeroMenuID 测试菜单ID为0时添加按钮组策略
func TestButtonAddGroupPolicyWithZeroMenuID(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)

	// 创建按钮
	button := &sysmodel.ButtonModel{}
	button.ID = 1
	button.MenuID = 0

	// 测试添加组策略
	err := repo.AddGroupPolicy(context.Background(), button)
	if err == nil {
		t.Fatal("菜单ID为0时添加按钮组策略应该返回错误")
	}
}

// TestButtonAddGroupPolicyWithCanceledContext 测试上下文已取消时添加按钮组策略
func TestButtonAddGroupPolicyWithCanceledContext(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)

	// 创建一个已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 创建按钮
	button := &sysmodel.ButtonModel{}
	button.ID = 1
	button.MenuID = 1

	// 测试添加组策略
	err := repo.AddGroupPolicy(ctx, button)
	if err == nil {
		t.Fatal("上下文已取消时添加按钮组策略应该返回错误")
	}
}

// TestButtonAddGroupPolicyWithNilButton 测试按钮为nil时添加组策略
func TestButtonAddGroupPolicyWithNilButton(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)

	// 测试添加组策略
	err := repo.AddGroupPolicy(context.Background(), nil)
	if err == nil {
		t.Fatal("按钮为nil时添加组策略应该返回错误")
	}
}

// TestButtonAddGroupPolicyWithZeroID 测试按钮ID为0时添加组策略
func TestButtonAddGroupPolicyWithZeroID(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)

	// 创建按钮
	button := &sysmodel.ButtonModel{}
	button.ID = 0
	button.MenuID = 1

	// 测试添加组策略
	err := repo.AddGroupPolicy(context.Background(), button)
	if err == nil {
		t.Fatal("按钮ID为0时添加组策略应该返回错误")
	}
}

// TestButtonRemoveGroupPolicy 测试删除按钮组策略
func TestButtonRemoveGroupPolicy(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&sysmodel.MenuModel{}, &sysmodel.ButtonModel{}, &sysmodel.ApiModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)
	menuRepo := NewMenuRepo(logger, db, dbTimeout, enforcer)

	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err := menuRepo.CreateModel(context.Background(), menu, nil)
	if err != nil {
		t.Fatalf("创建菜单失败: %v", err)
	}

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = repo.CreateModel(context.Background(), button, nil)
	if err != nil {
		t.Fatalf("创建按钮失败: %v", err)
	}

	// 先添加组策略
	err = repo.AddGroupPolicy(context.Background(), button)
	if err != nil {
		t.Fatalf("添加按钮组策略失败: %v", err)
	}

	// 测试删除组策略
	err = repo.RemoveGroupPolicy(context.Background(), button, true)
	if err != nil {
		t.Fatalf("删除按钮组策略失败: %v", err)
	}

	// 测试删除组策略（不删除继承）
	err = repo.AddGroupPolicy(context.Background(), button)
	if err != nil {
		t.Fatalf("添加按钮组策略失败: %v", err)
	}

	err = repo.RemoveGroupPolicy(context.Background(), button, false)
	if err != nil {
		t.Fatalf("删除按钮组策略失败: %v", err)
	}
}

// TestButtonRemoveGroupPolicyWithCanceledContext 测试上下文已取消时删除按钮组策略
func TestButtonRemoveGroupPolicyWithCanceledContext(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)

	// 创建一个已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 创建按钮
	button := &sysmodel.ButtonModel{}
	button.ID = 1

	// 测试删除组策略
	err := repo.RemoveGroupPolicy(ctx, button, true)
	if err == nil {
		t.Fatal("上下文已取消时删除按钮组策略应该返回错误")
	}
}

// TestButtonRemoveGroupPolicyWithNilButton 测试按钮为nil时删除组策略
func TestButtonRemoveGroupPolicyWithNilButton(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)

	// 测试删除组策略
	err := repo.RemoveGroupPolicy(context.Background(), nil, true)
	if err == nil {
		t.Fatal("按钮为nil时删除组策略应该返回错误")
	}
}

// TestButtonRemoveGroupPolicyWithZeroID 测试按钮ID为0时删除组策略
func TestButtonRemoveGroupPolicyWithZeroID(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewButtonRepo(logger, db, dbTimeout, enforcer)

	// 创建按钮
	button := &sysmodel.ButtonModel{}
	button.ID = 0

	// 测试删除组策略
	err := repo.RemoveGroupPolicy(context.Background(), button, true)
	if err == nil {
		t.Fatal("按钮ID为0时删除组策略应该返回错误")
	}
}
