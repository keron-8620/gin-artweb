package job

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	jobmodel "gin-artweb/internal/model/job"
	jobrepo "gin-artweb/internal/repo/job"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/test"
)

func CreateTestScheduleUpsertDTO() jobmodel.ScheduleUpsertDTO {
	return jobmodel.ScheduleUpsertDTO{
		Name:          fmt.Sprintf("test_schedule_%s", uuid.NewString()),
		Specification: "0 0 * * *", // 每天凌晨执行
		IsEnabled:     true,
		EnvVars:       "{\"TEST_VAR\": \"test_value\"}",
		CommandArgs:   "arg1 arg2",
		WorkDir:       "/tmp",
		Timeout:       300,
		IsRetry:       true,
		RetryInterval: 60,
		MaxRetries:    3,
		ScriptID:      1,
	}
}

// createScheduleTestContext 创建带有测试JWT claims的上下文
func createScheduleTestContext() context.Context {
	ctx := context.Background()
	claims := &auth.JwtClaims{
		UserInfo: auth.UserInfo{
			UserID:   1,
			Username: "test_user",
			RoleID:   1,
			IsStaff:  true,
		},
	}
	return ctxutil.SetJwtClaims(ctx, claims)
}

type ScheduleTestSuite struct {
	suite.Suite
	scheduleService *ScheduleService
	db              *gorm.DB
}

func (suite *ScheduleTestSuite) SetupSuite() {
	suite.db = test.NewTestGormDBWithConfig(nil)
	if err := suite.db.AutoMigrate(&jobmodel.ScheduleModel{}, &jobmodel.ScriptModel{}); err != nil {
		suite.Error(err, "数据库迁移失败")
	}

	suite.db.Exec("DELETE FROM job_schedule")

	dbTimeout := test.NewTestDBTimeouts()
	dbSlowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()

	var script jobmodel.ScriptModel
	result := suite.db.Where("id = ?", 1).First(&script)
	if result.Error != nil {
		script = jobmodel.ScriptModel{
			Name:      "test_script.sh",
			Descr:     "测试脚本",
			Project:   "test_project",
			Label:     "test_label",
			Language:  "bash",
			Status:    true,
			IsBuiltin: false,
			Username:  "test_user",
		}
		suite.db.Create(&script)
	}

	crontab := cron.New()

	recordService := NewScriptRecordService(
		logger,
		jobrepo.NewScriptRepo(
			logger,
			suite.db,
			dbTimeout,
			dbSlowThreshold,
		),
		jobrepo.NewRecordRepo(
			logger,
			suite.db,
			dbTimeout,
			dbSlowThreshold,
		),
	)

	suite.scheduleService = NewScheduleService(
		logger,
		jobrepo.NewScriptRepo(
			logger,
			suite.db,
			dbTimeout,
			dbSlowThreshold,
		),
		jobrepo.NewScheduleRepo(
			logger,
			suite.db,
			dbTimeout,
			dbSlowThreshold,
		),
		recordService,
		crontab,
	)
}

func (suite *ScheduleTestSuite) TearDownSuite() {
	suite.db.Exec("DELETE FROM job_schedule")
	suite.db.Exec("DELETE FROM job_script")
	for id := range suite.scheduleService.entryMap {
		suite.scheduleService.crontab.Remove(suite.scheduleService.entryMap[id])
	}
	suite.scheduleService.entryMap = make(map[uint32]cron.EntryID)
}

func (suite *ScheduleTestSuite) TearDownTest() {
	suite.db.Exec("DELETE FROM job_schedule")
	for id := range suite.scheduleService.entryMap {
		suite.scheduleService.crontab.Remove(suite.scheduleService.entryMap[id])
	}
	suite.scheduleService.entryMap = make(map[uint32]cron.EntryID)
}

