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

// CreateTestRoleModel 创建测试用的角色模型
func CreateTestRoleModel() *model.RoleModel {
	return &model.RoleModel{
		Name:  uuid.NewString(),
		Descr: "这是一个测试角色",
	}
}

type RoleTestSuite struct {
	suite.Suite
	apiRepo    *ApiRepo
	menuRepo   *MenuRepo
	buttonRepo *ButtonRepo
	roleRepo   *RoleRepo
}

func (suite *RoleTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&model.ApiModel{},
		&model.MenuModel{},
		&model.ButtonModel{},
		&model.RoleModel{},
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
	suite.roleRepo = &RoleRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}
}

func (suite *RoleTestSuite) TestCreateRole() {
	// 测试创建角色
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 测试查询刚创建的角色
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "查询刚创建的角色应该成功")
	suite.Equal(role.ID, fm.ID)
	suite.Equal(role.Name, fm.Name)
	suite.Equal(role.Descr, fm.Descr)
}

func (suite *RoleTestSuite) TestUpdateRole() {
	// 创建角色
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 测试更新角色
	updatedName := "更新的测试角色"
	updatedDescr := "更新后的角色描述"

	err = suite.roleRepo.UpdateModel(context.Background(), map[string]any{
		"name":  updatedName,
		"descr": updatedDescr,
	}, nil, nil, nil, "id = ?", role.ID)
	suite.NoError(err, "更新角色应该成功")

	// 测试查询更新后的角色
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "查询更新后的角色应该成功")
	suite.Equal(role.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Greater(fm.UpdatedAt, role.UpdatedAt)
}

func (suite *RoleTestSuite) TestDeleteRole() {
	// 创建角色
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 测试查询刚创建的角色
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "查询刚创建的角色应该成功")
	suite.Equal(role.ID, fm.ID)

	// 测试删除角色
	err = suite.roleRepo.DeleteModel(context.Background(), "id = ?", role.ID)
	suite.NoError(err, "删除角色应该成功")

	// 测试查询已删除的角色
	_, err = suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *RoleTestSuite) TestGetRole() {
	// 创建角色
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 测试根据ID查询角色
	m, err := suite.roleRepo.GetModel(context.Background(), []string{}, role.ID)
	suite.NoError(err, "根据ID查询角色应该成功")
	suite.Equal(role.ID, m.ID)
	suite.Equal(role.Name, m.Name)
	suite.Equal(role.Descr, m.Descr)
}

