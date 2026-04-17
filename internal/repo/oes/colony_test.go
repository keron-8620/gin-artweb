package oes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	monmodel "gin-artweb/internal/model/mon"
	oesmodel "gin-artweb/internal/model/oes"
	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

// TestOesColonyModel 测试数据结构
type TestOesColonyModel struct {
	SystemType    string
	ColonyNum     string
	ExtractedName string
	IsEnable      bool
	PackageID     uint32
	XCounterID    uint32
	MonNodeID     uint32
}

// 创建测试OesColony模型的选项
type CreateTestOesColonyOption func(*TestOesColonyModel)

// WithSystemType 设置系统类型
func WithSystemType(systemType string) CreateTestOesColonyOption {
	return func(model *TestOesColonyModel) {
		model.SystemType = systemType
	}
}

// WithColonyNum 设置集群编号
func WithColonyNum(colonyNum string) CreateTestOesColonyOption {
	return func(model *TestOesColonyModel) {
		model.ColonyNum = colonyNum
	}
}

// WithExtractedName 设置提取名称
func WithExtractedName(extractedName string) CreateTestOesColonyOption {
	return func(model *TestOesColonyModel) {
		model.ExtractedName = extractedName
	}
}

// WithIsEnable 设置启用状态
func WithIsEnable(isEnable bool) CreateTestOesColonyOption {
	return func(model *TestOesColonyModel) {
		model.IsEnable = isEnable
	}
}

// WithPackageID 设置包ID
func WithPackageID(packageID uint32) CreateTestOesColonyOption {
	return func(model *TestOesColonyModel) {
		model.PackageID = packageID
	}
}

// WithXCounterID 设置XCounter包ID
func WithXCounterID(xCounterID uint32) CreateTestOesColonyOption {
	return func(model *TestOesColonyModel) {
		model.XCounterID = xCounterID
	}
}

// WithMonNodeID 设置Mon节点ID
func WithMonNodeID(monNodeID uint32) CreateTestOesColonyOption {
	return func(model *TestOesColonyModel) {
		model.MonNodeID = monNodeID
	}
}

var oesColonyCounter int = 0

// CreateTestOesColonyModel 创建测试用的OesColony模型
func CreateTestOesColonyModel(options ...CreateTestOesColonyOption) *oesmodel.OesColonyModel {
	oesColonyCounter++

	// 默认值
	model := &TestOesColonyModel{
		SystemType:    "STK",
		ColonyNum:     fmt.Sprintf("%02d", oesColonyCounter%100),
		ExtractedName: fmt.Sprintf("oes-%s", uuid.NewString()),
		IsEnable:      true,
		PackageID:     1, // 假设程序包ID为1
		XCounterID:    2, // 假设xcounter包ID为2
		MonNodeID:     1, // 假设Mon节点ID为1
	}

	// 应用选项
	for _, option := range options {
		option(model)
	}

	// 转换为实际的模型
	return &oesmodel.OesColonyModel{
		SystemType:    model.SystemType,
		ColonyNum:     model.ColonyNum,
		ExtractedName: model.ExtractedName,
		IsEnable:      model.IsEnable,
		PackageID:     model.PackageID,
		XCounterID:    model.XCounterID,
		MonNodeID:     model.MonNodeID,
	}
}

type OesColonyTestSuite struct {
	suite.Suite
	colonyRepo *OesColonyRepo
	log        *zap.Logger
	gormDB     *gorm.DB
	timeouts   *config.DBTimeout
}

