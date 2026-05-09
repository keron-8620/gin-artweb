package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	commodel "gin-artweb/internal/model/common"
	jobmodel "gin-artweb/internal/model/job"
	jobrepo "gin-artweb/internal/repo/job"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/test"
)

type ScriptRecordHandlerTestSuite struct {
	suite.Suite
	router           *gin.Engine
	handler          *ScriptRecordHandler
	recordService    *jobsvc.RecordService
	scriptRepo       *jobrepo.ScriptRepo
	recordRepo       *jobrepo.RecordRepo
	testScriptID     uint32
	testUsername     string
	testDir          string
	scriptStorageDir string
	logStorageDir    string
	db               *gorm.DB
}

func (s *ScriptRecordHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	testDir, err := os.MkdirTemp("", "record-handler-test-*")
	s.Require().NoError(err, "创建测试目录失败")
	s.testDir = testDir

	s.scriptStorageDir = filepath.Join(testDir, "storage", "script")
	s.logStorageDir = filepath.Join(testDir, "storage", "logs")
	s.Require().NoError(os.MkdirAll(s.scriptStorageDir, 0755), "创建脚本存储目录失败")
	s.Require().NoError(os.MkdirAll(s.logStorageDir, 0755), "创建日志存储目录失败")

	config.StorageDir = filepath.Join(testDir, "storage")
	config.TmpDir = filepath.Join(testDir, "tmp")
	s.Require().NoError(os.MkdirAll(config.TmpDir, 0755), "创建临时目录失败")

	dbPath := filepath.Join(testDir, "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	s.Require().NoError(err, "创建测试数据库失败")
	s.db = db

	s.Require().NoError(s.db.AutoMigrate(&jobmodel.ScriptModel{}, &jobmodel.ScriptRecordModel{}), "数据库迁移失败")

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()

	s.scriptRepo = jobrepo.NewScriptRepo(logger, s.db, dbTimeout, slowThreshold)
	s.recordRepo = jobrepo.NewRecordRepo(logger, s.db, dbTimeout, slowThreshold)

	s.recordService = jobsvc.NewScriptRecordService(logger, s.scriptRepo, s.recordRepo)

	s.handler = NewScriptRecordHandler(logger, s.recordService)

	s.testUsername = "test-user"
}

func (s *ScriptRecordHandlerTestSuite) TearDownSuite() {
	if s.testDir != "" {
		os.RemoveAll(s.testDir)
	}
}

func mockRecordAuthMiddleware(username string) gin.HandlerFunc {
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

func (s *ScriptRecordHandlerTestSuite) SetupTest() {
	s.db.Exec("PRAGMA foreign_keys = OFF")
	s.Require().NoError(s.db.Exec("DELETE FROM job_script_record").Error, "清理执行记录数据失败")
	s.Require().NoError(s.db.Exec("DELETE FROM job_script").Error, "清理脚本数据失败")
	s.db.Exec("PRAGMA foreign_keys = ON")

	s.router = gin.Default()
	recordGroup := s.router.Group("/api/v1/jobs")
	recordGroup.Use(mockRecordAuthMiddleware(s.testUsername))
	s.handler.LoadRouter(recordGroup)

	script := &jobmodel.ScriptModel{
		Name:      fmt.Sprintf("test-script-%d.sh", uuid.New().ID()),
		Descr:     "Test script description",
		Project:   "test-project",
		Label:     "test-label",
		Language:  "bash",
		Status:    true,
		IsBuiltin: false,
		Username:  s.testUsername,
	}

	scriptPath := filepath.Join(s.scriptStorageDir, script.Project, script.Label, script.Name)
	s.Require().NoError(os.MkdirAll(filepath.Dir(scriptPath), 0755), "创建脚本目录失败")
	s.Require().NoError(os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'test output'\nsleep 1"), 0755), "创建测试脚本文件失败")

	s.Require().NoError(s.scriptRepo.CreateModel(context.Background(), script))
	s.testScriptID = script.ID
}

func (s *ScriptRecordHandlerTestSuite) createExecScriptRequest(method, path string, body interface{}) (*http.Request, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (s *ScriptRecordHandlerTestSuite) TestExecScriptRecord_Success() {
	reqBody := jobmodel.CreateScriptRecordDTO{
		ScriptID:    s.testScriptID,
		CommandArgs: "arg1 arg2",
		EnvVars:     `{"TEST_VAR": "test_value"}`,
		Timeout:     300,
		WorkDir:     "/tmp",
	}

	req, err := s.createExecScriptRequest("POST", "/api/v1/jobs/record", reqBody)
	s.Require().NoError(err, "创建请求失败")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")

	var resp jobmodel.ScriptRecordResp
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "解析响应失败")
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data.Script.ID)
}

func (s *ScriptRecordHandlerTestSuite) TestExecScriptRecord_MissingScriptID() {
	reqBody := map[string]interface{}{
		"timeout": 300,
	}

	req, err := s.createExecScriptRequest("POST", "/api/v1/jobs/record", reqBody)
	s.Require().NoError(err, "创建请求失败")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code, "缺少必填参数时应返回400")
}

func (s *ScriptRecordHandlerTestSuite) TestExecScriptRecord_InvalidScriptID() {
	reqBody := jobmodel.CreateScriptRecordDTO{
		ScriptID: 99999,
		Timeout:  300,
	}

	req, err := s.createExecScriptRequest("POST", "/api/v1/jobs/record", reqBody)
	s.Require().NoError(err, "创建请求失败")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code, "不存在的脚本ID时应返回404")
}

