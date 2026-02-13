package data

import (
	"context"
	"errors"
	"fmt"
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
	apiRepo *ApiRepo
}

func (suite *ApiTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&model.ApiModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	suite.apiRepo = &ApiRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}
}

func (suite *ApiTestSuite) TestCreateApi() {
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建API应该成功")

	fm, err := suite.apiRepo.GetModel(context.Background(), sm.ID)
	suite.NoError(err, "查询刚创建的API应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(sm.URL, fm.URL)
	suite.Equal(sm.Method, fm.Method)
	suite.Equal(sm.Label, fm.Label)
	suite.Equal(sm.Descr, fm.Descr)
}

func (suite *ApiTestSuite) TestUpdateApi() {
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建API应该成功")

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
	suite.NoError(err, "更新API应该成功")

	fm, err := suite.apiRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的API应该成功")
	suite.Equal(fm.ID, sm.ID)
	suite.Equal(updatedURL, fm.URL)
	suite.Equal(updatedMethod, fm.Method)
	suite.Equal(updatedLabel, fm.Label)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Greater(fm.UpdatedAt, sm.UpdatedAt)
}

func (suite *ApiTestSuite) TestDeleteApi() {
	sm := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建API应该成功")

	fm, err := suite.apiRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的API应该成功")
	suite.Equal(sm.ID, fm.ID)

	err = suite.apiRepo.DeleteModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "删除API应该成功")

	_, err = suite.apiRepo.GetModel(context.Background(), "id = ?", sm.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *ApiTestSuite) TestListApi() {
	// 清理可能存在的数据并创建测试数据
	for range 10 {
		sm := CreateTestApiModel()
		err := suite.apiRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建API应该成功")
	}

	qp := database.QueryParams{
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	total, ms, err := suite.apiRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列出API应该成功")
	suite.NotNil(ms, "API列表不应该为nil")
	suite.GreaterOrEqual(total, int64(10), "API总数应该至少有10条")

	qpPaginated := database.QueryParams{
		Size:    5,
		Page:    0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.apiRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页列出API应该成功")
	suite.NotNil(pMs, "分页API列表不应该为nil")
	suite.Equal(5, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(5), "分页总数应该至少等于limit")
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

func (suite *ApiTestSuite) TestAddPolicyWithZeroID() {
	// 测试添加策略时ID为0
	m := model.ApiModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: 0,
			},
		},
		URL:    "/api/test",
		Method: "GET",
	}
	err := suite.apiRepo.AddPolicy(context.Background(), m)
	suite.Error(err, "添加策略时ID为0应该返回错误")
}

func (suite *ApiTestSuite) TestAddPolicyWithEmptyURL() {
	// 测试添加策略时URL为空
	m := model.ApiModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: 1,
			},
		},
		URL:    "",
		Method: "GET",
	}
	err := suite.apiRepo.AddPolicy(context.Background(), m)
	suite.Error(err, "添加策略时URL为空应该返回错误")
}

func (suite *ApiTestSuite) TestAddPolicyWithEmptyMethod() {
	// 测试添加策略时Method为空
	m := model.ApiModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: 1,
			},
		},
		URL:    "/api/test",
		Method: "",
	}
	err := suite.apiRepo.AddPolicy(context.Background(), m)
	suite.Error(err, "添加策略时Method为空应该返回错误")
}

func (suite *ApiTestSuite) TestAddPolicyWithCanceledContext() {
	// 测试添加策略时上下文已取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), m)
	suite.NoError(err, "创建API应该成功")
	err = suite.apiRepo.AddPolicy(ctx, *m)
	suite.Error(err, "添加策略时上下文已取消应该返回错误")
}

func (suite *ApiTestSuite) TestRemovePolicyWithZeroID() {
	// 测试删除策略时ID为0
	m := model.ApiModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: 0,
			},
		},
		URL:    "/api/test",
		Method: "GET",
	}
	err := suite.apiRepo.RemovePolicy(context.Background(), m, true)
	suite.Error(err, "删除策略时ID为0应该返回错误")
}

func (suite *ApiTestSuite) TestRemovePolicyWithEmptyURL() {
	// 测试删除策略时URL为空
	m := model.ApiModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: 1,
			},
		},
		URL:    "",
		Method: "GET",
	}
	err := suite.apiRepo.RemovePolicy(context.Background(), m, true)
	suite.Error(err, "删除策略时URL为空应该返回错误")
}

func (suite *ApiTestSuite) TestRemovePolicyWithEmptyMethod() {
	// 测试删除策略时Method为空
	m := model.ApiModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: 1,
			},
		},
		URL:    "/api/test",
		Method: "",
	}
	err := suite.apiRepo.RemovePolicy(context.Background(), m, true)
	suite.Error(err, "删除策略时Method为空应该返回错误")
}

func (suite *ApiTestSuite) TestRemovePolicyWithCanceledContext() {
	// 测试删除策略时上下文已取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), m)
	suite.NoError(err, "创建API应该成功")
	err = suite.apiRepo.AddPolicy(context.Background(), *m)
	suite.NoError(err, "添加策略应该成功")
	err = suite.apiRepo.RemovePolicy(ctx, *m, true)
	suite.Error(err, "删除策略时上下文已取消应该返回错误")
}

