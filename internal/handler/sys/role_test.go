package sys

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	sysmodel "gin-artweb/internal/model/sys"
	sysrepo "gin-artweb/internal/repo/sys"
	syssvc "gin-artweb/internal/service/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/test"
)

func createTestRoleDTO() sysmodel.RoleUpsertDTO {
	return sysmodel.RoleUpsertDTO{
		Name:      fmt.Sprintf("test-role-%s", uuid.NewString()),
		Descr:     "Test role description",
		ApiIDs:    []uint32{},
		MenuIDs:   []uint32{},
		ButtonIDs: []uint32{},
	}
}

func createTestRoleModel() *sysmodel.RoleModel {
	return &sysmodel.RoleModel{
		Name:  fmt.Sprintf("test-role-%s", uuid.NewString()),
		Descr: "Test role description",
	}
}

type RoleHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	handler     *RoleHandler
	roleService *syssvc.RoleService
	roleRepo    *sysrepo.RoleRepo
	apiRepo     *sysrepo.ApiRepo
	menuRepo    *sysrepo.MenuRepo
	buttonRepo  *sysrepo.ButtonRepo
}

func (s *RoleHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.AutoMigrate(
		&sysmodel.RoleModel{},
		&sysmodel.ApiModel{},
		&sysmodel.MenuModel{},
		&sysmodel.ButtonModel{},
	), "数据库迁移失败")

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, err := auth.NewCasbinEnforcer()
	s.Require().NoError(err, "创建Casbin enforcer失败")

	s.apiRepo = sysrepo.NewApiRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.menuRepo = sysrepo.NewMenuRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.buttonRepo = sysrepo.NewButtonRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.roleRepo = sysrepo.NewRoleRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.roleService = syssvc.NewRoleService(logger, s.apiRepo, s.menuRepo, s.buttonRepo, s.roleRepo)
	s.handler = NewRoleHandler(logger, s.roleService)

	s.router = gin.Default()
	roleGroup := s.router.Group("/api/v1/customer")
	s.handler.LoadRouter(roleGroup)
}

func (s *RoleHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.Exec("DELETE FROM sys_role_api").Error, "清理角色API关联数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_role_menu").Error, "清理角色菜单关联数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_role_button").Error, "清理角色按钮关联数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_role").Error, "清理角色数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_api").Error, "清理API数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_menu").Error, "清理菜单数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_button").Error, "清理按钮数据失败")
}

func (s *RoleHandlerTestSuite) TestCreateRole_Success() {
	dto := createTestRoleDTO()
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("POST", "/api/v1/customer/role", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var resp sysmodel.RoleResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusCreated, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(dto.Name, resp.Data.Name)
	s.Equal(dto.Descr, resp.Data.Descr)
}

func (s *RoleHandlerTestSuite) TestCreateRole_InvalidRequest() {
	invalidDTO := map[string]any{
		"name": "",
	}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/customer/role", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusCreated, w.Code)
}

func (s *RoleHandlerTestSuite) TestGetRole_Success() {
	testID := uint32(uuid.New().ID())
	roleModel := createTestRoleModel()
	roleModel.ID = testID
	s.Require().NoError(s.roleRepo.CreateModel(context.Background(), roleModel, nil, nil, nil))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/customer/role/%d", testID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.RoleResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(testID, resp.Data.ID)
}

func (s *RoleHandlerTestSuite) TestGetRole_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/customer/role/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *RoleHandlerTestSuite) TestUpdateRole_Success() {
	testID := uint32(uuid.New().ID())
	roleModel := createTestRoleModel()
	roleModel.ID = testID
	s.Require().NoError(s.roleRepo.CreateModel(context.Background(), roleModel, nil, nil, nil))

	updateDTO := sysmodel.RoleUpsertDTO{
		Name:      "updated-role",
		Descr:     "Updated role description",
		ApiIDs:    []uint32{},
		MenuIDs:   []uint32{},
		ButtonIDs: []uint32{},
	}
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/customer/role/%d", testID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.RoleResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(updateDTO.Name, resp.Data.Name)
	s.Equal(updateDTO.Descr, resp.Data.Descr)
}

func (s *RoleHandlerTestSuite) TestUpdateRole_NotFound() {
	updateDTO := sysmodel.RoleUpsertDTO{
		Name:  "updated-role",
		Descr: "Updated role description",
	}
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", "/api/v1/customer/role/999999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *RoleHandlerTestSuite) TestDeleteRole_Success() {
	testID := uint32(uuid.New().ID())
	roleModel := createTestRoleModel()
	roleModel.ID = testID
	s.Require().NoError(s.roleRepo.CreateModel(context.Background(), roleModel, nil, nil, nil))

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/customer/role/%d", testID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	_, err := s.roleRepo.GetModel(context.Background(), nil, testID)
	s.Error(err)
}

func (s *RoleHandlerTestSuite) TestDeleteRole_NotFound() {
	req := httptest.NewRequest("DELETE", "/api/v1/customer/role/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *RoleHandlerTestSuite) TestListRole_Success() {
	for i := 0; i < 5; i++ {
		testID := uint32(uuid.New().ID())
		roleModel := createTestRoleModel()
		roleModel.ID = testID
		s.Require().NoError(s.roleRepo.CreateModel(context.Background(), roleModel, nil, nil, nil))
	}

	req := httptest.NewRequest("GET", "/api/v1/customer/role?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagRoleResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(5))
}

func (s *RoleHandlerTestSuite) TestListRole_WithFilter() {
	testID := uint32(uuid.New().ID())
	roleModel := createTestRoleModel()
	roleModel.ID = testID
	roleModel.Name = "unique-role-123"
	s.Require().NoError(s.roleRepo.CreateModel(context.Background(), roleModel, nil, nil, nil))

	req := httptest.NewRequest("GET", "/api/v1/customer/role?name=unique-role-123", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagRoleResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(1))
}

func TestRoleHandlerTestSuite(t *testing.T) {
	suite.Run(t, &RoleHandlerTestSuite{})
}

func TestNewRoleHandler(t *testing.T) {
	logger := test.NewTestZapLogger()
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, _ := auth.NewCasbinEnforcer()
	apiRepo := sysrepo.NewApiRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	menuRepo := sysrepo.NewMenuRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	buttonRepo := sysrepo.NewButtonRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	roleRepo := sysrepo.NewRoleRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	roleService := syssvc.NewRoleService(logger, apiRepo, menuRepo, buttonRepo, roleRepo)

	handler := NewRoleHandler(logger, roleService)

	if handler == nil {
		t.Fatal("NewRoleHandler should return non-nil handler")
	}
	if handler.log == nil {
		t.Fatal("Handler log should not be nil")
	}
	if handler.roleSvc == nil {
		t.Fatal("Handler roleSvc should not be nil")
	}
}
