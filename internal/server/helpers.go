package server

import (
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"time"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/expvarhandler"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/valyala/fasthttp/pprofhandler"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

func getRequestHandler(config schema.Configuration, providers middlewares.Providers) fasthttp.RequestHandler {
	rememberMe := strconv.FormatBool(config.Session.RememberMeDuration != schema.RememberMeDisabled)
	resetPassword := strconv.FormatBool(!config.AuthenticationBackend.DisableResetPassword)

	duoSelfEnrollment := f
	if config.DuoAPI != nil {
		duoSelfEnrollment = strconv.FormatBool(config.DuoAPI.EnableSelfEnrollment)
	}

	https := config.Server.TLS.Key != "" && config.Server.TLS.Certificate != ""

	serveIndexHandler := ServeTemplatedFile(embeddedAssets, indexFile, config.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, config.Session.Name, config.Theme, https)
	serveSwaggerHandler := ServeTemplatedFile(swaggerAssets, indexFile, config.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, config.Session.Name, config.Theme, https)
	serveSwaggerAPIHandler := ServeTemplatedFile(swaggerAssets, apiFile, config.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, config.Session.Name, config.Theme, https)

	embeddedPath, _ := fs.Sub(assets, "public_html")
	embeddedFS := fasthttpadaptor.NewFastHTTPHandler(http.FileServer(http.FS(embeddedPath)))
	middleware := middlewares.AutheliaMiddleware(config, providers)

	r := router.New()
	r.GET("/", middleware(serveIndexHandler))
	r.OPTIONS("/", middleware(handleOPTIONS))

	for _, f := range rootFiles {
		r.GET("/"+f, middlewares.AssetOverrideMiddleware(config.Server.AssetPath, embeddedFS))
	}

	// Swagger.
	r.GET("/api/", middleware(serveSwaggerHandler))
	r.GET("/api/"+apiFile, middleware(serveSwaggerAPIHandler))

	for _, f := range swaggerFiles {
		r.GET("/api/"+f, embeddedFS)
	}

	//r.ANY("/api/{filepath:*}", embeddedFS)

	r.GET("/static/{filepath:*}", middlewares.AssetOverrideMiddleware(config.Server.AssetPath, embeddedFS))

	r.GET("/api/health", middleware(handlers.HealthGet))
	r.GET("/api/state", middleware(handlers.StateGet))

	r.GET("/api/configuration", middleware(middlewares.RequireFirstFactor(handlers.ConfigurationGet)))

	r.GET("/api/verify", middleware(handlers.VerifyGet(config.AuthenticationBackend)))
	r.HEAD("/api/verify", middleware(handlers.VerifyGet(config.AuthenticationBackend)))

	r.POST("/api/checks/safe-redirection", middleware(handlers.CheckSafeRedirection))

	r.POST("/api/firstfactor", middleware(handlers.FirstFactorPost(middlewares.TimingAttackDelay(10, 250, 85, time.Second))))
	r.POST("/api/logout", middleware(handlers.LogoutPost))

	// Only register endpoints if forgot password is not disabled.
	if !config.AuthenticationBackend.DisableResetPassword {
		// Password reset related endpoints.
		r.POST("/api/reset-password/identity/start", middleware(handlers.ResetPasswordIdentityStart))
		r.POST("/api/reset-password/identity/finish", middleware(handlers.ResetPasswordIdentityFinish))
		r.POST("/api/reset-password", middleware(handlers.ResetPasswordPost))
	}

	// Information about the user.
	r.GET("/api/user/info", middleware(middlewares.RequireFirstFactor(handlers.UserInfoGet)))
	r.POST("/api/user/info/2fa_method", middleware(middlewares.RequireFirstFactor(handlers.MethodPreferencePost)))

	if !config.TOTP.Disable {
		// TOTP related endpoints.
		r.GET("/api/user/info/totp", middleware(middlewares.RequireFirstFactor(handlers.UserTOTPGet)))

		r.POST("/api/secondfactor/totp/identity/start", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityStart)))
		r.POST("/api/secondfactor/totp/identity/finish", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityFinish)))
		r.POST("/api/secondfactor/totp", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorTOTPPost)))
	}

	if !config.Webauthn.Disable {
		// Webauthn Endpoints.
		r.POST("/api/secondfactor/webauthn/identity/start", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorWebauthnIdentityStart)))
		r.POST("/api/secondfactor/webauthn/identity/finish", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorWebauthnIdentityFinish)))
		r.POST("/api/secondfactor/webauthn/attestation", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorWebauthnAttestationPOST)))

		r.GET("/api/secondfactor/webauthn/assertion", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorWebauthnAssertionGET)))
		r.POST("/api/secondfactor/webauthn/assertion", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorWebauthnAssertionPOST)))
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

		r.GET("/api/secondfactor/duo_devices", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorDuoDevicesGet(duoAPI))))

		r.POST("/api/secondfactor/duo", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorDuoPost(duoAPI))))

		r.POST("/api/secondfactor/duo_device", middleware(middlewares.RequireFirstFactor(handlers.SecondFactorDuoDevicePost)))
	}

	if config.Server.EnablePprof {
		r.GET("/debug/pprof/{name?}", pprofhandler.PprofHandler)
	}

	if config.Server.EnableExpvars {
		r.GET("/debug/vars", expvarhandler.ExpvarHandler)
	}

	if providers.OpenIDConnect.Fosite != nil {
		r.GET("/api/oidc/consent", middleware(handlers.OpenIDConnectConsentGET))
		r.POST("/api/oidc/consent", middleware(handlers.OpenIDConnectConsentPOST))

		allowedOrigins := utils.StringSliceFromURLs(config.IdentityProviders.OIDC.CORS.AllowedOrigins)

		corsGETPublic := middlewares.NewCORSMiddleware().
			WithAllowedMethods("OPTIONS", "GET").
			WithAllowedOrigins("*")

		r.OPTIONS(oidc.WellKnownOpenIDConfigurationPath, corsGETPublic.HandleOPTIONS)
		r.GET(oidc.WellKnownOpenIDConfigurationPath, corsGETPublic.Middleware(middleware(handlers.OpenIDConnectConfigurationWellKnownGET)))

		r.OPTIONS(oidc.WellKnownOAuthAuthorizationServerPath, corsGETPublic.HandleOPTIONS)
		r.GET(oidc.WellKnownOAuthAuthorizationServerPath, corsGETPublic.Middleware(middleware(handlers.OAuthAuthorizationServerWellKnownGET)))

		r.OPTIONS(oidc.JWKsPath, corsGETPublic.HandleOPTIONS)
		r.GET(oidc.JWKsPath, corsGETPublic.Middleware(middleware(handlers.JSONWebKeySetGET)))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/jwks", corsGETPublic.HandleOPTIONS)
		r.GET("/api/oidc/jwks", corsGETPublic.Middleware(middleware(handlers.JSONWebKeySetGET)))

		corsAuthorization := middlewares.NewCORSMiddleware().
			WithAllowedMethods("OPTIONS", "GET").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.AuthorizationEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints))

		r.OPTIONS(oidc.AuthorizationPath, corsAuthorization.HandleOnlyOPTIONS)
		r.GET(oidc.AuthorizationPath, middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectAuthorizationGET)))

		// TODO (james-d-elliott): Remove in GA. This is a legacy endpoint.
		r.OPTIONS("/api/oidc/authorize", corsAuthorization.HandleOnlyOPTIONS)
		r.GET("/api/oidc/authorize", middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectAuthorizationGET)))

		corsToken := middlewares.NewCORSMiddleware().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.TokenEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints))

		r.OPTIONS(oidc.TokenPath, corsToken.HandleOPTIONS)
		r.POST(oidc.TokenPath, corsToken.Middleware(middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectTokenPOST))))

		corsUserInfo := middlewares.NewCORSMiddleware().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "GET", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.UserinfoEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints))

		r.OPTIONS(oidc.UserinfoPath, corsUserInfo.HandleOPTIONS)
		r.GET(oidc.UserinfoPath, corsUserInfo.Middleware(middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectUserinfo))))
		r.POST(oidc.UserinfoPath, corsUserInfo.Middleware(middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OpenIDConnectUserinfo))))

		corsIntrospection := middlewares.NewCORSMiddleware().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.IntrospectionEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints))

		r.OPTIONS(oidc.IntrospectionPath, corsIntrospection.HandleOPTIONS)
		r.POST(oidc.IntrospectionPath, corsIntrospection.Middleware(middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthIntrospectionPOST))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/introspect", corsIntrospection.HandleOPTIONS)
		r.POST("/api/oidc/introspect", corsIntrospection.Middleware(middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthIntrospectionPOST))))

		corsRevocation := middlewares.NewCORSMiddleware().
			WithAllowCredentials(true).
			WithAllowedMethods("OPTIONS", "POST").
			WithAllowedOrigins(allowedOrigins...).
			WithEnabled(utils.IsStringInSlice(oidc.RevocationEndpoint, config.IdentityProviders.OIDC.CORS.Endpoints))

		r.OPTIONS(oidc.RevocationPath, corsRevocation.HandleOPTIONS)
		r.POST(oidc.RevocationPath, corsRevocation.Middleware(middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthRevocationPOST))))

		// TODO (james-d-elliott): Remove in GA. This is a legacy implementation of the above endpoint.
		r.OPTIONS("/api/oidc/revoke", corsRevocation.HandleOPTIONS)
		r.POST("/api/oidc/revoke", corsRevocation.Middleware(middleware(middlewares.NewHTTPToAutheliaHandlerAdaptor(handlers.OAuthRevocationPOST))))
	}

	r.NotFound = handleNotFound(middleware(serveIndexHandler))

	handler := middlewares.LogRequestMiddleware(r.Handler)
	if config.Server.Path != "" {
		handler = middlewares.StripPathMiddleware(config.Server.Path, handler)
	}

	return handler
}
