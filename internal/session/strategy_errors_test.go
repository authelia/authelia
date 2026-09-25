// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestDefaultStrategy_GetConfigShouldReturnTheSessionCookieConfiguration(t *testing.T) {
	strategy := newTestStrategy(t, nil)

	config := strategy.GetConfig()

	assert.Equal(t, testDomain, config.Domain)
	assert.Equal(t, testName, config.Name)
	assert.Equal(t, testExpiration, config.Expiration)
}

func TestDefaultStrategy_SaveShouldRejectNilSession(t *testing.T) {
	strategy := newTestStrategy(t, nil)

	assert.EqualError(t, strategy.Save(newTestContext(), nil), "error occurred saving session: it is nil")
}

func TestDefaultStrategy_SaveShouldReturnErrors(t *testing.T) {
	testCases := []struct {
		name     string
		codec    func(codec *failingCodec)
		repo     func(repository *erroringRepository)
		expected string
	}{
		{
			"ShouldReturnErrorGeneratingSessionID",
			func(codec *failingCodec) { codec.errSessionID = errTestFailure },
			nil,
			"error occurred generating session ID: bad stuff",
		},
		{
			"ShouldReturnErrorGeneratingPublicID",
			func(codec *failingCodec) { codec.errPublicID = errTestFailure },
			nil,
			"error occurred generating session public ID: bad stuff",
		},
		{
			"ShouldReturnErrorSealingSession",
			func(codec *failingCodec) { codec.errSeal = errTestFailure },
			nil,
			"error occurred encoding session: bad stuff",
		},
		{
			"ShouldReturnErrorSavingSessionToRepository",
			nil,
			func(repository *erroringRepository) { repository.errSave = errTestFailure },
			"error occurred saving session to registry: bad stuff",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			codec := &failingCodec{Codec: newTestCodec(t)}
			repository := &erroringRepository{testRepository: newTestRepository()}

			if tc.codec != nil {
				tc.codec(codec)
			}

			if tc.repo != nil {
				tc.repo(repository)
			}

			strategy := newTestStrategyWithCodec(t, codec, repository)
			ctx := newTestContext()

			userSession := strategy.NewDefault()
			userSession.Username = testUsername

			assert.EqualError(t, strategy.Save(ctx, &userSession), tc.expected)
			assert.NotContains(t, ctx.cookies, testName)
		})
	}
}

func TestDefaultStrategy_RegenerateShouldReturnErrorWhenTheSessionCannotBeRetrieved(t *testing.T) {
	strategy := newTestStrategyWithRepository(t, &failingRepository{testRepository: newTestRepository()}, nil)
	ctx := newTestContext()

	ctx.cookies[testName] = "an-identifier-which-cannot-be-retrieved"

	err := strategy.Regenerate(ctx)

	assert.ErrorIs(t, err, ErrRepositoryGet)
	assert.Equal(t, "an-identifier-which-cannot-be-retrieved", ctx.cookies[testName])
}

func TestDefaultStrategy_RegenerateShouldReturnErrors(t *testing.T) {
	testCases := []struct {
		name     string
		persist  bool
		codec    func(codec *failingCodec)
		repo     func(repository *erroringRepository)
		expected string
	}{
		{
			"ShouldReturnErrorGeneratingSessionID",
			false,
			func(codec *failingCodec) { codec.errSessionID = errTestFailure },
			nil,
			"error occurred generating session ID: bad stuff",
		},
		{
			"ShouldReturnErrorSealingSession",
			true,
			func(codec *failingCodec) { codec.errSeal = errTestFailure },
			nil,
			"error occurred encoding session: bad stuff",
		},
		{
			"ShouldReturnErrorChangingSessionID",
			true,
			nil,
			func(repository *erroringRepository) { repository.errChangeID = errTestFailure },
			"error occurred changing session ID: bad stuff",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			codec := &failingCodec{Codec: newTestCodec(t)}
			repository := &erroringRepository{testRepository: newTestRepository()}

			strategy := newTestStrategyWithCodec(t, codec, repository)
			ctx := newTestContext()

			if tc.persist {
				userSession := strategy.NewDefault()
				userSession.Username = testUsername

				require.NoError(t, strategy.Save(ctx, &userSession))
			}

			original := ctx.cookies[testName]

			if tc.codec != nil {
				tc.codec(codec)
			}

			if tc.repo != nil {
				tc.repo(repository)
			}

			assert.EqualError(t, strategy.Regenerate(ctx), tc.expected)
			assert.Equal(t, original, ctx.cookies[testName])
		})
	}
}

func TestDefaultStrategy_DestroyShouldReturnErrorWhenTheRepositoryFailsToDelete(t *testing.T) {
	repository := &erroringRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.NewDefault()
	userSession.Username = testUsername

	require.NoError(t, strategy.Save(ctx, &userSession))

	repository.errDelete = errTestFailure

	assert.EqualError(t, strategy.Destroy(ctx), "error occurred deleting session from backend: bad stuff")

	assert.Contains(t, ctx.cookies, testName)
	assert.Nil(t, ctx.cleared)

	repository.errDelete = nil

	require.NoError(t, strategy.Destroy(ctx))

	assert.NotContains(t, ctx.cookies, testName)
	assert.NotNil(t, ctx.cleared)
	assert.Empty(t, repository.data)
}

