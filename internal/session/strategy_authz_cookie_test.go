// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestDefaultStrategy_SaveShouldSetBothCookies(t *testing.T) {
	testCases := []struct {
		name           string
		have           string
		expected       http.SameSite
		expectedDomain string
	}{
		{"ShouldHandleLax", "lax", http.SameSiteLaxMode, "." + testDomain},
		{"ShouldHandleNone", "none", http.SameSiteNoneMode, "." + testDomain},
		{"ShouldHandleStrict", "strict", http.SameSiteStrictMode, testDomain},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := newTestStrategy(t, func(config *schema.SessionCookie) {
				config.SameSite = tc.have
			})

			ctx := newTestContext()

			userSession := strategy.New(testUsername)

			userSession.SetOneFactorPassword(time.Now(), false)

			require.NoError(t, strategy.Save(ctx, &userSession))

			primary, authz := ctx.jar[testName], ctx.jar[testNameAuthz]

			require.NotNil(t, primary)
			require.NotNil(t, authz)

			assert.Equal(t, http.SameSiteStrictMode, primary.SameSite)
			assert.Equal(t, testDomain, primary.Domain)

			assert.Equal(t, tc.expected, authz.SameSite)
			assert.Equal(t, tc.expectedDomain, authz.Domain)
			assert.Equal(t, "/", authz.Path)
			assert.True(t, authz.Secure)
			assert.True(t, authz.HttpOnly)
			assert.Equal(t, primary.Expires, authz.Expires)
			assert.WithinDuration(t, time.Now().Add(sessionCookieLifetimeMinimum), primary.Expires, time.Minute)
			assert.NotEqual(t, primary.Value, authz.Value)

			id, err := base64.RawURLEncoding.DecodeString(authz.Value)

			require.NoError(t, err)
			assert.Len(t, id, sessionIDLength)
		})
	}
}

func TestDefaultStrategy_GetShouldIgnoreTheAuthzCookie(t *testing.T) {
	strategy := newTestStrategy(t, nil)
	ctx := newTestContext()

	userSession := strategy.New(testUsername)

	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	delete(ctx.cookies, testName)

	actual, err := strategy.Get(ctx)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
}

func TestDefaultStrategy_SaveShouldKeepTheAuthzCookieItIssued(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.New(testUsername)

	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	issued := ctx.cookies[testNameAuthz]

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Equal(t, issued, ctx.cookies[testNameAuthz])
	assert.Len(t, repository.data, 2)
}

func TestDefaultStrategy_SaveShouldNotAdoptAPlantedAuthzCookie(t *testing.T) {
	testCases := []struct {
		name     string
		existing bool
	}{
		{"ShouldHandleNewSession", false},
		{"ShouldHandleExistingSession", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()
			strategy := newTestStrategyWithRepository(t, repository, nil)
			ctx := newTestContext()

			userSession := strategy.New(testUsername)

			userSession.SetOneFactorPassword(time.Now(), false)

			if tc.existing {
				require.NoError(t, strategy.Save(ctx, &userSession))
			}

			planted := newTestCookie("planted")

			ctx.cookies[testNameAuthz] = planted

			require.NoError(t, strategy.Save(ctx, &userSession))

			assert.NotEqual(t, planted, ctx.cookies[testNameAuthz])
			assert.Len(t, repository.data, 2)
		})
	}
}

func TestDefaultStrategy_SaveShouldAdoptTheAuthzCookieIssuedByRegenerate(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.New(testUsername)

	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	stale := ctx.cookies[testNameAuthz]

	require.NoError(t, strategy.Regenerate(ctx))

	regenerated := ctx.cookies[testNameAuthz]

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Equal(t, regenerated, ctx.cookies[testNameAuthz])
	assert.NotEqual(t, stale, ctx.cookies[testNameAuthz])
	assert.Len(t, repository.data, 2)
}

func TestDefaultStrategy_RegenerateShouldRotateTheAuthzCookie(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.New(testUsername)

	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	original := ctx.cookies[testNameAuthz]

	require.NoError(t, strategy.Regenerate(ctx))

	assert.NotEqual(t, original, ctx.cookies[testNameAuthz])
	assert.Len(t, repository.data, 2)
}

