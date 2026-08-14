package mon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	monmodel "gin-artweb/internal/model/mon"
	resourceModel "gin-artweb/internal/model/resource"
	monrepo "gin-artweb/internal/repo/mon"
	resourceRepo "gin-artweb/internal/repo/resource"
	resourceService "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
	"gin-artweb/pkg/archive"
	"gin-artweb/pkg/fileutil"
	"gin-artweb/pkg/serializer"
)

type monNodeServiceTestContext struct {
	nodeService *MonNodeService
	nodeRepo    *monrepo.MonNodeRepo
	pkgSvc      *resourceService.PackageService
	db          *gorm.DB
	tmpDir      string
	testLogger  *zap.Logger
}

func newMonNodeServiceTestContext(t *testing.T) *monNodeServiceTestContext {
	testLogger := test.NewTestZapLogger()
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	if err := db.AutoMigrate(
		&resourceModel.HostModel{},
		&resourceModel.PackageModel{},
		&monmodel.MonNodeModel{},
	); err != nil {
		t.Fatalf("failed to migrate mon node model: %v", err)
	}

	dbTimeout := test.NewTestDBTimeouts()
	dbSlowThreshold := test.NewTestDBSlowThreshold()
	nodeRepo := monrepo.NewMonNodeRepo(testLogger, db, dbTimeout, dbSlowThreshold)
	pkgRepo := resourceRepo.NewPackageRepo(testLogger, db, dbTimeout, dbSlowThreshold)
	pkgSvc := resourceService.NewPackageService(testLogger, pkgRepo, filepath.Join(config.StorageDir, "packages"))
	nodeService := NewMonNodeService(testLogger, nodeRepo, pkgSvc)

	return &monNodeServiceTestContext{
		nodeService: nodeService,
		nodeRepo:    nodeRepo,
		pkgSvc:      pkgSvc,
		db:          db,
		tmpDir:      tmpDir,
		testLogger:  testLogger,
	}
}

