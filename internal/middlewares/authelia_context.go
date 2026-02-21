package middlewares

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewRequestLogger create a new request logger for the given request.
func NewRequestLogger(ctx *fasthttp.RequestCtx) (entry *logrus.Entry) {
	fields := logrus.Fields{
		logging.FieldMethod:   string(ctx.Method()),
		logging.FieldRemoteIP: RequestCtxRemoteIP(ctx).String(),
		logging.FieldPath:     string(ctx.Path()),
	}

	if uri, ok := ctx.UserValue(UserValueKeyRawURI).(string); ok {
		fields[logging.FieldPathRaw] = uri
	}

	return logging.Logger().WithFields(fields)
}

// NewAutheliaCtx instantiate an AutheliaCtx out of a RequestCtx.
func NewAutheliaCtx(requestCTX *fasthttp.RequestCtx, configuration schema.Configuration, providers Providers) (ctx *AutheliaCtx) {
	ctx = new(AutheliaCtx)
	ctx.RequestCtx = requestCTX
	ctx.Providers = providers
	ctx.Configuration = configuration
	ctx.Logger = NewRequestLogger(ctx.RequestCtx)
	ctx.Clock = clock.New()

	return ctx
}

// AvailableSecondFactorMethods returns the available 2FA methods.
func (ctx *AutheliaCtx) AvailableSecondFactorMethods() (methods []string) {
	methods = make([]string, 0, 3)

	if !ctx.Configuration.TOTP.Disable {
		methods = append(methods, model.SecondFactorMethodTOTP)
	}

	if !ctx.Configuration.WebAuthn.Disable {
		methods = append(methods, model.SecondFactorMethodWebAuthn)
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
	if err := ctx.ReplyJSON(ErrorResponse{Status: "KO", Message: message}, 0); err != nil {
		ctx.Logger.Error(err)
	}
}

// ReplyStatusCode resets a response and replies with the given status code and relevant message.
func (ctx *AutheliaCtx) ReplyStatusCode(statusCode int) {
	ctx.Response.Reset()
	ctx.SetStatusCode(statusCode)
	ctx.SetContentTypeTextPlain()
	ctx.SetBodyString(fmt.Sprintf("%d %s", statusCode, fasthttp.StatusMessage(statusCode)))
}

// ReplyJSON writes a JSON response.
func (ctx *AutheliaCtx) ReplyJSON(data any, statusCode int) (err error) {
	var (
		body []byte
	)

	if body, err = json.Marshal(data); err != nil {
		return fmt.Errorf("unable to marshal JSON body: %w", err)
	}

	if statusCode > 0 {
		ctx.SetStatusCode(statusCode)
	}

	ctx.SetContentTypeApplicationJSON()
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

// XForwardedMethod returns the content of the X-Forwarded-Method header.
func (ctx *AutheliaCtx) XForwardedMethod() (method []byte) {
	return ctx.Request.Header.PeekBytes(headerXForwardedMethod)
}

// XForwardedProto returns the content of the X-Forwarded-Proto header.
func (ctx *AutheliaCtx) XForwardedProto() (proto []byte) {
	proto = ctx.Request.Header.PeekBytes(headerXForwardedProto)

	if len(proto) == 0 {
		if ctx.IsTLS() {
			return protoHTTPS
		}

		return protoHTTP
	}

	return proto
}

// XForwardedHost returns the content of the X-Forwarded-Host header.
func (ctx *AutheliaCtx) XForwardedHost() (host []byte) {
	return ctx.Request.Header.PeekBytes(headerXForwardedHost)
}

// GetXForwardedHost returns the content of the X-Forwarded-Host header falling back to the Host header.
func (ctx *AutheliaCtx) GetXForwardedHost() (host []byte) {
	host = ctx.XForwardedHost()

	if host == nil {
		return ctx.Host()
	}

	return host
}

// XForwardedURI returns the content of the X-Forwarded-URI header.
func (ctx *AutheliaCtx) XForwardedURI() (uri []byte) {
	return ctx.Request.Header.PeekBytes(headerXForwardedURI)
}

// GetXForwardedURI returns the content of the X-Forwarded-URI header, falling back to the start-line request path.
func (ctx *AutheliaCtx) GetXForwardedURI() (uri []byte) {
	uri = ctx.XForwardedURI()

	if len(uri) == 0 {
		return ctx.RequestURI()
	}

	return uri
}

// XOriginalMethod returns the content of the X-Original-Method header.
func (ctx *AutheliaCtx) XOriginalMethod() (method []byte) {
	return ctx.Request.Header.PeekBytes(headerXOriginalMethod)
}

// XOriginalURL returns the content of the X-Original-URL header.
func (ctx *AutheliaCtx) XOriginalURL() (uri []byte) {
	return ctx.Request.Header.PeekBytes(headerXOriginalURL)
}

// XAutheliaURL returns the content of the X-Authelia-URL header which is used to communicate the location of the
// portal when using proxies like Envoy.
func (ctx *AutheliaCtx) XAutheliaURL() []byte {
	return ctx.Request.Header.PeekBytes(headerXAutheliaURL)
}

// QueryArgRedirect returns the content of the 'rd' query argument.
func (ctx *AutheliaCtx) QueryArgRedirect() []byte {
	return ctx.QueryArgs().PeekBytes(qryArgRedirect)
}

// QueryArgAutheliaURL returns the content of the 'authelia_url' query argument.
func (ctx *AutheliaCtx) QueryArgAutheliaURL() []byte {
	return ctx.QueryArgs().PeekBytes(qryArgAutheliaURL)
}

// AuthzPath returns the 'authz_path' value.
func (ctx *AutheliaCtx) AuthzPath() (uri []byte) {
	if uv := ctx.UserValue(UserValueRouterKeyExtAuthzPath); uv != nil {
		return []byte(uv.(string))
	}

	return nil
}

// BasePath returns the base_url as per the path visited by the client.
func (ctx *AutheliaCtx) BasePath() string {
	if baseURL := ctx.UserValue(UserValueKeyBaseURL); baseURL != nil {
		return baseURL.(string)
	}

	return ""
}

// BasePathSlash is the same as BasePath but returns a final slash as well.
func (ctx *AutheliaCtx) BasePathSlash() string {
	if baseURL := ctx.UserValue(UserValueKeyBaseURL); baseURL != nil {
		if value := baseURL.(string); value[len(value)-1] == '/' {
			return value
		} else {
			return value + strSlash
		}
	}

	return strSlash
}

// RootURL returns the Root URL.
func (ctx *AutheliaCtx) RootURL() (issuerURL *url.URL) {
	return &url.URL{
		Scheme: string(ctx.XForwardedProto()),
		Host:   string(ctx.GetXForwardedHost()),
		Path:   ctx.BasePath(),
	}
}

// RootURLSlash is the same as RootURL but includes a final slash as well.
func (ctx *AutheliaCtx) RootURLSlash() (issuerURL *url.URL) {
	return &url.URL{
		Scheme: string(ctx.XForwardedProto()),
		Host:   string(ctx.GetXForwardedHost()),
		Path:   ctx.BasePathSlash(),
	}
}

// GetCookieDomainFromTargetURI returns the session provider for the targetURI domain.
func (ctx *AutheliaCtx) GetCookieDomainFromTargetURI(targetURI *url.URL) string {
	if targetURI == nil {
		return ""
	}

	hostname := targetURI.Hostname()

	for _, domain := range ctx.Configuration.Session.Cookies {
		if utils.HasDomainSuffix(hostname, domain.Domain) {
			return domain.Domain
		}
	}

	return ""
}

// IsSafeRedirectionTargetURI returns true if the targetURI is within the scope of a cookie domain and secure.
func (ctx *AutheliaCtx) IsSafeRedirectionTargetURI(targetURI *url.URL) bool {
	if targetURI == nil {
		return false
	}

	if !utils.IsURISecure(targetURI) {
		return false
	}

	return ctx.GetCookieDomainFromTargetURI(targetURI) != ""
}

// GetCookieDomain returns the cookie domain for the current request.
func (ctx *AutheliaCtx) GetCookieDomain() (domain string, err error) {
	var targetURI *url.URL

	if targetURI, err = ctx.GetXOriginalURLOrXForwardedURL(); err != nil {
		return "", fmt.Errorf("unable to retrieve cookie domain: %s", err)
	}

	return ctx.GetCookieDomainFromTargetURI(targetURI), nil
}

// GetSessionProviderByTargetURI returns the session provider for the Request's domain.
func (ctx *AutheliaCtx) GetSessionProviderByTargetURI(targetURL *url.URL) (provider *session.Session, err error) {
	domain := ctx.GetCookieDomainFromTargetURI(targetURL)

	if domain == "" {
		return nil, fmt.Errorf("unable to retrieve session cookie domain provider: no configured session cookie domain matches the url '%s'", targetURL)
	}

	return ctx.Providers.SessionProvider.Get(domain)
}

// GetSessionProvider returns the session provider for the Request's domain.
func (ctx *AutheliaCtx) GetSessionProvider() (provider *session.Session, err error) {
	if ctx.session == nil {
		var targetURI *url.URL

		if targetURI, err = ctx.GetXOriginalURLOrXForwardedURL(); err != nil {
			return nil, fmt.Errorf("unable to retrieve session cookie domain: %w", err)
		}

		if ctx.session, err = ctx.GetSessionProviderByTargetURI(targetURI); err != nil {
			return nil, err
		}
	}

	return ctx.session, nil
}

func (ctx *AutheliaCtx) NewSession() (userSession session.UserSession) {
	if provider, err := ctx.GetSessionProvider(); err != nil {
		return session.NewDefaultUserSession()
	} else {
		return provider.NewDefaultUserSession()
	}
}

func (ctx *AutheliaCtx) GetSessionConfig() (config schema.SessionCookie) {
	if provider, err := ctx.GetSessionProvider(); err != nil {
		return config
	} else {
		return provider.Config
	}
}

// GetCookieDomainSessionProvider returns the session provider for the provided domain.
func (ctx *AutheliaCtx) GetCookieDomainSessionProvider(domain string) (provider *session.Session, err error) {
	if domain == "" {
		return nil, fmt.Errorf("unable to retrieve session cookie domain provider: no configured session cookie domain matches the domain '%s'", domain)
	}

	if ctx.Providers.SessionProvider == nil {
		return nil, fmt.Errorf("unable to retrieve session cookie domain provider: no session provider is configured")
	}

	return ctx.Providers.SessionProvider.Get(domain)
}

// GetSession returns the user session provided the cookie provider could be discovered. It is recommended to get the
// provider itself if you also need to update or destroy sessions.
func (ctx *AutheliaCtx) GetSession() (userSession session.UserSession, err error) {
	var provider *session.Session

	if provider, err = ctx.GetSessionProvider(); err != nil {
		return userSession, err
	}

	if userSession, err = provider.GetSession(ctx.RequestCtx); err != nil {
		ctx.Logger.Error("Unable to retrieve user session")
		return provider.NewDefaultUserSession(), nil
	}

	if userSession.CookieDomain != provider.Config.Domain {
		ctx.Logger.Warnf("Destroying session cookie as the cookie domain '%s' does not match the requests detected cookie domain '%s' which may be a sign a user tried to move this cookie from one domain to another", userSession.CookieDomain, provider.Config.Domain)

		if err = provider.DestroySession(ctx.RequestCtx); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred trying to destroy the session cookie")
		}

		userSession = provider.NewDefaultUserSession()

		if err = provider.SaveSession(ctx.RequestCtx, userSession); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred trying to save the new session cookie")
		}
	}

	return userSession, nil
}

