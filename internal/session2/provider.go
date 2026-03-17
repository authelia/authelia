package session2

import (
	"context"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/valyala/fasthttp"
	"strings"
	"sync"
	"time"
)

// StorageProvider interface implemented by Provider storage backends.
type StorageProvider interface {
	// GetSessionByPrivateID returns a session given a private ID which is the cookie value.
	GetSessionByPrivateID(ctx context.Context, id []byte) (raw []byte, err error)
	GetSessionByPublicID(ctx context.Context, id string) (raw []byte, err error)
	GetSessionByInternalID(ctx context.Context, id string) (raw []byte, err error)
	GetSessionIDsByUsername(ctx context.Context, username string) (ids [][]byte, err error)
	SaveSession(ctx context.Context, id, data []byte, publicID, internalID string, expiration time.Duration) (err error)
	DestroySession(ctx context.Context, id []byte, publicID, internalID string) error
	RegenerateSession(ctx context.Context, oldID, id []byte, publicID, internalID string, expiration time.Duration) (err error)
	CountSession(ctx context.Context) (count int)
	NeedSessionGC() (needed bool)
	SessionGC() (err error)
}

type Provider interface {
}

func New(config *schema.Session, cookie *schema.SessionCookie, store StorageProvider, rand random.Provider) (provider Provider) {
	encoder, err := NewEncoder(rand, []byte(config.Secret), nil)

	p := &ProductionProvider{
		store: store,
		config: struct {
			Name     string
			Domain   string
			SameSite fasthttp.CookieSameSite
			MaxAge   time.Duration
		}{
			Name:   cookie.Name,
			Domain: cookie.Domain,
		},
	}

	switch strings.ToLower(cookie.SameSite) {
	case "strict":
		p.config.SameSite = fasthttp.CookieSameSiteStrictMode
	case "none":
		p.config.SameSite = fasthttp.CookieSameSiteNoneMode
	case "lax":
		p.config.SameSite = fasthttp.CookieSameSiteLaxMode
	default:
		p.config.SameSite = fasthttp.CookieSameSiteLaxMode
	}

	if store.NeedSessionGC() {
		go p.startGC()
	}

	return
}

type Session struct {
	ID         string
	PublicID   string
	InternalID string
	Data       []byte
}

type ProductionProvider struct {
	store      StorageProvider
	random     random.Provider
	pool       sync.Pool
	encoder    *Encoder
	stopGCChan chan struct{}
	config     struct {
		Name     string
		Domain   string
		SameSite fasthttp.CookieSameSite
		MaxAge   time.Duration
	}
}

func (p *ProductionProvider) startGC() {
	p.stopGCChan = make(chan struct{})

	ticker := time.NewTicker(1 * time.Minute)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.store.SessionGC(); err != nil {
				logging.Logger().WithError(err).Error("Error occurred during session garbage collection")
			}
		case <-p.stopGCChan:
			return
		}
	}
}

func (p *ProductionProvider) stopGC() {
	if p.stopGCChan != nil {
		p.stopGCChan <- struct{}{}
	}
}

func (p *ProductionProvider) Save(ctx *middlewares.AutheliaCtx, session Session) (err error) {
	return p.SaveWithExpiration(ctx, session, p.config.MaxAge)
}

func (p *ProductionProvider) SaveWithExpiration(ctx *middlewares.AutheliaCtx, session Session, expiration time.Duration) (err error) {
	id := p.httpGet(ctx)

	if len(id) == 0 {
		id = genSessionID(p.random)
	}

	sid, encoded, err := p.encoder.Encode(id, session)
	if err != nil {
		return err
	}

	if err = p.store.SaveSession(ctx, sid, encoded, session.PublicID, session.InternalID, storeExp(expiration)); err != nil {
		return err
	}

	p.httpSet(ctx, id, expiration)

	return nil
}

func (p *ProductionProvider) httpSet(ctx *middlewares.AutheliaCtx, id []byte, expiration time.Duration) {
	cookie := fasthttp.AcquireCookie()

	cookie.SetKey(p.config.Name)
	cookie.SetPath("/")
	cookie.SetHTTPOnly(true)
	cookie.SetDomain(p.config.Domain)
	cookie.SetValueBytes(id)
	cookie.SetSameSite(p.config.SameSite)
	cookie.SetSecure(true)

	if expiration >= 0 {
		if expiration == 0 {
			cookie.SetExpire(fasthttp.CookieExpireUnlimited)
		} else {
			cookie.SetExpire(time.Now().Add(expiration))
		}
	}

	ctx.Request.Header.SetCookieBytesKV(cookie.Key(), cookie.Value())
	ctx.Response.Header.SetCookie(cookie)

	fasthttp.ReleaseCookie(cookie)
}

func (p *ProductionProvider) httpDelete(ctx *middlewares.AutheliaCtx) {
	ctx.Request.Header.DelCookie(p.config.Name)
	ctx.Response.Header.DelCookie(p.config.Name)

	cookie := fasthttp.AcquireCookie()

	cookie.SetKey(p.config.Name)
	cookie.SetValue("")
	cookie.SetPath("/")
	cookie.SetHTTPOnly(true)
	cookie.SetExpire(time.Now().Add(-1 * time.Minute))

	ctx.Response.Header.SetCookie(cookie)

	fasthttp.ReleaseCookie(cookie)
}

func (p *ProductionProvider) httpGet(ctx *middlewares.AutheliaCtx) (id []byte) {
	if id = ctx.Request.Header.Cookie(p.config.Name); len(id) > 0 {
		return id
	}

	return nil
}
