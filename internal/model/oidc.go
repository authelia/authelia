package model

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
)

// NewOAuth2ConsentSession creates a new OAuth2ConsentSession.
func NewOAuth2ConsentSession(subject uuid.UUID, r fosite.Requester) (consent *OAuth2ConsentSession, err error) {
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
		openidSession *OpenIDSession
		sessData      []byte
	)

	openidSession = r.GetSession().(*OpenIDSession)
	if openidSession == nil {
		return nil, errors.New("unexpected session type")
	}

	subject = openidSession.GetSubject()

	if sessData, err = json.Marshal(openidSession); err != nil {
		return nil, err
	}

	return &OAuth2Session{
		ChallengeID:       openidSession.ChallengeID,
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
		Session:           sessData,
	}, nil
}

// NewOAuth2BlacklistedJTI creates a new OAuth2BlacklistedJTI.
func NewOAuth2BlacklistedJTI(jti string, exp time.Time) (jtiBlacklist *OAuth2BlacklistedJTI) {
	return &OAuth2BlacklistedJTI{
		Signature: fmt.Sprintf("%x", sha256.Sum256([]byte(jti))),
		ExpiresAt: exp,
	}
}

// OAuth2ConsentSession stores information about an OAuth2.0 Consent.
type OAuth2ConsentSession struct {
	ID          int       `db:"id"`
	ChallengeID uuid.UUID `db:"challenge_id"`
	ClientID    string    `db:"client_id"`
	Subject     uuid.UUID `db:"subject"`

	Authorized bool `db:"authorized"`
	Granted    bool `db:"granted"`

	RequestedAt time.Time  `db:"requested_at"`
	RespondedAt *time.Time `db:"responded_at"`

	Form string `db:"form_data"`

	RequestedScopes   StringSlicePipeDelimited `db:"requested_scopes"`
	GrantedScopes     StringSlicePipeDelimited `db:"granted_scopes"`
	RequestedAudience StringSlicePipeDelimited `db:"requested_audience"`
	GrantedAudience   StringSlicePipeDelimited `db:"granted_audience"`
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

// OpenIDSession holds OIDC Session information.
type OpenIDSession struct {
	*openid.DefaultSession `json:"idToken"`

	ChallengeID uuid.UUID `db:"challenge_id"`
	ClientID    string

	Extra map[string]interface{} `json:"extra"`
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
