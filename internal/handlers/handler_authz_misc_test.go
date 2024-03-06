package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestAuthzImplementation(t *testing.T) {
	assert.Equal(t, "Legacy", AuthzImplLegacy.String())
	assert.Equal(t, "", AuthzImplementation(-1).String())
}

func TestFriendlyMethod(t *testing.T) {
	assert.Equal(t, "unknown", friendlyMethod(""))
	assert.Equal(t, "GET", friendlyMethod(fasthttp.MethodGet))
}

func TestGenerateVerifySessionHasUpToDateProfileTraceLogs(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	generateVerifySessionHasUpToDateProfileTraceLogs(mock.Ctx, &session.UserSession{Username: "john", DisplayName: "example", Groups: []string{"abc"}, Emails: []string{"user@example.com", "test@example.com"}}, &authentication.UserDetails{Username: "john", Groups: []string{"123"}, DisplayName: "notexample", Emails: []string{"notuser@example.com"}})
	generateVerifySessionHasUpToDateProfileTraceLogs(mock.Ctx, &session.UserSession{Username: "john", DisplayName: "example"}, &authentication.UserDetails{Username: "john", DisplayName: "example"})
	generateVerifySessionHasUpToDateProfileTraceLogs(mock.Ctx, &session.UserSession{Username: "john", DisplayName: "example", Emails: []string{"abc@example.com"}}, &authentication.UserDetails{Username: "john", DisplayName: "example"})
	generateVerifySessionHasUpToDateProfileTraceLogs(mock.Ctx, &session.UserSession{Username: "john", DisplayName: "example"}, &authentication.UserDetails{Username: "john", DisplayName: "example", Emails: []string{"abc@example.com"}})
}
