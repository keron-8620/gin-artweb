package mds

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	jobmodel "gin-artweb/internal/model/job"
	mdsmodel "gin-artweb/internal/model/mds"
	monmodel "gin-artweb/internal/model/mon"
	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestMdsCronModel() *mdsmodel.MdsCronModel {
	return &mdsmodel.MdsCronModel{
		MdsColonyID: 1, // 假设mds集群ID为1
		ScheduleID:  1, // 假设计划任务ID为1
	}
}

type MdsCronTestSuite struct {
	suite.Suite
	cronRepo *MdsCronRepo
}

func (suite *MdsCronTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&resomodel.HostModel{}, &monmodel.MonNodeModel{}, &resomodel.PackageModel{}, &mdsmodel.MdsColonyModel{}, &jobmodel.ScriptModel{}, &mdsmodel.MdsCronModel{})

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

	// 创建测试数据：程序包
	packageModel := &resomodel.PackageModel{
		Label:           "test",
		StorageFilename: "test-package.tar.gz",
		OriginFilename:  "test-package.tar.gz",
		Version:         "1.0.0",
	}
	db.Create(packageModel)

	// 创建测试数据：Mds集群
	colonyModel := &mdsmodel.MdsColonyModel{
		ColonyNum:     "01",
		ExtractedName: "test-mds-colony",
		IsEnable:      true,
		PackageID:     1,
		MonNodeID:     1,
	}
	db.Create(colonyModel)

	// 创建测试数据：脚本
	scriptModel := &jobmodel.ScriptModel{
		Name:      "test-script",
		Descr:     "Test script",
		Project:   "test",
		Label:     "test",
		Language:  "bash",
		IsBuiltin: false,
		Username:  "test",
	}
	db.Create(scriptModel)

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.cronRepo = &MdsCronRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *MdsCronTestSuite) TestCreateModel() {
	// 测试正常创建
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron应该成功")
	suite.NotZero(cm.ID, "MdsCron ID应该不为零")

	// 测试边界情况：创建空模型
	err = suite.cronRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空MdsCron模型应该返回错误")
}

func (suite *MdsCronTestSuite) TestUpdateModel() {
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron用于更新测试应该成功")

	updateData := map[string]any{
		"mds_colony_id": 1,
	}
	err = suite.cronRepo.UpdateModel(context.Background(), updateData, "id = ?", cm.ID)
	suite.NoError(err, "更新MdsCron应该成功")

	fm, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询更新后的MdsCron应该成功")
	suite.Equal(uint32(1), fm.MdsColonyID, "MdsColonyID应该被更新")

	err = suite.cronRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", cm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	err = suite.cronRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的MdsCron应该成功（无操作）")
}

func (suite *MdsCronTestSuite) TestDeleteModel() {
	// 创建测试数据
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron用于删除测试应该成功")

	// 测试正常删除
	err = suite.cronRepo.DeleteModel(context.Background(), "id = ?", cm.ID)
	suite.NoError(err, "删除MdsCron应该成功")

	// 验证删除结果
	fm, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.Error(err, "查询已删除的MdsCron应该返回错误")
	suite.Nil(fm, "已删除的MdsCron应该为nil")

	// 测试边界情况：删除不存在的MdsCron
	err = suite.cronRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的MdsCron应该成功（无操作）")
}

func (suite *MdsCronTestSuite) TestGetModel() {
	// 创建测试数据
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.cronRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询MdsCron应该成功")
	suite.Equal(cm.ID, fm.ID)
	suite.Equal(cm.MdsColonyID, fm.MdsColonyID)
	suite.Equal(cm.ScheduleID, fm.ScheduleID)

	// 测试边界情况：查询不存在的MdsCron
	fm, err = suite.cronRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的MdsCron应该返回错误")
	suite.Nil(fm, "查询不存在的MdsCron应该返回nil")

	// 测试边界情况：使用预加载
	fm, err = suite.cronRepo.GetModel(context.Background(), []string{"MdsColony", "Schedule"}, "id = ?", cm.ID)
	suite.NoError(err, "使用预加载查询MdsCron应该成功")
	suite.Equal(cm.ID, fm.ID)
}

func (suite *MdsCronTestSuite) TestListModel() {
	for i := 0; i < 3; i++ {
		cm := CreateTestMdsCronModel()
		err := suite.cronRepo.CreateModel(context.Background(), cm)
		suite.NoError(err, "创建MdsCron用于列表测试应该成功")
	}

	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Limit:   10,
		Offset:  0,
	}
	models, err := suite.cronRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询MdsCron列表应该成功")
	suite.Greater(int64(len(models)), int64(0), "MdsCron列表数量应该大于0")
	suite.NotNil(models, "MdsCron列表应该不为nil")

	qp2 := database.QueryParams{
		Query: map[string]any{"mds_colony_id": 999999},
	}
	models2, err := suite.cronRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的MdsCron列表应该成功")
	suite.Equal(int64(0), int64(len(models2)), "不存在的MdsCron列表数量应该为0")
	suite.NotNil(models2, "不存在的MdsCron列表应该不为nil")
	suite.Len(models2, 0, "不存在的MdsCron列表长度应该为0")
}

func (suite *MdsCronTestSuite) TestCountModel() {
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron用于计数测试应该成功")

	count, err := suite.cronRepo.CountModel(context.Background(), map[string]any{"mds_colony_id": 1})
	suite.NoError(err, "计数MdsCron应该成功")
	suite.Greater(count, int64(0), "MdsCron计数应该大于0")

	count, err = suite.cronRepo.CountModel(context.Background(), map[string]any{"mds_colony_id": 999999})
	suite.NoError(err, "计数不存在的MdsCron应该成功")
	suite.Equal(int64(0), count, "不存在的MdsCron计数应该为0")
}

func (suite *MdsCronTestSuite) TestContextTimeout() {
	// 创建测试数据
	cm := CreateTestMdsCronModel()
	err := suite.cronRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsCron用于超时测试应该成功")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err = suite.cronRepo.GetModel(timeoutCtx, nil, "id = ?", cm.ID)
	suite.Error(err, "上下文超时后查询MdsCron应该返回错误")
}

func TestMdsCronTestSuite(t *testing.T) {
	pts := &MdsCronTestSuite{}
	suite.Run(t, pts)
}
