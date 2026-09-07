package handlers

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

var testOIDCKeyRSA2048 *rsa.PrivateKey

var regexpJSONWebKeySetIssuerError = regexp.MustCompile(`^JSON Web Key Set Request could not be processed: `)

var (
	regexpOAuth2DiscoveryIssuerError        = regexp.MustCompile(`^OAuth 2\.0 Discovery Request could not be processed: `)
	regexpOpenIDConnectDiscoveryIssuerError = regexp.MustCompile(`^OpenID Connect 1\.0 Discovery Request could not be processed: `)
)

var regexpUnableToFindJSONWebKey = regexp.MustCompile(`Unable to find JSON web key with kid '', use 'sig', and alg 'RS256' in JSON Web Key Set$`)

var (
	regexpAccessRequestIssuerError = regexp.MustCompile(`^Access Request could not be processed: `)
	regexpAccessRequestFailed      = regexp.MustCompile(`^Access Request failed with error: `)
)

var (
	regexpIntrospectionRequestIssuerError = regexp.MustCompile(`^Introspection Request with id '[^']+' could not be processed: `)
	regexpIntrospectionRequestFailed      = regexp.MustCompile(`^Introspection Request with id '[^']+' failed with error$`)
)

var regexpErrorInvalidClient = regexp.MustCompile(`^The request could not be authorized\.`)

var regexpIntrospectionIssuerMismatch = regexp.MustCompile(`The original request and the introspection request occurred at endpoints where the origin or effective issuer did not match\.`)

var (
	regexpRevocationRequestIssuerError = regexp.MustCompile(`^Revocation Request with id '[^']+' could not be processed: `)
	regexpRevocationRequestFailed      = regexp.MustCompile(`^Revocation Request with id '[^']+' failed with error: `)
)

var (
	regexpPARIssuerError = regexp.MustCompile(`^Pushed Authorization Request could not be processed: `)
	regexpPARFailed      = regexp.MustCompile(`^Pushed Authorization Request failed with error: `)
)

var (
	regexpUserInfoIssuerError    = regexp.MustCompile(`^User Info Request with id '[^']+' could not be processed: `)
	regexpUserInfoFailed         = regexp.MustCompile(`^User Info Request with id '[^']+' failed with error: `)
	regexpUserInfoIssuerMismatch = regexp.MustCompile(`^User Info Request with id '[^']+' could not be processed: .*issuer did not match`)
)

var (
	regexpDeviceAuthorizationIssuerError         = regexp.MustCompile(`^Device Authorization Request could not be processed: `)
	regexpDeviceAuthorizationUserFlowIssuerError = regexp.MustCompile(`^Device Authorization Request during the User Authorization Flow could not be processed: `)
)

var (
	regexpAnyError                        = regexp.MustCompile(`.`)
	regexpErrorClientAuthenticationFailed = regexp.MustCompile(`^Client authentication failed`)
)

var (
	regexpAuthorizationIssuerError          = regexp.MustCompile(`^Authorization Request could not be processed: `)
	regexpAuthorizationRequestIDIssuerError = regexp.MustCompile(`^Authorization Request with id '[^']*' could not be processed: `)
	regexpAuthorizationFailed               = regexp.MustCompile(`^Authorization Request .*(failed|could not be processed)`)
	regexpAuthorizationPromptNoneAnonymous  = regexp.MustCompile(`the 'prompt' type of 'none' was requested but the user is not logged in$`)
	regexpAuthorizationMultipartError       = regexp.MustCompile(`^Authorization Request with id '[^']*' had an error parsing a multipart form\.$`)
)

var regexpConsentMalformedChallengeID = regexp.MustCompile(`error occurred parsing the consent id [(]challenge[)] 'not-a-uuid'`)

var regexpAccessRequestIssuerMismatch = regexp.MustCompile(`^Access Request with id '[^']+' could not be processed: `)

