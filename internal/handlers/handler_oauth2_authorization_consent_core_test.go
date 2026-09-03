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

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestHandleOAuth2AuthorizationConsentGenerate(t *testing.T) {
	mustParseURI := func(t *testing.T, in string) *url.URL {
		result, err := url.Parse(in)
		require.NoError(t, err)

		return result
	}

	clientTest := &oidc.RegisteredClient{
		ID: testValue,
	}

	sub := uuid.MustParse("e79b6494-8852-4439-860c-159f2cba83dc")

	testCases := []struct {
		name        string
		issuer      *url.URL
		client      oidc.Client
		userSession session.UserSession
		subject     uuid.UUID
		requester   oauthelia2.AuthorizeRequester
		expected    *model.OAuth2ConsentSession
		handled     bool
		setup       func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expect      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			name:        "ShouldHandleDBError",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
				},
			},
			expected: (*model.OAuth2ConsentSession)(nil),
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(fmt.Errorf("invalid")),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred saving consent: invalid$`), nil)
			},
		},
		{
			name:        "ShouldHandleQueryArgsError",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
				},
			},
			expected: (*model.OAuth2ConsentSession)(nil),
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.QueryArgs().Add(queryArgConsentID, "abc")
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred generating consent: consent id value was present when it should be absent$`), nil)
			},
		},
		{
			name:        "ShouldHandlePromptLoginNotRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{
				ClientID:    "test",
				Subject:     uuid.NullUUID{UUID: sub, Valid: true},
				Form:        "prompt=login",
				RequestedAt: time.Unix(1000000, 0),
			},
			handled: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleMaxAgeNotRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterMaximumAge: []string{"10"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{
				ClientID:    "test",
				Subject:     uuid.NullUUID{UUID: sub, Valid: true},
				Form:        "max_age=10",
				RequestedAt: time.Unix(1000000, 0),
			},
			handled: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandlePromptLoginRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleMaxAgeRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 0, SecondFactorAuthnTimestamp: 0},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterMaximumAge: []string{"10"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Clock.Set(time.Unix(1000000, 0))

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID: "abc",
							},
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, mock.StorageMock, mock.Ctx.Providers.Templates)

			rw := httptest.NewRecorder()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent, handled := handleOAuth2AuthorizationConsentGenerate(mock.Ctx, tc.issuer, tc.client, tc.userSession, tc.subject, rw, httptest.NewRequest("GET", "https://example.com", nil), tc.requester)

			assert.Equal(t, tc.handled, handled)

			if tc.expected == nil {
				assert.Nil(t, consent)
			} else {
				require.NotNil(t, consent)
				assert.Equal(t, tc.expected.ClientID, consent.ClientID)
				assert.Equal(t, tc.expected.Subject, consent.Subject)
				assert.Equal(t, tc.expected.Granted, consent.Granted)
				assert.Equal(t, tc.expected.Authorized, consent.Authorized)
				assert.Equal(t, tc.expected.RequestedAt, consent.RequestedAt)
				assert.Equal(t, tc.expected.RespondedAt, consent.RespondedAt)
				assert.Equal(t, tc.expected.Form, consent.Form)
				assert.Equal(t, tc.expected.RequestedScopes, consent.RequestedScopes)
				assert.Equal(t, tc.expected.RequestedAudience, consent.RequestedAudience)
				assert.Equal(t, tc.expected.GrantedScopes, consent.GrantedScopes)
				assert.Equal(t, tc.expected.GrantedAudience, consent.GrantedAudience)
				assert.Equal(t, tc.expected.PreConfiguration, consent.PreConfiguration)
			}

			if tc.expect != nil {
				tc.expect(t, mock)
			}
		})
	}
}

