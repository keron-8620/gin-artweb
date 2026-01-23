package data

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/suite"
// 	"gorm.io/gorm"

// 	"gin-artweb/internal/customer/biz"
// 	"gin-artweb/internal/shared/auth"
// 	"gin-artweb/internal/shared/database"
// 	"gin-artweb/internal/shared/test"
// )

// func CreateTestMenuModel(pk uint32) *biz.MenuModel {
// 	return &biz.MenuModel{
// 		StandardModel: database.StandardModel{
// 			BaseModel: database.BaseModel{
// 				ID: pk,
// 			},
// 		},
// 		Path:         fmt.Sprintf("/menu/%d", pk),
// 		Component:    fmt.Sprintf("Component%d", pk),
// 		Name:         fmt.Sprintf("menu_%d", pk),
// 		Meta:         biz.Meta{Title: fmt.Sprintf("Menu %d", pk), Icon: "icon"},
// 		ArrangeOrder: pk,
// 		IsActive:     true,
// 		Descr:        fmt.Sprintf("这是一个测试菜单_%d", pk),
// 	}
// }

// type MenuTestSuite struct {
// 	suite.Suite
// 	menuRepo *menuRepo
// }

// func (suite *MenuTestSuite) SetupSuite() {
// 	db := test.NewTestGormDBWithConfig(nil)
// 	db.AutoMigrate(&biz.MenuModel{})
// 	dbTimeout := test.NewTestDBTimeouts()
// 	logger := test.NewTestZapLogger()
// 	enforcer, err := auth.NewCasbinEnforcer()
// 	if err != nil {
// 		panic(err)
// 	}
// 	suite.menuRepo = &menuRepo{
// 		log:      logger,
// 		gormDB:   db,
// 		timeouts: dbTimeout,
// 		enforcer: enforcer,
// 	}
// }

// func (suite *MenuTestSuite) TearDownSuite() {
// 	test.CloseTestGormDB(suite.menuRepo.gormDB)
// }

// func (suite *MenuTestSuite) TestCreateMenu() {
// 	sm := CreateTestMenuModel(1)
// 	emptyPerms := []biz.PermissionModel{}
// 	err := suite.menuRepo.CreateModel(context.Background(), sm, &emptyPerms)
// 	suite.NoError(err, "创建菜单应该成功")

// 	fm, err := suite.menuRepo.FindModel(context.Background(), []string{}, "id = ?", sm.ID)
// 	suite.NoError(err, "查询刚创建的菜单应该成功")
// 	suite.Equal(sm.ID, fm.ID)
// 	suite.Equal(sm.Path, fm.Path)
// 	suite.Equal(sm.Component, fm.Component)
// 	suite.Equal(sm.Name, fm.Name)
// 	suite.Equal(sm.ArrangeOrder, fm.ArrangeOrder)
// 	suite.Equal(sm.IsActive, fm.IsActive)
// 	suite.Equal(sm.Descr, fm.Descr)
// }

// func (suite *MenuTestSuite) TestUpdateMenu() {
// 	sm := CreateTestMenuModel(2)
// 	emptyPerms := []biz.PermissionModel{}
// 	err := suite.menuRepo.CreateModel(context.Background(), sm, &emptyPerms)
// 	suite.NoError(err, "创建菜单应该成功")

// 	updatedPath := fmt.Sprintf("/updated-menu/%d", sm.ID)
// 	updatedComponent := fmt.Sprintf("UpdatedComponent%d", sm.ID)
// 	updatedName := fmt.Sprintf("updated_menu_%d", sm.ID)
// 	updatedDescr := "这是更新的测试菜单"

// 	err = suite.menuRepo.UpdateModel(context.Background(), map[string]any{
// 		"path":      updatedPath,
// 		"component": updatedComponent,
// 		"name":      updatedName,
// 		"descr":     updatedDescr,
// 	}, nil, "id = ?", sm.ID)
// 	suite.NoError(err, "更新菜单应该成功")

// 	fm, err := suite.menuRepo.FindModel(context.Background(), []string{}, "id = ?", sm.ID)
// 	suite.NoError(err, "查询更新后的菜单应该成功")
// 	suite.Equal(fm.ID, sm.ID)
// 	suite.Equal(updatedPath, fm.Path)
// 	suite.Equal(updatedComponent, fm.Component)
// 	suite.Equal(updatedName, fm.Name)
// 	suite.Equal(updatedDescr, fm.Descr)
// 	suite.Greater(fm.UpdatedAt, sm.UpdatedAt)
// }

// func (suite *MenuTestSuite) TestDeleteMenu() {
// 	sm := CreateTestMenuModel(3)
// 	emptyPerms := []biz.PermissionModel{}
// 	err := suite.menuRepo.CreateModel(context.Background(), sm, &emptyPerms)
// 	suite.NoError(err, "创建菜单应该成功")

