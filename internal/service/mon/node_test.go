package mon

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

	monmodel "gin-artweb/internal/model/mon"
	resourceModel "gin-artweb/internal/model/resource"
	monrepo "gin-artweb/internal/repo/mon"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

type monNodeServiceTestContext struct {
	nodeService *MonNodeService
	nodeRepo    *monrepo.MonNodeRepo
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

	if err := db.AutoMigrate(&monmodel.MonNodeModel{}, &resourceModel.HostModel{}); err != nil {
		t.Fatalf("failed to migrate mon node model: %v", err)
	}

	dbTimeout := test.NewTestDBTimeouts()
	dbSlowThreshold := test.NewTestDBSlowThreshold()
	nodeRepo := monrepo.NewMonNodeRepo(testLogger, db, dbTimeout, dbSlowThreshold)
	nodeService := NewMonNodeService(testLogger, nodeRepo)

	return &monNodeServiceTestContext{
		nodeService: nodeService,
		nodeRepo:    nodeRepo,
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

func createTestMonNodeModel(t *testing.T, ctx *monNodeServiceTestContext, hostID uint32) *monmodel.MonNodeModel {
	dto := createTestMonNodeDTO(t, hostID)
	m := dto.ToModel()
	if err := ctx.nodeRepo.CreateModel(context.Background(), &m); err != nil {
		t.Fatalf("failed to create mon node model: %v", err)
	}
	return &m
}

func TestMonNodeService_CreateMonNode(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	dto := createTestMonNodeDTO(t, host.ID)

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

func TestMonNodeService_UpdateMonNodeByID(t *testing.T) {
	t.Parallel()

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
		dto := monmodel.MonNodeUpsertDTO{
			Name:        fmt.Sprintf("mon-node-%d", i),
			DeployPath:  "/opt/mon",
			OutportPath: "/opt/mon/out",
			JavaHome:    "/usr/local/java",
			URL:         fmt.Sprintf("http://192.168.1.%d:8080/list-%d", i+1, i),
			HostID:      host.ID,
		}
		m := dto.ToModel()
		if err := ctx.nodeRepo.CreateModel(context.Background(), &m); err != nil {
			t.Fatalf("创建MonNode应该成功: %v", err)
		}
	}

	dto := monmodel.ListMonNodeDTO{}
	count, list, svcErr := ctx.nodeService.ListMonNode(context.Background(), 1, 10, dto)
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
	_, _, svcErr := ctx.nodeService.ListMonNode(canceledCtx, 1, 10, dto)
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
	count, list, svcErr := ctx.nodeService.ListMonNode(context.Background(), 1, 10, dto)
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
	m := dto.ToModel()
	if err := ctx.nodeRepo.CreateModel(context.Background(), &m); err != nil {
		t.Fatalf("创建MonNode应该成功: %v", err)
	}

	filterDTO := monmodel.ListMonNodeDTO{
		HostID: host.ID,
	}
	count, list, svcErr := ctx.nodeService.ListMonNode(context.Background(), 1, 10, filterDTO)
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

func TestMonNodeService_ExportMonNode(t *testing.T) {
	t.Parallel()

	ctx := newMonNodeServiceTestContext(t)
	host := createTestHostModelForMon(t, ctx)
	m := createTestMonNodeModel(t, ctx, host.ID)

	outputPath := GetMonNodeExportPath(m.ID)
	svcErr := ctx.nodeService.ExportMonNode(context.Background(), *m, outputPath)
	if svcErr != nil {
		t.Fatalf("导出MonNode应该成功: %v", svcErr)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("导出文件应该存在: %s", outputPath)
	}
}

func TestGetMonNodeExportPath(t *testing.T) {
	nodeID := uint32(123)
	path := GetMonNodeExportPath(nodeID)
	expectedDir := filepath.Join(config.StorageDir, "mon", "config", "123")
	if !strings.Contains(path, expectedDir) {
		t.Errorf("路径应该包含 %s, got: %s", expectedDir, path)
	}
	if !strings.Contains(path, "mon.yaml") {
		t.Errorf("路径应该包含 mon.yaml, got: %s", path)
	}
}