func (suite *OesColonyTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(
		&resomodel.HostModel{},
		&monmodel.MonNodeModel{},
		&resomodel.PackageModel{},
		&oesmodel.OesColonyModel{},
	); err != nil {
		suite.Error(err, "数据库迁移失败")
	}

	// 创建测试数据:主机
	hostModel := &resomodel.HostModel{
		Name:    "test-host",
		Label:   "test",
		SSHIP:   "127.0.0.1",
		SSHPort: 22,
		SSHUser: "root",
		PyPath:  "/usr/bin/python3",
		Remark:  "",
	}
	db.Create(hostModel)

	// 创建测试数据:Mon节点
	monNodeModel := &monmodel.MonNodeModel{
		Name:        "test-mon-node",
		DeployPath:  "/opt/mon",
		OutportPath: "/opt/mon/outport",
		JavaHome:    "/usr/lib/jvm/java-11-openjdk-amd64",
		URL:         "http://localhost:8080/mon",
		HostID:      1,
	}
	db.Create(monNodeModel)

	// 创建测试数据:程序包1
	packageModel1 := &resomodel.PackageModel{
		Label:           "test-oes",
		StorageFilename: "test-oes.tar.gz",
		OriginFilename:  "test-oes.tar.gz",
		Version:         "1.0.0",
	}
	db.Create(packageModel1)

	// 创建测试数据:程序包2 (xcounter)
	packageModel2 := &resomodel.PackageModel{
		Label:           "test-xcounter",
		StorageFilename: "test-xcounter.tar.gz",
		OriginFilename:  "test-xcounter.tar.gz",
		Version:         "1.0.0",
	}
	db.Create(packageModel2)

	suite.timeouts = test.NewTestDBTimeouts()
	suite.log = test.NewTestZapLogger()
	suite.gormDB = db
	slowThreshold := test.NewTestDBSlowThreshold()
	suite.colonyRepo = &OesColonyRepo{
		log:           suite.log,
		gormDB:        suite.gormDB,
		timeouts:      suite.timeouts,
		slowThreshold: slowThreshold,
	}
}

func (suite *OesColonyTestSuite) TestCreateModel() {
	// 测试正常创建
	cm := CreateTestOesColonyModel()
	err := suite.colonyRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "正常创建OesColony模型应该成功")
	suite.NotZero(cm.ID, "创建成功后OesColony ID应该不为零")

	// 测试边界情况:创建空模型
	err = suite.colonyRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空OesColony模型应该返回错误")
}

func (suite *OesColonyTestSuite) TestUpdateModel() {
	// 创建测试数据
	cm := CreateTestOesColonyModel()
	err := suite.colonyRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesColony用于更新测试应该成功")

	// 测试正常更新 - 不更新ColonyNum以避免唯一约束冲突
	updateData := map[string]any{
		"ExtractedName": "updated-oes",
		"IsEnable":      false,
	}
	err = suite.colonyRepo.UpdateModel(context.Background(), updateData, "id = ?", cm.ID)
	suite.NoError(err, "正常更新OesColony应该成功")

	// 验证更新结果
	fm, err := suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询更新后的OesColony应该成功")
	suite.Equal(cm.ColonyNum, fm.ColonyNum, "ColonyNum应该保持不变")
	suite.Equal("updated-oes", fm.ExtractedName, "ExtractedName应该被更新为'updated-oes'")
	suite.False(fm.IsEnable, "IsEnable应该被更新为false")

	// 测试边界情况:更新数据为空
	err = suite.colonyRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", cm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况:更新不存在的OesColony
	err = suite.colonyRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的OesColony应该成功（无操作）")
}

func (suite *OesColonyTestSuite) TestDeleteModel() {
	// 创建测试数据
	cm := CreateTestOesColonyModel()
	err := suite.colonyRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesColony用于删除测试应该成功")

	// 测试正常删除
	err = suite.colonyRepo.DeleteModel(context.Background(), "id = ?", cm.ID)
	suite.NoError(err, "正常删除OesColony应该成功")

	// 验证删除结果
	fm, err := suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.Error(err, "查询已删除的OesColony应该返回错误")
	suite.Nil(fm, "已删除的OesColony应该为nil")

	// 测试边界情况:删除不存在的OesColony
	err = suite.colonyRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的OesColony应该成功（无操作）")
}

func (suite *OesColonyTestSuite) TestGetModel() {
	// 创建测试数据
	cm := CreateTestOesColonyModel()
	err := suite.colonyRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesColony用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "正常查询OesColony应该成功")
	suite.Equal(cm.ID, fm.ID, "查询结果的ID应该与创建时的ID一致")
	suite.Equal(cm.ColonyNum, fm.ColonyNum, "查询结果的ColonyNum应该与创建时的一致")
	suite.Equal(cm.IsEnable, fm.IsEnable, "查询结果的IsEnable应该与创建时的一致")

	// 测试边界情况:查询不存在的OesColony
	fm, err = suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的OesColony应该返回错误")
	suite.Nil(fm, "查询不存在的OesColony应该返回nil")

	// 测试边界情况:使用预加载
	fm, err = suite.colonyRepo.GetModel(context.Background(), []string{"Package", "XCounter", "MonNode"}, "id = ?", cm.ID)
	suite.NoError(err, "使用预加载查询OesColony应该成功")
	suite.Equal(cm.ID, fm.ID, "预加载查询结果的ID应该与创建时的ID一致")
}

