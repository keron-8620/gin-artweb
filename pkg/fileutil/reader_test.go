package fileutil

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriteReaderToFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "write-reader-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		name        string
		content     string
		filePath    string
		fileMode    os.FileMode
		overwrite   bool
		expectError bool
		errorMsg    string
		cancelCtx   bool
	}{
		{
			name:        "正常保存文件",
			content:     "test content",
			filePath:    filepath.Join(tempDir, "test1.txt"),
			fileMode:    0644,
			overwrite:   false,
			expectError: false,
			cancelCtx:   false,
		},
		{
			name:        "覆盖已存在文件",
			content:     "updated content",
			filePath:    filepath.Join(tempDir, "test2.txt"),
			fileMode:    0644,
			overwrite:   true,
			expectError: false,
			cancelCtx:   false,
		},
		{
			name:        "不允许覆盖已存在文件",
			content:     "new content",
			filePath:    filepath.Join(tempDir, "test3.txt"),
			fileMode:    0644,
			overwrite:   false,
			expectError: true,
			errorMsg:    "文件已存在",
			cancelCtx:   false,
		},
		{
			name:        "创建嵌套目录",
			content:     "nested content",
			filePath:    filepath.Join(tempDir, "subdir1", "subdir2", "test4.txt"),
			fileMode:    0644,
			overwrite:   false,
			expectError: false,
			cancelCtx:   false,
		},
		{
			name:        "空内容",
			content:     "",
			filePath:    filepath.Join(tempDir, "test5.txt"),
			fileMode:    0644,
			overwrite:   false,
			expectError: false,
			cancelCtx:   false,
		},
		{
			name:        "nil reader",
			content:     "",
			filePath:    filepath.Join(tempDir, "test6.txt"),
			fileMode:    0644,
			overwrite:   false,
			expectError: true,
			errorMsg:    "fileReader 不能为空",
			cancelCtx:   false,
		},
		{
			name:        "空文件路径",
			content:     "test content",
			filePath:    "",
			fileMode:    0644,
			overwrite:   false,
			expectError: true,
			errorMsg:    "filePath 不能为空",
			cancelCtx:   false,
		},
		{
			name:        "仅空白字符的文件路径",
			content:     "test content",
			filePath:    "   ",
			fileMode:    0644,
			overwrite:   false,
			expectError: true,
			errorMsg:    "filePath 不能为空",
			cancelCtx:   false,
		},
		{
			name:        "上下文取消",
			content:     "test content",
			filePath:    filepath.Join(tempDir, "test7.txt"),
			fileMode:    0644,
			overwrite:   false,
			expectError: true,
			cancelCtx:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ctx context.Context
			var cancel context.CancelFunc
			if tc.cancelCtx {
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx = context.Background()
			}

			if tc.overwrite {
				initialContent := "initial content"
				if err := os.WriteFile(tc.filePath, []byte(initialContent), tc.fileMode); err != nil {
					t.Fatalf("Failed to create initial file: %v", err)
				}
			}

			if !tc.overwrite && tc.expectError && tc.errorMsg == "文件已存在" {
				initialContent := "initial content"
				if err := os.WriteFile(tc.filePath, []byte(initialContent), tc.fileMode); err != nil {
					t.Fatalf("Failed to create initial file: %v", err)
				}
			}

			var reader io.Reader
			if tc.name != "nil reader" {
				reader = bytes.NewBufferString(tc.content)
			}

			err := WriteReaderToFile(ctx, reader, tc.filePath, tc.fileMode, tc.overwrite)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				actualContent, err := os.ReadFile(tc.filePath)
				assert.NoError(t, err)
				assert.Equal(t, tc.content, string(actualContent))

				fileInfo, err := os.Stat(tc.filePath)
				assert.NoError(t, err)
				assert.Equal(t, tc.fileMode, fileInfo.Mode().Perm())
			}
		})
	}
}

