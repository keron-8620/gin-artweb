package resource

import (
	"context"
	"fmt"
	"os"
	"strings"
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

	// 测试边界情况:创建空模型
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

	// 测试边界情况:删除不存在的Package
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

	// 测试边界情况:查询不存在的Package
	fm, err = suite.packageRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err, "查询不存在的Package应该返回错误")
	suite.Nil(fm, "查询不存在的Package应该返回nil")

	// 测试边界情况:使用预加载（虽然PackageModel可能没有关联关系，但测试方法调用）
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
		Limit:   10,
		Offset:  0,
	}
	models, err := suite.packageRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "查询Package列表应该成功")
	suite.Greater(int64(len(models)), int64(0), "Package列表数量应该大于0")
	suite.NotNil(models, "Package列表应该不为nil")
	suite.Greater(int64(len(models)), int64(0), "Package列表长度应该大于0")

	// 测试边界情况:空列表（如果之前没有数据）
	// 注意:由于测试套件是共享数据库，这里可能不会为空，但我们仍然测试方法调用
	qp2 := database.QueryParams{
		Query: map[string]any{"label": "non-existent-label"},
	}
	models2, err := suite.packageRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的Package列表应该成功")
	suite.Equal(int64(0), int64(len(models2)), "不存在的Package列表数量应该为0")
	suite.NotNil(models2, "不存在的Package列表应该不为nil")
	suite.Len(models2, 0, "不存在的Package列表长度应该为0")
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

func (suite *PackageTestSuite) TestCreateModelWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的创建操作
	pm := CreateTestPackageModel()
	err := suite.packageRepo.CreateModel(timeoutCtx, pm)
	suite.Error(err, "上下文超时后创建Package应该返回错误")
}

func (suite *PackageTestSuite) TestDeleteModelWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的删除操作
	err := suite.packageRepo.DeleteModel(timeoutCtx, "id = ?", 1)
	suite.Error(err, "上下文超时后删除Package应该返回错误")
}

func (suite *PackageTestSuite) TestListModelWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的列表操作
	qp := database.QueryParams{}
	_, err := suite.packageRepo.ListModel(timeoutCtx, qp)
	suite.Error(err, "上下文超时后查询Package列表应该返回错误")
}

func (suite *PackageTestSuite) TestCountModelWithTimeout() {
	// 测试上下文超时情况
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的计数操作
	_, err := suite.packageRepo.CountModel(timeoutCtx, map[string]any{"label": "test"})
	suite.Error(err, "上下文超时后计数Package应该返回错误")
}

func (suite *PackageTestSuite) TestCountModel() {
	// 使用唯一的标签来避免数据残留的影响
	uniqueLabel := fmt.Sprintf("test-label-%s", uuid.NewString()[:8])

	pm := CreateTestPackageModel()
	pm.Label = uniqueLabel
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err)

	count, err := suite.packageRepo.CountModel(context.Background(), map[string]any{"label": uniqueLabel})
	suite.NoError(err)
	suite.Equal(int64(1), count)

	count, err = suite.packageRepo.CountModel(context.Background(), map[string]any{"label": "nonexistent"})
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

func (suite *PackageTestSuite) TestCountModelMultipleRecords() {
	for i := 0; i < 5; i++ {
		pm := CreateTestPackageModel()
		pm.Label = "count-label"
		err := suite.packageRepo.CreateModel(context.Background(), pm)
		suite.NoError(err)
	}

	count, err := suite.packageRepo.CountModel(context.Background(), map[string]any{"label": "count-label"})
	suite.NoError(err)
	suite.GreaterOrEqual(count, int64(5))
}

func (suite *PackageTestSuite) TestSavePackageFile() {
	// 测试正常保存
	testContent := "test package content"
	reader := strings.NewReader(testContent)
	tempFile := fmt.Sprintf("/tmp/test-package-%s.tar.gz", uuid.NewString())
	defer os.Remove(tempFile)

	err := suite.packageRepo.SavePackageFile(context.Background(), reader, tempFile, false)
	suite.NoError(err, "保存程序包文件应该成功")

	// 验证文件内容
	content, err := os.ReadFile(tempFile)
	suite.NoError(err)
	suite.Equal(testContent, string(content))

	// 测试覆盖保存
	newContent := "new test package content"
	reader2 := strings.NewReader(newContent)
	err = suite.packageRepo.SavePackageFile(context.Background(), reader2, tempFile, true)
	suite.NoError(err, "覆盖保存程序包文件应该成功")

	// 验证新内容
	content, err = os.ReadFile(tempFile)
	suite.NoError(err)
	suite.Equal(newContent, string(content))
}

func (suite *PackageTestSuite) TestSavePackageFileWithTimeout() {
	// 测试上下文超时情况
	testContent := "test package content"
	reader := strings.NewReader(testContent)
	tempFile := fmt.Sprintf("/tmp/test-package-timeout-%s.tar.gz", uuid.NewString())
	defer os.Remove(tempFile)

	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的保存操作
	err := suite.packageRepo.SavePackageFile(timeoutCtx, reader, tempFile, false)
	suite.Error(err, "上下文超时后保存程序包文件应该返回错误")
}

