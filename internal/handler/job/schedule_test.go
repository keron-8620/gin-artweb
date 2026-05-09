package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	jobmodel "gin-artweb/internal/model/job"
	jobrepo "gin-artweb/internal/repo/job"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/test"
)

type ScheduleHandlerTestSuite struct {
	suite.Suite
	router       *gin.Engine
	handler      *ScheduleHandler
	scheduleRepo *jobrepo.ScheduleRepo
	scriptRepo   *jobrepo.ScriptRepo
	logger       *zap.Logger
	crontab      *cron.Cron
	testScriptID uint32
	testUsername string
}

func (suite *ScheduleHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	suite.Require().NoError(db.AutoMigrate(
		&jobmodel.ScriptModel{},
		&jobmodel.ScheduleModel{},
	), "数据库迁移失败")

	dbTimeout := test.NewTestDBTimeouts()
	dbSlowThreshold := test.NewTestDBSlowThreshold()
	suite.logger = test.NewTestZapLogger()

	suite.scriptRepo = jobrepo.NewScriptRepo(suite.logger, db, dbTimeout, dbSlowThreshold)
	suite.scheduleRepo = jobrepo.NewScheduleRepo(suite.logger, db, dbTimeout, dbSlowThreshold)

	suite.crontab = cron.New()
	suite.crontab.Start()

	// 创建一个简单的ScheduleService，不依赖完整的RecordService
	scheduleService := jobsvc.NewScheduleService(
		suite.logger,
		suite.scriptRepo,
		suite.scheduleRepo,
		nil, // recordService可以为nil，因为我们的测试不会实际触发任务执行
		suite.crontab,
	)

	suite.handler = NewScheduleHandler(suite.logger, scheduleService)

	suite.testUsername = "test_user"

	scriptModel := createTestScriptModel()
	suite.Require().NoError(suite.scriptRepo.CreateModel(context.Background(), scriptModel))
	suite.testScriptID = scriptModel.ID
}

func (suite *ScheduleHandlerTestSuite) TearDownSuite() {
	suite.crontab.Stop()
}

func (suite *ScheduleHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	suite.Require().NoError(db.Exec("DELETE FROM job_schedule").Error, "清理计划任务数据失败")

	suite.router = gin.Default()
	jobGroup := suite.router.Group("/api/v1/jobs")
	jobGroup.Use(mockAuthMiddlewareForSchedule(suite.testUsername))
	suite.handler.LoadRouter(jobGroup)
}

