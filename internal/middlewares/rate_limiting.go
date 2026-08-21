package middlewares

import (
	"context"
	"math"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewRateLimiter takes functional options and crafts a RateLimiter out of it.
func NewRateLimiter(opts ...RateLimiterOption) (limiter *RateLimiter) {
	options := &RateLimiterOptions{}

	for _, opt := range opts {
		opt(options)
	}

	if options.disabled {
		return &RateLimiter{}
	}

	if options.NewBucket == nil {
		options.NewBucket = NewIPRateLimitBucket
	}

	if options.Handler == nil {
		options.Handler = HandlerRateLimitAPI
	}

	buckets := make([]RateLimitBucket, len(options.Buckets))

	var frequency time.Duration

	for i, b := range options.Buckets {
		buckets[i] = options.NewBucket(b)

		if b.Period > 0 && (frequency == 0 || b.Period < frequency) {
			frequency = b.Period
		}
	}

	limiter = &RateLimiter{
		buckets:           buckets,
		frequency:         frequency,
		handler:           options.Handler,
		exemptStatusCodes: options.ExemptStatusCodes,
	}

	options.Collector.Register(limiter)

	return limiter
}

// RateLimiter is a collection of RateLimitBucket which produces the middleware that enforces them and which performs
// the garbage collection of the buckets themselves.
type RateLimiter struct {
	buckets           []RateLimitBucket
	frequency         time.Duration
	handler           RateLimitRequestHandler
	exemptStatusCodes []int
}

// Middleware returns the AutheliaMiddleware which enforces the buckets of this RateLimiter. A RateLimiter without any
// configured buckets returns a passthrough middleware.
func (l *RateLimiter) Middleware() AutheliaMiddleware {
	if len(l.buckets) == 0 {
		return func(next RequestHandler) RequestHandler { return next }
	}

	return func(next RequestHandler) RequestHandler {
		return newRateLimiterHandler(next, l.buckets, l.handler, l.exemptStatusCodes)
	}
}

// GarbageCollectionFrequency returns the frequency at which the garbage collection of the buckets is performed. This
// implements the service.GarbageCollector interface.
func (l *RateLimiter) GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	return l.frequency
}

// GarbageCollection performs the garbage collection process of the buckets. This implements the
// service.GarbageCollector interface.
func (l *RateLimiter) GarbageCollection(ctx context.Context) (err error) {
	for _, bucket := range l.buckets {
		if err = ctx.Err(); err != nil {
			return err
		}

		bucket.GarbageCollection()
	}

	return nil
}

// RateLimitBucket describes an implementation of a bucket which can be leveraged for rate limiting.
type RateLimitBucket interface {
	// FetchCtx fetches the *BucketLimiter given the *AutheliaCtx.
	FetchCtx(ctx *AutheliaCtx) (limiter *BucketLimiter)

	// GarbageCollection garbage collects the buckets that are no longer being used.
	GarbageCollection()
}

// The RateLimitBucketConfig describes a limit (number of seconds), and a burst (number of events) that can occur for a
// given rate limiter.
type RateLimitBucketConfig struct {
	Period   time.Duration
	Requests int
}

// NewRateLimiterFunc is a function type that constructs a RateLimitBucket from a RateLimitBucketConfig.
type NewRateLimiterFunc func(bucket RateLimitBucketConfig) RateLimitBucket

// RateLimitRequestHandler is a function type invoked when a request exceeds the rate limit, handling the response accordingly.
type RateLimitRequestHandler = func(ctx *AutheliaCtx, retryAfter time.Duration)

// RateLimiterOptions holds the configurable values for a NewRateLimiter.
type RateLimiterOptions struct {
	disabled bool

	NewBucket         NewRateLimiterFunc
	Handler           RateLimitRequestHandler
	Buckets           []RateLimitBucketConfig
	ExemptStatusCodes []int
	Collector         *GarbageCollector
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
			options.disabled = true
			options.Buckets = nil

			return
		}

		options.disabled = false

		options.Buckets = NewRateLimitBucketsConfig(config)
	}
}