func (suite *RoleTestSuite) TestListRole() {
	// 测试创建多个角色
	for range 5 {
		role := CreateTestRoleModel()
		err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
		suite.NoError(err, "创建角色应该成功")
	}

	// 测试查询角色列表
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	count, ms, err := suite.roleRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询角色列表应该成功")
	suite.NotNil(ms, "角色列表不应该为nil")
	suite.GreaterOrEqual(count, int64(5), "角色总数应该至少有5条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:   2,
		Offset:  0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.roleRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询角色列表应该成功")
	suite.NotNil(pMs, "分页角色列表不应该为nil")
	suite.Equal(2, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(2), "分页总数应该至少等于limit")
}

func (suite *RoleTestSuite) TestAddGroupPolicy() {
	// 创建API用于测试
	api := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")

	// 创建菜单用于测试
	menu := CreateTestMenuModel(nil)
	err = suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮用于测试
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 创建角色
	role := CreateTestRoleModel()
	apis := []model.ApiModel{*api}
	menus := []model.MenuModel{*menu}
	buttons := []model.ButtonModel{*button}
	err = suite.roleRepo.CreateModel(context.Background(), role, &apis, &menus, &buttons)
	suite.NoError(err, "创建角色应该成功")

	// 测试添加权限策略
	err = suite.roleRepo.AddGroupPolicy(context.Background(), role)
	suite.NoError(err, "添加角色权限策略应该成功")
}

func (suite *RoleTestSuite) TestRemoveGroupPolicy() {
	// 创建API用于测试
	api := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")

	// 创建菜单用于测试
	menu := CreateTestMenuModel(nil)
	err = suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建按钮用于测试
	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 创建角色
	role := CreateTestRoleModel()
	apis := []model.ApiModel{*api}
	menus := []model.MenuModel{*menu}
	buttons := []model.ButtonModel{*button}
	err = suite.roleRepo.CreateModel(context.Background(), role, &apis, &menus, &buttons)
	suite.NoError(err, "创建角色应该成功")

	// 添加权限策略
	err = suite.roleRepo.AddGroupPolicy(context.Background(), role)
	suite.NoError(err, "添加角色权限策略应该成功")

	// 测试删除权限策略
	err = suite.roleRepo.RemoveGroupPolicy(context.Background(), role)
	suite.NoError(err, "删除角色权限策略应该成功")
}

func (suite *RoleTestSuite) TestAddGroupPolicyWithNilRole() {
	// 测试添加权限策略时传入nil角色
	err := suite.roleRepo.AddGroupPolicy(context.Background(), nil)
	suite.Error(err, "传入nil角色应该返回错误")
	suite.Contains(err.Error(), "AddGroupPolicy操作失败: 角色模型不能为空")
}

func (suite *RoleTestSuite) TestAddGroupPolicyWithZeroID() {
	// 测试添加权限策略时传入ID为0的角色
	role := &model.RoleModel{
		Name:  "测试角色",
		Descr: "测试角色描述",
	}
	err := suite.roleRepo.AddGroupPolicy(context.Background(), role)
	suite.Error(err, "传入ID为0的角色应该返回错误")
	suite.Contains(err.Error(), "AddGroupPolicy操作失败: 角色ID不能为0")
}

func (suite *RoleTestSuite) TestRemoveGroupPolicyWithNilRole() {
	// 测试删除权限策略时传入nil角色
	err := suite.roleRepo.RemoveGroupPolicy(context.Background(), nil)
	suite.Error(err, "传入nil角色应该返回错误")
	suite.Contains(err.Error(), "RemoveGroupPolicy操作失败: 角色模型不能为空")
}

func (suite *RoleTestSuite) TestRemoveGroupPolicyWithZeroID() {
	// 测试删除权限策略时传入ID为0的角色
	role := &model.RoleModel{
		Name:  "测试角色",
		Descr: "测试角色描述",
	}
	err := suite.roleRepo.RemoveGroupPolicy(context.Background(), role)
	suite.Error(err, "传入ID为0的角色应该返回错误")
	suite.Contains(err.Error(), "RemoveGroupPolicy操作失败: 角色ID不能为0")
}

func (suite *RoleTestSuite) TestCreateRoleWithNilModel() {
	// 测试创建角色时传入nil模型
	err := suite.roleRepo.CreateModel(context.Background(), nil, nil, nil, nil)
	suite.Error(err, "传入nil模型应该返回错误")
	suite.Contains(err.Error(), "创建角色模型失败: 模型为空")
}

func (suite *RoleTestSuite) TestCreateRoleWithEmptyRelations() {
	// 测试创建角色时传入空的关联列表
	role := CreateTestRoleModel()
	emptyApis := []model.ApiModel{}
	emptyMenus := []model.MenuModel{}
	emptyButtons := []model.ButtonModel{}
	err := suite.roleRepo.CreateModel(context.Background(), role, &emptyApis, &emptyMenus, &emptyButtons)
	suite.NoError(err, "传入空的关联列表应该成功创建角色")

	// 验证角色创建成功
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "查询刚创建的角色应该成功")
	suite.Equal(role.ID, fm.ID)
}

func (suite *RoleTestSuite) TestUpdateRoleWithEmptyData() {
	// 创建角色
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 传入空的data映射
	err = suite.roleRepo.UpdateModel(context.Background(), map[string]any{}, nil, nil, nil, "id = ?", role.ID)
	suite.Error(err, "传入空data更新角色应该返回错误")
	suite.Contains(err.Error(), "更新角色模型失败: 更新数据为空")
}

func (suite *RoleTestSuite) TestGetRoleWithEmptyConditions() {
	// 测试查询时传入空条件
	_, err := suite.roleRepo.GetModel(context.Background(), []string{})
	// 当传入空条件时，GetModel方法会尝试获取数据库中的第一条记录
	// 如果数据库为空，会返回record not found错误
	// 如果数据库不为空，会返回第一条记录
	if err != nil {
		// 如果返回错误，应该是record not found
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询时传入空条件应该返回记录未找到错误")
	} else {
		// 如果返回结果，应该是一个有效的角色模型
		// 这里不做断言，因为可能没有数据
	}
}

func (suite *RoleTestSuite) TestGetRoleWithNonExistentID() {
	// 测试查询不存在的角色ID
	_, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", uint32(999999))
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询不存在的角色ID应该返回记录未找到错误")
}

func (suite *RoleTestSuite) TestDeleteRoleWithNonExistentID() {
	// 测试删除不存在的角色ID
	err := suite.roleRepo.DeleteModel(context.Background(), "id = ?", uint32(999999))
	suite.NoError(err, "删除不存在的角色ID应该成功")
}

