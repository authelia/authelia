// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStrategy_GetShouldHealTheAuthzCookie(t *testing.T) {
	testCases := []struct {
		name string
		have string
	}{
		{"ShouldHandleMissingCookie", ""},
		{"ShouldHandlePlantedCookie", newTestCookie("planted")},
		{"ShouldHandleMalformedCookie", "not base64!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()
			strategy := newTestStrategyWithRepository(t, repository, nil)
			saved := newTestAuthenticatedContext(t, strategy)

			ctx := newTestContext()

			ctx.cookies[testNameSession] = saved.cookies[testNameSession]

			if tc.have != "" {
				ctx.cookies[testNameAuthz] = tc.have
			}

			actual, err := strategy.Get(ctx)

			require.NoError(t, err)
			assert.Equal(t, testUsername, actual.Username)
			assert.False(t, actual.IsAnonymous())

			assert.Equal(t, 1, ctx.set)
			assert.NotContains(t, ctx.jar, testNameSession)
			require.NotNil(t, ctx.jar[testNameAuthz])

			healed := ctx.jar[testNameAuthz]

			assert.Equal(t, saved.cookies[testNameAuthz], healed.Value)
			assert.NotEqual(t, tc.have, healed.Value)
			assert.Equal(t, http.SameSiteLaxMode, healed.SameSite)
			assert.Equal(t, "."+testDomain, healed.Domain)
			assert.False(t, healed.Expires.IsZero())
			assert.Equal(t, saved.cookies[testNameSession], ctx.cookies[testNameSession])
			assert.Len(t, repository.data, 2)

			view := strategy.Authz()

			resolved, err := view.Get(newTestAuthzContext(ctx))

			require.NoError(t, err)
			assert.Equal(t, testUsername, resolved.Username)
			assert.False(t, resolved.IsAnonymous())

			previous, err := view.Get(newTestAuthzContext(saved))

			require.NoError(t, err)
			assert.Equal(t, testUsername, previous.Username)

			presented := newTestContext()

			presented.authz = true
			presented.cookies[testNameAuthz] = tc.have

			planted, err := view.Get(presented)

			require.NoError(t, err)
			assert.True(t, planted.IsAnonymous())

			again, err := strategy.Get(ctx)

			require.NoError(t, err)
			assert.Equal(t, testUsername, again.Username)
			assert.Equal(t, 1, ctx.set)
		})
	}
}

func TestDefaultStrategy_GetShouldCacheTheHealedSession(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	saved := newTestAuthenticatedContext(t, strategy)

	ctx := newTestCachingContext()

	ctx.cookies[testNameSession] = saved.cookies[testNameSession]

	anonymous, err := strategy.Authz().Get(ctx)

	require.NoError(t, err)
	require.True(t, anonymous.IsAnonymous())

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	require.Contains(t, ctx.sessions, testDomain)
	assert.Equal(t, testUsername, actual.Username)

	assert.Equal(t, saved.cookies[testNameAuthz], ctx.cookies[testNameAuthz])

	ctx.authz = true

	resolved, err := strategy.Authz().Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, testUsername, resolved.Username)
}

func TestDefaultStrategy_GetShouldNotHealTheAuthzCookie(t *testing.T) {
	testCases := []struct {
		name          string
		authenticated bool
		authz         bool
	}{
		{"ShouldHandleMatchingCookie", true, true},
		{"ShouldHandleAnonymousSessionWithoutTheCookie", false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := &observingRepository{testRepository: newTestRepository()}
			strategy := newTestStrategyWithRepository(t, repository, nil)

			var saved *testContext

			if tc.authenticated {
				saved = newTestAuthenticatedContext(t, strategy)
			} else {
				saved = newTestContext()

				anonymous := strategy.New(testUsername)

				require.NoError(t, strategy.Save(saved, &anonymous))
				require.NotContains(t, saved.cookies, testNameAuthz)
			}

			ctx := newTestContext()

			ctx.cookies[testNameSession] = saved.cookies[testNameSession]

			if tc.authz {
				ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]
			}

			saves, deletes, changes := repository.saves, repository.deletes, repository.changes

			actual, err := strategy.Get(ctx)

			require.NoError(t, err)
			assert.Equal(t, testUsername, actual.Username)
			assert.Equal(t, tc.authenticated, !actual.IsAnonymous())

			assert.Equal(t, 0, ctx.set)
			assert.Empty(t, ctx.jar)
			assert.Equal(t, saves, repository.saves)
			assert.Equal(t, deletes, repository.deletes)
			assert.Equal(t, changes, repository.changes)

			resolved, err := strategy.Authz().Get(newTestAuthzContext(saved))

			require.NoError(t, err)
			assert.Equal(t, tc.authenticated, !resolved.IsAnonymous())
		})
	}
}

