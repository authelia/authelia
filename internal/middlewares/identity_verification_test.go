package middlewares_test

import (
	"fmt"
	"net/mail"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

const testJWTSecret = "abc"

func newArgs(retriever func(ctx *middlewares.AutheliaCtx) (*session.Identity, error)) middlewares.IdentityVerificationStartArgs {
	return middlewares.IdentityVerificationStartArgs{
		ActionClaim:           "Claim",
		MailButtonContent:     "Register",
		MailTitle:             "Title",
		TargetEndpoint:        "/target",
		IdentityRetrieverFunc: retriever,
	}
}

func defaultRetriever(ctx *middlewares.AutheliaCtx) (*session.Identity, error) {
	return &session.Identity{
		Username: "john",
		Email:    "john@example.com",
	}, nil
}

func TestIdentityVerificationStart_ShouldPanic(t *testing.T) {
	assert.PanicsWithError(t, "identity verification requires an identity retriever", func() {
		middlewares.IdentityVerificationStart(newArgs(nil), nil)
	})
}

func TestShouldFailStartingProcessIfUserHasNoEmailAddress(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	retriever := func(ctx *middlewares.AutheliaCtx) (*session.Identity, error) {
		return nil, fmt.Errorf("User does not have any email")
	}

	middlewares.IdentityVerificationStart(newArgs(retriever), nil)(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, "User does not have any email", mock.Hook.LastEntry().Message)
}

func TestShouldFailIfJWTCannotBeSaved(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret = testJWTSecret
	mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTAlgorithm = ""

	mock.StorageMock.EXPECT().
		SaveIdentityVerification(mock.Ctx, gomock.Any()).
		Return(fmt.Errorf("cannot save"))

	args := newArgs(defaultRetriever)
	middlewares.IdentityVerificationStart(args, func(ctx *middlewares.AutheliaCtx, requestTime time.Time, successful *bool) {
		time.Sleep(time.Millisecond * 10)
	})(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, "cannot save", mock.Hook.LastEntry().Message)
}

func TestShouldFailSendingAnEmail(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret = testJWTSecret
	mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTAlgorithm = "HS512"
	mock.Ctx.Request.Header.Add(fasthttp.HeaderXForwardedProto, "http")
	mock.Ctx.Request.Header.Add(fasthttp.HeaderXForwardedHost, "host")

	mock.StorageMock.EXPECT().
		SaveIdentityVerification(mock.Ctx, gomock.Any()).
		Return(nil)

	mock.NotifierMock.EXPECT().
		Send(gomock.Eq(mock.Ctx), gomock.Eq(mail.Address{Address: "john@example.com"}), gomock.Eq("Title"), gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("no notif"))

	args := newArgs(defaultRetriever)
	middlewares.IdentityVerificationStart(args, nil)(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, "no notif", mock.Hook.LastEntry().Message)
}

func TestShouldSucceedIdentityVerificationStartProcess(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret = testJWTSecret
	mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTAlgorithm = "HS384"
	mock.Ctx.Request.Header.Add(fasthttp.HeaderXForwardedProto, "http")
	mock.Ctx.Request.Header.Add(fasthttp.HeaderXForwardedHost, "host")

	mock.StorageMock.EXPECT().
		SaveIdentityVerification(mock.Ctx, gomock.Any()).
		Return(nil)

	mock.NotifierMock.EXPECT().
		Send(gomock.Eq(mock.Ctx), gomock.Eq(mail.Address{Address: "john@example.com"}), gomock.Eq("Title"), gomock.Any(), gomock.Any()).
		Return(nil)

	args := newArgs(defaultRetriever)
	middlewares.IdentityVerificationStart(args, nil)(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	defer mock.Close()
}

func TestShouldSucceedIdentityVerificationStartProcessHS256(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret = testJWTSecret
	mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTAlgorithm = "HS256"
	mock.Ctx.Request.Header.Add(fasthttp.HeaderXForwardedProto, "http")
	mock.Ctx.Request.Header.Add(fasthttp.HeaderXForwardedHost, "host")

	mock.StorageMock.EXPECT().
		SaveIdentityVerification(mock.Ctx, gomock.Any()).
		Return(nil)

	mock.NotifierMock.EXPECT().
		Send(gomock.Eq(mock.Ctx), gomock.Eq(mail.Address{Address: "john@example.com"}), gomock.Eq("Title"), gomock.Any(), gomock.Any()).
		Return(nil)

	args := newArgs(defaultRetriever)
	middlewares.IdentityVerificationStart(args, nil)(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	defer mock.Close()
}

// Test Finish process.
type IdentityVerificationFinishProcess struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *IdentityVerificationFinishProcess) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())

	s.mock.Ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret = testJWTSecret
}

func (s *IdentityVerificationFinishProcess) TearDownTest() {
	s.mock.Close()
}

func createToken(ctx *mocks.MockAutheliaCtx, username, action string, expiresAt time.Time) (data string, verification model.IdentityVerification) {
	verification = model.NewIdentityVerification(uuid.New(), username, action, ctx.Ctx.RemoteIP(), time.Minute*5)

	verification.ExpiresAt = expiresAt

	claims := verification.ToIdentityVerificationClaim()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString([]byte(ctx.Ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret))

	return ss, verification
}

func next(ctx *middlewares.AutheliaCtx, username string) {}

func newFinishArgs() middlewares.IdentityVerificationFinishArgs {
	return middlewares.IdentityVerificationFinishArgs{
		ActionClaim:          "EXP_ACTION",
		IsTokenUserValidFunc: func(ctx *middlewares.AutheliaCtx, username string) bool { return true },
	}
}

