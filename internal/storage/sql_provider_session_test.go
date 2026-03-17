// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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
	assert.Equal(t, session.NewRecord(signature, data), actual)

	actual, err = provider.SessionGetByPublicID(ctx, issuer, publicID)

	require.NoError(t, err)
	assert.Equal(t, session.NewRecord(signature, data), actual)

	ids, err := provider.SessionGetIDsByUsername(ctx, issuer, username)

	require.NoError(t, err)
	assert.Equal(t, []string{signature}, ids)
}

func TestStorageSessionShouldReturnNoDataWhenUnknown(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	record, err := provider.SessionGet(ctx, "an-issuer", "a-signature-which-does-not-exist")

	require.NoError(t, err)
	assert.Nil(t, record)

	record, err = provider.SessionGetByPublicID(ctx, "an-issuer", "a-public-id-which-does-not-exist")

	require.NoError(t, err)
	assert.Nil(t, record)
}

func TestStorageSessionShouldNotReturnDataForAnotherIssuer(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("data")))

	record, err := provider.SessionGet(ctx, "another-issuer", "a-signature")

	require.NoError(t, err)
	assert.Nil(t, record)
}

func TestStorageSessionShouldNotReturnExpiredData(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", -time.Hour, []byte("data")))

	record, err := provider.SessionGet(ctx, "an-issuer", "a-signature")

	require.NoError(t, err)
	assert.Nil(t, record)

	require.NoError(t, provider.SessionGarbageCollection(ctx))
}

func TestStorageSessionSaveShouldReplaceExistingSession(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("first")))
	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("second")))

	record, err := provider.SessionGet(ctx, "an-issuer", "a-signature")

	require.NoError(t, err)
	assert.Equal(t, session.NewRecord("a-signature", []byte("second")), record)
}

func TestStorageSessionSaveDataShouldUpdateData(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("first")))
	require.NoError(t, provider.SessionSaveData(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("second")))

	record, err := provider.SessionGet(ctx, "an-issuer", "a-signature")

	require.NoError(t, err)
	assert.Equal(t, session.NewRecord("a-signature", []byte("second")), record)
}

func TestStorageSessionChangeIDShouldMoveSession(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "old-signature", "a-public-id", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionChangeID(ctx, "an-issuer", "old-signature", "new-signature", "a-public-id", "john", time.Hour, []byte("resealed")))

	record, err := provider.SessionGet(ctx, "an-issuer", "old-signature")

	require.NoError(t, err)
	assert.Nil(t, record)

	record, err = provider.SessionGet(ctx, "an-issuer", "new-signature")

	require.NoError(t, err)
	assert.Equal(t, session.NewRecord("new-signature", []byte("resealed")), record)

	record, err = provider.SessionGetByPublicID(ctx, "an-issuer", "a-public-id")

	require.NoError(t, err)
	assert.Equal(t, session.NewRecord("new-signature", []byte("resealed")), record)
}

func TestStorageSessionSaveShouldDiscardStaleSaveAfterChangeID(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "old-signature", "a-public-id", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionChangeID(ctx, "an-issuer", "old-signature", "new-signature", "a-public-id", "john", time.Hour, []byte("resealed")))

	assert.ErrorIs(t, provider.SessionSave(ctx, "an-issuer", "old-signature", "a-public-id", "john", time.Hour, []byte("stale")), session.ErrSessionSuperseded)

	record, err := provider.SessionGet(ctx, "an-issuer", "old-signature")

	require.NoError(t, err)
	assert.Nil(t, record)

	record, err = provider.SessionGet(ctx, "an-issuer", "new-signature")

	require.NoError(t, err)
	assert.Equal(t, session.NewRecord("new-signature", []byte("resealed")), record)

	record, err = provider.SessionGetByPublicID(ctx, "an-issuer", "a-public-id")

	require.NoError(t, err)
	assert.Equal(t, session.NewRecord("new-signature", []byte("resealed")), record)
}

func TestStorageSessionSaveShouldReplacePublicIDOfExistingSignature(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", -time.Hour, []byte("expired")))
	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "another-public-id", "john", time.Hour, []byte("data")))

	record, err := provider.SessionGetByPublicID(ctx, "an-issuer", "another-public-id")

	require.NoError(t, err)
	assert.Equal(t, session.NewRecord("a-signature", []byte("data")), record)

	record, err = provider.SessionGetByPublicID(ctx, "an-issuer", "a-public-id")

	require.NoError(t, err)
	assert.Nil(t, record)
}

