package biz

import (
	"context"
	"fmt"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/infra/customer/data"
	"gin-artweb/internal/infra/customer/model"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestApiModel() *model.ApiModel {
	return &model.ApiModel{
		URL:    fmt.Sprintf("/api/test/%s/", uuid.NewString()),
		Method: "GET",
		Label:  "test",
		Descr:  "这是一个测试接口",
	}
}

type ApiTestSuite struct {
	suite.Suite
	enforcer *casbin.Enforcer
	uc       *ApiUsecase
}

func (suite *ApiTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&model.ApiModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	suite.enforcer = enforcer
	suite.uc = &ApiUsecase{
		log: logger,
		apiRepo: data.NewApiRepo(
			logger,
			db,
			dbTimeout,
			enforcer,
		),
	}
}

func (suite *ApiTestSuite) TestCreateApi() {
	sm := CreateTestApiModel()
	fm, err := suite.uc.CreateApi(context.Background(), *sm)
	suite.Nil(err, "创建API应该成功")
	suite.Greater(fm.ID, uint32(0), "API ID应该大于0")
	suite.Equal(sm.URL, fm.URL)
	suite.Equal(sm.Method, fm.Method)
	suite.Equal(sm.Label, fm.Label)
	suite.Equal(sm.Descr, fm.Descr)
	sub := auth.ApiToSubject(fm.ID)
	ok, enforceErr := suite.enforcer.Enforce(sub, fm.URL, fm.Method)
	suite.NoError(enforceErr, "Enforce应该成功")
	suite.True(ok, "Enforce应该返回true")
}

func (suite *ApiTestSuite) TestFindApiByID() {
	sm := CreateTestApiModel()
	fm, err := suite.uc.CreateApi(context.Background(), *sm)
	suite.Nil(err, "创建API应该成功")
	suite.Greater(fm.ID, uint32(0), "API ID应该大于0")

	fm, err = suite.uc.FindApiByID(context.Background(), fm.ID)
	suite.Nil(err, "查询刚创建的API应该成功")
	suite.Greater(fm.ID, uint32(0), "API ID应该大于0")
	suite.Equal(sm.URL, fm.URL)
	suite.Equal(sm.Method, fm.Method)
	suite.Equal(sm.Label, fm.Label)
	suite.Equal(sm.Descr, fm.Descr)

}

func (suite *ApiTestSuite) TestFindApiByID_NotFound() {
	_, err := suite.uc.FindApiByID(context.Background(), 0)
	suite.NotNil(err, "查询不存在的API应该失败")
}

func (suite *ApiTestSuite) TestDeleteApi() {
	sm := CreateTestApiModel()
	fm, err := suite.uc.CreateApi(context.Background(), *sm)
	suite.Nil(err, "创建API应该成功")

	err = suite.uc.DeleteApiByID(context.Background(), fm.ID)
	suite.Nil(err, "删除刚创建的API应该成功")

	_, err = suite.uc.FindApiByID(context.Background(), fm.ID)
	suite.NotNil(err, "查询已删除的API应该失败")
}

func (suite *ApiTestSuite) TestDeleteApi_NotFound() {
	err := suite.uc.DeleteApiByID(context.Background(), 0)
	suite.NotNil(err, "删除不存在的API应该失败")
}

func (suite *ApiTestSuite) TestUpdateApiByID() {
	sm := CreateTestApiModel()
	fm, err := suite.uc.CreateApi(context.Background(), *sm)
	suite.Nil(err, "创建API应该成功")

	// 准备更新数据
	updateData := map[string]any{
		"label": "updated_test",
		"descr": "这是一个更新后的测试接口",
	}

	// 执行更新
	updatedFm, err := suite.uc.UpdateApiByID(context.Background(), fm.ID, updateData)
	suite.Nil(err, "更新API应该成功")
	suite.Equal(fm.ID, updatedFm.ID)
	suite.Equal(sm.URL, updatedFm.URL)
	suite.Equal(sm.Method, updatedFm.Method)
	suite.Equal(updateData["label"], updatedFm.Label)
	suite.Equal(updateData["descr"], updatedFm.Descr)

	// 验证权限策略更新
	sub := auth.ApiToSubject(updatedFm.ID)
	ok, enforceErr := suite.enforcer.Enforce(sub, updatedFm.URL, updatedFm.Method)
	suite.NoError(enforceErr, "Enforce应该成功")
	suite.True(ok, "Enforce应该返回true")
}

func (suite *ApiTestSuite) TestUpdateApiByID_NotFound() {
	updateData := map[string]any{
		"label": "updated_test",
	}
	_, err := suite.uc.UpdateApiByID(context.Background(), 0, updateData)
	suite.NotNil(err, "更新不存在的API应该失败")
}

func (suite *ApiTestSuite) TestListApi() {
	// 创建多个API
	apiCount := 3
	for i := 0; i < apiCount; i++ {
		sm := CreateTestApiModel()
		_, err := suite.uc.CreateApi(context.Background(), *sm)
		suite.Nil(err, "创建API应该成功")
	}

	// 测试列出所有API
	qp := database.QueryParams{}
	count, apiList, err := suite.uc.ListApi(context.Background(), qp)
	suite.Nil(err, "列出API应该成功")
	suite.GreaterOrEqual(int(count), apiCount, "返回的API数量应该大于等于创建的数量")
	suite.NotNil(apiList, "返回的API列表不应该为nil")
}

func (suite *ApiTestSuite) TestLoadApiPolicy() {
	// 创建几个API
	apiCount := 2
	createdApis := make([]model.ApiModel, 0, apiCount)
	for i := 0; i < apiCount; i++ {
		sm := CreateTestApiModel()
		fm, err := suite.uc.CreateApi(context.Background(), *sm)
		suite.Nil(err, "创建API应该成功")
		createdApis = append(createdApis, *fm)
	}

	// 加载API策略
	err := suite.uc.LoadApiPolicy(context.Background())
	suite.Nil(err, "加载API策略应该成功")

	// 验证权限策略是否正确加载
	for _, api := range createdApis {
		sub := auth.ApiToSubject(api.ID)
		ok, enforceErr := suite.enforcer.Enforce(sub, api.URL, api.Method)
		suite.NoError(enforceErr, "Enforce应该成功")
		suite.True(ok, "Enforce应该返回true")
	}
}

func (suite *ApiTestSuite) TestCreateApi_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文创建API
	sm := CreateTestApiModel()
	_, err := suite.uc.CreateApi(ctx, *sm)
	suite.NotNil(err, "上下文错误时创建API应该失败")
}

