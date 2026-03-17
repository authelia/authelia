// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/cache"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestAutheliaCtxShouldReturnTheErrorWhenTheSessionBackendFails(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	provider, err := session.NewProvider(&mock.Ctx.Configuration, []byte("cd21a70d4d24b23e0d1ae0a5cf6c4e1d3e2b8a5f7c9d0e1f2a3b4c5d6e7f8a9b"), mock.Ctx.Providers.Clock, mock.Ctx.Providers.Random, &failingSessionRepository{Repository: cache.NewSessionRepository(cache.NewMemory())})
	require.NoError(t, err)

	mock.Ctx.Providers.Session = provider

	mock.Ctx.Request.Header.SetCookie("authelia_session", "aaaaaaaaaaaaaaaaaaaa")

	_, err = mock.Ctx.GetSession()

	assert.ErrorIs(t, err, session.ErrRepositoryGet)
}

func TestAutheliaCtxShouldImplementTheSessionCachingContext(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	caching, ok := any(mock.Ctx).(session.CachingContext)

	require.True(t, ok, "the request context does not retain the session for the request")

	_, ok = caching.CachedSession("example.com")
	assert.False(t, ok)

	userSession := session.NewUserSession("john")

	caching.CacheSession("example.com", &userSession)

	cached, ok := caching.CachedSession("example.com")

	require.True(t, ok)
	assert.Equal(t, "john", cached.Username)

	_, ok = caching.CachedSession("example.org")
	assert.False(t, ok, "the session retained for one cookie domain was served to another")

	caching.CacheSession("example.com", nil)

	_, ok = caching.CachedSession("example.com")
	assert.False(t, ok)
}

func TestAutheliaCtxShouldReadTheSessionOncePerRequest(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	userSession := session.NewUserSession("john")
	userSession.CookieDomain = "example.com"

	require.NoError(t, mock.Ctx.SaveSession(&userSession))

	first, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	second, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	assert.Equal(t, "john", first.Username)
	assert.Equal(t, "john", second.Username)
}

type failingSessionRepository struct {
	session.Repository
}

func (r *failingSessionRepository) Get(_ context.Context, _, _ string) (record session.Record, err error) {
	return nil, context.DeadlineExceeded
}
