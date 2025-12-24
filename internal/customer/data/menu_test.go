package data

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestMenuModel(pk, pid uint32) *biz.MenuModel {
	m := &biz.MenuModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: pk,
			},
		},
		Path:      fmt.Sprintf("/menu/%d", pk),
		Component: "test_menu",
		Name:      fmt.Sprintf("test_menu_%d", pk),
		Meta: biz.Meta{
			Title: "test",
			Icon:  "test_icon",
		},
		ArrangeOrder: 1000,
		IsActive:     true,
	}
	if pid != 0 {
		m.ParentID = &pid
	}
	return m
}

type MenuTestSuite struct {
	suite.Suite
	permRepo *permissionRepo
	menuRepo *menuRepo
	perms    *[]biz.PermissionModel
}

func (suite *MenuTestSuite) SetupSuite() {
	// 确保menuRepo已初始化
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&biz.PermissionModel{},
		&biz.MenuModel{},
	)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, err := test.NewTestEnforcer(test.TestSecretKey)
	if err != nil {
		panic(err)
	}

	suite.permRepo = &permissionRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}

	suite.menuRepo = &menuRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}

	// 初始化数据
	for i := 30; i < 40; i++ {
		sm := CreateTestPermModel(uint32(i))
		err := suite.permRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建权限应该成功")
	}
	qp := database.QueryParams{
		Limit:   5,
		Offset:  0,
		IsCount: false,
	}
	_, pms, err := suite.permRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询权限列表应该成功")
	suite.NotNil(pms, "查询权限列表应该不为nil")
	suite.Greater(len(*pms), 0, "查询权限列表数量应该大于0")
	suite.perms = pms
}

func (suite *MenuTestSuite) TearDownSuite() {
	test.CloseTestGormDB(suite.menuRepo.gormDB)
}

func (suite *MenuTestSuite) TestCreateMenu() {
	mm := CreateTestMenuModel(51, 0)
	err := suite.menuRepo.CreateModel(context.Background(), mm, nil)
	suite.NoError(err, "创建父菜单应该成功")
	fmm, err := suite.menuRepo.FindModel(context.Background(), []string{"Parent", "Permissions"}, "id = ?", mm.ID)
	suite.NoError(err, "查询父菜单应该成功")
	suite.NotNil(fmm, "查询父菜单应该不为nil")
	suite.Equal(mm.ID, fmm.ID, "查询的菜单ID应该等于创建的菜单ID")

	cm := CreateTestMenuModel(52, 51)
	err = suite.menuRepo.CreateModel(context.Background(), cm, suite.perms)
	suite.NoError(err, "创建子菜单应该成功")
	fcm, err := suite.menuRepo.FindModel(context.Background(), []string{"Parent", "Permissions"}, "id = ?", cm.ID)
	suite.NoError(err, "查询子菜单应该成功")
	suite.NotNil(fcm, "查询子菜单应该不为nil")
	suite.Equal(cm.ID, fcm.ID, "查询的菜单ID应该等于创建的菜单ID")
	suite.NotNil(fcm.ParentID, "查询的子菜单的父ID不应该为nil")
	suite.Equal(mm.ID, *fcm.ParentID, "查询的菜单的父ID应该等于创建的菜单的父ID")
	suite.NotNil(fcm.Parent, "查询的菜单的父菜单应该不为nil")
	suite.Equal(mm.ID, fcm.Parent.ID, "查询的菜单的父菜单ID应该等于创建的菜单的父菜单ID")
}

// func (suite *MenuTestSuite) TestUpdateMenu() {
// }

// func (suite *MenuTestSuite) TestDeleteMenu() {
// }

// func (suite *MenuTestSuite) TestFindMenu() {
// }

// func (suite *MenuTestSuite) TestListMenu() {
// }

// func (suite *MenuTestSuite) TestAddPolicy() {
// }

// func (suite *MenuTestSuite) TestRemovePolicy() {
// }

// 每个测试文件都需要这个入口函数
func TestMenuTestSuite(t *testing.T) {
	mts := &MenuTestSuite{}
	suite.Run(t, mts)
}
