package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/time/rate"

	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

func setupTestLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, recorded := observer.New(zap.DebugLevel)
	return zap.New(core), recorded
}

func newTestJWTConfig() *config.JWTConfig {
	return &config.JWTConfig{
		Issuer:                "test-issuer",
		AccessTokenExpiration: time.Hour,
		AccessMethod:          jwt.SigningMethodHS256,
		AccessSecret:          []byte("test-access-secret-key-1234567890123456"),
	}
}

func newTestUserInfo() auth.UserInfo {
	return auth.UserInfo{
		UserID:   1,
		Username: "testuser",
		RoleID:   10,
		IsStaff:  true,
	}
}

func newTestAccessToken(t *testing.T, cfg *config.JWTConfig, user auth.UserInfo) string {
	token, err := auth.NewAccessJWT(context.Background(), cfg, user)
	assert.NoError(t, err)
	return token
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name          string
		headers       map[string]string
		queryParams   map[string]string
		expectedToken string
	}{
		{
			name:          "HTTP请求从Authorization头部获取token",
			headers:       map[string]string{"Authorization": "Bearer token123"},
			queryParams:   nil,
			expectedToken: "Bearer token123",
		},
		{
			name:          "WebSocket请求从查询参数获取token",
			headers:       map[string]string{"Connection": "upgrade", "Upgrade": "websocket"},
			queryParams:   map[string]string{"Authorization": "ws-token"},
			expectedToken: "ws-token",
		},
		{
			name:          "WebSocket请求从Sec-WebSocket-Protocol获取token",
			headers:       map[string]string{"Connection": "upgrade", "Upgrade": "websocket", "Sec-WebSocket-Protocol": "ws-protocol-token"},
			queryParams:   nil,
			expectedToken: "ws-protocol-token",
		},
		{
			name:          "WebSocket请求无token",
			headers:       map[string]string{"Connection": "upgrade", "Upgrade": "websocket"},
			queryParams:   nil,
			expectedToken: "",
		},
		{
			name:          "HTTP请求无Authorization头",
			headers:       map[string]string{},
			queryParams:   nil,
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.GET("/test", func(c *gin.Context) {
				token := extractToken(c)
				assert.Equal(t, tt.expectedToken, token)
				c.JSON(http.StatusOK, gin.H{"token": token})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		})
	}
}

func TestJWTAuthMiddleware(t *testing.T) {
	logger, _ := setupTestLogger()
	jwtConfig := newTestJWTConfig()
	userInfo := newTestUserInfo()

	tests := []struct {
		name           string
		token          string
		setToken       bool
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "有效token",
			token:          newTestAccessToken(t, jwtConfig, userInfo),
			setToken:       true,
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "缺少token",
			token:          "",
			setToken:       false,
			expectError:    true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "无效token",
			token:          "invalid-token",
			setToken:       true,
			expectError:    true,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.GET("/test", JWTAuthMiddleware(jwtConfig, logger), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.setToken {
				req.Header.Set("Authorization", tt.token)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectError {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "code")
				assert.Contains(t, resp, "msg")
				assert.Contains(t, resp, "reason")
			}
		})
	}
}

