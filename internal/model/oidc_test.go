package model_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewOAuth2SessionFromRequest(t *testing.T) {
	challenge := model.NullUUID(uuid.Must(uuid.Parse("a9e4638d-e273-4636-a43e-3b34cc9a76ee")))
	session := &oidc.Session{
		ChallengeID: challenge,
		DefaultSession: &openid.DefaultSession{
			Subject: "sub",
		},
	}

	sessionBytes, _ := json.Marshal(session)

	testCaases := []struct {
		name      string
		signature string
		have      oauthelia2.Requester
		expected  *model.OAuth2Session
		err       string
	}{
		{
			"ShouldNewUpStandard",
			"abc",
			&oauthelia2.Request{
				ID: "example",
				Client: &oauthelia2.DefaultClient{
					ID: "client_id",
				},
				Session:        session,
				RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID},
				GrantedScope:   oauthelia2.Arguments{oidc.ScopeOpenID},
			},
			&model.OAuth2Session{
				ChallengeID:     challenge,
				RequestID:       "example",
				ClientID:        "client_id",
				Signature:       "abc",
				Subject:         sql.NullString{String: "sub", Valid: true},
				RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
				GrantedScopes:   model.StringSlicePipeDelimited{oidc.ScopeOpenID},
				Active:          true,
				Session:         sessionBytes,
			},
			"",
		},
		{
			"ShouldNewUpWithoutScopes",
			"abc",
			&oauthelia2.Request{
				ID: "example",
				Client: &oauthelia2.DefaultClient{
					ID: "client_id",
				},
				Session:        session,
				RequestedScope: nil,
				GrantedScope:   nil,
			},
			&model.OAuth2Session{
				ChallengeID:     challenge,
				RequestID:       "example",
				ClientID:        "client_id",
				Signature:       "abc",
				Subject:         sql.NullString{String: "sub", Valid: true},
				RequestedScopes: model.StringSlicePipeDelimited{},
				GrantedScopes:   model.StringSlicePipeDelimited{},
				Active:          true,
				Session:         sessionBytes,
			},
			"",
		},
		{
			"ShouldRaiseErrorOnInvalidSessionType",
			"abc",
			&oauthelia2.Request{
				ID: "example",
				Client: &oauthelia2.DefaultClient{
					ID: "client_id",
				},
				Session:        &openid.DefaultSession{},
				RequestedScope: nil,
				GrantedScope:   nil,
			},
			nil,
			"failed to create new *model.OAuth2Session: the session type OpenIDSession was expected but the type '*openid.DefaultSession' was used",
		},
		{
			"ShouldRaiseErrorOnNilRequester",
			"abc",
			nil,
			nil,
			"failed to create new *model.OAuth2Session: the oauthelia2.Requester was nil",
		},
	}

	for _, tc := range testCaases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := model.NewOAuth2SessionFromRequest(tc.signature, tc.have)

			if len(tc.err) > 0 {
				assert.Nil(t, actual)
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, actual)

				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestOAuth2Session_SetSubject(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected sql.NullString
	}{
		{
			"ShouldParseValidNullString",
			"example",
			sql.NullString{String: "example", Valid: true},
		},
		{
			"ShouldParseEmptyNullString",
			"",
			sql.NullString{String: "", Valid: false},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			x := &model.OAuth2Session{}

			assert.Equal(t, x.Subject, sql.NullString{})

			x.SetSubject(tc.have)

			assert.Equal(t, tc.expected, x.Subject)
		})
	}
}

