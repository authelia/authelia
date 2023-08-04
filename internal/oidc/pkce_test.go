package oidc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ory/fosite"
	"github.com/ory/fosite/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestPKCEHandler_Misc(t *testing.T) {
	store := storage.NewMemoryStore()
	strategy := &MockCodeStrategy{}
	config := &oidc.Config{}

	handler := &oidc.PKCEHandler{Storage: store, AuthorizeCodeStrategy: strategy, Config: config}

	assert.Nil(t, handler.PopulateTokenEndpointResponse(context.Background(), nil, nil))
	assert.False(t, handler.CanSkipClientAuth(context.Background(), nil))
}

func TestPKCEHandler_CanHandleTokenEndpointRequest(t *testing.T) {
	testCases := []struct {
		name     string
		have     fosite.AccessRequester
		expected bool
	}{
		{
			"ShouldHandleAuthorizeCode",
			&fosite.AccessRequest{
				GrantTypes: fosite.Arguments{oidc.GrantTypeAuthorizationCode},
			},
			true,
		},
		{
			"ShouldNotHandleRefreshToken",
			&fosite.AccessRequest{
				GrantTypes: fosite.Arguments{oidc.GrantTypeRefreshToken},
			},
			false,
		},
		{
			"ShouldNotHandleClientCredentials",
			&fosite.AccessRequest{
				GrantTypes: fosite.Arguments{oidc.GrantTypeClientCredentials},
			},
			false,
		},
		{
			"ShouldNotHandleImplicit",
			&fosite.AccessRequest{
				GrantTypes: fosite.Arguments{oidc.GrantTypeImplicit},
			},
			false,
		},
	}

	store := storage.NewMemoryStore()
	strategy := &MockCodeStrategy{}
	config := &oidc.Config{}

	handler := &oidc.PKCEHandler{Storage: store, AuthorizeCodeStrategy: strategy, Config: config}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, handler.CanHandleTokenEndpointRequest(context.Background(), tc.have))
		})
	}
}

func TestPKCEHandler_HandleAuthorizeEndpointRequest(t *testing.T) {
	store := storage.NewMemoryStore()
	strategy := &MockCodeStrategy{}
	config := &oidc.Config{}

	client := &oidc.BaseClient{ID: "test"}

	handler := &oidc.PKCEHandler{Storage: store, AuthorizeCodeStrategy: strategy, Config: config}

	testCases := []struct {
		name                                      string
		types                                     fosite.Arguments
		enforce, enforcePublicClients, allowPlain bool
		method, challenge, code                   string
		expected                                  string
		client                                    *oidc.BaseClient
	}{
		{
			"ShouldNotHandleBlankResponseModes",
			fosite.Arguments{},
			false,
			false,
			false,
			oidc.PKCEChallengeMethodPlain,
			"challenge",
			"",
			"",
			client,
		},
		{
			"ShouldHandleAuthorizeCodeFlow",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			false,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"challenge",
			"abc123",
			"",
			client,
		},
		{
			"ShouldErrorHandleAuthorizeCodeFlowWithoutCode",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			false,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"challenge",
			"",
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. The PKCE handler must be loaded after the authorize code handler.",
			client,
		},
		{
			"ShouldErrorHandleAuthorizeCodeFlowWithoutChallengeMethodWhenEnforced",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			true,
			false,
			true,
			"",
			"",
			"abc123",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for clients.",
			client,
		},
		{
			"ShouldErrorHandleAuthorizeCodeFlowWithoutChallengeWhenEnforced",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			true,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"",
			"abc123",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for clients.",
			client,
		},
		{
			"ShouldSkipNotEnforced",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			false,
			false,
			true,
			"",
			"",
			"abc123",
			"",
			client,
		},
		{
			"ShouldErrorUnknownChallengeMethod",
			fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
			true,
			false,
			true,
			"abc",
			"abc",
			"abc123",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The code_challenge_method is not supported, use S256 instead.",
			client,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config.ProofKeyCodeExchange.Enforce = tc.enforce
			config.ProofKeyCodeExchange.EnforcePublicClients = tc.enforcePublicClients
			config.ProofKeyCodeExchange.AllowPlainChallengeMethod = tc.allowPlain

			requester := fosite.NewAuthorizeRequest()

			requester.Client = tc.client

			if len(tc.method) > 0 {
				requester.Form.Add(oidc.FormParameterCodeChallengeMethod, tc.method)
			}

			if len(tc.challenge) > 0 {
				requester.Form.Add(oidc.FormParameterCodeChallenge, tc.challenge)
			}

			requester.ResponseTypes = tc.types

			responder := fosite.NewAuthorizeResponse()

			if len(tc.code) > 0 {
				responder.AddParameter(oidc.FormParameterAuthorizationCode, tc.code)
			}

			err := handler.HandleAuthorizeEndpointRequest(context.Background(), requester, responder)

			err = ErrorToRFC6749ErrorTest(err)

			if len(tc.expected) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expected)
			}
		})
	}
}

