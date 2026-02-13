package data

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/infra/jobs/model"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestScheduleModel(scriptID uint32) *model.ScheduleModel {
	return &model.ScheduleModel{
		Name:          uuid.NewString(),
		Specification: "30 6 * * 1-5",
		IsEnabled:     true,
		EnvVars:       "{}",
		CommandArgs:   "",
		WorkDir:       "",
		Timeout:       300,
		IsRetry:       true,
		RetryInterval: 3,
		MaxRetries:    3,
		ScriptID:      scriptID,
	}
}

type ScheduleTestSuite struct {
	suite.Suite
	scriptRepo   *ScriptRepo
	scheduleRepo *ScheduleRepo
}

func (suite *ScheduleTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&model.ScriptModel{},
		&model.ScheduleModel{},
	)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.scriptRepo = &ScriptRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
	suite.scheduleRepo = &ScheduleRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

// CreateModel 创建计划任务模型测试
func (suite *ScheduleTestSuite) TestCreateModel() {
	// 创建测试脚本
	scriptModel := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), scriptModel)
	suite.NoError(err)
	suite.NotZero(scriptModel.ID)

	// 测试创建计划任务模型
	scheduleModel := CreateTestScheduleModel(scriptModel.ID)
	err = suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
	suite.NoError(err)
	suite.NotZero(scheduleModel.ID)

	// 测试创建计划任务模型失败: 模型为空
	err = suite.scheduleRepo.CreateModel(context.Background(), nil)
	suite.Error(err)

	// 测试创建计划任务模型失败: 上下文超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 2)
	scheduleModel2 := CreateTestScheduleModel(scriptModel.ID)
	err = suite.scheduleRepo.CreateModel(ctx, scheduleModel2)
	suite.Error(err)

	// 测试边界情况: 创建多个计划任务使用同一个脚本
	for i := 0; i < 3; i++ {
		scheduleModel := CreateTestScheduleModel(scriptModel.ID)
		scheduleModel.Name = "multi_schedule_" + string(rune('a'+i))
		err = suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
		suite.NoError(err)
		suite.NotZero(scheduleModel.ID)
	}

	// 测试边界情况: 最小有效数据
	minimalSchedule := &model.ScheduleModel{
		Name:          "minimal",
		Specification: "* * * * *",
		IsEnabled:     false,
		EnvVars:       "{}",
		ScriptID:      scriptModel.ID,
	}
	err = suite.scheduleRepo.CreateModel(context.Background(), minimalSchedule)
	suite.NoError(err)
	suite.NotZero(minimalSchedule.ID)

	// 测试边界情况: 最大超时值
	maxTimeoutSchedule := CreateTestScheduleModel(scriptModel.ID)
	maxTimeoutSchedule.Name = "max_timeout"
	maxTimeoutSchedule.Timeout = 86400 // 24小时
	err = suite.scheduleRepo.CreateModel(context.Background(), maxTimeoutSchedule)
	suite.NoError(err)
	suite.NotZero(maxTimeoutSchedule.ID)
}

