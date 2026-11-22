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
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestHandleOAuth2ConsentSessionBind(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		existing bool
		err      string
	}{
		{
			name:     "ShouldBindToTheExistingBindingOfTheUserAgent",
			existing: true,
		},
		{
			name: "ShouldBindToANewBindingWhenTheUserAgentHasNone",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.DelCookie(session.UserAgentBindingCookieName("authelia_session"))
			},
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

			binding, err := handleOAuth2ConsentSessionBind(mock.Ctx, consent)

			assert.Nil(t, mock.Ctx.Response.Header.PeekCookie(session.UserAgentBindingCookieName("authelia_session")))

			switch {
			case tc.err != "":
				assert.EqualError(t, err, tc.err)
				assert.Empty(t, binding)
				assert.Empty(t, consent.Binding)
			case tc.existing:
				assert.NoError(t, err)
				assert.Equal(t, mocks.UserAgentBinding, binding)
				assert.Equal(t, mocks.UserAgentBindingDigest, consent.Binding)
			default:
				assert.NoError(t, err)
				assert.Len(t, binding, 43)
				assert.NotEqual(t, mocks.UserAgentBinding, binding)
				assert.Equal(t, session.UserAgentBindingDigest(binding), consent.Binding)
			}
		})
	}
}

func TestHandleOAuth2ConsentSessionIssueBinding(t *testing.T) {
	t.Run("ShouldIssueTheBindingToTheUserAgent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		require.NoError(t, handleOAuth2ConsentSessionIssueBinding(mock.Ctx, mocks.UserAgentBinding))

		assert.Equal(t, mocks.UserAgentBinding, assertTestOAuth2UserAgentBindingCookie(t, mock))
	})

	t.Run("ShouldHandleSessionProviderError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		clearForwardedHeaders(mock)

		assert.EqualError(t, handleOAuth2ConsentSessionIssueBinding(mock.Ctx, mocks.UserAgentBinding), "error occurred obtaining the session provider to issue the consent session binding: unable to retrieve session cookie domain: missing required X-Forwarded-Host header")
		assert.Nil(t, mock.Ctx.Response.Header.PeekCookie(session.UserAgentBindingCookieName("authelia_session")))
	})
}