func TestDefaultStrategy_GetShouldIgnoreAFailureToHealTheAuthzCookie(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"ShouldHandleRepositoryFailure", errTestFailure},
		{"ShouldHandleSuperseded", ErrSessionSuperseded},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := &failingIndexSaveRepository{testRepository: newTestRepository(), err: tc.err}
			strategy := newTestStrategyWithRepository(t, repository, nil)
			saved := newTestAuthenticatedContext(t, strategy)

			ctx := newTestCachingContext()

			ctx.cookies[testNameSession] = saved.cookies[testNameSession]

			repository.fail = true

			actual, err := strategy.Get(ctx)

			require.NoError(t, err)
			assert.Equal(t, testUsername, actual.Username)
			assert.False(t, actual.IsAnonymous())

			assert.Equal(t, 0, ctx.set)
			assert.NotContains(t, ctx.cookies, testNameAuthz)

			require.Contains(t, ctx.sessions, testDomain)

			resolved, err := strategy.Authz().Get(newTestAuthzContext(saved))

			require.NoError(t, err)
			assert.Equal(t, testUsername, resolved.Username)
		})
	}
}

func TestDefaultStrategy_GetShouldHealConcurrentRequestsWithTheSameAuthzCookie(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestAuthenticatedContext(t, strategy)

	primary := testPrimaryKey(t, strategy, saved)
	stored := append([]byte(nil), repository.data[primary]...)

	first, second := newTestContext(), newTestContext()

	first.cookies[testNameSession] = saved.cookies[testNameSession]
	second.cookies[testNameSession] = saved.cookies[testNameSession]

	_, err := strategy.Get(first)
	require.NoError(t, err)

	_, err = strategy.Get(second)
	require.NoError(t, err)

	require.NotEmpty(t, first.cookies[testNameAuthz])
	assert.Equal(t, first.cookies[testNameAuthz], second.cookies[testNameAuthz])
	assert.Equal(t, saved.cookies[testNameAuthz], first.cookies[testNameAuthz])

	assert.Equal(t, stored, repository.data[primary])
	assert.Len(t, repository.data, 2)

	for _, ctx := range []*testContext{first, second, saved} {
		resolved, err := strategy.Authz().Get(newTestAuthzContext(ctx))

		require.NoError(t, err)
		assert.Equal(t, testUsername, resolved.Username)
	}
}

func TestDefaultStrategy_SaveShouldNotExtendTheSessionBeyondItsAuthzIndex(t *testing.T) {
	inner := newTestRepository()
	repository := &failingAuthzIndexSaveRepository{testRepository: inner}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestAuthenticatedContext(t, strategy)

	repository.issuer = strategy.(*DefaultStrategy).authzIssuer

	primary := testPrimaryKey(t, strategy, saved)
	stored := append([]byte(nil), inner.data[primary]...)
	expiration := inner.expirations[primary]

	ctx := newTestContext()

	ctx.cookies[testNameSession] = saved.cookies[testNameSession]
	ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

	userSession, err := strategy.Get(ctx)
	require.NoError(t, err)

	repository.fail = true

	userSession.SetTwoFactorTOTP(time.Now())

	err = strategy.Save(ctx, userSession)

	assert.ErrorIs(t, err, errTestFailure)
	assert.Equal(t, stored, inner.data[primary])
	assert.Equal(t, expiration, inner.expirations[primary])
	assert.Equal(t, saved.cookies[testNameAuthz], ctx.cookies[testNameAuthz])
	assert.Empty(t, ctx.cleared)

	resolved, err := strategy.Authz().Get(newTestAuthzContext(saved))

	require.NoError(t, err)
	assert.Equal(t, testUsername, resolved.Username)
	assert.False(t, resolved.AuthenticationMethodRefs.TOTP)
}

