package server

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"

	"github.com/duosecurity/duo_api_golang"
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
func handleError() func(ctx *fasthttp.RequestCtx, err error) {
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
			message = "Request from client exceeded the server buffer sizes."
		case *net.OpError:
			switch {
			case e.Timeout():
				statusCode = fasthttp.StatusRequestTimeout
				message = "Timeout network error occurred while handling request from the client."
			case e.Temporary():
				statusCode = fasthttp.StatusServiceUnavailable
				message = "Temporary network error occurred while handling the request from the client."
			default:
				statusCode = fasthttp.StatusBadRequest
				message = "An unknown network error occurred while handling a request from the client."
			}
		default:
			statusCode = fasthttp.StatusBadRequest
			message = "An unknown error occurred while handling a request from the client."
		}

		logging.Logger().WithFields(logrus.Fields{
			"method":      string(ctx.Method()),
			"path":        string(ctx.Path()),
			"remote_ip":   getRemoteIP(ctx),
			"status_code": statusCode,
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
		WithPreMiddlewares(middlewares.SecurityHeaders).Build()

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
	r.HEAD("/api/", bridge(serveOpenAPIHandler))
	r.GET("/api/", bridge(serveOpenAPIHandler))
	r.OPTIONS("/api/", policyCORSPublicGET.HandleOPTIONS)

	r.HEAD("/api/index.html", bridge(serveOpenAPIHandler))
	r.GET("/api/index.html", bridge(serveOpenAPIHandler))
	r.OPTIONS("/api/index.html", policyCORSPublicGET.HandleOPTIONS)

	r.HEAD("/api/openapi.yml", policyCORSPublicGET.Middleware(bridge(serveOpenAPISpecHandler)))
	r.GET("/api/openapi.yml", policyCORSPublicGET.Middleware(bridge(serveOpenAPISpecHandler)))
	r.OPTIONS("/api/openapi.yml", policyCORSPublicGET.HandleOPTIONS)

	for _, file := range filesSwagger {
		r.HEAD("/api/"+file, handlerPublicHTML)
		r.GET("/api/"+file, handlerPublicHTML)
	}

	middlewareAPI := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeaders, middlewares.SecurityHeadersNoStore, middlewares.SecurityHeadersCSPNone).
		Build()

	middleware1FA := middlewares.NewBridgeBuilder(*config, providers).
		WithPreMiddlewares(middlewares.SecurityHeaders, middlewares.SecurityHeadersNoStore, middlewares.SecurityHeadersCSPNone).
		WithPostMiddlewares(middlewares.Require1FA).
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
				WithField("impl", endpoint.Implementation).
				WithField("methods", []string{"*"}).
				Trace("Registering Authz Endpoint")

			r.ANY(pathAuthzLegacy, handler)
			r.ANY(path.Join(pathAuthzLegacy, pathParamAuthzEnvoy), handler)
		default:
			switch endpoint.Implementation {
			case handlers.AuthzImplLegacy.String(), handlers.AuthzImplExtAuthz.String():
				log.
					WithField("path_prefix", uri).
					WithField("impl", endpoint.Implementation).
					WithField("methods", []string{"*"}).
					Trace("Registering Authz Endpoint")

				r.ANY(uri, handler)
				r.ANY(path.Join(uri, pathParamAuthzEnvoy), handler)
			default:
				log.
					WithField("path", uri).
					WithField("impl", endpoint.Implementation).
					WithField("methods", []string{fasthttp.MethodGet, fasthttp.MethodHead}).
					Trace("Registering Authz Endpoint")

				r.HEAD(uri, handler)
				r.GET(uri, handler)
			}
		}
	}

	r.POST("/api/checks/safe-redirection", middlewareAPI(handlers.CheckSafeRedirectionPOST))

	r.POST("/api/firstfactor", middlewareAPI(handlers.FirstFactorPOST))
	r.POST("/api/logout", middlewareAPI(handlers.LogoutPOST))

	// Only register endpoints if forgot password is not disabled.
	if !config.AuthenticationBackend.PasswordReset.Disable &&
		config.AuthenticationBackend.PasswordReset.CustomURL.String() == "" {
		// Password reset related endpoints.
		r.POST("/api/reset-password/identity/start", middlewareAPI(handlers.ResetPasswordIdentityStart))
		r.POST("/api/reset-password/identity/finish", middlewareAPI(handlers.ResetPasswordIdentityFinish))
		r.POST("/api/reset-password", middlewareAPI(handlers.ResetPasswordPOST))
	}

	// Information about the user.
	r.GET("/api/user/info", middleware1FA(handlers.UserInfoGET))
	r.POST("/api/user/info", middleware1FA(handlers.UserInfoPOST))
	r.POST("/api/user/info/2fa_method", middleware1FA(handlers.MethodPreferencePOST))

	if !config.TOTP.Disable {
		// TOTP related endpoints.
		r.GET("/api/user/info/totp", middleware1FA(handlers.UserTOTPInfoGET))
		r.POST("/api/secondfactor/totp/identity/start", middleware1FA(handlers.TOTPIdentityStart))
		r.POST("/api/secondfactor/totp/identity/finish", middleware1FA(handlers.TOTPIdentityFinish))
		r.POST("/api/secondfactor/totp", middleware1FA(handlers.TimeBasedOneTimePasswordPOST))
	}

	if !config.Webauthn.Disable {
		// Webauthn Endpoints.
		r.POST("/api/secondfactor/webauthn/identity/start", middleware1FA(handlers.WebauthnIdentityStart))
		r.POST("/api/secondfactor/webauthn/identity/finish", middleware1FA(handlers.WebauthnIdentityFinish))
		r.POST("/api/secondfactor/webauthn/attestation", middleware1FA(handlers.WebauthnAttestationPOST))

		r.GET("/api/secondfactor/webauthn/assertion", middleware1FA(handlers.WebauthnAssertionGET))
		r.POST("/api/secondfactor/webauthn/assertion", middleware1FA(handlers.WebauthnAssertionPOST))
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
			middlewares.SecurityHeaders, middlewares.SecurityHeadersCSPNoneOpenIDConnect, middlewares.SecurityHeadersNoStore,
		).Build()

		r.GET("/api/oidc/consent", bridgeOIDC(handlers.OpenIDConnectConsentGET))
		r.POST("/api/oidc/consent", bridgeOIDC(handlers.OpenIDConnectConsentPOST))

		allowedOrigins := utils.StringSliceFromURLs(config.IdentityProviders.OIDC.CORS.AllowedOrigins)

		r.OPTIONS(oidc.EndpointPathWellKnownOpenIDConfiguration, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.EndpointPathWellKnownOpenIDConfiguration, policyCORSPublicGET.Middleware(bridgeOIDC(handlers.OpenIDConnectConfigurationWellKnownGET)))

		r.OPTIONS(oidc.EndpointPathWellKnownOAuthAuthorizationServer, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.EndpointPathWellKnownOAuthAuthorizationServer, policyCORSPublicGET.Middleware(bridgeOIDC(handlers.OAuthAuthorizationServerWellKnownGET)))

		r.OPTIONS(oidc.EndpointPathJWKs, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.EndpointPathJWKs, policyCORSPublicGET.Middleware(middlewareAPI(handlers.JSONWebKeySetGET)))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/jwks", policyCORSPublicGET.HandleOPTIONS)
		r.GET("/api/oidc/jwks", policyCORSPublicGET.Middleware(bridgeOIDC(handlers.JSONWebKeySetGET)))

		policyCORSAuthorization := middlewares.NewCORSPolicyBuilder().
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodGet, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.EndpointAuthorization, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.EndpointPathAuthorization, policyCORSAuthorization.HandleOnlyOPTIONS)
		r.GET(oidc.EndpointPathAuthorization, policyCORSAuthorization.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectAuthorization))))
		r.POST(oidc.EndpointPathAuthorization, policyCORSAuthorization.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectAuthorization))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy endpoint.
		r.OPTIONS("/api/oidc/authorize", policyCORSAuthorization.HandleOnlyOPTIONS)
		r.GET("/api/oidc/authorize", policyCORSAuthorization.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectAuthorization))))
		r.POST("/api/oidc/authorize", policyCORSAuthorization.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectAuthorization))))

		policyCORSToken := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.EndpointToken, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.EndpointPathToken, policyCORSToken.HandleOPTIONS)
		r.POST(oidc.EndpointPathToken, policyCORSToken.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectTokenPOST))))

		policyCORSUserinfo := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodGet, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.EndpointUserinfo, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.EndpointPathUserinfo, policyCORSUserinfo.HandleOPTIONS)
		r.GET(oidc.EndpointPathUserinfo, policyCORSUserinfo.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectUserinfo))))
		r.POST(oidc.EndpointPathUserinfo, policyCORSUserinfo.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectUserinfo))))

		policyCORSIntrospection := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.EndpointIntrospection, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.EndpointPathIntrospection, policyCORSIntrospection.HandleOPTIONS)
		r.POST(oidc.EndpointPathIntrospection, policyCORSIntrospection.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthIntrospectionPOST))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/introspect", policyCORSIntrospection.HandleOPTIONS)
		r.POST("/api/oidc/introspect", policyCORSIntrospection.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthIntrospectionPOST))))

		policyCORSRevocation := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodPost).
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.EndpointRevocation, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.EndpointPathRevocation, policyCORSRevocation.HandleOPTIONS)
		r.POST(oidc.EndpointPathRevocation, policyCORSRevocation.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthRevocationPOST))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/revoke", policyCORSRevocation.HandleOPTIONS)
		r.POST("/api/oidc/revoke", policyCORSRevocation.Middleware(bridgeOIDC(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthRevocationPOST))))
	}

	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = handleMethodNotAllowed
	r.NotFound = handleNotFound(bridge(serveIndexHandler))

	handler := middlewares.LogRequest(r.Handler)
	if config.Server.Path != "" {
		handler = middlewares.StripPath(config.Server.Path)(handler)
	}

	handler = middlewares.Wrap(middlewares.NewMetricsRequest(providers.Metrics), handler)

	return handler
}

func handleMetrics() fasthttp.RequestHandler {
	r := router.New()

	r.GET("/metrics", fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))

	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = handlers.Status(fasthttp.StatusMethodNotAllowed)
	r.NotFound = handlers.Status(fasthttp.StatusNotFound)

	return r.Handler
}
