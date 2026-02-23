package middlewares

import (
	"context"
	"net"
	"net/url"
	"strings"

	"github.com/go-spop/spop/message"
	"github.com/go-spop/spop/request"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
)

// NewSPOPRequestLogger create a new request logger for the given request.
func NewSPOPRequestLogger(request *request.Request) (entry *logrus.Entry) {
	fields := logrus.Fields{
		// logging.FieldMethod:   string(ctx.Method()),
		// logging.FieldRemoteIP: RequestCtxRemoteIP(ctx).String(),
		// logging.FieldPath:     string(ctx.Path()),
	}

	/*
		if uri, ok := ctx.UserValue(UserValueKeyRawURI).(string); ok {
			fields[logging.FieldPathRaw] = uri
		}
	*/

	return logging.Logger().WithFields(fields)
}

func NewAutheliaSPOPCtx(request *request.Request, message *message.Message, configuration schema.Configuration, providers Providers) *AutheliaSPOPCtx {
	return &AutheliaSPOPCtx{
		Context:       context.Background(),
		Request:       request,
		Message:       message,
		Configuration: configuration,
		Providers:     providers,
		Logger:        NewSPOPRequestLogger(request),
		Clock:         clock.New(),
	}
}

type AutheliaSPOPCtx struct {
	context.Context

	Request       *request.Request
	Message       *message.Message
	Configuration schema.Configuration
	Providers     Providers
	Clock         clock.Provider
	Logger        *logrus.Entry
}

func (ctx AutheliaSPOPCtx) GetLogger() *logrus.Entry {
	return ctx.Logger
}

func (ctx AutheliaSPOPCtx) GetConfiguration() schema.Configuration {
	return ctx.Configuration
}

func (ctx AutheliaSPOPCtx) GetClock() clock.Provider {
	return ctx.Clock
}

func (ctx AutheliaSPOPCtx) GetProviders() Providers {
	return ctx.Providers
}

func (ctx AutheliaSPOPCtx) GetUserProvider() authentication.UserProvider {
	return ctx.Providers.UserProvider
}

func (ctx AutheliaSPOPCtx) GetRandom() (random random.Provider) {
	return ctx.Providers.Random
}

func (ctx AutheliaSPOPCtx) GetProviderUserAttributeResolver() expression.UserAttributeResolver {
	return ctx.Providers.UserAttributeResolver
}

func (ctx AutheliaSPOPCtx) GetJWTWithTimeFuncOption() (option jwt.ParserOption) {
	return jwt.WithTimeFunc(ctx.Clock.Now)
}

func (ctx AutheliaSPOPCtx) Method() (method []byte) {
	value, ok := getStringSPOE("method", ctx.Message)
	if !ok {
		return nil
	}

	return []byte(value)
}

func (ctx AutheliaSPOPCtx) Host() (host []byte) {
	value, ok := getStringSPOE("host", ctx.Message)
	if !ok {
		return nil
	}

	return []byte(value)
}

func (ctx AutheliaSPOPCtx) XForwardedMethod() (method []byte) {
	if value, ok := getStringSPOE("xfm", ctx.Message); !ok {
		return nil
	} else {
		return []byte(value)
	}
}

func (ctx AutheliaSPOPCtx) XForwardedProto() (proto []byte) {
	if value, ok := getStringSPOE("xfp", ctx.Message); !ok {
		return nil
	} else {
		return []byte(value)
	}
}

func (ctx AutheliaSPOPCtx) XForwardedHost() (host []byte) {
	if value, ok := getStringSPOE("xfh", ctx.Message); !ok {
		return nil
	} else {
		return []byte(value)
	}
}

func (ctx AutheliaSPOPCtx) XForwardedURI() (uri []byte) {
	if value, ok := getStringSPOE("xfu", ctx.Message); !ok {
		return nil
	} else {
		return []byte(value)
	}
}

func (ctx AutheliaSPOPCtx) XOriginalMethod() (method []byte) {
	if value, ok := getStringSPOE("xom", ctx.Message); !ok {
		return nil
	} else {
		return []byte(value)
	}
}

func (ctx AutheliaSPOPCtx) XOriginalURL() (uri []byte) {
	if value, ok := getStringSPOE("xou", ctx.Message); !ok {
		return nil
	} else {
		return []byte(value)
	}
}

func (ctx AutheliaSPOPCtx) GetXOriginalURLOrXForwardedURL() (requestURI *url.URL, err error) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) XAutheliaURL() []byte {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) QueryArgAutheliaURL() []byte {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) RootURL() (issuerURL *url.URL) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) IssuerURL() (issuerURL *url.URL, err error) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) GetSessionManagerByTargetURI(targetURL *url.URL) (manager session.Manager, err error) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) GetRequestQueryArgValue(key []byte) (value []byte) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) GetRequestHeaderValue(key []byte) (value []byte) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) SetResponseHeaderValue(key []byte, value string) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) SetResponseHeaderValueBytes(key, value []byte) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) AuthzPath() (uri []byte) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) IsXHR() (xhr bool) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) AcceptsMIME(mime string) (ok bool) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) ReplyStatusCode(statusCode int) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) ReplyUnauthorized() {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) ReplyForbidden() {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) SpecialRedirect(uri string, statusCode int) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) SpecialRedirectNoBody(uri string, statusCode int) {
	//TODO implement me
	panic("implement me")
}

func (ctx AutheliaSPOPCtx) RecordAuthn(success, banned bool, authType string) {
	if ctx.Providers.Metrics == nil {
		return
	}

	ctx.Providers.Metrics.RecordAuthn(success, banned, authType)
}

func (ctx AutheliaSPOPCtx) ip() net.IP {
	value, ok := getStringSPOE("ip", ctx.Message)
	if !ok {
		return nil
	}

	return net.ParseIP(value)
}

func (ctx AutheliaSPOPCtx) RemoteIP() net.IP {
	if header, ok := getStringSPOE("xff", ctx.Message); ok {
		ips := strings.SplitN(header, ",", 2)
		if len(ips) != 0 {
			if ip := net.ParseIP(strings.Trim(ips[0], " ")); ip != nil {
				return ip
			}
		}
	}

	return ctx.ip()
}

func (ctx AutheliaSPOPCtx) GetSessionProviderByTargetURI(targetURL *url.URL) (provider *session.Session, err error) {
	//TODO implement me
	panic("implement me")
}

func getStringSPOE(key string, message *message.Message) (value string, ok bool) {
	v, ok := message.KV.Get(key)
	if !ok {
		return "", false
	}

	m, ok := v.(string)
	if !ok {
		return "", false
	}

	return m, true
}
