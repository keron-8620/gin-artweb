package job

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	jobmodel "gin-artweb/internal/model/job"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

// CreateTestScheduleModel 创建测试计划任务模型
// 参数：
// - scriptID: 脚本ID
// - options: 可选配置，用于覆盖默认值
func CreateTestScheduleModel(scriptID uint32, options ...func(*jobmodel.ScheduleModel)) *jobmodel.ScheduleModel {
	model := &jobmodel.ScheduleModel{
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

	// 应用可选配置
	for _, option := range options {
		option(model)
	}

	return model
}

// WithScheduleName 设置计划任务名称
func WithScheduleName(name string) func(*jobmodel.ScheduleModel) {
	return func(model *jobmodel.ScheduleModel) {
		model.Name = name
	}
}

// WithScheduleEnabled 设置计划任务启用状态
func WithScheduleEnabled(enabled bool) func(*jobmodel.ScheduleModel) {
	return func(model *jobmodel.ScheduleModel) {
		model.IsEnabled = enabled
	}
}

// WithScheduleTimeout 设置计划任务超时时间
func WithScheduleTimeout(timeout int) func(*jobmodel.ScheduleModel) {
	return func(model *jobmodel.ScheduleModel) {
		model.Timeout = timeout
	}
}

// WithScheduleRetry 设置计划任务重试配置
func WithScheduleRetry(isRetry bool, interval, maxRetries int) func(*jobmodel.ScheduleModel) {
	return func(model *jobmodel.ScheduleModel) {
		model.IsRetry = isRetry
		model.RetryInterval = interval
		model.MaxRetries = maxRetries
	}
}

// WithScheduleSpec 设置计划任务调度表达式
func WithScheduleSpec(spec string) func(*jobmodel.ScheduleModel) {
	return func(model *jobmodel.ScheduleModel) {
		model.Specification = spec
	}
}

type ScheduleTestSuite struct {
	suite.Suite
	scriptRepo   *ScriptRepo
	scheduleRepo *ScheduleRepo
	testScriptID uint32
}

func (suite *ScheduleTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(
		&jobmodel.ScriptModel{},
		&jobmodel.ScheduleModel{},
	); err != nil {
		suite.Error(err, "数据库迁移失败")
	}
	dbTimeout := test.NewTestDBTimeouts()
	dbSlowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()
	suite.scriptRepo = &ScriptRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: dbSlowThreshold,
	}
	suite.scheduleRepo = &ScheduleRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: dbSlowThreshold,
	}

	// 创建测试脚本用于所有测试
	scriptModel := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), scriptModel)
	suite.NoError(err)
	suite.NotZero(scriptModel.ID)
	suite.testScriptID = scriptModel.ID
}

// createTestSchedule 创建测试计划任务并返回ID
func (suite *ScheduleTestSuite) createTestSchedule(options ...func(*jobmodel.ScheduleModel)) uint32 {
	scheduleModel := CreateTestScheduleModel(suite.testScriptID, options...)
	err := suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
	suite.NoError(err)
	suite.NotZero(scheduleModel.ID)
	return scheduleModel.ID
}

// getScheduleByID 根据ID获取计划任务
func (suite *ScheduleTestSuite) getScheduleByID(scheduleID uint32) *jobmodel.ScheduleModel {
	model, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleID)
	suite.NoError(err, "获取计划任务失败，ID: %d", scheduleID)
	suite.NotNil(model, "计划任务不存在，ID: %d", scheduleID)
	return model
}

// assertScheduleDeleted 断言计划任务已被删除
func (suite *ScheduleTestSuite) assertScheduleDeleted(scheduleID uint32) {
	model, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleID)
	suite.Error(err, "计划任务应该已被删除，ID: %d", scheduleID)
	suite.Nil(model, "计划任务不应该存在，ID: %d", scheduleID)
}

