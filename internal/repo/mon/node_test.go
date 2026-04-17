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
	"gin-artweb/internal/shared/config"
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

// createTestMonNode 创建并保存测试MonNode到数据库
func (suite *MonNodeTestSuite) createTestMonNode() *monmodel.MonNodeModel {
	nm := CreateTestMonNodeModel()
	err := suite.nodeRepo.CreateModel(context.Background(), nm)
	suite.NoError(err, "创建测试MonNode应该成功")
	return nm
}

// createMultipleTestMonNodes 创建多个测试MonNode到数据库
func (suite *MonNodeTestSuite) createMultipleTestMonNodes(count int, namePrefix string) []*monmodel.MonNodeModel {
	nodes := make([]*monmodel.MonNodeModel, 0, count)
	for i := 0; i < count; i++ {
		nm := CreateTestMonNodeModel()
		nm.Name = fmt.Sprintf("%s-%d", namePrefix, i)
		err := suite.nodeRepo.CreateModel(context.Background(), nm)
		suite.NoError(err, "创建测试MonNode应该成功")
		nodes = append(nodes, nm)
	}
	return nodes
}

type MonNodeTestSuite struct {
	suite.Suite
	nodeRepo *MonNodeRepo
}

func (suite *MonNodeTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&resomodel.HostModel{}, &monmodel.MonNodeModel{}); err != nil {
		suite.Error(err, "数据库迁移失败")
	}

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
	slowThreshold := &config.DBSlowThreshold{
		WriteSlow: 100 * time.Millisecond,
		ReadSlow:  50 * time.Millisecond,
	}
	suite.nodeRepo = &MonNodeRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: slowThreshold,
	}
}

func (suite *MonNodeTestSuite) TestCreateModel() {
	// 测试正常创建
	nm := CreateTestMonNodeModel()
	err := suite.nodeRepo.CreateModel(context.Background(), nm)
	suite.NoError(err, "创建MonNode应该成功")
	suite.NotZero(nm.ID, "MonNode ID应该不为零")

	// 测试边界情况:创建空模型
	err = suite.nodeRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空MonNode模型应该返回错误")
}

func (suite *MonNodeTestSuite) TestUpdateModel() {
	// 创建测试数据
	nm := suite.createTestMonNode()

	// 表格驱动测试
	testCases := []struct {
		name        string
		updateData  map[string]any
		condition   string
		args        []any
		expectError bool
		verifyFunc  func()
	}{
		{
			name: "正常更新",
			updateData: map[string]any{
				"Name":       "updated-mon-node",
				"DeployPath": "/opt/mon-updated",
				"JavaHome":   "/usr/lib/jvm/java-17-openjdk-amd64",
			},
			condition:   "id = ?",
			args:        []any{nm.ID},
			expectError: false,
			verifyFunc: func() {
				fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", nm.ID)
				suite.NoError(err, "查询更新后的MonNode应该成功")
				suite.Equal("updated-mon-node", fm.Name)
				suite.Equal("/opt/mon-updated", fm.DeployPath)
				suite.Equal("/usr/lib/jvm/java-17-openjdk-amd64", fm.JavaHome)
			},
		},
		{
			name:        "更新数据为空",
			updateData:  map[string]any{},
			condition:   "id = ?",
			args:        []any{nm.ID},
			expectError: true,
			verifyFunc:  nil,
		},
		{
			name: "更新不存在的MonNode",
			updateData: map[string]any{
				"Name": "updated-mon-node",
			},
			condition:   "id = ?",
			args:        []any{999999},
			expectError: false,
			verifyFunc:  nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 构建参数，将condition和args合并
			params := append([]any{tc.condition}, tc.args...)
			err := suite.nodeRepo.UpdateModel(context.Background(), tc.updateData, params...)
			if tc.expectError {
				suite.Error(err, tc.name+"应该返回错误")
			} else {
				suite.NoError(err, tc.name+"应该成功")
				if tc.verifyFunc != nil {
					tc.verifyFunc()
				}
			}
		})
	}
}

func (suite *MonNodeTestSuite) TestDeleteModel() {
	// 创建测试数据
	nm := suite.createTestMonNode()

	// 表格驱动测试
	testCases := []struct {
		name        string
		condition   string
		args        []any
		expectError bool
		verifyFunc  func()
	}{
		{
			name:        "正常删除",
			condition:   "id = ?",
			args:        []any{nm.ID},
			expectError: false,
			verifyFunc: func() {
				fm, err := suite.nodeRepo.GetModel(context.Background(), nil, "id = ?", nm.ID)
				suite.Error(err, "查询已删除的MonNode应该返回错误")
				suite.Nil(fm, "已删除的MonNode应该为nil")
			},
		},
		{
			name:        "删除不存在的MonNode",
			condition:   "id = ?",
			args:        []any{999999},
			expectError: false,
			verifyFunc:  nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 构建参数，将condition和args合并
			params := append([]any{tc.condition}, tc.args...)
			err := suite.nodeRepo.DeleteModel(context.Background(), params...)
			if tc.expectError {
				suite.Error(err, tc.name+"应该返回错误")
			} else {
				suite.NoError(err, tc.name+"应该成功")
				if tc.verifyFunc != nil {
					tc.verifyFunc()
				}
			}
		})
	}
}