func (suite *ScheduleTestSuite) TestCreateSchedule() {
	dto := CreateTestScheduleUpsertDTO()
	ctx := createScheduleTestContext()

	schedule, err := suite.scheduleService.CreateSchedule(ctx, dto)
	suite.Nil(err, "创建计划任务应该成功")
	suite.Greater(schedule.ID, uint32(0), "计划任务ID应该大于0")
	suite.Equal(dto.Name, schedule.Name)
	suite.Equal(dto.Specification, schedule.Specification)
	suite.Equal(dto.IsEnabled, schedule.IsEnabled)
	suite.Equal(dto.EnvVars, schedule.EnvVars)
	suite.Equal(dto.CommandArgs, schedule.CommandArgs)
	suite.Equal(dto.WorkDir, schedule.WorkDir)
	suite.Equal(dto.Timeout, schedule.Timeout)
	suite.Equal(dto.IsRetry, schedule.IsRetry)
	suite.Equal(dto.RetryInterval, schedule.RetryInterval)
	suite.Equal(dto.MaxRetries, schedule.MaxRetries)
	suite.Equal(dto.ScriptID, schedule.ScriptID)
}

func (suite *ScheduleTestSuite) TestUpdateScheduleByID() {
	dto := CreateTestScheduleUpsertDTO()
	ctx := createScheduleTestContext()

	// 先创建一个计划任务
	schedule, err := suite.scheduleService.CreateSchedule(ctx, dto)
	suite.Nil(err, "创建计划任务应该成功")
	suite.Greater(schedule.ID, uint32(0), "计划任务ID应该大于0")

	// 准备更新数据
	updateDTO := jobmodel.ScheduleUpsertDTO{
		Name:          "updated_schedule",
		Specification: "0 1 * * *", // 每天凌晨1点执行
		IsEnabled:     true,
		EnvVars:       "{\"TEST_VAR\": \"updated_value\"}",
		CommandArgs:   "arg1 arg2 arg3",
		WorkDir:       "/tmp/updated",
		Timeout:       600,
		IsRetry:       false,
		RetryInterval: 30,
		MaxRetries:    1,
		ScriptID:      1,
	}

	// 执行更新
	updatedSchedule, err := suite.scheduleService.UpdateScheduleByID(ctx, schedule.ID, updateDTO)
	suite.Nil(err, "更新计划任务应该成功")
	suite.Equal(schedule.ID, updatedSchedule.ID)
	suite.Equal(updateDTO.Name, updatedSchedule.Name)
	suite.Equal(updateDTO.Specification, updatedSchedule.Specification)
	suite.Equal(updateDTO.IsEnabled, updatedSchedule.IsEnabled)
	suite.Equal(updateDTO.EnvVars, updatedSchedule.EnvVars)
	suite.Equal(updateDTO.CommandArgs, updatedSchedule.CommandArgs)
	suite.Equal(updateDTO.WorkDir, updatedSchedule.WorkDir)
	suite.Equal(updateDTO.Timeout, updatedSchedule.Timeout)
	suite.Equal(updateDTO.IsRetry, updatedSchedule.IsRetry)
	suite.Equal(updateDTO.RetryInterval, updatedSchedule.RetryInterval)
	suite.Equal(updateDTO.MaxRetries, updatedSchedule.MaxRetries)
	suite.Equal(updateDTO.ScriptID, updatedSchedule.ScriptID)
}

func (suite *ScheduleTestSuite) TestFindScheduleByID() {
	dto := CreateTestScheduleUpsertDTO()
	ctx := createScheduleTestContext()

	// 先创建一个计划任务
	schedule, err := suite.scheduleService.CreateSchedule(ctx, dto)
	suite.Nil(err, "创建计划任务应该成功")
	suite.Greater(schedule.ID, uint32(0), "计划任务ID应该大于0")

	// 然后查询
	foundSchedule, err := suite.scheduleService.FindScheduleByID(ctx, []string{"Script"}, schedule.ID)
	suite.Nil(err, "查询计划任务应该成功")
	suite.Greater(foundSchedule.ID, uint32(0), "计划任务ID应该大于0")
	suite.Equal(schedule.ID, foundSchedule.ID)
	suite.Equal(schedule.Name, foundSchedule.Name)
	suite.Equal(schedule.Specification, foundSchedule.Specification)
}

