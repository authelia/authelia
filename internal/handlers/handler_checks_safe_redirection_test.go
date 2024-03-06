package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
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
			session.UserSession{CookieDomain: "example.com", AuthenticationLevel: authentication.NotAuthenticated},
			"http://myapp.example.com",
			fasthttp.StatusUnauthorized,
			false,
		},
		{
			"ShouldReturnTrueOnGoodDomain",
			session.UserSession{CookieDomain: "example.com", Username: "john", AuthenticationLevel: authentication.OneFactor},
			"https://myapp.example.com",
			fasthttp.StatusOK,
			true,
		},
		{
			"ShouldReturnFalseOnGoodDomainWithBadScheme",
			session.UserSession{CookieDomain: "example.com", Username: "john", AuthenticationLevel: authentication.OneFactor},
			"http://myapp.example.com",
			fasthttp.StatusOK,
			false,
		},
		{
			"ShouldReturnFalseOnBadDomainWithGoodScheme",
			session.UserSession{CookieDomain: "example.com", Username: "john", AuthenticationLevel: authentication.OneFactor},
			"https://myapp.notgood.com",
			fasthttp.StatusOK,
			false,
		},
		{
			"ShouldReturnFalseOnBadDomainWithBadScheme",
			session.UserSession{CookieDomain: "example.com", Username: "john", AuthenticationLevel: authentication.OneFactor},
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
		CookieDomain:        exampleDotCom,
		Username:            "john",
		AuthenticationLevel: authentication.OneFactor,
	})

	defer mock.Close()

	mock.Ctx.Configuration.Session.Domain = exampleDotCom //nolint:staticcheck

	mock.SetRequestBody(t, "not a valid json")

	CheckSafeRedirectionPOST(mock.Ctx)
	mock.Assert200KO(t, "Operation failed.")
}

func TestShouldFailOnInvalidURL(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{
		CookieDomain:        exampleDotCom,
		Username:            "john",
		AuthenticationLevel: authentication.OneFactor,
	})
	defer mock.Close()

	mock.Ctx.Configuration.Session.Domain = exampleDotCom //nolint:staticcheck

	mock.SetRequestBody(t, checkURIWithinDomainRequestBody{
		URI: "https//invalid-url",
	})

	CheckSafeRedirectionPOST(mock.Ctx)
	mock.Assert200KO(t, "Operation failed.")
}
