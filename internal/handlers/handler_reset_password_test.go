package handlers

import (
	"database/sql"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestResetPasswordDELETE(t *testing.T) {
	jti := uuid.Must(uuid.Parse("01020304-0506-4722-8910-111213141500"))

	testCases := []struct {
		name           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleBodyParseError",
			nil,
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing reset password delete body", "unable to parse body: unexpected end of JSON input")
			},
		},
		{
			"ShouldHandleIssuerError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)
				mock.Ctx.Request.SetBodyString(`{"token":"abc"}`)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred determining the issuer", "missing required X-Forwarded-Host header")
			},
		},
		{
			"ShouldHandleMalformedToken",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.SetBodyString(`{"token":"abc"}`)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating the identity verification token as it appears to be malformed, this potentially can occur if you've not copied the full link", "token is malformed: token contains an invalid number of segments")
			},
		},
		{
			"ShouldHandleExpiredToken",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, ActionResetPassword)
				claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS256, testJWTSecret)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating the identity verification token validity period as it appears to be expired", "token has invalid claims: token is expired")
			},
		},
		{
			"ShouldHandleTokenNotValidYet",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, ActionResetPassword)
				claims.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Minute))

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS256, testJWTSecret)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating the identity verification token validity period as it appears to only be valid in the future", "token has invalid claims: token is not valid yet")
			},
		},
		{
			"ShouldHandleInvalidSignature",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, ActionResetPassword)

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS256, "not-the-configured-secret")
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating the identity verification token signature", "token signature is invalid: signature is invalid")
			},
		},
		{
			"ShouldHandleInvalidSigningAlgorithm",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, ActionResetPassword)

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS512, testJWTSecret)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating the identity verification token signature", "token signature is invalid: signing method HS512 is invalid")
			},
		},
		{
			"ShouldHandleInvalidIssuerClaim",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, ActionResetPassword)
				claims.Issuer = "https://auth.notexample.com"

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS256, testJWTSecret)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating the identity verification token", "token has invalid claims: token has invalid issuer")
			},
		},
		{
			"ShouldHandleMalformedClaims",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, ActionResetPassword)
				claims.ID = "not-a-uuid"

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS256, testJWTSecret)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating the identity verification token claims as they appear to be malformed", "invalid UUID length: 10")
			},
		},
		{
			"ShouldHandleMismatchedAction",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, "OtherAction")

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS256, testJWTSecret)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking the identity verification token, the token action 'OtherAction' does not match the endpoint action 'ResetPassword' which is not allowed", nil)
			},
		},
		{
			"ShouldHandleStorageLoadError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, ActionResetPassword)

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS256, testJWTSecret)

				mock.StorageMock.
					EXPECT().
					LoadIdentityVerification(mock.Ctx, jti.String()).
					Return(nil, fmt.Errorf("failed to load"))
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred looking up identity verification during the revocation phase", "failed to load")
			},
		},
		{
			"ShouldHandleAlreadyRevoked",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, ActionResetPassword)

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS256, testJWTSecret)

				mock.StorageMock.
					EXPECT().
					LoadIdentityVerification(mock.Ctx, jti.String()).
					Return(&model.IdentityVerification{
						JTI:       jti,
						Username:  testUsername,
						Action:    ActionResetPassword,
						RevokedAt: sql.NullTime{Valid: true, Time: time.Now()},
					}, nil)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking identity verification token as it's already revoked", nil)
			},
		},
		{
			"ShouldHandleStorageRevokeError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, ActionResetPassword)

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS256, testJWTSecret)

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadIdentityVerification(mock.Ctx, jti.String()).
						Return(&model.IdentityVerification{JTI: jti, Username: testUsername, Action: ActionResetPassword}, nil),
					mock.StorageMock.
						EXPECT().
						RevokeIdentityVerification(mock.Ctx, jti.String(), model.NewNullIP(net.ParseIP("0.0.0.0"))).
						Return(fmt.Errorf("failed to revoke")),
				)
			},
			`{"status":"KO","message":"Operation failed."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking identity verification when attempting to save the revocation status to the database", "failed to revoke")
			},
		},
		{
			"ShouldHandleValidToken",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				claims := newTestIdentityVerificationClaim(t, mock, jti, ActionResetPassword)

				setTestIdentityVerificationTokenBody(t, mock, claims, jwt.SigningMethodHS256, testJWTSecret)

				gomock.InOrder(
					mock.StorageMock.
						EXPECT().
						LoadIdentityVerification(mock.Ctx, jti.String()).
						Return(&model.IdentityVerification{JTI: jti, Username: testUsername, Action: ActionResetPassword}, nil),
					mock.StorageMock.
						EXPECT().
						RevokeIdentityVerification(mock.Ctx, jti.String(), model.NewNullIP(net.ParseIP("0.0.0.0"))).
						Return(nil),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret = testJWTSecret
			mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTAlgorithm = "HS256"

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			ResetPasswordDELETE(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestResetPasswordPOST(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected       string
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleNoIdentityVerification",
			nil,
			`{"status":"KO","message":"Unable to reset your password."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred resetting the password as no identity verification process has been initiated", nil)
			},
		},
		{
			"ShouldHandleBodyParseError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setTestPasswordResetUsername(t, mock)
			},
			`{"status":"KO","message":"Unable to reset your password."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing the reset password request body", "unable to parse body: unexpected end of JSON input")
			},
		},
		{
			"ShouldHandlePasswordPolicyError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setTestPasswordResetUsername(t, mock)

				mock.Ctx.Providers.PasswordPolicy = middlewares.NewPasswordPolicyProvider(schema.PasswordPolicy{
					Standard: schema.PasswordPolicyStandard{Enabled: true, MinLength: 8},
				})

				mock.Ctx.Request.SetBodyString(`{"password":"abc"}`)
			},
			`{"status":"KO","message":"Your supplied password does not meet the password policy requirements."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred checking the new password against the password policy", "the supplied password does not met the security policy")
			},
		},
		{
			"ShouldHandleBackendComplexityError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setTestPasswordResetUsername(t, mock)

				mock.Ctx.Request.SetBodyString(`{"password":"password123"}`)

				mock.UserProviderMock.
					EXPECT().
					UpdatePassword(testUsername, "password123").
					Return(fmt.Errorf("LDAP Result Code 19 \"Constraint Violation\": Password fails quality checking policy"))
			},
			`{"status":"KO","message":"0000052D."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred updating the user password as it does not meet the complexity requirements of the backend", "LDAP Result Code 19 \"Constraint Violation\": Password fails quality checking policy")
			},
		},
		{
			"ShouldHandleBackendUpdateError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setTestPasswordResetUsername(t, mock)

				mock.Ctx.Request.SetBodyString(`{"password":"password123"}`)

				mock.UserProviderMock.
					EXPECT().
					UpdatePassword(testUsername, "password123").
					Return(fmt.Errorf("failed to update"))
			},
			`{"status":"KO","message":"Unable to reset your password."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred updating the user password", "failed to update")
			},
		},
		{
			"ShouldHandleUserDetailsError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setTestPasswordResetUsername(t, mock)

				mock.Ctx.Request.SetBodyString(`{"password":"password123"}`)

				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						UpdatePassword(testUsername, "password123").
						Return(nil),
					mock.UserProviderMock.
						EXPECT().
						GetDetails(testUsername).
						Return(nil, fmt.Errorf("failed to get details")),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred retrieving user details", "failed to get details")

				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)
				assert.Nil(t, us.PasswordResetUsername)
			},
		},
		{
			"ShouldHandleUserWithoutEmail",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setTestPasswordResetUsername(t, mock)

				mock.Ctx.Request.SetBodyString(`{"password":"password123"}`)

				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						UpdatePassword(testUsername, "password123").
						Return(nil),
					mock.UserProviderMock.
						EXPECT().
						GetDetails(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName}, nil),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred retrieving user details: user has no email address configured", nil)
			},
		},
		{
			"ShouldHandleNotificationError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setTestPasswordResetUsername(t, mock)

				mock.Ctx.Request.SetBodyString(`{"password":"password123"}`)

				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						UpdatePassword(testUsername, "password123").
						Return(nil),
					mock.UserProviderMock.
						EXPECT().
						GetDetails(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.NotifierMock.
						EXPECT().
						Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: testEmail}, "Password changed successfully", gomock.Any(), gomock.Any()).
						Return(fmt.Errorf("failed to notify")),
				)
			},
			`{"status":"OK"}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "failed to notify", nil)
			},
		},
		{
			"ShouldHandleSuccess",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setTestPasswordResetUsername(t, mock)

				mock.Ctx.Request.SetBodyString(`{"password":"password123"}`)

				gomock.InOrder(
					mock.UserProviderMock.
						EXPECT().
						UpdatePassword(testUsername, "password123").
						Return(nil),
					mock.UserProviderMock.
						EXPECT().
						GetDetails(testUsername).
						Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil),
					mock.NotifierMock.
						EXPECT().
						Send(mock.Ctx, mail.Address{Name: testDisplayName, Address: testEmail}, "Password changed successfully", gomock.Any(), templates.EmailEventValues{
							Title:       "Password changed successfully",
							DisplayName: testDisplayName,
							RemoteIP:    "0.0.0.0",
							Details:     map[string]any{"Action": "Password Reset"},
							BodyPrefix:  eventEmailActionPasswordModifyPrefix,
							BodyEvent:   eventEmailActionPasswordReset,
							BodySuffix:  eventEmailActionPasswordModifySuffix,
						}).
						Return(nil),
				)
			},
			``,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()

				require.NoError(t, err)
				assert.Nil(t, us.PasswordResetUsername)
			},
		},
		{
			"ShouldHandleGetSessionError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
			},
			`{"status":"KO","message":"Unable to reset your password."}`,
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), errStrUserSessionData, "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.PasswordPolicy = middlewares.NewPasswordPolicyProvider(schema.PasswordPolicy{})

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			ResetPasswordPOST(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestIdentityRetrieverFromStorage(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected *session.Identity
		err      string
	}{
		{
			"ShouldHandleBodyParseError",
			nil,
			nil,
			"unexpected end of JSON input",
		},
		{
			"ShouldHandleUserProviderError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.SetBodyString(`{"username":"john"}`)

				mock.UserProviderMock.
					EXPECT().
					GetDetails(testUsername).
					Return(nil, fmt.Errorf("failed to get details"))
			},
			nil,
			"failed to get details",
		},
		{
			"ShouldHandleUserWithoutEmail",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.SetBodyString(`{"username":"john"}`)

				mock.UserProviderMock.
					EXPECT().
					GetDetails(testUsername).
					Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName}, nil)
			},
			nil,
			"user john has no email address configured",
		},
		{
			"ShouldReturnIdentity",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.SetBodyString(`{"username":"john"}`)

				mock.UserProviderMock.
					EXPECT().
					GetDetails(testUsername).
					Return(&authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}, nil)
			},
			&session.Identity{Username: testUsername, DisplayName: testDisplayName, Email: testEmail},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			identity, err := identityRetrieverFromStorage(mock.Ctx)

			if tc.err == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, identity)
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, identity)
			}
		})
	}
}

func TestResetPasswordIdentityVerificationFinish(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	resetPasswordIdentityVerificationFinish(mock.Ctx, testUsername)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK"}`, string(mock.Ctx.Response.Body()))

	us, err := mock.Ctx.GetSession()

	require.NoError(t, err)
	require.NotNil(t, us.PasswordResetUsername)
	assert.Equal(t, testUsername, *us.PasswordResetUsername)
}

