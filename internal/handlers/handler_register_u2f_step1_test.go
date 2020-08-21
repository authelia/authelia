package handlers

import (
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/mocks"
)

type HandlerRegisterU2FStep1Suite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *HandlerRegisterU2FStep1Suite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())

	userSession := s.mock.Ctx.GetSession()
	userSession.Username = testUsername
	s.mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
}

func (s *HandlerRegisterU2FStep1Suite) TearDownTest() {
	s.mock.Close()
}

func createToken(secret string, username string, action string, expiresAt time.Time) string {
	claims := &middlewares.IdentityVerificationClaim{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			Issuer:    "Authelia",
		},
		Action:   action,
		Username: username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString([]byte(secret))

	return ss
}

func (s *HandlerRegisterU2FStep1Suite) TestShouldRaiseWhenXForwardedProtoIsMissing() {
	token := createToken(s.mock.Ctx.Configuration.JWTSecret, "john", U2FRegistrationAction,
		time.Now().Add(1*time.Minute))
	s.mock.Ctx.Request.SetBodyString(fmt.Sprintf("{\"token\":\"%s\"}", token))

	s.mock.StorageProviderMock.EXPECT().
		FindIdentityVerificationToken(gomock.Eq(token)).
		Return(true, nil)

	s.mock.StorageProviderMock.EXPECT().
		RemoveIdentityVerificationToken(gomock.Eq(token)).
		Return(nil)

	SecondFactorU2FIdentityFinish(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "Missing header X-Forwarded-Proto", s.mock.Hook.LastEntry().Message)
}

func (s *HandlerRegisterU2FStep1Suite) TestShouldRaiseWhenXForwardedHostIsMissing() {
	s.mock.Ctx.Request.Header.Add("X-Forwarded-Proto", "http")
	token := createToken(s.mock.Ctx.Configuration.JWTSecret, "john", U2FRegistrationAction,
		time.Now().Add(1*time.Minute))
	s.mock.Ctx.Request.SetBodyString(fmt.Sprintf("{\"token\":\"%s\"}", token))

	s.mock.StorageProviderMock.EXPECT().
		FindIdentityVerificationToken(gomock.Eq(token)).
		Return(true, nil)

	s.mock.StorageProviderMock.EXPECT().
		RemoveIdentityVerificationToken(gomock.Eq(token)).
		Return(nil)

	SecondFactorU2FIdentityFinish(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "Missing header X-Forwarded-Host", s.mock.Hook.LastEntry().Message)
}

func TestShouldRunHandlerRegisterU2FStep1Suite(t *testing.T) {
	suite.Run(t, new(HandlerRegisterU2FStep1Suite))
}
