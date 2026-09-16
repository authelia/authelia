// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

func TestDefaultStrategy_NewDefault(t *testing.T) {
	testCases := []struct {
		name     string
		samesite string
		domain   string
		expected string
	}{
		{
			"ShouldUseConfiguredDomainWithLax",
			"lax",
			testDomain,
			testDomain,
		},
		{
			"ShouldUseConfiguredDomainWithStrict",
			"strict",
			testDomain,
			testDomain,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := newTestStrategy(t, func(config *schema.SessionCookie) {
				config.SameSite = tc.samesite
				config.Domain = tc.domain
			})

			userSession := strategy.NewDefault()

			assert.Equal(t, tc.expected, userSession.CookieDomain)
		})
	}
}

func TestDefaultStrategy_GetShouldReturnDefaultSessionForAnonymousRequest(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession, err := strategy.Get(ctx)

	require.NoError(t, err)
	require.NotNil(t, userSession)

	assert.True(t, userSession.IsAnonymous())
	assert.Equal(t, testDomain, userSession.CookieDomain)
}

func TestDefaultStrategy_GetShouldReturnDefaultSessionForUnknownCookie(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	ctx.cookies[testName] = "an-identifier-which-was-never-saved"

	userSession, err := strategy.Get(ctx)

	require.NoError(t, err)
	require.NotNil(t, userSession)

	assert.True(t, userSession.IsAnonymous())
}

func TestDefaultStrategy_SaveAndGetShouldRoundTrip(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	cookie, ok := ctx.cookies[testName]
	require.True(t, ok)
	assert.NotEmpty(t, cookie)

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	require.NotNil(t, actual)

	assert.Equal(t, testUsername, actual.Username)
	assert.Equal(t, testDomain, actual.CookieDomain)
	assert.NotEmpty(t, actual.PublicID)
}

func TestDefaultStrategy_SaveShouldRejectMismatchedDomain(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.CookieDomain = "notexample.com"

	err := strategy.Save(ctx, &userSession)

	assert.EqualError(t, err, "error occurred saving session: domain does not match cookie domain")
}

func TestDefaultStrategy_SaveShouldSetCookieDomainWhenEmpty(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := NewDefaultUserSession()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, testDomain, actual.CookieDomain)
}

func TestDefaultStrategy_SaveShouldApplyRememberMeExpiration(t *testing.T) {
	testCases := []struct {
		name              string
		keepMeLoggedIn    bool
		disableRememberMe bool
		expected          time.Duration
	}{
		{
			"ShouldUseExpirationWhenNotRemembered",
			false,
			false,
			testExpiration,
		},
		{
			"ShouldUseRememberMeWhenRemembered",
			true,
			false,
			testRememberMe,
		},
		{
			"ShouldUseExpirationWhenRememberMeDisabled",
			true,
			true,
			testExpiration,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()

			strategy := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
				config.DisableRememberMe = tc.disableRememberMe
			})

			ctx := newTestContext()

			userSession := strategy.NewDefault()
			userSession.Username = testUsername
			userSession.KeepMeLoggedIn = tc.keepMeLoggedIn

			require.NoError(t, strategy.Save(ctx, &userSession))

			require.Len(t, repository.expirations, 1)

			for _, expiration := range repository.expirations {
				assert.Equal(t, tc.expected, expiration)
			}
		})
	}
}

func TestDefaultStrategy_RegenerateShouldChangeCookieAndPreserveSession(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	original := ctx.cookies[testName]

	require.NoError(t, strategy.Regenerate(ctx))

	regenerated := ctx.cookies[testName]

	assert.NotEqual(t, original, regenerated)

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
}

func TestDefaultStrategy_RegenerateShouldNotErrorForAnonymousRequest(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	require.NoError(t, strategy.Regenerate(ctx))

	assert.NotEmpty(t, ctx.cookies[testName])
}

func TestDefaultStrategy_DestroyShouldClearSession(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))
	require.NoError(t, strategy.Destroy(ctx))

	assert.NotContains(t, ctx.cookies, testName)

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
}

func TestDefaultStrategy_DestroyShouldDeleteBackendRecordWhenSessionCannotBeDecoded(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))
	require.Len(t, repository.data, 1)

	for key := range repository.data {
		repository.data[key] = []byte("not-a-decodable-session")
	}

	require.NoError(t, strategy.Destroy(ctx))

	assert.NotContains(t, ctx.cookies, testName)
	assert.Empty(t, repository.data)
}

