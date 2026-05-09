package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGetBaseDir(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		cleanup func()
		wantErr bool
	}{
		{
			name:    "normal execution path",
			setup:   func() {},
			cleanup: func() {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			result := getBaseDir()
			if result == "" {
				t.Error("getBaseDir() returned empty string")
			}
		})
	}
}

func TestPathVariables(t *testing.T) {
	if BaseDir == "" {
		t.Error("BaseDir should not be empty")
	}
	if ConfigDir == "" {
		t.Error("ConfigDir should not be empty")
	}
	if LogDir == "" {
		t.Error("LogDir should not be empty")
	}
	if ResourceDir == "" {
		t.Error("ResourceDir should not be empty")
	}
	if StorageDir == "" {
		t.Error("StorageDir should not be empty")
	}
	if TmpDir == "" {
		t.Error("TmpDir should not be empty")
	}

	expectedConfigDir := filepath.Join(BaseDir, "config")
	if ConfigDir != expectedConfigDir {
		t.Errorf("ConfigDir = %v, want %v", ConfigDir, expectedConfigDir)
	}

	expectedLogDir := filepath.Join(BaseDir, "logs")
	if LogDir != expectedLogDir {
		t.Errorf("LogDir = %v, want %v", LogDir, expectedLogDir)
	}
}

func TestNewJWTConfig(t *testing.T) {
	tests := []struct {
		name             string
		accessExp        time.Duration
		refreshExp       time.Duration
		accessMethod     string
		refreshMethod    string
		accessSecret     []byte
		refreshSecret    []byte
		expectPanic      bool
		wantAccessMethod jwt.SigningMethod
	}{
		{
			name:             "valid HS256 configuration",
			accessExp:        time.Hour,
			refreshExp:       24 * time.Hour,
			accessMethod:     "HS256",
			refreshMethod:    "HS256",
			accessSecret:     []byte("test-access-secret"),
			refreshSecret:    []byte("test-refresh-secret"),
			expectPanic:      false,
			wantAccessMethod: jwt.SigningMethodHS256,
		},
		{
			name:             "valid HS512 configuration",
			accessExp:        30 * time.Minute,
			refreshExp:       7 * 24 * time.Hour,
			accessMethod:     "HS512",
			refreshMethod:    "HS512",
			accessSecret:     []byte("long-test-access-secret-key"),
			refreshSecret:    []byte("long-test-refresh-secret-key"),
			expectPanic:      false,
			wantAccessMethod: jwt.SigningMethodHS512,
		},
		{
			name:          "invalid access method",
			accessExp:     time.Hour,
			refreshExp:    24 * time.Hour,
			accessMethod:  "INVALID",
			refreshMethod: "HS256",
			accessSecret:  []byte("test-secret"),
			refreshSecret: []byte("test-secret"),
			expectPanic:   true,
		},
		{
			name:          "invalid refresh method",
			accessExp:     time.Hour,
			refreshExp:    24 * time.Hour,
			accessMethod:  "HS256",
			refreshMethod: "INVALID",
			accessSecret:  []byte("test-secret"),
			refreshSecret: []byte("test-secret"),
			expectPanic:   true,
		},
		{
			name:          "empty access secret",
			accessExp:     time.Hour,
			refreshExp:    24 * time.Hour,
			accessMethod:  "HS256",
			refreshMethod: "HS256",
			accessSecret:  []byte{},
			refreshSecret: []byte("test-secret"),
			expectPanic:   true,
		},
		{
			name:          "empty refresh secret",
			accessExp:     time.Hour,
			refreshExp:    24 * time.Hour,
			accessMethod:  "HS256",
			refreshMethod: "HS256",
			accessSecret:  []byte("test-secret"),
			refreshSecret: []byte{},
			expectPanic:   true,
		},
		{
			name:          "both secrets empty",
			accessExp:     time.Hour,
			refreshExp:    24 * time.Hour,
			accessMethod:  "HS256",
			refreshMethod: "HS256",
			accessSecret:  []byte{},
			refreshSecret: []byte{},
			expectPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); (r != nil) != tt.expectPanic {
					if tt.expectPanic {
						t.Errorf("NewJWTConfig() did not panic when expected")
					} else {
						t.Errorf("NewJWTConfig() panicked unexpectedly: %v", r)
					}
				}
			}()

			result := NewJWTConfig(
				tt.accessExp,
				tt.refreshExp,
				tt.accessMethod,
				tt.refreshMethod,
				tt.accessSecret,
				tt.refreshSecret,
			)

			if !tt.expectPanic {
				if result == nil {
					t.Error("NewJWTConfig() returned nil")
					return
				}

				if result.AccessTokenExpiration != tt.accessExp {
					t.Errorf("AccessTokenExpiration = %v, want %v", result.AccessTokenExpiration, tt.accessExp)
				}

				if result.RefreshTokenExpiration != tt.refreshExp {
					t.Errorf("RefreshTokenExpiration = %v, want %v", result.RefreshTokenExpiration, tt.refreshExp)
				}

				if string(result.AccessSecret) != string(tt.accessSecret) {
					t.Errorf("AccessSecret = %v, want %v", string(result.AccessSecret), string(tt.accessSecret))
				}

				if string(result.RefreshSecret) != string(tt.refreshSecret) {
					t.Errorf("RefreshSecret = %v, want %v", string(result.RefreshSecret), string(tt.refreshSecret))
				}

				if result.AccessMethod.Alg() != tt.wantAccessMethod.Alg() {
					t.Errorf("AccessMethod.Alg() = %v, want %v", result.AccessMethod.Alg(), tt.wantAccessMethod.Alg())
				}
			}
		})
	}
}