func (suite *ScheduleTestSuite) TestDeleteScheduleByID() {
	dto := CreateTestScheduleUpsertDTO()
	ctx := createScheduleTestContext()

	// 先创建一个计划任务
	schedule, err := suite.scheduleService.CreateSchedule(ctx, dto)
	suite.Nil(err, "创建计划任务应该成功")
	suite.Greater(schedule.ID, uint32(0), "计划任务ID应该大于0")

	// 然后删除
	err = suite.scheduleService.DeleteScheduleByID(ctx, schedule.ID)
	suite.Nil(err, "删除计划任务应该成功")

	// 再次查询应该失败
	_, err = suite.scheduleService.FindScheduleByID(ctx, []string{}, schedule.ID)
	suite.NotNil(err, "查询已删除的计划任务应该失败")
}

func (suite *ScheduleTestSuite) TestListSchedule() {
	ctx := createScheduleTestContext()

	// 创建多个计划任务
	scheduleCount := 3
	for i := 0; i < scheduleCount; i++ {
		dto := CreateTestScheduleUpsertDTO()
		_, err := suite.scheduleService.CreateSchedule(ctx, dto)
		suite.Nil(err, "创建计划任务应该成功")
	}

	// 测试列出所有计划任务
	listDTO := jobmodel.ListScheduleDTO{}
	page, size := 1, 10
	count, scheduleList, err := suite.scheduleService.ListSchedule(ctx, page, size, listDTO)
	suite.Nil(err, "列出计划任务应该成功")
	suite.GreaterOrEqual(int(count), scheduleCount, "返回的计划任务数量应该大于等于创建的数量")
	suite.NotNil(scheduleList, "返回的计划任务列表不应该为nil")
}

func (suite *ScheduleTestSuite) TestListScheduleJob() {
	ctx := createScheduleTestContext()

	// 创建一个计划任务
	dto := CreateTestScheduleUpsertDTO()
	_, err := suite.scheduleService.CreateSchedule(ctx, dto)
	suite.Nil(err, "创建计划任务应该成功")

	// 测试获取调度器任务列表
	jobList, err := suite.scheduleService.ListJob(ctx)
	suite.Nil(err, "获取调度器任务列表应该成功")
	suite.NotNil(jobList, "返回的调度器任务列表不应该为nil")
}

func (suite *ScheduleTestSuite) TestLoadSchedule() {
	ctx := createScheduleTestContext()

	// 创建一个计划任务
	dto := CreateTestScheduleUpsertDTO()
	_, err := suite.scheduleService.CreateSchedule(ctx, dto)
	suite.Nil(err, "创建计划任务应该成功")

	// 测试加载计划任务
	err = suite.scheduleService.LoadSchedule(ctx)
	suite.Nil(err, "加载计划任务应该成功")
}

func (suite *ScheduleTestSuite) TestAddJob() {
	ctx := createScheduleTestContext()

	// 创建一个计划任务模型
	schedule := &jobmodel.ScheduleModel{
		Name:          "test_schedule",
		Specification: "0 0 * * *", // 每天凌晨执行
		IsEnabled:     true,
		EnvVars:       "{\"TEST_VAR\": \"test_value\"}",
		CommandArgs:   "arg1 arg2",
		WorkDir:       "/tmp",
		Timeout:       300,
		IsRetry:       true,
		RetryInterval: 60,
		MaxRetries:    3,
		Username:      "test_user",
		ScriptID:      1,
	}

	// 测试添加计划任务
	err := suite.scheduleService.AddJob(ctx, *schedule)
	suite.Nil(err, "添加计划任务应该成功")
}

