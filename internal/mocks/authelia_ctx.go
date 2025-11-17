package mocks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

// MockAutheliaCtx a mock of AutheliaCtx.
type MockAutheliaCtx struct {
	// Logger hook.
	Hook *test.Hook
	Ctx  *middlewares.AutheliaCtx
	Ctrl *gomock.Controller

	// Providers.
	UserProviderMock *MockUserProvider
	StorageMock      *MockStorage
	NotifierMock     *MockNotifier
	TOTPMock         *MockTOTP
	RandomMock       *MockRandom
	DuoMock          *MockDuoProvider

	Clock clock.Fixed
}

// NewMockAutheliaCtx create an instance of AutheliaCtx mock.
func NewMockAutheliaCtx(t *testing.T) *MockAutheliaCtx {
	mockAuthelia := new(MockAutheliaCtx)
	mockAuthelia.Clock = clock.Fixed{}

	datetime, _ := time.Parse("2006-Jan-02", "2013-Feb-03")
	mockAuthelia.Clock.Set(datetime)

	config := schema.Configuration{}

	config.Session.Cookies = []schema.SessionCookie{
		{
			SessionCookieCommon: schema.SessionCookieCommon{
				Name:       "authelia_session",
				RememberMe: schema.DefaultSessionConfiguration.RememberMe,
				Expiration: schema.DefaultSessionConfiguration.Expiration,
			},
			Domain:                "example.com",
			DefaultRedirectionURL: &url.URL{Scheme: "https", Host: "www.example.com"},
		},
		{
			SessionCookieCommon: schema.SessionCookieCommon{
				Name:       "authelia_session",
				RememberMe: schema.DefaultSessionConfiguration.RememberMe,
				Expiration: schema.DefaultSessionConfiguration.Expiration,
			},
			Domain:                "example2.com",
			DefaultRedirectionURL: &url.URL{Scheme: "https", Host: "www.example2.com"},
		},
	}

	config.AuthenticationBackend.RefreshInterval = schema.NewRefreshIntervalDuration(schema.RefreshIntervalDefault)
	config.AccessControl = schema.AccessControl{
		DefaultPolicy: "deny",
		Rules: []schema.AccessControlRule{
			{
				Domains: []string{"bypass.example.com"},
				Policy:  "bypass",
			},
			{
				Domains: []string{"bypass-get.example.com"},
				Policy:  "bypass",
				Methods: []string{fasthttp.MethodGet},
			},
			{
				Domains: []string{"bypass-head.example.com"},
				Policy:  "bypass",
				Methods: []string{fasthttp.MethodHead},
			},
			{
				Domains: []string{"bypass-options.example.com"},
				Policy:  "bypass",
				Methods: []string{fasthttp.MethodOptions},
			},
			{
				Domains: []string{"bypass-trace.example.com"},
				Policy:  "bypass",
				Methods: []string{fasthttp.MethodTrace},
			},
			{
				Domains: []string{"bypass-put.example.com"},
				Policy:  "bypass",
				Methods: []string{fasthttp.MethodPut},
			},
			{
				Domains: []string{"bypass-patch.example.com"},
				Policy:  "bypass",
				Methods: []string{fasthttp.MethodPatch},
			},
			{
				Domains: []string{"bypass-post.example.com"},
				Policy:  "bypass",
				Methods: []string{fasthttp.MethodPost},
			},
			{
				Domains: []string{"bypass-delete.example.com"},
				Policy:  "bypass",
				Methods: []string{fasthttp.MethodDelete},
			},
			{
				Domains: []string{"bypass-connect.example.com"},
				Policy:  "bypass",
				Methods: []string{fasthttp.MethodConnect},
			},
			{
				Domains: []string{
					"bypass-get.example.com", "bypass-head.example.com", "bypass-options.example.com",
					"bypass-trace.example.com", "bypass-put.example.com", "bypass-patch.example.com",
					"bypass-post.example.com", "bypass-delete.example.com", "bypass-connect.example.com",
				},
				Policy: "one_factor",
			},
			{
				Domains: []string{"one-factor.example.com"},
				Policy:  "one_factor",
			},
			{
				Domains: []string{"two-factor.example.com"},
				Policy:  "two_factor",
			},
			{
				Domains: []string{"deny.example.com"},
				Policy:  "deny",
			},
			{
				Domains:  []string{"admin.example.com"},
				Policy:   "two_factor",
				Subjects: [][]string{{"group:admin"}},
			},
			{
				Domains:  []string{"grafana.example.com"},
				Policy:   "two_factor",
				Subjects: [][]string{{"group:grafana"}},
			},
			{
				Domains: []string{"bypass.example2.com"},
				Policy:  "bypass",
			},
			{
				Domains: []string{"one-factor.example2.com"},
				Policy:  "one_factor",
			},
			{
				Domains: []string{"two-factor.example2.com"},
				Policy:  "two_factor",
			},
			{
				Domains: []string{"deny.example2.com"},
				Policy:  "deny",
			},
			{
				Domains:  []string{"admin.example2.com"},
				Policy:   "two_factor",
				Subjects: [][]string{{"group:admin"}},
			},
			{
				Domains:  []string{"grafana.example2.com"},
				Policy:   "two_factor",
				Subjects: [][]string{{"group:grafana"}},
			},
		},
	}

	providers := middlewares.Providers{}

	mockAuthelia.Ctrl = gomock.NewController(t)
	mockAuthelia.UserProviderMock = NewMockUserProvider(mockAuthelia.Ctrl)
	providers.UserProvider = mockAuthelia.UserProviderMock

	mockAuthelia.StorageMock = NewMockStorage(mockAuthelia.Ctrl)
	providers.StorageProvider = mockAuthelia.StorageMock

	mockAuthelia.NotifierMock = NewMockNotifier(mockAuthelia.Ctrl)
	providers.Notifier = mockAuthelia.NotifierMock

	providers.Authorizer = authorization.NewAuthorizer(
		&config)

	providers.SessionProvider = session.NewProvider(
		config.Session, nil)

	providers.Regulator = regulation.NewRegulator(config.Regulation, providers.StorageProvider, &mockAuthelia.Clock)

	mockAuthelia.DuoMock = NewMockDuoProvider(mockAuthelia.Ctrl)

	mockAuthelia.TOTPMock = NewMockTOTP(mockAuthelia.Ctrl)
	providers.TOTP = mockAuthelia.TOTPMock

	mockAuthelia.RandomMock = NewMockRandom(mockAuthelia.Ctrl)

	providers.Random = random.NewMathematical()

	var err error
	if providers.Templates, err = templates.New(templates.Config{}); err != nil {
		panic(err)
	}

	request := &fasthttp.RequestCtx{}
	// Set a cookie to identify this client throughout the test.
	// request.Request.Header.SetCookie("authelia_session", "client_cookie").

	// Set X-Forwarded-Host for compatibility with multi-root-domain implementation.
	request.Request.Header.Set(fasthttp.HeaderXForwardedHost, "example.com")

	ctx := middlewares.NewAutheliaCtx(request, config, providers)
	mockAuthelia.Ctx = ctx

	logger, hook := test.NewNullLogger()
	mockAuthelia.Hook = hook

	mockAuthelia.Ctx.Logger = logrus.NewEntry(logger)

	return mockAuthelia
}

