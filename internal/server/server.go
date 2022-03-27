package server

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/authelia/authelia/v4/internal/logging"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/expvarhandler"
	"github.com/valyala/fasthttp/pprofhandler"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func registerRoutes(configuration schema.Configuration, providers middlewares.Providers) fasthttp.RequestHandler {
	autheliaMiddleware := middlewares.AutheliaMiddleware(configuration, providers)
	rememberMe := strconv.FormatBool(configuration.Session.RememberMeDuration != schema.RememberMeDisabled)
	resetPassword := strconv.FormatBool(!configuration.AuthenticationBackend.DisableResetPassword)

	resetPasswordCustomURL := configuration.AuthenticationBackend.PasswordReset.CustomURL.String()

	duoSelfEnrollment := f
	if configuration.DuoAPI != nil {
		duoSelfEnrollment = strconv.FormatBool(configuration.DuoAPI.EnableSelfEnrollment)
	}

	handlerPublicHTML := newPublicHTMLEmbeddedHandler()
	handlerLocales := newLocalesEmbeddedHandler()

	https := configuration.Server.TLS.Key != "" && configuration.Server.TLS.Certificate != ""

	serveIndexHandler := ServeTemplatedFile(embeddedAssets, indexFile, configuration.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, resetPasswordCustomURL, configuration.Session.Name, configuration.Theme, https)
	serveSwaggerHandler := ServeTemplatedFile(swaggerAssets, indexFile, configuration.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, resetPasswordCustomURL, configuration.Session.Name, configuration.Theme, https)
	serveSwaggerAPIHandler := ServeTemplatedFile(swaggerAssets, apiFile, configuration.Server.AssetPath, duoSelfEnrollment, rememberMe, resetPassword, resetPasswordCustomURL, configuration.Session.Name, configuration.Theme, https)

	r := router.New()
	r.GET("/", autheliaMiddleware(serveIndexHandler))
	r.OPTIONS("/", autheliaMiddleware(handleOPTIONS))

	for _, f := range rootFiles {
		r.GET("/"+f, handlerPublicHTML)
	}

	r.GET("/api/", autheliaMiddleware(serveSwaggerHandler))
	r.GET("/api/"+apiFile, autheliaMiddleware(serveSwaggerAPIHandler))

	for _, file := range swaggerFiles {
		r.GET("/api/"+file, handlerPublicHTML)
	}

	r.GET("/favicon.ico", middlewares.AssetOverrideMiddleware(configuration.Server.AssetPath, 0, handlerPublicHTML))
	r.GET("/static/media/logo.png", middlewares.AssetOverrideMiddleware(configuration.Server.AssetPath, 2, handlerPublicHTML))
	r.GET("/static/{filepath:*}", handlerPublicHTML)

	r.GET("/locales/{language:[a-z]{1,3}}-{variant:[a-z0-9-]+}/{namespace:[a-z]+}.json", middlewares.AssetOverrideMiddleware(configuration.Server.AssetPath, 0, handlerLocales))
	r.GET("/locales/{language:[a-z]{1,3}}/{namespace:[a-z]+}.json", middlewares.AssetOverrideMiddleware(configuration.Server.AssetPath, 0, handlerLocales))

	r.GET("/api/health", autheliaMiddleware(handlers.HealthGet))
	r.GET("/api/state", autheliaMiddleware(handlers.StateGet))

	r.GET("/api/configuration", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.ConfigurationGet)))

	r.GET("/api/configuration/password-policy", autheliaMiddleware(handlers.PasswordPolicyConfigurationGet))

	r.GET("/api/verify", autheliaMiddleware(handlers.VerifyGet(configuration.AuthenticationBackend)))
	r.HEAD("/api/verify", autheliaMiddleware(handlers.VerifyGet(configuration.AuthenticationBackend)))

	r.POST("/api/checks/safe-redirection", autheliaMiddleware(handlers.CheckSafeRedirection))

	r.POST("/api/firstfactor", autheliaMiddleware(handlers.FirstFactorPost(middlewares.TimingAttackDelay(10, 250, 85, time.Second))))
	r.POST("/api/logout", autheliaMiddleware(handlers.LogoutPost))

	// Only register endpoints if forgot password is not disabled.
	if !configuration.AuthenticationBackend.DisableResetPassword &&
		configuration.AuthenticationBackend.PasswordReset.CustomURL.String() == "" {
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

	r.NotFound = handleNotFound(autheliaMiddleware(serveIndexHandler))

	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = func(ctx *fasthttp.RequestCtx) {
		handlers.SetStatusCodeResponse(ctx, fasthttp.StatusMethodNotAllowed)
	}

	handler := middlewares.LogRequestMiddleware(r.Handler)
	if configuration.Server.Path != "" {
		handler = middlewares.StripPathMiddleware(configuration.Server.Path, handler)
	}

	if providers.OpenIDConnect.Fosite != nil {
		handlers.RegisterOIDC(r, autheliaMiddleware)
	}

	return handler
}

