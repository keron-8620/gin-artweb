package sys

import (
	"context"
	"fmt"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	sysmodel "gin-artweb/internal/model/sys"
	syssvc "gin-artweb/internal/repo/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/test"
)

func CreateTestApiDTO() sysmodel.CreateApiDTO {
	return sysmodel.CreateApiDTO{
		ID:     uint32(uuid.New().ID()),
		URL:    fmt.Sprintf("/api/test/%s/", uuid.NewString()),
		Method: "GET",
		Label:  "test",
		Descr:  "test api description",
	}
}

func CreateTestApiModel() *sysmodel.ApiModel {
	return &sysmodel.ApiModel{
		URL:    fmt.Sprintf("/api/test/%s/", uuid.NewString()),
		Method: "GET",
		Label:  "test",
		Descr:  "test api description",
	}
}

type apiTestSuite struct {
	suite.Suite
	enforcer   *casbin.Enforcer
	apiService *ApiService
}

func (s *apiTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(&sysmodel.ApiModel{}); err != nil {
		s.Error(err, "数据库迁移失败")
	}
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, err := auth.NewCasbinEnforcer()
	if err != nil {
		s.Error(err, "创建Casbinforcer失败")
	}
	s.enforcer = enforcer
	s.apiService = NewApiService(
		logger,
		syssvc.NewApiRepo(
			logger,
			db,
			dbTimeout,
			slowThreshold,
			enforcer,
		),
	)
}

func (s *apiTestSuite) newTestApiDTO() sysmodel.CreateApiDTO {
	return CreateTestApiDTO()
}

func (s *apiTestSuite) createTestApi() *sysmodel.ApiModel {
	dto := s.newTestApiDTO()
	api, err := s.apiService.CreateApi(context.Background(), dto)
	s.Require().Nil(err, "createTestApi: creation should succeed")
	s.Require().NotNil(api)
	return api
}

func (s *apiTestSuite) verifyApiFieldsEqual(expected, actual *sysmodel.ApiModel) {
	s.Equal(expected.URL, actual.URL)
	s.Equal(expected.Method, actual.Method)
	s.Equal(expected.Label, actual.Label)
	s.Equal(expected.Descr, actual.Descr)
}

func (s *apiTestSuite) verifyCasbinPolicy(apiID uint32, url, method string) {
	sub := auth.ApiToSubject(apiID)
	ok, err := s.enforcer.Enforce(sub, url, method)
	s.NoError(err, "Casbin Enforce should succeed")
	s.True(ok, "Casbin policy should exist")
}

func (s *apiTestSuite) TestCreateApi() {
	dto := s.newTestApiDTO()
	api, err := s.apiService.CreateApi(context.Background(), dto)

	s.Nil(err)
	s.NotNil(api)
	s.Greater(api.ID, uint32(0))
	s.verifyCasbinPolicy(api.ID, api.URL, api.Method)
}

func (s *apiTestSuite) TestFindApiByID() {
	created := s.createTestApi()

	found, err := s.apiService.FindApiByID(context.Background(), created.ID)
	s.Nil(err)
	s.NotNil(found)
	s.Equal(created.ID, found.ID)
	s.verifyApiFieldsEqual(created, found)
}

func (s *apiTestSuite) TestFindApiByID_NotFound() {
	_, err := s.apiService.FindApiByID(context.Background(), 0)
	s.NotNil(err)
}

func (s *apiTestSuite) TestDeleteApi() {
	created := s.createTestApi()

	err := s.apiService.DeleteApiByID(context.Background(), created.ID)
	s.Nil(err)

	_, err = s.apiService.FindApiByID(context.Background(), created.ID)
	s.NotNil(err)
}

func (s *apiTestSuite) TestDeleteApi_NotFound() {
	err := s.apiService.DeleteApiByID(context.Background(), 0)
	s.NotNil(err)
}

func (s *apiTestSuite) TestUpdateApiByID() {
	created := s.createTestApi()

	updateDTO := sysmodel.UpdateApiDTO{
		URL:    created.URL,
		Method: created.Method,
		Label:  "updated_label",
		Descr:  "updated description",
	}

	updated, err := s.apiService.UpdateApiByID(context.Background(), created.ID, updateDTO)
	s.Nil(err)
	s.NotNil(updated)
	s.Equal(created.ID, updated.ID)
	s.Equal(updateDTO.Label, updated.Label)
	s.Equal(updateDTO.Descr, updated.Descr)
	s.verifyCasbinPolicy(updated.ID, updated.URL, updated.Method)
}

func (s *apiTestSuite) TestUpdateApiByID_NotFound() {
	updateDTO := sysmodel.UpdateApiDTO{
		URL:    "/api/test/",
		Method: "GET",
		Label:  "updated_test",
		Descr:  "test update",
	}
	_, err := s.apiService.UpdateApiByID(context.Background(), 0, updateDTO)
	s.NotNil(err)
}

func (s *apiTestSuite) TestListApi() {
	created := make([]*sysmodel.ApiModel, 3)
	for i := range created {
		created[i] = s.createTestApi()
	}

	listDTO := sysmodel.ListApiDTO{}
	_, _, count, list, err := s.apiService.ListApi(context.Background(), listDTO)

	s.Nil(err)
	s.GreaterOrEqual(int(count), len(created))
	s.NotNil(list)
}

