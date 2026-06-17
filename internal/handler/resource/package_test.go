package resource

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

	resomodel "gin-artweb/internal/model/resource"
	resorepo "gin-artweb/internal/repo/resource"
	resosvc "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/test"
)

func createTestHandlerPackageModel() *resomodel.PackageModel {
	return &resomodel.PackageModel{
		Label:           fmt.Sprintf("test-label-%s", uuid.NewString()[:8]),
		StorageFilename: fmt.Sprintf("test-%s.tar.gz", uuid.NewString()[:8]),
		OriginFilename:  fmt.Sprintf("original-%s.tar.gz", uuid.NewString()[:8]),
		Version:         "1.0.0",
	}
}

type PackageHandlerTestSuite struct {
	suite.Suite
	router        *gin.Engine
	handler       *PackageHandler
	packageRepo   *resorepo.PackageRepo
	testPackageID uint32
}

func (s *PackageHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	_ = db.AutoMigrate(&resomodel.PackageModel{})

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()

	s.packageRepo = resorepo.NewPackageRepo(logger, db, dbTimeout, slowThreshold)
	pkgSvc := resosvc.NewPackageService(logger, s.packageRepo, "/tmp")
	s.handler = NewPackageHandler(logger, pkgSvc, 500)

	s.router = gin.Default()
	group := s.router.Group("/api/v1/resource")
	s.handler.LoadRouter(group)
}

func (s *PackageHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	_ = db.Exec("DELETE FROM resource_package").Error

	pkg := createTestHandlerPackageModel()
	_ = s.packageRepo.CreateModel(context.Background(), pkg)
	s.testPackageID = pkg.ID
}

func (s *PackageHandlerTestSuite) TestNewPackageHandler() {
	logger := test.NewTestZapLogger()

	handler := NewPackageHandler(logger, nil, 500)

	s.NotNil(handler)
	s.NotNil(handler.log)
}

func (s *PackageHandlerTestSuite) TestDeletePackage_InvalidID() {
	req := httptest.NewRequest("DELETE", "/api/v1/resource/package/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *PackageHandlerTestSuite) TestDeletePackage_NotFound() {
	req := httptest.NewRequest("DELETE", "/api/v1/resource/package/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *PackageHandlerTestSuite) TestGetPackage_Success() {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/resource/package/%d", s.testPackageID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp resomodel.PackageResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(s.testPackageID, resp.Data.ID)
}

func (s *PackageHandlerTestSuite) TestGetPackage_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/resource/package/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *PackageHandlerTestSuite) TestGetPackage_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/resource/package/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *PackageHandlerTestSuite) TestListPackage_Success() {
	for i := 0; i < 3; i++ {
		pkg := createTestHandlerPackageModel()
		_ = s.packageRepo.CreateModel(context.Background(), pkg)
	}

	req := httptest.NewRequest("GET", "/api/v1/resource/package?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp resomodel.PagPackageResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(1))
}

func (s *PackageHandlerTestSuite) TestListPackage_EmptyResult() {
	req := httptest.NewRequest("GET", "/api/v1/resource/package?label=non-existent-label-xyz", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp resomodel.PagPackageResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
}

func (s *PackageHandlerTestSuite) TestUploadPackage_InvalidRequest() {
	invalidDTO := map[string]any{"label": ""}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/resource/package", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusCreated, w.Code)
}

func TestPackageHandlerTestSuite(t *testing.T) {
	suite.Run(t, &PackageHandlerTestSuite{})
}
