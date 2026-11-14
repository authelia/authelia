// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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
)

func TestHandleOAuth2ConsentSessionBind(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected string
		err      string
	}{
		{
			name:     "ShouldBindToTheUserAgent",
			expected: mocks.AuthorizationBinding,
		},
		{
			name: "ShouldHandleSessionProviderError",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				clearForwardedHeaders(mock)
			},
			err: "error occurred obtaining the session provider to bind the consent session: unable to retrieve session cookie domain: missing required X-Forwarded-Host header",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent := &model.OAuth2ConsentSession{}

			err := handleOAuth2ConsentSessionBind(mock.Ctx, consent)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Empty(t, consent.Binding)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, consent.Binding)
			}
		})
	}
}

func TestHandleOAuth2ConsentSessionLoadBound(t *testing.T) {
	challenge := uuid.MustParse("6d7e6b0e-5d3a-4c8e-9b1e-1f2a3b4c5d6e")

	testCases := []struct {
		name  string
		setup func(t *testing.T, mock *mocks.MockAutheliaCtx)
		err   string
	}{
		{
			name: "ShouldLoadBoundConsentSession",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(&model.OAuth2ConsentSession{ChallengeID: challenge, Binding: mocks.AuthorizationBinding}, nil)
			},
		},
		{
			name: "ShouldRejectConsentSessionBoundToAnotherUserAgent",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(&model.OAuth2ConsentSession{ChallengeID: challenge, Binding: newTestSessionCookie("another-user-agent")}, nil)
			},
			err: "the consent session is not bound to the user agent",
		},
		{
			name: "ShouldRejectUnboundConsentSession",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(&model.OAuth2ConsentSession{ChallengeID: challenge}, nil)
			},
			err: "the consent session is not bound to the user agent",
		},
		{
			name: "ShouldHandleStorageError",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(nil, sql.ErrNoRows)
			},
			err: "sql: no rows in result set",
		},
		{
			name: "ShouldHandleSessionProviderError",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				clearForwardedHeaders(mock)
			},
			err: "error occurred obtaining the session provider to verify the consent session binding: unable to retrieve session cookie domain: missing required X-Forwarded-Host header",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			consent, err := handleOAuth2ConsentSessionLoadBound(mock.Ctx, challenge)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, consent)
			} else {
				require.NoError(t, err)
				assert.Equal(t, challenge, consent.ChallengeID)
			}
		})
	}
}

func TestHandleOAuth2AuthorizationConsentPending(t *testing.T) {
	challenge := uuid.MustParse("6d7e6b0e-5d3a-4c8e-9b1e-1f2a3b4c5d6e")

	client := &oidc.RegisteredClient{ID: "test"}

	newPending := func(mock *mocks.MockAutheliaCtx) *model.OAuth2ConsentSession {
		return &model.OAuth2ConsentSession{
			ChallengeID:     challenge,
			ClientID:        "test",
			Binding:         mocks.AuthorizationBinding,
			ExpiresAt:       mock.Ctx.GetClock().Now().Add(time.Minute),
			RequestedScopes: []string{oidc.ScopeOpenID},
			Form:            url.Values{oidc.FormParameterState: []string{"abc"}}.Encode(),
		}
	}

	testCases := []struct {
		name      string
		consentID string
		setup     func(t *testing.T, mock *mocks.MockAutheliaCtx)
		pending   bool
		handled   bool
		log       *regexp.Regexp
		err       *regexp.Regexp
	}{
		{
			name: "ShouldReturnNothingWithoutConsentID",
		},
		{
			name:      "ShouldReturnPendingConsentSession",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(newPending(mock), nil)
			},
			pending: true,
		},
		{
			name:      "ShouldHandleMalformedConsentID",
			consentID: "not-a-uuid",
			handled:   true,
			log:       regexpConsentMalformedChallengeID,
		},
		{
			name:      "ShouldHandleLoadError",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(nil, fmt.Errorf("bad"))
			},
			handled: true,
			log:     regexp.MustCompile(`error occurred while loading session: bad$`),
		},
		{
			name:      "ShouldRejectRespondedConsentSession",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				consent := newPending(mock)
				consent.SetRespondedAt(mock.Ctx.GetClock().Now(), 0)

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(consent, nil)
			},
			handled: true,
			log:     regexp.MustCompile(`error occurred generating consent: the consent session is not pending a response$`),
		},
		{
			name:      "ShouldRejectConsentSessionWithSubject",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				consent := newPending(mock)
				consent.SetSubject(challenge)

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(consent, nil)
			},
			handled: true,
			log:     regexp.MustCompile(`error occurred generating consent: the consent session is not pending a response$`),
		},
		{
			name:      "ShouldRejectExpiredConsentSession",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				consent := newPending(mock)
				consent.ExpiresAt = mock.Ctx.GetClock().Now().Add(-time.Minute)

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(consent, nil)
			},
			handled: true,
			log:     regexp.MustCompile(`error occurred generating consent: the consent session is not pending a response$`),
		},
		{
			name:      "ShouldRejectConsentSessionWhichDoesNotMatchTheRequest",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				consent := newPending(mock)
				consent.Form = url.Values{oidc.FormParameterState: []string{"xyz"}}.Encode()

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(consent, nil)
			},
			handled: true,
			log:     regexp.MustCompile(`error occurred matching the consent session to the request$`),
			err:     regexpAnyError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			setupTestOIDCProvider(t, mock, nil)

			if tc.consentID != "" {
				mock.Ctx.Request.SetRequestURI(oidc.EndpointPathAuthorization + "?" + url.Values{queryArgConsentID: []string{tc.consentID}}.Encode())
			}

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			requester := &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client:         client,
					RequestedScope: oauthelia2.Arguments{oidc.ScopeOpenID},
					Form:           url.Values{oidc.FormParameterState: []string{"abc"}},
				},
			}

			consent, handled := handleOAuth2AuthorizationConsentPending(mock.Ctx, client, httptest.NewRecorder(), requester)

			assert.Equal(t, tc.handled, handled)

			if tc.pending {
				require.NotNil(t, consent)
				assert.Equal(t, challenge, consent.ChallengeID)
			}

			if tc.log != nil {
				mock.AssertLastLogMessageRegexp(t, tc.log, tc.err)
			}
		})
	}
}
