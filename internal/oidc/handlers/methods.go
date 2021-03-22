package handlers

import (
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/utils"
)

// IsConsentMissing compares the requestedScopes and requestedAudience to the workflows
// GrantedScopes and GrantedAudience and returns true if they do not match or the workflow is nil.
func IsConsentMissing(workflow *session.OIDCWorkflowSession, requestedScopes, requestedAudience []string) (isMissing bool) {
	if workflow == nil {
		return true
	}

	return len(requestedScopes) > 0 && utils.IsStringSlicesDifferent(requestedScopes, workflow.GrantedScopes) ||
		len(requestedAudience) > 0 && utils.IsStringSlicesDifferentFold(requestedAudience, workflow.GrantedAudience)
}

func scopeNamesToScopes(scopeSlice []string) (scopes []Scope) {
	for _, name := range scopeSlice {
		if val, ok := scopeDescriptions[name]; ok {
			scopes = append(scopes, Scope{name, val})
		} else {
			scopes = append(scopes, Scope{name, name})
		}
	}

	return scopes
}

func audienceNamesToAudience(scopeSlice []string) (audience []Audience) {
	for _, name := range scopeSlice {
		if val, ok := audienceDescriptions[name]; ok {
			audience = append(audience, Audience{name, val})
		} else {
			audience = append(audience, Audience{name, name})
		}
	}

	return audience
}

func newSession(ctx *middlewares.AutheliaCtx, ar fosite.AuthorizeRequester) *openid.DefaultSession {
	userSession := ctx.GetSession()

	scopes := ar.GetGrantedScopes()

	extra := map[string]interface{}{}

	if len(userSession.Emails) != 0 && scopes.Has("email") {
		extra["email"] = userSession.Emails[0]
		extra["email_verified"] = true
	}

	if scopes.Has("groups") {
		extra["groups"] = userSession.Groups
	}

	if scopes.Has("profile") {
		extra["name"] = userSession.DisplayName
	}

	/*
		TODO: Adjust auth backends to return more profile information.
		It's probably ideal to adjust the auth providers at this time to not store 'extra' information in the session
		storage, and instead create a memory only storage for them.
		This is a simple design, have a map with a key of username, and a struct with the relevant information.
	*/

	oidcSession := newDefaultSession(ctx)
	oidcSession.Claims.Extra = extra
	oidcSession.Claims.Subject = userSession.Username
	oidcSession.Claims.Audience = ar.GetGrantedAudience()

	return oidcSession
}

func newDefaultSession(ctx *middlewares.AutheliaCtx) *openid.DefaultSession {
	issuer, err := ctx.ForwardedProtoHost()

	if err != nil {
		issuer = fallbackOIDCIssuer
	}

	return &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:      issuer,
			Subject:     "",
			ExpiresAt:   time.Now().Add(time.Hour * 6),
			IssuedAt:    time.Now(),
			RequestedAt: time.Now(),
			AuthTime:    time.Now(),
			Extra:       make(map[string]interface{}),
		},
		Headers: &jwt.Headers{
			Extra: make(map[string]interface{}),
		},
	}
}
