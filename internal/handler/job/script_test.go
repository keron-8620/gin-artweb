package job

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	jobmodel "gin-artweb/internal/model/job"
	jobrepo "gin-artweb/internal/repo/job"
	jobsvc "gin-artweb/internal/service/job"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/test"
)

func createTestScriptModel() *jobmodel.ScriptModel {
	return &jobmodel.ScriptModel{
		Name:      fmt.Sprintf("test-script-%d.sh", uuid.New().ID()),
		Descr:     "Test script description",
		Project:   "test-project",
		Label:     "test-label",
		Language:  "bash",
		Status:    true,
		IsBuiltin: false,
		Username:  "test-user",
	}
}

type ScriptHandlerTestSuite struct {
	suite.Suite
	router        *gin.Engine
	handler       *ScriptHandler
	scriptService *jobsvc.ScriptService
	scriptRepo    *jobrepo.ScriptRepo
	testScriptID  uint32
	testUsername  string
	testDir       string
}

func (s *ScriptHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	testDir, err := os.MkdirTemp("", "script-test-*")
	s.Require().NoError(err, "创建测试目录失败")
	s.testDir = testDir

	config.StorageDir = filepath.Join(testDir, "storage")
	config.TmpDir = filepath.Join(testDir, "tmp")
	s.Require().NoError(os.MkdirAll(config.StorageDir, 0755), "创建存储目录失败")
	s.Require().NoError(os.MkdirAll(config.TmpDir, 0755), "创建临时目录失败")

	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.AutoMigrate(&jobmodel.ScriptModel{}), "数据库迁移失败")

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()

	s.scriptRepo = jobrepo.NewScriptRepo(logger, db, dbTimeout, slowThreshold)

	s.scriptService = jobsvc.NewScriptService(logger, s.scriptRepo)

	s.handler = NewScriptHandler(logger, s.scriptService, 500<<20)

	s.testUsername = "test-user"
}

func (s *ScriptHandlerTestSuite) TearDownSuite() {
	if s.testDir != "" {
		os.RemoveAll(s.testDir)
	}
}

