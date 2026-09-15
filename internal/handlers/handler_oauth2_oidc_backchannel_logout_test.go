// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOIDCBackChannelLogoutClientsBySector(t *testing.T) {
	testCases := []struct {
		name     string
		clients  []oauthelia2.Client
		expected map[string][]string
	}{
		{
			"ShouldHandleNoClients",
			nil,
			map[string][]string{},
		},
		{
			"ShouldGroupClientsWithNoSectorIdentifierTogether",
			[]oauthelia2.Client{
				&oidc.RegisteredClient{ID: "a"},
				&oidc.RegisteredClient{ID: "b"},
			},
			map[string][]string{"": {"a", "b"}},
		},
		{
			"ShouldSeparateClientsByDistinctSectorIdentifier",
			[]oauthelia2.Client{
				&oidc.RegisteredClient{ID: "a", SectorIdentifierURI: &url.URL{Scheme: "https", Host: "one.example.com"}},
				&oidc.RegisteredClient{ID: "b", SectorIdentifierURI: &url.URL{Scheme: "https", Host: "two.example.com"}},
				&oidc.RegisteredClient{ID: "c", SectorIdentifierURI: &url.URL{Scheme: "https", Host: "one.example.com"}},
				&oidc.RegisteredClient{ID: "d"},
			},
			map[string][]string{
				"https://one.example.com": {"a", "c"},
				"https://two.example.com": {"b"},
				"":                        {"d"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := oidcBackChannelLogoutClientsBySector(tc.clients)

			require.Len(t, actual, len(tc.expected))

			for sector, expectedIDs := range tc.expected {
				sectorClients, ok := actual[sector]

				require.True(t, ok, "expected a group for sector identifier '%s'", sector)

				ids := make([]string, len(sectorClients))

				for i, client := range sectorClients {
					ids[i] = client.GetID()
				}

				assert.Equal(t, expectedIDs, ids)
			}
		})
	}
}

func TestOIDCBackChannelLogoutClients(t *testing.T) {
	// TODO: This records the current behavior of the participation lookup, which returns no clients until the
	// session rewrite provides the session identifier it needs. Replace the assertion when it does.
	assert.Empty(t, oidcBackChannelLogoutClients(nil, "john", ""))
}
