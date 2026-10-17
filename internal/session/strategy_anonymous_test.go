// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestDefaultStrategy_SaveShouldNotPersistAnonymousSessionWithoutState(t *testing.T) {
	repository := &observingRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.LastActivity = time.Now().Unix()

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Empty(t, repository.data)
	assert.Empty(t, userSession.PublicID)
	assert.Equal(t, 0, ctx.set)
	assert.Equal(t, 0, repository.deletes)
	assert.Empty(t, ctx.cleared)
	assert.NotContains(t, ctx.cookies, testName)
}

func TestDefaultStrategy_SaveShouldNotPersistAnonymousSessionWithoutStateForMalformedCookie(t *testing.T) {
	repository := &observingRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	ctx.cookies[testName] = "not!valid"

	userSession := strategy.NewDefault()

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Empty(t, repository.data)
	assert.Equal(t, 0, ctx.set)
	assert.Equal(t, 0, repository.deletes)
}

func TestDefaultStrategy_SaveShouldPersistAnonymousSessionWithState(t *testing.T) {
	now := time.Unix(1700000000, 0)

	repository := newTestRepository()
	strategy := NewStrategy(schema.SessionCookie{
		SessionCookieCommon: schema.SessionCookieCommon{
			Name:                testName,
			SameSite:            "lax",
			Expiration:          testExpiration,
			RememberMe:          testRememberMe,
			AnonymousExpiration: testAnonymousExpiration,
		},
		Domain: testDomain,
	}, clock.NewFixed(now), newTestCodec(t), repository)

	ctx := newTestContext()

	userSession := newTestAnonymousSessionWithState(strategy)

	require.NoError(t, strategy.Save(ctx, &userSession))

	require.Len(t, repository.data, 1)
	require.Len(t, repository.expirations, 1)

	for _, expiration := range repository.expirations {
		assert.Equal(t, testAnonymousExpiration, expiration)
	}

	assert.NotEmpty(t, userSession.PublicID)
	assert.NotEmpty(t, ctx.cookies[testName])
	assert.Equal(t, now.Add(testAnonymousExpiration).Unix(), userSession.Expires)

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	require.NotNil(t, actual.PasswordResetUsername)
	assert.Equal(t, testUsername, *actual.PasswordResetUsername)
	assert.Equal(t, now.Add(testAnonymousExpiration).Unix(), actual.Expires)
}

func TestDefaultStrategy_SaveShouldDiscardAnonymousSessionWhenItNoLongerHasState(t *testing.T) {
	repository := &observingRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := newTestAnonymousSessionWithState(strategy)

	require.NoError(t, strategy.Save(ctx, &userSession))
	require.Len(t, repository.data, 1)

	original := ctx.cookies[testName]

	require.NotEmpty(t, original)

	userSession.PasswordResetUsername = nil

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Empty(t, repository.data)
	assert.Empty(t, repository.publicIDs)
	// The session record and its authz index are both deleted.
	assert.Equal(t, 2, repository.deletes)
	assert.NotContains(t, ctx.cookies, testName)
	require.NotEmpty(t, ctx.cleared)

	for _, cookie := range ctx.cleared {
		assert.True(t, cookie.Expires.Before(time.Now()))
	}

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	assert.Nil(t, actual.PasswordResetUsername)

	t.Run("ShouldPersistAgainWithNewIdentifier", func(t *testing.T) {
		again := newTestAnonymousSessionWithState(strategy)

		require.NoError(t, strategy.Save(ctx, &again))

		assert.Len(t, repository.data, 1)
		assert.NotEmpty(t, ctx.cookies[testName])
		assert.NotEqual(t, original, ctx.cookies[testName])
	})
}

func TestDefaultStrategy_SaveShouldDiscardTheAuthzCookieAndIndex(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	authenticated := strategy.New(testUsername)

	authenticated.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &authenticated))
	require.Contains(t, ctx.jar, testNameAuthz)
	require.Len(t, repository.data, 2)

	anonymous := strategy.NewDefault()

	require.NoError(t, strategy.Save(ctx, &anonymous))

	assert.Empty(t, repository.data)
	assert.NotContains(t, ctx.cookies, testName)
	assert.NotContains(t, ctx.cookies, testNameAuthz)

	names := make([]string, 0, len(ctx.cleared))

	for _, cookie := range ctx.cleared {
		names = append(names, cookie.Name)
	}

	assert.ElementsMatch(t, []string{testName, testNameAuthz}, names)
}

func TestDefaultStrategy_SaveShouldReturnErrorWhenDiscardFails(t *testing.T) {
	repository := &erroringRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := newTestAnonymousSessionWithState(strategy)

	require.NoError(t, strategy.Save(ctx, &userSession))

	repository.errDelete = errTestFailure

	userSession.PasswordResetUsername = nil

	assert.EqualError(t, strategy.Save(ctx, &userSession), "error occurred deleting session from backend: bad stuff")
	assert.Contains(t, ctx.cookies, testName)
	assert.Empty(t, ctx.cleared)
}

