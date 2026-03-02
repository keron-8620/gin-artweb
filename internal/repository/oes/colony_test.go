package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	monmodel "gin-artweb/internal/model/mon"
	oesmodel "gin-artweb/internal/model/oes"
	resomodel "gin-artweb/internal/model/resource"
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
}

func (suite *OesColonyTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(
		&resomodel.HostModel{},
		&monmodel.MonNodeModel{},
		&resomodel.PackageModel{},
		&oesmodel.OesColonyModel{},
	)

	// 创建测试数据：主机
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

	// 创建测试数据：Mon节点
	monNodeModel := &monmodel.MonNodeModel{
		Name:        "test-mon-node",
		DeployPath:  "/opt/mon",
		OutportPath: "/opt/mon/outport",
		JavaHome:    "/usr/lib/jvm/java-11-openjdk-amd64",
		URL:         "http://localhost:8080/mon",
		HostID:      1,
	}
	db.Create(monNodeModel)

	// 创建测试数据：程序包1
	packageModel1 := &resomodel.PackageModel{
		Label:           "test-oes",
		StorageFilename: "test-oes.tar.gz",
		OriginFilename:  "test-oes.tar.gz",
		Version:         "1.0.0",
	}
	db.Create(packageModel1)

	// 创建测试数据：程序包2 (xcounter)
	packageModel2 := &resomodel.PackageModel{
		Label:           "test-xcounter",
		StorageFilename: "test-xcounter.tar.gz",
		OriginFilename:  "test-xcounter.tar.gz",
		Version:         "1.0.0",
	}
	db.Create(packageModel2)

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.colonyRepo = &OesColonyRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *OesColonyTestSuite) TestCreateModel() {
	// 测试正常创建
	cm := CreateTestOesColonyModel()
	err := suite.colonyRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesColony应该成功")
	suite.NotZero(cm.ID, "OesColony ID应该不为零")

	// 测试边界情况：创建空模型
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

	// 测试边界情况：更新数据为空
	err = suite.colonyRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", cm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况：更新不存在的OesColony
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

	// 测试边界情况：删除不存在的OesColony
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

	// 测试边界情况：查询不存在的OesColony
	fm, err = suite.colonyRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的OesColony应该返回错误")
	suite.Nil(fm, "查询不存在的OesColony应该返回nil")

	// 测试边界情况：使用预加载
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
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	count, models, err := suite.colonyRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询OesColony列表应该成功")
	suite.Greater(count, int64(0), "OesColony列表数量应该大于0")
	suite.NotNil(models, "OesColony列表应该不为nil")
	suite.Greater(len(*models), 0, "OesColony列表长度应该大于0")

	// 测试边界情况：空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"colony_num": "99"},
	}
	count2, models2, err := suite.colonyRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的OesColony列表应该成功")
	suite.Equal(int64(0), count2, "不存在的OesColony列表数量应该为0")
	suite.NotNil(models2, "不存在的OesColony列表应该不为nil")
	suite.Len(*models2, 0, "不存在的OesColony列表长度应该为0")
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

func TestOesColonyTestSuite(t *testing.T) {
	pts := &OesColonyTestSuite{}
	suite.Run(t, pts)
}
