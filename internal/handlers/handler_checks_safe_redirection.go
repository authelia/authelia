package handlers

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

// CheckSafeRedirectionPOST handler checking whether the redirection to a given URL provided in body is safe.
func CheckSafeRedirectionPOST(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	if userSession.AuthenticationLevel == authentication.NotAuthenticated {
		ctx.ReplyUnauthorized()
		return
	}

	var reqBody checkURIWithinDomainRequestBody

	err := ctx.ParseBody(&reqBody)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to parse request body: %w", err), messageOperationFailed)
		return
	}

	safe, err := utils.IsRedirectionURISafe(reqBody.URI, ctx.Configuration.Session.ProtectedDomains)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to determine if uri %s is safe to redirect to: %w", reqBody.URI, err), messageOperationFailed)
		return
	}

	err = ctx.SetJSONBody(checkURIWithinDomainResponseBody{
		OK: safe,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create response body: %w", err), messageOperationFailed)
		return
	}
}
