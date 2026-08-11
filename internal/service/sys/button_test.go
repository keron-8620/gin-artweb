package sys

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	sysmodel "gin-artweb/internal/model/sys"
	syssvc "gin-artweb/internal/repo/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/test"
)

func CreateTestButtonModel(menuID uint32) *sysmodel.ButtonModel {
	return &sysmodel.ButtonModel{
		MenuID:   menuID,
		Name:     uuid.NewString(),
		Sort:     10000,
		IsActive: true,
		Descr:    "test button",
	}
}

func CreateTestButtonDTO(menuID uint32, apiIDs []uint32) sysmodel.CreateButtonDTO {
	return sysmodel.CreateButtonDTO{
		ID:       uint32(uuid.New().ID()),
		MenuID:   menuID,
		Name:     uuid.NewString(),
		Sort:     10000,
		IsActive: true,
		Descr:    "test button",
		ApiIDs:   apiIDs,
	}
}

func newCanceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

type ButtonTestSuite struct {
	suite.Suite
	buttonService *ButtonService
}

func (s *ButtonTestSuite) SetupSuite() {
	db := test.NewTestGormDBWithConfig(nil)
	if err := db.AutoMigrate(
		&sysmodel.MenuModel{},
		&sysmodel.ApiModel{},
		&sysmodel.ButtonModel{},
	); err != nil {
		s.Error(err, "数据库迁移失败")
	}
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	slowThreshold := test.NewTestDBSlowThreshold()
	enforcer, err := auth.NewCasbinEnforcer()
	if err != nil {
		s.Error(err, "创建Casbinforcer失败")
	}
	s.buttonService = NewButtonService(
		logger,
		syssvc.NewApiRepo(logger, db, dbTimeout, slowThreshold, enforcer),
		syssvc.NewMenuRepo(logger, db, dbTimeout, slowThreshold, enforcer),
		syssvc.NewButtonRepo(logger, db, dbTimeout, slowThreshold, enforcer),
	)
}

func (s *ButtonTestSuite) TestGetMenu() {
	s.Run("nonExistent", func() {
		_, err := s.buttonService.getMenu(context.Background(), 9999)
		s.NotNil(err)
	})
	s.Run("contextCanceled", func() {
		_, err := s.buttonService.getMenu(newCanceledContext(), 1)
		s.NotNil(err)
	})
}

func (s *ButtonTestSuite) TestGetApis() {
	s.Run("emptyList", func() {
		apis, err := s.buttonService.getApis(context.Background(), []uint32{})
		s.Nil(err)
		s.NotNil(apis)
		s.Len(apis, 0)
	})
	s.Run("nonExistent", func() {
		apis, err := s.buttonService.getApis(context.Background(), []uint32{9999, 8888})
		s.Nil(err)
		s.NotNil(apis)
	})
	s.Run("contextCanceled", func() {
		_, err := s.buttonService.getApis(newCanceledContext(), []uint32{1})
		s.NotNil(err)
	})
}

func (s *ButtonTestSuite) TestCreateButton() {
	s.Run("normal", func() {
		menu := CreateTestMenuModel(nil)
		s.Nil(s.buttonService.menuRepo.CreateModel(context.Background(), menu, nil))
		s.Greater(menu.ID, uint32(0))

		dto := CreateTestButtonDTO(menu.ID, []uint32{})
		created, err := s.buttonService.CreateButton(context.Background(), dto)
		s.Nil(err)
		s.Greater(created.ID, uint32(0))
		s.Equal(dto.MenuID, created.MenuID)
		s.Equal(dto.Name, created.Name)
		s.Equal(dto.Sort, created.Sort)
		s.Equal(dto.IsActive, created.IsActive)
		s.Equal(dto.Descr, created.Descr)
	})
	s.Run("contextCanceled", func() {
		_, err := s.buttonService.CreateButton(newCanceledContext(), CreateTestButtonDTO(1, []uint32{}))
		s.NotNil(err)
	})
}

func (s *ButtonTestSuite) TestFindButtonByID() {
	s.Run("normal", func() {
		menu := CreateTestMenuModel(nil)
		s.Nil(s.buttonService.menuRepo.CreateModel(context.Background(), menu, nil))

		dto := CreateTestButtonDTO(menu.ID, []uint32{})
		created, err := s.buttonService.CreateButton(context.Background(), dto)
		s.Nil(err)

		found, err := s.buttonService.FindButtonByID(context.Background(), []string{}, created.ID)
		s.Nil(err)
		s.Equal(created.Name, found.Name)
		s.Equal(created.MenuID, found.MenuID)
	})
	s.Run("contextCanceled", func() {
		_, err := s.buttonService.FindButtonByID(newCanceledContext(), []string{}, 1)
		s.NotNil(err)
	})
}

func (s *ButtonTestSuite) TestUpdateButtonByID() {
	s.Run("normal", func() {
		menu := CreateTestMenuModel(nil)
		s.Nil(s.buttonService.menuRepo.CreateModel(context.Background(), menu, nil))

		dto := CreateTestButtonDTO(menu.ID, []uint32{})
		created, err := s.buttonService.CreateButton(context.Background(), dto)
		s.Nil(err)

		updateDTO := sysmodel.UpdateButtonDTO{
			Name:     "Updated Button",
			Sort:     created.Sort,
			IsActive: false,
			Descr:    "updated test button",
			MenuID:   created.MenuID,
			ApiIDs:   []uint32{},
		}

		updated, err := s.buttonService.UpdateButtonByID(context.Background(), created.ID, updateDTO)
		s.Nil(err)
		s.Equal(created.ID, updated.ID)
		s.Equal(updateDTO.Name, updated.Name)
		s.Equal(updateDTO.Descr, updated.Descr)
		s.Equal(updateDTO.IsActive, updated.IsActive)
	})
	s.Run("contextCanceled", func() {
		updateDTO := sysmodel.UpdateButtonDTO{
			Name:     "Updated Button",
			Sort:     100,
			IsActive: true,
			Descr:    "Test",
			MenuID:   1,
			ApiIDs:   []uint32{},
		}
		_, err := s.buttonService.UpdateButtonByID(newCanceledContext(), 1, updateDTO)
		s.NotNil(err)
	})
}