func TestOAuth2PARContext_ToAuthorizeRequest(t *testing.T) {
	const (
		parclientid = "par-client-id"
		requestid   = "rid123"
	)

	testCases := []struct {
		name     string
		setup    func(mock *mocks.MockOAuth2Storage)
		have     *model.OAuth2PARContext
		expected *oauthelia2.AuthorizeRequest
		err      string
	}{
		{
			"ShouldErrorInvalidJSONData",
			nil,
			&model.OAuth2PARContext{},
			&oauthelia2.AuthorizeRequest{},
			"error occurred while mapping PAR context back to an Authorize Request while trying to unmarshal the JSON session data: unexpected end of JSON input",
		},
		{
			"ShouldErrorInvalidClient",
			func(mock *mocks.MockOAuth2Storage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(nil, oauthelia2.ErrNotFound)
			},
			&model.OAuth2PARContext{
				ClientID: parclientid,
				Session:  []byte("{}"),
			},
			&oauthelia2.AuthorizeRequest{},
			"error occurred while mapping PAR context back to an Authorize Request while trying to lookup the registered client: not_found",
		},
		{
			"ShouldErrorOnBadForm",
			func(mock *mocks.MockOAuth2Storage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(&oidc.RegisteredClient{ID: parclientid}, nil)
			},
			&model.OAuth2PARContext{
				ID:        1,
				Signature: fmt.Sprintf("%sexample", oidc.RedirectURIPrefixPushedAuthorizationRequestURN),
				RequestID: requestid,
				ClientID:  parclientid,
				Session:   []byte("{}"),
				Form:      ";;;&;;;!@IO#JNM@($*!H@#(&*)!H#E*()!@&GE*)!@QGE*)@G#E*!@&G",
			},
			nil,
			"error occurred while mapping PAR context back to an Authorize Request while trying to parse the original form: invalid semicolon separator in query",
		},
		{
			"ShouldErrorOnBadFormRedirectURI",
			func(mock *mocks.MockOAuth2Storage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(&oidc.RegisteredClient{ID: parclientid}, nil)
			},
			&model.OAuth2PARContext{
				ID:        1,
				Signature: fmt.Sprintf("%sexample", oidc.RedirectURIPrefixPushedAuthorizationRequestURN),
				RequestID: requestid,
				ClientID:  parclientid,
				Session:   []byte("{}"),
				Form:      fmt.Sprintf("redirect_uri=%s", string([]byte{0x00})),
			},
			nil,
			"error occurred while mapping PAR context back to an Authorize Request while trying to parse the original redirect uri: parse \"\\x00\": net/url: invalid control character in URL",
		},
		{
			"ShouldRestoreAuthorizeRequest",
			func(mock *mocks.MockOAuth2Storage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(&oidc.RegisteredClient{ID: parclientid}, nil)
			},
			&model.OAuth2PARContext{
				ID:          1,
				Signature:   fmt.Sprintf("%sexample", oidc.RedirectURIPrefixPushedAuthorizationRequestURN),
				RequestID:   requestid,
				ClientID:    parclientid,
				Session:     []byte(`{"id_token":{"requested_at":"2018-12-10T13:49:51.141Z"}}`),
				RequestedAt: time.Unix(10000000, 0),
				Scopes:      model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeOffline},
				Audience:    model.StringSlicePipeDelimited{parclientid},
				Form: url.Values{
					oidc.FormParameterRedirectURI:  []string{"https://example.com"},
					oidc.FormParameterState:        []string{"abc123"},
					oidc.FormParameterResponseType: []string{oidc.ResponseTypeAuthorizationCodeFlow},
				}.Encode(),
				ResponseMode:         oidc.ResponseModeQuery,
				DefaultResponseMode:  oidc.ResponseModeQuery,
				HandledResponseTypes: model.StringSlicePipeDelimited{oidc.ResponseTypeAuthorizationCodeFlow},
			},
			&oauthelia2.AuthorizeRequest{
				RedirectURI:          MustParseRequestURI(t, "https://example.com"),
				State:                "abc123",
				ResponseMode:         oauthelia2.ResponseModeQuery,
				DefaultResponseMode:  oauthelia2.ResponseModeQuery,
				ResponseTypes:        oauthelia2.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
				HandledResponseTypes: oauthelia2.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
				Request: oauthelia2.Request{
					ID:                requestid,
					Client:            &oidc.RegisteredClient{ID: parclientid},
					RequestedScope:    oauthelia2.Arguments{oidc.ScopeOpenID, oidc.ScopeOffline},
					RequestedAudience: oauthelia2.Arguments{parclientid},
					RequestedAt:       time.Unix(10000000, 0),
					Session:           oidc.NewSessionWithRequestedAt(time.UnixMicro(1544449791141000)),
					Form: url.Values{
						oidc.FormParameterRedirectURI:  []string{"https://example.com"},
						oidc.FormParameterState:        []string{"abc123"},
						oidc.FormParameterResponseType: []string{oidc.ResponseTypeAuthorizationCodeFlow},
					},
				},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			defer ctrl.Finish()

			mock := mocks.NewMockOAuth2Storage(ctrl)

			if tc.setup != nil {
				tc.setup(mock)
			}

			actual, err := tc.have.ToAuthorizeRequest(context.TODO(), oidc.NewSession(), mock)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			}
		})
	}
}

