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

func CreateTestApiModel(opts ...func(*sysmodel.ApiModel)) *sysmodel.ApiModel {
	model := &sysmodel.ApiModel{
		URL:    fmt.Sprintf("/api/test/%s/", uuid.NewString()),
		Method: "GET",
		Label:  "test",
		Descr:  "这是一个测试接口",
	}

	// 应用可选配置
	for _, opt := range opts {
		opt(model)
	}

	return model
}

// WithMethod 设置API方法
func WithMethod(method string) func(*sysmodel.ApiModel) {
	return func(m *sysmodel.ApiModel) {
		m.Method = method
	}
}

// WithURL 设置API URL
func WithURL(url string) func(*sysmodel.ApiModel) {
	return func(m *sysmodel.ApiModel) {
		m.URL = url
	}
}

// WithLabel 设置API标签
func WithLabel(label string) func(*sysmodel.ApiModel) {
	return func(m *sysmodel.ApiModel) {
		m.Label = label
	}
}

// WithDescr 设置API描述
func WithDescr(descr string) func(*sysmodel.ApiModel) {
	return func(m *sysmodel.ApiModel) {
		m.Descr = descr
	}
}

type ApiTestSuite struct {
	suite.Suite
	apiRepo *ApiRepo
}

func (suite *ApiTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&sysmodel.ApiModel{}); err != nil {
		suite.Error(err, "数据库迁移失败")
	}
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()
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
}

func (suite *ApiTestSuite) TestCreateApi() {
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.Require().NoError(err, "创建API应该成功")

	fm, err := suite.apiRepo.GetModel(context.Background(), sm.ID)
	suite.Require().NoError(err, "查询刚创建的API应该成功")
	suite.Equal(sm.ID, fm.ID, "API ID应该匹配")
	suite.Equal(sm.URL, fm.URL, "API URL应该匹配")
	suite.Equal(sm.Method, fm.Method, "API Method应该匹配")
	suite.Equal(sm.Label, fm.Label, "API Label应该匹配")
	suite.Equal(sm.Descr, fm.Descr, "API Descr应该匹配")
}

func (suite *ApiTestSuite) TestUpdateApi() {
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.Require().NoError(err, "创建API应该成功")

	updatedURL := fmt.Sprintf("/api/%d", sm.ID)
	updatedMethod := "POST"
	updatedLabel := "updated_test"
	updatedDescr := "这是更新的测试接口"

	err = suite.apiRepo.UpdateModel(context.Background(), map[string]any{
		"url":    updatedURL,
		"method": updatedMethod,
		"label":  updatedLabel,
		"descr":  updatedDescr,
	}, "id = ?", sm.ID)
	suite.Require().NoError(err, "更新API应该成功")

	fm, err := suite.apiRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.Require().NoError(err, "查询更新后的API应该成功")
	suite.Equal(fm.ID, sm.ID, "API ID应该保持不变")
	suite.Equal(updatedURL, fm.URL, "API URL应该被更新")
	suite.Equal(updatedMethod, fm.Method, "API Method应该被更新")
	suite.Equal(updatedLabel, fm.Label, "API Label应该被更新")
	suite.Equal(updatedDescr, fm.Descr, "API Descr应该被更新")
	suite.Greater(fm.UpdatedAt, sm.UpdatedAt, "API UpdatedAt应该被更新")
}

func (suite *ApiTestSuite) TestDeleteApi() {
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.Require().NoError(err, "创建API应该成功")

	fm, err := suite.apiRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.Require().NoError(err, "查询刚创建的API应该成功")
	suite.Equal(sm.ID, fm.ID, "API ID应该匹配")

	err = suite.apiRepo.DeleteModel(context.Background(), "id = ?", sm.ID)
	suite.Require().NoError(err, "删除API应该成功")

	_, err = suite.apiRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.Error(err, "查询已删除的API应该返回错误")
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
}

