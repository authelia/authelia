package middlewares

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newTestAutheliaCtx(remoteIP string) *AutheliaCtx {
	var (
		reqCtx fasthttp.RequestCtx
		req    fasthttp.Request
	)

	reqCtx.Init(&req, &net.TCPAddr{IP: net.ParseIP(remoteIP), Port: 12345}, nil)

	return NewAutheliaCtx(&reqCtx, schema.Configuration{}, NewProvidersBasic())
}

func TestNewRateLimitBucketsConfig(t *testing.T) {
	testCases := []struct {
		name             string
		config           schema.ServerEndpointRateLimit
		expectedLen      int
		expectedPeriod   time.Duration
		expectedRequests int
	}{
		{
			"ShouldConvertSingleBucket",
			schema.ServerEndpointRateLimit{
				Buckets: []schema.ServerEndpointRateLimitBucket{
					{Period: 10 * time.Second, Requests: 5},
				},
			},
			1,
			10 * time.Second,
			5,
		},
		{
			"ShouldConvertMultipleBuckets",
			schema.ServerEndpointRateLimit{
				Buckets: []schema.ServerEndpointRateLimitBucket{
					{Period: 10 * time.Second, Requests: 5},
					{Period: time.Minute, Requests: 20},
				},
			},
			2,
			10 * time.Second,
			5,
		},
		{
			"ShouldReturnEmptyForNoBuckets",
			schema.ServerEndpointRateLimit{
				Buckets: nil,
			},
			0,
			0,
			0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewRateLimitBucketsConfig(tc.config)

			assert.Len(t, result, tc.expectedLen)

			if tc.expectedLen > 0 {
				assert.Equal(t, tc.expectedPeriod, result[0].Period)
				assert.Equal(t, tc.expectedRequests, result[0].Requests)
			}
		})
	}
}

func TestWithRateLimitConfig(t *testing.T) {
	testCases := []struct {
		name         string
		config       schema.ServerEndpointRateLimit
		expectStatus int
	}{
		{
			"ShouldPassThroughWhenDisabled",
			schema.ServerEndpointRateLimit{
				Enable: false,
				Buckets: []schema.ServerEndpointRateLimitBucket{
					{Period: time.Second, Requests: 1},
				},
			},
			fasthttp.StatusOK,
		},
		{
			"ShouldPassThroughWhenNoBuckets",
			schema.ServerEndpointRateLimit{
				Enable:  true,
				Buckets: nil,
			},
			fasthttp.StatusOK,
		},
		{
			"ShouldWrapWhenEnabled",
			schema.ServerEndpointRateLimit{
				Enable: true,
				Buckets: []schema.ServerEndpointRateLimitBucket{
					{Period: time.Minute, Requests: 1},
				},
			},
			fasthttp.StatusTooManyRequests,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			middleware := NewRateLimiter(WithRateLimitConfig(tc.config), WithRateLimitContext(t.Context()))
			require.NotNil(t, middleware)

			handler := middleware(func(ctx *AutheliaCtx) {
				ctx.SetStatusCode(fasthttp.StatusOK)
			})

			var last *AutheliaCtx

			for i := 0; i < 2; i++ {
				last = newTestAutheliaCtx("192.168.1.1")
				handler(last)
			}

			assert.Equal(t, tc.expectStatus, last.Response.StatusCode())
		})
	}
}

func TestWithRateLimitConfigDisabledClearsPriorBuckets(t *testing.T) {
	options := &RateLimiterOptions{}

	WithRateLimitBuckets(RateLimitBucketConfig{Period: time.Minute, Requests: 1})(options)
	require.Len(t, options.Buckets, 1)

	WithRateLimitConfig(schema.ServerEndpointRateLimit{Enable: false, Buckets: []schema.ServerEndpointRateLimitBucket{{Period: time.Second, Requests: 5}}})(options)
	assert.Nil(t, options.Buckets)
}

