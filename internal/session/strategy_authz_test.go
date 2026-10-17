// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestDefaultAuthzStrategy_GetShouldResolveTheSessionFromTheAuthzCookie(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestSavedContext(t, strategy, testUsername)

	actual, err := strategy.Authz().Get(newTestAuthzContext(ctx))

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
	assert.Equal(t, testDomain, actual.CookieDomain)
}

func TestDefaultAuthzStrategy_GetShouldReturnAnAnonymousSession(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T, strategy Strategy, repository *testRepository, saved, ctx *testContext)
	}{
		{
			"ShouldHandleNoCookie",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, ctx *testContext) {},
		},
		{
			"ShouldHandleOnlyThePrimaryCookie",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, ctx *testContext) {
				ctx.cookies[testName] = saved.cookies[testName]
			},
		},
		{
			"ShouldHandleThePrimaryValueInTheAuthzCookie",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, ctx *testContext) {
				ctx.cookies[testNameAuthz] = saved.cookies[testName]
			},
		},
		{
			"ShouldHandleCookieWhichIsNotBase64",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, ctx *testContext) {
				ctx.cookies[testNameAuthz] = "not base64!"
			},
		},
		{
			"ShouldHandleCookieOfTheWrongLength",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, ctx *testContext) {
				ctx.cookies[testNameAuthz] = base64.RawURLEncoding.EncodeToString([]byte("short"))
			},
		},
		{
			"ShouldHandleUnknownCookie",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, ctx *testContext) {
				ctx.cookies[testNameAuthz] = newTestCookie("unknown")
			},
		},
		{
			"ShouldHandleDanglingPointer",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, ctx *testContext) {
				ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

				delete(repository.data, testPrimaryKey(t, strategy, saved))
			},
		},
		{
			"ShouldHandleValueFromBeforeRegenerate",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, ctx *testContext) {
				ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

				require.NoError(t, strategy.Regenerate(saved))
			},
		},
		{
			"ShouldHandlePointerToAnotherSession",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, ctx *testContext) {
				ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

				other := newTestSavedContext(t, strategy, "harry")

				repository.data[testPointerKey(t, strategy, saved)] = []byte(getTestSessionSignature(t, strategy, other))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()
			strategy := newTestStrategyWithRepository(t, repository, nil)
			saved := newTestSavedContext(t, strategy, testUsername)
			ctx := newTestContext()

			tc.setup(t, strategy, repository, saved, ctx)

			actual, err := strategy.Authz().Get(ctx)

			require.NoError(t, err)
			assert.True(t, actual.IsAnonymous())
			assert.Equal(t, testDomain, actual.CookieDomain)
		})
	}
}

func TestDefaultAuthzStrategy_GetShouldReturnRepositoryErrorWhenTheBackendFails(t *testing.T) {
	strategy := newTestStrategyWithRepository(t, &failingRepository{testRepository: newTestRepository()}, nil)
	ctx := newTestContext()

	ctx.authz = true
	ctx.cookies[testNameAuthz] = newTestCookie("any")

	actual, err := strategy.Authz().Get(ctx)

	assert.Nil(t, actual)
	assert.ErrorIs(t, err, ErrRepositoryGet)
}

func TestDefaultAuthzStrategy_GetShouldNotShareTheCacheOfThePrimaryStrategy(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	ctx := newTestCachingContext()

	ctx.cookies[testName] = saved.cookies[testName]
	ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

	_, err := strategy.Get(ctx)
	require.NoError(t, err)

	ctx.authz = true

	_, err = strategy.Authz().Get(ctx)
	require.NoError(t, err)

	assert.Contains(t, ctx.sessions, testDomain)
	assert.Contains(t, ctx.sessions, cacheKeyPrefixAuthz+testDomain)
}

