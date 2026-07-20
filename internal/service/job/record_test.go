package job

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	jobmodel "gin-artweb/internal/model/job"
	jobrepo "gin-artweb/internal/repo/job"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/test"
)

func createRecordTestContext() context.Context {
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

type RecordServiceTestSuite struct {
	suite.Suite
	recordService *RecordService
	db            *gorm.DB
	scriptID      uint32
}

func (suite *RecordServiceTestSuite) SetupSuite() {
	suite.db = test.NewTestGormDBWithConfig(nil)
	if err := suite.db.AutoMigrate(&jobmodel.ScriptModel{}, &jobmodel.ScriptRecordModel{}); err != nil {
		suite.Error(err, "数据库迁移失败")
	}

	suite.db.Exec("DELETE FROM job_script_record")
	suite.db.Exec("DELETE FROM job_script")

	dbTimeout := test.NewTestDBTimeouts()
	dbSlowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()

	scriptRepo := jobrepo.NewScriptRepo(logger, suite.db, dbTimeout, dbSlowThreshold)
	recordRepo := jobrepo.NewRecordRepo(logger, suite.db, dbTimeout, dbSlowThreshold)

	suite.recordService = NewScriptRecordService(logger, scriptRepo, recordRepo)

	script := &jobmodel.ScriptModel{
		Name:      "test_script.sh",
		Descr:     "测试脚本",
		Project:   "test_project",
		Label:     "test_label",
		Language:  "bash",
		Status:    true,
		IsBuiltin: false,
		Username:  "test_user",
	}
	suite.db.Create(script)
	suite.scriptID = script.ID
}

func (suite *RecordServiceTestSuite) TestStoreAndGetCancel() {
	_, cancel := context.WithCancel(context.Background())
	id := uint32(123)

	suite.recordService.StoreCancel(id, cancel)

	retrievedCancel := suite.recordService.GetCancel(id)
	suite.NotNil(retrievedCancel, "Cancel function should be retrieved")

	retrievedCancel = suite.recordService.GetCancel(999)
	suite.Nil(retrievedCancel, "Cancel function for non-existent ID should be nil")
}

func (suite *RecordServiceTestSuite) TestDeleteCancel() {
	_, cancel := context.WithCancel(context.Background())
	id := uint32(456)

	suite.recordService.StoreCancel(id, cancel)

	suite.recordService.DeleteCancel(id)

	retrievedCancel := suite.recordService.GetCancel(id)
	suite.Nil(retrievedCancel, "Cancel function should be nil after deletion")
}

func (suite *RecordServiceTestSuite) TestGenerateScriptLogPath() {
	testTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	logName := "test_script.log"

	path := suite.recordService.GenerateScriptLogPath(testTime, logName)

	suite.NotEmpty(path)
	assert.Contains(suite.T(), path, "2024-01-15")
	assert.Contains(suite.T(), path, logName)
}

func (suite *RecordServiceTestSuite) TestCreateScriptRecord() {
	ctx := createRecordTestContext()

	execBiz := jobmodel.ExecuteScriptDTO{
		TriggerType: "api",
		ScriptID:    suite.scriptID,
		CommandArgs: "arg1 arg2",
		EnvVars:     `{"TEST_VAR": "test_value"}`,
		Timeout:     300,
		WorkDir:     "/tmp",
		Username:    "test_user",
	}

	record, err := suite.recordService.CreateScriptRecord(ctx, execBiz)

	suite.Nil(err, "CreateScriptRecord should succeed")
	suite.NotNil(record, "Record should not be nil")
	suite.Greater(record.ID, uint32(0), "Record ID should be greater than 0")
	suite.Equal(execBiz.TriggerType, record.TriggerType)
	suite.Equal(execBiz.ScriptID, record.ScriptID)
	suite.Equal(execBiz.CommandArgs, record.CommandArgs)
	suite.Equal(execBiz.EnvVars, record.EnvVars)
	suite.Equal(execBiz.Timeout, record.Timeout)
	suite.Equal(execBiz.WorkDir, record.WorkDir)
	suite.Equal(1, record.Status, "Initial status should be 1 (executing)")
}

func (suite *RecordServiceTestSuite) TestCreateScriptRecord_ContextError() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	execBiz := jobmodel.ExecuteScriptDTO{
		TriggerType: "api",
		ScriptID:    suite.scriptID,
		Timeout:     300,
	}

	record, err := suite.recordService.CreateScriptRecord(ctx, execBiz)

	suite.NotNil(err, "CreateScriptRecord should fail with cancelled context")
	suite.Nil(record, "Record should be nil")
}

