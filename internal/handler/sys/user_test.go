package sys

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	sysmodel "gin-artweb/internal/model/sys"
	sysrepo "gin-artweb/internal/repo/sys"
	syssvc "gin-artweb/internal/service/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/test"
	"gin-artweb/pkg/crypto"
)

func createTestUserHandlerRoleModel() *sysmodel.RoleModel {
	return &sysmodel.RoleModel{
		Name:  fmt.Sprintf("test-role-%s", uuid.NewString()),
		Descr: "Test role description",
	}
}

func createTestCreateUserDTO(roleID uint32) sysmodel.CreateUserDTO {
	return sysmodel.CreateUserDTO{
		Username: fmt.Sprintf("test-%d", uuid.New().ID()),
		Password: "Test123!@#$%",
		IsActive: true,
		IsStaff:  false,
		RoleID:   roleID,
	}
}

type UserHandlerTestSuite struct {
	suite.Suite
	router       *gin.Engine
	handler      *UserHandler
	userService  *syssvc.UserService
	roleRepo     *sysrepo.RoleRepo
	userRepo     *sysrepo.UserRepo
	recordRepo   *sysrepo.LoginRecordRepo
	testRoleID   uint32
	testUserID   uint32
	testUsername string
}

func (s *UserHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.AutoMigrate(
		&sysmodel.RoleModel{},
		&sysmodel.UserModel{},
		&sysmodel.LoginRecordModel{},
	), "数据库迁移失败")

	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, err := auth.NewCasbinEnforcer()
	s.Require().NoError(err, "创建Casbin enforcer失败")

	s.roleRepo = sysrepo.NewRoleRepo(logger, db, dbTimeout, slowThreshold, enforcer)
	s.userRepo = sysrepo.NewUserRepo(logger, db, dbTimeout, slowThreshold)
	s.recordRepo = sysrepo.NewLoginRecordRepo(logger, db, dbTimeout, slowThreshold, time.Minute*10, time.Minute*10, 5)

	jwtConfig := config.NewJWTConfig(
		time.Second*10,
		time.Minute*10,
		"HS256",
		"HS256",
		[]byte("test-access-secret-key-1234567890123456"),
		[]byte("test-refresh-secret-key-123456789012345"),
	)

	s.userService = syssvc.NewUserService(
		logger,
		s.roleRepo,
		s.userRepo,
		s.recordRepo,
		crypto.NewBcryptHasher(12),
		jwtConfig,
		syssvc.SecuritySettings{
			MaxFailedAttempts: 5,
			LockDuration:      time.Second * 5,
			PasswordStrength:  3,
		},
	)

	s.handler = NewUserHandler(logger, s.userService)

	s.router = gin.Default()
	userGroup := s.router.Group("/api/v1/customer")
	s.handler.LoadRouter(userGroup)

	s.router.POST("/api/v1/login", s.handler.Login)
	s.router.POST("/api/v1/refresh/token", s.handler.RefreshToken)

	testRole := createTestUserHandlerRoleModel()
	s.Require().NoError(s.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil))
	s.testRoleID = testRole.ID

	testUserDTO := createTestCreateUserDTO(s.testRoleID)
	createdUser, rErr := s.userService.CreateUser(context.Background(), testUserDTO)
	s.Require().Nil(rErr, "创建测试用户失败")
	s.Require().NotNil(createdUser)
	s.testUserID = createdUser.ID
	s.testUsername = createdUser.Username
}

