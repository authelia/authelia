// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestDefaultStrategy_SaveShouldAnchorRemoteIP(t *testing.T) {
	testCases := []struct {
		name     string
		anchor   bool
		expected net.IP
	}{
		{
			"ShouldRecordRemoteIPWhenEnabled",
			true,
			testRemoteIP,
		},
		{
			"ShouldNotRecordRemoteIPWhenDisabled",
			false,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := newTestStrategy(t, func(config *schema.SessionCookie) {
				config.AnchorRemoteIP = tc.anchor
			})

			ctx := newTestContext()

			userSession := strategy.NewDefault()
			userSession.Username = testUsername

			require.NoError(t, strategy.Save(ctx, &userSession))

			assert.Equal(t, tc.expected, userSession.RemoteIP)

			actual, err := strategy.Get(ctx)

			require.NoError(t, err)
			assert.Equal(t, testUsername, actual.Username)
			assert.Equal(t, tc.expected, actual.RemoteIP)
		})
	}
}

func TestDefaultStrategy_SaveShouldNotReplaceAnchoredRemoteIP(t *testing.T) {
	strategy := newTestStrategy(t, func(config *schema.SessionCookie) {
		config.AnchorRemoteIP = true
	})

	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.RemoteIP = testOtherRemoteIP

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Equal(t, testOtherRemoteIP, userSession.RemoteIP)
}

func TestDefaultStrategy_GetShouldReturnAnchoredSessionForTheSameRemoteIP(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
		config.AnchorRemoteIP = true
	})

	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	// A later request from the same IP, which has no session retained against its context.
	next := newTestContext()
	next.cookies[testName] = ctx.cookies[testName]

	actual, err := strategy.Get(next)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
	assert.Len(t, repository.data, 1)
	assert.Nil(t, next.cleared)
}

func TestDefaultStrategy_GetShouldDestroyAnchoredSessionForAnotherRemoteIP(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
		config.AnchorRemoteIP = true
	})

	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))
	require.Len(t, repository.data, 1)

	original := ctx.cookies[testName]

	next := newTestContext()
	next.remoteIP = testOtherRemoteIP
	next.cookies[testName] = original

	actual, err := strategy.Get(next)

	require.NoError(t, err)
	require.NotNil(t, actual)

	assert.True(t, actual.IsAnonymous())
	assert.Nil(t, actual.RemoteIP)
	assert.Equal(t, testDomain, actual.CookieDomain)
	assert.Empty(t, repository.data)
	assert.Empty(t, repository.publicIDs)
	assert.Empty(t, repository.usernames)
	assert.NotContains(t, next.cookies, testName)
	require.NotNil(t, next.cleared)
	assert.Equal(t, testName, next.cleared.Name)

	// The anonymous session which replaces it is issued a new identifier anchored to the new remote IP.
	require.NoError(t, strategy.Save(next, actual))

	assert.NotEmpty(t, next.cookies[testName])
	assert.NotEqual(t, original, next.cookies[testName])
	assert.Equal(t, testOtherRemoteIP, actual.RemoteIP)
}

func TestDefaultStrategy_GetShouldDestroyAnchoredSessionWithoutRemoteIP(t *testing.T) {
	repository := newTestRepository()

	unanchored := newTestStrategyWithRepository(t, repository, nil)
	anchored := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
		config.AnchorRemoteIP = true
	})

	ctx := newTestContext()

	userSession := unanchored.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, unanchored.Save(ctx, &userSession))
	require.Nil(t, userSession.RemoteIP)
	require.Len(t, repository.data, 1)

	next := newTestContext()
	next.cookies[testName] = ctx.cookies[testName]

	actual, err := anchored.Get(next)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
	assert.Empty(t, repository.data)
	assert.NotContains(t, next.cookies, testName)
}

func TestDefaultStrategy_GetShouldNotDestroyUnanchoredSessionForAnotherRemoteIP(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)

	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	next := newTestContext()
	next.remoteIP = testOtherRemoteIP
	next.cookies[testName] = ctx.cookies[testName]

	actual, err := strategy.Get(next)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
	assert.Len(t, repository.data, 1)
}

func TestDefaultStrategy_RegenerateShouldNotPreserveAnchoredSessionForAnotherRemoteIP(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
		config.AnchorRemoteIP = true
	})

	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	next := newTestContext()
	next.remoteIP = testOtherRemoteIP
	next.cookies[testName] = ctx.cookies[testName]

	require.NoError(t, strategy.Regenerate(next))

	assert.Empty(t, repository.data)

	actual, err := strategy.Get(next)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
}

var (
	testRemoteIP      = net.ParseIP("192.0.2.1")
	testOtherRemoteIP = net.ParseIP("198.51.100.1")
)
