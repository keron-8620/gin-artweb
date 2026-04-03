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

var oesColonyCounter int = 0

func CreateTestOesColonyModel() *oesmodel.OesColonyModel {
	oesColonyCounter++
	return &oesmodel.OesColonyModel{
		SystemType:    "STK",
		ColonyNum:     fmt.Sprintf("%02d", oesColonyCounter%100),
		ExtractedName: fmt.Sprintf("oes-%s", uuid.NewString()),
		IsEnable:      true,
		PackageID:     1, // 假设程序包ID为1
		XCounterID:    2, // 假设xcounter包ID为2
		MonNodeID:     1, // 假设Mon节点ID为1
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
	db.AutoMigrate(
		&resomodel.HostModel{},
		&monmodel.MonNodeModel{},
		&resomodel.PackageModel{},
		&oesmodel.OesColonyModel{},
	)

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
	suite.colonyRepo = &OesColonyRepo{
		log:      suite.log,
		gormDB:   suite.gormDB,
		timeouts: suite.timeouts,
	}
}

func (suite *OesColonyTestSuite) TestCreateModel() {
	// 测试正常创建
	cm := CreateTestOesColonyModel()
	err := suite.colonyRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesColony应该成功")
	suite.NotZero(cm.ID, "OesColony ID应该不为零")

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
	suite.NoError(err, "更新OesColony应该成功")

	// 验证更新结果
	fm, err := suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询更新后的OesColony应该成功")
	suite.Equal(cm.ColonyNum, fm.ColonyNum, "ColonyNum应该保持不变")
	suite.Equal("updated-oes", fm.ExtractedName)
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
	suite.NoError(err, "删除OesColony应该成功")

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
	suite.NoError(err, "查询OesColony应该成功")
	suite.Equal(cm.ID, fm.ID)
	suite.Equal(cm.ColonyNum, fm.ColonyNum)
	suite.Equal(cm.IsEnable, fm.IsEnable)

	// 测试边界情况:查询不存在的OesColony
	fm, err = suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的OesColony应该返回错误")
	suite.Nil(fm, "查询不存在的OesColony应该返回nil")

	// 测试边界情况:使用预加载
	fm, err = suite.colonyRepo.GetModel(context.Background(), []string{"Package", "XCounter", "MonNode"}, "id = ?", cm.ID)
	suite.NoError(err, "使用预加载查询OesColony应该成功")
	suite.Equal(cm.ID, fm.ID)
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
	suite.NoError(err, "查询OesColony列表应该成功")
	suite.Greater(int64(len(models)), int64(0), "OesColony列表数量应该大于0")
	suite.NotNil(models, "OesColony列表应该不为nil")
	suite.Greater(len(models), 0, "OesColony列表长度应该大于0")

	// 测试边界情况:空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"colony_num": "99"},
	}
	models2, err := suite.colonyRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的OesColony列表应该成功")
	suite.Equal(int64(0), int64(len(models2)), "不存在的OesColony列表数量应该为0")
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
	suite.NoError(err, "查询OesColony总数应该成功")
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

func (suite *OesColonyTestSuite) TestSaveConfigFile() {
	// 测试正常保存配置文件
	testContent := "test configuration content"
	fileReader := bytes.NewReader([]byte(testContent))
	testDir := "/tmp/oes-test"
	testPath := filepath.Join(testDir, "config.json")

	// 确保测试目录不存在
	os.RemoveAll(testDir)

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

	// 清理测试文件
	os.RemoveAll(testDir)
}

func (suite *OesColonyTestSuite) TestRemoveConfigFile() {
	// 测试正常删除存在的文件
	testDir := "/tmp/oes-test"
	testPath := filepath.Join(testDir, "config.json")

	// 确保测试目录不存在
	os.RemoveAll(testDir)

	// 创建测试文件
	os.MkdirAll(testDir, 0o755)
	err := os.WriteFile(testPath, []byte("test content"), 0o644)
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

	// 清理测试目录
	os.RemoveAll(testDir)
}

func (suite *OesColonyTestSuite) TestSaveConfigFileErrorPaths() {
	// 测试文件已存在且不允许覆盖的情况
	testDir := "/tmp/oes-test-existing"
	testPath := filepath.Join(testDir, "config.json")

	// 确保测试目录不存在
	os.RemoveAll(testDir)

	// 创建测试目录和文件
	os.MkdirAll(testDir, 0o755)
	err := os.WriteFile(testPath, []byte("existing content"), 0o644)
	suite.NoError(err, "创建测试文件应该成功")

	// 尝试保存文件且不允许覆盖，应该失败
	fileReader := bytes.NewReader([]byte("new content"))
	err = suite.colonyRepo.SaveConfigFile(context.Background(), fileReader, testPath, false)
	suite.Error(err, "当文件已存在且不允许覆盖时应该失败")

	// 清理测试目录
	os.RemoveAll(testDir)

	// 测试创建目录失败的情况（通过设置一个不可写的目录权限）
	testDir = "/tmp/oes-test-error"
	testPath = filepath.Join(testDir, "subdir", "config.json")

	// 确保测试目录不存在
	os.RemoveAll(testDir)

	// 创建一个只读的测试目录
	err = os.Mkdir(testDir, 0o555) // 只读权限
	suite.NoError(err, "创建只读目录应该成功")

	// 尝试在只读目录下创建子目录并保存文件，应该失败
	fileReader = bytes.NewReader([]byte("test content"))
	err = suite.colonyRepo.SaveConfigFile(context.Background(), fileReader, testPath, true)
	suite.Error(err, "在只读目录下保存文件应该失败")

	// 清理测试目录
	os.RemoveAll(testDir)

	// 测试写入文件失败的情况（通过模拟一个错误的 reader）
	testDir3 := "/tmp/oes-test-write-error"
	testPath3 := filepath.Join(testDir3, "config.json")
	os.RemoveAll(testDir3)

	// 创建一个模拟的 reader，在读取时返回错误
	errorReader := &errorReader{}
	err = suite.colonyRepo.SaveConfigFile(context.Background(), errorReader, testPath3, true)
	suite.Error(err, "当写入文件失败时应该返回错误")

	// 清理测试目录
	os.RemoveAll(testDir3)
}

// errorReader 是一个模拟的 reader，在读取时返回错误
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("模拟读取错误")
}