func (s *ScriptRecordHandlerTestSuite) TestExecScriptRecord_MissingTimeout() {
	reqBody := map[string]interface{}{
		"script_id": s.testScriptID,
	}

	req, err := s.createExecScriptRequest("POST", "/api/v1/jobs/record", reqBody)
	s.Require().NoError(err, "创建请求失败")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code, "缺少必填参数timeout时应返回400")
}

func (s *ScriptRecordHandlerTestSuite) TestExecScriptRecord_InvalidJSON() {
	req := httptest.NewRequest("POST", "/api/v1/jobs/record", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code, "无效JSON时应返回400")
}

func (s *ScriptRecordHandlerTestSuite) TestExecScriptRecord_EmptyBody() {
	req := httptest.NewRequest("POST", "/api/v1/jobs/record", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code, "空请求体时应返回400")
}

func (s *ScriptRecordHandlerTestSuite) TestGetScriptRecord_Success() {
	execReqBody := jobmodel.CreateScriptRecordDTO{
		ScriptID:    s.testScriptID,
		CommandArgs: "arg1",
		Timeout:     300,
		WorkDir:     "/tmp",
	}
	execReq, err := s.createExecScriptRequest("POST", "/api/v1/jobs/record", execReqBody)
	s.Require().NoError(err)
	execW := httptest.NewRecorder()
	s.router.ServeHTTP(execW, execReq)
	s.Require().Equal(http.StatusOK, execW.Code, "创建执行记录失败")

	var execResp jobmodel.ScriptRecordResp
	err = json.Unmarshal(execW.Body.Bytes(), &execResp)
	s.Require().NoError(err)
	recordID := execResp.Data.ID

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/record/%d", recordID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")

	var resp jobmodel.ScriptRecordResp
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.Code)
	s.Equal(recordID, resp.Data.ID)
}

func (s *ScriptRecordHandlerTestSuite) TestGetScriptRecord_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/record/999999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code, "不存在的记录ID时应返回404")
}

func (s *ScriptRecordHandlerTestSuite) TestGetScriptRecord_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/record/invalid", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code, "无效ID格式时应返回400")
}

