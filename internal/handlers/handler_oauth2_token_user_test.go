// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2TokenPOSTRefreshUserState(t *testing.T) {
	testCases := []struct {
		name     string
		identity func() (*model.UserOpaqueIdentifier, error)
		details  error
		status   int
		expected string
	}{
		{
			"ShouldIssueTokensWhenUserActive",
			nil,
			nil,
			http.StatusOK,
			"",
		},
		{
			"ShouldRejectWhenUserDisabled",
			nil,
			authentication.ErrUserNotFound,
			http.StatusBadRequest,
			"invalid_grant",
		},
		{
			"ShouldRejectWhenSubjectUnknown",
			func() (*model.UserOpaqueIdentifier, error) {
				return nil, nil
			},
			nil,
			http.StatusBadRequest,
			"invalid_grant",
		},
		{
			"ShouldFailWhenStorageFails",
			func() (*model.UserOpaqueIdentifier, error) {
				return nil, fmt.Errorf("error in db")
			},
			nil,
			http.StatusInternalServerError,
			"server_error",
		},
		{
			"ShouldFailWhenUserProviderFails",
			nil,
			fmt.Errorf("ldap unavailable"),
			http.StatusInternalServerError,
			"server_error",
		},
	}

	for _, jwtProfile := range []bool{false, true} {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s/JWTProfile=%t", tc.name, jwtProfile), func(t *testing.T) {
				mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
				defer mock.Close()

				client := newTestOIDCAuthorizationCodeClient(t)
				client.ConsentMode = "explicit"
				client.GrantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken}
				client.Scopes = []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeOfflineAccess}

				if jwtProfile {
					client.AccessTokenSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
					client.AccessTokenSignedResponseKeyID = testOIDCKeyID
				}

				config := newTestOIDCConfig(t)
				config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

				setupTestOIDCProvider(t, mock, config)
				setupTestOIDCSessionStore(t, mock)

				mock.StorageMock.EXPECT().
					LoadOAuth2RefreshTokenSessionAccessSignature(gomock.Any(), gomock.Any()).
					AnyTimes().
					Return("", nil)

				identifier := model.UserOpaqueIdentifier{Service: "openid", Username: testUsername, Identifier: uuid.Must(uuid.NewRandom())}

				var broken bool

				mock.StorageMock.EXPECT().
					LoadUserOpaqueIdentifierBySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes().
					Return(&identifier, nil)

				mock.StorageMock.EXPECT().
					LoadUserOpaqueIdentifier(gomock.Any(), gomock.Any()).
					AnyTimes().
					DoAndReturn(func(_ any, _ uuid.UUID) (*model.UserOpaqueIdentifier, error) {
						if broken && tc.identity != nil {
							return tc.identity()
						}

						value := identifier

						return &value, nil
					})

				details := newTestOIDCUserDetails()

				mock.UserProviderMock.EXPECT().
					GetDetails(testUsername).
					AnyTimes().
					Return(details, nil)

				mock.UserProviderMock.EXPECT().
					GetDetailsExtended(testUsername).
					AnyTimes().
					DoAndReturn(func(_ string) (*authentication.UserDetailsExtended, error) {
						if broken && tc.details != nil {
							return nil, tc.details
						}

						return &authentication.UserDetailsExtended{UserDetails: details}, nil
					})

				values := newTestOIDCAuthorizationValues()
				values.Set(oidc.FormParameterScope, strings.Join([]string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}, " "))

				consent := newTestOIDCConsentSession(t, mock, identifier.Identifier)
				consent.Form = values.Encode()
				consent.Authorized = true
				consent.RequestedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}
				consent.GrantedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}
				consent.RespondedAt = sql.NullTime{Time: mock.Ctx.GetClock().Now(), Valid: true}

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
					AnyTimes().
					Return(consent, nil)

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
					AnyTimes().
					Return(nil)

				mock.Ctx.Request.SetRequestURI(testOIDCAuthorizationEndpoint + "?consent_id=" + consent.ChallengeID.String())

				rwa, ra := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

				OAuth2AuthorizationGET(mock.Ctx, rwa, ra)

				require.Equal(t, http.StatusSeeOther, rwa.Code)

				location, err := url.Parse(rwa.Header().Get(fasthttp.HeaderLocation))

				require.NoError(t, err)
				require.Empty(t, location.Query().Get("error"), location.Query().Get("error_description"))

				rwc, rc := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
					testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
					oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
					testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
					testOIDCFormParameterCode:         []string{location.Query().Get(testOIDCFormParameterCode)},
					oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
				})

				OAuth2TokenPOST(mock.Ctx, rwc, rc)

				require.Equal(t, http.StatusOK, rwc.Code)

				refreshToken, ok := getTestOAuth2ErrorResponse(t, rwc)["refresh_token"].(string)

				require.True(t, ok)

				broken = true

				refresh := func() *httptest.ResponseRecorder {
					rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
						testOIDCFormParameterGrantType:    []string{oidc.GrantTypeRefreshToken},
						oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
						testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
						"refresh_token":                   []string{refreshToken},
					})

					OAuth2TokenPOST(mock.Ctx, rw, r)

					return rw
				}

				rw := refresh()
				response := getTestOAuth2ErrorResponse(t, rw)

				assert.Equal(t, tc.status, rw.Code, response)

				if tc.expected == "" {
					assert.NotEmpty(t, response["access_token"])

					return
				}

				assert.Equal(t, tc.expected, response["error"])
				assert.NotContains(t, response, "access_token")
				assert.NotContains(t, response, "id_token")

				if tc.status == http.StatusInternalServerError {
					broken = false

					assert.Equal(t, http.StatusOK, refresh().Code, "a transient failure should not consume the refresh token")
				}
			})
		}
	}
}

