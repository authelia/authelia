package server

import (
	"fmt"
	"os"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/expvarhandler"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/valyala/fasthttp/pprofhandler"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/duo"
	"github.com/authelia/authelia/internal/handlers"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/middlewares"
)

// StartServer start Authelia server with the given configuration and providers.
func StartServer(configuration schema.Configuration, providers middlewares.Providers) {
	autheliaMiddleware := middlewares.AutheliaMiddleware(configuration, providers)
	embeddedAssets := "/public_html"
	// TODO: Remove in v4.18.0.
	if os.Getenv("PUBLIC_DIR") != "" {
		logging.Logger().Warn("PUBLIC_DIR environment variable has been deprecated, assets are now embedded.")
	}

	router := router.New()

	router.GET("/", ServeIndex(embeddedAssets))
	router.GET("/static/{filepath:*}", fasthttpadaptor.NewFastHTTPHandler(br.Serve(embeddedAssets)))

	router.GET("/api/state", autheliaMiddleware(handlers.StateGet))

	router.GET("/api/configuration", autheliaMiddleware(handlers.ConfigurationGet))
	router.GET("/api/configuration/extended", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.ExtendedConfigurationGet)))

	router.GET("/api/verify", autheliaMiddleware(handlers.VerifyGet(configuration.AuthenticationBackend)))
	router.HEAD("/api/verify", autheliaMiddleware(handlers.VerifyGet(configuration.AuthenticationBackend)))

	router.POST("/api/firstfactor", autheliaMiddleware(handlers.FirstFactorPost))
	router.POST("/api/logout", autheliaMiddleware(handlers.LogoutPost))

	// Only register endpoints if forgot password is not disabled.
	if !configuration.AuthenticationBackend.DisableResetPassword {
		// Password reset related endpoints.
		router.POST("/api/reset-password/identity/start", autheliaMiddleware(
			handlers.ResetPasswordIdentityStart))
		router.POST("/api/reset-password/identity/finish", autheliaMiddleware(
			handlers.ResetPasswordIdentityFinish))
		router.POST("/api/reset-password", autheliaMiddleware(
			handlers.ResetPasswordPost))
	}

	// Information about the user.
	router.GET("/api/user/info", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.UserInfoGet)))
	router.POST("/api/user/info/2fa_method", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.MethodPreferencePost)))

	// TOTP related endpoints.
	router.POST("/api/secondfactor/totp/identity/start", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityStart)))
	router.POST("/api/secondfactor/totp/identity/finish", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityFinish)))
	router.POST("/api/secondfactor/totp", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPPost(&handlers.TOTPVerifierImpl{
			Period: uint(configuration.TOTP.Period),
			Skew:   uint(*configuration.TOTP.Skew),
		}))))

	// U2F related endpoints.
	router.POST("/api/secondfactor/u2f/identity/start", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FIdentityStart)))
	router.POST("/api/secondfactor/u2f/identity/finish", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FIdentityFinish)))

	router.POST("/api/secondfactor/u2f/register", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FRegister)))

	router.POST("/api/secondfactor/u2f/sign_request", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FSignGet)))

	router.POST("/api/secondfactor/u2f/sign", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FSignPost(&handlers.U2FVerifierImpl{}))))

	// Configure DUO api endpoint only if configuration exists.
	if configuration.DuoAPI != nil {
		var duoAPI duo.API
		if os.Getenv("ENVIRONMENT") == "dev" {
			duoAPI = duo.NewDuoAPI(duoapi.NewDuoApi(
				configuration.DuoAPI.IntegrationKey,
				configuration.DuoAPI.SecretKey,
				configuration.DuoAPI.Hostname, "", duoapi.SetInsecure()))
		} else {
			duoAPI = duo.NewDuoAPI(duoapi.NewDuoApi(
				configuration.DuoAPI.IntegrationKey,
				configuration.DuoAPI.SecretKey,
				configuration.DuoAPI.Hostname, ""))
		}

		router.POST("/api/secondfactor/duo", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoPost(duoAPI))))
	}

	// If trace is set, enable pprofhandler and expvarhandler.
	if configuration.LogLevel == "trace" {
		router.GET("/debug/pprof/{name?}", pprofhandler.PprofHandler)
		router.GET("/debug/vars", expvarhandler.ExpvarHandler)
	}

	router.NotFound = ServeIndex(embeddedAssets)

	server := &fasthttp.Server{
		ErrorHandler:          autheliaErrorHandler,
		Handler:               middlewares.LogRequestMiddleware(router.Handler),
		NoDefaultServerHeader: true,
		ReadBufferSize:        configuration.Server.ReadBufferSize,
		WriteBufferSize:       configuration.Server.WriteBufferSize,
	}

	addrPattern := fmt.Sprintf("%s:%d", configuration.Host, configuration.Port)

	if configuration.TLSCert != "" && configuration.TLSKey != "" {
		logging.Logger().Infof("Authelia is listening for TLS connections on %s", addrPattern)
		logging.Logger().Fatal(server.ListenAndServeTLS(addrPattern, configuration.TLSCert, configuration.TLSKey))
	} else {
		logging.Logger().Infof("Authelia is listening for non-TLS connections on %s", addrPattern)
		logging.Logger().Fatal(server.ListenAndServe(addrPattern))
	}
}