func (suite *RoleTestSuite) TestUpdateRoleWithNonExistentID() {
	// 测试更新不存在的角色ID
	err := suite.roleRepo.UpdateModel(context.Background(), map[string]any{
		"name": "更新的测试角色",
	}, nil, nil, nil, "id = ?", uint32(999999))
	suite.NoError(err, "更新不存在的角色ID应该成功")
}

func (suite *RoleTestSuite) TestListRoleWithPaginationBoundaries() {
	// 测试Limit=0的情况
	qpZeroLimit := database.QueryParams{
		Limit:   0,
		Offset:  0,
		IsCount: true,
	}
	_, msZero, err := suite.roleRepo.ListModel(context.Background(), qpZeroLimit)
	suite.NoError(err, "Limit=0应该成功查询")
	suite.NotNil(msZero, "角色列表不应该为nil")

	// 测试较大的Offset值
	qpLargeOffset := database.QueryParams{
		Limit:   10,
		Offset:  999999,
		IsCount: true,
	}
	_, msLarge, err := suite.roleRepo.ListModel(context.Background(), qpLargeOffset)
	suite.NoError(err, "较大的Offset值应该成功查询")
	suite.NotNil(msLarge, "角色列表不应该为nil")
	suite.LessOrEqual(len(*msLarge), 10, "返回的记录数应该不超过Limit")
}

func (suite *RoleTestSuite) TestListRoleWithNoRecords() {
	// 测试查询不存在的条件
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
		Query:   map[string]any{"id": uint32(999999)},
	}
	count, ms, err := suite.roleRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询无记录的角色列表应该成功")
	suite.NotNil(ms, "角色列表不应该为nil")
	suite.Equal(int64(0), count, "无记录时计数应该为0")
	suite.Len(*ms, 0, "无记录时角色列表长度应该为0")
}

func (suite *RoleTestSuite) TestContextTimeout() {
	// 创建一个会立即超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// 等待上下文超时
	time.Sleep(10 * time.Nanosecond)

	// 创建角色用于测试
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 测试CreateModel方法
	sm := CreateTestRoleModel()
	err = suite.roleRepo.CreateModel(ctx, sm, nil, nil, nil)
	suite.Error(err, "上下文超时后创建角色应该返回错误")

	// 测试UpdateModel方法
	err = suite.roleRepo.UpdateModel(ctx, map[string]any{
		"name": "测试角色",
	}, nil, nil, nil, "id = ?", role.ID)
	suite.Error(err, "上下文超时后更新角色应该返回错误")

	// 测试DeleteModel方法
	err = suite.roleRepo.DeleteModel(ctx, "id = ?", role.ID)
	suite.Error(err, "上下文超时后删除角色应该返回错误")

	// 测试GetModel方法
	_, err = suite.roleRepo.GetModel(ctx, []string{}, "id = ?", role.ID)
	suite.Error(err, "上下文超时后获取角色应该返回错误")

	// 测试ListModel方法
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	_, _, err = suite.roleRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文超时后列出角色应该返回错误")

	// 测试AddGroupPolicy方法
	err = suite.roleRepo.AddGroupPolicy(ctx, role)
	suite.Error(err, "上下文超时后添加权限策略应该返回错误")

	// 测试RemoveGroupPolicy方法
	err = suite.roleRepo.RemoveGroupPolicy(ctx, role)
	suite.Error(err, "上下文超时后删除权限策略应该返回错误")
}

func (suite *RoleTestSuite) TestContextCancel() {
	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	// 立即取消上下文
	cancel()

	// 创建角色用于测试
	role := CreateTestRoleModel()
	err := suite.roleRepo.CreateModel(context.Background(), role, nil, nil, nil)
	suite.NoError(err, "创建角色应该成功")

	// 测试CreateModel方法
	sm := CreateTestRoleModel()
	err = suite.roleRepo.CreateModel(ctx, sm, nil, nil, nil)
	suite.Error(err, "上下文取消后创建角色应该返回错误")

	// 测试UpdateModel方法
	err = suite.roleRepo.UpdateModel(ctx, map[string]any{
		"name": "测试角色",
	}, nil, nil, nil, "id = ?", role.ID)
	suite.Error(err, "上下文取消后更新角色应该返回错误")

	// 测试DeleteModel方法
	err = suite.roleRepo.DeleteModel(ctx, "id = ?", role.ID)
	suite.Error(err, "上下文取消后删除角色应该返回错误")

	// 测试GetModel方法
	_, err = suite.roleRepo.GetModel(ctx, []string{}, "id = ?", role.ID)
	suite.Error(err, "上下文取消后获取角色应该返回错误")

	// 测试ListModel方法
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	_, _, err = suite.roleRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文取消后列出角色应该返回错误")

	// 测试AddGroupPolicy方法
	err = suite.roleRepo.AddGroupPolicy(ctx, role)
	suite.Error(err, "上下文取消后添加权限策略应该返回错误")

	// 测试RemoveGroupPolicy方法
	err = suite.roleRepo.RemoveGroupPolicy(ctx, role)
	suite.Error(err, "上下文取消后删除权限策略应该返回错误")
}

