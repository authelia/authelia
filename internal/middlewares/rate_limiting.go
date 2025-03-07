package middlewares

import (
	"net/http"
	"sync"
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

// NewIPRateLimit given a series of RateLimitBucketConfig items produces an AutheliaMiddleware which handles requests based
// on the IPRateLimitBucket.
func NewIPRateLimit(bs ...RateLimitBucketConfig) AutheliaMiddleware {
	return NewRateLimiter(NewIPRateLimitBucket, HandlerRateLimitAPI, bs...)
}

type NewRateLimiterFunc func(bucket RateLimitBucketConfig) RateLimitBucket

func HandlerRateLimitAPI(ctx *AutheliaCtx) {
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

	ticker := time.NewTicker(time.Minute * 30)

	go func() {
		for range ticker.C {
			for _, bucket := range buckets {
				bucket.GC()
			}
		}
	}()

	return func(next RequestHandler) RequestHandler {
		return func(ctx *AutheliaCtx) {
			var (
				retryAfter time.Duration
			)

			for i, bucket := range buckets {
				limiter := bucket.FetchCtx(ctx)

				if !limiter.Allow() {
					reservation := limiter.ReserveN(time.Now().UTC(), 1)
					limiter.updated = time.Now().UTC()

					ctx.Logger.WithFields(map[string]any{"bucket": i + 1, "delay": reservation.Delay().Seconds()}).Warn("Rate Limit Exceeded")

					if reservation.Delay() > retryAfter {
						retryAfter = reservation.Delay()
					}

					reservation.Cancel()
				}
			}

			if retryAfter > 0 {
				ctx.Response.Header.SetBytesK(headerRetryAfter, time.Now().UTC().Add(retryAfter).Format(http.TimeFormat))
				ctx.SetStatusCode(fasthttp.StatusTooManyRequests)

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
		bucket: make(map[string]*RateLimiter),
		mu:     sync.Mutex{},
		p:      bucket.Period,
		r:      rate.Every(bucket.Period),
		b:      bucket.Requests,
	}
}

type RateLimiter struct {
	*rate.Limiter

	updated time.Time
}

// IPRateLimitBucket is a RateLimitBucket which limits requests based on each of the buckets delimited by IP.
type IPRateLimitBucket struct {
	bucket map[string]*RateLimiter
	mu     sync.Mutex
	p      time.Duration
	r      rate.Limit
	b      int
}

func (l *IPRateLimitBucket) Fetch(key string) (limiter *RateLimiter) {
	l.mu.Lock()

	defer l.mu.Unlock()

	var ok bool

	if limiter, ok = l.bucket[key]; !ok {
		limiter = l.new(key)
	}

	return limiter
}

func (l *IPRateLimitBucket) GC() {
	if len(l.bucket) == 0 {
		return
	}

	l.mu.Lock()

	defer l.mu.Unlock()

	for k, limiter := range l.bucket {
		if limiter.updated.Add(l.p).Before(time.Now().UTC()) {
			delete(l.bucket, k)
		}
	}
}

func (l *IPRateLimitBucket) FetchCtx(ctx *AutheliaCtx) (limiter *RateLimiter) {
	return l.Fetch(ctx.RemoteIP().String())
}

func (l *IPRateLimitBucket) new(ip string) (limiter *RateLimiter) {
	limiter = &RateLimiter{Limiter: rate.NewLimiter(l.r, l.b), updated: time.Now().UTC()}

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
