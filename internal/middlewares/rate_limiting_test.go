package middlewares

import (
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

func TestNewRateLimit(t *testing.T) {
	testCases := []struct {
		name     string
		config   schema.ServerEndpointRateLimit
		expectNl bool
	}{
		{
			"ShouldReturnNilWhenDisabled",
			schema.ServerEndpointRateLimit{
				Enable: false,
				Buckets: []schema.ServerEndpointRateLimitBucket{
					{Period: time.Second, Requests: 1},
				},
			},
			true,
		},
		{
			"ShouldReturnNilWhenNoBuckets",
			schema.ServerEndpointRateLimit{
				Enable:  true,
				Buckets: nil,
			},
			true,
		},
		{
			"ShouldReturnMiddlewareWhenEnabled",
			schema.ServerEndpointRateLimit{
				Enable: true,
				Buckets: []schema.ServerEndpointRateLimitBucket{
					{Period: time.Second, Requests: 10},
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewRateLimit(tc.config)

			if tc.expectNl {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestNewRateLimitHandler(t *testing.T) {
	nextCalled := false

	next := func(_ *AutheliaCtx) {
		nextCalled = true
	}

	testCases := []struct {
		name       string
		config     schema.ServerEndpointRateLimit
		expectNext bool
	}{
		{
			"ShouldPassThroughWhenDisabled",
			schema.ServerEndpointRateLimit{Enable: false},
			true,
		},
		{
			"ShouldPassThroughWhenNoBuckets",
			schema.ServerEndpointRateLimit{Enable: true, Buckets: nil},
			true,
		},
		{
			"ShouldWrapWhenEnabled",
			schema.ServerEndpointRateLimit{
				Enable: true,
				Buckets: []schema.ServerEndpointRateLimitBucket{
					{Period: time.Second, Requests: 100},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled = false

			handler := NewRateLimitHandler(tc.config, next)
			require.NotNil(t, handler)

			ctx := newTestAutheliaCtx("192.168.1.1")

			handler(ctx)

			assert.Equal(t, tc.expectNext, nextCalled)
		})
	}
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
					limiter.updated = time.Now().UTC().Add(tc.updateAge)
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

			middleware := NewRateLimiter(NewIPRateLimitBucket, nil, tc.buckets...)

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
	middleware := NewRateLimiter(NewIPRateLimitBucket, nil, RateLimitBucketConfig{
		Period:   time.Minute,
		Requests: 1,
	})

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
	middleware := NewRateLimiter(NewIPRateLimitBucket, nil, RateLimitBucketConfig{
		Period:   time.Minute,
		Requests: 1,
	})

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
	middleware := NewRateLimiter(NewIPRateLimitBucket, nil,
		RateLimitBucketConfig{Period: time.Minute, Requests: 2},
		RateLimitBucketConfig{Period: time.Hour, Requests: 5},
	)

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
	middleware := NewRateLimiter(NewIPRateLimitBucket, nil, RateLimitBucketConfig{
		Period:   time.Minute,
		Requests: 1,
	})

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

	middleware := NewRateLimiter(NewIPRateLimitBucket, func(ctx *AutheliaCtx) {
		customHandlerCalled = true

		ctx.SetStatusCode(fasthttp.StatusTooManyRequests)
		ctx.SetBodyString("custom rate limit response")
	}, RateLimitBucketConfig{
		Period:   time.Minute,
		Requests: 1,
	})

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

func TestHandlerRateLimitAPI(t *testing.T) {
	ctx := newTestAutheliaCtx("192.168.1.1")

	HandlerRateLimitAPI(ctx)

	body := string(ctx.Response.Body())
	assert.Contains(t, body, "Too Many Requests")
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

	stale.updated = time.Now().UTC().Add(-time.Second)
	_ = fresh

	assert.Len(t, bucket.bucket, 2)

	bucket.GC()

	assert.Len(t, bucket.bucket, 1)
	assert.Contains(t, bucket.bucket, "192.168.1.1")
	assert.NotContains(t, bucket.bucket, "10.0.0.1")
}
