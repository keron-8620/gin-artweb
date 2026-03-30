package resource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

type pkgServiceTestContext struct {
	pkgService *PackageService
	pkgRepo    *resourcerepo.PackageRepo
	db         *gorm.DB
	tmpDir     string
	storageDir string
	testLogger *zap.Logger
}

func newPkgServiceTestContext(t *testing.T) *pkgServiceTestContext {
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

	db.AutoMigrate(&resomodel.PackageModel{})
	dbTimeout := test.NewTestDBTimeouts()
	pkgRepo := resourcerepo.NewPackageRepo(testLogger, db, dbTimeout)
	pkgService := NewPackageService(
		testLogger,
		pkgRepo,
		storageDir,
	)

	return &pkgServiceTestContext{
		pkgService: pkgService,
		pkgRepo:    pkgRepo,
		db:         db,
		tmpDir:     tmpDir,
		storageDir: storageDir,
		testLogger: testLogger,
	}
}

func createTestPackageModel(t *testing.T) *resomodel.PackageModel {
	return &resomodel.PackageModel{
		BaseModel: database.BaseModel{
			ID: 0,
		},
		Label:           "test",
		StorageFilename: fmt.Sprintf("test-package-%s.tar.gz", uuid.NewString()),
		OriginFilename:  fmt.Sprintf("test-package-%s.tar.gz", uuid.NewString()),
		Version:         fmt.Sprintf("1.0.0-%s", uuid.NewString()[:8]),
	}
}

func TestPackageService_FindPackageByID(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	pm := createTestPackageModel(t)
	err := ctx.pkgRepo.CreateModel(context.Background(), pm)
	if err != nil {
		t.Fatalf("创建Package应该成功: %v", err)
	}

	fm, svcErr := ctx.pkgService.FindPackageByID(context.Background(), pm.ID)
	if svcErr != nil {
		t.Fatalf("查询Package应该成功: %v", svcErr)
	}
	if fm == nil {
		t.Fatal("Package应该不为nil")
	}
	if fm.ID != pm.ID {
		t.Errorf("ID不匹配: expected %d, got %d", pm.ID, fm.ID)
	}
	if fm.Label != pm.Label {
		t.Errorf("Label不匹配: expected %s, got %s", pm.Label, fm.Label)
	}
}

func TestPackageService_FindPackageByID_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	fm, svcErr := ctx.pkgService.FindPackageByID(canceledCtx, 1)
	if svcErr == nil {
		t.Error("上下文取消后查询Package应该返回错误")
	}
	if fm != nil {
		t.Error("Package应该为nil")
	}
}

func TestPackageService_FindPackageByID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	fm, svcErr := ctx.pkgService.FindPackageByID(context.Background(), 999999)
	if svcErr == nil {
		t.Error("查询不存在的Package应该返回错误")
	}
	if fm != nil {
		t.Error("Package应该为nil")
	}
}

func TestPackageService_ListPackage(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	for i := 0; i < 3; i++ {
		pm := createTestPackageModel(t)
		pm.OriginFilename = fmt.Sprintf("test-package-%d.tar.gz", i)
		err := ctx.pkgRepo.CreateModel(context.Background(), pm)
		if err != nil {
			t.Fatalf("创建Package应该成功: %v", err)
		}
	}

	dto := resomodel.ListPackageDTO{}
	count, list, svcErr := ctx.pkgService.ListPackage(context.Background(), 1, 10, dto)
	if svcErr != nil {
		t.Fatalf("查询Package列表应该成功: %v", svcErr)
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

func TestPackageService_ListPackage_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	dto := resomodel.ListPackageDTO{}
	_, _, svcErr := ctx.pkgService.ListPackage(canceledCtx, 1, 10, dto)
	if svcErr == nil {
		t.Error("上下文取消后查询Package列表应该返回错误")
	}
}

func TestPackageService_ListPackage_ZeroCount(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	dto := resomodel.ListPackageDTO{
		Label: "non-existent-label",
	}
	count, list, svcErr := ctx.pkgService.ListPackage(context.Background(), 1, 10, dto)
	if svcErr != nil {
		t.Fatalf("查询不存在的Package列表应该成功: %v", svcErr)
	}
	if count != 0 {
		t.Errorf("总数应该为0, got %d", count)
	}
	if list != nil {
		t.Error("列表应该为nil")
	}
}

func TestPackageService_ListPackage_WithFilters(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	pm := createTestPackageModel(t)
	pm.Label = "specific-label"
	err := ctx.pkgRepo.CreateModel(context.Background(), pm)
	if err != nil {
		t.Fatalf("创建Package应该成功: %v", err)
	}

	dto := resomodel.ListPackageDTO{
		Label: "specific-label",
	}
	count, list, svcErr := ctx.pkgService.ListPackage(context.Background(), 1, 10, dto)
	if svcErr != nil {
		t.Fatalf("带过滤条件查询Package列表应该成功: %v", svcErr)
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

func TestPackageService_DeletePackageByID(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	pm := createTestPackageModel(t)
	err := ctx.pkgRepo.CreateModel(context.Background(), pm)
	if err != nil {
		t.Fatalf("创建Package应该成功: %v", err)
	}

	svcErr := ctx.pkgService.DeletePackageByID(context.Background(), pm.ID)
	if svcErr != nil {
		t.Fatalf("删除Package应该成功: %v", svcErr)
	}

	_, getErr := ctx.pkgService.FindPackageByID(context.Background(), pm.ID)
	if getErr == nil {
		t.Error("查询已删除的Package应该返回错误")
	}
}

func TestPackageService_DeletePackageByID_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	svcErr := ctx.pkgService.DeletePackageByID(canceledCtx, 1)
	if svcErr == nil {
		t.Error("上下文取消后删除Package应该返回错误")
	}
}

func TestPackageService_DeletePackageByID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	svcErr := ctx.pkgService.DeletePackageByID(context.Background(), 999999)
	if svcErr == nil {
		t.Error("删除不存在的Package应该返回错误")
	}
}

func TestPackageService_DeletePackageByID_FindFailed(t *testing.T) {
	t.Parallel()

	ctx := newPkgServiceTestContext(t)

	pm := createTestPackageModel(t)
	err := ctx.pkgRepo.CreateModel(context.Background(), pm)
	if err != nil {
		t.Fatalf("创建Package应该成功: %v", err)
	}

	ctx.pkgRepo.DeleteModel(context.Background(), pm.ID)

	svcErr := ctx.pkgService.DeletePackageByID(context.Background(), pm.ID)
	if svcErr == nil {
		t.Error("删除已删除的Package应该返回错误")
	}
}

func TestGetPackageStoragePath(t *testing.T) {
	filename := "test-package.tar.gz"
	path := GetPackageStoragePath(filename)
	if !strings.Contains(path, filename) {
		t.Errorf("路径应该包含文件名, got: %s", path)
	}
	if !strings.Contains(path, "packages") {
		t.Errorf("路径应该包含packages, got: %s", path)
	}
}
