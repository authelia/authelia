package model_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewOAuth2SessionFromRequest(t *testing.T) {
	challenge := model.NullUUID(uuid.Must(uuid.Parse("a9e4638d-e273-4636-a43e-3b34cc9a76ee")))
	session := &model.OpenIDSession{
		ChallengeID: challenge,
		DefaultSession: &openid.DefaultSession{
			Subject: "sub",
		},
	}

	sessionBytes, _ := json.Marshal(session)

	testCaases := []struct {
		name      string
		signature string
		have      fosite.Requester
		expected  *model.OAuth2Session
		err       string
	}{
		{
			"ShouldNewUpStandard",
			"abc",
			&fosite.Request{
				ID: "example",
				Client: &fosite.DefaultClient{
					ID: "client_id",
				},
				Session:        session,
				RequestedScope: fosite.Arguments{oidc.ScopeOpenID},
				GrantedScope:   fosite.Arguments{oidc.ScopeOpenID},
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
			&fosite.Request{
				ID: "example",
				Client: &fosite.DefaultClient{
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
			&fosite.Request{
				ID: "example",
				Client: &fosite.DefaultClient{
					ID: "client_id",
				},
				Session:        &openid.DefaultSession{},
				RequestedScope: nil,
				GrantedScope:   nil,
			},
			nil,
			"failed to create new *model.OAuth2Session: the session type *model.OpenIDSession was expected but the type '*openid.DefaultSession' was used",
		},
		{
			"ShouldRaiseErrorOnNilRequester",
			"abc",
			nil,
			nil,
			"failed to create new *model.OAuth2Session: the fosite.Requester was nil",
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
		setup    func(mock *mocks.MockFositeStorage)
		have     *model.OAuth2PARContext
		expected *fosite.AuthorizeRequest
		err      string
	}{
		{
			"ShouldErrorInvalidJSONData",
			nil,
			&model.OAuth2PARContext{},
			&fosite.AuthorizeRequest{},
			"error occurred while mapping PAR context back to an Authorize Request while trying to unmarshal the JSON session data: unexpected end of JSON input",
		},
		{
			"ShouldErrorInvalidClient",
			func(mock *mocks.MockFositeStorage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(nil, fosite.ErrNotFound)
			},
			&model.OAuth2PARContext{
				ClientID: parclientid,
				Session:  []byte("{}"),
			},
			&fosite.AuthorizeRequest{},
			"error occurred while mapping PAR context back to an Authorize Request while trying to lookup the registered client: not_found",
		},
		{
			"ShouldErrorOnBadForm",
			func(mock *mocks.MockFositeStorage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(&oidc.BaseClient{ID: parclientid}, nil)
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
			func(mock *mocks.MockFositeStorage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(&oidc.BaseClient{ID: parclientid}, nil)
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
			func(mock *mocks.MockFositeStorage) {
				mock.EXPECT().GetClient(context.TODO(), parclientid).Return(&oidc.BaseClient{ID: parclientid}, nil)
			},
			&model.OAuth2PARContext{
				ID:          1,
				Signature:   fmt.Sprintf("%sexample", oidc.RedirectURIPrefixPushedAuthorizationRequestURN),
				RequestID:   requestid,
				ClientID:    parclientid,
				Session:     []byte("{}"),
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
			&fosite.AuthorizeRequest{
				RedirectURI:          MustParseRequestURI("https://example.com"),
				State:                "abc123",
				ResponseMode:         fosite.ResponseModeQuery,
				DefaultResponseMode:  fosite.ResponseModeQuery,
				ResponseTypes:        fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
				HandledResponseTypes: fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
				Request: fosite.Request{
					ID:                requestid,
					Client:            &oidc.BaseClient{ID: parclientid},
					RequestedScope:    fosite.Arguments{oidc.ScopeOpenID, oidc.ScopeOffline},
					RequestedAudience: fosite.Arguments{parclientid},
					RequestedAt:       time.Unix(10000000, 0),
					Session:           oidc.NewSession(),
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

			mock := mocks.NewMockFositeStorage(ctrl)

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
		have     fosite.AuthorizeRequester
		id       string
		expected *model.OAuth2PARContext
		err      string
	}{
		{
			"ShouldCreatePARContext",
			&fosite.AuthorizeRequest{
				HandledResponseTypes: fosite.Arguments{oidc.ResponseTypeHybridFlowIDToken},
				ResponseMode:         fosite.ResponseModeQuery,
				DefaultResponseMode:  fosite.ResponseModeFragment,
				Request: fosite.Request{
					ID:                "a-id",
					RequestedAt:       time.Time{},
					Client:            &oidc.BaseClient{ID: "a-client"},
					RequestedScope:    fosite.Arguments{oidc.ScopeOpenID},
					Form:              url.Values{oidc.FormParameterRedirectURI: []string{"https://example.com"}},
					Session:           &model.OpenIDSession{},
					RequestedAudience: fosite.Arguments{"a-client"},
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
				Session:              []byte{0x7b, 0x22, 0x69, 0x64, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x2c, 0x22, 0x43, 0x68, 0x61, 0x6c, 0x6c, 0x65, 0x6e, 0x67, 0x65, 0x49, 0x44, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x2c, 0x22, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x65, 0x78, 0x74, 0x72, 0x61, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c, 0x7d},
			},
			"",
		},
		{
			"ShouldFailCreateWrongSessionType",
			&fosite.AuthorizeRequest{
				HandledResponseTypes: fosite.Arguments{oidc.ResponseTypeHybridFlowIDToken},
				ResponseMode:         fosite.ResponseModeQuery,
				DefaultResponseMode:  fosite.ResponseModeFragment,
				Request: fosite.Request{
					ID:                "a-id",
					RequestedAt:       time.Time{},
					Client:            &oidc.BaseClient{ID: "a-client"},
					RequestedScope:    fosite.Arguments{oidc.ScopeOpenID},
					Form:              url.Values{oidc.FormParameterRedirectURI: []string{"https://example.com"}},
					Session:           &openid.DefaultSession{},
					RequestedAudience: fosite.Arguments{"a-client"},
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

func TestOpenIDSession(t *testing.T) {
	session := &model.OpenIDSession{
		DefaultSession: &openid.DefaultSession{},
	}

	assert.Nil(t, session.GetIDTokenClaims())
	assert.NotNil(t, session.Clone())

	session = nil

	assert.Nil(t, session.Clone())
}

func MustParseRequestURI(uri string) (parsed *url.URL) {
	var err error

	if parsed, err = url.ParseRequestURI(uri); err != nil {
		panic(err)
	}

	return parsed
}
