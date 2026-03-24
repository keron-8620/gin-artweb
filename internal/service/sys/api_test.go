package sys

import (
	"context"
	"fmt"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	sysmodel "gin-artweb/internal/model/sys"
	syssvc "gin-artweb/internal/repo/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/test"
)

func CreateTestApiDTO() sysmodel.CreateApiDTO {
	return sysmodel.CreateApiDTO{
		ID:     uint32(uuid.New().ID()),
		URL:    fmt.Sprintf("/api/test/%s/", uuid.NewString()),
		Method: "GET",
		Label:  "test",
		Descr:  "这是一个测试接口",
	}
}

func CreateTestApiModel() *sysmodel.ApiModel {
	return &sysmodel.ApiModel{
		URL:    fmt.Sprintf("/api/test/%s/", uuid.NewString()),
		Method: "GET",
		Label:  "test",
		Descr:  "这是一个测试接口",
	}
}

type ApiTestSuite struct {
	suite.Suite
	enforcer   *casbin.Enforcer
	apiservice *ApiService
}

func (suite *ApiTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&sysmodel.ApiModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	suite.enforcer = enforcer
	suite.apiservice = NewApiService(
		logger,
		syssvc.NewApiRepo(
			logger,
			db,
			dbTimeout,
			enforcer,
		),
	)
}

func (suite *ApiTestSuite) TestCreateApi() {
	dto := CreateTestApiDTO()
	fm, err := suite.apiservice.CreateApi(context.Background(), dto)
	suite.Nil(err, "创建API应该成功")
	suite.Greater(fm.ID, uint32(0), "API ID应该大于0")
	suite.Equal(dto.URL, fm.URL)
	suite.Equal(dto.Method, fm.Method)
	suite.Equal(dto.Label, fm.Label)
	suite.Equal(dto.Descr, fm.Descr)
	sub := auth.ApiToSubject(fm.ID)
	ok, enforceErr := suite.enforcer.Enforce(sub, fm.URL, fm.Method)
	suite.NoError(enforceErr, "Enforce应该成功")
	suite.True(ok, "Enforce应该返回true")
}

func (suite *ApiTestSuite) TestFindApiByID() {
	dto := CreateTestApiDTO()
	fm, err := suite.apiservice.CreateApi(context.Background(), dto)
	suite.Nil(err, "创建API应该成功")
	suite.Greater(fm.ID, uint32(0), "API ID应该大于0")

	fm, err = suite.apiservice.FindApiByID(context.Background(), fm.ID)
	suite.Nil(err, "查询刚创建的API应该成功")
	suite.Greater(fm.ID, uint32(0), "API ID应该大于0")
	suite.Equal(dto.URL, fm.URL)
	suite.Equal(dto.Method, fm.Method)
	suite.Equal(dto.Label, fm.Label)
	suite.Equal(dto.Descr, fm.Descr)

}

func (suite *ApiTestSuite) TestFindApiByID_NotFound() {
	_, err := suite.apiservice.FindApiByID(context.Background(), 0)
	suite.NotNil(err, "查询不存在的API应该失败")
}

func (suite *ApiTestSuite) TestDeleteApi() {
	dto := CreateTestApiDTO()
	fm, err := suite.apiservice.CreateApi(context.Background(), dto)
	suite.Nil(err, "创建API应该成功")

	err = suite.apiservice.DeleteApiByID(context.Background(), fm.ID)
	suite.Nil(err, "删除刚创建的API应该成功")

	_, err = suite.apiservice.FindApiByID(context.Background(), fm.ID)
	suite.NotNil(err, "查询已删除的API应该失败")
}

func (suite *ApiTestSuite) TestDeleteApi_NotFound() {
	err := suite.apiservice.DeleteApiByID(context.Background(), 0)
	suite.NotNil(err, "删除不存在的API应该失败")
}