func TestOAuth2TokenPOSTGrantAuthorizationPolicy(t *testing.T) {
	_, lan, err := net.ParseCIDR("192.168.2.0/24")

	require.NoError(t, err)

	const policyName = "token_policy"

	permissive := schema.IdentityProvidersOpenIDConnectPolicy{DefaultPolicy: "one_factor"}

	testCases := []struct {
		name     string
		policy   schema.IdentityProvidersOpenIDConnectPolicy
		status   int
		expected string
	}{
		{
			"ShouldIssueTokensWhenPolicyPermitsUser",
			permissive,
			http.StatusOK,
			"",
		},
		{
			"ShouldRejectWhenPolicyDeniesUserGroup",
			schema.IdentityProvidersOpenIDConnectPolicy{
				DefaultPolicy: "one_factor",
				Rules:         []schema.IdentityProvidersOpenIDConnectPolicyRule{{Policy: "deny", Subjects: schema.AccessControlRuleSubjects{{"group:dev"}}}},
			},
			http.StatusBadRequest,
			"invalid_grant",
		},
		{
			"ShouldRejectWhenPolicyDeniesByDefault",
			schema.IdentityProvidersOpenIDConnectPolicy{DefaultPolicy: "deny"},
			http.StatusBadRequest,
			"invalid_grant",
		},
		{
			"ShouldRejectWhenPolicyRequiresHigherLevel",
			schema.IdentityProvidersOpenIDConnectPolicy{DefaultPolicy: "two_factor"},
			http.StatusBadRequest,
			"invalid_grant",
		},
		{
			"ShouldIssueTokensWhenPolicyDependsOnNetwork",
			schema.IdentityProvidersOpenIDConnectPolicy{
				DefaultPolicy: "deny",
				Rules:         []schema.IdentityProvidersOpenIDConnectPolicyRule{{Policy: "one_factor", Networks: []*net.IPNet{lan}}},
			},
			http.StatusOK,
			"",
		},
	}

	grants := []struct {
		name  string
		setup func(t *testing.T, mock *mocks.MockAutheliaCtx, config func(policy schema.IdentityProvidersOpenIDConnectPolicy) *schema.IdentityProvidersOpenIDConnect) (exchange func() *httptest.ResponseRecorder)
	}{
		{oidc.GrantTypeAuthorizationCode, setupTestOAuth2TokenUserAuthorizationCode},
		{oidc.GrantTypeRefreshToken, setupTestOAuth2TokenUserRefreshToken},
		{oidc.GrantTypeDeviceCode, setupTestOAuth2TokenUserDeviceCode},
	}

	for _, grant := range grants {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s/%s", tc.name, grant.name), func(t *testing.T) {
				mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
				defer mock.Close()

				config := func(policy schema.IdentityProvidersOpenIDConnectPolicy) *schema.IdentityProvidersOpenIDConnect {
					c := newTestOIDCConfig(t)
					c.AuthorizationPolicies = map[string]schema.IdentityProvidersOpenIDConnectPolicy{policyName: policy}

					return c
				}

				exchange := grant.setup(t, mock, config)

				setupTestOIDCProvider(t, mock, withTestOIDCClients(config(tc.policy), mock.Ctx.Configuration.IdentityProviders.OIDC.Clients))

				rw := exchange()
				response := getTestOAuth2ErrorResponse(t, rw)

				assert.Equal(t, tc.status, rw.Code, response)

				if tc.expected == "" {
					assert.NotEmpty(t, response["access_token"])
				} else {
					assert.Equal(t, tc.expected, response["error"])
					assert.NotContains(t, response, "access_token")
				}
			})
		}
	}
}