func TestDefaultStrategy_DestroyShouldClearCookieMatchingTheCookieItSet(t *testing.T) {
	testCases := []struct {
		Name     string
		SameSite string
		Expected string
	}{
		{"ShouldMatchDomainAttributeForLax", "lax", "." + testDomain},
		{"ShouldMatchDomainAttributeForStrict", "strict", testDomain},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			strategy := newTestStrategy(t, func(config *schema.SessionCookie) {
				config.SameSite = tc.SameSite
			})

			ctx := newTestContext()

			userSession := strategy.NewDefault()
			userSession.Username = testUsername

			require.NoError(t, strategy.Save(ctx, &userSession))
			require.NoError(t, strategy.Destroy(ctx))

			require.NotNil(t, ctx.cleared)

			assert.Equal(t, testName, ctx.cleared.Name)
			assert.Equal(t, tc.Expected, ctx.cleared.Domain)
			assert.Equal(t, "/", ctx.cleared.Path)
			assert.True(t, ctx.cleared.Secure)
			assert.True(t, ctx.cleared.HttpOnly)
			assert.True(t, ctx.cleared.Expires.Before(time.Now()))
		})
	}
}

func TestDefaultStrategy_CSRFTokenShouldBeEmptyForAnonymousRequest(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	token, err := strategy.CSRFToken(ctx)

	assert.Empty(t, token)
	assert.ErrorIs(t, err, ErrCSRFTokenNoSession)
	assert.ErrorIs(t, strategy.VerifyCSRFToken(ctx, ""), ErrCSRFTokenNoSession)
	assert.ErrorIs(t, strategy.VerifyCSRFToken(ctx, "not-a-token"), ErrCSRFTokenNoSession)
	assert.ErrorIs(t, strategy.RegenerateCSRFToken(ctx), ErrCSRFTokenNoSession)

	assert.NoError(t, strategy.SetCSRFCookie(ctx))

	assert.NotContains(t, ctx.cookies, testNameCSRF)
}

func TestDefaultStrategy_CSRFTokenShouldReturnNoSessionForUnknownSessionCookie(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	ctx.cookies[testName] = "an-unknown-session-cookie-value"

	token, err := strategy.CSRFToken(ctx)

	assert.Empty(t, token)
	assert.ErrorIs(t, err, ErrCSRFTokenNoSession)
	assert.ErrorIs(t, strategy.VerifyCSRFToken(ctx, "not-a-token"), ErrCSRFTokenNoSession)

	assert.NoError(t, strategy.SetCSRFCookie(ctx))

	assert.NotContains(t, ctx.cookies, testNameCSRF)
}

func TestDefaultStrategy_SaveShouldDeliverVerifiableCSRFToken(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	token := ctx.cookies[testNameCSRF]

	require.NotEmpty(t, token)
	assert.Len(t, userSession.CSRF, 32)

	actual, err := strategy.CSRFToken(ctx)

	require.NoError(t, err)
	assert.Equal(t, actual, token)
	assert.NoError(t, strategy.VerifyCSRFToken(ctx, token))
	assert.ErrorIs(t, strategy.VerifyCSRFToken(ctx, ""), ErrCSRFTokenInvalid)
	assert.ErrorIs(t, strategy.VerifyCSRFToken(ctx, token+"00"), ErrCSRFTokenInvalid)
	assert.ErrorIs(t, strategy.VerifyCSRFToken(ctx, ctx.cookies[testName]), ErrCSRFTokenInvalid)

	assert.NotContains(t, token, string(userSession.CSRF))

	assert.NotEqual(t, ctx.cookies[testName], token)

	for key := range repository.data {
		assert.NotContains(t, key, token)
	}

	cookie := ctx.set[testNameCSRF]

	require.NotNil(t, cookie)
	assert.Equal(t, "", cookie.Domain)
	assert.Equal(t, "/", cookie.Path)
	assert.True(t, cookie.Secure)
	assert.False(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.False(t, cookie.Expires.IsZero())
	assert.Equal(t, ctx.set[testName].Expires, cookie.Expires)
}

func TestDefaultStrategy_VerifyCSRFTokenShouldRejectTokenOfAnotherCookieDomain(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	other := newTestStrategy(t, func(config *schema.SessionCookie) {
		config.Domain = "example.org"
	})

	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	token := ctx.cookies[testNameCSRF]

	require.NotEmpty(t, token)
	assert.NoError(t, strategy.VerifyCSRFToken(ctx, token))

	assert.NotEqual(t, strategy.(*DefaultStrategy).csrfToken(userSession.CSRF), other.(*DefaultStrategy).csrfToken(userSession.CSRF))
	assert.Error(t, other.VerifyCSRFToken(ctx, token))
}

func TestDefaultStrategy_RegenerateShouldRotateCSRFToken(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	original := ctx.cookies[testNameCSRF]

	require.NoError(t, strategy.Regenerate(ctx))

	regenerated := ctx.cookies[testNameCSRF]

	require.NotEmpty(t, regenerated)
	assert.NotEqual(t, original, regenerated)
	assert.NoError(t, strategy.VerifyCSRFToken(ctx, regenerated))
	assert.ErrorIs(t, strategy.VerifyCSRFToken(ctx, original), ErrCSRFTokenInvalid)
}

func TestDefaultStrategy_DestroyShouldClearCSRFToken(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	token := ctx.cookies[testNameCSRF]

	require.NoError(t, strategy.Destroy(ctx))

	assert.NotContains(t, ctx.cookies, testNameCSRF)

	actual, err := strategy.CSRFToken(ctx)

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, ErrCSRFTokenNoSession)
	assert.ErrorIs(t, strategy.VerifyCSRFToken(ctx, token), ErrCSRFTokenNoSession)
}

