package middlewares

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
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
			middleware := NewRateLimiter(WithRateLimitConfig(tc.config)).Middleware()
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

			var first *BucketLimiter

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

func TestIPRateLimitBucketFetchCtxKeysByFullAddress(t *testing.T) {
	testCases := []struct {
		name string
		a    string
		b    string
	}{
		{
			"ShouldSeparateIPv4Addresses",
			"192.168.1.1",
			"192.168.1.2",
		},
		{
			"ShouldSeparateIPv6AddressesWithinTheSamePrefix",
			"2001:db8::1",
			"2001:db8::2",
		},
		{
			"ShouldSeparateIPv6AddressesInDifferentPrefixes",
			"2001:db8:0:1::1",
			"2001:db8:0:2::1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
				Period:   time.Second,
				Requests: 10,
			}).(*IPRateLimitBucket)

			a := bucket.FetchCtx(newTestAutheliaCtx(tc.a))
			b := bucket.FetchCtx(newTestAutheliaCtx(tc.b))

			require.NotNil(t, a)
			require.NotNil(t, b)

			assert.NotSame(t, a, b)
			assert.Same(t, a, bucket.Fetch(tc.a))
			assert.Same(t, b, bucket.Fetch(tc.b))
			assert.Len(t, bucket.bucket, 2)
		})
	}
}

func TestIPRateLimitBucketGC(t *testing.T) {
	testCases := []struct {
		Name        string
		Period      time.Duration
		Requests    int
		Setup       func(t *testing.T, bucket *IPRateLimitBucket, now time.Time)
		ExpectedLen int
	}{
		{
			Name:     "ShouldNotGCRecentEntries",
			Period:   time.Hour,
			Requests: 10,
			Setup: func(t *testing.T, bucket *IPRateLimitBucket, now time.Time) {
				limiter := bucket.Fetch("192.168.1.1")

				require.NotNil(t, limiter)

				limiter.updated.Store(now.UnixNano())
			},
			ExpectedLen: 1,
		},
		{
			Name:     "ShouldGCExpiredEntries",
			Period:   time.Millisecond,
			Requests: 10,
			Setup: func(t *testing.T, bucket *IPRateLimitBucket, now time.Time) {
				limiter := bucket.Fetch("192.168.1.1")

				require.NotNil(t, limiter)

				limiter.updated.Store(now.Add(-time.Second).UnixNano())
			},
			ExpectedLen: 0,
		},
		{
			Name:        "ShouldHandleEmptyBucket",
			Period:      time.Second,
			Requests:    10,
			Setup:       nil,
			ExpectedLen: 0,
		},
		{
			Name:     "ShouldNotGCExhaustedEntryAfterOneIdlePeriod",
			Period:   time.Minute,
			Requests: 5,
			Setup: func(t *testing.T, bucket *IPRateLimitBucket, now time.Time) {
				drained := now.Add(-2 * time.Minute)
				limiter := bucket.Fetch("192.168.1.1")

				require.True(t, limiter.AllowN(drained, 5))
				require.False(t, limiter.AllowN(drained, 1))

				limiter.updated.Store(drained.UnixNano())
			},
			ExpectedLen: 1,
		},
		{
			Name:     "ShouldNotGCPartiallyRefilledEntry",
			Period:   time.Minute,
			Requests: 5,
			Setup: func(t *testing.T, bucket *IPRateLimitBucket, now time.Time) {
				drained := now.Add(-2 * time.Minute)
				limiter := bucket.Fetch("192.168.1.1")

				require.True(t, limiter.AllowN(drained, 4))

				limiter.updated.Store(drained.UnixNano())
			},
			ExpectedLen: 1,
		},
		{
			Name:     "ShouldGCFullyRefilledEntry",
			Period:   time.Minute,
			Requests: 5,
			Setup: func(t *testing.T, bucket *IPRateLimitBucket, now time.Time) {
				drained := now.Add(-2 * time.Hour)
				limiter := bucket.Fetch("192.168.1.1")

				require.True(t, limiter.AllowN(drained, 5))

				limiter.updated.Store(drained.UnixNano())
			},
			ExpectedLen: 0,
		},
		{
			Name:     "ShouldGCEntryWithZeroPeriod",
			Period:   0,
			Requests: 5,
			Setup: func(t *testing.T, bucket *IPRateLimitBucket, now time.Time) {
				limiter := bucket.Fetch("192.168.1.1")

				require.True(t, limiter.AllowN(now, 5))

				limiter.updated.Store(now.Add(-time.Second).UnixNano())
			},
			ExpectedLen: 0,
		},
		{
			Name:     "ShouldGCEntryWithZeroRequests",
			Period:   time.Second,
			Requests: 0,
			Setup: func(t *testing.T, bucket *IPRateLimitBucket, now time.Time) {
				limiter := bucket.Fetch("192.168.1.1")

				require.False(t, limiter.AllowN(now, 1))
				require.Equal(t, 0.0, limiter.TokensAt(now))

				limiter.updated.Store(now.Add(-time.Minute).UnixNano())
			},
			ExpectedLen: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
				Period:   tc.Period,
				Requests: tc.Requests,
			}).(*IPRateLimitBucket)

			if tc.Setup != nil {
				tc.Setup(t, bucket, time.Now().UTC())
			}

			bucket.GarbageCollection()

			assert.Len(t, bucket.bucket, tc.ExpectedLen)
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

			middleware := NewRateLimiter(WithRateLimitBuckets(tc.buckets...)).Middleware()

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
	})).Middleware()

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

