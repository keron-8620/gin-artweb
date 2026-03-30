package job

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	jobmodel "gin-artweb/internal/model/job"
	jobrepo "gin-artweb/internal/repo/job"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/test"
)

func CreateTestUploadScriptBiz() jobmodel.UploadScriptBiz {
	return jobmodel.UploadScriptBiz{
		Filename:  fmt.Sprintf("test_script_%s.sh", uuid.NewString()),
		Descr:     "测试脚本",
		Project:   "test_project",
		Label:     "test_label",
		Language:  "bash",
		Status:    true,
		File:      strings.NewReader("echo 'Hello World'"),
	}
}

// createScriptTestContext 创建带有测试JWT claims的上下文
func createScriptTestContext() context.Context {
	ctx := context.Background()
	claims := &auth.JwtClaims{
		UserInfo: auth.UserInfo{
			UserID:   1,
			Username: "test_user",
			RoleID:   1,
			IsStaff:  true,
		},
	}
	return ctxutil.SetJwtClaims(ctx, claims)
}



type ScriptTestSuite struct {
	suite.Suite
	scriptService *ScriptService
}

func (suite *ScriptTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&jobmodel.ScriptModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	suite.scriptService = NewScriptService(
		logger,
		jobrepo.NewScriptRepo(
			logger,
			db,
			dbTimeout,
		),
	)
}

func (suite *ScriptTestSuite) TestCreateScript() {
	dto := CreateTestUploadScriptBiz()
	ctx := createScriptTestContext()
	
	fm, err := suite.scriptService.CreateScript(ctx, dto)
	suite.Nil(err, "创建脚本应该成功")
	suite.Greater(fm.ID, uint32(0), "脚本ID应该大于0")
	suite.Equal(dto.Filename, fm.Name)
	suite.Equal(dto.Descr, fm.Descr)
	suite.Equal(dto.Project, fm.Project)
	suite.Equal(dto.Label, fm.Label)
	suite.Equal(dto.Language, fm.Language)
	suite.Equal(dto.Status, fm.Status)
	suite.False(fm.IsBuiltin, "脚本不应该是内置的")
}

func (suite *ScriptTestSuite) TestFindScriptByID() {
	dto := CreateTestUploadScriptBiz()
	ctx := createScriptTestContext()
	
	// 先创建一个脚本
	fm, err := suite.scriptService.CreateScript(ctx, dto)
	suite.Nil(err, "创建脚本应该成功")
	suite.Greater(fm.ID, uint32(0), "脚本ID应该大于0")
	
	// 然后查询
	foundFm, err := suite.scriptService.FindScriptByID(ctx, fm.ID)
	suite.Nil(err, "查询脚本应该成功")
	suite.Greater(foundFm.ID, uint32(0), "脚本ID应该大于0")
	suite.Equal(fm.Name, foundFm.Name)
	suite.Equal(fm.Descr, foundFm.Descr)
	suite.Equal(fm.Project, foundFm.Project)
	suite.Equal(fm.Label, foundFm.Label)
	suite.Equal(fm.Language, foundFm.Language)
	suite.Equal(fm.Status, foundFm.Status)
}

func (suite *ScriptTestSuite) TestDeleteScriptByID() {
	dto := CreateTestUploadScriptBiz()
	ctx := createScriptTestContext()
	
	// 先创建一个脚本
	fm, err := suite.scriptService.CreateScript(ctx, dto)
	suite.Nil(err, "创建脚本应该成功")
	suite.Greater(fm.ID, uint32(0), "脚本ID应该大于0")
	
	// 然后删除
	err = suite.scriptService.DeleteScriptByID(ctx, fm.ID)
	suite.Nil(err, "删除脚本应该成功")
	
	// 再次查询应该失败
	_, err = suite.scriptService.FindScriptByID(ctx, fm.ID)
	suite.NotNil(err, "查询已删除的脚本应该失败")
}

func (suite *ScriptTestSuite) TestListScript() {
	// 创建多个脚本
	scriptCount := 3
	ctx := createScriptTestContext()
	
	for i := 0; i < scriptCount; i++ {
		dto := CreateTestUploadScriptBiz()
		_, err := suite.scriptService.CreateScript(ctx, dto)
		suite.Nil(err, "创建脚本应该成功")
	}
	
	// 测试列出所有脚本
	listDTO := jobmodel.ListScriptDTO{}
	page, size := 1, 10
	count, scriptList, err := suite.scriptService.ListScript(ctx, page, size, listDTO)
	suite.Nil(err, "列出脚本应该成功")
	suite.GreaterOrEqual(int(count), scriptCount, "返回的脚本数量应该大于等于创建的数量")
	suite.NotNil(scriptList, "返回的脚本列表不应该为nil")
}

