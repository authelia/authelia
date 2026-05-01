// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

// ServeTemplatedFile serves a templated version of a specified file,
// this is utilized to pass information between the backend and frontend
// and generate a nonce to support a restrictive CSP for the styles the
// frontend injects at runtime.
func ServeTemplatedFile(t templates.Template, opts *TemplatedFileOptions) middlewares.RequestHandler {
	ext := path.Ext(t.Name())

	return func(ctx *middlewares.AutheliaCtx) {
		var err error

		lang := "en"

		if c := ctx.Request.Header.Cookie("language"); c != nil {
			if reValidLanguageCodes.MatchString(string(c)) {
				lang = string(c)
			}
		}

		logoOverride := strFalse

		if opts.AssetPath != "" {
			if _, err = os.Stat(filepath.Join(opts.AssetPath, fileLogo)); err == nil {
				logoOverride = strTrue
			}
		}

		middlewares.SetBaseSecurityHeaders(ctx.RequestCtx)

		switch ext {
		case utils.ExtHTML:
			ctx.SetContentTypeTextHTML()
		case utils.ExtJSON:
			ctx.SetContentTypeApplicationJSON()
		default:
			ctx.SetContentTypeTextPlain()
		}

		nonce := ctx.Providers.Random.StringCustom(32, random.CharSetAlphaNumeric)
		oidcClientLogoURIs := resolveOIDCConsentLogoURI(ctx)

		switch {
		case ctx.Configuration.Server.Headers.CSPTemplate != "":
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, expandCSPTemplate(string(ctx.Configuration.Server.Headers.CSPTemplate), nonce, oidcClientLogoURIs))
		case utils.Dev:
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, expandCSPTemplate(tmplCSPDevelopment, nonce, oidcClientLogoURIs))
		default:
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, expandCSPTemplate(tmplCSPDefault, nonce, oidcClientLogoURIs))
		}

		var (
			rememberMe     string
			csrfCookieName string
			baseURL        string
			domain         string
			provider       session.Strategy
			origin         *url.URL
		)

		baseURL = ctx.TemplateRootURL().String()

		if provider, err = ctx.GetSessionProvider(); err == nil {
			domain = provider.GetConfig().Domain
			rememberMe = strconv.FormatBool(!provider.GetConfig().DisableRememberMe)
			csrfCookieName = session.CSRFCookieName(provider.GetConfig().Name)
		}

		origin, err = ctx.GetOrigin()

		passkeyLogin := strconv.FormatBool(err == nil && ctx.Configuration.WebAuthn.EnablePasskeyLogin && webauthn.IsOriginAvailable(ctx.Configuration.WebAuthn, origin))

		data := &bytes.Buffer{}

		if err = t.Execute(data, opts.CommonData(ctx.BasePath(), baseURL, domain, nonce, lang, logoOverride, rememberMe, passkeyLogin, csrfCookieName)); err != nil {
			ctx.Error(errMessageServerGeneric, fasthttp.StatusServiceUnavailable)
			ctx.Logger.WithError(err).Errorf("Error occurred rendering template")

			return
		}

		switch {
		case ctx.IsHead():
			ctx.Response.ResetBody()
			ctx.Response.SkipBody = true
			ctx.Response.Header.Set(fasthttp.HeaderContentLength, strconv.Itoa(data.Len()))
		default:
			if _, err = data.WriteTo(ctx.Response.BodyWriter()); err != nil {
				ctx.GetLogger().WithError(err).Error("Error occurred writing response body")
				ctx.Error(errMessageServerGeneric, fasthttp.StatusServiceUnavailable)

				return
			}
		}
	}
}

