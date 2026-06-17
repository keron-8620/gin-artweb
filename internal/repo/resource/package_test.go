package resource

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestPackageModel(opts ...func(*resomodel.PackageModel)) *resomodel.PackageModel {
	pm := &resomodel.PackageModel{
		Label:           "test",
		StorageFilename: fmt.Sprintf("test-package-%s.tar.gz", uuid.NewString()),
		OriginFilename:  fmt.Sprintf("test-package-%s.tar.gz", uuid.NewString()),
		Version:         fmt.Sprintf("1.0.0-%s", uuid.NewString()[:8]),
	}

	// 应用可选的定制函数
	for _, opt := range opts {
		opt(pm)
	}

	return pm
}

// 辅助函数，用于定制测试模型
func WithLabel(label string) func(*resomodel.PackageModel) {
	return func(pm *resomodel.PackageModel) {
		pm.Label = label
	}
}

func WithStorageFilename(filename string) func(*resomodel.PackageModel) {
	return func(pm *resomodel.PackageModel) {
		pm.StorageFilename = filename
	}
}

func WithOriginFilename(filename string) func(*resomodel.PackageModel) {
	return func(pm *resomodel.PackageModel) {
		pm.OriginFilename = filename
	}
}

func WithVersion(version string) func(*resomodel.PackageModel) {
	return func(pm *resomodel.PackageModel) {
		pm.Version = version
	}
}

type PackageTestSuite struct {
	suite.Suite
	packageRepo *PackageRepo
}

func (suite *PackageTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&resomodel.PackageModel{}); err != nil {
		suite.Error(err, "数据库迁移失败")
	}
	dbTimeout := test.NewTestDBTimeouts()
	dbSlowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()
	suite.packageRepo = &PackageRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: dbSlowThreshold,
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
	fm, err := suite.packageRepo.GetModel(context.Background(), "id = ?", pm.ID)
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
	fm, err := suite.packageRepo.GetModel(context.Background(), "id = ?", pm.ID)
	suite.NoError(err, "查询Package应该成功")
	suite.Equal(pm.ID, fm.ID, "Package ID应该一致")
	suite.Equal(pm.StorageFilename, fm.StorageFilename, "Package StorageFilename应该一致")
	suite.Equal(pm.Version, fm.Version, "Package Version应该一致")

	// 测试边界情况:查询不存在的Package
	fm, err = suite.packageRepo.GetModel(context.Background(), "id = ?", 999999)
	suite.Error(err, "查询不存在的Package应该返回错误")
	suite.Nil(fm, "查询不存在的Package应该返回nil")

	// 测试边界情况:使用预加载（虽然PackageModel可能没有关联关系，但测试方法调用）
	fm, err = suite.packageRepo.GetModel(context.Background(), "id = ?", pm.ID)
	suite.NoError(err, "使用空预加载查询Package应该成功")
	suite.Equal(pm.ID, fm.ID, "Package ID应该一致")
}

func (suite *PackageTestSuite) TestListModel() {
	// 创建多个测试数据
	for i := 0; i < 5; i++ {
		pm := CreateTestPackageModel(
			WithStorageFilename(fmt.Sprintf("test-package-%d.tar.gz", i)),
			WithOriginFilename(fmt.Sprintf("test-package-%d.tar.gz", i)),
		)
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
	suite.NotNil(models, "Package列表应该不为nil")
	suite.Greater(len(models), 0, "Package列表长度应该大于0")

	// 测试边界情况:空列表
	qp2 := database.QueryParams{
		Query: map[string]any{"label": "non-existent-label"},
	}
	models2, err := suite.packageRepo.ListModel(context.Background(), qp2)
	suite.NoError(err, "查询不存在的Package列表应该成功")
	suite.NotNil(models2, "不存在的Package列表应该不为nil")
	suite.Len(models2, 0, "不存在的Package列表长度应该为0")

	// 测试分页查询功能
	qp3 := database.QueryParams{
		OrderBy: []string{"id desc"},
		Limit:   5,
		Offset:  0,
	}
	models3, err := suite.packageRepo.ListModel(context.Background(), qp3)
	suite.NoError(err, "分页查询Package列表应该成功")
	suite.NotNil(models3, "分页查询结果应该不为nil")
	suite.LessOrEqual(len(models3), 5, "分页查询结果长度应该小于等于5")

	// 测试带条件查询
	specialLabel := fmt.Sprintf("special-label-%s", uuid.NewString()[:8])
	pm := CreateTestPackageModel(WithLabel(specialLabel))
	err = suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err, "创建带特殊标签的Package应该成功")

	qp4 := database.QueryParams{
		Query: map[string]any{"label": specialLabel},
	}
	models4, err := suite.packageRepo.ListModel(context.Background(), qp4)
	suite.NoError(err, "带条件查询Package列表应该成功")
	suite.Len(models4, 1, "带条件查询结果应该只有一条")
	suite.Equal(specialLabel, models4[0].Label, "查询结果标签应该一致")
}