func TestHandleOIDCAuthorizationConsentNotAuthenticated(t *testing.T) {
	mustParseURI := func(t *testing.T, in string) *url.URL {
		result, err := url.Parse(in)
		require.NoError(t, err)

		return result
	}

	clientTest := &oidc.RegisteredClient{
		ID: testValue,
	}

	sub := uuid.MustParse("e79b6494-8852-4439-860c-159f2cba83dc")

	testCases := []struct {
		name        string
		issuer      *url.URL
		client      oidc.Client
		userSession session.UserSession
		subject     uuid.UUID
		requester   oauthelia2.AuthorizeRequester
		expected    *model.OAuth2ConsentSession
		handled     bool
		setup       func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expect      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			name:        "ShouldHandleDBError",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
				},
			},
			expected: (*model.OAuth2ConsentSession)(nil),
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(fmt.Errorf("invalid")),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred saving consent: invalid$`), nil)
			},
		},
		{
			name:        "ShouldHandlePromptLoginRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleMaxAgeRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 0, SecondFactorAuthnTimestamp: 0},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterMaximumAge: []string{"10"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Clock.Set(time.Unix(1000000, 0))

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID: "abc",
							},
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, mock.StorageMock, mock.Ctx.Providers.Templates)

			rw := httptest.NewRecorder()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent, handled := handleOAuth2AuthorizationConsentNotAuthenticated(mock.Ctx, tc.issuer, tc.client, tc.userSession, tc.subject, rw, httptest.NewRequest("GET", "https://example.com", nil), tc.requester)

			assert.Equal(t, tc.handled, handled)

			if tc.expected == nil {
				assert.Nil(t, consent)
			} else {
				require.NotNil(t, consent)
				assert.Equal(t, tc.expected.ClientID, consent.ClientID)
				assert.Equal(t, tc.expected.Subject, consent.Subject)
				assert.Equal(t, tc.expected.Granted, consent.Granted)
				assert.Equal(t, tc.expected.Authorized, consent.Authorized)
				assert.Equal(t, tc.expected.RequestedAt, consent.RequestedAt)
				assert.Equal(t, tc.expected.RespondedAt, consent.RespondedAt)
				assert.Equal(t, tc.expected.Form, consent.Form)
				assert.Equal(t, tc.expected.RequestedScopes, consent.RequestedScopes)
				assert.Equal(t, tc.expected.RequestedAudience, consent.RequestedAudience)
				assert.Equal(t, tc.expected.GrantedScopes, consent.GrantedScopes)
				assert.Equal(t, tc.expected.GrantedAudience, consent.GrantedAudience)
				assert.Equal(t, tc.expected.PreConfiguration, consent.PreConfiguration)
			}

			if tc.expect != nil {
				tc.expect(t, mock)
			}
		})
	}
}

func TestHandleOAuth2AuthorizationConsentModeImplicitWithoutID(t *testing.T) {
	mustParseURI := func(t *testing.T, in string) *url.URL {
		result, err := url.Parse(in)
		require.NoError(t, err)

		return result
	}

	clientTest := &oidc.RegisteredClient{
		ID:            testValue,
		ConsentPolicy: oidc.ClientConsentPolicy{Mode: oidc.ClientConsentModeImplicit},
	}

	sub := uuid.MustParse("e79b6494-8852-4439-860c-159f2cba83dc")

	testCases := []struct {
		name        string
		issuer      *url.URL
		client      oidc.Client
		userSession session.UserSession
		subject     uuid.UUID
		requester   oauthelia2.AuthorizeRequester
		expected    *model.OAuth2ConsentSession
		handled     bool
		setup       func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expect      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			name:        "ShouldHandleDBError",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(fmt.Errorf("invalid")),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'implicit' could not be processed: error occurred performing consent for consent session with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}': error occurred saving consent session: invalid$`), nil)
			},
		},
		{
			name:        "ShouldHandleLoadSessionFromDBError",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil, fmt.Errorf("bad")),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'implicit' could not be processed: error occurred performing consent for consent session with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}': error occurred saving consent session: bad$`), nil)
			},
		},
		{
			name:        "ShouldHandleFormPromptLoginNotRequiredErrorSave",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(&model.OAuth2ConsentSession{
							ID:          1,
							ClientID:    "test",
							Subject:     uuid.NullUUID{UUID: sub, Valid: true},
							Form:        "prompt=login",
							RequestedAt: time.Unix(1000000, 0),
						}, nil),
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSessionResponse(gomock.Eq(mock.Ctx), gomock.Any(), gomock.Eq(false)).
						Return(fmt.Errorf("bad conn")),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'implicit' could not be processed: error occurred performing consent for consent session with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}': error occurred saving consent session response: bad conn$`), nil)
			},
		},
		{
			name:        "ShouldHandleFormPromptLoginNotRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{
				ClientID:    "test",
				Subject:     uuid.NullUUID{UUID: sub, Valid: true},
				Form:        "prompt=login",
				RequestedAt: time.Unix(1000000, 0),
				RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true},
			},
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(&model.OAuth2ConsentSession{
							ID:          1,
							ClientID:    "test",
							Subject:     uuid.NullUUID{UUID: sub, Valid: true},
							Form:        "prompt=login",
							RequestedAt: time.Unix(1000000, 0),
						}, nil),
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSessionResponse(gomock.Eq(mock.Ctx), gomock.Any(), gomock.Eq(false)).
						Return(nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleFormMaxAgeNotRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterMaximumAge: []string{"10"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{
				ClientID:    "test",
				Subject:     uuid.NullUUID{UUID: sub, Valid: true},
				Form:        "max_age=10",
				RequestedAt: time.Unix(1000000, 0),
				RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true},
			},
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(&model.OAuth2ConsentSession{
							ID:          1,
							ClientID:    "test",
							Subject:     uuid.NullUUID{UUID: sub, Valid: true},
							Form:        "max_age=10",
							RequestedAt: time.Unix(1000000, 0),
						}, nil),
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSessionResponse(gomock.Eq(mock.Ctx), gomock.Any(), gomock.Eq(false)).
						Return(nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandlePromptLoginRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(&model.OAuth2ConsentSession{
							ID:          1,
							ClientID:    "test",
							Subject:     uuid.NullUUID{UUID: sub, Valid: true},
							Form:        "prompt=login",
							RequestedAt: time.Unix(1000000, 0),
						}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleMaxAgeRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 0, SecondFactorAuthnTimestamp: 0},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterMaximumAge: []string{"10"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSession(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(&model.OAuth2ConsentSession{
							ID:          1,
							ClientID:    "test",
							Subject:     uuid.NullUUID{UUID: sub, Valid: true},
							Form:        "max_age=10",
							RequestedAt: time.Unix(1000000, 0),
						}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Clock.Set(time.Unix(1000000, 0))

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID: "abc",
							},
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, mock.StorageMock, mock.Ctx.Providers.Templates)

			rw := httptest.NewRecorder()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent, handled := handleOAuth2AuthorizationConsentModeImplicitWithoutID(mock.Ctx, tc.issuer, tc.client, tc.userSession, tc.subject, rw, httptest.NewRequest("GET", "https://example.com", nil), tc.requester)

			assert.Equal(t, tc.handled, handled)

			if tc.expected == nil {
				assert.Nil(t, consent)
			} else {
				require.NotNil(t, consent)

				if tc.expected.ID != 0 {
					assert.Equal(t, tc.expected.ID, consent.ID)
				}

				assert.Equal(t, tc.expected.ClientID, consent.ClientID)
				assert.Equal(t, tc.expected.Subject, consent.Subject)
				assert.Equal(t, tc.expected.Granted, consent.Granted)
				assert.Equal(t, tc.expected.Authorized, consent.Authorized)
				assert.Equal(t, tc.expected.RequestedAt, consent.RequestedAt)
				assert.Equal(t, tc.expected.RespondedAt, consent.RespondedAt)
				assert.Equal(t, tc.expected.Form, consent.Form)
				assert.Equal(t, tc.expected.RequestedScopes, consent.RequestedScopes)
				assert.Equal(t, tc.expected.RequestedAudience, consent.RequestedAudience)
				assert.Equal(t, tc.expected.GrantedScopes, consent.GrantedScopes)
				assert.Equal(t, tc.expected.GrantedAudience, consent.GrantedAudience)
				assert.Equal(t, tc.expected.PreConfiguration, consent.PreConfiguration)
			}

			if tc.expect != nil {
				tc.expect(t, mock)
			}
		})
	}
}

