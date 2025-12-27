package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"path"
	"strings"
	"time"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/fasthttp/router"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/expvarhandler"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/valyala/fasthttp/pprofhandler"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

// Replacement for the default error handler in fasthttp.
func handleError(cpath string) func(ctx *fasthttp.RequestCtx, err error) {
	headerXForwardedFor := []byte(fasthttp.HeaderXForwardedFor)

	getRemoteIP := func(ctx *fasthttp.RequestCtx) string {
		if hdr := ctx.Request.Header.PeekBytes(headerXForwardedFor); hdr != nil {
			ips := strings.Split(string(hdr), ",")

			if len(ips) > 0 {
				return strings.Trim(ips[0], " ")
			}
		}

		return ctx.RemoteIP().String()
	}

	return func(ctx *fasthttp.RequestCtx, err error) {
		var (
			statusCode int
			message    string
		)

		var (
			fsbErr *fasthttp.ErrSmallBuffer
			noErr  *net.OpError
		)

		switch {
		case errors.As(err, &fsbErr):
			statusCode = fasthttp.StatusRequestHeaderFieldsTooLarge
			message = fmt.Sprintf(errFmtMessageServerReadBuffer, cpath)
		case errors.As(err, &noErr):
			if noErr.Timeout() {
				statusCode = fasthttp.StatusRequestTimeout
				message = errMessageServerRequestTimeout
			} else {
				statusCode = fasthttp.StatusBadRequest
				message = errMessageServerNetwork
			}
		default:
			statusCode = fasthttp.StatusBadRequest
			message = errMessageServerGeneric

			if ctx.IsTLS() {
				// We don't need to bother doing the TLS handshake check if we're sure it's already a TLS connection.
				break
			}

			if matches := reTLSRequestOnPlainTextSocketErr.FindStringSubmatch(err.Error()); len(matches) == 3 {
				if version, verr := utils.TLSVersionFromBytesString(matches[1] + matches[2]); verr == nil && version != -1 {
					statusCode = fasthttp.StatusBadRequest
					message = fmt.Sprintf(errFmtMessageServerTLSVersion, tls.VersionName(uint16(version))) //nolint:gosec // This conversion is safe as the only versions potentially returned are from the crypto/tls pkg.
				}
			}
		}

		logging.Logger().WithFields(logrus.Fields{
			logging.FieldMethod:     string(ctx.Method()),
			logging.FieldPath:       string(ctx.Path()),
			logging.FieldRemoteIP:   getRemoteIP(ctx),
			logging.FieldStatusCode: statusCode,
		}).WithError(err).Error(message)

		handlers.SetStatusCodeResponse(ctx, statusCode)
	}
}

func handleNotFound(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		uri := strings.ToLower(string(ctx.Path()))

		for i := 0; i < len(dirsHTTPServer); i++ {
			if uri == dirsHTTPServer[i].name || strings.HasPrefix(uri, dirsHTTPServer[i].prefix) {
				handlers.SetStatusCodeResponse(ctx, fasthttp.StatusNotFound)

				return
			}
		}

		next(ctx)
	}
}

func handleMethodNotAllowed(ctx *fasthttp.RequestCtx) {
	middlewares.SetContentTypeTextPlain(ctx)

	ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
	ctx.SetBodyString(fmt.Sprintf("%d %s", fasthttp.StatusMethodNotAllowed, fasthttp.StatusMessage(fasthttp.StatusMethodNotAllowed)))
}

type RegisterRoutesBridgedFunc = func(r *router.Router, config *schema.Configuration, providers middlewares.Providers, bridge middlewares.Bridge)

