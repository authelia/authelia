package authentication_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
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

func (s *DBUserProviderSuite) TestGetDetails() {
	var provider = s.mock.Ctx.Providers.UserProvider.(*authentication.DBUserProvider)

	var username = "john"

	var testCases = []struct {
		name        string
		username    string
		provider    authentication.UserProvider
		setup       func(mock *mocks.MockAutheliaCtx)
		expectError error
	}{
		{
			"ShouldFailIfUserNotFound",
			username,
			provider,
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{}, storage.ErrUserNotFound)
			},
			authentication.ErrUserNotFound,
		},
		{
			"ShouldFailIfUserIsDisabled",
			username,
			provider,
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{
						Username:    "john",
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Disabled:    true,
					}, nil)
			},
			authentication.ErrUserNotFound,
		},
		{
			"ShouldPassIfValidUsername",
			username,
			provider,
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{
						Username:    "john",
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Disabled:    false,
					}, nil)
			},
			nil,
		},
		{
			"ShouldAllowGetByusername",
			username,
			provider,
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{
						Username:    "john",
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Disabled:    false,
					}, nil)
			},
			nil,
		},
		{
			"ShouldAllowGetByEmail",
			"john@example.com",
			authentication.NewDBUserProvider(&schema.AuthenticationBackendDB{
				Search: schema.AuthenticationBackendDBSearch{
					Email: true,
				},
			}, s.mock.StorageMock),
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq("john@example.com"), true).
					Return(model.User{
						Username:    "john",
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Disabled:    false,
					}, nil)
			},
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup(s.mock)

			_, err := tc.provider.GetDetails(tc.username)

			s.ErrorIs(err, tc.expectError)
		})
	}
}

func (s *DBUserProviderSuite) TestCheckPassword() {
	var username = "john"

	var hashedPassword = []byte("$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/")

	var testCases = []struct {
		name        string
		password    string
		setup       func(mock *mocks.MockAutheliaCtx)
		expect      bool
		expectError error
	}{
		{
			"ShouldSuccessIfUserExistsAndPasswordMatch",
			"password",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{
						Username:    username,
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Password:    hashedPassword,
						Disabled:    false,
					}, nil)
			},
			true,
			nil,
		},
		{
			"ShouldFailIfPasswordNotMatch",
			"incorrect_password",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{
						Username:    username,
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Password:    hashedPassword,
						Disabled:    false,
					}, nil)
			},
			false,
			nil,
		},
		{
			"ShouldFailIfStoredPasswordIsEmpty",
			"password",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{
						Username:    username,
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Password:    []byte(""),
						Disabled:    false,
					}, nil)
			},
			false,
			authentication.ErrInvalidPassword,
		},
		{
			"ShouldFailIfUserNotFound",
			"password",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{}, errors.New("user not found"))
			},
			false,
			authentication.ErrUserNotFound,
		},
		{
			"ShouldFailIfUserIsDisabled",
			"password",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{
						Username:    username,
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Password:    hashedPassword,
						Disabled:    true,
					}, nil)
			},
			false,
			authentication.ErrUserNotFound,
		},
	}

	provider := s.mock.Ctx.Providers.UserProvider.(*authentication.DBUserProvider)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup(s.mock)

			valid, err := provider.CheckUserPassword(username, tc.password)

			s.Equal(tc.expect, valid)

			s.ErrorIs(err, tc.expectError)
		})
	}
}