func (suite *OesColonyTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		cm := CreateTestOesColonyModel()
		err := suite.colonyRepo.CreateModel(context.Background(), cm)
		suite.NoError(err, "创建OesColony用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Limit:   10,
		Offset:  0,
	}
	models, err := suite.colonyRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "正常查询OesColony列表应该成功")
	suite.NotNil(models, "OesColony列表应该不为nil")
	suite.Greater(len(models), 0, "OesColony列表长度应该大于0")

	// 测试边界情况:空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"colony_num": "99"},
	}
	models2, err := suite.colonyRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的OesColony列表应该成功")
	suite.NotNil(models2, "不存在的OesColony列表应该不为nil")
	suite.Len(models2, 0, "不存在的OesColony列表长度应该为0")
}

func (suite *OesColonyTestSuite) TestContextTimeout() {
	// 创建测试数据
	cm := CreateTestOesColonyModel()
	err := suite.colonyRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesColony用于超时测试应该成功")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err = suite.colonyRepo.GetModel(timeoutCtx, nil, "id = ?", cm.ID)
	suite.Error(err, "上下文超时后查询OesColony应该返回错误")
}

func (suite *OesColonyTestSuite) TestCountModel() {
	// 创建测试数据
	for i := 0; i < 3; i++ {
		cm := CreateTestOesColonyModel()
		err := suite.colonyRepo.CreateModel(context.Background(), cm)
		suite.NoError(err, "创建OesColony用于计数测试应该成功")
	}

	// 测试正常计数
	count, err := suite.colonyRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "正常查询OesColony总数应该成功")
	suite.Greater(count, int64(0), "OesColony总数应该大于0")

	// 测试带条件计数
	count2, err := suite.colonyRepo.CountModel(context.Background(), map[string]any{"system_type": "STK"})
	suite.NoError(err, "带条件查询OesColony总数应该成功")
	suite.GreaterOrEqual(count2, int64(3), "带条件的OesColony总数应该大于等于3")

	// 测试边界情况:查询不存在的类型
	count3, err := suite.colonyRepo.CountModel(context.Background(), map[string]any{"system_type": "NON-EXISTENT"})
	suite.NoError(err, "查询不存在类型的OesColony总数应该成功")
	suite.Equal(int64(0), count3, "不存在类型的OesColony总数应该为0")
}

// 辅助函数：清理测试目录
func (suite *OesColonyTestSuite) cleanupTestDir(dir string) {
	os.RemoveAll(dir)
}

// 辅助函数：创建测试文件
func (suite *OesColonyTestSuite) createTestFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// errorReader 是一个模拟的 reader，在读取时返回错误
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("模拟读取错误")
}

func (suite *OesColonyTestSuite) TestSaveConfigFile() {
	// 测试正常保存配置文件
	testContent := "test configuration content"
	fileReader := bytes.NewReader([]byte(testContent))
	testDir := "/tmp/oes-test"
	testPath := filepath.Join(testDir, "config.json")

	// 确保测试目录不存在
	suite.cleanupTestDir(testDir)
	defer suite.cleanupTestDir(testDir)

	// 测试正常保存
	err := suite.colonyRepo.SaveConfigFile(context.Background(), fileReader, testPath, true)
	suite.NoError(err, "保存配置文件应该成功")

	// 验证文件内容
	savedContent, err := os.ReadFile(testPath)
	suite.NoError(err, "读取保存的配置文件应该成功")
	suite.Equal(testContent, string(savedContent), "保存的配置文件内容应该正确")

	// 测试文件已存在且不允许覆盖的情况
	fileReader2 := bytes.NewReader([]byte("new content"))
	err = suite.colonyRepo.SaveConfigFile(context.Background(), fileReader2, testPath, false)
	suite.Error(err, "当文件已存在且不允许覆盖时应该返回错误")

	// 测试文件已存在但允许覆盖的情况
	fileReader3 := bytes.NewReader([]byte("new content"))
	err = suite.colonyRepo.SaveConfigFile(context.Background(), fileReader3, testPath, true)
	suite.NoError(err, "当文件已存在但允许覆盖时应该成功")

	// 验证文件内容已更新
	savedContent2, err := os.ReadFile(testPath)
	suite.NoError(err, "读取更新后的配置文件应该成功")
	suite.Equal("new content", string(savedContent2), "更新后的配置文件内容应该正确")

	// 测试创建目录失败的情况（通过设置一个不可写的目录权限）
	testDir = "/tmp/oes-test-error"
	testPath = filepath.Join(testDir, "subdir", "config.json")

	// 确保测试目录不存在
	suite.cleanupTestDir(testDir)
	defer suite.cleanupTestDir(testDir)

	// 创建一个只读的测试目录
	err = os.Mkdir(testDir, 0o555) // 只读权限
	suite.NoError(err, "创建只读目录应该成功")

	// 尝试在只读目录下创建子目录并保存文件，应该失败
	fileReader = bytes.NewReader([]byte("test content"))
	err = suite.colonyRepo.SaveConfigFile(context.Background(), fileReader, testPath, true)
	suite.Error(err, "在只读目录下保存文件应该失败")

	// 测试写入文件失败的情况（通过模拟一个错误的 reader）
	testDir3 := "/tmp/oes-test-write-error"
	testPath3 := filepath.Join(testDir3, "config.json")
	suite.cleanupTestDir(testDir3)
	defer suite.cleanupTestDir(testDir3)

	// 创建一个模拟的 reader，在读取时返回错误
	errorReader := &errorReader{}
	err = suite.colonyRepo.SaveConfigFile(context.Background(), errorReader, testPath3, true)
	suite.Error(err, "当写入文件失败时应该返回错误")

	// 测试创建目录权限失败的边界情况
	testPath4 := "/root/oes-test-config.yaml" // 通常只有root用户可以写入/root目录
	_ = suite.colonyRepo.SaveConfigFile(context.Background(), bytes.NewReader([]byte("test content")), testPath4, true)
	// 我们不强制断言错误，只确保函数能够执行完成
	// 因为在不同的系统和权限设置下，结果可能不同
}