// CreateServer Create Authelia's internal webserver with the given configuration and providers.
func CreateServer(configuration schema.Configuration, providers middlewares.Providers) (*fasthttp.Server, net.Listener) {
	handler := registerRoutes(configuration, providers)

	server := &fasthttp.Server{
		ErrorHandler:          autheliaErrorHandler,
		Handler:               handler,
		NoDefaultServerHeader: true,
		ReadBufferSize:        configuration.Server.ReadBufferSize,
		WriteBufferSize:       configuration.Server.WriteBufferSize,
	}
	logger := logging.Logger()

	address := net.JoinHostPort(configuration.Server.Host, strconv.Itoa(configuration.Server.Port))

	var (
		listener         net.Listener
		err              error
		connectionType   string
		connectionScheme string
	)

	if configuration.Server.TLS.Certificate != "" && configuration.Server.TLS.Key != "" {
		connectionType, connectionScheme = "TLS", schemeHTTPS
		err = server.AppendCert(configuration.Server.TLS.Certificate, configuration.Server.TLS.Key)

		if err != nil {
			logger.Fatalf("unable to load certificate: %v", err)
		}

		if len(configuration.Server.TLS.ClientCertificates) > 0 {
			caCertPool := x509.NewCertPool()

			for _, path := range configuration.Server.TLS.ClientCertificates {
				cert, err := os.ReadFile(path)
				if err != nil {
					logger.Fatalf("Cannot read client TLS certificate %s: %s", path, err)
				}

				caCertPool.AppendCertsFromPEM(cert)
			}

			// ClientCAs should never be nil, otherwise the system cert pool is used for client authentication
			// but we don't want everybody on the Internet to be able to authenticate.
			server.TLSConfig.ClientCAs = caCertPool
			server.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		listener, err = tls.Listen("tcp", address, server.TLSConfig.Clone())
		if err != nil {
			logger.Fatalf("Error initializing listener: %s", err)
		}
	} else {
		connectionType, connectionScheme = "non-TLS", schemeHTTP
		listener, err = net.Listen("tcp", address)
		if err != nil {
			logger.Fatalf("Error initializing listener: %s", err)
		}
	}

	if err = writeHealthCheckEnv(configuration.Server.DisableHealthcheck, connectionScheme, configuration.Server.Host,
		configuration.Server.Path, configuration.Server.Port); err != nil {
		logger.Fatalf("Could not configure healthcheck: %v", err)
	}

	actualAddress := listener.Addr().(*net.TCPAddr).AddrPort()
	if configuration.Server.Path == "" {
		logger.Infof("Initializing server for %s connections on '%s' path '/'", connectionType, actualAddress)
	} else {
		logger.Infof("Initializing server for %s connections on '%s' paths '/' and '%s'", connectionType, actualAddress, configuration.Server.Path)
	}

	return server, listener
}
