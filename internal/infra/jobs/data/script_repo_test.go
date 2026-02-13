package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"emperror.dev/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/jobs/model"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestScriptModel(isBuiltin bool) *model.ScriptModel {
	return &model.ScriptModel{
		Name:      fmt.Sprintf("test-%s.sh", uuid.NewString()),
		Descr:     "这是一个测试脚本",
		Project:   "test_project",
		Label:     "test_label",
		Language:  "bash",
		IsBuiltin: isBuiltin,
		Username:  "test_user",
	}
}

type ScriptTestSuite struct {
	suite.Suite
	scriptRepo *ScriptRepo
}

func (suite *ScriptTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&model.ScriptModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.scriptRepo = &ScriptRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

func (suite *ScriptTestSuite) TestCreateModel() {
	sm := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建脚本模型应该成功")

	fm, err := suite.scriptRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的脚本模型应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(sm.Name, fm.Name)
	suite.Equal(sm.Descr, fm.Descr)
	suite.Equal(sm.Project, fm.Project)
	suite.Equal(sm.Label, fm.Label)
	suite.Equal(sm.Language, fm.Language)
	suite.Equal(sm.IsBuiltin, fm.IsBuiltin)
	suite.Equal(sm.Username, fm.Username)
}

func (suite *ScriptTestSuite) TestCreateModelWithNil() {
	// 测试创建脚本时传入空数据
	err := suite.scriptRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建脚本时传入nil应该返回错误")
}

func (suite *ScriptTestSuite) TestCreateModelWithContextTimeout() {
	// 测试创建脚本时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	sm := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(ctx, sm)
	suite.Error(err, "创建脚本时上下文超时应该返回错误")
}

func (suite *ScriptTestSuite) TestUpdateModel() {
	sm := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建脚本模型应该成功")

	// 准备更新数据
	updatedName := fmt.Sprintf("updated-%s.sh", uuid.NewString())
	updatedDescr := "这是一个更新后的测试脚本"
	updatedProject := "updated_project"
	updatedLabel := "updated_label"
	updatedLanguage := "python"
	updatedIsBuiltin := true
	updatedUsername := "updated_user"

	err = suite.scriptRepo.UpdateModel(context.Background(), map[string]any{
		"name":       updatedName,
		"descr":      updatedDescr,
		"project":    updatedProject,
		"label":      updatedLabel,
		"language":   updatedLanguage,
		"is_builtin": updatedIsBuiltin,
		"username":   updatedUsername,
	}, "id = ?", sm.ID)
	suite.NoError(err, "更新脚本模型应该成功")

	fm, err := suite.scriptRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "查询更新后的脚本模型应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(updatedName, fm.Name)
	suite.Equal(updatedDescr, fm.Descr)
	suite.Equal(updatedProject, fm.Project)
	suite.Equal(updatedLabel, fm.Label)
	suite.Equal(updatedLanguage, fm.Language)
	suite.Equal(updatedIsBuiltin, fm.IsBuiltin)
	suite.Equal(updatedUsername, fm.Username)
	suite.Greater(fm.UpdatedAt, sm.UpdatedAt)
}

func (suite *ScriptTestSuite) TestUpdateModelWithEmptyData() {
	// 测试更新时传入空数据
	err := suite.scriptRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", 1)
	suite.Error(err, "更新脚本时传入空数据应该返回错误")
}

func (suite *ScriptTestSuite) TestUpdateModelNonExistent() {
	// 测试更新不存在的脚本
	err := suite.scriptRepo.UpdateModel(context.Background(), map[string]any{
		"name": "non-existent.sh",
	}, "id = ?", 999999)
	suite.NoError(err, "更新不存在的脚本不应该返回错误")
}

func (suite *ScriptTestSuite) TestUpdateModelWithContextTimeout() {
	// 测试更新脚本时上下文超时
	// 先创建一个脚本
	sm := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建脚本应该成功")

	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试更新脚本
	err = suite.scriptRepo.UpdateModel(ctx, map[string]any{
		"name": "updated.sh",
	}, "id = ?", sm.ID)
	suite.Error(err, "更新脚本时上下文超时应该返回错误")
}

func (suite *ScriptTestSuite) TestDeleteModel() {
	sm := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建脚本模型应该成功")

	fm, err := suite.scriptRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "查询刚创建的脚本模型应该成功")
	suite.Equal(sm.ID, fm.ID)

	err = suite.scriptRepo.DeleteModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "删除脚本模型应该成功")

	_, err = suite.scriptRepo.GetModel(context.Background(), "id = ?", sm.ID)
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
	} else {
		suite.Fail("应该返回错误，但没有返回")
	}
}