func createTestHostModelForMon(t *testing.T, ctx *monNodeServiceTestContext) *resourceModel.HostModel {
	uuidStr := uuid.NewString()
	port := 2222 + (len(uuidStr) % 1000)
	host := &resourceModel.HostModel{
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
	if err := ctx.db.Create(host).Error; err != nil {
		t.Fatalf("failed to create test host: %v", err)
	}
	return host
}

func createTestMonNodeDTO(t *testing.T, hostID uint32) monmodel.MonNodeUpsertDTO {
	uuidStr := uuid.NewString()
	return monmodel.MonNodeUpsertDTO{
		Name:        fmt.Sprintf("mon-node-%s", uuidStr[:8]),
		DeployPath:  "/opt/mon",
		OutportPath: "/opt/mon/out",
		JavaHome:    "/usr/local/java",
		URL:         fmt.Sprintf("http://192.168.1.%d:8080", len(uuidStr)%255),
		HostID:      hostID,
	}
}

func createTestPackageModel(t *testing.T, ctx *monNodeServiceTestContext, label, version string) *resourceModel.PackageModel {
	return createTestPackageModelWithConfig(t, ctx, label, version, true)
}

func createTestPackageModelWithConfig(t *testing.T, ctx *monNodeServiceTestContext, label, version string, withConfig bool) *resourceModel.PackageModel {
	filename := fmt.Sprintf("%s-%s-%s.tar.gz", label, version, uuid.NewString())
	packageModel := &resourceModel.PackageModel{
		Label:           label,
		StorageFilename: filename,
		OriginFilename:  filename,
		Version:         version,
	}

	storagePath := resourceService.GetPackageStoragePath(filename)
	sourceDir := t.TempDir()
	rootDir := filepath.Join(sourceDir, "mon-package")
	if err := os.MkdirAll(rootDir, 0750); err != nil {
		t.Fatalf("failed to create test package root directory: %v", err)
	}
	if withConfig {
		if err := os.MkdirAll(filepath.Join(rootDir, "config"), 0750); err != nil {
			t.Fatalf("failed to create test package directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(rootDir, "config", "default.conf"), []byte("test"), 0600); err != nil {
			t.Fatalf("failed to create test package config: %v", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(storagePath), 0750); err != nil {
		t.Fatalf("failed to create package storage directory: %v", err)
	}
	if err := archive.TarGz(rootDir, storagePath); err != nil {
		t.Fatalf("failed to create test package archive: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(storagePath) })
	if err := ctx.db.Create(packageModel).Error; err != nil {
		t.Fatalf("failed to create test package model: %v", err)
	}

	return packageModel
}

func createTestMonNodeModelWithoutArchive(t *testing.T, ctx *monNodeServiceTestContext, hostID uint32) *monmodel.MonNodeModel {
	dto := createTestMonNodeDTO(t, hostID)
	monPackage := createTestPackageRecord(t, ctx, "mon", uuid.NewString())
	jdkPackage := createTestPackageRecord(t, ctx, "jdk", uuid.NewString())
	dto.PackageID = monPackage.ID
	dto.JdkID = jdkPackage.ID
	m := dto.ToModel()
	if err := ctx.nodeRepo.CreateModel(context.Background(), &m); err != nil {
		t.Fatalf("failed to create mon node model: %v", err)
	}
	return &m
}

func createTestPackageRecord(t *testing.T, ctx *monNodeServiceTestContext, label, version string) *resourceModel.PackageModel {
	packageModel := &resourceModel.PackageModel{
		Label:           label,
		StorageFilename: fmt.Sprintf("%s-%s-%s.tar.gz", label, version, uuid.NewString()),
		OriginFilename:  "test.tar.gz",
		Version:         version,
	}
	if err := ctx.db.Create(packageModel).Error; err != nil {
		t.Fatalf("failed to create test package model: %v", err)
	}
	return packageModel
}

func createTestMonNodeModel(t *testing.T, ctx *monNodeServiceTestContext, hostID uint32) *monmodel.MonNodeModel {
	dto := createTestMonNodeDTO(t, hostID)
	monPackage := createTestPackageModel(t, ctx, "mon", "1.0.0")
	jdkPackage := createTestPackageModel(t, ctx, "jdk", "17")
	dto.PackageID = monPackage.ID
	dto.JdkID = jdkPackage.ID
	m := dto.ToModel()
	if err := ctx.nodeRepo.CreateModel(context.Background(), &m); err != nil {
		t.Fatalf("failed to create mon node model: %v", err)
	}
	if err := ctx.db.Preload("Host").Preload("Package").Preload("Jdk").First(&m, m.ID).Error; err != nil {
		t.Fatalf("failed to preload test mon node relations: %v", err)
	}
	return &m
}

func TestMonNodeService_CreateMonNode(t *testing.T) {
	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	dto := createTestMonNodeDTO(t, host.ID)
	monPackage := createTestPackageModel(t, ctx, "mon", "1.0.0")
	jdkPackage := createTestPackageModel(t, ctx, "jdk", "17")
	dto.PackageID = monPackage.ID
	dto.JdkID = jdkPackage.ID

	result, svcErr := ctx.nodeService.CreateMonNode(context.Background(), dto)
	if svcErr != nil {
		t.Fatalf("创建MonNode应该成功: %v", svcErr)
	}
	if result == nil {
		t.Fatal("MonNode应该不为nil")
	}
	if result.ID == 0 {
		t.Error("MonNode ID应该不为0")
	}
	if result.Name != dto.Name {
		t.Errorf("Name不匹配: expected %s, got %s", dto.Name, result.Name)
	}
	if result.DeployPath != dto.DeployPath {
		t.Errorf("DeployPath不匹配: expected %s, got %s", dto.DeployPath, result.DeployPath)
	}
	if result.HostID != dto.HostID {
		t.Errorf("HostID不匹配: expected %d, got %d", dto.HostID, result.HostID)
	}
	if result.PackageID == nil || *result.PackageID != dto.PackageID {
		t.Errorf("PackageID不匹配: expected %d, got %d", dto.PackageID, result.PackageID)
	}
	if result.JdkID == nil || *result.JdkID != dto.JdkID {
		t.Errorf("JdkID不匹配: expected %d, got %d", dto.JdkID, result.JdkID)
	}
	if result.Package.ID != dto.PackageID {
		t.Errorf("程序包关联不匹配: expected %d, got %d", dto.PackageID, result.Package.ID)
	}
	if result.Jdk.ID != dto.JdkID {
		t.Errorf("JDK关联不匹配: expected %d, got %d", dto.JdkID, result.Jdk.ID)
	}
}

func TestMonNodeService_CreateMonNode_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	dto := createTestMonNodeDTO(t, host.ID)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	result, svcErr := ctx.nodeService.CreateMonNode(canceledCtx, dto)
	if svcErr == nil {
		t.Error("上下文取消后创建MonNode应该返回错误")
	}
	if result != nil {
		t.Error("MonNode应该为nil")
	}
}

func TestMonNodeService_CreateMonNode_DatabaseError(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)
	dto := monmodel.MonNodeUpsertDTO{
		Name:        "test-mon",
		DeployPath:  "/opt/mon",
		OutportPath: "/opt/mon/out",
		JavaHome:    "/usr/local/java",
		URL:         "http://localhost:8080",
		HostID:      999999,
	}

	result, svcErr := ctx.nodeService.CreateMonNode(context.Background(), dto)
	if svcErr == nil {
		t.Error("创建不存在的Host关联的MonNode应该返回错误")
	}
	if result != nil {
		t.Error("MonNode应该为nil")
	}
}

func TestMonNodeService_CreateMonNode_RequiresPackages(t *testing.T) {
	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	dto := createTestMonNodeDTO(t, host.ID)

	result, svcErr := ctx.nodeService.CreateMonNode(context.Background(), dto)
	if svcErr == nil {
		t.Fatal("未指定程序包和JDK时创建MonNode应该返回错误")
	}
	if result != nil {
		t.Fatal("校验失败时MonNode应该为nil")
	}
}

func TestMonNodeService_UpdateMonNodeByID(t *testing.T) {
	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	original := createTestMonNodeModel(t, ctx, host.ID)

	updateDTO := monmodel.MonNodeUpsertDTO{
		Name:        fmt.Sprintf("updated-%s", original.Name),
		DeployPath:  "/updated/mon",
		OutportPath: "/updated/mon/out",
		JavaHome:    "/updated/java",
		URL:         "http://updated:8080",
		HostID:      host.ID,
		PackageID:   *original.PackageID,
		JdkID:       *original.JdkID,
	}

	result, svcErr := ctx.nodeService.UpdateMonNodeByID(context.Background(), original.ID, updateDTO)
	if svcErr != nil {
		t.Fatalf("更新MonNode应该成功: %v", svcErr)
	}
	if result == nil {
		t.Fatal("MonNode应该不为nil")
	}
	if result.Name != updateDTO.Name {
		t.Errorf("Name不匹配: expected %s, got %s", updateDTO.Name, result.Name)
	}
	if result.DeployPath != updateDTO.DeployPath {
		t.Errorf("DeployPath不匹配: expected %s, got %s", updateDTO.DeployPath, result.DeployPath)
	}
	if result.PackageID == nil || *result.PackageID != updateDTO.PackageID {
		t.Errorf("PackageID不匹配: expected %d, got %d", updateDTO.PackageID, result.PackageID)
	}
	if result.JdkID == nil || *result.JdkID != updateDTO.JdkID {
		t.Errorf("JdkID不匹配: expected %d, got %d", updateDTO.JdkID, result.JdkID)
	}
}

func TestMonNodeService_UpdateMonNodeByID_RequiresPackages(t *testing.T) {
	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	original := createTestMonNodeModel(t, ctx, host.ID)
	updateDTO := createTestMonNodeDTO(t, host.ID)

	result, svcErr := ctx.nodeService.UpdateMonNodeByID(context.Background(), original.ID, updateDTO)
	if svcErr == nil {
		t.Fatal("未指定程序包和JDK时更新MonNode应该返回错误")
	}
	if result != nil {
		t.Fatal("校验失败时MonNode应该为nil")
	}
}

func TestMonNodeService_UpdateMonNodeByID_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	original := createTestMonNodeModel(t, ctx, host.ID)

	updateDTO := createTestMonNodeDTO(t, host.ID)
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	result, svcErr := ctx.nodeService.UpdateMonNodeByID(canceledCtx, original.ID, updateDTO)
	if svcErr == nil {
		t.Error("上下文取消后更新MonNode应该返回错误")
	}
	if result != nil {
		t.Error("MonNode应该为nil")
	}
}

func TestMonNodeService_UpdateMonNodeByID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	updateDTO := createTestMonNodeDTO(t, host.ID)

	result, svcErr := ctx.nodeService.UpdateMonNodeByID(context.Background(), 999999, updateDTO)
	if svcErr == nil {
		t.Error("更新不存在的MonNode应该返回错误")
	}
	if result != nil {
		t.Error("MonNode应该为nil")
	}
}

func TestMonNodeService_DeleteMonNodeByID(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	original := createTestMonNodeModel(t, ctx, host.ID)
	_ = fileutil.RemoveAll(context.Background(), GetMonNodeBinDir(original.ID))
	_ = fileutil.RemoveAll(context.Background(), GetMonNodeConfDir(original.ID))

	svcErr := ctx.nodeService.DeleteMonNodeByID(context.Background(), original.ID)
	if svcErr != nil {
		t.Fatalf("删除MonNode应该成功: %v", svcErr)
	}

	_, getErr := ctx.nodeService.FindMonNodeByID(context.Background(), []string{}, original.ID)
	if getErr == nil {
		t.Error("查询已删除的MonNode应该返回错误")
	}
}

func TestMonNodeService_DeleteMonNodeByID_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	svcErr := ctx.nodeService.DeleteMonNodeByID(canceledCtx, 1)
	if svcErr == nil {
		t.Error("上下文取消后删除MonNode应该返回错误")
	}
}

func TestMonNodeService_DeleteMonNodeByID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)

	svcErr := ctx.nodeService.DeleteMonNodeByID(context.Background(), 999999)
	if svcErr != nil {
		t.Error("删除不存在的MonNode应该成功（无操作）")
	}
}

