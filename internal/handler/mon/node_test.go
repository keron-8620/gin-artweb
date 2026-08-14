package mon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	monmodel "gin-artweb/internal/model/mon"
	resomodel "gin-artweb/internal/model/resource"
	monrepo "gin-artweb/internal/repo/mon"
	resorepo "gin-artweb/internal/repo/resource"
	monsvc "gin-artweb/internal/service/mon"
	resosvc "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/test"
)

func createTestMonNodeUpsertDTO(hostID uint32) monmodel.MonNodeUpsertDTO {
	return monmodel.MonNodeUpsertDTO{
		Name:        fmt.Sprintf("mon-node-%s", uuid.NewString()[:8]),
		DeployPath:  "/opt/mon",
		OutportPath: "/data/mon",
		JavaHome:    "/usr/lib/jvm/java-11",
		URL:         fmt.Sprintf("http://%s.example.com:8080", uuid.NewString()[:8]),
		HostID:      hostID,
	}
}

func createTestHostModel() *resomodel.HostModel {
	return &resomodel.HostModel{
		Name:    fmt.Sprintf("host-%s", uuid.NewString()[:8]),
		Label:   "test",
		SSHIP:   "192.168.1.1",
		SSHPort: 22,
		SSHUser: "root",
		PyPath:  "/usr/bin/python3",
	}
}

type MonNodeHandlerTestSuite struct {
	suite.Suite
	router  *gin.Engine
	handler *NodeHandler
	hostID  uint32
}

func (s *MonNodeHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	_ = db.AutoMigrate(&monmodel.MonNodeModel{}, &resomodel.HostModel{})

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()

	hostRepo := resorepo.NewHostRepo(logger, db, dbTimeout, slowThreshold)
	nodeRepo := monrepo.NewMonNodeRepo(logger, db, dbTimeout, slowThreshold)
	pkgRepo := resorepo.NewPackageRepo(logger, db, dbTimeout, slowThreshold)
	pkgSvc := resosvc.NewPackageService(logger, pkgRepo, filepath.Join(config.StorageDir, "packages"))
	nodeSvc := monsvc.NewMonNodeService(logger, nodeRepo, pkgSvc)

	s.handler = NewNodeHandler(logger, nodeSvc)

	s.router = gin.Default()
	group := s.router.Group("/api/v1/mon")
	s.handler.LoadRouter(group)

	host := createTestHostModel()
	_ = hostRepo.CreateModel(context.Background(), host)
	s.hostID = host.ID
}

func (s *MonNodeHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	_ = db.Exec("DELETE FROM mon_node").Error
	_ = db.Exec("DELETE FROM resource_host").Error
	_ = db.AutoMigrate(&monmodel.MonNodeModel{}, &resomodel.HostModel{})
}

func (s *MonNodeHandlerTestSuite) TestNewNodeHandler() {
	logger := test.NewTestZapLogger()

	handler := NewNodeHandler(logger, nil)

	s.NotNil(handler)
	s.NotNil(handler.log)
}

func (s *MonNodeHandlerTestSuite) TestCreateMonNode_InvalidRequest() {
	invalidDTO := map[string]any{"name": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/mon/node", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusCreated, w.Code)
}

func (s *MonNodeHandlerTestSuite) TestUpdateMonNode_InvalidRequest() {
	invalidDTO := map[string]any{"name": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("PUT", "/api/v1/mon/node/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MonNodeHandlerTestSuite) TestUpdateMonNode_InvalidID() {
	dto := createTestMonNodeUpsertDTO(s.hostID)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("PUT", "/api/v1/mon/node/invalid", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MonNodeHandlerTestSuite) TestDeleteMonNode_InvalidID() {
	req := httptest.NewRequest("DELETE", "/api/v1/mon/node/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MonNodeHandlerTestSuite) TestGetMonNode_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/mon/node/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MonNodeHandlerTestSuite) TestGetMonNode_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/mon/node/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MonNodeHandlerTestSuite) TestListMonNode_Success() {
	req := httptest.NewRequest("GET", "/api/v1/mon/node", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp monmodel.PagMonNodeResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
}

func (s *MonNodeHandlerTestSuite) TestListMonNode_WithFilter() {
	req := httptest.NewRequest("GET", "/api/v1/mon/node?name=non-existent", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp monmodel.PagMonNodeResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
}

func TestMonNodeHandlerTestSuite(t *testing.T) {
	suite.Run(t, &MonNodeHandlerTestSuite{})
}
