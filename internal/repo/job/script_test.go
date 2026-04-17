package job

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"emperror.dev/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	jobmodel "gin-artweb/internal/model/job"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/test"
)

func CreateTestScriptModel(isBuiltin bool) *jobmodel.ScriptModel {
	return &jobmodel.ScriptModel{
		Name:      fmt.Sprintf("test-%s.sh", uuid.NewString()),
		Descr:     "这是一个测试脚本",
		Project:   "test_project",
		Label:     "test_label",
		Language:  "bash",
		IsBuiltin: isBuiltin,
		Username:  "test_user",
	}
}

func CreateTestScriptModelWithProject(project string) *jobmodel.ScriptModel {
	return &jobmodel.ScriptModel{
		Name:      fmt.Sprintf("test-%s.sh", uuid.NewString()),
		Descr:     "这是一个测试脚本",
		Project:   project,
		Label:     "test_label",
		Language:  "bash",
		IsBuiltin: false,
		Username:  "test_user",
	}
}

func CreateTestScriptModelWithLabel(label string) *jobmodel.ScriptModel {
	return &jobmodel.ScriptModel{
		Name:      fmt.Sprintf("test-%s.sh", uuid.NewString()),
		Descr:     "这是一个测试脚本",
		Project:   "test_project",
		Label:     label,
		Language:  "bash",
		IsBuiltin: false,
		Username:  "test_user",
	}
}

type ScriptTestSuite struct {
	suite.Suite
	scriptRepo *ScriptRepo
}

func (suite *ScriptTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&jobmodel.ScriptModel{}); err != nil {
		suite.Error(err, "数据库迁移失败")
	}
	dbTimeout := test.NewTestDBTimeouts()
	dbSlowThreshold := test.NewTestDBSlowThreshold()
	logger := test.NewTestZapLogger()
	suite.scriptRepo = &ScriptRepo{
		log:           logger,
		gormDB:        db,
		timeouts:      dbTimeout,
		slowThreshold: dbSlowThreshold,
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
	err := suite.scriptRepo.CreateModel(context.Background(), nil)
	suite.Error(err, "创建脚本时传入nil应该返回错误")
}

func (suite *ScriptTestSuite) TestUpdateModel() {
	sm := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建脚本模型应该成功")

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
	err := suite.scriptRepo.UpdateModel(context.Background(), map[string]any{}, "id = ?", 1)
	suite.Error(err, "更新脚本时传入空数据应该返回错误")
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
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "应该返回记录未找到错误")
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

func (suite *ScriptTestSuite) TestGetModelWithEmptyConditions() {
	result, err := suite.scriptRepo.GetModel(context.Background())
	if err != nil {
		suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询时传入空条件应该返回记录未找到错误")
	} else {
		suite.NotNil(result, "查询时传入空条件应该返回有效的脚本模型")
		suite.Greater(result.ID, uint32(0), "返回的脚本模型ID应该大于0")
	}
}

func (suite *ScriptTestSuite) TestListModel() {
	for range 10 {
		sm := CreateTestScriptModel(false)
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	qp := database.QueryParams{
		Limit:  10,
		Offset: 0,
	}
	ms, err := suite.scriptRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列出脚本模型应该成功")
	suite.NotNil(ms, "脚本模型列表不应该为nil")
	suite.GreaterOrEqual(len(ms), 10, "脚本模型总数应该至少有10条")

	qpPaginated := database.QueryParams{
		Limit:  5,
		Offset: 0,
	}
	ms, err = suite.scriptRepo.ListModel(context.Background(), qpPaginated)
	suite.NoError(err, "分页列出脚本模型应该成功")
	suite.NotNil(ms, "分页脚本模型列表不应该为nil")
	suite.Equal(5, len(ms), "分页查询应该返回指定数量的记录")
}

func (suite *ScriptTestSuite) TestListModelWithEmptyParams() {
	qp := database.QueryParams{}
	ms, err := suite.scriptRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "列表查询时传入空参数应该成功")
	suite.NotNil(ms, "脚本模型列表不应该为nil")
	suite.GreaterOrEqual(int64(len(ms)), int64(0), "脚本模型总数应该大于等于0")
}