func TestDefaultStrategy_DestroyShouldClearBothCookies(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.New(testUsername)

	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))
	require.NoError(t, strategy.Destroy(ctx))

	require.Len(t, ctx.cleared, 3)

	assert.Equal(t, testName, ctx.cleared[0].Name)
	assert.Equal(t, testDomain, ctx.cleared[0].Domain)
	assert.Equal(t, testNameAuthz, ctx.cleared[1].Name)
	assert.Equal(t, "."+testDomain, ctx.cleared[1].Domain)
	assert.Equal(t, testNameCSRF, ctx.cleared[2].Name)
	assert.Empty(t, repository.data)
}

func TestDefaultStrategy_SaveShouldRestoreThePointerWhenItIsGone(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.New(testUsername)

	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	issued := ctx.cookies[testNameAuthz]
	pointer := testPointerKey(t, strategy, ctx)

	delete(repository.data, pointer)

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.Equal(t, issued, ctx.cookies[testNameAuthz])
	assert.Contains(t, repository.data, pointer)
	assert.Len(t, repository.data, 2)
}

func TestDefaultStrategy_SaveShouldNotAdoptTheAuthzCookieOfAnotherSession(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctxOne, ctxTwo := newTestContext(), newTestContext()

	one, two := strategy.New(testUsername), strategy.New("harry")

	one.SetOneFactorPassword(time.Now(), false)
	two.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctxOne, &one))
	require.NoError(t, strategy.Save(ctxTwo, &two))

	key := testPointerKey(t, strategy, ctxTwo)
	pointer := append([]byte(nil), repository.data[key]...)

	require.NotEmpty(t, pointer)

	ctxOne.cookies[testNameAuthz] = ctxTwo.cookies[testNameAuthz]

	require.NoError(t, strategy.Save(ctxOne, &one))

	assert.NotEqual(t, ctxTwo.cookies[testNameAuthz], ctxOne.cookies[testNameAuthz])
	assert.Equal(t, pointer, repository.data[key])
	assert.Len(t, repository.data, 4)
}

func TestDefaultStrategy_DestroyShouldNotDeleteThePointerOfAnotherSession(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctxOne, ctxTwo := newTestContext(), newTestContext()

	one, two := strategy.New(testUsername), strategy.New("harry")

	one.SetOneFactorPassword(time.Now(), false)
	two.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctxOne, &one))
	require.NoError(t, strategy.Save(ctxTwo, &two))

	key := testPointerKey(t, strategy, ctxTwo)
	pointer := append([]byte(nil), repository.data[key]...)

	require.NotEmpty(t, pointer)

	repository.data[testPrimaryKey(t, strategy, ctxOne)] = []byte("not-a-decodable-session")

	ctxOne.cookies[testNameAuthz] = ctxTwo.cookies[testNameAuthz]

	require.NoError(t, strategy.Destroy(ctxOne))

	assert.Equal(t, pointer, repository.data[key])

	actual, err := strategy.Get(ctxTwo)

	require.NoError(t, err)
	assert.Equal(t, "harry", actual.Username)
}

func TestDefaultStrategy_DestroyShouldFollowTheAuthzCookieWithoutThePrimaryCookie(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	saved := newTestSavedContext(t, strategy, testUsername)
	ctx := newTestAuthzContext(saved)

	ctx.authz = false

	require.Len(t, repository.data, 2)
	require.NoError(t, strategy.Destroy(ctx))

	assert.Empty(t, repository.data)
	assert.Empty(t, repository.publicIDs)
	assert.Empty(t, ctx.cookies)

	require.Len(t, ctx.cleared, 3)

	assert.Equal(t, testName, ctx.cleared[0].Name)
	assert.Equal(t, testNameAuthz, ctx.cleared[1].Name)
	assert.Equal(t, testNameCSRF, ctx.cleared[2].Name)

	actual, err := strategy.Authz().Get(newTestAuthzContext(saved))

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())

	actual, err = strategy.Get(saved)

	require.NoError(t, err)
	assert.True(t, actual.IsAnonymous())
}