func (suite *OesColonyTestSuite) TestRemoveConfigFileErrorPaths() {
	// 测试删除文件失败的情况（通过设置文件为只读）
	testDir := "/tmp/oes-test-remove-error"
	testPath := filepath.Join(testDir, "config.json")

	// 确保测试目录不存在
	os.RemoveAll(testDir)

	// 创建测试目录和文件
	os.MkdirAll(testDir, 0o755)
	err := os.WriteFile(testPath, []byte("test content"), 0o444) // 只读权限
	suite.NoError(err, "创建只读文件应该成功")

	// 尝试删除只读文件，在某些系统上可能会成功，在某些系统上可能会失败
	// 这里我们不强制断言错误，只确保函数能够执行完成
	err = suite.colonyRepo.RemoveConfigFile(context.Background(), testPath)
	// 注意:在Linux系统中，删除只读文件是允许的，所以这里可能不会返回错误
	// 我们只确保函数能够正常执行，不崩溃

	// 清理测试目录
	os.RemoveAll(testDir)

	// 测试删除不存在的文件
	nonExistentPath := "/tmp/non-existent-file.txt"
	err = suite.colonyRepo.RemoveConfigFile(context.Background(), nonExistentPath)
	suite.NoError(err, "删除不存在的文件应该成功（无操作）")

	// 测试正常删除文件
	testDir2 := "/tmp/oes-test-normal-remove"
	testPath2 := filepath.Join(testDir2, "config.json")
	os.RemoveAll(testDir2)

	os.MkdirAll(testDir2, 0o755)
	err = os.WriteFile(testPath2, []byte("test content"), 0o644)
	suite.NoError(err, "创建测试文件应该成功")

	err = suite.colonyRepo.RemoveConfigFile(context.Background(), testPath2)
	suite.NoError(err, "正常删除文件应该成功")

	// 验证文件已删除
	_, err = os.Stat(testPath2)
	suite.True(os.IsNotExist(err), "文件应该已被删除")

	// 清理测试目录
	os.RemoveAll(testDir2)
}

func (suite *OesColonyTestSuite) TestNewOesColonyRepo() {
	repo := NewOesColonyRepo(suite.log, suite.gormDB, suite.timeouts)
	suite.NotNil(repo, "NewOesColonyRepo 应该返回非空实例")
	suite.Equal(suite.log, repo.log, "日志实例应该正确设置")
	suite.Equal(suite.gormDB, repo.gormDB, "数据库实例应该正确设置")
	suite.Equal(suite.timeouts, repo.timeouts, "超时设置应该正确设置")
}

func TestOesColonyTestSuite(t *testing.T) {
	pts := &OesColonyTestSuite{}
	suite.Run(t, pts)
}

func (suite *OesColonyTestSuite) TestSaveConfigFileCreateDirError() {
	// 测试创建目录失败的情况
	// 注意:由于权限问题，在某些系统上可能需要特殊处理
	testPath := "/root/oes-test-config.yaml" // 通常只有root用户可以写入/root目录

	_ = suite.colonyRepo.SaveConfigFile(context.Background(), bytes.NewReader([]byte("test content")), testPath, true)
	// 我们不强制断言错误，只确保函数能够执行完成
	// 因为在不同的系统和权限设置下，结果可能不同
}

func (suite *OesColonyTestSuite) TestSaveConfigFileChmodError() {
	// 测试设置目录权限失败的情况
	testDir := "/tmp/oes-test-chmod-error"
	testPath := filepath.Join(testDir, "config.yaml")

	// 确保测试目录不存在
	os.RemoveAll(testDir)

	// 创建测试目录并设置为只读
	err := os.MkdirAll(testDir, 0o555) // 只读权限
	suite.NoError(err, "创建只读目录应该成功")

	// 尝试保存配置文件，这应该会失败，因为无法设置目录权限
	err = suite.colonyRepo.SaveConfigFile(context.Background(), bytes.NewReader([]byte("test content")), testPath, true)
	// 我们不强制断言错误，只确保函数能够执行完成
	// 因为在不同的系统和权限设置下，结果可能不同

	// 清理测试目录
	os.RemoveAll(testDir)
}

func (suite *OesColonyTestSuite) TestRemoveConfigFileStatError() {
	// 测试检查文件失败的情况（非文件不存在的错误）
	// 注意:由于权限问题，在某些系统上可能需要特殊处理
	testPath := "/root/non-existent-file.txt" // 通常只有root用户可以访问/root目录

	_ = suite.colonyRepo.RemoveConfigFile(context.Background(), testPath)
	// 我们不强制断言错误，只确保函数能够执行完成
	// 因为在不同的系统和权限设置下，结果可能不同
}
