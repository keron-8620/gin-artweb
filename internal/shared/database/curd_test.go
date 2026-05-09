package database

import (
	"context"
	"testing"
	"time"

	"gin-artweb/internal/shared/config"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name  string `gorm:"column:name"`
	Email string `gorm:"column:email"`
	Age   int    `gorm:"column:age"`
}

type Role struct {
	gorm.Model
	Name string `gorm:"column:name"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

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

	err = db.AutoMigrate(&User{}, &Role{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestDBCreate(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	user := &User{Name: "test", Email: "test@example.com", Age: 25}

	err := DBCreate(ctx, db, &User{}, user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user ID to be set after creation")
	}
}

func TestDBCreateTX(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	user := &User{Name: "transaction_test", Email: "tx@example.com", Age: 30}

	err := DBCreateTX(ctx, db, &User{}, user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user ID to be set after transactional creation")
	}
}

func TestDBUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	user := &User{Name: "update_test", Email: "update@example.com", Age: 20}
	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	updateData := map[string]any{"Age": 25}
	err = DBUpdate(ctx, db, &User{}, updateData, "id = ?", user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var updatedUser User
	err = db.First(&updatedUser, user.ID).Error
	if err != nil {
		t.Fatalf("failed to fetch updated user: %v", err)
	}

	if updatedUser.Age != 25 {
		t.Errorf("expected age to be 25, got %d", updatedUser.Age)
	}
}

func TestDBUpdate_NoData(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := DBUpdate(ctx, db, &User{}, map[string]any{}, "id = ?", 1)
	if err != nil {
		t.Errorf("expected no error when updating with empty data, got %v", err)
	}
}

func TestDBUpdate_NoConditions(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := DBUpdate(ctx, db, &User{}, map[string]any{"Age": 25})
	if err == nil {
		t.Fatal("expected error when updating without conditions")
	}
}

func TestDBUpdateTx(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	user := &User{Name: "update_tx_test", Email: "update_tx@example.com", Age: 20}
	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	updateData := map[string]any{"Age": 30}
	err = DBUpdateTx(ctx, db, &User{}, updateData, "id = ?", user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var updatedUser User
	err = db.First(&updatedUser, user.ID).Error
	if err != nil {
		t.Fatalf("failed to fetch updated user: %v", err)
	}

	if updatedUser.Age != 30 {
		t.Errorf("expected age to be 30, got %d", updatedUser.Age)
	}
}

func TestDBDeleteTx(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	user := &User{Name: "delete_test", Email: "delete@example.com", Age: 25}
	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	err = DBDeleteTx(ctx, db, &User{}, "id = ?", user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var count int64
	err = db.Model(&User{}).Where("id = ?", user.ID).Count(&count).Error
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}

	if count != 0 {
		t.Errorf("expected user to be deleted, got %d records", count)
	}
}

func TestDBDeleteTx_NoConditions(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := DBDeleteTx(ctx, db, &User{})
	if err == nil {
		t.Fatal("expected error when deleting without conditions")
	}
}

func TestDBGet(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	user := &User{Name: "get_test", Email: "get@example.com", Age: 25}
	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	var result User
	err = DBGet(ctx, db, nil, &result, "id = ?", user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Name != "get_test" {
		t.Errorf("expected name to be 'get_test', got '%s'", result.Name)
	}
}

func TestDBGet_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	var result User
	err := DBGet(ctx, db, nil, &result, "id = ?", 999)
	if err == nil {
		t.Fatal("expected error when record not found")
	}
}

func TestDBList(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		err := db.Create(&User{Name: "list_test", Email: "list_test@example.com", Age: 20 + i}).Error
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}
	}

	var users []User
	query := QueryParams{
		Query:   map[string]any{"name": "list_test"},
		OrderBy: []string{"age ASC"},
		Limit:   3,
		Offset:  0,
	}

	err := DBList(ctx, db, &User{}, &users, query)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}

func TestDBList_WithColumns(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	err := db.Create(&User{Name: "column_test", Email: "column@example.com", Age: 30}).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	var users []User
	query := QueryParams{
		Columns: []string{"name"},
	}

	err = DBList(ctx, db, &User{}, &users, query)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) == 0 {
		t.Error("expected at least one user")
	}
}

func TestDBList_WithOmit(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	err := db.Create(&User{Name: "omit_test", Email: "omit@example.com", Age: 30}).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	var users []User
	query := QueryParams{
		Omit: []string{"email"},
	}

	err = DBList(ctx, db, &User{}, &users, query)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) == 0 {
		t.Error("expected at least one user")
	}
}

func TestDBCount(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		err := db.Create(&User{Name: "count_test", Email: "count@example.com", Age: 20 + i}).Error
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}
	}

	count, err := DBCount(ctx, db, &User{}, map[string]any{"name": "count_test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 5 {
		t.Errorf("expected count to be 5, got %d", count)
	}
}

func TestDBCount_EmptyQuery(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	count, err := DBCount(ctx, db, &User{}, map[string]any{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 0 {
		t.Errorf("expected count to be 0, got %d", count)
	}
}

func TestDBCreateRelationTx(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	user := &User{Name: "relation_test", Email: "relation@example.com", Age: 25}
	upmap := map[string]any{}

	err := DBCreateRelationTx(ctx, db, &User{}, user, upmap)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDBUpdateRelationTx(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	user := &User{Name: "relation_update_test", Email: "relation_update@example.com", Age: 25}
	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	data := map[string]any{"Age": 30}
	upmap := map[string]any{}

	err = DBUpdateRelationTx(ctx, db, user, data, upmap, "id = ?", user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDBPanic_NoPanic(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := DBPanic(ctx, db, nil)
	if err != nil {
		t.Errorf("expected no error when no panic occurred, got %v", err)
	}
}

func TestDBPanic_WithError(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	testErr := gorm.ErrRecordNotFound

	err := DBPanic(ctx, db, testErr)
	if err == nil {
		t.Error("expected error to be returned")
	}
}

func TestDBPanic_WithNilDB(t *testing.T) {
	ctx := context.Background()

	err := DBPanic(ctx, nil, nil)
	if err != nil {
		t.Errorf("expected no error when no panic occurred with nil db, got %v", err)
	}
}

func TestDBPanic_WithNilCtx(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	//nolint:staticcheck // intentionally testing nil context behavior
	err := DBPanic(nil, db, nil)
	if err != nil {
		t.Errorf("expected no error when no panic occurred with nil ctx, got %v", err)
	}
}

func TestQueryParams_MarshalLogObject(t *testing.T) {
	qp := &QueryParams{
		Preloads: []string{"Roles", "Profile"},
		Query:    map[string]any{"name": "test", "age": 25},
		OrderBy:  []string{"created_at DESC"},
		Limit:    10,
		Offset:   0,
		Omit:     []string{"password"},
		Columns:  nil,
	}

	assert.NotNil(t, qp)

	qp2 := &QueryParams{}
	assert.NotNil(t, qp2)

	encoder := zapcore.NewMapObjectEncoder()
	err := qp.MarshalLogObject(encoder)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	qp3 := &QueryParams{}
	err = qp3.MarshalLogObject(encoder)
	if err != nil {
		t.Errorf("expected no error for empty QueryParams, got %v", err)
	}
}

func TestDBGet_WithPreloads(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	user := &User{Name: "preload_test", Email: "preload@example.com", Age: 25}
	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	var result User
	err = DBGet(ctx, db, []string{}, &result, "id = ?", user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Name != "preload_test" {
		t.Errorf("expected name to be 'preload_test', got '%s'", result.Name)
	}
}

func TestDBList_EmptyOrderBy(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	err := db.Create(&User{Name: "empty_order", Email: "empty@example.com", Age: 30}).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	var users []User
	query := QueryParams{
		Query:   map[string]any{"name": "empty_order"},
		OrderBy: []string{},
	}

	err = DBList(ctx, db, &User{}, &users, query)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

func TestDBList_NoLimit(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		err := db.Create(&User{Name: "no_limit", Email: "no_limit@example.com", Age: 20 + i}).Error
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}
	}

	var users []User
	query := QueryParams{
		Query: map[string]any{"name": "no_limit"},
		Limit: 0,
	}

	err := DBList(ctx, db, &User{}, &users, query)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}

func TestDBUpdateTx_NoConditions(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := DBUpdateTx(ctx, db, &User{}, map[string]any{"Age": 25})
	if err == nil {
		t.Fatal("expected error when updating without conditions")
	}
}

func TestDBUpdateTx_NoData(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := DBUpdateTx(ctx, db, &User{}, map[string]any{}, "id = ?", 1)
	if err != nil {
		t.Errorf("expected no error when updating with empty data, got %v", err)
	}
}

func TestDBCount_NilQuery(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	count, err := DBCount(ctx, db, &User{}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 0 {
		t.Errorf("expected count to be 0, got %d", count)
	}
}

func TestDBCount_EmptyKey(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()
	err := db.Create(&User{Name: "empty_key", Email: "empty@example.com", Age: 25}).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	count, err := DBCount(ctx, db, &User{}, map[string]any{"": "value"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 1 {
		t.Errorf("expected count to be 1, got %d", count)
	}
}

func TestDBUpdateRelationTx_NoConditions(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := DBUpdateRelationTx(ctx, db, &User{}, map[string]any{"Age": 30}, map[string]any{})
	if err == nil {
		t.Fatal("expected error when updating without conditions")
	}
}

func TestDBUpdateRelationTx_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := DBUpdateRelationTx(ctx, db, &User{}, map[string]any{"Age": 30}, map[string]any{}, "id = ?", 999)
	if err == nil {
		t.Fatal("expected error when record not found")
	}
}

func TestDBCreateTX_FailedCommit(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := db.Exec("DROP TABLE IF EXISTS users").Error
	if err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}

	user := &User{Name: "failed_commit", Email: "failed@example.com", Age: 25}
	err = DBCreateTX(ctx, db, &User{}, user)
	if err == nil {
		t.Error("expected error when creating record in non-existent table")
	}
}

func TestDBCreateRelationTx_FailedCommit(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := db.Exec("DROP TABLE IF EXISTS users").Error
	if err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}

	user := &User{Name: "failed_relation", Email: "failed@example.com", Age: 25}
	err = DBCreateRelationTx(ctx, db, &User{}, user, map[string]any{})
	if err == nil {
		t.Error("expected error when creating record in non-existent table")
	}
}

func TestDBCreate_Failed(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := db.Exec("DROP TABLE IF EXISTS users").Error
	if err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}

	user := &User{Name: "failed_create", Email: "failed@example.com", Age: 25}
	err = DBCreate(ctx, db, &User{}, user)
	if err == nil {
		t.Error("expected error when creating record in non-existent table")
	}
}

func TestDBDeleteTx_Failed(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	err := db.Exec("DROP TABLE IF EXISTS users").Error
	if err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}

	err = DBDeleteTx(ctx, db, &User{}, "id = ?", 1)
	if err == nil {
		t.Error("expected error when deleting from non-existent table")
	}
}

func TestDBUpdateTx_FailedCommit(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = CloseGormDB(db) }()

	ctx := context.Background()

	user := &User{Name: "update_failed", Email: "update@example.com", Age: 25}
	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	err = db.Exec("DROP TABLE IF EXISTS users").Error
	if err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}

	err = DBUpdateTx(ctx, db, &User{}, map[string]any{"Age": 30}, "id = ?", user.ID)
	if err == nil {
		t.Error("expected error when updating in non-existent table")
	}
}
