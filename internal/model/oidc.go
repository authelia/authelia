package model

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/mohae/deepcopy"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"

	"github.com/authelia/authelia/v4/internal/utils"
)

// NewOAuth2ConsentSession creates a new OAuth2ConsentSession.
func NewOAuth2ConsentSession(subject NullUUID, r fosite.Requester) (consent *OAuth2ConsentSession, err error) {
	consent = &OAuth2ConsentSession{
		ClientID:          r.GetClient().GetID(),
		Subject:           subject,
		Form:              r.GetRequestForm().Encode(),
		RequestedAt:       r.GetRequestedAt(),
		RequestedScopes:   StringSlicePipeDelimited(r.GetRequestedScopes()),
		RequestedAudience: StringSlicePipeDelimited(r.GetRequestedAudience()),
		GrantedScopes:     StringSlicePipeDelimited(r.GetGrantedScopes()),
		GrantedAudience:   StringSlicePipeDelimited(r.GetGrantedAudience()),
	}

	if consent.ChallengeID, err = uuid.NewRandom(); err != nil {
		return nil, err
	}

	return consent, nil
}

// NewOAuth2SessionFromRequest creates a new OAuth2Session from a signature and fosite.Requester.
func NewOAuth2SessionFromRequest(signature string, r fosite.Requester) (session *OAuth2Session, err error) {
	var (
		subject       string
		sessionOpenID *OpenIDSession
		ok            bool
		sessionData   []byte
	)

	sessionOpenID, ok = r.GetSession().(*OpenIDSession)
	if !ok {
		return nil, fmt.Errorf("can't convert type '%T' to an *OAuth2Session", r.GetSession())
	}

	subject = sessionOpenID.GetSubject()

	if sessionData, err = json.Marshal(sessionOpenID); err != nil {
		return nil, err
	}

	return &OAuth2Session{
		ChallengeID:       sessionOpenID.ChallengeID,
		RequestID:         r.GetID(),
		ClientID:          r.GetClient().GetID(),
		Signature:         signature,
		RequestedAt:       r.GetRequestedAt(),
		Subject:           subject,
		RequestedScopes:   StringSlicePipeDelimited(r.GetRequestedScopes()),
		GrantedScopes:     StringSlicePipeDelimited(r.GetGrantedScopes()),
		RequestedAudience: StringSlicePipeDelimited(r.GetRequestedAudience()),
		GrantedAudience:   StringSlicePipeDelimited(r.GetGrantedAudience()),
		Active:            true,
		Revoked:           false,
		Form:              r.GetRequestForm().Encode(),
		Session:           sessionData,
	}, nil
}

// NewOAuth2BlacklistedJTI creates a new OAuth2BlacklistedJTI.
func NewOAuth2BlacklistedJTI(jti string, exp time.Time) (jtiBlacklist OAuth2BlacklistedJTI) {
	return OAuth2BlacklistedJTI{
		Signature: fmt.Sprintf("%x", sha256.Sum256([]byte(jti))),
		ExpiresAt: exp,
	}
}

// OAuth2ConsentSession stores information about an OAuth2.0 Consent.
type OAuth2ConsentSession struct {
	ID          int       `db:"id"`
	ChallengeID uuid.UUID `db:"challenge_id"`
	ClientID    string    `db:"client_id"`
	Subject     NullUUID  `db:"subject"`

	Authorized bool `db:"authorized"`
	Granted    bool `db:"granted"`

	RequestedAt time.Time  `db:"requested_at"`
	RespondedAt *time.Time `db:"responded_at"`
	ExpiresAt   *time.Time `db:"expires_at"`

	Form string `db:"form_data"`

	RequestedScopes   StringSlicePipeDelimited `db:"requested_scopes"`
	GrantedScopes     StringSlicePipeDelimited `db:"granted_scopes"`
	RequestedAudience StringSlicePipeDelimited `db:"requested_audience"`
	GrantedAudience   StringSlicePipeDelimited `db:"granted_audience"`
}

// HasExactGrants returns true if the granted audience and scopes of this consent matches exactly with another
// audience and set of scopes.
func (s OAuth2ConsentSession) HasExactGrants(scopes, audience []string) (has bool) {
	return s.HasExactGrantedScopes(scopes) && s.HasExactGrantedAudience(audience)
}

// HasExactGrantedAudience returns true if the granted audience of this consent matches exactly with another audience.
func (s OAuth2ConsentSession) HasExactGrantedAudience(audience []string) (has bool) {
	return !utils.IsStringSlicesDifferent(s.GrantedAudience, audience)
}