func mockScriptAuthMiddleware(username string) gin.HandlerFunc {
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

func (s *ScriptHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.Exec("DELETE FROM job_script").Error, "清理脚本数据失败")

	s.router = gin.Default()
	scriptGroup := s.router.Group("/api/v1/jobs")
	scriptGroup.Use(mockScriptAuthMiddleware(s.testUsername))
	s.handler.LoadRouter(scriptGroup)

	testScript := createTestScriptModel()
	scriptPath := filepath.Join(config.StorageDir, "script", testScript.Project, testScript.Label, testScript.Name)
	s.Require().NoError(os.MkdirAll(filepath.Dir(scriptPath), 0755), "创建脚本目录失败")
	s.Require().NoError(os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'test'"), 0755), "创建测试脚本文件失败")

	s.Require().NoError(s.scriptRepo.CreateModel(context.Background(), testScript))
	s.testScriptID = testScript.ID
}

func (s *ScriptHandlerTestSuite) createMultipartRequest(method, path string, filename, project, label, language string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	_, err = part.Write([]byte("#!/bin/bash\necho 'test'"))
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("descr", "Test script")
	if err != nil {
		return nil, err
	}
	err = writer.WriteField("project", project)
	if err != nil {
		return nil, err
	}
	err = writer.WriteField("label", label)
	if err != nil {
		return nil, err
	}
	err = writer.WriteField("language", language)
	if err != nil {
		return nil, err
	}
	err = writer.WriteField("status", "true")
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func (s *ScriptHandlerTestSuite) TestCreateScript_Success() {
	filename := fmt.Sprintf("test-create-%d.sh", uuid.New().ID())
	req, err := s.createMultipartRequest("POST", "/api/v1/jobs/script", filename, "test-project", "test-label", "bash")
	s.Require().NoError(err, "创建请求失败")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestCreateScript_InvalidRequest() {
	req := httptest.NewRequest("POST", "/api/v1/jobs/script", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ScriptHandlerTestSuite) TestCreateScript_MissingProject() {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.sh")
	s.Require().NoError(err)
	_, err = part.Write([]byte("#!/bin/bash"))
	s.Require().NoError(err)
	s.Require().NoError(writer.WriteField("label", "test"))
	s.Require().NoError(writer.WriteField("language", "bash"))
	s.Require().NoError(writer.Close())

	req := httptest.NewRequest("POST", "/api/v1/jobs/script", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestGetScript_Success() {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/script/%d", s.testScriptID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestGetScript_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/script/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ScriptHandlerTestSuite) TestGetScript_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/script/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ScriptHandlerTestSuite) TestUpdateScript_Success() {
	filename := fmt.Sprintf("test-update-%d.sh", uuid.New().ID())
	req, err := s.createMultipartRequest("PUT", fmt.Sprintf("/api/v1/jobs/script/%d", s.testScriptID), filename, "test-project", "test-label", "bash")
	s.Require().NoError(err, "创建请求失败")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestUpdateScript_NotFound() {
	filename := fmt.Sprintf("test-update-%d.sh", uuid.New().ID())
	req, err := s.createMultipartRequest("PUT", "/api/v1/jobs/script/999999", filename, "test-project", "test-label", "bash")
	s.Require().NoError(err, "创建请求失败")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ScriptHandlerTestSuite) TestDeleteScript_Success() {
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/jobs/script/%d", s.testScriptID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestDeleteScript_NotFound() {
	req := httptest.NewRequest("DELETE", "/api/v1/jobs/script/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ScriptHandlerTestSuite) TestDeleteScript_InvalidID() {
	req := httptest.NewRequest("DELETE", "/api/v1/jobs/script/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ScriptHandlerTestSuite) TestListScript_Success() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/script", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestListScript_WithFilter() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/script?project=test-project", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestDownloadScript_Success() {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/script/%d/download", s.testScriptID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.True(strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment"))
}

func (s *ScriptHandlerTestSuite) TestDownloadScript_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/script/999999/download", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ScriptHandlerTestSuite) TestListProject_Success() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/script/project", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestListLabel_Success() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/script/label", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestCreateScript_MissingLabel() {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.sh")
	s.Require().NoError(err)
	_, err = part.Write([]byte("#!/bin/bash"))
	s.Require().NoError(err)
	s.Require().NoError(writer.WriteField("project", "test"))
	s.Require().NoError(writer.WriteField("language", "bash"))
	s.Require().NoError(writer.Close())

	req := httptest.NewRequest("POST", "/api/v1/jobs/script", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestCreateScript_MissingLanguage() {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.sh")
	s.Require().NoError(err)
	_, err = part.Write([]byte("#!/bin/bash"))
	s.Require().NoError(err)
	s.Require().NoError(writer.WriteField("project", "test"))
	s.Require().NoError(writer.WriteField("label", "test"))
	s.Require().NoError(writer.Close())

	req := httptest.NewRequest("POST", "/api/v1/jobs/script", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestListScript_EmptyResult() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/script?project=nonexistent", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestListScript_WithPagination() {
	req := httptest.NewRequest("GET", "/api/v1/jobs/script?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *ScriptHandlerTestSuite) TestDownloadScript_FileNotExist() {
	ctx := context.Background()
	scriptModel, err := s.scriptRepo.GetModel(ctx, s.testScriptID)
	s.Require().NoError(err, "查询脚本失败")

	scriptPath := jobsvc.GetScriptStoragePath(scriptModel.Project, scriptModel.Label, scriptModel.Name, scriptModel.IsBuiltin)
	os.Remove(scriptPath)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/script/%d/download", s.testScriptID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

func TestScriptHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ScriptHandlerTestSuite))
}