func TestPKCEHandler_HandleTokenEndpointRequest(t *testing.T) {
	store := storage.NewMemoryStore()
	strategy := &MockCodeStrategy{}
	config := &oidc.Config{}

	handler := &oidc.PKCEHandler{Storage: store, AuthorizeCodeStrategy: strategy, Config: config}

	challenge := "GM6jKJIR6JxgxU5m5Y79WzudqoNmo7PogrhI1_F8eGw"
	verifier := "Nt6MpT7QXtfme55cKv9b23KEAvSEHyjRVtQt5jcjUUmWU9bTzd"

	clientConfidential := &oidc.BaseClient{
		ID:     "test",
		Public: false,
	}

	clientPublic := &oidc.BaseClient{
		ID:     "test",
		Public: true,
	}

	testCases := []struct {
		name                                      string
		grant                                     string
		enforce, enforcePublicClients, allowPlain bool
		method, challenge, verifier               string
		code                                      string
		expected                                  string
		client                                    *oidc.BaseClient
	}{
		{
			"ShouldFailNotAuthCode",
			oidc.GrantTypeClientCredentials,
			false,
			false,
			false,
			oidc.PKCEChallengeMethodSHA256,
			challenge,
			verifier,
			"code-0",
			"The handler is not responsible for this request.",
			clientConfidential,
		},
		{
			"ShouldPassPlainWithConfidentialClientWhenEnforcedWhenAllowPlain",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodPlain,
			"sW6e3dnNWdLMoT9rLrSMgC8xfVNuwnNdAShrGWbysqVoc8s3HK",
			"sW6e3dnNWdLMoT9rLrSMgC8xfVNuwnNdAShrGWbysqVoc8s3HK",
			"code-1",
			"",
			clientConfidential,
		},
		{
			"ShouldFailWithConfidentialClientWithNotAllowPlainWhenPlainWhenEnforced",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			false,
			oidc.PKCEChallengeMethodPlain,
			"sW6e3dnNWdLMoT9rLrSMgC8xfVNuwnNdAShrGWbysqVoc8s3HK",
			"sW6e3dnNWdLMoT9rLrSMgC8xfVNuwnNdAShrGWbysqVoc8s3HK",
			"code-2",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must use the 'S256' PKCE 'code_challenge_method' but the 'plain' method was requested. The server is configured in a way that enforces PKCE 'S256' as challenge method for clients.",
			clientConfidential,
		},
		{
			"ShouldPassWithConfidentialClientWhenNotProvidedWhenNotEnforced",
			oidc.GrantTypeAuthorizationCode,
			false,
			false,
			false,
			"",
			"",
			"",
			"code-3",
			"",
			clientConfidential,
		},
		{
			"ShouldPassWithPublicClientWhenNotProvidedWhenNotEnforced",
			oidc.GrantTypeAuthorizationCode,
			false,
			false,
			false,
			"",
			"",
			"",
			"code-4",
			"",
			clientPublic,
		},
		{
			"ShouldFailWithPublicClientWhenNotProvidedWhenNotEnforcedWhenEnforcedForPublicClients",
			oidc.GrantTypeAuthorizationCode,
			false,
			true,
			false,
			"",
			"",
			"",
			"code-5",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. This client must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for this client.",
			clientPublic,
		},
		{
			"ShouldPassS256WithConfidentialClientWhenEnforcedWhenAllowPlain",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodSHA256,
			challenge,
			verifier,
			"code-6",
			"",
			clientConfidential,
		},
		{
			"ShouldPassS256WithConfidentialClientWhenEnforcedWhenNotAllowPlain",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			false,
			oidc.PKCEChallengeMethodSHA256,
			challenge,
			verifier,
			"code-7",
			"",
			clientConfidential,
		},
		{
			"ShouldFailS256WithConfidentialClientWhenEnforcedWhenAllowPlain",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodSHA256,
			challenge,
			"jSfm3f5nS9eD4eYSHaeQxVBVKxXnfmbWAFQiiAdMAK98EhNifm",
			"code-8",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code challenge did not match the code verifier.",
			clientConfidential,
		},
		{
			"ShouldFailS256WithConfidentialClientWhenVerifierTooShort",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodSHA256,
			challenge,
			"aaaaaaaaaaaaa",
			"code-9",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code verifier must be at least 43 characters.",
			clientConfidential,
		},
		{
			"ShouldFailPlainWithConfidentialClientWhenNotMatching",
			oidc.GrantTypeAuthorizationCode,
			true,
			true,
			true,
			oidc.PKCEChallengeMethodPlain,
			"jqNhtkWBNbP9oz6BjA2ufPWxADqnHJcUNhS8VNzBWLd44ynFzi",
			"uoPgkXQNCiiMzQ4aXeXdzBQaDArGNN9bke8gQWo7qZZ2djrcJZ",
			"code-10",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code challenge did not match the code verifier.",
			clientConfidential,
		},
		{
			"ShouldPassNoPKCESessionWhenNotEnforced",
			oidc.GrantTypeAuthorizationCode,
			false,
			false,
			true,
			"",
			"",
			"",
			"code-11",
			"",
			clientConfidential,
		},
		{
			"ShouldFailNoPKCESessionWhenEnforced",
			oidc.GrantTypeAuthorizationCode,
			true,
			false,
			true,
			"",
			"",
			"",
			"code-12",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for clients.",
			clientConfidential,
		},
		{
			"ShouldFailLongVerifier",
			oidc.GrantTypeAuthorizationCode,
			true,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"fUcMF5287hMEieVQcjvxViWUEUGD9NjG63hELzLtSyPiETpwCjuLuZYJCMJkeAMb3wg6WRHXRzj6KSScu48J7KRDJScEAZbRXjMjR79KQavdqHLVDpv4WQra7teJvGjJfUcMF5287hMEieVQcjvxViWUEUGD9NjG63hELzLtSyPiETpwCjuLuZYJCMJkeAMb3wg6WRHXRzj6KSScu48J7KRDJScEAZbRXjMjR79KQavdqHLVDpv4WQra7teJvGjJ",
			"fUcMF5287hMEieVQcjvxViWUEUGD9NjG63hELzLtSyPiETpwCjuLuZYJCMJkeAMb3wg6WRHXRzj6KSScu48J7KRDJScEAZbRXjMjR79KQavdqHLVDpv4WQra7teJvGjJfUcMF5287hMEieVQcjvxViWUEUGD9NjG63hELzLtSyPiETpwCjuLuZYJCMJkeAMb3wg6WRHXRzj6KSScu48J7KRDJScEAZbRXjMjR79KQavdqHLVDpv4WQra7teJvGjJ",
			"code-13",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code verifier can not be longer than 128 characters.",
			clientConfidential,
		},
		{
			"ShouldFailBadChars",
			oidc.GrantTypeAuthorizationCode,
			true,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"U5katcVxpTd8jU5YpUD9xdoivT55zzMWdMiy2TDA4DH9FJ5bTK45Mhd@",
			"U5katcVxpTd8jU5YpUD9xdoivT55zzMWdMiy2TDA4DH9FJ5bTK45Mhd@",
			"code-14",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code verifier must only contain [a-Z], [0-9], '-', '.', '_', '~'.",
			clientConfidential,
		},
		{
			"ShouldFailNoChallenge",
			oidc.GrantTypeAuthorizationCode,
			false,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"",
			"U5katcVxpTd8jU5YpUD9xdoivT55zzMWdMiy2TDA4DH9FJ5bTK45Mhd",
			"code-15",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The PKCE code verifier was provided but the code challenge was absent from the authorization request.",
			clientConfidential,
		},
		{
			"ShouldFailNoPKCESessionWhenEnforcedNotFound",
			oidc.GrantTypeAuthorizationCode,
			true,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"",
			"abc123",
			"",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. Unable to find initial PKCE data tied to this request The recorded error is: not_found.",
			clientConfidential,
		},
		{
			"ShouldFailInvalidCode",
			oidc.GrantTypeAuthorizationCode,
			true,
			false,
			true,
			oidc.PKCEChallengeMethodPlain,
			"WXNqfH6FCXcJH5oT9eqTM3HTdTh4b2aSvVe9KWkxcHCJJ3FaXF",
			"WXNqfH6FCXcJH5oT9eqTM3HTdTh4b2aSvVe9KWkxcHCJJ3FaXF",
			"BADCODE",
			"The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. Unable to find initial PKCE data tied to this request The recorded error is: not_found.",
			clientConfidential,
		},
		{
			"ShouldPassNotEnforcedNoSession",
			oidc.GrantTypeAuthorizationCode,
			false,
			false,
			true,
			"",
			"",
			"",
			"",
			"",
			clientConfidential,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			config.ProofKeyCodeExchange.Enforce = tc.enforce
			config.ProofKeyCodeExchange.EnforcePublicClients = tc.enforcePublicClients
			config.ProofKeyCodeExchange.AllowPlainChallengeMethod = tc.allowPlain

			if len(tc.code) > 0 {
				strategy.signature = tc.code
			} else {
				strategy.signature = fmt.Sprintf("code-%d", i)
			}

			ar := fosite.NewAuthorizeRequest()

			ar.Client = tc.client

			if len(tc.challenge) != 0 {
				ar.Form.Add(oidc.FormParameterCodeChallenge, tc.challenge)
			}

			if len(tc.method) != 0 {
				ar.Form.Add(oidc.FormParameterCodeChallengeMethod, tc.method)
			}

			if len(tc.code) > 0 {
				require.NoError(t, store.CreatePKCERequestSession(ctx, fmt.Sprintf("code-%d", i), ar))
			}

			r := fosite.NewAccessRequest(nil)
			r.Client = tc.client
			r.GrantTypes = fosite.Arguments{tc.grant}

			if len(tc.verifier) != 0 {
				r.Form.Add(oidc.FormParameterCodeVerifier, tc.verifier)
			}

			err := handler.HandleTokenEndpointRequest(context.Background(), r)
			err = ErrorToRFC6749ErrorTest(err)

			if len(tc.expected) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expected)
			}
		})
	}
}