func (s *ScriptRecordHandlerTestSuite) createRecordViaHTTP() *jobmodel.ScriptRecordDetailOut {
	execReqBody := jobmodel.CreateScriptRecordDTO{
		ScriptID:    s.testScriptID,
		CommandArgs: "arg1 arg2",
		Timeout:     300,
		WorkDir:     "/tmp",
	}
	execReq, err := s.createExecScriptRequest("POST", "/api/v1/jobs/record", execReqBody)
	s.Require().NoError(err)
	execW := httptest.NewRecorder()
	s.router.ServeHTTP(execW, execReq)
	s.Require().Equal(http.StatusOK, execW.Code)

	var execResp jobmodel.ScriptRecordResp
	err = json.Unmarshal(execW.Body.Bytes(), &execResp)
	s.Require().NoError(err)
	return &execResp.Data
}

func (s *ScriptRecordHandlerTestSuite) TestListScriptRecord_Success() {
	for i := 0; i < 3; i++ {
		s.createRecordViaHTTP()
	}

	req := httptest.NewRequest("GET", "/api/v1/jobs/record", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")

	var resp jobmodel.PagScriptRecordResp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "解析响应失败")
	s.Equal(http.StatusOK, resp.Code)
	s.GreaterOrEqual(len(resp.Data.Items), 3, "应返回至少3条记录")
}

func (s *ScriptRecordHandlerTestSuite) TestListScriptRecord_WithPagination() {
	for i := 0; i < 5; i++ {
		s.createRecordViaHTTP()
	}

	req := httptest.NewRequest("GET", "/api/v1/jobs/record?page=1&size=2", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")

	var resp jobmodel.PagScriptRecordResp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "解析响应失败")
	s.Equal(http.StatusOK, resp.Code)
	s.Equal(1, resp.Data.Page)
	s.GreaterOrEqual(resp.Data.Size, 2, "分页大小应至少为请求的2")
	s.GreaterOrEqual(int(resp.Data.Total), 5, "总记录数应至少为5")
}

func (s *ScriptRecordHandlerTestSuite) TestListScriptRecord_FilterByTriggerType() {
	record := s.createRecordViaHTTP()
	s.Require().Equal("api", record.TriggerType)

	req := httptest.NewRequest("GET", "/api/v1/jobs/record?trigger_type=api", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")

	var resp jobmodel.PagScriptRecordResp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "解析响应失败")
	s.GreaterOrEqual(len(resp.Data.Items), 1, "应返回至少1条记录")
	for _, item := range resp.Data.Items {
		s.Equal("api", item.TriggerType, "所有记录应为api触发类型")
	}
}

func (s *ScriptRecordHandlerTestSuite) TestListScriptRecord_FilterByStatus() {
	s.createRecordViaHTTP()

	req := httptest.NewRequest("GET", "/api/v1/jobs/record?status=1", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")

	var resp jobmodel.PagScriptRecordResp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "解析响应失败")
	s.GreaterOrEqual(len(resp.Data.Items), 1, "应返回至少1条记录")
}

func (s *ScriptRecordHandlerTestSuite) TestListScriptRecord_FilterByScriptID() {
	record := s.createRecordViaHTTP()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/record?script_id=%d", record.Script.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")

	var resp jobmodel.PagScriptRecordResp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "解析响应失败")
	s.GreaterOrEqual(len(resp.Data.Items), 1, "应返回至少1条记录")
	for _, item := range resp.Data.Items {
		s.Equal(record.Script.ID, item.Script.ID, "所有记录的ScriptID应匹配")
	}
}

func (s *ScriptRecordHandlerTestSuite) TestListScriptRecord_FilterByUsername() {
	record := s.createRecordViaHTTP()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/record?username=%s", record.Username), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")

	var resp jobmodel.PagScriptRecordResp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "解析响应失败")
	s.GreaterOrEqual(len(resp.Data.Items), 1, "应返回至少1条记录")
	for _, item := range resp.Data.Items {
		s.Equal(record.Username, item.Username, "所有记录的用户名应匹配")
	}
}

func (s *ScriptRecordHandlerTestSuite) TestListScriptRecord_EmptyResult() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/record?script_id=99999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")

	var resp jobmodel.PagScriptRecordResp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "解析响应失败")
	s.Equal(int64(0), resp.Data.Total, "总记录数应为0")
}