func TestIPRateLimitBucketFetch(t *testing.T) {
	testCases := []struct {
		name string
		keys []string
	}{
		{
			"ShouldCreateNewLimiter",
			[]string{"192.168.1.1"},
		},
		{
			"ShouldReturnExistingLimiter",
			[]string{"192.168.1.1", "192.168.1.1"},
		},
		{
			"ShouldCreateSeparateLimitersPerIP",
			[]string{"192.168.1.1", "10.0.0.1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
				Period:   time.Second,
				Requests: 10,
			}).(*IPRateLimitBucket)

			var first *RateLimiter

			for i, key := range tc.keys {
				limiter := bucket.Fetch(key)

				require.NotNil(t, limiter)

				if i == 0 {
					first = limiter
				}
			}

			if len(tc.keys) >= 2 && tc.keys[0] == tc.keys[1] {
				assert.Same(t, first, bucket.Fetch(tc.keys[0]))
			}
		})
	}
}

func TestIPRateLimitBucketFetchCtx(t *testing.T) {
	bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
		Period:   time.Second,
		Requests: 10,
	}).(*IPRateLimitBucket)

	ctx := newTestAutheliaCtx("192.168.1.1")

	limiter := bucket.FetchCtx(ctx)

	require.NotNil(t, limiter)
	assert.Same(t, limiter, bucket.Fetch("192.168.1.1"))
}

func TestIPRateLimitBucketGC(t *testing.T) {
	testCases := []struct {
		name          string
		period        time.Duration
		updateAge     time.Duration
		expectedCount int
	}{
		{
			"ShouldNotGCRecentEntries",
			time.Hour,
			0,
			1,
		},
		{
			"ShouldGCExpiredEntries",
			time.Millisecond,
			-time.Second,
			0,
		},
		{
			"ShouldHandleEmptyBucket",
			time.Second,
			0,
			0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
				Period:   tc.period,
				Requests: 10,
			}).(*IPRateLimitBucket)

			if tc.name != "ShouldHandleEmptyBucket" {
				limiter := bucket.Fetch("192.168.1.1")
				if tc.updateAge != 0 {
					limiter.updated.Store(time.Now().UTC().Add(tc.updateAge).UnixNano())
				}
			}

			bucket.GC()

			assert.Len(t, bucket.bucket, tc.expectedCount)
		})
	}
}

func TestNewRateLimiterMiddleware(t *testing.T) {
	testCases := []struct {
		name           string
		buckets        []RateLimitBucketConfig
		requests       int
		expectedStatus int
		expectNext     bool
	}{
		{
			"ShouldAllowWithinLimit",
			[]RateLimitBucketConfig{{Period: time.Second, Requests: 10}},
			1,
			fasthttp.StatusOK,
			true,
		},
		{
			"ShouldBlockExceedingLimit",
			[]RateLimitBucketConfig{{Period: time.Minute, Requests: 2}},
			3,
			fasthttp.StatusTooManyRequests,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false

			middleware := NewRateLimiter(WithRateLimitBuckets(tc.buckets...), WithRateLimitContext(t.Context()))

			handler := middleware(func(ctx *AutheliaCtx) {
				nextCalled = true

				ctx.SetStatusCode(fasthttp.StatusOK)
			})

			var lastCtx *AutheliaCtx

			for i := 0; i < tc.requests; i++ {
				lastCtx = newTestAutheliaCtx("192.168.1.1")
				nextCalled = false

				handler(lastCtx)
			}

			assert.Equal(t, tc.expectedStatus, lastCtx.Response.StatusCode())
			assert.Equal(t, tc.expectNext, nextCalled)
		})
	}
}

