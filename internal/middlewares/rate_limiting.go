package middlewares

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// RateLimitBucket describes an implementation of a bucket which can be leveraged for rate limiting.
type RateLimitBucket interface {
	FetchCtx(ctx *AutheliaCtx) (limiter *RateLimiter)
	GC()
}

// The RateLimitBucketConfig describes a limit (number of seconds), and a burst (number of events) that can occur for a
// given rate limiter.
type RateLimitBucketConfig struct {
	Period   time.Duration
	Requests int
}

// NewIPRateLimit given a series of RateLimitBucketConfig items produces an AutheliaMiddleware which handles requests
// based on the IPRateLimitBucket. Responses whose status code is present in exemptStatusCodes do not increment the
// rate limit. Regardless of the response status code, the rate limit is enforced when a bucket is already full.
func NewIPRateLimit(exemptStatusCodes []int, bs ...RateLimitBucketConfig) AutheliaMiddleware {
	return NewRateLimiter(WithRateLimitBuckets(bs...), WithRateLimitExemptStatusCodes(exemptStatusCodes...))
}

type NewRateLimiterFunc func(bucket RateLimitBucketConfig) RateLimitBucket

type RateLimitRequestHandler = func(ctx *AutheliaCtx, retryAfter time.Duration)

// RateLimiterOptions holds the configurable values for a NewRateLimiter middleware.
type RateLimiterOptions struct {
	NewBucket         NewRateLimiterFunc
	Handler           RateLimitRequestHandler
	Buckets           []RateLimitBucketConfig
	ExemptStatusCodes []int
	Ctx               context.Context
	GCInterval        time.Duration
}

// RateLimiterOption configures a NewRateLimiter middleware.
type RateLimiterOption func(*RateLimiterOptions)

// WithRateLimitBucketFunc sets the function used to construct a RateLimitBucket from a RateLimitBucketConfig.
func WithRateLimitBucketFunc(f NewRateLimiterFunc) RateLimiterOption {
	return func(options *RateLimiterOptions) {
		options.NewBucket = f
	}
}

// WithRateLimitBuckets sets the bucket configurations for the rate limiter.
func WithRateLimitBuckets(buckets ...RateLimitBucketConfig) RateLimiterOption {
	return func(options *RateLimiterOptions) {
		options.Buckets = buckets
	}
}

// WithRateLimitErrorHandler sets the RequestHandler invoked when a request is rate limited. A nil handler is ignored
// so callers can apply this option unconditionally without clobbering a handler set by an earlier option.
func WithRateLimitErrorHandler(handler RateLimitRequestHandler) RateLimiterOption {
	return func(options *RateLimiterOptions) {
		if handler == nil {
			return
		}

		options.Handler = handler
	}
}

// WithRateLimitExemptStatusCodes sets response status codes which do not increment the rate limit. Regardless of the
// status code the rate limit is still enforced when a bucket is already full.
func WithRateLimitExemptStatusCodes(codes ...int) RateLimiterOption {
	return func(options *RateLimiterOptions) {
		options.ExemptStatusCodes = codes
	}
}

// WithRateLimitConfig replaces the rate limiter buckets with those derived from a ServerEndpointRateLimit schema
// config. A disabled config clears any previously configured buckets so the resulting NewRateLimiter middleware is a
// passthrough regardless of option ordering.
func WithRateLimitConfig(config schema.ServerEndpointRateLimit) RateLimiterOption {
	return func(options *RateLimiterOptions) {
		if !config.Enable {
			options.Buckets = nil

			return
		}

		options.Buckets = NewRateLimitBucketsConfig(config)
	}
}

// WithRateLimitContext binds the rate limiter's GC goroutine to a context. When the context is cancelled the GC ticker
// is stopped and the goroutine exits. If unset the GC goroutine runs for the lifetime of the process.
func WithRateLimitContext(ctx context.Context) RateLimiterOption {
	return func(options *RateLimiterOptions) {
		options.Ctx = ctx
	}
}

