package oidc_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOpenIDConnectProvider_FlowClientCredentialsTokenHandler(t *testing.T) {
	provider := oidc.NewOpenIDConnectProvider(&schema.OpenIDConnect{
		IssuerCertificateChain:   schema.X509CertificateChain{},
		IssuerPrivateKey:         keyRSA2048,
		EnablePKCEPlainChallenge: true,
		HMACSecret:               badhmac,
		Clients: []schema.OpenIDConnectClient{
			{
				ID:                  myclient,
				Secret:              tOpenIDConnectPlainTextClientSecret,
				SectorIdentifier:    url.URL{Host: examplecomsid},
				AuthorizationPolicy: onefactor,
				RedirectURIs: []string{
					examplecom,
				},
			},
		},
	}, nil, nil)

	testCases := []struct {
		name     string
		have     fosite.AccessRequester
		client   fosite.Client
		expected fosite.AccessRequester
	}{
		{
			"ShouldGrantAll",
			&fosite.AccessRequest{
				Request: fosite.Request{
					RequestedScope:    fosite.Arguments{},
					RequestedAudience: fosite.Arguments{},
				},
			},

			&oidc.BaseClient{
				Scopes:   fosite.Arguments{"abc", "123"},
				Audience: fosite.Arguments{"aud", "123"},
			},
			&fosite.AccessRequest{
				Request: fosite.Request{
					GrantedScope:    fosite.Arguments{"abc", "123"},
					GrantedAudience: fosite.Arguments(nil),
				},
			},
		},
		{
			"ShouldGrantRequested",
			&fosite.AccessRequest{
				Request: fosite.Request{
					RequestedScope:    fosite.Arguments{"abc"},
					RequestedAudience: fosite.Arguments{"aud"},
				},
			},

			&oidc.BaseClient{
				Scopes:   fosite.Arguments{"abc", "123"},
				Audience: fosite.Arguments{"aud", "123"},
			},
			&fosite.AccessRequest{
				Request: fosite.Request{
					GrantedScope:    fosite.Arguments{"abc"},
					GrantedAudience: fosite.Arguments{"aud"},
				},
			},
		},
		{
			"ShouldNotGrantRequestedUnknown",
			&fosite.AccessRequest{
				Request: fosite.Request{
					RequestedScope:    fosite.Arguments{"abc", "987"},
					RequestedAudience: fosite.Arguments{"aud", "987"},
				},
			},

			&oidc.BaseClient{
				Scopes:   fosite.Arguments{"abc", "123"},
				Audience: fosite.Arguments{"aud", "123"},
			},
			&fosite.AccessRequest{
				Request: fosite.Request{
					GrantedScope:    fosite.Arguments{"abc"},
					GrantedAudience: fosite.Arguments{"aud"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Nil(t, tc.have.GetGrantedScopes())
			require.Nil(t, tc.have.GetGrantedAudience())

			provider.FlowClientCredentialsTokenHandler(context.TODO(), tc.have, tc.client)

			assert.Equal(t, tc.expected.GetGrantedScopes(), tc.have.GetGrantedScopes())
			assert.Equal(t, tc.expected.GetGrantedAudience(), tc.have.GetGrantedAudience())
		})
	}
}
