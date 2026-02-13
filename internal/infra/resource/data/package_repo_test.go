package data

import (
	"context"
	"testing"

	"gin-artweb/internal/infra/resource/model"
	"gin-artweb/internal/shared/database"

	"github.com/stretchr/testify/assert"
)

func TestPackageRepo_CreateModel(t *testing.T) {
	t.Run("成功创建程序包模型", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestPackageRepo()
		ctx := context.Background()
		packageModel := &model.PackageModel{
			Label:           "test",
			StorageFilename: "test-package-1.0.0.tar.gz",
			OriginFilename:  "test-package-1.0.0.tar.gz",
			Version:         "1.0.0",
		}

		// 执行测试
		err := repo.CreateModel(ctx, packageModel)

		// 验证结果
		assert.NoError(t, err)
		assert.NotZero(t, packageModel.ID)
		assert.NotZero(t, packageModel.UploadedAt)
	})

	t.Run("创建空程序包模型失败", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestPackageRepo()
		ctx := context.Background()

		// 执行测试
		err := repo.CreateModel(ctx, nil)

		// 验证结果
		assert.Error(t, err)
	})
}

func TestPackageRepo_DeleteModel(t *testing.T) {
	t.Run("成功删除程序包模型", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestPackageRepo()
		ctx := context.Background()
		packageModel := &model.PackageModel{
			Label:           "test",
			StorageFilename: "test-package-1.0.0.tar.gz",
			OriginFilename:  "test-package-1.0.0.tar.gz",
			Version:         "1.0.0",
		}

		// 先创建模型
		err := repo.CreateModel(ctx, packageModel)
		assert.NoError(t, err)

		// 执行删除操作
		err = repo.DeleteModel(ctx, "id = ?", packageModel.ID)

		// 验证结果
		assert.NoError(t, err)
	})
}

func TestPackageRepo_GetModel(t *testing.T) {
	t.Run("成功获取程序包模型", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestPackageRepo()
		ctx := context.Background()
		packageModel := &model.PackageModel{
			Label:           "test",
			StorageFilename: "test-package-1.0.0.tar.gz",
			OriginFilename:  "test-package-1.0.0.tar.gz",
			Version:         "1.0.0",
		}

		// 先创建模型
		err := repo.CreateModel(ctx, packageModel)
		assert.NoError(t, err)

		// 执行查询操作
		result, err := repo.GetModel(ctx, []string{}, "id = ?", packageModel.ID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, packageModel.StorageFilename, result.StorageFilename)
		assert.Equal(t, packageModel.Version, result.Version)
	})
}

func TestPackageRepo_ListModel(t *testing.T) {
	t.Run("成功获取程序包模型列表", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestPackageRepo()
		ctx := context.Background()

		// 创建多个测试模型
		for i := 1; i <= 3; i++ {
			packageModel := &model.PackageModel{
				Label:           "test",
				StorageFilename: "test-package-" + string(rune('0'+i)) + ".tar.gz",
				OriginFilename:  "test-package-" + string(rune('0'+i)) + ".tar.gz",
				Version:         "1.0.0",
			}
			err := repo.CreateModel(ctx, packageModel)
			assert.NoError(t, err)
		}

		// 执行查询操作
		queryParams := database.QueryParams{
			OrderBy: []string{"id desc"},
			Size:    10,
			Page:    0,
			IsCount: true,
		}
		count, result, err := repo.ListModel(ctx, queryParams)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, count, int64(3))
		assert.Len(t, *result, 3)
	})
}
