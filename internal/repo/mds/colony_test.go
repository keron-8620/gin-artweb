package mds

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	mdsmodel "gin-artweb/internal/model/mds"
	monmodel "gin-artweb/internal/model/mon"
	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

var colonyCounter int = 0

func CreateTestMdsColonyModel(overrides ...map[string]any) *mdsmodel.MdsColonyModel {
	colonyCounter++
	model := &mdsmodel.MdsColonyModel{
		ColonyNum:     fmt.Sprintf("%02d", colonyCounter%100),
		ExtractedName: fmt.Sprintf("mds-%s", uuid.NewString()),
		IsEnable:      true,
		PackageID:     1, // 假设程序包ID为1
		MonNodeID:     1, // 假设Mon节点ID为1
	}

	// 应用覆盖值
	if len(overrides) > 0 {
		for key, value := range overrides[0] {
			switch key {
			case "ColonyNum":
				if val, ok := value.(string); ok {
					model.ColonyNum = val
				}
			case "ExtractedName":
				if val, ok := value.(string); ok {
					model.ExtractedName = val
				}
			case "IsEnable":
				if val, ok := value.(bool); ok {
					model.IsEnable = val
				}
			case "PackageID":
				if val, ok := value.(uint32); ok {
					model.PackageID = val
				}
			case "MonNodeID":
				if val, ok := value.(uint32); ok {
					model.MonNodeID = val
				}
			}
		}
	}

	return model
}

type MdsColonyTestSuite struct {
	suite.Suite
	colonyRepo *MdsColonyRepo
	testModel  *mdsmodel.MdsColonyModel
}

func (suite *MdsColonyTestSuite) SetupTest() {
	// 在每个测试前创建测试数据
	suite.testModel = CreateTestMdsColonyModel()
	err := suite.colonyRepo.CreateModel(context.Background(), suite.testModel)
	suite.NoError(err, "创建测试MdsColony应该成功")
	suite.NotZero(suite.testModel.ID, "MdsColony ID应该不为零")
}

func (suite *MdsColonyTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&resomodel.HostModel{}, &monmodel.MonNodeModel{}, &resomodel.PackageModel{}, &mdsmodel.MdsColonyModel{}); err != nil {
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

	// 创建测试数据:程序包
	packageModel := &resomodel.PackageModel{
		Label:           "test",
		StorageFilename: "test-package.tar.gz",
		OriginFilename:  "test-package.tar.gz",
		Version:         "1.0.0",
	}
	db.Create(packageModel)

	dbTimeout := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()
	suite.colonyRepo = &MdsColonyRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: slowThreshold,
	}
}

func (suite *MdsColonyTestSuite) TestCreateModel() {
	// 测试边界情况:创建空模型
	err := suite.colonyRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空MdsColony模型应该返回错误")
}

func (suite *MdsColonyTestSuite) TestUpdateModel() {
	// 测试正常更新 - 不更新ColonyNum以避免唯一约束冲突
	updateData := map[string]any{
		"ExtractedName": "updated-mds",
		"IsEnable":      false,
	}
	err := suite.colonyRepo.UpdateModel(context.Background(), updateData, "id = ?", suite.testModel.ID)
	suite.NoError(err, "更新MdsColony应该成功")

	// 验证更新结果
	fm, err := suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", suite.testModel.ID)
	suite.NoError(err, "查询更新后的MdsColony应该成功")
	suite.Equal(suite.testModel.ColonyNum, fm.ColonyNum, "ColonyNum应该保持不变")
	suite.Equal("updated-mds", fm.ExtractedName)
	suite.False(fm.IsEnable, "IsEnable应该被更新为false")

	// 测试边界情况:更新数据为空
	err = suite.colonyRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", suite.testModel.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况:更新不存在的MdsColony
	err = suite.colonyRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的MdsColony应该成功（无操作）")
}

func (suite *MdsColonyTestSuite) TestDeleteModel() {
	// 测试正常删除
	err := suite.colonyRepo.DeleteModel(context.Background(), "id = ?", suite.testModel.ID)
	suite.NoError(err, "删除MdsColony应该成功")

	// 验证删除结果
	fm, err := suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", suite.testModel.ID)
	suite.Error(err, "查询已删除的MdsColony应该返回错误")
	suite.Nil(fm, "已删除的MdsColony应该为nil")

	// 测试边界情况:删除不存在的MdsColony
	err = suite.colonyRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的MdsColony应该成功（无操作）")
}

