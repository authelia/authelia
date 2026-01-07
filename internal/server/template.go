package server

import (
	"bytes"
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ServeTemplatedFile serves a templated version of a specified file,
// this is utilised to pass information between the backend and frontend
// and generate a nonce to support a restrictive CSP while using material-ui.
func ServeTemplatedFile(t templates.Template, opts *TemplatedFileOptions) middlewares.RequestHandler {
	ext := path.Ext(t.Name())

	return func(ctx *middlewares.AutheliaCtx) {
		var err error

		lang := "en"
		if c := ctx.Request.Header.Cookie("language"); c != nil {
			lang = string(c)
		}

		logoOverride := strFalse

		if opts.AssetPath != "" {
			if _, err = os.Stat(filepath.Join(opts.AssetPath, fileLogo)); err == nil {
				logoOverride = strTrue
			}
		}

		middlewares.SetBaseSecurityHeaders(ctx.RequestCtx)

		switch ext {
		case extHTML:
			ctx.SetContentTypeTextHTML()
		case extJSON:
			ctx.SetContentTypeApplicationJSON()
		default:
			ctx.SetContentTypeTextPlain()
		}

		nonce := ctx.Providers.Random.StringCustom(32, random.CharSetAlphaNumeric)

		switch {
		case ctx.Configuration.Server.Headers.CSPTemplate != "":
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, strings.ReplaceAll(string(ctx.Configuration.Server.Headers.CSPTemplate), placeholderCSPNonce, nonce))
		case utils.Dev:
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, fmt.Sprintf(tmplCSPDevelopment, nonce))
		default:
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, fmt.Sprintf(tmplCSPDefault, nonce))
		}

		var (
			rememberMe string
			baseURL    string
			domain     string
			provider   *session.Session
		)

		baseURL = ctx.RootURLSlash().String()

		if provider, err = ctx.GetSessionProvider(); err == nil {
			domain = provider.Config.Domain
			rememberMe = strconv.FormatBool(!provider.Config.DisableRememberMe)
		}

		data := &bytes.Buffer{}

		if err = t.Execute(data, opts.CommonData(ctx.BasePath(), baseURL, domain, nonce, lang, logoOverride, rememberMe)); err != nil {
			ctx.RequestCtx.Error("an error occurred", fasthttp.StatusServiceUnavailable)
			ctx.Logger.WithError(err).Errorf("Error occcurred rendering template")

			return
		}

		switch {
		case ctx.IsHead():
			ctx.Response.ResetBody()
			ctx.Response.SkipBody = true
			ctx.Response.Header.Set(fasthttp.HeaderContentLength, strconv.Itoa(data.Len()))
		default:
			if _, err = data.WriteTo(ctx.Response.BodyWriter()); err != nil {
				ctx.RequestCtx.Error("an error occurred", fasthttp.StatusServiceUnavailable)
				ctx.Logger.WithError(err).Errorf("Error occcurred writing body")

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
		case extHTML:
			nonce = ctx.Providers.Random.StringCustom(32, random.CharSetAlphaNumeric)
			ctx.Response.Header.Del(fasthttp.HeaderContentSecurityPolicy)
			ctx.Response.Header.Add(fasthttp.HeaderContentSecurityPolicy, fmt.Sprintf(tmplCSPSwagger, nonce))
			ctx.SetContentTypeTextHTML()
		case extYML:
			ctx.SetContentTypeApplicationYAML()
		default:
			ctx.SetContentTypeTextPlain()
		}

		var (
			baseURL  string
			domain   string
			provider *session.Session
			err      error
		)

		baseURL = ctx.RootURLSlash().String()

		if provider, err = ctx.GetSessionProvider(); err == nil {
			domain = provider.Config.Domain
		}

		data := &bytes.Buffer{}
		if err = t.Execute(data, opts.OpenAPIData(ctx.BasePath(), baseURL, domain, nonce)); err != nil {
			ctx.RequestCtx.Error("an error occurred", fasthttp.StatusServiceUnavailable)
			ctx.Logger.WithError(err).Errorf("Error occcurred rendering template")

			return
		}

		switch {
		case ctx.IsHead():
			ctx.Response.ResetBody()
			ctx.Response.SkipBody = true
			ctx.Response.Header.Set(fasthttp.HeaderContentLength, strconv.Itoa(data.Len()))
		default:
			if _, err = data.WriteTo(ctx.Response.BodyWriter()); err != nil {
				ctx.RequestCtx.Error("an error occurred", fasthttp.StatusServiceUnavailable)
				ctx.Logger.WithError(err).Errorf("Error occcurred writing body")

				return
			}
		}
	}
}

