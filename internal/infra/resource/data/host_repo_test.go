package data

import (
	"context"
	"testing"
	"time"

	"gin-artweb/internal/infra/resource/model"
	"gin-artweb/internal/shared/database"

	"github.com/stretchr/testify/assert"
)

func TestHostRepo_CreateModel(t *testing.T) {
	t.Run("成功创建主机模型", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestHostRepo()
		ctx := context.Background()
		hostModel := &model.HostModel{
			Name:    "test-host",
			Label:   "test",
			SSHIP:   "192.168.1.1",
			SSHPort: 22,
			SSHUser: "admin",
			PyPath:  "/usr/bin/python3",
			Remark:  "测试主机",
		}

		// 执行测试
		err := repo.CreateModel(ctx, hostModel)

		// 验证结果
		assert.NoError(t, err)
		assert.NotZero(t, hostModel.ID)
		assert.NotZero(t, hostModel.CreatedAt)
		assert.NotZero(t, hostModel.UpdatedAt)
	})

	t.Run("创建空主机模型失败", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestHostRepo()
		ctx := context.Background()

		// 执行测试
		err := repo.CreateModel(ctx, nil)

		// 验证结果
		assert.Error(t, err)
	})
}

func TestHostRepo_UpdateModel(t *testing.T) {
	t.Run("成功更新主机模型", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestHostRepo()
		ctx := context.Background()
		hostModel := &model.HostModel{
			Name:    "test-host",
			Label:   "test",
			SSHIP:   "192.168.1.1",
			SSHPort: 22,
			SSHUser: "admin",
			PyPath:  "/usr/bin/python3",
			Remark:  "测试主机",
		}

		// 先创建模型
		err := repo.CreateModel(ctx, hostModel)
		assert.NoError(t, err)

		// 准备更新数据
		updateData := map[string]any{
			"name":   "updated-test-host",
			"remark": "更新后的测试主机",
		}

		// 执行更新操作
		err = repo.UpdateModel(ctx, updateData, "id = ?", hostModel.ID)

		// 验证结果
		assert.NoError(t, err)

		// 验证更新是否生效
		updatedHost, err := repo.GetModel(ctx, []string{}, "id = ?", hostModel.ID)
		assert.NoError(t, err)
		assert.Equal(t, "updated-test-host", updatedHost.Name)
		assert.Equal(t, "更新后的测试主机", updatedHost.Remark)
	})

	t.Run("更新空数据失败", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestHostRepo()
		ctx := context.Background()

		// 执行测试
		err := repo.UpdateModel(ctx, map[string]any{}, "id = ?", 1)

		// 验证结果
		assert.Error(t, err)
	})
}

func TestHostRepo_DeleteModel(t *testing.T) {
	t.Run("成功删除主机模型", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestHostRepo()
		ctx := context.Background()
		hostModel := &model.HostModel{
			Name:    "test-host",
			Label:   "test",
			SSHIP:   "192.168.1.1",
			SSHPort: 22,
			SSHUser: "admin",
			PyPath:  "/usr/bin/python3",
			Remark:  "测试主机",
		}

		// 先创建模型
		err := repo.CreateModel(ctx, hostModel)
		assert.NoError(t, err)

		// 执行删除操作
		err = repo.DeleteModel(ctx, "id = ?", hostModel.ID)

		// 验证结果
		assert.NoError(t, err)
	})
}

func TestHostRepo_GetModel(t *testing.T) {
	t.Run("成功获取主机模型", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestHostRepo()
		ctx := context.Background()
		hostModel := &model.HostModel{
			Name:    "test-host",
			Label:   "test",
			SSHIP:   "192.168.1.1",
			SSHPort: 22,
			SSHUser: "admin",
			PyPath:  "/usr/bin/python3",
			Remark:  "测试主机",
		}

		// 先创建模型
		err := repo.CreateModel(ctx, hostModel)
		assert.NoError(t, err)

		// 执行查询操作
		result, err := repo.GetModel(ctx, []string{}, "id = ?", hostModel.ID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, hostModel.Name, result.Name)
		assert.Equal(t, hostModel.SSHIP, result.SSHIP)
		assert.Equal(t, hostModel.SSHPort, result.SSHPort)
	})
}

func TestHostRepo_ListModel(t *testing.T) {
	t.Run("成功获取主机模型列表", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestHostRepo()
		ctx := context.Background()

		// 创建多个测试模型
		for i := 1; i <= 3; i++ {
			hostModel := &model.HostModel{
				Name:    "test-host-" + string(rune('0'+i)),
				Label:   "test",
				SSHIP:   "192.168.1." + string(rune('0'+i)),
				SSHPort: 22,
				SSHUser: "admin",
				PyPath:  "/usr/bin/python3",
				Remark:  "测试主机 " + string(rune('0'+i)),
			}
			err := repo.CreateModel(ctx, hostModel)
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

func TestHostRepo_NewSSHClient(t *testing.T) {
	t.Run("上下文取消时创建SSH客户端失败", func(t *testing.T) {
		// 准备测试数据
		repo := NewTestHostRepo()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消上下文

		// 执行测试
		client, err := repo.NewSSHClient(ctx, "192.168.1.1", 22, "admin", nil, 5*time.Second)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}