func TestMonNodeService_FindMonNodeByID(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	original := createTestMonNodeModel(t, ctx, host.ID)

	result, svcErr := ctx.nodeService.FindMonNodeByID(context.Background(), []string{"Host"}, original.ID)
	if svcErr != nil {
		t.Fatalf("查询MonNode应该成功: %v", svcErr)
	}
	if result == nil {
		t.Fatal("MonNode应该不为nil")
	}
	if result.ID != original.ID {
		t.Errorf("ID不匹配: expected %d, got %d", original.ID, result.ID)
	}
	if result.Name != original.Name {
		t.Errorf("Name不匹配: expected %s, got %s", original.Name, result.Name)
	}
}

func TestMonNodeService_FindMonNodeByID_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	result, svcErr := ctx.nodeService.FindMonNodeByID(canceledCtx, []string{}, 1)
	if svcErr == nil {
		t.Error("上下文取消后查询MonNode应该返回错误")
	}
	if result != nil {
		t.Error("MonNode应该为nil")
	}
}

func TestMonNodeService_FindMonNodeByID_NotFound(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)

	result, svcErr := ctx.nodeService.FindMonNodeByID(context.Background(), []string{}, 999999)
	if svcErr == nil {
		t.Error("查询不存在的MonNode应该返回错误")
	}
	if result != nil {
		t.Error("MonNode应该为nil")
	}
}

