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

func TestDefaultStrategy_SaveShouldAnchorRemoteNetwork(t *testing.T) {
	testCases := []struct {
		name            string
		anchor          *schema.SessionCookieAnchorRemoteIP
		remoteIP        net.IP
		expectedNetwork net.IP
		expectedBits    uint8
	}{
		{
			"ShouldRecordIPv4Host",
			&schema.SessionCookieAnchorRemoteIP{IPv4Mask: 32, IPv6Mask: 128},
			net.ParseIP("192.0.2.1"),
			net.IP{192, 0, 2, 1},
			32,
		},
		{
			"ShouldRecordIPv4Network",
			&schema.SessionCookieAnchorRemoteIP{IPv4Mask: 24, IPv6Mask: 64},
			net.ParseIP("192.0.2.130"),
			net.IP{192, 0, 2, 0},
			24,
		},
		{
			"ShouldRecordIPv4MappedIPv6AsIPv4",
			&schema.SessionCookieAnchorRemoteIP{IPv4Mask: 24, IPv6Mask: 64},
			net.ParseIP("::ffff:192.0.2.130"),
			net.IP{192, 0, 2, 0},
			24,
		},
		{
			"ShouldRecordIPv6Host",
			&schema.SessionCookieAnchorRemoteIP{IPv4Mask: 32, IPv6Mask: 128},
			net.ParseIP("2001:db8:1:2:3:4:5:6"),
			net.ParseIP("2001:db8:1:2:3:4:5:6"),
			128,
		},
		{
			"ShouldRecordIPv6Network",
			&schema.SessionCookieAnchorRemoteIP{IPv4Mask: 32, IPv6Mask: 64},
			net.ParseIP("2001:db8:1:2:3:4:5:6"),
			net.ParseIP("2001:db8:1:2::"),
			64,
		},
		{
			"ShouldNotRecordWhenDisabled",
			nil,
			net.ParseIP("192.0.2.1"),
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := newTestStrategy(t, func(config *schema.SessionCookie) {
				config.AnchorRemoteIP = tc.anchor
			})

			ctx := newTestContext()
			ctx.remoteIP = tc.remoteIP

			userSession := strategy.NewDefault()
			userSession.Username = testUsername

			require.NoError(t, strategy.Save(ctx, &userSession))

			assert.Equal(t, tc.expectedNetwork, userSession.RemoteNetwork)
			assert.Equal(t, tc.expectedBits, userSession.RemoteNetworkBits)

			actual, err := strategy.Get(ctx)

			require.NoError(t, err)
			assert.Equal(t, testUsername, actual.Username)
			assert.Equal(t, tc.expectedNetwork, actual.RemoteNetwork)
			assert.Equal(t, tc.expectedBits, actual.RemoteNetworkBits)
		})
	}
}

func TestDefaultStrategy_SaveShouldNotReplaceAnchoredRemoteNetwork(t *testing.T) {
	strategy := newTestStrategy(t, func(config *schema.SessionCookie) {
		config.AnchorRemoteIP = testAnchorHost
	})

	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.RemoteNetwork = net.IP{198, 51, 100, 0}
	userSession.RemoteNetworkBits = 24

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Equal(t, net.IP{198, 51, 100, 0}, userSession.RemoteNetwork)
	assert.Equal(t, uint8(24), userSession.RemoteNetworkBits)
}

