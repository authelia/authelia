package session

import (
	"context"
	"testing"

	"github.com/authelia/authelia/v4/internal/configuration/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrategyShouldCacheTheSessionForTheRequest(t *testing.T) {
	repository := newCountingRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)

	previous := newTestCachingContext()

	userSession := strategy.New("john")

	require.NoError(t, strategy.Save(previous, &userSession))

	ctx := newTestCachingContext()
	ctx.cookies[testName] = previous.cookies[testName]

	repository.reads = 0

	first, err := strategy.Get(ctx)
	require.NoError(t, err)

	second, err := strategy.Get(ctx)
	require.NoError(t, err)

	assert.Equal(t, "john", first.Username)
	assert.Equal(t, "john", second.Username)
	assert.Equal(t, 1, repository.reads)
}

func TestStrategyShouldNotReloadTheSessionItJustSaved(t *testing.T) {
	repository := newCountingRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	ctx := newTestCachingContext()

	userSession := strategy.New("john")

	require.NoError(t, strategy.Save(ctx, &userSession))

	repository.reads = 0

	saved, err := strategy.Get(ctx)
	require.NoError(t, err)

	assert.Equal(t, "john", saved.Username)
	assert.Equal(t, 0, repository.reads)
}

func TestStrategyShouldNotCacheForAContextWhichDoesNotSupportIt(t *testing.T) {
	repository := newCountingRepository()
	strategy := newTestStrategyWithRepository(t, repository, nil)
	previous := newTestContext()

	userSession := strategy.New("john")

	require.NoError(t, strategy.Save(previous, &userSession))

	ctx := newTestContext()
	ctx.cookies[testName] = previous.cookies[testName]

	repository.reads = 0

	_, err := strategy.Get(ctx)
	require.NoError(t, err)

	_, err = strategy.Get(ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, repository.reads)
}

func TestStrategyCacheShouldObserveWrites(t *testing.T) {
	testCases := []struct {
		Name     string
		Act      func(t *testing.T, strategy Strategy, ctx Context)
		Expected string
	}{
		{
			"ShouldObserveASave",
			func(t *testing.T, strategy Strategy, ctx Context) {
				updated := strategy.New("jane")

				require.NoError(t, strategy.Save(ctx, &updated))
			},
			"jane",
		},
		{
			"ShouldObserveADestroy",
			func(t *testing.T, strategy Strategy, ctx Context) {
				require.NoError(t, strategy.Destroy(ctx))
			},
			"",
		},
		{
			"ShouldObserveARegenerate",
			func(t *testing.T, strategy Strategy, ctx Context) {
				require.NoError(t, strategy.Regenerate(ctx))
			},
			"john",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			repository := newCountingRepository()
			strategy := newTestStrategyWithRepository(t, repository, nil)
			ctx := newTestCachingContext()

			userSession := strategy.New("john")

			require.NoError(t, strategy.Save(ctx, &userSession))

			cached, err := strategy.Get(ctx)
			require.NoError(t, err)
			require.Equal(t, "john", cached.Username)

			tc.Act(t, strategy, ctx)

			after, err := strategy.Get(ctx)
			require.NoError(t, err)

			assert.Equal(t, tc.Expected, after.Username)
		})
	}
}

func TestStrategyCacheShouldNotShareAcrossCookieDomains(t *testing.T) {
	repository := newCountingRepository()

	one := newTestStrategyWithRepository(t, repository, nil)
	two := newTestStrategyWithRepository(t, repository, func(config *schema.SessionCookie) {
		config.Domain = "example.org"
	})

	ctx := newTestCachingContext()

	userSession := one.New("john")

	require.NoError(t, one.Save(ctx, &userSession))

	first, err := one.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, "john", first.Username)

	second, err := two.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", second.Username, "the session cached for one cookie domain was served to another")
}

type countingRepository struct {
	*testRepository

	reads int
}

func newCountingRepository() *countingRepository {
	return &countingRepository{testRepository: newTestRepository()}
}

func (r *countingRepository) Get(ctx context.Context, issuer, id string) (data []byte, err error) {
	r.reads++

	return r.testRepository.Get(ctx, issuer, id)
}

type testCachingContext struct {
	*testContext

	sessions map[string]*UserSession
}

func newTestCachingContext() *testCachingContext {
	return &testCachingContext{testContext: newTestContext(), sessions: map[string]*UserSession{}}
}

func (c *testCachingContext) CachedSession(domain string) (userSession *UserSession, ok bool) {
	userSession, ok = c.sessions[domain]

	return userSession, ok
}

func (c *testCachingContext) CacheSession(domain string, userSession *UserSession) {
	if userSession == nil {
		delete(c.sessions, domain)

		return
	}

	c.sessions[domain] = userSession
}