func TestHandleOAuth2AuthorizationConsentModeImplicitWithID(t *testing.T) {
	mustParseURI := func(t *testing.T, in string) *url.URL {
		result, err := url.Parse(in)
		require.NoError(t, err)

		return result
	}

	clientTest := &oidc.RegisteredClient{
		ID:            testValue,
		ConsentPolicy: oidc.ClientConsentPolicy{Mode: oidc.ClientConsentModeImplicit},
	}

	challenge := uuid.MustParse("11303e1f-f8af-436a-9a72-c7361bfc9f37")
	sub := uuid.MustParse("e79b6494-8852-4439-860c-159f2cba83dc")

	testCases := []struct {
		name        string
		issuer      *url.URL
		client      oidc.Client
		userSession session.UserSession
		subject     uuid.UUID
		requester   oauthelia2.AuthorizeRequester
		expected    *model.OAuth2ConsentSession
		handled     bool
		setup       func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expect      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			name:        "ShouldHandleDBError",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(nil, fmt.Errorf("error in db")),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'implicit' could not be processed: error occurred performing consent for consent session with id '11303e1f-f8af-436a-9a72-c7361bfc9f37': error occurred while loading session: error in db$`), nil)
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongID",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: challenge, Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'implicit' could not be processed: error occurred performing consent for consent session with id '11303e1f-f8af-436a-9a72-c7361bfc9f37': user 'test' with subject 'e79b6494-8852-4439-860c-159f2cba83dc' is not authorized to consent for subject '11303e1f-f8af-436a-9a72-c7361bfc9f37'$`), nil)
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongClientID",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ChallengeID: challenge, ClientID: "other", Subject: uuid.NullUUID{UUID: sub, Valid: true}, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10)}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'implicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested client id 'test' does not match the requested client id 'other' from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongNonce",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterNonce: []string{"def456"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Form: "nonce=abc123", ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10)}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'implicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested nonce does not match the requested nonce from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongState",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterNonce: []string{"abc123"},
						oidc.FormParameterState: []string{"uvw321"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Form: "nonce=abc123&state=xyz789", ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10)}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'implicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested state does not match the requested state from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleFormPromptLoginNotRequiredErrorSave",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10)}, nil),
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSessionResponse(gomock.Eq(mock.Ctx), gomock.Any(), gomock.Eq(false)).
						Return(fmt.Errorf("bad conn")),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'implicit' could not be processed: error occurred performing consent for consent session with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}': error occurred saving consent session response: bad conn$`), nil)
			},
		},
		{
			name:        "ShouldHandleFormPromptLoginNotRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{
				ID:          40,
				ClientID:    "test",
				Subject:     uuid.NullUUID{UUID: sub, Valid: true},
				Form:        "prompt=login",
				RequestedAt: time.Unix(1000000, 0),
				ExpiresAt:   time.Unix(1000000, 0).Add(time.Second * 10),
				RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true},
			},
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 40, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Form: "prompt=login", RequestedAt: time.Unix(1000000, 0), ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10)}, nil),
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSessionResponse(gomock.Eq(mock.Ctx), gomock.Any(), gomock.Eq(false)).
						Return(nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleFormPromptLoginRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 100000, SecondFactorAuthnTimestamp: 100000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 40, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Form: "prompt=login", RequestedAt: time.Unix(1000000, 0)}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleFormMaxAgeRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 100000, SecondFactorAuthnTimestamp: 100000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterMaximumAge: []string{"10"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 40, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Form: "max_age=10", RequestedAt: time.Unix(1000000, 0)}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleFormMaxAgeNotRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterMaximumAge: []string{"10"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{
				ID:          40,
				ClientID:    "test",
				Subject:     uuid.NullUUID{UUID: sub, Valid: true},
				Form:        "max_age=10",
				RequestedAt: time.Unix(1000000, 0),
				ExpiresAt:   time.Unix(1000000, 0).Add(time.Second * 10),
				RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true},
			},
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 40, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Form: "max_age=10", RequestedAt: time.Unix(1000000, 0), ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10)}, nil),
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSessionResponse(gomock.Eq(mock.Ctx), gomock.Any(), gomock.Eq(false)).
						Return(nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Clock.Set(time.Unix(1000000, 0))

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID: "abc",
							},
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, mock.StorageMock, mock.Ctx.Providers.Templates)

			rw := httptest.NewRecorder()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent, handled := handleOAuth2AuthorizationConsentModeImplicitWithID(mock.Ctx, tc.issuer, tc.client, tc.userSession, tc.subject, challenge, rw, httptest.NewRequest("GET", "https://example.com", nil), tc.requester)

			assert.Equal(t, tc.handled, handled)

			if tc.expected == nil {
				assert.Nil(t, consent)
			} else {
				require.NotNil(t, consent)

				if tc.expected.ID != 0 {
					assert.Equal(t, tc.expected.ID, consent.ID)
				}

				assert.Equal(t, tc.expected.ClientID, consent.ClientID)
				assert.Equal(t, tc.expected.Subject, consent.Subject)
				assert.Equal(t, tc.expected.Granted, consent.Granted)
				assert.Equal(t, tc.expected.Authorized, consent.Authorized)
				assert.Equal(t, tc.expected.RequestedAt, consent.RequestedAt)
				assert.Equal(t, tc.expected.RespondedAt, consent.RespondedAt)
				assert.Equal(t, tc.expected.Form, consent.Form)
				assert.Equal(t, tc.expected.RequestedScopes, consent.RequestedScopes)
				assert.Equal(t, tc.expected.RequestedAudience, consent.RequestedAudience)
				assert.Equal(t, tc.expected.GrantedScopes, consent.GrantedScopes)
				assert.Equal(t, tc.expected.GrantedAudience, consent.GrantedAudience)
				assert.Equal(t, tc.expected.PreConfiguration, consent.PreConfiguration)
			}

			if tc.expect != nil {
				tc.expect(t, mock)
			}
		})
	}
}

