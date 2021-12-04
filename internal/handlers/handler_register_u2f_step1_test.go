package handlers

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/models"
)

type HandlerRegisterU2FStep1Suite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *HandlerRegisterU2FStep1Suite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())

	userSession := s.mock.Ctx.GetSession()
	userSession.Username = testUsername
	err := s.mock.Ctx.SaveSession(userSession)
	require.NoError(s.T(), err)
}

func (s *HandlerRegisterU2FStep1Suite) TearDownTest() {
	s.mock.Close()
}

func createToken(ctx *mocks.MockAutheliaCtx, username, action string, expiresAt time.Time) (data string, verification models.IdentityVerification) {
	verification = models.NewIdentityVerification(uuid.New(), username, action, ctx.Ctx.RemoteIP())

	verification.ExpiresAt = expiresAt

	claims := verification.ToIdentityVerificationClaim()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString([]byte(ctx.Ctx.Configuration.JWTSecret))

	return ss, verification
}

func (s *HandlerRegisterU2FStep1Suite) TestShouldRaiseWhenXForwardedProtoIsMissing() {
	token, verification := createToken(s.mock, "john", ActionU2FRegistration,
		time.Now().Add(1*time.Minute))
	s.mock.Ctx.Request.SetBodyString(fmt.Sprintf("{\"token\":\"%s\"}", token))

	s.mock.StorageMock.EXPECT().
		FindIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String())).
		Return(true, nil)

	s.mock.StorageMock.EXPECT().
		ConsumeIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String()), gomock.Eq(models.NewNullIP(s.mock.Ctx.RemoteIP()))).
		Return(nil)

	SecondFactorU2FIdentityFinish(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "missing header X-Forwarded-Proto", s.mock.Hook.LastEntry().Message)
}

func (s *HandlerRegisterU2FStep1Suite) TestShouldRaiseWhenXForwardedHostIsMissing() {
	s.mock.Ctx.Request.Header.Add("X-Forwarded-Proto", "http")
	token, verification := createToken(s.mock, "john", ActionU2FRegistration,
		time.Now().Add(1*time.Minute))
	s.mock.Ctx.Request.SetBodyString(fmt.Sprintf("{\"token\":\"%s\"}", token))

	s.mock.StorageMock.EXPECT().
		FindIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String())).
		Return(true, nil)

	s.mock.StorageMock.EXPECT().
		ConsumeIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String()), gomock.Eq(models.NewNullIP(s.mock.Ctx.RemoteIP()))).
		Return(nil)

	SecondFactorU2FIdentityFinish(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "missing header X-Forwarded-Host", s.mock.Hook.LastEntry().Message)
}

func TestShouldRunHandlerRegisterU2FStep1Suite(t *testing.T) {
	suite.Run(t, new(HandlerRegisterU2FStep1Suite))
}