// SaveSession saves the content of the session.
func (ctx *AutheliaCtx) SaveSession(userSession session.UserSession) error {
	provider, err := ctx.GetSessionProvider()
	if err != nil {
		return fmt.Errorf("unable to save user session: %s", err)
	}

	return provider.SaveSession(ctx.RequestCtx, userSession)
}

// RegenerateSession regenerates a user session.
func (ctx *AutheliaCtx) RegenerateSession() (err error) {
	provider, err := ctx.GetSessionProvider()
	if err != nil {
		return fmt.Errorf("unable to regenerate user session: %s", err)
	}

	return provider.RegenerateSession(ctx.RequestCtx)
}

// DestroySession destroys a user session.
func (ctx *AutheliaCtx) DestroySession() (err error) {
	provider, err := ctx.GetSessionProvider()
	if err != nil {
		return fmt.Errorf("unable to destroy user session: %s", err)
	}

	return provider.DestroySession(ctx.RequestCtx)
}

// GetDefaultRedirectionURL retrieves the default redirection URL for the request.
func (ctx *AutheliaCtx) GetDefaultRedirectionURL() *url.URL {
	if provider, err := ctx.GetSessionProvider(); err == nil {
		return provider.Config.DefaultRedirectionURL
	}

	return nil
}

// ReplyOK is a helper method to reply ok.
func (ctx *AutheliaCtx) ReplyOK() {
	ctx.SetContentTypeApplicationJSON()
	ctx.SetBody(okMessageBytes)
}

