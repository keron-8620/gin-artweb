package resource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	resomodel "gin-artweb/internal/model/resource"
	resorepo "gin-artweb/internal/repo/resource"
	resosvc "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/test"
)

func createTestHandlerHostModel() *resomodel.HostModel {
	return &resomodel.HostModel{
		Name:    fmt.Sprintf("host-%s", uuid.NewString()[:8]),
		Label:   "test-label",
		SSHIP:   "192.168.1.1",
		SSHPort: 22,
		SSHUser: "root",
		PyPath:  "/usr/bin/python3",
	}
}

func createTestHostUpsertDTO() resomodel.HostUpsertDTO {
	return resomodel.HostUpsertDTO{
		Name:        fmt.Sprintf("host-%s", uuid.NewString()[:8]),
		Label:       "test-label",
		SSHIP:       "192.168.1.1",
		SSHPort:     22,
		SSHUser:     "root",
		SSHPassword: "password",
		PyPath:      "/usr/bin/python3",
	}
}

type HostHandlerTestSuite struct {
	suite.Suite
	router   *gin.Engine
	handler  *HostHandler
	hostRepo *resorepo.HostRepo
	hostSvc  *resosvc.HostService
}

func (s *HostHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	_ = db.AutoMigrate(&resomodel.HostModel{})

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()

	s.hostRepo = resorepo.NewHostRepo(logger, db, dbTimeout, slowThreshold)
	s.hostSvc = resosvc.NewHostService(logger, s.hostRepo, time.Second*5, nil, nil)
	s.handler = NewHostHandler(logger, s.hostSvc)

	s.router = gin.Default()
	group := s.router.Group("/api/v1/resource")
	s.handler.LoadRouter(group)
}

func (s *HostHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	_ = db.Exec("DELETE FROM resource_host").Error
}

func (s *HostHandlerTestSuite) createHost() uint32 {
	host := createTestHandlerHostModel()
	_ = s.hostRepo.CreateModel(context.Background(), host)
	return host.ID
}

func (s *HostHandlerTestSuite) TestNewHostHandler() {
	logger := test.NewTestZapLogger()
	handler := NewHostHandler(logger, nil)
	s.NotNil(handler)
	s.NotNil(handler.log)
}

func (s *HostHandlerTestSuite) TestCreateHost_InvalidRequest() {
	invalidDTO := map[string]any{"name": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/resource/host", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusCreated, w.Code)
}

func (s *HostHandlerTestSuite) TestUpdateHost_InvalidRequest() {
	hostID := s.createHost()

	invalidDTO := map[string]any{"name": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/resource/host/%d", hostID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *HostHandlerTestSuite) TestUpdateHost_InvalidID() {
	dto := createTestHostUpsertDTO()
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("PUT", "/api/v1/resource/host/invalid", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *HostHandlerTestSuite) TestDeleteHost_InvalidID() {
	req := httptest.NewRequest("DELETE", "/api/v1/resource/host/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *HostHandlerTestSuite) TestGetHost_Success() {
	hostID := s.createHost()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/resource/host/%d", hostID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp resomodel.HostResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(hostID, resp.Data.ID)
}

func (s *HostHandlerTestSuite) TestGetHost_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/resource/host/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *HostHandlerTestSuite) TestGetHost_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/resource/host/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *HostHandlerTestSuite) TestListHost_Success() {
	_ = s.createHost()

	req := httptest.NewRequest("GET", "/api/v1/resource/host?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp resomodel.PagHostResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(1))
}

func (s *HostHandlerTestSuite) TestListHost_WithFilter() {
	req := httptest.NewRequest("GET", "/api/v1/resource/host?name="+string(uuid.NewString()[:8]), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp resomodel.PagHostResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
}

func (s *HostHandlerTestSuite) TestListHost_EmptyResult() {
	req := httptest.NewRequest("GET", "/api/v1/resource/host?name=non-existent-host-xyz", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp resomodel.PagHostResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
}

func TestHostHandlerTestSuite(t *testing.T) {
	suite.Run(t, &HostHandlerTestSuite{})
}
