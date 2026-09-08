package handlers

import (
	"database/sql"
	"fmt"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestHandleOAuth2AuthorizationConsentModePreConfigured(t *testing.T) {
	testCases := []struct {
		name      string
		consentID string
		handled   bool
		setup     func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expect    func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			name:      "ShouldDispatchToWithoutIDWhenNoConsentID",
			consentID: "",
			handled:   true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock)

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(fmt.Errorf("error in db"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`could not be processed: error occurred saving consent: error in db$`), nil)
			},
		},
		{
			name:      "ShouldHandleMalformedConsentID",
			consentID: "not-a-uuid",
			handled:   true,
			setup:     nil,
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`error occurred parsing the consent id \(challenge\) 'not-a-uuid'`), nil)
			},
		},
		{
			name:      "ShouldDispatchToWithIDWhenConsentIDPresent",
			consentID: preConfChallenge.String(),
			handled:   true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(preConfChallenge)).
					Return(nil, fmt.Errorf("error in db"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`error occurred while loading session: error in db`), nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			setupTestOIDCProvider(t, mock, nil)

			if tc.consentID != "" {
				mock.Ctx.Request.SetRequestURI("https://auth.example.com/api/oidc/authorization?consent_id=" + tc.consentID)
			}

			rw := httptest.NewRecorder()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent, handled := handleOAuth2AuthorizationConsentModePreConfigured(mock.Ctx, preConfIssuer, preConfClient, newPreConfUserSession(), preConfSubject, rw, httptest.NewRequest("GET", "https://auth.example.com", nil), newPreConfRequester(nil))

			assert.Equal(t, tc.handled, handled)
			assert.Nil(t, consent)

			if tc.expect != nil {
				tc.expect(t, mock)
			}
		})
	}
}