// ServeTemplatedOpenAPI serves templated OpenAPI related files.
func ServeTemplatedOpenAPI(t templates.Template, opts *TemplatedFileOptions) middlewares.RequestHandler {
	ext := path.Ext(t.Name())

	return func(ctx *middlewares.AutheliaCtx) {
		var nonce string

		switch ext {
		case utils.ExtHTML:
			nonce = ctx.Providers.Random.StringCustom(32, random.CharSetAlphaNumeric)
			ctx.Response.Header.Del(fasthttp.HeaderContentSecurityPolicy)
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, fmt.Sprintf(tmplCSPSwagger, nonce))
			ctx.SetContentTypeTextHTML()
		case utils.ExtYAML, utils.ExtYML:
			ctx.SetContentTypeApplicationYAML()
		default:
			ctx.SetContentTypeTextPlain()
		}

		var (
			baseURL  string
			domain   string
			cookie   string
			provider session.Strategy
			err      error
		)

		baseURL = ctx.TemplateRootURL().String()

		if provider, err = ctx.GetSessionProvider(); err == nil {
			config := provider.GetConfig()

			domain, cookie = config.Domain, config.Name
		}

		data := &bytes.Buffer{}
		if err = t.Execute(data, opts.OpenAPIData(ctx.BasePath(), baseURL, domain, cookie, nonce)); err != nil {
			ctx.Error(errMessageServerGeneric, fasthttp.StatusServiceUnavailable)
			ctx.Logger.WithError(err).Errorf("Error occurred rendering template")

			return
		}

		switch {
		case ctx.IsHead():
			ctx.Response.ResetBody()
			ctx.Response.SkipBody = true
			ctx.Response.Header.Set(fasthttp.HeaderContentLength, strconv.Itoa(data.Len()))
		default:
			if _, err = data.WriteTo(ctx.Response.BodyWriter()); err != nil {
				ctx.Error(errMessageServerGeneric, fasthttp.StatusServiceUnavailable)
				ctx.Logger.WithError(err).Errorf("Error occurred writing body")

				return
			}
		}
	}
}

// ETagRootURL dynamically matches the If-None-Match header and adds the ETag header.
func ETagRootURL(next middlewares.RequestHandler) middlewares.RequestHandler {
	etags := map[string][]byte{}

	mu := &sync.Mutex{}

	return func(ctx *middlewares.AutheliaCtx) {
		k := ctx.TemplateRootURL().String()

		mu.Lock()

		etag, ok := etags[k]

		mu.Unlock()

		if ok && bytes.Equal(etag, ctx.Request.Header.PeekBytes(headerIfNoneMatch)) {
			ctx.Response.Header.SetBytesKV(headerETag, etag)
			ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueCacheControlETaggedAssets)

			ctx.SetStatusCode(fasthttp.StatusNotModified)

			return
		}

		next(ctx)

		if ctx.Response.SkipBody || ctx.Response.StatusCode() != fasthttp.StatusOK {
			// Skip generating the ETag as the response body should be empty.
			return
		}

		etagNew := generateEtag(ctx.Response.Body())

		mu.Lock()

		if !ok || !bytes.Equal(etag, etagNew) {
			etags[k] = etagNew
		}

		mu.Unlock()

		ctx.Response.Header.SetBytesKV(headerETag, etagNew)
		ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueCacheControlETaggedAssets)
	}
}