// ETagRootURL dynamically matches the If-None-Match header and adds the ETag header.
func ETagRootURL(next middlewares.RequestHandler) middlewares.RequestHandler {
	etags := map[string][]byte{}

	h := sha1.New() //nolint:gosec // Usage is for collision avoidance not security.
	mu := &sync.Mutex{}

	return func(ctx *middlewares.AutheliaCtx) {
		k := ctx.RootURLSlash().String()

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

		mu.Lock()

		h.Write(ctx.Response.Body())
		sum := h.Sum(nil)
		h.Reset()

		etagNew := make([]byte, hex.EncodedLen(len(sum)))

		hex.Encode(etagNew, sum)

		if !ok || !bytes.Equal(etag, etagNew) {
			etags[k] = etagNew
		}

		mu.Unlock()

		ctx.Response.Header.SetBytesKV(headerETag, etagNew)
		ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueCacheControlETaggedAssets)
	}
}

func writeHealthCheckEnv(disabled bool, scheme, host, path string, port uint16) (err error) {
	if disabled {
		return nil
	}

	_, err = os.Stat("/app/healthcheck.sh")
	if err != nil {
		return nil
	}

	_, err = os.Stat("/app/.healthcheck.env")
	if err != nil {
		return nil
	}

	file, err := os.OpenFile("/app/.healthcheck.env", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	if host == "0.0.0.0" {
		host = localhost
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}

	if path == "/" {
		path = ""
	}

	_, err = fmt.Fprintf(file, healthCheckEnv, scheme, host, port, path)

	return err
}

// NewTemplatedFileOptions returns a new *TemplatedFileOptions.
func NewTemplatedFileOptions(config *schema.Configuration) (opts *TemplatedFileOptions) {
	opts = &TemplatedFileOptions{
		AssetPath:               config.Server.AssetPath,
		DuoSelfEnrollment:       strFalse,
		PasskeyLogin:            strconv.FormatBool(config.WebAuthn.EnablePasskeyLogin),
		SPNEGOLogin:             strconv.FormatBool(config.SPNEGO.Enabled),
		RememberMe:              strconv.FormatBool(!config.Session.DisableRememberMe),
		ResetPassword:           strconv.FormatBool(!config.AuthenticationBackend.PasswordReset.Disable),
		ResetPasswordCustomURL:  config.AuthenticationBackend.PasswordReset.CustomURL.String(),
		PasswordChange:          strconv.FormatBool(!config.AuthenticationBackend.PasswordChange.Disable),
		PrivacyPolicyURL:        "",
		PrivacyPolicyAccept:     strFalse,
		Session:                 "",
		Theme:                   config.Theme,
		EndpointsPasswordReset:  !config.AuthenticationBackend.PasswordReset.Disable && config.AuthenticationBackend.PasswordReset.CustomURL.String() == "",
		EndpointsPasswordChange: !config.AuthenticationBackend.PasswordChange.Disable,
		EndpointsWebAuthn:       !config.WebAuthn.Disable,
		EndpointsPasskeys:       !config.WebAuthn.Disable && config.WebAuthn.EnablePasskeyLogin,
		EndpointsSPNEGO:         config.SPNEGO.Enabled,
		EndpointsTOTP:           !config.TOTP.Disable,
		EndpointsDuo:            !config.DuoAPI.Disable,
		EndpointsOpenIDConnect:  config.IdentityProviders.OIDC != nil,
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

// TemplatedFileOptions is a struct which is used for many templated files.
type TemplatedFileOptions struct {
	AssetPath              string
	DuoSelfEnrollment      string
	PasskeyLogin           string
	SPNEGOLogin            string
	RememberMe             string
	ResetPassword          string
	ResetPasswordCustomURL string
	PasswordChange         string
	PrivacyPolicyURL       string
	PrivacyPolicyAccept    string
	Session                string
	Theme                  string

	EndpointsPasswordReset  bool
	EndpointsPasswordChange bool
	EndpointsWebAuthn       bool
	EndpointsPasskeys       bool
	EndpointsTOTP           bool
	EndpointsDuo            bool
	EndpointsOpenIDConnect  bool
	EndpointsSPNEGO         bool

	EndpointsAuthz map[string]schema.ServerEndpointsAuthz
}

// CommonData returns a TemplatedFileCommonData with the dynamic options.
func (options *TemplatedFileOptions) CommonData(base, baseURL, domain, nonce, language, logoOverride, rememberMe string) TemplatedFileCommonData {
	if rememberMe != "" {
		return options.commonDataWithRememberMe(base, baseURL, domain, nonce, language, logoOverride, rememberMe)
	}

	return TemplatedFileCommonData{
		Base:     base,
		BaseURL:  baseURL,
		Domain:   domain,
		CSPNonce: nonce,
		Language: language,

		LogoOverride:           logoOverride,
		DuoSelfEnrollment:      options.DuoSelfEnrollment,
		PasskeyLogin:           options.PasskeyLogin,
		SPNEGO:                 options.SPNEGOLogin,
		RememberMe:             options.RememberMe,
		ResetPassword:          options.ResetPassword,
		ResetPasswordCustomURL: options.ResetPasswordCustomURL,
		PrivacyPolicyURL:       options.PrivacyPolicyURL,
		PrivacyPolicyAccept:    options.PrivacyPolicyAccept,
		Session:                options.Session,
		Theme:                  options.Theme,
	}
}

// CommonDataWithRememberMe returns a TemplatedFileCommonData with the dynamic options.
func (options *TemplatedFileOptions) commonDataWithRememberMe(base, baseURL, domain, nonce, language, logoOverride, rememberMe string) TemplatedFileCommonData {
	return TemplatedFileCommonData{
		Base:                   base,
		BaseURL:                baseURL,
		Domain:                 domain,
		CSPNonce:               nonce,
		Language:               language,
		LogoOverride:           logoOverride,
		DuoSelfEnrollment:      options.DuoSelfEnrollment,
		PasskeyLogin:           options.PasskeyLogin,
		SPNEGO:                 options.SPNEGOLogin,
		RememberMe:             rememberMe,
		ResetPassword:          options.ResetPassword,
		ResetPasswordCustomURL: options.ResetPasswordCustomURL,
		PrivacyPolicyURL:       options.PrivacyPolicyURL,
		PrivacyPolicyAccept:    options.PrivacyPolicyAccept,
		Session:                options.Session,
		Theme:                  options.Theme,
	}
}

// OpenAPIData returns a TemplatedFileOpenAPIData with the dynamic options.
func (options *TemplatedFileOptions) OpenAPIData(base, baseURL, domain, nonce string) TemplatedFileOpenAPIData {
	return TemplatedFileOpenAPIData{
		Base:           base,
		BaseURL:        baseURL,
		Domain:         domain,
		CSPNonce:       nonce,
		Session:        options.Session,
		PasswordReset:  options.EndpointsPasswordReset,
		PasswordChange: options.EndpointsPasswordChange,
		WebAuthn:       options.EndpointsWebAuthn,
		Passkeys:       options.EndpointsPasskeys,
		SPNEGO:         options.EndpointsSPNEGO,
		TOTP:           options.EndpointsTOTP,
		Duo:            options.EndpointsDuo,
		OpenIDConnect:  options.EndpointsOpenIDConnect,
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
	DuoSelfEnrollment      string
	PasskeyLogin           string
	SPNEGO                 string
	RememberMe             string
	ResetPassword          string
	ResetPasswordCustomURL string
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
	PasswordReset  bool
	PasswordChange bool
	WebAuthn       bool
	Passkeys       bool
	SPNEGO         bool
	TOTP           bool
	Duo            bool
	OpenIDConnect  bool

	EndpointsAuthz map[string]schema.ServerEndpointsAuthz
}
