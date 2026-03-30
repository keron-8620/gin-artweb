package oes

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"

	monmodel "gin-artweb/internal/model/mon"
	oesmodel "gin-artweb/internal/model/oes"
	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestOesNodeModel(oesColonyID, hostID uint32) *oesmodel.OesNodeModel {
	return &oesmodel.OesNodeModel{
		NodeRole:    "master",
		IsEnable:    true,
		OesColonyID: oesColonyID,
		HostID:      hostID,
	}
}

type OesNodeTestSuite struct {
	suite.Suite
	nodeRepo *OesNodeRepo
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func (suite *OesNodeTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&resomodel.HostModel{}, &monmodel.MonNodeModel{}, &resomodel.PackageModel{}, &oesmodel.OesColonyModel{}, &oesmodel.OesNodeModel{})

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

	// 创建测试数据：Oes集群
	esColonyModel := &oesmodel.OesColonyModel{
		SystemType:    "STK",
		ColonyNum:     "01",
		ExtractedName: "test-oes-colony",
		IsEnable:      true,
		PackageID:     1,
		XCounterID:    2,
		MonNodeID:     1,
	}
	db.Create(esColonyModel)

	suite.timeouts = test.NewTestDBTimeouts()
	suite.log = test.NewTestZapLogger()
	suite.gormDB = db
	suite.nodeRepo = &OesNodeRepo{
		log:      suite.log,
		gormDB:   suite.gormDB,
		timeouts: suite.timeouts,
	}
}

func (suite *OesNodeTestSuite) TestCreateModel() {
	// 测试正常创建
	cm := CreateTestOesNodeModel(1, 1) // 使用已创建的集群和主机ID
	err := suite.nodeRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesNode应该成功")
	suite.NotZero(cm.ID, "OesNode ID应该不为零")

	// 测试边界情况：创建空模型
	err = suite.nodeRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空OesNode模型应该返回错误")
}

func (suite *OesNodeTestSuite) TestUpdateModel() {
	// 创建测试数据
	cm := CreateTestOesNodeModel(1, 1)
	err := suite.nodeRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesNode用于更新测试应该成功")

	// 测试正常更新
	updateData := map[string]any{
		"NodeRole": "follow",
		"IsEnable": false,
	}
	err = suite.nodeRepo.UpdateModel(context.Background(), updateData, "id = ?", cm.ID)
	suite.NoError(err, "更新OesNode应该成功")

	// 验证更新结果
	fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询更新后的OesNode应该成功")
	suite.Equal("follow", fm.NodeRole)
	suite.False(fm.IsEnable, "IsEnable应该被更新为false")

	// 测试边界情况：更新数据为空
	err = suite.nodeRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", cm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况：更新不存在的OesNode
	err = suite.nodeRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的OesNode应该成功（无操作）")
}

func (suite *OesNodeTestSuite) TestDeleteModel() {
	// 创建测试数据
	cm := CreateTestOesNodeModel(1, 1)
	err := suite.nodeRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesNode用于删除测试应该成功")

	// 测试正常删除
	err = suite.nodeRepo.DeleteModel(context.Background(), "id = ?", cm.ID)
	suite.NoError(err, "删除OesNode应该成功")

	// 验证删除结果
	fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.Error(err, "查询已删除的OesNode应该返回错误")
	suite.Nil(fm, "已删除的OesNode应该为nil")

	// 测试边界情况：删除不存在的OesNode
	err = suite.nodeRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的OesNode应该成功（无操作）")
}

func (suite *OesNodeTestSuite) TestGetModel() {
	// 创建测试数据
	cm := CreateTestOesNodeModel(1, 1)
	err := suite.nodeRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesNode用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询OesNode应该成功")
	suite.Equal(cm.ID, fm.ID)
	suite.Equal(cm.NodeRole, fm.NodeRole)
	suite.Equal(cm.IsEnable, fm.IsEnable)

	// 测试边界情况：查询不存在的OesNode
	fm, err = suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的OesNode应该返回错误")
	suite.Nil(fm, "查询不存在的OesNode应该返回nil")

	// 测试边界情况：使用预加载
	fm, err = suite.nodeRepo.GetModel(context.Background(), []string{"OesColony", "Host"}, "id = ?", cm.ID)
	suite.NoError(err, "使用预加载查询OesNode应该成功")
	suite.Equal(cm.ID, fm.ID)
}

func (suite *OesNodeTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		cm := CreateTestOesNodeModel(1, 1)
		err := suite.nodeRepo.CreateModel(context.Background(), cm)
		suite.NoError(err, "创建OesNode用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Limit:   10,
		Offset:  0,
	}
	models, err := suite.nodeRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询OesNode列表应该成功")
	suite.Greater(int64(len(models)), int64(0), "OesNode列表数量应该大于0")
	suite.NotNil(models, "OesNode列表应该不为nil")
	suite.Greater(len(models), 0, "OesNode列表长度应该大于0")

	// 测试边界情况：空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"node_role": "non-existent"},
	}
	models2, err := suite.nodeRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的OesNode列表应该成功")
	suite.Equal(int64(0), int64(len(models2)), "不存在的OesNode列表数量应该为0")
	suite.NotNil(models2, "不存在的OesNode列表应该不为nil")
	suite.Len(models2, 0, "不存在的OesNode列表长度应该为0")
}

func (suite *OesNodeTestSuite) TestContextTimeout() {
	// 创建测试数据
	cm := CreateTestOesNodeModel(1, 1)
	err := suite.nodeRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建OesNode用于超时测试应该成功")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err = suite.nodeRepo.GetModel(timeoutCtx, nil, "id = ?", cm.ID)
	suite.Error(err, "上下文超时后查询OesNode应该返回错误")
}

func (suite *OesNodeTestSuite) TestCountModel() {
	// 创建测试数据
	for i := 0; i < 3; i++ {
		cm := CreateTestOesNodeModel(1, 1)
		err := suite.nodeRepo.CreateModel(context.Background(), cm)
		suite.NoError(err, "创建OesNode用于计数测试应该成功")
	}

	// 测试正常计数
	count, err := suite.nodeRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "查询OesNode总数应该成功")
	suite.Greater(count, int64(0), "OesNode总数应该大于0")

	// 测试带条件计数
	count2, err := suite.nodeRepo.CountModel(context.Background(), map[string]any{"node_role": "master"})
	suite.NoError(err, "带条件查询OesNode总数应该成功")
	suite.GreaterOrEqual(count2, int64(3), "带条件的OesNode总数应该大于等于3")

	// 测试边界情况：查询不存在的角色
	count3, err := suite.nodeRepo.CountModel(context.Background(), map[string]any{"node_role": "non-existent"})
	suite.NoError(err, "查询不存在角色的OesNode总数应该成功")
	suite.Equal(int64(0), count3, "不存在角色的OesNode总数应该为0")
}

func (suite *OesNodeTestSuite) TestNewOesNodeRepo() {
	repo := NewOesNodeRepo(suite.log, suite.gormDB, suite.timeouts)
	suite.NotNil(repo, "NewOesNodeRepo 应该返回非空实例")
	suite.Equal(suite.log, repo.log, "日志实例应该正确设置")
	suite.Equal(suite.gormDB, repo.gormDB, "数据库实例应该正确设置")
	suite.Equal(suite.timeouts, repo.timeouts, "超时设置应该正确设置")
}

func TestOesNodeTestSuite(t *testing.T) {
	pts := &OesNodeTestSuite{}
	suite.Run(t, pts)
}
