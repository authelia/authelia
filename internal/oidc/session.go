package oidc

import (
	"github.com/google/uuid"
	"github.com/mohae/deepcopy"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
)

// Session holds OpenID Connect 1.0 Session information.
type Session struct {
	*openid.DefaultSession `json:"id_token"`

	ChallengeID           uuid.NullUUID  `json:"challenge_id"`
	KID                   string         `json:"kid"`
	ClientID              string         `json:"client_id"`
	ExcludeNotBeforeClaim bool           `json:"exclude_nbf_claim"`
	AllowedTopLevelClaims []string       `json:"allowed_top_level_claims"`
	Extra                 map[string]any `json:"extra"`
}

// GetChallengeID returns the challenge id.
func (s *Session) GetChallengeID() uuid.NullUUID {
	return s.ChallengeID
}

// GetIDTokenClaims returns the *jwt.IDTokenClaims for this session.
func (s *Session) GetIDTokenClaims() *jwt.IDTokenClaims {
	if s.DefaultSession == nil {
		return nil
	}

	return s.DefaultSession.Claims
}

// GetExtraClaims returns the Extra/Unregistered claims for this session.
func (s *Session) GetExtraClaims() map[string]any {
	if s.DefaultSession != nil && s.DefaultSession.Claims != nil {
		return s.DefaultSession.Claims.Extra
	}

	return s.Extra
}

// Clone copies the OpenIDSession to a new fosite.Session.
func (s *Session) Clone() fosite.Session {
	if s == nil {
		return nil
	}

	return deepcopy.Copy(s).(fosite.Session)
}
