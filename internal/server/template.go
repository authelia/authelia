package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/valyala/fasthttp"
)

// ServeTemplatedFile serves a templated version of a specified file,
// this is utilised to pass information between the backend and frontend
// and generate a nonce to support a restrictive CSP while using material-ui.
func ServeTemplatedFile(publicDir, file, assetPath, duoSelfEnrollment, rememberMe, resetPassword, resetPasswordCustomURL, session, theme string, https bool) middlewares.RequestHandler {
	logger := logging.Logger()

	a, err := assets.Open(publicDir + file)
	if err != nil {
		logger.Fatalf("Unable to open %s: %s", file, err)
	}

	b, err := io.ReadAll(a)
	if err != nil {
		logger.Fatalf("Unable to read %s: %s", file, err)
	}

	tmpl, err := template.New("file").Parse(string(b))
	if err != nil {
		logger.Fatalf("Unable to parse %s template: %s", file, err)
	}

	return func(ctx *middlewares.AutheliaCtx) {
		base := ""
		if baseURL := ctx.UserValueBytes(middlewares.UserValueKeyBaseURL); baseURL != nil {
			base = baseURL.(string)
		}

		logoOverride := f

		if assetPath != "" {
			if _, err := os.Stat(filepath.Join(assetPath, logoFile)); err == nil {
				logoOverride = t
			}
		}

		var scheme = schemeHTTPS

		if !https {
			proto := string(ctx.XForwardedProto())
			switch proto {
			case "":
				break
			case schemeHTTP, schemeHTTPS:
				scheme = proto
			}
		}

		baseURL := scheme + "://" + string(ctx.XForwardedHost()) + base + "/"
		nonce := utils.RandomString(32, utils.AlphaNumericCharacters, true)

		switch filepath.Ext(file) {
		case ".html":
			ctx.SetContentTypeTextHTML()
		default:
			ctx.SetContentTypeTextPlain()
		}

		switch {
		case publicDir == swaggerAssets:
			ctx.SetContentSecurityPolicy(fmt.Sprintf("base-uri 'self'; default-src 'self'; img-src 'self' https://validator.swagger.io data:; object-src 'none'; script-src 'self' 'unsafe-inline' 'nonce-%s'; style-src 'self' 'nonce-%s'", nonce, nonce))
		case ctx.Configuration.Server.Headers.CSPTemplate != "":
			ctx.SetContentSecurityPolicy(strings.ReplaceAll(ctx.Configuration.Server.Headers.CSPTemplate, cspNoncePlaceholder, nonce))
		case os.Getenv("ENVIRONMENT") == dev:
			ctx.SetContentSecurityPolicy(fmt.Sprintf(cspDefaultTemplate, " 'unsafe-eval'", nonce))
		default:
			ctx.SetContentSecurityPolicy(fmt.Sprintf(cspDefaultTemplate, "", nonce))
		}

		err = tmpl.Execute(ctx.Response.BodyWriter(), struct{ Base, BaseURL, CSPNonce, DuoSelfEnrollment, LogoOverride, RememberMe, ResetPassword, ResetPasswordCustomURL, Session, Theme string }{Base: base, BaseURL: baseURL, CSPNonce: nonce, DuoSelfEnrollment: duoSelfEnrollment, LogoOverride: logoOverride, RememberMe: rememberMe, ResetPassword: resetPassword, ResetPasswordCustomURL: resetPasswordCustomURL, Session: session, Theme: theme})
		if err != nil {
			ctx.ReplyStatusCode(fasthttp.StatusServiceUnavailable)
			ctx.Logger.Errorf("Unable to execute template: %v", err)

			return
		}
	}
}

func writeHealthCheckEnv(disabled bool, scheme, host, path string, port int) (err error) {
	if disabled {
		return nil
	}

	_, err = os.Stat("/app/healthcheck.sh")
	if err != nil {
		return nil
	}

	_, err = os.Stat("/app/.healthcheck.env")
	if err != nil {
		return nil
	}

	file, err := os.OpenFile("/app/.healthcheck.env", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	if host == "0.0.0.0" {
		host = localhost
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}

	_, err = file.WriteString(fmt.Sprintf(healthCheckEnv, scheme, host, port, path))

	return err
}
