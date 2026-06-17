package oes

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

	monmodel "gin-artweb/internal/model/mon"
	oesmodel "gin-artweb/internal/model/oes"
	resomodel "gin-artweb/internal/model/resource"
	monrepo "gin-artweb/internal/repo/mon"
	oesrepo "gin-artweb/internal/repo/oes"
	resorepo "gin-artweb/internal/repo/resource"
	oessvc "gin-artweb/internal/service/oes"
	"gin-artweb/internal/shared/test"
)

func createTestHostForOes() *resomodel.HostModel {
	return &resomodel.HostModel{
		Name:    fmt.Sprintf("host-oes-%s", uuid.NewString()[:8]),
		Label:   "oes-test",
		SSHIP:   "192.168.1.1",
		SSHPort: 22,
		SSHUser: "root",
	}
}

func createTestPkgForOes() *resomodel.PackageModel {
	return &resomodel.PackageModel{
		Label:           fmt.Sprintf("oes-pkg-%s", uuid.NewString()[:8]),
		StorageFilename: fmt.Sprintf("oes-%s.tar.gz", uuid.NewString()[:8]),
		OriginFilename:  fmt.Sprintf("oes-orig-%s.tar.gz", uuid.NewString()[:8]),
		Version:         "1.0.0",
	}
}

func createTestMonNodeForOes(hostID uint32) *monmodel.MonNodeModel {
	return &monmodel.MonNodeModel{
		Name:        fmt.Sprintf("mon-oes-%s", uuid.NewString()[:8]),
		DeployPath:  "/opt/mon",
		OutportPath: "/data/mon",
		JavaHome:    "/usr/lib/jvm/java-11",
		URL:         fmt.Sprintf("http://mon-oes-%s:8080", uuid.NewString()[:8]),
		HostID:      hostID,
	}
}

func createTestOesColonyDTO(packageID, xcounterID, monNodeID uint32) oesmodel.OesColonyUpsertDTO {
	return oesmodel.OesColonyUpsertDTO{
		SystemType:    "STK",
		ColonyNum:     fmt.Sprintf("0%d", uuid.New().ID()%100),
		ExtractedName: "oes-extracted",
		IsEnable:      true,
		PackageID:     packageID,
		XCounterID:    xcounterID,
		MonNodeID:     monNodeID,
	}
}

type OesColonyHandlerTestSuite struct {
	suite.Suite
	router         *gin.Engine
	handler        *OesColonyHandler
	colonyRepo     *oesrepo.OesColonyRepo
	packageRepo    *resorepo.PackageRepo
	monNodeRepo    *monrepo.MonNodeRepo
	hostRepo       *resorepo.HostRepo
	testColonyID   uint32
	testPackageID  uint32
	testXCounterID uint32
	testMonNodeID  uint32
	testHostID     uint32
}

func (s *OesColonyHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	_ = db.AutoMigrate(
		&oesmodel.OesColonyModel{},
		&monmodel.MonNodeModel{},
		&resomodel.PackageModel{},
		&resomodel.HostModel{},
	)

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()

	s.colonyRepo = oesrepo.NewOesColonyRepo(logger, db, dbTimeout, slowThreshold)
	s.packageRepo = resorepo.NewPackageRepo(logger, db, dbTimeout, slowThreshold)
	s.monNodeRepo = monrepo.NewMonNodeRepo(logger, db, dbTimeout, slowThreshold)
	s.hostRepo = resorepo.NewHostRepo(logger, db, dbTimeout, slowThreshold)

	colonySvc := oessvc.NewOesColonyService(logger, s.colonyRepo, nil)
	s.handler = NewOesColonyHandler(logger, colonySvc, nil, nil, nil)

	s.router = gin.Default()
	group := s.router.Group("/api/v1/oes")
	s.handler.LoadRouter(group)

	host := createTestHostForOes()
	_ = s.hostRepo.CreateModel(context.Background(), host)
	s.testHostID = host.ID

	pkg := createTestPkgForOes()
	_ = s.packageRepo.CreateModel(context.Background(), pkg)
	s.testPackageID = pkg.ID

	xcounter := createTestPkgForOes()
	_ = s.packageRepo.CreateModel(context.Background(), xcounter)
	s.testXCounterID = xcounter.ID

	monNode := createTestMonNodeForOes(s.testHostID)
	_ = s.monNodeRepo.CreateModel(context.Background(), monNode)
	s.testMonNodeID = monNode.ID

	colonyDTO := createTestOesColonyDTO(s.testPackageID, s.testXCounterID, s.testMonNodeID)
	colony := colonyDTO.ToModel()
	_ = s.colonyRepo.CreateModel(context.Background(), &colony)
	s.testColonyID = colony.ID
}