func TestDefaultStrategy_SetCSRFCookieShouldDeliverTokenForExistingSession(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	token := ctx.cookies[testNameCSRF]

	delete(ctx.cookies, testNameCSRF)

	require.NoError(t, strategy.SetCSRFCookie(ctx))

	assert.Equal(t, token, ctx.cookies[testNameCSRF])
	assert.NoError(t, strategy.VerifyCSRFToken(ctx, ctx.cookies[testNameCSRF]))
}

func TestDefaultStrategy_SaveShouldPreserveCSRFSecret(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	original := ctx.cookies[testNameCSRF]

	userSession.KeepMeLoggedIn = true

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Equal(t, original, ctx.cookies[testNameCSRF])
	assert.NoError(t, strategy.VerifyCSRFToken(ctx, original))
}

func TestDefaultStrategy_RegenerateCSRFTokenShouldRotateWithoutChangingSessionIdentifier(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	sessionCookie := ctx.cookies[testName]
	original := ctx.cookies[testNameCSRF]

	require.NoError(t, strategy.RegenerateCSRFToken(ctx))

	regenerated := ctx.cookies[testNameCSRF]

	assert.Equal(t, sessionCookie, ctx.cookies[testName])
	require.NotEmpty(t, regenerated)
	assert.NotEqual(t, original, regenerated)
	assert.NoError(t, strategy.VerifyCSRFToken(ctx, regenerated))
	assert.ErrorIs(t, strategy.VerifyCSRFToken(ctx, original), ErrCSRFTokenInvalid)

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)

	fresh := newTestContext()
	fresh.cookies[testName] = sessionCookie

	assert.NoError(t, strategy.VerifyCSRFToken(fresh, regenerated))
	assert.ErrorIs(t, strategy.VerifyCSRFToken(fresh, original), ErrCSRFTokenInvalid)
}

func TestDefaultStrategy_SetCSRFCookieShouldGenerateSecretForSessionWithoutOne(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	concrete := strategy.(*DefaultStrategy)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	userSession.CSRF = nil

	sid := concrete.codec.Sign([]byte(ctx.cookies[testName]))

	data, err := concrete.codec.Seal(testDomain, sid, userSession)
	require.NoError(t, err)

	repository.data[repository.key(sid, concrete.issuer)] = data

	delete(ctx.cookies, testNameCSRF)

	token, err := strategy.CSRFToken(ctx)

	assert.Empty(t, token)
	assert.ErrorIs(t, err, ErrCSRFTokenInvalid)

	require.NoError(t, strategy.SetCSRFCookie(ctx))

	token = ctx.cookies[testNameCSRF]

	require.NotEmpty(t, token)
	assert.NoError(t, strategy.VerifyCSRFToken(ctx, token))
}

func TestDefaultStrategy_DestroyShouldNotErrorForAnonymousRequest(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	require.NoError(t, strategy.Destroy(ctx))

	assert.Empty(t, repository.data)
}

func TestDefaultStrategy_SaveShouldPropagateGeneratedPublicID(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.Empty(t, userSession.PublicID)
	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.NotEmpty(t, userSession.PublicID)
}

func TestDefaultStrategy_SaveShouldNotOrphanPublicIDAcrossRegenerate(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()

	require.NoError(t, strategy.Save(ctx, &userSession))

	pid := userSession.PublicID
	require.NotEmpty(t, pid)

	require.NoError(t, strategy.Regenerate(ctx))

	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Equal(t, pid, userSession.PublicID)
	assert.Len(t, repository.data, 1)
	assert.Len(t, repository.publicIDs, 1)
}