func TestNewRateLimiterLogsOncePerRateLimitedRequest(t *testing.T) {
	testCases := []struct {
		Name            string
		Buckets         []RateLimitBucketConfig
		Requests        int
		ExpectedEntries int
		ExpectedBucket  int
		ExpectedDelay   float64
	}{
		{
			Name:            "ShouldNotLogWhenWithinLimits",
			Buckets:         []RateLimitBucketConfig{{Period: time.Minute, Requests: 2}},
			Requests:        2,
			ExpectedEntries: 0,
		},
		{
			Name:            "ShouldLogOnceForSingleBucket",
			Buckets:         []RateLimitBucketConfig{{Period: time.Minute, Requests: 1}},
			Requests:        2,
			ExpectedEntries: 1,
			ExpectedBucket:  1,
			ExpectedDelay:   60,
		},
		{
			Name: "ShouldLogOnceWhenEveryBucketExceeded",
			Buckets: []RateLimitBucketConfig{
				{Period: time.Minute, Requests: 1},
				{Period: 2 * time.Minute, Requests: 1},
				{Period: 10 * time.Minute, Requests: 1},
			},
			Requests:        2,
			ExpectedEntries: 1,
			ExpectedBucket:  3,
			ExpectedDelay:   600,
		},
		{
			Name: "ShouldLogOnceForEachRateLimitedRequest",
			Buckets: []RateLimitBucketConfig{
				{Period: time.Minute, Requests: 1},
				{Period: 2 * time.Minute, Requests: 1},
			},
			Requests:        4,
			ExpectedEntries: 3,
			ExpectedBucket:  2,
			ExpectedDelay:   120,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			middleware := NewRateLimiter(WithRateLimitBuckets(tc.Buckets...)).Middleware()

			handler := middleware(func(ctx *AutheliaCtx) {
				ctx.SetStatusCode(fasthttp.StatusOK)
			})

			logger, hook := test.NewNullLogger()

			logger.SetLevel(logrus.TraceLevel)

			for range tc.Requests {
				ctx := newTestAutheliaCtx("10.0.0.1")
				ctx.Logger = logrus.NewEntry(logger)

				handler(ctx)
			}

			entries := hook.AllEntries()

			require.Len(t, entries, tc.ExpectedEntries)

			for _, entry := range entries {
				assert.Equal(t, logrus.WarnLevel, entry.Level)
				assert.Equal(t, "Rate Limit Exceeded", entry.Message)
				assert.Equal(t, tc.ExpectedBucket, entry.Data["bucket"])
				assert.InDelta(t, tc.ExpectedDelay, entry.Data["delay"], 1)
			}
		})
	}
}

func TestNewRateLimiterDifferentIPs(t *testing.T) {
	middleware := NewRateLimiter(WithRateLimitBuckets(RateLimitBucketConfig{
		Period:   time.Minute,
		Requests: 1,
	})).Middleware()

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
	)).Middleware()

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
	})).Middleware()

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
	).Middleware()

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
			).Middleware()

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

