package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

var okMessageBytes = []byte("{\"status\":\"OK\"}")

func TestHealthOk(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{
		Username: "john",
		AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
			UsernameAndPassword: true,
		},
	})
	defer mock.Close()

	HealthGET(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, okMessageBytes, mock.Ctx.Response.Body())
}
