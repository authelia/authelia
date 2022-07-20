package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

var exampleDotComDomain = "example.com"

func TestCheckSafeRedirection_ForbiddenCall(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{
		Username:            "john",
		AuthenticationLevel: authentication.NotAuthenticated,
	})
	defer mock.Close()
	mock.Ctx.Configuration.Session.Domain = exampleDotComDomain

	mock.SetRequestBody(t, checkURIWithinDomainRequestBody{
		URI: "http://myapp.example.com",
	})

	CheckSafeRedirectionPOST(mock.Ctx)
	assert.Equal(t, 401, mock.Ctx.Response.StatusCode())
}

func TestCheckSafeRedirection_UnsafeRedirection(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{
		Username:            "john",
		AuthenticationLevel: authentication.OneFactor,
	})
	defer mock.Close()
	mock.Ctx.Configuration.Session.Domain = exampleDotComDomain

	mock.SetRequestBody(t, checkURIWithinDomainRequestBody{
		URI: "http://myapp.com",
	})

	CheckSafeRedirectionPOST(mock.Ctx)
	mock.Assert200OK(t, checkURIWithinDomainResponseBody{
		OK: false,
	})
}

func TestCheckSafeRedirection_SafeRedirection(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{
		Username:            "john",
		AuthenticationLevel: authentication.OneFactor,
	})
	defer mock.Close()
	mock.Ctx.Configuration.Session.Domain = exampleDotComDomain

	mock.SetRequestBody(t, checkURIWithinDomainRequestBody{
		URI: "https://myapp.example.com",
	})

	CheckSafeRedirectionPOST(mock.Ctx)
	mock.Assert200OK(t, checkURIWithinDomainResponseBody{
		OK: true,
	})
}
