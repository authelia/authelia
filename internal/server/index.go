//go:generate broccoli -src ../../public_html -o public_html
package server

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

var alphaNumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// ServeIndex serve the index.html file with nonce generated for supporting
// restrictive CSP while using material-ui from the local or embedded
// virtual filesystem.
func ServeIndex(publicDir string) fasthttp.RequestHandler {
	var f io.Reader
	var err error

	if publicDir == "/public_html" {
		f, err = br.Open(publicDir + "/index.html")
	} else {
		f, err = os.Open(publicDir + "/index.html")
	}

	if err != nil {
		logging.Logger().Fatalf("Unable to open index.html: %v", err)
		return func(ctx *fasthttp.RequestCtx) {
			ctx.Error("An error occurred", 500)
		}
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		logging.Logger().Fatalf("Unable to read index.html: %v", err)
		return func(ctx *fasthttp.RequestCtx) {
			ctx.Error("An error occurred", 500)
		}
	}

	tmpl, err := template.New("index").Parse(string(b))
	if err != nil {
		logging.Logger().Fatalf("Unable to parse index.html template: %v", err)
		return func(ctx *fasthttp.RequestCtx) {
			ctx.Error("An error occurred", 500)
		}
	}

	return func(ctx *fasthttp.RequestCtx) {
		nonce := utils.RandomString(32, alphaNumericRunes)
		ctx.SetContentType("text/html; charset=utf-8")
		ctx.Response.Header.Add("Content-Security-Policy", fmt.Sprintf("default-src 'self'; style-src 'self' 'nonce-%s'", nonce))
		err := tmpl.Execute(ctx.Response.BodyWriter(), struct{ CSPNonce string }{CSPNonce: nonce})
		if err != nil {
			ctx.Error("An error occurred", 503)
			logging.Logger().Errorf("Unable to execute template: %v", err)
			return
		}
	}
}