func TestDefaultStrategy_GetShouldEnforceAnchoredRemoteNetwork(t *testing.T) {
	testCases := []struct {
		name      string
		anchor    *schema.SessionCookieAnchorRemoteIP
		first     net.IP
		next      net.IP
		destroyed bool
	}{
		{
			"ShouldReturnForSameIPv4Host",
			testAnchorHost,
			net.ParseIP("192.0.2.1"),
			net.ParseIP("192.0.2.1"),
			false,
		},
		{
			"ShouldDestroyForOtherIPv4Host",
			testAnchorHost,
			net.ParseIP("192.0.2.1"),
			net.ParseIP("192.0.2.2"),
			true,
		},
		{
			"ShouldReturnForSameIPv4Network",
			testAnchorNetwork,
			net.ParseIP("192.0.2.1"),
			net.ParseIP("192.0.2.254"),
			false,
		},
		{
			"ShouldReturnForSameIPv4NetworkWhenIPv4MappedIPv6",
			testAnchorNetwork,
			net.ParseIP("192.0.2.1"),
			net.ParseIP("::ffff:192.0.2.254"),
			false,
		},
		{
			"ShouldDestroyForOtherIPv4Network",
			testAnchorNetwork,
			net.ParseIP("192.0.2.1"),
			net.ParseIP("192.0.3.1"),
			true,
		},
		{
			"ShouldReturnForSameIPv6Host",
			testAnchorHost,
			net.ParseIP("2001:db8:1:2:3:4:5:6"),
			net.ParseIP("2001:db8:1:2:3:4:5:6"),
			false,
		},
		{
			"ShouldDestroyForOtherIPv6Host",
			testAnchorHost,
			net.ParseIP("2001:db8:1:2:3:4:5:6"),
			net.ParseIP("2001:db8:1:2:3:4:5:7"),
			true,
		},
		{
			"ShouldReturnForSameIPv6Network",
			testAnchorNetwork,
			net.ParseIP("2001:db8:1:2:3:4:5:6"),
			net.ParseIP("2001:db8:1:2:ffff:ffff:ffff:ffff"),
			false,
		},
		{
			"ShouldDestroyForOtherIPv6Network",
			testAnchorNetwork,
			net.ParseIP("2001:db8:1:2:3:4:5:6"),
			net.ParseIP("2001:db8:1:3:3:4:5:6"),
			true,
		},
		{
			"ShouldDestroyForIPv6WhenAnchoredToIPv4",
			testAnchorNetwork,
			net.ParseIP("192.0.2.1"),
			net.ParseIP("2001:db8:1:2:3:4:5:6"),
			true,
		},
		{
			"ShouldDestroyForIPv4WhenAnchoredToIPv6",
			testAnchorNetwork,
			net.ParseIP("2001:db8:1:2:3:4:5:6"),
			net.ParseIP("192.0.2.1"),
			true,
		},
		{
			"ShouldDestroyWithoutRemoteIP",
			testAnchorNetwork,
			net.ParseIP("192.0.2.1"),
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()
			strategy := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
				config.AnchorRemoteIP = tc.anchor
			})

			ctx := newTestContext()
			ctx.remoteIP = tc.first

			userSession := strategy.NewDefault()
			userSession.Username = testUsername

			require.NoError(t, strategy.Save(ctx, &userSession))
			require.Len(t, repository.data, 1)

			// A later request, which has no session retained against its context.
			next := newTestContext()
			next.remoteIP = tc.next
			next.cookies[testName] = ctx.cookies[testName]

			actual, err := strategy.Get(next)

			require.NoError(t, err)
			require.NotNil(t, actual)

			if !tc.destroyed {
				assert.Equal(t, testUsername, actual.Username)
				assert.Len(t, repository.data, 1)
				assert.Nil(t, next.cleared)

				return
			}

			assert.True(t, actual.IsAnonymous())
			assert.Nil(t, actual.RemoteNetwork)
			assert.Zero(t, actual.RemoteNetworkBits)
			assert.Equal(t, testDomain, actual.CookieDomain)
			assert.Empty(t, repository.data)
			assert.Empty(t, repository.publicIDs)
			assert.Empty(t, repository.usernames)
			assert.NotContains(t, next.cookies, testName)
			require.NotNil(t, next.cleared)
			assert.Equal(t, testName, next.cleared.Name)
		})
	}
}

