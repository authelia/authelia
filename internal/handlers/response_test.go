// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestHandle1FAResponse(t *testing.T) {
	testCases := []struct {
		name      string
		targetURI string
		method    string
		expected  string
		expectedf func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldRedirectWithoutRequestMethod",
			"https://app.example.com/",
			"",
			`{"status":"OK","data":{"redirect":"https://app.example.com/"}}`,
			nil,
		},
		{
			"ShouldRedirectWithUppercaseRequestMethod",
			"https://app.example.com/",
			fasthttp.MethodGet,
			`{"status":"OK","data":{"redirect":"https://app.example.com/"}}`,
			nil,
		},
		{
			"ShouldRedirectWithLowercaseRequestMethod",
			"https://app.example.com/",
			"get",
			`{"status":"OK","data":{"redirect":"https://app.example.com/"}}`,
			nil,
		},
		{
			"ShouldPreserveFragment",
			"https://app.example.com/#/dashboard",
			fasthttp.MethodGet,
			`{"status":"OK","data":{"redirect":"https://app.example.com/#/dashboard"}}`,
			nil,
		},
		{
			"ShouldPreserveQuery",
			"https://app.example.com/?a=1&b=2",
			fasthttp.MethodGet,
			`{"status":"OK","data":{"redirect":"https://app.example.com/?a=1\u0026b=2"}}`,
			nil,
		},
		{
			"ShouldHandleInvalidRequestMethod",
			"https://app.example.com/",
			"GET1",
			`{"status":"KO","message":"Authentication failed. Check your credentials."}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing the target URL 'https://app.example.com/'", "method header with value 'GET1' has invalid characters")
			},
		},
		{
			"ShouldHandleInvalidTargetURI",
			"notaurl",
			fasthttp.MethodGet,
			`{"status":"KO","message":"Authentication failed. Check your credentials."}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing the target URL 'notaurl'", "error occurred parsing object url: parse \"notaurl\": invalid URI for request")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			Handle1FAResponse(mock.Ctx, tc.targetURI, tc.method, testUsername, nil)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestHandlePasskeyResponse(t *testing.T) {
	t.Run("ShouldUseTwoFactorResponseWhenTwoFactor", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		HandlePasskeyResponse(mock.Ctx, testRedirectionURLString, fasthttp.MethodGet, testUsername, nil, true)

		mock.Assert200OK(t, redirectResponse{Redirect: testRedirectionURLString})
	})

	t.Run("ShouldUseOneFactorResponseWhenNotTwoFactor", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
			AccessControl: schema.AccessControl{DefaultPolicy: "one_factor"},
		})

		HandlePasskeyResponse(mock.Ctx, testRedirectionURLString, fasthttp.MethodGet, testUsername, nil, false)

		mock.Assert200OK(t, redirectResponse{Redirect: testRedirectionURLString})
	})

	t.Run("ShouldFallBackToDefaultRedirectionURLWhenNoTargetURL", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		HandlePasskeyResponse(mock.Ctx, "", fasthttp.MethodGet, testUsername, nil, true)

		mock.Assert200OK(t, redirectResponse{Redirect: testRedirectionURLString})
	})
}

func TestDoMarkAuthenticationAttempt(t *testing.T) {
	t.Run("ShouldRecordRequestDetailsFromReferer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Request.Header.SetReferer("https://login.example.com:8080/?rd=https%3A%2F%2Fapp.example.com%2Fpath&rm=POST")

		mock.StorageMock.EXPECT().
			AppendAuthenticationLog(gomock.Any(), gomock.Eq(model.AuthenticationAttempt{
				Username:      testUsername,
				Successful:    true,
				Banned:        false,
				Time:          mock.Clock.Now(),
				Type:          regulation.AuthType1FA,
				RemoteIP:      model.NewNullIPFromString("0.0.0.0"),
				RequestURI:    "https://app.example.com/path",
				RequestMethod: fasthttp.MethodPost,
			})).
			Return(nil)

		expectAuthnSuccess(mock, testUsername, events.StageFirstFactor, events.MethodPassword)

		doMarkAuthenticationAttempt(mock.Ctx, true, regulation.NewBan(regulation.BanTypeNone, testUsername, nil), regulation.AuthType1FA, nil)
	})

	t.Run("ShouldIgnoreMalformedReferer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Request.Header.SetReferer("not a url")

		mock.StorageMock.EXPECT().
			AppendAuthenticationLog(gomock.Any(), gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: false,
				Banned:     false,
				Time:       mock.Clock.Now(),
				Type:       regulation.AuthType1FA,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil)

		expectAuthnFailure(mock, testUsername, events.StageFirstFactor, events.MethodPassword, events.ReasonInvalidCredentials)

		doMarkAuthenticationAttempt(mock.Ctx, false, regulation.NewBan(regulation.BanTypeNone, testUsername, nil), regulation.AuthType1FA, nil)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Unsuccessful 1FA authentication attempt by user 'john'", nil)
	})

	t.Run("ShouldLogBannedUser", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.StorageMock.EXPECT().
			AppendAuthenticationLog(gomock.Any(), gomock.Any()).
			Return(nil)

		expectAuthnFailure(mock, testUsername, events.StageFirstFactor, events.MethodPassword, events.ReasonBanned)

		doMarkAuthenticationAttempt(mock.Ctx, false, regulation.NewBan(regulation.BanTypeUser, testUsername, nil), regulation.AuthType1FA, nil)

		assert.Contains(t, mock.Hook.LastEntry().Message, "they are banned until")
	})
}

