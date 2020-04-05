package middlewares

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/utils"
)

// NewRequestLogger create a new request logger for the given request.
func NewRequestLogger(ctx *AutheliaCtx) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"method":    string(ctx.Method()),
		"path":      string(ctx.Path()),
		"remote_ip": ctx.RemoteIP().String(),
	})
}

// NewAutheliaCtx instantiate an AutheliaCtx out of a RequestCtx.
func NewAutheliaCtx(ctx *fasthttp.RequestCtx, configuration schema.Configuration, providers Providers) (*AutheliaCtx, error) {
	autheliaCtx := new(AutheliaCtx)
	autheliaCtx.RequestCtx = ctx
	autheliaCtx.Providers = providers
	autheliaCtx.Configuration = configuration
	autheliaCtx.Logger = NewRequestLogger(autheliaCtx)
	autheliaCtx.Clock = utils.RealClock{}
	return autheliaCtx, nil
}

// AutheliaMiddleware is wrapping the RequestCtx into an AutheliaCtx providing Authelia related objects.
func AutheliaMiddleware(configuration schema.Configuration, providers Providers) func(next RequestHandler) fasthttp.RequestHandler {
	return func(next RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			autheliaCtx, err := NewAutheliaCtx(ctx, configuration, providers)
			if err != nil {
				autheliaCtx.Error(err, operationFailedMessage)
				return
			}
			next(autheliaCtx)
		}
	}
}

// Error reply with an error and display the stack trace in the logs.
func (c *AutheliaCtx) Error(err error, message string) {
	b, marshalErr := json.Marshal(ErrorResponse{Status: "KO", Message: message})

	if marshalErr != nil {
		c.Logger.Error(marshalErr)
	}

	c.SetContentType("application/json")
	c.SetBody(b)
	c.Logger.Error(err)
}

// ReplyError reply with an error but does not display any stack trace in the logs
func (c *AutheliaCtx) ReplyError(err error, message string) {
	b, marshalErr := json.Marshal(ErrorResponse{Status: "KO", Message: message})

	if marshalErr != nil {
		c.Logger.Error(marshalErr)
	}

	c.SetContentType("application/json")
	c.SetBody(b)
	c.Logger.Debug(err)
}

// ReplyUnauthorized response sent when user is unauthorized
func (c *AutheliaCtx) ReplyUnauthorized() {
	c.RequestCtx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
	// c.Response.Header.Set("WWW-Authenticate", "Basic realm=Restricted")
}

// ReplyForbidden response sent when access is forbidden to user
func (c *AutheliaCtx) ReplyForbidden() {
	c.RequestCtx.Error(fasthttp.StatusMessage(fasthttp.StatusForbidden), fasthttp.StatusForbidden)
}

// XForwardedProto return the content of the header X-Forwarded-Proto
func (c *AutheliaCtx) XForwardedProto() []byte {
	return c.RequestCtx.Request.Header.Peek(xForwardedProtoHeader)
}

// XForwardedHost return the content of the header X-Forwarded-Host
func (c *AutheliaCtx) XForwardedHost() []byte {
	return c.RequestCtx.Request.Header.Peek(xForwardedHostHeader)
}

// XForwardedURI return the content of the header X-Forwarded-URI
func (c *AutheliaCtx) XForwardedURI() []byte {
	return c.RequestCtx.Request.Header.Peek(xForwardedURIHeader)
}

// XOriginalURL return the content of the header X-Original-URL
func (c *AutheliaCtx) XOriginalURL() []byte {
	return c.RequestCtx.Request.Header.Peek(xOriginalURLHeader)
}

// GetSession return the user session. Any update will be saved in cache.
func (c *AutheliaCtx) GetSession() session.UserSession {
	userSession, err := c.Providers.SessionProvider.GetSession(c.RequestCtx)
	if err != nil {
		c.Logger.Error("Unable to retrieve user session")
		return session.NewDefaultUserSession()
	}
	return userSession
}

// SaveSession save the content of the session.
func (c *AutheliaCtx) SaveSession(userSession session.UserSession) error {
	return c.Providers.SessionProvider.SaveSession(c.RequestCtx, userSession)
}

// ReplyOK is a helper method to reply ok
func (c *AutheliaCtx) ReplyOK() {
	c.SetContentType(applicationJSONContentType)
	c.SetBody(okMessageBytes)
}

// ParseBody parse the request body into the type of value
func (c *AutheliaCtx) ParseBody(value interface{}) error {
	err := json.Unmarshal(c.PostBody(), &value)

	if err != nil {
		return fmt.Errorf("Unable to parse body: %s", err)
	}

	valid, err := govalidator.ValidateStruct(value)

	if err != nil {
		return fmt.Errorf("Unable to validate body: %s", err)
	}

	if !valid {
		return fmt.Errorf("Body is not valid")
	}
	return nil
}

// SetJSONBody Set json body
func (c *AutheliaCtx) SetJSONBody(value interface{}) error {
	b, err := json.Marshal(OKResponse{Status: "OK", Data: value})
	if err != nil {
		return fmt.Errorf("Unable to marshal JSON body")
	}

	c.SetContentType("application/json")
	c.SetBody(b)
	return nil
}

// RemoteIP return the remote IP taking X-Forwarded-For header into account if provided.
func (c *AutheliaCtx) RemoteIP() net.IP {
	XForwardedFor := c.Request.Header.Peek("X-Forwarded-For")
	if XForwardedFor != nil {
		ips := strings.Split(string(XForwardedFor), ",")

		if len(ips) > 0 {
			return net.ParseIP(strings.Trim(ips[0], " "))
		}
	}
	return c.RequestCtx.RemoteIP()
}