func (s *DBUserProviderSuite) TestUpdatePassword() {
	var username, password = "john", "password"

	var testCases = []struct {
		name        string
		setup       func(mock *mocks.MockAutheliaCtx)
		expectError error
	}{
		{
			"ShouldSuccessIfUserExists",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{
						Username:    username,
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Disabled:    false,
					}, nil)

				s.mock.StorageMock.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Eq(username), gomock.Any()).
					Return(nil)
			},
			nil,
		},
		{
			"ShouldFailIfUserNotExists",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{}, storage.ErrUserNotFound)
			},
			authentication.ErrUserNotFound,
		},
		{
			"ShouldFailIfUserIsDisabled",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{
						Username:    username,
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Disabled:    true,
					}, nil)
			},
			authentication.ErrUserNotFound,
		},
		{
			"ShouldFailIfHaveErrorSavingChanges",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().LoadUser(gomock.Any(), gomock.Eq(username), false).
					Return(model.User{
						Username:    username,
						Email:       "john@example.com",
						DisplayName: "John Doe",
						Groups:      []string{"admins", "dev"},
						Disabled:    false,
					}, nil)

				s.mock.StorageMock.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Eq(username), gomock.Any()).
					Return(errors.New("some error"))
			},
			authentication.ErrUpdatingUserPassword,
		},
	}

	provider := s.mock.Ctx.Providers.UserProvider.(*authentication.DBUserProvider)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup(s.mock)

			err := provider.UpdatePassword(username, password)

			s.ErrorIs(err, tc.expectError)
		})
	}
}