func (suite *RecordServiceTestSuite) TestUpdateScriptRecord() {
	ctx := createRecordTestContext()

	execBiz := jobmodel.ExecuteScriptDTO{
		TriggerType: "api",
		ScriptID:    suite.scriptID,
		Timeout:     300,
		Username:    "test_user",
	}

	record, err := suite.recordService.CreateScriptRecord(ctx, execBiz)
	suite.Nil(err, "CreateScriptRecord should succeed")

	taskInfo := &jobmodel.TaskInfo{
		ExitCode: 0,
		Status:   2,
		ErrMSG:   "",
	}

	err = suite.recordService.UpdateScriptRecord(ctx, record.ID, taskInfo)
	suite.Nil(err, "UpdateScriptRecord should succeed")

	updatedRecord, err := suite.recordService.FindScriptRecordByID(ctx, []string{}, record.ID)
	suite.Nil(err, "FindScriptRecordByID should succeed")
	suite.Equal(2, updatedRecord.Status, "Status should be updated to 2")
	suite.Equal(0, updatedRecord.ExitCode, "ExitCode should be updated")
}

func (suite *RecordServiceTestSuite) TestUpdateScriptRecord_ContextError() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	taskInfo := &jobmodel.TaskInfo{
		ExitCode: 0,
		Status:   2,
	}

	err := suite.recordService.UpdateScriptRecord(ctx, 1, taskInfo)
	suite.NotNil(err, "UpdateScriptRecord should fail with cancelled context")
}

func (suite *RecordServiceTestSuite) TestFindScriptRecordByID() {
	ctx := createRecordTestContext()

	execBiz := jobmodel.ExecuteScriptDTO{
		TriggerType: "api",
		ScriptID:    suite.scriptID,
		Timeout:     300,
		Username:    "test_user",
	}

	record, err := suite.recordService.CreateScriptRecord(ctx, execBiz)
	suite.Nil(err, "CreateScriptRecord should succeed")

	foundRecord, err := suite.recordService.FindScriptRecordByID(ctx, []string{}, record.ID)
	suite.Nil(err, "FindScriptRecordByID should succeed")
	suite.NotNil(foundRecord)
	suite.Equal(record.ID, foundRecord.ID)
	suite.Equal(record.TriggerType, foundRecord.TriggerType)
}

func (suite *RecordServiceTestSuite) TestFindScriptRecordByID_ContextError() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := suite.recordService.FindScriptRecordByID(ctx, []string{}, 1)
	suite.NotNil(err, "FindScriptRecordByID should fail with cancelled context")
}

func (suite *RecordServiceTestSuite) TestListScriptRecord() {
	ctx := createRecordTestContext()

	for i := 0; i < 3; i++ {
		execBiz := jobmodel.ExecuteScriptDTO{
			TriggerType: "api",
			ScriptID:    suite.scriptID,
			Timeout:     300,
			Username:    "test_user",
		}
		_, err := suite.recordService.CreateScriptRecord(ctx, execBiz)
		suite.Nil(err, "CreateScriptRecord should succeed")
	}

	listDTO := &jobmodel.ListScriptRecordDTO{}
	page, size := 1, 10
	count, records, err := suite.recordService.ListScriptRecord(ctx, page, size, listDTO)

	suite.Nil(err, "ListScriptRecord should succeed")
	suite.GreaterOrEqual(count, int64(3), "Count should be at least 3")
	suite.NotNil(records)
	suite.GreaterOrEqual(len(records), 3, "Should return at least 3 records")
}

func (suite *RecordServiceTestSuite) TestListScriptRecord_ContextError() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	listDTO := &jobmodel.ListScriptRecordDTO{}
	_, _, err := suite.recordService.ListScriptRecord(ctx, 1, 10, listDTO)

	suite.NotNil(err, "ListScriptRecord should fail with cancelled context")
}

