// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/storage"
)

// TestRequestedAudienceResourceImplicit ensures the requested audience and resource modes of a client are applied by
// the provider when a request includes neither the audience nor the resource parameter, and only then.
func TestRequestedAudienceResourceImplicit(t *testing.T) {
	const (
		clientID = "implicit-target-client"
		audience = "https://aud.example.com"
		resource = "https://api.example.com"
	)

	newProvider := func(t *testing.T, audienceMode, resourceMode string) *oidc.OpenIDConnectProvider {
		t.Helper()

		store, err := storage.NewSQLiteProvider(&schema.Configuration{
			Storage: schema.Storage{
				EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
				Local:         &schema.StorageLocal{Path: filepath.Join(t.TempDir(), "db.sqlite3")},
			},
		})

		require.NoError(t, err)
		require.NoError(t, store.StartupCheck())

		provider := oidc.NewOpenIDConnectProvider(&schema.Configuration{
			IdentityProviders: schema.IdentityProviders{
				OIDC: &schema.IdentityProvidersOpenIDConnect{
					IssuerCertificateChain: schema.X509CertificateChain{},
					IssuerPrivateKey:       x509PrivateKeyRSA2048,
					HMACSecret:             badhmac,
					Clients: []schema.IdentityProvidersOpenIDConnectClient{
						{
							ID:                      clientID,
							Secret:                  tOpenIDConnectPlainTextClientSecret,
							AuthorizationPolicy:     onefactor,
							TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
							GrantTypes:              []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeClientCredentials},
							ResponseTypes:           []string{oidc.ResponseTypeAuthorizationCodeFlow},
							Scopes:                  []string{oidc.ScopeOpenID, "read"},
							RedirectURIs:            []string{examplecom},
							Audience:                []string{audience},
							Resource:                []string{resource},
							RequestedAudienceMode:   audienceMode,
							RequestedResourceMode:   resourceMode,
						},
					},
				},
			},
		}, store, nil)

		require.NotNil(t, provider)

		return provider
	}

	clientCredentials := func(provider *oidc.OpenIDConnectProvider, extra url.Values) (oauthelia2.AccessRequester, error) {
		form := url.Values{
			"grant_type":               []string{oidc.GrantTypeClientCredentials},
			"scope":                    []string{"read"},
			oidc.FormParameterClientID: []string{clientID},
			"client_secret":            []string{"client-secret"},
		}

		maps.Copy(form, extra)

		r := httptest.NewRequest(http.MethodPost, "https://auth.example.com/api/oidc/token", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		return provider.NewAccessRequest(context.Background(), r, oidc.NewSession())
	}

	authorization := func(provider *oidc.OpenIDConnectProvider, extra url.Values) (oauthelia2.AuthorizeRequester, error) {
		query := url.Values{
			"response_type":            []string{oidc.ResponseTypeAuthorizationCodeFlow},
			"scope":                    []string{oidc.ScopeOpenID},
			"redirect_uri":             []string{examplecom},
			"state":                    []string{"abcdefghijklmnopqrstuvwxyz"},
			oidc.FormParameterClientID: []string{clientID},
		}

		maps.Copy(query, extra)

		r := httptest.NewRequest(http.MethodGet, "https://auth.example.com/api/oidc/authorization?"+query.Encode(), nil)

		return provider.NewAuthorizeRequest(context.Background(), r)
	}

	testCases := []struct {
		name             string
		audienceMode     string
		resourceMode     string
		extra            url.Values
		expectedAudience []string
		expectedResource []string
	}{
		{
			name:             "ShouldRequestBothImplicitly",
			audienceMode:     "implicit",
			resourceMode:     "implicit",
			expectedAudience: []string{audience},
			expectedResource: []string{resource},
		},
		{
			name:             "ShouldRequestOnlyTheImplicitResource",
			audienceMode:     "explicit",
			resourceMode:     "implicit",
			expectedResource: []string{resource},
		},
		{
			name:             "ShouldRequestOnlyTheImplicitAudience",
			audienceMode:     "implicit",
			resourceMode:     "explicit",
			expectedAudience: []string{audience},
		},
		{
			name:         "ShouldRequestNothingExplicitly",
			audienceMode: "explicit",
			resourceMode: "explicit",
		},
		{
			name:             "ShouldNotApplyImplicitModesWhenTheResourceIsRequested",
			audienceMode:     "implicit",
			resourceMode:     "implicit",
			extra:            url.Values{"resource": []string{resource}},
			expectedResource: []string{resource},
		},
		{
			name:             "ShouldNotApplyImplicitModesWhenTheAudienceIsRequested",
			audienceMode:     "implicit",
			resourceMode:     "implicit",
			extra:            url.Values{"audience": []string{audience}},
			expectedAudience: []string{audience},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := newProvider(t, tc.audienceMode, tc.resourceMode)

			t.Run("ClientCredentials", func(t *testing.T) {
				requester, err := clientCredentials(provider, tc.extra)
				require.NoError(t, err)

				assert.ElementsMatch(t, tc.expectedAudience, requester.GetRequestedAudience())
				assert.ElementsMatch(t, tc.expectedResource, requester.GetRequestedResource())
			})

			t.Run("Authorization", func(t *testing.T) {
				requester, err := authorization(provider, tc.extra)
				require.NoError(t, err)

				assert.ElementsMatch(t, tc.expectedAudience, requester.GetRequestedAudience())
				assert.ElementsMatch(t, tc.expectedResource, requester.GetRequestedResource())
			})
		})
	}
}