// testContextTimeout 测试上下文超时场景
func (suite *ScheduleTestSuite) testContextTimeout(action func(ctx context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	time.Sleep(time.Millisecond * 2)
	err := action(ctx)
	suite.Error(err, "操作应该在上下文超时时失败")
}

// TestCreateModel 创建计划任务模型测试
func (suite *ScheduleTestSuite) TestCreateModel() {
	// 测试创建计划任务模型
	scheduleModel := CreateTestScheduleModel(suite.testScriptID)
	err := suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
	suite.NoError(err, "创建计划任务失败")
	suite.NotZero(scheduleModel.ID, "计划任务ID应该不为零")

	// 测试创建计划任务模型失败: 模型为空
	err = suite.scheduleRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建计划任务时模型为空应该失败")

	// 测试创建计划任务模型失败: 上下文超时
	suite.testContextTimeout(func(ctx context.Context) error {
		scheduleModel2 := CreateTestScheduleModel(suite.testScriptID)
		return suite.scheduleRepo.CreateModel(ctx, scheduleModel2)
	})

	// 测试边界情况: 创建多个计划任务使用同一个脚本
	for i := range 3 {
		scheduleModel := CreateTestScheduleModel(
			suite.testScriptID,
			WithScheduleName(fmt.Sprintf("multi_schedule_%c", 'a'+i)),
		)
		err = suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
		suite.NoError(err)
		suite.NotZero(scheduleModel.ID)
	}

	// 测试边界情况: 最小有效数据
	minimalSchedule := &jobmodel.ScheduleModel{
		Name:          "minimal",
		Specification: "* * * * *",
		IsEnabled:     false,
		EnvVars:       "{}",
		ScriptID:      suite.testScriptID,
	}
	err = suite.scheduleRepo.CreateModel(context.Background(), minimalSchedule)
	suite.NoError(err)
	suite.NotZero(minimalSchedule.ID)

	// 测试边界情况: 最大超时值
	maxTimeoutSchedule := CreateTestScheduleModel(
		suite.testScriptID,
		WithScheduleName("max_timeout"),
		WithScheduleTimeout(86400), // 24小时
	)
	err = suite.scheduleRepo.CreateModel(context.Background(), maxTimeoutSchedule)
	suite.NoError(err)
	suite.NotZero(maxTimeoutSchedule.ID)
}

// TestUpdateModel 更新计划任务模型测试
func (suite *ScheduleTestSuite) TestUpdateModel() {
	// 创建测试计划任务
	scheduleID := suite.createTestSchedule()

	// 测试更新计划任务模型
	updateData := map[string]any{
		"name":       "updated_schedule",
		"is_enabled": false,
	}
	err := suite.scheduleRepo.UpdateModel(context.Background(), updateData, "id = ?", scheduleID)
	suite.NoError(err, "更新计划任务失败，ID: %d", scheduleID)

	// 验证更新结果
	updatedModel := suite.getScheduleByID(scheduleID)
	suite.Equal("updated_schedule", updatedModel.Name, "计划任务名称应该更新为 'updated_schedule'")
	suite.Equal(false, updatedModel.IsEnabled, "计划任务应该被禁用")

	// 测试更新计划任务模型失败: 更新数据为空
	err = suite.scheduleRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", scheduleID)
	suite.Error(err, "更新计划任务时数据为空应该失败")

	// 测试更新计划任务模型失败: 上下文超时
	suite.testContextTimeout(func(ctx context.Context) error {
		return suite.scheduleRepo.UpdateModel(ctx, updateData, "id = ?", scheduleID)
	})

	// 测试边界情况: 部分更新（只更新一个字段）
	partialUpdate := map[string]any{
		"name": "partial_update",
	}
	err = suite.scheduleRepo.UpdateModel(context.Background(), partialUpdate, "id = ?", scheduleID)
	suite.NoError(err)
	updatedModel, err = suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleID)
	suite.NoError(err)
	suite.Equal("partial_update", updatedModel.Name)
	suite.Equal(false, updatedModel.IsEnabled) // 其他字段应保持不变

	// 测试边界情况: 更新为零值
	zeroValueUpdate := map[string]any{
		"timeout":    0,
		"is_enabled": false,
	}
	err = suite.scheduleRepo.UpdateModel(context.Background(), zeroValueUpdate, "id = ?", scheduleID)
	suite.NoError(err)
	updatedModel, err = suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleID)
	suite.NoError(err)
	suite.Equal(0, updatedModel.Timeout)
	suite.Equal(false, updatedModel.IsEnabled)

	// 测试边界情况: 更新为最大值
	maxValueUpdate := map[string]any{
		"timeout":     86400, // 24小时
		"max_retries": 100,
	}
	err = suite.scheduleRepo.UpdateModel(context.Background(), maxValueUpdate, "id = ?", scheduleID)
	suite.NoError(err)
	updatedModel, err = suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleID)
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
	for i := range 3 {
		multiUpdate := map[string]any{
			"name": fmt.Sprintf("multi_update_%c", 'a'+i),
		}
		err = suite.scheduleRepo.UpdateModel(context.Background(), multiUpdate, "id = ?", scheduleID)
		suite.NoError(err)
	}
	updatedModel, err = suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleID)
	suite.NoError(err)
	suite.Equal("multi_update_c", updatedModel.Name)
}