func (suite *ScheduleTestSuite) TestAddJob_WithoutRetry() {
	ctx := createScheduleTestContext()

	// 创建一个计划任务模型，禁用重试
	schedule := &jobmodel.ScheduleModel{
		Name:          "test_schedule_no_retry",
		Specification: "0 0 * * *", // 每天凌晨执行
		IsEnabled:     true,
		EnvVars:       "{\"TEST_VAR\": \"test_value\"}",
		CommandArgs:   "arg1 arg2",
		WorkDir:       "/tmp",
		Timeout:       300,
		IsRetry:       false, // 禁用重试
		RetryInterval: 60,
		MaxRetries:    3,
		Username:      "test_user",
		ScriptID:      1,
	}

	// 测试添加计划任务
	err := suite.scheduleService.AddJob(ctx, *schedule)
	suite.Nil(err, "添加计划任务应该成功")
}

func (suite *ScheduleTestSuite) TestAddJob_WithZeroMaxRetries() {
	ctx := createScheduleTestContext()

	// 创建一个计划任务模型，最大重试次数为0
	schedule := &jobmodel.ScheduleModel{
		Name:          "test_schedule_zero_retry",
		Specification: "0 0 * * *", // 每天凌晨执行
		IsEnabled:     true,
		EnvVars:       "{\"TEST_VAR\": \"test_value\"}",
		CommandArgs:   "arg1 arg2",
		WorkDir:       "/tmp",
		Timeout:       300,
		IsRetry:       true,
		RetryInterval: 60,
		MaxRetries:    0, // 最大重试次数为0
		Username:      "test_user",
		ScriptID:      1,
	}

	// 测试添加计划任务
	err := suite.scheduleService.AddJob(ctx, *schedule)
	suite.Nil(err, "添加计划任务应该成功")
}

func (suite *ScheduleTestSuite) TestRemoveJob() {
	ctx := createScheduleTestContext()

	// 创建一个计划任务模型
	schedule := &jobmodel.ScheduleModel{
		Name:          "test_schedule",
		Specification: "0 0 * * *", // 每天凌晨执行
		IsEnabled:     true,
		EnvVars:       "{\"TEST_VAR\": \"test_value\"}",
		CommandArgs:   "arg1 arg2",
		WorkDir:       "/tmp",
		Timeout:       300,
		IsRetry:       true,
		RetryInterval: 60,
		MaxRetries:    3,
		Username:      "test_user",
		ScriptID:      1,
	}

	// 先添加计划任务
	err := suite.scheduleService.AddJob(ctx, *schedule)
	suite.Nil(err, "添加计划任务应该成功")

	// 然后移除计划任务
	err = suite.scheduleService.RemoveJob(ctx, schedule.ID)
	suite.Nil(err, "移除计划任务应该成功")
}

func (suite *ScheduleTestSuite) TestCreateSchedule_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文创建计划任务
	dto := CreateTestScheduleUpsertDTO()
	_, err := suite.scheduleService.CreateSchedule(ctx, dto)
	suite.NotNil(err, "上下文错误时创建计划任务应该失败")
}

func (suite *ScheduleTestSuite) TestUpdateScheduleByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文更新计划任务
	dto := CreateTestScheduleUpsertDTO()
	_, err := suite.scheduleService.UpdateScheduleByID(ctx, 1, dto)
	suite.NotNil(err, "上下文错误时更新计划任务应该失败")
}

func (suite *ScheduleTestSuite) TestDeleteScheduleByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文删除计划任务
	err := suite.scheduleService.DeleteScheduleByID(ctx, 1)
	suite.NotNil(err, "上下文错误时删除计划任务应该失败")
}

func (suite *ScheduleTestSuite) TestFindScheduleByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文查找计划任务
	_, err := suite.scheduleService.FindScheduleByID(ctx, []string{}, 1)
	suite.NotNil(err, "上下文错误时查找计划任务应该失败")
}

func (suite *ScheduleTestSuite) TestListSchedule_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文列出计划任务
	listDTO := jobmodel.ListScheduleDTO{}
	_, _, err := suite.scheduleService.ListSchedule(ctx, 1, 10, listDTO)
	suite.NotNil(err, "上下文错误时列出计划任务应该失败")
}