var regexpUserInfoNotAccessToken = regexp.MustCompile(`bearer authorization failed as the token is not an Access Token$`)

var regexpUnsafeTargetURI = regexp.MustCompile(`^Error occurred determining if the URI '.*' is safe to redirect to as it could not be parsed$`)

var regexpDeviceAuthorizationClaimsParseError = regexp.MustCompile(`error occurred parsing the claims parameter$`)

var regexpAccessResponseCreationFailed = regexp.MustCompile(`^Access Response for Request with id '[^']+' failed to be created with error: `)

type testOIDCSessionStore struct {
	FailSaves bool
}

func mustGetTestOIDCKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	if testOIDCKeyRSA2048 == nil {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		testOIDCKeyRSA2048 = key
	}

	return testOIDCKeyRSA2048
}

func newTestOIDCConfig(t *testing.T) *schema.IdentityProvidersOpenIDConnect {
	t.Helper()

	return &schema.IdentityProvidersOpenIDConnect{
		HMACSecret: "abcdefghijklmnopqrstuvwxyz1234567890",
		JSONWebKeys: []schema.JWK{
			{
				KeyID:            testOIDCKeyID,
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgRSAUsingSHA256,
				Key:              mustGetTestOIDCKey(t),
				CertificateChain: schema.X509CertificateChain{},
			},
		},
		ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
			testOIDCClaimsPolicyMerged: {IDTokenAudienceMode: oidc.IDTokenAudienceModeExperimentalMerged},
		},
	}
}

func setupTestOIDCProvider(t *testing.T, mock *mocks.MockAutheliaCtx, config *schema.IdentityProvidersOpenIDConnect) {
	t.Helper()

	if config == nil {
		config = newTestOIDCConfig(t)
	}

	mock.Ctx.Configuration.IdentityProviders.OIDC = config
	mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(&mock.Ctx.Configuration, mock.StorageMock, mock.Ctx.Providers.Templates)
	mock.Ctx.Providers.UserAttributeResolver = expression.NewUserAttributes(&mock.Ctx.Configuration)

	require.NotNil(t, mock.Ctx.Providers.OpenIDConnect)
}

func clearForwardedHeaders(mock *mocks.MockAutheliaCtx) {
	mock.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)
	mock.Ctx.Request.Header.Del("X-Forwarded-Host")
}

func newTestOIDCConfigNoRSAKey(t *testing.T) *schema.IdentityProvidersOpenIDConnect {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	return &schema.IdentityProvidersOpenIDConnect{
		HMACSecret: "abcdefghijklmnopqrstuvwxyz1234567890",
		JSONWebKeys: []schema.JWK{
			{
				KeyID:            "ecdsa-default",
				Use:              oidc.KeyUseSignature,
				Algorithm:        oidc.SigningAlgECDSAUsingP256AndSHA256,
				Key:              key,
				CertificateChain: schema.X509CertificateChain{},
			},
		},
	}
}

func newTestOAuth2Request(t *testing.T, method, uri string, values url.Values) (rw *httptest.ResponseRecorder, r *http.Request) {
	t.Helper()

	var body io.Reader

	if values != nil && method == fasthttp.MethodPost {
		body = strings.NewReader(values.Encode())
	}

	r = httptest.NewRequest(method, uri, body)

	if values != nil {
		if method == fasthttp.MethodPost {
			r.Header.Set(fasthttp.HeaderContentType, "application/x-www-form-urlencoded")
		} else {
			r.URL.RawQuery = values.Encode()
		}
	}

	return httptest.NewRecorder(), r
}

func getTestOAuth2ErrorResponse(t *testing.T, rw *httptest.ResponseRecorder) (response map[string]any) {
	t.Helper()

	response = map[string]any{}

	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &response))

	return response
}

func mustDecodeTestSecret(t *testing.T, value string) *schema.PasswordDigest {
	t.Helper()

	digest, err := schema.DecodePasswordDigest(value)

	require.NoError(t, err)

	return digest
}