func (s *apiTestSuite) TestLoadApiPolicy() {
	created := make([]*sysmodel.ApiModel, 2)
	for i := range created {
		created[i] = s.createTestApi()
	}

	err := s.apiService.LoadApiPolicy(context.Background())
	s.Nil(err)

	for _, api := range created {
		s.verifyCasbinPolicy(api.ID, api.URL, api.Method)
	}
}

func (s *apiTestSuite) TestLoadApiPolicy_Empty() {
	err := s.apiService.LoadApiPolicy(context.Background())
	s.Nil(err)
}

func (s *apiTestSuite) TestNotFoundCases() {
	tests := []struct {
		name   string
		method func() error
	}{
		{"FindApiByID", func() error {
			_, err := s.apiService.FindApiByID(context.Background(), 0)
			return err
		}},
		{"DeleteApiByID", func() error {
			return s.apiService.DeleteApiByID(context.Background(), 0)
		}},
		{"UpdateApiByID", func() error {
			updateDTO := sysmodel.UpdateApiDTO{
				URL:    "/api/test/",
				Method: "GET",
				Label:  "updated_test",
				Descr:  "test update",
			}
			_, err := s.apiService.UpdateApiByID(context.Background(), 0, updateDTO)
			return err
		}},
	}

	for _, tt := range tests {
		s.Run(tt.name+"_NotFound", func() {
			err := tt.method()
			s.NotNil(err)
		})
	}
}

func (s *apiTestSuite) TestContextCancelledCases() {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name   string
		method func(ctx context.Context) error
	}{
		{
			"CreateApi",
			func(ctx context.Context) error {
				_, err := s.apiService.CreateApi(ctx, s.newTestApiDTO())
				return err
			},
		},
		{
			"FindApiByID",
			func(ctx context.Context) error {
				_, err := s.apiService.FindApiByID(ctx, 1)
				return err
			},
		},
		{
			"UpdateApiByID",
			func(ctx context.Context) error {
				updateDTO := sysmodel.UpdateApiDTO{
					URL:    "/api/test/",
					Method: "GET",
					Label:  "updated_test",
					Descr:  "test update",
				}
				_, err := s.apiService.UpdateApiByID(ctx, 1, updateDTO)
				return err
			},
		},
		{
			"DeleteApiByID",
			func(ctx context.Context) error {
				return s.apiService.DeleteApiByID(ctx, 1)
			},
		},
		{
			"ListApi",
			func(ctx context.Context) error {
				listDTO := sysmodel.ListApiDTO{}
				_, _, _, _, err := s.apiService.ListApi(ctx, listDTO)
				return err
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name+"_ContextError", func() {
			err := tt.method(cancelledCtx)
			s.NotNil(err)
		})
	}
}

func (s *apiTestSuite) TestUpdateApiByID_VerifyOldPolicyRemoved() {
	created := s.createTestApi()
	oldURL := created.URL
	oldMethod := created.Method
	sub := auth.ApiToSubject(created.ID)
	ok, _ := s.enforcer.Enforce(sub, oldURL, oldMethod)
	s.True(ok, "Old policy should exist before update")

	updateDTO := sysmodel.UpdateApiDTO{
		URL:    created.URL + "updated/",
		Method: "POST",
		Label:  "updated_label",
		Descr:  "updated description",
	}

	updated, err := s.apiService.UpdateApiByID(context.Background(), created.ID, updateDTO)
	s.Nil(err)
	s.NotNil(updated)

	ok, _ = s.enforcer.Enforce(sub, oldURL, oldMethod)
	s.False(ok, "Old policy should be removed after update")

	s.verifyCasbinPolicy(updated.ID, updated.URL, updated.Method)
}

func (s *apiTestSuite) TestDeleteApi_VerifyPolicyRemoved() {
	created := s.createTestApi()
	sub := auth.ApiToSubject(created.ID)
	ok, _ := s.enforcer.Enforce(sub, created.URL, created.Method)
	s.True(ok, "Policy should exist before delete")

	err := s.apiService.DeleteApiByID(context.Background(), created.ID)
	s.Nil(err)

	ok, _ = s.enforcer.Enforce(sub, created.URL, created.Method)
	s.False(ok, "Policy should be removed after delete")
}

func (s *apiTestSuite) TestUpdateApiByID_MultipleFields() {
	created := s.createTestApi()

	originalLabel := created.Label
	originalDescr := created.Descr

	updateDTO := sysmodel.UpdateApiDTO{
		URL:    created.URL,
		Method: created.Method,
		Label:  "new_label",
		Descr:  "new description",
	}

	updated, err := s.apiService.UpdateApiByID(context.Background(), created.ID, updateDTO)
	s.Nil(err)
	s.NotNil(updated)
	s.Equal(created.ID, updated.ID)
	s.Equal("new_label", updated.Label)
	s.Equal("new description", updated.Descr)
	s.Equal(originalLabel, created.Label)
	s.Equal(originalDescr, created.Descr)
}

func (s *apiTestSuite) TestListApi_WithFilter() {
	created := s.createTestApi()

	listDTO := sysmodel.ListApiDTO{
		URL:    created.URL,
		Method: created.Method,
	}
	_, _, count, list, err := s.apiService.ListApi(context.Background(), listDTO)

	s.Nil(err)
	s.GreaterOrEqual(int(count), 1)
	s.NotNil(list)

	found := false
	for _, api := range list {
		if api.ID == created.ID {
			found = true
			break
		}
	}
	s.True(found, "Created API should be in filtered list")
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, &apiTestSuite{})
}