// NewTemplatedFileOptions returns a new *TemplatedFileOptions.
func NewTemplatedFileOptions(config *schema.Configuration) (opts *TemplatedFileOptions) {
	opts = &TemplatedFileOptions{
		AssetPath:               config.Server.AssetPath,
		CSRFCookieName:          session.CSRFCookieName(config.Session.Name),
		SessionName:             config.Session.Name,
		DuoSelfEnrollment:       strFalse,
		PasskeyLogin:            strconv.FormatBool(config.WebAuthn.EnablePasskeyLogin),
		ExternalIdentityLogin:   strconv.FormatBool(config.AuthenticationBackend.ExternalIdentity != nil && len(config.AuthenticationBackend.ExternalIdentity.Providers) != 0),
		RememberMe:              strconv.FormatBool(!config.Session.DisableRememberMe),
		ResetPassword:           strconv.FormatBool(!config.AuthenticationBackend.PasswordReset.Disable),
		ResetPasswordCustomURL:  config.AuthenticationBackend.PasswordReset.CustomURL.String(),
		TOTPAppAppleStore:       totpAppStoreLink(config.TOTP.Apps.AppleStore),
		TOTPAppGooglePlay:       totpAppStoreLink(config.TOTP.Apps.GooglePlay),
		RegistrationURL:         config.AuthenticationBackend.Registration.CustomURL.String(),
		PasswordChange:          strconv.FormatBool(!config.AuthenticationBackend.PasswordChange.Disable),
		PrivacyPolicyURL:        "",
		PrivacyPolicyAccept:     strFalse,
		Session:                 "",
		Theme:                   config.Theme,
		EndpointsPasswordReset:  !config.AuthenticationBackend.PasswordReset.Disable && config.AuthenticationBackend.PasswordReset.CustomURL.String() == "",
		EndpointsPasswordChange: !config.AuthenticationBackend.PasswordChange.Disable,
		EndpointsWebAuthn:       !config.WebAuthn.Disable,
		EndpointsPasskeys:       !config.WebAuthn.Disable && config.WebAuthn.EnablePasskeyLogin,
		EndpointsTOTP:           !config.TOTP.Disable,
		EndpointsDuo:            !config.DuoAPI.Disable,
		EndpointsRecoveryCodes:  !config.RecoveryCodes.Disable,
		EndpointsOpenIDConnect:  config.IdentityProviders.OIDC != nil,
		EndpointsHealthVerbose:  config.Server.Endpoints.Health.Verbose,
		EndpointsAuthz:          config.Server.Endpoints.Authz,
	}

	if config.PrivacyPolicy.Enabled {
		opts.PrivacyPolicyURL = config.PrivacyPolicy.PolicyURL.String()
		opts.PrivacyPolicyAccept = strconv.FormatBool(config.PrivacyPolicy.RequireUserAcceptance)
	}

	if !config.DuoAPI.Disable {
		opts.DuoSelfEnrollment = strconv.FormatBool(config.DuoAPI.EnableSelfEnrollment)
	}

	return opts
}

// totpAppStoreLink returns the store listing link, or an empty value when the store is disabled.
func totpAppStoreLink(store schema.TOTPAppsStore) string {
	if store.Disable {
		return ""
	}

	return store.URL.String()
}

// TemplatedFileOptions is a struct which is used for many templated files.
type TemplatedFileOptions struct {
	AssetPath              string
	CSRFCookieName         string
	DuoSelfEnrollment      string
	PasskeyLogin           string
	ExternalIdentityLogin  string
	RememberMe             string
	ResetPassword          string
	ResetPasswordCustomURL string
	TOTPAppAppleStore      string
	TOTPAppGooglePlay      string
	RegistrationURL        string
	PasswordChange         string
	PrivacyPolicyURL       string
	PrivacyPolicyAccept    string
	Session                string
	SessionName            string
	Theme                  string

	EndpointsPasswordReset  bool
	EndpointsPasswordChange bool
	EndpointsWebAuthn       bool
	EndpointsPasskeys       bool
	EndpointsTOTP           bool
	EndpointsDuo            bool
	EndpointsRecoveryCodes  bool
	EndpointsOpenIDConnect  bool
	EndpointsHealthVerbose  bool

	EndpointsAuthz map[string]schema.ServerEndpointsAuthz
}