func TestDefaultStrategy_SaveShouldNotSetCookieOrCacheWhenTheSessionIsSuperseded(t *testing.T) {
	repository := &erroringRepository{testRepository: newTestRepository()}
	strategy := newTestStrategyWithRepository(t, repository, nil)

	previous := newTestContext()

	userSession := strategy.New(testUsername)

	require.NoError(t, strategy.Save(previous, &userSession))

	ctx := newTestCachingContext()
	ctx.cookies[testName] = previous.cookies[testName]

	loaded, err := strategy.Get(ctx)
	require.NoError(t, err)

	repository.errSave = ErrSessionSuperseded

	loaded.LastActivity = 1234

	require.NoError(t, strategy.Save(ctx, loaded))

	assert.Equal(t, 0, ctx.set)
	assert.Equal(t, previous.cookies[testName], ctx.cookies[testName])

	cached, ok := ctx.CachedSession(testDomain)
	require.True(t, ok)
	assert.Zero(t, cached.LastActivity)
}

func TestDefaultStrategy_GetShouldRejectSessionRecordingAnotherCookieDomain(t *testing.T) {
	codec := newTestCodec(t)
	repository := newTestRepository()
	strategy := newTestStrategyWithCodec(t, codec, repository)
	ctx := newTestContext()

	id := []byte("an-identifier-for-a-session-of-another-domain")

	sid := codec.Sign(id)

	userSession := NewUserSession(testUsername)
	userSession.CookieDomain = "another.example.com"

	data, err := codec.Seal(testDomain, sid, userSession)
	require.NoError(t, err)

	require.NoError(t, repository.Save(ctx, codec.Sign([]byte(testDomain)), sid, "a-public-id", testUsername, testExpiration, data))

	ctx.cookies[testName] = encodeCookieID(id)

	actual, err := strategy.Get(ctx)

	assert.Nil(t, actual)
	assert.EqualError(t, err, "error occurred getting session: domain does not match cookie domain")
}

func TestDefaultStrategy_GetExpires(t *testing.T) {
	strategy, ok := newTestStrategy(t, nil).(*DefaultStrategy)
	require.True(t, ok)

	assert.Equal(t, expireUnlimited, strategy.getExpires(0))

	now := time.Now()

	assert.WithinDuration(t, now.Add(testExpiration), strategy.getExpires(testExpiration), time.Second*5)
}

func TestNewSameSite(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected http.SameSite
	}{
		{"ShouldParseStrict", "strict", http.SameSiteStrictMode},
		{"ShouldParseLaxCaseInsensitive", "LAX", http.SameSiteLaxMode},
		{"ShouldParseNone", "none", http.SameSiteNoneMode},
		{"ShouldDefaultWhenEmpty", "", http.SameSiteDefaultMode},
		{"ShouldDefaultWhenUnknown", "bogus", http.SameSiteDefaultMode},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, newSameSite(tc.have))
		})
	}
}

var errTestFailure = errors.New("bad stuff")

var (
	_ Codec      = (*failingCodec)(nil)
	_ Repository = (*erroringRepository)(nil)
)

type failingCodec struct {
	Codec

	errSessionID error
	errPublicID  error
	errSeal      error
}

func (c *failingCodec) GenerateSessionID() (id []byte, err error) {
	if c.errSessionID != nil {
		return nil, c.errSessionID
	}

	return c.Codec.GenerateSessionID()
}

func (c *failingCodec) GeneratePublicID() (id string, err error) {
	if c.errPublicID != nil {
		return "", c.errPublicID
	}

	return c.Codec.GeneratePublicID()
}

func (c *failingCodec) Seal(domain, id string, session UserSession) (data []byte, err error) {
	if c.errSeal != nil {
		return nil, c.errSeal
	}

	return c.Codec.Seal(domain, id, session)
}

type erroringRepository struct {
	*testRepository

	errSave     error
	errChangeID error
	errDelete   error
}

func (r *erroringRepository) Save(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	if r.errSave != nil {
		return r.errSave
	}

	return r.testRepository.Save(ctx, issuer, id, pid, username, expiration, data)
}

func (r *erroringRepository) ChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	if r.errChangeID != nil {
		return r.errChangeID
	}

	return r.testRepository.ChangeID(ctx, issuer, oldID, id, pid, username, expiration, data)
}

func (r *erroringRepository) Delete(ctx context.Context, issuer, id, pid, username string) (err error) {
	if r.errDelete != nil {
		return r.errDelete
	}

	return r.testRepository.Delete(ctx, issuer, id, pid, username)
}

func newTestStrategyWithCodec(t *testing.T, codec Codec, repository Repository) Strategy {
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

	return NewStrategy(config, clock.New(), codec, repository)
}
