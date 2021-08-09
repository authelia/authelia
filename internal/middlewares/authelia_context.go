package middlewares

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"path"
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
func AutheliaMiddleware(configuration schema.Configuration, providers Providers) RequestHandlerBridge {
	return func(next RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			autheliaCtx, err := NewAutheliaCtx(ctx, configuration, providers)
			if err != nil {
				autheliaCtx.Error(err, messageOperationFailed)
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

	c.SetContentType(contentTypeApplicationJSON)
	c.SetBody(b)
	c.Logger.Error(err)
}

// ReplyError reply with an error but does not display any stack trace in the logs.
func (c *AutheliaCtx) ReplyError(err error, message string) {
	b, marshalErr := json.Marshal(ErrorResponse{Status: "KO", Message: message})

	if marshalErr != nil {
		c.Logger.Error(marshalErr)
	}

	c.SetContentType(contentTypeApplicationJSON)
	c.SetBody(b)
	c.Logger.Debug(err)
}

// ReplyUnauthorized response sent when user is unauthorized.
func (c *AutheliaCtx) ReplyUnauthorized() {
	c.RequestCtx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
}

// ReplyForbidden response sent when access is forbidden to user.
func (c *AutheliaCtx) ReplyForbidden() {
	c.RequestCtx.Error(fasthttp.StatusMessage(fasthttp.StatusForbidden), fasthttp.StatusForbidden)
}

// ReplyBadRequest response sent when bad request has been sent.
func (c *AutheliaCtx) ReplyBadRequest() {
	c.RequestCtx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest), fasthttp.StatusBadRequest)
}

// XForwardedProto return the content of the X-Forwarded-Proto header.
func (c *AutheliaCtx) XForwardedProto() []byte {
	return c.RequestCtx.Request.Header.Peek(headerXForwardedProto)
}

// XForwardedMethod return the content of the X-Forwarded-Method header.
func (c *AutheliaCtx) XForwardedMethod() []byte {
	return c.RequestCtx.Request.Header.Peek(headerXForwardedMethod)
}

// XForwardedHost return the content of the X-Forwarded-Host header.
func (c *AutheliaCtx) XForwardedHost() []byte {
	return c.RequestCtx.Request.Header.Peek(headerXForwardedHost)
}

// XForwardedURI return the content of the X-Forwarded-URI header.
func (c *AutheliaCtx) XForwardedURI() []byte {
	return c.RequestCtx.Request.Header.Peek(headerXForwardedURI)
}

// ForwardedProtoHost gets the X-Forwarded-Proto and X-Forwarded-Host headers and forms them into a URL.
func (c AutheliaCtx) ForwardedProtoHost() (string, error) {
	XForwardedProto := c.XForwardedProto()

	if XForwardedProto == nil {
		return "", errMissingXForwardedProto
	}

	XForwardedHost := c.XForwardedHost()

	if XForwardedHost == nil {
		return "", errMissingXForwardedHost
	}

	return fmt.Sprintf("%s://%s", XForwardedProto,
		XForwardedHost), nil
}

// BasePath returns the base_url as per the path visited by the client.
func (c *AutheliaCtx) BasePath() (base string) {
	if baseURL := c.UserValue("base_url"); baseURL != nil {
		return baseURL.(string)
	}

	return base
}

// GetExternalRootURL gets the X-Forwarded-Proto, X-Forwarded-Host headers and the BasePath and forms them into a URL.
func (c *AutheliaCtx) GetExternalRootURL() (string, error) {
	externalRootURL, err := c.ForwardedProtoHost()
	if err != nil {
		return "", err
	}

	if base := c.BasePath(); base != "" {
		externalBaseURL, err := url.Parse(externalRootURL)
		if err != nil {
			return "", err
		}

		externalBaseURL.Path = path.Join(externalBaseURL.Path, base)

		return externalBaseURL.String(), nil
	}

	return externalRootURL, nil
}