func (suite *ApiTestSuite) TestListApi() {
	// 清理可能存在的数据并创建测试数据
	for range 10 {
		sm := CreateTestApiModel()
		err := suite.apiRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建API应该成功")
	}

	// 测试CountModel
	count, err := suite.apiRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取API总数应该成功")
	suite.GreaterOrEqual(count, int64(10), "API总数应该至少有10条")

	// 测试基本列表查询
	qp := database.QueryParams{
		Limit:  10,
		Offset: 0,
	}
	ms, err := suite.apiRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列出API应该成功")
	suite.NotNil(ms, "API列表不应该为nil")
	suite.GreaterOrEqual(len(ms), 10, "API列表应该至少有10条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:  5,
		Offset: 0,
	}
	pMs, err := suite.apiRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页列出API应该成功")
	suite.NotNil(pMs, "分页API列表不应该为nil")
	suite.Equal(5, len(pMs), "分页查询应该返回指定数量的记录")

	// 测试空参数查询
	qpEmpty := database.QueryParams{}
	msEmpty, err := suite.apiRepo.ListModel(context.Background(), qpEmpty)
	suite.NoError(err, "列表查询时传入空参数应该成功")
	suite.NotNil(msEmpty, "API列表不应该为nil")

	// 测试无效分页参数查询
	qpInvalid := database.QueryParams{
		Limit:  -1,
		Offset: -1,
	}
	msInvalid, err := suite.apiRepo.ListModel(context.Background(), qpInvalid)
	suite.NoError(err, "列表查询时传入无效分页参数应该成功")
	suite.NotNil(msInvalid, "API列表不应该为nil")

	// 测试排序查询
	// 先创建多个API用于测试
	for i := 0; i < 5; i++ {
		sm := CreateTestApiModel(
			WithURL(fmt.Sprintf("/api/test/%d/", i)),
			WithLabel(fmt.Sprintf("test_%d", i)),
			WithDescr(fmt.Sprintf("这是测试接口 %d", i)),
		)
		err := suite.apiRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建API应该成功")
	}

	// 测试按ID降序排序
	qpSort := database.QueryParams{
		OrderBy: []string{"id DESC"},
	}
	msSort, err := suite.apiRepo.ListModel(context.Background(), qpSort)
	suite.NoError(err, "按ID降序排序查询应该成功")
	suite.NotNil(msSort, "API列表不应该为nil")
	if len(msSort) > 1 {
		// 验证排序结果
		prevID := msSort[0].ID
		for _, api := range msSort {
			suite.LessOrEqual(api.ID, prevID, "API应该按ID降序排序")
			prevID = api.ID
		}
	}

	// 测试过滤查询
	// 创建一个特定标签的API
	testLabel := "filter_test"
	smFilter := CreateTestApiModel(
		WithURL("/api/filter/test/"),
		WithLabel(testLabel),
		WithDescr("这是一个用于过滤测试的接口"),
	)
	err = suite.apiRepo.CreateModel(context.Background(), smFilter)
	suite.NoError(err, "创建API应该成功")

	// 测试按标签过滤
	qpFilter := database.QueryParams{
		Query: map[string]any{
			"label": testLabel,
		},
	}
	msFilter, err := suite.apiRepo.ListModel(context.Background(), qpFilter)
	suite.NoError(err, "按标签过滤查询应该成功")
	suite.NotNil(msFilter, "API列表不应该为nil")
	// 验证过滤结果
	for _, api := range msFilter {
		suite.Equal(testLabel, api.Label, "API应该按标签过滤")
	}

	// 测试CountModel带过滤条件
	countFilter, err := suite.apiRepo.CountModel(context.Background(), map[string]any{
		"label": testLabel,
	})
	suite.NoError(err, "带过滤条件的API总数查询应该成功")
	suite.GreaterOrEqual(countFilter, int64(1), "带过滤条件的API总数应该至少为1")
}

func (suite *ApiTestSuite) TestAddPolicy() {
	// 创建一个API用于测试策略添加
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建API应该成功")

	err = suite.apiRepo.AddPolicy(context.Background(), *sm)
	suite.NoError(err, "添加策略应该成功")

	sub := auth.ApiToSubject(sm.ID)
	ok, err := suite.apiRepo.enforcer.Enforce(sub, sm.URL, sm.Method)
	suite.NoError(err, "检查授权应该成功")
	suite.True(ok, "添加策略后应该有API")

	// 测试重复添加策略
	err = suite.apiRepo.AddPolicy(context.Background(), *sm)
	suite.NoError(err, "重复添加策略应该成功")
}

func (suite *ApiTestSuite) TestPolicyWithInvalidModel() {
	// 测试添加策略时传入无效的模型
	testCases := []struct {
		name     string
		model    sysmodel.ApiModel
		expected string
	}{
		{
			name: "ID为0",
			model: sysmodel.ApiModel{
				URL:    "/api/test",
				Method: "GET",
			},
			expected: "添加策略时ID为0应该返回错误",
		},
		{
			name: "URL为空",
			model: sysmodel.ApiModel{
				StandardModel: database.StandardModel{
					BaseModel: database.BaseModel{
						ID: 1,
					},
				},
				Method: "GET",
			},
			expected: "添加策略时URL为空应该返回错误",
		},
		{
			name: "Method为空",
			model: sysmodel.ApiModel{
				StandardModel: database.StandardModel{
					BaseModel: database.BaseModel{
						ID: 1,
					},
				},
				URL: "/api/test",
			},
			expected: "添加策略时Method为空应该返回错误",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.apiRepo.AddPolicy(context.Background(), tc.model)
			suite.Error(err, tc.expected)

			err = suite.apiRepo.RemovePolicy(context.Background(), tc.model, true)
			suite.Error(err, "删除策略时"+tc.name+"应该返回错误")
		})
	}
}

