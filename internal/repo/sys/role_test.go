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

func CreateTestRoleModel() *sysmodel.RoleModel {
	return &sysmodel.RoleModel{
		Name:  fmt.Sprintf("test_role_%s", uuid.NewString()),
		Descr: "这是一个测试角色",
	}
}

type RoleTestSuite struct {
	suite.Suite
	roleRepo   *RoleRepo
	apiRepo    *ApiRepo
	menuRepo   *MenuRepo
	buttonRepo *ButtonRepo
}

func (suite *RoleTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&sysmodel.RoleModel{},
		&sysmodel.ApiModel{},
		&sysmodel.MenuModel{},
		&sysmodel.ButtonModel{},
	)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	suite.roleRepo = &RoleRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}
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

func (suite *RoleTestSuite) TestCreateRole() {
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, role.ID)
	suite.NoError(err, "查询刚创建的角色应该成功")
	suite.Equal(role.ID, fm.ID)
	suite.Equal(role.Name, fm.Name)
	suite.Equal(role.Descr, fm.Descr)
}

func (suite *RoleTestSuite) TestUpdateRole() {
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	updatedName := "updated_role"
	updatedDescr := "这是更新的测试角色描述"

	err = suite.roleRepo.UpdateModel(context.Background(), map[string]any{
		"name":  updatedName,
		"descr": updatedDescr,
	}, nil, nil, nil, "id = ?", role.ID)
	suite.NoError(err, "更新角色应该成功")

	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "查询更新后的角色应该成功")
	suite.Equal(role.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Greater(fm.UpdatedAt, role.UpdatedAt)
}

func (suite *RoleTestSuite) TestDeleteRole() {
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "查询刚创建的角色应该成功")
	suite.Equal(role.ID, fm.ID)

	err = suite.roleRepo.DeleteModel(context.Background(), "id = ?", role.ID)
	suite.NoError(err, "删除角色应该成功")

	_, err = suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *RoleTestSuite) TestListRoles() {
	// 测试创建多个角色
	for range 5 {
		role := CreateTestRoleModel()
		err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
		suite.NoError(err, "创建角色应该成功")
	}

	// 测试CountModel
	count, err := suite.roleRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取角色总数应该成功")
	suite.GreaterOrEqual(count, int64(5), "角色总数应该至少有5条")

	// 测试查询角色列表
	qp := database.QueryParams{
		Limit:  10,
		Offset: 0,
	}
	ms, err := suite.roleRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询角色列表应该成功")
	suite.NotNil(ms, "角色列表不应该为nil")
	suite.GreaterOrEqual(len(ms), 5, "角色列表应该至少有5条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:  2,
		Offset: 0,
	}
	pMs, err := suite.roleRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询角色列表应该成功")
	suite.NotNil(pMs, "分页角色列表不应该为nil")
	suite.Equal(2, len(pMs), "分页查询应该返回指定数量的记录")
}

func (suite *RoleTestSuite) TestAddGroupPolicy() {
	// 创建API用于测试
	api := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")
	err = suite.apiRepo.AddPolicy(context.Background(), *api)
	suite.NoError(err, "添加API策略应该成功")

	// 创建菜单用于测试
	menu := CreateTestMenuModel(nil)
	err = suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")
	err = suite.menuRepo.AddGroupPolicy(context.Background(), menu)
	suite.NoError(err, "添加菜单策略应该成功")

	// 创建按钮用于测试
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")
	err = suite.buttonRepo.AddGroupPolicy(context.Background(), button)
	suite.NoError(err, "添加按钮策略应该成功")

	// 创建角色
	role := CreateTestRoleModel()
	apis := []sysmodel.ApiModel{*api}
	menus := []sysmodel.MenuModel{*menu}
	buttons := []sysmodel.ButtonModel{*button}

	err = suite.roleRepo.CreateModel(context.Background(), role, apis, menus, buttons)
	suite.NoError(err, "创建角色并添加策略应该成功")
	err = suite.roleRepo.AddGroupPolicy(context.Background(), role)
	suite.NoError(err, "添加角色策略应该成功")

	// 验证策略是否添加成功
	sub := auth.RoleToSubject(role.ID)
	ok, err := suite.roleRepo.enforcer.Enforce(sub, api.URL, api.Method)
	suite.NoError(err, "检查授权应该成功")
	suite.True(ok, "添加策略后应该有API权限")
}