func (suite *ScriptTestSuite) TestListModelWithSorting() {
	for i := 0; i < 5; i++ {
		sm := &jobmodel.ScriptModel{
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

	qp := database.QueryParams{
		OrderBy: []string{"id DESC"},
	}
	ms, err := suite.scriptRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按ID降序排序查询应该成功")
	suite.NotNil(ms, "脚本模型列表不应该为nil")
	if len(ms) > 1 {
		prevID := ms[0].ID
		for _, script := range ms {
			suite.LessOrEqual(script.ID, prevID, "脚本应该按ID降序排序")
			prevID = script.ID
		}
	}
}

func (suite *ScriptTestSuite) TestListModelWithFiltering() {
	testLabel := "filter_test"
	sm := &jobmodel.ScriptModel{
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

	qp := database.QueryParams{
		Query: map[string]any{
			"label": testLabel,
		},
	}
	ms, err := suite.scriptRepo.ListModel(context.Background(), qp)
	suite.NoError(err, "按标签过滤查询应该成功")
	suite.NotNil(ms, "脚本模型列表不应该为nil")
	for _, script := range ms {
		suite.Equal(testLabel, script.Label, "脚本应该按标签过滤")
	}
}

func (suite *ScriptTestSuite) TestListProjectsAndLabels() {
	// 测试ListProjects
	projects := []string{"project1", "project2", "project3"}
	for _, project := range projects {
		sm := CreateTestScriptModelWithProject(project)
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	result, err := suite.scriptRepo.ListProjects(context.Background(), nil)
	suite.NoError(err, "查询项目名称应该成功")
	suite.NotNil(result, "项目名称列表不应该为nil")
	for _, project := range projects {
		suite.Contains(result, project, "项目名称列表应该包含所有创建的项目")
	}

	testProject := "filter_project"
	for i := 0; i < 3; i++ {
		sm := CreateTestScriptModelWithProject(testProject)
		sm.Label = fmt.Sprintf("label-%d", i)
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	query := map[string]any{
		"project": testProject,
	}
	result, err = suite.scriptRepo.ListProjects(context.Background(), query)
	suite.NoError(err, "带条件查询项目名称应该成功")
	suite.NotNil(result, "项目名称列表不应该为nil")
	suite.Equal(1, len(result), "带条件查询应该只返回匹配的项目")
	suite.Equal(testProject, result[0], "项目名称应该匹配查询条件")

	// 测试ListLabels
	labels := []string{"label1", "label2", "label3"}
	for _, label := range labels {
		sm := CreateTestScriptModelWithLabel(label)
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	result, err = suite.scriptRepo.ListLabels(context.Background(), nil)
	suite.NoError(err, "查询标签名称应该成功")
	suite.NotNil(result, "标签名称列表不应该为nil")
	for _, label := range labels {
		suite.Contains(result, label, "标签名称列表应该包含所有创建的标签")
	}

	testLabel := "filter_label"
	for i := 0; i < 3; i++ {
		sm := CreateTestScriptModelWithLabel(testLabel)
		sm.Project = fmt.Sprintf("project-%d", i)
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	query = map[string]any{
		"label": testLabel,
	}
	result, err = suite.scriptRepo.ListLabels(context.Background(), query)
	suite.NoError(err, "带条件查询标签名称应该成功")
	suite.NotNil(result, "标签名称列表不应该为nil")
	suite.Equal(1, len(result), "带条件查询应该只返回匹配的标签")
	suite.Equal(testLabel, result[0], "标签名称应该匹配查询条件")
}

func (suite *ScriptTestSuite) TestCountModel() {
	for i := 0; i < 5; i++ {
		sm := CreateTestScriptModel(false)
		err := suite.scriptRepo.CreateModel(context.Background(), sm)
		suite.NoError(err, "创建脚本模型应该成功")
	}

	count, err := suite.scriptRepo.CountModel(context.Background(), nil)
	suite.NoError(err, "查询脚本总数应该成功")
	suite.GreaterOrEqual(count, int64(5), "脚本总数应该至少为 5")

	query := map[string]any{
		"project": "test_project",
	}
	count, err = suite.scriptRepo.CountModel(context.Background(), query)
	suite.NoError(err, "带条件查询脚本总数应该成功")
	suite.GreaterOrEqual(count, int64(5), "带条件查询脚本总数应该至少为 5")

	query = map[string]any{
		"project": "non_existent_project",
	}
	count, err = suite.scriptRepo.CountModel(context.Background(), query)
	suite.NoError(err, "查询不存在的脚本总数应该成功")
	suite.Equal(int64(0), count, "查询不存在的脚本总数应该为 0")
}

func (suite *ScriptTestSuite) TestSaveScriptFile() {
	tempDir, err := os.MkdirTemp("", "script_test")
	suite.NoError(err, "创建临时目录应该成功")
	defer os.RemoveAll(tempDir)

	scriptContent := []byte("#!/bin/bash\necho 'Hello, World!'")
	scriptPath := filepath.Join(tempDir, "test_script.sh")

	err = suite.scriptRepo.SaveScriptFile(context.Background(), bytes.NewReader(scriptContent), scriptPath, false)
	suite.NoError(err, "保存脚本文件应该成功")

	fileInfo, err := os.Stat(scriptPath)
	suite.NoError(err, "验证脚本文件存在应该成功")
	suite.True(fileInfo.Mode().IsRegular(), "脚本文件应该是普通文件")

	content, err := os.ReadFile(scriptPath)
	suite.NoError(err, "读取脚本文件应该成功")
	suite.Equal(scriptContent, content, "脚本文件内容应该与写入的一致")

	err = suite.scriptRepo.SaveScriptFile(context.Background(), bytes.NewReader([]byte("#!/bin/bash\necho 'Updated'")), scriptPath, false)
	suite.Error(err, "保存脚本文件时文件已存在且不允许覆盖应该返回错误")

	err = suite.scriptRepo.SaveScriptFile(context.Background(), bytes.NewReader([]byte("#!/bin/bash\necho 'Updated'")), scriptPath, true)
	suite.NoError(err, "保存脚本文件时文件已存在但允许覆盖应该成功")

	content, err = os.ReadFile(scriptPath)
	suite.NoError(err, "读取更新后的脚本文件应该成功")
	suite.Equal([]byte("#!/bin/bash\necho 'Updated'"), content, "脚本文件内容应该已更新")

	invalidPath := filepath.Join("/root", "test_script.sh")
	err = suite.scriptRepo.SaveScriptFile(context.Background(), bytes.NewReader(scriptContent), invalidPath, false)
	suite.Error(err, "保存脚本文件时创建目录失败应该返回错误")
}

func (suite *ScriptTestSuite) TestRemoveScriptFile() {
	tempDir, err := os.MkdirTemp("", "script_test")
	suite.NoError(err, "创建临时目录应该成功")
	defer os.RemoveAll(tempDir)

	scriptContent := []byte("#!/bin/bash\necho 'Hello, World!'")
	scriptPath := filepath.Join(tempDir, "test_script.sh")

	err = suite.scriptRepo.SaveScriptFile(context.Background(), bytes.NewReader(scriptContent), scriptPath, false)
	suite.NoError(err, "保存脚本文件应该成功")

	err = suite.scriptRepo.RemoveScriptFile(context.Background(), scriptPath)
	suite.NoError(err, "删除脚本文件应该成功")

	_, err = os.Stat(scriptPath)
	suite.True(os.IsNotExist(err), "脚本文件应该已删除")

	err = suite.scriptRepo.RemoveScriptFile(context.Background(), scriptPath)
	suite.NoError(err, "删除不存在的脚本文件应该成功")

	invalidPath := filepath.Join("/root", "test_script.sh")
	err = suite.scriptRepo.RemoveScriptFile(context.Background(), invalidPath)
	suite.Error(err, "删除脚本文件时检查文件失败应该返回错误")

	err = os.Chmod(scriptPath, 0o444)
	if err == nil {
		err = suite.scriptRepo.RemoveScriptFile(context.Background(), scriptPath)
		suite.NoError(err, "删除脚本文件时文件权限不足应该成功")
	}
}

func (suite *ScriptTestSuite) TestOperationWithNonExistentRecord() {
	// 测试更新不存在的脚本
	err := suite.scriptRepo.UpdateModel(context.Background(), map[string]any{
		"name": "non-existent.sh",
	}, "id = ?", 999999)
	suite.NoError(err, "更新不存在的脚本不应该返回错误")

	// 测试删除不存在的脚本
	err = suite.scriptRepo.DeleteModel(context.Background(), "id = ?", 999999)
	suite.NoError(err, "删除不存在的脚本不应该返回错误")

	// 测试查询不存在的脚本
	_, err = suite.scriptRepo.GetModel(context.Background(), 999999)
	suite.True(errors.Is(err, gorm.ErrRecordNotFound), "查询不存在的脚本应该返回记录未找到错误")
}

func (suite *ScriptTestSuite) TestOperationWithContextTimeout() {
	// 创建一个脚本用于测试
	sm := CreateTestScriptModel(false)
	err := suite.scriptRepo.CreateModel(context.Background(), sm)
	suite.NoError(err, "创建脚本应该成功")

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()
	time.Sleep(time.Millisecond * 5)

	// 测试创建脚本时上下文超时
	sm2 := CreateTestScriptModel(false)
	err = suite.scriptRepo.CreateModel(ctx, sm2)
	suite.Error(err, "创建脚本时上下文超时应该返回错误")

	// 测试更新脚本时上下文超时
	err = suite.scriptRepo.UpdateModel(ctx, map[string]any{
		"name": "updated.sh",
	}, "id = ?", sm.ID)
	suite.Error(err, "更新脚本时上下文超时应该返回错误")

	// 测试删除脚本时上下文超时
	err = suite.scriptRepo.DeleteModel(ctx, "id = ?", sm.ID)
	suite.Error(err, "删除脚本时上下文超时应该返回错误")

	// 测试查询脚本时上下文已取消
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2()
	_, err = suite.scriptRepo.GetModel(ctx2, "id = ?", 1)
	suite.Error(err, "查询脚本时上下文已取消应该返回错误")

	// 测试列表查询时上下文超时
	qp := database.QueryParams{}
	ms, err := suite.scriptRepo.ListModel(ctx, qp)
	suite.Error(err, "列表查询时上下文超时应该返回错误")
	suite.Nil(ms, "超时查询脚本模型列表应该为nil")

	// 测试查询项目时上下文超时
	_, err = suite.scriptRepo.ListProjects(ctx, nil)
	suite.Error(err, "查询项目时上下文超时应该返回错误")

	// 测试查询标签时上下文超时
	_, err = suite.scriptRepo.ListLabels(ctx, nil)
	suite.Error(err, "查询标签时上下文超时应该返回错误")

	// 测试查询脚本总数时上下文超时
	_, err = suite.scriptRepo.CountModel(ctx, nil)
	suite.Error(err, "查询脚本总数时上下文超时应该返回错误")
}

func TestScriptTestSuite(t *testing.T) {
	pts := &ScriptTestSuite{}
	suite.Run(t, pts)
}