func TestDefaultStrategy_SaveShouldApplyAnonymousExpiration(t *testing.T) {
	testCases := []struct {
		name           string
		username       string
		keepMeLoggedIn bool
		anonymous      time.Duration
		expected       time.Duration
	}{
		{
			"ShouldUseAnonymousExpirationForAnonymousSession",
			"",
			false,
			testAnonymousExpiration,
			testAnonymousExpiration,
		},
		{
			"ShouldUseAnonymousExpirationForRememberedAnonymousSession",
			"",
			true,
			testAnonymousExpiration,
			testAnonymousExpiration,
		},
		{
			"ShouldUseExpirationForAnonymousSessionWhenUnconfigured",
			"",
			false,
			0,
			testExpiration,
		},
		{
			"ShouldUseExpirationForUserSession",
			testUsername,
			false,
			testAnonymousExpiration,
			testExpiration,
		},
		{
			"ShouldUseRememberMeForRememberedUserSession",
			testUsername,
			true,
			testAnonymousExpiration,
			testRememberMe,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()

			strategy := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
				config.AnonymousExpiration = tc.anonymous
			})

			ctx := newTestContext()

			userSession := strategy.New(tc.username)

			username := testUsername

			userSession.PasswordResetUsername = &username
			userSession.KeepMeLoggedIn = tc.keepMeLoggedIn

			require.NoError(t, strategy.Save(ctx, &userSession))

			require.Len(t, repository.expirations, 1)

			for _, expiration := range repository.expirations {
				assert.Equal(t, tc.expected, expiration)
			}
		})
	}
}

func TestDefaultStrategy_RegenerateShouldApplyAnonymousExpiration(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := newTestAnonymousSessionWithState(strategy)

	require.NoError(t, strategy.Save(ctx, &userSession))
	require.NoError(t, strategy.Regenerate(ctx))

	require.Len(t, repository.expirations, 1)

	for _, expiration := range repository.expirations {
		assert.Equal(t, testAnonymousExpiration, expiration)
	}
}

func TestUserSession_RequiresPersistence(t *testing.T) {
	username := testUsername

	testCases := []struct {
		name     string
		modify   func(userSession *UserSession)
		expected bool
	}{
		{"ShouldNotRequireForDefault", func(userSession *UserSession) {}, false},
		{"ShouldNotRequireForIdentifiersAndActivity", func(userSession *UserSession) {
			userSession.CookieDomain = testDomain
			userSession.PublicID = "abc"
			userSession.LastActivity = 1
			userSession.KeepMeLoggedIn = true
			userSession.RefreshTTL = time.Unix(1, 0)
		}, false},
		{"ShouldRequireForUsername", func(userSession *UserSession) { userSession.Username = testUsername }, true},
		{"ShouldRequireForFirstFactorTimestamp", func(userSession *UserSession) { userSession.FirstFactorAuthnTimestamp = 1 }, true},
		{"ShouldRequireForSecondFactorTimestamp", func(userSession *UserSession) { userSession.SecondFactorAuthnTimestamp = 1 }, true},
		{"ShouldRequireForAMR", func(userSession *UserSession) { userSession.AuthenticationMethodRefs.WebAuthnUserVerified = true }, true},
		{"ShouldRequireForExtraAMR", func(userSession *UserSession) { userSession.AuthenticationMethodRefs.Extra = []string{"x"} }, true},
		{"ShouldRequireForWebAuthn", func(userSession *UserSession) { userSession.WebAuthn = &WebAuthn{} }, true},
		{"ShouldRequireForTOTP", func(userSession *UserSession) { userSession.TOTP = &TOTP{} }, true},
		{"ShouldRequireForPasswordReset", func(userSession *UserSession) { userSession.PasswordResetUsername = &username }, true},
		{"ShouldRequireForElevation", func(userSession *UserSession) { userSession.Elevations.User = &Elevation{} }, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userSession := NewDefaultUserSession()

			tc.modify(&userSession)

			assert.Equal(t, tc.expected, userSession.RequiresPersistence())
		})
	}
}

// TestUserSession_RequiresPersistenceShouldConsiderEveryField ensures a field added to the UserSession is deliberately
// classified, as an unclassified field holding state would otherwise be silently discarded for anonymous sessions.
func TestUserSession_RequiresPersistenceShouldConsiderEveryField(t *testing.T) {
	disregarded := map[string]bool{
		"CookieDomain":   true,
		"PublicID":       true,
		"KeepMeLoggedIn": true,
		"LastActivity":   true,
		"RefreshTTL":     true,
		"Expires":        true,
	}

	fields := reflect.TypeFor[UserSession]()

	for i := range fields.NumField() {
		field := fields.Field(i)

		t.Run(field.Name, func(t *testing.T) {
			userSession := NewDefaultUserSession()

			setTestNonZero(t, reflect.ValueOf(&userSession).Elem().Field(i))

			assert.Equal(t, !disregarded[field.Name], userSession.RequiresPersistence(), "the field '%s' must either be considered by requiresPersistence or be added to the disregarded fields of this test", field.Name)
		})
	}
}

func setTestNonZero(t *testing.T, value reflect.Value) {
	t.Helper()

	switch value.Kind() {
	case reflect.String:
		value.SetString("x")
	case reflect.Bool:
		value.SetBool(true)
	case reflect.Int, reflect.Int64:
		value.SetInt(1)
	case reflect.Pointer:
		value.Set(reflect.New(value.Type().Elem()))
	case reflect.Struct:
		if value.Type() == reflect.TypeFor[time.Time]() {
			value.Set(reflect.ValueOf(time.Unix(1, 0)))

			return
		}

		setTestNonZero(t, value.Field(0))
	default:
		t.Fatalf("unhandled kind '%s' of type '%s'", value.Kind(), value.Type())
	}
}