// WithRateLimitGCInterval sets the interval between rate limiter GC ticks. Non-positive values are ignored. Defaults to
// 30 minutes.
func WithRateLimitGCInterval(interval time.Duration) RateLimiterOption {
	return func(options *RateLimiterOptions) {
		if interval <= 0 {
			return
		}

		options.GCInterval = interval
	}
}

// HandlerRateLimitAPI handles general API responses for rate limiting.
func HandlerRateLimitAPI(ctx *AutheliaCtx, retryAfter time.Duration) {
	ctx.SetStatusCode(fasthttp.StatusTooManyRequests)

	ctx.Response.Header.SetBytesK(headerRetryAfter, time.Now().UTC().Add(retryAfter).Format(http.TimeFormat))
	ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueNoStore)
	ctx.Response.Header.SetBytesKV(headerPragma, headerValueNoCache)

	ctx.SetJSONError(fasthttp.StatusMessage(fasthttp.StatusTooManyRequests))
}

// HandlerRateLimitOpenIDConnect handles responses for the OpenID Connect 1.0 endpoints.
func HandlerRateLimitOpenIDConnect(ctx *AutheliaCtx, retryAfter time.Duration) {
	ctx.SetStatusCode(fasthttp.StatusTooManyRequests)

	ctx.Response.Header.SetBytesK(headerRetryAfter, strconv.Itoa(int(retryAfter.Seconds())))
	ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueNoStore)
	ctx.Response.Header.SetBytesKV(headerPragma, headerValueNoCache)
	ctx.Response.Header.SetBytesKV(headerContentType, contentTypeApplicationJSON)

	ctx.Response.SetBodyRaw(bodyOpenIDConnectRateLimitExceeded)
}

// NewRateLimiter takes functional options and crafts a RateLimiter middleware out of it.
func NewRateLimiter(opts ...RateLimiterOption) AutheliaMiddleware {
	options := &RateLimiterOptions{}

	for _, opt := range opts {
		opt(options)
	}

	if options.NewBucket == nil {
		options.NewBucket = NewIPRateLimitBucket
	}

	if options.Handler == nil {
		options.Handler = HandlerRateLimitAPI
	}

	if len(options.Buckets) == 0 {
		return func(next RequestHandler) RequestHandler { return next }
	}

	buckets := make([]RateLimitBucket, len(options.Buckets))

	for i, b := range options.Buckets {
		buckets[i] = options.NewBucket(b)
	}

	handler := options.Handler
	exemptStatusCodes := options.ExemptStatusCodes
	disableExemption := len(exemptStatusCodes) == 0

	ctx := options.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	gcInterval := options.GCInterval
	if gcInterval <= 0 {
		gcInterval = time.Minute * 30
	}

	go runRateLimitGC(ctx, buckets, gcInterval)

	return func(next RequestHandler) RequestHandler {
		return newRateLimiterHandler(next, buckets, handler, exemptStatusCodes, disableExemption)
	}
}

func runRateLimitGC(ctx context.Context, buckets []RateLimitBucket, interval time.Duration) {
	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, bucket := range buckets {
				bucket.GC()
			}
		}
	}
}

func newRateLimiterHandler(next RequestHandler, buckets []RateLimitBucket, handler RateLimitRequestHandler, exemptStatusCodes []int, disableExemption bool) RequestHandler {
	return func(ctx *AutheliaCtx) {
		var (
			retryAfter   time.Duration
			reservations []*rate.Reservation
		)

		if !disableExemption {
			reservations = make([]*rate.Reservation, 0, len(buckets))
		}

		now := time.Now().UTC()

		for i, bucket := range buckets {
			limiter := bucket.FetchCtx(ctx)
			reservation := limiter.ReserveN(now, 1)
			delay := reservation.DelayFrom(now)

			if delay > 0 {
				ctx.GetLogger().WithFields(map[string]any{"bucket": i + 1, "delay": delay.Seconds()}).Warn("Rate Limit Exceeded")

				if delay > retryAfter {
					retryAfter = delay
				}

				reservation.CancelAt(now)

				continue
			}

			if !disableExemption {
				reservations = append(reservations, reservation)
			}
		}

		if retryAfter > 0 {
			handler(ctx, retryAfter)

			return
		}

		next(ctx)

		if disableExemption {
			return
		}

		if isStatusCodeExempt(ctx.Response.StatusCode(), exemptStatusCodes) {
			for _, r := range reservations {
				r.CancelAt(now)
			}
		}
	}
}