func (suite *RoleTestSuite) TestCreateRoleWithInvalidData() {
	// 测试创建角色时传入空数据
	err := suite.roleRepo.CreateModel(context.Background(), nil, nil, nil, nil)
	suite.Error(err, "创建角色时传入nil应该返回错误")
}

func (suite *RoleTestSuite) TestFindNonExistentRole() {
	// 测试查找不存在的角色
	_, err := suite.roleRepo.GetModel(context.Background(), []string{}, 999999)
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查找不存在的角色应该返回记录未找到错误")
}

func (suite *RoleTestSuite) TestUpdateRoleWithEmptyData() {
	// 测试更新时传入空数据
	err := suite.roleRepo.UpdateModel(context.Background(), map[string]any{}, nil, nil, nil, "id = ?", 1)
	suite.Error(err, "更新角色时传入空数据应该返回错误")
}

func (suite *RoleTestSuite) TestUpdateNonExistentRole() {
	// 测试更新不存在的角色
	err := suite.roleRepo.UpdateModel(context.Background(), map[string]any{
		"name": "updated_role",
	}, nil, nil, nil, "id = ?", 999999)
	suite.NoError(err, "更新不存在的角色不应该返回错误")
}

func (suite *RoleTestSuite) TestDeleteRoleWithEmptyConditions() {
	// 测试删除时传入空条件
	err := suite.roleRepo.DeleteModel(context.Background())
	suite.Error(err, "删除时传入空条件应该返回错误")
}

func (suite *RoleTestSuite) TestDeleteNonExistentRole() {
	// 测试删除不存在的角色
	err := suite.roleRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的角色不应该返回错误")
}

func (suite *RoleTestSuite) TestGetRoleWithEmptyConditions() {
	// 测试查询时传入空条件
	result, err := suite.roleRepo.GetModel(context.Background(), []string{})
	// 当传入空条件时，GetModel方法会尝试获取数据库中的第一条记录
	// 如果数据库为空，会返回record not found错误
	// 如果数据库不为空，会返回第一条记录
	if err != nil {
		// 如果返回错误，应该是record not found
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询时传入空条件应该返回记录未找到错误")
	} else {
		// 如果返回结果，应该是一个有效的角色模型
		suite.NotNil(result, "查询时传入空条件应该返回有效的角色模型")
		suite.Greater(result.ID, uint32(0), "返回的角色模型ID应该大于0")
	}
}

func (suite *RoleTestSuite) TestGetRoleWithContextTimeout() {
	// 测试查询时上下文已取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := suite.roleRepo.GetModel(ctx, []string{}, 1)
	suite.Error(err, "查询时上下文已取消应该返回错误")
}

func (suite *RoleTestSuite) TestCreateRoleWithContextTimeout() {
	// 测试创建角色时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(ctx, role, nil, nil, nil)
	suite.Error(err, "创建角色时上下文超时应该返回错误")
}

func (suite *RoleTestSuite) TestUpdateRoleWithContextTimeout() {
	// 测试更新角色时上下文超时
	// 先创建一个角色
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试更新角色
	err = suite.roleRepo.UpdateModel(ctx, map[string]any{
		"name": "updated_role",
	}, nil, nil, nil, "id = ?", role.ID)
	suite.Error(err, "更新角色时上下文超时应该返回错误")
}

func (suite *RoleTestSuite) TestDeleteRoleWithContextTimeout() {
	// 测试删除角色时上下文超时
	// 先创建一个角色
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试删除角色
	err = suite.roleRepo.DeleteModel(ctx, "id = ?", role.ID)
	suite.Error(err, "删除角色时上下文超时应该返回错误")
}