// ParseBody parse the request body into the type of value.
func (ctx *AutheliaCtx) ParseBody(value any) (err error) {
	if err = json.Unmarshal(ctx.PostBody(), &value); err != nil {
		return fmt.Errorf("unable to parse body: %w", err)
	}

	if _, err = govalidator.ValidateStruct(value); err != nil {
		return fmt.Errorf("unable to validate body: %w", err)
	}

	return nil
}

// SetContentTypeApplicationJSON sets the Content-Type header to 'application/json; charset=utf-8'.
func (ctx *AutheliaCtx) SetContentTypeApplicationJSON() {
	ctx.SetContentTypeBytes(contentTypeApplicationJSON)
}

// SetContentTypeTextPlain efficiently sets the Content-Type header to 'text/plain; charset=utf-8'.
func (ctx *AutheliaCtx) SetContentTypeTextPlain() {
	ctx.SetContentTypeBytes(contentTypeTextPlain)
}

// SetContentTypeTextHTML efficiently sets the Content-Type header to 'text/html; charset=utf-8'.
func (ctx *AutheliaCtx) SetContentTypeTextHTML() {
	ctx.SetContentTypeBytes(contentTypeTextHTML)
}

// SetContentTypeApplicationYAML efficiently sets the Content-Type header to 'application/yaml; charset=utf-8'.
func (ctx *AutheliaCtx) SetContentTypeApplicationYAML() {
	ctx.SetContentTypeBytes(contentTypeApplicationYAML)
}

