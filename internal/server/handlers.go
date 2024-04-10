package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
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

		switch e := err.(type) {
		case *fasthttp.ErrSmallBuffer:
			statusCode = fasthttp.StatusRequestHeaderFieldsTooLarge
			message = fmt.Sprintf(errFmtMessageServerReadBuffer, cpath)
		case *net.OpError:
			if e.Timeout() {
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
					message = fmt.Sprintf(errFmtMessageServerTLSVersion, tls.VersionName(uint16(version)))
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

//nolint:gocyclo
func handleRouter(config *schema.Configuration, providers middlewares.Providers) fasthttp.RequestHandler {
	log := logging.Logger()

	optsTemplatedFile := NewTemplatedFileOptions(config)

	serveIndexHandler := ServeTemplatedFile(providers.Templates.GetAssetIndexTemplate(), optsTemplatedFile)
	serveOpenAPIHandler := ServeTemplatedOpenAPI(providers.Templates.GetAssetOpenAPIIndexTemplate(), optsTemplatedFile)
	serveOpenAPISpecHandler := ETagRootURL(ServeTemplatedOpenAPI(providers.Templates.GetAssetOpenAPISpecTemplate(), optsTemplatedFile))

	handlerPublicHTML := newPublicHTMLEmbeddedHandler()
	handlerLocales := newLocalesEmbeddedHandler()

	bridge := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeadersBase).Build()

	bridgeSwagger := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeadersRelaxed).Build()

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
	r.HEAD("/locales/{language:[a-z]{1,3}}-{variant:[a-zA-Z0-9-]+}/{namespace:[a-z]+}.json", middlewares.AssetOverride(config.Server.AssetPath, 0, handlerLocales))
	r.GET("/locales/{language:[a-z]{1,3}}-{variant:[a-zA-Z0-9-]+}/{namespace:[a-z]+}.json", middlewares.AssetOverride(config.Server.AssetPath, 0, handlerLocales))

	r.HEAD("/locales/{language:[a-z]{1,3}}/{namespace:[a-z]+}.json", middlewares.AssetOverride(config.Server.AssetPath, 0, handlerLocales))
	r.GET("/locales/{language:[a-z]{1,3}}/{namespace:[a-z]+}.json", middlewares.AssetOverride(config.Server.AssetPath, 0, handlerLocales))

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

		handler := middlewares.Wrap(metricsVRMW, bridge(authz.Handler))

		switch name {
		case "legacy":
			log.
				WithField("path_prefix", pathAuthzLegacy).
				WithField("implementation", endpoint.Implementation).
				WithField("methods", "*").
				Trace("Registering Authz Endpoint")

			r.ANY(pathAuthzLegacy, handler)
			r.ANY(path.Join(pathAuthzLegacy, pathParamAuthzEnvoy), handler)
		default:
			switch endpoint.Implementation {
			case handlers.AuthzImplLegacy.String(), handlers.AuthzImplExtAuthz.String():
				log.
					WithField("path_prefix", uri).
					WithField("implementation", endpoint.Implementation).
					WithField("methods", "*").
					Trace("Registering Authz Endpoint")

				r.ANY(uri, handler)
				r.ANY(path.Join(uri, pathParamAuthzEnvoy), handler)
			default:
				log.
					WithField("path", uri).
					WithField("implementation", endpoint.Implementation).
					WithField("methods", []string{fasthttp.MethodGet, fasthttp.MethodHead}).
					Trace("Registering Authz Endpoint")

				r.GET(uri, handler)
				r.HEAD(uri, handler)
			}
		}
	}

	r.POST("/api/checks/safe-redirection", middlewareAPI(handlers.CheckSafeRedirectionPOST))

	delayFunc := middlewares.TimingAttackDelay(10, 250, 85, time.Second, true)

	r.POST("/api/firstfactor", middlewareAPI(handlers.FirstFactorPOST(delayFunc)))
	r.POST("/api/logout", middlewareAPI(handlers.LogoutPOST))

	// Only register endpoints if forgot password is not disabled.
	if !config.AuthenticationBackend.PasswordReset.Disable &&
		config.AuthenticationBackend.PasswordReset.CustomURL.String() == "" {
		// Password reset related endpoints.
		r.POST("/api/reset-password/identity/start", middlewareAPI(handlers.ResetPasswordIdentityStart))
		r.POST("/api/reset-password/identity/finish", middlewareAPI(handlers.ResetPasswordIdentityFinish))

		r.POST("/api/reset-password", middlewareAPI(handlers.ResetPasswordPOST))
		r.DELETE("/api/reset-password", middlewareAPI(handlers.ResetPasswordDELETE))
	}

	// Information about the user.
	r.GET("/api/user/info", middleware1FA(handlers.UserInfoGET))
	r.POST("/api/user/info", middleware1FA(handlers.UserInfoPOST))
	r.POST("/api/user/info/2fa_method", middleware1FA(handlers.MethodPreferencePOST))

	// User Session Elevation.
	middlewareDelaySecond := middlewares.ArbitraryDelay(time.Second)

	r.GET("/api/user/session/elevation", middleware1FA(handlers.UserSessionElevationGET))
	r.POST("/api/user/session/elevation", middleware1FA(handlers.UserSessionElevationPOST))
	r.PUT("/api/user/session/elevation", middlewareDelaySecond(middleware1FA(handlers.UserSessionElevationPUT)))

	r.DELETE("/api/user/session/elevation/{id}", middlewareAPI(handlers.UserSessionElevateDELETE))

	if !config.TOTP.Disable {
		// TOTP related endpoints.
		r.GET("/api/secondfactor/totp", middleware1FA(handlers.TimeBasedOneTimePasswordGET))
		r.POST("/api/secondfactor/totp", middleware1FA(handlers.TimeBasedOneTimePasswordPOST))
		r.DELETE("/api/secondfactor/totp", middleware1FA(handlers.TOTPConfigurationDELETE))

		r.GET("/api/secondfactor/totp/register", middlewareElevated1FA(handlers.TOTPRegisterGET))
		r.PUT("/api/secondfactor/totp/register", middlewareElevated1FA(handlers.TOTPRegisterPUT))
		r.POST("/api/secondfactor/totp/register", middlewareElevated1FA(handlers.TOTPRegisterPOST))
		r.DELETE("/api/secondfactor/totp/register", middlewareElevated1FA(handlers.TOTPRegisterDELETE))
	}

	if !config.WebAuthn.Disable {
		r.GET("/api/secondfactor/webauthn", middleware1FA(handlers.WebAuthnAssertionGET))
		r.POST("/api/secondfactor/webauthn", middleware1FA(handlers.WebAuthnAssertionPOST))

		// Management of the WebAuthn credentials.
		r.GET("/api/secondfactor/webauthn/credentials", middleware1FA(handlers.WebAuthnCredentialsGET))

		r.PUT("/api/secondfactor/webauthn/credential/register", middlewareElevated1FA(handlers.WebAuthnRegistrationPUT))
		r.POST("/api/secondfactor/webauthn/credential/register", middlewareElevated1FA(handlers.WebAuthnRegistrationPOST))
		r.DELETE("/api/secondfactor/webauthn/credential/register", middlewareElevated1FA(handlers.WebAuthnRegistrationDELETE))

		r.PUT("/api/secondfactor/webauthn/credential/{credentialID}", middlewareElevated1FA(handlers.WebAuthnCredentialPUT))
		r.DELETE("/api/secondfactor/webauthn/credential/{credentialID}", middlewareElevated1FA(handlers.WebAuthnCredentialDELETE))
	}

	// Configure DUO api endpoint only if configuration exists.
	if !config.DuoAPI.Disable {
		var duoAPI duo.API
		if os.Getenv("ENVIRONMENT") == dev {
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

		r.GET("/api/secondfactor/duo_devices", middleware1FA(handlers.DuoDevicesGET(duoAPI)))
		r.POST("/api/secondfactor/duo", middleware1FA(handlers.DuoPOST(duoAPI)))
		r.POST("/api/secondfactor/duo_device", middleware1FA(handlers.DuoDevicePOST))
	}

	if config.Server.Endpoints.EnablePprof {
		r.GET("/debug/pprof/{name?}", pprofhandler.PprofHandler)
	}

	if config.Server.Endpoints.EnableExpvars {
		r.GET("/debug/vars", expvarhandler.ExpvarHandler)
	}

	if providers.OpenIDConnect != nil {
		bridgeOIDC := middlewares.NewBridgeBuilder(*config, providers).WithPreMiddlewares(
			middlewares.SecurityHeadersBase, middlewares.SecurityHeadersCSPNoneOpenIDConnect, middlewares.SecurityHeadersNoStore,
		).Build()

		r.GET("/api/oidc/consent", bridgeOIDC(handlers.OpenIDConnectConsentGET))
		r.POST("/api/oidc/consent", bridgeOIDC(handlers.OpenIDConnectConsentPOST))

		allowedOrigins := utils.StringSliceFromURLs(config.IdentityProviders.OIDC.CORS.AllowedOrigins)

		r.OPTIONS(oidc.EndpointPathWellKnownOpenIDConfiguration, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.EndpointPathWellKnownOpenIDConfiguration, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, "openid_configuration"), policyCORSPublicGET.Middleware(bridgeOIDC(handlers.OpenIDConnectConfigurationWellKnownGET))))

		r.OPTIONS(oidc.EndpointPathWellKnownOAuthAuthorizationServer, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.EndpointPathWellKnownOAuthAuthorizationServer, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, "oauth_configuration"), policyCORSPublicGET.Middleware(bridgeOIDC(handlers.OAuthAuthorizationServerWellKnownGET))))

		r.OPTIONS(oidc.EndpointPathJWKs, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.EndpointPathJWKs, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, "jwks"), policyCORSPublicGET.Middleware(middlewareAPI(handlers.JSONWebKeySetGET))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/jwks", policyCORSPublicGET.HandleOPTIONS)
		r.GET("/api/oidc/jwks", middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, "jwks"), policyCORSPublicGET.Middleware(bridgeOIDC(handlers.JSONWebKeySetGET))))

		policyCORSAuthorization := middlewares.NewCORSPolicyBuilder().
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodGet, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.EndpointAuthorization, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		authorization := middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointAuthorization), policyCORSAuthorization.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectAuthorization))))

		r.OPTIONS(oidc.EndpointPathAuthorization, policyCORSAuthorization.HandleOnlyOPTIONS)
		r.GET(oidc.EndpointPathAuthorization, authorization)
		r.POST(oidc.EndpointPathAuthorization, authorization)

		// TODO (james-d-elliott): Remove in GA. This is a legacy endpoint.
		r.OPTIONS("/api/oidc/authorize", policyCORSAuthorization.HandleOnlyOPTIONS)
		r.GET("/api/oidc/authorize", authorization)
		r.POST("/api/oidc/authorize", authorization)

		policyCORSPAR := middlewares.NewCORSPolicyBuilder().
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSliceFold(oidc.EndpointPushedAuthorizationRequest, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.EndpointPathPushedAuthorizationRequest, policyCORSPAR.HandleOnlyOPTIONS)
		r.POST(oidc.EndpointPathPushedAuthorizationRequest, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointPushedAuthorizationRequest), policyCORSPAR.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectPushedAuthorizationRequest)))))

		policyCORSToken := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.EndpointToken, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.EndpointPathToken, policyCORSToken.HandleOPTIONS)
		r.POST(oidc.EndpointPathToken, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointToken), policyCORSToken.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectTokenPOST)))))

		policyCORSUserinfo := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodGet, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.EndpointUserinfo, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.EndpointPathUserinfo, policyCORSUserinfo.HandleOPTIONS)
		r.GET(oidc.EndpointPathUserinfo, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointUserinfo), policyCORSUserinfo.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectUserinfo)))))
		r.POST(oidc.EndpointPathUserinfo, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointUserinfo), policyCORSUserinfo.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectUserinfo)))))

		policyCORSIntrospection := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.EndpointIntrospection, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.EndpointPathIntrospection, policyCORSIntrospection.HandleOPTIONS)
		r.POST(oidc.EndpointPathIntrospection, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointIntrospection), policyCORSIntrospection.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthIntrospectionPOST)))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/introspect", policyCORSIntrospection.HandleOPTIONS)
		r.POST("/api/oidc/introspect", middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointIntrospection), policyCORSIntrospection.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthIntrospectionPOST)))))

		policyCORSRevocation := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.EndpointRevocation, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.EndpointPathRevocation, policyCORSRevocation.HandleOPTIONS)
		r.POST(oidc.EndpointPathRevocation, middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointRevocation), policyCORSRevocation.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthRevocationPOST)))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/revoke", policyCORSRevocation.HandleOPTIONS)
		r.POST("/api/oidc/revoke", middlewares.Wrap(middlewares.NewMetricsRequestOpenIDConnect(providers.Metrics, oidc.EndpointRevocation), policyCORSRevocation.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthRevocationPOST)))))
	}

	r.RedirectFixedPath = false
	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = handleMethodNotAllowed
	r.NotFound = handleNotFound(bridge(serveIndexHandler))

	handler := middlewares.LogRequest(r.Handler)
	if config.Server.Address.RouterPath() != "/" {
		handler = middlewares.StripPath(config.Server.Address.RouterPath())(handler)
	}

	handler = middlewares.MultiWrap(handler, middlewares.RecoverPanic, middlewares.NewMetricsRequest(providers.Metrics))

	return handler
}

func handleMetrics(path string) fasthttp.RequestHandler {
	r := router.New()

	r.GET(path, fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))

	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = handlers.Status(fasthttp.StatusMethodNotAllowed)
	r.NotFound = handlers.Status(fasthttp.StatusNotFound)

	return r.Handler
}