func newTestOIDCClientCredentialsClient(t *testing.T) schema.IdentityProvidersOpenIDConnectClient {
	t.Helper()

	return schema.IdentityProvidersOpenIDConnectClient{
		ID:                      testOIDCClientCredentialsID,
		Secret:                  mustDecodeTestSecret(t, testOIDCClientSecretDigest),
		AuthorizationPolicy:     "one_factor",
		GrantTypes:              []string{oidc.GrantTypeClientCredentials},
		Scopes:                  []string{testOIDCScopeBearerAuthz},
		Audience:                []string{"https://app.example.com"},
		TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
	}
}

func setupTestOIDCSessionStore(t *testing.T, mock *mocks.MockAutheliaCtx) (store *testOIDCSessionStore) {
	t.Helper()

	store = &testOIDCSessionStore{}

	sessions := map[storage.OAuth2SessionType]map[string]model.OAuth2Session{}

	mock.StorageMock.EXPECT().
		BeginTX(gomock.Any()).
		AnyTimes().
		DoAndReturn(func(ctx context.Context) (context.Context, error) {
			return ctx, nil
		})

	mock.StorageMock.EXPECT().
		Commit(gomock.Any()).
		AnyTimes().
		Return(nil)

	mock.StorageMock.EXPECT().
		Rollback(gomock.Any()).
		AnyTimes().
		Return(nil)

	mock.StorageMock.EXPECT().
		SaveOAuth2Session(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, sessionType storage.OAuth2SessionType, session model.OAuth2Session) (err error) {
			if store.FailSaves {
				return fmt.Errorf("error in db")
			}

			if _, ok := sessions[sessionType]; !ok {
				sessions[sessionType] = map[string]model.OAuth2Session{}
			}

			sessions[sessionType][session.Signature] = session

			return nil
		})

	mock.StorageMock.EXPECT().
		LoadOAuth2Session(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, sessionType storage.OAuth2SessionType, signature string) (session *model.OAuth2Session, err error) {
			if s, ok := sessions[sessionType][signature]; ok {
				return &s, nil
			}

			return nil, sql.ErrNoRows
		})

	mock.StorageMock.EXPECT().
		RevokeOAuth2Session(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, sessionType storage.OAuth2SessionType, signature string) (err error) {
			if s, ok := sessions[sessionType][signature]; ok {
				s.Active = false
				sessions[sessionType][signature] = s
			}

			return nil
		})

	mock.StorageMock.EXPECT().
		DeactivateOAuth2Session(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, sessionType storage.OAuth2SessionType, signature string) (err error) {
			if s, ok := sessions[sessionType][signature]; ok {
				s.Active = false
				sessions[sessionType][signature] = s
			}

			return nil
		})

	mock.StorageMock.EXPECT().
		RevokeOAuth2SessionByRequestID(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, sessionType storage.OAuth2SessionType, requestID string) (err error) {
			for signature, s := range sessions[sessionType] {
				if s.RequestID == requestID {
					s.Active, s.Revoked = false, true
					sessions[sessionType][signature] = s
				}
			}

			return nil
		})

	mock.StorageMock.EXPECT().
		DeactivateOAuth2SessionByRequestID(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, sessionType storage.OAuth2SessionType, requestID string) (err error) {
			for signature, s := range sessions[sessionType] {
				if s.RequestID == requestID {
					s.Active = false
					sessions[sessionType][signature] = s
				}
			}

			return nil
		})

	return store
}

func mustGetTestOIDCClientCredentialsToken(t *testing.T, mock *mocks.MockAutheliaCtx) (token string) {
	t.Helper()

	rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCTokenEndpoint, url.Values{
		testOIDCFormParameterGrantType:    []string{oidc.GrantTypeClientCredentials},
		oidc.FormParameterClientID:        []string{testOIDCClientCredentialsID},
		testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		oidc.FormParameterScope:           []string{testOIDCScopeBearerAuthz},
		testOIDCFormParameterAudience:     []string{"https://app.example.com"},
	})

	OAuth2TokenPOST(mock.Ctx, rw, r)

	require.Equal(t, http.StatusOK, rw.Code)

	response := getTestOAuth2ErrorResponse(t, rw)

	token, ok := response["access_token"].(string)

	require.True(t, ok)
	require.NotEmpty(t, token)

	return token
}