func (suite *PackageTestSuite) TestCountModel() {
	// 使用唯一的标签来避免数据残留的影响
	uniqueLabel := fmt.Sprintf("test-label-%s", uuid.NewString()[:8])

	pm := CreateTestPackageModel(WithLabel(uniqueLabel))
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err, "创建Package用于计数测试应该成功")

	count, err := suite.packageRepo.CountModel(context.Background(), map[string]any{"label": uniqueLabel})
	suite.NoError(err, "计数Package应该成功")
	suite.Equal(int64(1), count, "计数结果应该为1")

	count, err = suite.packageRepo.CountModel(context.Background(), map[string]any{"label": "nonexistent"})
	suite.NoError(err, "计数不存在的Package应该成功")
	suite.Equal(int64(0), count, "计数不存在的Package结果应该为0")

	// 测试多个记录的计数
	countLabel := fmt.Sprintf("count-label-%s", uuid.NewString()[:8])
	for i := 0; i < 5; i++ {
		pm := CreateTestPackageModel(WithLabel(countLabel))
		err := suite.packageRepo.CreateModel(context.Background(), pm)
		suite.NoError(err, "创建Package用于多记录计数测试应该成功")
	}

	count, err = suite.packageRepo.CountModel(context.Background(), map[string]any{"label": countLabel})
	suite.NoError(err, "计数多个Package应该成功")
	suite.GreaterOrEqual(count, int64(5), "计数结果应该大于等于5")
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
	suite.NoError(err, "读取保存的文件应该成功")
	suite.Equal(testContent, string(content), "文件内容应该与原始内容一致")

	// 测试覆盖保存
	newContent := "new test package content"
	reader2 := strings.NewReader(newContent)
	err = suite.packageRepo.SavePackageFile(context.Background(), reader2, tempFile, true)
	suite.NoError(err, "覆盖保存程序包文件应该成功")

	// 验证新内容
	content, err = os.ReadFile(tempFile)
	suite.NoError(err, "读取覆盖保存的文件应该成功")
	suite.Equal(newContent, string(content), "文件内容应该与新内容一致")
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

	// 测试删除不存在的文件
	nonExistentFile := fmt.Sprintf("/tmp/non-existent-package-%s.tar.gz", uuid.NewString())

	// 验证文件不存在
	_, err = os.Stat(nonExistentFile)
	suite.Error(err, "文件应该不存在")
	suite.True(os.IsNotExist(err), "错误应该是文件不存在")

	// 测试删除不存在的文件
	err = suite.packageRepo.RemovePackageFile(context.Background(), nonExistentFile)
	suite.NoError(err, "删除不存在的程序包文件应该成功")
}

func (suite *PackageTestSuite) TestOperationsWithTimeout() {
	pm := CreateTestPackageModel()
	err := suite.packageRepo.CreateModel(context.Background(), pm)
	suite.NoError(err, "创建Package用于超时测试应该成功")

	timeoutCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = suite.packageRepo.GetModel(timeoutCtx, "id = ?", pm.ID)
	suite.Error(err, "上下文取消后查询Package应该返回错误")

	pm2 := CreateTestPackageModel()
	err = suite.packageRepo.CreateModel(timeoutCtx, pm2)
	suite.Error(err, "上下文取消后创建Package应该返回错误")

	err = suite.packageRepo.DeleteModel(timeoutCtx, "id = ?", pm.ID)
	suite.Error(err, "上下文取消后删除Package应该返回错误")

	qp := database.QueryParams{}
	_, err = suite.packageRepo.ListModel(timeoutCtx, qp)
	suite.Error(err, "上下文取消后查询Package列表应该返回错误")

	_, err = suite.packageRepo.CountModel(timeoutCtx, map[string]any{"label": "test"})
	suite.Error(err, "上下文取消后计数Package应该返回错误")

	testContent := "test package content"
	reader := strings.NewReader(testContent)
	tempFile := fmt.Sprintf("/tmp/test-package-timeout-%s.tar.gz", uuid.NewString())
	defer os.Remove(tempFile)

	err = suite.packageRepo.SavePackageFile(timeoutCtx, reader, tempFile, false)
	suite.Error(err, "上下文取消后保存程序包文件应该返回错误")

	tempFile2 := fmt.Sprintf("/tmp/test-package-remove-timeout-%s.tar.gz", uuid.NewString())
	err = suite.packageRepo.RemovePackageFile(timeoutCtx, tempFile2)
	suite.Error(err, "上下文取消后删除程序包文件应该返回错误")
}

func TestPackageTestSuite(t *testing.T) {
	pts := &PackageTestSuite{}
	suite.Run(t, pts)
}
