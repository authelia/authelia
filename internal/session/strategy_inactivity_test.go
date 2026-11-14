// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestDefaultStrategy_SaveShouldRetainTheSessionForTheInactivityTimeout(t *testing.T) {
	testCases := []struct {
		name           string
		inactivity     time.Duration
		keepMeLoggedIn bool
		disable        bool
		expected       time.Duration
	}{
		{"ShouldRetainForTheInactivityTimeout", time.Second * 10, false, false, time.Second * 10},
		{"ShouldRetainForTheExpirationWhenShorter", time.Minute, false, false, testExpiration},
		{"ShouldRetainForTheExpirationWhenInactivityIsDisabled", 0, false, false, testExpiration},
		{"ShouldRetainRememberedSessionForTheRememberMeDuration", time.Second * 10, true, false, testRememberMe},
		{"ShouldRetainForTheInactivityTimeoutWhenRememberMeIsDisabled", time.Second * 10, true, true, time.Second * 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()
			strategy, _ := newTestInactivityStrategy(t, repository, func(config *schema.SessionCookie) {
				config.Inactivity = tc.inactivity
				config.DisableRememberMe = tc.disable
			})

			ctx := newTestContext()

			userSession := strategy.New(testUsername)
			userSession.SetOneFactorPassword(time.Now(), tc.keepMeLoggedIn)

			require.NoError(t, strategy.Save(ctx, &userSession))

			assert.Equal(t, tc.expected, repository.expirations[testPrimaryKey(t, strategy, ctx)])
			assert.Equal(t, strategy.(*DefaultStrategy).getExpiration(userSession), repository.expirations[testPointerKey(t, strategy, ctx)])
		})
	}
}

func TestDefaultStrategy_SaveShouldNotApplyTheInactivityTimeoutToAnonymousSessions(t *testing.T) {
	repository := newTestRepository()
	strategy, _ := newTestInactivityStrategy(t, repository, func(config *schema.SessionCookie) {
		config.Inactivity = time.Second * 5
	})

	ctx := newTestContext()

	username := testUsername

	userSession := strategy.NewDefault()
	userSession.PasswordResetUsername = &username

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Equal(t, testAnonymousExpiration, repository.expirations[testPrimaryKey(t, strategy, ctx)])
}

func TestDefaultStrategy_ActivityShouldExtendTheSession(t *testing.T) {
	testCases := []struct {
		name  string
		authz bool
	}{
		{"ShouldExtendOnSessionCookieRequest", false},
		{"ShouldExtendOnAuthzCookieRequest", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()
			strategy, fixed := newTestInactivityStrategy(t, repository, func(config *schema.SessionCookie) {
				config.Inactivity = time.Second * 10
			})

			saved := newTestInactivityContext(t, strategy, fixed)
			key := testPrimaryKey(t, strategy, saved)

			get := func() *UserSession {
				var (
					actual *UserSession
					err    error
				)

				if tc.authz {
					actual, err = strategy.Authz().Get(newTestAuthzContext(saved))
				} else {
					ctx := newTestContext()
					ctx.cookies[testNameSession] = saved.cookies[testNameSession]

					actual, err = strategy.Get(ctx)
				}

				require.NoError(t, err)

				return actual
			}

			advanceTestInactivity(fixed, repository, key, time.Millisecond*500)

			assert.Equal(t, testUsername, get().Username)
			assert.Equal(t, 0, repository.touches)

			advanceTestInactivity(fixed, repository, key, time.Second*5)

			assert.Equal(t, testUsername, get().Username)
			assert.Equal(t, 1, repository.touches)
			assert.Equal(t, time.Second*10, repository.expirations[key])

			advanceTestInactivity(fixed, repository, key, time.Second*8)

			assert.Equal(t, testUsername, get().Username)
			assert.Equal(t, 2, repository.touches)
			assert.Equal(t, time.Second*10, repository.expirations[key])
		})
	}
}

func TestDefaultStrategy_ActivityShouldNotExtendTheSessionBeyondItsExpiration(t *testing.T) {
	repository := newTestRepository()
	strategy, fixed := newTestInactivityStrategy(t, repository, func(config *schema.SessionCookie) {
		config.Inactivity = time.Second * 30
	})

	saved := newTestInactivityContext(t, strategy, fixed)
	key := testPrimaryKey(t, strategy, saved)

	require.Equal(t, time.Second*30, repository.expirations[key])

	advanceTestInactivity(fixed, repository, key, time.Second*25)

	actual, err := strategy.Authz().Get(newTestAuthzContext(saved))

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
	assert.Equal(t, 1, repository.touches)
	assert.Equal(t, testExpiration-time.Second*25, repository.expirations[key])

	advanceTestInactivity(fixed, repository, key, time.Second*10)

	actual, err = strategy.Authz().Get(newTestAuthzContext(saved))

	require.NoError(t, err)
	assert.Equal(t, testUsername, actual.Username)
	assert.Equal(t, 1, repository.touches)
}

