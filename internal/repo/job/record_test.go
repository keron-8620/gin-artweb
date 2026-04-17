package job

import (
	"context"
	"fmt"
	"testing"
	"time"

	"emperror.dev/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	jobmodel "gin-artweb/internal/model/job"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

// CreateTestScriptRecordModel 创建测试脚本执行记录模型
func CreateTestScriptRecordModel(scriptID uint32) *jobmodel.ScriptRecordModel {
	return &jobmodel.ScriptRecordModel{
		TriggerType: "cron",
		Status:      0,
		ExitCode:    0,
		EnvVars:     "{}",
		CommandArgs: "",
		WorkDir:     "",
		Timeout:     300,
		LogName:     fmt.Sprintf("test-%s.log", uuid.NewString()),
		Username:    "test_user",
		ScriptID:    scriptID,
	}
}

// CreateTestScriptRecordModelWithOptions 创建带选项的测试脚本执行记录模型
func CreateTestScriptRecordModelWithOptions(scriptID uint32, options map[string]interface{}) *jobmodel.ScriptRecordModel {
	record := CreateTestScriptRecordModel(scriptID)

	// 应用选项
	if val, ok := options["TriggerType"].(string); ok {
		record.TriggerType = val
	}
	if val, ok := options["Status"].(int); ok {
		record.Status = val
	}
	if val, ok := options["ExitCode"].(int); ok {
		record.ExitCode = val
	}
	if val, ok := options["EnvVars"].(string); ok {
		record.EnvVars = val
	}
	if val, ok := options["CommandArgs"].(string); ok {
		record.CommandArgs = val
	}
	if val, ok := options["WorkDir"].(string); ok {
		record.WorkDir = val
	}
	if val, ok := options["Timeout"].(int); ok {
		record.Timeout = val
	}
	if val, ok := options["LogName"].(string); ok {
		record.LogName = val
	}
	if val, ok := options["Username"].(string); ok {
		record.Username = val
	}

	return record
}

// GetTestRecordOptions 获取常用的测试记录选项
func GetTestRecordOptions() map[string]interface{} {
	return map[string]interface{}{
		"TriggerType": "manual",
		"Status":      1,
		"ExitCode":    1,
		"EnvVars":     "{\"TEST\": \"value\"}",
		"CommandArgs": "--test",
		"WorkDir":     "/tmp",
		"Timeout":     600,
		"Username":    "updated_user",
	}
}

// setupTestScript 创建测试脚本
func (suite *RecordTestSuite) setupTestScript() *jobmodel.ScriptModel {
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")
	return script
}

// setupTestRecord 创建测试记录
func (suite *RecordTestSuite) setupTestRecord(scriptID uint32) *jobmodel.ScriptRecordModel {
	record := CreateTestScriptRecordModel(scriptID)
	err := suite.recordRepo.CreateModel(context.Background(), record)
	suite.NoError(err, "创建脚本执行记录模型应该成功")
	return record
}

// createTimeoutContext 创建超时上下文
func createTimeoutContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	time.Sleep(time.Millisecond * 5) // 等待超时
	return ctx, cancel
}

type RecordTestSuite struct {
	suite.Suite
	scriptRepo *ScriptRepo
	recordRepo *RecordRepo
}

func (suite *RecordTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(
		&jobmodel.ScriptModel{},
		&jobmodel.ScriptRecordModel{},
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
	suite.recordRepo = &RecordRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: dbSlowThreshold,
	}
}

func (suite *RecordTestSuite) TestCreateModel() {
	// 先创建一个脚本模型用于测试
	script := suite.setupTestScript()

	// 测试创建脚本执行记录
	record := suite.setupTestRecord(script.ID)

	// 测试查询刚创建的记录
	fm, err := suite.recordRepo.GetModel(context.Background(), []string{}, "id = ?", record.ID)
	suite.NoError(err, "查询刚创建的脚本执行记录模型应该成功")

	// 验证记录字段
	suite.Equal(record.ID, fm.ID)
	suite.Equal(record.TriggerType, fm.TriggerType)
	suite.Equal(record.Status, fm.Status)
	suite.Equal(record.ExitCode, fm.ExitCode)
	suite.Equal(record.EnvVars, fm.EnvVars)
	suite.Equal(record.CommandArgs, fm.CommandArgs)
	suite.Equal(record.WorkDir, fm.WorkDir)
	suite.Equal(record.Timeout, fm.Timeout)
	suite.Equal(record.LogName, fm.LogName)
	suite.Equal(record.Username, fm.Username)
	suite.Equal(record.ScriptID, fm.ScriptID)
}