// NewMockAutheliaCtxWithUserSession create an instance of AutheliaCtx mock with predefined user session.
func NewMockAutheliaCtxWithUserSession(t *testing.T, userSession session.UserSession) *MockAutheliaCtx {
	t.Helper()

	mock := NewMockAutheliaCtx(t)
	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	return mock
}

// Close close the mock.
func (m *MockAutheliaCtx) Close() {
	m.Hook.Reset()
	m.Ctrl.Finish()
}

func (m *MockAutheliaCtx) SetLogLevel(level logrus.Level) {
	logger := logrus.New()
	logger.Out = io.Discard
	logger.SetLevel(level)

	m.Hook = test.NewLocal(logger)
	m.Ctx.Logger = logrus.NewEntry(logger)
}

// SetRequestBody set the request body from a struct with json tags.
func (m *MockAutheliaCtx) SetRequestBody(t *testing.T, body interface{}) {
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)
	m.Ctx.Request.SetBody(bodyBytes)
}

func (m *MockAutheliaCtx) LogEntryN(n int) *logrus.Entry {
	entries := m.Hook.AllEntries()

	if i := len(entries) - (1 + n); i < 0 {
		return nil
	} else {
		return entries[i]
	}
}

// AssertKO assert an error response from the service.
func (m *MockAutheliaCtx) AssertKO(t *testing.T, message string, code int) {
	assert.Equal(t, code, m.Ctx.Response.StatusCode())
	assert.Equal(t, fmt.Sprintf("{\"status\":\"KO\",\"message\":\"%s\"}", message), string(m.Ctx.Response.Body()))
}

