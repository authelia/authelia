// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStrategy_ShouldNotHonorTheSessionCookieOnAnAuthorizationRequest(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestAuthenticatedContext(t, strategy)

	ctx := newTestContext()

	ctx.authz = true
	ctx.cookies[testName] = saved.cookies[testName]

	records := len(repository.data)

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
	assert.Equal(t, "", actual.Username)
	assert.Equal(t, 0, ctx.set)
	assert.Len(t, repository.data, records)

	userSession := strategy.New(testUsername)

	userSession.SetOneFactorPassword(time.Now(), false)

	assert.ErrorIs(t, strategy.Save(ctx, &userSession), ErrNotPermittedOnAuthzEndpointRequest)
	assert.ErrorIs(t, strategy.Regenerate(ctx), ErrNotPermittedOnAuthzEndpointRequest)
	assert.ErrorIs(t, strategy.Destroy(ctx), ErrNotPermittedOnAuthzEndpointRequest)

	assert.Equal(t, 0, ctx.set)
	assert.Empty(t, ctx.cleared)
	assert.Len(t, repository.data, records)

	actual, err = strategy.Get(saved)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
}

func TestDefaultAuthzStrategy_ShouldNotHonorTheAuthzCookieOnAPortalRequest(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestAuthenticatedContext(t, strategy)

	ctx := newTestContext()

	ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

	records := len(repository.data)

	view := strategy.Authz()

	stored, err := view.Get(newTestAuthzContext(saved))

	require.NoError(t, err)
	require.Equal(t, testUsername, stored.Username)
	require.NotEqual(t, int64(100), stored.LastActivity)

	expected := stored.LastActivity

	actual, err := view.Get(ctx)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
	assert.Equal(t, "", actual.Username)

	actual.LastActivity = 100

	require.NoError(t, view.SaveActivity(ctx, actual))
	assert.ErrorIs(t, view.Destroy(ctx), ErrNotPermittedOnPortalRequest)

	assert.Equal(t, 0, ctx.set)
	assert.Empty(t, ctx.cleared)
	assert.Len(t, repository.data, records)

	authz := newTestAuthzContext(saved)

	actual, err = view.Get(authz)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
	assert.Equal(t, expected, actual.LastActivity)
}

func TestDefaultStrategy_ShouldNotShareAnAnonymousResultAcrossRequestKinds(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	saved := newTestAuthenticatedContext(t, strategy)

	ctx := newTestCachingContext()

	ctx.authz = true
	ctx.cookies[testName] = saved.cookies[testName]
	ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
	assert.NotContains(t, ctx.sessions, testDomain)

	actual, err = strategy.Authz().Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
}

func TestDefaultAuthzStrategy_ShouldNotShareAnAnonymousResultAcrossRequestKinds(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	saved := newTestAuthenticatedContext(t, strategy)

	ctx := newTestCachingContext()

	ctx.cookies[testName] = saved.cookies[testName]
	ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

	actual, err := strategy.Authz().Get(ctx)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
	assert.NotContains(t, ctx.sessions, cacheKeyPrefixAuthz+testDomain)

	actual, err = strategy.Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
}