func TestHandle2FAResponse(t *testing.T) {
	t.Run("ShouldReplyOKWhenNoDefaultRedirectionURL", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := mock.Ctx.Configuration.Session

		config.Cookies[0].DefaultRedirectionURL = nil

		mock.Ctx.Providers.SessionProvider = session.NewProvider(config, nil)

		Handle2FAResponse(mock.Ctx, "")

		mock.Assert200OK(t, nil)
	})

	t.Run("ShouldHandleMalformedTargetURI", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		Handle2FAResponse(mock.Ctx, "!@#not a uri")

		mock.Assert200KO(t, messageMFAValidationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpUnsafeTargetURI, regexpAnyError)
	})

	t.Run("ShouldReplyOKWhenTargetURIUnsafe", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		Handle2FAResponse(mock.Ctx, "https://not-a-configured-domain.local")

		mock.Assert200OK(t, nil)
	})
}

func TestHandleFlowResponseOpenIDConnectNoSubflow(t *testing.T) {
	newConsent := func(t *testing.T, mock *mocks.MockAutheliaCtx) *model.OAuth2ConsentSession {
		t.Helper()

		return newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))
	}

	t.Run("ShouldHandleAnonymousUser", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newConsent(t, mock)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		userSession := session.UserSession{}

		handleFlowResponse(mock.Ctx, &userSession, consent.ChallengeID.String(), flowNameOpenIDConnect, "", "")

		mock.Assert200KO(t, messageAuthenticationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to redirect for consent as the user is anonymous", nil)
	})

	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newConsent(t, mock)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		clearForwardedHeaders(mock)

		userSession := newTestOIDCUserSession(1)

		handleFlowResponse(mock.Ctx, &userSession, consent.ChallengeID.String(), flowNameOpenIDConnect, "", "")

		mock.Assert200KO(t, messageAuthenticationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred determining the issuer", "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldRedirectToConsentDecisionWhenFormRequiresLogin", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newConsent(t, mock)
		consent.Form = url.Values{oidc.FormParameterPrompt: []string{oidc.PromptLogin}}.Encode()
		consent.RequestedAt = mock.Ctx.GetClock().Now().Add(time.Hour)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		userSession := newTestOIDCUserSession(1)

		handleFlowResponse(mock.Ctx, &userSession, consent.ChallengeID.String(), flowNameOpenIDConnect, "", "")

		body := redirectResponse{}

		mock.GetResponseData(t, &body)

		target, err := url.Parse(body.Redirect)

		require.NoError(t, err)

		assert.Equal(t, oidc.FrontendEndpointPathConsentDecision, target.Path)
		assert.Equal(t, consent.ChallengeID.String(), target.Query().Get(queryArgFlowID))
	})

	t.Run("ShouldRedirectToAuthorizationWhenAuthenticationSufficient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newConsent(t, mock)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		userSession := newTestOIDCUserSession(1)

		handleFlowResponse(mock.Ctx, &userSession, consent.ChallengeID.String(), flowNameOpenIDConnect, "", "")

		body := redirectResponse{}

		mock.GetResponseData(t, &body)

		target, err := url.Parse(body.Redirect)

		require.NoError(t, err)

		assert.Equal(t, oidc.EndpointPathAuthorization, target.Path)
		assert.Equal(t, consent.ChallengeID.String(), target.Query().Get(queryArgConsentID))
	})

	t.Run("ShouldHandleMalformedFormOnConsentSession", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newConsent(t, mock)
		consent.Form = "%zz"

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		userSession := newTestOIDCUserSession(1)

		handleFlowResponse(mock.Ctx, &userSession, consent.ChallengeID.String(), flowNameOpenIDConnect, "", "")

		mock.Assert200KO(t, messageAuthenticationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred getting the original form from the consent session", regexpAnyError)
	})
}

func TestAuthenticationEventStageAndMethod(t *testing.T) {
	testCases := []struct {
		name          string
		authType      string
		stage, method string
	}{
		{"ShouldMapFirstFactor", regulation.AuthType1FA, events.StageFirstFactor, events.MethodPassword},
		{"ShouldMapPasskeyToFirstFactorWebAuthn", regulation.AuthTypePasskey, events.StageFirstFactor, events.MethodWebAuthn},
		{"ShouldMapPasswordToSecondFactor", regulation.AuthTypePassword, events.StageSecondFactor, events.MethodPassword},
		{"ShouldMapTOTP", regulation.AuthTypeTOTP, events.StageSecondFactor, events.MethodTOTP},
		{"ShouldMapWebAuthn", regulation.AuthTypeWebAuthn, events.StageSecondFactor, events.MethodWebAuthn},
		{"ShouldMapDuo", regulation.AuthTypeDuo, events.StageSecondFactor, events.MethodDuo},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stage, method := doGetAuthenticationEventStageAndMethod(tc.authType)

			assert.Equal(t, tc.stage, stage)
			assert.Equal(t, tc.method, method)
		})
	}
}