func TestNewRateLimiterUserValueExempt(t *testing.T) {
	testCases := []struct {
		Name             string
		ExemptStatuses   []int
		BucketRequests   int
		ExemptRequests   []bool
		ExpectedStatuses []int
	}{
		{
			Name:             "ShouldNotConsumeTokensForExemptUserValueWithoutExemptStatusCodes",
			ExemptStatuses:   nil,
			BucketRequests:   2,
			ExemptRequests:   []bool{true, true, true, true},
			ExpectedStatuses: []int{fasthttp.StatusOK, fasthttp.StatusOK, fasthttp.StatusOK, fasthttp.StatusOK},
		},
		{
			Name:             "ShouldConsumeTokensForNonExemptUserValueWithoutExemptStatusCodes",
			ExemptStatuses:   nil,
			BucketRequests:   2,
			ExemptRequests:   []bool{false, false, false},
			ExpectedStatuses: []int{fasthttp.StatusOK, fasthttp.StatusOK, fasthttp.StatusTooManyRequests},
		},
		{
			Name:             "ShouldMixExemptAndNonExemptUserValues",
			ExemptStatuses:   nil,
			BucketRequests:   2,
			ExemptRequests:   []bool{true, false, true, false, false},
			ExpectedStatuses: []int{fasthttp.StatusOK, fasthttp.StatusOK, fasthttp.StatusOK, fasthttp.StatusOK, fasthttp.StatusTooManyRequests},
		},
		{
			Name:             "ShouldEnforceLimitForExemptUserValueWhenBucketAlreadyFull",
			ExemptStatuses:   nil,
			BucketRequests:   1,
			ExemptRequests:   []bool{false, true},
			ExpectedStatuses: []int{fasthttp.StatusOK, fasthttp.StatusTooManyRequests},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var exempt bool

			middleware := NewRateLimiter(
				WithRateLimitBuckets(RateLimitBucketConfig{
					Period:   time.Minute,
					Requests: tc.BucketRequests,
				}),
				WithRateLimitExemptStatusCodes(tc.ExemptStatuses...),
			).Middleware()

			handler := middleware(func(ctx *AutheliaCtx) {
				ctx.SetStatusCode(fasthttp.StatusOK)

				if exempt {
					ctx.SetUserValue(UserValueRateLimitExempt, true)
				}
			})

			for i, expected := range tc.ExpectedStatuses {
				exempt = tc.ExemptRequests[i]
				ctx := newTestAutheliaCtx("10.0.0.1")
				handler(ctx)
				assert.Equal(t, expected, ctx.Response.StatusCode(), "request %d", i+1)
			}
		})
	}
}

func TestNewIsRateLimitExempt(t *testing.T) {
	testCases := []struct {
		Name              string
		ExemptStatusCodes []int
		UserValue         any
		StatusCode        int
		Expected          bool
	}{
		{
			Name:              "ShouldReturnFalseWhenNoUserValueAndStatusNotExempt",
			ExemptStatusCodes: []int{fasthttp.StatusOK},
			UserValue:         nil,
			StatusCode:        fasthttp.StatusUnauthorized,
			Expected:          false,
		},
		{
			Name:              "ShouldReturnTrueWhenStatusExempt",
			ExemptStatusCodes: []int{fasthttp.StatusOK},
			UserValue:         nil,
			StatusCode:        fasthttp.StatusOK,
			Expected:          true,
		},
		{
			Name:              "ShouldReturnTrueWhenUserValueTrue",
			ExemptStatusCodes: nil,
			UserValue:         true,
			StatusCode:        fasthttp.StatusUnauthorized,
			Expected:          true,
		},
		{
			Name:              "ShouldReturnFalseWhenUserValueFalse",
			ExemptStatusCodes: nil,
			UserValue:         false,
			StatusCode:        fasthttp.StatusUnauthorized,
			Expected:          false,
		},
		{
			Name:              "ShouldIgnoreNonBoolUserValue",
			ExemptStatusCodes: nil,
			UserValue:         "true",
			StatusCode:        fasthttp.StatusUnauthorized,
			Expected:          false,
		},
		{
			Name:              "ShouldReturnTrueWhenUserValueTrueAndStatusExempt",
			ExemptStatusCodes: []int{fasthttp.StatusOK},
			UserValue:         true,
			StatusCode:        fasthttp.StatusOK,
			Expected:          true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			isRateLimitExempt := newIsRateLimitExempt(tc.ExemptStatusCodes)

			ctx := newTestAutheliaCtx("10.0.0.1")
			ctx.Response.SetStatusCode(tc.StatusCode)

			if tc.UserValue != nil {
				ctx.SetUserValue(UserValueRateLimitExempt, tc.UserValue)
			}

			assert.Equal(t, tc.Expected, isRateLimitExempt(ctx))
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

func TestHandlerRateLimitOpenIDConnect(t *testing.T) {
	testCases := []struct {
		name       string
		retryAfter time.Duration
		expected   string
	}{
		{
			"ShouldRenderWholeSeconds",
			time.Second * 30,
			"30",
		},
		{
			"ShouldRoundPartialSecondsUp",
			time.Millisecond * 1500,
			"2",
		},
		{
			"ShouldRenderSubSecondAsOne",
			time.Millisecond,
			"1",
		},
		{
			"ShouldRenderZero",
			0,
			"0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newTestAutheliaCtx("192.168.1.1")

			HandlerRateLimitOpenIDConnect(ctx, tc.retryAfter)

			assert.Equal(t, fasthttp.StatusTooManyRequests, ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(ctx.Response.Header.Peek(fasthttp.HeaderRetryAfter)))
			assert.Equal(t, HeaderCacheControlNotStore, string(ctx.Response.Header.Peek(fasthttp.HeaderCacheControl)))
			assert.Equal(t, HeaderPragmaNoCache, string(ctx.Response.Header.Peek(fasthttp.HeaderPragma)))
			assert.Equal(t, ContentTypeApplicationJSON, string(ctx.Response.Header.Peek(fasthttp.HeaderContentType)))
			assert.JSONEq(t, `{"error":"temporarily_unavailable","error_description":"Too many requests. The endpoint is temporarily unavailable. Try again later."}`, string(ctx.Response.Body()))
		})
	}
}

func TestIPRateLimitBucketFetchIsRaceFree(t *testing.T) {
	const (
		keys = 32
		n    = 16
	)

	bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
		Period:   time.Second,
		Requests: 10,
	}).(*IPRateLimitBucket)

	limiters := make([][]*BucketLimiter, keys)

	var ready atomic.Int64

	wg := &sync.WaitGroup{}

	for k := range keys {
		limiters[k] = make([]*BucketLimiter, n)

		key := fmt.Sprintf("192.168.1.%d", k)

		for i := range n {
			wg.Add(1)

			go func() {
				defer wg.Done()

				ready.Add(1)

				for ready.Load() < keys*n {
					runtime.Gosched()
				}

				limiters[k][i] = bucket.Fetch(key)
			}()
		}
	}

	wg.Wait()

	for k := range keys {
		require.NotNil(t, limiters[k][0])

		for i := range n {
			assert.Same(t, limiters[k][0], limiters[k][i])
		}
	}

	assert.Len(t, bucket.bucket, keys)
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

	bucket.GarbageCollection()

	assert.Len(t, bucket.bucket, 1)
	assert.Contains(t, bucket.bucket, "192.168.1.1")
	assert.NotContains(t, bucket.bucket, "10.0.0.1")
}