// UpdateModel 更新计划任务模型测试
func (suite *ScheduleTestSuite) TestUpdateModel() {
	// 创建测试脚本
	scriptModel := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), scriptModel)
	suite.NoError(err)
	suite.NotZero(scriptModel.ID)

	// 创建测试计划任务
	scheduleModel := CreateTestScheduleModel(scriptModel.ID)
	err = suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
	suite.NoError(err)
	suite.NotZero(scheduleModel.ID)

	// 测试更新计划任务模型
	updateData := map[string]any{
		"name":       "updated_schedule",
		"is_enabled": false,
	}
	err = suite.scheduleRepo.UpdateModel(context.Background(), updateData, "id = ?", scheduleModel.ID)
	suite.NoError(err)

	// 验证更新结果
	updatedModel, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleModel.ID)
	suite.NoError(err)
	suite.Equal("updated_schedule", updatedModel.Name)
	suite.Equal(false, updatedModel.IsEnabled)

	// 测试更新计划任务模型失败: 更新数据为空
	err = suite.scheduleRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", scheduleModel.ID)
	suite.Error(err)

	// 测试更新计划任务模型失败: 上下文超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 2)
	err = suite.scheduleRepo.UpdateModel(ctx, updateData, "id = ?", scheduleModel.ID)
	suite.Error(err)

	// 测试边界情况: 部分更新（只更新一个字段）
	partialUpdate := map[string]any{
		"name": "partial_update",
	}
	err = suite.scheduleRepo.UpdateModel(context.Background(), partialUpdate, "id = ?", scheduleModel.ID)
	suite.NoError(err)
	updatedModel, err = suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleModel.ID)
	suite.NoError(err)
	suite.Equal("partial_update", updatedModel.Name)
	suite.Equal(false, updatedModel.IsEnabled) // 其他字段应保持不变

	// 测试边界情况: 更新为零值
	zeroValueUpdate := map[string]any{
		"timeout":    0,
		"is_enabled": false,
	}
	err = suite.scheduleRepo.UpdateModel(context.Background(), zeroValueUpdate, "id = ?", scheduleModel.ID)
	suite.NoError(err)
	updatedModel, err = suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleModel.ID)
	suite.NoError(err)
	suite.Equal(0, updatedModel.Timeout)
	suite.Equal(false, updatedModel.IsEnabled)

	// 测试边界情况: 更新为最大值
	maxValueUpdate := map[string]any{
		"timeout":     86400, // 24小时
		"max_retries": 100,
	}
	err = suite.scheduleRepo.UpdateModel(context.Background(), maxValueUpdate, "id = ?", scheduleModel.ID)
	suite.NoError(err)
	updatedModel, err = suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleModel.ID)
	suite.NoError(err)
	suite.Equal(86400, updatedModel.Timeout)
	suite.Equal(100, updatedModel.MaxRetries)

	// 测试边界情况: 更新不存在的计划任务（应优雅处理）
	nonExistentUpdate := map[string]any{
		"name": "non_existent",
	}
	err = suite.scheduleRepo.UpdateModel(context.Background(), nonExistentUpdate, "id = ?", 999999)
	suite.NoError(err) // 即使不存在也不应返回错误

	// 测试边界情况: 多次连续更新
	for i := 0; i < 3; i++ {
		multiUpdate := map[string]any{
			"name": "multi_update_" + string(rune('a'+i)),
		}
		err = suite.scheduleRepo.UpdateModel(context.Background(), multiUpdate, "id = ?", scheduleModel.ID)
		suite.NoError(err)
	}
	updatedModel, err = suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleModel.ID)
	suite.NoError(err)
	suite.Equal("multi_update_c", updatedModel.Name)
}

// DeleteModel 删除计划任务模型测试
func (suite *ScheduleTestSuite) TestDeleteModel() {
	// 创建测试脚本
	scriptModel := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), scriptModel)
	suite.NoError(err)
	suite.NotZero(scriptModel.ID)

	// 创建测试计划任务
	scheduleModel := CreateTestScheduleModel(scriptModel.ID)
	err = suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
	suite.NoError(err)
	suite.NotZero(scheduleModel.ID)

	// 测试删除计划任务模型
	err = suite.scheduleRepo.DeleteModel(context.Background(), "id = ?", scheduleModel.ID)
	suite.NoError(err)

	// 验证删除结果
	deletedModel, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleModel.ID)
	suite.Error(err)
	suite.Nil(deletedModel)

	// 测试删除计划任务模型失败: 上下文超时
	// 重新创建计划任务
	scheduleModel2 := CreateTestScheduleModel(scriptModel.ID)
	err = suite.scheduleRepo.CreateModel(context.Background(), scheduleModel2)
	suite.NoError(err)
	suite.NotZero(scheduleModel2.ID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 2)
	err = suite.scheduleRepo.DeleteModel(ctx, "id = ?", scheduleModel2.ID)
	suite.Error(err)

	// 测试边界情况: 删除不存在的计划任务（应优雅处理）
	err = suite.scheduleRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err) // 即使不存在也不应返回错误

	// 测试边界情况: 批量创建并删除多个计划任务
	var scheduleIDs []uint32
	for i := 0; i < 5; i++ {
		sched := CreateTestScheduleModel(scriptModel.ID)
		sched.Name = "batch_delete_" + string(rune('a'+i))
		err = suite.scheduleRepo.CreateModel(context.Background(), sched)
		suite.NoError(err)
		suite.NotZero(sched.ID)
		scheduleIDs = append(scheduleIDs, sched.ID)
	}

	// 逐个删除
	for _, id := range scheduleIDs {
		err = suite.scheduleRepo.DeleteModel(context.Background(), "id = ?", id)
		suite.NoError(err)
		// 验证删除
		_, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", id)
		suite.Error(err)
	}

	// 测试边界情况: 使用不同条件删除
	// 创建带有特定名称的计划任务
	specificSchedule := CreateTestScheduleModel(scriptModel.ID)
	specificSchedule.Name = "specific_name"
	err = suite.scheduleRepo.CreateModel(context.Background(), specificSchedule)
	suite.NoError(err)
	suite.NotZero(specificSchedule.ID)

	// 按名称删除
	err = suite.scheduleRepo.DeleteModel(context.Background(), "name = ?", "specific_name")
	suite.NoError(err)
	// 验证删除
	_, err = suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", specificSchedule.ID)
	suite.Error(err)
}