func TestDefaultAuthzStrategy_SaveActivityShouldOnlyPersistActivity(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)
	ctx := newTestAuthzContext(saved)

	view := strategy.Authz()

	userSession, err := view.Get(ctx)
	require.NoError(t, err)

	refresh := time.Unix(200, 0)

	userSession.LastActivity = 100
	userSession.RefreshTTL = refresh
	userSession.Username = "harry"
	userSession.KeepMeLoggedIn = true
	userSession.AuthenticationMethodRefs.TOTP = true

	repository.expirations[testPrimaryKey(t, strategy, saved)] = 0
	repository.expirations[testPointerKey(t, strategy, saved)] = 0

	require.NoError(t, view.SaveActivity(ctx, userSession))

	assert.Equal(t, 0, ctx.set)
	assert.Empty(t, ctx.jar)
	assert.InDelta(t, testExpiration.Seconds(), repository.expirations[testPrimaryKey(t, strategy, saved)].Seconds(), 2)
	assert.LessOrEqual(t, repository.expirations[testPrimaryKey(t, strategy, saved)], testExpiration)
	assert.Equal(t, saved.cookies[testNameAuthz], ctx.cookies[testNameAuthz])
	assert.NotContains(t, ctx.cookies, testName)
	assert.Equal(t, repository.expirations[testPrimaryKey(t, strategy, saved)], repository.expirations[testPointerKey(t, strategy, saved)])

	actual, err := strategy.Get(saved)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
	assert.Equal(t, int64(100), actual.LastActivity)
	assert.True(t, refresh.Equal(actual.RefreshTTL))
	assert.False(t, actual.KeepMeLoggedIn)
	assert.False(t, actual.AuthenticationMethodRefs.TOTP)
}

func TestDefaultAuthzStrategy_SaveActivityShouldNotCreateASession(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T, strategy Strategy, saved *testContext)
	}{
		{"ShouldHandleAnonymousRequest", func(t *testing.T, strategy Strategy, saved *testContext) {
			delete(saved.cookies, testNameAuthz)
		}},
		{"ShouldHandleSessionDestroyedByAnotherRequest", func(t *testing.T, strategy Strategy, saved *testContext) {
			other := newTestContext()

			other.cookies[testName] = saved.cookies[testName]
			other.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

			require.NoError(t, strategy.Destroy(other))
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()
			strategy := newTestStrategyWithRepository(t, repository, nil)
			saved := newTestSavedContext(t, strategy, testUsername)

			view := strategy.Authz()

			userSession, err := view.Get(newTestAuthzContext(saved))
			require.NoError(t, err)

			tc.setup(t, strategy, saved)

			records := len(repository.data)

			ctx := newTestAuthzContext(saved)

			userSession.LastActivity = 100

			require.NoError(t, view.SaveActivity(ctx, userSession))

			assert.Equal(t, 0, ctx.set)
			assert.Len(t, repository.data, records)
		})
	}
}

func TestDefaultAuthzStrategy_SaveActivityShouldHandleRepositoryFailures(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		vanish   bool
		expected string
		set      int
	}{
		{"ShouldHandleSuperseded", ErrSessionSuperseded, false, "", 0},
		{"ShouldReturnGenericErrorWhenSessionResolves", errors.New("backend failure"), false, "error occurred saving session activity", 0},
		{"ShouldIgnoreGenericErrorWhenSessionVanished", errors.New("backend failure"), true, "", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := &failingSaveRepository{testRepository: newTestRepository(), err: tc.err}
			strategy := newTestStrategyWithRepository(t, repository, nil)
			saved := newTestSavedContext(t, strategy, testUsername)

			view := strategy.Authz()

			ctx := newTestCachingContext()

			ctx.authz = true
			ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

			userSession, err := view.Get(ctx)
			require.NoError(t, err)

			repository.fail = true

			if tc.vanish {
				repository.vanish = testPrimaryKey(t, strategy, saved)
			}

			err = view.SaveActivity(ctx, userSession)

			if tc.expected == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.expected)
			}

			assert.Equal(t, tc.set, ctx.set)
			assert.NotContains(t, ctx.sessions, cacheKeyPrefixAuthz+testDomain)
		})
	}
}

func TestDefaultAuthzStrategy_SaveActivityShouldNotRefreshTheCookieOfAMovedSession(t *testing.T) {
	repository := &failingSaveRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	view := strategy.Authz()

	ctx := newTestCachingContext()

	ctx.authz = true
	ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

	userSession, err := view.Get(ctx)

	require.NoError(t, err)
	require.Equal(t, testUsername, userSession.Username)

	repository.fail = true
	repository.vanish = testPrimaryKey(t, strategy, saved)

	userSession.LastActivity = 100

	require.NoError(t, view.SaveActivity(ctx, userSession))

	assert.Equal(t, 0, ctx.set)
	assert.Empty(t, ctx.jar)
	assert.NotContains(t, ctx.sessions, cacheKeyPrefixAuthz+testDomain)
}

func TestDefaultAuthzStrategy_GetShouldReturnRepositoryErrorWhenTheSessionRecordFails(t *testing.T) {
	repository := &failingSessionRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	repository.fail = true

	actual, err := strategy.Authz().Get(newTestAuthzContext(saved))

	assert.Nil(t, actual)
	assert.ErrorIs(t, err, ErrRepositoryGet)
}