func (suite *RecordServiceTestSuite) TestListScriptRecordByIDs() {
	ctx := createRecordTestContext()

	var recordIDs []uint32
	for i := 0; i < 3; i++ {
		execBiz := jobmodel.ExecuteScriptDTO{
			TriggerType: "api",
			ScriptID:    suite.scriptID,
			Timeout:     300,
			Username:    "test_user",
		}
		record, err := suite.recordService.CreateScriptRecord(ctx, execBiz)
		suite.Nil(err, "CreateScriptRecord should succeed")
		recordIDs = append(recordIDs, record.ID)
	}

	records, err := suite.recordService.ListScriptRecordByIDs(ctx, []string{}, recordIDs)

	suite.Nil(err, "ListScriptRecordByIDs should succeed")
	suite.Len(records, 3, "Should return 3 records")
}

func (suite *RecordServiceTestSuite) TestListScriptRecordByIDs_ContextError() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := suite.recordService.ListScriptRecordByIDs(ctx, []string{}, []uint32{1})
	suite.NotNil(err, "ListScriptRecordByIDs should fail with cancelled context")
}

func (suite *RecordServiceTestSuite) TestListScriptRecord_EmptyResult() {
	ctx := createRecordTestContext()

	listDTO := &jobmodel.ListScriptRecordDTO{ScriptID: 99999}
	page, size := 1, 10
	count, records, err := suite.recordService.ListScriptRecord(ctx, page, size, listDTO)

	suite.Nil(err, "ListScriptRecord should succeed")
	suite.Equal(int64(0), count, "Count should be 0 for non-existent records")
	suite.Nil(records, "Records should be nil for empty result")
}

func (suite *RecordServiceTestSuite) TestCreateScriptRecord_DisabledScript() {
	ctx := createRecordTestContext()

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
	suite.db.Create(disabledScript)

	execBiz := jobmodel.ExecuteScriptDTO{
		TriggerType: "api",
		ScriptID:    disabledScript.ID,
		Timeout:     300,
	}

	record, err := suite.recordService.CreateScriptRecord(ctx, execBiz)

	suite.NotNil(err, "CreateScriptRecord should fail for disabled script")
	suite.Nil(record, "Record should be nil for disabled script")
}

func (suite *RecordServiceTestSuite) TestGetScriptLogStoragePath() {
	testTime := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	logName := "test_script.log"

	path := GetScriptLogStoragePath(testTime, logName)

	suite.NotEmpty(path)
	assert.Contains(suite.T(), path, "2024-06-15")
	assert.Contains(suite.T(), path, logName)
}

func (suite *RecordServiceTestSuite) TestAsyncExecuteScript() {
	ctx := createRecordTestContext()

	execBiz := jobmodel.ExecuteScriptDTO{
		TriggerType: "api",
		ScriptID:    suite.scriptID,
		Timeout:     300,
		Username:    "test_user",
	}

	record, err := suite.recordService.AsyncExecuteScript(ctx, execBiz)

	suite.Nil(err, "AsyncExecuteScript should succeed")
	suite.NotNil(record, "Record should not be nil")
}

func (suite *RecordServiceTestSuite) TestSyncExecuteScript() {
	ctx := createRecordTestContext()

	execBiz := jobmodel.ExecuteScriptDTO{
		TriggerType: "api",
		ScriptID:    suite.scriptID,
		Timeout:     300,
		Username:    "test_user",
	}

	taskInfo, err := suite.recordService.SyncExecuteScript(ctx, execBiz)

	suite.Nil(err, "SyncExecuteScript should succeed")
	suite.NotNil(taskInfo, "TaskInfo should not be nil")
}

func (suite *RecordServiceTestSuite) TestCancel_NotFound() {
	ctx := createRecordTestContext()

	suite.recordService.Cancel(ctx, 99999)

	cancelFunc := suite.recordService.GetCancel(99999)
	suite.Nil(cancelFunc, "Cancel function for non-existent ID should be nil")
}

func (suite *RecordServiceTestSuite) TestCancel() {
	ctx := createRecordTestContext()

	execBiz := jobmodel.ExecuteScriptDTO{
		TriggerType: "api",
		ScriptID:    suite.scriptID,
		Timeout:     300,
		Username:    "test_user",
	}

	record, err := suite.recordService.CreateScriptRecord(ctx, execBiz)
	suite.Require().Nil(err, "CreateScriptRecord should succeed")
	suite.Require().NotNil(record, "Record should not be nil")

	suite.recordService.Cancel(ctx, record.ID)

	cancelFunc := suite.recordService.GetCancel(record.ID)
	suite.Nil(cancelFunc, "Cancel function should be removed after Cancel is called")
}

func TestRecordServiceTestSuite(t *testing.T) {
	pts := &RecordServiceTestSuite{}
	suite.Run(t, pts)
}
