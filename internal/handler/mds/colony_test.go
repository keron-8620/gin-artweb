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

	mdsmodel "gin-artweb/internal/model/mds"
	monmodel "gin-artweb/internal/model/mon"
	resomodel "gin-artweb/internal/model/resource"
	mdsrepo "gin-artweb/internal/repo/mds"
	monrepo "gin-artweb/internal/repo/mon"
	resorepo "gin-artweb/internal/repo/resource"
	mdssvc "gin-artweb/internal/service/mds"
	"gin-artweb/internal/shared/test"
)

func createTestPackageForMds() *resomodel.PackageModel {
	return &resomodel.PackageModel{
		Label:           "mds-pkg",
		StorageFilename: fmt.Sprintf("mds-%s.tar.gz", uuid.NewString()[:8]),
		OriginFilename:  fmt.Sprintf("mds-orig-%s.tar.gz", uuid.NewString()[:8]),
		Version:         "1.0.0",
	}
}

func createTestMonNodeForMds(hostID uint32) *monmodel.MonNodeModel {
	return &monmodel.MonNodeModel{
		Name:        fmt.Sprintf("mon-%s", uuid.NewString()[:8]),
		DeployPath:  "/opt/mon",
		OutportPath: "/data/mon",
		JavaHome:    "/usr/lib/jvm/java-11",
		URL:         fmt.Sprintf("http://mon-%s:8080", uuid.NewString()[:8]),
		HostID:      hostID,
	}
}

func createTestMdsColonyDTO(packageID, monNodeID uint32) mdsmodel.MdsColonyUpsertDTO {
	return mdsmodel.MdsColonyUpsertDTO{
		ColonyNum:     fmt.Sprintf("0%d", uuid.New().ID()%100),
		ExtractedName: "mds-extracted",
		IsEnable:      true,
		PackageID:     packageID,
		MonNodeID:     monNodeID,
	}
}

type MdsColonyHandlerTestSuite struct {
	suite.Suite
	router        *gin.Engine
	handler       *MdsColonyHandler
	colonyRepo    *mdsrepo.MdsColonyRepo
	packageRepo   *resorepo.PackageRepo
	monNodeRepo   *monrepo.MonNodeRepo
	hostRepo      *resorepo.HostRepo
	testColonyID  uint32
	testPackageID uint32
	testMonNodeID uint32
	testHostID    uint32
}

func (s *MdsColonyHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	_ = db.AutoMigrate(
		&mdsmodel.MdsColonyModel{},
		&mdsmodel.MdsNodeModel{},
		&monmodel.MonNodeModel{},
		&resomodel.PackageModel{},
		&resomodel.HostModel{},
	)

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()

	s.colonyRepo = mdsrepo.NewMdsColonyRepo(logger, db, dbTimeout, slowThreshold)
	s.packageRepo = resorepo.NewPackageRepo(logger, db, dbTimeout, slowThreshold)
	s.monNodeRepo = monrepo.NewMonNodeRepo(logger, db, dbTimeout, slowThreshold)
	s.hostRepo = resorepo.NewHostRepo(logger, db, dbTimeout, slowThreshold)

	colonySvc := mdssvc.NewMdsColonyService(logger, s.colonyRepo, nil)
	s.handler = NewMdsColonyHandler(logger, colonySvc, nil)

	s.router = gin.Default()
	group := s.router.Group("/api/v1/mds")
	s.handler.LoadRouter(group)

	host := &resomodel.HostModel{
		Name:    "mds-host",
		Label:   "mds",
		SSHIP:   "192.168.1.1",
		SSHPort: 22,
		SSHUser: "root",
	}
	_ = s.hostRepo.CreateModel(context.Background(), host)
	s.testHostID = host.ID

	pkg := createTestPackageForMds()
	_ = s.packageRepo.CreateModel(context.Background(), pkg)
	s.testPackageID = pkg.ID

	monNode := createTestMonNodeForMds(s.testHostID)
	_ = s.monNodeRepo.CreateModel(context.Background(), monNode)
	s.testMonNodeID = monNode.ID

	colonyDTO := createTestMdsColonyDTO(s.testPackageID, s.testMonNodeID)
	colony := colonyDTO.ToModel()
	_ = s.colonyRepo.CreateModel(context.Background(), &colony)
	s.testColonyID = colony.ID
}

