package server

import (
	"net"
	"os"
	"strconv"

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
	rememberMe := strconv.FormatBool(configuration.Session.RememberMeDuration != "0")
	resetPassword := strconv.FormatBool(!configuration.AuthenticationBackend.DisableResetPassword)

	rootFiles := []string{"favicon.ico", "manifest.json", "robots.txt"}

	serveIndexHandler := ServeIndex(embeddedAssets, configuration.Server.Path, rememberMe, resetPassword)

	r := router.New()
	r.GET("/", serveIndexHandler)

	for _, f := range rootFiles {
		r.GET("/"+f, fasthttpadaptor.NewFastHTTPHandler(br.Serve(embeddedAssets)))
	}

	r.GET("/static/{filepath:*}", fasthttpadaptor.NewFastHTTPHandler(br.Serve(embeddedAssets)))

	r.GET("/api/state", autheliaMiddleware(handlers.StateGet))

	r.GET("/api/configuration", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.ConfigurationGet)))

	r.GET("/api/verify", autheliaMiddleware(handlers.VerifyGet(configuration.AuthenticationBackend)))
	r.HEAD("/api/verify", autheliaMiddleware(handlers.VerifyGet(configuration.AuthenticationBackend)))

	r.POST("/api/firstfactor", autheliaMiddleware(handlers.FirstFactorPost(1000, true)))
	r.POST("/api/logout", autheliaMiddleware(handlers.LogoutPost))

	// Only register endpoints if forgot password is not disabled.
	if !configuration.AuthenticationBackend.DisableResetPassword {
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
		middlewares.RequireFirstFactor(handlers.UserInfoGet)))
	r.POST("/api/user/info/2fa_method", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.MethodPreferencePost)))

	// TOTP related endpoints.
	r.POST("/api/secondfactor/totp/identity/start", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityStart)))
	r.POST("/api/secondfactor/totp/identity/finish", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityFinish)))
	r.POST("/api/secondfactor/totp", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPPost(&handlers.TOTPVerifierImpl{
			Period: uint(configuration.TOTP.Period),
			Skew:   uint(*configuration.TOTP.Skew),
		}))))

	// U2F related endpoints.
	r.POST("/api/secondfactor/u2f/identity/start", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FIdentityStart)))
	r.POST("/api/secondfactor/u2f/identity/finish", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FIdentityFinish)))

	r.POST("/api/secondfactor/u2f/register", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FRegister)))

	r.POST("/api/secondfactor/u2f/sign_request", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorU2FSignGet)))

	r.POST("/api/secondfactor/u2f/sign", autheliaMiddleware(
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

		r.POST("/api/secondfactor/duo", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoPost(duoAPI))))
	}

	// If trace is set, enable pprofhandler and expvarhandler.
	if configuration.LogLevel == "trace" {
		r.GET("/debug/pprof/{name?}", pprofhandler.PprofHandler)
		r.GET("/debug/vars", expvarhandler.ExpvarHandler)
	}

	r.NotFound = serveIndexHandler

	handler := middlewares.LogRequestMiddleware(r.Handler)
	if configuration.Server.Path != "" {
		handler = middlewares.StripPathMiddleware(handler)
	}

	server := &fasthttp.Server{
		ErrorHandler:          autheliaErrorHandler,
		Handler:               handler,
		NoDefaultServerHeader: true,
		ReadBufferSize:        configuration.Server.ReadBufferSize,
		WriteBufferSize:       configuration.Server.WriteBufferSize,
	}

	addrPattern := net.JoinHostPort(configuration.Host, strconv.Itoa(configuration.Port))

	listener, err := net.Listen("tcp", addrPattern)
	if err != nil {
		logging.Logger().Fatalf("Error initializing listener: %s", err)
	}

	if configuration.TLSCert != "" && configuration.TLSKey != "" {
		logging.Logger().Infof("Authelia is listening for TLS connections on %s%s", addrPattern, configuration.Server.Path)
		logging.Logger().Fatal(server.ServeTLS(listener, configuration.TLSCert, configuration.TLSKey))
	} else {
		logging.Logger().Infof("Authelia is listening for non-TLS connections on %s%s", addrPattern, configuration.Server.Path)
		logging.Logger().Fatal(server.Serve(listener))
	}
}
