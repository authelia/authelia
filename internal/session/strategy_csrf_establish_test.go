// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStrategy_ConcurrentRequestsShouldEstablishTheSameCSRFSecret(t *testing.T) {
	testCases := []struct {
		name string
		save func(t *testing.T, strategy Strategy, ctx *testCachingContext)
	}{
		{"ShouldAgreeOnSave", func(t *testing.T, strategy Strategy, ctx *testCachingContext) {
			userSession, err := strategy.Get(ctx)
			require.NoError(t, err)
			require.Empty(t, userSession.CSRF)

			require.NoError(t, strategy.Save(ctx, userSession))
		}},
		{"ShouldAgreeOnSetCSRFCookie", func(t *testing.T, strategy Strategy, ctx *testCachingContext) {
			require.NoError(t, strategy.SetCSRFCookie(ctx))
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()
			strategy := newTestStrategyWithRepository(t, repository, nil)
			saved := newTestLegacySessionContext(t, strategy, repository)

			first, second := newTestCachingContext(), newTestCachingContext()
			first.cookies[testNameSession], second.cookies[testNameSession] = saved.cookies[testNameSession], saved.cookies[testNameSession]

			_, err := strategy.Get(first)
			require.NoError(t, err)

			_, err = strategy.Get(second)
			require.NoError(t, err)

			tc.save(t, strategy, first)
			tc.save(t, strategy, second)

			require.NotEmpty(t, first.cookies[testNameCSRF])
			assert.Equal(t, first.cookies[testNameCSRF], second.cookies[testNameCSRF])

			ctx := newTestContext()
			ctx.cookies[testNameSession] = saved.cookies[testNameSession]

			assert.NoError(t, strategy.VerifyCSRFToken(ctx, first.cookies[testNameCSRF]))
		})
	}
}

func TestDefaultStrategy_SaveShouldAdoptTheCSRFSecretOfARequestWhichWonTheRace(t *testing.T) {
	inner := newTestRepository()
	repository := &racingEstablishRepository{testRepository: inner}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestLegacySessionContext(t, strategy, inner)

	ctx := newTestContext()
	ctx.cookies[testNameSession] = saved.cookies[testNameSession]

	winner := newTestContext()
	winner.cookies[testNameSession] = saved.cookies[testNameSession]

	repository.race = func() {
		require.NoError(t, strategy.SetCSRFCookie(winner))
	}

	userSession, err := strategy.Get(ctx)
	require.NoError(t, err)

	require.NoError(t, strategy.Save(ctx, userSession))

	assert.Equal(t, 2, repository.attempts)
	require.NotEmpty(t, winner.cookies[testNameCSRF])
	assert.Equal(t, winner.cookies[testNameCSRF], ctx.cookies[testNameCSRF])

	verify := newTestContext()
	verify.cookies[testNameSession] = saved.cookies[testNameSession]

	assert.NoError(t, strategy.VerifyCSRFToken(verify, ctx.cookies[testNameCSRF]))
}

func TestDefaultStrategy_SaveShouldReturnTheFailureToEstablishTheCSRFSecret(t *testing.T) {
	inner := newTestRepository()
	repository := &erroringSaveDataRepository{testRepository: inner, err: errTestFailure}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestLegacySessionContext(t, strategy, inner)

	ctx := newTestContext()
	ctx.cookies[testNameSession] = saved.cookies[testNameSession]

	userSession, err := strategy.Get(ctx)
	require.NoError(t, err)

	assert.ErrorIs(t, strategy.Save(ctx, userSession), errTestFailure)
	assert.NotContains(t, ctx.cookies, testNameCSRF)
}

func newTestLegacySessionContext(t *testing.T, strategy Strategy, repository *testRepository) (ctx *testContext) {
	t.Helper()

	concrete := strategy.(*DefaultStrategy)

	ctx = newTestContext()

	userSession := strategy.New(testUsername)
	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	userSession.CSRF = nil

	sid := getTestSessionSignature(t, strategy, ctx)

	data, err := concrete.codec.Seal(testDomain, sid, userSession)
	require.NoError(t, err)

	repository.data[repository.key(sid, concrete.issuer)] = data

	delete(ctx.cookies, testNameCSRF)

	return ctx
}