func (suite *ApiTestSuite) TestRemovePolicy() {
	// 创建一个API用于测试策略删除
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建API应该成功")

	err = suite.apiRepo.AddPolicy(context.Background(), *sm)
	suite.NoError(err, "添加策略应该成功")

	sub := auth.ApiToSubject(sm.ID)
	ok, err := suite.apiRepo.enforcer.Enforce(sub, sm.URL, sm.Method)
	suite.NoError(err, "检查授权应该成功")
	suite.True(ok, "添加策略后应该有API")

	// 测试删除策略并删除继承的组策略
	err = suite.apiRepo.RemovePolicy(context.Background(), *sm, true)
	suite.NoError(err, "删除策略并删除继承的组策略应该成功")

	// 验证策略已移除
	ok, err = suite.apiRepo.enforcer.Enforce(sub, sm.URL, sm.Method)
	suite.NoError(err, "检查授权应该成功")
	suite.False(ok, "移除策略后不应该有API")

	// 再次尝试删除相同的策略，应该不会报错
	err = suite.apiRepo.RemovePolicy(context.Background(), *sm, false)
	suite.NoError(err, "删除不存在的策略应该成功")

	// 重新添加策略并测试不删除继承的组策略
	err = suite.apiRepo.AddPolicy(context.Background(), *sm)
	suite.NoError(err, "重新添加策略应该成功")

	err = suite.apiRepo.RemovePolicy(context.Background(), *sm, false)
	suite.NoError(err, "删除策略但不删除继承的组策略应该成功")

	// 验证策略已移除
	ok, err = suite.apiRepo.enforcer.Enforce(sub, sm.URL, sm.Method)
	suite.NoError(err, "检查授权应该成功")
	suite.False(ok, "移除策略后不应该有API")
}

func (suite *ApiTestSuite) TestContextTimeout() {
	// 创建一个会立即超时的上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 创建API用于测试
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建API应该成功")

	// 测试CreateModel方法
	sm2 := CreateTestApiModel()
	err = suite.apiRepo.CreateModel(ctx, sm2)
	suite.Error(err, "上下文超时后创建API应该返回错误")

	// 测试UpdateModel方法
	err = suite.apiRepo.UpdateModel(ctx, map[string]any{
		"url": "/api/test",
	}, "id = ?", sm.ID)
	suite.Error(err, "上下文超时后更新API应该返回错误")

	// 测试DeleteModel方法
	err = suite.apiRepo.DeleteModel(ctx, "id = ?", sm.ID)
	suite.Error(err, "上下文超时后删除API应该返回错误")

	// 测试GetModel方法
	_, err = suite.apiRepo.GetModel(ctx, sm.ID)
	suite.Error(err, "上下文超时后获取API应该返回错误")

	// 测试ListModel方法
	qp := database.QueryParams{}
	_, err = suite.apiRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文超时后查询API列表应该返回错误")

	// 测试CountModel方法
	_, err = suite.apiRepo.CountModel(ctx, nil)
	suite.Error(err, "上下文超时后计数查询应该返回错误")

	// 测试AddPolicy方法
	err = suite.apiRepo.AddPolicy(ctx, *sm)
	suite.Error(err, "上下文超时后添加策略应该返回错误")

	// 测试RemovePolicy方法
	err = suite.apiRepo.RemovePolicy(ctx, *sm, true)
	suite.Error(err, "上下文超时后删除策略应该返回错误")
}

func (suite *ApiTestSuite) TestCreateApiWithInvalidData() {
	// 测试创建API时传入空数据
	err := suite.apiRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建API时传入nil应该返回错误")
}

func (suite *ApiTestSuite) TestFindNonExistentApi() {
	// 测试查找不存在的API
	_, err := suite.apiRepo.GetModel(context.Background(), 999999)
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查找不存在的API应该返回记录未找到错误")
}

func (suite *ApiTestSuite) TestUpdateApiWithEmptyData() {
	// 测试更新时传入空数据
	err := suite.apiRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", 1)
	suite.Error(err, "更新API时传入空数据应该返回错误")
}

func (suite *ApiTestSuite) TestUpdateNonExistentApi() {
	// 测试更新不存在的API
	err := suite.apiRepo.UpdateModel(context.Background(), map[string]any{
		"url": "/api/non-existent",
	}, "id = ?", 999999)
	suite.NoError(err, "更新不存在的API不应该返回错误")
}

func (suite *ApiTestSuite) TestDeleteApiWithEmptyConditions() {
	// 测试删除时传入空条件
	err := suite.apiRepo.DeleteModel(context.Background())
	suite.Error(err, "删除时传入空条件应该返回错误")
}

