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

func CreateTestPackageModel() *resomodel.PackageModel {
	return &resomodel.PackageModel{
		Label:           "test",
		StorageFilename: fmt.Sprintf("test-package-%s.tar.gz", uuid.NewString()),
		OriginFilename:  fmt.Sprintf("test-package-%s.tar.gz", uuid.NewString()),
		Version:         fmt.Sprintf("1.0.0-%s", uuid.NewString()[:8]),
	}
}

type PackageTestSuite struct {
	suite.Suite
	packageRepo *PackageRepo
}

func (suite *PackageTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&resomodel.PackageModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.packageRepo = &PackageRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *PackageTestSuite) TestCreateModel() {
	// 测试正常创建
	pm := CreateTestPackageModel()
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err, "创建Package应该成功")
	suite.NotZero(pm.ID, "Package ID应该不为零")
	suite.NotZero(pm.UploadedAt, "Package UploadedAt应该不为零")

	// 测试边界情况：创建空模型
	err = suite.packageRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空Package模型应该返回错误")
}

func (suite *PackageTestSuite) TestDeleteModel() {
	// 创建测试数据
	pm := CreateTestPackageModel()
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err, "创建Package用于删除测试应该成功")

	// 测试正常删除
	err = suite.packageRepo.DeleteModel(context.Background(), "id = ?", pm.ID)
	suite.NoError(err, "删除Package应该成功")

	// 验证删除结果
	fm, err := suite.packageRepo.GetModel(context.Background(), nil, "id = ?", pm.ID)
	suite.Error(err, "查询已删除的Package应该返回错误")
	suite.Nil(fm, "已删除的Package应该为nil")

	// 测试边界情况：删除不存在的Package
	err = suite.packageRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的Package应该成功（无操作）")
}

func (suite *PackageTestSuite) TestGetModel() {
	// 创建测试数据
	pm := CreateTestPackageModel()
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err, "创建Package用于查询测试应该成功")

	// 测试正常查询
	fm, err := suite.packageRepo.GetModel(context.Background(), nil, "id = ?", pm.ID)
	suite.NoError(err, "查询Package应该成功")
	suite.Equal(pm.ID, fm.ID)
	suite.Equal(pm.StorageFilename, fm.StorageFilename)
	suite.Equal(pm.Version, fm.Version)

	// 测试边界情况：查询不存在的Package
	fm, err = suite.packageRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的Package应该返回错误")
	suite.Nil(fm, "查询不存在的Package应该返回nil")

	// 测试边界情况：使用预加载（虽然PackageModel可能没有关联关系，但测试方法调用）
	fm, err = suite.packageRepo.GetModel(context.Background(), []string{}, "id = ?", pm.ID)
	suite.NoError(err, "使用空预加载查询Package应该成功")
	suite.Equal(pm.ID, fm.ID)
}

func (suite *PackageTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		pm := CreateTestPackageModel()
		pm.StorageFilename = fmt.Sprintf("test-package-%d.tar.gz", i)
		pm.OriginFilename = fmt.Sprintf("test-package-%d.tar.gz", i)
		err := suite.packageRepo.CreateModel(context.Background(), pm)
		suite.NoError(err, "创建Package用于列表测试应该成功")
	}

	// 测试正常查询列表
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	count, models, err := suite.packageRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询Package列表应该成功")
	suite.Greater(count, int64(0), "Package列表数量应该大于0")
	suite.NotNil(models, "Package列表应该不为nil")
	suite.Greater(len(*models), 0, "Package列表长度应该大于0")

	// 测试边界情况：空列表（如果之前没有数据）
	// 注意：由于测试套件是共享数据库，这里可能不会为空，但我们仍然测试方法调用
	qp2 := database.QueryParams{
		Query: map[string]any{"label": "non-existent-label"},
	}
	count2, models2, err := suite.packageRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的Package列表应该成功")
	suite.Equal(int64(0), count2, "不存在的Package列表数量应该为0")
	suite.NotNil(models2, "不存在的Package列表应该不为nil")
	suite.Len(*models2, 0, "不存在的Package列表长度应该为0")
}

func (suite *PackageTestSuite) TestContextTimeout() {
	// 创建测试数据
	pm := CreateTestPackageModel()
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err, "创建Package用于超时测试应该成功")

	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的操作
	_, err = suite.packageRepo.GetModel(timeoutCtx, nil, "id = ?", pm.ID)
	suite.Error(err, "上下文超时后查询Package应该返回错误")
}

func TestPackageTestSuite(t *testing.T) {
	pts := &PackageTestSuite{}
	suite.Run(t, pts)
}