func TestDefaultStrategy_GetShouldNotReadSessionOfAnotherDomain(t *testing.T) {
	repository := newTestRepository()

	first := newTestStrategyWithRepository(t, repository, nil)
	second := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
		config.Domain = "notexample.com"
	})

	ctx := newTestContext()

	userSession := first.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, first.Save(ctx, &userSession))

	actual, err := second.Get(ctx)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
}

func TestDefaultStrategy_GetShouldRejectSessionSealedForAnotherDomain(t *testing.T) {
	repository := newTestRepository()

	autheliaURL := &url.URL{Scheme: "https", Host: "auth.example.com"}

	first := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
		config.AutheliaURL = autheliaURL
	})

	second := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
		config.AutheliaURL = autheliaURL
		config.Domain = "notexample.com"
	})

	ctx := newTestContext()

	userSession := first.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, first.Save(ctx, &userSession))

	_, err := second.Get(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error occurred decoding session")
}

func TestDefaultStrategy_GetShouldRejectSessionMovedToAnotherRecord(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))
	require.Len(t, repository.data, 1)

	var data []byte

	for _, value := range repository.data {
		data = value
	}

	codec := newTestCodec(t)
	cookie := "a-cookie-value-which-was-never-issued"

	require.NoError(t, repository.Save(ctx, codec.Sign([]byte(testDomain)), codec.Sign([]byte(cookie)), "a-public-id", testUsername, testExpiration, data))

	ctx.cookies[testName] = cookie

	_, err := strategy.Get(ctx)

	assert.EqualError(t, err, "error occurred decoding session: unable to decrypt session: cipher: message authentication failed")
}

func TestDefaultStrategy_GetShouldRejectSessionWhenTheBackendReportsAnotherSignature(t *testing.T) {
	strategy := newTestStrategyWithRepository(t, &mismatchedRepository{testRepository: newTestRepository()}, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	_, err := strategy.Get(ctx)

	assert.EqualError(t, err, "error occurred getting session: signature does not match the session identifier")
	assert.NotErrorIs(t, err, ErrRepositoryGet)
}

func TestDefaultStrategy_GetShouldReturnRepositoryErrorWhenTheBackendFails(t *testing.T) {
	strategy := newTestStrategyWithRepository(t, &failingRepository{testRepository: newTestRepository()}, nil)
	ctx := newTestContext()

	ctx.cookies[testName] = "an-identifier-which-cannot-be-retrieved"

	_, err := strategy.Get(ctx)

	assert.ErrorIs(t, err, ErrRepositoryGet)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.EqualError(t, err, "error occurred getting session from backend: context deadline exceeded")
}

func TestRepositoryShouldReturnASignatureWhichOpensASessionFoundByPublicID(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	codec := newTestCodec(t)

	record, err := repository.GetByPublicID(ctx, codec.Sign([]byte(testDomain)), userSession.PublicID)

	require.NoError(t, err)
	require.NotNil(t, record)
	require.NotEmpty(t, record.GetSessionData())
	require.NotEmpty(t, record.GetSessionSignature())

	actual := &UserSession{}

	require.NoError(t, codec.Open(testDomain, record, actual))
	assert.Equal(t, testUsername, actual.Username)
}

func TestDefaultStrategy_RegenerateShouldResealSessionAgainstTheNewIdentifier(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))
	require.Len(t, repository.data, 1)

	var sealed []byte

	for _, value := range repository.data {
		sealed = value
	}

	require.NoError(t, strategy.Regenerate(ctx))
	require.Len(t, repository.data, 1)

	var resealed []byte

	for _, value := range repository.data {
		resealed = value
	}

	assert.NotEqual(t, sealed, resealed)

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
}

var (
	_ Repository = (*testRepository)(nil)
	_ Repository = (*mismatchedRepository)(nil)
	_ Context    = (*testContext)(nil)
)

type failingRepository struct {
	*testRepository
}

func (r *failingRepository) Get(_ context.Context, _, _ string) (record Record, err error) {
	return nil, context.DeadlineExceeded
}

type testContext struct {
	context.Context

	cookies map[string]string
	set     map[string]*http.Cookie
	cleared *http.Cookie
}

func (c *testContext) GetCookie(name string) string {
	return c.cookies[name]
}

func (c *testContext) SetCookie(cookie *http.Cookie) {
	c.cookies[cookie.Name] = cookie.Value
	c.set[cookie.Name] = cookie
}