func (suite *ScheduleTestSuite) TestReLoadSchedule_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文重新加载计划任务
	err := suite.scheduleService.LoadSchedule(ctx)
	suite.NotNil(err, "上下文错误时加载计划任务应该失败")
}

func (suite *ScheduleTestSuite) TestUpdateScheduleByIDs() {
	ctx := createScheduleTestContext()

	// 创建多个计划任务
	scheduleCount := 2
	createdSchedules := make([]uint32, 0, scheduleCount)

	for i := 0; i < scheduleCount; i++ {
		dto := CreateTestScheduleUpsertDTO()
		schedule, err := suite.scheduleService.CreateSchedule(ctx, dto)
		suite.Nil(err, "创建计划任务应该成功")
		createdSchedules = append(createdSchedules, schedule.ID)
	}

	// 准备更新数据
	updateData := map[string]any{
		"is_enabled": false,
		"timeout":    600,
	}

	// 执行批量更新
	err := suite.scheduleService.UpdateScheduleByIDs(ctx, createdSchedules, updateData)
	suite.Nil(err, "批量更新计划任务应该成功")
}

func (suite *ScheduleTestSuite) TestUpdateScheduleByIDs_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文批量更新计划任务
	updateData := map[string]any{
		"name":       "batch_updated_schedule",
		"is_enabled": false,
	}
	err := suite.scheduleService.UpdateScheduleByIDs(ctx, []uint32{1}, updateData)
	suite.NotNil(err, "上下文错误时批量更新计划任务应该失败")
}

func (suite *ScheduleTestSuite) TestDeleteScheduleByIDs() {
	ctx := createScheduleTestContext()

	// 创建多个计划任务
	scheduleCount := 2
	createdSchedules := make([]uint32, 0, scheduleCount)

	for i := 0; i < scheduleCount; i++ {
		dto := CreateTestScheduleUpsertDTO()
		schedule, err := suite.scheduleService.CreateSchedule(ctx, dto)
		suite.Nil(err, "创建计划任务应该成功")
		createdSchedules = append(createdSchedules, schedule.ID)
	}

	// 执行批量删除
	err := suite.scheduleService.DeleteScheduleByIDs(ctx, createdSchedules)
	suite.Nil(err, "批量删除计划任务应该成功")

	// 验证删除是否成功
	for _, id := range createdSchedules {
		_, err := suite.scheduleService.FindScheduleByID(ctx, []string{}, id)
		suite.NotNil(err, "查询已删除的计划任务应该失败")
	}
}

func (suite *ScheduleTestSuite) TestDeleteScheduleByIDs_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文批量删除计划任务
	err := suite.scheduleService.DeleteScheduleByIDs(ctx, []uint32{1})
	suite.NotNil(err, "上下文错误时批量删除计划任务应该失败")
}

func (suite *ScheduleTestSuite) TestRemoveJob_NotFound() {
	ctx := createScheduleTestContext()

	err := suite.scheduleService.RemoveJob(ctx, 99999)
	suite.Nil(err, "移除不存在的计划任务应该成功")
}

func (suite *ScheduleTestSuite) TestCreateSchedule_DisabledScript() {
	ctx := createScheduleTestContext()

	disabledScript := &jobmodel.ScriptModel{
		Name:      "disabled_script.sh",
		Descr:     "Disabled script",
		Project:   "test_project",
		Label:     "test_label",
		Language:  "bash",
		Status:    false,
		IsBuiltin: false,
		Username:  "test_user",
	}
	err := suite.scheduleService.scriptRepo.CreateModel(ctx, disabledScript)
	suite.Nil(err, "创建禁用脚本应该成功")

	dto := jobmodel.ScheduleUpsertDTO{
		Name:          "test_schedule_disabled",
		Specification: "0 0 * * *",
		IsEnabled:     true,
		ScriptID:      disabledScript.ID,
	}

	schedule, err := suite.scheduleService.CreateSchedule(ctx, dto)
	suite.Nil(err, "创建使用禁用脚本的计划任务应该成功(脚本验证在执行时进行)")
	suite.NotNil(schedule, "计划任务不应该为nil")
	suite.Equal(disabledScript.ID, schedule.ScriptID, "ScriptID should match")
	suite.Equal(disabledScript.ID, schedule.Script.ID, "Script.ID should match")
}