func (suite *RecordTestSuite) TestCreateModelWithNil() {
	// 测试创建脚本执行记录时传入空数据
	err := suite.recordRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建脚本执行记录时传入nil应该返回错误")
}

func (suite *RecordTestSuite) TestUpdateModel() {
	// 先创建一个脚本模型用于测试
	script := suite.setupTestScript()

	// 创建脚本执行记录
	record := suite.setupTestRecord(script.ID)

	// 准备更新数据
	updatedOptions := GetTestRecordOptions()
	updatedLogName := fmt.Sprintf("updated-%s.log", uuid.NewString())
	updatedOptions["LogName"] = updatedLogName

	// 测试更新脚本执行记录
	err := suite.recordRepo.UpdateModel(context.Background(), map[string]any{
		"trigger_type": updatedOptions["TriggerType"],
		"status":       updatedOptions["Status"],
		"exit_code":    updatedOptions["ExitCode"],
		"env_vars":     updatedOptions["EnvVars"],
		"command_args": updatedOptions["CommandArgs"],
		"work_dir":     updatedOptions["WorkDir"],
		"timeout":      updatedOptions["Timeout"],
		"log_name":     updatedOptions["LogName"],
		"username":     updatedOptions["Username"],
	}, "id = ?", record.ID)
	suite.NoError(err, "更新脚本执行记录模型应该成功")

	// 测试查询更新后的记录
	fm, err := suite.recordRepo.GetModel(context.Background(), []string{}, "id = ?", record.ID)
	suite.NoError(err, "查询更新后的脚本执行记录模型应该成功")
	suite.Equal(record.ID, fm.ID)
	suite.Equal(updatedOptions["TriggerType"], fm.TriggerType)
	suite.Equal(updatedOptions["Status"], fm.Status)
	suite.Equal(updatedOptions["ExitCode"], fm.ExitCode)
	suite.Equal(updatedOptions["EnvVars"], fm.EnvVars)
	suite.Equal(updatedOptions["CommandArgs"], fm.CommandArgs)
	suite.Equal(updatedOptions["WorkDir"], fm.WorkDir)
	suite.Equal(updatedOptions["Timeout"], fm.Timeout)
	suite.Equal(updatedOptions["LogName"], fm.LogName)
	suite.Equal(updatedOptions["Username"], fm.Username)
	suite.Greater(fm.UpdatedAt, record.UpdatedAt)
}

func (suite *RecordTestSuite) TestUpdateModelWithEmptyData() {
	// 测试更新时传入空数据
	err := suite.recordRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", 1)
	suite.Error(err, "更新脚本执行记录时传入空数据应该返回错误")
}

func (suite *RecordTestSuite) TestUpdateModelNonExistent() {
	// 测试更新不存在的记录
	err := suite.recordRepo.UpdateModel(context.Background(), map[string]any{
		"status": 1,
	}, "id = ?", 999999)
	suite.NoError(err, "更新不存在的脚本执行记录不应该返回错误")
}

func (suite *RecordTestSuite) TestDeleteModel() {
	// 先创建一个脚本模型用于测试
	script := suite.setupTestScript()

	// 创建脚本执行记录
	record := suite.setupTestRecord(script.ID)

	// 测试查询刚创建的记录
	fm, err := suite.recordRepo.GetModel(context.Background(), []string{}, "id = ?", record.ID)
	suite.NoError(err, "查询刚创建的脚本执行记录模型应该成功")
	suite.Equal(record.ID, fm.ID)

	// 测试删除脚本执行记录
	err = suite.recordRepo.DeleteModel(context.Background(), "id = ?", record.ID)
	suite.NoError(err, "删除脚本执行记录模型应该成功")

	// 测试查询已删除的记录
	_, err = suite.recordRepo.GetModel(context.Background(), []string{}, "id = ?", record.ID)
	suite.Error(err, "查询已删除的记录应该返回错误")
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "错误应该是记录未找到")
}

func (suite *RecordTestSuite) TestDeleteModelNonExistent() {
	// 测试删除不存在的记录
	err := suite.recordRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的脚本执行记录不应该返回错误")
}

