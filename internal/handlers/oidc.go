package handlers

import (
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"

	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/utils"
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

func newOpenIDSession(subject string) *OpenIDSession {
	return &OpenIDSession{
		DefaultSession: &openid.DefaultSession{
			Claims:  new(jwt.IDTokenClaims),
			Headers: new(jwt.Headers),
			Subject: subject,
		},
		Extra: map[string]interface{}{},
	}
}

type OpenIDSession struct {
	*openid.DefaultSession `json:"idToken"`

	Extra    map[string]interface{} `json:"extra"`
	ClientID string
}