func (suite *RoleTestSuite) TestCreateRoleWithApis() {
	// 测试创建与API关联的角色
	api := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")

	// 创建角色并关联API
	role := CreateTestRoleModel()
	apis := []model.ApiModel{*api}
	err = suite.roleRepo.CreateModel(context.Background(), role, &apis, nil, nil)
	suite.NoError(err, "创建与API关联的角色应该成功")

	// 验证角色创建成功
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "查询角色应该成功")
	suite.Equal(role.ID, fm.ID)
}

func (suite *RoleTestSuite) TestCreateRoleWithMenus() {
	// 测试创建与Menu关联的角色
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	// 创建角色并关联Menu
	role := CreateTestRoleModel()
	menus := []model.MenuModel{*menu}
	err = suite.roleRepo.CreateModel(context.Background(), role, nil, &menus, nil)
	suite.NoError(err, "创建与Menu关联的角色应该成功")

	// 验证角色创建成功
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "查询角色应该成功")
	suite.Equal(role.ID, fm.ID)
}

func (suite *RoleTestSuite) TestCreateRoleWithButtons() {
	// 测试创建与Button关联的角色
	menu := CreateTestMenuModel(nil)
	err := suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 创建角色并关联Button
	role := CreateTestRoleModel()
	buttons := []model.ButtonModel{*button}
	err = suite.roleRepo.CreateModel(context.Background(), role, nil, nil, &buttons)
	suite.NoError(err, "创建与Button关联的角色应该成功")

	// 验证角色创建成功
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "查询角色应该成功")
	suite.Equal(role.ID, fm.ID)
}

func (suite *RoleTestSuite) TestCreateRoleWithAllRelations() {
	// 测试创建与所有类型关联的角色
	api := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")

	menu := CreateTestMenuModel(nil)
	err = suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 创建角色并关联所有类型
	role := CreateTestRoleModel()
	apis := []model.ApiModel{*api}
	menus := []model.MenuModel{*menu}
	buttons := []model.ButtonModel{*button}
	err = suite.roleRepo.CreateModel(context.Background(), role, &apis, &menus, &buttons)
	suite.NoError(err, "创建与所有类型关联的角色应该成功")

	// 验证角色创建成功
	fm, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "查询角色应该成功")
	suite.Equal(role.ID, fm.ID)
}

func (suite *RoleTestSuite) TestPreloadRelations() {
	// 测试预加载关联数据
	api := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), api)
	suite.NoError(err, "创建API应该成功")

	menu := CreateTestMenuModel(nil)
	err = suite.menuRepo.CreateModel(context.Background(), menu, nil)
	suite.NoError(err, "创建菜单应该成功")

	button := CreateTestButtonModel(menu.ID)
	err = suite.buttonRepo.CreateModel(context.Background(), button, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 创建角色并关联所有类型
	role := CreateTestRoleModel()
	apis := []model.ApiModel{*api}
	menus := []model.MenuModel{*menu}
	buttons := []model.ButtonModel{*button}
	err = suite.roleRepo.CreateModel(context.Background(), role, &apis, &menus, &buttons)
	suite.NoError(err, "创建与所有类型关联的角色应该成功")

	// 不预加载关联数据查询角色
	roleWithoutRelations, err := suite.roleRepo.GetModel(context.Background(), []string{}, "id = ?", role.ID)
	suite.NoError(err, "不预加载关联数据查询角色应该成功")

	// 预加载关联数据查询角色
	roleWithRelations, err := suite.roleRepo.GetModel(context.Background(), []string{"Apis", "Menus", "Buttons"}, "id = ?", role.ID)
	suite.NoError(err, "预加载关联数据查询角色应该成功")
	suite.Equal(roleWithoutRelations.ID, roleWithRelations.ID)
	suite.NotEmpty(roleWithRelations.Apis, "预加载后应该包含API")
	suite.NotEmpty(roleWithRelations.Menus, "预加载后应该包含Menu")
	suite.NotEmpty(roleWithRelations.Buttons, "预加载后应该包含Button")
}

// 每个测试文件都需要这个入口函数
func TestRoleTestSuite(t *testing.T) {
	pts := &RoleTestSuite{}
	suite.Run(t, pts)
}