func (suite *RoleTestSuite) TestListRolesWithContextTimeout() {
	// 测试列表查询时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试列表查询
	qp := database.QueryParams{}
	_, err := suite.roleRepo.ListModel(ctx, qp)
	suite.Error(err, "列表查询时上下文超时应该返回错误")
}

func (suite *RoleTestSuite) TestCountRolesWithContextTimeout() {
	// 测试计数查询时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试计数查询
	_, err := suite.roleRepo.CountModel(ctx, nil)
	suite.Error(err, "计数查询时上下文超时应该返回错误")
}

func (suite *RoleTestSuite) TestListRolesWithEmptyParams() {
	// 测试列表查询时传入空参数
	qp := database.QueryParams{}
	ms, err := suite.roleRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列表查询时传入空参数应该成功")
	suite.NotNil(ms, "角色列表不应该为nil")

	// 测试CountModel
	count, err := suite.roleRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取角色总数应该成功")
	suite.GreaterOrEqual(count, int64(0), "角色总数应该大于等于0")
}

func (suite *RoleTestSuite) TestListRolesWithSorting() {
	// 测试列表查询时传入排序参数
	// 先创建多个角色用于测试
	for i := 0; i < 5; i++ {
		role := CreateTestRoleModel()
		err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
		suite.NoError(err, "创建角色应该成功")
	}

	// 测试按ID降序排序
	qp := database.QueryParams{
		OrderBy: []string{"id DESC"},
	}
	ms, err := suite.roleRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按ID降序排序查询应该成功")
	suite.NotNil(ms, "角色列表不应该为nil")
	if len(ms) > 1 {
		// 验证排序结果
		prevID := ms[0].ID
		for _, role := range ms {
			suite.LessOrEqual(role.ID, prevID, "角色应该按ID降序排序")
			prevID = role.ID
		}
	}
}

func (suite *RoleTestSuite) TestListRolesWithFiltering() {
	// 测试列表查询时传入过滤参数
	// 创建一个特定名称的角色
	testName := "filter_test_role"
	role := CreateTestRoleModel()
	role.Name = testName
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 测试按名称过滤
	qp := database.QueryParams{
		Query: map[string]any{
			"name": testName,
		},
	}
	ms, err := suite.roleRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按名称过滤查询应该成功")
	suite.NotNil(ms, "角色列表不应该为nil")
	// 验证过滤结果
	for _, role := range ms {
		suite.Equal(testName, role.Name, "角色应该按名称过滤")
	}

	// 测试CountModel带过滤条件
	count, err := suite.roleRepo.CountModel(context.Background(), map[string]any{
		"name": testName,
	})
	suite.NoError(err, "带过滤条件的角色总数查询应该成功")
	suite.GreaterOrEqual(count, int64(1), "带过滤条件的角色总数应该至少为1")
}

// 每个测试文件都需要这个入口函数
func TestRoleTestSuite(t *testing.T) {
	pts := &RoleTestSuite{}
	suite.Run(t, pts)
}

