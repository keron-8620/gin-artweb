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

func createTestMenuDTO(id uint32) sysmodel.CreateMenuDTO {
	return sysmodel.CreateMenuDTO{
		ID:        id,
		Path:      fmt.Sprintf("/menu/%s", uuid.NewString()),
		Component: "/components/test",
		Name:      fmt.Sprintf("test-menu-%s", uuid.NewString()),
		Meta: sysmodel.MetaSchemas{
			Title: "Test Menu",
			Icon:  "test-icon",
		},
		Sort:     100,
		IsActive: true,
		Descr:    "Test menu description",
	}
}

func createTestMenuModel(id uint32) *sysmodel.MenuModel {
	return &sysmodel.MenuModel{
		Path:      fmt.Sprintf("/menu/%s", uuid.NewString()),
		Component: "/components/test",
		Name:      fmt.Sprintf("test-menu-%s", uuid.NewString()),
		Meta: sysmodel.MetaSchemas{
			Title: "Test Menu",
			Icon:  "test-icon",
		},
		Sort:     100,
		IsActive: true,
		Descr:    "Test menu description",
	}
}

type MenuHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	handler     *MenuHandler
	menuService *syssvc.MenuService
	menuRepo    *sysrepo.MenuRepo
	apiRepo     *sysrepo.ApiRepo
}

func (s *MenuHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.AutoMigrate(&sysmodel.MenuModel{}, &sysmodel.ApiModel{}), "数据库迁移失败")

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, err := auth.NewCasbinEnforcer()
	s.Require().NoError(err, "创建Casbin enforcer失败")

	s.apiRepo = sysrepo.NewApiRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.menuRepo = sysrepo.NewMenuRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.menuService = syssvc.NewMenuService(logger, s.apiRepo, s.menuRepo)
	s.handler = NewMenuHandler(logger, s.menuService)

	s.router = gin.Default()
	menuGroup := s.router.Group("/api/v1/customer")
	s.handler.LoadRouter(menuGroup)
}

func (s *MenuHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.Exec("DELETE FROM sys_menu_api").Error, "清理菜单API关联数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_api").Error, "清理API数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_menu").Error, "清理菜单数据失败")
}

func (s *MenuHandlerTestSuite) TestCreateMenu_Success() {
	testID := uint32(uuid.New().ID())
	dto := createTestMenuDTO(testID)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("POST", "/api/v1/customer/menu", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var resp sysmodel.MenuResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusCreated, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(dto.Path, resp.Data.Path)
	s.Equal(dto.Component, resp.Data.Component)
	s.Equal(dto.Name, resp.Data.Name)
	s.Equal(dto.Meta.Title, resp.Data.Meta.Title)
	s.Equal(dto.Meta.Icon, resp.Data.Meta.Icon)
	s.Equal(dto.Sort, resp.Data.Sort)
	s.Equal(dto.IsActive, resp.Data.IsActive)
	s.Equal(dto.Descr, resp.Data.Descr)
}

func (s *MenuHandlerTestSuite) TestCreateMenu_InvalidRequest() {
	invalidDTO := map[string]any{
		"id": 0,
	}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/customer/menu", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusCreated, w.Code)
}

func (s *MenuHandlerTestSuite) TestGetMenu_Success() {
	testID := uint32(uuid.New().ID())
	menuModel := createTestMenuModel(testID)
	menuModel.ID = testID
	s.Require().NoError(s.menuRepo.CreateModel(context.Background(), menuModel, nil))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/customer/menu/%d", testID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.MenuResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(testID, resp.Data.ID)
}

func (s *MenuHandlerTestSuite) TestGetMenu_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/customer/menu/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MenuHandlerTestSuite) TestUpdateMenu_Success() {
	testID := uint32(uuid.New().ID())
	menuModel := createTestMenuModel(testID)
	menuModel.ID = testID
	s.Require().NoError(s.menuRepo.CreateModel(context.Background(), menuModel, nil))

	updateDTO := sysmodel.UpdateMenuDTO{
		Path:      "/menu/updated",
		Component: "/components/updated",
		Name:      "updated-menu",
		Meta: sysmodel.MetaSchemas{
			Title: "Updated Menu",
			Icon:  "updated-icon",
		},
		Sort:     200,
		IsActive: false,
		Descr:    "Updated menu description",
	}
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/customer/menu/%d", testID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.MenuResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(updateDTO.Path, resp.Data.Path)
	s.Equal(updateDTO.Component, resp.Data.Component)
	s.Equal(updateDTO.Name, resp.Data.Name)
	s.Equal(updateDTO.Meta.Title, resp.Data.Meta.Title)
	s.Equal(updateDTO.Meta.Icon, resp.Data.Meta.Icon)
	s.Equal(updateDTO.Sort, resp.Data.Sort)
	s.Equal(updateDTO.IsActive, resp.Data.IsActive)
	s.Equal(updateDTO.Descr, resp.Data.Descr)
}

func (s *MenuHandlerTestSuite) TestUpdateMenu_NotFound() {
	updateDTO := sysmodel.UpdateMenuDTO{
		Path:      "/menu/updated",
		Component: "/components/updated",
		Name:      "updated-menu",
		Meta: sysmodel.MetaSchemas{
			Title: "Updated Menu",
			Icon:  "updated-icon",
		},
		Sort: 200,
	}
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", "/api/v1/customer/menu/999999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MenuHandlerTestSuite) TestDeleteMenu_Success() {
	testID := uint32(uuid.New().ID())
	menuModel := createTestMenuModel(testID)
	menuModel.ID = testID
	s.Require().NoError(s.menuRepo.CreateModel(context.Background(), menuModel, nil))

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/customer/menu/%d", testID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	_, err := s.menuRepo.GetModel(context.Background(), nil, testID)
	s.Error(err)
}

func (s *MenuHandlerTestSuite) TestDeleteMenu_NotFound() {
	req := httptest.NewRequest("DELETE", "/api/v1/customer/menu/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *MenuHandlerTestSuite) TestListMenu_Success() {
	for i := 0; i < 5; i++ {
		testID := uint32(uuid.New().ID())
		menuModel := createTestMenuModel(testID)
		menuModel.ID = testID
		s.Require().NoError(s.menuRepo.CreateModel(context.Background(), menuModel, nil))
	}

	req := httptest.NewRequest("GET", "/api/v1/customer/menu?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagMenuResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(5))
}

func (s *MenuHandlerTestSuite) TestListMenu_WithFilter() {
	testID := uint32(uuid.New().ID())
	menuModel := createTestMenuModel(testID)
	menuModel.ID = testID
	menuModel.Name = "unique-menu-123"
	s.Require().NoError(s.menuRepo.CreateModel(context.Background(), menuModel, nil))

	req := httptest.NewRequest("GET", "/api/v1/customer/menu?name=unique-menu-123", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagMenuResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(1))
}

func TestMenuHandlerTestSuite(t *testing.T) {
	suite.Run(t, &MenuHandlerTestSuite{})
}

func TestNewMenuHandler(t *testing.T) {
	logger := test.NewTestZapLogger()
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, _ := auth.NewCasbinEnforcer()
	apiRepo := sysrepo.NewApiRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	menuRepo := sysrepo.NewMenuRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	menuService := syssvc.NewMenuService(logger, apiRepo, menuRepo)

	handler := NewMenuHandler(logger, menuService)

	if handler == nil {
		t.Fatal("NewMenuHandler should return non-nil handler")
	}
	if handler.log == nil {
		t.Fatal("Handler log should not be nil")
	}
	if handler.menuSvc == nil {
		t.Fatal("Handler menuSvc should not be nil")
	}
}