func withTestOIDCClients(config *schema.IdentityProvidersOpenIDConnect, clients []schema.IdentityProvidersOpenIDConnectClient) *schema.IdentityProvidersOpenIDConnect {
	config.Clients = clients

	return config
}

func setupTestOAuth2TokenUserAuthorizationCode(t *testing.T, mock *mocks.MockAutheliaCtx, config func(policy schema.IdentityProvidersOpenIDConnectPolicy) *schema.IdentityProvidersOpenIDConnect) func() *httptest.ResponseRecorder {
	t.Helper()

	client := newTestOIDCAuthorizationCodeClient(t)
	client.ConsentMode = "implicit"
	client.AuthorizationPolicy = "token_policy"

	setupTestOIDCProvider(t, mock, withTestOIDCClients(config(schema.IdentityProvidersOpenIDConnectPolicy{DefaultPolicy: "one_factor"}), []schema.IdentityProvidersOpenIDConnectClient{client}))
	setupTestOIDCSessionStore(t, mock)
	setupTestOIDCConsentStore(t, mock)
	setupTestOIDCSubjectStore(t, mock)
	setupTestOIDCUserDetails(t, mock)

	code := mustGetTestOIDCAuthorizationCode(t, mock, newTestOIDCAuthorizationValues())

	return func() *httptest.ResponseRecorder {
		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			testOIDCFormParameterCode:         []string{code},
			oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		return rw
	}
}