func (suite *ApiTestSuite) TestUpdateApiByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文更新API
	updateData := map[string]any{
		"label": "updated_test",
	}
	_, err := suite.uc.UpdateApiByID(ctx, 1, updateData)
	suite.NotNil(err, "上下文错误时更新API应该失败")
}

func (suite *ApiTestSuite) TestDeleteApiByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文删除API
	err := suite.uc.DeleteApiByID(ctx, 1)
	suite.NotNil(err, "上下文错误时删除API应该失败")
}

func (suite *ApiTestSuite) TestFindApiByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文查找API
	_, err := suite.uc.FindApiByID(ctx, 1)
	suite.NotNil(err, "上下文错误时查找API应该失败")
}

func (suite *ApiTestSuite) TestListApi_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文列出API
	qp := database.QueryParams{}
	_, _, err := suite.uc.ListApi(ctx, qp)
	suite.NotNil(err, "上下文错误时列出API应该失败")
}

func (suite *ApiTestSuite) TestLoadApiPolicy_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文加载API策略
	err := suite.uc.LoadApiPolicy(ctx)
	suite.NotNil(err, "上下文错误时加载API策略应该失败")
}

// 每个测试文件都需要这个入口函数
func TestApiTestSuite(t *testing.T) {
	pts := &ApiTestSuite{}
	suite.Run(t, pts)
}
