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

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewRequestLogger create a new request logger for the given request.
func NewRequestLogger(ctx *AutheliaCtx) *logrus.Entry {
	return logging.Logger().WithFields(logrus.Fields{
		"method":    string(ctx.Method()),
		"path":      string(ctx.Path()),
		"remote_ip": ctx.RemoteIP().String(),
	})
}

// NewAutheliaCtx instantiate an AutheliaCtx out of a RequestCtx.
func NewAutheliaCtx(requestCTX *fasthttp.RequestCtx, configuration schema.Configuration, providers Providers) (ctx *AutheliaCtx) {
	ctx = new(AutheliaCtx)
	ctx.RequestCtx = requestCTX
	ctx.Providers = providers
	ctx.Configuration = configuration
	ctx.Logger = NewRequestLogger(ctx)
	ctx.Clock = utils.RealClock{}

	return ctx
}

// AvailableSecondFactorMethods returns the available 2FA methods.
func (ctx *AutheliaCtx) AvailableSecondFactorMethods() (methods []string) {
	methods = make([]string, 0, 3)

	if !ctx.Configuration.TOTP.Disable {
		methods = append(methods, model.SecondFactorMethodTOTP)
	}

	if !ctx.Configuration.Webauthn.Disable {
		methods = append(methods, model.SecondFactorMethodWebauthn)
	}

	if !ctx.Configuration.DuoAPI.Disable {
		methods = append(methods, model.SecondFactorMethodDuo)
	}

	return methods
}

// Error reply with an error and display the stack trace in the logs.
func (ctx *AutheliaCtx) Error(err error, message string) {
	ctx.SetJSONError(message)

	ctx.Logger.Error(err)
}

// SetJSONError sets the body of the response to an JSON error KO message.
func (ctx *AutheliaCtx) SetJSONError(message string) {
	if replyErr := ctx.ReplyJSON(ErrorResponse{Status: "KO", Message: message}, 0); replyErr != nil {
		ctx.Logger.Error(replyErr)
	}
}

// ReplyError reply with an error but does not display any stack trace in the logs.
func (ctx *AutheliaCtx) ReplyError(err error, message string) {
	b, marshalErr := json.Marshal(ErrorResponse{Status: "KO", Message: message})

	if marshalErr != nil {
		ctx.Logger.Error(marshalErr)
	}

	ctx.SetContentTypeBytes(contentTypeApplicationJSON)
	ctx.SetBody(b)
	ctx.Logger.Debug(err)
}

// ReplyStatusCode resets a response and replies with the given status code and relevant message.
func (ctx *AutheliaCtx) ReplyStatusCode(statusCode int) {
	ctx.Response.Reset()
	ctx.SetStatusCode(statusCode)
	ctx.SetContentTypeBytes(contentTypeTextPlain)
	ctx.SetBodyString(fmt.Sprintf("%d %s", statusCode, fasthttp.StatusMessage(statusCode)))
}

// ReplyJSON writes a JSON response.
func (ctx *AutheliaCtx) ReplyJSON(data interface{}, statusCode int) (err error) {
	var (
		body []byte
	)

	if body, err = json.Marshal(data); err != nil {
		return fmt.Errorf("unable to marshal JSON body: %w", err)
	}

	if statusCode > 0 {
		ctx.SetStatusCode(statusCode)
	}

	ctx.SetContentTypeBytes(contentTypeApplicationJSON)
	ctx.SetBody(body)

	return nil
}

// ReplyUnauthorized response sent when user is unauthorized.
func (ctx *AutheliaCtx) ReplyUnauthorized() {
	ctx.ReplyStatusCode(fasthttp.StatusUnauthorized)
}

// ReplyForbidden response sent when access is forbidden to user.
func (ctx *AutheliaCtx) ReplyForbidden() {
	ctx.ReplyStatusCode(fasthttp.StatusForbidden)
}

// ReplyBadRequest response sent when bad request has been sent.
func (ctx *AutheliaCtx) ReplyBadRequest() {
	ctx.ReplyStatusCode(fasthttp.StatusBadRequest)
}

// XForwardedProto return the content of the X-Forwarded-Proto header.
func (ctx *AutheliaCtx) XForwardedProto() (proto []byte) {
	proto = ctx.RequestCtx.Request.Header.PeekBytes(headerXForwardedProto)

	if proto == nil {
		if ctx.RequestCtx.IsTLS() {
			return protoHTTPS
		}

		return protoHTTP
	}

	return proto
}