func (suite *ApiTestSuite) TestUpdateApiByID() {
	dto := CreateTestApiDTO()
	fm, err := suite.apiservice.CreateApi(context.Background(), dto)
	suite.Nil(err, "创建API应该成功")

	// 准备更新数据
	updateDTO := sysmodel.UpdateApiDTO{
		URL:    dto.URL,
		Method: dto.Method,
		Label:  "updated_test",
		Descr:  "这是一个更新后的测试接口",
	}

	// 执行更新
	updatedFm, err := suite.apiservice.UpdateApiByID(context.Background(), fm.ID, updateDTO)
	suite.Nil(err, "更新API应该成功")
	suite.Equal(fm.ID, updatedFm.ID)
	suite.Equal(dto.URL, updatedFm.URL)
	suite.Equal(dto.Method, updatedFm.Method)
	suite.Equal(updateDTO.Label, updatedFm.Label)
	suite.Equal(updateDTO.Descr, updatedFm.Descr)

	// 验证权限策略更新
	sub := auth.ApiToSubject(updatedFm.ID)
	ok, enforceErr := suite.enforcer.Enforce(sub, updatedFm.URL, updatedFm.Method)
	suite.NoError(enforceErr, "Enforce应该成功")
	suite.True(ok, "Enforce应该返回true")
}

func (suite *ApiTestSuite) TestUpdateApiByID_NotFound() {
	updateDTO := sysmodel.UpdateApiDTO{
		URL:    "/api/test/",
		Method: "GET",
		Label:  "updated_test",
		Descr:  "测试更新",
	}
	_, err := suite.apiservice.UpdateApiByID(context.Background(), 0, updateDTO)
	suite.NotNil(err, "更新不存在的API应该失败")
}

func (suite *ApiTestSuite) TestListApi() {
	// 创建多个API
	apiCount := 3
	for i := 0; i < apiCount; i++ {
		dto := CreateTestApiDTO()
		_, err := suite.apiservice.CreateApi(context.Background(), dto)
		suite.Nil(err, "创建API应该成功")
	}

	// 测试列出所有API
	listDTO := sysmodel.ListApiDTO{}
	page, size := listDTO.StandardModelQuery.GetPageParam()
	count, apiList, err := suite.apiservice.ListApi(context.Background(), page, size, listDTO)
	suite.Nil(err, "列出API应该成功")
	suite.GreaterOrEqual(int(count), apiCount, "返回的API数量应该大于等于创建的数量")
	suite.NotNil(apiList, "返回的API列表不应该为nil")
}

func (suite *ApiTestSuite) TestLoadApiPolicy() {
	// 创建几个API
	apiCount := 2
	createdApis := make([]sysmodel.ApiModel, 0, apiCount)
	for range apiCount {
		dto := CreateTestApiDTO()
		fm, err := suite.apiservice.CreateApi(context.Background(), dto)
		suite.Nil(err, "创建API应该成功")
		createdApis = append(createdApis, *fm)
	}

	// 加载API策略
	err := suite.apiservice.LoadApiPolicy(context.Background())
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
	dto := CreateTestApiDTO()
	_, err := suite.apiservice.CreateApi(ctx, dto)
	suite.NotNil(err, "上下文错误时创建API应该失败")
}

func (suite *ApiTestSuite) TestUpdateApiByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文更新API
	updateDTO := sysmodel.UpdateApiDTO{
		URL:    "/api/test/",
		Method: "GET",
		Label:  "updated_test",
		Descr:  "测试更新",
	}
	_, err := suite.apiservice.UpdateApiByID(ctx, 1, updateDTO)
	suite.NotNil(err, "上下文错误时更新API应该失败")
}

func (suite *ApiTestSuite) TestDeleteApiByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文删除API
	err := suite.apiservice.DeleteApiByID(ctx, 1)
	suite.NotNil(err, "上下文错误时删除API应该失败")
}

func (suite *ApiTestSuite) TestFindApiByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文查找API
	_, err := suite.apiservice.FindApiByID(ctx, 1)
	suite.NotNil(err, "上下文错误时查找API应该失败")
}

func (suite *ApiTestSuite) TestListApi_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文列出API
	listDTO := sysmodel.ListApiDTO{}
	page, size := listDTO.StandardModelQuery.GetPageParam()
	_, _, err := suite.apiservice.ListApi(ctx, page, size, listDTO)
	suite.NotNil(err, "上下文错误时列出API应该失败")
}

func (suite *ApiTestSuite) TestLoadApiPolicy_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文加载API策略
	err := suite.apiservice.LoadApiPolicy(ctx)
	suite.NotNil(err, "上下文错误时加载API策略应该失败")
}

// 每个测试文件都需要这个入口函数
func TestApiTestSuite(t *testing.T) {
	pts := &ApiTestSuite{}
	suite.Run(t, pts)
}