func isStatusCodeExempt(status int, exemptStatusCodes []int) bool {
	for _, code := range exemptStatusCodes {
		if code == status {
			return true
		}
	}

	return false
}

// NewIPRateLimitBucket returns a IPRateLimitBucket given a RateLimitBucketConfig.
func NewIPRateLimitBucket(bucket RateLimitBucketConfig) (limiter RateLimitBucket) {
	return &IPRateLimitBucket{
		bucket: make(map[string]*RateLimiter),
		p:      bucket.Period,
		r:      rate.Every(bucket.Period),
		b:      bucket.Requests,
	}
}

// RateLimiter is a struct which holds the important information related to a specific rate limit instance. The
// updated field stores the UnixNano of the most recent Fetch and is accessed atomically so the request hot path only
// needs to take an RLock on the parent bucket.
type RateLimiter struct {
	*rate.Limiter

	updated atomic.Int64
}

// IPRateLimitBucket is a RateLimitBucket which limits requests based on each of the buckets delimited by IP.
type IPRateLimitBucket struct {
	bucket map[string]*RateLimiter
	mu     sync.RWMutex
	p      time.Duration
	r      rate.Limit
	b      int
}

// Fetch the *RateLimiter for the specific key from the dict. The common path where the limiter already exists takes
// only an RLock and refreshes the timestamp atomically; the write lock is reserved for first-time limiter creation.
func (l *IPRateLimitBucket) Fetch(key string) (limiter *RateLimiter) {
	now := time.Now().UTC().UnixNano()

	l.mu.RLock()

	if limiter, ok := l.bucket[key]; ok {
		limiter.updated.Store(now)
		l.mu.RUnlock()

		return limiter
	}

	l.mu.RUnlock()

	l.mu.Lock()

	defer l.mu.Unlock()

	if limiter, ok := l.bucket[key]; ok {
		limiter.updated.Store(now)

		return limiter
	}

	limiter = l.new(key)
	limiter.updated.Store(now)

	return limiter
}

// GC the rate limit bucket.
func (l *IPRateLimitBucket) GC() {
	threshold := time.Now().UTC().Add(-l.p).UnixNano()

	l.mu.Lock()

	defer l.mu.Unlock()

	for k, limiter := range l.bucket {
		if limiter.updated.Load() < threshold {
			delete(l.bucket, k)
		}
	}
}

// FetchCtx fetches the *RateLimiter given the *AutheliaCtx.
func (l *IPRateLimitBucket) FetchCtx(ctx *AutheliaCtx) (limiter *RateLimiter) {
	return l.Fetch(ctx.RemoteIP().String())
}

// new constructs and inserts a new RateLimiter for the given key. The caller must hold l.mu as a write lock and is
// responsible for storing the initial value of updated.
func (l *IPRateLimitBucket) new(ip string) (limiter *RateLimiter) {
	limiter = &RateLimiter{Limiter: rate.NewLimiter(l.r, l.b)}

	l.bucket[ip] = limiter

	return limiter
}

// NewRateLimitBucketsConfig converts a schema.ServerEndpointRateLimit to a RateLimitBucketConfig slice.
func NewRateLimitBucketsConfig(config schema.ServerEndpointRateLimit) []RateLimitBucketConfig {
	buckets := make([]RateLimitBucketConfig, len(config.Buckets))

	for i, bucket := range config.Buckets {
		buckets[i] = RateLimitBucketConfig{Period: bucket.Period, Requests: bucket.Requests}
	}

	return buckets
}