// XForwardedMethod return the content of the X-Forwarded-Method header.
func (ctx *AutheliaCtx) XForwardedMethod() (method []byte) {
	return ctx.RequestCtx.Request.Header.PeekBytes(headerXForwardedMethod)
}

// XForwardedHost return the content of the X-Forwarded-Host header.
func (ctx *AutheliaCtx) XForwardedHost() (host []byte) {
	host = ctx.RequestCtx.Request.Header.PeekBytes(headerXForwardedHost)

	if host == nil {
		return ctx.RequestCtx.Host()
	}

	return host
}

// XForwardedURI return the content of the X-Forwarded-URI header.
func (ctx *AutheliaCtx) XForwardedURI() (uri []byte) {
	uri = ctx.RequestCtx.Request.Header.PeekBytes(headerXForwardedURI)

	if len(uri) == 0 {
		return ctx.RequestCtx.RequestURI()
	}

	return uri
}

// BasePath returns the base_url as per the path visited by the client.
func (ctx *AutheliaCtx) BasePath() (base string) {
	if baseURL := ctx.UserValueBytes(UserValueKeyBaseURL); baseURL != nil {
		return baseURL.(string)
	}

	return base
}

// ExternalRootURL gets the X-Forwarded-Proto, X-Forwarded-Host headers and the BasePath and forms them into a URL.
func (ctx *AutheliaCtx) ExternalRootURL() (string, error) {
	protocol := ctx.XForwardedProto()
	if protocol == nil {
		return "", errMissingXForwardedProto
	}

	host := ctx.XForwardedHost()
	if host == nil {
		return "", errMissingXForwardedHost
	}

	externalRootURL := fmt.Sprintf("%s://%s", protocol, host)

	if base := ctx.BasePath(); base != "" {
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
func (ctx *AutheliaCtx) XOriginalURL() []byte {
	return ctx.RequestCtx.Request.Header.PeekBytes(headerXOriginalURL)
}

// GetDomainSession returns the session provider for the Request's domain.
func (ctx *AutheliaCtx) GetDomainSession() (*session.Session, error) {
	targetURL, err := ctx.GetOriginalURL()

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve domain session: %s", err)
	}

	isUnderProtectedDomain, cookieDomain := utils.IsURLUnderProtectedDomain(targetURL, ctx.Configuration.Session.Domains)
	if !isUnderProtectedDomain || cookieDomain == "" {
		return nil, fmt.Errorf("unable to retrieve domain session: %s", err)
	}

	return ctx.Providers.SessionProvider.Get(cookieDomain)
}

// GetSession return the user session. Any update will be saved in cache.
func (ctx *AutheliaCtx) GetSession() session.UserSession {
	domainSession, err := ctx.GetDomainSession()
	if err != nil {
		ctx.Logger.Error("Unable to retrieve domain session")
		return session.NewDefaultUserSession()
	}

	userSession, err := domainSession.GetSession(ctx.RequestCtx)
	if err != nil {
		ctx.Logger.Error("Unable to retrieve user session")
		return session.NewDefaultUserSession()
	}

	return userSession
}

// SaveSession save the content of the session.
func (ctx *AutheliaCtx) SaveSession(userSession session.UserSession) error {
	domainSession, err := ctx.GetDomainSession()
	if err != nil {
		return fmt.Errorf("unable to save user session: %s", err)
	}

	return domainSession.SaveSession(ctx.RequestCtx, userSession)
}

// RegenerateSession regenereates user session.
func (ctx *AutheliaCtx) RegenerateSession() error {
	domainSession, err := ctx.GetDomainSession()
	if err != nil {
		return fmt.Errorf("unable to regenerate user session: %s", err)
	}

	return domainSession.RegenerateSession(ctx.RequestCtx)
}

// DestroySession destroy user session.
func (ctx *AutheliaCtx) DestroySession() error {
	domainSession, err := ctx.GetDomainSession()
	if err != nil {
		return fmt.Errorf("unable to destroy user session: %s", err)
	}

	return domainSession.DestroySession(ctx.RequestCtx)
}

// ReplyOK is a helper method to reply ok.
func (ctx *AutheliaCtx) ReplyOK() {
	ctx.SetContentTypeBytes(contentTypeApplicationJSON)
	ctx.SetBody(okMessageBytes)
}

// ParseBody parse the request body into the type of value.
func (ctx *AutheliaCtx) ParseBody(value interface{}) error {
	err := json.Unmarshal(ctx.PostBody(), &value)

	if err != nil {
		return fmt.Errorf("unable to parse body: %w", err)
	}

	valid, err := govalidator.ValidateStruct(value)

	if err != nil {
		return fmt.Errorf("unable to validate body: %w", err)
	}

	if !valid {
		return fmt.Errorf("Body is not valid")
	}

	return nil
}

// SetJSONBody Set json body.
func (ctx *AutheliaCtx) SetJSONBody(value interface{}) error {
	return ctx.ReplyJSON(OKResponse{Status: "OK", Data: value}, 0)
}

// RemoteIP return the remote IP taking X-Forwarded-For header into account if provided.
func (ctx *AutheliaCtx) RemoteIP() net.IP {
	XForwardedFor := ctx.Request.Header.PeekBytes(headerXForwardedFor)
	if XForwardedFor != nil {
		ips := strings.Split(string(XForwardedFor), ",")

		if len(ips) > 0 {
			return net.ParseIP(strings.Trim(ips[0], " "))
		}
	}

	return ctx.RequestCtx.RemoteIP()
}

// GetOriginalURL extract the URL from the request headers (X-Original-URL or X-Forwarded-* headers).
func (ctx *AutheliaCtx) GetOriginalURL() (*url.URL, error) {
	originalURL := ctx.XOriginalURL()
	if originalURL != nil {
		parsedURL, err := url.ParseRequestURI(string(originalURL))
		if err != nil {
			return nil, fmt.Errorf("Unable to parse URL extracted from X-Original-URL header: %v", err)
		}

		ctx.Logger.Trace("Using X-Original-URL header content as targeted site URL")

		return parsedURL, nil
	}

	forwardedProto, forwardedHost, forwardedURI := ctx.XForwardedProto(), ctx.XForwardedHost(), ctx.XForwardedURI()

	if forwardedProto == nil {
		return nil, errMissingXForwardedProto
	}

	if forwardedHost == nil {
		return nil, errMissingXForwardedHost
	}

	var requestURI string

	forwardedProto = append(forwardedProto, protoHostSeparator...)
	requestURI = string(append(forwardedProto,
		append(forwardedHost, forwardedURI...)...))

	parsedURL, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL %s: %v", requestURI, err)
	}

	ctx.Logger.Tracef("Using X-Fowarded-Proto, X-Forwarded-Host and X-Forwarded-URI headers " +
		"to construct targeted site URL")

	return parsedURL, nil
}

