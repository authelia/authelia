package oidc_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"testing"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	fjwt "authelia.com/provider/oauth2/token/jwt"
	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestSortedJSONWebKey(t *testing.T) {
	testCases := []struct {
		name     string
		have     []jose.JSONWebKey
		expected []jose.JSONWebKey
	}{
		{
			"ShouldOrderByKID",
			[]jose.JSONWebKey{
				{KeyID: abc},
				{KeyID: "123"},
			},
			[]jose.JSONWebKey{
				{KeyID: "123"},
				{KeyID: abc},
			},
		},
		{
			"ShouldOrderByAlg",
			[]jose.JSONWebKey{
				{Algorithm: "RS256"},
				{Algorithm: "HS256"},
			},
			[]jose.JSONWebKey{
				{Algorithm: "HS256"},
				{Algorithm: "RS256"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(oidc.SortedJSONWebKey(tc.have))

			assert.Equal(t, tc.expected, tc.have)
		})
	}
}

func TestRFC6750Header(t *testing.T) {
	testCases := []struct {
		name     string
		have     *oauthelia2.RFC6749Error
		realm    string
		scope    string
		expected string
	}{
		{
			"ShouldEncodeAll",
			&oauthelia2.RFC6749Error{
				ErrorField:       "invalid_example",
				DescriptionField: "A description",
			},
			"abc",
			"openid",
			`realm="abc",error="invalid_example",error_description="A description",scope="openid"`,
		},
		{
			"ShouldEncodeBasic",
			&oauthelia2.RFC6749Error{
				ErrorField:       "invalid_example",
				DescriptionField: "A description",
			},
			"",
			"",
			`error="invalid_example",error_description="A description"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.RFC6750Header(tc.realm, tc.scope, tc.have))
		})
	}
}

func TestIsJWTProfileAccessToken(t *testing.T) {
	testCases := []struct {
		name     string
		have     *fjwt.Token
		expected bool
	}{
		{
			"ShouldReturnFalseOnNilTokenHeader",
			&fjwt.Token{Header: nil},
			false,
		},
		{
			"ShouldReturnFalseOnEmptyHeader",
			&fjwt.Token{Header: map[string]any{}},
			false,
		},
		{
			"ShouldReturnFalseOnInvalidKeyTypeHeaderType",
			&fjwt.Token{Header: map[string]any{oidc.JWTHeaderKeyType: 123}},
			false,
		},
		{
			"ShouldReturnFalseOnInvalidKeyTypeHeaderValue",
			&fjwt.Token{Header: map[string]any{oidc.JWTHeaderKeyType: "JWT"}},
			false,
		},
		{
			"ShouldReturnTrue",
			&fjwt.Token{Header: map[string]any{oidc.JWTHeaderKeyType: oidc.JWTHeaderTypeValueAccessTokenJWT}},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.IsJWTProfileAccessToken(tc.have.Header))
		})
	}
}

func TestGetLangFromRequester(t *testing.T) {
	testCases := []struct {
		name     string
		have     oauthelia2.Requester
		expected language.Tag
	}{
		{
			"ShouldReturnDefault",
			&TestGetLangRequester{},
			language.English,
		},
		{
			"ShouldReturnEmpty",
			&oauthelia2.Request{},
			language.Tag{},
		},
		{
			"ShouldReturnValueFromRequest",
			&oauthelia2.Request{
				Lang: language.French,
			},
			language.French,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.GetLangFromRequester(tc.have))
		})
	}
}

func TestRequesterRequiresLogin(t *testing.T) {
	testCases := []struct {
		name                     string
		have                     oauthelia2.Requester
		requested, authenticated int64
		expected                 bool
	}{
		{
			name: "ShouldNotRequireWithoutPrompt",
			have: &oauthelia2.Request{
				Form: url.Values{},
			},
			requested:     0,
			authenticated: 0,
			expected:      false,
		},
		{
			name:          "ShouldHandleNilForm",
			have:          &oauthelia2.Request{},
			requested:     0,
			authenticated: 0,
			expected:      false,
		},
		{
			name:          "ShouldHandleNil",
			have:          nil,
			requested:     0,
			authenticated: 0,
			expected:      false,
		},
		{
			name: "ShouldNotRequireWithPromptNone",
			have: &oauthelia2.Request{
				Form: url.Values{oidc.FormParameterPrompt: []string{oidc.PromptNone}},
			},
			requested:     0,
			authenticated: 0,
			expected:      false,
		},
		{
			name: "ShouldNotRequireWithPromptNonePastAuthenticated",
			have: &oauthelia2.Request{
				Form: url.Values{oidc.FormParameterPrompt: []string{oidc.PromptNone}},
			},
			requested:     100000,
			authenticated: 0,
			expected:      false,
		},
		{
			name: "ShouldNotRequireWithPromptLogin",
			have: &oauthelia2.Request{
				Form: url.Values{oidc.FormParameterPrompt: []string{oidc.PromptLogin}},
			},
			requested:     0,
			authenticated: 0,
			expected:      false,
		},
		{
			name: "ShouldRequireWithPromptLoginPastAuthenticated",
			have: &oauthelia2.Request{
				Form: url.Values{oidc.FormParameterPrompt: []string{oidc.PromptLogin}},
			},
			requested:     100000,
			authenticated: 0,
			expected:      true,
		},
		{
			name: "ShouldNotRequireWithMaxAge",
			have: &oauthelia2.Request{
				Form: url.Values{oidc.FormParameterMaximumAge: []string{"100"}},
			},
			requested:     0,
			authenticated: 0,
			expected:      false,
		},
		{
			name: "ShouldRequireWithMaxAgePastAuthenticated",
			have: &oauthelia2.Request{
				Form: url.Values{oidc.FormParameterMaximumAge: []string{"100"}},
			},
			requested:     100000,
			authenticated: 0,
			expected:      true,
		},
		{
			name: "ShouldRequireWithMaxAgePastAuthenticatedInvalid",
			have: &oauthelia2.Request{
				Form: url.Values{oidc.FormParameterMaximumAge: []string{"not100"}},
			},
			requested:     1,
			authenticated: 0,
			expected:      true,
		},
		{
			name:          "ShouldHandleDeviceCode",
			have:          &oauthelia2.DeviceAuthorizeRequest{},
			requested:     0,
			authenticated: 0,
			expected:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.RequesterRequiresLogin(tc.have, time.Unix(tc.requested, 0), time.Unix(tc.authenticated, 0)))
		})
	}
}

func TestRequesterRequiresExplicitConsent(t *testing.T) {
	testCases := []struct {
		name     string
		have     oauthelia2.Requester
		expected bool
	}{
		{
			"ShouldHandleNil",
			nil,
			false,
		},
		{
			"ShouldHandleDefault",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{},
			},
			false,
		},
		{
			"ShouldHandlePromptConsent",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptConsent},
					},
				},
			},
			true,
		},
		{
			"ShouldHandlePromptConsentLogin",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Form: url.Values{
						oidc.FormParameterPrompt: {fmt.Sprintf("%s %s", oidc.PromptConsent, oidc.PromptLogin)},
					},
				},
			},
			true,
		},
		{
			"ShouldHandlePromptLogin",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Form: url.Values{
						oidc.FormParameterPrompt: {fmt.Sprintf(oidc.PromptLogin)},
					},
				},
			},
			true,
		},
		{
			"ShouldHandleScopeWithoutResponseType",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Form: url.Values{
						oidc.FormParameterScope: {fmt.Sprintf("%s %s", oidc.ScopeOpenID, oidc.ScopeOfflineAccess)},
					},
				},
			},
			false,
		},
		{
			"ShouldHandleScopeWithResponseTypeRequestedScopes",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Form: url.Values{
						oidc.FormParameterScope:        {fmt.Sprintf("%s %s", oidc.ScopeOpenID, oidc.ScopeOfflineAccess)},
						oidc.FormParameterResponseType: {oidc.ResponseTypeAuthorizationCodeFlow},
					},
					RequestedScope: []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
				},
				ResponseTypes: []string{oidc.ResponseTypeAuthorizationCodeFlow},
			},
			true,
		},
		{
			"ShouldHandleScopeWithResponseTypeGrantedScopes",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Form: url.Values{
						oidc.FormParameterScope:        {fmt.Sprintf("%s %s", oidc.ScopeOpenID, oidc.ScopeOfflineAccess)},
						oidc.FormParameterResponseType: {oidc.ResponseTypeAuthorizationCodeFlow},
					},
					GrantedScope: []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
				},
				ResponseTypes: []string{oidc.ResponseTypeAuthorizationCodeFlow},
			},
			true,
		},
		{
			"ShouldHandleScopeWithResponseTypeNoScopes",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Form: url.Values{
						oidc.FormParameterScope: {oidc.ScopeOpenID},
					},
					GrantedScope:   []string{oidc.ScopeOpenID},
					RequestedScope: []string{oidc.ScopeOpenID},
				},
				ResponseTypes: []string{oidc.ResponseTypeAuthorizationCodeFlow},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.RequesterRequiresExplicitConsent(tc.have))

			if tc.have != nil {
				assert.Equal(t, tc.expected, oidc.FormRequiresExplicitConsent(tc.have.GetRequestForm()))
			}
		})
	}
}

func TestRequesterIsAuthorizeCodeFlow(t *testing.T) {
	testCases := []struct {
		name     string
		have     oauthelia2.Requester
		expected bool
	}{
		{
			"ShouldHandleNil",
			nil,
			false,
		},
		{
			"ShouldHandleEmpty",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{},
			},
			false,
		},
		{
			"ShouldHandleResponseType",
			&oauthelia2.AuthorizeRequest{
				Request:       oauthelia2.Request{},
				ResponseTypes: []string{oidc.ResponseTypeAuthorizationCodeFlow},
			},
			true,
		},
		{
			"ShouldHandleResponseType",
			&oauthelia2.AuthorizeRequest{
				Request:       oauthelia2.Request{},
				ResponseTypes: []string{oidc.ResponseTypeImplicitFlowToken},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.RequesterIsAuthorizeCodeFlow(tc.have))
		})
	}
}

func TestIsPushedAuthorizedRequest(t *testing.T) {
	testCases := []struct {
		name     string
		have     oauthelia2.Requester
		expected bool
	}{
		{
			"ShouldHandleNil",
			nil,
			false,
		},
		{
			"ShouldHandleEmpty",
			&oauthelia2.AuthorizeRequest{},
			false,
		},
		{
			"ShouldHandleNormalRequestObject",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Form: url.Values{
						oidc.FormParameterRequestURI: {"https://example.com/x.jwt"},
					},
				},
			},
			false,
		},
		{
			"ShouldHandleNormalRequestObject",
			&oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Form: url.Values{
						oidc.FormParameterRequestURI: {fmt.Sprintf("%sabc123", oidc.RedirectURIPrefixPushedAuthorizationRequestURN)},
					},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.IsPushedAuthorizedRequest(tc.have, oidc.RedirectURIPrefixPushedAuthorizationRequestURN))
		})
	}
}

func TestAccessResponderToClearMap(t *testing.T) {
	testCases := []struct {
		name     string
		have     oauthelia2.AccessResponder
		expected map[string]any
	}{
		{
			"ShouldHandleNil",
			nil,
			nil,
		},
		{
			"ShouldHandleDefault",
			&oauthelia2.AccessResponse{
				Extra: map[string]any{},
			},
			map[string]any{
				"access_token": "authelia_at_**************",
				"token_type":   "",
			},
		},
		{
			"ShouldHandleTokens",
			&oauthelia2.AccessResponse{
				TokenType: "bearer",
				Extra: map[string]any{
					"refresh_token": "secret_value",
					"id_token":      "secret_value",
					"active":        true,
				},
			},
			map[string]any{
				"access_token":  "authelia_at_**************",
				"refresh_token": "authelia_rt_**************",
				"id_token":      "*********.***********.*************",
				"token_type":    "bearer",
				"active":        true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.AccessResponderToClearMap(tc.have))
		})
	}
}

func TestHydrateClientCredentialsFlowSessionWithAccessRequest(t *testing.T) {
	testCases := []struct {
		name     string
		ctx      oidc.Context
		have     oauthelia2.Client
		expected *oidc.Session
		err      string
	}{
		{
			"ShouldHandleNil",
			&TestContext{Context: context.Background()},
			nil,
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{},
			},
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Failed to get the client for the request.",
		},
		{
			"ShouldHandleIssuerError",
			&TestContext{Context: context.Background(), IssuerURLFunc: func() (issuerURL *url.URL, err error) {
				return nil, errors.New("issuer error")
			}},
			nil,
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{},
			},
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Failed to determine the issuer with error: issuer error.",
		},
		{
			"ShouldHandleNormalCase",
			&TestContext{
				Context: context.Background(),
				Clock:   clock.NewFixed(time.Unix(1000, 0).UTC()),
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return url.ParseRequestURI("https://auth.example.com")
				},
			},
			&oidc.RegisteredClient{
				ID: "example",
			},
			&oidc.Session{
				ClientID:          "example",
				ClientCredentials: true,
				DefaultSession: &openid.DefaultSession{
					Headers: &fjwt.Headers{
						Extra: map[string]any{},
					},
					Claims: &fjwt.IDTokenClaims{
						Issuer:   "https://auth.example.com",
						Subject:  "example",
						IssuedAt: fjwt.NewNumericDate(time.Unix(1000, 0).UTC()),
						Extra:    map[string]any{},
					},
					RequestedAt: time.Unix(1000, 0).UTC(),
				},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			session := &oidc.Session{
				DefaultSession: &openid.DefaultSession{},
			}

			err := oidc.HydrateClientCredentialsFlowSessionWithAccessRequest(tc.ctx, tc.have, session)

			if tc.err != "" {
				assert.EqualError(t, oauthelia2.ErrorToDebugRFC6749Error(err), tc.err)
			} else {
				assert.NoError(t, oauthelia2.ErrorToDebugRFC6749Error(err))
			}

			assert.Equal(t, tc.expected, session)
		})
	}
}

func TestInitializeSessionDefaults(t *testing.T) {
	testCases := []struct {
		name string
		have *oidc.Session
	}{
		{
			"ShouldHandleDefault",
			&oidc.Session{},
		},
		{
			"ShouldHandleNil",
			nil,
		},
		{
			"ShouldHandleMissingExtra",
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{
					Claims: &fjwt.IDTokenClaims{},
				},
			},
		},
		{
			"ShouldHandleMissingClaims",
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var expected *oidc.Session

			if tc.have != nil {
				expected = &oidc.Session{
					DefaultSession: &openid.DefaultSession{
						Headers: &fjwt.Headers{
							Extra: map[string]any{},
						},
						Claims: &fjwt.IDTokenClaims{
							Extra: map[string]any{},
						},
					},
					AccessToken: nil,
				}
			}

			oidc.InitializeSessionDefaults(tc.have)

			assert.Equal(t, expected, tc.have)
		})
	}
}

func TestIsAccessToken(t *testing.T) {
	testCases := []struct {
		name     string
		ctx      oidc.Context
		value    string
		expected bool
		err      string
	}{
		{
			"ShouldHandleNilCtx",
			nil,
			"abc",
			false,
			"error occurred getting configuration: context wasn't provided",
		},
		{
			"ShouldHandleNoOIDC",
			&TestContext{},
			"authelia_at_example",
			false,
			"",
		},
		{
			"ShouldHandleNoBearer",
			&TestContext{Config: schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{},
				},
			}},
			"authelia_at_example",
			false,
			"",
		},
		{
			"ShouldHandleOpaqueWithoutChecksum",
			&TestContext{Config: schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
							BearerAuthorization: true,
						},
					},
				},
			}},
			"authelia_at_example",
			false,
			"",
		},
		{
			"ShouldHandleOpaque",
			&TestContext{Config: schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
							BearerAuthorization: true,
						},
					},
				},
			}},
			"authelia_at_example.checksum",
			true,
			"",
		},
		{
			"ShouldHandleMaybeSignedJWTDecodeError",
			&TestContext{
				Config: schema.Configuration{
					IdentityProviders: schema.IdentityProviders{
						OIDC: &schema.IdentityProvidersOpenIDConnect{
							Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
								BearerAuthorization: true,
							},
						},
					},
				},
			},
			"abc.123.zyz",
			false,
			"error occurred parsing bearer token: token is malformed: could not JSON decode header: invalid character 'i' looking for beginning of value",
		},
		{
			"ShouldHandleMaybeSignedJWTIssuerError",
			&TestContext{
				Config: schema.Configuration{
					IdentityProviders: schema.IdentityProviders{
						OIDC: &schema.IdentityProvidersOpenIDConnect{
							Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
								BearerAuthorization: true,
							},
						},
					},
				},
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return nil, errors.New("issuer error")
				},
			},
			"abc.123.zyz",
			false,
			"error occurred determining the issuer: issuer error",
		},
		{
			"ShouldHandleMaybeSignedJWTNotAccessToken",
			&TestContext{
				Config: schema.Configuration{
					IdentityProviders: schema.IdentityProviders{
						OIDC: &schema.IdentityProvidersOpenIDConnect{
							Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
								BearerAuthorization: true,
							},
						},
					},
				},
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return url.ParseRequestURI("https://auth.example.com")
				},
			},
			"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6IjgyM2M2MjM3MjA4Y2EyNTMxZmZlNDAxMzU4OTMyYzQ3In0.eyJpc3MiOiJodHRwczovL2lkcC5sb2NhbCIsImF1ZCI6Im15X2NsaWVudF9hcHAiLCJzdWIiOiI1YmU4NjM1OTA3M2M0MzRiYWQyZGEzOTMyMjIyZGFiZSIsImV4cCI6MTc1ODQzMTQ2NywiaWF0IjoxNzU4NDMxMTY3fQ.vn9HvU8O5xbV2KQzdhr2NPnyPwNrMzFoFQ8OC2TszgIiTMDmWd35CG1Vu_8IQjVMz_NN1Ka8iGltaB7YT4l4l2L2l6pkGy2ttUo_QtUCHB8PEDmaxb5nybYcvA9pmdVfuZ06jZ4yPmQddoajXCznci9yQ1itDix5kBGQv3D8oOPqD9Np6e5nEauMWr-G_jDaEt5R4u9RWiEnYnESRMmRavH-LcSXCG80Blih8TuejtS9hmRxMD-QQMdtgMm7VEF3sLUUyHP4Bvj9s-7kW42dzIEOxQtPPlkFdoO6mA01Kx57wMkgznCpVZQNCLYqnsLguWT8vV7vwq8RR8lC7trHIw",
			false,
			"error occurred checking the token: the token is not a JWT profile access token",
		},
		{
			"ShouldHandleMaybeSignedJWTWrongIssuer",
			&TestContext{
				Config: schema.Configuration{
					IdentityProviders: schema.IdentityProviders{
						OIDC: &schema.IdentityProvidersOpenIDConnect{
							Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
								BearerAuthorization: true,
							},
						},
					},
				},
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return url.ParseRequestURI("https://auth.example.com")
				},
			},
			"eyJ0eXAiOiJhdCtqd3QiLCJhbGciOiJSUzI1NiIsImtpZCI6IjgyM2M2MjM3MjA4Y2EyNTMxZmZlNDAxMzU4OTMyYzQ3In0.eyJpc3MiOiJodHRwczovL2lkcC5sb2NhbCIsImF1ZCI6ImFwaTEiLCJzdWIiOiI1YmU4NjM1OTA3M2M0MzRiYWQyZGEzOTMyMjIyZGFiZSIsImNsaWVudF9pZCI6Im15X2NsaWVudF9hcHAiLCJleHAiOjE3NTg0MzQ2NjEsImlhdCI6MTc1ODQzMTA2MSwianRpIjoiYmYxYTZmMDhkMDY0YjkwOGM2NzdmN2IwYTYwOTRlNWQifQ.JwlHb8spi7yryR5OLW-4xAtSHEzfrNdJYrT8zpZZVZnxoAqZ93s3wyoUkEbBPtakTDUl-pcfN1D_nZ7wbgRyinsMjusSrKfkj93dXvezs7vjSJ13cdCuyB4SbhamaIuPk3JyBVpTImczBwC0SQ6khBT4G1A7EEEkkoqa3ldvgI3Z2aY8Zdfb9HF8iEcm5lWpu3Vy50yXNG0IgSxegvmo8wb47C-EFUC1Dh4CVV3_YlvyWrlLNsSTQxnik_8cYfmod8TpGCiSbstH869e-JLSnIkADSN_J-RlQ5V2YFtiEhS9MvCnwjTp0hTXzMkE_k_iIsRMxMNu3VGCURP46s9Fgg",
			false,
			"error occurred checking the token: the token issuer 'https://idp.local' does not match the expected 'https://auth.example.com'",
		},
		{
			"ShouldHandleMaybeSignedJWT",
			&TestContext{
				Config: schema.Configuration{
					IdentityProviders: schema.IdentityProviders{
						OIDC: &schema.IdentityProvidersOpenIDConnect{
							Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
								BearerAuthorization: true,
							},
						},
					},
				},
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return url.ParseRequestURI("https://auth.example.com")
				},
			},
			"eyJ0eXAiOiJhdCtqd3QiLCJhbGciOiJSUzI1NiIsImtpZCI6IjgyM2M2MjM3MjA4Y2EyNTMxZmZlNDAxMzU4OTMyYzQ3In0.eyJpc3MiOiJodHRwczovL2F1dGguZXhhbXBsZS5jb20iLCJhdWQiOiJhcGkxIiwic3ViIjoiNWJlODYzNTkwNzNjNDM0YmFkMmRhMzkzMjIyMmRhYmUiLCJjbGllbnRfaWQiOiJteV9jbGllbnRfYXBwIiwiZXhwIjoxNzU4NDM0ODQ1LCJpYXQiOjE3NTg0MzEyNDUsImp0aSI6IjZlMDRhYmY4NGRjYmE1ODg5NGJkZjRhOWVkOWUyNTU4In0.VaDejpvD6pUvCj07LQ3AuoUJWA609oHhnIXXQ0DNzTdufk8zAEDsVlDzotMlXxph0QXJtCkJ-HCTG1M7fNZUOGHbImQQASFoGy15sdSvmn-4HcAM-FPwnLQ8AVwEAwkCV_2EHYfLgORm0HxX9YLKiaXhn0VyMn3oz9oNnCBC0jEYA-rGZR1LwtEQZacIUPOJxksrGrMu1icgjTk0JsUeYBF7WYpdJwiZBgkcLBtVDSDyJb6mdmk8ha1ZR89RI66lJlhXH8YVlYgUz-WuhgH7BQ9x-l8mE8j0SuiyNNGZug9T2Zo51VI_QPowalQjX-QXen29K4J_imMOvQ1Gh5I6cA",
			true,
			"",
		},
	}

	//
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := oidc.IsAccessToken(tc.ctx, tc.value)
			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}

type TestGetLangRequester struct {
}

func (t TestGetLangRequester) SetRequestedAt(rat time.Time) {}

func (t TestGetLangRequester) SetID(id string) {}

func (t TestGetLangRequester) GetID() string {
	return ""
}

func (t TestGetLangRequester) GetRequestedAt() (requestedAt time.Time) {
	return time.Now().UTC()
}

func (t TestGetLangRequester) GetClient() (client oauthelia2.Client) {
	return nil
}

func (t TestGetLangRequester) GetRequestedScopes() (scopes oauthelia2.Arguments) {
	return nil
}

func (t TestGetLangRequester) GetRequestedAudience() (audience oauthelia2.Arguments) {
	return nil
}

func (t TestGetLangRequester) SetRequestedScopes(scopes oauthelia2.Arguments) {}

func (t TestGetLangRequester) SetRequestedAudience(audience oauthelia2.Arguments) {}

func (t TestGetLangRequester) AppendRequestedScope(scope string) {}

func (t TestGetLangRequester) GetGrantedScopes() (grantedScopes oauthelia2.Arguments) {
	return nil
}

func (t TestGetLangRequester) GetGrantedAudience() (grantedAudience oauthelia2.Arguments) {
	return nil
}

func (t TestGetLangRequester) GrantScope(scope string) {}

func (t TestGetLangRequester) GrantAudience(audience string) {}

func (t TestGetLangRequester) GetSession() (session oauthelia2.Session) {
	return nil
}

func (t TestGetLangRequester) SetSession(session oauthelia2.Session) {}

func (t TestGetLangRequester) GetRequestForm() url.Values {
	return nil
}

func (t TestGetLangRequester) Merge(requester oauthelia2.Requester) {}

func (t TestGetLangRequester) Sanitize(allowedParameters []string) oauthelia2.Requester {
	return nil
}
