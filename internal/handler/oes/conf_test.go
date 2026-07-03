package oes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	oesmodel "gin-artweb/internal/model/oes"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/test"
)

type OesConfHandlerTestSuite struct {
	suite.Suite
}

func (s *OesConfHandlerTestSuite) TestNewOesConfHandler() {
	logger := test.NewTestZapLogger()

	handler := NewOesConfHandler(logger, 500)

	s.NotNil(handler)
	s.NotNil(handler.log)
}

// TestDownloadOesConfBinding 回归测试: 验证下载接口能正确绑定路径参数与查询参数。
// 复现报错 URL: /api/v1/oes/conf/01/download?dir_name=all%2Ftemplate&filename=BJSBS.DBF.template
// 之前 OesConfFileDTO 将 uri 与 form 的 required 字段混在一个结构体中,
// ShouldBindUri 会对整个结构体做校验,导致 DirName/Filename 在未填充时即触发 required 失败。
// 拆分为 OesConfUriDTO 与 OesConfFileQueryDTO 后,各自只校验自身绑定的字段。
func (s *OesConfHandlerTestSuite) TestDownloadOesConfBinding() {
	logger := test.NewTestZapLogger()
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		uri        string
		wantOK     bool
		wantColony string
		wantDir    string
		wantFile   string
	}{
		{
			name:       "reported url",
			uri:        "/api/v1/oes/conf/01/download?dir_name=all%2Ftemplate&filename=BJSBS.DBF.template",
			wantOK:     true,
			wantColony: "01",
			wantDir:    "all/template",
			wantFile:   "BJSBS.DBF.template",
		},
		{
			name:   "missing query params",
			uri:    "/api/v1/oes/conf/01/download",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var gotURI oesmodel.OesConfUriDTO
			var gotQuery oesmodel.OesConfFileQueryDTO
			ok := true

			r := gin.New()
			r.GET("/api/v1/oes/conf/:colony_num/download", func(c *gin.Context) {
				if !common.ShouldBindUri(c, logger, &gotURI, "test bind uri") {
					ok = false
					return
				}
				if !common.ShouldBindQuery(c, logger, &gotQuery, "test bind query") {
					ok = false
					return
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.uri, nil)
			r.ServeHTTP(w, req)

			s.Equal(tt.wantOK, ok)
			if tt.wantOK {
				s.Equal(tt.wantColony, gotURI.ColonyNum)
				s.Equal(tt.wantDir, gotQuery.DirName)
				s.Equal(tt.wantFile, gotQuery.Filename)
			}
		})
	}
}

func TestOesConfHandlerTestSuite(t *testing.T) {
	suite.Run(t, &OesConfHandlerTestSuite{})
}