func TestIPRateLimitBucketGCDoesNotResetExhaustedLimiter(t *testing.T) {
	bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
		Period:   time.Minute,
		Requests: 5,
	}).(*IPRateLimitBucket)

	now := time.Now().UTC()
	drained := now.Add(-2 * time.Minute)

	limiter := bucket.Fetch("10.0.0.1")

	require.True(t, limiter.AllowN(drained, 5))
	require.False(t, limiter.AllowN(drained, 1))

	limiter.updated.Store(drained.UnixNano())

	bucket.GarbageCollection()

	require.Len(t, bucket.bucket, 1)

	assert.Same(t, limiter, bucket.Fetch("10.0.0.1"))
	assert.False(t, limiter.AllowN(now, 5))
	assert.InDelta(t, 2.0, limiter.TokensAt(now), 0.01)
}

func TestIPRateLimitBucketGCEvictsRefilledLimiters(t *testing.T) {
	bucket := NewIPRateLimitBucket(RateLimitBucketConfig{
		Period:   time.Minute,
		Requests: 5,
	}).(*IPRateLimitBucket)

	now := time.Now().UTC()

	exhausted := bucket.Fetch("10.0.0.1")
	refilled := bucket.Fetch("10.0.0.2")
	untouched := bucket.Fetch("10.0.0.3")

	require.True(t, exhausted.AllowN(now.Add(-2*time.Minute), 5))
	require.True(t, refilled.AllowN(now.Add(-2*time.Hour), 5))

	exhausted.updated.Store(now.Add(-2 * time.Minute).UnixNano())
	refilled.updated.Store(now.Add(-2 * time.Hour).UnixNano())
	untouched.updated.Store(now.Add(-2 * time.Minute).UnixNano())

	require.Len(t, bucket.bucket, 3)

	bucket.GarbageCollection()

	assert.Len(t, bucket.bucket, 1)
	assert.Contains(t, bucket.bucket, "10.0.0.1")
	assert.NotContains(t, bucket.bucket, "10.0.0.2")
	assert.NotContains(t, bucket.bucket, "10.0.0.3")
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

func TestNewRateLimiterRegistersWithCollector(t *testing.T) {
	testCases := []struct {
		name      string
		collector *GarbageCollector
		buckets   []RateLimitBucketConfig
		expected  int
	}{
		{
			"ShouldRegisterOnceWithBuckets",
			NewGarbageCollector(),
			[]RateLimitBucketConfig{
				{Period: time.Minute, Requests: 10},
				{Period: time.Hour, Requests: 20},
			},
			1,
		},
		{
			"ShouldRegisterWithoutBuckets",
			NewGarbageCollector(),
			nil,
			1,
		},
		{
			"ShouldHandleNilCollector",
			nil,
			[]RateLimitBucketConfig{
				{Period: time.Minute, Requests: 10},
			},
			0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limiter := NewRateLimiter(WithRateLimitBuckets(tc.buckets...), WithRateLimitCollector(tc.collector))

			require.NotNil(t, limiter)
			assert.NotNil(t, limiter.Middleware())
			assert.Equal(t, tc.expected, tc.collector.Len())
		})
	}
}