func TestDefaultStrategy_ActivityShouldNotExtendTheSessionWhenNotApplicable(t *testing.T) {
	testCases := []struct {
		name           string
		inactivity     time.Duration
		keepMeLoggedIn bool
		expires        bool
	}{
		{"ShouldNotExtendWhenInactivityIsDisabled", 0, false, false},
		{"ShouldNotExtendRememberedSession", time.Second * 10, true, false},
		{"ShouldNotExtendSessionWithoutExpiration", time.Second * 10, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()
			strategy, fixed := newTestInactivityStrategy(t, repository, func(config *schema.SessionCookie) {
				config.Inactivity = tc.inactivity
			})

			saved := newTestContext()

			userSession := strategy.New(testUsername)
			userSession.SetOneFactorPassword(fixed.Now(), tc.keepMeLoggedIn)

			require.NoError(t, strategy.Save(saved, &userSession))

			key := testPrimaryKey(t, strategy, saved)

			if tc.expires {
				p := strategy.(*DefaultStrategy)

				userSession.Expires = 0

				data, err := p.codec.Seal(testDomain, getTestSessionSignature(t, strategy, saved), userSession)
				require.NoError(t, err)

				repository.data[key] = data
			}

			advanceTestInactivity(fixed, repository, key, time.Second*5)

			ctx := newTestContext()
			ctx.cookies[testNameSession] = saved.cookies[testNameSession]

			actual, err := strategy.Get(ctx)

			require.NoError(t, err)
			assert.Equal(t, testUsername, actual.Username)
			assert.Equal(t, 0, repository.touches)
		})
	}
}

func TestDefaultStrategy_ActivityShouldNotExtendTheSessionWhenCached(t *testing.T) {
	repository := newTestRepository()
	strategy, fixed := newTestInactivityStrategy(t, repository, func(config *schema.SessionCookie) {
		config.Inactivity = time.Second * 10
	})

	saved := newTestInactivityContext(t, strategy, fixed)
	key := testPrimaryKey(t, strategy, saved)

	advanceTestInactivity(fixed, repository, key, time.Second*5)

	ctx := newTestCachingContext()
	ctx.cookies[testNameSession] = saved.cookies[testNameSession]

	for range 3 {
		_, err := strategy.Get(ctx)
		require.NoError(t, err)
	}

	assert.Equal(t, 1, repository.touches)
}

func TestDefaultStrategy_RegenerateShouldRetainTheSessionForTheInactivityTimeout(t *testing.T) {
	repository := newTestRepository()
	strategy, fixed := newTestInactivityStrategy(t, repository, func(config *schema.SessionCookie) {
		config.Inactivity = time.Second * 10
	})

	saved := newTestInactivityContext(t, strategy, fixed)

	ctx := newTestContext()
	ctx.cookies[testNameSession] = saved.cookies[testNameSession]

	require.NoError(t, strategy.Regenerate(ctx))

	assert.Equal(t, time.Second*10, repository.expirations[testPrimaryKey(t, strategy, ctx)])
	assert.Equal(t, testExpiration, repository.expirations[testPointerKey(t, strategy, ctx)])
}

func newTestInactivityStrategy(t *testing.T, repository Repository, modify func(config *schema.SessionCookie)) (Strategy, *clock.Fixed) {
	t.Helper()

	fixed := clock.NewFixed(time.Unix(1000000, 0))

	config := schema.SessionCookie{
		SessionCookieCommon: schema.SessionCookieCommon{
			Name:       testName,
			SameSite:   "lax",
			Expiration: testExpiration,
			RememberMe: testRememberMe,

			AnonymousExpiration: testAnonymousExpiration,
		},
		Domain: testDomain,
	}

	if modify != nil {
		modify(&config)
	}

	return NewStrategy(config, fixed, newTestCodec(t), repository), fixed
}

func newTestInactivityContext(t *testing.T, strategy Strategy, fixed *clock.Fixed) (ctx *testContext) {
	t.Helper()

	ctx = newTestContext()

	userSession := strategy.New(testUsername)
	userSession.SetOneFactorPassword(fixed.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	return ctx
}

func advanceTestInactivity(fixed *clock.Fixed, repository *testRepository, key string, elapsed time.Duration) {
	fixed.Set(fixed.Now().Add(elapsed))

	repository.expirations[key] -= elapsed
}

func TestDefaultStrategy_GetTouchInterval(t *testing.T) {
	testCases := []struct {
		name       string
		inactivity time.Duration
		expiration time.Duration
		expected   time.Duration
	}{
		{"ShouldUseTheDivisor", time.Minute * 5, time.Hour, time.Second * 30},
		{"ShouldUseTheFloor", time.Second * 5, time.Hour, time.Second},
		{"ShouldUseTheExpirationWhenShorter", time.Hour, time.Minute * 5, time.Second * 30},
		{"ShouldStayBelowAOneSecondInactivity", time.Second, time.Hour, time.Millisecond * 500},
		{"ShouldStayBelowASubSecondInactivity", time.Millisecond * 500, time.Hour, time.Millisecond * 250},
		{"ShouldStayBelowAShortInactivity", time.Millisecond * 1500, time.Hour, time.Millisecond * 750},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &DefaultStrategy{config: schema.SessionCookie{SessionCookieCommon: schema.SessionCookieCommon{Expiration: tc.expiration}}}

			assert.Equal(t, tc.expected, p.getTouchInterval(tc.inactivity))
		})
	}
}