func mockAuthMiddleware(userID uint32, username string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := &auth.JwtClaims{
			UserInfo: auth.UserInfo{
				UserID:   userID,
				Username: username,
			},
		}
		ctx := context.WithValue(c.Request.Context(), ctxutil.JwtClaimsKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (s *UserHandlerTestSuite) SetupTest() {
	db := test.NewTestGormDBWithConfig(nil)
	s.Require().NoError(db.Exec("DELETE FROM sys_login_record").Error, "清理登录记录数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_user").Error, "清理用户数据失败")
	s.Require().NoError(db.Exec("DELETE FROM sys_role").Error, "清理角色数据失败")

	testRole := createTestUserHandlerRoleModel()
	s.Require().NoError(s.roleRepo.CreateModel(context.Background(), testRole, nil, nil, nil))
	s.testRoleID = testRole.ID

	testUserDTO := createTestCreateUserDTO(s.testRoleID)
	createdUser, rErr := s.userService.CreateUser(context.Background(), testUserDTO)
	s.Require().Nil(rErr, "创建测试用户失败")
	s.Require().NotNil(createdUser)
	s.testUserID = createdUser.ID
	s.testUsername = createdUser.Username

	s.router = gin.Default()
	userGroup := s.router.Group("/api/v1/customer")
	userGroup.Use(mockAuthMiddleware(s.testUserID, s.testUsername))
	s.handler.LoadRouter(userGroup)
	userGroup.PATCH("/me/password", s.handler.PatchPassword)
	userGroup.GET("/me/record/login", s.handler.ListMeLoginRecord)
	s.router.POST("/api/v1/login", s.handler.Login)
	s.router.POST("/api/v1/refresh/token", s.handler.RefreshToken)
}

func (s *UserHandlerTestSuite) TestCreateUser_Success() {
	dto := createTestCreateUserDTO(s.testRoleID)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("POST", "/api/v1/customer/user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var resp sysmodel.UserResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusCreated, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(dto.Username, resp.Data.Username)
	s.Equal(dto.IsActive, resp.Data.IsActive)
	s.Equal(dto.IsStaff, resp.Data.IsStaff)
	s.Equal(dto.RoleID, resp.Data.Role.ID)
}

func (s *UserHandlerTestSuite) TestCreateUser_InvalidRequest() {
	invalidDTO := map[string]any{
		"username": "",
	}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/customer/user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusCreated, w.Code)
}

func (s *UserHandlerTestSuite) TestCreateUser_InvalidRole() {
	dto := createTestCreateUserDTO(999999)
	jsonData, _ := json.Marshal(dto)

	req := httptest.NewRequest("POST", "/api/v1/customer/user", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusCreated, w.Code)
}

func (s *UserHandlerTestSuite) TestGetUser_Success() {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/customer/user/%d", s.testUserID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.UserResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(s.testUserID, resp.Data.ID)
	s.Equal(s.testUsername, resp.Data.Username)
}

func (s *UserHandlerTestSuite) TestGetUser_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/customer/user/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestGetUser_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/customer/user/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestUpdateUser_Success() {
	updateDTO := sysmodel.UpdateUserDTO{
		Username: fmt.Sprintf("updated-%d", uuid.New().ID()),
		IsActive: false,
		IsStaff:  true,
		RoleID:   s.testRoleID,
	}
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/customer/user/%d", s.testUserID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.UserResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(updateDTO.Username, resp.Data.Username)
	s.Equal(updateDTO.IsActive, resp.Data.IsActive)
	s.Equal(updateDTO.IsStaff, resp.Data.IsStaff)
}

func (s *UserHandlerTestSuite) TestUpdateUser_NotFound() {
	updateDTO := sysmodel.UpdateUserDTO{
		Username: "updated-username",
		RoleID:   s.testRoleID,
	}
	jsonData, _ := json.Marshal(updateDTO)

	req := httptest.NewRequest("PUT", "/api/v1/customer/user/999999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestUpdateUser_InvalidRequest() {
	invalidDTO := map[string]any{
		"username": "",
	}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/customer/user/%d", s.testUserID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestDeleteUser_Success() {
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/customer/user/%d", s.testUserID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	_, err := s.userRepo.GetModel(context.Background(), nil, s.testUserID)
	s.Error(err)
}

func (s *UserHandlerTestSuite) TestDeleteUser_NotFound() {
	req := httptest.NewRequest("DELETE", "/api/v1/customer/user/999999", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestDeleteUser_InvalidID() {
	req := httptest.NewRequest("DELETE", "/api/v1/customer/user/invalid", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestListUser_Success() {
	for i := 0; i < 5; i++ {
		dto := createTestCreateUserDTO(s.testRoleID)
		_, rErr := s.userService.CreateUser(context.Background(), dto)
		s.Require().Nil(rErr)
	}

	req := httptest.NewRequest("GET", "/api/v1/customer/user?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagUserResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(6))
}

func (s *UserHandlerTestSuite) TestListUser_WithFilter() {
	dto := createTestCreateUserDTO(s.testRoleID)
	dto.Username = "unique-user-123"
	_, rErr := s.userService.CreateUser(context.Background(), dto)
	s.Require().Nil(rErr)

	req := httptest.NewRequest("GET", "/api/v1/customer/user?username=unique-user-123", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagUserResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(1))
}

func (s *UserHandlerTestSuite) TestListUser_EmptyResult() {
	req := httptest.NewRequest("GET", "/api/v1/customer/user?username=non-existent-user", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagUserResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.Equal(int64(0), resp.Data.Total)
}

func (s *UserHandlerTestSuite) TestResetPassword_Success() {
	resetDTO := sysmodel.ResetPasswordDTO{
		NewPassword:     "NewTest123!@#$%",
		ConfirmPassword: "NewTest123!@#$%",
	}
	jsonData, _ := json.Marshal(resetDTO)

	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/customer/user/password/%d", s.testUserID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestResetPassword_NotFound() {
	resetDTO := sysmodel.ResetPasswordDTO{
		NewPassword:     "NewTest123!@#$%",
		ConfirmPassword: "NewTest123!@#$%",
	}
	jsonData, _ := json.Marshal(resetDTO)

	req := httptest.NewRequest("PATCH", "/api/v1/customer/user/password/999999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestResetPassword_PasswordMismatch() {
	resetDTO := sysmodel.ResetPasswordDTO{
		NewPassword:     "NewTest123!@#$%",
		ConfirmPassword: "DifferentPassword!@#",
	}
	jsonData, _ := json.Marshal(resetDTO)

	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/customer/user/password/%d", s.testUserID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestResetPassword_WeakPassword() {
	resetDTO := sysmodel.ResetPasswordDTO{
		NewPassword:     "weak",
		ConfirmPassword: "weak",
	}
	jsonData, _ := json.Marshal(resetDTO)

	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/customer/user/password/%d", s.testUserID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestPatchPassword_Success() {
	patchDTO := sysmodel.PatchPasswordDTO{
		OldPassword:     "Test123!@#$%",
		NewPassword:     "NewTest123!@#$%",
		ConfirmPassword: "NewTest123!@#$%",
	}
	jsonData, _ := json.Marshal(patchDTO)

	req := httptest.NewRequest("PATCH", "/api/v1/customer/me/password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestPatchPassword_WrongOldPassword() {
	patchDTO := sysmodel.PatchPasswordDTO{
		OldPassword:     "wrong_password",
		NewPassword:     "NewTest123!@#$%",
		ConfirmPassword: "NewTest123!@#$%",
	}
	jsonData, _ := json.Marshal(patchDTO)

	req := httptest.NewRequest("PATCH", "/api/v1/customer/me/password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestPatchPassword_PasswordMismatch() {
	patchDTO := sysmodel.PatchPasswordDTO{
		OldPassword:     "Test123!@#$%",
		NewPassword:     "NewTest123!@#$%",
		ConfirmPassword: "DifferentPassword!@#",
	}
	jsonData, _ := json.Marshal(patchDTO)

	req := httptest.NewRequest("PATCH", "/api/v1/customer/me/password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestPatchPassword_WeakPassword() {
	patchDTO := sysmodel.PatchPasswordDTO{
		OldPassword:     "Test123!@#$%",
		NewPassword:     "weak",
		ConfirmPassword: "weak",
	}
	jsonData, _ := json.Marshal(patchDTO)

	req := httptest.NewRequest("PATCH", "/api/v1/customer/me/password", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestLogin_Success() {
	loginDTO := sysmodel.LoginDTO{
		Username: s.testUsername,
		Password: "Test123!@#$%",
	}
	jsonData, _ := json.Marshal(loginDTO)

	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.LoginResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotEmpty(resp.Data.AccessToken)
	s.NotEmpty(resp.Data.RefreshToken)
}

func (s *UserHandlerTestSuite) TestLogin_WrongPassword() {
	loginDTO := sysmodel.LoginDTO{
		Username: s.testUsername,
		Password: "wrong_password",
	}
	jsonData, _ := json.Marshal(loginDTO)

	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestLogin_UserNotFound() {
	loginDTO := sysmodel.LoginDTO{
		Username: "non-existent-user",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(loginDTO)

	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestLogin_InvalidRequest() {
	invalidDTO := map[string]any{
		"username": "",
	}
	jsonData, _ := json.Marshal(invalidDTO)

	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestRefreshToken_Success() {
	loginDTO := sysmodel.LoginDTO{
		Username: s.testUsername,
		Password: "Test123!@#$%",
	}
	jsonData, _ := json.Marshal(loginDTO)

	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	var loginResp sysmodel.LoginResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &loginResp))

	refreshDTO := sysmodel.RefreshTokenDTO{
		RefreshToken: loginResp.Data.RefreshToken,
	}
	refreshJson, _ := json.Marshal(refreshDTO)

	refreshReq := httptest.NewRequest("POST", "/api/v1/refresh/token", bytes.NewBuffer(refreshJson))
	refreshReq.Header.Set("Content-Type", "application/json")
	refreshW := httptest.NewRecorder()

	s.router.ServeHTTP(refreshW, refreshReq)

	s.Equal(http.StatusOK, refreshW.Code)

	var refreshResp sysmodel.LoginResp
	s.Require().NoError(json.Unmarshal(refreshW.Body.Bytes(), &refreshResp))
	s.Equal(http.StatusOK, refreshResp.Code)
	s.NotEmpty(refreshResp.Data.AccessToken)
	s.NotEmpty(refreshResp.Data.RefreshToken)
}

func (s *UserHandlerTestSuite) TestRefreshToken_InvalidToken() {
	refreshDTO := sysmodel.RefreshTokenDTO{
		RefreshToken: "invalid_token",
	}
	jsonData, _ := json.Marshal(refreshDTO)

	req := httptest.NewRequest("POST", "/api/v1/refresh/token", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestRefreshToken_EmptyToken() {
	refreshDTO := sysmodel.RefreshTokenDTO{
		RefreshToken: "",
	}
	jsonData, _ := json.Marshal(refreshDTO)

	req := httptest.NewRequest("POST", "/api/v1/refresh/token", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func (s *UserHandlerTestSuite) TestListLoginRecord_Success() {
	loginDTO := sysmodel.LoginDTO{
		Username: s.testUsername,
		Password: "Test123!@#$%",
	}
	for i := 0; i < 3; i++ {
		jsonData, _ := json.Marshal(loginDTO)
		req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		s.Equal(http.StatusOK, w.Code)
	}

	req := httptest.NewRequest("GET", "/api/v1/customer/user/record/login?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagLoginRecordResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
	s.GreaterOrEqual(resp.Data.Total, int64(3))
}

func (s *UserHandlerTestSuite) TestListLoginRecord_WithFilter() {
	req := httptest.NewRequest("GET", "/api/v1/customer/user/record/login?username="+s.testUsername, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagLoginRecordResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
}

func (s *UserHandlerTestSuite) TestListMeLoginRecord_Success() {
	loginDTO := sysmodel.LoginDTO{
		Username: s.testUsername,
		Password: "Test123!@#$%",
	}
	for i := 0; i < 2; i++ {
		jsonData, _ := json.Marshal(loginDTO)
		req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		s.Equal(http.StatusOK, w.Code)
	}

	req := httptest.NewRequest("GET", "/api/v1/customer/me/record/login?page=1&size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp sysmodel.PagLoginRecordResp
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Equal(http.StatusOK, resp.Code)
	s.NotNil(resp.Data)
}

func (s *UserHandlerTestSuite) TestNewUserHandler() {
	logger := test.NewTestZapLogger()

	handler := NewUserHandler(logger, s.userService)

	s.NotNil(handler)
	s.NotNil(handler.log)
	s.NotNil(handler.userSvc)
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, &UserHandlerTestSuite{})
}