func (s *DBUserProviderSuite) TestAddUser() {
	var empty, username, password, email, displayname = "", "john", "password", "john@example.com", "John Doe"

	var testCases = []struct {
		name        string
		username    string
		password    string
		displayname string
		options     []func(options *authentication.NewUserDetailsOpts)
		setup       func(mock *mocks.MockAutheliaCtx)
		expectError error
	}{
		{
			"ShouldPass",
			username,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail(email)},
			func(mock *mocks.MockAutheliaCtx) {
				ctx := context.Background()
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq("john")).
						Return(false, nil),
					s.mock.StorageMock.EXPECT().CreateUser(gomock.Any(), gomock.Eq("john"), gomock.Eq("john@example.com"), gomock.Any()).
						Return(nil),
					s.mock.StorageMock.EXPECT().UpdateUserGroups(gomock.Any(), gomock.Any()).
						Return(nil),
					s.mock.StorageMock.EXPECT().UpdateUserDisplayName(gomock.Any(), gomock.Eq("john"), gomock.Any()).
						Return(nil),
					s.mock.StorageMock.EXPECT().Commit(gomock.Any()).
						Return(nil),
				)
			},
			nil,
		},
		{
			"ShouldFailIfEmptyPassword",
			username,
			empty,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail(email)},
			func(mock *mocks.MockAutheliaCtx) {},
			authentication.ErrInvalidPassword,
		},
		{
			"ShouldFailIfEmptyUsername",
			empty,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail(email)},
			func(mock *mocks.MockAutheliaCtx) {},
			authentication.ErrInvalidUsername,
		},
		{
			"ShouldFailIfNoEmailProvided",
			username,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){},
			func(mock *mocks.MockAutheliaCtx) {},
			authentication.ErrInvalidEmail,
		},
		{
			"ShouldFailIfInvalidEmailProvided",
			username,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail("not_a_email")},
			func(mock *mocks.MockAutheliaCtx) {},
			authentication.ErrInvalidEmail,
		},
		{
			"ShouldFailIfUserExists",
			username,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail(email)},
			func(mock *mocks.MockAutheliaCtx) {
				ctx := context.Background()

				gomock.InOrder(
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq("john")).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().Rollback(gomock.Any()).
						Return(nil),
				)
			},
			authentication.ErrUserExists,
		},
		{
			"ShouldFailIfCantBeginTx",
			username,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail(email)},
			func(mock *mocks.MockAutheliaCtx) {
				ctx := context.Background()

				s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
					Return(ctx, errors.New("some tx error"))
			},
			authentication.ErrCreatingUser,
		},
		{
			"ShouldFailIfCantCreateUser",
			username,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail(email)},
			func(mock *mocks.MockAutheliaCtx) {
				ctx := context.Background()

				gomock.InOrder(
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq("john")).
						Return(false, nil),
					s.mock.StorageMock.EXPECT().CreateUser(gomock.Any(), gomock.Eq("john"), gomock.Eq("john@example.com"), gomock.Any()).
						Return(errors.New("error creating user")),
					s.mock.StorageMock.EXPECT().Rollback(gomock.Any()).
						Return(nil),
				)
			},
			authentication.ErrCreatingUser,
		},
		{
			"ShouldFailIfCantCreateUserAndCantRollback",
			username,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail(email)},
			func(mock *mocks.MockAutheliaCtx) {
				ctx := context.Background()

				gomock.InOrder(
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq("john")).
						Return(false, nil),
					s.mock.StorageMock.EXPECT().CreateUser(gomock.Any(), gomock.Eq("john"), gomock.Eq("john@example.com"), gomock.Any()).
						Return(errors.New("error creating user")),
					s.mock.StorageMock.EXPECT().Rollback(gomock.Any()).
						Return(errors.New("error rolling back!")),
				)
			},
			authentication.ErrCreatingUser,
		},
		{
			"ShouldFailIfCantCommitChanges",
			username,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail(email)},
			func(mock *mocks.MockAutheliaCtx) {
				ctx := context.Background()

				gomock.InOrder(
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq("john")).
						Return(false, nil),
					s.mock.StorageMock.EXPECT().CreateUser(gomock.Any(), gomock.Eq("john"), gomock.Eq("john@example.com"), gomock.Any()).
						Return(nil),
					s.mock.StorageMock.EXPECT().UpdateUserGroups(gomock.Any(), gomock.Any()).
						Return(nil),
					s.mock.StorageMock.EXPECT().UpdateUserDisplayName(gomock.Any(), gomock.Eq("john"), gomock.Any()).
						Return(nil),
					s.mock.StorageMock.EXPECT().Commit(gomock.Any()).
						Return(errors.New("error committing changes!")),
				)
			},
			authentication.ErrCreatingUser,
		},
		{
			"ShouldFailIfCantUpdateDisplayName",
			username,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail(email)},
			func(mock *mocks.MockAutheliaCtx) {
				ctx := context.Background()

				gomock.InOrder(
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq("john")).
						Return(false, nil),
					s.mock.StorageMock.EXPECT().CreateUser(gomock.Any(), gomock.Eq("john"), gomock.Eq("john@example.com"), gomock.Any()).
						Return(nil),
					s.mock.StorageMock.EXPECT().UpdateUserGroups(gomock.Any(), gomock.Any()).
						Return(nil),
					s.mock.StorageMock.EXPECT().UpdateUserDisplayName(gomock.Any(), gomock.Eq("john"), gomock.Any()).
						Return(errors.New("error updating display name")),
					s.mock.StorageMock.EXPECT().Rollback(gomock.Any()).
						Return(nil),
				)
			},
			authentication.ErrCreatingUser,
		},
		{
			"ShouldFailIfCantUpdateGroups",
			username,
			password,
			displayname,
			[]func(options *authentication.NewUserDetailsOpts){authentication.WithEmail(email)},
			func(mock *mocks.MockAutheliaCtx) {
				ctx := context.Background()

				gomock.InOrder(
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq("john")).
						Return(false, nil),
					s.mock.StorageMock.EXPECT().CreateUser(gomock.Any(), gomock.Eq("john"), gomock.Eq("john@example.com"), gomock.Any()).
						Return(nil),
					s.mock.StorageMock.EXPECT().UpdateUserGroups(gomock.Any(), gomock.Any()).
						Return(errors.New("cant update user's groups")),
					s.mock.StorageMock.EXPECT().Rollback(gomock.Any()).
						Return(nil),
				)
			},
			authentication.ErrCreatingUser,
		},
	}

	provider := s.mock.Ctx.Providers.UserProvider.(*authentication.DBUserProvider)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup(s.mock)

			err := provider.AddUser(tc.username, tc.displayname, tc.password, tc.options...)

			s.ErrorIs(err, tc.expectError)
		})
	}
}

