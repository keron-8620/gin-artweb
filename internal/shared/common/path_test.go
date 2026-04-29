package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadUint32FromFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		filePath    string
		fileContent string
		setupFile   bool
		want        uint32
		wantErr     bool
	}{
		{"file not exist", filepath.Join(tmpDir, "notexist.txt"), "", false, 0, false},
		{"valid number", filepath.Join(tmpDir, "valid.txt"), "12345\n", true, 12345, false},
		{"number with whitespace", filepath.Join(tmpDir, "whitespace.txt"), "  999  \n", true, 999, false},
		{"max uint32", filepath.Join(tmpDir, "max.txt"), "4294967295\n", true, 4294967295, false},
		{"zero", filepath.Join(tmpDir, "zero.txt"), "0\n", true, 0, false},
		{"invalid content", filepath.Join(tmpDir, "invalid.txt"), "abc123\n", true, 0, true},
		{"empty file", filepath.Join(tmpDir, "empty.txt"), "", true, 0, true},
		{"overflow", filepath.Join(tmpDir, "overflow.txt"), "4294967296\n", true, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFile {
				if err := os.WriteFile(tt.filePath, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			got, err := ReadUint32FromFile(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadUint32FromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadUint32FromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
