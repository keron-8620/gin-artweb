package mds

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/shared/test"
)

type MdsConfHandlerTestSuite struct {
	suite.Suite
	router  *gin.Engine
	handler *MdsConfHandler
}

func (s *MdsConfHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	logger := test.NewTestZapLogger()
	s.handler = NewMdsConfHandler(logger, 500)

	s.router = gin.Default()
	group := s.router.Group("/api/v1/mds")
	s.handler.LoadRouter(group)
}

func (s *MdsConfHandlerTestSuite) TestNewMdsConfHandler() {
	logger := test.NewTestZapLogger()

	handler := NewMdsConfHandler(logger, 500)

	s.NotNil(handler)
	s.NotNil(handler.log)
}

func (s *MdsConfHandlerTestSuite) TestUploadMdsConf_InvalidURI() {
	req := httptest.NewRequest("POST", "/api/v1/mds/invalid/conf/dir", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsConfHandlerTestSuite) TestListMdsConf_Success() {
	// ListMdsConf for a non-existent colony directory should still work
	// (it returns 200 with empty list or 200 with error data)
	req := httptest.NewRequest("GET", "/api/v1/mds/nonexistent/conf", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	var resp map[string]interface{}
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Contains(resp, "code")
}

func TestMdsConfHandlerTestSuite(t *testing.T) {
	suite.Run(t, &MdsConfHandlerTestSuite{})
}