func TestStorageSessionDeleteShouldRemoveSession(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionDelete(ctx, "an-issuer", "a-signature", "a-public-id", "john"))

	record, err := provider.SessionGet(ctx, "an-issuer", "a-signature")

	require.NoError(t, err)
	assert.Nil(t, record)
}

func TestStorageSessionSaveShouldDiscardStaleSaveAfterDelete(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionDelete(ctx, "an-issuer", "a-signature", "a-public-id", "john"))

	assert.ErrorIs(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("stale")), session.ErrSessionSuperseded)

	record, err := provider.SessionGet(ctx, "an-issuer", "a-signature")

	require.NoError(t, err)
	assert.Nil(t, record)

	record, err = provider.SessionGetByPublicID(ctx, "an-issuer", "a-public-id")

	require.NoError(t, err)
	assert.Nil(t, record)

	ids, err := provider.SessionGetIDsByUsername(ctx, "an-issuer", "john")

	require.NoError(t, err)
	assert.Empty(t, ids)

	require.NoError(t, provider.SessionChangeID(ctx, "an-issuer", "a-signature", "a-new-signature", "a-public-id", "john", time.Hour, []byte("moved")))

	record, err = provider.SessionGet(ctx, "an-issuer", "a-new-signature")

	require.NoError(t, err)
	assert.Nil(t, record)

	require.NoError(t, provider.SessionSaveData(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("stale")))

	record, err = provider.SessionGet(ctx, "an-issuer", "a-signature")

	require.NoError(t, err)
	assert.Nil(t, record)
}

func TestStorageSessionGarbageCollectionShouldRemoveExpiredDestroyedSessions(t *testing.T) {
	ctx, provider := newTestSessionProvider(t)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", -time.Second, []byte("data")))
	require.NoError(t, provider.SessionDelete(ctx, "an-issuer", "a-signature", "a-public-id", "john"))
	require.NoError(t, provider.SessionGarbageCollection(ctx))

	var count int

	require.NoError(t, provider.db.GetContext(ctx, &count, "SELECT COUNT(id) FROM session;"))
	assert.Equal(t, 0, count)

	require.NoError(t, provider.SessionSave(ctx, "an-issuer", "a-signature", "a-public-id", "john", time.Hour, []byte("data")))
}

func TestStorageSessionStrategyShouldNotRestoreASessionDestroyedByAnotherRequest(t *testing.T) {
	strategy := newTestSessionStrategy(t)

	login := &testSessionContext{Context: context.Background(), cookies: map[string]string{}}

	userSession := strategy.New("john")

	require.NoError(t, strategy.Save(login, &userSession))

	cookie := login.cookies["authelia_session"]

	require.NotEmpty(t, cookie)

	stale := &testSessionContext{Context: context.Background(), cookies: map[string]string{"authelia_session": cookie}}

	loaded, err := strategy.Get(stale)

	require.NoError(t, err)
	require.Equal(t, "john", loaded.Username)

	logout := &testSessionContext{Context: context.Background(), cookies: map[string]string{"authelia_session": cookie}}

	require.NoError(t, strategy.Destroy(logout))

	loaded.LastActivity++

	require.NoError(t, strategy.Save(stale, loaded))

	assert.Equal(t, 0, stale.set)

	fresh := &testSessionContext{Context: context.Background(), cookies: map[string]string{"authelia_session": cookie}}

	actual, err := strategy.Get(fresh)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
}

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
	set     int
}

func (c *testSessionContext) GetCookie(name string) string {
	return c.cookies[name]
}

func (c *testSessionContext) SetCookie(cookie *http.Cookie) {
	c.set++
	c.cookies[cookie.Name] = cookie.Value
}

func (c *testSessionContext) ClearCookie(cookie *http.Cookie) {
	delete(c.cookies, cookie.Name)
}

func newTestSessionStrategy(t *testing.T) session.Strategy {
	t.Helper()

	_, provider := newTestSessionProvider(t)

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

	return strategy
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
