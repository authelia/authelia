package handlers

import (
	"bytes"
	"net/url"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

var bytesEmpty = []byte("")

func ctxGetPortalURL(ctx *middlewares.AutheliaCtx) (portalURL *url.URL) {
	var rawURL []byte

	if rawURL = ctx.QueryArgRedirect(); rawURL != nil && !bytes.Equal(rawURL, bytesEmpty) {
		portalURL, _ = url.ParseRequestURI(string(rawURL))

		return portalURL
	} else if rawURL = ctx.XAutheliaURL(); rawURL != nil && !bytes.Equal(rawURL, bytesEmpty) {
		portalURL, _ = url.ParseRequestURI(string(rawURL))

		return portalURL
	}

	return nil
}