func TestHandleOAuth2AuthorizationConsentModePreConfiguredWithoutID(t *testing.T) {
	testCases := []struct {
		name       string
		form       url.Values
		expected   *model.OAuth2ConsentSession
		anyConsent bool
		handled    bool
		setup      func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expect     func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder)
	}{
		{
			name:     "ShouldHandlePreConfigurationLookupError",
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentPreConfigurations(gomock.Eq(mock.Ctx), gomock.Eq(testValue), gomock.Eq(preConfSubject), gomock.Any()).
					Return(nil, fmt.Errorf("error in db"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`had error looking up pre-configured consent sessions: error loading rows: error in db$`), nil)
			},
		},
		{
			name:     "ShouldHandleMalformedClaimsRequest",
			form:     url.Values{oidc.FormParameterClaims: []string{"not json"}},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`had error looking up pre-configured consent sessions: error parsing claim requests`), nil)
			},
		},
		{
			name:     "ShouldDenyPromptNoneWithoutPreConfiguration",
			form:     url.Values{oidc.FormParameterPrompt: []string{oidc.PromptNone}},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`the 'prompt' type of 'none' was requested but client is configured to require consent or pre-configured consent and the pre-configured consent was absent$`), nil)
			},
		},
		{
			name:       "ShouldGenerateConsentWithoutMatchingPreConfiguration",
			expected:   nil,
			anyConsent: true,
			handled:    true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock)

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(nil)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				assert.Equal(t, 302, rw.Code)
				assert.Contains(t, rw.Header().Get("Location"), "flow=openid_connect")
			},
		},
		{
			name:       "ShouldSkipRevokedPreConfiguration",
			expected:   nil,
			anyConsent: true,
			handled:    true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				config := newPreConfig(mock)
				config.Revoked = true

				expectPreConfigRows(t, mock, config)

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(nil)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				assert.Equal(t, 302, rw.Code)
			},
		},
		{
			name:       "ShouldSkipPreConfigurationWithDifferentGrants",
			expected:   nil,
			anyConsent: true,
			handled:    true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				config := newPreConfig(mock)
				config.Scopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeProfile}

				expectPreConfigRows(t, mock, config)

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(nil)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				assert.Equal(t, 302, rw.Code)
			},
		},
		{
			name:       "ShouldSkipPreConfigurationWithDifferentClaimsSignature",
			expected:   nil,
			anyConsent: true,
			handled:    true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				config := newPreConfig(mock)
				config.SignatureClaims = sql.NullString{String: "a-different-signature", Valid: true}

				expectPreConfigRows(t, mock, config)

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(nil)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				assert.Equal(t, 302, rw.Code)
			},
		},
		{
			name:     "ShouldHandleSaveConsentSessionError",
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock, newPreConfig(mock))

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(fmt.Errorf("error in db"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`error occurred saving consent session: error in db$`), nil)
			},
		},
		{
			name:     "ShouldHandleLoadConsentSessionError",
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock, newPreConfig(mock))

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(nil)

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(nil, fmt.Errorf("error in db"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`error occurred saving consent session: error in db$`), nil)
			},
		},
		{
			name:     "ShouldHandleSaveConsentSessionResponseError",
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock, newPreConfig(mock))

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(nil)

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(&model.OAuth2ConsentSession{ChallengeID: preConfChallenge, ClientID: testValue, RequestedAt: time.Unix(1000000, 0)}, nil)

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSessionResponse(gomock.Eq(mock.Ctx), gomock.Any(), gomock.Eq(false)).
					Return(fmt.Errorf("error in db"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`error occurred saving consent session response: error in db$`), nil)
			},
		},
		{
			name: "ShouldGrantWithMatchingPreConfiguration",
			expected: &model.OAuth2ConsentSession{
				ChallengeID:      preConfChallenge,
				ClientID:         testValue,
				Granted:          false,
				Authorized:       false,
				GrantedScopes:    model.StringSlicePipeDelimited{oidc.ScopeOpenID},
				PreConfiguration: sql.NullInt64{Int64: 10, Valid: true},
			},
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock, newPreConfig(mock))

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(nil)

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(&model.OAuth2ConsentSession{ChallengeID: preConfChallenge, ClientID: testValue, RequestedAt: time.Unix(1000000, 0), RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID}}, nil)

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSessionResponse(gomock.Eq(mock.Ctx), gomock.Any(), gomock.Eq(false)).
					Return(nil)
			},
			expect: nil,
		},
		{
			name:     "ShouldRedirectForReauthenticationWhenPromptLogin",
			form:     url.Values{oidc.FormParameterPrompt: []string{oidc.PromptLogin}},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock, newPreConfig(mock))

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(nil)

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Any()).
					Return(&model.OAuth2ConsentSession{ChallengeID: preConfChallenge, ClientID: testValue, RequestedAt: time.Unix(2000000, 0)}, nil)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				assert.Equal(t, 302, rw.Code)
				assert.Contains(t, rw.Header().Get("Location"), oidc.FrontendEndpointPathConsentDecision)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			setupTestOIDCProvider(t, mock, nil)

			rw := httptest.NewRecorder()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent, handled := handleOAuth2AuthorizationConsentModePreConfiguredWithoutID(mock.Ctx, preConfIssuer, preConfClient, newPreConfUserSession(), preConfSubject, rw, httptest.NewRequest("GET", "https://auth.example.com", nil), newPreConfRequester(tc.form))

			assert.Equal(t, tc.handled, handled)

			switch {
			case tc.anyConsent:
				break
			case tc.expected == nil:
				assert.Nil(t, consent)
			default:
				require.NotNil(t, consent)

				assert.Equal(t, tc.expected.ChallengeID, consent.ChallengeID)
				assert.Equal(t, tc.expected.ClientID, consent.ClientID)
				assert.Equal(t, tc.expected.Authorized, consent.Authorized)
				assert.Equal(t, tc.expected.Granted, consent.Granted)
				assert.Equal(t, tc.expected.GrantedScopes, consent.GrantedScopes)
				assert.Equal(t, tc.expected.PreConfiguration, consent.PreConfiguration)
				assert.True(t, consent.RespondedAt.Valid)
			}

			if tc.expect != nil {
				tc.expect(t, mock, rw)
			}
		})
	}
}