// TestNewRoleRepo 测试创建角色仓库实例
func TestNewRoleRepo(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()

	repo := NewRoleRepo(logger, db, dbTimeout, enforcer)
	if repo == nil {
		t.Fatal("NewRoleRepo should return a non-nil repository")
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

// TestRoleAddGroupPolicy 测试添加角色组策略
func TestRoleAddGroupPolicy(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&sysmodel.RoleModel{}, &sysmodel.ApiModel{}, &sysmodel.MenuModel{}, &sysmodel.ButtonModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewRoleRepo(logger, db, dbTimeout, enforcer)
	apiRepo := NewApiRepo(logger, db, dbTimeout, enforcer)
	menuRepo := NewMenuRepo(logger, db, dbTimeout, enforcer)
	buttonRepo := NewButtonRepo(logger, db, dbTimeout, enforcer)

	// 创建API
	api := CreateTestApiModel()
	err := apiRepo.CreateModel(context.Background(), api)
	if err != nil {
		t.Fatalf("创建API失败: %v", err)
	}

	// 创建菜单
	menu := CreateTestMenuModel(nil)
	err = menuRepo.CreateModel(context.Background(), menu, nil)
	if err != nil {
		t.Fatalf("创建菜单失败: %v", err)
	}

	// 创建按钮
	button := CreateTestButtonModel(menu.ID)
	err = buttonRepo.CreateModel(context.Background(), button, nil)
	if err != nil {
		t.Fatalf("创建按钮失败: %v", err)
	}

	// 创建角色并关联API、菜单和按钮
	role := CreateTestRoleModel()
	role.Apis = []sysmodel.ApiModel{*api}
	role.Menus = []sysmodel.MenuModel{*menu}
	role.Buttons = []sysmodel.ButtonModel{*button}
	err = repo.CreateModel(context.Background(), role, role.Apis, role.Menus, role.Buttons)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 测试添加组策略
	err = repo.AddGroupPolicy(context.Background(), role)
	if err != nil {
		t.Fatalf("添加角色组策略失败: %v", err)
	}
}

// TestRoleAddGroupPolicyWithInvalidData 测试添加包含无效数据的角色组策略
func TestRoleAddGroupPolicyWithInvalidData(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&sysmodel.RoleModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewRoleRepo(logger, db, dbTimeout, enforcer)

	// 创建角色
	role := CreateTestRoleModel()
	err := repo.CreateModel(context.Background(), role, nil, nil, nil)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 手动设置无效数据
	role.Apis = []sysmodel.ApiModel{{URL: "/api/test", Method: "GET"}} // ID为0
	role.Menus = []sysmodel.MenuModel{{Name: "test_menu", Path: "/test"}} // ID为0
	role.Buttons = []sysmodel.ButtonModel{{Name: "test_button", MenuID: 1}} // ID为0

	// 测试添加组策略（应该跳过无效数据）
	err = repo.AddGroupPolicy(context.Background(), role)
	if err != nil {
		t.Fatalf("添加包含无效数据的角色组策略失败: %v", err)
	}
}

// TestRoleRemoveGroupPolicy 测试删除角色组策略
func TestRoleRemoveGroupPolicy(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&sysmodel.RoleModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewRoleRepo(logger, db, dbTimeout, enforcer)

	// 创建角色
	role := CreateTestRoleModel()
	err := repo.CreateModel(context.Background(), role, nil, nil, nil)
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}

	// 测试删除组策略
	err = repo.RemoveGroupPolicy(context.Background(), role)
	if err != nil {
		t.Fatalf("删除角色组策略失败: %v", err)
	}
}

// TestRoleRemoveGroupPolicyWithCanceledContext 测试上下文已取消时删除角色组策略
func TestRoleRemoveGroupPolicyWithCanceledContext(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewRoleRepo(logger, db, dbTimeout, enforcer)

	// 创建一个已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 创建角色
	role := &sysmodel.RoleModel{}
	role.ID = 1

	// 测试删除组策略
	err := repo.RemoveGroupPolicy(ctx, role)
	if err == nil {
		t.Fatal("上下文已取消时删除角色组策略应该返回错误")
	}
}

// TestRoleRemoveGroupPolicyWithNilRole 测试角色为nil时删除组策略
func TestRoleRemoveGroupPolicyWithNilRole(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewRoleRepo(logger, db, dbTimeout, enforcer)

	// 测试删除组策略
	err := repo.RemoveGroupPolicy(context.Background(), nil)
	if err == nil {
		t.Fatal("角色为nil时删除组策略应该返回错误")
	}
}

// TestRoleRemoveGroupPolicyWithZeroID 测试角色ID为0时删除组策略
func TestRoleRemoveGroupPolicyWithZeroID(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	repo := NewRoleRepo(logger, db, dbTimeout, enforcer)

	// 创建角色
	role := &sysmodel.RoleModel{}
	role.ID = 0

	// 测试删除组策略
	err := repo.RemoveGroupPolicy(context.Background(), role)
	if err == nil {
		t.Fatal("角色ID为0时删除组策略应该返回错误")
	}
}