func (suite *OesColonyTestSuite) TestRemoveConfigFile() {
	// 测试正常删除存在的文件
	testDir := "/tmp/oes-test"
	testPath := filepath.Join(testDir, "config.json")

	// 确保测试目录不存在
	suite.cleanupTestDir(testDir)
	defer suite.cleanupTestDir(testDir)

	// 创建测试文件
	err := suite.createTestFile(testPath, "test content")
	suite.NoError(err, "创建测试文件应该成功")

	// 测试正常删除
	err = suite.colonyRepo.RemoveConfigFile(context.Background(), testPath)
	suite.NoError(err, "删除配置文件应该成功")

	// 验证文件已删除
	_, err = os.Stat(testPath)
	suite.True(os.IsNotExist(err), "配置文件应该已被删除")

	// 测试删除不存在的文件
	err = suite.colonyRepo.RemoveConfigFile(context.Background(), testPath)
	suite.NoError(err, "删除不存在的配置文件应该成功（无操作）")

	// 测试删除文件失败的情况（通过设置文件为只读）
	testDir2 := "/tmp/oes-test-remove-error"
	testPath2 := filepath.Join(testDir2, "config.json")
	suite.cleanupTestDir(testDir2)
	defer suite.cleanupTestDir(testDir2)

	// 创建测试目录和文件
	err = os.MkdirAll(testDir2, 0o755)
	suite.NoError(err, "创建测试目录应该成功")
	err = os.WriteFile(testPath2, []byte("test content"), 0o444) // 只读权限
	suite.NoError(err, "创建只读文件应该成功")

	err = suite.colonyRepo.RemoveConfigFile(context.Background(), testPath2)
	suite.NoError(err, "删除只读文件时文件权限不足应该成功")

	// 测试检查文件失败的边界情况
	testPath3 := "/root/non-existent-file.txt" // 通常只有root用户可以访问/root目录
	_ = suite.colonyRepo.RemoveConfigFile(context.Background(), testPath3)
	// 我们不强制断言错误，只确保函数能够执行完成
	// 因为在不同的系统和权限设置下，结果可能不同
}

func (suite *OesColonyTestSuite) TestNewOesColonyRepo() {
	slowThreshold := test.NewTestDBSlowThreshold()
	repo := NewOesColonyRepo(suite.log, suite.gormDB, suite.timeouts, slowThreshold)
	suite.NotNil(repo, "NewOesColonyRepo 应该返回非空实例")
	suite.Equal(suite.log, repo.log, "日志实例应该正确设置")
	suite.Equal(suite.gormDB, repo.gormDB, "数据库实例应该正确设置")
	suite.Equal(suite.timeouts, repo.timeouts, "超时设置应该正确设置")
	suite.Equal(slowThreshold, repo.slowThreshold, "慢查询阈值设置应该正确设置")
}

func TestOesColonyTestSuite(t *testing.T) {
	pts := &OesColonyTestSuite{}
	suite.Run(t, pts)
}
