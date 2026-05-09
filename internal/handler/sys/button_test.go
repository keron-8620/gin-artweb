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

func createTestButtonDTO(id, menuID uint32) sysmodel.CreateButtonDTO {
	return sysmodel.CreateButtonDTO{
		ID:       id,
		Name:     fmt.Sprintf("test-button-%s", uuid.NewString()),
		Sort:     100,
		IsActive: true,
		Descr:    "Test button description",
		MenuID:   menuID,
	}
}

func createTestButtonModel(id, menuID uint32) *sysmodel.ButtonModel {
	return &sysmodel.ButtonModel{
		Name:     fmt.Sprintf("test-button-%s", uuid.NewString()),
		Sort:     100,
		IsActive: true,
		Descr:    "Test button description",
		MenuID:   menuID,
	}
}

type ButtonHandlerTestSuite struct {
	suite.Suite
	router        *gin.Engine
	handler       *ButtonHandler
	buttonService *syssvc.ButtonService
	buttonRepo    *sysrepo.ButtonRepo
	apiRepo       *sysrepo.ApiRepo
	menuRepo      *sysrepo.MenuRepo
	testMenuID    uint32
}

func (s *ButtonHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.AutoMigrate(&sysmodel.ButtonModel{}, &sysmodel.MenuModel{}, &sysmodel.ApiModel{}), "数据库迁移失败")

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, err := auth.NewCasbinEnforcer()
	s.Require().NoError(err, "创建Casbin enforcer失败")

	s.apiRepo = sysrepo.NewApiRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.menuRepo = sysrepo.NewMenuRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.buttonRepo = sysrepo.NewButtonRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.buttonService = syssvc.NewButtonService(logger, s.apiRepo, s.menuRepo, s.buttonRepo)
	s.handler = NewButtonHandler(logger, s.buttonService)

	s.router = gin.Default()
	buttonGroup := s.router.Group("/api/v1/customer")
	s.handler.LoadRouter(buttonGroup)
}

func (s *ButtonHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.Exec("DELETE FROM sys_button_api").Error, "清理按钮API关联数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_button").Error, "清理按钮数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_menu").Error, "清理菜单数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_api").Error, "清理API数据失败")

	// 创建一个测试菜单用于测试按钮
	s.testMenuID = uint32(uuid.New().ID())
	testMenu := createTestMenuModel(s.testMenuID)
	testMenu.ID = s.testMenuID
	s.Require().NoError(s.menuRepo.CreateModel(context.Background(), testMenu, nil))
}

func (s *ButtonHandlerTestSuite) TestCreateButton_Success() {
	testID := uint32(uuid.New().ID())
	dto := createTestButtonDTO(testID, s.testMenuID)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("POST", "/api/v1/customer/button", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var resp sysmodel.ButtonResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusCreated, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(dto.Name, resp.Data.Name)
	s.Equal(dto.Sort, resp.Data.Sort)
	s.Equal(dto.IsActive, resp.Data.IsActive)
	s.Equal(dto.Descr, resp.Data.Descr)
	s.Equal(dto.MenuID, s.testMenuID)
}

func (s *ButtonHandlerTestSuite) TestCreateButton_InvalidRequest() {
	invalidDTO := map[string]any{
		"id": 0,
	}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/customer/button", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusCreated, w.Code)
}

func (s *ButtonHandlerTestSuite) TestGetButton_Success() {
	testID := uint32(uuid.New().ID())
	buttonModel := createTestButtonModel(testID, s.testMenuID)
	buttonModel.ID = testID
	s.Require().NoError(s.buttonRepo.CreateModel(context.Background(), buttonModel, nil))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/customer/button/%d", testID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.ButtonResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(testID, resp.Data.ID)
}

func (s *ButtonHandlerTestSuite) TestGetButton_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/customer/button/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *ButtonHandlerTestSuite) TestUpdateButton_Success() {
	testID := uint32(uuid.New().ID())
	buttonModel := createTestButtonModel(testID, s.testMenuID)
	buttonModel.ID = testID
	s.Require().NoError(s.buttonRepo.CreateModel(context.Background(), buttonModel, nil))

	updateDTO := sysmodel.UpdateButtonDTO{
		Name:     "updated-button",
		Sort:     200,
		IsActive: false,
		Descr:    "Updated button description",
		MenuID:   s.testMenuID,
	}
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/customer/button/%d", testID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.ButtonResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(updateDTO.Name, resp.Data.Name)
	s.Equal(updateDTO.Sort, resp.Data.Sort)
	s.Equal(updateDTO.IsActive, resp.Data.IsActive)
	s.Equal(updateDTO.Descr, resp.Data.Descr)
}

func (s *ButtonHandlerTestSuite) TestUpdateButton_NotFound() {
	updateDTO := sysmodel.UpdateButtonDTO{
		Name:     "updated-button",
		Sort:     200,
		IsActive: false,
		MenuID:   s.testMenuID,
	}
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", "/api/v1/customer/button/999999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *ButtonHandlerTestSuite) TestDeleteButton_Success() {
	testID := uint32(uuid.New().ID())
	buttonModel := createTestButtonModel(testID, s.testMenuID)
	buttonModel.ID = testID
	s.Require().NoError(s.buttonRepo.CreateModel(context.Background(), buttonModel, nil))

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/customer/button/%d", testID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	_, err := s.buttonRepo.GetModel(context.Background(), nil, testID)
	s.Error(err)
}

func (s *ButtonHandlerTestSuite) TestDeleteButton_NotFound() {
	req := httptest.NewRequest("DELETE", "/api/v1/customer/button/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *ButtonHandlerTestSuite) TestListButton_Success() {
	for i := 0; i < 5; i++ {
		testID := uint32(uuid.New().ID())
		buttonModel := createTestButtonModel(testID, s.testMenuID)
		buttonModel.ID = testID
		s.Require().NoError(s.buttonRepo.CreateModel(context.Background(), buttonModel, nil))
	}

	req := httptest.NewRequest("GET", "/api/v1/customer/button?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagButtonResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(5))
}

func (s *ButtonHandlerTestSuite) TestListButton_WithFilter() {
	testID := uint32(uuid.New().ID())
	buttonModel := createTestButtonModel(testID, s.testMenuID)
	buttonModel.ID = testID
	buttonModel.Name = "unique-button-123"
	s.Require().NoError(s.buttonRepo.CreateModel(context.Background(), buttonModel, nil))

	req := httptest.NewRequest("GET", "/api/v1/customer/button?name=unique-button-123", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagButtonResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(1))
}

func TestButtonHandlerTestSuite(t *testing.T) {
	suite.Run(t, &ButtonHandlerTestSuite{})
}

func TestNewButtonHandler(t *testing.T) {
	logger := test.NewTestZapLogger()
	db := test.NewTestGormDBWithConfig(nil)
	dbTimeout := test.NewTestDBTimeouts()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, _ := auth.NewCasbinEnforcer()
	apiRepo := sysrepo.NewApiRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	menuRepo := sysrepo.NewMenuRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	buttonRepo := sysrepo.NewButtonRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	buttonService := syssvc.NewButtonService(logger, apiRepo, menuRepo, buttonRepo)

	handler := NewButtonHandler(logger, buttonService)

	if handler == nil {
		t.Fatal("NewButtonHandler should return non-nil handler")
	}
	if handler.log == nil {
		t.Fatal("Handler log should not be nil")
	}
	if handler.buttonSvc == nil {
		t.Fatal("Handler buttonSvc should not be nil")
	}
}