func TestResetPasswordIdentityVerificationFinishShouldHandleGetSessionError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")

	resetPasswordIdentityVerificationFinish(mock.Ctx, testUsername)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK"}`, string(mock.Ctx.Response.Body()))

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Unable to get session to enable password reset in session for user ''", "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
}

func newTestIdentityVerificationClaim(t *testing.T, mock *mocks.MockAutheliaCtx, jti uuid.UUID, action string) *model.IdentityVerificationClaim {
	t.Helper()

	var (
		issuerURL *url.URL
		err       error
	)

	issuerURL, err = mock.Ctx.IssuerURL()

	require.NoError(t, err)

	verification := model.NewIdentityVerification(jti, testUsername, action, mock.Ctx.RemoteIP(), time.Minute*5)

	return verification.ToIdentityVerificationClaim(issuerURL)
}

func setTestIdentityVerificationTokenBody(t *testing.T, mock *mocks.MockAutheliaCtx, claims *model.IdentityVerificationClaim, method jwt.SigningMethod, secret string) {
	t.Helper()

	signed, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))

	require.NoError(t, err)

	mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"token":"%s"}`, signed))
}

func setTestPasswordResetUsername(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	us, err := mock.Ctx.GetSession()

	require.NoError(t, err)

	username := testUsername

	us.PasswordResetUsername = &username

	require.NoError(t, mock.Ctx.SaveSession(us))
}