func TestRateLimiterGarbageCollectionFrequency(t *testing.T) {
	testCases := []struct {
		name     string
		buckets  []RateLimitBucketConfig
		expected time.Duration
	}{
		{
			"ShouldReturnShortestPeriod",
			[]RateLimitBucketConfig{
				{Period: time.Hour, Requests: 100},
				{Period: time.Minute, Requests: 10},
				{Period: time.Minute * 10, Requests: 50},
			},
			time.Minute,
		},
		{
			"ShouldReturnOnlyPeriod",
			[]RateLimitBucketConfig{
				{Period: time.Minute * 15, Requests: 10},
			},
			time.Minute * 15,
		},
		{
			"ShouldReturnZeroWithoutBuckets",
			nil,
			0,
		},
		{
			"ShouldIgnoreZeroPeriods",
			[]RateLimitBucketConfig{
				{Period: 0, Requests: 10},
				{Period: time.Minute * 5, Requests: 10},
			},
			time.Minute * 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limiter := NewRateLimiter(WithRateLimitBuckets(tc.buckets...))

			assert.Equal(t, tc.expected, limiter.GarbageCollectionFrequency(t.Context()))
		})
	}
}

func TestRateLimiterGarbageCollection(t *testing.T) {
	limiter := NewRateLimiter(WithRateLimitBuckets(
		RateLimitBucketConfig{Period: time.Minute, Requests: 10},
		RateLimitBucketConfig{Period: time.Hour, Requests: 20},
	))

	require.Len(t, limiter.buckets, 2)

	for _, bucket := range limiter.buckets {
		b := bucket.(*IPRateLimitBucket)

		b.Fetch("192.168.1.1").updated.Store(time.Now().UTC().Add(-2 * time.Hour).UnixNano())
		b.Fetch("192.168.1.2")
	}

	require.NoError(t, limiter.GarbageCollection(t.Context()))

	for _, bucket := range limiter.buckets {
		b := bucket.(*IPRateLimitBucket)

		assert.Len(t, b.bucket, 1)
		assert.NotContains(t, b.bucket, "192.168.1.1")
	}
}

func TestRateLimiterGarbageCollectionCancelledContext(t *testing.T) {
	limiter := NewRateLimiter(WithRateLimitBuckets(RateLimitBucketConfig{Period: time.Minute, Requests: 10}))

	bucket := limiter.buckets[0].(*IPRateLimitBucket)

	bucket.Fetch("192.168.1.1").updated.Store(time.Now().UTC().Add(-2 * time.Hour).UnixNano())

	ctx, cancel := context.WithCancel(t.Context())

	cancel()

	assert.EqualError(t, limiter.GarbageCollection(ctx), "context canceled")
	assert.Len(t, bucket.bucket, 1)
}

func TestRateLimiterGarbageCollectionDoesNotResetLimits(t *testing.T) {
	limiter := NewRateLimiter(WithRateLimitBuckets(RateLimitBucketConfig{Period: time.Minute, Requests: 2}))

	handler := limiter.Middleware()(func(ctx *AutheliaCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	for range 2 {
		ctx := newTestAutheliaCtx("10.0.0.1")

		handler(ctx)

		require.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	}

	ctx := newTestAutheliaCtx("10.0.0.1")

	handler(ctx)

	require.Equal(t, fasthttp.StatusTooManyRequests, ctx.Response.StatusCode())

	bucket := limiter.buckets[0].(*IPRateLimitBucket)

	require.Len(t, bucket.bucket, 1)

	bucket.bucket["10.0.0.1"].updated.Store(time.Now().UTC().Add(-2 * time.Minute).UnixNano())

	require.NoError(t, limiter.GarbageCollection(t.Context()))

	assert.Len(t, bucket.bucket, 1)

	ctx = newTestAutheliaCtx("10.0.0.1")

	handler(ctx)

	assert.Equal(t, fasthttp.StatusTooManyRequests, ctx.Response.StatusCode())
}
