package data

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"gin-artweb/internal/infra/jobs/model"
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