func TestDefaultAuthzStrategy_DestroyShouldReturnRepositoryErrorWhenTheSessionRecordFails(t *testing.T) {
	repository := &failingSessionRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)
	ctx := newTestAuthzContext(saved)

	repository.fail = true

	err := strategy.Authz().Destroy(ctx)

	assert.ErrorIs(t, err, ErrRepositoryGet)
	require.Len(t, ctx.cleared, 2)
	assert.Contains(t, repository.data, testPrimaryKey(t, strategy, saved))
	assert.Contains(t, repository.data, testPointerKey(t, strategy, saved))
}

func TestDefaultAuthzStrategy_DestroyShouldClearBothCookiesWhenTheSessionDeleteFails(t *testing.T) {
	repository := &failingPointerDeleteRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	p, ok := strategy.(*DefaultStrategy)

	require.True(t, ok)

	repository.issuer = p.issuer

	ctx := newTestCachingContext()

	ctx.cookies[testName] = saved.cookies[testName]
	ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

	_, err := strategy.Get(ctx)

	require.NoError(t, err)

	ctx.authz = true

	_, err = strategy.Authz().Get(ctx)

	require.NoError(t, err)
	require.Len(t, ctx.sessions, 2)

	err = strategy.Authz().Destroy(ctx)

	assert.ErrorIs(t, err, errTestFailure)
	assert.ErrorContains(t, err, "error occurred deleting session from backend")
	assert.Equal(t, 1, repository.failed)

	require.Len(t, ctx.cleared, 2)

	assert.Equal(t, testNameAuthz, ctx.cleared[0].Name)
	assert.Equal(t, testName, ctx.cleared[1].Name)
	assert.Empty(t, ctx.cookies)
	assert.Empty(t, ctx.sessions)
}

func TestDefaultAuthzStrategy_DestroyShouldRemoveTheSessionAndClearBothCookies(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)
	ctx := newTestAuthzContext(saved)

	require.NoError(t, strategy.Authz().Destroy(ctx))

	require.Len(t, ctx.cleared, 2)

	assert.Equal(t, testNameAuthz, ctx.cleared[0].Name)
	assert.Equal(t, testName, ctx.cleared[1].Name)
	assert.Empty(t, repository.data)

	actual, err := strategy.Get(saved)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
}

func TestDefaultAuthzStrategy_DestroyShouldNotDeleteASessionItCannotResolve(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)
	other := newTestSavedContext(t, strategy, "harry")

	repository.data[testPointerKey(t, strategy, saved)] = []byte(getTestSessionSignature(t, strategy, other))

	require.NoError(t, strategy.Authz().Destroy(newTestAuthzContext(saved)))

	actual, err := strategy.Get(other)

	require.NoError(t, err)
	assert.Equal(t, "harry", actual.Username)
}

func TestEncapsulatedAuthzSession(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	saved := newTestSavedContext(t, strategy, testUsername)
	ctx := newTestAuthzContext(saved)

	manager := NewEncapsulatedAuthzSession(strategy.Authz(), ctx)

	assert.Equal(t, testDomain, manager.GetSessionConfig().Domain)

	anonymous := manager.NewDefaultUserSession()

	assert.True(t, anonymous.IsAnonymous())

	userSession, err := manager.GetSession()

	require.NoError(t, err)
	assert.Equal(t, testUsername, userSession.Username)

	userSession.LastActivity = 100

	require.NoError(t, manager.SaveSessionActivity(&userSession))
	require.NoError(t, manager.DestroySession())

	userSession, err = manager.GetSession()

	require.NoError(t, err)
	assert.True(t, userSession.IsAnonymous())
}

func newTestSavedContext(t *testing.T, strategy Strategy, username string) (ctx *testContext) {
	t.Helper()

	ctx = newTestContext()

	userSession := strategy.New(username)

	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	return ctx
}

func newTestAuthzContext(saved *testContext) (ctx *testContext) {
	ctx = newTestContext()

	ctx.authz = true
	ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

	return ctx
}

func getTestSessionSignature(t *testing.T, strategy Strategy, saved *testContext) (sid string) {
	t.Helper()

	p, ok := strategy.(*DefaultStrategy)
	require.True(t, ok)

	id, err := base64.RawURLEncoding.DecodeString(saved.cookies[testName])
	require.NoError(t, err)

	return p.codec.Sign(id)
}

