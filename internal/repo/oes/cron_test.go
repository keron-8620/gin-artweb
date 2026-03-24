package oes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	oesmodel "gin-artweb/internal/model/oes"
	jobmodel "gin-artweb/internal/model/job"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestOesCronModel(oesColonyID, scheduleID uint32) *oesmodel.OesCronModel {
	return &oesmodel.OesCronModel{
		OesColonyID: oesColonyID,
		ScheduleID:  scheduleID,
	}
}

type OesCronTestSuite struct {
	suite.Suite
	cronRepo *OesCronRepo
	oesColonyID uint32
	scheduleID uint32
}

func (suite *OesCronTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&jobmodel.ScriptModel{}, &jobmodel.ScheduleModel{}, &oesmodel.OesColonyModel{}, &oesmodel.OesCronModel{})

	// 创建测试数据：Script
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

	// 创建测试数据：Schedule
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

	// 创建测试数据：OesColony
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

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.cronRepo = &OesCronRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}

	suite.oesColonyID = esColonyModel.ID
	suite.scheduleID = scheduleModel.ID
}

func (suite *OesCronTestSuite) TestCreateModel() {
	// 测试正常创建
	cm := CreateTestOesCronModel(suite.oesColonyID, suite.scheduleID)
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesCron应该成功")
	suite.NotZero(cm.ID, "OesCron ID应该不为零")

	// 测试边界情况：创建空模型
	err = suite.cronRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空OesCron模型应该返回错误")
}

func (suite *OesCronTestSuite) TestUpdateModel() {
	// 创建测试数据
	cm := CreateTestOesCronModel(suite.oesColonyID, suite.scheduleID)
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesCron用于更新测试应该成功")

	// 测试正常更新
	updateData := map[string]any{
		"oes_colony_id": 2,
		"schedule_id":   2,
	}
	err = suite.cronRepo.UpdateModel(context.Background(), updateData, "id = ?", cm.ID)
	suite.NoError(err, "更新OesCron应该成功")

	// 验证更新结果
	fm, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询更新后的OesCron应该成功")
	suite.Equal(uint32(2), fm.OesColonyID)
	suite.Equal(uint32(2), fm.ScheduleID)

	// 测试边界情况：更新数据为空
	err = suite.cronRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", cm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况：更新不存在的OesCron
	err = suite.cronRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的OesCron应该成功（无操作）")
}

func (suite *OesCronTestSuite) TestDeleteModel() {
	// 创建测试数据
	cm := CreateTestOesCronModel(suite.oesColonyID, suite.scheduleID)
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesCron用于删除测试应该成功")

	// 测试正常删除
	err = suite.cronRepo.DeleteModel(context.Background(), "id = ?", cm.ID)
	suite.NoError(err, "删除OesCron应该成功")

	// 验证删除结果
	fm, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.Error(err, "查询已删除的OesCron应该返回错误")
	suite.Nil(fm, "已删除的OesCron应该为nil")

	// 测试边界情况：删除不存在的OesCron
	err = suite.cronRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的OesCron应该成功（无操作）")
}

func (suite *OesCronTestSuite) TestGetModel() {
	// 创建测试数据
	cm := CreateTestOesCronModel(suite.oesColonyID, suite.scheduleID)
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesCron用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询OesCron应该成功")
	suite.Equal(cm.ID, fm.ID)
	suite.Equal(cm.OesColonyID, fm.OesColonyID)
	suite.Equal(cm.ScheduleID, fm.ScheduleID)

	// 测试边界情况：查询不存在的OesCron
	fm, err = suite.cronRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的OesCron应该返回错误")
	suite.Nil(fm, "查询不存在的OesCron应该返回nil")
}

func (suite *OesCronTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		cm := CreateTestOesCronModel(suite.oesColonyID, suite.scheduleID)
		err := suite.cronRepo.CreateModel(context.Background(), cm)
		suite.NoError(err, "创建OesCron用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Limit:   10,
		Offset:  0,
	}
	models, err := suite.cronRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询OesCron列表应该成功")
	suite.Greater(int64(len(models)), int64(0), "OesCron列表数量应该大于0")
	suite.NotNil(models, "OesCron列表应该不为nil")
	suite.Greater(len(models), 0, "OesCron列表长度应该大于0")

	// 测试边界情况：空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"oes_colony_id": 999999},
	}
	models2, err := suite.cronRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的OesCron列表应该成功")
	suite.Equal(int64(0), int64(len(models2)), "不存在的OesCron列表数量应该为0")
	suite.NotNil(models2, "不存在的OesCron列表应该不为nil")
	suite.Len(models2, 0, "不存在的OesCron列表长度应该为0")
}

func (suite *OesCronTestSuite) TestCountModel() {
	// 创建测试数据
	for i := 0; i < 3; i++ {
		cm := CreateTestOesCronModel(suite.oesColonyID, suite.scheduleID)
		err := suite.cronRepo.CreateModel(context.Background(), cm)
		suite.NoError(err, "创建OesCron用于计数测试应该成功")
	}

	// 测试正常计数
	count, err := suite.cronRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "查询OesCron总数应该成功")
	suite.Greater(count, int64(0), "OesCron总数应该大于0")

	// 测试带条件计数
	count2, err := suite.cronRepo.CountModel(context.Background(), map[string]any{"oes_colony_id": suite.oesColonyID})
	suite.NoError(err, "带条件查询OesCron总数应该成功")
	suite.GreaterOrEqual(count2, int64(3), "带条件的OesCron总数应该大于等于3")

	// 测试边界情况：查询不存在的类型
	count3, err := suite.cronRepo.CountModel(context.Background(), map[string]any{"oes_colony_id": 999999})
	suite.NoError(err, "查询不存在类型的OesCron总数应该成功")
	suite.Equal(int64(0), count3, "不存在类型的OesCron总数应该为0")
}

func (suite *OesCronTestSuite) TestContextTimeout() {
	// 创建测试数据
	cm := CreateTestOesCronModel(suite.oesColonyID, suite.scheduleID)
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesCron用于超时测试应该成功")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err = suite.cronRepo.GetModel(timeoutCtx, nil, "id = ?", cm.ID)
	suite.Error(err, "上下文超时后查询OesCron应该返回错误")
}

func TestOesCronTestSuite(t *testing.T) {
	pts := &OesCronTestSuite{}
	suite.Run(t, pts)
}
