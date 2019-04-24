package mocks

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/clems4ever/authelia/regulation"
	"github.com/stretchr/testify/assert"

	"github.com/clems4ever/authelia/authorization"
	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/clems4ever/authelia/middlewares"
	"github.com/clems4ever/authelia/session"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/valyala/fasthttp"
)

// MockAutheliaCtx a mock of AutheliaCtx
type MockAutheliaCtx struct {
	// Logger hook
	Hook *test.Hook
	Ctx  *middlewares.AutheliaCtx
	Ctrl *gomock.Controller

	// Providers
	UserProviderMock    *MockUserProvider
	StorageProviderMock *MockStorageProvider
	NotifierMock        *MockNotifier

	UserSession *session.UserSession
}

// NewMockAutheliaCtx create an instance of AutheliaCtx mock
func NewMockAutheliaCtx(t *testing.T) *MockAutheliaCtx {
	mockAuthelia := new(MockAutheliaCtx)

	configuration := schema.Configuration{
		AccessControl: new(schema.AccessControlConfiguration),
	}
	configuration.Session.Name = "authelia_session"
	configuration.AccessControl.DefaultPolicy = "deny"
	configuration.AccessControl.Rules = []schema.ACLRule{schema.ACLRule{
		Domain: "bypass.example.com",
		Policy: "bypass",
	}, schema.ACLRule{
		Domain: "one-factor.example.com",
		Policy: "one_factor",
	}, schema.ACLRule{
		Domain: "two-factor.example.com",
		Policy: "two_factor",
	}, schema.ACLRule{
		Domain: "deny.example.com",
		Policy: "deny",
	}}

	providers := middlewares.Providers{}

	mockAuthelia.Ctrl = gomock.NewController(t)
	mockAuthelia.UserProviderMock = NewMockUserProvider(mockAuthelia.Ctrl)
	providers.UserProvider = mockAuthelia.UserProviderMock

	mockAuthelia.StorageProviderMock = NewMockStorageProvider(mockAuthelia.Ctrl)
	providers.StorageProvider = mockAuthelia.StorageProviderMock

	mockAuthelia.NotifierMock = NewMockNotifier(mockAuthelia.Ctrl)
	providers.Notifier = mockAuthelia.NotifierMock

	providers.Authorizer = authorization.NewAuthorizer(
		*configuration.AccessControl)

	providers.SessionProvider = session.NewProvider(
		configuration.Session)

	providers.Regulator = regulation.NewRegulator(configuration.Regulation, providers.StorageProvider)

	request := &fasthttp.RequestCtx{}
	// Set a cookie to identify this client throughout the test
	// request.Request.Header.SetCookie("authelia_session", "client_cookie")

	autheliaCtx, _ := middlewares.NewAutheliaCtx(request, configuration, providers)
	mockAuthelia.Ctx = autheliaCtx

	logger, hook := test.NewNullLogger()
	mockAuthelia.Hook = hook

	mockAuthelia.Ctx.Logger = logrus.NewEntry(logger)
	return mockAuthelia
}

// Close close the mock
func (m *MockAutheliaCtx) Close() {
	m.Hook.Reset()
}

// Assert200KO assert an error response from the service.
func (m *MockAutheliaCtx) Assert200KO(t *testing.T, message string) {
	assert.Equal(t, 200, m.Ctx.Response.StatusCode())
	assert.Equal(t, fmt.Sprintf("{\"status\":\"KO\",\"message\":\"%s\"}", message), string(m.Ctx.Response.Body()))
}

// Assert200OK assert a successful response from the service.
func (m *MockAutheliaCtx) Assert200OK(t *testing.T, data interface{}) {
	assert.Equal(t, 200, m.Ctx.Response.StatusCode())
	response := middlewares.OKResponse{
		Status: "OK",
		Data:   data,
	}

	b, err := json.Marshal(response)

	assert.NoError(t, err)
	assert.Equal(t, string(b), string(m.Ctx.Response.Body()))
}
