package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gin-artweb/internal/shared/config"
)

func TestCleanSQLComments_SingleLineComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove double slash comment",
			input:    "SELECT * FROM users // this is a comment",
			expected: "SELECT * FROM users",
		},
		{
			name:     "remove dash comment",
			input:    "SELECT * FROM users -- this is a comment",
			expected: "SELECT * FROM users",
		},
		{
			name:     "remove hash comment",
			input:    "SELECT * FROM users # this is a comment",
			expected: "SELECT * FROM users",
		},
		{
			name:     "comment at start",
			input:    "// comment\nSELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "mixed comments",
			input:    "SELECT * FROM users -- comment1\nSELECT * FROM posts // comment2",
			expected: "SELECT * FROM users \nSELECT * FROM posts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanSQLComments(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCleanSQLComments_MultiLineComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove multiline comment",
			input:    "SELECT * FROM /* comment */ users",
			expected: "SELECT * FROM  users",
		},
		{
			name:     "remove nested multiline comment",
			input:    "SELECT /* inner */ * FROM users",
			expected: "SELECT  * FROM users",
		},
		{
			name:     "multiline comment across lines",
			input:    "SELECT * /*\ncomment\n*/ FROM users",
			expected: "SELECT *  FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanSQLComments(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCleanSQLComments_StringProtection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "comment inside string should be preserved",
			input:    `SELECT '// not a comment' FROM users`,
			expected: `SELECT '// not a comment' FROM users`,
		},
		{
			name:     "multiline comment inside string",
			input:    `SELECT '/* not a comment */' FROM users`,
			expected: `SELECT '/* not a comment */' FROM users`,
		},
		{
			name:     "mixed string and comment",
			input:    `SELECT 'value --' FROM users -- real comment`,
			expected: `SELECT 'value --' FROM users`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanSQLComments(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSplitSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single statement",
			input:    "SELECT * FROM users",
			expected: []string{"SELECT * FROM users"},
		},
		{
			name:     "multiple statements",
			input:    "SELECT * FROM users; SELECT * FROM posts",
			expected: []string{"SELECT * FROM users", "SELECT * FROM posts"},
		},
		{
			name:     "statement with semicolon in string",
			input:    `SELECT ';not a separator' FROM users; SELECT * FROM posts`,
			expected: []string{`SELECT ';not a separator' FROM users`, "SELECT * FROM posts"},
		},
		{
			name:     "empty statements",
			input:    ";;SELECT * FROM users;;",
			expected: []string{"SELECT * FROM users"},
		},
		{
			name:     "whitespace handling",
			input:    "\nSELECT * FROM users ;\n\nSELECT * FROM posts ; ",
			expected: []string{"SELECT * FROM users", "SELECT * FROM posts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitSQL(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d statements, got %d", len(tt.expected), len(result))
			}
			for i, stmt := range result {
				if stmt != tt.expected[i] {
					t.Errorf("statement %d: expected %q, got %q", i, tt.expected[i], stmt)
				}
			}
		})
	}
}

func TestExecSQLFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sql_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sqlContent := `
CREATE TABLE test_table (
    id INTEGER PRIMARY KEY,
    name TEXT
);
INSERT INTO test_table (name) VALUES ('test');
`
	sqlPath := filepath.Join(tmpDir, "test.sql")
	err = os.WriteFile(sqlPath, []byte(sqlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write SQL file: %v", err)
	}

	c := &config.DBConf{
		Type:            "sqlite",
		Dsn:             ":memory:",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
	}

	gc := NewGormConfig(nil)
	db, err := NewGormDB(c, gc)
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	err = ExecSQLFile(ctx, db, sqlPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var count int64
	err = db.Model(&struct{}{}).Table("test_table").Count(&count).Error
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 record, got %d", count)
	}
}

func TestExecSQLFile_FileNotFound(t *testing.T) {
	c := &config.DBConf{
		Type:            "sqlite",
		Dsn:             ":memory:",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
	}

	gc := NewGormConfig(nil)
	db, err := NewGormDB(c, gc)
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	err = ExecSQLFile(ctx, db, "/nonexistent/path/test.sql")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestExecSQLFile_InvalidSQL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sql_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sqlContent := `
INVALID SQL SYNTAX;
`
	sqlPath := filepath.Join(tmpDir, "invalid.sql")
	err = os.WriteFile(sqlPath, []byte(sqlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write SQL file: %v", err)
	}

	c := &config.DBConf{
		Type:            "sqlite",
		Dsn:             ":memory:",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
	}

	gc := NewGormConfig(nil)
	db, err := NewGormDB(c, gc)
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	err = ExecSQLFile(ctx, db, sqlPath)
	if err == nil {
		t.Fatal("expected error for invalid SQL")
	}
}

func TestExecSQLFile_WithComments(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sql_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sqlContent := `
// This is a comment
CREATE TABLE test_comments (
    id INTEGER PRIMARY KEY -- inline comment
);
/* 
Multi-line comment
*/
INSERT INTO test_comments (id) VALUES (1); # another comment
`
	sqlPath := filepath.Join(tmpDir, "test_with_comments.sql")
	err = os.WriteFile(sqlPath, []byte(sqlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write SQL file: %v", err)
	}

	c := &config.DBConf{
		Type:            "sqlite",
		Dsn:             ":memory:",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
	}

	gc := NewGormConfig(nil)
	db, err := NewGormDB(c, gc)
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	err = ExecSQLFile(ctx, db, sqlPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var count int64
	err = db.Model(&struct{}{}).Table("test_comments").Count(&count).Error
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 record, got %d", count)
	}
}

func TestSplitSQL_EmptyInput(t *testing.T) {
	result := splitSQL("")
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input, got %v", result)
	}
}

func TestSplitSQL_OnlyWhitespace(t *testing.T) {
	result := splitSQL("   \n  ;  \n   ")
	if len(result) != 0 {
		t.Errorf("expected empty result for whitespace only input, got %v", result)
	}
}

func TestCleanSQLComments_EmptyInput(t *testing.T) {
	result := cleanSQLComments("")
	if result != "" {
		t.Errorf("expected empty result for empty input, got %q", result)
	}
}

func TestExecSQLFile_NilDB(t *testing.T) {
	c := &config.DBConf{
		Type:            "sqlite",
		Dsn:             ":memory:",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
	}

	gc := NewGormConfig(nil)
	db, err := NewGormDB(c, gc)
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	defer func() { _ = CloseGormDB(db) }()

	tmpDir, err := os.MkdirTemp("", "sql_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sqlContent := "SELECT 1;"
	sqlPath := filepath.Join(tmpDir, "test.sql")
	err = os.WriteFile(sqlPath, []byte(sqlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write SQL file: %v", err)
	}

	ctx := context.Background()
	err = ExecSQLFile(ctx, db, sqlPath)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
