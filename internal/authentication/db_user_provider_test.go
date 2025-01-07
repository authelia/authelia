package authentication_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestShouldRunStorageSuite(t *testing.T) {
	suite.Run(t, new(DBUserProviderSuite))
}

type DBUserProviderSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *DBUserProviderSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	s.mock.Ctx.Configuration.AuthenticationBackend = schema.AuthenticationBackend{
		DB: &schema.DefaultDBAuthenticationBackendConfig,
	}

	provider := authentication.NewDBUserProvider(s.mock.Ctx.Configuration.AuthenticationBackend.DB, s.mock.StorageMock)
	err := provider.StartupCheck()
	s.NoError(err)
	s.mock.Ctx.Providers.UserProvider = provider
}
func (s *DBUserProviderSuite) TearDownTest() {
	s.mock.Ctrl.Finish()
}

func (s *DBUserProviderSuite) TestStartupCheckShouldPass() {
	provider := authentication.NewDBUserProvider(s.mock.Ctx.Configuration.AuthenticationBackend.DB, s.mock.StorageMock)

	s.IsType(&authentication.DBUserProvider{}, provider)
	s.NoError(provider.StartupCheck())
}

func (s *DBUserProviderSuite) TestStartupCheckShouldFailIfInvalidPasswordAlgorithm() {
	provider := authentication.NewDBUserProvider(&schema.AuthenticationBackendDB{
		Password: schema.AuthenticationBackendPassword{
			Algorithm: "invalid",
		},
	}, s.mock.StorageMock)

	s.ErrorContains(provider.StartupCheck(), "algorithm 'invalid' is unknown")
}

func (s *DBUserProviderSuite) TestGetUserShouldFailIfUserNotFound() {
	provider := s.mock.Ctx.Providers.UserProvider
	s.mock.StorageMock.EXPECT().LoadUserByUsername(gomock.Any(), gomock.Eq("ada")).
		Return(model.User{}, errors.New("user not found"))

	user, err := provider.GetDetails("ada")
	s.Nil(user)
	s.ErrorContains(err, "user not found")
}

func (s *DBUserProviderSuite) TestGetUserShouldGetUserByUsername() {
	provider := s.mock.Ctx.Providers.UserProvider

	s.mock.StorageMock.EXPECT().LoadUserByUsername(gomock.Any(), gomock.Eq("john")).
		Return(model.User{
			Username:    "john",
			Email:       "john@example.com",
			DisplayName: "John Doe",
			Groups:      []string{"admins", "dev"},
			Disabled:    false,
		}, nil)

	user, err := provider.GetDetails("john")
	s.NoError(err)
	s.NotNil(user)
	s.Equal("john@example.com", user.Emails[0])
}

func (s *DBUserProviderSuite) TestGetUserShouldGetUserByEmail() {
	provider := authentication.NewDBUserProvider(&schema.AuthenticationBackendDB{
		Search: schema.AuthenticationBackendDBSearch{
			Email: true,
		},
	}, s.mock.StorageMock)

	s.mock.StorageMock.EXPECT().LoadUserByEmail(gomock.Any(), gomock.Eq("john@example.com")).
		Return(model.User{
			Username:    "john",
			Email:       "john@example.com",
			DisplayName: "John Doe",
			Groups:      []string{"admins", "dev"},
			Disabled:    false,
		}, nil)

	user, err := provider.GetDetails("john@example.com")
	s.NotNil(user)
	s.NoError(err)
	s.Equal("john", user.Username)
}

func (s *DBUserProviderSuite) TestCheckPasswordOk() {
	provider := s.mock.Ctx.Providers.UserProvider

	s.mock.StorageMock.EXPECT().LoadUserByUsername(gomock.Any(), gomock.Eq("john")).
		Return(model.User{
			Username:    "john",
			Email:       "john@example.com",
			DisplayName: "John Doe",
			Groups:      []string{"admins", "dev"},
			Disabled:    false,
			Password:    []byte("$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"),
		}, nil)

	valid, err := provider.CheckUserPassword("john", "password")
	s.NoError(err)
	s.True(valid)
}