func (s *OesColonyHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	_ = db.Exec("DELETE FROM oes_colony").Error
	_ = db.Exec("DELETE FROM mon_node").Error
	_ = db.Exec("DELETE FROM resource_package").Error
	_ = db.Exec("DELETE FROM resource_host").Error
	_ = db.AutoMigrate(
		&oesmodel.OesColonyModel{},
		&monmodel.MonNodeModel{},
		&resomodel.PackageModel{},
		&resomodel.HostModel{},
	)

	host := createTestHostForOes()
	_ = s.hostRepo.CreateModel(context.Background(), host)
	s.testHostID = host.ID

	pkg := createTestPkgForOes()
	_ = s.packageRepo.CreateModel(context.Background(), pkg)
	s.testPackageID = pkg.ID

	xcounter := createTestPkgForOes()
	_ = s.packageRepo.CreateModel(context.Background(), xcounter)
	s.testXCounterID = xcounter.ID

	monNode := createTestMonNodeForOes(s.testHostID)
	_ = s.monNodeRepo.CreateModel(context.Background(), monNode)
	s.testMonNodeID = monNode.ID

	colonyDTO := createTestOesColonyDTO(s.testPackageID, s.testXCounterID, s.testMonNodeID)
	colony := colonyDTO.ToModel()
	_ = s.colonyRepo.CreateModel(context.Background(), &colony)
	s.testColonyID = colony.ID
}

func (s *OesColonyHandlerTestSuite) TestNewOesColonyHandler() {
	logger := test.NewTestZapLogger()

	handler := NewOesColonyHandler(logger, nil, nil, nil, nil)

	s.NotNil(handler)
	s.NotNil(handler.log)
}

func (s *OesColonyHandlerTestSuite) TestCreateOesColony_InvalidRequest() {
	invalidDTO := map[string]any{"system_type": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/oes/colony", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *OesColonyHandlerTestSuite) TestGetOesColony_Success() {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/oes/colony/%d", s.testColonyID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp oesmodel.OesColonyResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(s.testColonyID, resp.Data.ID)
}

func (s *OesColonyHandlerTestSuite) TestGetOesColony_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/oes/colony/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *OesColonyHandlerTestSuite) TestGetOesColony_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/oes/colony/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *OesColonyHandlerTestSuite) TestUpdateOesColony_InvalidRequest() {
	invalidDTO := map[string]any{"system_type": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/oes/colony/%d", s.testColonyID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *OesColonyHandlerTestSuite) TestUpdateOesColony_InvalidID() {
	dto := createTestOesColonyDTO(s.testPackageID, s.testXCounterID, s.testMonNodeID)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("PUT", "/api/v1/oes/colony/invalid", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *OesColonyHandlerTestSuite) TestListOesColony_Success() {
	for i := 0; i < 3; i++ {
		colonyDTO := createTestOesColonyDTO(s.testPackageID, s.testXCounterID, s.testMonNodeID)
		colony := colonyDTO.ToModel()
		colony.ColonyNum = fmt.Sprintf("0%d", (i+5)*10)
		_ = s.colonyRepo.CreateModel(context.Background(), &colony)
	}

	req := httptest.NewRequest("GET", "/api/v1/oes/colony?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp oesmodel.PagOesColonyResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(4))
}

func (s *OesColonyHandlerTestSuite) TestListOesColony_WithFilter() {
	req := httptest.NewRequest("GET", "/api/v1/oes/colony?system_type=STK", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp oesmodel.PagOesColonyResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
}

func (s *OesColonyHandlerTestSuite) TestDeleteOesColony_InvalidID() {
	req := httptest.NewRequest("DELETE", "/api/v1/oes/colony/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func TestOesColonyHandlerTestSuite(t *testing.T) {
	suite.Run(t, &OesColonyHandlerTestSuite{})
}