func (c *testContext) ClearCookie(cookie *http.Cookie) {
	c.cleared = cookie

	delete(c.cookies, cookie.Name)
}

type testRepository struct {
	data        map[string][]byte
	publicIDs   map[string]string
	usernames   map[string][]string
	expirations map[string]time.Duration
}

func (r *testRepository) key(id, issuer string) string {
	return issuer + ":" + id
}

func (r *testRepository) Get(_ context.Context, issuer, id string) (record Record, err error) {
	data, ok := r.data[r.key(id, issuer)]
	if !ok {
		return nil, nil
	}

	return NewRecord(id, data), nil
}

func (r *testRepository) GetByPublicID(_ context.Context, issuer, pid string) (record Record, err error) {
	id, ok := r.publicIDs[r.key(pid, issuer)]
	if !ok {
		return nil, nil
	}

	data, ok := r.data[r.key(id, issuer)]
	if !ok {
		return nil, nil
	}

	return NewRecord(id, data), nil
}

func (r *testRepository) GetIDsByUsername(_ context.Context, issuer, username string) (ids []string, err error) {
	return r.usernames[r.key(username, issuer)], nil
}

func (r *testRepository) Save(_ context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	key := r.key(id, issuer)

	r.data[key] = data
	r.expirations[key] = expiration
	r.publicIDs[r.key(pid, issuer)] = id

	if username != "" {
		usernameKey := r.key(username, issuer)
		r.usernames[usernameKey] = append(r.usernames[usernameKey], id)
	}

	return nil
}

func (r *testRepository) SaveData(_ context.Context, issuer, id, _, _ string, expiration time.Duration, data []byte) (err error) {
	key := r.key(id, issuer)

	r.data[key] = data
	r.expirations[key] = expiration

	return nil
}

func (r *testRepository) Delete(_ context.Context, issuer, id, pid, username string) (err error) {
	delete(r.data, r.key(id, issuer))
	delete(r.expirations, r.key(id, issuer))
	delete(r.publicIDs, r.key(pid, issuer))
	delete(r.usernames, r.key(username, issuer))

	return nil
}

func (r *testRepository) ChangeID(_ context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	oldKey, key := r.key(oldID, issuer), r.key(id, issuer)

	if _, ok := r.data[oldKey]; !ok {
		return nil
	}

	r.data[key] = data
	r.expirations[key] = expiration

	delete(r.data, oldKey)
	delete(r.expirations, oldKey)

	r.publicIDs[r.key(pid, issuer)] = id

	return nil
}

func (r *testRepository) GarbageCollection(_ context.Context) (err error) {
	return nil
}

func (r *testRepository) GarbageCollectionFrequency(_ context.Context) (frequency time.Duration) {
	return 0
}

type mismatchedRepository struct {
	*testRepository
}

func (r *mismatchedRepository) Get(ctx context.Context, issuer, id string) (record Record, err error) {
	if record, err = r.testRepository.Get(ctx, issuer, id); err != nil || record == nil {
		return record, err
	}

	return NewRecord("a-signature-which-was-not-requested", record.GetSessionData()), nil
}

func newTestStrategy(t *testing.T, modify func(config *schema.SessionCookie)) Strategy {
	t.Helper()

	return newTestStrategyWithRepository(t, newTestRepository(), modify)
}

func newTestStrategyWithRepository(t *testing.T, repository Repository, modify func(config *schema.SessionCookie)) Strategy {
	t.Helper()

	config := schema.SessionCookie{
		SessionCookieCommon: schema.SessionCookieCommon{
			Name:       testName,
			SameSite:   "lax",
			Expiration: testExpiration,
			RememberMe: testRememberMe,
		},
		Domain: testDomain,
	}

	if modify != nil {
		modify(&config)
	}

	return NewStrategy(config, clock.New(), newTestCodec(t), repository)
}

func newTestCodec(t *testing.T) Codec {
	t.Helper()

	codec, err := NewCodec(testSecret, []byte(testHMACKey), []byte(testHMACKey+"-csrf"), random.NewMathematical())
	require.NoError(t, err)

	return codec
}

func newTestContext() *testContext {
	return &testContext{Context: context.Background(), cookies: map[string]string{}, set: map[string]*http.Cookie{}}
}

func newTestRepository() *testRepository {
	return &testRepository{
		data:        map[string][]byte{},
		publicIDs:   map[string]string{},
		usernames:   map[string][]string{},
		expirations: map[string]time.Duration{},
	}
}
