package data

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/customer/biz"
	"gin-artweb/internal/shared/database"
)

// CreateTestButtonModel 创建测试用的按钮模型
func CreateTestButtonModel(pk uint32, menuID uint32) *biz.ButtonModel {
	return &biz.ButtonModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: pk,
			},
		},
		MenuID:       menuID,
		Name:         "test_button_" + string(rune('0'+pk)),
		ArrangeOrder: pk,
		IsActive:     true,
		Descr:        "测试按钮" + string(rune('0'+pk)),
	}
}

type ButtonTestSuite struct {
	suite.Suite
	buttonRepo *buttonRepo
}

func (suite *ButtonTestSuite) SetupSuite() {
	suite.buttonRepo = NewTestButtonRepo()
}

func (suite *ButtonTestSuite) TestCreateButton() {
	// 测试创建按钮
	sm := CreateTestButtonModel(1, 1)
	perms := []biz.PermissionModel{}
	err := suite.buttonRepo.CreateModel(context.Background(), sm, &perms)
	suite.NoError(err, "创建按钮应该成功")

	// 测试查询刚创建的按钮
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的按钮应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(sm.MenuID, fm.MenuID)
	suite.Equal(sm.Name, fm.Name)
	suite.Equal(sm.ArrangeOrder, fm.ArrangeOrder)
	suite.Equal(sm.IsActive, fm.IsActive)
	suite.Equal(sm.Descr, fm.Descr)
}

func (suite *ButtonTestSuite) TestUpdateButton() {
	// 测试创建按钮
	sm := CreateTestButtonModel(2, 1)
	err := suite.buttonRepo.CreateModel(context.Background(), sm, nil)
	suite.NoError(err, "创建按钮应该成功")

	// 测试更新按钮
	updatedName := "更新的测试按钮"
	updatedArrangeOrder := uint32(10)
	updatedIsActive := false
	updatedDescr := "更新的按钮描述"

	err = suite.buttonRepo.UpdateModel(context.Background(), map[string]any{
		"name":          updatedName,
		"arrange_order": updatedArrangeOrder,
		"is_active":     updatedIsActive,
		"descr":         updatedDescr,
	}, nil, "id = ?", sm.ID)
	suite.NoError(err, "更新按钮应该成功")

	// 测试查询更新后的按钮
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的按钮应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedArrangeOrder, fm.ArrangeOrder)
	suite.Equal(updatedIsActive, fm.IsActive)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Greater(fm.UpdatedAt, sm.UpdatedAt)
}

func (suite *ButtonTestSuite) TestDeleteButton() {
	// 测试创建按钮
	sm := CreateTestButtonModel(3, 1)
	perms := []biz.PermissionModel{}
	err := suite.buttonRepo.CreateModel(context.Background(), sm, &perms)
	suite.NoError(err, "创建按钮应该成功")

	// 测试查询刚创建的按钮
	fm, err := suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的按钮应该成功")
	suite.Equal(sm.ID, fm.ID)

	// 测试删除按钮
	err = suite.buttonRepo.DeleteModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "删除按钮应该成功")

	// 测试查询已删除的按钮
	_, err = suite.buttonRepo.GetModel(context.Background(), []string{}, "id = ?", sm.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *ButtonTestSuite) TestGetButtonByID() {
	// 测试创建按钮
	sm := CreateTestButtonModel(4, 1)
	perms := []biz.PermissionModel{}
	err := suite.buttonRepo.CreateModel(context.Background(), sm, &perms)
	suite.NoError(err, "创建按钮应该成功")

	// 测试根据ID查询按钮
	m, err := suite.buttonRepo.GetModel(context.Background(), []string{}, sm.ID)
	suite.NoError(err, "根据ID查询按钮应该成功")
	suite.Equal(sm.ID, m.ID)
	suite.Equal(sm.MenuID, m.MenuID)
	suite.Equal(sm.Name, m.Name)
	suite.Equal(sm.ArrangeOrder, m.ArrangeOrder)
	suite.Equal(sm.IsActive, m.IsActive)
	suite.Equal(sm.Descr, m.Descr)
}

func (suite *ButtonTestSuite) TestListButtons() {
	// 测试创建多个按钮
	for i := 5; i < 10; i++ {
		sm := CreateTestButtonModel(uint32(i), 1)
		perms := []biz.PermissionModel{}
		err := suite.buttonRepo.CreateModel(context.Background(), sm, &perms)
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

func (suite *ButtonTestSuite) TestCreateButtonWithInvalidData() {
	// 测试创建按钮时传入空数据
	err := suite.buttonRepo.CreateModel(context.Background(), nil, nil)
	suite.Error(err, "创建按钮时传入nil应该返回错误")
}

// 每个测试文件都需要这个入口函数
func TestButtonTestSuite(t *testing.T) {
	pts := &ButtonTestSuite{}
	suite.Run(t, pts)
}