func (s *MdsColonyHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	_ = db.Exec("DELETE FROM mds_node").Error
	_ = db.Exec("DELETE FROM mds_colony").Error
	_ = db.Exec("DELETE FROM mon_node").Error
	_ = db.Exec("DELETE FROM resource_package").Error
	_ = db.Exec("DELETE FROM resource_host").Error
	_ = db.AutoMigrate(
		&mdsmodel.MdsColonyModel{},
		&mdsmodel.MdsNodeModel{},
		&monmodel.MonNodeModel{},
		&resomodel.PackageModel{},
		&resomodel.HostModel{},
	)

	host := &resomodel.HostModel{
		Name:    "mds-host",
		Label:   "mds",
		SSHIP:   "192.168.1.1",
		SSHPort: 22,
		SSHUser: "root",
	}
	_ = s.hostRepo.CreateModel(context.Background(), host)
	s.testHostID = host.ID

	pkg := createTestPackageForMds()
	_ = s.packageRepo.CreateModel(context.Background(), pkg)
	s.testPackageID = pkg.ID

	monNode := createTestMonNodeForMds(s.testHostID)
	_ = s.monNodeRepo.CreateModel(context.Background(), monNode)
	s.testMonNodeID = monNode.ID

	colonyDTO := createTestMdsColonyDTO(s.testPackageID, s.testMonNodeID)
	colony := colonyDTO.ToModel()
	_ = s.colonyRepo.CreateModel(context.Background(), &colony)
	s.testColonyID = colony.ID
}

func (s *MdsColonyHandlerTestSuite) TestNewMdsColonyHandler() {
	logger := test.NewTestZapLogger()

	handler := NewMdsColonyHandler(logger, nil, nil)

	s.NotNil(handler)
	s.NotNil(handler.log)
}

func (s *MdsColonyHandlerTestSuite) TestCreateMdsColony_InvalidRequest() {
	invalidDTO := map[string]any{"colony_num": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/mds/colony", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsColonyHandlerTestSuite) TestGetMdsColony_Success() {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/mds/colony/%d", s.testColonyID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp mdsmodel.MdsColonyResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(s.testColonyID, resp.Data.ID)
}

func (s *MdsColonyHandlerTestSuite) TestGetMdsColony_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/mds/colony/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsColonyHandlerTestSuite) TestGetMdsColony_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/mds/colony/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsColonyHandlerTestSuite) TestUpdateMdsColony_InvalidRequest() {
	invalidDTO := map[string]any{"colony_num": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/mds/colony/%d", s.testColonyID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsColonyHandlerTestSuite) TestUpdateMdsColony_InvalidID() {
	dto := createTestMdsColonyDTO(s.testPackageID, s.testMonNodeID)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("PUT", "/api/v1/mds/colony/invalid", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsColonyHandlerTestSuite) TestListMdsColony_Success() {
	for i := 0; i < 3; i++ {
		colonyDTO := createTestMdsColonyDTO(s.testPackageID, s.testMonNodeID)
		colony := colonyDTO.ToModel()
		colony.ColonyNum = fmt.Sprintf("0%d", (i+5)*10)
		_ = s.colonyRepo.CreateModel(context.Background(), &colony)
	}

	req := httptest.NewRequest("GET", "/api/v1/mds/colony?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp mdsmodel.PagMdsColonyResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(4))
}

func (s *MdsColonyHandlerTestSuite) TestListMdsColony_WithFilter() {
	req := httptest.NewRequest("GET", "/api/v1/mds/colony?colony_num=99", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp mdsmodel.PagMdsColonyResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
}

func (s *MdsColonyHandlerTestSuite) TestDeleteMdsColony_InvalidID() {
	req := httptest.NewRequest("DELETE", "/api/v1/mds/colony/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func TestMdsColonyHandlerTestSuite(t *testing.T) {
	suite.Run(t, &MdsColonyHandlerTestSuite{})
}
