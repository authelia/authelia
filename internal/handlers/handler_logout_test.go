package handlers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/mocks"
)

type LogoutSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *LogoutSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	provider, err := s.mock.Ctx.GetSessionProvider()
	s.Assert().NoError(err)

	userSession, err := provider.GetSession(s.mock.Ctx.RequestCtx)
	s.Assert().NoError(err)

	userSession.Username = testUsername
	s.Assert().NoError(provider.SaveSession(s.mock.Ctx.RequestCtx, userSession))
}

func (s *LogoutSuite) TearDownTest() {
	s.mock.Close()
}

func (s *LogoutSuite) TestShouldDestroySession() {
	LogoutPOST(s.mock.Ctx)
	b := s.mock.Ctx.Response.Header.PeekCookie("authelia_session")

	// Reset the cookie, meaning it resets the value and expires the cookie by setting
	// date to one minute in the past.
	assert.True(s.T(), strings.HasPrefix(string(b), "authelia_session=;"))
}

func TestLogoutPOST(t *testing.T) {
	testCases := []struct {
		name      string
		have      string
		expected  string
		expectedf func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleBodyParseError",
			"not a valid json",
			`{"status":"OK","data":{"safeTargetURL":false}}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing the logout request body", "unable to parse body: invalid character 'o' in literal null (expecting 'u')")
			},
		},
		{
			"ShouldHandleNoTargetURL",
			`{}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
			nil,
		},
		{
			"ShouldHandleSafeTargetURL",
			`{"targetURL":"https://www.example.com"}`,
			`{"status":"OK","data":{"safeTargetURL":true}}`,
			nil,
		},
		{
			"ShouldHandleUnsafeTargetURL",
			`{"targetURL":"https://www.notexample.com"}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
			nil,
		},
		{
			"ShouldHandleMalformedTargetURL",
			`{"targetURL":"https//www.example.com"}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			us, err := mock.Ctx.GetSession()

			require.NoError(t, err)

			us.Username = testUsername

			require.NoError(t, mock.Ctx.SaveSession(us))

			mock.Ctx.Request.SetBodyString(tc.have)

			LogoutPOST(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			assert.True(t, strings.HasPrefix(string(mock.Ctx.Response.Header.PeekCookie("authelia_session")), "authelia_session=;"))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestRunLogoutSuite(t *testing.T) {
	s := new(LogoutSuite)
	suite.Run(t, s)
}