func (suite *RecordTestSuite) TestGetModel() {
	// 先创建一个脚本模型用于测试
	script := suite.setupTestScript()

	// 创建脚本执行记录
	record := suite.setupTestRecord(script.ID)

	// 测试查询脚本执行记录
	fm, err := suite.recordRepo.GetModel(context.Background(), []string{}, "id = ?", record.ID)
	suite.NoError(err, "查询脚本执行记录模型应该成功")

	// 验证记录字段
	suite.Equal(record.ID, fm.ID)
	suite.Equal(record.TriggerType, fm.TriggerType)
	suite.Equal(record.Status, fm.Status)
	suite.Equal(record.ExitCode, fm.ExitCode)
	suite.Equal(record.EnvVars, fm.EnvVars)
	suite.Equal(record.CommandArgs, fm.CommandArgs)
	suite.Equal(record.WorkDir, fm.WorkDir)
	suite.Equal(record.Timeout, fm.Timeout)
	suite.Equal(record.LogName, fm.LogName)
	suite.Equal(record.Username, fm.Username)
	suite.Equal(record.ScriptID, fm.ScriptID)
}

func (suite *RecordTestSuite) TestGetModelNonExistent() {
	// 测试查询不存在的脚本执行记录
	_, err := suite.recordRepo.GetModel(context.Background(), []string{}, 999999)
	suite.Error(err, "查询不存在的脚本执行记录应该返回错误")
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "错误应该是记录未找到")
}

func (suite *RecordTestSuite) TestGetModelWithEmptyConditions() {
	// 测试查询时传入空条件
	result, err := suite.recordRepo.GetModel(context.Background(), []string{})

	// 当传入空条件时，GetModel方法会尝试获取数据库中的第一条记录
	if err != nil {
		// 如果返回错误，应该是record not found
		suite.Error(err, "查询时传入空条件可能返回错误")
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "错误应该是记录未找到")
	} else {
		// 如果返回结果，应该是一个有效的脚本执行记录模型
		suite.NotNil(result, "查询时传入空条件应该返回有效的脚本执行记录模型")
		suite.Greater(result.ID, uint32(0), "返回的脚本执行记录模型ID应该大于0")
	}
}

func (suite *RecordTestSuite) TestListModel() {
	// 先创建一个脚本模型用于测试
	script := suite.setupTestScript()

	// 清理可能存在的数据并创建测试数据
	for range 10 {
		record := CreateTestScriptRecordModel(script.ID)
		err := suite.recordRepo.CreateModel(context.Background(), record)
		suite.NoError(err, "创建脚本执行记录模型应该成功")
	}

	// 测试查询脚本执行记录列表
	qp := database.QueryParams{
		Limit:  10,
		Offset: 0,
	}
	ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列出脚本执行记录模型应该成功")
	suite.NotNil(ms, "脚本执行记录模型列表不应该为nil")
	suite.GreaterOrEqual(int64(len(ms)), int64(10), "脚本执行记录模型总数应该至少有10条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Limit:  5,
		Offset: 0,
	}
	ms, err = suite.recordRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页列出脚本执行记录模型应该成功")
	suite.NotNil(ms, "分页脚本执行记录模型列表不应该为nil")
	suite.Equal(5, len(ms), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(int64(len(ms)), int64(5), "分页总数应该至少等于limit")
}

func (suite *RecordTestSuite) TestListModelWithEmptyParams() {
	// 测试列表查询时传入空参数
	qp := database.QueryParams{}
	ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列表查询时传入空参数应该成功")
	suite.NotNil(ms, "脚本执行记录模型列表不应该为nil")
	suite.GreaterOrEqual(int64(len(ms)), int64(0), "脚本执行记录模型总数应该大于等于0")
}

func (suite *RecordTestSuite) TestListModelWithSorting() {
	// 先创建一个脚本模型用于测试
	script := suite.setupTestScript()

	// 测试创建多个脚本执行记录
	for i := 0; i < 5; i++ {
		record := CreateTestScriptRecordModel(script.ID)
		err := suite.recordRepo.CreateModel(context.Background(), record)
		suite.NoError(err, "创建脚本执行记录模型应该成功")
	}

	// 测试按ID降序排序
	qp := database.QueryParams{
		OrderBy: []string{"id DESC"},
	}
	ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按ID降序排序查询应该成功")
	suite.NotNil(ms, "脚本执行记录模型列表不应该为nil")
	if len(ms) > 1 {
		// 验证排序结果
		prevID := ms[0].ID
		for _, record := range ms {
			suite.LessOrEqual(record.ID, prevID, "脚本执行记录应该按ID降序排序")
			prevID = record.ID
		}
	}
}

func (suite *RecordTestSuite) TestListModelWithFiltering() {
	// 先创建一个脚本模型用于测试
	script := suite.setupTestScript()

	// 创建一个特定状态的脚本执行记录
	testStatus := 1
	record := CreateTestScriptRecordModel(script.ID)
	record.Status = testStatus
	err := suite.recordRepo.CreateModel(context.Background(), record)
	suite.NoError(err, "创建脚本执行记录模型应该成功")

	// 测试按状态过滤
	qp := database.QueryParams{
		Query: map[string]any{
			"status": testStatus,
		},
	}
	ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按状态过滤查询应该成功")
	suite.NotNil(ms, "脚本执行记录模型列表不应该为nil")
	// 验证过滤结果
	for _, record := range ms {
		suite.Equal(testStatus, record.Status, "脚本执行记录应该按状态过滤")
	}
}

func (suite *RecordTestSuite) TestCountModel() {
	// 先创建一个脚本模型用于测试
	script := suite.setupTestScript()

	// 创建多个脚本执行记录
	for i := 0; i < 5; i++ {
		record := CreateTestScriptRecordModel(script.ID)
		err := suite.recordRepo.CreateModel(context.Background(), record)
		suite.NoError(err, "创建脚本执行记录模型应该成功")
	}

	// 测试查询脚本执行记录总数
	count, err := suite.recordRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "查询脚本执行记录总数应该成功")
	suite.GreaterOrEqual(count, int64(5), "脚本执行记录总数应该至少为 5")

	// 测试带条件查询脚本执行记录总数
	query := map[string]any{
		"script_id": script.ID,
	}
	count, err = suite.recordRepo.CountModel(context.Background(), query)
	suite.NoError(err, "带条件查询脚本执行记录总数应该成功")
	suite.GreaterOrEqual(count, int64(5), "带条件查询脚本执行记录总数应该至少为 5")

	// 测试查询不存在的脚本执行记录总数
	query = map[string]any{
		"script_id": 999999,
	}
	count, err = suite.recordRepo.CountModel(context.Background(), query)
	suite.NoError(err, "查询不存在的脚本执行记录总数应该成功")
	suite.Equal(int64(0), count, "查询不存在的脚本执行记录总数应该为 0")

}