func TestDefaultStrategy_GetShouldHealWhenTheSessionIsCached(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		authz    func(saved *testContext) string
		expected func(t *testing.T, saved, ctx *testContext)
	}{
		{
			"ShouldIssueTheAuthzCookie",
			testUsername,
			func(saved *testContext) string { return "" },
			func(t *testing.T, saved, ctx *testContext) {
				assert.Equal(t, saved.cookies[testNameAuthz], ctx.cookies[testNameAuthz])
			},
		},
		{
			"ShouldReplaceAMismatchedAuthzCookie",
			testUsername,
			func(saved *testContext) string { return newTestCookie("other") },
			func(t *testing.T, saved, ctx *testContext) {
				assert.Equal(t, saved.cookies[testNameAuthz], ctx.cookies[testNameAuthz])
			},
		},
		{
			"ShouldClearTheAuthzCookieOfAnAnonymousSession",
			"",
			func(saved *testContext) string { return newTestCookie("other") },
			func(t *testing.T, saved, ctx *testContext) {
				assert.NotContains(t, ctx.cookies, testNameAuthz)
				require.Len(t, ctx.cleared, 1)
				assert.Equal(t, testNameAuthz, ctx.cleared[0].Name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := newTestStrategy(t, nil)
			saved := newTestContext()

			userSession := newTestAnonymousSessionWithState(strategy)

			if tc.username != "" {
				userSession = strategy.New(tc.username)
				userSession.SetOneFactorPassword(time.Now(), false)
			}

			require.NoError(t, strategy.Save(saved, &userSession))

			ctx := newTestCachingContext()

			ctx.cookies[testNameSession] = saved.cookies[testNameSession]

			if value := tc.authz(saved); value != "" {
				ctx.cookies[testNameAuthz] = value
			}

			_, _, err := strategy.(*DefaultStrategy).get(ctx, false)
			require.NoError(t, err)
			require.Contains(t, ctx.sessions, testDomain)

			_, err = strategy.Get(ctx)
			require.NoError(t, err)

			tc.expected(t, saved, ctx.testContext)
		})
	}
}

func newTestAuthenticatedContext(t *testing.T, strategy Strategy) (ctx *testContext) {
	t.Helper()

	ctx = newTestContext()

	userSession := strategy.New(testUsername)

	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	return ctx
}

type failingIndexSaveRepository struct {
	*testRepository

	err  error
	fail bool
}

func (r *failingIndexSaveRepository) Save(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	if r.fail {
		return r.err
	}

	return r.testRepository.Save(ctx, issuer, id, pid, username, expiration, data)
}

type failingAuthzIndexSaveRepository struct {
	*testRepository

	issuer string
	fail   bool
}

func (r *failingAuthzIndexSaveRepository) Save(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	if r.fail && issuer == r.issuer {
		return errTestFailure
	}

	return r.testRepository.Save(ctx, issuer, id, pid, username, expiration, data)
}

func TestDefaultStrategy_GetShouldNotResolveTheCookieOfARegeneratedSession(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	stale := newTestContext()
	stale.cookies[testNameSession] = saved.cookies[testNameSession]
	stale.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

	old := newTestAuthzContext(stale)

	request := newTestCachingContext()
	request.cookies[testNameSession] = saved.cookies[testNameSession]

	loaded, err := strategy.Get(request)
	require.NoError(t, err)

	require.NoError(t, strategy.Regenerate(request))

	loaded.SetTwoFactorTOTP(time.Now())

	require.NoError(t, strategy.Save(request, loaded))
	require.NotEqual(t, stale.cookies[testNameSession], request.cookies[testNameSession])

	for range 2 {
		actual, err := strategy.Get(stale)

		require.NoError(t, err)
		assert.True(t, actual.IsAnonymous())
		assert.Empty(t, actual.Username)
	}

	assert.Zero(t, stale.set)
	assert.Empty(t, stale.jar)
	assert.Empty(t, stale.cleared)
	assert.NotEqual(t, request.cookies[testNameSession], stale.cookies[testNameSession])

	resolved, err := strategy.Authz().Get(old)

	require.NoError(t, err)
	assert.True(t, resolved.IsAnonymous())

	after := newTestContext()
	after.cookies[testNameSession] = request.cookies[testNameSession]

	actual, err := strategy.Get(after)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)

	resolved, err = strategy.Authz().Get(newTestAuthzContext(request.testContext))

	require.NoError(t, err)
	assert.Equal(t, testUsername, resolved.Username)
}

