package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

type SecondFactorDuoPostSuite struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *SecondFactorDuoPostSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())

	s.mock.Ctx.Providers.Duo = s.mock.DuoMock

	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	userSession.Username = testUsername

	s.Assert().NoError(s.mock.Ctx.SaveSession(userSession))
}

func (s *SecondFactorDuoPostSuite) TearDownTest() {
	s.mock.Close()
}

func (s *SecondFactorDuoPostSuite) TestShouldEnroll() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(nil, errors.New("no Duo device and method saved"))

	var enrollURL = "https://api-example.duosecurity.com/portal?code=1234567890ABCDEF&akey=12345ABCDEFGHIJ67890"

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = enroll
	preAuthResponse.EnrollPortalURL = enrollURL

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), DuoSignResponse{
		Result:    enroll,
		EnrollURL: enrollURL,
	})
}

func (s *SecondFactorDuoPostSuite) TestShouldAutoSelect() {
	s.mock.StorageMock.EXPECT().LoadPreferredDuoDevice(s.mock.Ctx, "john").Return(nil, errors.New("no Duo device and method saved"))

	var duoDevices = []duo.Device{
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
		{Capabilities: []string{"auto", "sms", "mobile_otp"}, Number: "+123456789****", Device: "1234567890ABCDEFGHIJ", DisplayName: "Test Device 2"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	s.mock.StorageMock.EXPECT().
		SavePreferredDuoDevice(s.mock.Ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}).
		Return(nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeDuo,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		})).
		Return(nil)

	values = url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "12345ABCDEFGHIJ67890")
	values.Set("pushinfo", "target%20url=https://target.example.com")

	authResponse := duo.AuthResponse{}
	authResponse.Result = allow

	s.mock.DuoMock.EXPECT().AuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&authResponse, nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{TargetURL: "https://target.example.com"})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
}

func (s *SecondFactorDuoPostSuite) TestShouldDenyAutoSelect() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(nil, errors.New("no Duo device and method saved"))

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = deny

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	values = url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "12345ABCDEFGHIJ67890")

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), DuoSignResponse{
		Result: deny,
	})
}

func (s *SecondFactorDuoPostSuite) TestShouldFailAutoSelect() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(nil, errors.New("no Duo device and method saved"))

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Any()).Return(nil, fmt.Errorf("Connection error"))

	bodyBytes, err := json.Marshal(bodySignDuoRequest{TargetURL: "https://target.example.com"})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)

	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")
}

func (s *SecondFactorDuoPostSuite) TestShouldDeleteOldDeviceAndEnroll() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "NOTEXISTENT", Method: "push"}, nil)

	var enrollURL = "https://api-example.duosecurity.com/portal?code=1234567890ABCDEF&akey=12345ABCDEFGHIJ67890"

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = enroll
	preAuthResponse.EnrollPortalURL = enrollURL

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	s.mock.StorageMock.EXPECT().DeletePreferredDuoDevice(s.mock.Ctx, "john").Return(nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), DuoSignResponse{
		Result:    enroll,
		EnrollURL: enrollURL,
	})
}

func (s *SecondFactorDuoPostSuite) TestShouldDeleteOldDeviceAndCallPreauthAPIWithInvalidDevicesAndEnroll() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "NOTEXISTENT", Method: "push"}, nil)

	var duoDevices = []duo.Device{
		{Capabilities: []string{"sms"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	s.mock.StorageMock.EXPECT().DeletePreferredDuoDevice(s.mock.Ctx, "john").Return(nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), DuoSignResponse{
		Result: enroll,
	})
}

func (s *SecondFactorDuoPostSuite) TestShouldUseOldDeviceAndSelect() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "NOTEXISTENT", Method: "push"}, nil)

	var duoDevices = []duo.Device{
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: "+123456789****", Device: "1234567890ABCDEFGHIJ", DisplayName: "Test Device 2"},
		{Capabilities: []string{"auto", "sms", "mobile_otp"}, Number: "+123456789****", Device: "1234567890ABCDEFGHIJ", DisplayName: "Test Device 3"},
	}

	var apiDevices = []DuoDevice{
		{Capabilities: []string{"push"}, Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
		{Capabilities: []string{"push"}, Device: "1234567890ABCDEFGHIJ", DisplayName: "Test Device 2"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), DuoDevicesResponse{Result: auth, Devices: apiDevices})
}