func (suite *MdsColonyTestSuite) TestGetModel() {
	// 测试正常查询
	fm, err := suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", suite.testModel.ID)
	suite.NoError(err, "查询MdsColony应该成功")
	suite.Equal(suite.testModel.ID, fm.ID)
	suite.Equal(suite.testModel.ColonyNum, fm.ColonyNum)
	suite.Equal(suite.testModel.IsEnable, fm.IsEnable)

	// 测试边界情况:查询不存在的MdsColony
	fm, err = suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的MdsColony应该返回错误")
	suite.Nil(fm, "查询不存在的MdsColony应该返回nil")

	// 测试边界情况:使用预加载
	fm, err = suite.colonyRepo.GetModel(context.Background(), []string{"Package", "MonNode"}, "id = ?", suite.testModel.ID)
	suite.NoError(err, "使用预加载查询MdsColony应该成功")
	suite.Equal(suite.testModel.ID, fm.ID)
}

func (suite *MdsColonyTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		cm := CreateTestMdsColonyModel()
		err := suite.colonyRepo.CreateModel(context.Background(), cm)
		suite.NoError(err, "创建MdsColony用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Limit:   10,
		Offset:  0,
	}
	models, err := suite.colonyRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询MdsColony列表应该成功")
	suite.Greater(int64(len(models)), int64(0), "MdsColony列表数量应该大于0")
	suite.NotNil(models, "MdsColony列表应该不为nil")

	// 测试边界情况:空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"colony_num": "99"},
	}
	models2, err := suite.colonyRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的MdsColony列表应该成功")
	suite.Len(models2, 0, "不存在的MdsColony列表长度应该为0")
}

func (suite *MdsColonyTestSuite) TestContextTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err := suite.colonyRepo.GetModel(timeoutCtx, nil, "id = ?", suite.testModel.ID)
	suite.Error(err, "上下文超时后查询MdsColony应该返回错误")
}

func (suite *MdsColonyTestSuite) TestCountModel() {
	// 测试正常计数
	count, err := suite.colonyRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "计数MdsColony应该成功")
	suite.Greater(count, int64(0), "MdsColony计数应该大于0")

	// 测试带条件计数
	count, err = suite.colonyRepo.CountModel(context.Background(), map[string]any{"is_enable": true})
	suite.NoError(err, "带条件计数MdsColony应该成功")
	suite.GreaterOrEqual(count, int64(0), "MdsColony计数应该大于等于0")

	// 测试计数不存在的条件
	count, err = suite.colonyRepo.CountModel(context.Background(), map[string]any{"colony_num": "9999"})
	suite.NoError(err, "计数不存在的MdsColony应该成功")
	suite.Equal(int64(0), count, "不存在的MdsColony计数应该为0")
}

func (suite *MdsColonyTestSuite) TestSaveConfigFile() {
	// 测试SaveConfigFile函数
	testContent := "test config content"
	testReader := strings.NewReader(testContent)
	testPath := "/tmp/test-mds-config.txt"

	// 清理测试文件
	defer func() {
		os.Remove(testPath)
	}()

	// 测试正常保存
	err := suite.colonyRepo.SaveConfigFile(context.Background(), testReader, testPath, true)
	suite.NoError(err, "保存配置文件应该成功")

	// 验证文件内容
	content, err := os.ReadFile(testPath)
	suite.NoError(err, "读取保存的配置文件应该成功")
	suite.Equal(testContent, string(content), "配置文件内容应该正确")

	// 测试覆盖保存
	testContent2 := "updated config content"
	testReader2 := strings.NewReader(testContent2)
	err = suite.colonyRepo.SaveConfigFile(context.Background(), testReader2, testPath, true)
	suite.NoError(err, "覆盖保存配置文件应该成功")

	// 验证更新后的内容
	content, err = os.ReadFile(testPath)
	suite.NoError(err, "读取更新后的配置文件应该成功")
	suite.Equal(testContent2, string(content), "更新后的配置文件内容应该正确")

	// 测试不允许覆盖已存在的文件
	testReader3 := strings.NewReader("new content")
	err = suite.colonyRepo.SaveConfigFile(context.Background(), testReader3, testPath, false)
	suite.Error(err, "不允许覆盖时应该返回错误")
}

func (suite *MdsColonyTestSuite) TestRemoveConfigFile() {
	// 测试RemoveConfigFile函数
	testPath := "/tmp/test-mds-config-remove.txt"

	// 创建测试文件
	err := os.WriteFile(testPath, []byte("test content"), 0644)
	suite.NoError(err, "创建测试文件应该成功")

	// 测试正常删除
	err = suite.colonyRepo.RemoveConfigFile(context.Background(), testPath)
	suite.NoError(err, "删除配置文件应该成功")

	// 验证文件已删除
	_, err = os.Stat(testPath)
	suite.True(os.IsNotExist(err), "文件应该被成功删除")

	// 测试删除不存在的文件
	err = suite.colonyRepo.RemoveConfigFile(context.Background(), testPath)
	suite.NoError(err, "删除不存在的配置文件应该成功")
}

func TestMdsColonyTestSuite(t *testing.T) {
	pts := &MdsColonyTestSuite{}
	suite.Run(t, pts)
}
