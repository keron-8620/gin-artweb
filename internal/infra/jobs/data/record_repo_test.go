package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"emperror.dev/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/jobs/model"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestScriptRecordModel(scriptID uint32) *model.ScriptRecordModel {
	return &model.ScriptRecordModel{
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

type RecordTestSuite struct {
	suite.Suite
	scriptRepo *ScriptRepo
	recordRepo *RecordRepo
}

func (suite *RecordTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&model.ScriptModel{},
		&model.ScriptRecordModel{},
	)
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.scriptRepo = &ScriptRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
	suite.recordRepo = &RecordRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *RecordTestSuite) TestCreateModel() {
	// 先创建一个脚本模型用于测试
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")

	// 测试创建脚本执行记录
	record := CreateTestScriptRecordModel(script.ID)
	err = suite.recordRepo.CreateModel(context.Background(), record)
	suite.NoError(err, "创建脚本执行记录模型应该成功")

	// 测试查询刚创建的记录
	fm, err := suite.recordRepo.GetModel(context.Background(), []string{}, "id = ?", record.ID)
	suite.NoError(err, "查询刚创建的脚本执行记录模型应该成功")
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

func (suite *RecordTestSuite) TestCreateModelWithContextTimeout() {
	// 测试创建脚本执行记录时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 先创建一个脚本模型用于测试
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")

	// 尝试在超时上下文中创建记录
	record := CreateTestScriptRecordModel(script.ID)
	err = suite.recordRepo.CreateModel(ctx, record)
	suite.Error(err, "创建脚本执行记录时上下文超时应该返回错误")
}

func (suite *RecordTestSuite) TestUpdateModel() {
	// 先创建一个脚本模型用于测试
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")

	// 创建脚本执行记录
	record := CreateTestScriptRecordModel(script.ID)
	err = suite.recordRepo.CreateModel(context.Background(), record)
	suite.NoError(err, "创建脚本执行记录模型应该成功")

	// 准备更新数据
	updatedTriggerType := "manual"
	updatedStatus := 1
	updatedExitCode := 1
	updatedEnvVars := "{\"TEST\": \"value\"}"
	updatedCommandArgs := "--test"
	updatedWorkDir := "/tmp"
	updatedTimeout := 600
	updatedLogName := fmt.Sprintf("updated-%s.log", uuid.NewString())
	updatedUsername := "updated_user"

	// 测试更新脚本执行记录
	err = suite.recordRepo.UpdateModel(context.Background(), map[string]any{
		"trigger_type": updatedTriggerType,
		"status":       updatedStatus,
		"exit_code":    updatedExitCode,
		"env_vars":     updatedEnvVars,
		"command_args": updatedCommandArgs,
		"work_dir":     updatedWorkDir,
		"timeout":      updatedTimeout,
		"log_name":     updatedLogName,
		"username":     updatedUsername,
	}, "id = ?", record.ID)
	suite.NoError(err, "更新脚本执行记录模型应该成功")

	// 测试查询更新后的记录
	fm, err := suite.recordRepo.GetModel(context.Background(), []string{}, "id = ?", record.ID)
	suite.NoError(err, "查询更新后的脚本执行记录模型应该成功")
	suite.Equal(record.ID, fm.ID)
	suite.Equal(updatedTriggerType, fm.TriggerType)
	suite.Equal(updatedStatus, fm.Status)
	suite.Equal(updatedExitCode, fm.ExitCode)
	suite.Equal(updatedEnvVars, fm.EnvVars)
	suite.Equal(updatedCommandArgs, fm.CommandArgs)
	suite.Equal(updatedWorkDir, fm.WorkDir)
	suite.Equal(updatedTimeout, fm.Timeout)
	suite.Equal(updatedLogName, fm.LogName)
	suite.Equal(updatedUsername, fm.Username)
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

func (suite *RecordTestSuite) TestUpdateModelWithContextTimeout() {
	// 先创建一个脚本模型用于测试
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")

	// 创建脚本执行记录
	record := CreateTestScriptRecordModel(script.ID)
	err = suite.recordRepo.CreateModel(context.Background(), record)
	suite.NoError(err, "创建脚本执行记录模型应该成功")

	// 测试更新脚本执行记录时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试在超时上下文中更新记录
	err = suite.recordRepo.UpdateModel(ctx, map[string]any{
		"status": 1,
	}, "id = ?", record.ID)
	suite.Error(err, "更新脚本执行记录时上下文超时应该返回错误")
}

func (suite *RecordTestSuite) TestDeleteModel() {
	// 先创建一个脚本模型用于测试
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")

	// 创建脚本执行记录
	record := CreateTestScriptRecordModel(script.ID)
	err = suite.recordRepo.CreateModel(context.Background(), record)
	suite.NoError(err, "创建脚本执行记录模型应该成功")

	// 测试查询刚创建的记录
	fm, err := suite.recordRepo.GetModel(context.Background(), []string{}, "id = ?", record.ID)
	suite.NoError(err, "查询刚创建的脚本执行记录模型应该成功")
	suite.Equal(record.ID, fm.ID)

	// 测试删除脚本执行记录
	err = suite.recordRepo.DeleteModel(context.Background(), "id = ?", record.ID)
	suite.NoError(err, "删除脚本执行记录模型应该成功")

	// 测试查询已删除的记录
	_, err = suite.recordRepo.GetModel(context.Background(), []string{}, "id = ?", record.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *RecordTestSuite) TestDeleteModelNonExistent() {
	// 测试删除不存在的记录
	err := suite.recordRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的脚本执行记录不应该返回错误")
}

func (suite *RecordTestSuite) TestDeleteModelWithContextTimeout() {
	// 先创建一个脚本模型用于测试
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")

	// 创建脚本执行记录
	record := CreateTestScriptRecordModel(script.ID)
	err = suite.recordRepo.CreateModel(context.Background(), record)
	suite.NoError(err, "创建脚本执行记录模型应该成功")

	// 测试删除脚本执行记录时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试在超时上下文中删除记录
	err = suite.recordRepo.DeleteModel(ctx, "id = ?", record.ID)
	suite.Error(err, "删除脚本执行记录时上下文超时应该返回错误")
}

func (suite *RecordTestSuite) TestGetModel() {
	// 先创建一个脚本模型用于测试
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")

	// 创建脚本执行记录
	record := CreateTestScriptRecordModel(script.ID)
	err = suite.recordRepo.CreateModel(context.Background(), record)
	suite.NoError(err, "创建脚本执行记录模型应该成功")

	// 测试查询脚本执行记录
	fm, err := suite.recordRepo.GetModel(context.Background(), []string{}, "id = ?", record.ID)
	suite.NoError(err, "查询脚本执行记录模型应该成功")
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
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询不存在的脚本执行记录应该返回记录未找到错误")
}

func (suite *RecordTestSuite) TestGetModelWithEmptyConditions() {
	// 测试查询时传入空条件
	result, err := suite.recordRepo.GetModel(context.Background(), []string{})
	// 当传入空条件时，GetModel方法会尝试获取数据库中的第一条记录
	// 如果数据库为空，会返回record not found错误
	// 如果数据库不为空，会返回第一条记录
	if err != nil {
		// 如果返回错误，应该是record not found
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询时传入空条件应该返回记录未找到错误")
	} else {
		// 如果返回结果，应该是一个有效的脚本执行记录模型
		suite.NotNil(result, "查询时传入空条件应该返回有效的脚本执行记录模型")
		suite.Greater(result.ID, uint32(0), "返回的脚本执行记录模型ID应该大于0")
	}
}

func (suite *RecordTestSuite) TestGetModelWithContextTimeout() {
	// 测试查询脚本执行记录时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试在超时上下文中查询记录
	_, err := suite.recordRepo.GetModel(ctx, []string{}, 1)
	suite.Error(err, "查询脚本执行记录时上下文超时应该返回错误")
}

func (suite *RecordTestSuite) TestListModel() {
	// 先创建一个脚本模型用于测试
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")

	// 清理可能存在的数据并创建测试数据
	for range 10 {
		record := CreateTestScriptRecordModel(script.ID)
		err := suite.recordRepo.CreateModel(context.Background(), record)
		suite.NoError(err, "创建脚本执行记录模型应该成功")
	}

	// 测试查询脚本执行记录列表
	qp := database.QueryParams{
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	total, ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列出脚本执行记录模型应该成功")
	suite.NotNil(ms, "脚本执行记录模型列表不应该为nil")
	suite.GreaterOrEqual(total, int64(10), "脚本执行记录模型总数应该至少有10条")

	// 测试分页查询
	qpPaginated := database.QueryParams{
		Size:    5,
		Page:    0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.recordRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页列出脚本执行记录模型应该成功")
	suite.NotNil(pMs, "分页脚本执行记录模型列表不应该为nil")
	suite.Equal(5, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(5), "分页总数应该至少等于limit")
}

func (suite *RecordTestSuite) TestListModelWithEmptyParams() {
	// 测试列表查询时传入空参数
	qp := database.QueryParams{}
	total, ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列表查询时传入空参数应该成功")
	suite.NotNil(ms, "脚本执行记录模型列表不应该为nil")
	suite.GreaterOrEqual(total, int64(0), "脚本执行记录模型总数应该大于等于0")
}

func (suite *RecordTestSuite) TestListModelWithSorting() {
	// 先创建一个脚本模型用于测试
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")

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
	_, ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按ID降序排序查询应该成功")
	suite.NotNil(ms, "脚本执行记录模型列表不应该为nil")
	if len(*ms) > 1 {
		// 验证排序结果
		prevID := (*ms)[0].ID
		for _, record := range *ms {
			suite.LessOrEqual(record.ID, prevID, "脚本执行记录应该按ID降序排序")
			prevID = record.ID
		}
	}
}

func (suite *RecordTestSuite) TestListModelWithFiltering() {
	// 先创建一个脚本模型用于测试
	script := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), script)
	suite.NoError(err, "创建脚本模型应该成功")

	// 创建一个特定状态的脚本执行记录
	testStatus := 1
	record := CreateTestScriptRecordModel(script.ID)
	record.Status = testStatus
	err = suite.recordRepo.CreateModel(context.Background(), record)
	suite.NoError(err, "创建脚本执行记录模型应该成功")

	// 测试按状态过滤
	qp := database.QueryParams{
		Query: map[string]any{
			"status": testStatus,
		},
	}
	_, ms, err := suite.recordRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按状态过滤查询应该成功")
	suite.NotNil(ms, "脚本执行记录模型列表不应该为nil")
	// 验证过滤结果
	for _, record := range *ms {
		suite.Equal(testStatus, record.Status, "脚本执行记录应该按状态过滤")
	}
}

func (suite *RecordTestSuite) TestListModelWithContextTimeout() {
	// 测试列表查询时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试在超时上下文中列表查询
	qp := database.QueryParams{}
	_, _, err := suite.recordRepo.ListModel(ctx, qp)
	suite.Error(err, "列表查询时上下文超时应该返回错误")
}

// 每个测试文件都需要这个入口函数
func TestRecordTestSuite(t *testing.T) {
	pts := &RecordTestSuite{}
	suite.Run(t, pts)
}
