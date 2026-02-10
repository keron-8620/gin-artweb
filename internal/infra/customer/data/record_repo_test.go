package data

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/infra/customer/biz"
	"gin-artweb/internal/shared/database"
)

// CreateTestLoginRecordModel 创建测试用的登录记录模型
func CreateTestLoginRecordModel(pk uint32, ip string) *biz.LoginRecordModel {
	return &biz.LoginRecordModel{
		BaseModel: database.BaseModel{
			ID: pk,
		},
		Username:  "test_user",
		IPAddress: ip,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		Status:    true,
	}
}

type RecordTestSuite struct {
	suite.Suite
	recordRepo *loginRecordRepo
}

func (suite *RecordTestSuite) SetupSuite() {
	suite.recordRepo = NewTestLoginRecordRepo()
}

func (suite *RecordTestSuite) TestCreateLoginRecord() {
	// 测试创建登录记录
	sm := CreateTestLoginRecordModel(1, "192.168.1.1")
	err := suite.recordRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建登录记录应该成功")

	// 测试查询登录记录列表
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	count, ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询登录记录列表应该成功")
	suite.NotNil(ms, "登录记录列表不应该为nil")
	suite.GreaterOrEqual(count, int64(1), "登录记录总数应该至少有1条")
}

func (suite *RecordTestSuite) TestGetLoginFailNum() {
	// 测试获取登录失败次数（新IP）
	num, err := suite.recordRepo.GetLoginFailNum(context.Background(), "192.168.1.2")
	suite.NoError(err, "获取登录失败次数应该成功")
	suite.Equal(5, num, "新IP应该返回最大允许失败次数")

	// 测试设置登录失败次数
	err = suite.recordRepo.SetLoginFailNum(context.Background(), "192.168.1.2", 3)
	suite.NoError(err, "设置登录失败次数应该成功")

	// 测试获取登录失败次数（已设置的IP）
	num, err = suite.recordRepo.GetLoginFailNum(context.Background(), "192.168.1.2")
	suite.NoError(err, "获取登录失败次数应该成功")
	suite.Equal(3, num, "已设置的IP应该返回设置的失败次数")
}

func (suite *RecordTestSuite) TestGetLoginFailNumWithEmptyIP() {
	// 测试获取登录失败次数时传入空IP
	_, err := suite.recordRepo.GetLoginFailNum(context.Background(), "")
	suite.Error(err, "获取登录失败次数时传入空IP应该返回错误")
}

func (suite *RecordTestSuite) TestSetLoginFailNumWithEmptyIP() {
	// 测试设置登录失败次数时传入空IP
	err := suite.recordRepo.SetLoginFailNum(context.Background(), "", 3)
	suite.Error(err, "设置登录失败次数时传入空IP应该返回错误")
}

func (suite *RecordTestSuite) TestListLoginRecords() {
	// 测试创建多个登录记录
	for i := 10; i <= 14; i++ {
		sm := CreateTestLoginRecordModel(uint32(i), "192.168.1."+string(rune('0'+(i-9))))
		err := suite.recordRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建登录记录应该成功")
	}

	// 测试查询登录记录列表
	qp := database.QueryParams{
		Limit:   10,
		Offset:  0,
		IsCount: true,
	}
	count, ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询登录记录列表应该成功")
	suite.NotNil(ms, "登录记录列表不应该为nil")
	suite.GreaterOrEqual(count, int64(5), "登录记录总数应该至少有5条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:   2,
		Offset:  0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.recordRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页查询登录记录列表应该成功")
	suite.NotNil(pMs, "分页登录记录列表不应该为nil")
	suite.Equal(2, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(2), "分页总数应该至少等于limit")
}

// 每个测试文件都需要这个入口函数
func TestRecordTestSuite(t *testing.T) {
	pts := &RecordTestSuite{}
	suite.Run(t, pts)
}