func (s *IdentityVerificationFinishProcess) TestShouldFailIfJSONBodyIsMalformed() {
	middlewares.IdentityVerificationFinish(newFinishArgs(), next)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed")
	assert.Equal(s.T(), "unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
}

func (s *IdentityVerificationFinishProcess) TestShouldFailIfTokenIsNotProvided() {
	s.mock.Ctx.Request.SetBodyString("{}")
	middlewares.IdentityVerificationFinish(newFinishArgs(), next)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed")
	assert.Equal(s.T(), "no token provided", s.mock.Hook.LastEntry().Message)
}

func (s *IdentityVerificationFinishProcess) TestShouldFailIfTokenIsNotFoundInDB() {
	token, verification := createToken(s.mock, "john", "Login",
		time.Now().Add(1*time.Minute))

	s.mock.Ctx.Request.SetBodyString(fmt.Sprintf("{\"token\":\"%s\"}", token))

	s.mock.StorageMock.EXPECT().
		FindIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String())).
		Return(false, nil)

	middlewares.IdentityVerificationFinish(newFinishArgs(), next)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "The identity verification token has already been used")
	assert.Equal(s.T(), "Error occurred looking up identity verification during the validation phase, the token was not found in the database which could indicate it was never generated or was already used", s.mock.Hook.LastEntry().Message)
}

func (s *IdentityVerificationFinishProcess) TestShouldFailIfTokenIsInvalid() {
	s.mock.Ctx.Request.SetBodyString("{\"token\":\"abc\"}")

	middlewares.IdentityVerificationFinish(newFinishArgs(), next)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed")
	assert.Equal(s.T(), "Error occurred validating the identity verification token as it appears to be malformed, this potentially can occur if you've not copied the full link", s.mock.Hook.LastEntry().Message)
}

func (s *IdentityVerificationFinishProcess) TestShouldFailIfTokenExpired() {
	args := newArgs(defaultRetriever)
	token, _ := createToken(s.mock, "john", args.ActionClaim,
		time.Now().Add(-1*time.Minute))
	s.mock.Ctx.Request.SetBodyString(fmt.Sprintf("{\"token\":\"%s\"}", token))

	middlewares.IdentityVerificationFinish(newFinishArgs(), next)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "The identity verification token has expired")
	assert.Equal(s.T(), "Error occurred validating the identity verification token validity period as it appears to be expired", s.mock.Hook.LastEntry().Message)
}

func (s *IdentityVerificationFinishProcess) TestShouldFailForWrongAction() {
	token, verification := createToken(s.mock, "", "",
		time.Now().Add(1*time.Minute))
	s.mock.Ctx.Request.SetBodyString(fmt.Sprintf("{\"token\":\"%s\"}", token))

	s.mock.StorageMock.EXPECT().
		FindIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String())).
		Return(true, nil)

	middlewares.IdentityVerificationFinish(newFinishArgs(), next)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed")
	assert.Equal(s.T(), "Error occurred handling the identity verification token, the token action '' does not match the endpoint action 'EXP_ACTION' which is not allowed", s.mock.Hook.LastEntry().Message)
}

func (s *IdentityVerificationFinishProcess) TestShouldFailForWrongUser() {
	token, verification := createToken(s.mock, "harry", "EXP_ACTION",
		time.Now().Add(1*time.Minute))
	s.mock.Ctx.Request.SetBodyString(fmt.Sprintf("{\"token\":\"%s\"}", token))

	s.mock.StorageMock.EXPECT().
		FindIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String())).
		Return(true, nil)

	args := newFinishArgs()
	args.IsTokenUserValidFunc = func(ctx *middlewares.AutheliaCtx, username string) bool { return false }
	middlewares.IdentityVerificationFinish(args, next)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed")
	assert.Equal(s.T(), "Error occurred handling the identity verification token, the user is not allowed to use this token", s.mock.Hook.LastEntry().Message)
}

func (s *IdentityVerificationFinishProcess) TestShouldFailIfTokenCannotBeRemovedFromDB() {
	token, verification := createToken(s.mock, "john", "EXP_ACTION",
		time.Now().Add(1*time.Minute))
	s.mock.Ctx.Request.SetBodyString(fmt.Sprintf("{\"token\":\"%s\"}", token))

	s.mock.StorageMock.EXPECT().
		FindIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String())).
		Return(true, nil)

	s.mock.StorageMock.EXPECT().
		ConsumeIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String()), gomock.Eq(model.NewNullIP(s.mock.Ctx.RemoteIP()))).
		Return(fmt.Errorf("cannot remove"))

	middlewares.IdentityVerificationFinish(newFinishArgs(), next)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed")
	assert.Equal(s.T(), "Error occurred consuming the identity verification during the validation phase", s.mock.Hook.LastEntry().Message)
}

func (s *IdentityVerificationFinishProcess) TestShouldReturn200OnFinishComplete() {
	token, verification := createToken(s.mock, "john", "EXP_ACTION",
		time.Now().Add(1*time.Minute))
	s.mock.Ctx.Request.SetBodyString(fmt.Sprintf("{\"token\":\"%s\"}", token))

	s.mock.StorageMock.EXPECT().
		FindIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String())).
		Return(true, nil)

	s.mock.StorageMock.EXPECT().
		ConsumeIdentityVerification(s.mock.Ctx, gomock.Eq(verification.JTI.String()), gomock.Eq(model.NewNullIP(s.mock.Ctx.RemoteIP()))).
		Return(nil)

	middlewares.IdentityVerificationFinish(newFinishArgs(), next)(s.mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
}

func TestRunIdentityVerificationFinish(t *testing.T) {
	s := new(IdentityVerificationFinishProcess)
	suite.Run(t, s)
}
