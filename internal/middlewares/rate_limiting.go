package middlewares

import (
	"sync"

	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

// RateLimitBucket describes an implementation of a bucket which can be leveraged for rate limiting.
type RateLimitBucket interface {
	FetchCtx(ctx *AutheliaCtx) (limiter *rate.Limiter)
}

// The RateLimitBucketConfig describes a limit (number of seconds), and a burst (number of events) that can occur for a
// given rate limiter.
type RateLimitBucketConfig struct {
	Limit float64
	Burst int
}

// NewIPRateLimit given a series of RateLimitBucketConfig items produces an AutheliaMiddleware which handles requests based
// on the IPRateLimitBucket.
func NewIPRateLimit(bs ...RateLimitBucketConfig) AutheliaMiddleware {
	buckets := make([]RateLimitBucket, len(bs))

	for i, b := range bs {
		buckets[i] = NewIPRateLimitBucket(b)
	}

	return func(next RequestHandler) RequestHandler {
		return func(ctx *AutheliaCtx) {
			var exceeded bool

			for _, bucket := range buckets {
				limiter := bucket.FetchCtx(ctx)

				if !limiter.Allow() {
					exceeded = true
				}
			}

			if exceeded {
				ctx.Logger.Warn("Rate Limit Exceeded")

				ctx.SetStatusCode(fasthttp.StatusTooManyRequests)
				ctx.SetJSONError(fasthttp.StatusMessage(fasthttp.StatusTooManyRequests))

				return
			}

			next(ctx)
		}
	}
}

// NewIPRateLimitBucket returns a IPRateLimitBucket given a RateLimitBucketConfig.
func NewIPRateLimitBucket(bucket RateLimitBucketConfig) (limiter *IPRateLimitBucket) {
	return &IPRateLimitBucket{
		bucket: make(map[string]*rate.Limiter),
		mu:     sync.Mutex{},
		r:      rate.Limit(bucket.Limit),
		b:      bucket.Burst,
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

func NewRateLimitFirstFactor() AutheliaMiddleware {
	config := []RateLimitBucketConfig{
		{Limit: 60, Burst: 10},
		{Limit: 120, Burst: 15},
	}

	return NewIPRateLimit(config...)
}

func NewRateLimitResetPassword() AutheliaMiddleware {
	config := []RateLimitBucketConfig{
		{Limit: 20, Burst: 5},
		{Limit: 60, Burst: 10},
	}

	return NewIPRateLimit(config...)
}

func NewRateLimitResetPasswordStart() AutheliaMiddleware {
	config := []RateLimitBucketConfig{
		{Limit: 600, Burst: 5},
		{Limit: 900, Burst: 10},
	}

	return NewIPRateLimit(config...)
}

func NewRateLimitTOTP() AutheliaMiddleware {
	config := []RateLimitBucketConfig{
		{Limit: 60, Burst: 30},
		{Limit: 120, Burst: 40},
	}

	return NewIPRateLimit(config...)
}

func NewRateLimitDUO() AutheliaMiddleware {
	config := []RateLimitBucketConfig{
		{Limit: 60, Burst: 10},
		{Limit: 120, Burst: 15},
	}

	return NewIPRateLimit(config...)
}