func (suite *ScriptTestSuite) TestDeleteModelNonExistent() {
	// 测试删除不存在的脚本
	err := suite.scriptRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的脚本不应该返回错误")
}

func (suite *ScriptTestSuite) TestDeleteModelWithContextTimeout() {
	// 测试删除脚本时上下文超时
	// 先创建一个脚本
	sm := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建脚本应该成功")

	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试删除脚本
	err = suite.scriptRepo.DeleteModel(ctx, "id = ?", sm.ID)
	suite.Error(err, "删除脚本时上下文超时应该返回错误")
}

func (suite *ScriptTestSuite) TestGetModel() {
	sm := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建脚本模型应该成功")

	fm, err := suite.scriptRepo.GetModel(context.Background(), "id = ?", sm.ID)
	suite.NoError(err, "查询脚本模型应该成功")
	suite.Equal(sm.ID, fm.ID)
	suite.Equal(sm.Name, fm.Name)
	suite.Equal(sm.Descr, fm.Descr)
	suite.Equal(sm.Project, fm.Project)
	suite.Equal(sm.Label, fm.Label)
	suite.Equal(sm.Language, fm.Language)
	suite.Equal(sm.IsBuiltin, fm.IsBuiltin)
	suite.Equal(sm.Username, fm.Username)
}

func (suite *ScriptTestSuite) TestGetModelNonExistent() {
	// 测试查询不存在的脚本
	_, err := suite.scriptRepo.GetModel(context.Background(), 999999)
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询不存在的脚本应该返回记录未找到错误")
}

func (suite *ScriptTestSuite) TestGetModelWithEmptyConditions() {
	// 测试查询时传入空条件
	result, err := suite.scriptRepo.GetModel(context.Background())
	// 当传入空条件时，GetModel方法会尝试获取数据库中的第一条记录
	// 如果数据库为空，会返回record not found错误
	// 如果数据库不为空，会返回第一条记录
	if err != nil {
		// 如果返回错误，应该是record not found
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询时传入空条件应该返回记录未找到错误")
	} else {
		// 如果返回结果，应该是一个有效的脚本模型
		suite.NotNil(result, "查询时传入空条件应该返回有效的脚本模型")
		suite.Greater(result.ID, uint32(0), "返回的脚本模型ID应该大于0")
	}
}

func (suite *ScriptTestSuite) TestGetModelWithContextTimeout() {
	// 测试查询脚本时上下文已取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := suite.scriptRepo.GetModel(ctx, 1)
	suite.Error(err, "查询脚本时上下文已取消应该返回错误")
}