func setupTestOIDCConsentStore(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	consents := map[uuid.UUID]*model.OAuth2ConsentSession{}

	id := 0

	mock.StorageMock.EXPECT().
		SaveOAuth2ConsentSession(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, consent *model.OAuth2ConsentSession) (err error) {
			id++

			stored := *consent
			stored.ID = id

			consents[consent.ChallengeID] = &stored

			return nil
		})

	mock.StorageMock.EXPECT().
		LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, challengeID uuid.UUID) (consent *model.OAuth2ConsentSession, err error) {
			if stored, ok := consents[challengeID]; ok {
				value := *stored

				return &value, nil
			}

			return nil, sql.ErrNoRows
		})

	mock.StorageMock.EXPECT().
		SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, consent *model.OAuth2ConsentSession, rejection bool) (err error) {
			stored := *consent
			stored.Authorized = !rejection

			consents[consent.ChallengeID] = &stored

			return nil
		})

	mock.StorageMock.EXPECT().
		SaveOAuth2ConsentSessionGranted(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, id int) (err error) {
			for _, consent := range consents {
				if consent.ID == id {
					consent.Granted = true
				}
			}

			return nil
		})
}

func setupTestOIDCSubjectStore(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	identifiers := map[string]model.UserOpaqueIdentifier{}

	mock.StorageMock.EXPECT().
		LoadUserOpaqueIdentifierBySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, service, sectorID, username string) (identifier *model.UserOpaqueIdentifier, err error) {
			key := service + ":" + sectorID + ":" + username

			if value, ok := identifiers[key]; ok {
				return &value, nil
			}

			value := model.UserOpaqueIdentifier{
				Service:    service,
				SectorID:   sectorID,
				Username:   username,
				Identifier: uuid.Must(uuid.NewRandom()),
			}

			identifiers[key] = value

			return &value, nil
		})

	mock.StorageMock.EXPECT().
		SaveUserOpaqueIdentifier(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil)

	mock.StorageMock.EXPECT().
		LoadUserOpaqueIdentifier(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, identifier uuid.UUID) (subject *model.UserOpaqueIdentifier, err error) {
			for _, value := range identifiers {
				if value.Identifier == identifier {
					v := value

					return &v, nil
				}
			}

			return nil, sql.ErrNoRows
		})
}

func newTestOIDCUserSession(level int) session.UserSession {
	now := time.Now().Unix()

	us := session.UserSession{
		CookieDomain:             exampleDotCom,
		Username:                 testUsername,
		DisplayName:              testDisplayName,
		Emails:                   []string{testEmail},
		Groups:                   []string{"dev"},
		AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{UsernameAndPassword: true},
		LastActivity:             now,
	}

	if level >= 1 {
		us.FirstFactorAuthnTimestamp = now
	}

	if level >= 2 {
		us.SecondFactorAuthnTimestamp = now
		us.AuthenticationMethodRefs.TOTP = true
	}

	return us
}

func setupTestOIDCUserDetails(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	details := &authentication.UserDetails{
		Username:    testUsername,
		DisplayName: testDisplayName,
		Emails:      []string{testEmail},
		Groups:      []string{"dev"},
	}

	mock.UserProviderMock.EXPECT().
		GetDetails(testUsername).
		AnyTimes().
		Return(details, nil)

	mock.UserProviderMock.EXPECT().
		GetDetailsExtended(testUsername).
		AnyTimes().
		Return(&authentication.UserDetailsExtended{UserDetails: details}, nil)
}

