package mon

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	monsvc "gin-artweb/internal/service/mon"
	"gin-artweb/internal/shared/test"
)

type MonConfHandlerTestSuite struct {
	suite.Suite
	router  *gin.Engine
	handler *MonConfHandler
	nodeID  uint32
	confDir string
}

func (s *MonConfHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.handler = NewMonConfHandler(test.NewTestZapLogger(), 1024*1024)
	s.router = gin.New()
	s.handler.LoadRouter(s.router.Group("/api/v1/mon"))

	s.nodeID = uint32(100000 + len(uuid.NewString()))
	s.confDir = monsvc.GetMonNodeConfDir(s.nodeID)
	s.Require().NoError(os.MkdirAll(s.confDir, 0750))
	s.T().Cleanup(func() { _ = os.RemoveAll(s.confDir) })
}

func (s *MonConfHandlerTestSuite) TestNewMonConfHandler() {
	logger := test.NewTestZapLogger()
	handler := NewMonConfHandler(logger, 500)

	s.NotNil(handler)
	s.Same(logger, handler.log)
	s.Equal(int64(500), handler.maxSize)
}

func (s *MonConfHandlerTestSuite) TestUploadListDownloadDeleteMonConf() {
	filename := "mon-" + uuid.NewString() + ".yaml"
	content := []byte("name: test-mon\nhost_id: 1\n")

	status, body := s.upload(filename, content)
	s.Equal(http.StatusOK, status)
	s.Contains(body, `"code":200`)

	s.FileExists(filepath.Join(s.confDir, filename))
	s.Equal(content, s.readFile(filepath.Join(s.confDir, filename)))

	listResp := s.request(http.MethodGet, "/api/v1/mon/conf/"+strconv.FormatUint(uint64(s.nodeID), 10), nil, "")
	s.Equal(http.StatusOK, listResp.Code)
	s.Contains(listResp.Body.String(), filename)

	downloadResp := s.request(
		http.MethodGet,
		"/api/v1/mon/conf/"+strconv.FormatUint(uint64(s.nodeID), 10)+"/download?filename="+filename,
		nil,
		"",
	)
	s.Equal(http.StatusOK, downloadResp.Code)
	s.Equal(content, downloadResp.Body.Bytes())

	deleteResp := s.request(
		http.MethodDelete,
		"/api/v1/mon/conf/"+strconv.FormatUint(uint64(s.nodeID), 10)+"?filename="+filename,
		nil,
		"",
	)
	s.Equal(http.StatusOK, deleteResp.Code)
	s.NoFileExists(filepath.Join(s.confDir, filename))
}

func (s *MonConfHandlerTestSuite) TestUploadMonConf_MissingFile() {
	resp := s.request(
		http.MethodPost,
		"/api/v1/mon/conf/"+strconv.FormatUint(uint64(s.nodeID), 10),
		bytes.NewBufferString(""),
		"multipart/form-data; boundary=missing",
	)
	s.NotEqual(http.StatusOK, resp.Code)
}

func (s *MonConfHandlerTestSuite) TestDownloadMonConf_MissingFilename() {
	resp := s.request(
		http.MethodGet,
		"/api/v1/mon/conf/"+strconv.FormatUint(uint64(s.nodeID), 10)+"/download",
		nil,
		"",
	)
	s.NotEqual(http.StatusOK, resp.Code)
}

func (s *MonConfHandlerTestSuite) TestDeleteMonConf_MissingFilename() {
	resp := s.request(
		http.MethodDelete,
		"/api/v1/mon/conf/"+strconv.FormatUint(uint64(s.nodeID), 10),
		nil,
		"",
	)
	s.NotEqual(http.StatusOK, resp.Code)
}

func (s *MonConfHandlerTestSuite) TestListMonConf_InvalidID() {
	resp := s.request(http.MethodGet, "/api/v1/mon/conf/invalid", nil, "")
	s.NotEqual(http.StatusOK, resp.Code)
}

func (s *MonConfHandlerTestSuite) TestRoutesDoNotContainNodePrefix() {
	filename := "route-check-" + uuid.NewString() + ".yaml"
	status, _ := s.upload(filename, []byte("test"))
	s.Equal(http.StatusOK, status)

	resp := s.request(http.MethodGet, "/api/v1/mon/node/conf/"+strconv.FormatUint(uint64(s.nodeID), 10), nil, "")
	s.Equal(http.StatusNotFound, resp.Code)
}

func (s *MonConfHandlerTestSuite) upload(filename string, content []byte) (int, string) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	s.Require().NoError(err)
	_, err = part.Write(content)
	s.Require().NoError(err)
	s.Require().NoError(writer.Close())

	resp := s.request(
		http.MethodPost,
		"/api/v1/mon/conf/"+strconv.FormatUint(uint64(s.nodeID), 10),
		&body,
		writer.FormDataContentType(),
	)
	return resp.Code, resp.Body.String()
}

func (s *MonConfHandlerTestSuite) request(method, path string, body io.Reader, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)
	return resp
}

func (s *MonConfHandlerTestSuite) readFile(path string) []byte {
	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	return content
}

func TestMonConfHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MonConfHandlerTestSuite))
}