func TestDefaultStrategy_DestroyShouldNotFollowAnAuthzCookieWhichDoesNotResolve(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(t *testing.T, strategy Strategy, repository *testRepository, saved, other, ctx *testContext)
		records int
	}{
		{
			"ShouldHandleCookieBelongingToNothing",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, other, ctx *testContext) {
				ctx.cookies[testNameAuthz] = newTestCookie("unknown")
			},
			4,
		},
		{
			"ShouldHandleThePrimaryValueInTheAuthzCookie",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, other, ctx *testContext) {
				ctx.cookies[testNameAuthz] = other.cookies[testName]
			},
			4,
		},
		{
			"ShouldHandlePointerToAnotherSession",
			func(t *testing.T, strategy Strategy, repository *testRepository, saved, other, ctx *testContext) {
				ctx.cookies[testNameAuthz] = saved.cookies[testNameAuthz]

				repository.data[testPointerKey(t, strategy, saved)] = []byte(getTestSessionSignature(t, strategy, other))
			},
			3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := newTestRepository()
			strategy := newTestStrategyWithRepository(t, repository, nil)
			saved := newTestSavedContext(t, strategy, testUsername)
			other := newTestSavedContext(t, strategy, "harry")
			ctx := newTestContext()

			tc.setup(t, strategy, repository, saved, other, ctx)

			require.NoError(t, strategy.Destroy(ctx))

			assert.Len(t, repository.data, tc.records)
			assert.Contains(t, repository.data, testPrimaryKey(t, strategy, saved))
			assert.Contains(t, repository.data, testPrimaryKey(t, strategy, other))
			assert.Contains(t, repository.data, testPointerKey(t, strategy, other))

			require.Len(t, ctx.cleared, 3)

			actual, err := strategy.Get(other)

			require.NoError(t, err)
			assert.Equal(t, "harry", actual.Username)

			actual, err = strategy.Authz().Get(newTestAuthzContext(other))

			require.NoError(t, err)
			assert.Equal(t, "harry", actual.Username)
		})
	}
}

func TestStrategyShouldIgnoreAFailureToDeleteThePointer(t *testing.T) {
	testCases := []struct {
		name    string
		have    func(strategy Strategy, ctx *testContext, userSession *UserSession) error
		set     int
		cleared []string
		rotated bool
	}{
		{
			"ShouldRegenerate",
			func(strategy Strategy, ctx *testContext, _ *UserSession) error {
				return strategy.Regenerate(ctx)
			},
			3,
			nil,
			true,
		},
		{
			"ShouldDestroy",
			func(strategy Strategy, ctx *testContext, _ *UserSession) error {
				return strategy.Destroy(ctx)
			},
			0,
			[]string{testName, testNameAuthz, testNameCSRF},
			false,
		},
		{
			"ShouldDestroyFromTheAuthzView",
			func(strategy Strategy, ctx *testContext, _ *UserSession) error {
				ctx.authz = true

				// The CSRF cookie is a host only cookie of the portal so it never accompanies a request to the authz
				// endpoints.
				delete(ctx.cookies, testNameCSRF)

				return strategy.Authz().Destroy(ctx)
			},
			0,
			[]string{testNameAuthz, testName},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := &failingPointerDeleteRepository{testRepository: newTestRepository()}
			strategy := newTestStrategyWithRepository(t, repository, nil)

			p, ok := strategy.(*DefaultStrategy)

			require.True(t, ok)

			repository.issuer = p.authzIssuer

			ctx := newTestContext()

			userSession := strategy.New(testUsername)

			userSession.SetOneFactorPassword(time.Now(), false)

			require.NoError(t, strategy.Save(ctx, &userSession))
			require.Zero(t, repository.failed)

			primary, authz, set := ctx.cookies[testName], ctx.cookies[testNameAuthz], ctx.set

			require.NoError(t, tc.have(strategy, ctx, &userSession))

			assert.Equal(t, 1, repository.failed)
			assert.Equal(t, tc.set, ctx.set-set)

			require.Len(t, ctx.cleared, len(tc.cleared))

			for i, name := range tc.cleared {
				assert.Equal(t, name, ctx.cleared[i].Name)
			}

			if tc.cleared != nil {
				assert.Empty(t, ctx.cookies)

				return
			}

			assert.NotEmpty(t, ctx.cookies[testName])
			assert.NotEmpty(t, ctx.cookies[testNameAuthz])
			assert.NotEqual(t, authz, ctx.cookies[testNameAuthz])
			assert.NotEqual(t, newTestCookie("planted"), ctx.cookies[testNameAuthz])

			ctx.authz = true

			actual, err := strategy.Authz().Get(ctx)

			require.NoError(t, err)
			assert.Equal(t, testUsername, actual.Username)

			stale := newTestContext()

			stale.authz = true
			stale.cookies[testNameAuthz] = authz

			previous, err := strategy.Authz().Get(stale)

			require.NoError(t, err)
			assert.True(t, previous.IsAnonymous())

			assert.Equal(t, tc.rotated, primary != ctx.cookies[testName])
		})
	}
}

