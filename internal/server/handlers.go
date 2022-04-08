package server

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/expvarhandler"
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
func handlerError() func(ctx *fasthttp.RequestCtx, err error) {
	logger := logging.Logger()

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
		switch e := err.(type) {
		case *fasthttp.ErrSmallBuffer:
			logger.Tracef("Request was too large to handle from client %s. Response Code %d.", getRemoteIP(ctx), fasthttp.StatusRequestHeaderFieldsTooLarge)
			ctx.Error("request header too large", fasthttp.StatusRequestHeaderFieldsTooLarge)
		case *net.OpError:
			if e.Timeout() {
				logger.Tracef("Request timeout occurred while handling from client %s: %s. Response Code %d.", getRemoteIP(ctx), ctx.RequestURI(), fasthttp.StatusRequestTimeout)
				ctx.Error("request timeout", fasthttp.StatusRequestTimeout)
			} else {
				logger.Tracef("An unknown error occurred while handling a request from client %s: %s. Response Code %d.", getRemoteIP(ctx), ctx.RequestURI(), fasthttp.StatusBadRequest)
				ctx.Error("error when parsing request", fasthttp.StatusBadRequest)
			}
		default:
			logger.Tracef("An unknown error occurred while handling a request from client %s: %s. Response Code %d.", getRemoteIP(ctx), ctx.RequestURI(), fasthttp.StatusBadRequest)
			ctx.Error("error when parsing request", fasthttp.StatusBadRequest)
		}
	}
}

func handlerNotFound(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := strings.ToLower(string(ctx.Path()))

		for i := 0; i < len(httpServerDirs); i++ {
			if path == httpServerDirs[i].name || strings.HasPrefix(path, httpServerDirs[i].prefix) {
				handlers.SetStatusCodeResponse(ctx, fasthttp.StatusNotFound)

				return
			}
		}

		next(ctx)
	}
}

func handlerMethodNotAllowed(ctx *fasthttp.RequestCtx) {
	handlers.SetStatusCodeResponse(ctx, fasthttp.StatusMethodNotAllowed)
}