// IsXHR returns true if the request is a XMLHttpRequest.
func (ctx AutheliaCtx) IsXHR() (xhr bool) {
	requestedWith := ctx.Request.Header.PeekBytes(headerXRequestedWith)

	return requestedWith != nil && strings.EqualFold(string(requestedWith), headerValueXRequestedWithXHR)
}

// AcceptsMIME takes a mime type and returns true if the request accepts that type or the wildcard type.
func (ctx AutheliaCtx) AcceptsMIME(mime string) (acceptsMime bool) {
	accepts := strings.Split(string(ctx.Request.Header.PeekBytes(headerAccept)), ",")

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
func (ctx *AutheliaCtx) SpecialRedirect(uri string, statusCode int) {
	if statusCode < fasthttp.StatusMovedPermanently || (statusCode > fasthttp.StatusSeeOther && statusCode != fasthttp.StatusTemporaryRedirect && statusCode != fasthttp.StatusPermanentRedirect && statusCode != fasthttp.StatusUnauthorized) {
		statusCode = fasthttp.StatusFound
	}

	ctx.SetContentTypeBytes(contentTypeTextHTML)
	ctx.SetStatusCode(statusCode)

	u := fasthttp.AcquireURI()

	ctx.URI().CopyTo(u)
	u.Update(uri)

	ctx.Response.Header.SetBytesKV(headerLocation, u.FullURI())

	ctx.SetBodyString(fmt.Sprintf("<a href=\"%s\">%d %s</a>", utils.StringHTMLEscape(string(u.FullURI())), statusCode, fasthttp.StatusMessage(statusCode)))

	fasthttp.ReleaseURI(u)
}

// RecordAuthentication records authentication metrics.
func (ctx *AutheliaCtx) RecordAuthentication(success, regulated bool, method string) {
	if ctx.Providers.Metrics == nil {
		return
	}

	ctx.Providers.Metrics.RecordAuthentication(success, regulated, method)
}