func TestNewOAuth2PARContext(t *testing.T) {
	testCases := []struct {
		name     string
		have     oauthelia2.AuthorizeRequester
		id       string
		expected *model.OAuth2PARContext
		err      string
	}{
		{
			"ShouldCreatePARContext",
			&oauthelia2.AuthorizeRequest{
				HandledResponseTypes: oauthelia2.Arguments{oidc.ResponseTypeHybridFlowIDToken},
				ResponseMode:         oauthelia2.ResponseModeQuery,
				DefaultResponseMode:  oauthelia2.ResponseModeFragment,
				Request: oauthelia2.Request{
					ID:                "a-id",
					RequestedAt:       time.Time{},
					Client:            &oidc.RegisteredClient{ID: "a-client"},
					RequestedScope:    oauthelia2.Arguments{oidc.ScopeOpenID},
					Form:              url.Values{oidc.FormParameterRedirectURI: []string{"https://example.com"}},
					Session:           &oidc.Session{},
					RequestedAudience: oauthelia2.Arguments{"a-client"},
				},
			},
			"123",
			&model.OAuth2PARContext{
				Signature:            "123",
				RequestID:            "a-id",
				ClientID:             "a-client",
				RequestedAt:          time.Time{},
				Scopes:               model.StringSlicePipeDelimited{oidc.ScopeOpenID},
				Audience:             model.StringSlicePipeDelimited{"a-client"},
				HandledResponseTypes: model.StringSlicePipeDelimited{oidc.ResponseTypeHybridFlowIDToken},
				ResponseMode:         oidc.ResponseModeQuery,
				DefaultResponseMode:  oidc.ResponseModeFragment,
				Form:                 "redirect_uri=https%3A%2F%2Fexample.com",
				Session:              []byte(`{"id_token":null,"challenge_id":null,"kid":"","client_id":"","client_credentials":false,"exclude_nbf_claim":false,"allowed_top_level_claims":null,"extra":null}`),
			},
			"",
		},
		{
			"ShouldFailCreateWrongSessionType",
			&oauthelia2.AuthorizeRequest{
				HandledResponseTypes: oauthelia2.Arguments{oidc.ResponseTypeHybridFlowIDToken},
				ResponseMode:         oauthelia2.ResponseModeQuery,
				DefaultResponseMode:  oauthelia2.ResponseModeFragment,
				Request: oauthelia2.Request{
					ID:                "a-id",
					RequestedAt:       time.Time{},
					Client:            &oidc.RegisteredClient{ID: "a-client"},
					RequestedScope:    oauthelia2.Arguments{oidc.ScopeOpenID},
					Form:              url.Values{oidc.FormParameterRedirectURI: []string{"https://example.com"}},
					Session:           &openid.DefaultSession{},
					RequestedAudience: oauthelia2.Arguments{"a-client"},
				},
			},
			"123",
			nil,
			"failed to create new PAR context: can't assert type '*openid.DefaultSession' to an *OAuth2Session",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := model.NewOAuth2PARContext(tc.id, tc.have)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			}
		})
	}
}

