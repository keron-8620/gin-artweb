package mon

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	monmodel "gin-artweb/internal/model/mon"
	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestMonNodeModel() *monmodel.MonNodeModel {
	return &monmodel.MonNodeModel{
		Name:        fmt.Sprintf("mon-node-%s", uuid.NewString()),
		DeployPath:  "/opt/mon",
		OutportPath: "/opt/mon/outport",
		JavaHome:    "/usr/lib/jvm/java-11-openjdk-amd64",
		URL:         fmt.Sprintf("http://localhost:8080/%s", uuid.NewString()),
		HostID:      1, // 假设主机ID为1
	}
}

type MonNodeTestSuite struct {
	suite.Suite
	nodeRepo *MonNodeRepo
}

func (suite *MonNodeTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&resomodel.HostModel{}, &monmodel.MonNodeModel{})

	// 创建一个测试主机，因为MonNodeModel需要关联HostID
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

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.nodeRepo = &MonNodeRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *MonNodeTestSuite) TestCreateModel() {
	// 测试正常创建
	nm := CreateTestMonNodeModel()
	err := suite.nodeRepo.CreateModel(context.Background(), nm)
	suite.NoError(err, "创建MonNode应该成功")
	suite.NotZero(nm.ID, "MonNode ID应该不为零")

	// 测试边界情况：创建空模型
	err = suite.nodeRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空MonNode模型应该返回错误")
}

func (suite *MonNodeTestSuite) TestUpdateModel() {
	// 创建测试数据
	nm := CreateTestMonNodeModel()
	err := suite.nodeRepo.CreateModel(context.Background(), nm)
	suite.NoError(err, "创建MonNode用于更新测试应该成功")

	// 测试正常更新
	updateData := map[string]any{
		"Name":       "updated-mon-node",
		"DeployPath": "/opt/mon-updated",
		"JavaHome":   "/usr/lib/jvm/java-17-openjdk-amd64",
	}
	err = suite.nodeRepo.UpdateModel(context.Background(), updateData, "id = ?", nm.ID)
	suite.NoError(err, "更新MonNode应该成功")

	// 验证更新结果
	fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", nm.ID)
	suite.NoError(err, "查询更新后的MonNode应该成功")
	suite.Equal("updated-mon-node", fm.Name)
	suite.Equal("/opt/mon-updated", fm.DeployPath)
	suite.Equal("/usr/lib/jvm/java-17-openjdk-amd64", fm.JavaHome)

	// 测试边界情况：更新数据为空
	err = suite.nodeRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", nm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况：更新不存在的MonNode
	err = suite.nodeRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的MonNode应该成功（无操作）")
}

func (suite *MonNodeTestSuite) TestDeleteModel() {
	// 创建测试数据
	nm := CreateTestMonNodeModel()
	err := suite.nodeRepo.CreateModel(context.Background(), nm)
	suite.NoError(err, "创建MonNode用于删除测试应该成功")

	// 测试正常删除
	err = suite.nodeRepo.DeleteModel(context.Background(), "id = ?", nm.ID)
	suite.NoError(err, "删除MonNode应该成功")

	// 验证删除结果
	fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", nm.ID)
	suite.Error(err, "查询已删除的MonNode应该返回错误")
	suite.Nil(fm, "已删除的MonNode应该为nil")

	// 测试边界情况：删除不存在的MonNode
	err = suite.nodeRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的MonNode应该成功（无操作）")
}

func (suite *MonNodeTestSuite) TestGetModel() {
	// 创建测试数据
	nm := CreateTestMonNodeModel()
	err := suite.nodeRepo.CreateModel(context.Background(), nm)
	suite.NoError(err, "创建MonNode用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", nm.ID)
	suite.NoError(err, "查询MonNode应该成功")
	suite.Equal(nm.ID, fm.ID)
	suite.Equal(nm.Name, fm.Name)
	suite.Equal(nm.URL, fm.URL)

	// 测试边界情况：查询不存在的MonNode
	fm, err = suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的MonNode应该返回错误")
	suite.Nil(fm, "查询不存在的MonNode应该返回nil")

	// 测试边界情况：使用预加载
	fm, err = suite.nodeRepo.GetModel(context.Background(), []string{"Host"}, "id = ?", nm.ID)
	suite.NoError(err, "使用预加载查询MonNode应该成功")
	suite.Equal(nm.ID, fm.ID)
}

func (suite *MonNodeTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		nm := CreateTestMonNodeModel()
		nm.Name = fmt.Sprintf("mon-node-%d", i)
		err := suite.nodeRepo.CreateModel(context.Background(), nm)
		suite.NoError(err, "创建MonNode用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	count, models, err := suite.nodeRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询MonNode列表应该成功")
	suite.Greater(count, int64(0), "MonNode列表数量应该大于0")
	suite.NotNil(models, "MonNode列表应该不为nil")
	suite.Greater(len(*models), 0, "MonNode列表长度应该大于0")

	// 测试边界情况：空列表（如果之前没有数据）
	// 注意：由于测试套件是共享数据库，这里可能不会为空，但我们仍然测试方法调用
	qp2 := database.QueryParams{
		Query: map[string]any{"name": "non-existent-mon-node"},
	}
	count2, models2, err := suite.nodeRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的MonNode列表应该成功")
	suite.Equal(int64(0), count2, "不存在的MonNode列表数量应该为0")
	suite.NotNil(models2, "不存在的MonNode列表应该不为nil")
	suite.Len(*models2, 0, "不存在的MonNode列表长度应该为0")
}

func (suite *MonNodeTestSuite) TestContextTimeout() {
	// 创建测试数据
	nm := CreateTestMonNodeModel()
	err := suite.nodeRepo.CreateModel(context.Background(), nm)
	suite.NoError(err, "创建MonNode用于超时测试应该成功")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err = suite.nodeRepo.GetModel(timeoutCtx, nil, "id = ?", nm.ID)
	suite.Error(err, "上下文超时后查询MonNode应该返回错误")
}

func TestMonNodeTestSuite(t *testing.T) {
	pts := &MonNodeTestSuite{}
	suite.Run(t, pts)
}