// HasExactGrantedScopes returns true if the granted scopes of this consent matches exactly with another set of scopes.
func (s OAuth2ConsentSession) HasExactGrantedScopes(scopes []string) (has bool) {
	return !utils.IsStringSlicesDifferent(s.GrantedScopes, scopes)
}

// IsAuthorized returns true if the user has responded to the consent session and it was authorized.
func (s OAuth2ConsentSession) IsAuthorized() bool {
	return s.Responded() && s.Authorized
}

// CanGrant returns true if the user has responded to the consent session, it was authorized, and it either hast not
// previously been granted or the ability to grant has not expired.
func (s OAuth2ConsentSession) CanGrant() bool {
	if !s.Responded() {
		return false
	}

	if s.Granted && (s.ExpiresAt == nil || s.ExpiresAt.Before(time.Now())) {
		return false
	}

	return true
}

// IsDenied returns true if the user has responded to the consent session and it was not authorized.
func (s OAuth2ConsentSession) IsDenied() bool {
	return s.Responded() && !s.Authorized
}

// Responded returns true if the user has responded to the consent session.
func (s OAuth2ConsentSession) Responded() bool {
	return s.RespondedAt != nil
}

// GetForm returns the form.
func (s OAuth2ConsentSession) GetForm() (form url.Values, err error) {
	return url.ParseQuery(s.Form)
}

// OAuth2BlacklistedJTI represents a blacklisted JTI used with OAuth2.0.
type OAuth2BlacklistedJTI struct {
	ID        int       `db:"id"`
	Signature string    `db:"signature"`
	ExpiresAt time.Time `db:"expires_at"`
}

// OAuth2Session represents a OAuth2.0 session.
type OAuth2Session struct {
	ID                int                      `db:"id"`
	ChallengeID       uuid.UUID                `db:"challenge_id"`
	RequestID         string                   `db:"request_id"`
	ClientID          string                   `db:"client_id"`
	Signature         string                   `db:"signature"`
	RequestedAt       time.Time                `db:"requested_at"`
	Subject           string                   `db:"subject"`
	RequestedScopes   StringSlicePipeDelimited `db:"requested_scopes"`
	GrantedScopes     StringSlicePipeDelimited `db:"granted_scopes"`
	RequestedAudience StringSlicePipeDelimited `db:"requested_audience"`
	GrantedAudience   StringSlicePipeDelimited `db:"granted_audience"`
	Active            bool                     `db:"active"`
	Revoked           bool                     `db:"revoked"`
	Form              string                   `db:"form_data"`
	Session           []byte                   `db:"session_data"`
}

// SetSubject implements an interface required for RFC7523.
func (s *OAuth2Session) SetSubject(subject string) {
	s.Subject = subject
}

// ToRequest converts an OAuth2Session into a fosite.Request given a fosite.Session and fosite.Storage.
func (s OAuth2Session) ToRequest(ctx context.Context, session fosite.Session, store fosite.Storage) (request *fosite.Request, err error) {
	sessionData := s.Session

	if session != nil {
		if err = json.Unmarshal(sessionData, session); err != nil {
			return nil, err
		}
	}

	client, err := store.GetClient(ctx, s.ClientID)
	if err != nil {
		return nil, err
	}

	values, err := url.ParseQuery(s.Form)
	if err != nil {
		return nil, err
	}

	return &fosite.Request{
		ID:                s.RequestID,
		RequestedAt:       s.RequestedAt,
		Client:            client,
		RequestedScope:    fosite.Arguments(s.RequestedScopes),
		GrantedScope:      fosite.Arguments(s.GrantedScopes),
		RequestedAudience: fosite.Arguments(s.RequestedAudience),
		GrantedAudience:   fosite.Arguments(s.GrantedAudience),
		Form:              values,
		Session:           session,
	}, nil
}

// OpenIDSession holds OIDC Session information.
type OpenIDSession struct {
	*openid.DefaultSession `json:"id_token"`

	ChallengeID uuid.UUID `db:"challenge_id"`
	ClientID    string

	Extra map[string]interface{} `json:"extra"`
}

// Clone copies the OpenIDSession to a new fosite.Session.
func (s *OpenIDSession) Clone() fosite.Session {
	if s == nil {
		return nil
	}

	return deepcopy.Copy(s).(fosite.Session)
}