func TestOAuth2Session_ToRequest(t *testing.T) {
	const (
		parclientid = "par-client-id"
		requestid   = "rid123"
	)

	testCases := []struct {
		name     string
		setup    func(mock *mocks.MockOAuth2Storage)
		have     *model.OAuth2Session
		expected *oauthelia2.Request
		err      string
	}{
		{
			"ShouldErrorInvalidJSONData",
			nil,
			&model.OAuth2Session{},
			&oauthelia2.Request{},
			"error occurred while mapping OAuth 2.0 Session back to a Request while trying to unmarshal the JSON session data: unexpected end of JSON input",
		},
		{
			"ShouldErrorInvalidClient",
			func(mock *mocks.MockOAuth2Storage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(nil, oauthelia2.ErrNotFound)
			},
			&model.OAuth2Session{
				ClientID: parclientid,
				Session:  []byte("{}"),
			},
			&oauthelia2.Request{},
			"error occurred while mapping OAuth 2.0 Session back to a Request while trying to lookup the registered client: not_found",
		},
		{
			"ShouldErrorOnBadForm",
			func(mock *mocks.MockOAuth2Storage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(&oidc.RegisteredClient{ID: parclientid}, nil)
			},
			&model.OAuth2Session{
				ID:        1,
				Signature: fmt.Sprintf("%sexample", oidc.RedirectURIPrefixPushedAuthorizationRequestURN),
				RequestID: requestid,
				ClientID:  parclientid,
				Session:   []byte("{}"),
				Form:      ";;;&;;;!@IO#JNM@($*!H@#(&*)!H#E*()!@&GE*)!@QGE*)@G#E*!@&G",
			},
			nil,
			"error occurred while mapping OAuth 2.0 Session back to a Request while trying to parse the original form: invalid semicolon separator in query",
		},
		{
			"ShouldRestoreRequest",
			func(mock *mocks.MockOAuth2Storage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(&oidc.RegisteredClient{ID: parclientid}, nil)
			},
			&model.OAuth2Session{
				ID:                1,
				Signature:         fmt.Sprintf("%sexample", oidc.RedirectURIPrefixPushedAuthorizationRequestURN),
				RequestID:         requestid,
				ClientID:          parclientid,
				Session:           []byte(`{"id_token":{"requested_at":"2018-12-10T13:49:51.141Z"}}`),
				RequestedAt:       time.Unix(10000000, 0),
				RequestedScopes:   model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeOffline},
				RequestedAudience: model.StringSlicePipeDelimited{parclientid},
				Form: url.Values{
					oidc.FormParameterRedirectURI:  []string{"https://example.com"},
					oidc.FormParameterState:        []string{"abc123"},
					oidc.FormParameterResponseType: []string{oidc.ResponseTypeAuthorizationCodeFlow},
				}.Encode(),
			},
			&oauthelia2.Request{
				ID:                requestid,
				Client:            &oidc.RegisteredClient{ID: parclientid},
				RequestedScope:    oauthelia2.Arguments{oidc.ScopeOpenID, oidc.ScopeOffline},
				RequestedAudience: oauthelia2.Arguments{parclientid},
				RequestedAt:       time.Unix(10000000, 0),
				Session:           oidc.NewSessionWithRequestedAt(time.UnixMicro(1544449791141000)),
				Form: url.Values{
					oidc.FormParameterRedirectURI:  []string{"https://example.com"},
					oidc.FormParameterState:        []string{"abc123"},
					oidc.FormParameterResponseType: []string{oidc.ResponseTypeAuthorizationCodeFlow},
				},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			defer ctrl.Finish()

			mock := mocks.NewMockOAuth2Storage(ctrl)

			if tc.setup != nil {
				tc.setup(mock)
			}

			actual, err := tc.have.ToRequest(context.TODO(), oidc.NewSession(), mock)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			}
		})
	}
}

