package middlewares_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

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