// TestDeleteModel 删除计划任务模型测试
func (suite *ScheduleTestSuite) TestDeleteModel() {
	// 创建测试计划任务
	scheduleID := suite.createTestSchedule()

	// 测试删除计划任务模型
	err := suite.scheduleRepo.DeleteModel(context.Background(), "id = ?", scheduleID)
	suite.NoError(err, "删除计划任务失败，ID: %d", scheduleID)

	// 验证删除结果
	suite.assertScheduleDeleted(scheduleID)

	// 测试删除计划任务模型失败: 上下文超时
	// 重新创建计划任务
	scheduleID2 := suite.createTestSchedule()

	suite.testContextTimeout(func(ctx context.Context) error {
		return suite.scheduleRepo.DeleteModel(ctx, "id = ?", scheduleID2)
	})

	// 测试边界情况: 删除不存在的计划任务（应优雅处理）
	err = suite.scheduleRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err) // 即使不存在也不应返回错误

	// 测试边界情况: 批量创建并删除多个计划任务
	var scheduleIDs []uint32
	for i := 0; i < 5; i++ {
		sched := CreateTestScheduleModel(
			suite.testScriptID,
			WithScheduleName(fmt.Sprintf("batch_delete_%c", 'a'+i)),
		)
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
	specificSchedule := CreateTestScheduleModel(
		suite.testScriptID,
		WithScheduleName("specific_name"),
	)
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

// TestGetModel 查询计划任务模型测试
func (suite *ScheduleTestSuite) TestGetModel() {
	// 创建测试计划任务
	scheduleID := suite.createTestSchedule()
	scheduleModel := suite.getScheduleByID(scheduleID)

	// 测试查询计划任务模型
	retrievedModel, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleID)
	suite.NoError(err)
	suite.NotNil(retrievedModel)
	suite.Equal(scheduleID, retrievedModel.ID)
	suite.Equal(scheduleModel.Name, retrievedModel.Name)

	// 测试查询计划任务模型失败: 上下文超时
	suite.testContextTimeout(func(ctx context.Context) error {
		_, err := suite.scheduleRepo.GetModel(ctx, nil, "id = ?", scheduleID)
		return err
	})

	// 测试边界情况: 查询不存在的计划任务
	nonExistentModel, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err)
	suite.Nil(nonExistentModel)

	// 测试边界情况: 使用不同条件查询
	// 按名称查询
	namedModel, err := suite.scheduleRepo.GetModel(context.Background(), nil, "name = ?", scheduleModel.Name)
	suite.NoError(err)
	suite.NotNil(namedModel)
	suite.Equal(scheduleID, namedModel.ID)

	// 测试边界情况: 使用空的预加载
	emptyPreloadModel, err := suite.scheduleRepo.GetModel(context.Background(), []string{}, "id = ?", scheduleID)
	suite.NoError(err)
	suite.NotNil(emptyPreloadModel)
	suite.Equal(scheduleID, emptyPreloadModel.ID)

	// 测试边界情况: 使用多个条件查询
	multiConditionModel, err := suite.scheduleRepo.GetModel(
		context.Background(),
		nil,
		"id = ? AND name = ?",
		scheduleID,
		scheduleModel.Name,
	)
	suite.NoError(err)
	suite.NotNil(multiConditionModel)
	suite.Equal(scheduleID, multiConditionModel.ID)

	// 测试边界情况: 查询已删除的计划任务
	err = suite.scheduleRepo.DeleteModel(context.Background(), "id = ?", scheduleID)
	suite.NoError(err)
	deletedModel, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleID)
	suite.Error(err)
	suite.Nil(deletedModel)
}

