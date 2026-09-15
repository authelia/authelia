// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/cache"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestAutheliaCtxShouldFallBackToADefaultSessionWhenTheSessionCannotBeOpened(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	provider, err := session.NewProvider(&mock.Ctx.Configuration, []byte("cd21a70d4d24b23e0d1ae0a5cf6c4e1d3e2b8a5f7c9d0e1f2a3b4c5d6e7f8a9b"), []byte("a-csrf-hmac-key"), mock.Ctx.Providers.Clock, mock.Ctx.Providers.Random, &mismatchedSessionRepository{Repository: cache.NewSessionRepository(cache.NewMemory())})
	require.NoError(t, err)

	mock.Ctx.Providers.Session = provider

	mock.Ctx.Request.Header.SetCookie("authelia_session", "aaaaaaaaaaaaaaaaaaaa")

	userSession, err := mock.Ctx.GetSession()

	require.NoError(t, err)

	assert.True(t, userSession.IsAnonymous())
	assert.Equal(t, "example.com", userSession.CookieDomain)

	mock.AssertLastLogMessage(t, "Unable to retrieve user session", "error occurred getting session: signature does not match the session identifier")
}

func TestAutheliaCtxShouldResolveTheSessionFromTheRequestDomain(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	manager, err := mock.Ctx.GetSessionManagerByTargetURI(&url.URL{Scheme: "https", Host: "app.example.com"})

	require.NoError(t, err)
	require.NotNil(t, manager)

	assert.Equal(t, "example.com", manager.GetSessionConfig().Domain)
	assert.Equal(t, "example.com", mock.Ctx.NewSession().CookieDomain)
	assert.Equal(t, "example.com", mock.Ctx.GetSessionConfig().Domain)

	userSession := session.NewUserSession("john")
	userSession.CookieDomain = "example.com"

	require.NoError(t, mock.Ctx.SaveSession(&userSession))
	require.NoError(t, mock.Ctx.DestroySession())
}

func TestAutheliaCtxShouldReturnDefaultsWhenTheSessionProviderIsUnavailable(t *testing.T) {
	ctx := middlewares.NewAutheliaCtx(&fasthttp.RequestCtx{}, schema.Configuration{}, middlewares.NewProvidersBasic())

	userSession := ctx.NewSession()

	assert.True(t, userSession.IsAnonymous())
	assert.Equal(t, schema.SessionCookie{}, ctx.GetSessionConfig())

	manager, err := ctx.GetSessionManagerByTargetURI(&url.URL{Scheme: "https", Host: "app.example.com"})

	assert.Nil(t, manager)
	assert.EqualError(t, err, "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://app.example.com'")
}

func TestAutheliaCtxSetCookieShouldApplySameSite(t *testing.T) {
	testCases := []struct {
		name     string
		have     http.SameSite
		expected fasthttp.CookieSameSite
	}{
		{"ShouldNotApplyWhenUnset", 0, fasthttp.CookieSameSiteDisabled},
		{"ShouldApplyDefault", http.SameSiteDefaultMode, fasthttp.CookieSameSiteDefaultMode},
		{"ShouldApplyLax", http.SameSiteLaxMode, fasthttp.CookieSameSiteLaxMode},
		{"ShouldApplyStrict", http.SameSiteStrictMode, fasthttp.CookieSameSiteStrictMode},
		{"ShouldApplyNone", http.SameSiteNoneMode, fasthttp.CookieSameSiteNoneMode},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			//nolint:gosec // The SameSite attribute is deliberately varied, including insecure values, to observe how each is applied.
			mock.Ctx.SetCookie(&http.Cookie{Name: "test", Value: "value", Domain: "example.com", Path: "/", Secure: true, HttpOnly: true, SameSite: tc.have})

			cookie := fasthttp.AcquireCookie()

			defer fasthttp.ReleaseCookie(cookie)

			cookie.SetKey("test")

			require.True(t, mock.Ctx.Response.Header.Cookie(cookie))

			assert.Equal(t, tc.expected, cookie.SameSite())
			assert.Equal(t, "value", string(cookie.Value()))
			assert.Equal(t, "value", string(mock.Ctx.Request.Header.Cookie("test")))
		})
	}
}

func TestAutheliaCtxClearCookieShouldExpireTheCookie(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.SetCookie(&http.Cookie{Name: "test", Value: "value", Domain: "example.com", Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})

	mock.Ctx.ClearCookie(&http.Cookie{Name: "test", Value: "value", Domain: "example.com", Path: "/", Expires: time.Unix(0, 0), Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})

	cookie := fasthttp.AcquireCookie()

	defer fasthttp.ReleaseCookie(cookie)

	cookie.SetKey("test")

	require.True(t, mock.Ctx.Response.Header.Cookie(cookie))

	assert.Empty(t, cookie.Value())
	assert.Equal(t, "example.com", string(cookie.Domain()))
	assert.Equal(t, "/", string(cookie.Path()))
	assert.True(t, cookie.Secure())
	assert.True(t, cookie.HTTPOnly())
	assert.True(t, cookie.Expire().Before(time.Now()))
	assert.Empty(t, mock.Ctx.Request.Header.Cookie("test"))
}

type mismatchedSessionRepository struct {
	session.Repository
}

func (r *mismatchedSessionRepository) Get(_ context.Context, _, _ string) (record session.Record, err error) {
	return session.NewRecord("a-signature-which-was-not-requested", []byte("data")), nil
}
