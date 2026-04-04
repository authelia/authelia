package oidc_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestSubjectUUIDFromSubjectString(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected uuid.UUID
		err      bool
	}{
		{"ShouldParseValidUUID", "fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70", uuid.MustParse("fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70"), false},
		{"ShouldBeDeterministic", "fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70", uuid.MustParse("fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70"), false},
		{"ShouldErrInvalidUUID", "not-a-uuid", uuid.UUID{}, true},
		{"ShouldErrEmptyString", "", uuid.UUID{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subject, err := oidc.SubjectUUIDFromSubjectString(tc.value)

			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expected, subject)
		})
	}
}

func TestSubjectUUIDFromClaims(t *testing.T) {
	testCases := []struct {
		name     string
		claims   map[string]any
		expected uuid.UUID
		err      bool
	}{
		{"ShouldParseValidSubjectClaim", map[string]any{oidc.ClaimSubject: "fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70"}, uuid.MustParse("fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70"), false},
		{"ShouldErrMissingSubjectClaim", map[string]any{}, uuid.UUID{}, true},
		{"ShouldErrSubjectClaimNotString", map[string]any{oidc.ClaimSubject: 12345}, uuid.UUID{}, true},
		{"ShouldErrSubjectClaimInvalidUUID", map[string]any{oidc.ClaimSubject: "not-a-uuid"}, uuid.UUID{}, true},
		{"ShouldErrNilClaims", nil, uuid.UUID{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subject, err := oidc.SubjectUUIDFromClaims(tc.claims)

			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expected, subject)
		})
	}
}

func TestUserDetailerFromSubject(t *testing.T) {
	subjectUUID := uuid.MustParse("fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70")

	testCases := []struct {
		name         string
		setup        func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider)
		subject      uuid.UUID
		err          bool
		expectedUser string
	}{
		{
			"ShouldSucceedReturningDetailer",
			func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider) {
				store := mocks.NewMockStorage(ctrl)
				up := mocks.NewMockUserProvider(ctrl)

				store.EXPECT().LoadUserOpaqueIdentifier(gomock.Any(), subjectUUID).Return(&model.UserOpaqueIdentifier{
					Username: "john",
				}, nil)

				up.EXPECT().GetDetailsExtended("john").Return(&authentication.UserDetailsExtended{}, nil)

				return store, up
			},
			subjectUUID,
			false,
			"john",
		},
		{
			"ShouldErrWhenStorageFails",
			func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider) {
				store := mocks.NewMockStorage(ctrl)
				up := mocks.NewMockUserProvider(ctrl)

				store.EXPECT().LoadUserOpaqueIdentifier(gomock.Any(), subjectUUID).Return(nil, fmt.Errorf("storage error"))

				return store, up
			},
			subjectUUID,
			true,
			"",
		},
		{
			"ShouldErrWhenUserProviderFails",
			func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider) {
				store := mocks.NewMockStorage(ctrl)
				up := mocks.NewMockUserProvider(ctrl)

				store.EXPECT().LoadUserOpaqueIdentifier(gomock.Any(), subjectUUID).Return(&model.UserOpaqueIdentifier{
					Username: "john",
				}, nil)

				up.EXPECT().GetDetailsExtended("john").Return(nil, fmt.Errorf("user not found"))

				return store, up
			},
			subjectUUID,
			true,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store, up := tc.setup(ctrl)

			ctx := &TestContext{
				Storage:      store,
				UserProvider: up,
			}

			detailer, err := oidc.UserDetailerFromSubject(ctx, tc.subject)

			if tc.err {
				assert.Error(t, err)
				assert.Nil(t, detailer)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, detailer)
			}
		})
	}
}

func TestUserDetailerFromSubjectString(t *testing.T) {
	subjectUUID := uuid.MustParse("fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70")

	testCases := []struct {
		name    string
		setup   func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider)
		subject string
		err     bool
	}{
		{
			"ShouldSucceed",
			func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider) {
				store := mocks.NewMockStorage(ctrl)
				up := mocks.NewMockUserProvider(ctrl)

				store.EXPECT().LoadUserOpaqueIdentifier(gomock.Any(), subjectUUID).Return(&model.UserOpaqueIdentifier{
					Username: "john",
				}, nil)

				up.EXPECT().GetDetailsExtended("john").Return(&authentication.UserDetailsExtended{}, nil)

				return store, up
			},
			"fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70",
			false,
		},
		{
			"ShouldErrInvalidUUID",
			func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider) {
				return mocks.NewMockStorage(ctrl), mocks.NewMockUserProvider(ctrl)
			},
			"not-a-uuid",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store, up := tc.setup(ctrl)

			ctx := &TestContext{
				Storage:      store,
				UserProvider: up,
			}

			detailer, err := oidc.UserDetailerFromSubjectString(ctx, tc.subject)

			if tc.err {
				assert.Error(t, err)
				assert.Nil(t, detailer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, detailer)
			}
		})
	}
}

func TestUserDetailerFromClaims(t *testing.T) {
	subjectUUID := uuid.MustParse("fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70")

	testCases := []struct {
		name   string
		setup  func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider)
		claims map[string]any
		err    bool
	}{
		{
			"ShouldSucceed",
			func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider) {
				store := mocks.NewMockStorage(ctrl)
				up := mocks.NewMockUserProvider(ctrl)

				store.EXPECT().LoadUserOpaqueIdentifier(gomock.Any(), subjectUUID).Return(&model.UserOpaqueIdentifier{
					Username: "john",
				}, nil)

				up.EXPECT().GetDetailsExtended("john").Return(&authentication.UserDetailsExtended{}, nil)

				return store, up
			},
			map[string]any{oidc.ClaimSubject: "fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70"},
			false,
		},
		{
			"ShouldErrMissingSubject",
			func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider) {
				return mocks.NewMockStorage(ctrl), mocks.NewMockUserProvider(ctrl)
			},
			map[string]any{},
			true,
		},
		{
			"ShouldErrInvalidSubjectUUID",
			func(ctrl *gomock.Controller) (*mocks.MockStorage, *mocks.MockUserProvider) {
				return mocks.NewMockStorage(ctrl), mocks.NewMockUserProvider(ctrl)
			},
			map[string]any{oidc.ClaimSubject: "not-a-uuid"},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store, up := tc.setup(ctrl)

			ctx := &TestContext{
				Storage:      store,
				UserProvider: up,
			}

			detailer, err := oidc.UserDetailerFromClaims(ctx, tc.claims)

			if tc.err {
				assert.Error(t, err)
				assert.Nil(t, detailer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, detailer)
			}
		})
	}
}