func mustGetTestOIDCSubject(t *testing.T, mock *mocks.MockAutheliaCtx, clientID string) (subject uuid.UUID) {
	t.Helper()

	client, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, clientID)

	require.NoError(t, err)

	subject, err = mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, client.GetSectorIdentifierURI(), testUsername)

	require.NoError(t, err)

	return subject
}

func setupTestOIDCDeviceCodeStore(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	sessions := map[string]*model.OAuth2DeviceCodeSession{}

	save := func(_ context.Context, session *model.OAuth2DeviceCodeSession) (err error) {
		stored := *session

		if stored.ID == 0 {
			stored.ID = len(sessions) + 1
		}

		sessions[stored.UserCodeSignature] = &stored

		return nil
	}

	mock.StorageMock.EXPECT().
		SaveOAuth2DeviceCodeSession(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(save)

	mock.StorageMock.EXPECT().
		UpdateOAuth2DeviceCodeSession(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(save)

	mock.StorageMock.EXPECT().
		UpdateOAuth2DeviceCodeSessionData(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(save)

	mock.StorageMock.EXPECT().
		LoadOAuth2DeviceCodeSessionByUserCode(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, signature string) (session *model.OAuth2DeviceCodeSession, err error) {
			if stored, ok := sessions[signature]; ok {
				value := *stored

				return &value, nil
			}

			return nil, sql.ErrNoRows
		})
}

func mustGetTestOIDCUserCode(t *testing.T, mock *mocks.MockAutheliaCtx) (userCode string) {
	t.Helper()

	rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
		oidc.FormParameterClientID:        []string{testOIDCDeviceCodeID},
		testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
		oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
	})

	OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

	require.Equal(t, http.StatusOK, rw.Code)

	response := getTestOAuth2ErrorResponse(t, rw)

	userCode, ok := response[oidc.FormParameterUserCode].(string)

	require.True(t, ok)
	require.NotEmpty(t, userCode)

	return userCode
}

func mustGetTestOIDCAuthorizationCode(t *testing.T, mock *mocks.MockAutheliaCtx, values url.Values) (code string) {
	t.Helper()

	rw, r := newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, values)

	OAuth2AuthorizationGET(mock.Ctx, rw, r)

	require.Equal(t, http.StatusSeeOther, rw.Code)

	location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

	require.NoError(t, err)
	require.Empty(t, location.Query().Get("error"), location.Query().Get("error_description"))

	code = location.Query().Get(testOIDCFormParameterCode)

	require.NotEmpty(t, code)

	return code
}

func setupTestOIDCPARStore(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	sessions := map[string]model.OAuth2PushedAuthorizationSession{}

	save := func(_ context.Context, par model.OAuth2PushedAuthorizationSession) (err error) {
		sessions[par.Signature] = par

		return nil
	}

	mock.StorageMock.EXPECT().
		SaveOAuth2PushedAuthorizationSession(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(save)

	mock.StorageMock.EXPECT().
		UpdateOAuth2PushedAuthorizationSession(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(save)

	mock.StorageMock.EXPECT().
		LoadOAuth2PushedAuthorizationSession(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, signature string) (par *model.OAuth2PushedAuthorizationSession, err error) {
			if value, ok := sessions[signature]; ok {
				return &value, nil
			}

			return nil, sql.ErrNoRows
		})

	mock.StorageMock.EXPECT().
		RevokeOAuth2PushedAuthorizationSession(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, signature string) (err error) {
			if value, ok := sessions[signature]; ok {
				value.Revoked = true
				sessions[signature] = value
			}

			return nil
		})
}

func newTestOIDCPreConfiguredClient(t *testing.T) schema.IdentityProvidersOpenIDConnectClient {
	t.Helper()

	client := newTestOIDCAuthorizationCodeClient(t)

	client.ConsentMode = "pre-configured"
	client.ConsentPreConfiguredDuration = &testOIDCPreConfiguredDuration

	return client
}
