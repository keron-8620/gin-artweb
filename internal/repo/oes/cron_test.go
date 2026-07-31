package oes

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"

	jobmodel "gin-artweb/internal/model/job"
	oesmodel "gin-artweb/internal/model/oes"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

type OesCronTestSuite struct {
	suite.Suite
	cronRepo       *OesCronRepo
	oesColonyID    uint32
	scheduleID     uint32
	log            *zap.Logger
	gormDB         *gorm.DB
	timeouts       *config.DBTimeout
	nextScheduleID uint32
}

func (suite *OesCronTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&jobmodel.ScriptModel{}, &jobmodel.ScheduleModel{}, &oesmodel.OesColonyModel{}, &oesmodel.OesCronModel{}); err != nil {
		suite.Error(err, "数据库迁移失败")
	}

	// 创建测试数据:Script
	scriptModel := &jobmodel.ScriptModel{
		Name:     "test-script.sh",
		Descr:    "test script",
		Project:  "oes",
		Label:    "cron",
		Language: "bash",
		Status:   true,
		Username: "admin",
	}
	db.Create(scriptModel)

	// 创建测试数据:Schedule
	scheduleModel := &jobmodel.ScheduleModel{
		Name:          "test-schedule",
		Specification: "0 * * * *",
		IsEnabled:     true,
		Timeout:       300,
		CreateType:    1,
		Username:      "admin",
		ScriptID:      scriptModel.ID,
	}
	db.Create(scheduleModel)

	// 创建测试数据:OesColony
	esColonyModel := &oesmodel.OesColonyModel{
		SystemType:    "STK",
		ColonyNum:     "01",
		ExtractedName: "test-oes-colony",
		IsEnable:      true,
		PackageID:     1,
		XCounterID:    1,
		MonNodeID:     1,
	}
	db.Create(esColonyModel)

	suite.timeouts = test.NewTestDBTimeouts()
	suite.log = test.NewTestZapLogger()
	suite.gormDB = db
	slowThreshold := test.NewTestDBSlowThreshold()
	suite.cronRepo = &OesCronRepo{
		log:           suite.log,
		gormDB:        suite.gormDB,
		timeouts:      suite.timeouts,
		slowThreshold: slowThreshold,
	}

	suite.oesColonyID = esColonyModel.ID
	suite.scheduleID = scheduleModel.ID
}

// SetupTest 在每个测试方法运行前执行
func (suite *OesCronTestSuite) SetupTest() {
	suite.nextScheduleID = 0
	suite.NoError(suite.gormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&oesmodel.OesCronModel{}).Error)
}

// createTestOesCronModel 创建测试用的 OesCronModel
func (suite *OesCronTestSuite) createTestOesCronModel() *oesmodel.OesCronModel {
	suite.nextScheduleID++
	return &oesmodel.OesCronModel{
		OesColonyID: suite.oesColonyID,
		ScheduleID:  suite.nextScheduleID,
	}
}

// createTestOesCron 创建并保存测试用的 OesCronModel
func (suite *OesCronTestSuite) createTestOesCron() *oesmodel.OesCronModel {
	model := suite.createTestOesCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), model)
	suite.NoError(err, "创建测试 OesCron 应该成功")
	suite.NotZero(model.ID, "OesCron ID 应该不为零")
	return model
}

func (suite *OesCronTestSuite) TestCreateModel() {
	// 测试正常创建
	model := suite.createTestOesCron()
	suite.NotZero(model.ID, "OesCron ID 应该不为零")

	// 测试边界情况:创建空模型
	err := suite.cronRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空 OesCron 模型应该返回错误")
}

func (suite *OesCronTestSuite) TestUpdateModel() {
	// 创建测试数据
	model := suite.createTestOesCron()

	// 测试正常更新
	updateData := map[string]any{
		"oes_colony_id": 2,
		"schedule_id":   2,
	}
	err := suite.cronRepo.UpdateModel(context.Background(), updateData, "id = ?", model.ID)
	suite.NoError(err, "更新 OesCron 应该成功")

	// 验证更新结果
	updatedModel, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", model.ID)
	suite.NoError(err, "查询更新后的 OesCron 应该成功")
	suite.Equal(uint32(2), updatedModel.OesColonyID)
	suite.Equal(uint32(2), updatedModel.ScheduleID)

	// 测试边界情况:更新数据为空
	err = suite.cronRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", model.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况:更新不存在的 OesCron
	err = suite.cronRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的 OesCron 应该成功（无操作）")
}

