package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2DeviceAuthorizationPOST(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID: []string{testOIDCDeviceCodeID},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_request", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpDeviceAuthorizationIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleUnknownClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:        []string{"not-a-client"},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusUnauthorized, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error during the Device Authorization Flow", regexpErrorClientAuthenticationFailed)
	})

	t.Run("ShouldHandleClientWithoutDeviceCodeGrant", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error during the Device Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldIssueDeviceCode", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.StorageMock.EXPECT().
			SaveOAuth2DeviceCodeSession(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:        []string{testOIDCDeviceCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.NotEmpty(t, response["device_code"])
		assert.NotEmpty(t, response["user_code"])
		assert.NotEmpty(t, response["verification_uri"])
	})
}

func TestOAuth2DeviceAuthorizationPUT(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, nil)

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpDeviceAuthorizationUserFlowIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleUnknownUserCode", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByUserCode(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil, sql.ErrNoRows)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{"ABCDEFGH"},
			oidc.FormParameterFlowID:   []string{uuid.Must(uuid.NewRandom()).String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleMalformedFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{"not-a-uuid"},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error to parse the flow ID during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleUnknownConsentSession", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCConsentStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{uuid.Must(uuid.NewRandom()).String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error to load the consent session during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSubjectMismatch", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to match the session to the subject during the User Authorization Flow", nil)
	})

	t.Run("ShouldHandleInsufficientAuthenticationLevel", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCDeviceCodeClient(t)
		client.AuthorizationPolicy = "two_factor"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed as the user did not satisfy the client authorization policy during the User Authorization Flow", nil)
	})

	t.Run("ShouldApproveUserAuthorization", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		subject := mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID)

		consent := newTestOIDCConsentSession(t, mock, subject)
		consent.ClientID = testOIDCDeviceCodeID
		consent.Authorized = true
		consent.GrantedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID}

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Times(1).
			Return(nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)
	})
}

func TestOAuth2DeviceAuthorizationPUTExtra(t *testing.T) {
	setupPUT := func(t *testing.T, mutate func(client *schema.IdentityProvidersOpenIDConnectClient)) (mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
		t.Helper()

		mock = mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))

		client := newTestOIDCDeviceCodeClient(t)

		if mutate != nil {
			mutate(&client)
		}

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode = mustGetTestOIDCUserCode(t, mock)

		subject := mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID)

		consent = newTestOIDCConsentSession(t, mock, subject)
		consent.ClientID = testOIDCDeviceCodeID
		consent.Authorized = true
		consent.GrantedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID}

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		return mock, userCode, consent
	}

	newRequest := func(t *testing.T, userCode string, consent *model.OAuth2ConsentSession) (*httptest.ResponseRecorder, *http.Request) {
		t.Helper()

		return newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})
	}

	t.Run("ShouldHandleUserDetailsError", func(t *testing.T) {
		mock, userCode, consent := setupPUT(t, nil)
		defer mock.Close()

		mock.UserProviderMock.EXPECT().
			GetDetailsExtended(gomock.Eq(testUsername)).
			Return(nil, fmt.Errorf("error in backend"))

		rw, r := newRequest(t, userCode, consent)

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to obtain the user details during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSaveConsentSessionGrantedError", func(t *testing.T) {
		mock, userCode, consent := setupPUT(t, nil)
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Return(fmt.Errorf("error in db"))

		rw, r := newRequest(t, userCode, consent)

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request had an error while saving the session during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldMergeAudienceWhenClaimsStrategyRequires", func(t *testing.T) {
		mock, userCode, consent := setupPUT(t, func(client *schema.IdentityProvidersOpenIDConnectClient) {
			client.ClaimsPolicy = testOIDCClaimsPolicyMerged
		})

		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Return(nil)

		rw, r := newRequest(t, userCode, consent)

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusOK, rw.Code)
	})
}

func TestOAuth2DeviceAuthorizationErrorPaths(t *testing.T) {
	t.Run("ShouldHandleDeviceAuthorizeResponseError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.StorageMock.EXPECT().
			SaveOAuth2DeviceCodeSession(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("error in db"))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:        []string{testOIDCDeviceCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request had an error while trying to create a response during the Device Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSessionProviderError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.Ctx.Request.Header.Set("X-Original-URL", "https://unknown.example.org/api/oidc/device-authorization")

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to obtain the user session during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSubjectLookupError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			LoadUserOpaqueIdentifierBySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("error in db"))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to obtain the user subject value for the user during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleMalformedClaimsParameter", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		rwd, rd := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:        []string{testOIDCDeviceCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
			oidc.FormParameterClaims:          []string{"not json"},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rwd, rd)

		require.Equal(t, http.StatusOK, rwd.Code)

		userCode, ok := getTestOAuth2ErrorResponse(t, rwd)[oidc.FormParameterUserCode].(string)

		require.True(t, ok)

		subject := mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID)

		consent := newTestOIDCConsentSession(t, mock, subject)
		consent.ClientID = testOIDCDeviceCodeID

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpDeviceAuthorizationClaimsParseError, regexpAnyError)
	})

	t.Run("ShouldHandleUserAuthorizeResponseError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		subject := mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID)

		consent := newTestOIDCConsentSession(t, mock, subject)
		consent.ClientID = testOIDCDeviceCodeID
		consent.Authorized = true
		consent.GrantedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID}

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2Session(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(fmt.Errorf("error in db"))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request had an error while attempting to generate the responder during the User Authorization Flow", regexpAnyError)
	})
}

const testOIDCDeviceAuthorizationEndpoint = "https://login.example.com:8080/api/oidc/device-authorization"

func newTestOIDCDeviceCodeClient(t *testing.T) schema.IdentityProvidersOpenIDConnectClient {
	t.Helper()

	return schema.IdentityProvidersOpenIDConnectClient{
		ID:                      testOIDCDeviceCodeID,
		Secret:                  mustDecodeTestSecret(t, testOIDCClientSecretDigest),
		AuthorizationPolicy:     "one_factor",
		GrantTypes:              []string{oidc.GrantTypeDeviceCode},
		ResponseTypes:           []string{oidc.ResponseTypeAuthorizationCodeFlow},
		Scopes:                  []string{oidc.ScopeOpenID, oidc.ScopeProfile},
		RedirectURIs:            []string{testOIDCRedirectURI},
		TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
	}
}
