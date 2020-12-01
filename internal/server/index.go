package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

var alphaNumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// ServeIndex serve the index.html file with nonce generated for supporting
// restrictive CSP while using material-ui from the embedded virtual filesystem.
//go:generate broccoli -src ../../public_html -o public_html
func ServeIndex(publicDir, base, rememberMe, resetPassword string) fasthttp.RequestHandler {
	f, err := br.Open(publicDir + "/index.html")
	if err != nil {
		logging.Logger().Fatalf("Unable to open index.html: %v", err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		logging.Logger().Fatalf("Unable to read index.html: %v", err)
	}

	tmpl, err := template.New("index").Parse(string(b))
	if err != nil {
		logging.Logger().Fatalf("Unable to parse index.html template: %v", err)
	}

	return func(ctx *fasthttp.RequestCtx) {
		nonce := utils.RandomString(32, alphaNumericRunes)

		ctx.SetContentType("text/html; charset=utf-8")

		switch {
		case os.Getenv("ENVIRONMENT") == dev:
			ctx.Response.Header.Add("Content-Security-Policy", fmt.Sprintf("default-src 'self' 'unsafe-eval'; object-src 'none'; style-src 'self' 'nonce-%s'", nonce))
		case publicDir == "/public_html/api":
			ctx.Response.Header.Add("Content-Security-Policy", fmt.Sprintf("base-uri 'self' ; default-src 'self' ; img-src 'self' https://validator.swagger.io data: ; object-src 'none' ; script-src 'self' 'unsafe-inline' 'nonce-%s' ; style-src 'self' 'nonce-%s'", nonce, nonce))
		default:
			ctx.Response.Header.Add("Content-Security-Policy", fmt.Sprintf("default-src 'self' ; object-src 'none'; style-src 'self' 'nonce-%s'", nonce))
		}

		err := tmpl.Execute(ctx.Response.BodyWriter(), struct{ Base, CSPNonce, RememberMe, ResetPassword string }{Base: base, CSPNonce: nonce, RememberMe: rememberMe, ResetPassword: resetPassword})
		if err != nil {
			ctx.Error("An error occurred", 503)
			logging.Logger().Errorf("Unable to execute template: %v", err)

			return
		}
	}
}
