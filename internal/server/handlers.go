package server

import (
	"net"
	"os"
	"strconv"
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
	"github.com/authelia/authelia/v4/internal/handler"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middleware"
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
			if e.Timeout() {
				statusCode = fasthttp.StatusRequestTimeout
				message = "Request timeout occurred while handling request from client."
			} else {
				statusCode = fasthttp.StatusBadRequest
				message = "An unknown network error occurred while handling a request from client."
			}
		default:
			statusCode = fasthttp.StatusBadRequest
			message = "An unknown error occurred while handling a request from client."
		}

		logging.Logger().WithFields(logrus.Fields{
			"method":      string(ctx.Method()),
			"path":        string(ctx.Path()),
			"remote_ip":   getRemoteIP(ctx),
			"status_code": statusCode,
		}).WithError(err).Error(message)

		handler.SetStatusCodeResponse(ctx, statusCode)
	}
}

func handleNotFound(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := strings.ToLower(string(ctx.Path()))

		for i := 0; i < len(httpServerDirs); i++ {
			if path == httpServerDirs[i].name || strings.HasPrefix(path, httpServerDirs[i].prefix) {
				handler.SetStatusCodeResponse(ctx, fasthttp.StatusNotFound)

				return
			}
		}

		next(ctx)
	}
}

