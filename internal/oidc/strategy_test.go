// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewAudienceStrategy(t *testing.T) {
	testCases := []struct {
		Name     string
		Strategy string
		Haystack []string
		Needle   []string
		Nil      bool
		Err      string
	}{
		{Name: "ShouldReturnNilEmpty", Strategy: "", Nil: true},
		{Name: "ShouldReturnNilInvalid", Strategy: "prefix", Nil: true},
		{Name: "ShouldMatchExact", Strategy: oidc.StrategyExact, Haystack: []string{"https://app.example.com/api"}, Needle: []string{"https://app.example.com/api"}},
		{Name: "ShouldNotMatchExactSubpath", Strategy: oidc.StrategyExact, Haystack: []string{"https://app.example.com/api"}, Needle: []string{"https://app.example.com/api/users"}, Err: "invalid_target"},
		{Name: "ShouldNotMatchExactWildcard", Strategy: oidc.StrategyExact, Haystack: []string{"https://app.example.com/api/*"}, Needle: []string{"https://app.example.com/api/users"}, Err: "invalid_target"},
		{Name: "ShouldMatchWildcard", Strategy: oidc.StrategyWildcard, Haystack: []string{"https://app.example.com/api/*"}, Needle: []string{"https://app.example.com/api/users"}},
		{Name: "ShouldMatchWildcardOpaque", Strategy: oidc.StrategyWildcard, Haystack: []string{"opaque"}, Needle: []string{"opaque"}},
		{Name: "ShouldNotMatchWildcardDotSegment", Strategy: oidc.StrategyWildcard, Haystack: []string{"https://app.example.com/api/*"}, Needle: []string{"https://app.example.com/api/../admin"}, Err: "invalid_target"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			strategy := oidc.NewAudienceStrategy(tc.Strategy)

			if tc.Nil {
				assert.Nil(t, strategy)

				return
			}

			require.NotNil(t, strategy)

			err := strategy(tc.Haystack, tc.Needle)

			if tc.Err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.Err)
			}
		})
	}
}

func TestNewResourceStrategy(t *testing.T) {
	testCases := []struct {
		Name     string
		Strategy string
		Haystack []string
		Needle   []string
		Nil      bool
		Err      string
	}{
		{Name: "ShouldReturnNilEmpty", Strategy: "", Nil: true},
		{Name: "ShouldReturnNilInvalid", Strategy: "magic", Nil: true},
		{Name: "ShouldMatchExact", Strategy: oidc.StrategyExact, Haystack: []string{"https://app.example.com/api"}, Needle: []string{"https://app.example.com/api"}},
		{Name: "ShouldNotMatchExactSubpath", Strategy: oidc.StrategyExact, Haystack: []string{"https://app.example.com/api"}, Needle: []string{"https://app.example.com/api/users"}, Err: "invalid_target"},
		{Name: "ShouldMatchPrefixSubpath", Strategy: oidc.StrategyPrefix, Haystack: []string{"https://app.example.com/api"}, Needle: []string{"https://app.example.com/api/users"}},
		{Name: "ShouldNotMatchPrefixSegmentBoundary", Strategy: oidc.StrategyPrefix, Haystack: []string{"https://app.example.com/api"}, Needle: []string{"https://app.example.com/api2"}, Err: "invalid_target"},
		{Name: "ShouldMatchWildcard", Strategy: oidc.StrategyWildcard, Haystack: []string{"https://app.example.com/api/*"}, Needle: []string{"https://app.example.com/api/users"}},
		{Name: "ShouldNotMatchWildcardWithoutSuffix", Strategy: oidc.StrategyWildcard, Haystack: []string{"https://app.example.com/api"}, Needle: []string{"https://app.example.com/api/users"}, Err: "invalid_target"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			strategy := oidc.NewResourceStrategy(tc.Strategy)

			if tc.Nil {
				assert.Nil(t, strategy)

				return
			}

			require.NotNil(t, strategy)

			err := strategy(tc.Haystack, tc.Needle)

			if tc.Err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.Err)
			}
		})
	}
}

func TestStrategyResolution(t *testing.T) {
	testCases := []struct {
		Name             string
		Provider         schema.IdentityProvidersOpenIDConnect
		Client           schema.IdentityProvidersOpenIDConnectClient
		Haystack         []string
		Needle           []string
		ExpectedAudience bool
		ExpectedResource bool
	}{
		{
			Name:             "ShouldUseProviderStrategies",
			Provider:         schema.IdentityProvidersOpenIDConnect{AudienceStrategy: oidc.StrategyExact, ResourceStrategy: oidc.StrategyPrefix},
			Client:           schema.IdentityProvidersOpenIDConnectClient{ID: "test"},
			Haystack:         []string{"https://app.example.com/api"},
			Needle:           []string{"https://app.example.com/api/users"},
			ExpectedAudience: false,
			ExpectedResource: true,
		},
		{
			Name:             "ShouldPreferClientStrategies",
			Provider:         schema.IdentityProvidersOpenIDConnect{AudienceStrategy: oidc.StrategyExact, ResourceStrategy: oidc.StrategyPrefix},
			Client:           schema.IdentityProvidersOpenIDConnectClient{ID: "test", AudienceStrategy: oidc.StrategyWildcard, ResourceStrategy: oidc.StrategyExact},
			Haystack:         []string{"https://app.example.com/api/*"},
			Needle:           []string{"https://app.example.com/api/users"},
			ExpectedAudience: true,
			ExpectedResource: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()

			config := oidc.NewConfig(&tc.Provider, nil, nil)
			client := oidc.NewClient(tc.Client, &tc.Provider, nil)

			assert.Equal(t, tc.ExpectedAudience, oauthelia2.GetAudienceStrategy(ctx, config, client)(tc.Haystack, tc.Needle) == nil)
			assert.Equal(t, tc.ExpectedResource, oauthelia2.GetResourceStrategy(ctx, config, client)(tc.Haystack, tc.Needle) == nil)
		})
	}
}

func TestNewConfigClientAssertionIssuerAudience(t *testing.T) {
	testCases := []struct {
		Name     string
		Legacy   bool
		Expected bool
	}{
		{"ShouldEnforceByDefault", false, false},
		{"ShouldAllowLegacy", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{EnableClientAssertionLegacyAudience: tc.Legacy}, nil, nil)

			assert.Equal(t, !tc.Legacy, config.EnforceClientAssertionIssuerAudience)
			assert.Equal(t, tc.Expected, config.GetDisableClientAssertionIssuerAudience(context.Background()))
		})
	}
}