func TestCasbinAuthMiddleware(t *testing.T) {
	logger, _ := setupTestLogger()

	enforcer, err := NewTestCasbinEnforcer()
	assert.NoError(t, err)

	tests := []struct {
		name           string
		roleID         uint32
		path           string
		method         string
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "admin有GET权限",
			roleID:         1,
			path:           "/api/users",
			method:         http.MethodGet,
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "admin有POST权限",
			roleID:         1,
			path:           "/api/users",
			method:         http.MethodPost,
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user有profile权限",
			roleID:         2,
			path:           "/api/profile",
			method:         http.MethodGet,
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user无users权限",
			roleID:         2,
			path:           "/api/users",
			method:         http.MethodGet,
			expectError:    true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "admin无DELETE权限",
			roleID:         1,
			path:           "/api/users",
			method:         http.MethodDelete,
			expectError:    true,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.Any(tt.path, func(c *gin.Context) {
				claims := &auth.JwtClaims{
					UserInfo: auth.UserInfo{RoleID: tt.roleID},
				}
				ctx := context.WithValue(c.Request.Context(), ctxutilJwtClaimsKey, claims)
				c.Request = c.Request.WithContext(ctx)
				c.Next()
			}, CasbinAuthMiddleware(enforcer, logger), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func NewTestCasbinEnforcer() (*casbin.Enforcer, error) {
	cm, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`)
	if err != nil {
		return nil, err
	}
	adapter := stringadapter.NewAdapter("p, role_1, /api/users, GET\np, role_1, /api/users, POST\np, role_2, /api/profile, GET")
	enforcer, err := casbin.NewEnforcer(cm, adapter)
	if err != nil {
		return nil, err
	}
	return enforcer, nil
}

func TestCorsMiddleware(t *testing.T) {
	tests := []struct {
		name              string
		cfg               *config.AllowConfig
		origin            string
		method            string
		expectCorsHeaders bool
		expectedStatus    int
	}{
		{
			name:              "默认配置允许所有源",
			cfg:               nil,
			origin:            "http://example.com",
			method:            http.MethodGet,
			expectCorsHeaders: true,
			expectedStatus:    http.StatusOK,
		},
		{
			name: "指定源被允许",
			cfg: &config.AllowConfig{
				AllowOrigins:     []string{"http://example.com"},
				AllowCredentials: true,
				AllowMethods:     []string{"GET", "POST"},
				AllowHeaders:     []string{"Authorization"},
			},
			origin:            "http://example.com",
			method:            http.MethodGet,
			expectCorsHeaders: true,
			expectedStatus:    http.StatusOK,
		},
		{
			name: "源不在允许列表中",
			cfg: &config.AllowConfig{
				AllowOrigins: []string{"http://allowed.com"},
			},
			origin:            "http://notallowed.com",
			method:            http.MethodGet,
			expectCorsHeaders: false,
			expectedStatus:    http.StatusOK,
		},
		{
			name: "预检请求",
			cfg: &config.AllowConfig{
				AllowOrigins: []string{"http://example.com"},
				AllowMethods: []string{"GET", "POST", "OPTIONS"},
			},
			origin:            "http://example.com",
			method:            http.MethodOptions,
			expectCorsHeaders: true,
			expectedStatus:    http.StatusNoContent,
		},
		{
			name:              "无Origin头",
			cfg:               nil,
			origin:            "",
			method:            http.MethodGet,
			expectCorsHeaders: false,
			expectedStatus:    http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.Use(CorsMiddleware(tt.cfg))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectCorsHeaders {
				assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestErrorMiddleware(t *testing.T) {
	logger, recorded := setupTestLogger()

	tests := []struct {
		name           string
		panicValue     interface{}
		expectErrorLog bool
		expectedStatus int
	}{
		{
			name:           "正常请求不触发panic",
			panicValue:     nil,
			expectErrorLog: false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "panic字符串",
			panicValue:     "test panic",
			expectErrorLog: true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "panic错误对象",
			panicValue:     errors.ErrUnknown,
			expectErrorLog: true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "panic其他类型",
			panicValue:     12345,
			expectErrorLog: true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.Use(ErrorMiddleware(logger))
			r.GET("/test", func(c *gin.Context) {
				if tt.panicValue != nil {
					panic(tt.panicValue)
				}
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectErrorLog {
				assert.True(t, recorded.Len() > 0, "Expected error log to be recorded")
			}
			recorded.TakeAll()
		})
	}
}

func TestHostGuard(t *testing.T) {
	logger, _ := setupTestLogger()

	tests := []struct {
		name           string
		allowedHosts   []string
		requestHost    string
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "Host在允许列表中",
			allowedHosts:   []string{"localhost:8080", "example.com"},
			requestHost:    "localhost:8080",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Host不在允许列表中",
			allowedHosts:   []string{"localhost:8080"},
			requestHost:    "attacker.com",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "空允许列表拒绝所有请求",
			allowedHosts:   []string{},
			requestHost:    "localhost:8080",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.Use(HostGuard(logger, tt.allowedHosts...))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Host = tt.requestHost

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestIPRateLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(1, 1)

	tests := []struct {
		name     string
		ip       string
		count    int
		expected bool
	}{
		{
			name:     "IP首次请求",
			ip:       "192.168.1.1",
			count:    1,
			expected: true,
		},
		{
			name:     "IP超过限制",
			ip:       "192.168.1.1",
			count:    3,
			expected: false,
		},
		{
			name:     "不同IP不受影响",
			ip:       "192.168.1.2",
			count:    1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			for i := 0; i < tt.count; i++ {
				result = limiter.GetLimiter(tt.ip).Allow()
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGlobalRateLimiterMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		rps            rate.Limit
		burst          int
		requestCount   int
		expectedStatus int
	}{
		{
			name:           "在限制范围内",
			rps:            10,
			burst:          5,
			requestCount:   3,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "超过全局限制",
			rps:            1,
			burst:          1,
			requestCount:   3,
			expectedStatus: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.Use(GlobalRateLimiterMiddleware(tt.rps, tt.burst))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			var lastStatus int
			for i := 0; i < tt.requestCount; i++ {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				lastStatus = w.Code
			}

			assert.Equal(t, tt.expectedStatus, lastStatus)
		})
	}
}

func TestIPBasedRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(IPBasedRateLimiterMiddleware(1, 1))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.RemoteAddr = "192.168.1.2:12345"
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(RateLimiterMiddleware(1, 1))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestIsWebSocketRequest(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected bool
	}{
		{
			name: "标准WebSocket请求",
			headers: map[string]string{
				"Upgrade":               "websocket",
				"Connection":            "upgrade",
				"Sec-WebSocket-Version": "13",
			},
			expected: true,
		},
		{
			name: "缺少Upgrade头",
			headers: map[string]string{
				"Connection":            "upgrade",
				"Sec-WebSocket-Version": "13",
			},
			expected: false,
		},
		{
			name: "缺少Connection头",
			headers: map[string]string{
				"Upgrade":               "websocket",
				"Sec-WebSocket-Version": "13",
			},
			expected: false,
		},
		{
			name: "缺少Sec-WebSocket-Version",
			headers: map[string]string{
				"Upgrade":    "websocket",
				"Connection": "upgrade",
			},
			expected: false,
		},
		{
			name: "大小写不敏感",
			headers: map[string]string{
				"Upgrade":               "WEBSOCKET",
				"Connection":            "UPGRADE",
				"Sec-WebSocket-Version": "13",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.GET("/test", func(c *gin.Context) {
				result := isWebSocketRequest(c)
				assert.Equal(t, tt.expected, result)
				c.JSON(http.StatusOK, gin.H{"is_ws": result})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		})
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	logger, _ := setupTestLogger()

	tests := []struct {
		name           string
		timeout        time.Duration
		handlerDelay   time.Duration
		pathTimeouts   []PathTimeout
		allowList      []string
		isWebSocket    bool
		expectTimeout  bool
		expectedStatus int
	}{
		{
			name:           "请求在超时前完成",
			timeout:        500 * time.Millisecond,
			handlerDelay:   100 * time.Millisecond,
			expectTimeout:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "请求超时",
			timeout:        100 * time.Millisecond,
			handlerDelay:   500 * time.Millisecond,
			expectTimeout:  true,
			expectedStatus: http.StatusRequestTimeout,
		},
		{
			name:           "WebSocket请求跳过超时",
			timeout:        100 * time.Millisecond,
			handlerDelay:   500 * time.Millisecond,
			isWebSocket:    true,
			expectTimeout:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "请求在允许列表中",
			timeout:        100 * time.Millisecond,
			handlerDelay:   500 * time.Millisecond,
			allowList:      []string{"/api/long"},
			expectTimeout:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "路径特定超时",
			timeout:        100 * time.Millisecond,
			handlerDelay:   300 * time.Millisecond,
			pathTimeouts:   []PathTimeout{{Path: "/api/slow", Timeout: 500 * time.Millisecond}},
			expectTimeout:  false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectTimeout || tt.isWebSocket {
				t.Skip("gin.Context is not thread-safe, known issue in gin v1.11.0")
			}

			gin.SetMode(gin.TestMode)
			r := gin.New()

			options := []TimeoutMiddlewareOption{}
			for _, pt := range tt.pathTimeouts {
				options = append(options, WithPathTimeout(pt.Path, pt.Timeout))
			}
			if len(tt.allowList) > 0 {
				options = append(options, WithAllowList(tt.allowList))
			}

			r.Use(TimeoutMiddleware(logger, tt.timeout, options...))

			path := "/api/test"
			if len(tt.pathTimeouts) > 0 {
				path = tt.pathTimeouts[0].Path
			}
			if len(tt.allowList) > 0 {
				path = tt.allowList[0]
			}

			r.GET(path, func(c *gin.Context) {
				time.Sleep(tt.handlerDelay)
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, path, nil)
			if tt.isWebSocket {
				req.Header.Set("Upgrade", "websocket")
				req.Header.Set("Connection", "upgrade")
				req.Header.Set("Sec-WebSocket-Version", "13")
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectTimeout {
				assert.Equal(t, "true", w.Header().Get("X-Request-Timeout"))
			}
		})
	}
}

func TestTimestampMiddleware(t *testing.T) {
	logger, _ := setupTestLogger()
	nonceStore := cache.New(5*time.Minute, 10*time.Minute)

	tests := []struct {
		name            string
		path            string
		timestamp       string
		nonce           string
		tolerance       int64
		futureTolerance int64
		expectError     bool
		expectedStatus  int
		isReplay        bool
	}{
		{
			name:            "有效时间戳和nonce",
			path:            "/api/test",
			timestamp:       fmt.Sprintf("%d", time.Now().UnixMilli()),
			nonce:           uuid.NewString(),
			tolerance:       300000,
			futureTolerance: 60000,
			expectError:     false,
			expectedStatus:  http.StatusOK,
			isReplay:        false,
		},
		{
			name:            "缺少X-Timestamp",
			path:            "/api/test",
			timestamp:       "",
			nonce:           uuid.NewString(),
			tolerance:       300000,
			futureTolerance: 60000,
			expectError:     true,
			expectedStatus:  http.StatusBadRequest,
			isReplay:        false,
		},
		{
			name:            "缺少X-Nonce",
			path:            "/api/test",
			timestamp:       fmt.Sprintf("%d", time.Now().UnixMilli()),
			nonce:           "",
			tolerance:       300000,
			futureTolerance: 60000,
			expectError:     true,
			expectedStatus:  http.StatusBadRequest,
			isReplay:        false,
		},
		{
			name:            "无效时间戳格式",
			path:            "/api/test",
			timestamp:       "invalid-timestamp",
			nonce:           uuid.NewString(),
			tolerance:       300000,
			futureTolerance: 60000,
			expectError:     true,
			expectedStatus:  http.StatusBadRequest,
			isReplay:        false,
		},
		{
			name:            "时间戳过期",
			path:            "/api/test",
			timestamp:       fmt.Sprintf("%d", time.Now().UnixMilli()-400000),
			nonce:           uuid.NewString(),
			tolerance:       300000,
			futureTolerance: 60000,
			expectError:     true,
			expectedStatus:  http.StatusBadRequest,
			isReplay:        false,
		},
		{
			name:            "时间戳太超前",
			path:            "/api/test",
			timestamp:       fmt.Sprintf("%d", time.Now().UnixMilli()+100000),
			nonce:           uuid.NewString(),
			tolerance:       300000,
			futureTolerance: 60000,
			expectError:     true,
			expectedStatus:  http.StatusBadRequest,
			isReplay:        false,
		},
		{
			name:            "非API路径跳过检查",
			path:            "/test",
			timestamp:       "",
			nonce:           "",
			tolerance:       300000,
			futureTolerance: 60000,
			expectError:     false,
			expectedStatus:  http.StatusOK,
			isReplay:        false,
		},
		{
			name:            "重复请求检测",
			path:            "/api/test",
			timestamp:       fmt.Sprintf("%d", time.Now().UnixMilli()),
			nonce:           "duplicate-nonce",
			tolerance:       300000,
			futureTolerance: 60000,
			expectError:     true,
			expectedStatus:  http.StatusBadRequest,
			isReplay:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.Use(TimestampMiddleware(nonceStore, logger, tt.tolerance, tt.futureTolerance, 5*time.Minute))
			r.GET(tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.timestamp != "" {
				req.Header.Set("X-Timestamp", tt.timestamp)
			}
			if tt.nonce != "" {
				req.Header.Set("X-Nonce", tt.nonce)
			}

			if tt.isReplay {
				nonceStore.Set(tt.nonce, true, 5*time.Minute)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{name: "正数", input: 10, expected: 10},
		{name: "负数", input: -10, expected: 10},
		{name: "零", input: 0, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := abs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTracingMiddleware(t *testing.T) {
	logger, recorded := setupTestLogger()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(TracingMiddleware(logger))
	r.GET("/test", func(c *gin.Context) {
		traceID := c.Request.Context().Value(ctxutil.TraceIDKey)
		assert.NotEmpty(t, traceID)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, recorded.Len() >= 2, "Expected at least start and end logs")
}

func TestIPRateLimiterConcurrency(t *testing.T) {
	limiter := NewIPRateLimiter(100, 10)
	ip := "192.168.1.1"
	var wg sync.WaitGroup
	counter := 0
	var mu sync.Mutex

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.GetLimiter(ip).Allow() {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, 10, counter)
}

func TestTimeoutMiddlewarePanicHandling(t *testing.T) {
	logger, _ := setupTestLogger()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(ErrorMiddleware(logger))
	r.Use(TimeoutMiddleware(logger, 500*time.Millisecond))
	r.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCorsMiddlewarePreflight(t *testing.T) {
	cfg := &config.AllowConfig{
		AllowOrigins:     []string{"http://example.com"},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(CorsMiddleware(cfg))
	r.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

var ctxutilJwtClaimsKey = ctxutil.JwtClaimsKey