func (suite *ApiTestSuite) TestRemovePolicyWithRemoveInherited() {
	// 测试删除策略时设置removeInherited为true
	m := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), m)
	suite.NoError(err, "创建API应该成功")

	err = suite.apiRepo.AddPolicy(context.Background(), *m)
	suite.NoError(err, "添加策略应该成功")

	// 测试删除策略并删除继承的组策略
	err = suite.apiRepo.RemovePolicy(context.Background(), *m, true)
	suite.NoError(err, "删除策略并删除继承的组策略应该成功")

	// 验证策略已移除
	sub := auth.ApiToSubject(m.ID)
	ok, err := suite.apiRepo.enforcer.Enforce(sub, m.URL, m.Method)
	suite.NoError(err, "检查授权应该成功")
	suite.False(ok, "移除策略后不应该有API")
}

func (suite *ApiTestSuite) TestRemovePolicyWithoutRemoveInherited() {
	// 测试删除策略时设置removeInherited为false
	m := CreateTestApiModel()
	err := suite.apiRepo.CreateModel(context.Background(), m)
	suite.NoError(err, "创建API应该成功")

	err = suite.apiRepo.AddPolicy(context.Background(), *m)
	suite.NoError(err, "添加策略应该成功")

	// 测试删除策略但不删除继承的组策略
	err = suite.apiRepo.RemovePolicy(context.Background(), *m, false)
	suite.NoError(err, "删除策略但不删除继承的组策略应该成功")

	// 验证策略已移除
	sub := auth.ApiToSubject(m.ID)
	ok, err := suite.apiRepo.enforcer.Enforce(sub, m.URL, m.Method)
	suite.NoError(err, "检查授权应该成功")
	suite.False(ok, "移除策略后不应该有API")
}

func (suite *ApiTestSuite) TestListApiWithEmptyParams() {
	// 测试列表查询时传入空参数
	qp := database.QueryParams{}
	total, ms, err := suite.apiRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列表查询时传入空参数应该成功")
	suite.NotNil(ms, "API列表不应该为nil")
	suite.GreaterOrEqual(total, int64(0), "API总数应该大于等于0")
}

func (suite *ApiTestSuite) TestListApiWithInvalidPagination() {
	// 测试列表查询时传入无效分页参数
	qp := database.QueryParams{
		Size: -1,
		Page: -1,
	}
	_, ms, err := suite.apiRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列表查询时传入无效分页参数应该成功")
	suite.NotNil(ms, "API列表不应该为nil")
}

func (suite *ApiTestSuite) TestListApiWithSorting() {
	// 测试列表查询时传入排序参数
	// 先创建多个API用于测试
	for i := 0; i < 5; i++ {
		sm := &model.ApiModel{
			URL:    fmt.Sprintf("/api/test/%d/", i),
			Method: "GET",
			Label:  fmt.Sprintf("test_%d", i),
			Descr:  fmt.Sprintf("这是测试接口 %d", i),
		}
		err := suite.apiRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建API应该成功")
	}

	// 测试按ID降序排序
	qp := database.QueryParams{
		OrderBy: []string{"id DESC"},
	}
	_, ms, err := suite.apiRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按ID降序排序查询应该成功")
	suite.NotNil(ms, "API列表不应该为nil")
	if len(*ms) > 1 {
		// 验证排序结果
		prevID := (*ms)[0].ID
		for _, api := range *ms {
			suite.LessOrEqual(api.ID, prevID, "API应该按ID降序排序")
			prevID = api.ID
		}
	}
}

func (suite *ApiTestSuite) TestListApiWithFiltering() {
	// 测试列表查询时传入过滤参数
	// 创建一个特定标签的API
	testLabel := "filter_test"
	sm := &model.ApiModel{
		URL:    "/api/filter/test/",
		Method: "GET",
		Label:  testLabel,
		Descr:  "这是一个用于过滤测试的接口",
	}
	err := suite.apiRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建API应该成功")

	// 测试按标签过滤
	qp := database.QueryParams{
		Query: map[string]any{
			"label": testLabel,
		},
	}
	_, ms, err := suite.apiRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按标签过滤查询应该成功")
	suite.NotNil(ms, "API列表不应该为nil")
	// 验证过滤结果
	for _, api := range *ms {
		suite.Equal(testLabel, api.Label, "API应该按标签过滤")
	}
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
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := suite.apiRepo.GetModel(ctx, 1)
	suite.Error(err, "查询时上下文已取消应该返回错误")
}

func (suite *ApiTestSuite) TestCreateApiWithContextTimeout() {
	// 测试创建API时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试列表查询
	qp := database.QueryParams{}
	_, _, err := suite.apiRepo.ListModel(ctx, qp)
	suite.Error(err, "列表查询时上下文超时应该返回错误")
}

func (suite *ApiTestSuite) TestCreateApiWithSameUrlDifferentMethod() {
	// 测试创建具有相同URL但不同Method的API
	sm1 := &model.ApiModel{
		URL:    "/api/same/url/test/",
		Method: "GET",
		Label:  "get_test",
		Descr:  "这是一个GET测试接口",
	}
	err := suite.apiRepo.CreateModel(context.Background(), sm1)
	suite.NoError(err, "创建GET方法的API应该成功")

	// 尝试创建相同URL但不同Method的API
	sm2 := &model.ApiModel{
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
