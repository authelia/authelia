package handlers

import (
	"net/url"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

// CheckSafeRedirectionPOST handler checking whether the redirection to a given URL provided in body is safe.
func CheckSafeRedirectionPOST(ctx *middlewares.AutheliaCtx) {
	var (
		s   session.UserSession
		err error
	)
	if s, err = ctx.GetSession(); err != nil {
		ctx.ReplyUnauthorized()
		return
	}

	if s.IsAnonymous() {
		ctx.ReplyUnauthorized()
		return
	}

	var (
		bodyJSON  checkURIWithinDomainRequestBody
		targetURI *url.URL
	)

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred parsing the safe redirection request body")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if targetURI, err = url.ParseRequestURI(bodyJSON.URI); err != nil {
		ctx.GetLogger().WithError(err).Errorf("Error occurred determining if the URI '%s' is safe to redirect to as it could not be parsed", bodyJSON.URI)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(checkURIWithinDomainResponseBody{OK: ctx.IsSafeRedirectionTargetURI(targetURI)}); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred setting the safe redirection response body")
		ctx.SetJSONError(messageOperationFailed)

		return
	}
}