func (s *DBUserProviderSuite) TestCheckPasswordFailsIfPasswordDoesNotMatch() {
	provider := s.mock.Ctx.Providers.UserProvider

	s.mock.StorageMock.EXPECT().LoadUserByUsername(gomock.Any(), gomock.Eq("john")).
		Return(model.User{
			Username:    "john",
			Email:       "john@example.com",
			DisplayName: "John Doe",
			Groups:      []string{"admins", "dev"},
			Disabled:    false,
			Password:    []byte("$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"),
		}, nil)

	valid, err := provider.CheckUserPassword("john", "incorrect")
	s.NoError(err)
	s.False(valid)
}

func (s *DBUserProviderSuite) TestCheckPasswordFailsIfPasswordInStorageIsEmpty() {
	provider := s.mock.Ctx.Providers.UserProvider

	s.mock.StorageMock.EXPECT().LoadUserByUsername(gomock.Any(), gomock.Eq("john")).
		Return(model.User{
			Username:    "john",
			Email:       "john@example.com",
			DisplayName: "John Doe",
			Groups:      []string{"admins", "dev"},
			Disabled:    false,
			Password:    []byte{},
		}, nil)

	valid, err := provider.CheckUserPassword("john", "any")
	s.ErrorContains(err, "provided encoded hash has an invalid format")
	s.False(valid)
}

func (s *DBUserProviderSuite) TestCheckPasswordFailsIfUserNotFound() {
	provider := s.mock.Ctx.Providers.UserProvider

	s.mock.StorageMock.EXPECT().LoadUserByUsername(gomock.Any(), gomock.Eq("ada")).
		Return(model.User{}, errors.New("user not found"))

	valid, err := provider.CheckUserPassword("ada", "any")
	s.ErrorContains(err, "user not found")
	s.False(valid)
}

// TODO: test that hashed password meets expected hash.
func (s *DBUserProviderSuite) TestUpdatePasswordOk() {
	provider := s.mock.Ctx.Providers.UserProvider

	s.mock.StorageMock.EXPECT().LoadUserByUsername(gomock.Any(), gomock.Eq("john")).
		Return(model.User{
			Username:    "john",
			Email:       "john@example.com",
			DisplayName: "John Doe",
			Groups:      []string{"admins", "dev"},
			Disabled:    false,
		}, nil)

	s.mock.StorageMock.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Eq("john"), gomock.Any()).
		Return(nil)

	err := provider.UpdatePassword("john", "password")
	s.NoError(err)
}

func (s *DBUserProviderSuite) TestUpdatePasswordFailsIfUserIsDisabled() {
	provider := s.mock.Ctx.Providers.UserProvider

	s.mock.StorageMock.EXPECT().LoadUserByUsername(gomock.Any(), gomock.Eq("john")).
		Return(model.User{
			Username:    "john",
			Email:       "john@example.com",
			DisplayName: "John Doe",
			Groups:      []string{"admins", "dev"},
			Disabled:    true,
		}, nil)

	err := provider.UpdatePassword("john", "password")
	s.ErrorContains(err, "user not found")
}

func (s *DBUserProviderSuite) TestUpdatePasswordFailsIfUserNotFound() {
	provider := s.mock.Ctx.Providers.UserProvider

	s.mock.StorageMock.EXPECT().LoadUserByUsername(gomock.Any(), gomock.Eq("ada")).
		Return(model.User{}, errors.New("user not found"))

	err := provider.UpdatePassword("ada", "password")
	s.ErrorContains(err, "user not found")
}

func (s *DBUserProviderSuite) TestUpdatePasswordFailsIfStorageBackendFails() {
	provider := s.mock.Ctx.Providers.UserProvider

	var spectedError = errors.New("some error")

	s.mock.StorageMock.EXPECT().LoadUserByUsername(gomock.Any(), gomock.Eq("john")).
		Return(model.User{
			Username:    "john",
			Email:       "john@example.com",
			DisplayName: "John Doe",
			Groups:      []string{"admins", "dev"},
			Disabled:    false,
		}, nil)

	s.mock.StorageMock.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Eq("john"), gomock.Any()).
		Return(spectedError)

	err := provider.UpdatePassword("john", "password")
	s.ErrorIs(err, spectedError)
}