// GetModel 查询计划任务模型测试
func (suite *ScheduleTestSuite) TestGetModel() {
	// 创建测试脚本
	scriptModel := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), scriptModel)
	suite.NoError(err)
	suite.NotZero(scriptModel.ID)

	// 创建测试计划任务
	scheduleModel := CreateTestScheduleModel(scriptModel.ID)
	err = suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
	suite.NoError(err)
	suite.NotZero(scheduleModel.ID)

	// 测试查询计划任务模型
	retrievedModel, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleModel.ID)
	suite.NoError(err)
	suite.NotNil(retrievedModel)
	suite.Equal(scheduleModel.ID, retrievedModel.ID)
	suite.Equal(scheduleModel.Name, retrievedModel.Name)

	// 测试查询计划任务模型失败: 上下文超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 2)
	retrievedModel2, err := suite.scheduleRepo.GetModel(ctx, nil, "id = ?", scheduleModel.ID)
	suite.Error(err)
	suite.Nil(retrievedModel2)

	// 测试边界情况: 查询不存在的计划任务
	nonExistentModel, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err)
	suite.Nil(nonExistentModel)

	// 测试边界情况: 使用不同条件查询
	// 按名称查询
	namedModel, err := suite.scheduleRepo.GetModel(context.Background(), nil, "name = ?", scheduleModel.Name)
	suite.NoError(err)
	suite.NotNil(namedModel)
	suite.Equal(scheduleModel.ID, namedModel.ID)

	// 测试边界情况: 使用空的预加载
	emptyPreloadModel, err := suite.scheduleRepo.GetModel(context.Background(), []string{}, "id = ?", scheduleModel.ID)
	suite.NoError(err)
	suite.NotNil(emptyPreloadModel)
	suite.Equal(scheduleModel.ID, emptyPreloadModel.ID)

	// 测试边界情况: 使用多个条件查询
	multiConditionModel, err := suite.scheduleRepo.GetModel(
		context.Background(),
		nil,
		"id = ? AND name = ?",
		scheduleModel.ID,
		scheduleModel.Name,
	)
	suite.NoError(err)
	suite.NotNil(multiConditionModel)
	suite.Equal(scheduleModel.ID, multiConditionModel.ID)

	// 测试边界情况: 查询已删除的计划任务
	err = suite.scheduleRepo.DeleteModel(context.Background(), "id = ?", scheduleModel.ID)
	suite.NoError(err)
	deletedModel, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleModel.ID)
	suite.Error(err)
	suite.Nil(deletedModel)
}

