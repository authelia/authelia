package handlers

import (
	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func oidcGrantRequests(ar fosite.AuthorizeRequester, consent *model.OAuth2ConsentSession, userSession *session.UserSession) (extraClaims map[string]any) {
	extraClaims = map[string]any{}

	for _, scope := range consent.GrantedScopes {
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
		for _, audience := range consent.GrantedAudience {
			ar.GrantAudience(audience)
		}
	}

	return extraClaims
}
