// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

func TestNewProvider(t *testing.T) {
	config := &schema.Configuration{
		Session: schema.Session{
			Secret: testSecret,
			Cookies: []schema.SessionCookie{
				{
					SessionCookieCommon: schema.SessionCookieCommon{Name: testName, SameSite: "lax", Expiration: testExpiration},
					Domain:              testDomain,
				},
				{
					SessionCookieCommon: schema.SessionCookieCommon{Name: "other_session", SameSite: "strict", Expiration: testExpiration},
					Domain:              "example.org",
				},
			},
		},
	}

	provider, err := NewProvider(config, []byte(testHMACKey), []byte(testHMACKey+"-csrf"), clock.New(), random.NewMathematical(), newTestRepository())

	require.NoError(t, err)
	require.NotNil(t, provider)

	testCases := []struct {
		name     string
		domain   string
		expected string
		err      string
	}{
		{"ShouldReturnStrategyForFirstDomain", testDomain, testName, ""},
		{"ShouldReturnStrategyForSecondDomain", "example.org", "other_session", ""},
		{"ShouldReturnErrorForUnknownDomain", "example.net", "", "not found"},
		{"ShouldReturnErrorForSubdomainOfConfiguredDomain", "sub.example.com", "", "not found"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy, err := provider.GetStrategy(tc.domain)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, strategy)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, strategy)

			assert.Equal(t, tc.domain, strategy.GetConfig().Domain)
			assert.Equal(t, tc.expected, strategy.GetConfig().Name)
		})
	}
}

func TestDefaultProvider_StartupCheck(t *testing.T) {
	provider, err := NewProvider(&schema.Configuration{Session: schema.Session{Secret: testSecret}}, []byte(testHMACKey), []byte(testHMACKey+"-csrf"), clock.New(), random.NewMathematical(), newTestRepository())

	require.NoError(t, err)

	defaultProvider, ok := provider.(*DefaultProvider)
	require.True(t, ok)

	assert.NoError(t, defaultProvider.StartupCheck())

	_, err = defaultProvider.GetStrategy(testDomain)
	assert.EqualError(t, err, "not found")
}