func setupTestOAuth2TokenUserRefreshToken(t *testing.T, mock *mocks.MockAutheliaCtx, config func(policy schema.IdentityProvidersOpenIDConnectPolicy) *schema.IdentityProvidersOpenIDConnect) func() *httptest.ResponseRecorder {
	t.Helper()

	client := newTestOIDCAuthorizationCodeClient(t)
	client.ConsentMode = "explicit"
	client.AuthorizationPolicy = "token_policy"
	client.GrantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken}
	client.Scopes = []string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeOfflineAccess}

	setupTestOIDCProvider(t, mock, withTestOIDCClients(config(schema.IdentityProvidersOpenIDConnectPolicy{DefaultPolicy: "one_factor"}), []schema.IdentityProvidersOpenIDConnectClient{client}))
	setupTestOIDCSessionStore(t, mock)
	setupTestOIDCSubjectStore(t, mock)
	setupTestOIDCUserDetails(t, mock)

	mock.StorageMock.EXPECT().
		LoadOAuth2RefreshTokenSessionAccessSignature(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return("", nil)

	subject := mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID)

	values := newTestOIDCAuthorizationValues()
	values.Set(oidc.FormParameterScope, strings.Join([]string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}, " "))

	consent := newTestOIDCConsentSession(t, mock, subject)
	consent.Form = values.Encode()
	consent.Authorized = true
	consent.RequestedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}
	consent.GrantedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}
	consent.RespondedAt = sql.NullTime{Time: mock.Ctx.GetClock().Now(), Valid: true}

	mock.StorageMock.EXPECT().
		LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
		AnyTimes().
		Return(consent, nil)

	mock.StorageMock.EXPECT().
		SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
		AnyTimes().
		Return(nil)

	mock.Ctx.Request.SetRequestURI(testOIDCAuthorizationEndpoint + "?consent_id=" + consent.ChallengeID.String())

	rwa, ra := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

	OAuth2AuthorizationGET(mock.Ctx, rwa, ra)

	require.Equal(t, http.StatusSeeOther, rwa.Code)

	location, err := url.Parse(rwa.Header().Get(fasthttp.HeaderLocation))

	require.NoError(t, err)
	require.Empty(t, location.Query().Get("error"), location.Query().Get("error_description"))

	rwc, rc := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
		testOIDCFormParameterGrantType:    []string{oidc.GrantTypeAuthorizationCode},
		oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
		testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		testOIDCFormParameterCode:         []string{location.Query().Get(testOIDCFormParameterCode)},
		oidc.FormParameterRedirectURI:     []string{testOIDCRedirectURI},
	})

	OAuth2TokenPOST(mock.Ctx, rwc, rc)

	require.Equal(t, http.StatusOK, rwc.Code)

	refreshToken, ok := getTestOAuth2ErrorResponse(t, rwc)["refresh_token"].(string)

	require.True(t, ok)

	return func() *httptest.ResponseRecorder {
		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeRefreshToken},
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			"refresh_token":                   []string{refreshToken},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		return rw
	}
}

func setupTestOAuth2TokenUserDeviceCode(t *testing.T, mock *mocks.MockAutheliaCtx, config func(policy schema.IdentityProvidersOpenIDConnectPolicy) *schema.IdentityProvidersOpenIDConnect) func() *httptest.ResponseRecorder {
	t.Helper()

	client := newTestOIDCDeviceCodeClient(t)
	client.AuthorizationPolicy = "token_policy"

	setupTestOIDCProvider(t, mock, withTestOIDCClients(config(schema.IdentityProvidersOpenIDConnectPolicy{DefaultPolicy: "one_factor"}), []schema.IdentityProvidersOpenIDConnectClient{client}))
	setupTestOIDCDeviceCodeStore(t, mock)
	setupTestOIDCSessionStore(t, mock)
	setupTestOIDCSubjectStore(t, mock)
	setupTestOIDCUserDetails(t, mock)

	rwd, rd := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
		oidc.FormParameterClientID:        []string{testOIDCDeviceCodeID},
		testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
	})

	OAuth2DeviceAuthorizationPOST(mock.Ctx, rwd, rd)

	require.Equal(t, http.StatusOK, rwd.Code)

	device := getTestOAuth2ErrorResponse(t, rwd)

	userCode, ok := device[oidc.FormParameterUserCode].(string)

	require.True(t, ok)

	deviceCode, ok := device["device_code"].(string)

	require.True(t, ok)

	consent := newTestOIDCDeviceConsentSession(t, mock, mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID), userCode)

	mock.StorageMock.EXPECT().
		LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
		AnyTimes().
		Return(consent, nil)

	mock.StorageMock.EXPECT().
		SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
		AnyTimes().
		Return(nil)

	rwp, rp := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
		oidc.FormParameterUserCode: []string{userCode},
		oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
	})

	OAuth2DeviceAuthorizationPUT(mock.Ctx, rwp, rp)

	require.Equal(t, http.StatusOK, rwp.Code)

	return func() *httptest.ResponseRecorder {
		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
			testOIDCFormParameterGrantType:    []string{oidc.GrantTypeDeviceCode},
			oidc.FormParameterClientID:        []string{testOIDCDeviceCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			"device_code":                     []string{deviceCode},
		})

		OAuth2TokenPOST(mock.Ctx, rw, r)

		return rw
	}
}