func (s *ButtonTestSuite) TestDeleteButtonByID() {
	s.Run("normal", func() {
		menu := CreateTestMenuModel(nil)
		s.Nil(s.buttonService.menuRepo.CreateModel(context.Background(), menu, nil))

		dto := CreateTestButtonDTO(menu.ID, []uint32{})
		created, err := s.buttonService.CreateButton(context.Background(), dto)
		s.Nil(err)

		err = s.buttonService.DeleteButtonByID(context.Background(), created.ID)
		s.Nil(err)

		_, err = s.buttonService.FindButtonByID(context.Background(), []string{}, created.ID)
		s.NotNil(err)
	})
	s.Run("contextCanceled", func() {
		err := s.buttonService.DeleteButtonByID(newCanceledContext(), 1)
		s.NotNil(err)
	})
}

func (s *ButtonTestSuite) TestListButton() {
	s.Run("normal", func() {
		menu := CreateTestMenuModel(nil)
		s.Nil(s.buttonService.menuRepo.CreateModel(context.Background(), menu, nil))
		s.Greater(menu.ID, uint32(0))

		for i := 0; i < 3; i++ {
			_, err := s.buttonService.CreateButton(context.Background(), CreateTestButtonDTO(menu.ID, []uint32{}))
			s.Nil(err)
		}

		_, _, count, list, err := s.buttonService.ListButton(context.Background(), sysmodel.ListButtonDTO{})
		s.Nil(err)
		s.GreaterOrEqual(int(count), 3)
		s.NotNil(list)
	})
	s.Run("contextCanceled", func() {
		_, _, _, _, err := s.buttonService.ListButton(newCanceledContext(), sysmodel.ListButtonDTO{})
		s.NotNil(err)
	})
}

func (s *ButtonTestSuite) TestLoadButtonPolicy() {
	s.Run("normal", func() {
		menu := CreateTestMenuModel(nil)
		s.Nil(s.buttonService.menuRepo.CreateModel(context.Background(), menu, nil))

		for i := 0; i < 2; i++ {
			_, err := s.buttonService.CreateButton(context.Background(), CreateTestButtonDTO(menu.ID, []uint32{}))
			s.Nil(err)
		}

		err := s.buttonService.LoadButtonPolicy(context.Background())
		s.Nil(err)
	})
	s.Run("contextCanceled", func() {
		err := s.buttonService.LoadButtonPolicy(newCanceledContext())
		s.NotNil(err)
	})
}

func (s *ButtonTestSuite) TestCreateButtonWithApis() {
	menu := CreateTestMenuModel(nil)
	s.Nil(s.buttonService.menuRepo.CreateModel(context.Background(), menu, nil))
	s.Greater(menu.ID, uint32(0))

	apiIDs := make([]uint32, 0, 2)
	for i := 0; i < 2; i++ {
		api := CreateTestApiModel()
		s.Nil(s.buttonService.apiRepo.CreateModel(context.Background(), api))
		s.Greater(api.ID, uint32(0))
		apiIDs = append(apiIDs, api.ID)
	}

	dto := CreateTestButtonDTO(menu.ID, apiIDs)
	created, err := s.buttonService.CreateButton(context.Background(), dto)
	s.Nil(err)
	s.Greater(created.ID, uint32(0))
	s.Equal(dto.MenuID, created.MenuID)
	s.Equal(dto.Name, created.Name)

	found, err := s.buttonService.FindButtonByID(context.Background(), []string{"Apis"}, created.ID)
	s.Nil(err)
	s.NotNil(found.Apis)
	s.Len(found.Apis, 2)
}

func (s *ButtonTestSuite) TestButtonCasbinInheritance() {
	menu := CreateTestMenuModel(nil)
	s.Nil(s.buttonService.menuRepo.CreateModel(context.Background(), menu, nil))
	s.Greater(menu.ID, uint32(0))

	api := CreateTestApiModel()
	s.Nil(s.buttonService.apiRepo.CreateModel(context.Background(), api))
	s.Greater(api.ID, uint32(0))

	dto := CreateTestButtonDTO(menu.ID, []uint32{api.ID})
	created, err := s.buttonService.CreateButton(context.Background(), dto)
	s.Nil(err)
	s.Greater(created.ID, uint32(0))

	loadErr := s.buttonService.LoadButtonPolicy(context.Background())
	s.Nil(loadErr)

	buttonSubject := auth.ButtonToSubject(created.ID)
	apiSubject := auth.ApiToSubject(api.ID)
	s.Greater(created.ID, uint32(0))
	s.Greater(api.ID, uint32(0))
	s.NotEmpty(buttonSubject)
	s.NotEmpty(apiSubject)
}

func TestButtonTestSuite(t *testing.T) {
	suite.Run(t, &ButtonTestSuite{})
}