func (suite *OesCronTestSuite) TestDeleteModel() {
	// 创建测试数据
	model := suite.createTestOesCron()

	// 测试正常删除
	err := suite.cronRepo.DeleteModel(context.Background(), "id = ?", model.ID)
	suite.NoError(err, "删除 OesCron 应该成功")

	// 验证删除结果
	deletedModel, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", model.ID)
	suite.Error(err, "查询已删除的 OesCron 应该返回错误")
	suite.Nil(deletedModel, "已删除的 OesCron 应该为 nil")

	// 测试边界情况:删除不存在的 OesCron
	err = suite.cronRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的 OesCron 应该成功（无操作）")
}

func (suite *OesCronTestSuite) TestGetModel() {
	// 创建测试数据
	model := suite.createTestOesCron()

	// 测试正常查询
	retrievedModel, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", model.ID)
	suite.NoError(err, "查询 OesCron 应该成功")
	suite.Equal(model.ID, retrievedModel.ID)
	suite.Equal(model.OesColonyID, retrievedModel.OesColonyID)
	suite.Equal(model.ScheduleID, retrievedModel.ScheduleID)

	// 测试边界情况:查询不存在的 OesCron
	retrievedModel, err = suite.cronRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的 OesCron 应该返回错误")
	suite.Nil(retrievedModel, "查询不存在的 OesCron 应该返回 nil")
}

func (suite *OesCronTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		suite.createTestOesCron()
	}

	// 测试正常查询列表
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Limit:   10,
		Offset:  0,
	}
	models, err := suite.cronRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询 OesCron 列表应该成功")
	suite.NotNil(models, "OesCron 列表应该不为 nil")
	suite.Greater(len(models), 0, "OesCron 列表长度应该大于 0")

	// 测试边界情况:空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"oes_colony_id": 999999},
	}
	models2, err := suite.cronRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的 OesCron 列表应该成功")
	suite.NotNil(models2, "不存在的 OesCron 列表应该不为 nil")
	suite.Len(models2, 0, "不存在的 OesCron 列表长度应该为 0")
}

func (suite *OesCronTestSuite) TestCountModel() {
	// 创建测试数据
	for i := 0; i < 3; i++ {
		suite.createTestOesCron()
	}

	// 测试正常计数
	count, err := suite.cronRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "查询 OesCron 总数应该成功")
	suite.Greater(count, int64(0), "OesCron 总数应该大于 0")

	// 测试带条件计数
	count2, err := suite.cronRepo.CountModel(context.Background(), map[string]any{"oes_colony_id": suite.oesColonyID})
	suite.NoError(err, "带条件查询 OesCron 总数应该成功")
	suite.GreaterOrEqual(count2, int64(3), "带条件的 OesCron 总数应该大于等于 3")

	// 测试边界情况:查询不存在的类型
	count3, err := suite.cronRepo.CountModel(context.Background(), map[string]any{"oes_colony_id": 999999})
	suite.NoError(err, "查询不存在类型的 OesCron 总数应该成功")
	suite.Equal(int64(0), count3, "不存在类型的 OesCron 总数应该为 0")
}

func (suite *OesCronTestSuite) TestContextTimeout() {
	// 创建测试数据
	model := suite.createTestOesCron()

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err := suite.cronRepo.GetModel(timeoutCtx, nil, "id = ?", model.ID)
	suite.Error(err, "上下文超时后查询 OesCron 应该返回错误")
}

func (suite *OesCronTestSuite) TestNewOesCronRepo() {
	slowThreshold := test.NewTestDBSlowThreshold()
	repo := NewOesCronRepo(suite.log, suite.gormDB, suite.timeouts, slowThreshold)
	suite.NotNil(repo, "NewOesCronRepo 应该返回非空实例")
	suite.Equal(suite.log, repo.log, "日志实例应该正确设置")
	suite.Equal(suite.gormDB, repo.gormDB, "数据库实例应该正确设置")
	suite.Equal(suite.timeouts, repo.timeouts, "超时设置应该正确设置")
	suite.Equal(slowThreshold, repo.slowThreshold, "慢查询阈值设置应该正确设置")
}

func TestOesCronTestSuite(t *testing.T) {
	suite.Run(t, &OesCronTestSuite{})
}