func (suite *ScriptTestSuite) TestListModel() {
	// 清理可能存在的数据并创建测试数据
	for range 10 {
		sm := CreateTestScriptModel(false)
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	qp := database.QueryParams{
		Size:    10,
		Page:    0,
		IsCount: true,
	}
	total, ms, err := suite.scriptRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列出脚本模型应该成功")
	suite.NotNil(ms, "脚本模型列表不应该为nil")
	suite.GreaterOrEqual(total, int64(10), "脚本模型总数应该至少有10条")

	qpPaginated := database.QueryParams{
		Size:    5,
		Page:    0,
		IsCount: true,
	}
	pTotal, pMs, err := suite.scriptRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页列出脚本模型应该成功")
	suite.NotNil(pMs, "分页脚本模型列表不应该为nil")
	suite.Equal(5, len(*pMs), "分页查询应该返回指定数量的记录")
	suite.GreaterOrEqual(pTotal, int64(5), "分页总数应该至少等于limit")
}

func (suite *ScriptTestSuite) TestListModelWithEmptyParams() {
	// 测试列表查询时传入空参数
	qp := database.QueryParams{}
	total, ms, err := suite.scriptRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列表查询时传入空参数应该成功")
	suite.NotNil(ms, "脚本模型列表不应该为nil")
	suite.GreaterOrEqual(total, int64(0), "脚本模型总数应该大于等于0")
}

func (suite *ScriptTestSuite) TestListModelWithSorting() {
	// 测试列表查询时传入排序参数
	// 先创建多个脚本用于测试
	for i := 0; i < 5; i++ {
		sm := &model.ScriptModel{
			Name:      fmt.Sprintf("test-sort-%d.sh", i),
			Descr:     "这是一个测试脚本",
			Project:   "test_project",
			Label:     "test_label",
			Language:  "bash",
			IsBuiltin: false,
			Username:  "test_user",
		}
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本应该成功")
	}

	// 测试按ID降序排序
	qp := database.QueryParams{
		OrderBy: []string{"id DESC"},
	}
	_, ms, err := suite.scriptRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按ID降序排序查询应该成功")
	suite.NotNil(ms, "脚本模型列表不应该为nil")
	if len(*ms) > 1 {
		// 验证排序结果
		prevID := (*ms)[0].ID
		for _, script := range *ms {
			suite.LessOrEqual(script.ID, prevID, "脚本应该按ID降序排序")
			prevID = script.ID
		}
	}
}

func (suite *ScriptTestSuite) TestListModelWithFiltering() {
	// 测试列表查询时传入过滤参数
	// 创建一个特定标签的脚本
	testLabel := "filter_test"
	sm := &model.ScriptModel{
		Name:      "test-filter.sh",
		Descr:     "这是一个用于过滤测试的脚本",
		Project:   "test_project",
		Label:     testLabel,
		Language:  "bash",
		IsBuiltin: false,
		Username:  "test_user",
	}
	err := suite.scriptRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建脚本应该成功")

	// 测试按标签过滤
	qp := database.QueryParams{
		Query: map[string]any{
			"label": testLabel,
		},
	}
	_, ms, err := suite.scriptRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按标签过滤查询应该成功")
	suite.NotNil(ms, "脚本模型列表不应该为nil")
	// 验证过滤结果
	for _, script := range *ms {
		suite.Equal(testLabel, script.Label, "脚本应该按标签过滤")
	}
}

func (suite *ScriptTestSuite) TestListModelWithContextTimeout() {
	// 测试列表查询时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试列表查询
	qp := database.QueryParams{}
	_, _, err := suite.scriptRepo.ListModel(ctx, qp)
	suite.Error(err, "列表查询时上下文超时应该返回错误")
}

func (suite *ScriptTestSuite) TestListProjects() {
	// 创建不同项目的脚本
	projects := []string{"project1", "project2", "project3"}
	for _, project := range projects {
		sm := &model.ScriptModel{
			Name:      fmt.Sprintf("test-%s.sh", uuid.NewString()),
			Descr:     "这是一个测试脚本",
			Project:   project,
			Label:     "test_label",
			Language:  "bash",
			IsBuiltin: false,
			Username:  "test_user",
		}
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	// 测试查询所有项目
	result, err := suite.scriptRepo.ListProjects(context.Background(), nil)
	suite.NoError(err, "查询项目名称应该成功")
	suite.NotNil(result, "项目名称列表不应该为nil")
	// 验证所有项目都在结果中
	for _, project := range projects {
		suite.Contains(result, project, "项目名称列表应该包含所有创建的项目")
	}
}

func (suite *ScriptTestSuite) TestListProjectsWithQuery() {
	// 创建不同项目和标签的脚本
	testProject := "filter_project"
	for i := 0; i < 3; i++ {
		sm := &model.ScriptModel{
			Name:      fmt.Sprintf("test-%s.sh", uuid.NewString()),
			Descr:     "这是一个测试脚本",
			Project:   testProject,
			Label:     fmt.Sprintf("label-%d", i),
			Language:  "bash",
			IsBuiltin: false,
			Username:  "test_user",
		}
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	// 测试带条件查询项目
	query := map[string]any{
		"project": testProject,
	}
	result, err := suite.scriptRepo.ListProjects(context.Background(), query)
	suite.NoError(err, "带条件查询项目名称应该成功")
	suite.NotNil(result, "项目名称列表不应该为nil")
	suite.Equal(1, len(result), "带条件查询应该只返回匹配的项目")
	suite.Equal(testProject, result[0], "项目名称应该匹配查询条件")
}

func (suite *ScriptTestSuite) TestListProjectsWithContextTimeout() {
	// 测试查询项目时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试查询项目
	_, err := suite.scriptRepo.ListProjects(ctx, nil)
	suite.Error(err, "查询项目时上下文超时应该返回错误")
}

func (suite *ScriptTestSuite) TestListLabels() {
	// 创建不同标签的脚本
	labels := []string{"label1", "label2", "label3"}
	for _, label := range labels {
		sm := &model.ScriptModel{
			Name:      fmt.Sprintf("test-%s.sh", uuid.NewString()),
			Descr:     "这是一个测试脚本",
			Project:   "test_project",
			Label:     label,
			Language:  "bash",
			IsBuiltin: false,
			Username:  "test_user",
		}
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	// 测试查询所有标签
	result, err := suite.scriptRepo.ListLabels(context.Background(), nil)
	suite.NoError(err, "查询标签名称应该成功")
	suite.NotNil(result, "标签名称列表不应该为nil")
	// 验证所有标签都在结果中
	for _, label := range labels {
		suite.Contains(result, label, "标签名称列表应该包含所有创建的标签")
	}
}

func (suite *ScriptTestSuite) TestListLabelsWithQuery() {
	// 创建不同标签的脚本
	testLabel := "filter_label"
	for i := 0; i < 3; i++ {
		sm := &model.ScriptModel{
			Name:      fmt.Sprintf("test-%s.sh", uuid.NewString()),
			Descr:     "这是一个测试脚本",
			Project:   fmt.Sprintf("project-%d", i),
			Label:     testLabel,
			Language:  "bash",
			IsBuiltin: false,
			Username:  "test_user",
		}
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	// 测试带条件查询标签
	query := map[string]any{
		"label": testLabel,
	}
	result, err := suite.scriptRepo.ListLabels(context.Background(), query)
	suite.NoError(err, "带条件查询标签名称应该成功")
	suite.NotNil(result, "标签名称列表不应该为nil")
	suite.Equal(1, len(result), "带条件查询应该只返回匹配的标签")
	suite.Equal(testLabel, result[0], "标签名称应该匹配查询条件")
}

func (suite *ScriptTestSuite) TestListLabelsWithContextTimeout() {
	// 测试查询标签时上下文超时
	// 创建一个非常短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	// 等待超时
	time.Sleep(time.Millisecond * 5)

	// 尝试查询标签
	_, err := suite.scriptRepo.ListLabels(ctx, nil)
	suite.Error(err, "查询标签时上下文超时应该返回错误")
}

// 每个测试文件都需要这个入口函数
func TestScriptTestSuite(t *testing.T) {
	pts := &ScriptTestSuite{}
	suite.Run(t, pts)
}