func (s *SecondFactorDuoPostSuite) TestShouldUseInvalidMethodAndAutoSelect() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "invalidmethod"}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeDuo,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		})).
		Return(nil)

	var duoDevices = []duo.Device{
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	s.mock.StorageMock.EXPECT().
		SavePreferredDuoDevice(s.mock.Ctx, model.DuoDevice{Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}).
		Return(nil)

	values = url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "12345ABCDEFGHIJ67890")
	values.Set("pushinfo", "target%20url=https://target.example.com")

	authResponse := duo.AuthResponse{}
	authResponse.Result = allow

	s.mock.DuoMock.EXPECT().AuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&authResponse, nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{TargetURL: "https://target.example.com"})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoPreauthAPIAndAllowAccess() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}, nil)

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = allow

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{TargetURL: "https://target.example.com"})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoPreauthAPIAndDenyAccess() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}, nil)

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = deny

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	values = url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "12345ABCDEFGHIJ67890")

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusUnauthorized, s.mock.Ctx.Response.StatusCode())
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoPreauthAPIAndFail() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}, nil)

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Any()).Return(nil, fmt.Errorf("Connection error"))

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)

	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoAPIAndDenyAccess() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: false,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeDuo,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		})).
		Return(nil)

	var duoDevices = []duo.Device{
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	values = url.Values{}
	values.Set("username", "john")
	values.Set("ipaddr", s.mock.Ctx.RemoteIP().String())
	values.Set("factor", "push")
	values.Set("device", "12345ABCDEFGHIJ67890")

	response := duo.AuthResponse{}
	response.Result = deny

	s.mock.DuoMock.EXPECT().AuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&response, nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusUnauthorized, s.mock.Ctx.Response.StatusCode())
}

func (s *SecondFactorDuoPostSuite) TestShouldCallDuoAPIAndFail() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}, nil)

	var duoDevices = []duo.Device{
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	s.mock.DuoMock.EXPECT().AuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Any()).Return(nil, fmt.Errorf("Connection error"))

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)

	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")
}

func (s *SecondFactorDuoPostSuite) TestShouldRedirectUserToDefaultURL() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeDuo,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		})).
		Return(nil)

	var duoDevices = []duo.Device{
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	response := duo.AuthResponse{}
	response.Result = allow

	s.mock.DuoMock.EXPECT().AuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Any()).Return(&response, nil)

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: testRedirectionURLString,
	})
}

func (s *SecondFactorDuoPostSuite) TestShouldNotReturnRedirectURL() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeDuo,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		})).
		Return(nil)

	var duoDevices = []duo.Device{
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	response := duo.AuthResponse{}
	response.Result = allow

	s.mock.DuoMock.EXPECT().AuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Any()).Return(&response, nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), &redirectResponse{Redirect: "https://www.example.com"})
}

func (s *SecondFactorDuoPostSuite) TestShouldRedirectUserToSafeTargetURL() {
	s.mock.Ctx.Configuration.Session.Cookies = []schema.SessionCookie{
		{
			Domain: "example.com",
		},
		{
			Domain: "mydomain.local",
		},
	}
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeDuo,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		})).
		Return(nil)

	var duoDevices = []duo.Device{
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	response := duo.AuthResponse{}
	response.Result = allow

	s.mock.DuoMock.EXPECT().AuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Any()).Return(&response, nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{
		TargetURL: "https://example.com",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: "https://example.com",
	})
}

func (s *SecondFactorDuoPostSuite) TestShouldNotRedirectToUnsafeURL() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeDuo,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		})).
		Return(nil)

	var duoDevices = []duo.Device{
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	response := duo.AuthResponse{}
	response.Result = allow

	s.mock.DuoMock.EXPECT().AuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Any()).Return(&response, nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{
		TargetURL: "http://example.com",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	DuoPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)
}

func (s *SecondFactorDuoPostSuite) TestShouldRegenerateSessionForPreventingSessionFixation() {
	s.mock.StorageMock.EXPECT().
		LoadPreferredDuoDevice(s.mock.Ctx, "john").
		Return(&model.DuoDevice{ID: 1, Username: "john", Device: "12345ABCDEFGHIJ67890", Method: "push"}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeDuo,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		})).
		Return(nil)

	var duoDevices = []duo.Device{
		{Capabilities: []string{"auto", "push", "sms", "mobile_otp"}, Number: " ", Device: "12345ABCDEFGHIJ67890", DisplayName: "Test Device 1"},
	}

	values := url.Values{}
	values.Set("username", "john")

	preAuthResponse := duo.PreAuthResponse{}
	preAuthResponse.Result = auth
	preAuthResponse.Devices = duoDevices

	s.mock.DuoMock.EXPECT().PreAuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Eq(values)).Return(&preAuthResponse, nil)

	response := duo.AuthResponse{}
	response.Result = allow

	s.mock.DuoMock.EXPECT().AuthCall(s.mock.Ctx, &session.UserSession{CookieDomain: "example.com", Username: "john"}, gomock.Any()).Return(&response, nil)

	bodyBytes, err := json.Marshal(bodySignDuoRequest{
		TargetURL: "http://example.com",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	DuoPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))
}

func TestRunSecondFactorDuoPostSuite(t *testing.T) {
	s := new(SecondFactorDuoPostSuite)
	suite.Run(t, s)
}