// XOriginalURL return the content of the X-Original-URL header.
func (c *AutheliaCtx) XOriginalURL() []byte {
	return c.RequestCtx.Request.Header.Peek(headerXOriginalURL)
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

// ReplyOK is a helper method to reply ok.
func (c *AutheliaCtx) ReplyOK() {
	c.SetContentType(contentTypeApplicationJSON)
	c.SetBody(okMessageBytes)
}

// ParseBody parse the request body into the type of value.
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

// SetJSONBody Set json body.
func (c *AutheliaCtx) SetJSONBody(value interface{}) error {
	b, err := json.Marshal(OKResponse{Status: "OK", Data: value})
	if err != nil {
		return fmt.Errorf("Unable to marshal JSON body")
	}

	c.SetContentType(contentTypeApplicationJSON)
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

// GetOriginalURL extract the URL from the request headers (X-Original-URI or X-Forwarded-* headers).
func (c *AutheliaCtx) GetOriginalURL() (*url.URL, error) {
	originalURL := c.XOriginalURL()
	if originalURL != nil {
		parsedURL, err := url.ParseRequestURI(string(originalURL))
		if err != nil {
			return nil, fmt.Errorf("Unable to parse URL extracted from X-Original-URL header: %v", err)
		}

		c.Logger.Trace("Using X-Original-URL header content as targeted site URL")

		return parsedURL, nil
	}

	forwardedProto := c.XForwardedProto()
	forwardedHost := c.XForwardedHost()
	forwardedURI := c.XForwardedURI()

	if forwardedProto == nil {
		return nil, errMissingXForwardedProto
	}

	if forwardedHost == nil {
		return nil, errMissingXForwardedHost
	}

	var requestURI string

	scheme := forwardedProto
	scheme = append(scheme, protoHostSeparator...)
	requestURI = string(append(scheme,
		append(forwardedHost, forwardedURI...)...))

	parsedURL, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL %s: %v", requestURI, err)
	}

	c.Logger.Tracef("Using X-Fowarded-Proto, X-Forwarded-Host and X-Forwarded-URI headers " +
		"to construct targeted site URL")

	return parsedURL, nil
}

// IsXHR returns true if the request is a XMLHttpRequest.
func (c AutheliaCtx) IsXHR() (xhr bool) {
	requestedWith := c.Request.Header.Peek(headerXRequestedWith)

	return requestedWith != nil && string(requestedWith) == headerValueXRequestedWithXHR
}

// AcceptsMIME takes a mime type and returns true if the request accepts that type or the wildcard type.
func (c AutheliaCtx) AcceptsMIME(mime string) (acceptsMime bool) {
	accepts := strings.Split(string(c.Request.Header.Peek("Accept")), ",")

	for i, accept := range accepts {
		mimeType := strings.Trim(strings.SplitN(accept, ";", 2)[0], " ")
		if mimeType == mime || (i == 0 && mimeType == "*/*") {
			return true
		}
	}

	return false
}

// SpecialRedirect performs a redirect similar to fasthttp.RequestCtx except it allows statusCode 401 and includes body
// content in the form of a link to the location.
func (c *AutheliaCtx) SpecialRedirect(uri string, statusCode int) {
	if statusCode < fasthttp.StatusMovedPermanently || (statusCode > fasthttp.StatusSeeOther && statusCode != fasthttp.StatusTemporaryRedirect && statusCode != fasthttp.StatusPermanentRedirect && statusCode != fasthttp.StatusUnauthorized) {
		statusCode = fasthttp.StatusFound
	}

	c.SetContentType(contentTypeTextHTML)
	c.SetStatusCode(statusCode)

	u := fasthttp.AcquireURI()

	c.URI().CopyTo(u)
	u.Update(uri)

	c.Response.Header.SetBytesV("Location", u.FullURI())

	c.SetBodyString(fmt.Sprintf("<a href=\"%s\">%s</a>", utils.StringHTMLEscape(string(u.FullURI())), fasthttp.StatusMessage(statusCode)))

	fasthttp.ReleaseURI(u)
}
