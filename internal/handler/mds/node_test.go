package mds

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
	"gorm.io/gorm"

	mdsmodel "gin-artweb/internal/model/mds"
	resomodel "gin-artweb/internal/model/resource"
	mdsrepo "gin-artweb/internal/repo/mds"
	resorepo "gin-artweb/internal/repo/resource"
	mdssvc "gin-artweb/internal/service/mds"
	"gin-artweb/internal/shared/test"
)

func createTestMdsColonySetup() *mdsmodel.MdsColonyModel {
	return &mdsmodel.MdsColonyModel{
		ColonyNum:     nextTestMdsColonyNum(),
		ExtractedName: "mds-test",
		IsEnable:      true,
	}
}

func createTestHostForMds() *resomodel.HostModel {
	return &resomodel.HostModel{
		Name:    fmt.Sprintf("host-mds-%s", uuid.NewString()[:8]),
		Label:   "mds-test",
		SSHIP:   "192.168.1.1",
		SSHPort: 22,
		SSHUser: "root",
	}
}

func createTestMdsNodeUpsertDTO(colonyID, hostID uint32) mdsmodel.MdsNodeUpsertDTO {
	return mdsmodel.MdsNodeUpsertDTO{
		NodeRole:    "master",
		IsEnable:    true,
		MdsColonyID: colonyID,
		HostID:      hostID,
	}
}

type MdsNodeHandlerTestSuite struct {
	suite.Suite
	db           *gorm.DB
	router       *gin.Engine
	handler      *MdsNodeHandler
	nodeRepo     *mdsrepo.MdsNodeRepo
	colonyRepo   *mdsrepo.MdsColonyRepo
	hostRepo     *resorepo.HostRepo
	testNodeID   uint32
	testColonyID uint32
	testHostID   uint32
}

func (s *MdsNodeHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.db = test.NewNamedTestGormDBWithConfig("handler_mds_node", nil)
	s.Require().NoError(s.db.AutoMigrate(
		&mdsmodel.MdsNodeModel{},
		&mdsmodel.MdsColonyModel{},
		&resomodel.HostModel{},
	))

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()

	s.nodeRepo = mdsrepo.NewMdsNodeRepo(logger, s.db, dbTimeout, slowThreshold)
	s.colonyRepo = mdsrepo.NewMdsColonyRepo(logger, s.db, dbTimeout, slowThreshold)
	s.hostRepo = resorepo.NewHostRepo(logger, s.db, dbTimeout, slowThreshold)

	nodeSvc := mdssvc.NewMdsNodeService(logger, s.nodeRepo)
	s.handler = NewMdsNodeHandler(logger, nodeSvc)

	s.router = gin.New()
	group := s.router.Group("/api/v1/mds")
	s.handler.LoadRouter(group)
}

func (s *MdsNodeHandlerTestSuite) TearDownSuite() {
	s.Require().NoError(test.CloseTestGormDB(s.db))
}

func (s *MdsNodeHandlerTestSuite) SetupTest() {
	s.Require().NoError(s.db.Exec("DELETE FROM mds_node").Error)
	s.Require().NoError(s.db.Exec("DELETE FROM mds_colony").Error)
	s.Require().NoError(s.db.Exec("DELETE FROM resource_host").Error)

	colony := createTestMdsColonySetup()
	s.Require().NoError(s.colonyRepo.CreateModel(context.Background(), colony))
	s.testColonyID = colony.ID

	host := createTestHostForMds()
	s.Require().NoError(s.hostRepo.CreateModel(context.Background(), host))
	s.testHostID = host.ID

	nodeDTO := createTestMdsNodeUpsertDTO(s.testColonyID, s.testHostID)
	node := nodeDTO.ToModel()
	s.Require().NoError(s.nodeRepo.CreateModel(context.Background(), &node))
	s.testNodeID = node.ID
}

func (s *MdsNodeHandlerTestSuite) TestNewMdsNodeHandler() {
	logger := test.NewTestZapLogger()

	handler := NewMdsNodeHandler(logger, nil)

	s.NotNil(handler)
	s.NotNil(handler.log)
}

func (s *MdsNodeHandlerTestSuite) TestCreateMdsNode_InvalidRequest() {
	invalidDTO := map[string]any{"node_role": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/mds/node", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsNodeHandlerTestSuite) TestUpdateMdsNode_InvalidRequest() {
	invalidDTO := map[string]any{"node_role": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/mds/node/%d", s.testNodeID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsNodeHandlerTestSuite) TestUpdateMdsNode_InvalidID() {
	dto := createTestMdsNodeUpsertDTO(s.testColonyID, s.testHostID)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("PUT", "/api/v1/mds/node/invalid", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsNodeHandlerTestSuite) TestGetMdsNode_Success() {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/mds/node/%d", s.testNodeID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp mdsmodel.MdsNodeResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(s.testNodeID, resp.Data.ID)
}

func (s *MdsNodeHandlerTestSuite) TestGetMdsNode_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/mds/node/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsNodeHandlerTestSuite) TestGetMdsNode_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/mds/node/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsNodeHandlerTestSuite) TestListMdsNode_Success() {
	for i := 0; i < 3; i++ {
		nodeDTO := createTestMdsNodeUpsertDTO(s.testColonyID, s.testHostID)
		node := nodeDTO.ToModel()
		s.Require().NoError(s.nodeRepo.CreateModel(context.Background(), &node))
	}

	req := httptest.NewRequest("GET", "/api/v1/mds/node?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp mdsmodel.PagMdsNodeResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(4))
}

func (s *MdsNodeHandlerTestSuite) TestListMdsNode_WithFilter() {
	req := httptest.NewRequest("GET", "/api/v1/mds/node?node_role=arbiter", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp mdsmodel.PagMdsNodeResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
}

func (s *MdsNodeHandlerTestSuite) TestDeleteMdsNode_InvalidID() {
	req := httptest.NewRequest("DELETE", "/api/v1/mds/node/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func TestMdsNodeHandlerTestSuite(t *testing.T) {
	suite.Run(t, &MdsNodeHandlerTestSuite{})
}
