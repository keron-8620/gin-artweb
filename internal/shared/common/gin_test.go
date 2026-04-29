package common

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestShouldBind(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Name string `json:"name" form:"name"`
		Age  int    `json:"age" form:"age"`
	}

	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		expectBind  bool
	}{
		{"valid json", http.MethodPost, "application/json", `{"name":"test","age":20}`, true},
		{"valid form", http.MethodPost, "application/x-www-form-urlencoded", "name=test&age=20", true},
		{"invalid json", http.MethodPost, "application/json", `{"name":"test"`, false},
		{"missing required", http.MethodPost, "application/json", `{"name":"test"}`, true},
		{"empty body json", http.MethodPost, "application/json", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tt.method, "/test", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", tt.contentType)

			var v TestStruct
			result := ShouldBind(c, logger, &v, "test bind error")

			if result != tt.expectBind {
				t.Errorf("ShouldBind() = %v, want %v", result, tt.expectBind)
			}
		})
	}
}

func TestShouldBindUri(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		ID   uint   `uri:"id"`
		Name string `uri:"name"`
	}

	tests := []struct {
		name       string
		path       string
		expectBind bool
	}{
		{"valid uri params", "/users/123/john", true},
		{"invalid id", "/users/abc/john", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			r.GET("/users/:id/:name", func(c *gin.Context) {
				var v TestStruct
				result := ShouldBindUri(c, logger, &v, "test bind uri error")

				if result != tt.expectBind {
					t.Errorf("ShouldBindUri() = %v, want %v", result, tt.expectBind)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			r.ServeHTTP(w, req)
		})
	}
}

func TestShouldBindQuery(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Name string `form:"name"`
		Age  int    `form:"age"`
	}

	tests := []struct {
		name       string
		query      string
		expectBind bool
	}{
		{"valid query", "/test?name=john&age=30", true},
		{"missing age", "/test?name=john", true},
		{"invalid age", "/test?name=john&age=abc", false},
		{"empty query", "/test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tt.query, nil)

			var v TestStruct
			result := ShouldBindQuery(c, logger, &v, "test bind query error")

			if result != tt.expectBind {
				t.Errorf("ShouldBindQuery() = %v, want %v", result, tt.expectBind)
			}
		})
	}
}
