// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestNewReauthenticationState(t *testing.T) {
	now := time.Unix(1700000000, 0)
	fresh := now.Add(-time.Minute).Unix()
	boundary := now.Add(-time.Minute * 5).Unix()
	stale := now.Add(-time.Minute*5 - time.Second).Unix()

	notRequired := middlewares.ReauthenticationState{Methods: []string{}}
	password := middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodPassword}}
	secondFactor := middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodSecondFactor}}
	both := middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodPassword, middlewares.ReauthenticationMethodSecondFactor}}

	testCases := []struct {
		name       string
		mode       string
		first      int64
		possession int64
		enrollment middlewares.Enrollment
		expected   middlewares.ReauthenticationState
	}{
		{"ShouldNotRequireWhenEmpty", "", 0, 0, middlewares.EnrollmentUnknown, notRequired},
		{"ShouldNotRequireWhenDisabled", schema.ElevatedSessionReauthenticationDisabled, 0, 0, middlewares.EnrollmentUnknown, notRequired},
		{"ShouldPasswordFresh", schema.ElevatedSessionReauthenticationPassword, fresh, 0, middlewares.EnrollmentUnknown, notRequired},
		{"ShouldPasswordBoundary", schema.ElevatedSessionReauthenticationPassword, boundary, 0, middlewares.EnrollmentUnknown, notRequired},
		{"ShouldPasswordStale", schema.ElevatedSessionReauthenticationPassword, stale, 0, middlewares.EnrollmentUnknown, password},
		{"ShouldPasswordNever", schema.ElevatedSessionReauthenticationPassword, 0, 0, middlewares.EnrollmentUnknown, password},
		{"ShouldPasswordIgnoreSecondFactor", schema.ElevatedSessionReauthenticationPassword, stale, fresh, middlewares.EnrollmentSecondFactor, password},
		{"ShouldSecondFactorEnrolledFresh", schema.ElevatedSessionReauthenticationSecondFactor, stale, fresh, middlewares.EnrollmentSecondFactor, notRequired},
		{"ShouldSecondFactorEnrolledStale", schema.ElevatedSessionReauthenticationSecondFactor, fresh, stale, middlewares.EnrollmentSecondFactor, secondFactor},
		{"ShouldSecondFactorNoneFallbackPasswordFresh", schema.ElevatedSessionReauthenticationSecondFactor, fresh, 0, middlewares.EnrollmentNone, notRequired},
		{"ShouldSecondFactorNoneFallbackPasswordStale", schema.ElevatedSessionReauthenticationSecondFactor, stale, 0, middlewares.EnrollmentNone, password},
		{"ShouldSecondFactorUnknownFailClosedEvenIfFresh", schema.ElevatedSessionReauthenticationSecondFactor, fresh, fresh, middlewares.EnrollmentUnknown, secondFactor},
		{"ShouldAnyEnrolledPasswordFresh", schema.ElevatedSessionReauthenticationAny, fresh, stale, middlewares.EnrollmentSecondFactor, notRequired},
		{"ShouldAnyEnrolledSecondFactorFresh", schema.ElevatedSessionReauthenticationAny, stale, fresh, middlewares.EnrollmentSecondFactor, notRequired},
		{"ShouldAnyEnrolledStale", schema.ElevatedSessionReauthenticationAny, stale, stale, middlewares.EnrollmentSecondFactor, both},
		{"ShouldAnyNonePasswordOnly", schema.ElevatedSessionReauthenticationAny, stale, fresh, middlewares.EnrollmentNone, password},
		{"ShouldAnyUnknownFailClosedPasswordOnly", schema.ElevatedSessionReauthenticationAny, fresh, fresh, middlewares.EnrollmentUnknown, password},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := schema.IdentityValidationElevatedSession{
				RequireReauthentication:  tc.mode,
				ReauthenticationLifespan: time.Minute * 5,
			}

			userSession := &session.UserSession{
				FirstFactorAuthnTimestamp:            tc.first,
				SecondFactorPossessionAuthnTimestamp: tc.possession,
			}

			assert.Equal(t, tc.expected, middlewares.NewReauthenticationState(config, userSession, now, tc.enrollment))
		})
	}
}

