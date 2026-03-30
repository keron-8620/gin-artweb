package resource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	resomodel "gin-artweb/internal/model/resource"
	resourcerepo "gin-artweb/internal/repo/resource"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

type hostServiceTestContext struct {
	hostService *HostService
	hostRepo    *resourcerepo.HostRepo
	db          *gorm.DB
	tmpDir      string
	storageDir  string
	testLogger  *zap.Logger
}

func newHostServiceTestContext(t *testing.T) *hostServiceTestContext {
	testLogger := test.NewTestZapLogger()
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, "storage")
	err := os.MkdirAll(storageDir, 0755)
	if err != nil {
		t.Fatalf("failed to create temp storage dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	db.AutoMigrate(&resomodel.HostModel{})
	dbTimeout := test.NewTestDBTimeouts()
	hostRepo := resourcerepo.NewHostRepo(testLogger, db, dbTimeout)
	hostService := NewHostService(
		testLogger,
		hostRepo,
		5*time.Second,
		nil,
		nil,
	)

	return &hostServiceTestContext{
		hostService: hostService,
		hostRepo:    hostRepo,
		db:          db,
		tmpDir:      tmpDir,
		storageDir:  storageDir,
		testLogger:  testLogger,
	}
}

func createTestHostModel(t *testing.T) *resomodel.HostModel {
	uuidStr := uuid.NewString()
	port := 2222 + (len(uuidStr) % 1000)
	return &resomodel.HostModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{
				ID: 0,
			},
		},
		Name:    fmt.Sprintf("host-%s", uuidStr),
		Label:   "test",
		SSHIP:   "127.0.0.1",
		SSHPort: uint16(port),
		SSHUser: fmt.Sprintf("root-%s", uuidStr[len(uuidStr)-4:]),
		PyPath:  "/usr/bin/python3",
		Remark:  "",
	}
}

func TestHostService_FindHostById(t *testing.T) {
	t.Parallel()

	ctx := newHostServiceTestContext(t)

	hm := createTestHostModel(t)
	err := ctx.hostRepo.CreateModel(context.Background(), hm)
	if err != nil {
		t.Fatalf("创建Host应该成功: %v", err)
	}

	fm, svcErr := ctx.hostService.FindHostById(context.Background(), hm.ID)
	if svcErr != nil {
		t.Fatalf("查询Host应该成功: %v", svcErr)
	}
	if fm == nil {
		t.Fatal("Host应该不为nil")
	}
	if fm.ID != hm.ID {
		t.Errorf("ID不匹配: expected %d, got %d", hm.ID, fm.ID)
	}
	if fm.Name != hm.Name {
		t.Errorf("Name不匹配: expected %s, got %s", hm.Name, fm.Name)
	}
}

func TestHostService_FindHostById_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newHostServiceTestContext(t)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	fm, svcErr := ctx.hostService.FindHostById(canceledCtx, 1)
	if svcErr == nil {
		t.Error("上下文取消后查询Host应该返回错误")
	}
	if fm != nil {
		t.Error("Host应该为nil")
	}
}

func TestHostService_FindHostById_NotFound(t *testing.T) {
	t.Parallel()

	ctx := newHostServiceTestContext(t)

	fm, svcErr := ctx.hostService.FindHostById(context.Background(), 999999)
	if svcErr == nil {
		t.Error("查询不存在的Host应该返回错误")
	}
	if fm != nil {
		t.Error("Host应该为nil")
	}
}

func TestHostService_ListHost(t *testing.T) {
	t.Parallel()

	ctx := newHostServiceTestContext(t)

	for i := 0; i < 3; i++ {
		hm := createTestHostModel(t)
		hm.Name = fmt.Sprintf("host-%d", i)
		err := ctx.hostRepo.CreateModel(context.Background(), hm)
		if err != nil {
			t.Fatalf("创建Host应该成功: %v", err)
		}
	}

	dto := resomodel.ListHostDTO{}
	count, list, svcErr := ctx.hostService.ListHost(context.Background(), 1, 10, dto)
	if svcErr != nil {
		t.Fatalf("查询Host列表应该成功: %v", svcErr)
	}
	if count <= 0 {
		t.Errorf("总数应该大于0, got %d", count)
	}
	if list == nil {
		t.Error("列表应该不为nil")
	}
	if len(list) <= 0 {
		t.Error("列表长度应该大于0")
	}
}