// 上下文超时测试用例
type timeoutTestCase struct {
	name     string
	execute  func(ctx context.Context) error
	expected string
}

func (suite *RecordTestSuite) TestWithContextTimeout() {
	script := suite.setupTestScript()
	record := suite.setupTestRecord(script.ID)

	testCases := []timeoutTestCase{
		{
			name: "CreateModel",
			execute: func(ctx context.Context) error {
				newRecord := CreateTestScriptRecordModel(script.ID)
				return suite.recordRepo.CreateModel(ctx, newRecord)
			},
			expected: "创建脚本执行记录时上下文超时应该返回错误",
		},
		{
			name: "UpdateModel",
			execute: func(ctx context.Context) error {
				return suite.recordRepo.UpdateModel(ctx, map[string]any{
					"status": 1,
				}, "id = ?", record.ID)
			},
			expected: "更新脚本执行记录时上下文超时应该返回错误",
		},
		{
			name: "DeleteModel",
			execute: func(ctx context.Context) error {
				return suite.recordRepo.DeleteModel(ctx, "id = ?", record.ID)
			},
			expected: "删除脚本执行记录时上下文超时应该返回错误",
		},
		{
			name: "GetModel",
			execute: func(ctx context.Context) error {
				_, err := suite.recordRepo.GetModel(ctx, []string{}, 1)
				return err
			},
			expected: "查询脚本执行记录时上下文超时应该返回错误",
		},
		{
			name: "ListModel",
			execute: func(ctx context.Context) error {
				qp := database.QueryParams{}
				_, err := suite.recordRepo.ListModel(ctx, qp)
				return err
			},
			expected: "列表查询时上下文超时应该返回错误",
		},
		{
			name: "CountModel",
			execute: func(ctx context.Context) error {
				_, err := suite.recordRepo.CountModel(ctx, nil)
				return err
			},
			expected: "查询脚本执行记录总数时上下文超时应该返回错误",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			ctx, cancel := createTimeoutContext()
			defer cancel()
			err := tc.execute(ctx)
			suite.Error(err, tc.expected)
		})
	}
}

// 每个测试文件都需要这个入口函数
func TestRecordTestSuite(t *testing.T) {
	pts := &RecordTestSuite{}
	suite.Run(t, pts)
}