func TestSystemConfStruct(t *testing.T) {
	conf := &SystemConf{
		Server: &ServerConfig{
			Host: "localhost",
			Port: 8080,
			SSL: SSLConfig{
				Enable: false,
			},
			Rate: RateLimitConfig{
				RPS:   100,
				Burst: 10,
			},
			Timeout: TimeoutConfig{
				Request:  30,
				Shutdown: 10,
			},
			Swagger: true,
		},
		Database: &DBConf{
			Type:            "mysql",
			Dsn:             "root:password@tcp(localhost:3306)/test",
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			LogSQL:          true,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 600,
			ReadTimeout:     10,
			WriteTimeout:    15,
			ListTimeout:     30,
			ReadSlow:        2,
			WriteSlow:       3,
			ListSlow:        5,
		},
		Log: &LogConfig{
			Level:      "info",
			MaxSize:    100,
			MaxAge:     30,
			MaxBackups: 10,
			LocalTime:  true,
			Compress:   true,
		},
		CORS: &AllowConfig{
			AllowOrigins:     []string{"*"},
			AllowCredentials: true,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			ExposeHeaders:    []string{"X-Custom-Header"},
		},
		Security: &SecurityConfig{
			HostGuard: HostGuardConfig{
				Enable:       true,
				TrustedHosts: []string{"localhost", "example.com"},
			},
			Timestamp: TimestampConfig{
				CheckTimestamp:  true,
				Tolerance:       30000,
				FutureTolerance: 5000,
			},
			Token: TokenConfig{
				AccessDuration:  time.Hour,
				RefreshDuration: 24 * time.Hour,
				AccessMethod:    "HS256",
				RefreshMethod:   "HS256",
			},
			Login: LoginSecurityConfig{
				MaxFailedAttempts: 5,
				LockDuration:      15 * time.Minute,
			},
			Password: PasswordConfig{
				StrengthLevel: 3,
			},
		},
		SSH: &SSHConfig{
			Private: "/home/user/.ssh/id_rsa",
			Timeout: 30,
		},
		Upload: &UploadConfig{
			MaxPkgSize:    50,
			MaxScriptSize: 10,
			MaxConfSize:   5,
		},
		API: &APIConfig{
			TotalTimeout:     60,
			BizCreateTimeout: 30,
			BizUpdateTimeout: 30,
			BizQueryTimeout:  10,
			BizDeleteTimeout: 10,
		},
	}

	if conf.Server.Host != "localhost" {
		t.Errorf("Server.Host = %v, want localhost", conf.Server.Host)
	}
	if conf.Database.Type != "mysql" {
		t.Errorf("Database.Type = %v, want mysql", conf.Database.Type)
	}
	if conf.Log.Level != "info" {
		t.Errorf("Log.Level = %v, want info", conf.Log.Level)
	}
	if !conf.CORS.AllowCredentials {
		t.Error("CORS.AllowCredentials should be true")
	}
	if !conf.Security.HostGuard.Enable {
		t.Error("Security.HostGuard.Enable should be true")
	}
	if conf.SSH.Timeout != 30 {
		t.Errorf("SSH.Timeout = %v, want 30", conf.SSH.Timeout)
	}
	if conf.Upload.MaxPkgSize != 50 {
		t.Errorf("Upload.MaxPkgSize = %v, want 50", conf.Upload.MaxPkgSize)
	}
}