func TestMonNodeService_ListMonNode(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)

	for i := 0; i < 3; i++ {
		m := createTestMonNodeModelWithoutArchive(t, ctx, host.ID)
		m.Name = fmt.Sprintf("mon-node-%d", i)
		m.URL = fmt.Sprintf("http://192.168.1.%d:8080/list-%d", i+1, i)
		if err := ctx.db.Model(&monmodel.MonNodeModel{}).Where("id = ?", m.ID).Updates(map[string]any{
			"name": m.Name,
			"url":  m.URL,
		}).Error; err != nil {
			t.Fatalf("更新测试MonNode应该成功: %v", err)
		}
	}

	dto := monmodel.ListMonNodeDTO{}
	_, _, count, list, svcErr := ctx.nodeService.ListMonNode(context.Background(), dto)
	if svcErr != nil {
		t.Fatalf("查询MonNode列表应该成功: %v", svcErr)
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

func TestMonNodeService_ListMonNode_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	dto := monmodel.ListMonNodeDTO{}
	_, _, _, _, svcErr := ctx.nodeService.ListMonNode(canceledCtx, dto)
	if svcErr == nil {
		t.Error("上下文取消后查询MonNode列表应该返回错误")
	}
}

func TestMonNodeService_ListMonNode_ZeroCount(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)

	dto := monmodel.ListMonNodeDTO{
		Name: "non-existent-mon",
	}
	_, _, count, list, svcErr := ctx.nodeService.ListMonNode(context.Background(), dto)
	if svcErr != nil {
		t.Fatalf("查询不存在的MonNode列表应该成功: %v", svcErr)
	}
	if count != 0 {
		t.Errorf("总数应该为0, got %d", count)
	}
	if list != nil {
		t.Error("列表应该为nil")
	}
}

