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
	req := httptest.NewRequest("POST", "/api/v1/mds/conf/invalid", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MdsConfHandlerTestSuite) TestListMdsConf_Success() {
	// ListMdsConf for a non-existent colony directory should still work
	// (it returns 200 with empty list or 200 with error data)
	req := httptest.NewRequest("GET", "/api/v1/mds/conf/nonexistent", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	var resp map[string]interface{}
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Contains(resp, "code")
}

// TestDownloadMdsConfBinding 回归测试: 验证下载接口的参数绑定不再误报验证错误。
// 之前 DownloadOrDeleteMdsConfRequest 将 uri 与 form 的 required 字段混在一个结构体中,
// ShouldBindUri 校验整个结构体时 DirName/Filename 尚未填充即触发 required 失败,
// 返回 400 ERROR_VALIDATION_FAILED。拆分为 MdsConfUriDTO 与 MdsConfFileQueryDTO 后,
// 各自只校验自身绑定的字段。
func (s *MdsConfHandlerTestSuite) TestDownloadMdsConfBinding() {
	tests := []struct {
		name       string
		uri        string
		wantCode   int
		wantReason string
	}{
		{
			// 合法的 dir_name(允许任意文件夹)与 filename,绑定应通过;
			// 文件不存在则落到下载业务错误(404),而非参数验证错误(400)。
			name:       "valid binding then file not found",
			uri:        "/api/v1/mds/conf/01/download?dir_name=unknown&filename=not_exist.conf",
			wantCode:   http.StatusNotFound,
			wantReason: "DOWNLOAD_FILE_NOT_FOUND",
		},
		{
			name:       "missing query params",
			uri:        "/api/v1/mds/conf/01/download",
			wantCode:   http.StatusBadRequest,
			wantReason: "ERROR_VALIDATION_FAILED",
		},
		{
			// dir_name 包含路径穿透(../),应在 query 绑定阶段失败。
			name:       "path traversal dir_name",
			uri:        "/api/v1/mds/conf/01/download?dir_name=..&filename=x",
			wantCode:   http.StatusBadRequest,
			wantReason: "ERROR_VALIDATION_FAILED",
		},
		{
			// filename 包含路径分隔符(/),应在 query 绑定阶段失败。
			name:       "path traversal filename",
			uri:        "/api/v1/mds/conf/01/download?dir_name=all&filename=etc/passwd",
			wantCode:   http.StatusBadRequest,
			wantReason: "ERROR_VALIDATION_FAILED",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, tt.uri, nil)
			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)

			s.Equal(tt.wantCode, w.Code)
			var resp map[string]interface{}
			s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
			s.Equal(tt.wantReason, resp["reason"])
		})
	}
}

func TestMdsConfHandlerTestSuite(t *testing.T) {
	suite.Run(t, &MdsConfHandlerTestSuite{})
}
