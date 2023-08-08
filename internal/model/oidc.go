package model

import (
	"context"
	"crypto/sha256"
	"database/sql"
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
func NewOAuth2ConsentSession(subject uuid.UUID, r fosite.Requester) (consent *OAuth2ConsentSession, err error) {
	consent = &OAuth2ConsentSession{
		ClientID:          r.GetClient().GetID(),
		Subject:           NullUUID(subject),
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

// NewOAuth2BlacklistedJTI creates a new OAuth2BlacklistedJTI.
func NewOAuth2BlacklistedJTI(jti string, exp time.Time) (jtiBlacklist OAuth2BlacklistedJTI) {
	return OAuth2BlacklistedJTI{
		Signature: fmt.Sprintf("%x", sha256.Sum256([]byte(jti))),
		ExpiresAt: exp,
	}
}

// NewOAuth2SessionFromRequest creates a new OAuth2Session from a signature and fosite.Requester.
func NewOAuth2SessionFromRequest(signature string, r fosite.Requester) (session *OAuth2Session, err error) {
	var (
		subject       sql.NullString
		sessionOpenID *OpenIDSession
		ok            bool
		sessionData   []byte
	)

	sessionOpenID, ok = r.GetSession().(*OpenIDSession)
	if !ok {
		return nil, fmt.Errorf("can't convert type '%T' to an *OAuth2Session", r.GetSession())
	}

	subject = sql.NullString{String: sessionOpenID.GetSubject()}

	subject.Valid = len(subject.String) > 0

	if sessionData, err = json.Marshal(sessionOpenID); err != nil {
		return nil, err
	}

	requested, granted := r.GetRequestedScopes(), r.GetGrantedScopes()

	if requested == nil {
		requested = fosite.Arguments{}
	}

	if granted == nil {
		granted = fosite.Arguments{}
	}

	return &OAuth2Session{
		ChallengeID:       sessionOpenID.ChallengeID,
		RequestID:         r.GetID(),
		ClientID:          r.GetClient().GetID(),
		Signature:         signature,
		RequestedAt:       r.GetRequestedAt(),
		Subject:           subject,
		RequestedScopes:   StringSlicePipeDelimited(requested),
		GrantedScopes:     StringSlicePipeDelimited(granted),
		RequestedAudience: StringSlicePipeDelimited(r.GetRequestedAudience()),
		GrantedAudience:   StringSlicePipeDelimited(r.GetGrantedAudience()),
		Active:            true,
		Revoked:           false,
		Form:              r.GetRequestForm().Encode(),
		Session:           sessionData,
	}, nil
}

// NewOAuth2PARContext creates a new Pushed Authorization Request Context as a OAuth2PARContext.
func NewOAuth2PARContext(contextID string, r fosite.AuthorizeRequester) (context *OAuth2PARContext, err error) {
	var (
		s       *OpenIDSession
		ok      bool
		req     *fosite.AuthorizeRequest
		session []byte
	)

	if s, ok = r.GetSession().(*OpenIDSession); !ok {
		return nil, fmt.Errorf("can't convert type '%T' to an *OAuth2Session", r.GetSession())
	}

	if session, err = json.Marshal(s); err != nil {
		return nil, err
	}

	var handled StringSlicePipeDelimited

	if req, ok = r.(*fosite.AuthorizeRequest); ok {
		handled = StringSlicePipeDelimited(req.HandledResponseTypes)
	}

	return &OAuth2PARContext{
		Signature:            contextID,
		RequestID:            r.GetID(),
		ClientID:             r.GetClient().GetID(),
		RequestedAt:          r.GetRequestedAt(),
		Scopes:               StringSlicePipeDelimited(r.GetRequestedScopes()),
		Audience:             StringSlicePipeDelimited(r.GetRequestedAudience()),
		HandledResponseTypes: handled,
		ResponseMode:         string(r.GetResponseMode()),
		DefaultResponseMode:  string(r.GetDefaultResponseMode()),
		Revoked:              false,
		Form:                 r.GetRequestForm().Encode(),
		Session:              session,
	}, nil
}

// OAuth2ConsentPreConfig stores information about an OAuth2.0 Pre-Configured Consent.
type OAuth2ConsentPreConfig struct {
	ID       int64     `db:"id"`
	ClientID string    `db:"client_id"`
	Subject  uuid.UUID `db:"subject"`

	CreatedAt time.Time    `db:"created_at"`
	ExpiresAt sql.NullTime `db:"expires_at"`

	Revoked bool `db:"revoked"`

	Scopes   StringSlicePipeDelimited `db:"scopes"`
	Audience StringSlicePipeDelimited `db:"audience"`
}

// HasExactGrants returns true if the granted audience and scopes of this consent pre-configuration matches exactly with
// another audience and set of scopes.
func (s *OAuth2ConsentPreConfig) HasExactGrants(scopes, audience []string) (has bool) {
	return s.HasExactGrantedScopes(scopes) && s.HasExactGrantedAudience(audience)
}

// HasExactGrantedAudience returns true if the granted audience of this consent matches exactly with another audience.
func (s *OAuth2ConsentPreConfig) HasExactGrantedAudience(audience []string) (has bool) {
	return !utils.IsStringSlicesDifferent(s.Audience, audience)
}

// HasExactGrantedScopes returns true if the granted scopes of this consent matches exactly with another set of scopes.
func (s *OAuth2ConsentPreConfig) HasExactGrantedScopes(scopes []string) (has bool) {
	return !utils.IsStringSlicesDifferent(s.Scopes, scopes)
}

// CanConsent returns true if this pre-configuration can still provide consent.
func (s *OAuth2ConsentPreConfig) CanConsent() bool {
	return !s.Revoked && (!s.ExpiresAt.Valid || s.ExpiresAt.Time.After(time.Now()))
}

// OAuth2ConsentSession stores information about an OAuth2.0 Consent.
type OAuth2ConsentSession struct {
	ID          int           `db:"id"`
	ChallengeID uuid.UUID     `db:"challenge_id"`
	ClientID    string        `db:"client_id"`
	Subject     uuid.NullUUID `db:"subject"`

	Authorized bool `db:"authorized"`
	Granted    bool `db:"granted"`

	RequestedAt time.Time    `db:"requested_at"`
	RespondedAt sql.NullTime `db:"responded_at"`

	Form string `db:"form_data"`

	RequestedScopes   StringSlicePipeDelimited `db:"requested_scopes"`
	GrantedScopes     StringSlicePipeDelimited `db:"granted_scopes"`
	RequestedAudience StringSlicePipeDelimited `db:"requested_audience"`
	GrantedAudience   StringSlicePipeDelimited `db:"granted_audience"`

	PreConfiguration sql.NullInt64
}

// Grant grants the requested scopes and audience.
func (s *OAuth2ConsentSession) Grant() {
	s.GrantedScopes = s.RequestedScopes
	s.GrantedAudience = s.RequestedAudience

	if !utils.IsStringInSlice(s.ClientID, s.GrantedAudience) {
		s.GrantedAudience = append(s.GrantedAudience, s.ClientID)
	}
}

// HasExactGrants returns true if the granted audience and scopes of this consent matches exactly with another
// audience and set of scopes.
func (s *OAuth2ConsentSession) HasExactGrants(scopes, audience []string) (has bool) {
	return s.HasExactGrantedScopes(scopes) && s.HasExactGrantedAudience(audience)
}

// HasExactGrantedAudience returns true if the granted audience of this consent matches exactly with another audience.
func (s *OAuth2ConsentSession) HasExactGrantedAudience(audience []string) (has bool) {
	return !utils.IsStringSlicesDifferent(s.GrantedAudience, audience)
}

// HasExactGrantedScopes returns true if the granted scopes of this consent matches exactly with another set of scopes.
func (s *OAuth2ConsentSession) HasExactGrantedScopes(scopes []string) (has bool) {
	return !utils.IsStringSlicesDifferent(s.GrantedScopes, scopes)
}

// Responded returns true if the user has responded to the consent session.
func (s *OAuth2ConsentSession) Responded() bool {
	return s.RespondedAt.Valid
}

// IsAuthorized returns true if the user has responded to the consent session and it was authorized.
func (s *OAuth2ConsentSession) IsAuthorized() bool {
	return s.Responded() && s.Authorized
}

// IsDenied returns true if the user has responded to the consent session and it was not authorized.
func (s *OAuth2ConsentSession) IsDenied() bool {
	return s.Responded() && !s.Authorized
}

// CanGrant returns true if the session can still grant a token. This is NOT indicative of if there is a user response
// to this consent request or if the user rejected the consent request.
func (s *OAuth2ConsentSession) CanGrant() bool {
	if !s.Subject.Valid || s.Granted {
		return false
	}

	return true
}

// GetForm returns the form.
func (s *OAuth2ConsentSession) GetForm() (form url.Values, err error) {
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
	ChallengeID       uuid.NullUUID            `db:"challenge_id"`
	RequestID         string                   `db:"request_id"`
	ClientID          string                   `db:"client_id"`
	Signature         string                   `db:"signature"`
	RequestedAt       time.Time                `db:"requested_at"`
	Subject           sql.NullString           `db:"subject"`
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
	s.Subject = sql.NullString{String: subject, Valid: len(subject) > 0}
}

// ToRequest converts an OAuth2Session into a fosite.Request given a fosite.Session and fosite.Storage.
func (s *OAuth2Session) ToRequest(ctx context.Context, session fosite.Session, store fosite.Storage) (request *fosite.Request, err error) {
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

// OAuth2PARContext holds relevant information about a Pushed Authorization Request in order to process the authorization.
type OAuth2PARContext struct {
	ID                   int                      `db:"id"`
	Signature            string                   `db:"signature"`
	RequestID            string                   `db:"request_id"`
	ClientID             string                   `db:"client_id"`
	RequestedAt          time.Time                `db:"requested_at"`
	Scopes               StringSlicePipeDelimited `db:"scopes"`
	Audience             StringSlicePipeDelimited `db:"audience"`
	HandledResponseTypes StringSlicePipeDelimited `db:"handled_response_types"`
	ResponseMode         string                   `db:"response_mode"`
	DefaultResponseMode  string                   `db:"response_mode_default"`
	Revoked              bool                     `db:"revoked"`
	Form                 string                   `db:"form_data"`
	Session              []byte                   `db:"session_data"`
}

func (par *OAuth2PARContext) ToAuthorizeRequest(ctx context.Context, session fosite.Session, store fosite.Storage) (request *fosite.AuthorizeRequest, err error) {
	if session != nil {
		if err = json.Unmarshal(par.Session, session); err != nil {
			return nil, err
		}
	}

	var (
		client fosite.Client
		form   url.Values
	)

	if client, err = store.GetClient(ctx, par.ClientID); err != nil {
		return nil, err
	}

	if form, err = url.ParseQuery(par.Form); err != nil {
		return nil, err
	}

	request = fosite.NewAuthorizeRequest()

	request.Request = fosite.Request{
		ID:                par.RequestID,
		RequestedAt:       par.RequestedAt,
		Client:            client,
		RequestedScope:    fosite.Arguments(par.Scopes),
		RequestedAudience: fosite.Arguments(par.Audience),
		Form:              form,
		Session:           session,
	}

	if par.ResponseMode != "" {
		request.ResponseMode = fosite.ResponseModeType(par.ResponseMode)
	}

	if par.DefaultResponseMode != "" {
		request.DefaultResponseMode = fosite.ResponseModeType(par.DefaultResponseMode)
	}

	if len(par.HandledResponseTypes) != 0 {
		request.HandledResponseTypes = fosite.Arguments(par.HandledResponseTypes)
	}

	return request, nil
}

// OpenIDSession holds OIDC Session information.
type OpenIDSession struct {
	*openid.DefaultSession `json:"id_token"`

	ChallengeID uuid.NullUUID `db:"challenge_id"`
	ClientID    string

	Extra map[string]any `json:"extra"`
}

// Clone copies the OpenIDSession to a new fosite.Session.
func (s *OpenIDSession) Clone() fosite.Session {
	if s == nil {
		return nil
	}

	return deepcopy.Copy(s).(fosite.Session)
}
