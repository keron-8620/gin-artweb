package sys

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/suite"

	sysmodel "gin-artweb/internal/model/sys"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

// CreateTestLoginRecordModel 创建测试用的登录记录模型
func CreateTestLoginRecordModel(ip string, status ...bool) *sysmodel.LoginRecordModel {
	isSuccess := true
	if len(status) > 0 {
		isSuccess = status[0]
	}
	return &sysmodel.LoginRecordModel{
		Username:  "test_user",
		IPAddress: ip,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		Status:    isSuccess,
	}
}

type RecordTestSuite struct {
	suite.Suite
	recordRepo *LoginRecordRepo
}

func (suite *RecordTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&sysmodel.LoginRecordModel{}); err != nil {
		suite.Error(err, "数据库迁移失败")
	}
	dbTimeout := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()
	suite.recordRepo = &LoginRecordRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: slowThreshold,
		cache:         cache.New(5*time.Minute, 10*time.Minute),
		maxNum:        5,
		ttl:           5 * time.Minute,
	}
}

func (suite *RecordTestSuite) TestCreateModel() {
	// 测试正常场景:创建登录记录模型
	sm := CreateTestLoginRecordModel("192.168.1.1")
	err := suite.recordRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建登录记录应该成功")

	// 验证登录时间被设置
	suite.NotZero(sm.LoginAt, "登录时间应该被设置")
}

func (suite *RecordTestSuite) TestCreateModelWithNil() {
	// 测试异常场景:传入nil模型参数
	err := suite.recordRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建登录记录时传入nil应该返回错误")
}

func (suite *RecordTestSuite) TestListAndCountModel() {
	// 清理可能存在的数据并创建测试数据
	testIP := "192.168.1.100"
	for i := range 10 {
		// 使用不同的ID范围，避免与其他测试冲突
		sm := CreateTestLoginRecordModel(fmt.Sprintf("192.168.1.%d", i+1))
		err := suite.recordRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建登录记录应该成功")
	}

	// 测试正常场景:查询登录记录列表
	qp := database.QueryParams{
		Limit:  10,
		Offset: 0,
	}
	ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列出登录记录应该成功")
	suite.NotNil(ms, "登录记录列表不应该为nil")
	suite.GreaterOrEqual(len(ms), 10, "登录记录列表应该至少有10条")

	// 测试CountModel
	count, err := suite.recordRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "获取登录记录总数应该成功")
	suite.GreaterOrEqual(count, int64(10), "登录记录总数应该至少有10条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:  5,
		Offset: 0,
	}
	pMs, err := suite.recordRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页列出登录记录应该成功")
	suite.NotNil(pMs, "分页登录记录列表不应该为nil")
	suite.Equal(5, len(pMs), "分页查询应该返回指定数量的记录")

	// 测试排序查询
	qpSorted := database.QueryParams{
		Limit:   10,
		Offset:  0,
		OrderBy: []string{"id DESC"},
	}
	sMs, err := suite.recordRepo.ListModel(context.Background(), qpSorted)
	suite.NoError(err, "排序查询登录记录应该成功")
	suite.NotNil(sMs, "排序登录记录列表不应该为nil")
	if len(sMs) > 1 {
		// 验证排序结果
		prevID := sMs[0].ID
		for _, record := range sMs {
			suite.LessOrEqual(record.ID, prevID, "登录记录应该按ID降序排序")
			prevID = record.ID
		}
	}

	// 测试带过滤条件的查询
	sm := CreateTestLoginRecordModel(testIP)
	err = suite.recordRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建登录记录应该成功")

	qpFiltered := database.QueryParams{
		Query: map[string]any{
			"ip_address": testIP,
		},
	}
	fMs, err := suite.recordRepo.ListModel(context.Background(), qpFiltered)
	suite.NoError(err, "按IP过滤查询应该成功")
	suite.NotNil(fMs, "登录记录列表不应该为nil")
	for _, record := range fMs {
		suite.Equal(testIP, record.IPAddress, "登录记录应该按IP过滤")
	}

	// 测试带过滤条件的CountModel
	filterIP := "192.168.1.1"
	count, err = suite.recordRepo.CountModel(context.Background(), map[string]any{"ip_address": filterIP})
	suite.NoError(err, "带过滤条件的登录记录总数查询应该成功")
	suite.GreaterOrEqual(count, int64(1), "带过滤条件的登录记录总数应该至少为1")

	count, err = suite.recordRepo.CountModel(context.Background(), map[string]any{"ip_address": "192.168.1.999"})
	suite.NoError(err, "带不存在的过滤条件的登录记录总数查询应该成功")
	suite.Equal(int64(0), count, "带不存在的过滤条件的登录记录总数应该为0")
}