func (suite *ScriptTestSuite) TestListScriptsByIDs() {
	ctx := createScriptTestContext()
	
	// 创建多个脚本
	scriptCount := 2
	createdScripts := make([]uint32, 0, scriptCount)
	
	for i := 0; i < scriptCount; i++ {
		dto := CreateTestUploadScriptBiz()
		fm, err := suite.scriptService.CreateScript(ctx, dto)
		suite.Nil(err, "创建脚本应该成功")
		createdScripts = append(createdScripts, fm.ID)
	}
	
	// 测试通过ID列表查询
	scriptList, err := suite.scriptService.ListScriptsByIDs(ctx, createdScripts)
	suite.Nil(err, "通过ID列表查询脚本应该成功")
	suite.Len(scriptList, scriptCount, "返回的脚本数量应该等于创建的数量")
	
	// 测试空ID列表
	scriptList, err = suite.scriptService.ListScriptsByIDs(ctx, []uint32{})
	suite.Nil(err, "空ID列表查询应该成功")
	suite.Len(scriptList, 0, "空ID列表应该返回空结果")
}

func (suite *ScriptTestSuite) TestListProjects() {
	dto := CreateTestUploadScriptBiz()
	ctx := createScriptTestContext()
	
	// 创建一个脚本
	_, err := suite.scriptService.CreateScript(ctx, dto)
	suite.Nil(err, "创建脚本应该成功")
	
	// 测试列出项目
	listDTO := jobmodel.ListScriptDTO{}
	projects, err := suite.scriptService.ListProjects(ctx, listDTO)
	suite.Nil(err, "列出项目应该成功")
	suite.NotEmpty(projects, "项目列表不应该为空")
}

func (suite *ScriptTestSuite) TestListLabels() {
	dto := CreateTestUploadScriptBiz()
	ctx := createScriptTestContext()
	
	// 创建一个脚本
	_, err := suite.scriptService.CreateScript(ctx, dto)
	suite.Nil(err, "创建脚本应该成功")
	
	// 测试列出标签
	listDTO := jobmodel.ListScriptDTO{}
	labels, err := suite.scriptService.ListLabels(ctx, listDTO)
	suite.Nil(err, "列出标签应该成功")
	suite.NotEmpty(labels, "标签列表不应该为空")
}

func (suite *ScriptTestSuite) TestCreateScript_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// 尝试使用已取消的上下文创建脚本
	dto := CreateTestUploadScriptBiz()
	_, err := suite.scriptService.CreateScript(ctx, dto)
	suite.NotNil(err, "上下文错误时创建脚本应该失败")
}

func (suite *ScriptTestSuite) TestFindScriptByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// 尝试使用已取消的上下文查找脚本
	_, err := suite.scriptService.FindScriptByID(ctx, 1)
	suite.NotNil(err, "上下文错误时查找脚本应该失败")
}

func (suite *ScriptTestSuite) TestDeleteScriptByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// 尝试使用已取消的上下文删除脚本
	err := suite.scriptService.DeleteScriptByID(ctx, 1)
	suite.NotNil(err, "上下文错误时删除脚本应该失败")
}

func (suite *ScriptTestSuite) TestListScript_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// 尝试使用已取消的上下文列出脚本
	listDTO := jobmodel.ListScriptDTO{}
	_, _, err := suite.scriptService.ListScript(ctx, 1, 10, listDTO)
	suite.NotNil(err, "上下文错误时列出脚本应该失败")
}

func (suite *ScriptTestSuite) TestListScriptsByIDs_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// 尝试使用已取消的上下文通过ID列表查询脚本
	_, err := suite.scriptService.ListScriptsByIDs(ctx, []uint32{1})
	suite.NotNil(err, "上下文错误时通过ID列表查询脚本应该失败")
}

func (suite *ScriptTestSuite) TestListProjects_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// 尝试使用已取消的上下文列出项目
	listDTO := jobmodel.ListScriptDTO{}
	_, err := suite.scriptService.ListProjects(ctx, listDTO)
	suite.NotNil(err, "上下文错误时列出项目应该失败")
}

func (suite *ScriptTestSuite) TestListLabels_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// 尝试使用已取消的上下文列出标签
	listDTO := jobmodel.ListScriptDTO{}
	_, err := suite.scriptService.ListLabels(ctx, listDTO)
	suite.NotNil(err, "上下文错误时列出标签应该失败")
}

func (suite *ScriptTestSuite) TestUpdateScriptByID() {
	// 先创建一个脚本
	dto := CreateTestUploadScriptBiz()
	ctx := createScriptTestContext()
	fm, err := suite.scriptService.CreateScript(ctx, dto)
	suite.Nil(err, "创建脚本应该成功")
	suite.Greater(fm.ID, uint32(0), "脚本ID应该大于0")
	
	// 准备更新数据
	updateDTO := jobmodel.UploadScriptBiz{
		Filename:  fmt.Sprintf("updated_test_script_%s.sh", uuid.NewString()),
		Descr:     "更新后的测试脚本",
		Project:   "updated_test_project",
		Label:     "updated_test_label",
		Language:  "python",
		Status:    false,
		File:      strings.NewReader("print('Hello Updated World')"),
	}
	
	// 执行更新
	updatedFm, err := suite.scriptService.UpdateScriptByID(ctx, fm.ID, updateDTO)
	suite.Nil(err, "更新脚本应该成功")
	suite.Greater(updatedFm.ID, uint32(0), "脚本ID应该大于0")
	suite.Equal(updateDTO.Filename, updatedFm.Name)
	suite.Equal(updateDTO.Descr, updatedFm.Descr)
	suite.Equal(updateDTO.Project, updatedFm.Project)
	suite.Equal(updateDTO.Label, updatedFm.Label)
	suite.Equal(updateDTO.Language, updatedFm.Language)
	suite.Equal(updateDTO.Status, updatedFm.Status)
}