func getHandler(config schema.Configuration, providers middlewares.Providers) fasthttp.RequestHandler {
	rememberMe := strconv.FormatBool(config.Session.RememberMeDuration != schema.RememberMeDisabled)
	resetPassword := strconv.FormatBool(!config.AuthenticationBackend.DisableResetPassword)

	resetPasswordCustomURL := config.AuthenticationBackend.PasswordReset.CustomURL.String()

	duoSelfEnrollment := f
	if config.DuoAPI != nil {
		duoSelfEnrollment = strconv.FormatBool(config.DuoAPI.EnableSelfEnrollment)
	}

	https := config.Server.TLS.Key != "" && config.Server.TLS.Certificate != ""

	serveIndexHandler := ServeTemplatedFile(embeddedAssets, indexFile, config.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, resetPasswordCustomURL, config.Session.Name, config.Theme, https)
	serveSwaggerHandler := ServeTemplatedFile(swaggerAssets, indexFile, config.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, resetPasswordCustomURL, config.Session.Name, config.Theme, https)
	serveSwaggerAPIHandler := ServeTemplatedFile(swaggerAssets, apiFile, config.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, resetPasswordCustomURL, config.Session.Name, config.Theme, https)

	handlerPublicHTML := newPublicHTMLEmbeddedHandler()
	handlerLocales := newLocalesEmbeddedHandler()

	autheliaMiddleware := middlewares.AutheliaMiddleware(config, providers)

	policyCORSPublicGET := middlewares.NewCORSPolicyBuilder().
		WithAllowedMethods("OPTIONS", "GET").
		WithAllowedOrigins("*").
		Build()

	r := router.New()

	// Static Assets.
	r.GET("/", autheliaMiddleware(serveIndexHandler))

	for _, f := range rootFiles {
		r.GET("/"+f, handlerPublicHTML)
	}

	r.GET("/favicon.ico", middlewares.AssetOverrideMiddleware(config.Server.AssetPath, 0, handlerPublicHTML))
	r.GET("/static/media/logo.png", middlewares.AssetOverrideMiddleware(config.Server.AssetPath, 2, handlerPublicHTML))
	r.GET("/static/{filepath:*}", handlerPublicHTML)

	// Locales.
	r.GET("/locales/{language:[a-z]{1,3}}-{variant:[a-z0-9-]+}/{namespace:[a-z]+}.json", middlewares.AssetOverrideMiddleware(config.Server.AssetPath, 0, handlerLocales))
	r.GET("/locales/{language:[a-z]{1,3}}/{namespace:[a-z]+}.json", middlewares.AssetOverrideMiddleware(config.Server.AssetPath, 0, handlerLocales))

	// Swagger.
	r.GET("/api/", autheliaMiddleware(serveSwaggerHandler))
	r.OPTIONS("/api/", policyCORSPublicGET.HandleOPTIONS)
	r.GET("/api/"+apiFile, policyCORSPublicGET.Middleware(autheliaMiddleware(serveSwaggerAPIHandler)))
	r.OPTIONS("/api/"+apiFile, policyCORSPublicGET.HandleOPTIONS)

	for _, file := range swaggerFiles {
		r.GET("/api/"+file, handlerPublicHTML)
	}

	r.GET("/api/health", autheliaMiddleware(handlers.HealthGet))
	r.GET("/api/state", autheliaMiddleware(handlers.StateGet))

	r.GET("/api/configuration", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.ConfigurationGet)))

	r.GET("/api/configuration/password-policy", autheliaMiddleware(handlers.PasswordPolicyConfigurationGet))

	r.GET("/api/verify", autheliaMiddleware(handlers.VerifyGet(config.AuthenticationBackend)))
	r.HEAD("/api/verify", autheliaMiddleware(handlers.VerifyGet(config.AuthenticationBackend)))

	r.POST("/api/checks/safe-redirection", autheliaMiddleware(handlers.CheckSafeRedirection))

	r.POST("/api/firstfactor", autheliaMiddleware(handlers.FirstFactorPost(middlewares.TimingAttackDelay(10, 250, 85, time.Second))))
	r.POST("/api/logout", autheliaMiddleware(handlers.LogoutPost))

	// Only register endpoints if forgot password is not disabled.
	if !config.AuthenticationBackend.DisableResetPassword &&
		config.AuthenticationBackend.PasswordReset.CustomURL.String() == "" {
		// Password reset related endpoints.
		r.POST("/api/reset-password/identity/start", autheliaMiddleware(
			handlers.ResetPasswordIdentityStart))
		r.POST("/api/reset-password/identity/finish", autheliaMiddleware(
			handlers.ResetPasswordIdentityFinish))
		r.POST("/api/reset-password", autheliaMiddleware(
			handlers.ResetPasswordPost))
	}

	// Information about the user.
	r.GET("/api/user/info", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.UserInfoGET)))
	r.POST("/api/user/info", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.UserInfoPOST)))
	r.POST("/api/user/info/2fa_method", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.MethodPreferencePost)))

	if !config.TOTP.Disable {
		// TOTP related endpoints.
		r.GET("/api/user/info/totp", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.UserTOTPGet)))

		r.POST("/api/secondfactor/totp/identity/start", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityStart)))
		r.POST("/api/secondfactor/totp/identity/finish", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityFinish)))
		r.POST("/api/secondfactor/totp", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorTOTPPost)))
	}

	if !config.Webauthn.Disable {
		// Webauthn Endpoints.
		r.POST("/api/secondfactor/webauthn/identity/start", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorWebauthnIdentityStart)))
		r.POST("/api/secondfactor/webauthn/identity/finish", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorWebauthnIdentityFinish)))
		r.POST("/api/secondfactor/webauthn/attestation", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorWebauthnAttestationPOST)))

		r.GET("/api/secondfactor/webauthn/assertion", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorWebauthnAssertionGET)))
		r.POST("/api/secondfactor/webauthn/assertion", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorWebauthnAssertionPOST)))
	}

	// Configure DUO api endpoint only if configuration exists.
	if config.DuoAPI != nil {
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

		r.GET("/api/secondfactor/duo_devices", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoDevicesGet(duoAPI))))

		r.POST("/api/secondfactor/duo", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoPost(duoAPI))))

		r.POST("/api/secondfactor/duo_device", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoDevicePost)))
	}

	if config.Server.EnablePprof {
		r.GET("/debug/pprof/{name?}", pprofhandler.PprofHandler)
	}

	if config.Server.EnableExpvars {
		r.GET("/debug/vars", expvarhandler.ExpvarHandler)
	}

	if providers.OpenIDConnect.Fosite != nil {
		r.GET("/api/oidc/consent", autheliaMiddleware(handlers.OpenIDConnectConsentGET))
		r.POST("/api/oidc/consent", autheliaMiddleware(handlers.OpenIDConnectConsentPOST))

		allowedOrigins := utils.StringSliceFromURLs(config.IdentityProviders.OIDC.CORS.AllowedOrigins)

		r.OPTIONS(oidc.WellKnownOpenIDConfigurationPath, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.WellKnownOpenIDConfigurationPath, policyCORSPublicGET.Middleware(autheliaMiddleware(handlers.OpenIDConnectConfigurationWellKnownGET)))

		r.OPTIONS(oidc.WellKnownOAuthAuthorizationServerPath, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.WellKnownOAuthAuthorizationServerPath, policyCORSPublicGET.Middleware(autheliaMiddleware(handlers.OAuthAuthorizationServerWellKnownGET)))

		r.OPTIONS(oidc.JWKsPath, policyCORSPublicGET.HandleOPTIONS)
		r.GET(oidc.JWKsPath, policyCORSPublicGET.Middleware(autheliaMiddleware(handlers.JSONWebKeySetGET)))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/jwks", policyCORSPublicGET.HandleOPTIONS)
		r.GET("/api/oidc/jwks", policyCORSPublicGET.Middleware(autheliaMiddleware(handlers.JSONWebKeySetGET)))

		policyCORSAuthorization := middlewares.NewCORSPolicyBuilder().
			WithAllowedMethods("OPTIONS", "GET").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.AuthorizationEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.AuthorizationPath, policyCORSAuthorization.HandleOnlyOPTIONS)
		r.GET(oidc.AuthorizationPath, autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectAuthorizationGET)))

		// TODO (james-d-elliott): Remove in GA. This is a legacy endpoint.
		r.OPTIONS("/api/oidc/authorize", policyCORSAuthorization.HandleOnlyOPTIONS)
		r.GET("/api/oidc/authorize", autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectAuthorizationGET)))

		policyCORSToken := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.TokenEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.TokenPath, policyCORSToken.HandleOPTIONS)
		r.POST(oidc.TokenPath, policyCORSToken.Middleware(autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectTokenPOST))))

		policyCORSUserinfo := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "GET", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.UserinfoEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.UserinfoPath, policyCORSUserinfo.HandleOPTIONS)
		r.GET(oidc.UserinfoPath, policyCORSUserinfo.Middleware(autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectUserinfo))))
		r.POST(oidc.UserinfoPath, policyCORSUserinfo.Middleware(autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectUserinfo))))

		policyCORSIntrospection := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.IntrospectionEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.IntrospectionPath, policyCORSIntrospection.HandleOPTIONS)
		r.POST(oidc.IntrospectionPath, policyCORSIntrospection.Middleware(autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthIntrospectionPOST))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/introspect", policyCORSIntrospection.HandleOPTIONS)
		r.POST("/api/oidc/introspect", policyCORSIntrospection.Middleware(autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthIntrospectionPOST))))

		policyCORSRevocation := middlewares.NewCORSPolicyBuilder().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.RevocationEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints)).
			Build()

		r.OPTIONS(oidc.RevocationPath, policyCORSRevocation.HandleOPTIONS)
		r.POST(oidc.RevocationPath, policyCORSRevocation.Middleware(autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthRevocationPOST))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/revoke", policyCORSRevocation.HandleOPTIONS)
		r.POST("/api/oidc/revoke", policyCORSRevocation.Middleware(autheliaMiddleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthRevocationPOST))))
	}

	r.NotFound = handlerNotFound(autheliaMiddleware(serveIndexHandler))

	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = handlerMethodNotAllowed

	handler := middlewares.LogRequestMiddleware(r.Handler)
	if config.Server.Path != "" {
		handler = middlewares.StripPathMiddleware(config.Server.Path, handler)
	}

	return handler
}
