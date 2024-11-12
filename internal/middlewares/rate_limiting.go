package middlewares

import (
	"sync"

	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

func NewIPRateLimit(r float64, b int) AutheliaMiddleware {
	bucket := NewIPRateLimitBucket(r, b)

	return func(next RequestHandler) RequestHandler {
		return func(ctx *AutheliaCtx) {
			limiter := bucket.FetchCtx(ctx)

			if !limiter.Allow() {
				ctx.SetStatusCode(fasthttp.StatusTooManyRequests)

				return
			}

			next(ctx)
		}
	}
}

type IPRateLimitBucket struct {
	bucket map[string]*rate.Limiter
	mu     *sync.Mutex
	r      rate.Limit
	b      int
}

func NewIPRateLimitBucket(r float64, b int) (limiter *IPRateLimitBucket) {
	return &IPRateLimitBucket{
		bucket: make(map[string]*rate.Limiter),
		mu:     &sync.Mutex{},
		r:      rate.Limit(r),
		b:      b,
	}
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