// CommonData returns a TemplatedFileCommonData with the dynamic options. The rememberMe, passkeyLogin, and
// csrfCookieName values override the static options when they're not empty, i.e. when the request resolved to a
// session cookie domain.
func (options *TemplatedFileOptions) CommonData(base, baseURL, domain, nonce, language, logoOverride, rememberMe, passkeyLogin, csrfCookieName string) TemplatedFileCommonData {
	data := TemplatedFileCommonData{
		Base:                   base,
		BaseURL:                baseURL,
		Domain:                 domain,
		CSPNonce:               nonce,
		Language:               language,
		LogoOverride:           logoOverride,
		CSRFCookieName:         options.CSRFCookieName,
		DuoSelfEnrollment:      options.DuoSelfEnrollment,
		PasskeyLogin:           options.PasskeyLogin,
		ExternalIdentityLogin:  options.ExternalIdentityLogin,
		RememberMe:             options.RememberMe,
		ResetPassword:          options.ResetPassword,
		ResetPasswordCustomURL: options.ResetPasswordCustomURL,
		TOTPAppAppleStore:      options.TOTPAppAppleStore,
		TOTPAppGooglePlay:      options.TOTPAppGooglePlay,
		RegistrationURL:        options.RegistrationURL,
		PrivacyPolicyURL:       options.PrivacyPolicyURL,
		PrivacyPolicyAccept:    options.PrivacyPolicyAccept,
		Session:                options.Session,
		Theme:                  options.Theme,
	}

	if rememberMe != "" {
		data.RememberMe = rememberMe
	}

	if passkeyLogin != "" {
		data.PasskeyLogin = passkeyLogin
	}

	if csrfCookieName != "" {
		data.CSRFCookieName = csrfCookieName
	}

	return data
}

// OpenAPIData returns a TemplatedFileOpenAPIData with the dynamic options. The cookie is the name of the session cookie
// for the domain of the request, and the configured session cookie name is used when it's empty, falling back to the
// default session cookie name.
func (options *TemplatedFileOptions) OpenAPIData(base, baseURL, domain, cookie, nonce string) TemplatedFileOpenAPIData {
	if cookie == "" {
		cookie = options.SessionName
	}

	if cookie == "" {
		cookie = schema.DefaultSessionConfiguration.Name
	}

	return TemplatedFileOpenAPIData{
		Base:           base,
		BaseURL:        baseURL,
		Domain:         domain,
		CSPNonce:       nonce,
		Session:        session.SessionCookieName(cookie),
		SessionAuthz:   session.AuthzCookieName(cookie),
		SessionBinding: session.UserAgentBindingCookieName(cookie),
		CSRFCookieName: session.CSRFCookieName(cookie),
		PasswordReset:  options.EndpointsPasswordReset,
		PasswordChange: options.EndpointsPasswordChange,
		WebAuthn:       options.EndpointsWebAuthn,
		Passkeys:       options.EndpointsPasskeys,
		TOTP:           options.EndpointsTOTP,
		Duo:            options.EndpointsDuo,
		RecoveryCodes:  options.EndpointsRecoveryCodes,
		OpenIDConnect:  options.EndpointsOpenIDConnect,
		HealthVerbose:  options.EndpointsHealthVerbose,
		EndpointsAuthz: options.EndpointsAuthz,
	}
}

// TemplatedFileCommonData is a struct which is used for many templated files.
type TemplatedFileCommonData struct {
	Base                   string
	BaseURL                string
	Domain                 string
	CSPNonce               string
	Language               string
	LogoOverride           string
	CSRFCookieName         string
	DuoSelfEnrollment      string
	PasskeyLogin           string
	ExternalIdentityLogin  string
	RememberMe             string
	ResetPassword          string
	ResetPasswordCustomURL string
	TOTPAppAppleStore      string
	TOTPAppGooglePlay      string
	RegistrationURL        string
	PrivacyPolicyURL       string
	PrivacyPolicyAccept    string
	Session                string
	Theme                  string
}

// TemplatedFileOpenAPIData is a struct which is used for the OpenAPI spec file.
type TemplatedFileOpenAPIData struct {
	Base           string
	BaseURL        string
	Domain         string
	CSPNonce       string
	Session        string
	SessionAuthz   string
	SessionBinding string
	CSRFCookieName string
	PasswordReset  bool
	PasswordChange bool
	WebAuthn       bool
	Passkeys       bool
	TOTP           bool
	Duo            bool
	RecoveryCodes  bool
	OpenIDConnect  bool
	HealthVerbose  bool

	EndpointsAuthz map[string]schema.ServerEndpointsAuthz
}