func mockAuthMiddlewareForSchedule(username string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := &auth.JwtClaims{
			UserInfo: auth.UserInfo{
				UserID:   1,
				Username: username,
			},
		}
		ctx := context.WithValue(c.Request.Context(), ctxutil.JwtClaimsKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func createTestScheduleUpsertDTO(scriptID uint32, options ...func(*jobmodel.ScheduleUpsertDTO)) *jobmodel.ScheduleUpsertDTO {
	dto := &jobmodel.ScheduleUpsertDTO{
		Name:          uuid.NewString(),
		Specification: "30 6 * * 1-5",
		IsEnabled:     true,
		Timeout:       300,
		IsRetry:       true,
		RetryInterval: 60,
		MaxRetries:    3,
		CreateType:    0,
		ScriptID:      scriptID,
	}

	for _, option := range options {
		option(dto)
	}

	return dto
}

func (suite *ScheduleHandlerTestSuite) createTestScheduleInDB(options ...func(*jobmodel.ScheduleUpsertDTO)) uint32 {
	dto := createTestScheduleUpsertDTO(suite.testScriptID, options...)
	model := dto.ToModel(suite.testUsername)
	suite.Require().NoError(suite.scheduleRepo.CreateModel(context.Background(), &model))
	return model.ID
}

func (suite *ScheduleHandlerTestSuite) TestNewScheduleHandler() {
	suite.NotNil(suite.handler)
	suite.NotNil(suite.handler.log)
	suite.NotNil(suite.handler.scheduleSvc)
}

func (suite *ScheduleHandlerTestSuite) TestCreateSchedule_Success() {
	dto := createTestScheduleUpsertDTO(suite.testScriptID)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("POST", "/api/v1/jobs/schedule", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp jobmodel.ScheduleResp
	suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	suite.Equal(http.StatusOK, resp.Code)
	suite.NotNil(resp.Data)
	suite.Equal(dto.Name, resp.Data.Name)
	suite.Equal(dto.Specification, resp.Data.Specification)
}

func (suite *ScheduleHandlerTestSuite) TestCreateSchedule_InvalidRequestBody() {
	req := httptest.NewRequest("POST", "/api/v1/jobs/schedule", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestCreateSchedule_InvalidCronExpression() {
	dto := createTestScheduleUpsertDTO(suite.testScriptID, func(d *jobmodel.ScheduleUpsertDTO) {
		d.Specification = "invalid cron"
	})
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("POST", "/api/v1/jobs/schedule", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestCreateSchedule_MissingRequiredFields() {
	dto := &jobmodel.ScheduleUpsertDTO{}
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("POST", "/api/v1/jobs/schedule", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestCreateSchedule_ScriptNotFound() {
	dto := createTestScheduleUpsertDTO(999999)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("POST", "/api/v1/jobs/schedule", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestUpdateSchedule_Success() {
	scheduleID := suite.createTestScheduleInDB()

	updateDTO := createTestScheduleUpsertDTO(suite.testScriptID, func(d *jobmodel.ScheduleUpsertDTO) {
		d.Name = "updated_name"
		d.IsEnabled = false
	})
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/jobs/schedule/%d", scheduleID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp jobmodel.ScheduleResp
	suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	suite.Equal(http.StatusOK, resp.Code)
	suite.NotNil(resp.Data)
	suite.Equal("updated_name", resp.Data.Name)
	suite.Equal(false, resp.Data.IsEnabled)
}

func (suite *ScheduleHandlerTestSuite) TestUpdateSchedule_NotFound() {
	updateDTO := createTestScheduleUpsertDTO(suite.testScriptID)
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", "/api/v1/jobs/schedule/999999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestUpdateSchedule_InvalidID() {
	updateDTO := createTestScheduleUpsertDTO(suite.testScriptID)
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", "/api/v1/jobs/schedule/invalid", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestUpdateSchedule_InvalidCron() {
	scheduleID := suite.createTestScheduleInDB()

	updateDTO := createTestScheduleUpsertDTO(suite.testScriptID, func(d *jobmodel.ScheduleUpsertDTO) {
		d.Specification = "invalid cron"
	})
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/jobs/schedule/%d", scheduleID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestDeleteSchedule_Success() {
	scheduleID := suite.createTestScheduleInDB()

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/jobs/schedule/%d", scheduleID), nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	model, err := suite.scheduleRepo.GetModel(context.Background(), nil, "id = ?", scheduleID)
	suite.Error(err)
	suite.Nil(model)
}

func (suite *ScheduleHandlerTestSuite) TestDeleteSchedule_NotFound() {
	req := httptest.NewRequest("DELETE", "/api/v1/jobs/schedule/999999", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestDeleteSchedule_InvalidID() {
	req := httptest.NewRequest("DELETE", "/api/v1/jobs/schedule/invalid", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestGetSchedule_Success() {
	scheduleID := suite.createTestScheduleInDB()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/schedule/%d", scheduleID), nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp jobmodel.ScheduleResp
	suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	suite.Equal(http.StatusOK, resp.Code)
	suite.NotNil(resp.Data)
	suite.Equal(scheduleID, resp.Data.ID)
}

func (suite *ScheduleHandlerTestSuite) TestGetSchedule_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/schedule/999999", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestGetSchedule_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/schedule/invalid", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.NotEqual(http.StatusOK, w.Code)
}

func (suite *ScheduleHandlerTestSuite) TestListSchedule_Success() {
	for i := 0; i < 5; i++ {
		suite.createTestScheduleInDB(func(d *jobmodel.ScheduleUpsertDTO) {
			d.Name = fmt.Sprintf("list_test_%d", i)
		})
	}

	req := httptest.NewRequest("GET", "/api/v1/jobs/schedule?page=1&size=10", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp jobmodel.PagScheduleResp
	suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	suite.Equal(http.StatusOK, resp.Code)
	suite.NotNil(resp.Data)
	suite.GreaterOrEqual(resp.Data.Total, int64(5))
}

func (suite *ScheduleHandlerTestSuite) TestListSchedule_WithNameFilter() {
	targetName := "unique_filter_name"
	suite.createTestScheduleInDB(func(d *jobmodel.ScheduleUpsertDTO) {
		d.Name = targetName
	})

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/schedule?name=%s", targetName), nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp jobmodel.PagScheduleResp
	suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	suite.Equal(http.StatusOK, resp.Code)
	suite.NotNil(resp.Data)
	suite.GreaterOrEqual(resp.Data.Total, int64(1))
}

func (suite *ScheduleHandlerTestSuite) TestListSchedule_WithEnabledFilter() {
	isEnabled := true
	suite.createTestScheduleInDB(func(d *jobmodel.ScheduleUpsertDTO) {
		d.IsEnabled = isEnabled
	})

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/schedule?is_enabled=%t", isEnabled), nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp jobmodel.PagScheduleResp
	suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	suite.Equal(http.StatusOK, resp.Code)
}

func (suite *ScheduleHandlerTestSuite) TestListSchedule_WithScriptIDFilter() {
	suite.createTestScheduleInDB()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/schedule?script_id=%d", suite.testScriptID), nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp jobmodel.PagScheduleResp
	suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	suite.Equal(http.StatusOK, resp.Code)
}

func (suite *ScheduleHandlerTestSuite) TestListSchedule_Pagination() {
	for i := 0; i < 15; i++ {
		suite.createTestScheduleInDB()
	}

	req := httptest.NewRequest("GET", "/api/v1/jobs/schedule?page=2&size=5", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp jobmodel.PagScheduleResp
	suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	suite.Equal(http.StatusOK, resp.Code)
	suite.NotNil(resp.Data)
	suite.Equal(2, resp.Data.Page)
	suite.Equal(int64(15), resp.Data.Total)
	// Check that we have items on page 2
	suite.Greater(len(resp.Data.Items), 0)
}

func (suite *ScheduleHandlerTestSuite) TestLoadRouter() {
	testRouter := gin.Default()
	testGroup := testRouter.Group("/test")
	suite.handler.LoadRouter(testGroup)
	// 不应该 panic
}

func TestScheduleHandlerTestSuite(t *testing.T) {
	suite.Run(t, &ScheduleHandlerTestSuite{})
}
