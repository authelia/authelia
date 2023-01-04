package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// CheckSafeRedirectionPOST handler checking whether the redirection to a given URL provided in body is safe.
func CheckSafeRedirectionPOST(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	if userSession.IsAnonymous() {
		ctx.ReplyUnauthorized()
		return
	}

	var (
		bodyJSON  checkURIWithinDomainRequestBody
		targetURI *url.URL
		err       error
	)

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request body: %w", err), messageOperationFailed)
		return
	}

	if targetURI, err = url.ParseRequestURI(bodyJSON.URI); err != nil {
		ctx.Error(fmt.Errorf("unable to determine if uri %s is safe to redirect to: failed to parse URI '%s': %w", bodyJSON.URI, bodyJSON.URI, err), messageOperationFailed)
		return
	}

	if err = ctx.SetJSONBody(checkURIWithinDomainResponseBody{OK: ctx.GetTargetURICookieDomain(targetURI) != ""}); err != nil {
		ctx.Error(fmt.Errorf("unable to create response body: %w", err), messageOperationFailed)
		return
	}
}