func TestHandleOAuth2AuthorizationConsentModePreConfiguredWithIDExtra(t *testing.T) {
	testCases := []struct {
		name       string
		form       url.Values
		consent    *model.OAuth2ConsentSession
		expected   *model.OAuth2ConsentSession
		anyConsent bool
		handled    bool
		setup      func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expect     func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder)
	}{
		{
			name: "ShouldGrantWithMatchingPreConfiguration",
			consent: &model.OAuth2ConsentSession{
				ID:              1,
				ChallengeID:     preConfChallenge,
				ClientID:        testValue,
				RequestedAt:     time.Unix(1000000, 0),
				ExpiresAt:       time.Unix(9000000000, 0),
				RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
			},
			expected: &model.OAuth2ConsentSession{
				ChallengeID:      preConfChallenge,
				ClientID:         testValue,
				GrantedScopes:    model.StringSlicePipeDelimited{oidc.ScopeOpenID},
				PreConfiguration: sql.NullInt64{Int64: 10, Valid: true},
			},
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock, newPreConfig(mock))

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSessionResponse(gomock.Eq(mock.Ctx), gomock.Any(), gomock.Eq(false)).
					Return(nil)
			},
		},
		{
			name: "ShouldHandleSaveConsentSessionResponseError",
			consent: &model.OAuth2ConsentSession{
				ID:              1,
				ChallengeID:     preConfChallenge,
				ClientID:        testValue,
				RequestedAt:     time.Unix(1000000, 0),
				ExpiresAt:       time.Unix(9000000000, 0),
				RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock, newPreConfig(mock))

				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSessionResponse(gomock.Eq(mock.Ctx), gomock.Any(), gomock.Eq(false)).
					Return(fmt.Errorf("error in db"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`error occurred saving consent session response: error in db$`), nil)
			},
		},
		{
			name: "ShouldHandlePreConfigurationLookupError",
			consent: &model.OAuth2ConsentSession{
				ID:              1,
				ChallengeID:     preConfChallenge,
				ClientID:        testValue,
				RequestedAt:     time.Unix(1000000, 0),
				ExpiresAt:       time.Unix(9000000000, 0),
				RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentPreConfigurations(gomock.Eq(mock.Ctx), gomock.Eq(testValue), gomock.Eq(preConfSubject), gomock.Any()).
					Return(nil, fmt.Errorf("error in db"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`had error looking up pre-configured consent sessions: error loading rows: error in db$`), nil)
			},
		},
		{
			name: "ShouldDenyExplicitlyRejectedConsentSession",
			consent: &model.OAuth2ConsentSession{
				ID:              1,
				ChallengeID:     preConfChallenge,
				ClientID:        testValue,
				RequestedAt:     time.Unix(1000000, 0),
				ExpiresAt:       time.Unix(9000000000, 0),
				RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
				RespondedAt:     sql.NullTime{Time: time.Unix(1000001, 0), Valid: true},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`the user explicitly rejected this consent session$`), nil)
			},
		},
		{
			name: "ShouldRedirectUnrespondedConsentSessionToConsentPage",
			consent: &model.OAuth2ConsentSession{
				ID:              1,
				ChallengeID:     preConfChallenge,
				ClientID:        testValue,
				RequestedAt:     time.Unix(1000000, 0),
				ExpiresAt:       time.Unix(9000000000, 0),
				RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				assert.Equal(t, 302, rw.Code)
				assert.Contains(t, rw.Header().Get("Location"), preConfChallenge.String())
			},
		},
		{
			name: "ShouldDenyPromptNoneOnAuthorizedSessionWithoutPreConfiguration",
			form: url.Values{oidc.FormParameterPrompt: []string{oidc.PromptNone}},
			consent: &model.OAuth2ConsentSession{
				ID:              1,
				ChallengeID:     preConfChallenge,
				ClientID:        testValue,
				RequestedAt:     time.Unix(1000000, 0),
				ExpiresAt:       time.Unix(9000000000, 0),
				RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
				Authorized:      true,
				RespondedAt:     sql.NullTime{Time: time.Unix(1000001, 0), Valid: true},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`the 'prompt' type of 'none' was requested but client is configured to require consent or pre-configured consent and the pre-configured consent was absent$`), nil)
			},
		},
		{
			name: "ShouldReturnAuthorizedSessionWithoutPreConfiguration",
			consent: &model.OAuth2ConsentSession{
				ID:              1,
				ChallengeID:     preConfChallenge,
				ClientID:        testValue,
				RequestedAt:     time.Unix(1000000, 0),
				ExpiresAt:       time.Unix(9000000000, 0),
				RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
				Authorized:      true,
				RespondedAt:     sql.NullTime{Time: time.Unix(1000001, 0), Valid: true},
			},
			anyConsent: true,
			handled:    false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				expectPreConfigRows(t, mock)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			setupTestOIDCProvider(t, mock, nil)

			mock.StorageMock.EXPECT().
				LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(preConfChallenge)).
				Return(tc.consent, nil)

			rw := httptest.NewRecorder()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent, handled := handleOAuth2AuthorizationConsentModePreConfiguredWithID(mock.Ctx, preConfIssuer, preConfClient, newPreConfUserSession(), preConfSubject, preConfChallenge, rw, httptest.NewRequest("GET", "https://auth.example.com", nil), newPreConfRequester(tc.form))

			assert.Equal(t, tc.handled, handled)

			switch {
			case tc.anyConsent:
				assert.NotNil(t, consent)
			case tc.expected == nil:
				assert.Nil(t, consent)
			default:
				require.NotNil(t, consent)

				assert.Equal(t, tc.expected.ChallengeID, consent.ChallengeID)
				assert.Equal(t, tc.expected.ClientID, consent.ClientID)
				assert.Equal(t, tc.expected.GrantedScopes, consent.GrantedScopes)
				assert.Equal(t, tc.expected.PreConfiguration, consent.PreConfiguration)
				assert.Equal(t, uuid.NullUUID{UUID: preConfSubject, Valid: true}, consent.Subject)
				assert.True(t, consent.RespondedAt.Valid)
			}

			if tc.expect != nil {
				tc.expect(t, mock, rw)
			}
		})
	}
}