type failingSessionRepository struct {
	*testRepository

	fail  bool
	calls int
}

func (r *failingSessionRepository) Get(ctx context.Context, issuer, id string) (record Record, err error) {
	if r.fail {
		r.calls++

		if r.calls >= 2 {
			return nil, context.DeadlineExceeded
		}
	}

	return r.testRepository.Get(ctx, issuer, id)
}

type failingSaveRepository struct {
	*testRepository

	err    error
	fail   bool
	vanish string
}

func (r *failingSaveRepository) SaveData(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, previous, data []byte) (err error) {
	if !r.fail {
		return r.testRepository.SaveData(ctx, issuer, id, pid, username, expiration, previous, data)
	}

	if r.vanish != "" {
		delete(r.data, r.vanish)
	}

	return r.err
}

func TestDefaultAuthzStrategy_SaveActivityShouldNotDiscardAConcurrentSave(t *testing.T) {
	repository := &racingSaveDataRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)
	ctx := newTestAuthzContext(saved)

	view := strategy.Authz()

	userSession, err := view.Get(ctx)
	require.NoError(t, err)

	// Another request raises the authentication level of the session after the activity request read it but before it
	// writes it back.
	repository.before = func() {
		portal := newTestContext()

		portal.cookies[testName] = saved.cookies[testName]

		current, err := strategy.Get(portal)
		require.NoError(t, err)

		current.SetTwoFactorTOTP(time.Now())

		require.NoError(t, strategy.Save(portal, current))
	}

	userSession.LastActivity = 100

	require.NoError(t, view.SaveActivity(ctx, userSession))
	require.True(t, repository.raced)

	actual, err := strategy.Get(saved)

	require.NoError(t, err)
	assert.True(t, actual.AuthenticationMethodRefs.TOTP)
	assert.NotEqual(t, int64(100), actual.LastActivity)
}

type racingSaveDataRepository struct {
	*testRepository

	before func()
	raced  bool
}

func (r *racingSaveDataRepository) SaveData(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, previous, data []byte) (err error) {
	if r.before != nil && !r.raced {
		r.raced = true

		r.before()
	}

	return r.testRepository.SaveData(ctx, issuer, id, pid, username, expiration, previous, data)
}

func TestDefaultAuthzStrategy_SaveActivityShouldNotExtendTheSessionExpiry(t *testing.T) {
	now := time.Unix(1700000000, 0)
	fixed := clock.NewFixed(now)
	repository := newTestRepository()

	config := schema.SessionCookie{
		SessionCookieCommon: schema.SessionCookieCommon{Name: testName, SameSite: "lax", Expiration: time.Minute, RememberMe: testRememberMe},
		Domain:              testDomain,
	}

	strategy := NewStrategy(config, fixed, newTestCodec(t), repository)

	portal := newTestContext()

	userSession := strategy.New(testUsername)

	userSession.SetOneFactorPassword(now, false)

	require.NoError(t, strategy.Save(portal, &userSession))

	assert.Equal(t, now.Add(time.Minute).Unix(), userSession.Expires)

	primary, pointer := testPrimaryKey(t, strategy, portal), testPointerKey(t, strategy, portal)

	fixed.Set(now.Add(time.Second * 20))

	ctx := newTestAuthzContext(portal)

	activity, err := strategy.Authz().Get(ctx)
	require.NoError(t, err)

	activity.LastActivity = fixed.Now().Unix()

	require.NoError(t, strategy.Authz().SaveActivity(ctx, activity))

	assert.Equal(t, time.Second*40, repository.expirations[primary])
	assert.Equal(t, time.Second*40, repository.expirations[pointer])

	fixed.Set(now.Add(time.Second * 30))

	current, err := strategy.Get(portal)
	require.NoError(t, err)

	require.NoError(t, strategy.Save(portal, current))

	assert.Equal(t, time.Minute, repository.expirations[primary])
	assert.Equal(t, fixed.Now().Add(time.Minute).Unix(), current.Expires)

	fixed.Set(now.Add(time.Second * 120))

	ctx = newTestAuthzContext(portal)

	expired, err := strategy.Authz().Get(ctx)
	require.NoError(t, err)

	expired.LastActivity = fixed.Now().Unix()

	// A session which has already expired is never written back, which would otherwise extend it.
	require.NoError(t, strategy.Authz().SaveActivity(ctx, expired))

	assert.Equal(t, time.Minute, repository.expirations[primary])
}