func TestDBTimeoutStruct(t *testing.T) {
	timeout := &DBTimeout{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		ListTimeout:  30 * time.Second,
	}

	if timeout.ReadTimeout != 10*time.Second {
		t.Errorf("ReadTimeout = %v, want 10s", timeout.ReadTimeout)
	}
	if timeout.WriteTimeout != 15*time.Second {
		t.Errorf("WriteTimeout = %v, want 15s", timeout.WriteTimeout)
	}
	if timeout.ListTimeout != 30*time.Second {
		t.Errorf("ListTimeout = %v, want 30s", timeout.ListTimeout)
	}
}

func TestDBSlowThresholdStruct(t *testing.T) {
	threshold := &DBSlowThreshold{
		ReadSlow:  2 * time.Second,
		WriteSlow: 3 * time.Second,
		ListSlow:  5 * time.Second,
	}

	if threshold.ReadSlow != 2*time.Second {
		t.Errorf("ReadSlow = %v, want 2s", threshold.ReadSlow)
	}
	if threshold.WriteSlow != 3*time.Second {
		t.Errorf("WriteSlow = %v, want 3s", threshold.WriteSlow)
	}
	if threshold.ListSlow != 5*time.Second {
		t.Errorf("ListSlow = %v, want 5s", threshold.ListSlow)
	}
}

func TestJWTConfigStruct(t *testing.T) {
	accessSecret := []byte("access-secret")
	refreshSecret := []byte("refresh-secret")

	jwtConf := &JWTConfig{
		Issuer:                 "test-issuer",
		AccessTokenExpiration:  time.Hour,
		RefreshTokenExpiration: 24 * time.Hour,
		AccessSecret:           accessSecret,
		RefreshSecret:          refreshSecret,
		AccessMethod:           jwt.SigningMethodHS256,
		RefreshMethod:          jwt.SigningMethodHS256,
	}

	if jwtConf.Issuer != "test-issuer" {
		t.Errorf("Issuer = %v, want test-issuer", jwtConf.Issuer)
	}
	if jwtConf.AccessTokenExpiration != time.Hour {
		t.Errorf("AccessTokenExpiration = %v, want 1h", jwtConf.AccessTokenExpiration)
	}
	if jwtConf.RefreshTokenExpiration != 24*time.Hour {
		t.Errorf("RefreshTokenExpiration = %v, want 24h", jwtConf.RefreshTokenExpiration)
	}
	if string(jwtConf.AccessSecret) != string(accessSecret) {
		t.Errorf("AccessSecret mismatch")
	}
	if string(jwtConf.RefreshSecret) != string(refreshSecret) {
		t.Errorf("RefreshSecret mismatch")
	}
	if jwtConf.AccessMethod.Alg() != "HS256" {
		t.Errorf("AccessMethod.Alg() = %v, want HS256", jwtConf.AccessMethod.Alg())
	}
}

func TestConfigDirExists(t *testing.T) {
	_, err := os.Stat(ConfigDir)
	if os.IsNotExist(err) {
		t.Logf("Config directory %s does not exist, this may be expected in test environment", ConfigDir)
	}
}

func TestEnvVariables(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		setup    func()
		cleanup  func()
	}{
		{
			name:     "test env var presence",
			envKey:   "TEST_CONFIG_ENV",
			envValue: "test_value",
			setup: func() {
				os.Setenv("TEST_CONFIG_ENV", "test_value")
			},
			cleanup: func() {
				os.Unsetenv("TEST_CONFIG_ENV")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			val := os.Getenv(tt.envKey)
			if val != tt.envValue {
				t.Errorf("os.Getenv(%q) = %q, want %q", tt.envKey, val, tt.envValue)
			}
		})
	}
}
