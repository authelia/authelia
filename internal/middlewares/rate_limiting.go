package middlewares

import (
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// RateLimitBucket describes an implementation of a bucket which can be leveraged for rate limiting.
type RateLimitBucket interface {
	FetchCtx(ctx *AutheliaCtx) (limiter *rate.Limiter)
}

// The RateLimitBucketConfig describes a limit (number of seconds), and a burst (number of events) that can occur for a
// given rate limiter.
type RateLimitBucketConfig struct {
	Period   time.Duration
	Requests int
}

// NewIPRateLimit given a series of RateLimitBucketConfig items produces an AutheliaMiddleware which handles requests based
// on the IPRateLimitBucket.
func NewIPRateLimit(bs ...RateLimitBucketConfig) AutheliaMiddleware {
	return NewRateLimiter(NewIPRateLimitBucket, HandlerRateLimitAPI, bs...)
}

type NewRateLimiterFunc func(bucket RateLimitBucketConfig) RateLimitBucket

func HandlerRateLimitAPI(ctx *AutheliaCtx) {
	ctx.SetStatusCode(fasthttp.StatusTooManyRequests)
	ctx.SetJSONError(fasthttp.StatusMessage(fasthttp.StatusTooManyRequests))
}

func NewRateLimiter(newBucket func(bucket RateLimitBucketConfig) RateLimitBucket, handler RequestHandler, bs ...RateLimitBucketConfig) AutheliaMiddleware {
	buckets := make([]RateLimitBucket, len(bs))

	for i, b := range bs {
		buckets[i] = newBucket(b)
	}

	if handler == nil {
		handler = HandlerRateLimitAPI
	}

	return func(next RequestHandler) RequestHandler {
		return func(ctx *AutheliaCtx) {
			var exceeded bool

			for i, bucket := range buckets {
				limiter := bucket.FetchCtx(ctx)

				if !limiter.Allow() {
					ctx.Logger.WithField("bucket", i+1).Warn("Rate Limit Exceeded")

					exceeded = true
				}
			}

			if exceeded {
				handler(ctx)

				return
			}

			next(ctx)
		}
	}
}

// NewIPRateLimitBucket returns a IPRateLimitBucket given a RateLimitBucketConfig.
func NewIPRateLimitBucket(bucket RateLimitBucketConfig) (limiter RateLimitBucket) {
	return &IPRateLimitBucket{
		bucket: make(map[string]*rate.Limiter),
		mu:     sync.Mutex{},
		r:      rate.Limit(bucket.Period.Seconds()),
		b:      bucket.Requests,
	}
}

// IPRateLimitBucket is a RateLimitBucket which limits requests based on each of the buckets delimited by IP.
type IPRateLimitBucket struct {
	bucket map[string]*rate.Limiter
	mu     sync.Mutex
	r      rate.Limit
	b      int
}

func (l *IPRateLimitBucket) Fetch(key string) (limiter *rate.Limiter) {
	l.mu.Lock()

	defer l.mu.Unlock()

	var ok bool

	if limiter, ok = l.bucket[key]; !ok {
		limiter = l.new(key)
	}

	return limiter
}

func (l *IPRateLimitBucket) FetchCtx(ctx *AutheliaCtx) (limiter *rate.Limiter) {
	return l.Fetch(ctx.RemoteIP().String())
}

func (l *IPRateLimitBucket) new(ip string) *rate.Limiter {
	limiter := rate.NewLimiter(l.r, l.b)

	l.bucket[ip] = limiter

	return limiter
}

func NewRateLimitHandler(config schema.ServerEndpointRateLimit, next RequestHandler) RequestHandler {
	if !config.Enable || len(config.Buckets) == 0 {
		return next
	}

	middleware := NewIPRateLimit(NewRateLimitBucketsConfig(config)...)

	return middleware(next)
}

func NewRateLimit(config schema.ServerEndpointRateLimit) AutheliaMiddleware {
	if !config.Enable || len(config.Buckets) == 0 {
		return nil
	}

	return NewIPRateLimit(NewRateLimitBucketsConfig(config)...)
}

func NewRateLimitBucketsConfig(config schema.ServerEndpointRateLimit) []RateLimitBucketConfig {
	buckets := make([]RateLimitBucketConfig, len(config.Buckets))

	for i, bucket := range config.Buckets {
		buckets[i] = RateLimitBucketConfig{Period: bucket.Period, Requests: bucket.Requests}
	}

	return buckets
}
