package sys

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
	"github.com/stretchr/testify/suite"

	sysmodel "gin-artweb/internal/model/sys"
	sysrepo "gin-artweb/internal/repo/sys"
	syssvc "gin-artweb/internal/service/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/test"
)

func createTestApiDTO(id uint32) sysmodel.CreateApiDTO {
	return sysmodel.CreateApiDTO{
		ID:     id,
		URL:    fmt.Sprintf("/api/test/%s/", uuid.NewString()),
		Method: "GET",
		Label:  "test_label",
		Descr:  "test_description",
	}
}

func createTestApiModel(id uint32) *sysmodel.ApiModel {
	return &sysmodel.ApiModel{
		URL:    fmt.Sprintf("/api/test/%s/", uuid.NewString()),
		Method: "GET",
		Label:  "test_label",
		Descr:  "test_description",
	}
}

type ApiHandlerTestSuite struct {
	suite.Suite
	router     *gin.Engine
	handler    *ApiHandler
	apiService *syssvc.ApiService
	apiRepo    *sysrepo.ApiRepo
}

func (s *ApiHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.AutoMigrate(&sysmodel.ApiModel{}), "数据库迁移失败")

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, err := auth.NewCasbinEnforcer()
	s.Require().NoError(err, "创建Casbin enforcer失败")

	s.apiRepo = sysrepo.NewApiRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.apiService = syssvc.NewApiService(logger, s.apiRepo)
	s.handler = NewApiHandler(logger, s.apiService)

	s.router = gin.Default()
	apiGroup := s.router.Group("/api/v1/customer")
	s.handler.LoadRouter(apiGroup)
}

func (s *ApiHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.Exec("DELETE FROM sys_api").Error, "清理测试数据失败")
}

func (s *ApiHandlerTestSuite) TestCreateApi_Success() {
	testID := uint32(uuid.New().ID())
	dto := createTestApiDTO(testID)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("POST", "/api/v1/customer/api", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var resp sysmodel.ApiResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusCreated, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(dto.URL, resp.Data.URL)
	s.Equal(dto.Method, resp.Data.Method)
	s.Equal(dto.Label, resp.Data.Label)
	s.Equal(dto.Descr, resp.Data.Descr)
}

func (s *ApiHandlerTestSuite) TestCreateApi_InvalidRequest() {
	invalidDTO := map[string]any{
		"id": 0,
	}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/customer/api", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusCreated, w.Code)
}

func (s *ApiHandlerTestSuite) TestGetApi_Success() {
	testID := uint32(uuid.New().ID())
	apiModel := createTestApiModel(testID)
	apiModel.ID = testID
	s.Require().NoError(s.apiRepo.CreateModel(context.Background(), apiModel))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/customer/api/%d", testID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.ApiResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(testID, resp.Data.ID)
}

func (s *ApiHandlerTestSuite) TestGetApi_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/customer/api/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *ApiHandlerTestSuite) TestUpdateApi_Success() {
	testID := uint32(uuid.New().ID())
	apiModel := createTestApiModel(testID)
	apiModel.ID = testID
	s.Require().NoError(s.apiRepo.CreateModel(context.Background(), apiModel))

	updateDTO := sysmodel.UpdateApiDTO{
		URL:    "/api/updated/",
		Method: "POST",
		Label:  "updated_label",
		Descr:  "updated_description",
	}
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/customer/api/%d", testID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.ApiResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(updateDTO.URL, resp.Data.URL)
	s.Equal(updateDTO.Method, resp.Data.Method)
	s.Equal(updateDTO.Label, resp.Data.Label)
	s.Equal(updateDTO.Descr, resp.Data.Descr)
}

func (s *ApiHandlerTestSuite) TestUpdateApi_NotFound() {
	updateDTO := sysmodel.UpdateApiDTO{
		URL:    "/api/updated/",
		Method: "POST",
		Label:  "updated_label",
	}
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", "/api/v1/customer/api/999999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *ApiHandlerTestSuite) TestDeleteApi_Success() {
	testID := uint32(uuid.New().ID())
	apiModel := createTestApiModel(testID)
	apiModel.ID = testID
	s.Require().NoError(s.apiRepo.CreateModel(context.Background(), apiModel))

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/customer/api/%d", testID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	_, err := s.apiRepo.GetModel(context.Background(), testID)
	s.Error(err)
}

func (s *ApiHandlerTestSuite) TestDeleteApi_NotFound() {
	req := httptest.NewRequest("DELETE", "/api/v1/customer/api/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *ApiHandlerTestSuite) TestListApi_Success() {
	for i := 0; i < 5; i++ {
		testID := uint32(uuid.New().ID())
		apiModel := createTestApiModel(testID)
		apiModel.ID = testID
		s.Require().NoError(s.apiRepo.CreateModel(context.Background(), apiModel))
	}

	req := httptest.NewRequest("GET", "/api/v1/customer/api?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagApiResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(5))
}

func (s *ApiHandlerTestSuite) TestListApi_WithFilter() {
	testID := uint32(uuid.New().ID())
	apiModel := createTestApiModel(testID)
	apiModel.ID = testID
	apiModel.Label = "unique_label_123"
	s.Require().NoError(s.apiRepo.CreateModel(context.Background(), apiModel))

	req := httptest.NewRequest("GET", "/api/v1/customer/api?label=unique_label_123", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagApiResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(1))
}

func TestApiHandlerTestSuite(t *testing.T) {
	suite.Run(t, &ApiHandlerTestSuite{})
}

func TestNewApiHandler(t *testing.T) {
	logger := test.NewTestZapLogger()
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, _ := auth.NewCasbinEnforcer()
	apiRepo := sysrepo.NewApiRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	apiService := syssvc.NewApiService(logger, apiRepo)

	handler := NewApiHandler(logger, apiService)

	if handler == nil {
		t.Fatal("NewApiHandler should return non-nil handler")
	}
	if handler.log == nil {
		t.Fatal("Handler log should not be nil")
	}
	if handler.apiSvc == nil {
		t.Fatal("Handler apiSvc should not be nil")
	}
}