// WithRateLimitCollector registers the rate limiter buckets with a *GarbageCollector which is responsible for
// scheduling their garbage collection. If unset the buckets are never garbage collected.
func WithRateLimitCollector(collector *GarbageCollector) RateLimiterOption {
	return func(options *RateLimiterOptions) {
		options.Collector = collector
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

	ctx.Response.Header.SetBytesK(headerRetryAfter, strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
	ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueNoStore)
	ctx.Response.Header.SetBytesKV(headerPragma, headerValueNoCache)
	ctx.Response.Header.SetBytesKV(headerContentType, contentTypeApplicationJSON)

	ctx.Response.SetBodyRaw(bodyOpenIDConnectRateLimitExceeded)
}

func newRateLimiterHandler(next RequestHandler, buckets []RateLimitBucket, handler RateLimitRequestHandler, exemptStatusCodes []int) RequestHandler {
	isRateLimitExempt := newIsRateLimitExempt(exemptStatusCodes)

	return func(ctx *AutheliaCtx) {
		var (
			retryAfter       time.Duration
			retryAfterBucket int
		)

		reservations := make([]*rate.Reservation, 0, len(buckets))

		now := time.Now().UTC()

		for i, bucket := range buckets {
			limiter := bucket.FetchCtx(ctx)
			reservation := limiter.ReserveN(now, 1)
			delay := reservation.DelayFrom(now)

			if delay > 0 {
				if delay > retryAfter {
					retryAfter, retryAfterBucket = delay, i+1
				}

				reservation.CancelAt(now)

				continue
			}

			reservations = append(reservations, reservation)
		}

		if retryAfter > 0 {
			ctx.GetLogger().WithFields(map[string]any{"bucket": retryAfterBucket, "delay": retryAfter.Seconds()}).Warn("Rate Limit Exceeded")

			handler(ctx, retryAfter)

			return
		}

		next(ctx)

		if isRateLimitExempt(ctx) {
			for _, r := range reservations {
				r.CancelAt(now)
			}
		}
	}
}

func newIsRateLimitExempt(exemptStatusCodes []int) func(ctx *AutheliaCtx) bool {
	return func(ctx *AutheliaCtx) bool {
		var exempt bool

		if value := ctx.Value(UserValueRateLimitExempt); value != nil {
			exempt, _ = value.(bool)
		}

		return exempt || slices.Contains(exemptStatusCodes, ctx.Response.StatusCode())
	}
}

// NewIPRateLimitBucket returns a IPRateLimitBucket given a RateLimitBucketConfig.
func NewIPRateLimitBucket(bucket RateLimitBucketConfig) (limiter RateLimitBucket) {
	return &IPRateLimitBucket{
		bucket: make(map[string]*BucketLimiter),
		p:      bucket.Period,
		r:      rate.Every(bucket.Period),
		b:      bucket.Requests,
	}
}

// BucketLimiter is a struct which holds the important information related to a specific rate limit instance. The
// updated field stores the UnixNano of the most recent Fetch and is accessed atomically so the request hot path only
// needs to take an RLock on the parent bucket.
type BucketLimiter struct {
	*rate.Limiter

	updated atomic.Int64
}

// IPRateLimitBucket is a RateLimitBucket which limits requests based on each of the buckets delimited by IP.
type IPRateLimitBucket struct {
	bucket map[string]*BucketLimiter
	mu     sync.RWMutex
	p      time.Duration
	r      rate.Limit
	b      int
}

// Fetch the *BucketLimiter for the specific key from the dict. The common path where the limiter already exists takes
// only an RLock and refreshes the timestamp atomically; the write lock is reserved for first-time limiter creation.
func (l *IPRateLimitBucket) Fetch(key string) (limiter *BucketLimiter) {
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

// GarbageCollection garbage collects the buckets that are no longer being used.
func (l *IPRateLimitBucket) GarbageCollection() {
	now := time.Now().UTC()
	threshold := now.Add(-l.p).UnixNano()

	l.mu.Lock()

	defer l.mu.Unlock()

	for k, limiter := range l.bucket {
		if limiter.updated.Load() < threshold && limiter.TokensAt(now) >= float64(l.b) {
			delete(l.bucket, k)
		}
	}
}

// FetchCtx fetches the *BucketLimiter given the *AutheliaCtx.
func (l *IPRateLimitBucket) FetchCtx(ctx *AutheliaCtx) (limiter *BucketLimiter) {
	return l.Fetch(ctx.RemoteIP().String())
}

func (l *IPRateLimitBucket) new(ip string) (limiter *BucketLimiter) {
	limiter = &BucketLimiter{Limiter: rate.NewLimiter(l.r, l.b)}

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
