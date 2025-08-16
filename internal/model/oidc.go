package model

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/utils"
)

// NewOAuth2ConsentSession creates a new OAuth2ConsentSession.
func NewOAuth2ConsentSession(expires time.Time, subject uuid.UUID, r oauthelia2.Requester) (consent *OAuth2ConsentSession, err error) {
	return NewOAuth2ConsentSessionWithForm(expires, subject, r, r.GetRequestForm())
}

// NewOAuth2ConsentSessionWithForm creates a new OAuth2ConsentSession with a custom form parameter.
func NewOAuth2ConsentSessionWithForm(expires time.Time, subject uuid.UUID, r oauthelia2.Requester, form url.Values) (consent *OAuth2ConsentSession, err error) {
	consent = &OAuth2ConsentSession{
		ClientID:          r.GetClient().GetID(),
		Subject:           NullUUID(subject),
		Form:              form.Encode(),
		RequestedAt:       r.GetRequestedAt(),
		ExpiresAt:         expires,
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

// NewOAuth2SessionFromRequest creates a new OAuth2Session from a signature and oauthelia2.Requester.
func NewOAuth2SessionFromRequest(signature string, r oauthelia2.Requester) (session *OAuth2Session, err error) {
	if r == nil {
		return nil, fmt.Errorf("failed to create new *model.OAuth2Session: the oauthelia2.Requester was nil")
	}

	var (
		subject     sql.NullString
		s           OpenIDSession
		ok          bool
		sessionData []byte
	)

	s, ok = r.GetSession().(OpenIDSession)
	if !ok {
		return nil, fmt.Errorf("failed to create new *model.OAuth2Session: the session type OpenIDSession was expected but the type '%T' was used", r.GetSession())
	}

	subject = sql.NullString{String: s.GetSubject()}

	subject.Valid = len(subject.String) > 0

	if sessionData, err = json.Marshal(s); err != nil {
		return nil, fmt.Errorf("failed to create new *model.OAuth2Session: an error was returned while attempting to marshal the session data to json: %w", err)
	}

	requested, granted := r.GetRequestedScopes(), r.GetGrantedScopes()

	if requested == nil {
		requested = oauthelia2.Arguments{}
	}

	if granted == nil {
		granted = oauthelia2.Arguments{}
	}

	return &OAuth2Session{
		ChallengeID:       s.GetChallengeID(),
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

// NewOAuth2DeviceCodeSessionFromRequest creates a new OAuth2DeviceCodeSession from a signature and oauthelia2.Requester.
func NewOAuth2DeviceCodeSessionFromRequest(r oauthelia2.DeviceAuthorizeRequester) (session *OAuth2DeviceCodeSession, err error) {
	if r == nil {
		return nil, fmt.Errorf("failed to create new *model.OAuth2DeviceCodeSession: the oauthelia2.DeviceAuthorizeRequester was nil")
	}

	var (
		subject     sql.NullString
		s           OpenIDSession
		ok          bool
		sessionData []byte
	)

	s, ok = r.GetSession().(OpenIDSession)
	if !ok {
		return nil, fmt.Errorf("failed to create new *model.OAuth2DeviceCodeSession: the session type OpenIDSession was expected but the type '%T' was used", r.GetSession())
	}

	subject = sql.NullString{String: s.GetSubject()}

	subject.Valid = len(subject.String) > 0

	if sessionData, err = json.Marshal(s); err != nil {
		return nil, fmt.Errorf("failed to create new *model.OAuth2DeviceCodeSession: an error was returned while attempting to marshal the session data to json: %w", err)
	}

	requested, granted := r.GetRequestedScopes(), r.GetGrantedScopes()

	if requested == nil {
		requested = oauthelia2.Arguments{}
	}

	if granted == nil {
		granted = oauthelia2.Arguments{}
	}

	return &OAuth2DeviceCodeSession{
		ChallengeID:       s.GetChallengeID(),
		RequestID:         r.GetID(),
		ClientID:          r.GetClient().GetID(),
		Signature:         r.GetDeviceCodeSignature(),
		UserCodeSignature: r.GetUserCodeSignature(),
		Status:            int(r.GetStatus()),
		Subject:           subject,
		RequestedAt:       r.GetRequestedAt(),
		CheckedAt:         r.GetLastChecked(),
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
func NewOAuth2PARContext(contextID string, r oauthelia2.AuthorizeRequester) (context *OAuth2PARContext, err error) {
	var (
		s       OpenIDSession
		ok      bool
		req     *oauthelia2.AuthorizeRequest
		session []byte
	)

	if s, ok = r.GetSession().(OpenIDSession); !ok {
		return nil, fmt.Errorf("failed to create new PAR context: can't assert type '%T' to an *OAuth2Session", r.GetSession())
	}

	if session, err = json.Marshal(s); err != nil {
		return nil, err
	}

	var handled StringSlicePipeDelimited

	if req, ok = r.(*oauthelia2.AuthorizeRequest); ok {
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

	RequestedClaims sql.NullString           `db:"requested_claims"`
	SignatureClaims sql.NullString           `db:"signature_claims"`
	GrantedClaims   StringSlicePipeDelimited `db:"granted_claims"`
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

// HasClaimsSignature returns true if the requested claims signature of this consent matches exactly with another request.
func (s *OAuth2ConsentPreConfig) HasClaimsSignature(signature string) (has bool) {
	return (s.SignatureClaims.Valid || len(signature) == 0) && strings.EqualFold(signature, s.SignatureClaims.String)
}

// CanConsent returns true if this pre-configuration can still provide consent.
func (s *OAuth2ConsentPreConfig) CanConsent() bool {
	return s.CanConsentAt(time.Now())
}

// CanConsentAt returns true if this pre-configuration can still provide consent at a particular time.
func (s *OAuth2ConsentPreConfig) CanConsentAt(now time.Time) bool {
	return !s.Revoked && (!s.ExpiresAt.Valid || s.ExpiresAt.Time.After(now))
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
	ExpiresAt   time.Time    `db:"expires_at"`
	RespondedAt sql.NullTime `db:"responded_at"`

	Form string `db:"form_data"`

	RequestedScopes   StringSlicePipeDelimited `db:"requested_scopes"`
	GrantedScopes     StringSlicePipeDelimited `db:"granted_scopes"`
	RequestedAudience StringSlicePipeDelimited `db:"requested_audience"`
	GrantedAudience   StringSlicePipeDelimited `db:"granted_audience"`
	GrantedClaims     StringSlicePipeDelimited `db:"granted_claims"`

	PreConfiguration sql.NullInt64
}

// GetRequestedAt returns the requested at value.
func (s *OAuth2ConsentSession) GetRequestedAt() time.Time {
	return s.RequestedAt
}

// SetSubject sets the subject value.
func (s *OAuth2ConsentSession) SetSubject(subject uuid.UUID) {
	s.Subject = uuid.NullUUID{UUID: subject, Valid: subject != uuid.Nil}
}

// SetRespondedAt sets the responded at value.
func (s *OAuth2ConsentSession) SetRespondedAt(t time.Time, preconf int64) {
	s.RespondedAt = sql.NullTime{Time: t, Valid: true}

	if preconf > 0 {
		s.PreConfiguration = sql.NullInt64{Int64: preconf, Valid: true}
	}
}

// GrantScopes grants all of the requested scopes.
func (s *OAuth2ConsentSession) GrantScopes() {
	s.GrantedScopes = s.RequestedScopes
}

// GrantScope grants the specified scope.
func (s *OAuth2ConsentSession) GrantScope(scope string) {
	s.GrantedScopes = append(s.GrantedScopes, scope)
}

// GrantClaims grants the specified claims.
func (s *OAuth2ConsentSession) GrantClaims(claims []string) {
	if len(claims) == 0 {
		return
	}

	s.GrantedClaims = claims
}

// GrantAudience grants all of the requested audiences.
func (s *OAuth2ConsentSession) GrantAudience() {
	s.GrantedAudience = s.RequestedAudience
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
func (s *OAuth2ConsentSession) CanGrant(now time.Time) bool {
	return !s.Granted && now.Before(s.ExpiresAt)
}

// GetForm returns the form.
func (s *OAuth2ConsentSession) GetForm() (form url.Values, err error) {
	return url.ParseQuery(s.Form)
}

func (s *OAuth2ConsentSession) GetRequestedScopes() []string {
	return s.RequestedScopes
}

func (s *OAuth2ConsentSession) GetGrantedScopes() []string {
	return s.GrantedScopes
}

func (s *OAuth2ConsentSession) GetRequestedAudience() []string {
	return s.RequestedAudience
}

func (s *OAuth2ConsentSession) GetGrantedAudience() []string {
	return s.GrantedAudience
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

// ToRequest converts an OAuth2Session into a oauthelia2.Request given a oauthelia2.Session and oauthelia2.Storage.
func (s *OAuth2Session) ToRequest(ctx context.Context, session oauthelia2.Session, store oauthelia2.Storage) (request *oauthelia2.Request, err error) {
	sessionData := s.Session

	if session != nil {
		if err = json.Unmarshal(sessionData, session); err != nil {
			return nil, fmt.Errorf("error occurred while mapping OAuth 2.0 Session back to a Request while trying to unmarshal the JSON session data: %w", err)
		}
	}

	client, err := store.GetClient(ctx, s.ClientID)
	if err != nil {
		return nil, fmt.Errorf("error occurred while mapping OAuth 2.0 Session back to a Request while trying to lookup the registered client: %w", err)
	}

	values, err := url.ParseQuery(s.Form)
	if err != nil {
		return nil, fmt.Errorf("error occurred while mapping OAuth 2.0 Session back to a Request while trying to parse the original form: %w", err)
	}

	return &oauthelia2.Request{
		ID:                s.RequestID,
		RequestedAt:       s.RequestedAt,
		Client:            client,
		RequestedScope:    oauthelia2.Arguments(s.RequestedScopes),
		GrantedScope:      oauthelia2.Arguments(s.GrantedScopes),
		RequestedAudience: oauthelia2.Arguments(s.RequestedAudience),
		GrantedAudience:   oauthelia2.Arguments(s.GrantedAudience),
		Form:              values,
		Session:           session,
	}, nil
}

// OAuth2DeviceCodeSession stores the Device Code Grant information.
type OAuth2DeviceCodeSession struct {
	ID                int                      `db:"id"`
	ChallengeID       uuid.NullUUID            `db:"challenge_id"`
	RequestID         string                   `db:"request_id"`
	ClientID          string                   `db:"client_id"`
	Signature         string                   `db:"signature"`
	UserCodeSignature string                   `db:"user_code_signature"`
	Status            int                      `db:"status"`
	Subject           sql.NullString           `db:"subject"`
	RequestedAt       time.Time                `db:"requested_at"`
	CheckedAt         time.Time                `db:"checked_at"`
	RequestedScopes   StringSlicePipeDelimited `db:"requested_scopes"`
	GrantedScopes     StringSlicePipeDelimited `db:"granted_scopes"`
	RequestedAudience StringSlicePipeDelimited `db:"requested_audience"`
	GrantedAudience   StringSlicePipeDelimited `db:"granted_audience"`
	Active            bool                     `db:"active"`
	Revoked           bool                     `db:"revoked"`
	Form              string                   `db:"form_data"`
	Session           []byte                   `db:"session_data"`
}

// GetRequestedAt returns the requested at value.
func (s *OAuth2DeviceCodeSession) GetRequestedAt() time.Time {
	return s.RequestedAt
}

// GetForm returns the form.
func (s *OAuth2DeviceCodeSession) GetForm() (form url.Values, err error) {
	return url.ParseQuery(s.Form)
}

func (s *OAuth2DeviceCodeSession) GetRequestedScopes() []string {
	return s.RequestedScopes
}

func (s *OAuth2DeviceCodeSession) GetGrantedScopes() []string {
	return s.GrantedScopes
}

func (s *OAuth2DeviceCodeSession) GetRequestedAudience() []string {
	return s.RequestedAudience
}

func (s *OAuth2DeviceCodeSession) GetGrantedAudience() []string {
	return s.GrantedAudience
}

// ToRequest converts an OAuth2Session into a oauthelia2.Request given an oauthelia2.Session and oauthelia2.Storage.
func (s *OAuth2DeviceCodeSession) ToRequest(ctx context.Context, session oauthelia2.Session, store oauthelia2.Storage) (request *oauthelia2.DeviceAuthorizeRequest, err error) {
	sessionData := s.Session

	if session != nil {
		if err = json.Unmarshal(sessionData, session); err != nil {
			return nil, fmt.Errorf("error occurred while mapping OAuth 2.0 Session back to a DeviceAuthorizeRequest while trying to unmarshal the JSON session data: %w", err)
		}
	}

	client, err := store.GetClient(ctx, s.ClientID)
	if err != nil {
		return nil, fmt.Errorf("error occurred while mapping OAuth 2.0 Session back to a DeviceAuthorizeRequest while trying to lookup the registered client: %w", err)
	}

	values, err := url.ParseQuery(s.Form)
	if err != nil {
		return nil, fmt.Errorf("error occurred while mapping OAuth 2.0 Session back to a DeviceAuthorizeRequest while trying to parse the original form: %w", err)
	}

	request = &oauthelia2.DeviceAuthorizeRequest{
		Request: oauthelia2.Request{
			ID:                s.RequestID,
			RequestedAt:       s.RequestedAt,
			Client:            client,
			RequestedScope:    oauthelia2.Arguments(s.RequestedScopes),
			GrantedScope:      oauthelia2.Arguments(s.GrantedScopes),
			RequestedAudience: oauthelia2.Arguments(s.RequestedAudience),
			GrantedAudience:   oauthelia2.Arguments(s.GrantedAudience),
			Form:              values,
			Session:           session,
		},
		DeviceCodeSignature: s.Signature,
		UserCodeSignature:   s.UserCodeSignature,
		Status:              oauthelia2.DeviceAuthorizeStatus(s.Status),
		LastChecked:         s.CheckedAt,
	}

	return request, nil
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

func (par *OAuth2PARContext) ToAuthorizeRequest(ctx context.Context, session oauthelia2.Session, store oauthelia2.Storage) (request *oauthelia2.AuthorizeRequest, err error) {
	if session != nil {
		if err = json.Unmarshal(par.Session, session); err != nil {
			return nil, fmt.Errorf("error occurred while mapping PAR context back to an Authorize Request while trying to unmarshal the JSON session data: %w", err)
		}
	}

	var (
		client oauthelia2.Client
		form   url.Values
	)

	if client, err = store.GetClient(ctx, par.ClientID); err != nil {
		return nil, fmt.Errorf("error occurred while mapping PAR context back to an Authorize Request while trying to lookup the registered client: %w", err)
	}

	if form, err = url.ParseQuery(par.Form); err != nil {
		return nil, fmt.Errorf("error occurred while mapping PAR context back to an Authorize Request while trying to parse the original form: %w", err)
	}

	request = oauthelia2.NewAuthorizeRequest()

	request.Request = oauthelia2.Request{
		ID:                par.RequestID,
		RequestedAt:       par.RequestedAt,
		Client:            client,
		RequestedScope:    oauthelia2.Arguments(par.Scopes),
		RequestedAudience: oauthelia2.Arguments(par.Audience),
		Form:              form,
		Session:           session,
	}

	request.State = form.Get("state")

	if form.Has("redirect_uri") {
		if request.RedirectURI, err = url.Parse(form.Get("redirect_uri")); err != nil {
			return nil, fmt.Errorf("error occurred while mapping PAR context back to an Authorize Request while trying to parse the original redirect uri: %w", err)
		}
	}

	if form.Has("response_type") {
		request.ResponseTypes = oauthelia2.RemoveEmpty(strings.Split(form.Get("response_type"), " "))
	}

	if par.ResponseMode != "" {
		request.ResponseMode = oauthelia2.ResponseModeType(par.ResponseMode)
	}

	if par.DefaultResponseMode != "" {
		request.DefaultResponseMode = oauthelia2.ResponseModeType(par.DefaultResponseMode)
	}

	if len(par.HandledResponseTypes) != 0 {
		request.HandledResponseTypes = oauthelia2.Arguments(par.HandledResponseTypes)
	}

	return request, nil
}

// OpenIDSession represents the types available for an oidc.Session that are required in the models package.
type OpenIDSession interface {
	oauthelia2.Session

	GetChallengeID() uuid.NullUUID
}