// 	fm, err := suite.menuRepo.FindModel(context.Background(), []string{}, "id = ?", sm.ID)
// 	suite.NoError(err, "查询刚创建的菜单应该成功")
// 	suite.Equal(sm.ID, fm.ID)

// 	err = suite.menuRepo.DeleteModel(context.Background(), "id = ?", sm.ID)
// 	suite.NoError(err, "删除菜单应该成功")

// 	_, err = suite.menuRepo.FindModel(context.Background(), []string{}, "id = ?", sm.ID)
// 	if err != nil {
// 		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
// 	} else {
// 		suite.Fail("应该返回错误，但没有返回")
// 	}
// }

// func (suite *MenuTestSuite) TestFindMenuByID() {
// 	// 创建一个菜单用于测试
// 	sm := CreateTestMenuModel(4)
// 	emptyPerms := []biz.PermissionModel{}
// 	err := suite.menuRepo.CreateModel(context.Background(), sm, &emptyPerms)
// 	suite.NoError(err, "创建菜单应该成功")

// 	m, err := suite.menuRepo.FindModel(context.Background(), []string{}, "id = ?", sm.ID)
// 	suite.NoError(err, "查询菜单应该成功")
// 	suite.Equal(sm.Path, m.Path)
// 	suite.Equal(sm.Component, m.Component)
// 	suite.Equal(sm.Name, m.Name)
// 	suite.Equal(sm.Descr, m.Descr)
// }

// func (suite *MenuTestSuite) TestListMenu() {
// 	// 清理可能存在的数据并创建测试数据
// 	for i := 10; i < 20; i++ {
// 		sm := CreateTestMenuModel(uint32(i))
// 		emptyPerms := []biz.PermissionModel{}
// 		err := suite.menuRepo.CreateModel(context.Background(), sm, &emptyPerms)
// 		suite.NoError(err, "创建菜单应该成功")
// 	}

// 	qp := database.QueryParams{
// 		Limit:   10,
// 		Offset:  0,
// 		IsCount: true,
// 	}
// 	total, ms, err := suite.menuRepo.ListModel(context.Background(), qp)
// 	suite.NoError(err, "列出菜单应该成功")
// 	suite.NotNil(ms, "菜单列表不应该为nil")
// 	suite.GreaterOrEqual(total, int64(10), "菜单总数应该至少有10条")

// 	qpPaginated := database.QueryParams{
// 		Limit:   5,
// 		Offset:  0,
// 		IsCount: true,
// 	}
// 	pTotal, pMs, err := suite.menuRepo.ListModel(context.Background(), qpPaginated)
// 	suite.NoError(err, "分页列出菜单应该成功")
// 	suite.NotNil(pMs, "分页菜单列表不应该为nil")
// 	suite.Equal(5, len(*pMs), "分页查询应该返回指定数量的记录")
// 	suite.GreaterOrEqual(pTotal, int64(5), "分页总数应该至少等于limit")
// }

// func (suite *MenuTestSuite) TestAddGroupPolicy() {
// 	// 创建一个菜单用于测试策略添加
// 	sm := CreateTestMenuModel(6)
// 	emptyPerms := []biz.PermissionModel{}
// 	err := suite.menuRepo.CreateModel(context.Background(), sm, &emptyPerms)
// 	suite.NoError(err, "创建菜单应该成功")

// 	err = suite.menuRepo.AddGroupPolicy(context.Background(), sm)
// 	suite.NoError(err, "添加组策略应该成功")
// }

// func (suite *MenuTestSuite) TestRemoveGroupPolicy() {
// 	// 创建一个菜单用于测试策略移除
// 	sm := CreateTestMenuModel(7)
// 	emptyPerms := []biz.PermissionModel{}
// 	err := suite.menuRepo.CreateModel(context.Background(), sm, &emptyPerms)
// 	suite.NoError(err, "创建菜单应该成功")

// 	// 先添加策略
// 	err = suite.menuRepo.AddGroupPolicy(context.Background(), sm)
// 	suite.NoError(err, "添加组策略应该成功")

// 	// 移除策略
// 	err = suite.menuRepo.RemoveGroupPolicy(context.Background(), sm, true)
// 	suite.NoError(err, "移除组策略应该成功")
// }

// func (suite *MenuTestSuite) TestCreateMenuWithInvalidData() {
// 	// 测试创建菜单时传入空数据
// 	err := suite.menuRepo.CreateModel(context.Background(), nil, nil)
// 	suite.Error(err, "创建菜单时传入nil应该返回错误")
// }

// func (suite *MenuTestSuite) TestFindNonExistentMenu() {
// 	// 测试查找不存在的菜单
// 	_, err := suite.menuRepo.FindModel(context.Background(), []string{}, "id = ?", 999999)
// 	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查找不存在的菜单应该返回记录未找到错误")
// }

// // 每个测试文件都需要这个入口函数
// func TestMenuTestSuite(t *testing.T) {
// mts := &MenuTestSuite{}
// 	suite.Run(t, mts)
// }