// SetJSONBody Set json body.
func (ctx *AutheliaCtx) SetJSONBody(value any) error {
	return ctx.ReplyJSON(OKResponse{Status: "OK", Data: value}, 0)
}

func (ctx *AutheliaCtx) GetRequestQueryArgValue(key []byte) (value []byte) {
	return ctx.QueryArgs().PeekBytes(key)
}

func (ctx *AutheliaCtx) GetRequestHeaderValue(key []byte) (value []byte) {
	return ctx.Request.Header.PeekBytes(key)
}

func (ctx *AutheliaCtx) SetResponseHeaderValue(key []byte, value string) {
	ctx.Response.Header.SetBytesK(key, value)
}

func (ctx *AutheliaCtx) SetResponseHeaderValueBytes(key, value []byte) {
	ctx.Response.Header.SetBytesKV(key, value)
}

// RemoteIP return the remote IP taking X-Forwarded-For header into account if provided.
func (ctx *AutheliaCtx) RemoteIP() net.IP {
	return RequestCtxRemoteIP(ctx.RequestCtx)
}

// GetXForwardedURL returns the parsed X-Forwarded-Proto, X-Forwarded-Host, and X-Forwarded-URI request header as a
// *url.URL.
func (ctx *AutheliaCtx) GetXForwardedURL() (requestURI *url.URL, err error) {
	forwardedProto, forwardedHost, forwardedURI := ctx.XForwardedProto(), ctx.GetXForwardedHost(), ctx.GetXForwardedURI()

	if len(forwardedHost) == 0 {
		return nil, ErrMissingXForwardedHost
	}

	value := utils.BytesJoin(forwardedProto, protoHostSeparator, forwardedHost, forwardedURI)

	if requestURI, err = url.ParseRequestURI(string(value)); err != nil {
		return nil, fmt.Errorf("failed to parse X-Forwarded Headers: %w", err)
	}

	return requestURI, nil
}

