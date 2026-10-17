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

			ctx.cookies[testName] = saved.cookies[testName]

			if tc.have != "" {
				ctx.cookies[testNameAuthz] = tc.have
			}

			actual, err := strategy.Get(ctx)

			require.NoError(t, err)
			assert.Equal(t, testUsername, actual.Username)
			assert.False(t, actual.IsAnonymous())

			assert.Equal(t, 1, ctx.set)
			assert.NotContains(t, ctx.jar, testName)
			require.NotNil(t, ctx.jar[testNameAuthz])

			healed := ctx.jar[testNameAuthz]

			assert.Equal(t, saved.cookies[testNameAuthz], healed.Value)
			assert.NotEqual(t, tc.have, healed.Value)
			assert.Equal(t, http.SameSiteLaxMode, healed.SameSite)
			assert.Equal(t, "."+testDomain, healed.Domain)
			assert.False(t, healed.Expires.IsZero())
			assert.Equal(t, saved.cookies[testName], ctx.cookies[testName])
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

	ctx.cookies[testName] = saved.cookies[testName]

	anonymous, err := strategy.Authz().Get(ctx)

	require.NoError(t, err)
	require.True(t, anonymous.IsAnonymous())

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	require.Contains(t, ctx.sessions, testDomain)
	assert.Equal(t, testUsername, actual.Username)

	// The authz cookie is derived from the session so it's issued with the same value every time it's healed.
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

			ctx.cookies[testName] = saved.cookies[testName]

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

			ctx.cookies[testName] = saved.cookies[testName]

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

func newTestAuthenticatedContext(t *testing.T, strategy Strategy) (ctx *testContext) {
	t.Helper()

	ctx = newTestContext()

	userSession := strategy.New(testUsername)

	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	return ctx
}

func TestDefaultStrategy_GetShouldHealConcurrentRequestsWithTheSameAuthzCookie(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestAuthenticatedContext(t, strategy)

	primary := testPrimaryKey(t, strategy, saved)
	stored := append([]byte(nil), repository.data[primary]...)

	first, second := newTestContext(), newTestContext()

	first.cookies[testName] = saved.cookies[testName]
	second.cookies[testName] = saved.cookies[testName]

	_, err := strategy.Get(first)
	require.NoError(t, err)

	_, err = strategy.Get(second)
	require.NoError(t, err)

	require.NotEmpty(t, first.cookies[testNameAuthz])
	assert.Equal(t, first.cookies[testNameAuthz], second.cookies[testNameAuthz])
	assert.Equal(t, saved.cookies[testNameAuthz], first.cookies[testNameAuthz])

	// Healing only ever writes the authz index, never the session record.
	assert.Equal(t, stored, repository.data[primary])
	assert.Len(t, repository.data, 2)

	for _, ctx := range []*testContext{first, second, saved} {
		resolved, err := strategy.Authz().Get(newTestAuthzContext(ctx))

		require.NoError(t, err)
		assert.Equal(t, testUsername, resolved.Username)
	}
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