func TestHandleOAuth2AuthorizationConsentModePreConfiguredWithID(t *testing.T) {
	mustParseURI := func(t *testing.T, in string) *url.URL {
		result, err := url.Parse(in)
		require.NoError(t, err)

		return result
	}

	clientTest := &oidc.RegisteredClient{
		ID:            testValue,
		ConsentPolicy: oidc.ClientConsentPolicy{Mode: oidc.ClientConsentModePreConfigured},
	}

	challenge := uuid.MustParse("11303e1f-f8af-436a-9a72-c7361bfc9f37")
	sub := uuid.MustParse("e79b6494-8852-4439-860c-159f2cba83dc")

	testCases := []struct {
		name        string
		issuer      *url.URL
		client      oidc.Client
		userSession session.UserSession
		subject     uuid.UUID
		challenge   uuid.UUID
		requester   oauthelia2.AuthorizeRequester
		expected    *model.OAuth2ConsentSession
		handled     bool
		setup       func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expect      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			name:        "ShouldHandleNilChallenge",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   uuid.Nil,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client:      clientTest,
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup:    nil,
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'pre-configured' could not be processed: the consent id had a zero value$`), nil)
			},
		},
		{
			name:        "ShouldHandleDBError",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client:      clientTest,
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(nil, fmt.Errorf("error in db")),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'pre-configured' could not be processed: error occurred performing consent for consent session with id '11303e1f-f8af-436a-9a72-c7361bfc9f37': error occurred while loading session: error in db$`), nil)
			},
		},
		{
			name:        "ShouldHandleWrongSubject",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client:      clientTest,
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: challenge, Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'pre-configured' could not be processed: error occurred performing consent for consent session with id '11303e1f-f8af-436a-9a72-c7361bfc9f37': user 'test' with subject 'e79b6494-8852-4439-860c-159f2cba83dc' is not authorized to consent for subject '11303e1f-f8af-436a-9a72-c7361bfc9f37'$`), nil)
			},
		},
		{
			name:        "ShouldHandleExpiredSession",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client:      clientTest,
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(-time.Second * 10)}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'pre-configured' could not be processed: error occurred performing consent for consent session with id '11303e1f-f8af-436a-9a72-c7361bfc9f37': the session does not appear to be valid for pre-configured consent: either the subject is null, the consent has been granted and is either not pre-configured, or the pre-configuration is expired$`), nil)
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongClientID",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,

					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "other", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'pre-configured' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested client id 'test' does not match the requested client id 'other' from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongScope",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client:         clientTest,
					RequestedScope: []string{"openid", "profile"},
					RequestedAt:    time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", RequestedScopes: []string{"openid"}, Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'pre-configured' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested scope 'openid profile' does not match the requested scope 'openid' from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongAudience",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client:            clientTest,
					RequestedAudience: []string{"https://b.example.com"},
					RequestedAt:       time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", RequestedAudience: []string{"https://a.example.com"}, Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'pre-configured' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested audience 'https://b\.example\.com' does not match the requested audience 'https://a\.example\.com' from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongNonce",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterNonce: []string{"def456"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Form: "nonce=abc123", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'pre-configured' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested nonce does not match the requested nonce from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongState",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterNonce: []string{"abc123"},
						oidc.FormParameterState: []string{"uvw321"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Form: "nonce=abc123&state=xyz789", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'pre-configured' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested state does not match the requested state from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleMatchingRequestWithoutPreConfiguration",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterNonce: []string{"abc123"},
						oidc.FormParameterState: []string{"xyz789"},
					},
					RequestedScope:    []string{"profile", "openid"},
					RequestedAudience: []string{"https://app.example.com"},
					RequestedAt:       time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "nonce=abc123&state=xyz789", RequestedScopes: []string{"openid", "profile"}, RequestedAudience: []string{"https://app.example.com"}, ExpiresAt: time.Unix(1000000, 0).Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}},
			handled:  false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "nonce=abc123&state=xyz789", RequestedScopes: []string{"openid", "profile"}, RequestedAudience: []string{"https://app.example.com"}, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentPreConfigurations(gomock.Eq(mock.Ctx), gomock.Eq("test"), gomock.Eq(sub), gomock.Eq(mock.Ctx.Providers.Clock.Now())).
						Return(&storage.ConsentPreConfigRows{}, nil),
				)
			},
			expect: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Clock.Set(time.Unix(1000000, 0))

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID: "abc",
							},
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, mock.StorageMock, mock.Ctx.Providers.Templates)

			rw := httptest.NewRecorder()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent, handled := handleOAuth2AuthorizationConsentModePreConfiguredWithID(mock.Ctx, tc.issuer, tc.client, tc.userSession, tc.subject, tc.challenge, rw, httptest.NewRequest("GET", "https://example.com", nil), tc.requester)

			assert.Equal(t, tc.handled, handled)

			if tc.expected == nil {
				assert.Nil(t, consent)
			} else {
				require.NotNil(t, consent)

				if tc.expected.ID != 0 {
					assert.Equal(t, tc.expected.ID, consent.ID)
				}

				assert.Equal(t, tc.expected.ClientID, consent.ClientID)
				assert.Equal(t, tc.expected.Subject, consent.Subject)
				assert.Equal(t, tc.expected.Granted, consent.Granted)
				assert.Equal(t, tc.expected.Authorized, consent.Authorized)
				assert.Equal(t, tc.expected.RequestedAt, consent.RequestedAt)
				assert.Equal(t, tc.expected.RespondedAt, consent.RespondedAt)
				assert.Equal(t, tc.expected.Form, consent.Form)
				assert.Equal(t, tc.expected.RequestedScopes, consent.RequestedScopes)
				assert.Equal(t, tc.expected.RequestedAudience, consent.RequestedAudience)
				assert.Equal(t, tc.expected.GrantedScopes, consent.GrantedScopes)
				assert.Equal(t, tc.expected.GrantedAudience, consent.GrantedAudience)
				assert.Equal(t, tc.expected.PreConfiguration, consent.PreConfiguration)
			}

			if tc.expect != nil {
				tc.expect(t, mock)
			}
		})
	}
}

