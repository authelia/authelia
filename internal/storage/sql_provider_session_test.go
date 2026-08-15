package storage

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestStorageSessionShouldRoundTrip(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	const (
		issuer    = "an-issuer"
		signature = "a-signature"
		publicID  = "a-public-id"
		username  = "john"
	)

	data := []byte("the session data")

	require.NoError(t, provider.SessionSave(ctx, issuer, signature, publicID, username, time.Hour, data))

	actual, err := provider.SessionGet(ctx, issuer, signature)

	require.NoError(t, err)
	assert.Equal(t, data, actual)

	actual, err = provider.SessionGetByPublicID(ctx, issuer, publicID)

	require.NoError(t, err)
	assert.Equal(t, data, actual)

	ids, err := provider.SessionGetIDsByUsername(ctx, issuer, username)

	require.NoError(t, err)
	assert.Equal(t, []string{signature}, ids)
}

func TestStorageSessionShouldReturnNoDataWhenUnknown(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	data, err := provider.SessionGet(ctx, "an-issuer", "a-signature-which-does-not-exist")

	require.NoError(t, err)
	assert.Nil(t, data)

	data, err = provider.SessionGetByPublicID(ctx, "an-issuer", "a-public-id-which-does-not-exist")

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestStorageSessionShouldNotReturnDataForAnotherIssuer(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("data")))

	data, err := provider.SessionGet(ctx, "another-issuer", "a-signature")

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestStorageSessionShouldNotReturnExpiredData(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", -time.Hour, []byte("data")))

	data, err := provider.SessionGet(ctx, "an-issuer", "a-signature")

	require.NoError(t, err)
	assert.Nil(t, data)

	require.NoError(t, provider.SessionGarbageCollection(ctx))
}

func TestStorageSessionSaveShouldReplaceExistingSession(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("first")))
	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("second")))

	data, err := provider.SessionGet(ctx, "an-issuer", "a-signature")

	require.NoError(t, err)
	assert.Equal(t, []byte("second"), data)
}

func TestStorageSessionSaveDataShouldUpdateData(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("first")))
	require.NoError(t, provider.SessionSaveData(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("second")))

	data, err := provider.SessionGet(ctx, "an-issuer", "a-signature")

	require.NoError(t, err)
	assert.Equal(t, []byte("second"), data)
}

func TestStorageSessionChangeIDShouldMoveSession(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "old-signature", "a-public-id", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionChangeID(ctx, "an-issuer", "old-signature", "new-signature", "a-public-id", "john", time.Hour))

	data, err := provider.SessionGet(ctx, "an-issuer", "old-signature")

	require.NoError(t, err)
	assert.Nil(t, data)

	data, err = provider.SessionGet(ctx, "an-issuer", "new-signature")

	require.NoError(t, err)
	assert.Equal(t, []byte("data"), data)

	data, err = provider.SessionGetByPublicID(ctx, "an-issuer", "a-public-id")

	require.NoError(t, err)
	assert.Equal(t, []byte("data"), data)
}

func TestStorageSessionDeleteShouldRemoveSession(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionDelete(ctx, "an-issuer", "a-signature", "a-public-id", "john"))

	data, err := provider.SessionGet(ctx, "an-issuer", "a-signature")

	require.NoError(t, err)
	assert.Nil(t, data)
}

// TestStorageSessionRepositoryShouldBackSessionStrategy exercises the full 'session.storage: internal' path: a session
// strategy persisting to and reading from the storage provider.
func TestStorageSessionRepositoryShouldBackSessionStrategy(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	config := &schema.Configuration{
		Session: schema.Session{
			SessionCookieCommon: schema.SessionCookieCommon{
				Name:       "authelia_session",
				SameSite:   "lax",
				Expiration: time.Hour,
				RememberMe: time.Hour * 24,
			},
			Cookies: []schema.SessionCookie{
				{
					SessionCookieCommon: schema.SessionCookieCommon{
						Name: "authelia_session", SameSite: "lax", Expiration: time.Hour, RememberMe: time.Hour * 24,
					},
					Domain: "example.com",
				},
			},
		},
	}

	sessionProvider, err := session.NewProvider(config, []byte("an-hmac-key"), clock.New(), random.NewMathematical(), NewSessionRepository(provider))
	require.NoError(t, err)

	strategy, err := sessionProvider.GetStrategy("example.com")
	require.NoError(t, err)

	sctx := &testSessionContext{Context: ctx, cookies: map[string]string{}}

	userSession := strategy.NewDefault()
	userSession.Username = "john"

	require.NoError(t, strategy.Save(sctx, &userSession))
	require.NotEmpty(t, sctx.cookies["authelia_session"])

	actual, err := strategy.Get(sctx)

	require.NoError(t, err)
	assert.Equal(t, "john", actual.Username)
	assert.Equal(t, "example.com", actual.CookieDomain)

	require.NoError(t, strategy.Destroy(sctx))

	actual, err = strategy.Get(sctx)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
}

type testSessionContext struct {
	context.Context

	cookies map[string]string
}

func (c *testSessionContext) GetCookie(name string) string {
	return c.cookies[name]
}

func (c *testSessionContext) SetCookie(cookie *http.Cookie) {
	c.cookies[cookie.Name] = cookie.Value
}

func (c *testSessionContext) ClearCookie(cookie *http.Cookie) {
	delete(c.cookies, cookie.Name)
}

func newTestSessionProvider(t *testing.T) (ctx context.Context, provider *SQLiteProvider) {
	t.Helper()

	config := &schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	}

	ctx = context.Background()

	migrator, err := NewSQLiteProvider(config)
	require.NoError(t, err)
	require.NoError(t, migrator.SchemaMigrate(ctx, true, SchemaLatest))
	require.NoError(t, migrator.Close())

	provider, err = NewSQLiteProvider(config)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = provider.Close()
	})

	return ctx, provider
}