func (suite *MonNodeTestSuite) TestGetModel() {
	// 创建测试数据
	nm := suite.createTestMonNode()

	// 表格驱动测试
	testCases := []struct {
		name        string
		preload     []string
		condition   string
		args        []any
		expectError bool
		verifyFunc  func(*monmodel.MonNodeModel)
	}{
		{
			name:        "正常查询",
			preload:     nil,
			condition:   "id = ?",
			args:        []any{nm.ID},
			expectError: false,
			verifyFunc: func(fm *monmodel.MonNodeModel) {
				suite.Equal(nm.ID, fm.ID)
				suite.Equal(nm.Name, fm.Name)
				suite.Equal(nm.URL, fm.URL)
			},
		},
		{
			name:        "查询不存在的MonNode",
			preload:     nil,
			condition:   "id = ?",
			args:        []any{999999},
			expectError: true,
			verifyFunc:  nil,
		},
		{
			name:        "使用预加载",
			preload:     []string{"Host"},
			condition:   "id = ?",
			args:        []any{nm.ID},
			expectError: false,
			verifyFunc: func(fm *monmodel.MonNodeModel) {
				suite.Equal(nm.ID, fm.ID)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// 构建参数，将condition和args合并
			params := append([]any{tc.condition}, tc.args...)
			fm, err := suite.nodeRepo.GetModel(context.Background(), tc.preload, params...)
			if tc.expectError {
				suite.Error(err, tc.name+"应该返回错误")
				suite.Nil(fm, tc.name+"应该返回nil")
			} else {
				suite.NoError(err, tc.name+"应该成功")
				suite.NotNil(fm, tc.name+"应该返回非nil")
				if tc.verifyFunc != nil {
					tc.verifyFunc(fm)
				}
			}
		})
	}
}

func (suite *MonNodeTestSuite) TestListModel() {
	// 创建多个测试数据
	suite.createMultipleTestMonNodes(5, "mon-node")

	// 测试正常查询列表
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Limit:   10,
		Offset:  0,
	}
	models, err := suite.nodeRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询MonNode列表应该成功")
	suite.NotNil(models, "MonNode列表应该不为nil")
	suite.Greater(len(models), 0, "MonNode列表数量应该大于0")

	// 测试边界情况:空列表（如果之前没有数据）
	// 注意:由于测试套件是共享数据库，这里可能不会为空，但我们仍然测试方法调用
	qp2 := database.QueryParams{
		Query: map[string]any{"name": "non-existent-mon-node"},
	}
	models2, err := suite.nodeRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的MonNode列表应该成功")
	suite.NotNil(models2, "不存在的MonNode列表应该不为nil")
	suite.Len(models2, 0, "不存在的MonNode列表长度应该为0")
}

func (suite *MonNodeTestSuite) TestCountModel() {
	// 创建多个测试数据
	suite.createMultipleTestMonNodes(3, "count-test")

	// 测试正常计数
	count, err := suite.nodeRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "计数MonNode应该成功")
	suite.Greater(count, int64(0), "MonNode计数应该大于0")

	// 测试带条件计数
	countWithQuery, err := suite.nodeRepo.CountModel(context.Background(), map[string]any{"name LIKE ?": "count-test-%"})
	suite.NoError(err, "带条件计数MonNode应该成功")
	suite.GreaterOrEqual(countWithQuery, int64(3), "带条件计数应该至少为3")

	// 测试计数不存在的MonNode
	countNonExistent, err := suite.nodeRepo.CountModel(context.Background(), map[string]any{"name": "non-existent-mon-node"})
	suite.NoError(err, "计数不存在的MonNode应该成功")
	suite.Zero(countNonExistent, "不存在的MonNode计数应该为0")
}

func (suite *MonNodeTestSuite) TestDatabaseErrorScenarios() {
	// 测试上下文超时导致的数据库操作失败
	testTimeoutError := func() {
		// 创建测试数据
		nm := suite.createTestMonNode()

		// 创建一个已经超时的上下文
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		// 等待上下文超时
		time.Sleep(time.Millisecond)

		// 测试各种操作在超时上下文下的行为
		_, err := suite.nodeRepo.GetModel(timeoutCtx, nil, "id = ?", nm.ID)
		suite.Error(err, "超时上下文下查询MonNode应该返回错误")

		err = suite.nodeRepo.UpdateModel(timeoutCtx, map[string]any{"name": "test"}, "id = ?", nm.ID)
		suite.Error(err, "超时上下文下更新MonNode应该返回错误")

		err = suite.nodeRepo.DeleteModel(timeoutCtx, "id = ?", nm.ID)
		suite.Error(err, "超时上下文下删除MonNode应该返回错误")

		_, err = suite.nodeRepo.ListModel(timeoutCtx, database.QueryParams{})
		suite.Error(err, "超时上下文下列表查询应该返回错误")

		_, err = suite.nodeRepo.CountModel(timeoutCtx, nil)
		suite.Error(err, "超时上下文下计数应该返回错误")
	}

	testTimeoutError()

	// 注意:空条件的删除和更新操作会失败，因为数据库操作需要WHERE条件
	// 这里我们不测试空条件操作，因为它们不是预期的用例
}

func TestMonNodeTestSuite(t *testing.T) {
	pts := &MonNodeTestSuite{}
	suite.Run(t, pts)
}