func TestMonNodeService_ListMonNode_WithHostIDFilter(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)

	dto := createTestMonNodeDTO(t, host.ID)
	dto.Name = "specific-mon-node"
	monPackage := createTestPackageRecord(t, ctx, "mon", uuid.NewString())
	jdkPackage := createTestPackageRecord(t, ctx, "jdk", uuid.NewString())
	dto.PackageID = monPackage.ID
	dto.JdkID = jdkPackage.ID
	m := dto.ToModel()
	if err := ctx.nodeRepo.CreateModel(context.Background(), &m); err != nil {
		t.Fatalf("创建MonNode应该成功: %v", err)
	}

	filterDTO := monmodel.ListMonNodeDTO{
		HostID: host.ID,
	}
	_, _, count, list, svcErr := ctx.nodeService.ListMonNode(context.Background(), filterDTO)
	if svcErr != nil {
		t.Fatalf("带HostID过滤条件查询MonNode列表应该成功: %v", svcErr)
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

func TestMonNodeService_OutportMonData(t *testing.T) {
	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	m := createTestMonNodeModel(t, ctx, host.ID)

	confDir := GetMonNodeConfDir(m.ID)
	t.Cleanup(func() { _ = os.RemoveAll(confDir) })
	svcErr := ctx.nodeService.OutportMonData(context.Background(), *m)
	if svcErr != nil {
		t.Fatalf("导出MonNode配置应该成功: %v", svcErr)
	}

	outputPath := filepath.Join(confDir, "mon.yaml")
	var vars monmodel.MonNodeVars
	if _, err := serializer.ReadYAML(outputPath, &vars); err != nil {
		t.Fatalf("导出的MonNode配置应该可以解析: %v", err)
	}
	if vars.ID != m.ID {
		t.Errorf("导出ID不匹配: expected %d, got %d", m.ID, vars.ID)
	}
	if vars.Name != m.Name {
		t.Errorf("导出Name不匹配: expected %s, got %s", m.Name, vars.Name)
	}
	if vars.DeployPath != m.DeployPath {
		t.Errorf("导出DeployPath不匹配: expected %s, got %s", m.DeployPath, vars.DeployPath)
	}
	if vars.OutportPath != m.OutportPath {
		t.Errorf("导出OutportPath不匹配: expected %s, got %s", m.OutportPath, vars.OutportPath)
	}
	if vars.JavaHome != m.JavaHome {
		t.Errorf("导出JavaHome不匹配: expected %s, got %s", m.JavaHome, vars.JavaHome)
	}
	if vars.URL != m.URL {
		t.Errorf("导出URL不匹配: expected %s, got %s", m.URL, vars.URL)
	}
	if vars.HostID != m.HostID {
		t.Errorf("导出HostID不匹配: expected %d, got %d", m.HostID, vars.HostID)
	}
}

func TestMonNodeService_OutportMonData_WithoutPackageConfig(t *testing.T) {
	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	m := createTestMonNodeModel(t, ctx, host.ID)
	monPackage := createTestPackageModelWithConfig(t, ctx, "mon", "1.0.1", false)
	m.Package = monPackage
	m.PackageID = &monPackage.ID

	confDir := GetMonNodeConfDir(m.ID)
	t.Cleanup(func() {
		_ = os.RemoveAll(GetMonNodeBinDir(m.ID))
		_ = os.RemoveAll(confDir)
	})
	if svcErr := ctx.nodeService.OutportMonData(context.Background(), *m); svcErr != nil {
		t.Fatalf("缺少config目录的MonNode程序包仍应该可以导出配置: %v", svcErr)
	}
	if _, err := os.Stat(filepath.Join(confDir, "mon.yaml")); err != nil {
		t.Fatalf("应该生成mon.yaml: %v", err)
	}
}

func TestGetMonNodeConfDir(t *testing.T) {
	nodeID := uint32(123)
	path := GetMonNodeConfDir(nodeID)
	expectedPath := filepath.Join(config.StorageDir, "mon", "config", "123")
	if path != expectedPath {
		t.Errorf("配置目录不匹配: expected %s, got: %s", expectedPath, path)
	}
}