func (suite *RecordTestSuite) TestLoginFailNum() {
	// 测试正常场景:获取不存在IP的登录失败次数，应该返回maxNum
	num, err := suite.recordRepo.GetLoginFailNum(context.Background(), "192.168.1.100")
	suite.NoError(err, "获取不存在IP的登录失败次数应该成功")
	suite.Equal(suite.recordRepo.maxNum, num, "获取不存在IP的登录失败次数应该返回maxNum")

	// 测试异常场景:传入空IP地址
	num, err = suite.recordRepo.GetLoginFailNum(context.Background(), "")
	suite.Error(err, "获取登录失败次数时传入空IP应该返回错误")
	suite.Equal(0, num, "传入空IP时应该返回0")

	// 测试正常场景:设置登录失败次数
	ip := "192.168.1.100"
	failNum := 3
	err = suite.recordRepo.SetLoginFailNum(context.Background(), ip, failNum)
	suite.NoError(err, "设置登录失败次数应该成功")

	// 验证设置是否成功
	num, err = suite.recordRepo.GetLoginFailNum(context.Background(), ip)
	suite.NoError(err, "获取登录失败次数应该成功")
	suite.Equal(failNum, num, "获取的登录失败次数应该等于设置的值")

	// 测试异常场景:设置时传入空IP地址
	err = suite.recordRepo.SetLoginFailNum(context.Background(), "", 3)
	suite.Error(err, "设置登录失败次数时传入空IP应该返回错误")
}

func (suite *RecordTestSuite) TestCacheExpiration() {
	// 创建一个具有短TTL的临时仓库用于测试缓存过期
	tempRepo := &LoginRecordRepo{
		log:      suite.recordRepo.log,
		gormDB:   suite.recordRepo.gormDB,
		timeouts: suite.recordRepo.timeouts,
		cache:    cache.New(100*time.Millisecond, 200*time.Millisecond), // 短TTL
		maxNum:   5,
		ttl:      100 * time.Millisecond,
	}

	// 设置登录失败次数
	ip := "192.168.1.200"
	failNum := 3
	err := tempRepo.SetLoginFailNum(context.Background(), ip, failNum)
	suite.NoError(err, "设置登录失败次数应该成功")

	// 验证设置是否成功
	num, err := tempRepo.GetLoginFailNum(context.Background(), ip)
	suite.NoError(err, "获取登录失败次数应该成功")
	suite.Equal(failNum, num, "获取的登录失败次数应该等于设置的值")

	// 等待缓存过期
	time.Sleep(200 * time.Millisecond)

	// 验证缓存过期后获取登录失败次数的行为
	num, err = tempRepo.GetLoginFailNum(context.Background(), ip)
	suite.NoError(err, "缓存过期后获取登录失败次数应该成功")
	suite.Equal(tempRepo.maxNum, num, "缓存过期后应该返回maxNum")
}

func (suite *RecordTestSuite) TestWithCanceledContext() {
	// 测试上下文取消场景
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试CreateModel
	sm := CreateTestLoginRecordModel("192.168.1.1")
	err := suite.recordRepo.CreateModel(ctx, sm)
	suite.Error(err, "上下文已取消时创建登录记录应该返回错误")

	// 测试GetLoginFailNum
	num, err := suite.recordRepo.GetLoginFailNum(ctx, "192.168.1.100")
	suite.Error(err, "获取登录失败次数时上下文被取消应该返回错误")
	suite.Equal(0, num, "上下文被取消时应该返回0")

	// 测试SetLoginFailNum
	err = suite.recordRepo.SetLoginFailNum(ctx, "192.168.1.100", 3)
	suite.Error(err, "设置登录失败次数时上下文被取消应该返回错误")

	// 测试ListModel
	qp := database.QueryParams{
		Limit:  10,
		Offset: 0,
	}
	_, err = suite.recordRepo.ListModel(ctx, qp)
	suite.Error(err, "上下文已取消时列出登录记录应该返回错误")

	// 测试CountModel
	_, err = suite.recordRepo.CountModel(ctx, nil)
	suite.Error(err, "上下文已取消时获取登录记录总数应该返回错误")
}

// 每个测试文件都需要这个入口函数
func TestRecordTestSuite(t *testing.T) {
	pts := &RecordTestSuite{}
	suite.Run(t, pts)
}

// TestNewLoginRecordRepo 测试创建登录记录仓库实例
func TestNewLoginRecordRepo(t *testing.T) {
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()
	lockTime := 5 * time.Minute
	clearTime := 10 * time.Minute
	maxNum := 5

	repo := NewLoginRecordRepo(logger, db, dbTimeout, slowThreshold, lockTime, clearTime, maxNum)
	if repo == nil {
		t.Fatal("NewLoginRecordRepo should return a non-nil repository")
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
	if repo.cache == nil {
		t.Fatal("Repo cache should not be nil")
	}
	if repo.maxNum != maxNum {
		t.Fatal("Repo maxNum should be set correctly")
	}
}