type racingEstablishRepository struct {
	*testRepository

	race     func()
	attempts int
}

func (r *racingEstablishRepository) SaveData(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, previous, data []byte) (err error) {
	if previous != nil {
		r.attempts++

		if race := r.race; race != nil {
			r.race = nil

			race()
		}
	}

	return r.testRepository.SaveData(ctx, issuer, id, pid, username, expiration, previous, data)
}

type erroringSaveDataRepository struct {
	*testRepository

	err error
}

func (r *erroringSaveDataRepository) SaveData(_ context.Context, _, _, _, _ string, _ time.Duration, _, _ []byte) (err error) {
	return r.err
}

func TestDefaultStrategy_SaveShouldNotReplaceTheCSRFSecretWhenItCannotBeEstablished(t *testing.T) {
	inner := newTestRepository()
	repository := &erroringSaveDataRepository{testRepository: inner, err: ErrSessionSuperseded}
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestLegacySessionContext(t, strategy, inner)

	ctx := newTestContext()
	ctx.cookies[testNameSession] = saved.cookies[testNameSession]

	userSession, err := strategy.Get(ctx)
	require.NoError(t, err)

	assert.ErrorIs(t, strategy.Save(ctx, userSession), ErrSessionSuperseded)
	assert.NotContains(t, ctx.cookies, testNameCSRF)

	stored, err := strategy.Get(saved)
	require.NoError(t, err)
	assert.Empty(t, stored.CSRF)
}

func TestDefaultStrategy_SaveShouldNotRestoreTheCSRFSecretFromBeforeRegenerate(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	ctx := newTestCachingContext()
	ctx.cookies[testNameSession] = saved.cookies[testNameSession]

	userSession, err := strategy.Get(ctx)
	require.NoError(t, err)

	original := saved.cookies[testNameCSRF]

	require.NoError(t, strategy.Regenerate(ctx))

	regenerated := ctx.cookies[testNameCSRF]

	require.NotEmpty(t, regenerated)
	require.NotEqual(t, original, regenerated)

	userSession.SetTwoFactorTOTP(time.Now())

	require.NoError(t, strategy.Save(ctx, userSession))

	assert.Equal(t, regenerated, ctx.cookies[testNameCSRF])

	fresh := newTestContext()
	fresh.cookies[testNameSession] = ctx.cookies[testNameSession]

	actual, err := strategy.Get(fresh)
	require.NoError(t, err)
	assert.True(t, actual.AuthenticationMethodRefs.TOTP)

	assert.NoError(t, strategy.VerifyCSRFToken(fresh, regenerated))
	assert.ErrorIs(t, strategy.VerifyCSRFToken(fresh, original), ErrCSRFTokenInvalid)
}

func TestDefaultStrategy_RegenerateCSRFTokenShouldPersistTheNewSecretWhenTheSessionIsCached(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	saved := newTestSavedContext(t, strategy, testUsername)

	ctx := newTestCachingContext()
	ctx.cookies[testNameSession] = saved.cookies[testNameSession]

	_, err := strategy.Get(ctx)
	require.NoError(t, err)

	original := saved.cookies[testNameCSRF]

	require.NoError(t, strategy.RegenerateCSRFToken(ctx))

	regenerated := ctx.cookies[testNameCSRF]

	require.NotEmpty(t, regenerated)
	assert.NotEqual(t, original, regenerated)

	fresh := newTestContext()
	fresh.cookies[testNameSession] = saved.cookies[testNameSession]

	assert.NoError(t, strategy.VerifyCSRFToken(fresh, regenerated))
	assert.ErrorIs(t, strategy.VerifyCSRFToken(fresh, original), ErrCSRFTokenInvalid)
}
