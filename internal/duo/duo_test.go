package duo_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	. "github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestAPIImpl_Call(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mock := mocks.NewMockDuoBaseProvider(ctrl)

	impl := NewDuoAPI(mock)

	assert.NotNil(t, impl.BaseProvider)
}

func TestDuoProvider_PreAuthCall_JSONError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := mocks.NewMockDuoBaseProvider(ctrl)
	provider := NewDuoAPI(base)
	assert.NotNil(t, provider.BaseProvider)

	ctx := NewTestCtx(net.ParseIP("127.0.0.1"))

	gomock.InOrder(
		base.EXPECT().
			SignedCall(fasthttp.MethodPost, "/auth/v2/preauth", url.Values{"username": {"Jane"}}).
			Return(&http.Response{}, []byte(""), nil),
	)

	response, err := provider.PreAuthCall(ctx, &session.UserSession{Username: "Jane"}, url.Values{"username": {"Jane"}})

	assert.EqualError(t, err, "error occurred making the preauth call to the duo api: error occurred parsing response: unexpected end of JSON input")
	assert.Nil(t, response)
}

func TestDuoProvider_PreAuthCall_APIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := mocks.NewMockDuoBaseProvider(ctrl)
	provider := NewDuoAPI(base)
	assert.NotNil(t, provider.BaseProvider)

	ctx := NewTestCtx(net.ParseIP("127.0.0.1"))

	gomock.InOrder(
		base.EXPECT().
			SignedCall(fasthttp.MethodPost, "/auth/v2/preauth", url.Values{"username": {"Jane"}}).
			Return(nil, nil, fmt.Errorf("uguu")),
	)

	response, err := provider.PreAuthCall(ctx, &session.UserSession{Username: "Jane"}, url.Values{"username": {"Jane"}})

	assert.EqualError(t, err, "error occurred making the preauth call to the duo api: error occurred making signed call: uguu")
	assert.Nil(t, response)
}

func TestDuoProvider_PreAuthCall_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := mocks.NewMockDuoBaseProvider(ctrl)
	provider := NewDuoAPI(base)
	assert.NotNil(t, provider.BaseProvider)

	ctx := NewTestCtx(net.ParseIP("127.0.0.1"))

	gomock.InOrder(
		base.EXPECT().
			SignedCall(fasthttp.MethodPost, "/auth/v2/preauth", url.Values{"username": {"Jane"}}).
			Return(nil, []byte(`{"stat":"FAIL","message":"There was a mock in the way of this request being successful","message_detail":"This failed intentionally to prove this code path works.","code":400}`), nil),
	)

	response, err := provider.PreAuthCall(ctx, &session.UserSession{Username: "Jane"}, url.Values{"username": {"Jane"}})

	assert.EqualError(t, err, "error occurred making the preauth call to the duo api: failure status was returned")
	assert.Nil(t, response)
}

func TestDuoProvider_PreAuthCall_Unknown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := mocks.NewMockDuoBaseProvider(ctrl)
	provider := NewDuoAPI(base)
	assert.NotNil(t, provider.BaseProvider)

	ctx := NewTestCtx(net.ParseIP("127.0.0.1"))

	gomock.InOrder(
		base.EXPECT().
			SignedCall(fasthttp.MethodPost, "/auth/v2/preauth", url.Values{"username": {"Jane"}}).
			Return(nil, []byte(`{"stat":"F","message":"There was a mock in the way of this request being successful","message_detail":"This failed intentionally to prove this code path works.","code":400}`), nil),
	)

	response, err := provider.PreAuthCall(ctx, &session.UserSession{Username: "Jane"}, url.Values{"username": {"Jane"}})

	assert.EqualError(t, err, "error occurred making the preauth call to the duo api: unknown status was returned")
	assert.Nil(t, response)
}

func TestDuoProvider_PreAuthCall_OKWithBadStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := mocks.NewMockDuoBaseProvider(ctrl)
	provider := NewDuoAPI(base)
	assert.NotNil(t, provider.BaseProvider)

	ctx := NewTestCtx(net.ParseIP("127.0.0.1"))

	gomock.InOrder(
		base.EXPECT().
			SignedCall(fasthttp.MethodPost, "/auth/v2/preauth", url.Values{"username": {"Jane"}}).
			Return(nil, []byte(`{"stat":"OK","message":"All Good!","message_detail":"Great job.","code":401}`), nil),
	)

	response, err := provider.PreAuthCall(ctx, &session.UserSession{Username: "Jane"}, url.Values{"username": {"Jane"}})

	assert.EqualError(t, err, "error occurred making the preauth call to the duo api: failure status code was returned")
	assert.Nil(t, response)
}