func (suite *ScriptTestSuite) TestUpdateScriptByID_ContextError() {
	// 创建一个可取消的上下文并立即取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 尝试使用已取消的上下文更新脚本
	dto := CreateTestUploadScriptBiz()
	_, err := suite.scriptService.UpdateScriptByID(ctx, 1, dto)
	suite.NotNil(err, "上下文错误时更新脚本应该失败")
}

func (suite *ScriptTestSuite) TestUpdateScriptByID_BuiltinScript() {
	// 先创建一个内置脚本
	builtinScript := &jobmodel.ScriptModel{
		Name:      "builtin_script.sh",
		Descr:     "Builtin script",
		Project:   "test_project",
		Label:     "test_label",
		Language:  "bash",
		Status:    true,
		IsBuiltin: true,
		Username:  "test_user",
	}
	err := suite.scriptService.scriptRepo.CreateModel(context.Background(), builtinScript)
	suite.Nil(err, "创建内置脚本应该成功")

	// 尝试更新内置脚本
	ctx := createScriptTestContext()
	updateDTO := CreateTestUploadScriptBiz()
	_, err = suite.scriptService.UpdateScriptByID(ctx, builtinScript.ID, updateDTO)
	suite.NotNil(err, "更新内置脚本应该失败")
	suite.Contains(err.Error(), "SCRIPT_IS_BUILTIN", "错误信息应该提到内置脚本")
}

func (suite *ScriptTestSuite) TestDeleteScriptByID_BuiltinScript() {
	// 先创建一个内置脚本
	builtinScript := &jobmodel.ScriptModel{
		Name:      "builtin_script_delete.sh",
		Descr:     "Builtin script for delete test",
		Project:   "test_project",
		Label:     "test_label",
		Language:  "bash",
		Status:    true,
		IsBuiltin: true,
		Username:  "test_user",
	}
	err := suite.scriptService.scriptRepo.CreateModel(context.Background(), builtinScript)
	suite.Nil(err, "创建内置脚本应该成功")

	// 尝试删除内置脚本
	ctx := createScriptTestContext()
	err = suite.scriptService.DeleteScriptByID(ctx, builtinScript.ID)
	suite.NotNil(err, "删除内置脚本应该失败")
	suite.Contains(err.Error(), "SCRIPT_IS_BUILTIN", "错误信息应该提到内置脚本")
}

func (suite *ScriptTestSuite) TestUpdateScriptByID_NotFound() {
	ctx := createScriptTestContext()

	updateDTO := CreateTestUploadScriptBiz()
	_, err := suite.scriptService.UpdateScriptByID(ctx, 99999, updateDTO)
	suite.NotNil(err, "更新不存在的脚本应该失败")
}

func (suite *ScriptTestSuite) TestDeleteScriptByID_NotFound() {
	ctx := createScriptTestContext()

	err := suite.scriptService.DeleteScriptByID(ctx, 99999)
	suite.NotNil(err, "删除不存在的脚本应该失败")
}

func (suite *ScriptTestSuite) TestListScript_EmptyResult() {
	ctx := createScriptTestContext()

	listDTO := jobmodel.ListScriptDTO{Project: "non_existent_project"}
	page, size := 1, 10
	count, scriptList, err := suite.scriptService.ListScript(ctx, page, size, listDTO)
	suite.Nil(err, "列出脚本应该成功")
	suite.Equal(int64(0), count, "Count should be 0 for non-existent project")
	suite.Nil(scriptList, "Script list should be nil for empty result")
}

func (suite *ScriptTestSuite) TestGetScriptStoragePath() {
	pathBuiltin := GetScriptStoragePath("test_project", "test_label", "test_script.sh", true)
	suite.NotEmpty(pathBuiltin)
	suite.Contains(pathBuiltin, "resource", "内置脚本路径应包含resource目录")

	pathNonBuiltin := GetScriptStoragePath("test_project", "test_label", "test_script.sh", false)
	suite.NotEmpty(pathNonBuiltin)
	suite.Contains(pathNonBuiltin, "storage", "非内置脚本路径应包含storage目录")
}

func TestScriptTestSuite(t *testing.T) {
	pts := &ScriptTestSuite{}
	suite.Run(t, pts)
}