// Assert401KO assert an error response from the service.
func (m *MockAutheliaCtx) Assert401KO(t *testing.T, message string) {
	m.AssertKO(t, message, fasthttp.StatusUnauthorized)
}

// Assert403KO assert an error response from the service.
func (m *MockAutheliaCtx) Assert403KO(t *testing.T, message string) {
	m.AssertKO(t, message, fasthttp.StatusForbidden)
}

// Assert404KO assert an error response from the service.
func (m *MockAutheliaCtx) Assert404KO(t *testing.T, message string) {
	m.AssertKO(t, message, fasthttp.StatusNotFound)
}

// Assert500KO assert an error response from the service.
func (m *MockAutheliaCtx) Assert500KO(t *testing.T, message string) {
	m.AssertKO(t, message, fasthttp.StatusInternalServerError)
}

// Assert200KO assert an error response from the service.
func (m *MockAutheliaCtx) Assert200KO(t *testing.T, message string) {
	m.AssertKO(t, message, fasthttp.StatusOK)
}

// Assert200OK assert a successful response from the service.
func (m *MockAutheliaCtx) Assert200OK(t *testing.T, data interface{}) {
	assert.Equal(t, fasthttp.StatusOK, m.Ctx.Response.StatusCode())

	response := middlewares.OKResponse{
		Status: "OK",
		Data:   data,
	}

	b, err := json.Marshal(response)

	assert.NoError(t, err)
	assert.Equal(t, string(b), string(m.Ctx.Response.Body()))
}

func (m *MockAutheliaCtx) AssertLastLogMessageRegexp(t *testing.T, message, err *regexp.Regexp) {
	entry := m.Hook.LastEntry()

	require.NotNil(t, entry)

	if message != nil {
		assert.Regexp(t, message, entry.Message)
	}

	v, ok := entry.Data["error"]

	if err == nil {
		assert.False(t, ok)
		assert.Nil(t, v)
	} else {
		assert.True(t, ok)
		require.NotNil(t, v)

		theErr, ok := v.(error)
		assert.True(t, ok)
		require.NotNil(t, theErr)

		assert.Regexp(t, err, theErr.Error())
	}
}

func (m *MockAutheliaCtx) AssertLastLogMessage(t *testing.T, message, err string) {
	entry := m.Hook.LastEntry()

	require.NotNil(t, entry)

	assert.Equal(t, message, entry.Message)

	v, ok := entry.Data["error"]

	if err == "" {
		assert.False(t, ok)
		assert.Nil(t, v)
	} else {
		assert.True(t, ok)
		require.NotNil(t, v)

		theErr, ok := v.(error)
		assert.True(t, ok)
		require.NotNil(t, theErr)

		assert.EqualError(t, theErr, err)
	}
}

func (m *MockAutheliaCtx) AssertLogEntryAdvanced(t *testing.T, n int, level logrus.Level, message any, fields map[string]any) {
	entry := m.LogEntryN(n)

	require.NotNil(t, entry)

	assert.Equal(t, level, entry.Level)

	AssertIsStringEqualOrRegexp(t, message, entry.Message)

	for field, expected := range fields {
		require.Contains(t, entry.Data, field)

		AssertIsStringEqualOrRegexp(t, expected, entry.Data[field])
	}
}

// GetResponseData retrieves a response from the service.
func (m *MockAutheliaCtx) GetResponseData(t *testing.T, data interface{}) {
	okResponse := middlewares.OKResponse{}
	okResponse.Data = data
	err := json.Unmarshal(m.Ctx.Response.Body(), &okResponse)
	require.NoError(t, err)
}

// GetResponseError retrieves an error response from the service.
func (m *MockAutheliaCtx) GetResponseError(t *testing.T) (errResponse middlewares.ErrorResponse) {
	err := json.Unmarshal(m.Ctx.Response.Body(), &errResponse)
	require.NoError(t, err)

	return errResponse
}

func AssertIsStringEqualOrRegexp(t *testing.T, expected any, actual any) {
	t.Helper()

	switch v := expected.(type) {
	case string:
		switch a := actual.(type) {
		case error:
			assert.EqualError(t, a, v)
		default:
			assert.Equal(t, v, a)
		}
	case *regexp.Regexp:
		switch a := actual.(type) {
		case error:
			assert.Regexp(t, v, a.Error())
		default:
			assert.Regexp(t, v, a)
		}
	case nil:
		break
	default:
		t.Fatal("Expected value must be a string or Regexp")
	}
}
