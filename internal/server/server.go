package server

import (
	"fmt"
	"os"
	"path"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/duo"
	"github.com/authelia/authelia/internal/handlers"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/middlewares"
	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// StartServer start Authelia server with the given configuration and providers.
func StartServer(configuration schema.Configuration, providers middlewares.Providers) {
	router := router.New()

	autheliaMiddleware := middlewares.AutheliaMiddleware(configuration, providers)

	publicDir := os.Getenv("PUBLIC_DIR")
	if publicDir == "" {
		publicDir = "./public_html"
	}
	fmt.Println("Selected public_html directory is ", publicDir)

	router.GET("/", fasthttp.FSHandler(publicDir, 0))
	router.ServeFiles("/static/*filepath", publicDir+"/static")

	router.GET("/api/state", autheliaMiddleware(handlers.StateGet))

	router.GET("/api/configuration", autheliaMiddleware(handlers.ConfigurationGet))
	router.GET("/api/configuration/extended", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.ExtendedConfigurationGet)))

	router.GET("/api/verify", autheliaMiddleware(handlers.VerifyGet))

	router.POST("/api/firstfactor", autheliaMiddleware(handlers.FirstFactorPost))
	router.POST("/api/logout", autheliaMiddleware(handlers.LogoutPost))

	// Password reset related endpoints.
	router.POST("/api/reset-password/identity/start", autheliaMiddleware(
		handlers.ResetPasswordIdentityStart))
	router.POST("/api/reset-password/identity/finish", autheliaMiddleware(
		handlers.ResetPasswordIdentityFinish))
	router.POST("/api/reset-password", autheliaMiddleware(
		handlers.ResetPasswordPost))

	// Information about the user
	router.GET("/api/user/info", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.UserInfoGet)))
	router.POST("/api/user/info/2fa_method", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.MethodPreferencePost)))

	// TOTP related endpoints
	router.POST("/api/secondfactor/totp/identity/start", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityStart)))
	router.POST("/api/secondfactor/totp/identity/finish", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPIdentityFinish)))
	router.POST("/api/secondfactor/totp", autheliaMiddleware(
		middlewares.RequireFirstFactor(handlers.SecondFactorTOTPPost(&handlers.TOTPVerifierImpl{}))))

	// U2F related endpoints
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

	// Configure DUO api endpoint only if configuration exists
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

	router.NotFound = func(ctx *fasthttp.RequestCtx) {
		ctx.SendFile(path.Join(publicDir, "index.html"))
	}

	portPattern := fmt.Sprintf("%s:%d", configuration.Host, configuration.Port)
	logging.Logger().Infof("Authelia is listening on %s", portPattern)

	logging.Logger().Fatal(fasthttp.ListenAndServe(portPattern,
		middlewares.LogRequestMiddleware(router.Handler)))
}