func TestNewRateLimiterRetryAfterHeader(t *testing.T) {
	middleware := NewRateLimiter(WithRateLimitBuckets(RateLimitBucketConfig{
		Period:   time.Minute,
		Requests: 1,
	}), WithRateLimitContext(t.Context()))

	handler := middleware(func(ctx *AutheliaCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	ctx := newTestAutheliaCtx("10.0.0.1")
	handler(ctx)
	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Empty(t, ctx.Response.Header.Peek(fasthttp.HeaderRetryAfter))

	ctx2 := newTestAutheliaCtx("10.0.0.1")
	handler(ctx2)
	assert.Equal(t, fasthttp.StatusTooManyRequests, ctx2.Response.StatusCode())
	assert.NotEmpty(t, ctx2.Response.Header.Peek(fasthttp.HeaderRetryAfter))
}

func TestNewRateLimiterDifferentIPs(t *testing.T) {
	middleware := NewRateLimiter(WithRateLimitBuckets(RateLimitBucketConfig{
		Period:   time.Minute,
		Requests: 1,
	}), WithRateLimitContext(t.Context()))

	handler := middleware(func(ctx *AutheliaCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	ctx1 := newTestAutheliaCtx("10.0.0.1")
	handler(ctx1)
	assert.Equal(t, fasthttp.StatusOK, ctx1.Response.StatusCode())

	ctx2 := newTestAutheliaCtx("10.0.0.2")
	handler(ctx2)
	assert.Equal(t, fasthttp.StatusOK, ctx2.Response.StatusCode())
}

func TestNewRateLimiterMultipleBuckets(t *testing.T) {
	middleware := NewRateLimiter(WithRateLimitBuckets(
		RateLimitBucketConfig{Period: time.Minute, Requests: 2},
		RateLimitBucketConfig{Period: time.Hour, Requests: 5},
	), WithRateLimitContext(t.Context()))

	handler := middleware(func(ctx *AutheliaCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	for i := 0; i < 2; i++ {
		ctx := newTestAutheliaCtx("10.0.0.1")
		handler(ctx)
		assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	}

	ctx := newTestAutheliaCtx("10.0.0.1")
	handler(ctx)
	assert.Equal(t, fasthttp.StatusTooManyRequests, ctx.Response.StatusCode())
}

func TestNewRateLimiterNilHandler(t *testing.T) {
	middleware := NewRateLimiter(WithRateLimitErrorHandler(nil), WithRateLimitBuckets(RateLimitBucketConfig{
		Period:   time.Minute,
		Requests: 1,
	}), WithRateLimitContext(t.Context()))

	handler := middleware(func(ctx *AutheliaCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	ctx := newTestAutheliaCtx("10.0.0.1")
	handler(ctx)

	ctx2 := newTestAutheliaCtx("10.0.0.1")
	handler(ctx2)
	assert.Equal(t, fasthttp.StatusTooManyRequests, ctx2.Response.StatusCode())

	body := string(ctx2.Response.Body())
	assert.Contains(t, body, "Too Many Requests")
}

func TestNewRateLimiterCustomHandler(t *testing.T) {
	customHandlerCalled := false

	middleware := NewRateLimiter(
		WithRateLimitBucketFunc(NewIPRateLimitBucket),
		WithRateLimitErrorHandler(func(ctx *AutheliaCtx, _ time.Duration) {
			customHandlerCalled = true

			ctx.SetStatusCode(fasthttp.StatusTooManyRequests)
			ctx.SetBodyString("custom rate limit response")
		}),
		WithRateLimitBuckets(RateLimitBucketConfig{
			Period:   time.Minute,
			Requests: 1,
		}),
		WithRateLimitContext(t.Context()),
	)

	handler := middleware(func(ctx *AutheliaCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	ctx := newTestAutheliaCtx("10.0.0.1")
	handler(ctx)

	ctx2 := newTestAutheliaCtx("10.0.0.1")
	handler(ctx2)

	assert.True(t, customHandlerCalled)
	assert.Equal(t, fasthttp.StatusTooManyRequests, ctx2.Response.StatusCode())
	assert.Contains(t, string(ctx2.Response.Body()), "custom rate limit response")
}

func TestNewRateLimiterExemptStatusCodes(t *testing.T) {
	testCases := []struct {
		Name             string
		ExemptStatuses   []int
		BucketRequests   int
		Sequence         []int
		ExpectedStatuses []int
	}{
		{
			Name:             "ShouldNotConsumeTokensForExemptStatuses",
			ExemptStatuses:   []int{fasthttp.StatusOK},
			BucketRequests:   2,
			Sequence:         []int{fasthttp.StatusOK, fasthttp.StatusOK, fasthttp.StatusOK, fasthttp.StatusOK},
			ExpectedStatuses: []int{fasthttp.StatusOK, fasthttp.StatusOK, fasthttp.StatusOK, fasthttp.StatusOK},
		},
		{
			Name:             "ShouldConsumeTokensForNonExemptStatuses",
			ExemptStatuses:   []int{fasthttp.StatusOK},
			BucketRequests:   2,
			Sequence:         []int{fasthttp.StatusUnauthorized, fasthttp.StatusUnauthorized, fasthttp.StatusUnauthorized},
			ExpectedStatuses: []int{fasthttp.StatusUnauthorized, fasthttp.StatusUnauthorized, fasthttp.StatusTooManyRequests},
		},
		{
			Name:             "ShouldEnforceLimitForExemptStatusWhenBucketAlreadyFull",
			ExemptStatuses:   []int{fasthttp.StatusOK},
			BucketRequests:   1,
			Sequence:         []int{fasthttp.StatusUnauthorized, fasthttp.StatusOK},
			ExpectedStatuses: []int{fasthttp.StatusUnauthorized, fasthttp.StatusTooManyRequests},
		},
		{
			Name:             "ShouldMixExemptAndNonExempt",
			ExemptStatuses:   []int{fasthttp.StatusOK},
			BucketRequests:   2,
			Sequence:         []int{fasthttp.StatusOK, fasthttp.StatusUnauthorized, fasthttp.StatusOK, fasthttp.StatusUnauthorized, fasthttp.StatusUnauthorized},
			ExpectedStatuses: []int{fasthttp.StatusOK, fasthttp.StatusUnauthorized, fasthttp.StatusOK, fasthttp.StatusUnauthorized, fasthttp.StatusTooManyRequests},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var nextStatus int

			middleware := NewRateLimiter(
				WithRateLimitBuckets(RateLimitBucketConfig{
					Period:   time.Minute,
					Requests: tc.BucketRequests,
				}),
				WithRateLimitExemptStatusCodes(tc.ExemptStatuses...),
				WithRateLimitContext(t.Context()),
			)

			handler := middleware(func(ctx *AutheliaCtx) {
				ctx.SetStatusCode(nextStatus)
			})

			for i, expected := range tc.ExpectedStatuses {
				nextStatus = tc.Sequence[i]
				ctx := newTestAutheliaCtx("10.0.0.1")
				handler(ctx)
				assert.Equal(t, expected, ctx.Response.StatusCode(), "request %d", i+1)
			}
		})
	}
}

func TestHandlerRateLimitAPI(t *testing.T) {
	ctx := newTestAutheliaCtx("192.168.1.1")

	HandlerRateLimitAPI(ctx, 30*time.Second)

	body := string(ctx.Response.Body())
	assert.Contains(t, body, "Too Many Requests")
	assert.Equal(t, fasthttp.StatusTooManyRequests, ctx.Response.StatusCode())
	assert.NotEmpty(t, ctx.Response.Header.Peek(fasthttp.HeaderRetryAfter))
}

func TestNewIPRateLimitBucket(t *testing.T) {
	bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
		Period:   5 * time.Second,
		Requests: 10,
	})

	require.NotNil(t, bucket)

	ipBucket, ok := bucket.(*IPRateLimitBucket)
	require.True(t, ok)

	assert.Equal(t, 10, ipBucket.b)
	assert.Equal(t, 5*time.Second, ipBucket.p)
	assert.NotNil(t, ipBucket.bucket)
	assert.Empty(t, ipBucket.bucket)
}

func TestIPRateLimitBucketGCMultipleEntries(t *testing.T) {
	bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
		Period:   time.Millisecond,
		Requests: 10,
	}).(*IPRateLimitBucket)

	fresh := bucket.Fetch("192.168.1.1")
	stale := bucket.Fetch("10.0.0.1")

	stale.updated.Store(time.Now().UTC().Add(-time.Second).UnixNano())

	_ = fresh

	assert.Len(t, bucket.bucket, 2)

	bucket.GC()

	assert.Len(t, bucket.bucket, 1)
	assert.Contains(t, bucket.bucket, "192.168.1.1")
	assert.NotContains(t, bucket.bucket, "10.0.0.1")
}

func TestIPRateLimitBucketFetchRefreshesUpdated(t *testing.T) {
	bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
		Period:   time.Hour,
		Requests: 10,
	}).(*IPRateLimitBucket)

	limiter := bucket.Fetch("192.168.1.1")
	limiter.updated.Store(time.Now().UTC().Add(-2 * time.Hour).UnixNano())

	bucket.Fetch("192.168.1.1")

	assert.WithinDuration(t, time.Now().UTC(), time.Unix(0, limiter.updated.Load()).UTC(), time.Second)
}

func TestRunRateLimitGCExitsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan struct{})

	go func() {
		runRateLimitGC(ctx, nil, time.Hour)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runRateLimitGC did not exit after context cancel")
	}
}