func handleRouter(config schema.Configuration, providers middleware.Providers) fasthttp.RequestHandler {
	rememberMe := strconv.FormatBool(config.Session.RememberMeDuration != schema.RememberMeDisabled)
	resetPassword := strconv.FormatBool(!config.AuthenticationBackend.PasswordReset.Disable)

	resetPasswordCustomURL := config.AuthenticationBackend.PasswordReset.CustomURL.String()

	duoSelfEnrollment := f
	if !config.DuoAPI.Disable {
		duoSelfEnrollment = strconv.FormatBool(config.DuoAPI.EnableSelfEnrollment)
	}

	https := config.Server.TLS.Key != "" && config.Server.TLS.Certificate != ""

	serveIndexHandler := ServeTemplatedFile(embeddedAssets, indexFile, config.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, resetPasswordCustomURL, config.Session.Name, config.Theme, https)
	serveSwaggerHandler := ServeTemplatedFile(swaggerAssets, indexFile, config.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, resetPasswordCustomURL, config.Session.Name, config.Theme, https)
	serveSwaggerAPIHandler := ServeTemplatedFile(swaggerAssets, apiFile, config.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, resetPasswordCustomURL, config.Session.Name, config.Theme, https)

	handlerPublicHTML := newPublicHTMLEmbeddedHandler()
	handlerLocales := newLocalesEmbeddedHandler()

	middlewarePublic := middleware.NewBridgeBuilder(config, providers).
		WithPreMiddlewares(middleware.SecurityHeaders).Build()

	policyCORSPublicGET := middleware.NewCORSPolicyBuilder().
		WithAllowedMethods("OPTIONS", "GET").
		WithAllowedOrigins("*").
		Build()

	r := router.New()

	// Static Assets.
	r.GET("/", middlewarePublic(serveIndexHandler))

	for _, f := range rootFiles {
		r.GET("/"+f, handlerPublicHTML)
	}

	r.GET("/favicon.ico", middleware.AssetOverride(config.Server.AssetPath, 0, handlerPublicHTML))
	r.GET("/static/media/logo.png", middleware.AssetOverride(config.Server.AssetPath, 2, handlerPublicHTML))
	r.GET("/static/{filepath:*}", handlerPublicHTML)

	// Locales.
	r.GET("/locales/{language:[a-z]{1,3}}-{variant:[a-zA-Z0-9-]+}/{namespace:[a-z]+}.json", middleware.AssetOverride(config.Server.AssetPath, 0, handlerLocales))
	r.GET("/locales/{language:[a-z]{1,3}}/{namespace:[a-z]+}.json", middleware.AssetOverride(config.Server.AssetPath, 0, handlerLocales))

	// Swagger.
	r.GET("/api/", middlewarePublic(serveSwaggerHandler))
	r.OPTIONS("/api/", policyCORSPublicGET.HandleOPTIONS)
	r.GET("/api/"+apiFile, policyCORSPublicGET.Middleware(middlewarePublic(serveSwaggerAPIHandler)))
	r.OPTIONS("/api/"+apiFile, policyCORSPublicGET.HandleOPTIONS)

	for _, file := range swaggerFiles {
		r.GET("/api/"+file, handlerPublicHTML)
	}

	middlewareAPI := middleware.NewBridgeBuilder(config, providers).
		WithPreMiddlewares(middleware.SecurityHeaders, middleware.SecurityHeadersNoStore, middleware.SecurityHeadersCSPNone).
		Build()

	middleware1FA := middleware.NewBridgeBuilder(config, providers).
		WithPreMiddlewares(middleware.SecurityHeaders, middleware.SecurityHeadersNoStore, middleware.SecurityHeadersCSPNone).
		WithPostMiddlewares(middleware.Require1FA).
		Build()

	r.GET("/api/health", middlewareAPI(handler.HealthGET))
	r.GET("/api/state", middlewareAPI(handler.StateGET))

	r.GET("/api/configuration", middleware1FA(handler.ConfigurationGET))

	r.GET("/api/configuration/password-policy", middlewareAPI(handler.PasswordPolicyConfigurationGET))

	metricsVRMW := middleware.NewMetricsVerifyRequest(providers.Metrics)

	r.GET("/api/verify", middleware.Wrap(metricsVRMW, middlewarePublic(handler.VerifyGET(config.AuthenticationBackend))))
	r.HEAD("/api/verify", middleware.Wrap(metricsVRMW, middlewarePublic(handler.VerifyGET(config.AuthenticationBackend))))

	r.GET("/api/verify/nginx", middleware.Wrap(metricsVRMW, middlewarePublic(handler.VerifyGETProxyNGINX(config.AuthenticationBackend))))
	r.HEAD("/api/verify/nginx", middleware.Wrap(metricsVRMW, middlewarePublic(handler.VerifyGETProxyNGINX(config.AuthenticationBackend))))

	r.POST("/api/checks/safe-redirection", middlewareAPI(handler.CheckSafeRedirectionPOST))

	delayFunc := middleware.TimingAttackDelay(10, 250, 85, time.Second, true)

	r.POST("/api/firstfactor", middlewareAPI(handler.FirstFactorPOST(delayFunc)))
	r.POST("/api/logout", middlewareAPI(handler.LogoutPOST))

	// Only register endpoints if forgot password is not disabled.
	if !config.AuthenticationBackend.PasswordReset.Disable &&
		config.AuthenticationBackend.PasswordReset.CustomURL.String() == "" {
		// Password reset related endpoints.
		r.POST("/api/reset-password/identity/start", middlewareAPI(handler.ResetPasswordIdentityStart))
		r.POST("/api/reset-password/identity/finish", middlewareAPI(handler.ResetPasswordIdentityFinish))
		r.POST("/api/reset-password", middlewareAPI(handler.ResetPasswordPOST))
	}

	// Information about the user.
	r.GET("/api/user/info", middleware1FA(handler.UserInfoGET))
	r.POST("/api/user/info", middleware1FA(handler.UserInfoPOST))
	r.POST("/api/user/info/2fa_method", middleware1FA(handler.MethodPreferencePOST))

	if !config.TOTP.Disable {
		// TOTP related endpoints.
		r.GET("/api/user/info/totp", middleware1FA(handler.UserTOTPInfoGET))
		r.POST("/api/secondfactor/totp/identity/start", middleware1FA(handler.TOTPIdentityStart))
		r.POST("/api/secondfactor/totp/identity/finish", middleware1FA(handler.TOTPIdentityFinish))
		r.POST("/api/secondfactor/totp", middleware1FA(handler.TimeBasedOneTimePasswordPOST))
	}

	if !config.Webauthn.Disable {
		// Webauthn Endpoints.
		r.POST("/api/secondfactor/webauthn/identity/start", middleware1FA(handler.WebauthnIdentityStart))
		r.POST("/api/secondfactor/webauthn/identity/finish", middleware1FA(handler.WebauthnIdentityFinish))
		r.POST("/api/secondfactor/webauthn/attestation", middleware1FA(handler.WebauthnAttestationPOST))

		r.GET("/api/secondfactor/webauthn/assertion", middleware1FA(handler.WebauthnAssertionGET))
		r.POST("/api/secondfactor/webauthn/assertion", middleware1FA(handler.WebauthnAssertionPOST))
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

		r.GET("/api/secondfactor/duo_devices", middleware1FA(handler.DuoDevicesGET(duoAPI)))
		r.POST("/api/secondfactor/duo", middleware1FA(handler.DuoPOST(duoAPI)))
		r.POST("/api/secondfactor/duo_device", middleware1FA(handler.DuoDevicePOST))
	}

	if config.Server.EnablePprof {
		r.GET("/debug/pprof/{name?}", pprofhandler.PprofHandler)
	}

	if config.Server.EnableExpvars {
		r.GET("/debug/vars", expvarhandler.ExpvarHandler)
	}

	if providers.OpenIDConnect.Fosite != nil {
		middlewareOIDC := middleware.NewBridgeBuilder(config, providers).WithPreMiddlewares(
			middleware.SecurityHeaders, middleware.SecurityHeadersCSPNone, middleware.SecurityHeadersNoStore,
		).Build()

		r.GET("/api/oidc/consent", middlewareOIDC(handler.OpenIDConnectConsentGET))
		r.POST("/api/oidc/consent", middlewareOIDC(handler.OpenIDConnectConsentPOST))

		allowedOrigins := utils.StringSliceFromURLs(config.IdentityProviders.OIDC.CORS.AllowedOrigins)

		r.OPTIONS(oidc.WellKnownOpenIDConfigurationPath, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.WellKnownOpenIDConfigurationPath, policyCORSPublicGET.Middleware(middlewareOIDC(handler.OpenIDConnectConfigurationWellKnownGET)))

		r.OPTIONS(oidc.WellKnownOAuthAuthorizationServerPath, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.WellKnownOAuthAuthorizationServerPath, policyCORSPublicGET.Middleware(middlewareOIDC(handler.OAuthAuthorizationServerWellKnownGET)))

		r.OPTIONS(oidc.JWKsPath, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.JWKsPath, policyCORSPublicGET.Middleware(middlewareAPI(handler.JSONWebKeySetGET)))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/jwks", policyCORSPublicGET.HandleOPTIONS)
		r.GET("/api/oidc/jwks", policyCORSPublicGET.Middleware(middlewareOIDC(handler.JSONWebKeySetGET)))

		policyCORSAuthorization := middleware.NewCORSPolicyBuilder().
			WithAllowedMethods("OPTIONS", "GET").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.AuthorizationEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.AuthorizationPath, policyCORSAuthorization.HandleOnlyOPTIONS)
		r.GET(oidc.AuthorizationPath, middlewareOIDC(middleware.NewHTTPToAutheliaHandlerAdaptor(handler.OpenIDConnectAuthorizationGET)))

		// TODO (james-d-elliott): Remove in GA. This is a legacy endpoint.
		r.OPTIONS("/api/oidc/authorize", policyCORSAuthorization.HandleOnlyOPTIONS)
		r.GET("/api/oidc/authorize", middlewareOIDC(middleware.NewHTTPToAutheliaHandlerAdaptor(handler.OpenIDConnectAuthorizationGET)))

		policyCORSToken := middleware.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.TokenEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.TokenPath, policyCORSToken.HandleOPTIONS)
		r.POST(oidc.TokenPath, policyCORSToken.Middleware(middlewareOIDC(middleware.NewHTTPToAutheliaHandlerAdaptor(handler.OpenIDConnectTokenPOST))))

		policyCORSUserinfo := middleware.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "GET", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.UserinfoEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.UserinfoPath, policyCORSUserinfo.HandleOPTIONS)
		r.GET(oidc.UserinfoPath, policyCORSUserinfo.Middleware(middlewareOIDC(middleware.NewHTTPToAutheliaHandlerAdaptor(handler.OpenIDConnectUserinfo))))
		r.POST(oidc.UserinfoPath, policyCORSUserinfo.Middleware(middlewareOIDC(middleware.NewHTTPToAutheliaHandlerAdaptor(handler.OpenIDConnectUserinfo))))

		policyCORSIntrospection := middleware.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.IntrospectionEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.IntrospectionPath, policyCORSIntrospection.HandleOPTIONS)
		r.POST(oidc.IntrospectionPath, policyCORSIntrospection.Middleware(middlewareOIDC(middleware.NewHTTPToAutheliaHandlerAdaptor(handler.OAuthIntrospectionPOST))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/introspect", policyCORSIntrospection.HandleOPTIONS)
		r.POST("/api/oidc/introspect", policyCORSIntrospection.Middleware(middlewareOIDC(middleware.NewHTTPToAutheliaHandlerAdaptor(handler.OAuthIntrospectionPOST))))

		policyCORSRevocation := middleware.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.RevocationEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.RevocationPath, policyCORSRevocation.HandleOPTIONS)
		r.POST(oidc.RevocationPath, policyCORSRevocation.Middleware(middlewareOIDC(middleware.NewHTTPToAutheliaHandlerAdaptor(handler.OAuthRevocationPOST))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/revoke", policyCORSRevocation.HandleOPTIONS)
		r.POST("/api/oidc/revoke", policyCORSRevocation.Middleware(middlewareOIDC(middleware.NewHTTPToAutheliaHandlerAdaptor(handler.OAuthRevocationPOST))))
	}

	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = handler.Status(fasthttp.StatusMethodNotAllowed)
	r.NotFound = handleNotFound(middlewarePublic(serveIndexHandler))

	handler := middleware.LogRequest(r.Handler)
	if config.Server.Path != "" {
		handler = middleware.StripPath(config.Server.Path)(handler)
	}

	handler = middleware.Wrap(middleware.NewMetricsRequest(providers.Metrics), handler)

	return handler
}

func handleMetrics() fasthttp.RequestHandler {
	r := router.New()

	r.GET("/metrics", fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))

	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = handler.Status(fasthttp.StatusMethodNotAllowed)
	r.NotFound = handler.Status(fasthttp.StatusNotFound)

	return r.Handler
}