//nolint:gocyclo
func handlerMain(config *schema.Configuration, providers middlewares.Providers) (handler fasthttp.RequestHandler, err error) {
	optsTemplatedFile := NewTemplatedFileOptions(config)

	serveIndexHandler := ServeTemplatedFile(providers.Templates.GetAssetIndexTemplate(), optsTemplatedFile)
	serveOpenAPIHandler := ServeTemplatedOpenAPI(providers.Templates.GetAssetOpenAPIIndexTemplate(), optsTemplatedFile)
	serveOpenAPISpecHandler := ETagRootURL(ServeTemplatedOpenAPI(providers.Templates.GetAssetOpenAPISpecTemplate(), optsTemplatedFile))

	handlerPublicHTML := newPublicHTMLEmbeddedHandler()

	var (
		handlerLocales, handlerLocalesList middlewares.RequestHandler
	)

	if handlerLocales, err = newLocalesEmbeddedHandler(); err != nil {
		return nil, err
	}

	if handlerLocalesList, err = newLocalesListHandler(); err != nil {
		return nil, err
	}

	bridge := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeadersBase).Build()

	bridgeSwagger := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeadersRelaxed, middlewares.SecurityHeadersCSPSelf).Build()

	policyCORSPublicGET := middlewares.NewCORSPolicyBuilder().
		WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodGet).
		WithAllowedOrigins("*").
		Build()

	r := router.New()

	// Static Assets.
	r.HEAD("/", bridge(serveIndexHandler))
	r.GET("/", bridge(serveIndexHandler))

	for _, f := range filesRoot {
		r.HEAD("/"+f, handlerPublicHTML)
		r.GET("/"+f, handlerPublicHTML)
	}

	r.HEAD("/favicon.ico", middlewares.AssetOverride(config.Server.AssetPath, 0, handlerPublicHTML))
	r.GET("/favicon.ico", middlewares.AssetOverride(config.Server.AssetPath, 0, handlerPublicHTML))

	r.HEAD("/static/media/logo.png", middlewares.AssetOverride(config.Server.AssetPath, 2, handlerPublicHTML))
	r.GET("/static/media/logo.png", middlewares.AssetOverride(config.Server.AssetPath, 2, handlerPublicHTML))

	r.HEAD("/static/{filepath:*}", handlerPublicHTML)
	r.GET("/static/{filepath:*}", handlerPublicHTML)

	// Locales.
	r.GET("/locales", bridge(handlerLocalesList))

	r.HEAD("/locales/{language:[a-z]{1,3}}-{variant:[a-zA-Z0-9-]+}/{namespace:[a-z]+}.json", middlewares.AssetOverride(config.Server.AssetPath, 0, bridge(handlerLocales)))
	r.GET("/locales/{language:[a-z]{1,3}}-{variant:[a-zA-Z0-9-]+}/{namespace:[a-z]+}.json", middlewares.AssetOverride(config.Server.AssetPath, 0, bridge(handlerLocales)))

	r.HEAD("/locales/{language:[a-z]{1,3}}/{namespace:[a-z]+}.json", middlewares.AssetOverride(config.Server.AssetPath, 0, bridge(handlerLocales)))
	r.GET("/locales/{language:[a-z]{1,3}}/{namespace:[a-z]+}.json", middlewares.AssetOverride(config.Server.AssetPath, 0, bridge(handlerLocales)))

	// Swagger.
	r.HEAD(prefixAPI, bridgeSwagger(serveOpenAPIHandler))
	r.GET(prefixAPI, bridgeSwagger(serveOpenAPIHandler))
	r.OPTIONS(prefixAPI, policyCORSPublicGET.HandleOPTIONS)

	r.HEAD("/api/index.html", bridgeSwagger(serveOpenAPIHandler))
	r.GET("/api/index.html", bridgeSwagger(serveOpenAPIHandler))
	r.OPTIONS("/api/index.html", policyCORSPublicGET.HandleOPTIONS)

	r.HEAD("/api/openapi.yml", policyCORSPublicGET.Middleware(bridgeSwagger(serveOpenAPISpecHandler)))
	r.GET("/api/openapi.yml", policyCORSPublicGET.Middleware(bridgeSwagger(serveOpenAPISpecHandler)))
	r.OPTIONS("/api/openapi.yml", policyCORSPublicGET.HandleOPTIONS)

	for _, file := range filesSwagger {
		r.HEAD(prefixAPI+file, handlerPublicHTML)
		r.GET(prefixAPI+file, handlerPublicHTML)
	}

	middlewareAPI := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeadersBase, middlewares.SecurityHeadersNoStore, middlewares.SecurityHeadersCSPNone).
		Build()

	middleware1FA := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeadersBase, middlewares.SecurityHeadersNoStore, middlewares.SecurityHeadersCSPNone).
		WithPostMiddlewares(middlewares.Require1FA).
		Build()

	middlewareElevated1FA := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeadersBase, middlewares.SecurityHeadersNoStore, middlewares.SecurityHeadersCSPNone).
		WithPostMiddlewares(middlewares.RequireElevated).
		Build()

	r.HEAD("/api/health", middlewareAPI(handlers.HealthGET))
	r.GET("/api/health", middlewareAPI(handlers.HealthGET))

	r.GET("/api/state", middlewareAPI(handlers.StateGET))

	r.GET("/api/configuration", middleware1FA(handlers.ConfigurationGET))

	r.GET("/api/configuration/password-policy", middlewareAPI(handlers.PasswordPolicyConfigurationGET))

	metricsVRMW := middlewares.NewMetricsAuthzRequest(providers.Metrics)

	for name, endpoint := range config.Server.Endpoints.Authz {
		uri := path.Join(pathAuthz, name)

		authz := handlers.NewAuthzBuilder().WithConfig(config).WithEndpointConfig(endpoint).Build()

		handlerAuthz := middlewares.Wrap(metricsVRMW, bridge(authz.Handler))

		switch name {
		case "legacy":
			r.ANY(pathAuthzLegacy, handlerAuthz)
			r.ANY(path.Join(pathAuthzLegacy, pathParamAuthzEnvoy), handlerAuthz)
		default:
			switch endpoint.Implementation {
			case handlers.AuthzImplLegacy.String(), handlers.AuthzImplExtAuthz.String():
				r.ANY(uri, handlerAuthz)
				r.ANY(path.Join(uri, pathParamAuthzEnvoy), handlerAuthz)
			default:
				r.GET(uri, handlerAuthz)
				r.HEAD(uri, handlerAuthz)
			}
		}
	}

	r.POST("/api/checks/safe-redirection", middlewareAPI(handlers.CheckSafeRedirectionPOST))

	funcDelayPassword := middlewares.TimingAttackDelay(10, 250, 85, time.Second, true)

	r.POST("/api/firstfactor", middlewareAPI(handlers.FirstFactorPasswordPOST(funcDelayPassword)))
	r.POST("/api/firstfactor/reauthenticate", middleware1FA(handlers.FirstFactorReauthenticatePOST(funcDelayPassword)))
	r.POST("/api/logout", middlewareAPI(handlers.LogoutPOST))

	// Only register endpoints if forgot password is not disabled.
	if !config.AuthenticationBackend.PasswordReset.Disable &&
		config.AuthenticationBackend.PasswordReset.CustomURL.String() == "" {
		resetPasswordTokenRL := middlewares.NewIPRateLimit(middlewares.NewRateLimitBucketsConfig(config.Server.Endpoints.RateLimits.ResetPasswordFinish)...)

		// Password reset related endpoints.
		r.POST("/api/reset-password/identity/start", middlewareAPI(middlewares.NewRateLimitHandler(config.Server.Endpoints.RateLimits.ResetPasswordStart, handlers.ResetPasswordIdentityStart)))
		r.POST("/api/reset-password/identity/finish", middlewareAPI(resetPasswordTokenRL(handlers.ResetPasswordIdentityFinish)))

		r.POST("/api/reset-password", middlewareAPI(handlers.ResetPasswordPOST))
		r.DELETE("/api/reset-password", middlewareAPI(resetPasswordTokenRL(handlers.ResetPasswordDELETE)))
	}

	if !config.AuthenticationBackend.PasswordChange.Disable {
		r.POST("/api/change-password", middlewareElevated1FA(handlers.ChangePasswordPOST))
	}

	// Information about the user.
	r.GET("/api/user/info", middleware1FA(handlers.UserInfoGET))
	r.POST("/api/user/info", middleware1FA(handlers.UserInfoPOST))
	r.POST("/api/user/info/2fa_method", middleware1FA(handlers.MethodPreferencePOST))

	// User Session Elevation.
	middlewareElevatePOST := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeadersBase, middlewares.SecurityHeadersNoStore, middlewares.SecurityHeadersCSPNone).
		WithPostMiddlewares(middlewares.NewRateLimit(config.Server.Endpoints.RateLimits.SessionElevationStart), middlewares.Require1FA).
		Build()

	middlewareElevatePUT := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeadersBase, middlewares.SecurityHeadersNoStore, middlewares.SecurityHeadersCSPNone, middlewares.ArbitraryDelay(time.Second)).
		WithPostMiddlewares(middlewares.NewRateLimit(config.Server.Endpoints.RateLimits.SessionElevationFinish), middlewares.Require1FA).
		Build()

	r.GET("/api/user/session/elevation", middleware1FA(handlers.UserSessionElevationGET))
	r.POST("/api/user/session/elevation", middlewareElevatePOST(handlers.UserSessionElevationPOST))
	r.PUT("/api/user/session/elevation", middlewareElevatePUT(handlers.UserSessionElevationPUT))

	r.DELETE("/api/user/session/elevation/{id}", middlewareAPI(handlers.UserSessionElevateDELETE))

	if !config.TOTP.Disable {
		middlewareRateLimitTOTP := middlewares.NewBridgeBuilder(*config, providers).
			WithPreMiddlewares(middlewares.SecurityHeadersBase, middlewares.SecurityHeadersNoStore, middlewares.SecurityHeadersCSPNone).
			WithPostMiddlewares(middlewares.NewRateLimit(config.Server.Endpoints.RateLimits.SecondFactorTOTP), middlewares.Require1FA).
			Build()

		// TOTP related endpoints.
		r.GET("/api/secondfactor/totp", middleware1FA(handlers.TimeBasedOneTimePasswordGET))
		r.POST("/api/secondfactor/totp", middlewareRateLimitTOTP(handlers.TimeBasedOneTimePasswordPOST))
		r.DELETE("/api/secondfactor/totp", middleware1FA(handlers.TOTPConfigurationDELETE))

		r.GET("/api/secondfactor/totp/register", middlewareElevated1FA(handlers.TOTPRegisterGET))
		r.PUT("/api/secondfactor/totp/register", middlewareElevated1FA(handlers.TOTPRegisterPUT))
		r.POST("/api/secondfactor/totp/register", middlewareElevated1FA(handlers.TOTPRegisterPOST))
		r.DELETE("/api/secondfactor/totp/register", middlewareElevated1FA(handlers.TOTPRegisterDELETE))
	}

	if !config.WebAuthn.Disable {
		r.GET("/api/secondfactor/webauthn", middleware1FA(handlers.WebAuthnAssertionGET))
		r.POST("/api/secondfactor/webauthn", middleware1FA(handlers.WebAuthnAssertionPOST))

		if config.WebAuthn.EnablePasskeyLogin {
			r.GET("/api/firstfactor/passkey", middlewareAPI(handlers.FirstFactorPasskeyGET))
			r.POST("/api/firstfactor/passkey", middlewareAPI(handlers.FirstFactorPasskeyPOST))
			r.POST("/api/secondfactor/password", middleware1FA(handlers.SecondFactorPasswordPOST(funcDelayPassword)))
		}

		// Management of the WebAuthn credentials.
		r.GET("/api/secondfactor/webauthn/credentials", middleware1FA(handlers.WebAuthnCredentialsGET))

		r.PUT("/api/secondfactor/webauthn/credential/register", middlewareElevated1FA(handlers.WebAuthnRegistrationPUT))
		r.POST("/api/secondfactor/webauthn/credential/register", middlewareElevated1FA(handlers.WebAuthnRegistrationPOST))
		r.DELETE("/api/secondfactor/webauthn/credential/register", middlewareElevated1FA(handlers.WebAuthnRegistrationDELETE))

		r.PUT("/api/secondfactor/webauthn/credential/{credentialID}", middlewareElevated1FA(handlers.WebAuthnCredentialPUT))
		r.DELETE("/api/secondfactor/webauthn/credential/{credentialID}", middlewareElevated1FA(handlers.WebAuthnCredentialDELETE))
	}

	// SPNEGO KRB5 Authentication related Endpoint.
	if !config.SPNEGO.Enabled {
		r.POST("/api/firstfactor/spnego", middlewareAPI(handlers.FirstFactorSPNEGOPOST))
	}

	// Configure DUO api endpoint only if configuration exists.
	if !config.DuoAPI.Disable {
		var duoAPI duo.Provider

		if utils.Dev {
			duoAPI = duo.NewDuoAPI(duoapi.NewDuoApi(
				config.DuoAPI.IntegrationKey,
				config.DuoAPI.SecretKey,
				config.DuoAPI.Hostname, "", duoapi.SetInsecure()))
		} else {
			duoAPI = duo.NewDuoAPI(duoapi.NewDuoApi(
				config.DuoAPI.IntegrationKey,
				config.DuoAPI.SecretKey,
				config.DuoAPI.Hostname, ""))
		}

		middlewareRateLimitDuo := middlewares.NewBridgeBuilder(*config, providers).
			WithPreMiddlewares(middlewares.SecurityHeadersBase, middlewares.SecurityHeadersNoStore, middlewares.SecurityHeadersCSPNone).
			WithPostMiddlewares(middlewares.NewRateLimit(config.Server.Endpoints.RateLimits.SecondFactorDuo), middlewares.Require1FA).
			Build()

		r.GET("/api/secondfactor/duo", middleware1FA(handlers.DuoGET))
		r.GET("/api/secondfactor/duo_devices", middleware1FA(handlers.DuoDevicesGET(duoAPI)))
		r.POST("/api/secondfactor/duo", middlewareRateLimitDuo(handlers.DuoPOST(duoAPI)))
		r.POST("/api/secondfactor/duo_device", middleware1FA(handlers.DuoDevicePOST))
	}

	if config.Server.Endpoints.EnablePprof {
		r.GET("/debug/pprof/{name?}", pprofhandler.PprofHandler)
	}

	if config.Server.Endpoints.EnableExpvars {
		r.GET("/debug/vars", expvarhandler.ExpvarHandler)
	}

	if providers.OpenIDConnect != nil {
		RegisterOpenIDConnectRoutes(r, config, providers)
	}

	r.RedirectFixedPath = false
	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = handleMethodNotAllowed
	r.NotFound = handleNotFound(bridge(serveIndexHandler))

	handler = middlewares.LogRequest(r.Handler)
	if config.Server.Address.RouterPath() != "/" {
		handler = middlewares.StripPath(config.Server.Address.RouterPath())(handler)
	}

	handler = middlewares.MultiWrap(handler, middlewares.RecoverPanic, middlewares.NewMetricsRequest(providers.Metrics))

	return handler, nil
}