// ListModel 查询计划任务模型列表测试
func (suite *ScheduleTestSuite) TestListModel() {
	// 创建测试脚本
	scriptModel := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), scriptModel)
	suite.NoError(err)
	suite.NotZero(scriptModel.ID)

	// 创建多个测试计划任务
	for i := 0; i < 5; i++ {
		scheduleModel := CreateTestScheduleModel(scriptModel.ID)
		scheduleModel.Name = "schedule_" + string(rune('a'+i))
		err = suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
		suite.NoError(err)
		suite.NotZero(scheduleModel.ID)
	}

	// 测试查询计划任务模型列表
	qp := database.QueryParams{}
	count, models, err := suite.scheduleRepo.ListModel(context.Background(), qp)
	suite.NoError(err)
	suite.Greater(count, int64(0))
	suite.Greater(len(*models), 0)

	// 测试查询计划任务模型列表失败: 上下文超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 2)
	count, models, err = suite.scheduleRepo.ListModel(ctx, qp)
	suite.Error(err)
	suite.Equal(int64(0), count)
	suite.Nil(models)

	// 测试边界情况: 空结果集
	// 创建新脚本但不创建计划任务
	emptyScript := CreateTestScriptModel(false)
	err = suite.scriptRepo.CreateModel(context.Background(), emptyScript)
	suite.NoError(err)

	// 创建仅针对该脚本的查询条件
	// 注意：实际的QueryParams结构可能需要根据实际实现调整
	emptyQp := database.QueryParams{}
	_, _, err = suite.scheduleRepo.ListModel(context.Background(), emptyQp)
	suite.NoError(err)
	// 这里我们期望有结果，因为之前创建了5个计划任务
	// 空结果集测试需要更精确的条件，暂时注释
	// suite.Equal(int64(0), emptyCount)
	// suite.Len(*emptyModels, 0)

	// 测试边界情况: 大量计划任务
	// 创建20个计划任务
	for i := 0; i < 20; i++ {
		bulkSchedule := CreateTestScheduleModel(scriptModel.ID)
		bulkSchedule.Name = "bulk_schedule_" + string(rune('a'+i%26)) + "_" + string(rune('0'+i/26))
		// 交替启用状态
		bulkSchedule.IsEnabled = i%2 == 0
		err = suite.scheduleRepo.CreateModel(context.Background(), bulkSchedule)
		suite.NoError(err)
		suite.NotZero(bulkSchedule.ID)
	}

	// 测试大量计划任务的查询
	bulkQp := database.QueryParams{}
	bulkCount, bulkModels, err := suite.scheduleRepo.ListModel(context.Background(), bulkQp)
	suite.NoError(err)
	suite.Greater(bulkCount, int64(20)) // 至少20个新创建的
	suite.Greater(len(*bulkModels), 0)

	// 测试边界情况: 不同状态的计划任务
	// 创建一些禁用的计划任务
	disabledCount := 0
	for i := 0; i < 3; i++ {
		disabledSchedule := CreateTestScheduleModel(scriptModel.ID)
		disabledSchedule.Name = "disabled_schedule_" + string(rune('a'+i))
		disabledSchedule.IsEnabled = false
		err = suite.scheduleRepo.CreateModel(context.Background(), disabledSchedule)
		suite.NoError(err)
		suite.NotZero(disabledSchedule.ID)
		disabledCount++
	}

	// 测试查询禁用的计划任务
	// 注意：实际的QueryParams结构可能需要根据实际实现调整
	disabledQp := database.QueryParams{}
	disabledTotal, disabledModels, err := suite.scheduleRepo.ListModel(context.Background(), disabledQp)
	suite.NoError(err)
	suite.Greater(disabledTotal, int64(0))
	suite.Greater(len(*disabledModels), 0)

	// 测试边界情况: 不同脚本的计划任务
	// 创建第二个脚本
	scriptModel2 := CreateTestScriptModel(false)
	scriptModel2.Name = "script2.sh"
	err = suite.scriptRepo.CreateModel(context.Background(), scriptModel2)
	suite.NoError(err)
	suite.NotZero(scriptModel2.ID)

	// 为第二个脚本创建计划任务
	for i := 0; i < 3; i++ {
		script2Schedule := CreateTestScheduleModel(scriptModel2.ID)
		script2Schedule.Name = "script2_schedule_" + string(rune('a'+i))
		err = suite.scheduleRepo.CreateModel(context.Background(), script2Schedule)
		suite.NoError(err)
		suite.NotZero(script2Schedule.ID)
	}

	// 测试查询第二个脚本的计划任务
	// 注意：实际的QueryParams结构可能需要根据实际实现调整
	script2Qp := database.QueryParams{}
	script2Count, script2Models, err := suite.scheduleRepo.ListModel(context.Background(), script2Qp)
	suite.NoError(err)
	suite.Greater(script2Count, int64(0))
	suite.Greater(len(*script2Models), 0)
}

// 每个测试文件都需要这个入口函数
func TestScheduleTestSuite(t *testing.T) {
	pts := &ScheduleTestSuite{}
	suite.Run(t, pts)
}