func TestWriteReaderToFileWithTimeout(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "write-reader-timeout-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	filePath := filepath.Join(tempDir, "test_timeout.txt")
	slowReader := &slowReader{delay: 100 * time.Millisecond, ctx: ctx}

	err = WriteReaderToFile(ctx, slowReader, filePath, 0644, false)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestWriteReaderToFileLargeContent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "write-large-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	contentSize := 1 * 1024 * 1024
	content := strings.Repeat("a", contentSize)
	filePath := filepath.Join(tempDir, "large_file.txt")

	reader := bytes.NewBufferString(content)
	err = WriteReaderToFile(context.Background(), reader, filePath, 0644, false)

	assert.NoError(t, err)

	actualContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, contentSize, len(actualContent))
	assert.Equal(t, content, string(actualContent))
}

func TestWriteReaderToFileReaderError(t *testing.T) {
	tempDir, err := os.MkdirTemp("/tmp", "write-reader-error-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "error_test.txt")
	reader := &errorReader{}

	err = WriteReaderToFile(context.Background(), reader, filePath, 0644, false)

	assert.Error(t, err)
}

func TestWriteReaderToFileBinaryContent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "write-binary-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	binaryData := make([]byte, 256)
	for i := range binaryData {
		binaryData[i] = byte(i)
	}

	filePath := filepath.Join(tempDir, "binary_file.bin")
	reader := bytes.NewBuffer(binaryData)

	err = WriteReaderToFile(context.Background(), reader, filePath, 0644, false)
	assert.NoError(t, err)

	actualContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, binaryData, actualContent)
}

func TestWriteReaderToFilePermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "write-perm-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testModes := []os.FileMode{0600, 0644, 0755, 0777}

	for _, mode := range testModes {
		t.Run(mode.String(), func(t *testing.T) {
			filePath := filepath.Join(tempDir, "perm_test_"+mode.String()+".txt")
			reader := bytes.NewBufferString("permission test")

			err := WriteReaderToFile(context.Background(), reader, filePath, mode, false)
			assert.NoError(t, err)

			info, err := os.Stat(filePath)
			assert.NoError(t, err)
			assert.Equal(t, mode, info.Mode().Perm())
		})
	}
}

func TestWriteReaderToFileContextCancellationMidway(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "write-cancel-midway-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithCancel(context.Background())
	filePath := filepath.Join(tempDir, "cancel_midway.txt")

	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	reader := &slowChunkReader{chunkSize: 1024, delay: 10 * time.Millisecond, totalSize: 50 * 1024, ctx: ctx}
	err = WriteReaderToFile(ctx, reader, filePath, 0644, false)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func BenchmarkWriteReaderToFileSmall(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "bench-small")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	content := "small test content"
	filePath := filepath.Join(tempDir, "small.txt")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		os.Remove(filePath)
		reader := bytes.NewBufferString(content)
		if err := WriteReaderToFile(context.Background(), reader, filePath, 0644, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteReaderToFileMedium(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "bench-medium")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	content := strings.Repeat("x", 64*1024)
	filePath := filepath.Join(tempDir, "medium.txt")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		os.Remove(filePath)
		reader := bytes.NewBufferString(content)
		if err := WriteReaderToFile(context.Background(), reader, filePath, 0644, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteReaderToFileLarge(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "bench-large")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	content := strings.Repeat("y", 10*1024*1024)
	filePath := filepath.Join(tempDir, "large.txt")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		os.Remove(filePath)
		reader := bytes.NewBufferString(content)
		if err := WriteReaderToFile(context.Background(), reader, filePath, 0644, false); err != nil {
			b.Fatal(err)
		}
	}
}

type slowReader struct {
	delay time.Duration
	ctx   context.Context
}

func (s *slowReader) Read(p []byte) (n int, err error) {
	select {
	case <-time.After(s.delay):
		return 0, io.EOF
	case <-s.ctx.Done():
		return 0, context.DeadlineExceeded
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

type slowChunkReader struct {
	chunkSize int
	delay     time.Duration
	totalSize int
	offset    int
	ctx       context.Context
}

func (s *slowChunkReader) Read(p []byte) (n int, err error) {
	if s.offset >= s.totalSize {
		return 0, io.EOF
	}

	select {
	case <-s.ctx.Done():
		return 0, s.ctx.Err()
	case <-time.After(s.delay):
	}

	remaining := s.totalSize - s.offset
	if len(p) > remaining {
		p = p[:remaining]
	}

	for i := range p {
		p[i] = 'a'
	}

	n = len(p)
	s.offset += n

	if s.offset >= s.totalSize {
		return n, io.EOF
	}

	return n, nil
}