func (suite *PackageTestSuite) TestRemovePackageFile() {
	// 测试正常删除
	testContent := "test package content"
	reader := strings.NewReader(testContent)
	tempFile := fmt.Sprintf("/tmp/test-package-remove-%s.tar.gz", uuid.NewString())
	defer os.Remove(tempFile)

	// 先创建文件
	err := suite.packageRepo.SavePackageFile(context.Background(), reader, tempFile, false)
	suite.NoError(err, "保存程序包文件应该成功")

	// 验证文件存在
	_, err = os.Stat(tempFile)
	suite.NoError(err, "文件应该存在")

	// 测试删除文件
	err = suite.packageRepo.RemovePackageFile(context.Background(), tempFile)
	suite.NoError(err, "删除程序包文件应该成功")

	// 验证文件不存在
	_, err = os.Stat(tempFile)
	suite.Error(err, "文件应该不存在")
	suite.True(os.IsNotExist(err), "错误应该是文件不存在")
}

func (suite *PackageTestSuite) TestRemovePackageFileNonExistent() {
	// 测试删除不存在的文件
	nonExistentFile := fmt.Sprintf("/tmp/non-existent-package-%s.tar.gz", uuid.NewString())

	// 验证文件不存在
	_, err := os.Stat(nonExistentFile)
	suite.Error(err, "文件应该不存在")
	suite.True(os.IsNotExist(err), "错误应该是文件不存在")

	// 测试删除不存在的文件
	err = suite.packageRepo.RemovePackageFile(context.Background(), nonExistentFile)
	suite.NoError(err, "删除不存在的程序包文件应该成功")
}

func (suite *PackageTestSuite) TestRemovePackageFileWithTimeout() {
	// 测试上下文超时情况
	tempFile := fmt.Sprintf("/tmp/test-package-remove-timeout-%s.tar.gz", uuid.NewString())

	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// 等待超时
	time.Sleep(time.Millisecond * 2)

	// 测试超时后的删除操作
	err := suite.packageRepo.RemovePackageFile(timeoutCtx, tempFile)
	suite.Error(err, "上下文超时后删除程序包文件应该返回错误")
}

func (suite *PackageTestSuite) TestCreateModelNilModel() {
	// 测试边界情况:创建空模型
	err := suite.packageRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建空Package模型应该返回错误")
}

func (suite *PackageTestSuite) TestListModelWithPagination() {
	// 测试分页查询功能
	qp := database.QueryParams{
		OrderBy: []string{"id desc"},
		Limit:   5,
		Offset:  0,
	}
	models, err := suite.packageRepo.ListModel(context.Background(), qp)
	suite.NoError(err)
	suite.NotNil(models)
	suite.LessOrEqual(len(models), 5)
}

func (suite *PackageTestSuite) TestListModelEmptyResult() {
	qp := database.QueryParams{
		Query: map[string]any{"label": "totally-nonexistent-label-xyz"},
	}
	models, err := suite.packageRepo.ListModel(context.Background(), qp)
	suite.NoError(err)
	suite.Len(models, 0)
}

func (suite *PackageTestSuite) TestGetModelNotFound() {
	fm, err := suite.packageRepo.GetModel(context.Background(), nil, "id = ?", 999999)
	suite.Error(err)
	suite.Nil(fm)
}

func (suite *PackageTestSuite) TestGetModelWithPreloads() {
	pm := CreateTestPackageModel()
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err)

	fm, err := suite.packageRepo.GetModel(context.Background(), []string{}, "id = ?", pm.ID)
	suite.NoError(err)
	suite.Equal(pm.ID, fm.ID)
}

func (suite *PackageTestSuite) TestDeleteModelNotFound() {
	err := suite.packageRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err)
}

func (suite *PackageTestSuite) TestDeleteModelSuccess() {
	pm := CreateTestPackageModel()
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err)

	err = suite.packageRepo.DeleteModel(context.Background(), "id = ?", pm.ID)
	suite.NoError(err)

	fm, err := suite.packageRepo.GetModel(context.Background(), nil, "id = ?", pm.ID)
	suite.Error(err)
	suite.Nil(fm)
}

func (suite *PackageTestSuite) TestCreateModelSuccess() {
	pm := CreateTestPackageModel()
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err)
	suite.NotZero(pm.ID)
	suite.NotZero(pm.UploadedAt)

	fm, err := suite.packageRepo.GetModel(context.Background(), nil, "id = ?", pm.ID)
	suite.NoError(err)
	suite.Equal(pm.ID, fm.ID)
}

func (suite *PackageTestSuite) TestListModelWithQuery() {
	pm := CreateTestPackageModel()
	pm.Label = "special-label"
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err)

	qp := database.QueryParams{
		Query: map[string]any{"label": "special-label"},
	}
	models, err := suite.packageRepo.ListModel(context.Background(), qp)
	suite.NoError(err)
	suite.Len(models, 1)
	suite.Equal(pm.Label, models[0].Label)
}

func TestPackageTestSuite(t *testing.T) {
	pts := &PackageTestSuite{}
	suite.Run(t, pts)
}