func TestHandleOAuth2ConsentSessionRenewBinding(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		existing bool
		err      string
	}{
		{
			name:     "ShouldRenewTheExistingBindingOfTheUserAgent",
			existing: true,
		},
		{
			name: "ShouldIssueANewBindingWhenTheUserAgentHasNone",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.DelCookie(session.UserAgentBindingCookieName("authelia_session"))
			},
		},
		{
			name: "ShouldHandleSessionProviderError",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				clearForwardedHeaders(mock)
			},
			err: "error occurred obtaining the session provider to renew the consent session binding: unable to retrieve session cookie domain: missing required X-Forwarded-Host header",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			err := handleOAuth2ConsentSessionRenewBinding(mock.Ctx)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, mock.Ctx.Response.Header.PeekCookie(session.UserAgentBindingCookieName("authelia_session")))

				return
			}

			require.NoError(t, err)

			binding := assertTestOAuth2UserAgentBindingCookie(t, mock)

			if tc.existing {
				assert.Equal(t, mocks.UserAgentBinding, binding)
			} else {
				assert.NotEqual(t, mocks.UserAgentBinding, binding)
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
					Return(&model.OAuth2ConsentSession{ChallengeID: challenge, Binding: mocks.UserAgentBindingDigest}, nil)
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

func TestHandleOAuth2ConsentSessionLoadBoundOrClaim(t *testing.T) {
	challenge := uuid.MustParse("6d7e6b0e-5d3a-4c8e-9b1e-1f2a3b4c5d6e")

	testCases := []struct {
		name    string
		setup   func(t *testing.T, mock *mocks.MockAutheliaCtx)
		err     string
		binding string
		issued  bool
	}{
		{
			name: "ShouldLoadBoundConsentSession",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(&model.OAuth2ConsentSession{ID: 1, ChallengeID: challenge, Binding: mocks.UserAgentBindingDigest}, nil)
			},
			binding: mocks.UserAgentBindingDigest,
		},
		{
			name: "ShouldRejectConsentSessionBoundToAnotherUserAgent",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(&model.OAuth2ConsentSession{ID: 1, ChallengeID: challenge, Binding: newTestSessionCookie("another-user-agent")}, nil)
			},
			err: "the consent session is not bound to the user agent",
		},
		{
			name: "ShouldClaimUnboundConsentSessionWithTheExistingBinding",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 1, ChallengeID: challenge}, nil),
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSessionBinding(gomock.Eq(mock.Ctx), gomock.Eq(&model.OAuth2ConsentSession{ID: 1, ChallengeID: challenge, Binding: mocks.UserAgentBindingDigest})).
						Return(nil),
				)
			},
			binding: mocks.UserAgentBindingDigest,
			issued:  true,
		},
		{
			name: "ShouldClaimUnboundConsentSessionWithANewBinding",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.DelCookie(session.UserAgentBindingCookieName("authelia_session"))

				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 1, ChallengeID: challenge}, nil),
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSessionBinding(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(nil),
				)
			},
			issued: true,
		},
		{
			name: "ShouldRejectConsentSessionClaimedConcurrently",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				gomock.InOrder(
					mock.StorageMock.EXPECT().
						LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
						Return(&model.OAuth2ConsentSession{ID: 1, ChallengeID: challenge}, nil),
					mock.StorageMock.EXPECT().
						SaveOAuth2ConsentSessionBinding(gomock.Eq(mock.Ctx), gomock.Any()).
						Return(fmt.Errorf("error updating oauth2 consent session (binding): %w", storage.ErrNoRowsAffected)),
				)
			},
			err: "error occurred claiming the consent session: error updating oauth2 consent session (binding): no rows affected",
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

			consent, err := handleOAuth2ConsentSessionLoadBoundOrClaim(mock.Ctx, challenge)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, consent)
				assert.Nil(t, mock.Ctx.Response.Header.PeekCookie(session.UserAgentBindingCookieName("authelia_session")))

				return
			}

			require.NoError(t, err)
			assert.Equal(t, challenge, consent.ChallengeID)

			if !tc.issued {
				assert.Equal(t, tc.binding, consent.Binding)
				assert.Nil(t, mock.Ctx.Response.Header.PeekCookie(session.UserAgentBindingCookieName("authelia_session")))

				return
			}

			binding := assertTestOAuth2UserAgentBindingCookie(t, mock)

			assert.Equal(t, session.UserAgentBindingDigest(binding), consent.Binding)

			if tc.binding != "" {
				assert.Equal(t, tc.binding, consent.Binding)
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
			Binding:         mocks.UserAgentBindingDigest,
			ExpiresAt:       mock.Ctx.GetClock().Now().Add(time.Minute),
			RequestedScopes: []string{oidc.ScopeOpenID},
			Form:            url.Values{oidc.FormParameterState: []string{"abc"}}.Encode(),
		}
	}

	testCases := []struct {
		name      string
		consentID string
		setup     func(t *testing.T, mock *mocks.MockAutheliaCtx)
		restart   bool
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
			name:      "ShouldRestartForRespondedConsentSession",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				consent := newPending(mock)
				consent.SetRespondedAt(mock.Ctx.GetClock().Now(), 0)

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(consent, nil)
			},
			restart: true,
		},
		{
			name:      "ShouldRestartForConsentSessionWithSubject",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				consent := newPending(mock)
				consent.SetSubject(challenge)

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(consent, nil)
			},
			restart: true,
		},
		{
			name:      "ShouldRestartForExpiredConsentSession",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				consent := newPending(mock)
				consent.ExpiresAt = mock.Ctx.GetClock().Now().Add(-time.Minute)

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(consent, nil)
			},
			restart: true,
		},
		{
			name:      "ShouldRestartForConsentSessionBoundToAnotherUserAgent",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				consent := newPending(mock)
				consent.Binding = testOAuth2UserAgentBindingAnotherUserAgent

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(consent, nil)
			},
			restart: true,
			log:     regexpConsentPendingRestart,
			err:     regexp.MustCompile(`^the consent session is not bound to the user agent$`),
		},
		{
			name:      "ShouldRestartForMissingConsentSession",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(nil, fmt.Errorf("error selecting oauth2 consent session: %w", sql.ErrNoRows))
			},
			restart: true,
			log:     regexpConsentPendingRestart,
			err:     regexp.MustCompile(`no rows in result set$`),
		},
		{
			name:      "ShouldRestartForMalformedConsentID",
			consentID: "not-a-uuid",
			restart:   true,
			log:       regexpConsentPendingRestart,
			err:       regexp.MustCompile(`^invalid UUID length: 10$`),
		},
		{
			name:      "ShouldNotRestartForLoadError",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(nil, fmt.Errorf("bad"))
			},
			restart: true,
			handled: true,
			log:     regexp.MustCompile(`error occurred while loading session: bad$`),
		},
		{
			name:      "ShouldRestartForConsentSessionWhichDoesNotMatchTheRequest",
			consentID: challenge.String(),
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				consent := newPending(mock)
				consent.Form = url.Values{oidc.FormParameterState: []string{"xyz"}}.Encode()

				mock.StorageMock.EXPECT().
					LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
					Return(consent, nil)
			},
			restart: true,
			log:     regexpConsentPendingRestart,
			err:     regexpAnyError,
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

			consent, handled := handleOAuth2AuthorizationConsentPending(mock.Ctx, client, httptest.NewRecorder(), requester, tc.restart)

			assert.Equal(t, tc.handled, handled)

			if tc.pending {
				require.NotNil(t, consent)
				assert.Equal(t, challenge, consent.ChallengeID)
			} else {
				assert.Nil(t, consent)
			}

			if tc.log != nil {
				mock.AssertLastLogMessageRegexp(t, tc.log, tc.err)
			}
		})
	}
}

var regexpConsentPendingRestart = regexp.MustCompile(`^Authorization Request with id '[^']*' on client with id 'test' is being restarted as the referenced consent session with id '[^']+' can't be used by this user agent$`)