// GetXOriginalURL returns the parsed X-OriginalURL request header as a *url.URL.
func (ctx *AutheliaCtx) GetXOriginalURL() (requestURI *url.URL, err error) {
	value := ctx.XOriginalURL()

	if value == nil {
		return nil, ErrMissingXOriginalURL
	}

	if requestURI, err = url.ParseRequestURI(string(value)); err != nil {
		return nil, fmt.Errorf("failed to parse X-Original-URL header: %w", err)
	}

	return requestURI, nil
}

// GetXOriginalURLOrXForwardedURL returns the parsed X-Original-URL request header if it's available or the parsed
// X-Forwarded request headers if not.
func (ctx *AutheliaCtx) GetXOriginalURLOrXForwardedURL() (requestURI *url.URL, err error) {
	requestURI, err = ctx.GetXOriginalURL()

	switch {
	case err == nil:
		return requestURI, nil
	case errors.Is(err, ErrMissingXOriginalURL):
		return ctx.GetXForwardedURL()
	default:
		return requestURI, err
	}
}

// GetOrigin returns the expected origin for requests from this endpoint.
func (ctx *AutheliaCtx) GetOrigin() (origin *url.URL, err error) {
	if origin, err = ctx.GetXOriginalURLOrXForwardedURL(); err != nil {
		return nil, err
	}

	origin.Path = ""
	origin.RawPath = ""

	return origin, nil
}

// IssuerURL returns the expected Issuer.
func (ctx *AutheliaCtx) IssuerURL() (issuerURL *url.URL, err error) {
	issuerURL = &url.URL{
		Scheme: string(ctx.XForwardedProto()),
		Host:   string(ctx.GetXForwardedHost()),
		Path:   ctx.BasePath(),
	}

	if len(issuerURL.Host) == 0 {
		return nil, ErrMissingXForwardedHost
	}

	return issuerURL, nil
}

// IsXHR returns true if the request is a XMLHttpRequest.
func (ctx *AutheliaCtx) IsXHR() (xhr bool) {
	if requestedWith := ctx.Request.Header.PeekBytes(headerXRequestedWith); requestedWith != nil && strings.EqualFold(string(requestedWith), headerValueXRequestedWithXHR) {
		return true
	}

	return false
}

