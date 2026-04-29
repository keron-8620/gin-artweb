package common

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestUploadFile(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		fileSize int64
		maxSize  int64
		savePath string
		filePerm os.FileMode
		wantErr  bool
	}{
		{"normal upload", 100, 1024, filepath.Join(tmpDir, "upload", "test.txt"), 0644, false},
		{"file too large", 2000, 1024, filepath.Join(tmpDir, "upload", "large.txt"), 0644, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			part, err := writer.CreateFormFile("file", "test.txt")
			if err != nil {
				t.Fatalf("failed to create form file: %v", err)
			}
			_, err = part.Write(make([]byte, tt.fileSize))
			if err != nil {
				t.Fatalf("failed to write to form file: %v", err)
			}
			writer.Close()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/upload", &buf)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())

			err = c.Request.ParseMultipartForm(32 << 20)
			if err != nil {
				t.Fatalf("failed to parse multipart form: %v", err)
			}

			files := c.Request.MultipartForm.File["file"]
			if len(files) == 0 {
				t.Fatal("no file found in form")
			}

			result := UploadFile(c, logger, tt.maxSize, tt.savePath, files[0], tt.filePerm)

			if (result != nil) != tt.wantErr {
				t.Errorf("UploadFile() error = %v, wantErr %v", result, tt.wantErr)
			}

			if !tt.wantErr {
				if _, err := os.Stat(tt.savePath); os.IsNotExist(err) {
					t.Errorf("UploadFile() file not saved at %s", tt.savePath)
				}
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "download_test.txt")
	err := os.WriteFile(testFile, []byte("hello world"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		rename   string
		wantErr  bool
	}{
		{"file exists", testFile, "", false},
		{"file exists with rename", testFile, "new_name.txt", false},
		{"file not exist", filepath.Join(tmpDir, "notexist.txt"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/download", nil)

			result := DownloadFile(c, logger, tt.filePath, tt.rename)

			if (result != nil) != tt.wantErr {
				t.Errorf("DownloadFile() error = %v, wantErr %v", result, tt.wantErr)
			}

			if !tt.wantErr {
				if w.Code != http.StatusOK {
					t.Errorf("DownloadFile() status code = %v, want %v", w.Code, http.StatusOK)
				}
				contentDisposition := w.Header().Get("Content-Disposition")
				if contentDisposition == "" {
					t.Error("DownloadFile() Content-Disposition header not set")
				}
			}
		})
	}
}