func TestHandleOAuth2AuthorizationConsentModeExplicitWithID(t *testing.T) {
	mustParseURI := func(t *testing.T, in string) *url.URL {
		result, err := url.Parse(in)
		require.NoError(t, err)

		return result
	}

	clientTest := &oidc.RegisteredClient{
		ID:            testValue,
		ConsentPolicy: oidc.ClientConsentPolicy{Mode: oidc.ClientConsentModeExplicit},
	}

	challenge := uuid.MustParse("11303e1f-f8af-436a-9a72-c7361bfc9f37")
	sub := uuid.MustParse("e79b6494-8852-4439-860c-159f2cba83dc")

	testCases := []struct {
		name        string
		issuer      *url.URL
		client      oidc.Client
		userSession session.UserSession
		subject     uuid.UUID
		challenge   uuid.UUID
		requester   oauthelia2.AuthorizeRequester
		expected    *model.OAuth2ConsentSession
		handled     bool
		setup       func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expect      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			name:        "ShouldHandleDBError",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(nil, fmt.Errorf("error in db")),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred performing consent for consent session with id '11303e1f-f8af-436a-9a72-c7361bfc9f37': error occurred while loading session: error in db$`), nil)
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongID",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: challenge, Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred performing consent for consent session with id '11303e1f-f8af-436a-9a72-c7361bfc9f37': user 'test' with subject 'e79b6494-8852-4439-860c-159f2cba83dc' is not authorized to consent for subject '11303e1f-f8af-436a-9a72-c7361bfc9f37'$`), nil)
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongClientID",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "other", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested client id 'test' does not match the requested client id 'other' from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleMatchingScopesAndAudience",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedScope:    []string{"profile", "openid"},
					RequestedAudience: []string{"https://app.example.com"},
					RequestedAt:       time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, RequestedScopes: []string{"openid", "profile"}, RequestedAudience: []string{"https://app.example.com"}, ExpiresAt: time.Unix(1000000, 0).Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}},
			handled:  false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, RequestedScopes: []string{"openid", "profile"}, RequestedAudience: []string{"https://app.example.com"}, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: nil,
		},
		{
			name:        "ShouldHandleLoadSessionWrongScope",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedScope: []string{"openid", "profile"},
					RequestedAt:    time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, RequestedScopes: []string{"openid"}, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested scope 'openid profile' does not match the requested scope 'openid' from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongAudience",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAudience: []string{"https://b.example.com"},
					RequestedAt:       time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, RequestedAudience: []string{"https://a.example.com"}, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested audience 'https://b\.example\.com' does not match the requested audience 'https://a\.example\.com' from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleMatchingNonceAndState",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterNonce: []string{"abc123"},
						oidc.FormParameterState: []string{"xyz789"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "nonce=abc123&state=xyz789", ExpiresAt: time.Unix(1000000, 0).Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}},
			handled:  false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "nonce=abc123&state=xyz789", ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: nil,
		},
		{
			name:        "ShouldHandlePushedAuthorizationRequestWithMergedNonceAndState",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterClientID:   []string{"test"},
						oidc.FormParameterRequestURI: []string{"urn:ietf:params:oauth:request_uri:abc123"},
						oidc.FormParameterNonce:      []string{"abc123"},
						oidc.FormParameterState:      []string{"xyz789"},
					},
					RequestedScope: []string{"openid"},
					RequestedAt:    time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "client_id=test&request_uri=urn%3Aietf%3Aparams%3Aoauth%3Arequest_uri%3Aabc123", RequestedScopes: []string{"openid"}, ExpiresAt: time.Unix(1000000, 0).Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}},
			handled:  false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "client_id=test&request_uri=urn%3Aietf%3Aparams%3Aoauth%3Arequest_uri%3Aabc123", RequestedScopes: []string{"openid"}, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: nil,
		},
		{
			name:        "ShouldHandlePushedAuthorizationRequestWrongRequestURI",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterClientID:   []string{"test"},
						oidc.FormParameterRequestURI: []string{"urn:ietf:params:oauth:request_uri:def456"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "client_id=test&request_uri=urn%3Aietf%3Aparams%3Aoauth%3Arequest_uri%3Aabc123", ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested request uri 'urn:ietf:params:oauth:request_uri:def456' does not match the requested request uri 'urn:ietf:params:oauth:request_uri:abc123' from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongNonce",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterNonce: []string{"def456"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "nonce=abc123", ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested nonce does not match the requested nonce from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionMissingNonce",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "nonce=abc123", ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested nonce does not match the requested nonce from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionWrongState",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterNonce: []string{"abc123"},
						oidc.FormParameterState: []string{"uvw321"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "nonce=abc123&state=xyz789", ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested state does not match the requested state from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionMissingState",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterNonce: []string{"abc123"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "nonce=abc123&state=xyz789", ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed\. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters\. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified\. The requested state does not match the requested state from the consent session\.$`))
			},
		},
		{
			name:        "ShouldHandleLoadSessionMalformedForm",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, Form: "nonce=%zz", ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred matching the consent session to the request$`), regexp.MustCompile(`^The authorization server encountered an unexpected condition that prevented it from fulfilling the request\. Error occurred parsing the consent request form\. invalid URL escape '%zz'$`))
			},
		},
		{
			name:        "ShouldHandleFormPromptLoginNotRequiredAuthorized",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: &model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, ExpiresAt: time.Unix(1000000, 0).Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}},
			handled:  false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Authorized: true, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleFormPromptLoginNotRequiredNotResponded",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleFormPromptLoginNotRequiredNotRejected",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10), RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred performing consent for consent session with id '11303e1f-f8af-436a-9a72-c7361bfc9f37': the user explicitly rejected this consent session$`), nil)
			},
		},
		{
			name:        "ShouldHandleFormPromptLoginNotRequiredCantGrant",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 44, Granted: true, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, RespondedAt: sql.NullTime{Time: time.Unix(1000000, 0), Valid: true}}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: error occurred performing consent for consent session with id '11303e1f-f8af-436a-9a72-c7361bfc9f37': the session does not appear to be valid for explicit consent: either the subject is null, the consent has already been granted, or the consent session is a pre-configured session$`), nil)
			},
		},
		{
			name:        "ShouldHandleFormPromptLoginNilChallenge",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000},
			subject:     sub,
			challenge:   uuid.Nil,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup:    func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertLastLogMessageRegexp(t, regexp.MustCompile(`^Authorization Request with id '[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}' on client with id 'test' using consent mode 'explicit' could not be processed: the consent id had a zero value$`), nil)
			},
		},
		{
			name:        "ShouldHandleFormPromptLoginRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 100000, SecondFactorAuthnTimestamp: 100000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterPrompt: []string{oidc.PromptLogin},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 40, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Form: "prompt=login", RequestedAt: time.Unix(1000000, 0)}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
		{
			name:        "ShouldHandleFormMaxAgeRequired",
			issuer:      mustParseURI(t, "https://auth.example.com"),
			client:      clientTest,
			userSession: session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 100000, SecondFactorAuthnTimestamp: 100000},
			subject:     sub,
			challenge:   challenge,
			requester: &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
					Form: url.Values{
						oidc.FormParameterMaximumAge: []string{"10"},
					},
					RequestedAt: time.Unix(1000000, 0),
				},
			},
			expected: nil,
			handled:  true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 40, ChallengeID: challenge, ClientID: "test", Subject: uuid.NullUUID{UUID: sub, Valid: true}, Form: "max_age=10", RequestedAt: time.Unix(1000000, 0)}, nil),
				)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Clock.Set(time.Unix(1000000, 0))

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID: "abc",
							},
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, mock.StorageMock, mock.Ctx.Providers.Templates)

			rw := httptest.NewRecorder()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent, handled := handleOAuth2AuthorizationConsentModeExplicitWithID(mock.Ctx, tc.issuer, tc.client, tc.userSession, tc.subject, tc.challenge, rw, httptest.NewRequest("GET", "https://example.com", nil), tc.requester)

			assert.Equal(t, tc.handled, handled)

			if tc.expected == nil {
				assert.Nil(t, consent)
			} else {
				require.NotNil(t, consent)

				if tc.expected.ID != 0 {
					assert.Equal(t, tc.expected.ID, consent.ID)
				}

				assert.Equal(t, tc.expected.ClientID, consent.ClientID)
				assert.Equal(t, tc.expected.Subject, consent.Subject)
				assert.Equal(t, tc.expected.Granted, consent.Granted)
				assert.Equal(t, tc.expected.Authorized, consent.Authorized)
				assert.Equal(t, tc.expected.RequestedAt, consent.RequestedAt)
				assert.Equal(t, tc.expected.RespondedAt, consent.RespondedAt)
				assert.Equal(t, tc.expected.Form, consent.Form)
				assert.Equal(t, tc.expected.RequestedScopes, consent.RequestedScopes)
				assert.Equal(t, tc.expected.RequestedAudience, consent.RequestedAudience)
				assert.Equal(t, tc.expected.GrantedScopes, consent.GrantedScopes)
				assert.Equal(t, tc.expected.GrantedAudience, consent.GrantedAudience)
				assert.Equal(t, tc.expected.PreConfiguration, consent.PreConfiguration)
			}

			if tc.expect != nil {
				tc.expect(t, mock)
			}
		})
	}
}