func TestDefaultStrategy_GetShouldAnchorReplacementSessionToTheNewRemoteNetwork(t *testing.T) {
	strategy := newTestStrategy(t, func(config *schema.SessionCookie) {
		config.AnchorRemoteIP = testAnchorNetwork
	})

	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	original := ctx.cookies[testName]

	next := newTestContext()
	next.remoteIP = testOtherRemoteIP
	next.cookies[testName] = original

	actual, err := strategy.Get(next)

	require.NoError(t, err)
	require.True(t, actual.IsAnonymous())

	// The anonymous session which replaces it is issued a new identifier anchored to the new remote network.
	require.NoError(t, strategy.Save(next, actual))

	assert.NotEmpty(t, next.cookies[testName])
	assert.NotEqual(t, original, next.cookies[testName])
	assert.Equal(t, net.IP{198, 51, 100, 0}, actual.RemoteNetwork)
	assert.Equal(t, uint8(24), actual.RemoteNetworkBits)
}

func TestDefaultStrategy_GetShouldUseTheMaskRecordedOnTheSession(t *testing.T) {
	testCases := []struct {
		name      string
		saved     *schema.SessionCookieAnchorRemoteIP
		current   *schema.SessionCookieAnchorRemoteIP
		destroyed bool
	}{
		{
			"ShouldReturnForSameNetworkWhenMaskNarrowed",
			testAnchorNetwork,
			testAnchorHost,
			false,
		},
		{
			"ShouldDestroyForOtherHostWhenMaskWidened",
			testAnchorHost,
			testAnchorNetwork,
			true,
		},
		{
			"ShouldReturnForOtherHostWhenAnchoringRemoved",
			testAnchorHost,
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()

			saved := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
				config.AnchorRemoteIP = tc.saved
			})

			current := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
				config.AnchorRemoteIP = tc.current
			})

			ctx := newTestContext()

			userSession := saved.NewDefault()
			userSession.Username = testUsername

			require.NoError(t, saved.Save(ctx, &userSession))

			next := newTestContext()
			next.remoteIP = net.ParseIP("192.0.2.2")
			next.cookies[testName] = ctx.cookies[testName]

			actual, err := current.Get(next)

			require.NoError(t, err)

			if tc.destroyed {
				assert.Empty(t, actual.Username)
				assert.Empty(t, repository.data)
			} else {
				assert.Equal(t, testUsername, actual.Username)
				assert.Len(t, repository.data, 1)
			}
		})
	}
}

func TestDefaultStrategy_GetShouldDestroySessionWithoutAnchorWhenAnchoring(t *testing.T) {
	testCases := []struct {
		name   string
		modify func(userSession *UserSession)
	}{
		{
			"ShouldDestroySessionWithoutRemoteNetwork",
			nil,
		},
		{
			"ShouldDestroySessionWithoutRemoteNetworkBits",
			func(userSession *UserSession) {
				userSession.RemoteNetwork = net.IP{192, 0, 2, 1}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()

			unanchored := newTestStrategyWithRepository(t, repository, nil)
			anchored := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
				config.AnchorRemoteIP = testAnchorHost
			})

			ctx := newTestContext()

			userSession := unanchored.NewDefault()
			userSession.Username = testUsername

			if tc.modify != nil {
				tc.modify(&userSession)
			}

			require.NoError(t, unanchored.Save(ctx, &userSession))
			require.Zero(t, userSession.RemoteNetworkBits)
			require.Len(t, repository.data, 1)

			next := newTestContext()
			next.cookies[testName] = ctx.cookies[testName]

			actual, err := anchored.Get(next)

			require.NoError(t, err)
			assert.True(t, actual.IsAnonymous())
			assert.Empty(t, repository.data)
			assert.NotContains(t, next.cookies, testName)
		})
	}
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

func TestDefaultStrategy_RegenerateShouldNotPreserveAnchoredSessionForAnotherRemoteNetwork(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
		config.AnchorRemoteIP = testAnchorNetwork
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

	testAnchorHost    = &schema.SessionCookieAnchorRemoteIP{IPv4Mask: 32, IPv6Mask: 128}
	testAnchorNetwork = &schema.SessionCookieAnchorRemoteIP{IPv4Mask: 24, IPv6Mask: 64}
)
