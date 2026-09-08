package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestCheckSafeRedirection(t *testing.T) {
	testCases := []struct {
		name        string
		userSession session.UserSession
		have        string
		expected    int
		ok          bool
	}{
		{
			"ShouldReturnUnauthorized",
			session.UserSession{CookieDomain: "example.com"},
			"http://myapp.example.com",
			fasthttp.StatusUnauthorized,
			false,
		},
		{
			"ShouldReturnTrueOnGoodDomain",
			session.UserSession{CookieDomain: "example.com", Username: "john", AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{UsernameAndPassword: true}},
			"https://myapp.example.com",
			fasthttp.StatusOK,
			true,
		},
		{
			"ShouldReturnFalseOnGoodDomainWithBadScheme",
			session.UserSession{CookieDomain: "example.com", Username: "john", AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{UsernameAndPassword: true}},
			"http://myapp.example.com",
			fasthttp.StatusOK,
			false,
		},
		{
			"ShouldReturnFalseOnBadDomainWithGoodScheme",
			session.UserSession{CookieDomain: "example.com", Username: "john", AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{UsernameAndPassword: true}},
			"https://myapp.notgood.com",
			fasthttp.StatusOK,
			false,
		},
		{
			"ShouldReturnFalseOnBadDomainWithBadScheme",
			session.UserSession{CookieDomain: "example.com", Username: "john", AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{UsernameAndPassword: true}},
			"http://myapp.notgood.com",
			fasthttp.StatusOK,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, tc.userSession)
			defer mock.Close()

			mock.SetRequestBody(t, checkURIWithinDomainRequestBody{
				URI: tc.have,
			})

			CheckSafeRedirectionPOST(mock.Ctx)

			assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())

			if tc.expected == fasthttp.StatusOK {
				mock.Assert200OK(t, checkURIWithinDomainResponseBody{
					OK: tc.ok,
				})
			}
		})
	}
}

func TestShouldFailOnInvalidBody(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{
		CookieDomain: exampleDotCom,
		Username:     "john",
		AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
			UsernameAndPassword: true,
		},
	})

	defer mock.Close()

	mock.Ctx.Configuration.Session.Domain = exampleDotCom //nolint:staticcheck

	mock.SetRequestBody(t, "not a valid json")

	CheckSafeRedirectionPOST(mock.Ctx)
	mock.Assert200KO(t, "Operation failed.")
	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing the safe redirection request body", "unable to parse body: json: cannot unmarshal string into Go value of type handlers.checkURIWithinDomainRequestBody")
}

func TestShouldFailOnGetSessionError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")

	CheckSafeRedirectionPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
}

func TestShouldFailOnInvalidURL(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{
		CookieDomain: exampleDotCom,
		Username:     "john",
		AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
			UsernameAndPassword: true,
		},
	})
	defer mock.Close()

	mock.Ctx.Configuration.Session.Domain = exampleDotCom //nolint:staticcheck

	mock.SetRequestBody(t, checkURIWithinDomainRequestBody{
		URI: "https//invalid-url",
	})

	CheckSafeRedirectionPOST(mock.Ctx)
	mock.Assert200KO(t, "Operation failed.")
	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred determining if the URI 'https//invalid-url' is safe to redirect to as it could not be parsed", "parse \"https//invalid-url\": invalid URI for request")
}
