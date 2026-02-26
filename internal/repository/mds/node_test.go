package mds

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	mdsmodel "gin-artweb/internal/model/mds"
	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestMdsNodeModel(mdsColonyID, hostID uint32) *mdsmodel.MdsNodeModel {
	return &mdsmodel.MdsNodeModel{
		NodeRole:    "master",
		IsEnable:    true,
		MdsColonyID: mdsColonyID,
		HostID:      hostID,
	}
}

type MdsNodeTestSuite struct {
	suite.Suite
	nodeRepo *MdsNodeRepo
}

func (suite *MdsNodeTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&resomodel.HostModel{}, &mdsmodel.MdsColonyModel{}, &mdsmodel.MdsNodeModel{})

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

	// 创建测试数据：Mds集群
	colonyModel := &mdsmodel.MdsColonyModel{
		ColonyNum:     "01",
		ExtractedName: "test-mds-colony",
		IsEnable:      true,
		PackageID:     1, // 假设程序包ID为1
		MonNodeID:     1, // 假设Mon节点ID为1
	}
	db.Create(colonyModel)

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.nodeRepo = &MdsNodeRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *MdsNodeTestSuite) TestCreateModel() {
	// 测试正常创建
	cm := CreateTestMdsNodeModel(1, 1) // 使用已创建的集群和主机ID
	err := suite.nodeRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsNode应该成功")
	suite.NotZero(cm.ID, "MdsNode ID应该不为零")

	// 测试边界情况：创建空模型
	err = suite.nodeRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空MdsNode模型应该返回错误")
}

func (suite *MdsNodeTestSuite) TestUpdateModel() {
	// 创建测试数据
	cm := CreateTestMdsNodeModel(1, 1)
	err := suite.nodeRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsNode用于更新测试应该成功")

	// 测试正常更新
	updateData := map[string]any{
		"NodeRole": "follow",
		"IsEnable": false,
	}
	err = suite.nodeRepo.UpdateModel(context.Background(), updateData, "id = ?", cm.ID)
	suite.NoError(err, "更新MdsNode应该成功")

	// 验证更新结果
	fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询更新后的MdsNode应该成功")
	suite.Equal("follow", fm.NodeRole)
	suite.False(fm.IsEnable, "IsEnable应该被更新为false")

	// 测试边界情况：更新数据为空
	err = suite.nodeRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", cm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况：更新不存在的MdsNode
	err = suite.nodeRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的MdsNode应该成功（无操作）")
}

func (suite *MdsNodeTestSuite) TestDeleteModel() {
	// 创建测试数据
	cm := CreateTestMdsNodeModel(1, 1)
	err := suite.nodeRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsNode用于删除测试应该成功")

	// 测试正常删除
	err = suite.nodeRepo.DeleteModel(context.Background(), "id = ?", cm.ID)
	suite.NoError(err, "删除MdsNode应该成功")

	// 验证删除结果
	fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.Error(err, "查询已删除的MdsNode应该返回错误")
	suite.Nil(fm, "已删除的MdsNode应该为nil")

	// 测试边界情况：删除不存在的MdsNode
	err = suite.nodeRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的MdsNode应该成功（无操作）")
}

func (suite *MdsNodeTestSuite) TestGetModel() {
	// 创建测试数据
	cm := CreateTestMdsNodeModel(1, 1)
	err := suite.nodeRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsNode用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", cm.ID)
	suite.NoError(err, "查询MdsNode应该成功")
	suite.Equal(cm.ID, fm.ID)
	suite.Equal(cm.NodeRole, fm.NodeRole)
	suite.Equal(cm.IsEnable, fm.IsEnable)

	// 测试边界情况：查询不存在的MdsNode
	fm, err = suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的MdsNode应该返回错误")
	suite.Nil(fm, "查询不存在的MdsNode应该返回nil")

	// 测试边界情况：使用预加载
	fm, err = suite.nodeRepo.GetModel(context.Background(), []string{"MdsColony", "Host"}, "id = ?", cm.ID)
	suite.NoError(err, "使用预加载查询MdsNode应该成功")
	suite.Equal(cm.ID, fm.ID)
}

func (suite *MdsNodeTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		cm := CreateTestMdsNodeModel(1, 1)
		err := suite.nodeRepo.CreateModel(context.Background(), cm)
		suite.NoError(err, "创建MdsNode用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	count, models, err := suite.nodeRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询MdsNode列表应该成功")
	suite.Greater(count, int64(0), "MdsNode列表数量应该大于0")
	suite.NotNil(models, "MdsNode列表应该不为nil")
	suite.Greater(len(*models), 0, "MdsNode列表长度应该大于0")

	// 测试边界情况：空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"node_role": "non-existent"},
	}
	count2, models2, err := suite.nodeRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的MdsNode列表应该成功")
	suite.Equal(int64(0), count2, "不存在的MdsNode列表数量应该为0")
	suite.NotNil(models2, "不存在的MdsNode列表应该不为nil")
	suite.Len(*models2, 0, "不存在的MdsNode列表长度应该为0")
}

func (suite *MdsNodeTestSuite) TestContextTimeout() {
	// 创建测试数据
	cm := CreateTestMdsNodeModel(1, 1)
	err := suite.nodeRepo.CreateModel(context.Background(), cm)
	suite.NoError(err, "创建MdsNode用于超时测试应该成功")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err = suite.nodeRepo.GetModel(timeoutCtx, nil, "id = ?", cm.ID)
	suite.Error(err, "上下文超时后查询MdsNode应该返回错误")
}

func TestMdsNodeTestSuite(t *testing.T) {
	pts := &MdsNodeTestSuite{}
	suite.Run(t, pts)
}