func TestDefaultStrategy_RegenerateShouldFailWhenAConcurrentRequestRegeneratedTheSession(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	first := newTestCachingContext()
	first.cookies[testNameSession] = saved.cookies[testNameSession]

	second := newTestCachingContext()
	second.cookies[testNameSession] = saved.cookies[testNameSession]

	loaded, err := strategy.Get(first)
	require.NoError(t, err)

	_, err = strategy.Get(second)
	require.NoError(t, err)

	require.NoError(t, strategy.Regenerate(first))

	loaded.SetTwoFactorTOTP(time.Now())

	require.NoError(t, strategy.Save(first, loaded))

	set := second.set

	delete(second.jar, testNameAuthz)

	assert.ErrorIs(t, strategy.Regenerate(second), ErrSessionSuperseded)

	assert.Equal(t, set, second.set)
	assert.Empty(t, second.jar)
	assert.Equal(t, saved.cookies[testNameSession], second.cookies[testNameSession])
	assert.NotEqual(t, first.cookies[testNameSession], second.cookies[testNameSession])

	after := newTestContext()
	after.cookies[testNameSession] = first.cookies[testNameSession]

	actual, err := strategy.Get(after)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
	assert.Equal(t, loaded.PublicID, actual.PublicID)
	assert.Len(t, liveTestRecords(t, strategy, repository, repository.data), 2)
}

func TestDefaultStrategy_RegenerateShouldFailWhenAConcurrentRequestDestroyedTheSession(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	first := newTestCachingContext()
	first.cookies[testNameSession] = saved.cookies[testNameSession]

	second := newTestCachingContext()
	second.cookies[testNameSession] = saved.cookies[testNameSession]

	_, err := strategy.Get(first)
	require.NoError(t, err)

	_, err = strategy.Get(second)
	require.NoError(t, err)

	require.NoError(t, strategy.Regenerate(first))
	require.NoError(t, strategy.Destroy(first))

	assert.ErrorIs(t, strategy.Regenerate(second), ErrSessionSuperseded)
	assert.Equal(t, saved.cookies[testNameSession], second.cookies[testNameSession])
}

func TestDefaultStrategy_GetShouldClearTheAuthzCookieOfADestroyedSession(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	stale := newTestContext()
	stale.cookies[testNameSession] = saved.cookies[testNameSession]
	stale.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

	require.NoError(t, strategy.Destroy(saved))

	actual, err := strategy.Get(stale)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
	require.Len(t, stale.cleared, 1)
	assert.Equal(t, testNameAuthz, stale.cleared[0].Name)
}

func TestDefaultStrategy_GetShouldNotClearTheAuthzCookieOfAnotherSession(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)
	other := newTestSavedContext(t, strategy, "harry")

	stale := newTestContext()
	stale.cookies[testNameSession] = saved.cookies[testNameSession]
	stale.cookies[testNameAuthz] = other.cookies[testNameAuthz]

	require.NoError(t, strategy.Destroy(saved))

	actual, err := strategy.Get(stale)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
	assert.Empty(t, stale.cleared)
}

func TestDefaultStrategy_GetShouldClearAMalformedAuthzCookie(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	stale := newTestContext()
	stale.cookies[testNameSession] = saved.cookies[testNameSession]
	stale.cookies[testNameAuthz] = "not-a-cookie-value"

	require.NoError(t, strategy.Destroy(saved))

	_, err := strategy.Get(stale)

	require.NoError(t, err)
	require.Len(t, stale.cleared, 1)
	assert.Equal(t, testNameAuthz, stale.cleared[0].Name)
}