func TestHostService_ListHost_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newHostServiceTestContext(t)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	dto := resomodel.ListHostDTO{}
	_, _, svcErr := ctx.hostService.ListHost(canceledCtx, 1, 10, dto)
	if svcErr == nil {
		t.Error("上下文取消后查询Host列表应该返回错误")
	}
}

func TestHostService_ListHost_ZeroCount(t *testing.T) {
	t.Parallel()

	ctx := newHostServiceTestContext(t)

	dto := resomodel.ListHostDTO{
		Name: "non-existent-host",
	}
	count, list, svcErr := ctx.hostService.ListHost(context.Background(), 1, 10, dto)
	if svcErr != nil {
		t.Fatalf("查询不存在的Host列表应该成功: %v", svcErr)
	}
	if count != 0 {
		t.Errorf("总数应该为0, got %d", count)
	}
	if list != nil {
		t.Error("列表应该为nil")
	}
}

func TestHostService_ListHost_WithFilters(t *testing.T) {
	t.Parallel()

	ctx := newHostServiceTestContext(t)

	hm := createTestHostModel(t)
	hm.Label = "specific-label"
	err := ctx.hostRepo.CreateModel(context.Background(), hm)
	if err != nil {
		t.Fatalf("创建Host应该成功: %v", err)
	}

	dto := resomodel.ListHostDTO{
		Label: "specific-label",
	}
	count, list, svcErr := ctx.hostService.ListHost(context.Background(), 1, 10, dto)
	if svcErr != nil {
		t.Fatalf("带过滤条件查询Host列表应该成功: %v", svcErr)
	}
	if count < 1 {
		t.Errorf("总数应该大于等于1, got %d", count)
	}
	if list == nil {
		t.Error("列表应该不为nil")
	}
	if len(list) < 1 {
		t.Error("列表长度应该大于等于1")
	}
}

func TestHostService_DeleteHostById(t *testing.T) {
	t.Parallel()

	ctx := newHostServiceTestContext(t)

	hm := createTestHostModel(t)
	err := ctx.hostRepo.CreateModel(context.Background(), hm)
	if err != nil {
		t.Fatalf("创建Host应该成功: %v", err)
	}

	svcErr := ctx.hostService.DeleteHostById(context.Background(), hm.ID)
	if svcErr != nil {
		t.Fatalf("删除Host应该成功: %v", svcErr)
	}

	_, getErr := ctx.hostService.FindHostById(context.Background(), hm.ID)
	if getErr == nil {
		t.Error("查询已删除的Host应该返回错误")
	}
}

func TestHostService_DeleteHostById_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newHostServiceTestContext(t)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	svcErr := ctx.hostService.DeleteHostById(canceledCtx, 1)
	if svcErr == nil {
		t.Error("上下文取消后删除Host应该返回错误")
	}
}

func TestHostService_DeleteHostById_NotFound(t *testing.T) {
	t.Parallel()

	ctx := newHostServiceTestContext(t)

	svcErr := ctx.hostService.DeleteHostById(context.Background(), 999999)
	if svcErr != nil {
		t.Error("删除不存在的Host应该成功（无操作）")
	}
}

func TestGetHostVarsExportPath(t *testing.T) {
	hostID := uint32(123)
	path := GetHostVarsExportPath(hostID)
	if !strings.Contains(path, "host_123.yaml") {
		t.Errorf("路径应该包含host_123.yaml, got: %s", path)
	}
	if !strings.Contains(path, "host_vars") {
		t.Errorf("路径应该包含host_vars, got: %s", path)
	}
}
