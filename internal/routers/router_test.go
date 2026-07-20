package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDiagnosticsAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		expected   string
		provided   string
		wantStatus int
	}{
		{name: "有效Bearer", expected: "diagnostics-secret-token-1234567890", provided: "Bearer diagnostics-secret-token-1234567890", wantStatus: http.StatusOK},
		{name: "有效原始Token", expected: "diagnostics-secret-token-1234567890", provided: "diagnostics-secret-token-1234567890", wantStatus: http.StatusOK},
		{name: "错误Token", expected: "diagnostics-secret-token-1234567890", provided: "wrong", wantStatus: http.StatusUnauthorized},
		{name: "服务端未配置Token", expected: "", provided: "anything", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/metrics", diagnosticsAuth(tt.expected), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			if tt.provided != "" {
				req.Header.Set("Authorization", tt.provided)
			}
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantStatus, resp.Code)
		})
	}
}
