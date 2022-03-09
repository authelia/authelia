package server

import (
	"embed"
	"io/fs"
	"net"
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
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

//go:embed public_html
var assets embed.FS

func registerRoutes(configuration schema.Configuration, providers middlewares.Providers) fasthttp.RequestHandler {
	autheliaMiddleware := middlewares.AutheliaMiddleware(configuration, providers)
	rememberMe := strconv.FormatBool(configuration.Session.RememberMeDuration != -1)
	resetPassword := strconv.FormatBool(!configuration.AuthenticationBackend.DisableResetPassword)

	duoSelfEnrollment := f
	if configuration.DuoAPI != nil {
		duoSelfEnrollment = strconv.FormatBool(configuration.DuoAPI.EnableSelfEnrollment)
	}

	embeddedPath, _ := fs.Sub(assets, "public_html")
	embeddedFS := fasthttpadaptor.NewFastHTTPHandler(http.FileServer(http.FS(embeddedPath)))

	https := configuration.Server.TLS.Key != "" && configuration.Server.TLS.Certificate != ""

	serveIndexHandler := ServeTemplatedFile(embeddedAssets, indexFile, configuration.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, configuration.Session.Name, configuration.Theme, https)
	serveSwaggerHandler := ServeTemplatedFile(swaggerAssets, indexFile, configuration.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, configuration.Session.Name, configuration.Theme, https)
	serveSwaggerAPIHandler := ServeTemplatedFile(swaggerAssets, apiFile, configuration.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, configuration.Session.Name, configuration.Theme, https)

	r := router.New()
	r.GET("/", autheliaMiddleware(serveIndexHandler))
	r.OPTIONS("/", autheliaMiddleware(handleOPTIONS))

	r.GET("/api/", autheliaMiddleware(serveSwaggerHandler))
	r.GET("/api/"+apiFile, autheliaMiddleware(serveSwaggerAPIHandler))

	for _, f := range rootFiles {
		r.GET("/"+f, middlewares.AssetOverrideMiddleware(configuration.Server.AssetPath, embeddedFS))
	}

	r.GET("/static/{filepath:*}", middlewares.AssetOverrideMiddleware(configuration.Server.AssetPath, embeddedFS))
	r.ANY("/api/{filepath:*}", embeddedFS)

	r.GET("/api/health", autheliaMiddleware(handlers.HealthGet))
	r.GET("/api/state", autheliaMiddleware(handlers.StateGet))

	r.GET("/api/configuration", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.ConfigurationGet)))

	r.GET("/api/verify", autheliaMiddleware(handlers.VerifyGet(configuration.AuthenticationBackend)))
	r.HEAD("/api/verify", autheliaMiddleware(handlers.VerifyGet(configuration.AuthenticationBackend)))

	r.POST("/api/checks/safe-redirection", autheliaMiddleware(handlers.CheckSafeRedirection))

	r.POST("/api/firstfactor", autheliaMiddleware(handlers.FirstFactorPost(middlewares.TimingAttackDelay(10, 250, 85, time.Second))))
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

	if !configuration.TOTP.Disable {
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

	if !configuration.Webauthn.Disable {
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
	if configuration.DuoAPI != nil {
		var duoAPI duo.API
		if os.Getenv("ENVIRONMENT") == dev {
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

		r.GET("/api/secondfactor/duo_devices", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoDevicesGet(duoAPI))))

		r.POST("/api/secondfactor/duo", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoPost(duoAPI))))

		r.POST("/api/secondfactor/duo_device", autheliaMiddleware(
			middlewares.RequireFirstFactor(handlers.SecondFactorDuoDevicePost)))
	}

	if configuration.Server.EnablePprof {
		r.GET("/debug/pprof/{name?}", pprofhandler.PprofHandler)
	}

	if configuration.Server.EnableExpvars {
		r.GET("/debug/vars", expvarhandler.ExpvarHandler)
	}

	r.NotFound = autheliaMiddleware(serveIndexHandler)

	handler := middlewares.LogRequestMiddleware(r.Handler)
	if configuration.Server.Path != "" {
		handler = middlewares.StripPathMiddleware(configuration.Server.Path, handler)
	}

	if providers.OpenIDConnect.Fosite != nil {
		handlers.RegisterOIDC(r, autheliaMiddleware)
	}

	return handler
}

// Start Authelia's internal webserver with the given configuration and providers.
func Start(configuration schema.Configuration, providers middlewares.Providers) {
	logger := logging.Logger()

	handler := registerRoutes(configuration, providers)

	server := &fasthttp.Server{
		ErrorHandler:          autheliaErrorHandler,
		Handler:               handler,
		NoDefaultServerHeader: true,
		ReadBufferSize:        configuration.Server.ReadBufferSize,
		WriteBufferSize:       configuration.Server.WriteBufferSize,
	}

	addrPattern := net.JoinHostPort(configuration.Server.Host, strconv.Itoa(configuration.Server.Port))

	listener, err := net.Listen("tcp", addrPattern)
	if err != nil {
		logger.Fatalf("Error initializing listener: %s", err)
	}

	if configuration.Server.TLS.Certificate != "" && configuration.Server.TLS.Key != "" {
		if err = writeHealthCheckEnv(configuration.Server.DisableHealthcheck, "https", configuration.Server.Host, configuration.Server.Path, configuration.Server.Port); err != nil {
			logger.Fatalf("Could not configure healthcheck: %v", err)
		}

		if configuration.Server.Path == "" {
			logger.Infof("Listening for TLS connections on '%s' path '/'", addrPattern)
		} else {
			logger.Infof("Listening for TLS connections on '%s' paths '/' and '%s'", addrPattern, configuration.Server.Path)
		}

		logger.Fatal(server.ServeTLS(listener, configuration.Server.TLS.Certificate, configuration.Server.TLS.Key))
	} else {
		if err = writeHealthCheckEnv(configuration.Server.DisableHealthcheck, "http", configuration.Server.Host, configuration.Server.Path, configuration.Server.Port); err != nil {
			logger.Fatalf("Could not configure healthcheck: %v", err)
		}

		if configuration.Server.Path == "" {
			logger.Infof("Listening for non-TLS connections on '%s' path '/'", addrPattern)
		} else {
			logger.Infof("Listening for non-TLS connections on '%s' paths '/' and '%s'", addrPattern, configuration.Server.Path)
		}
		logger.Fatal(server.Serve(listener))
	}
}