func (suite *ApiTestSuite) TestDeleteNonExistentApi() {
	// 测试删除不存在的API
	err := suite.apiRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的API不应该返回错误")
}

func (suite *ApiTestSuite) TestGetApiWithEmptyConditions() {
	// 测试查询时传入空条件
	result, err := suite.apiRepo.GetModel(context.Background())
	// 当传入空条件时，GetModel方法会尝试获取数据库中的第一条记录
	// 如果数据库为空，会返回record not found错误
	// 如果数据库不为空，会返回第一条记录
	if err != nil {
		// 如果返回错误，应该是record not found
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询时传入空条件应该返回记录未找到错误")
	} else {
		// 如果返回结果，应该是一个有效的API模型
		suite.NotNil(result, "查询时传入空条件应该返回有效的API模型")
		suite.Greater(result.ID, uint32(0), "返回的API模型ID应该大于0")
	}
}

func (suite *ApiTestSuite) TestGetApiWithCanceledContext() {
	// 测试查询时上下文已取消
	testCtx := context.Background()
	ctx, cancel := context.WithCancel(testCtx)
	cancel()
	_, err := suite.apiRepo.GetModel(ctx, 1)
	suite.Error(err, "查询时上下文已取消应该返回错误")
}

func (suite *ApiTestSuite) TestCreateApiWithContextTimeout() {
	// 测试创建API时上下文超时
	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(ctx, sm)
	suite.Error(err, "创建API时上下文超时应该返回错误")
}

func (suite *ApiTestSuite) TestUpdateApiWithContextTimeout() {
	// 测试更新API时上下文超时
	// 先创建一个API
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建API应该成功")

	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试更新API
	err = suite.apiRepo.UpdateModel(ctx, map[string]any{
		"label": "updated_test",
	}, "id = ?", sm.ID)
	suite.Error(err, "更新API时上下文超时应该返回错误")
}

func (suite *ApiTestSuite) TestDeleteApiWithContextTimeout() {
	// 测试删除API时上下文超时
	// 先创建一个API
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建API应该成功")

	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试删除API
	err = suite.apiRepo.DeleteModel(ctx, "id = ?", sm.ID)
	suite.Error(err, "删除API时上下文超时应该返回错误")
}

func (suite *ApiTestSuite) TestGetApiWithContextTimeout() {
	// 测试获取API时上下文超时
	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试获取API
	_, err := suite.apiRepo.GetModel(ctx, 1)
	suite.Error(err, "获取API时上下文超时应该返回错误")
}

func (suite *ApiTestSuite) TestListApiWithContextTimeout() {
	// 测试列表查询时上下文超时
	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试列表查询
	qp := database.QueryParams{}
	_, err := suite.apiRepo.ListModel(ctx, qp)
	suite.Error(err, "列表查询时上下文超时应该返回错误")
}

func (suite *ApiTestSuite) TestCountApiWithContextTimeout() {
	// 测试计数查询时上下文超时
	// 创建一个非常短的超时上下文
	testCtx := context.Background()
	ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试计数查询
	_, err := suite.apiRepo.CountModel(ctx, nil)
	suite.Error(err, "计数查询时上下文超时应该返回错误")
}

func (suite *ApiTestSuite) TestCreateApiWithSameUrlDifferentMethod() {
	// 测试创建具有相同URL但不同Method的API
	sm1 := &sysmodel.ApiModel{
		URL:    "/api/same/url/test/",
		Method: "GET",
		Label:  "get_test",
		Descr:  "这是一个GET测试接口",
	}
	err := suite.apiRepo.CreateModel(context.Background(), sm1)
	suite.NoError(err, "创建GET方法的API应该成功")

	// 尝试创建相同URL但不同Method的API
	sm2 := &sysmodel.ApiModel{
		URL:    "/api/same/url/test/",
		Method: "POST",
		Label:  "post_test",
		Descr:  "这是一个POST测试接口",
	}
	err = suite.apiRepo.CreateModel(context.Background(), sm2)
	suite.NoError(err, "创建相同URL但不同Method的API应该成功")
}

// 每个测试文件都需要这个入口函数
func TestApiTestSuite(t *testing.T) {
	pts := &ApiTestSuite{}
	suite.Run(t, pts)
}

// TestNewApiRepo 测试创建API仓库实例
func TestNewApiRepo(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()

	repo := NewApiRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	if repo == nil {
		t.Fatal("NewApiRepo should return a non-nil repository")
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
	if repo.slowThreshold == nil {
		t.Fatal("Repo slowThreshold should not be nil")
	}
	if repo.enforcer == nil {
		t.Fatal("Repo enforcer should not be nil")
	}
}