func TestDuoProvider_PreAuthCall_OKWithoutResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	base := mocks.NewMockDuoBaseProvider(ctrl)
	provider := NewDuoAPI(base)
	assert.NotNil(t, provider.BaseProvider)

	ctx := NewTestCtx(net.ParseIP("127.0.0.1"))

	gomock.InOrder(
		base.EXPECT().
			SignedCall(fasthttp.MethodPost, "/auth/v2/preauth", url.Values{"username": {"Jane"}}).
			Return(nil, []byte(`{"stat":"OK","message":"All Good!","message_detail":"Great job.","code":200}`), nil),
	)

	response, err := provider.PreAuthCall(ctx, &session.UserSession{Username: "Jane"}, url.Values{"username": {"Jane"}})

	assert.EqualError(t, err, "error occurred parsing the duo api preauth json response: unexpected end of JSON input")
	assert.Nil(t, response)
}

func TestDuoProvider_PreAuthCall(t *testing.T) {
	testCases := []struct {
		name              string
		mockIP            string
		mockUsername      string
		mockResponse      *http.Response
		mockResponseBytes []byte
		mockErr           string
		response          *PreAuthResponse
		err               string
	}{
		{
			"ShouldHandleInitialJSONError",
			"127.0.0.1",
			"Jane",
			nil,
			[]byte(""),
			"",
			nil,
			"error occurred making the preauth call to the duo api: error occurred parsing response: unexpected end of JSON input",
		},
		{
			"ShouldHandleAPIError",
			"127.0.0.1",
			"Jane",
			nil,
			nil,
			"uguu",
			nil,
			"error occurred making the preauth call to the duo api: error occurred making signed call: uguu",
		},
		{
			"ShouldHandleAPIFAILStatus",
			"127.0.0.1",
			"Jane",
			&http.Response{StatusCode: 400},
			[]byte(`{"stat":"FAIL","message":"There was a mock in the way of this request being successful","message_detail":"This failed intentionally to prove this code path works.","code":400}`),
			"",
			nil,
			"error occurred making the preauth call to the duo api: failure status was returned",
		},
		{
			"ShouldHandleAPIUnknownStatus",
			"127.0.0.1",
			"Jane",
			&http.Response{StatusCode: 400},
			[]byte(`{"stat":"F","message":"There was a mock in the way of this request being successful","message_detail":"This failed intentionally to prove this code path works.","code":400}`),
			"",
			nil,
			"error occurred making the preauth call to the duo api: unknown status was returned",
		},
		{
			"ShouldHandleAPIOKStatusWithBadStatusCode",
			"127.0.0.1",
			"Jane",
			&http.Response{StatusCode: 401},
			[]byte(`{"stat":"OK","message":"All Good!","message_detail":"Great job.","code":401}`),
			"",
			nil,
			"error occurred making the preauth call to the duo api: failure status code was returned",
		},
		{
			"ShouldHandleAPIOKWithoutResponse",
			"127.0.0.1",
			"Jane",
			&http.Response{StatusCode: 200},
			[]byte(`{"stat":"OK","message":"All Good!","message_detail":"Great job.","code":200}`),
			"",
			nil,
			"error occurred parsing the duo api preauth json response: unexpected end of JSON input",
		},
		{
			"ShouldHandleAPIOKWithResponse",
			"127.0.0.1",
			"Jane",
			&http.Response{StatusCode: 200},
			[]byte(`{"stat":"OK","code":200,"response":{"devices":[{"capabilities":["auto","push","sms","phone","mobile_otp"],"device":"DPFZRS9FB0D46QFTM891","display_name":"iOS (XXX-XXX-0100)","name":"","number":"XXX-XXX-0100","type":"phone"},{"device":"DHEKH0JJIYC1LX3AZWO4","name":"0","type":"token"}],"result":"auth","status_msg":"Account is active"}}`),
			"",
			&PreAuthResponse{
				Result:        "auth",
				StatusMessage: "Account is active",
				Devices: []Device{
					{
						Capabilities: []string{"auto", "push", "sms", "phone", "mobile_otp"},
						Device:       "DPFZRS9FB0D46QFTM891",
						DisplayName:  "iOS (XXX-XXX-0100)",
						Number:       "XXX-XXX-0100",
						Type:         "phone",
					},
					{
						Device: "DHEKH0JJIYC1LX3AZWO4",
						Name:   "0",
						Type:   "token",
					},
				},
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			base := mocks.NewMockDuoBaseProvider(ctrl)
			provider := NewDuoAPI(base)
			assert.NotNil(t, provider.BaseProvider)

			ctx := NewTestCtx(net.ParseIP(tc.mockIP))

			var mockErr error

			if tc.mockErr != "" {
				mockErr = errors.New(tc.mockErr)
			}

			base.EXPECT().
				SignedCall(fasthttp.MethodPost, "/auth/v2/preauth", url.Values{"username": {tc.mockUsername}}).
				Return(tc.mockResponse, tc.mockResponseBytes, mockErr)

			response, err := provider.PreAuthCall(ctx, &session.UserSession{Username: tc.mockUsername}, url.Values{"username": {tc.mockUsername}})

			assert.Equal(t, tc.response, response)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDuoProvider_AuthCall(t *testing.T) {
	testCases := []struct {
		name              string
		mockIP            string
		mockUsername      string
		mockResponse      *http.Response
		mockResponseBytes []byte
		mockErr           string
		response          *AuthResponse
		err               string
	}{
		{
			"ShouldHandleInitialJSONError",
			"127.0.0.1",
			"Jane",
			nil,
			[]byte(""),
			"",
			nil,
			"error occurred making the auth call to the duo api: error occurred parsing response: unexpected end of JSON input",
		},
		{
			"ShouldHandleAPIError",
			"127.0.0.1",
			"Jane",
			nil,
			nil,
			"uguu",
			nil,
			"error occurred making the auth call to the duo api: error occurred making signed call: uguu",
		},
		{
			"ShouldHandleAPIFAILStatus",
			"127.0.0.1",
			"Jane",
			&http.Response{StatusCode: 400},
			[]byte(`{"stat":"FAIL","message":"There was a mock in the way of this request being successful","message_detail":"This failed intentionally to prove this code path works.","code":400}`),
			"",
			nil,
			"error occurred making the auth call to the duo api: failure status was returned",
		},
		{
			"ShouldHandleAPIUnknownStatus",
			"127.0.0.1",
			"Jane",
			&http.Response{StatusCode: 400},
			[]byte(`{"stat":"F","message":"There was a mock in the way of this request being successful","message_detail":"This failed intentionally to prove this code path works.","code":400}`),
			"",
			nil,
			"error occurred making the auth call to the duo api: unknown status was returned",
		},
		{
			"ShouldHandleAPIOKStatusWithBadStatusCode",
			"127.0.0.1",
			"Jane",
			&http.Response{StatusCode: 401},
			[]byte(`{"stat":"OK","message":"All Good!","message_detail":"Great job.","code":401}`),
			"",
			nil,
			"error occurred making the auth call to the duo api: failure status code was returned",
		},
		{
			"ShouldHandleAPIOKWithoutResponse",
			"127.0.0.1",
			"Jane",
			&http.Response{StatusCode: 200},
			[]byte(`{"stat":"OK","message":"All Good!","message_detail":"Great job.","code":200}`),
			"",
			nil,
			"error occurred parsing the duo api auth json response: unexpected end of JSON input",
		},
		{
			"ShouldHandleAPIOKWithResponse",
			"127.0.0.1",
			"Jane",
			&http.Response{StatusCode: 200},
			[]byte(`{"stat":"OK","message":"All Good!","message_detail":"Great job.","code":200,"response":{"result":"allow","status":"allow","status_msg":"Success. Logging you in...","trusted_device_token":"REkxSzP00Ld4ddEVTRZOUlYMEl8RFVVQkdJ05HwUldRRThJR1VTNE0=||1627133735|8356ef7779bb0ec4c28ca9b04dc50493c4d2e05e"}}`),
			"",
			&AuthResponse{
				Result:             "allow",
				Status:             "allow",
				StatusMessage:      "Success. Logging you in...",
				TrustedDeviceToken: "REkxSzP00Ld4ddEVTRZOUlYMEl8RFVVQkdJ05HwUldRRThJR1VTNE0=||1627133735|8356ef7779bb0ec4c28ca9b04dc50493c4d2e05e",
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			base := mocks.NewMockDuoBaseProvider(ctrl)
			provider := NewDuoAPI(base)
			assert.NotNil(t, provider.BaseProvider)

			ctx := NewTestCtx(net.ParseIP(tc.mockIP))

			var mockErr error

			if tc.mockErr != "" {
				mockErr = errors.New(tc.mockErr)
			}

			base.EXPECT().
				SignedCall(fasthttp.MethodPost, "/auth/v2/auth", url.Values{"username": {tc.mockUsername}}).
				Return(tc.mockResponse, tc.mockResponseBytes, mockErr)

			response, err := provider.AuthCall(ctx, &session.UserSession{Username: tc.mockUsername}, url.Values{"username": {tc.mockUsername}})

			assert.Equal(t, tc.response, response)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func NewTestCtx(ip net.IP) *TestCtx {
	logger, hook := test.NewNullLogger()

	ctx := &TestCtx{
		ip:      ip,
		logger:  logger.WithFields(map[string]any{}),
		hook:    hook,
		Context: context.Background(),
	}

	return ctx
}

type TestCtx struct {
	ip     net.IP
	logger *logrus.Entry
	hook   *test.Hook

	providers middlewares.Providers

	context.Context
}

func (t *TestCtx) GetClock() clock.Provider {
	if t.providers.Clock == nil {
		t.providers.Clock = clock.New()
	}

	return t.providers.Clock
}

func (t *TestCtx) GetRandom() random.Provider {
	if t.providers.Random == nil {
		t.providers.Random = random.New()
	}

	return t.providers.Random
}

func (t *TestCtx) GetLogger() (logger *logrus.Entry) {
	return t.logger
}

func (t *TestCtx) GetProviders() (providers middlewares.Providers) {
	return t.providers
}

func (t *TestCtx) GetConfiguration() (config *schema.Configuration) {
	return nil
}

func (t *TestCtx) RemoteIP() (ip net.IP) {
	return t.ip
}