func TestNewOAuth2DeviceCodeSessionFromRequest(t *testing.T) {
	testCases := []struct {
		name     string
		have     oauthelia2.DeviceAuthorizeRequester
		expected *model.OAuth2DeviceCodeSession
		err      string
	}{
		{
			"ShouldErrorNil",
			nil,
			nil,
			"failed to create new *model.OAuth2DeviceCodeSession: the oauthelia2.DeviceAuthorizeRequester was nil",
		},
		{
			"ShouldHandleValueTypes",
			&oauthelia2.DeviceAuthorizeRequest{
				Request: oauthelia2.Request{
					ID:          "1",
					RequestedAt: time.Time{},
					Client: &oauthelia2.DefaultClient{
						ID: "example",
					},
					RequestedScope: []string{"openid", "offline"},
					Form: url.Values{
						"scope": []string{"openid", "offline"},
					},
					Session: &oidc.Session{
						DefaultSession: &openid.DefaultSession{
							Username: "john",
							Subject:  "john",
						},
						KID:                   "abc",
						ClientID:              "example",
						ClientCredentials:     false,
						ExcludeNotBeforeClaim: false,
					},
				},
				DeviceCodeSignature: "abc",
				UserCodeSignature:   "123",
				Status:              oauthelia2.DeviceAuthorizeStatusNew,
				LastChecked:         time.Unix(100000000, 0),
			},
			&model.OAuth2DeviceCodeSession{ID: 0, ChallengeID: uuid.NullUUID{UUID: uuid.UUID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, Valid: false}, RequestID: "1", ClientID: "example", Signature: "abc", UserCodeSignature: "123", Status: 0, Subject: sql.NullString{String: "john", Valid: true}, RequestedAt: time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), CheckedAt: time.Date(1973, time.March, 3, 20, 46, 40, 0, time.Local), RequestedScopes: model.StringSlicePipeDelimited{"openid", "offline"}, GrantedScopes: model.StringSlicePipeDelimited{}, RequestedAudience: model.StringSlicePipeDelimited(nil), GrantedAudience: model.StringSlicePipeDelimited(nil), Active: true, Revoked: false, Form: "scope=openid&scope=offline", Session: []uint8{0x7b, 0x22, 0x69, 0x64, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x3a, 0x7b, 0x22, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x3a, 0x22, 0x6a, 0x6f, 0x68, 0x6e, 0x22, 0x2c, 0x22, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x22, 0x3a, 0x22, 0x6a, 0x6f, 0x68, 0x6e, 0x22, 0x2c, 0x22, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x22, 0x3a, 0x22, 0x30, 0x30, 0x30, 0x31, 0x2d, 0x30, 0x31, 0x2d, 0x30, 0x31, 0x54, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x5a, 0x22, 0x7d, 0x2c, 0x22, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x65, 0x6e, 0x67, 0x65, 0x5f, 0x69, 0x64, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x2c, 0x22, 0x6b, 0x69, 0x64, 0x22, 0x3a, 0x22, 0x61, 0x62, 0x63, 0x22, 0x2c, 0x22, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x22, 0x3a, 0x22, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x22, 0x2c, 0x22, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x22, 0x3a, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x22, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x5f, 0x6e, 0x62, 0x66, 0x5f, 0x63, 0x6c, 0x61, 0x69, 0x6d, 0x22, 0x3a, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x22, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x5f, 0x74, 0x6f, 0x70, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x5f, 0x63, 0x6c, 0x61, 0x69, 0x6d, 0x73, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x2c, 0x22, 0x65, 0x78, 0x74, 0x72, 0x61, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x7d}},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := model.NewOAuth2DeviceCodeSessionFromRequest(tc.have)
			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			}
		})
	}
}