func TestHandleOAuth2AuthorizationConsentModePreConfiguredMisc(t *testing.T) {
	t.Run("ShouldRedirectForReauthenticationWhenPromptLoginWithID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(preConfChallenge)).
			Return(&model.OAuth2ConsentSession{
				ID:              1,
				ChallengeID:     preConfChallenge,
				ClientID:        testValue,
				RequestedAt:     time.Unix(9000000000, 0),
				ExpiresAt:       time.Unix(9000000000, 0),
				RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
			}, nil)

		rw := httptest.NewRecorder()

		form := url.Values{oidc.FormParameterPrompt: []string{oidc.PromptLogin}}

		consent, handled := handleOAuth2AuthorizationConsentModePreConfiguredWithID(mock.Ctx, preConfIssuer, preConfClient, newPreConfUserSession(), preConfSubject, preConfChallenge, rw, httptest.NewRequest("GET", "https://auth.example.com", nil), newPreConfRequester(form))

		assert.True(t, handled)
		assert.Nil(t, consent)

		assert.Equal(t, 302, rw.Code)
		assert.Contains(t, rw.Header().Get("Location"), oidc.FrontendEndpointPathConsentDecision)
	})

	t.Run("ShouldReturnNoPreConfigurationWhenRowsAreClosed", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentPreConfigurations(gomock.Eq(mock.Ctx), gomock.Eq(testValue), gomock.Eq(preConfSubject), gomock.Any()).
			DoAndReturn(func(_ any, _ string, _ uuid.UUID, _ time.Time) (*storage.ConsentPreConfigRows, error) {
				rows, closer := mocks.NewConsentPreConfigRows(t, newPreConfig(mock))

				closer()

				return rows, nil
			})

		config, err := handleOAuth2AuthorizationConsentModePreConfiguredGetPreConfig(mock.Ctx, preConfClient, preConfSubject, newPreConfRequester(nil))

		assert.Nil(t, config)
		assert.Nil(t, err)
	})
}

var (
	preConfClient = &oidc.RegisteredClient{
		ID:            testValue,
		ConsentPolicy: oidc.ClientConsentPolicy{Mode: oidc.ClientConsentModePreConfigured},
	}

	preConfChallenge = uuid.MustParse("11303e1f-f8af-436a-9a72-c7361bfc9f37")
	preConfSubject   = uuid.MustParse("e79b6494-8852-4439-860c-159f2cba83dc")

	preConfIssuer = &url.URL{Scheme: "https", Host: "auth.example.com"}
)

func expectPreConfigRows(t *testing.T, mock *mocks.MockAutheliaCtx, configs ...model.OAuth2ConsentPreConfig) {
	t.Helper()

	mock.StorageMock.EXPECT().
		LoadOAuth2ConsentPreConfigurations(gomock.Eq(mock.Ctx), gomock.Eq(testValue), gomock.Eq(preConfSubject), gomock.Any()).
		DoAndReturn(func(_ any, _ string, _ uuid.UUID, _ time.Time) (rows *storage.ConsentPreConfigRows, err error) {
			rows, closer := mocks.NewConsentPreConfigRows(t, configs...)

			t.Cleanup(closer)

			return rows, nil
		})
}

func newPreConfRequester(form url.Values) *oauthelia2.AuthorizeRequest {
	if form == nil {
		form = url.Values{}
	}

	return &oauthelia2.AuthorizeRequest{
		Request: oauthelia2.Request{
			Client:         preConfClient,
			RequestedAt:    time.Unix(1000000, 0),
			RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID},
			Form:           form,
		},
	}
}

func newPreConfig(mock *mocks.MockAutheliaCtx) model.OAuth2ConsentPreConfig {
	return model.OAuth2ConsentPreConfig{
		ID:        10,
		ClientID:  testValue,
		Subject:   preConfSubject,
		CreatedAt: mock.Ctx.Providers.Clock.Now().Add(-time.Hour),
		ExpiresAt: sql.NullTime{Time: mock.Ctx.Providers.Clock.Now().Add(time.Hour), Valid: true},
		Scopes:    model.StringSlicePipeDelimited{oidc.ScopeOpenID},
		Audience:  model.StringSlicePipeDelimited{},
	}
}

func newPreConfUserSession() session.UserSession {
	return session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000}
}