func (s *DBUserProviderSuite) TestDeleteUser() {
	var username = "john"

	var testCases = []struct {
		name        string
		setup       func(mock *mocks.MockAutheliaCtx)
		expectError error
	}{
		{
			"ShouldSuccessIfUserExists",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().DeleteUser(gomock.Any(), gomock.Eq(username)).
						Return(nil),
				)
			},
			nil,
		},
		{
			"ShouldFailIfUserNotExists",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
					Return(false, nil)
			},
			authentication.ErrUserNotFound,
		},
		{
			"ShouldFailIfErrorWhileCheckingIfUserIxists",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
					Return(false, errors.New("some error"))
			},
			authentication.ErrDeletingUser,
		},
		{
			"ShouldFailIfHaveErrorSavingChanges",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().DeleteUser(gomock.Any(), gomock.Eq(username)).
						Return(errors.New("some error error")),
				)
			},
			authentication.ErrDeletingUser,
		},
	}

	provider := s.mock.Ctx.Providers.UserProvider.(*authentication.DBUserProvider)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup(s.mock)

			err := provider.DeleteUser(username)

			s.ErrorIs(err, tc.expectError)
		})
	}
}

func (s *DBUserProviderSuite) TestChangeUserDisplayName() {
	var username, displayname = "john", "john Doe"

	var testCases = []struct {
		name      string
		setup     func(mock *mocks.MockAutheliaCtx)
		expectErr error
	}{
		{
			"ShouldSuccessIfUserExists",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().UpdateUserDisplayName(gomock.Any(), gomock.Eq(username), gomock.Eq(displayname)).
						Return(nil),
				)
			},
			nil,
		},
		{
			"ShouldFailIfUserNotExists",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
					Return(false, nil)
			},
			authentication.ErrUserNotFound,
		},
		{
			"ShouldFailIfErrorWhileCheckingIfUserIxists",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
					Return(false, errors.New("some error"))
			},
			authentication.ErrUpdatingUser,
		},
		{
			"ShouldFailIfHaveErrorSavingChanges",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().UpdateUserDisplayName(gomock.Any(), gomock.Eq(username), gomock.Eq(displayname)).
						Return(errors.New("some error saving user")),
				)
			},
			authentication.ErrUpdatingUser,
		},
	}

	provider := s.mock.Ctx.Providers.UserProvider.(*authentication.DBUserProvider)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup(s.mock)

			err := provider.ChangeDisplayName(username, displayname)

			s.ErrorIs(err, tc.expectErr)
		})
	}
}

func (s *DBUserProviderSuite) TestChangeUserEmail() {
	var username, email = "john", "john@example.com"

	var testCases = []struct {
		name        string
		setup       func(mock *mocks.MockAutheliaCtx)
		expectError error
	}{
		{
			"ShouldSuccessIfUserExists",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().UpdateUserEmail(gomock.Any(), gomock.Eq(username), gomock.Eq(email)).
						Return(nil),
				)
			},
			nil,
		},
		{
			"ShouldFailIfUserNotExists",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
					Return(false, nil)
			},
			authentication.ErrUserNotFound,
		},
		{
			"ShouldFailIfErrorWhileCheckingIfUserIxists",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
					Return(false, errors.New("some error"))
			},
			authentication.ErrUpdatingUser,
		},
		{
			"ShouldFailIfHaveErrorSavingChanges",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().UpdateUserEmail(gomock.Any(), gomock.Eq(username), gomock.Eq(email)).
						Return(errors.New("some error saving user")),
				)
			},
			authentication.ErrUpdatingUser,
		},
	}

	provider := s.mock.Ctx.Providers.UserProvider.(*authentication.DBUserProvider)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup(s.mock)

			err := provider.ChangeEmail(username, email)

			s.ErrorIs(err, tc.expectError)
		})
	}
}

