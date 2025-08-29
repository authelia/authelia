package handlers

import (
	"net/http"
	"net/url"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func handleOAuth2AuthorizationClaims(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, _ *http.Request, flow string, userSession session.UserSession, details *authentication.UserDetailsExtended, client oidc.Client, requester oauthelia2.Requester, issuer *url.URL, consent *model.OAuth2ConsentSession, extra map[string]any) (requests *oidc.ClaimsRequests, handled bool) {
	var err error

	if requester.GetRequestedScopes().Has(oidc.ScopeOpenID) {
		if requests, err = oidc.NewClaimRequests(requester.GetRequestForm()); err != nil {
			ctx.Logger.WithError(err).Errorf("%s Request with id '%s' on client with id '%s' could not be processed: error occurred parsing the claims parameter", flow, requester.GetID(), client.GetID())

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, err)

			return nil, true
		}

		claimsStrategy := client.GetClaimsStrategy()
		scopeStrategy := ctx.Providers.OpenIDConnect.GetScopeStrategy(ctx)

		if err = claimsStrategy.ValidateClaimsRequests(ctx, scopeStrategy, client, requests); err != nil {
			ctx.Logger.WithError(oauthelia2.ErrorToDebugRFC6749Error(err)).Errorf("%s Request with id '%s' on client with id '%s' could not be processed: the client requested claims were not permitted.", flow, requester.GetID(), client.GetID())

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oauthelia2.ErrAccessDenied.WithHint("The requested subject was not the same subject that attempted to authorize the request."))

			return nil, true
		}

		if requested, ok := requests.MatchesIssuer(issuer); !ok {
			ctx.Logger.Errorf("%s Request with id '%s' on client with id '%s' could not be processed: the client requested issuer '%s' but the issuer for the token will be '%s' instead", flow, requester.GetID(), client.GetID(), requested, issuer.String())

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oauthelia2.ErrAccessDenied.WithHint("The requested issuer was not the same issuer that attempted to authorize the request."))

			return nil, true
		}

		if requested, ok := requests.MatchesSubject(consent.Subject.UUID.String()); !ok {
			ctx.Logger.Errorf("%s Request with id '%s' on client with id '%s' could not be processed: the client requested subject '%s' but the subject value for '%s' is '%s' for the '%s' sector identifier", flow, requester.GetID(), client.GetID(), requested, userSession.Username, consent.Subject.UUID, client.GetSectorIdentifierURI())

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, oauthelia2.ErrAccessDenied.WithHint("The requested subject was not the same subject that attempted to authorize the request."))

			return nil, true
		}

		oidc.GrantScopeAudienceConsent(requester, consent)

		var implicit bool

		if r, ok := requester.(oauthelia2.AuthorizeRequester); ok {
			implicit = r.GetResponseTypes().ExactOne(oidc.ResponseTypeImplicitFlowIDToken)
		}

		if err = claimsStrategy.HydrateIDTokenClaims(ctx, scopeStrategy, client, requester.GetGrantedScopes(), oauthelia2.Arguments(consent.GrantedClaims), requests.GetIDTokenRequests(), details, consent.RequestedAt, ctx.GetClock().Now(), nil, extra, implicit); err != nil {
			ctx.Logger.Errorf("%s Response for Request with id '%s' on client with id '%s' could not be created: %s", flow, requester.GetID(), client.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

			ctx.Providers.OpenIDConnect.WriteDynamicAuthorizeError(ctx, rw, requester, err)

			return nil, true
		}
	} else {
		oidc.GrantScopeAudienceConsent(requester, consent)
	}

	return requests, false
}
