package handlers

import (
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"

	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// isConsentMissing compares the requestedScopes and requestedAudience to the workflows
// GrantedScopes and GrantedAudience and returns true if they do not match or the workflow is nil.
func isConsentMissing(workflow *session.OIDCWorkflowSession, requestedScopes, requestedAudience []string) (isMissing bool) {
	if workflow == nil {
		return true
	}

	return len(requestedScopes) > 0 && utils.IsStringSlicesDifferent(requestedScopes, workflow.GrantedScopes) ||
		len(requestedAudience) > 0 && utils.IsStringSlicesDifferentFold(requestedAudience, workflow.GrantedAudience)
}

func newOpenIDSession(subject string) *oidc.OpenIDSession {
	return &oidc.OpenIDSession{
		DefaultSession: &openid.DefaultSession{
			Claims:  new(jwt.IDTokenClaims),
			Headers: new(jwt.Headers),
			Subject: subject,
		},
		Extra: map[string]interface{}{},
	}
}

func oidcGrantRequests(ar fosite.AuthorizeRequester, scopes, audiences []string, userSession *session.UserSession) (extraClaims map[string]interface{}) {
	extraClaims = map[string]interface{}{}

	for _, scope := range scopes {
		if ar != nil {
			ar.GrantScope(scope)
		}

		switch scope {
		case oidc.ScopeGroups:
			extraClaims[oidc.ClaimGroups] = userSession.Groups
		case oidc.ScopeProfile:
			extraClaims[oidc.ClaimPreferredUsername] = userSession.Username
			extraClaims[oidc.ClaimDisplayName] = userSession.DisplayName
		case oidc.ScopeEmail:
			if len(userSession.Emails) != 0 {
				extraClaims[oidc.ClaimEmail] = userSession.Emails[0]
				if len(userSession.Emails) > 1 {
					extraClaims[oidc.ClaimEmailAlts] = userSession.Emails[1:]
				}
				// TODO (james-d-elliott): actually verify emails and record that information.
				extraClaims[oidc.ClaimEmailVerified] = true
			}
		}
	}

	if ar != nil {
		for _, audience := range audiences {
			ar.GrantAudience(audience)
		}

		if !utils.IsStringInSlice(ar.GetClient().GetID(), ar.GetGrantedAudience()) {
			ar.GrantAudience(ar.GetClient().GetID())
		}
	}

	return extraClaims
}
