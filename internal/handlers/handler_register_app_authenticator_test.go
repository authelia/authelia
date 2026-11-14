// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers_test

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/appauthenticator"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/suites/appdevice"
)

const (
	appPortalHost      = "login.example.com:8080"
	appPortalBaseURL   = "https://login.example.com:8080"
	appPortalQRPrefix  = "authelia-app://login.example.com:8080/api/register/app/"
	appPortalCookie    = "authelia_session"
	appPortalUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0"
)

type appPortal struct {
	t    *testing.T
	env  *appTestEnv
	mock *mocks.MockAutheliaCtx
}

type appPortalSession struct {
	p      *appPortal
	cookie string
}

type appPortalResponse struct {
	Status  string          `json:"status"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func appPortalProxy(addr string) *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
			req.Header.Set("X-Forwarded-Host", req.URL.Host)
			req.URL.Scheme = "http"
			req.URL.Host = addr
			req.Host = addr

			return http.DefaultTransport.RoundTrip(req)
		}),
	}
}

func newAppPortal(t *testing.T, opts ...appTestEnvOption) *appPortal {
	_, loopback, err := net.ParseCIDR("127.0.0.1/32")
	require.NoError(t, err)

	env := newAppTestEnv(t, append([]appTestEnvOption{func(o *appTestEnvOptions) {
		o.proxies = schema.ServerHeadersTrustedProxies{Enable: true, Networks: []*net.IPNet{loopback}}
		o.client = appPortalProxy
	}}, opts...)...)

	mock := mocks.NewMockAutheliaCtx(t)

	t.Cleanup(mock.Close)

	mock.Ctx.Configuration.AppAuthenticator = env.config.AppAuthenticator
	mock.Ctx.Configuration.Regulation = env.config.Regulation
	mock.Ctx.Providers.StorageProvider = env.storage
	mock.Ctx.Providers.Clock = env.clock
	mock.Ctx.Providers.Regulator = regulation.NewRegulator(env.config.Regulation, env.logs, env.clock)

	return &appPortal{t: t, env: env, mock: mock}
}

func (p *appPortal) login(username string) *appPortalSession {
	s := &appPortalSession{p: p}

	s.do(fasthttp.MethodGet, nil, func(ctx *middlewares.AutheliaCtx) {
		userSession, err := ctx.GetSession()
		require.NoError(p.t, err)

		userSession.Username = username
		userSession.SetOneFactorPassword(p.env.clock.Now(), false)

		require.NoError(p.t, ctx.SaveSession(&userSession))
	})

	require.NotEmpty(p.t, s.cookie)

	return s
}

func (s *appPortalSession) do(method string, body any, handler middlewares.RequestHandler, values ...string) *middlewares.AutheliaCtx {
	s.p.t.Helper()

	req := &fasthttp.Request{}

	req.Header.SetMethod(method)
	req.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	req.Header.Set(fasthttp.HeaderXForwardedHost, appPortalHost)
	req.Header.Set(fasthttp.HeaderXForwardedFor, "198.51.100.4")
	req.Header.Set(fasthttp.HeaderUserAgent, appPortalUserAgent)

	if s.cookie != "" {
		req.Header.SetCookie(appPortalCookie, s.cookie)
	}

	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(s.p.t, err)

		req.SetBody(data)
	}

	request := &fasthttp.RequestCtx{}

	request.Init(req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 40000}, nil)

	ctx := middlewares.NewAutheliaCtx(request, s.p.mock.Ctx.Configuration, s.p.mock.Ctx.Providers)
	ctx.Logger = s.p.mock.Ctx.Logger

	for i := 0; i+1 < len(values); i += 2 {
		ctx.SetUserValue(values[i], values[i+1])
	}

	handler(ctx)

	s.cookie = string(ctx.Request.Header.Cookie(appPortalCookie))

	return ctx
}

func (s *appPortalSession) session() session.UserSession {
	s.p.t.Helper()

	var userSession session.UserSession

	s.do(fasthttp.MethodGet, nil, func(ctx *middlewares.AutheliaCtx) {
		var err error

		userSession, err = ctx.GetSession()
		require.NoError(s.p.t, err)
	})

	return userSession
}

func (s *appPortalSession) binding() []byte {
	userSession := s.session()

	require.NotNil(s.p.t, userSession.AppAuthenticator)
	require.NotEmpty(s.p.t, userSession.AppAuthenticator.Binding)

	return appauthenticator.HashID(userSession.AppAuthenticator.Binding)
}

func appPortalOK(t *testing.T, ctx *middlewares.AutheliaCtx, data any) {
	t.Helper()

	require.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode(), string(ctx.Response.Body()))

	response := appPortalResponse{}

	require.NoError(t, json.Unmarshal(ctx.Response.Body(), &response), string(ctx.Response.Body()))
	require.Equal(t, "OK", response.Status, string(ctx.Response.Body()))

	if data != nil {
		require.NoError(t, json.Unmarshal(response.Data, data), string(ctx.Response.Body()))
	}
}

func appPortalKO(t *testing.T, ctx *middlewares.AutheliaCtx, status int) {
	t.Helper()

	assert.Equal(t, status, ctx.Response.StatusCode(), string(ctx.Response.Body()))

	response := appPortalResponse{}

	require.NoError(t, json.Unmarshal(ctx.Response.Body(), &response), string(ctx.Response.Body()))
	assert.Equal(t, "KO", response.Status)
}

type appRegistrationPUTData struct {
	QR        string `json:"qr"`
	ExpiresAt int64  `json:"expires_at"`
}

type appRegistrationGETData map[string]any

func (s *appPortalSession) registrationPUT() appRegistrationPUTData {
	s.p.t.Helper()

	data := appRegistrationPUTData{}

	appPortalOK(s.p.t, s.do(fasthttp.MethodPut, nil, handlers.AppAuthenticatorRegistrationPUT), &data)

	return data
}

func (s *appPortalSession) registrationGET() appRegistrationGETData {
	s.p.t.Helper()

	data := appRegistrationGETData{}

	appPortalOK(s.p.t, s.do(fasthttp.MethodGet, nil, handlers.AppAuthenticatorRegistrationGET), &data)

	return data
}

func TestAppAuthenticatorRegistrationShouldRegisterDevice(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")

	now := p.env.clock.Now()

	created := s.registrationPUT()
	assert.True(t, strings.HasPrefix(created.QR, appPortalQRPrefix), created.QR)
	assert.Regexp(t, `^[A-Za-z0-9_-]{43}$`, strings.TrimPrefix(created.QR, appPortalQRPrefix))
	assert.Equal(t, now.Add(5*time.Minute).Unix(), created.ExpiresAt)

	assert.Equal(t, appRegistrationGETData{"state": "waiting", "expires_at": float64(now.Add(5 * time.Minute).Unix())}, s.registrationGET())

	d := p.env.device(appdevice.WithInfo(appdevice.Info{Name: "Jane's iPhone", Platform: "iOS", Model: "iPhone15,2"}))

	begin, err := d.RegisterBegin(created.QR)
	require.NoError(t, err)
	assert.Equal(t, appPortalBaseURL, begin.BaseURL)

	confirm := s.registrationGET()
	assert.Equal(t, "confirm_number", confirm["state"])
	assert.Equal(t, float64(2), confirm["nonce_length"])
	assert.Equal(t, float64(begin.ExpiresAt), confirm["expires_at"])
	assert.Equal(t, map[string]any{"platform": "iOS", "model": "iPhone15,2", "name": "Jane's iPhone"}, confirm["device"])

	number, ok := confirm["number"].(string)
	require.True(t, ok, "number must be a JSON string")
	assert.Regexp(t, `^[0-9]{2}$`, number)

	require.NoError(t, d.RegisterConfirm(number))
	assert.Equal(t, "john", d.Username)
	assert.Equal(t, appPortalBaseURL, d.BaseURL)

	device, err := p.env.storage.LoadAppAuthenticatorDeviceByDeviceID(t.Context(), d.DeviceID)
	require.NoError(t, err)

	assert.Equal(t, appRegistrationGETData{"state": "registered", "id": float64(device.ID)}, s.registrationGET())
	assert.Equal(t, "john", device.Username)
	assert.Equal(t, appPortalBaseURL, device.BaseURL)

	registration, err := p.env.storage.LoadAppAuthenticatorRegistrationByPortalBinding(t.Context(), s.binding())
	require.NoError(t, err)
	assert.Equal(t, "john", registration.Username)
	assert.Equal(t, model.AppAuthenticatorRegistrationStateCompleted, registration.State)
}

func TestAppAuthenticatorRegistrationShouldUseBasePath(t *testing.T) {
	p := newAppPortal(t)

	p.mock.Ctx.Configuration.Session.Cookies[0].AutheliaURL = &url.URL{Scheme: "https", Host: appPortalHost, Path: "/authelia/"}
	p.mock.ResetSessionProvider()

	s := p.login("john")

	created := s.registrationPUT()
	assert.True(t, strings.HasPrefix(created.QR, "authelia-app://login.example.com:8080/authelia/api/register/app/"), created.QR)
}

func TestAppAuthenticatorRegistrationShouldFailWithoutTrustedIdentities(t *testing.T) {
	p := newAppPortal(t, withAppMode(schema.AppAuthenticatorAttestationModeRequired))
	s := p.login("john")

	appPortalKO(t, s.do(fasthttp.MethodPut, nil, handlers.AppAuthenticatorRegistrationPUT), fasthttp.StatusForbidden)

	assert.Nil(t, s.session().AppAuthenticator)
}

func TestAppAuthenticatorRegistrationShouldFailWithInvalidPortalURL(t *testing.T) {
	p := newAppPortal(t)

	p.mock.Ctx.Configuration.Session.Cookies[0].AutheliaURL = &url.URL{Scheme: "https", Host: appPortalHost, Path: "/a b/"}
	p.mock.ResetSessionProvider()

	s := p.login("john")

	appPortalKO(t, s.do(fasthttp.MethodPut, nil, handlers.AppAuthenticatorRegistrationPUT), fasthttp.StatusInternalServerError)

	entry := p.mock.Hook.LastEntry()
	require.NotNil(t, entry)
	assert.Contains(t, entry.Message, "configuration")
}

func TestAppAuthenticatorRegistrationShouldCancelPreviousCode(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")

	first := s.registrationPUT()
	binding := s.binding()

	second := s.registrationPUT()
	assert.NotEqual(t, first.QR, second.QR)
	assert.Equal(t, binding, s.binding())

	_, err := p.env.device().RegisterBegin(first.QR)
	requireHTTPError(t, err, http.StatusNotFound, "unknown")

	d := p.env.device()

	_, err = d.RegisterBegin(second.QR)
	require.NoError(t, err)

	s.registrationPUT()

	assert.Equal(t, "waiting", s.registrationGET()["state"])

	err = d.RegisterConfirm("00")
	requireHTTPError(t, err, http.StatusConflict, "already_used")

	_, err = p.env.device().RegisterBegin(second.QR)
	requireHTTPError(t, err, http.StatusConflict, "already_used")
}

func TestAppAuthenticatorRegistrationShouldReportStates(t *testing.T) {
	t.Run("ShouldReportFailedWhenVoided", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")

		created := s.registrationPUT()

		d := p.env.device()

		_, err := d.RegisterBegin(created.QR)
		require.NoError(t, err)

		number, _ := s.registrationGET()["number"].(string)

		err = d.RegisterConfirm(wrongNumber(number))
		requireHTTPError(t, err, http.StatusUnauthorized, "bad_signature")

		assert.Equal(t, appRegistrationGETData{"state": "failed"}, s.registrationGET())
	})

	t.Run("ShouldReportExpiredCode", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")

		s.registrationPUT()

		p.env.clock.Advance(5*time.Minute + time.Second)

		assert.Equal(t, appRegistrationGETData{"state": "expired"}, s.registrationGET())
	})

	t.Run("ShouldReportExpiredSession", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")

		created := s.registrationPUT()

		_, err := p.env.device().RegisterBegin(created.QR)
		require.NoError(t, err)

		p.env.clock.Advance(5*time.Minute + time.Second)

		assert.Equal(t, appRegistrationGETData{"state": "expired"}, s.registrationGET())
	})

	t.Run("ShouldReportExpiredWhenSettled", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")

		created := s.registrationPUT()

		d := p.env.device()

		_, err := d.RegisterBegin(created.QR)
		require.NoError(t, err)

		p.env.clock.Advance(5*time.Minute + time.Second)

		err = d.RegisterConfirm("00")
		requireHTTPError(t, err, http.StatusGone, "expired")

		assert.Equal(t, appRegistrationGETData{"state": "expired"}, s.registrationGET())
	})

	t.Run("ShouldNotFindWithoutRegistration", func(t *testing.T) {
		p := newAppPortal(t)
		s := p.login("john")

		appPortalKO(t, s.do(fasthttp.MethodGet, nil, handlers.AppAuthenticatorRegistrationGET), fasthttp.StatusNotFound)
	})

	t.Run("ShouldNotShowAnotherSession", func(t *testing.T) {
		p := newAppPortal(t)
		owner := p.login("john")
		other := p.login("john")

		created := owner.registrationPUT()

		_, err := p.env.device().RegisterBegin(created.QR)
		require.NoError(t, err)

		appPortalKO(t, other.do(fasthttp.MethodGet, nil, handlers.AppAuthenticatorRegistrationGET), fasthttp.StatusNotFound)

		other.registrationPUT()

		assert.Equal(t, "confirm_number", owner.registrationGET()["state"])
		assert.Equal(t, "waiting", other.registrationGET()["state"])
	})
}

func TestAppAuthenticatorRegistrationShouldCancel(t *testing.T) {
	p := newAppPortal(t)
	s := p.login("john")

	appPortalOK(t, s.do(fasthttp.MethodDelete, nil, handlers.AppAuthenticatorRegistrationDELETE), nil)

	created := s.registrationPUT()

	appPortalOK(t, s.do(fasthttp.MethodDelete, nil, handlers.AppAuthenticatorRegistrationDELETE), nil)

	appPortalKO(t, s.do(fasthttp.MethodGet, nil, handlers.AppAuthenticatorRegistrationGET), fasthttp.StatusNotFound)

	_, err := p.env.device().RegisterBegin(created.QR)
	requireHTTPError(t, err, http.StatusNotFound, "unknown")
}

func TestAppAuthenticatorRegistrationShouldRejectAnonymous(t *testing.T) {
	p := newAppPortal(t)
	s := &appPortalSession{p: p}

	appPortalKO(t, s.do(fasthttp.MethodPut, nil, handlers.AppAuthenticatorRegistrationPUT), fasthttp.StatusForbidden)
	appPortalKO(t, s.do(fasthttp.MethodGet, nil, handlers.AppAuthenticatorRegistrationGET), fasthttp.StatusForbidden)
	appPortalKO(t, s.do(fasthttp.MethodDelete, nil, handlers.AppAuthenticatorRegistrationDELETE), fasthttp.StatusForbidden)
}

func (s *appPortalSession) registerDevice() *appdevice.Device {
	s.p.t.Helper()

	created := s.registrationPUT()

	d := s.p.env.device()

	_, err := d.RegisterBegin(created.QR)
	require.NoError(s.p.t, err)

	number, ok := s.registrationGET()["number"].(string)
	require.True(s.p.t, ok)

	require.NoError(s.p.t, d.RegisterConfirm(number))

	return d
}