func testPrimaryKey(t *testing.T, strategy Strategy, ctx *testContext) string {
	t.Helper()

	p, ok := strategy.(*DefaultStrategy)

	require.True(t, ok)

	id, err := base64.RawURLEncoding.DecodeString(ctx.cookies[testName])

	require.NoError(t, err)

	return newTestRepository().key(p.codec.Sign(id), p.issuer)
}

func testPointerKey(t *testing.T, strategy Strategy, ctx *testContext) string {
	t.Helper()

	p, ok := strategy.(*DefaultStrategy)

	require.True(t, ok)

	id, err := base64.RawURLEncoding.DecodeString(ctx.cookies[testNameAuthz])

	require.NoError(t, err)

	return newTestRepository().key(p.codec.SignAuthz(id), p.authzIssuer)
}

type failingPointerDeleteRepository struct {
	*testRepository

	issuer string
	failed int
}

func (r *failingPointerDeleteRepository) Delete(ctx context.Context, issuer, id, pid, username string) (err error) {
	if issuer == r.issuer {
		r.failed++

		return errTestFailure
	}

	return r.testRepository.Delete(ctx, issuer, id, pid, username)
}

func TestDefaultStrategy_SaveShouldNotIssueTheAuthzCookieToAnAnonymousSession(t *testing.T) {
	repository := newTestRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestContext()

	userSession := strategy.New(testUsername)

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.NotEmpty(t, ctx.cookies[testName])
	assert.NotContains(t, ctx.cookies, testNameAuthz)
	assert.Len(t, repository.data, 1)

	userSession.SetOneFactorPassword(time.Now(), false)

	require.NoError(t, strategy.Save(ctx, &userSession))

	assert.NotEmpty(t, ctx.cookies[testNameAuthz])
	assert.Len(t, repository.data, 2)

	resolved, err := strategy.Authz().Get(newTestAuthzContext(ctx))

	require.NoError(t, err)
	assert.Equal(t, testUsername, resolved.Username)
}

func TestDefaultStrategy_ShouldNotHealTheAuthzCookieBeforeReplacingTheSession(t *testing.T) {
	testCases := []struct {
		name string
		have func(strategy Strategy, ctx *testContext) error
	}{
		{"ShouldHandleRegenerate", func(strategy Strategy, ctx *testContext) error { return strategy.Regenerate(ctx) }},
		{"ShouldHandleDestroy", func(strategy Strategy, ctx *testContext) error { return strategy.Destroy(ctx) }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository := &observingRepository{testRepository: newTestRepository()}
			strategy := newTestStrategyWithRepository(t, repository, nil)
			saved := newTestAuthenticatedContext(t, strategy)

			ctx := newTestContext()

			ctx.cookies[testName] = saved.cookies[testName]

			saves := repository.saves

			require.NoError(t, tc.have(strategy, ctx))

			// Regenerate saves the authz index of the new identifier once, and Destroy saves nothing, so a heal of the
			// authz cookie of the identifier being replaced would show as an additional save.
			switch tc.name {
			case "ShouldHandleRegenerate":
				assert.Equal(t, saves+1, repository.saves)
			default:
				assert.Equal(t, saves, repository.saves)
			}
		})
	}
}