// AcceptsMIME takes a mime type and returns true if the request accepts that type or the wildcard type.
func (ctx *AutheliaCtx) AcceptsMIME(mime string) (acceptsMime bool) {
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
// content in the form of a link to the location if the request method was not head.
func (ctx *AutheliaCtx) SpecialRedirect(uri string, statusCode int) {
	var u []byte

	u, statusCode = ctx.setSpecialRedirect(uri, statusCode)

	ctx.SetContentTypeTextHTML()
	ctx.SetBodyString(fmt.Sprintf("<a href=\"%s\">%d %s</a>", utils.StringHTMLEscape(string(u)), statusCode, fasthttp.StatusMessage(statusCode)))
}

// SpecialRedirectNoBody performs a redirect similar to fasthttp.RequestCtx except it allows statusCode 401 and includes
// no body.
func (ctx *AutheliaCtx) SpecialRedirectNoBody(uri string, statusCode int) {
	_, _ = ctx.setSpecialRedirect(uri, statusCode)
}

func (ctx *AutheliaCtx) setSpecialRedirect(uri string, statusCode int) ([]byte, int) {
	if statusCode < fasthttp.StatusMovedPermanently || (statusCode > fasthttp.StatusSeeOther && statusCode != fasthttp.StatusTemporaryRedirect && statusCode != fasthttp.StatusPermanentRedirect && statusCode != fasthttp.StatusUnauthorized) {
		statusCode = fasthttp.StatusFound
	}

	ctx.SetStatusCode(statusCode)

	u := fasthttp.AcquireURI()

	ctx.URI().CopyTo(u)
	u.Update(uri)

	raw := u.FullURI()

	ctx.Response.Header.SetBytesKV(headerLocation, raw)

	fasthttp.ReleaseURI(u)

	return raw, statusCode
}

// RecordAuthn records authentication metrics.
func (ctx *AutheliaCtx) RecordAuthn(success, regulated bool, method string) {
	if ctx.Providers.Metrics == nil {
		return
	}

	ctx.Providers.Metrics.RecordAuthn(success, regulated, method)
}

// GetClock returns the clock. For use with interface fulfillment.
func (ctx *AutheliaCtx) GetClock() (clock clock.Provider) {
	return ctx.Clock
}

// GetRandom returns the random provider. For use with interface fulfillment.
func (ctx *AutheliaCtx) GetRandom() (random random.Provider) {
	return ctx.Providers.Random
}

// GetJWTWithTimeFuncOption returns the WithTimeFunc jwt.ParserOption. For use with interface fulfillment.
func (ctx *AutheliaCtx) GetJWTWithTimeFuncOption() (option jwt.ParserOption) {
	return jwt.WithTimeFunc(ctx.Clock.Now)
}

// GetConfiguration returns the current configuration.
func (ctx *AutheliaCtx) GetConfiguration() (config schema.Configuration) {
	return ctx.Configuration
}

// GetProviders returns the providers for this context.
func (ctx *AutheliaCtx) GetProviders() (providers Providers) {
	return ctx.Providers
}

func (ctx *AutheliaCtx) GetUserProvider() (provider authentication.UserProvider) {
	return ctx.Providers.UserProvider
}

func (ctx *AutheliaCtx) GetProviderUserAttributeResolver() expression.UserAttributeResolver {
	return ctx.Providers.UserAttributeResolver
}

func (ctx *AutheliaCtx) GetWebAuthnProvider() (w *webauthn.WebAuthn, err error) {
	var (
		origin *url.URL
	)

	if origin, err = ctx.GetOrigin(); err != nil {
		return nil, err
	}

	config := &webauthn.Config{
		RPID:                  origin.Hostname(),
		RPDisplayName:         ctx.Configuration.WebAuthn.DisplayName,
		RPOrigins:             []string{origin.String()},
		AttestationPreference: ctx.Configuration.WebAuthn.ConveyancePreference,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			AuthenticatorAttachment: ctx.Configuration.WebAuthn.SelectionCriteria.Attachment,
			ResidentKey:             ctx.Configuration.WebAuthn.SelectionCriteria.Discoverability,
			UserVerification:        ctx.Configuration.WebAuthn.SelectionCriteria.UserVerification,
		},
		Debug:                false,
		EncodeUserIDAsString: false,
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    ctx.Configuration.WebAuthn.Timeout,
				TimeoutUVD: ctx.Configuration.WebAuthn.Timeout,
			},
			Registration: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    ctx.Configuration.WebAuthn.Timeout,
				TimeoutUVD: ctx.Configuration.WebAuthn.Timeout,
			},
		},
		MDS: ctx.Providers.MetaDataService,
	}

	switch ctx.Configuration.WebAuthn.SelectionCriteria.Attachment {
	case protocol.Platform, protocol.CrossPlatform:
		config.AuthenticatorSelection.AuthenticatorAttachment = ctx.Configuration.WebAuthn.SelectionCriteria.Attachment
	}

	switch ctx.Configuration.WebAuthn.SelectionCriteria.Discoverability {
	case protocol.ResidentKeyRequirementRequired:
		config.AuthenticatorSelection.RequireResidentKey = protocol.ResidentKeyRequired()
	case protocol.ResidentKeyRequirementPreferred, protocol.ResidentKeyRequirementDiscouraged:
		config.AuthenticatorSelection.RequireResidentKey = protocol.ResidentKeyNotRequired()
	}

	ctx.Logger.Tracef("Creating new WebAuthn RP instance with ID %s and Origins %s", config.RPID, strings.Join(config.RPOrigins, ", "))

	return webauthn.New(config)
}

// Value is a shaded method of context.Context which returns the AutheliaCtx struct if the key is the internal key
// otherwise it returns the shaded value.
func (ctx *AutheliaCtx) Value(key any) any {
	if key == model.CtxKeyAutheliaCtx {
		return ctx
	}

	return ctx.RequestCtx.Value(key)
}

// GetLogger returns the logger for this request.
func (ctx *AutheliaCtx) GetLogger() *logrus.Entry {
	return ctx.Logger
}
