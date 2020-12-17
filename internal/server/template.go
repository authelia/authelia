package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

var alphaNumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// ServeTemplatedFile serves a templated version of a specified file,
// this is utilised to pass information between the backend and frontend
// and generate a nonce to support a restrictive CSP while using material-ui.
//go:generate broccoli -src ../../public_html -o public_html
func ServeTemplatedFile(publicDir, file, base, session, rememberMe, resetPassword string) fasthttp.RequestHandler {
	f, err := br.Open(publicDir + file)
	if err != nil {
		logging.Logger().Fatalf("Unable to open %s: %s", file, err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		logging.Logger().Fatalf("Unable to read %s: %s", file, err)
	}

	tmpl, err := template.New("file").Parse(string(b))
	if err != nil {
		logging.Logger().Fatalf("Unable to parse %s template: %s", file, err)
	}

	return func(ctx *fasthttp.RequestCtx) {
		nonce := utils.RandomString(32, alphaNumericRunes)

		switch extension := filepath.Ext(file); extension {
		case ".html":
			ctx.SetContentType("text/html; charset=utf-8")
		default:
			ctx.SetContentType("text/plain; charset=utf-8")
		}

		switch {
		case os.Getenv("ENVIRONMENT") == dev:
			ctx.Response.Header.Add("Content-Security-Policy", fmt.Sprintf("default-src 'self' 'unsafe-eval'; object-src 'none'; style-src 'self' 'nonce-%s'", nonce))
		case publicDir == "/public_html/api/":
			ctx.Response.Header.Add("Content-Security-Policy", fmt.Sprintf("base-uri 'self' ; default-src 'self' ; img-src 'self' https://validator.swagger.io data: ; object-src 'none' ; script-src 'self' 'unsafe-inline' 'nonce-%s' ; style-src 'self' 'nonce-%s'", nonce, nonce))
		default:
			ctx.Response.Header.Add("Content-Security-Policy", fmt.Sprintf("default-src 'self' ; object-src 'none'; style-src 'self' 'nonce-%s'", nonce))
		}

		err := tmpl.Execute(ctx.Response.BodyWriter(), struct{ Base, CSPNonce, Session, RememberMe, ResetPassword string }{Base: base, CSPNonce: nonce, Session: session, RememberMe: rememberMe, ResetPassword: resetPassword})
		if err != nil {
			ctx.Error("An error occurred", 503)
			logging.Logger().Errorf("Unable to execute template: %v", err)

			return
		}
	}
}