// RegisterOpenIDConnectRoutes handles registration of OpenID Connect 1.0 routes.
func RegisterOpenIDConnectRoutes(r *router.Router, config *schema.Configuration, providers middlewares.Providers) {
	middlewareAPI := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeadersBase, middlewares.SecurityHeadersNoStore, middlewares.SecurityHeadersCSPNone).
		Build()

	policyCORSPublicGET := middlewares.NewCORSPolicyBuilder().
		WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodGet).
		WithAllowedOrigins("*").
		Build()

	bridge := middlewares.NewBridgeBuilder(*config, providers).WithPreMiddlewares(
		middlewares.SecurityHeadersBase, middlewares.SecurityHeadersCSPNoneOpenIDConnect, middlewares.SecurityHeadersNoStore,
	).Build()

	r.GET(oidc.EndpointPathConsent, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointConsent), bridge(handlers.OAuth2ConsentGET)))
	r.POST(oidc.EndpointPathConsent, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointConsent), bridge(handlers.OAuth2ConsentPOST)))

	allowedOrigins := utils.StringSliceFromURLs(config.IdentityProviders.OIDC.CORS.AllowedOrigins)

	r.OPTIONS(oidc.EndpointPathWellKnownOAuthAuthorizationServer, policyCORSPublicGET.HandleOPTIONS)
	r.GET(oidc.EndpointPathWellKnownOAuthAuthorizationServer, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, "oauth_configuration"), policyCORSPublicGET.Middleware(bridge(handlers.WellKnownOAuthAuthorizationServerGET))))

	r.OPTIONS(oidc.EndpointPathWellKnownOpenIDConfiguration, policyCORSPublicGET.HandleOPTIONS)
	r.GET(oidc.EndpointPathWellKnownOpenIDConfiguration, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, "openid_configuration"), policyCORSPublicGET.Middleware(bridge(handlers.WellKnownOpenIDConfigurationGET))))

	r.OPTIONS(oidc.EndpointPathJWKs, policyCORSPublicGET.HandleOPTIONS)
	r.GET(oidc.EndpointPathJWKs, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, "jwks"), policyCORSPublicGET.Middleware(middlewareAPI(handlers.OAuth2JSONWebKeySetGET))))

	policyCORSAuthorization := middlewares.NewCORSPolicyBuilder().
		WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodGet, fasthttp.MethodPost).
		WithAllowedOrigins(allowedOrigins...).
		WithEnabled(utils.IsStringInSlice(oidc.EndpointAuthorization, config.IdentityProviders.OIDC.CORS.Endpoints)).
		Build()

	authorizationGET := middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointAuthorization), policyCORSAuthorization.Middleware(bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuth2AuthorizationGET))))
	authorizationPOST := policyCORSAuthorization.Middleware(bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuth2AuthorizationPOST)))

	r.OPTIONS(oidc.EndpointPathAuthorization, policyCORSAuthorization.HandleOnlyOPTIONS)
	r.GET(oidc.EndpointPathAuthorization, authorizationGET)
	r.POST(oidc.EndpointPathAuthorization, authorizationPOST)

	policyCORSDeviceAuthorization := middlewares.NewCORSPolicyBuilder().
		WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
		WithAllowedOrigins(allowedOrigins...).
		WithEnabled(utils.IsStringInSlice(oidc.EndpointDeviceAuthorization, config.IdentityProviders.OIDC.CORS.Endpoints)).
		Build()

	r.OPTIONS(oidc.EndpointPathDeviceAuthorization, policyCORSDeviceAuthorization.HandleOnlyOPTIONS)
	r.POST(oidc.EndpointPathDeviceAuthorization, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointDeviceAuthorization), policyCORSDeviceAuthorization.Middleware(bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuth2DeviceAuthorizationPOST)))))
	r.PUT(oidc.EndpointPathDeviceAuthorization, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointDeviceAuthorization), bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuth2DeviceAuthorizationPUT))))

	policyCORSPAR := middlewares.NewCORSPolicyBuilder().
		WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
		WithAllowedOrigins(allowedOrigins...).
		WithEnabled(utils.IsStringInSliceFold(oidc.EndpointPushedAuthorizationRequest, config.IdentityProviders.OIDC.CORS.Endpoints)).
		Build()

	r.OPTIONS(oidc.EndpointPathPushedAuthorizationRequest, policyCORSPAR.HandleOnlyOPTIONS)
	r.POST(oidc.EndpointPathPushedAuthorizationRequest, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointPushedAuthorizationRequest), policyCORSPAR.Middleware(bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuth2PushedAuthorizationRequest)))))

	policyCORSToken := middlewares.NewCORSPolicyBuilder().
		WithAllowCredentials(true).
		WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
		WithAllowedOrigins(allowedOrigins...).
		WithEnabled(utils.IsStringInSlice(oidc.EndpointToken, config.IdentityProviders.OIDC.CORS.Endpoints)).
		Build()

	r.OPTIONS(oidc.EndpointPathToken, policyCORSToken.HandleOPTIONS)
	r.POST(oidc.EndpointPathToken, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointToken), policyCORSToken.Middleware(bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuth2TokenPOST)))))

	policyCORSUserinfo := middlewares.NewCORSPolicyBuilder().
		WithAllowCredentials(true).
		WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodGet, fasthttp.MethodPost).
		WithAllowedOrigins(allowedOrigins...).
		WithEnabled(utils.IsStringInSlice(oidc.EndpointUserinfo, config.IdentityProviders.OIDC.CORS.Endpoints)).
		Build()

	r.OPTIONS(oidc.EndpointPathUserinfo, policyCORSUserinfo.HandleOPTIONS)
	r.GET(oidc.EndpointPathUserinfo, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointUserinfo), policyCORSUserinfo.Middleware(bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectUserinfo)))))
	r.POST(oidc.EndpointPathUserinfo, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointUserinfo), policyCORSUserinfo.Middleware(bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectUserinfo)))))

	policyCORSIntrospection := middlewares.NewCORSPolicyBuilder().
		WithAllowCredentials(true).
		WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
		WithAllowedOrigins(allowedOrigins...).
		WithEnabled(utils.IsStringInSlice(oidc.EndpointIntrospection, config.IdentityProviders.OIDC.CORS.Endpoints)).
		Build()

	r.OPTIONS(oidc.EndpointPathIntrospection, policyCORSIntrospection.HandleOPTIONS)
	r.POST(oidc.EndpointPathIntrospection, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointIntrospection), policyCORSIntrospection.Middleware(bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuth2IntrospectionPOST)))))

	policyCORSRevocation := middlewares.NewCORSPolicyBuilder().
		WithAllowCredentials(true).
		WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
		WithAllowedOrigins(allowedOrigins...).
		WithEnabled(utils.IsStringInSlice(oidc.EndpointRevocation, config.IdentityProviders.OIDC.CORS.Endpoints)).
		Build()

	r.OPTIONS(oidc.EndpointPathRevocation, policyCORSRevocation.HandleOPTIONS)
	r.POST(oidc.EndpointPathRevocation, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointRevocation), policyCORSRevocation.Middleware(bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuth2RevocationPOST)))))
}

func handlerMetrics(path string) fasthttp.RequestHandler {
	r := router.New()

	r.GET(path, fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))

	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = handlers.Status(fasthttp.StatusMethodNotAllowed)
	r.NotFound = handlers.Status(fasthttp.StatusNotFound)

	return r.Handler
}
