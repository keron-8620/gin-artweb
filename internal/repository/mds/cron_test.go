package mds

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	mdsmodel "gin-artweb/internal/model/mds"
	jobsmodel "gin-artweb/internal/model/jobs"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

var cronCounter int = 0

func CreateTestMdsCronModel() *mdsmodel.MdsCronModel {
	cronCounter++
	return &mdsmodel.MdsCronModel{
		TaskName:    fmt.Sprintf("test-task-%s", uuid.NewString()),
		MdsColonyID: 1, // 假设集群ID为1
		ScheduleID:  uint32(cronCounter), // 使用cronCounter作为schedule_id，确保唯一性
	}
}

type MdsCronTestSuite struct {
	suite.Suite
	cronRepo *MdsCronRepo
}

func (suite *MdsCronTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&mdsmodel.MdsColonyModel{}, &jobsmodel.ScriptModel{}, &jobsmodel.ScheduleModel{}, &mdsmodel.MdsCronModel{})

	// 创建测试数据：MdsColony
	colonyModel := &mdsmodel.MdsColonyModel{
		ColonyNum:     "01",
		ExtractedName: "test-colony",
		IsEnable:      true,
		PackageID:     1,
		MonNodeID:     1,
	}
	db.Create(colonyModel)

	// 创建测试数据：Script
	scriptModel := &jobsmodel.ScriptModel{
		Name:      "test-script",
		Descr:     "test script",
		Project:   "test",
		Label:     "test",
		Language:  "bash",
		Status:    true,
		IsBuiltin: false,
		Username:  "admin",
	}
	db.Create(scriptModel)

	// 创建多个测试数据：Schedule
	for i := 1; i <= 10; i++ {
		scheduleModel := &jobsmodel.ScheduleModel{
			Name:          fmt.Sprintf("test-schedule-%d", i),
			Specification: "0 0 * * *",
			IsEnabled:     true,
			CommandArgs:   "echo test",
			ScriptID:      1,
		}
		db.Create(scheduleModel)
	}

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.cronRepo = &MdsCronRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *MdsCronTestSuite) TestCreateModel() {
	// 测试正常创建
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron应该成功")
	suite.NotZero(cm.ID, "MdsCron ID应该不为零")

	// 测试边界情况：创建空模型
	err = suite.cronRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空MdsCron模型应该返回错误")
}

func (suite *MdsCronTestSuite) TestUpdateModel() {
	// 创建测试数据
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron用于更新测试应该成功")

	// 测试正常更新 - 只更新TaskName，不更新ScheduleID以避免唯一约束冲突
	updateData := map[string]any{
		"TaskName":    "updated-task",
		"MdsColonyID": 1,
	}
	err = suite.cronRepo.UpdateModel(context.Background(), updateData, "id = ?", cm.ID)
	suite.NoError(err, "更新MdsCron应该成功")

	// 验证更新结果
	fm, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询更新后的MdsCron应该成功")
	suite.Equal(cm.ID, fm.ID)
	suite.Equal("updated-task", fm.TaskName)
	suite.Equal(uint32(1), fm.MdsColonyID)
	suite.Equal(cm.ScheduleID, fm.ScheduleID)

	// 测试边界情况：更新数据为空
	err = suite.cronRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", cm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况：更新不存在的MdsCron
	err = suite.cronRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的MdsCron应该成功（无操作）")
}

func (suite *MdsCronTestSuite) TestDeleteModel() {
	// 创建测试数据
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron用于删除测试应该成功")

	// 测试正常删除
	err = suite.cronRepo.DeleteModel(context.Background(), "id = ?", cm.ID)
	suite.NoError(err, "删除MdsCron应该成功")

	// 验证删除结果
	fm, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.Error(err, "查询已删除的MdsCron应该返回错误")
	suite.Nil(fm, "已删除的MdsCron应该为nil")

	// 测试边界情况：删除不存在的MdsCron
	err = suite.cronRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的MdsCron应该成功（无操作）")
}

func (suite *MdsCronTestSuite) TestGetModel() {
	// 创建测试数据
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询MdsCron应该成功")
	suite.Equal(cm.ID, fm.ID)
	suite.Equal(cm.TaskName, fm.TaskName)
	suite.Equal(cm.MdsColonyID, fm.MdsColonyID)
	suite.Equal(cm.ScheduleID, fm.ScheduleID)

	// 测试边界情况：查询不存在的MdsCron
	fm, err = suite.cronRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的MdsCron应该返回错误")
	suite.Nil(fm, "查询不存在的MdsCron应该返回nil")

	// 测试边界情况：使用预加载
	fm, err = suite.cronRepo.GetModel(context.Background(), []string{}, "id = ?", cm.ID)
	suite.NoError(err, "使用预加载查询MdsCron应该成功")
	suite.Equal(cm.ID, fm.ID)
}

func (suite *MdsCronTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		cm := CreateTestMdsCronModel()
		err := suite.cronRepo.CreateModel(context.Background(), cm)
		suite.NoError(err, "创建MdsCron用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	count, models, err := suite.cronRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询MdsCron列表应该成功")
	suite.Greater(count, int64(0), "MdsCron列表数量应该大于0")
	suite.NotNil(models, "MdsCron列表应该不为nil")
	suite.Greater(len(*models), 0, "MdsCron列表长度应该大于0")

	// 测试边界情况：空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"task_name": "non-existent-task"},
	}
	count2, models2, err := suite.cronRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的MdsCron列表应该成功")
	suite.Equal(int64(0), count2, "不存在的MdsCron列表数量应该为0")
	suite.NotNil(models2, "不存在的MdsCron列表应该不为nil")
	suite.Len(*models2, 0, "不存在的MdsCron列表长度应该为0")
}

func (suite *MdsCronTestSuite) TestContextTimeout() {
	// 创建测试数据
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron用于超时测试应该成功")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err = suite.cronRepo.GetModel(timeoutCtx, nil, "id = ?", cm.ID)
	suite.Error(err, "上下文超时后查询MdsCron应该返回错误")
}

func TestMdsCronTestSuite(t *testing.T) {
	pts := &MdsCronTestSuite{}
	suite.Run(t, pts)
}