func (suite *ScheduleTestSuite) TestUpdateScheduleByID_DisableSchedule() {
	ctx := createScheduleTestContext()

	dto := CreateTestScheduleUpsertDTO()
	dto.IsEnabled = true
	schedule, err := suite.scheduleService.CreateSchedule(ctx, dto)
	suite.Nil(err, "创建计划任务应该成功")
	suite.True(schedule.IsEnabled, "计划任务应该是启用状态")

	updateDTO := jobmodel.ScheduleUpsertDTO{
		Name:          "updated_schedule_disabled",
		Specification: "0 1 * * *",
		IsEnabled:     false,
		ScriptID:      dto.ScriptID,
	}

	updatedSchedule, err := suite.scheduleService.UpdateScheduleByID(ctx, schedule.ID, updateDTO)
	suite.Nil(err, "更新计划任务应该成功")
	suite.Equal(updateDTO.IsEnabled, updatedSchedule.IsEnabled, "计划任务应该是禁用状态")
}

func (suite *ScheduleTestSuite) TestUpdateScheduleByIDs_EmptyList() {
	ctx := createScheduleTestContext()

	updateData := map[string]any{
		"is_enabled": false,
	}

	err := suite.scheduleService.UpdateScheduleByIDs(ctx, []uint32{}, updateData)
	suite.Nil(err, "批量更新空列表应该成功")
}

func (suite *ScheduleTestSuite) TestUpdateScheduleByIDs_EnabledScheduleTriggersAddJob() {
	ctx := createScheduleTestContext()

	dto := CreateTestScheduleUpsertDTO()
	dto.IsEnabled = true
	schedule, err := suite.scheduleService.CreateSchedule(ctx, dto)
	suite.Nil(err, "创建计划任务应该成功")
	suite.True(schedule.IsEnabled, "计划任务应该是启用状态")

	updateData := map[string]any{
		"name": "batch_updated_name",
	}

	err = suite.scheduleService.UpdateScheduleByIDs(ctx, []uint32{schedule.ID}, updateData)
	suite.Nil(err, "批量更新计划任务应该成功")
}

func (suite *ScheduleTestSuite) TestUpdateScheduleByIDs_NotFound() {
	ctx := createScheduleTestContext()

	updateData := map[string]any{
		"is_enabled": false,
	}

	err := suite.scheduleService.UpdateScheduleByIDs(ctx, []uint32{99999, 99998}, updateData)
	suite.Nil(err, "批量更新不存在的ID应该成功(不会返回错误)")
}

func (suite *ScheduleTestSuite) TestDeleteScheduleByIDs_EmptyList() {
	ctx := createScheduleTestContext()

	err := suite.scheduleService.DeleteScheduleByIDs(ctx, []uint32{})
	suite.NotNil(err, "批量删除空列表应该返回错误(GORM IN子句不支持空列表)")
}

func (suite *ScheduleTestSuite) TestListSchedule_EmptyResult() {
	ctx := createScheduleTestContext()

	listDTO := jobmodel.ListScheduleDTO{Name: "non_existent_schedule"}
	page, size := 1, 10
	count, scheduleList, err := suite.scheduleService.ListSchedule(ctx, page, size, listDTO)
	suite.Nil(err, "列出计划任务应该成功")
	suite.Equal(int64(0), count, "Count should be 0 for non-existent schedule")
	suite.Nil(scheduleList, "Schedule list should be nil for empty result")
}

func TestScheduleTestSuite(t *testing.T) {
	pts := &ScheduleTestSuite{}
	suite.Run(t, pts)
}
