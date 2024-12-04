package middlewares

import (
	"sync"

	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

// The RateLimitBucket describes a limit (number of seconds), and a burst (number of events) that can occur for a given
// rate limiter.
type RateLimitBucket struct {
	Limit float64
	Burst int
}

// NewIPRateLimit given a series of RateLimitBucket items produces an AutheliaMiddleware which handles requests based
// on the IPRateLimitBucket.
func NewIPRateLimit(bs ...RateLimitBucket) AutheliaMiddleware {
	buckets := make([]*IPRateLimitBucket, len(bs))

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

				return
			}

			next(ctx)
		}
	}
}

// NewIPRateLimitBucket returns a IPRateLimitBucket given a RateLimitBucket.
func NewIPRateLimitBucket(bucket RateLimitBucket) (limiter *IPRateLimitBucket) {
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

func (l *IPRateLimitBucket) Fetch(ip string) (limiter *rate.Limiter) {
	l.mu.Lock()

	defer l.mu.Unlock()

	var ok bool

	if limiter, ok = l.bucket[ip]; !ok {
		limiter = l.new(ip)
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