func TestOAuth2ConsentPreConfig(t *testing.T) {
	config := &model.OAuth2ConsentPreConfig{
		ClientID: "abc",
	}

	assert.True(t, config.CanConsent())

	config.Revoked = true

	assert.False(t, config.CanConsent())

	config.Revoked = false
	config.ExpiresAt = sql.NullTime{Valid: true}

	assert.False(t, config.CanConsent())

	assert.False(t, config.HasExactGrants([]string{oidc.ScopeProfile}, []string{"abc"}))

	config.Scopes = []string{oidc.ScopeProfile}

	assert.False(t, config.HasExactGrants([]string{oidc.ScopeProfile}, []string{"abc"}))

	config.Audience = []string{"abc"}

	assert.True(t, config.HasExactGrants([]string{oidc.ScopeProfile}, []string{"abc"}))
}

func TestOAuth2ConsentSession(t *testing.T) {
	session := &model.OAuth2ConsentSession{
		ID:        0,
		ClientID:  "a-client",
		ExpiresAt: time.Unix(10000000, 0),
	}

	before := time.Unix(9999999, 0)
	after := time.Unix(10000001, 0)

	assert.True(t, session.CanGrant(before))
	assert.False(t, session.CanGrant(after))

	session.Subject = uuid.NullUUID{Valid: true}

	assert.True(t, session.CanGrant(before))
	assert.False(t, session.IsDenied())
	assert.False(t, session.Responded())
	assert.False(t, session.IsAuthorized())

	session.RespondedAt = sql.NullTime{Valid: true}

	assert.True(t, session.Responded())
	assert.True(t, session.IsDenied())

	session.Authorized = true

	assert.True(t, session.IsAuthorized())
	assert.False(t, session.IsDenied())

	session.Granted = true

	assert.False(t, session.CanGrant(before))

	assert.False(t, session.HasExactGrants([]string{oidc.ScopeOpenID}, []string{"abc"}))

	session.GrantedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID}

	assert.False(t, session.HasExactGrants([]string{oidc.ScopeOpenID}, []string{"abc"}))

	session.GrantedAudience = model.StringSlicePipeDelimited{"abc"}

	assert.True(t, session.HasExactGrants([]string{oidc.ScopeOpenID}, []string{"abc"}))

	session.GrantedScopes = nil
	session.GrantedAudience = nil

	oidc.ConsentGrant(session, true, nil)

	assert.Nil(t, session.GrantedScopes)
	session.HasExactGrantedAudience([]string{"a-client"})

	session.RequestedScopes = []string{oidc.ScopeOpenID}
	session.RequestedAudience = []string{"abc"}

	oidc.ConsentGrant(session, true, nil)

	session.HasExactGrantedScopes([]string{oidc.ScopeOpenID})
	session.HasExactGrantedAudience([]string{"abc", "a-client"})

	form, err := session.GetForm()

	assert.NoError(t, err)
	assert.Equal(t, url.Values{}, form)

	session.Form = "scope=abc"

	form, err = session.GetForm()

	assert.NoError(t, err)
	assert.Equal(t, url.Values{oidc.FormParameterScope: []string{"abc"}}, form)

	session.Form = ";;;&;;;;"

	form, err = session.GetForm()

	assert.EqualError(t, err, "invalid semicolon separator in query")
	assert.Equal(t, url.Values{}, form)
}

func TestMisc(t *testing.T) {
	jti := model.NewOAuth2BlacklistedJTI("abc", time.Unix(10000, 0).UTC())

	assert.Equal(t, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", jti.Signature)
	assert.Equal(t, time.Unix(10000, 0).UTC(), jti.ExpiresAt)

	sub := uuid.MustParse("b9423f3a-65da-4ea8-8f6b-1dafb141f3a8")

	session, err := model.NewOAuth2ConsentSession(time.Now().UTC(), sub, &oauthelia2.Request{Client: &oidc.RegisteredClient{ID: "abc"}})

	assert.NoError(t, err)
	assert.NotNil(t, session)
}

func MustParseRequestURI(t *testing.T, uri string) (parsed *url.URL) {
	t.Helper()

	var err error

	parsed, err = url.ParseRequestURI(uri)
	require.NoError(t, err)

	return parsed
}