// TestListModel 查询计划任务模型列表测试
func (suite *ScheduleTestSuite) TestListModel() {
	// 创建多个测试计划任务
	for i := 0; i < 5; i++ {
		scheduleModel := CreateTestScheduleModel(
			suite.testScriptID,
			WithScheduleName(fmt.Sprintf("schedule_%c", 'a'+i)),
		)
		err := suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
		suite.NoError(err)
		suite.NotZero(scheduleModel.ID)
	}

	// 测试查询计划任务模型列表
	qp := database.QueryParams{}
	models, err := suite.scheduleRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询计划任务列表失败")
	suite.Greater(len(models), 0, "计划任务列表应该不为空")

	// 测试查询计划任务模型列表失败: 上下文超时
	suite.testContextTimeout(func(ctx context.Context) error {
		_, err := suite.scheduleRepo.ListModel(ctx, qp)
		return err
	})

	// 测试边界情况: 大量计划任务
	// 创建20个计划任务
	for i := 0; i < 20; i++ {
		bulkSchedule := CreateTestScheduleModel(
			suite.testScriptID,
			WithScheduleName(fmt.Sprintf("bulk_schedule_%c_%d", 'a'+i%26, i/26)),
			WithScheduleEnabled(i%2 == 0),
		)
		err := suite.scheduleRepo.CreateModel(context.Background(), bulkSchedule)
		suite.NoError(err, "创建批量计划任务失败，索引: %d", i)
		suite.NotZero(bulkSchedule.ID, "批量计划任务ID应该不为零，索引: %d", i)
	}

	// 测试大量计划任务的查询
	bulkQp := database.QueryParams{}
	bulkModels, err := suite.scheduleRepo.ListModel(context.Background(), bulkQp)
	suite.NoError(err, "查询大量计划任务列表失败")
	suite.Greater(len(bulkModels), 20, "计划任务列表应该包含至少20个新创建的计划任务")

	// 测试边界情况: 不同状态的计划任务
	// 创建一些禁用的计划任务
	for i := 0; i < 3; i++ {
		disabledSchedule := CreateTestScheduleModel(
			suite.testScriptID,
			WithScheduleName(fmt.Sprintf("disabled_schedule_%c", 'a'+i)),
			WithScheduleEnabled(false),
		)
		err := suite.scheduleRepo.CreateModel(context.Background(), disabledSchedule)
		suite.NoError(err)
		suite.NotZero(disabledSchedule.ID)
	}

	// 测试查询禁用的计划任务
	disabledQp := database.QueryParams{}
	disabledModels, err := suite.scheduleRepo.ListModel(context.Background(), disabledQp)
	suite.NoError(err)
	suite.Greater(len(disabledModels), 0)

	// 测试边界情况: 不同脚本的计划任务
	// 创建第二个脚本
	scriptModel2 := CreateTestScriptModel(false)
	scriptModel2.Name = "script2.sh"
	err = suite.scriptRepo.CreateModel(context.Background(), scriptModel2)
	suite.NoError(err)
	suite.NotZero(scriptModel2.ID)

	// 为第二个脚本创建计划任务
	for i := range 3 {
		script2Schedule := CreateTestScheduleModel(
			scriptModel2.ID,
			WithScheduleName(fmt.Sprintf("script2_schedule_%c", 'a'+i)),
		)
		err := suite.scheduleRepo.CreateModel(context.Background(), script2Schedule)
		suite.NoError(err)
		suite.NotZero(script2Schedule.ID)
	}

	// 测试查询第二个脚本的计划任务
	script2Qp := database.QueryParams{}
	script2Models, err := suite.scheduleRepo.ListModel(context.Background(), script2Qp)
	suite.NoError(err)
	suite.Greater(len(script2Models), 0)
}

// TestCountModel 测试计划任务模型计数
func (suite *ScheduleTestSuite) TestCountModel() {
	// 创建多个计划任务
	for i := 0; i < 5; i++ {
		scheduleModel := CreateTestScheduleModel(suite.testScriptID)
		err := suite.scheduleRepo.CreateModel(context.Background(), scheduleModel)
		suite.NoError(err, "创建计划任务模型失败，索引: %d", i)
	}

	// 测试查询计划任务总数
	count, err := suite.scheduleRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "查询计划任务总数失败")
	suite.GreaterOrEqual(count, int64(5), "计划任务总数应该至少为 5")

	// 测试带条件查询计划任务总数
	query := map[string]any{
		"script_id": suite.testScriptID,
	}
	count, err = suite.scheduleRepo.CountModel(context.Background(), query)
	suite.NoError(err, "带条件查询计划任务总数失败")
	suite.GreaterOrEqual(count, int64(5), "带条件查询计划任务总数应该至少为 5")

	// 测试查询不存在的计划任务总数
	query = map[string]any{
		"script_id": 999999,
	}
	count, err = suite.scheduleRepo.CountModel(context.Background(), query)
	suite.NoError(err, "查询不存在的计划任务总数失败")
	suite.Equal(int64(0), count, "查询不存在的计划任务总数应该为 0")

	// 测试查询计划任务总数时上下文超时
	suite.testContextTimeout(func(ctx context.Context) error {
		_, err := suite.scheduleRepo.CountModel(ctx, nil)
		return err
	})
}

// 每个测试文件都需要这个入口函数
func TestScheduleTestSuite(t *testing.T) {
	suite.Run(t, &ScheduleTestSuite{})
}
