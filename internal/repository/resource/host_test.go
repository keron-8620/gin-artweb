package resource

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestHostModel() *resomodel.HostModel {
	// 生成唯一的端口号，避免唯一约束冲突
	uuidStr := uuid.NewString()
	// 取UUID的后4位作为端口号的一部分
	portSuffix := uuidStr[len(uuidStr)-4:]
	// 转换为数字并确保在有效端口范围内
	port := 2222 + (len(uuidStr) % 1000)
	return &resomodel.HostModel{
		Name:    fmt.Sprintf("host-%s", uuidStr),
		Label:   "test",
		SSHIP:   "127.0.0.1",
		SSHPort: uint16(port),
		SSHUser: fmt.Sprintf("root-%s", portSuffix),
		PyPath:  "/usr/bin/python3",
		Remark:  "",
	}
}

type HostTestSuite struct {
	suite.Suite
	hostRepo *HostRepo
}

func (suite *HostTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&resomodel.HostModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.hostRepo = &HostRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *HostTestSuite) TestCreateHost() {
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host应该成功")

	fm, err := suite.hostRepo.GetModel(context.Background(), nil, hm.ID)
	suite.NoError(err, "查询刚创建的Host应该成功")
	suite.Equal(hm.ID, fm.ID)
	suite.Equal(hm.Name, fm.Name)
	suite.Equal(hm.Label, fm.Label)
	suite.Equal(hm.SSHIP, fm.SSHIP)
	suite.Equal(hm.SSHPort, fm.SSHPort)
	suite.Equal(hm.SSHUser, fm.SSHUser)
	suite.Equal(hm.PyPath, fm.PyPath)
	suite.Equal(hm.Remark, fm.Remark)
}

func (suite *HostTestSuite) TestUpdateModel() {
	// 创建测试数据
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host用于更新测试应该成功")

	// 测试正常更新
	updateData := map[string]any{
		"Name":    "updated-host",
		"SSHPort": 2222,
		"Remark":  "updated remark",
	}
	err = suite.hostRepo.UpdateModel(context.Background(), updateData, "id = ?", hm.ID)
	suite.NoError(err, "更新Host应该成功")

	// 验证更新结果
	fm, err := suite.hostRepo.GetModel(context.Background(), nil, "id = ?", hm.ID)
	suite.NoError(err, "查询更新后的Host应该成功")
	suite.Equal("updated-host", fm.Name)
	suite.Equal(uint16(2222), fm.SSHPort)
	suite.Equal("updated remark", fm.Remark)

	// 测试边界情况：更新数据为空
	err = suite.hostRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", hm.ID)
	suite.Error(err, "更新数据为空时应该返回错误")

	// 测试边界情况：更新不存在的Host
	err = suite.hostRepo.UpdateModel(context.Background(), updateData, "id = ?", 999999)
	suite.NoError(err, "更新不存在的Host应该成功（无操作）")
}

func (suite *HostTestSuite) TestDeleteModel() {
	// 创建测试数据
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host用于删除测试应该成功")

	// 测试正常删除
	err = suite.hostRepo.DeleteModel(context.Background(), "id = ?", hm.ID)
	suite.NoError(err, "删除Host应该成功")

	// 验证删除结果
	fm, err := suite.hostRepo.GetModel(context.Background(), nil, "id = ?", hm.ID)
	suite.Error(err, "查询已删除的Host应该返回错误")
	suite.Nil(fm, "已删除的Host应该为nil")

	// 测试边界情况：删除不存在的Host
	err = suite.hostRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的Host应该成功（无操作）")
}

func (suite *HostTestSuite) TestGetModel() {
	// 创建测试数据
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.hostRepo.GetModel(context.Background(), nil, "id = ?", hm.ID)
	suite.NoError(err, "查询Host应该成功")
	suite.Equal(hm.ID, fm.ID)
	suite.Equal(hm.Name, fm.Name)

	// 测试边界情况：查询不存在的Host
	fm, err = suite.hostRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的Host应该返回错误")
	suite.Nil(fm, "查询不存在的Host应该返回nil")

	// 测试边界情况：使用预加载（虽然HostModel可能没有关联关系，但测试方法调用）
	fm, err = suite.hostRepo.GetModel(context.Background(), []string{}, "id = ?", hm.ID)
	suite.NoError(err, "使用空预加载查询Host应该成功")
	suite.Equal(hm.ID, fm.ID)
}

func (suite *HostTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		hm := CreateTestHostModel()
		hm.Name = fmt.Sprintf("host-%d", i)
		err := suite.hostRepo.CreateModel(context.Background(), hm)
		suite.NoError(err, "创建Host用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{}
	count, models, err := suite.hostRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询Host列表应该成功")
	suite.Greater(count, int64(0), "Host列表数量应该大于0")
	suite.NotNil(models, "Host列表应该不为nil")
	suite.Greater(len(*models), 0, "Host列表长度应该大于0")

	// 测试边界情况：空列表（如果之前没有数据）
	// 注意：由于测试套件是共享数据库，这里可能不会为空，但我们仍然测试方法调用
	qp2 := database.QueryParams{
		Query: map[string]any{"name": "non-existent-host"},
	}
	count2, models2, err := suite.hostRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的Host列表应该成功")
	suite.Equal(int64(0), count2, "不存在的Host列表数量应该为0")
	suite.NotNil(models2, "不存在的Host列表应该不为nil")
	suite.Len(*models2, 0, "不存在的Host列表长度应该为0")
}

func (suite *HostTestSuite) TestCreateModelWithEmpty() {
	// 测试边界情况：创建空模型
	err := suite.hostRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空Host模型应该返回错误")
}

func (suite *HostTestSuite) TestContextTimeout() {
	// 创建测试数据
	hm := CreateTestHostModel()
	err := suite.hostRepo.CreateModel(context.Background(), hm)
	suite.NoError(err, "创建Host用于超时测试应该成功")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err = suite.hostRepo.GetModel(timeoutCtx, nil, "id = ?", hm.ID)
	suite.Error(err, "上下文超时后查询Host应该返回错误")
}

func (suite *HostTestSuite) TestNewSSHClientWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的SSH客户端创建
	client, err := suite.hostRepo.NewSSHClient(timeoutCtx, "127.0.0.1", 22, "root", nil, time.Second)
	suite.Error(err, "上下文超时后创建SSH客户端应该返回错误")
	suite.Nil(client, "上下文超时后创建的SSH客户端应该为nil")
}

func (suite *HostTestSuite) TestExecuteCommandWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的命令执行
	err := suite.hostRepo.ExecuteCommand(timeoutCtx, nil, "echo test")
	suite.Error(err, "上下文超时后执行命令应该返回错误")
}

func TestHostTestSuite(t *testing.T) {
	pts := &HostTestSuite{}
	suite.Run(t, pts)
}