func TestAuthenticationEventRejectedRendersAsTheWrappedError(t *testing.T) {
	inner := fmt.Errorf("duo auth result: 'deny', status: 'deny', message: 'Login request denied.'")

	rejected := newErrAuthenticationRejected(inner)

	assert.Equal(t, inner.Error(), rejected.Error())
	assert.ErrorIs(t, rejected, inner)
}

func TestAuthenticationEventReason(t *testing.T) {
	testCases := []struct {
		name       string
		successful bool
		ban        *regulation.Ban
		errAuth    error
		expected   string
	}{
		{"ShouldNotClassifyASuccess", true, regulation.NewBan(regulation.BanTypeNone, testUsername, nil), nil, ""},
		{"ShouldClassifyABannedUser", false, regulation.NewBan(regulation.BanTypeUser, testUsername, nil), nil, events.ReasonBanned},
		{"ShouldClassifyABannedIP", false, regulation.NewBan(regulation.BanTypeIP, "127.0.0.1", nil), nil, events.ReasonBanned},
		{"ShouldPreferTheBanOverTheError", false, regulation.NewBan(regulation.BanTypeUser, testUsername, nil), fmt.Errorf("backend unreachable"), events.ReasonBanned},
		{"ShouldClassifyAnErrorAsInternal", false, regulation.NewBan(regulation.BanTypeNone, testUsername, nil), fmt.Errorf("backend unreachable"), events.ReasonInternalError},
		{"ShouldClassifyARejectionAsInvalidCredentials", false, regulation.NewBan(regulation.BanTypeNone, testUsername, nil), newErrAuthenticationRejected(fmt.Errorf("duo auth result: 'deny'")), events.ReasonInvalidCredentials},
		{"ShouldUnwrapAWrappedRejection", false, regulation.NewBan(regulation.BanTypeNone, testUsername, nil), fmt.Errorf("outer: %w", newErrAuthenticationRejected(fmt.Errorf("duo auth result: 'deny'"))), events.ReasonInvalidCredentials},
		{"ShouldPreferTheBanOverTheRejection", false, regulation.NewBan(regulation.BanTypeUser, testUsername, nil), newErrAuthenticationRejected(fmt.Errorf("duo auth result: 'deny'")), events.ReasonBanned},
		{"ShouldClassifyTheAbsenceOfAnErrorAsInvalidCredentials", false, regulation.NewBan(regulation.BanTypeNone, testUsername, nil), nil, events.ReasonInvalidCredentials},
		{"ShouldClassifyANotFoundUser", false, regulation.NewBan(regulation.BanTypeUnknown, "", nil), authentication.ErrUserNotFound, events.ReasonUserNotFound},
		{"ShouldUnwrapAWrappedNotFoundUser", false, regulation.NewBan(regulation.BanTypeUnknown, "", nil), fmt.Errorf("failed to retrieve user details for user john: %w", authentication.ErrUserNotFound), events.ReasonUserNotFound},
		{"ShouldPreferTheBanOverTheNotFoundUser", false, regulation.NewBan(regulation.BanTypeIP, "127.0.0.1", nil), authentication.ErrUserNotFound, events.ReasonBanned},
		{"ShouldPreferTheRejectionOverTheNotFoundUser", false, regulation.NewBan(regulation.BanTypeNone, testUsername, nil), newErrAuthenticationRejected(authentication.ErrUserNotFound), events.ReasonInvalidCredentials},
		{"ShouldNotClassifyAnUnreachableBackendAsNotFound", false, regulation.NewBan(regulation.BanTypeUnknown, "", nil), fmt.Errorf("failed to connect: connection refused"), events.ReasonInternalError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, doGetAuthenticationEventReason(tc.successful, tc.ban, tc.errAuth))
		})
	}
}

func TestAuthenticationEventStageAndMethodCoversEveryAuthType(t *testing.T) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filepath.Join("..", "regulation", "const.go"), nil, 0)
	require.NoError(t, err)

	stages := []string{events.StageFirstFactor, events.StageSecondFactor}
	methods := []string{events.MethodPassword, events.MethodTOTP, events.MethodWebAuthn, events.MethodDuo}

	var n int

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}

		for _, spec := range gen.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || len(value.Names) != 1 || len(value.Values) != 1 || !strings.HasPrefix(value.Names[0].Name, "AuthType") {
				continue
			}

			lit, ok := value.Values[0].(*ast.BasicLit)
			require.True(t, ok, "%s is not a string literal", value.Names[0].Name)

			n++

			stage, method := doGetAuthenticationEventStageAndMethod(strings.Trim(lit.Value, `"`))

			assert.Contains(t, stages, stage, "%s does not map to a documented stage", value.Names[0].Name)
			assert.Contains(t, methods, method, "%s does not map to a documented method", value.Names[0].Name)
		}
	}

	assert.Equal(t, 6, n, "the number of authentication types changed, update the mapping and this expectation")
}