func (s *DBUserProviderSuite) TestChangeUserGroups() {
	var ctx = context.Background()

	var username, groups = "john", []string{"dev", "admins"}

	var testCases = []struct {
		name        string
		setup       func(mock *mocks.MockAutheliaCtx)
		expectError error
	}{
		{
			"ShouldSuccessIfUserExists",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UpdateUserGroups(gomock.Any(), gomock.Eq(username), gomock.Eq(groups)).
						Return(nil),
					s.mock.StorageMock.EXPECT().Commit(gomock.Any()).
						Return(nil),
				)
			},
			nil,
		},
		{
			"ShouldFailIfUserNotExists",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
					Return(false, nil)
			},
			authentication.ErrUserNotFound,
		},
		{
			"ShouldFailIfErrorWhileCheckingIfUserIxists",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
					Return(false, errors.New("some error"))
			},
			authentication.ErrUpdatingUser,
		},
		{
			"ShouldFailIfCantCreateTransaction",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, errors.New("some error")),
				)
			},
			authentication.ErrUpdatingUser,
		},
		{
			"ShouldFailIfHaveErrorSavingChanges",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UpdateUserGroups(gomock.Any(), gomock.Eq(username), gomock.Eq(groups)).
						Return(errors.New("some error")),
					s.mock.StorageMock.EXPECT().Rollback(gomock.Any()).
						Return(nil),
				)
			},
			authentication.ErrUpdatingUser,
		},
		{
			"ShouldFailIfCantRollbackAfterError",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UpdateUserGroups(gomock.Any(), gomock.Eq(username), gomock.Eq(groups)).
						Return(errors.New("some error")),
					s.mock.StorageMock.EXPECT().Rollback(gomock.Any()).
						Return(errors.New("no way")),
				)
			},
			authentication.ErrUpdatingUser,
		},
		{
			"ShouldFailIfCantCommitChanges",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					s.mock.StorageMock.EXPECT().UserExists(gomock.Any(), gomock.Eq(username)).
						Return(true, nil),
					s.mock.StorageMock.EXPECT().BeginTX(gomock.Any()).
						Return(ctx, nil),
					s.mock.StorageMock.EXPECT().UpdateUserGroups(gomock.Any(), gomock.Eq(username), gomock.Eq(groups)).
						Return(nil),
					s.mock.StorageMock.EXPECT().Commit(gomock.Any()).
						Return(errors.New("commit error")),
				)
			},
			authentication.ErrUpdatingUser,
		},
	}

	provider := s.mock.Ctx.Providers.UserProvider.(*authentication.DBUserProvider)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup(s.mock)

			err := provider.ChangeGroups(username, groups)

			s.ErrorIs(err, tc.expectError)
		})
	}
}

func (s *DBUserProviderSuite) TestListUsers() {
	var userList = []model.User{
		{
			Username:    "john",
			Email:       "john@example.com",
			DisplayName: "John Doe",
			Groups:      []string{"admins", "dev"},
			Disabled:    false,
		},
		{
			Username:    "john",
			Email:       "john@example.com",
			DisplayName: "John Doe",
			Groups:      []string{"dev"},
			Disabled:    false,
		},
	}

	var testCases = []struct {
		name        string
		setup       func(mock *mocks.MockAutheliaCtx)
		expect      []model.User
		expectError error
	}{
		{
			"ShouldGetUsers",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().ListUsers(gomock.Any()).
					Return(userList, nil)
			},
			userList,
			nil,
		},
		{
			"ShouldFailIfStorageFailed",
			func(mock *mocks.MockAutheliaCtx) {
				s.mock.StorageMock.EXPECT().ListUsers(gomock.Any()).
					Return([]model.User{}, errors.New("error loading users"))
			},
			userList,
			authentication.ErrListingUser,
		},
	}

	provider := s.mock.Ctx.Providers.UserProvider.(*authentication.DBUserProvider)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setup(s.mock)

			_, err := provider.ListUsers()

			s.ErrorIs(err, tc.expectError)
		})
	}
}
