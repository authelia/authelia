// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestIsPasswordChangeRequired(t *testing.T) {
	testCases := []struct {
		name      string
		attribute string
		setup     func(mock *mocks.MockAutheliaCtx)
		expected  bool
	}{
		{
			"ShouldNotBeRequiredWithoutAnAttribute",
			"",
			nil,
			false,
		},
		{
			"ShouldBeRequiredWhenTheAttributeIsTrue",
			"pwd_reset",
			func(mock *mocks.MockAutheliaCtx) {
				mock.UserProviderMock.EXPECT().
					GetDetailsExtended(testUsername).
					Return(newUserDetailsExtendedWithExtra(map[string]any{"pwd_reset": true}), nil)
			},
			true,
		},
		{
			"ShouldNotBeRequiredWhenTheAttributeIsFalse",
			"pwd_reset",
			func(mock *mocks.MockAutheliaCtx) {
				mock.UserProviderMock.EXPECT().
					GetDetailsExtended(testUsername).
					Return(newUserDetailsExtendedWithExtra(map[string]any{"pwd_reset": false}), nil)
			},
			false,
		},
		{
			"ShouldNotBeRequiredWhenTheAttributeIsAbsent",
			"pwd_reset",
			func(mock *mocks.MockAutheliaCtx) {
				mock.UserProviderMock.EXPECT().
					GetDetailsExtended(testUsername).
					Return(newUserDetailsExtendedWithExtra(map[string]any{}), nil)
			},
			false,
		},
		{
			"ShouldNotBeRequiredWhenTheAttributeIsNotABoolean",
			"pwd_reset",
			func(mock *mocks.MockAutheliaCtx) {
				mock.UserProviderMock.EXPECT().
					GetDetailsExtended(testUsername).
					Return(newUserDetailsExtendedWithExtra(map[string]any{"pwd_reset": "TRUE"}), nil)
			},
			false,
		},
		{
			"ShouldNotBeRequiredWhenTheDetailsAreUnavailable",
			"pwd_reset",
			func(mock *mocks.MockAutheliaCtx) {
				mock.UserProviderMock.EXPECT().
					GetDetailsExtended(testUsername).
					Return(nil, fmt.Errorf("failed to mock the details"))
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := newMockAutheliaCtxWithUserAttributes(t)

			defer mock.Close()

			mock.Ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute = tc.attribute

			if tc.setup != nil {
				tc.setup(mock)
			}

			assert.Equal(t, tc.expected, isPasswordChangeRequired(mock.Ctx, testUsername))
		})
	}
}

func TestClearPasswordChangeRequired(t *testing.T) {
	testCases := []struct {
		name      string
		attribute string
		setup     func(mock *mocks.MockAutheliaCtx)
		logged    string
	}{
		{
			"ShouldNotClearWithoutAClearAttribute",
			"",
			nil,
			"",
		},
		{
			"ShouldClearWhenTheAttributeStillHoldsTheUser",
			"pwd_reset",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.UserProviderMock.EXPECT().
						GetDetailsExtended(testUsername).
						Return(newUserDetailsExtendedWithExtra(map[string]any{"pwd_reset": true}), nil),
					mock.UserProviderMock.EXPECT().
						ClearExtraAttribute(testUsername, "pwd_reset").
						Return(nil),
				)
			},
			"",
		},
		{
			"ShouldNotClearWhenTheBackendHasClearedTheAttribute",
			"pwd_reset",
			func(mock *mocks.MockAutheliaCtx) {
				mock.UserProviderMock.EXPECT().
					GetDetailsExtended(testUsername).
					Return(newUserDetailsExtendedWithExtra(map[string]any{"pwd_reset": false}), nil)
			},
			"",
		},
		{
			"ShouldLogAnErrorWhenTheClearFails",
			"pwd_reset",
			func(mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.UserProviderMock.EXPECT().
						GetDetailsExtended(testUsername).
						Return(newUserDetailsExtendedWithExtra(map[string]any{"pwd_reset": true}), nil),
					mock.UserProviderMock.EXPECT().
						ClearExtraAttribute(testUsername, "pwd_reset").
						Return(fmt.Errorf("failed to mock the clear")),
				)
			},
			"Error occurred clearing the attribute which requires the user change their password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := newMockAutheliaCtxWithUserAttributes(t)

			defer mock.Close()

			mock.Ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute = "pwd_reset"
			mock.Ctx.Configuration.AuthenticationBackend.PasswordChange.ClearAttribute = tc.attribute

			if tc.setup != nil {
				tc.setup(mock)
			}

			clearPasswordChangeRequired(mock.Ctx, testUsername)

			if tc.logged == "" {
				assert.Empty(t, mock.Hook.Entries)

				return
			}

			mock.AssertLogEntryAdvanced(t, 0, logrus.ErrorLevel, tc.logged, map[string]any{"username": testUsername, "attribute": tc.attribute})
		})
	}
}

func newMockAutheliaCtxWithUserAttributes(t *testing.T) (mock *mocks.MockAutheliaCtx) {
	mock = mocks.NewMockAutheliaCtx(t)

	mock.Ctx.Providers.UserAttributeResolver = &expression.UserAttributes{}

	return mock
}

func newUserDetailsExtendedWithExtra(extra map[string]any) (details *authentication.UserDetailsExtended) {
	return &authentication.UserDetailsExtended{
		UserDetails: &authentication.UserDetails{Username: testUsername},
		Extra:       extra,
	}
}
