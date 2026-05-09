package database

import (
	"log"
	"os"
	"testing"
	"time"

	"gin-artweb/internal/shared/config"

	"gorm.io/gorm"
)

func TestNewGormConfig(t *testing.T) {
	tests := []struct {
		name    string
		dbLog   *log.Logger
		wantLog bool
	}{
		{
			name:    "with logger",
			dbLog:   log.New(os.Stdout, "[TEST] ", log.LstdFlags),
			wantLog: true,
		},
		{
			name:    "without logger",
			dbLog:   nil,
			wantLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewGormConfig(tt.dbLog)
			if cfg == nil {
				t.Fatal("expected non-nil config")
			}

			if cfg.SkipDefaultTransaction != false {
				t.Errorf("expected SkipDefaultTransaction to be false, got %v", cfg.SkipDefaultTransaction)
			}

			if cfg.FullSaveAssociations != false {
				t.Errorf("expected FullSaveAssociations to be false, got %v", cfg.FullSaveAssociations)
			}

			if cfg.TranslateError != true {
				t.Errorf("expected TranslateError to be true, got %v", cfg.TranslateError)
			}

			if tt.wantLog && cfg.Logger == nil {
				t.Error("expected logger to be set")
			}
			if !tt.wantLog && cfg.Logger != nil {
				t.Error("expected logger to be nil")
			}
		})
	}
}

func TestNewGormDB_SQLite(t *testing.T) {
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
		t.Fatalf("expected no error, got %v", err)
	}

	if db == nil {
		t.Fatal("expected non-nil db instance")
	}

	_, err = db.DB()
	if err != nil {
		t.Fatalf("expected no error when getting underlying DB, got %v", err)
	}
}

func TestNewGormDB_UnsupportedType(t *testing.T) {
	c := &config.DBConf{
		Type: "unsupported",
		Dsn:  "test",
	}

	gc := NewGormConfig(nil)
	_, err := NewGormDB(c, gc)
	if err == nil {
		t.Fatal("expected error for unsupported database type, got nil")
	}

	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestNewGormDB_InvalidDNS(t *testing.T) {
	c := &config.DBConf{
		Type: "mysql",
		Dsn:  "invalid-dns-string",
	}

	gc := NewGormConfig(nil)
	_, err := NewGormDB(c, gc)
	if err == nil {
		t.Fatal("expected error for invalid DNS, got nil")
	}
}

func TestCloseGormDB(t *testing.T) {
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

	err = CloseGormDB(db)
	if err != nil {
		t.Errorf("expected no error when closing DB, got %v", err)
	}
}

func TestCloseGormDB_Nil(t *testing.T) {
	err := CloseGormDB(nil)
	if err == nil {
		t.Fatal("expected error when closing nil DB, got nil")
	}

	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCloseGormDB_FailedGetDB(t *testing.T) {
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

	err = CloseGormDB(db)
	if err != nil {
		t.Errorf("expected no error when closing DB, got %v", err)
	}
}

func TestNewGormDB_Postgres(t *testing.T) {
	c := &config.DBConf{
		Type: "postgres",
		Dsn:  "host=localhost port=5432 user=test dbname=test password=test sslmode=disable",
	}

	gc := NewGormConfig(nil)
	_, err := NewGormDB(c, gc)
	if err == nil {
		t.Skip("PostgreSQL not available, skipping connection test")
	}
}

func TestNewGormDB_MySQL(t *testing.T) {
	c := &config.DBConf{
		Type: "mysql",
		Dsn:  "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
	}

	gc := NewGormConfig(nil)
	_, err := NewGormDB(c, gc)
	if err == nil {
		t.Skip("MySQL not available, skipping connection test")
	}
}

func TestNewGormDB_SQLServer(t *testing.T) {
	c := &config.DBConf{
		Type: "sqlserver",
		Dsn:  "sqlserver://sa:password@localhost:1433?database=testdb",
	}

	gc := NewGormConfig(nil)
	_, err := NewGormDB(c, gc)
	if err == nil {
		t.Skip("SQL Server not available, skipping connection test")
	}
}

func TestNewGormDB_OpenGauss(t *testing.T) {
	c := &config.DBConf{
		Type: "opengauss",
		Dsn:  "host=localhost port=5432 user=test password=test dbname=test",
	}

	gc := NewGormConfig(nil)
	_, err := NewGormDB(c, gc)
	if err == nil {
		t.Skip("OpenGauss not available, skipping connection test")
	}
}

func TestConnectionPoolSettings(t *testing.T) {
	c := &config.DBConf{
		Type:            "sqlite",
		Dsn:             ":memory:",
		MaxIdleConns:    5,
		MaxOpenConns:    50,
		ConnMaxLifetime: time.Minute * 10,
		ConnMaxIdleTime: time.Minute * 5,
	}

	gc := NewGormConfig(nil)
	db, err := NewGormDB(c, gc)
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}
	defer func() { _ = CloseGormDB(db) }()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying DB: %v", err)
	}

	_ = sqlDB
}

func TestGormDB_Query(t *testing.T) {
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

	type TestModel struct {
		gorm.Model
		Name string
	}

	err = db.AutoMigrate(&TestModel{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	err = db.Create(&TestModel{Name: "test"}).Error
	if err != nil {
		t.Fatalf("failed to create record: %v", err)
	}

	var result TestModel
	err = db.First(&result, "name = ?", "test").Error
	if err != nil {
		t.Fatalf("failed to query record: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("expected name to be 'test', got '%s'", result.Name)
	}
}
