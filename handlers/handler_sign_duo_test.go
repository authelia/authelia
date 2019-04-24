package handlers

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/clems4ever/authelia/duo"
	"github.com/clems4ever/authelia/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SecondFactorDuoPostSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *SecondFactorDuoPostSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = "john"
	s.mock.Ctx.SaveSession(userSession)
}

func (s *SecondFactorDuoPostSuite) TearDownTest() {
	s.mock.Close()
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoAPIAndAllowAccess() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	values := url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "auto")
	values.Set("pushinfo", "target%20url=https://target.example.com")

	response := duo.Response{}
	response.Response.Result = "allow"

	duoMock.EXPECT().Call(gomock.Eq(values)).Return(&response, nil)

	s.mock.Ctx.Request.SetBodyString("{\"targetURL\": \"https://target.example.com\"}")

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)

	assert.Equal(s.T(), s.mock.Ctx.Response.StatusCode(), 200)
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoAPIAndDenyAccess() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	values := url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "auto")
	values.Set("pushinfo", "target%20url=https://target.example.com")

	response := duo.Response{}
	response.Response.Result = "deny"

	duoMock.EXPECT().Call(gomock.Eq(values)).Return(&response, nil)

	s.mock.Ctx.Request.SetBodyString("{\"targetURL\": \"https://target.example.com\"}")

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)

	assert.Equal(s.T(), s.mock.Ctx.Response.StatusCode(), 401)
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoAPIAndFail() {
	duoMock := mocks.NewMockAPI(s.mock.Ctrl)

	values := url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "auto")
	values.Set("pushinfo", "target%20url=https://target.example.com")

	duoMock.EXPECT().Call(gomock.Eq(values)).Return(nil, fmt.Errorf("Connnection error"))

	s.mock.Ctx.Request.SetBodyString("{\"targetURL\": \"https://target.example.com\"}")

	SecondFactorDuoPost(duoMock)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Authentication failed, please retry later.")
}

func TestRunSecondeFactorDuoPostSuite(t *testing.T) {
	s := new(SecondFactorDuoPostSuite)
	suite.Run(t, s)
}