func TestNewReauthenticationStateShouldNotAcceptNonPossessionSecondFactor(t *testing.T) {
	now := time.Unix(1700000000, 0)
	fresh := now.Add(-time.Minute).Unix()
	stale := now.Add(-time.Hour).Unix()

	config := schema.IdentityValidationElevatedSession{
		RequireReauthentication:  schema.ElevatedSessionReauthenticationSecondFactor,
		ReauthenticationLifespan: time.Minute * 5,
	}

	userSession := &session.UserSession{
		FirstFactorAuthnTimestamp:  stale,
		SecondFactorAuthnTimestamp: fresh,
	}

	expected := middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodSecondFactor}}

	assert.Equal(t, expected, middlewares.NewReauthenticationState(config, userSession, now, middlewares.EnrollmentSecondFactor))

	config.RequireReauthentication = schema.ElevatedSessionReauthenticationAny

	expected = middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodPassword, middlewares.ReauthenticationMethodSecondFactor}}

	assert.Equal(t, expected, middlewares.NewReauthenticationState(config, userSession, now, middlewares.EnrollmentSecondFactor))
}

func TestGetReauthenticationState(t *testing.T) {
	now := time.Unix(1700000000, 0)

	testCases := []struct {
		name     string
		mode     string
		setup    func(mock *mocks.MockAutheliaCtx)
		expected middlewares.ReauthenticationState
	}{
		{
			"ShouldNotLoadUserInfoForPassword",
			schema.ElevatedSessionReauthenticationPassword,
			nil,
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodPassword}},
		},
		{
			"ShouldLoadEnrollmentForSecondFactor",
			schema.ElevatedSessionReauthenticationSecondFactor,
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{HasTOTP: true}, nil)
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodSecondFactor}},
		},
		{
			"ShouldFallbackToPasswordWhenNotEnrolled",
			schema.ElevatedSessionReauthenticationSecondFactor,
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{}, nil)
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodPassword}},
		},
		{
			"ShouldLoadEnrollmentForAny",
			schema.ElevatedSessionReauthenticationAny,
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{HasTOTP: true}, nil)
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodPassword, middlewares.ReauthenticationMethodSecondFactor}},
		},
		{
			"ShouldLoadEnrollmentForSecondFactorWebAuthnOnly",
			schema.ElevatedSessionReauthenticationSecondFactor,
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{HasWebAuthn: true}, nil)
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodSecondFactor}},
		},
		{
			"ShouldLoadEnrollmentForSecondFactorDuoOnly",
			schema.ElevatedSessionReauthenticationSecondFactor,
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{HasDuo: true}, nil)
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodSecondFactor}},
		},
		{
			"ShouldFallbackToPasswordWhenOnlyMethodTOTPDisabled",
			schema.ElevatedSessionReauthenticationSecondFactor,
			func(mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.TOTP.Disable = true
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{HasTOTP: true}, nil)
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodPassword}},
		},
		{
			"ShouldFallbackToPasswordWhenOnlyMethodWebAuthnDisabled",
			schema.ElevatedSessionReauthenticationSecondFactor,
			func(mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.WebAuthn.Disable = true
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{HasWebAuthn: true}, nil)
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodPassword}},
		},
		{
			"ShouldFallbackToPasswordWhenOnlyMethodDuoDisabled",
			schema.ElevatedSessionReauthenticationAny,
			func(mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.DuoAPI.Disable = true
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{HasDuo: true}, nil)
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodPassword}},
		},
		{
			"ShouldCountRemainingEnabledMethodWhenAnotherDisabled",
			schema.ElevatedSessionReauthenticationSecondFactor,
			func(mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Configuration.TOTP.Disable = true
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{HasTOTP: true, HasDuo: true}, nil)
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodSecondFactor}},
		},
		{
			"ShouldNotOfferPasswordOnStorageErrorForSecondFactor",
			schema.ElevatedSessionReauthenticationSecondFactor,
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{}, errors.New("bad storage"))
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodSecondFactor}},
		},
		{
			"ShouldOfferPasswordOnStorageErrorForAny",
			schema.ElevatedSessionReauthenticationAny,
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{}, errors.New("bad storage"))
			},
			middlewares.ReauthenticationState{Required: true, Methods: []string{middlewares.ReauthenticationMethodPassword}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Clock.Set(now)

			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.RequireReauthentication = tc.mode
			mock.Ctx.Configuration.IdentityValidation.ElevatedSession.ReauthenticationLifespan = time.Minute * 5

			if tc.setup != nil {
				tc.setup(mock)
			}

			userSession := &session.UserSession{Username: john}

			assert.Equal(t, tc.expected, middlewares.GetReauthenticationState(mock.Ctx, userSession))
		})
	}
}