func (s *ScriptRecordHandlerTestSuite) TestDownloadScriptRecordLog_Success() {
	record := s.createRecordViaHTTP()
	recordID := record.ID

	var model jobmodel.ScriptRecordModel
	err := s.db.First(&model, recordID).Error
	s.Require().NoError(err, "获取记录模型失败")

	logPath := s.recordService.GenerateScriptLogPath(time.Now(), model.LogName)
	logDir := filepath.Dir(logPath)
	s.Require().NoError(os.MkdirAll(logDir, 0755), "创建日志目录失败")
	logContent := "[2024-01-01T12:00:00Z] 开始执行脚本\n[2024-01-01T12:00:01Z] 脚本执行成功\n"
	s.Require().NoError(os.WriteFile(logPath, []byte(logContent), 0644), "创建日志文件失败")

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/record/%d/log", recordID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")
	s.NotEmpty(w.Body.Bytes(), "响应体不应为空")
}

func (s *ScriptRecordHandlerTestSuite) TestDownloadScriptRecordLog_RecordNotFound() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/record/999999/log", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code, "不存在的记录ID时应返回404")
}

func (s *ScriptRecordHandlerTestSuite) TestDownloadScriptRecordLog_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/record/invalid/log", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code, "无效ID格式时应返回400")
}

func (s *ScriptRecordHandlerTestSuite) TestDownloadScriptRecordLog_LogFileNotFound() {
	record := s.createRecordViaHTTP()
	recordID := record.ID

	var model jobmodel.ScriptRecordModel
	err := s.db.First(&model, recordID).Error
	s.Require().NoError(err, "获取记录模型失败")

	newLogName := "nonexistent.log"
	s.db.Model(&model).Update("log_name", newLogName)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/record/%d/log", recordID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code, "日志文件不存在时应返回404")
}

func (s *ScriptRecordHandlerTestSuite) TestStreamScriptRecordLog_Success() {
	s.T().Skip("SSE streaming test requires async infrastructure, skipping for now")
}

func (s *ScriptRecordHandlerTestSuite) TestStreamScriptRecordLog_RecordNotFound() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/record/999999/log/stream", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code, "不存在的记录ID时应返回404")
}

func (s *ScriptRecordHandlerTestSuite) TestStreamScriptRecordLog_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/record/invalid/log/stream", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code, "无效ID格式时应返回400")
}

func (s *ScriptRecordHandlerTestSuite) TestStreamScriptRecordLog_LogFileNotFound() {
	record := s.createRecordViaHTTP()
	recordID := record.ID

	var model jobmodel.ScriptRecordModel
	err := s.db.First(&model, recordID).Error
	s.Require().NoError(err, "获取记录模型失败")

	newLogName := "nonexistent.log"
	s.db.Model(&model).Update("log_name", newLogName)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/record/%d/log/stream", recordID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code, "日志文件不存在时应返回404")
}

func (s *ScriptRecordHandlerTestSuite) TestCancelScriptRecord_Success() {
	record := s.createRecordViaHTTP()

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/jobs/record/%d", record.ID), nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "响应状态码应为200")

	var resp commodel.MapAPIResp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "解析响应失败")
	s.Equal(http.StatusOK, resp.Code)
}

func (s *ScriptRecordHandlerTestSuite) TestCancelScriptRecord_NotFound() {
	req := httptest.NewRequest("DELETE", "/api/v1/jobs/record/999999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "取消不存在的记录应返回200")
}

func (s *ScriptRecordHandlerTestSuite) TestCancelScriptRecord_InvalidID() {
	req := httptest.NewRequest("DELETE", "/api/v1/jobs/record/invalid", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code, "无效ID格式时应返回400")
}

func TestScriptRecordHandlerTestSuite(t *testing.T) {
	pts := &ScriptRecordHandlerTestSuite{}
	suite.Run(t, pts)
}
