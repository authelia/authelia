package model

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
)

func NewOAuth2SessionFromRequest(signature string, r fosite.Requester) (session *OAuth2Session, err error) {
	var (
		subject  string
		sess     fosite.Session
		sessData []byte
	)

	sess = r.GetSession()

	if sess != nil {
		subject = sess.GetSubject()

		if sessData, err = json.Marshal(sess); err != nil {
			return nil, err
		}
	}

	return &OAuth2Session{
		RequestID:         r.GetID(),
		ClientID:          r.GetClient().GetID(),
		Signature:         signature,
		RequestedAt:       r.GetRequestedAt(),
		Subject:           subject,
		RequestedScopes:   StringSlicePipeDelimited(r.GetRequestedScopes()),
		GrantedScopes:     StringSlicePipeDelimited(r.GetGrantedScopes()),
		RequestedAudience: StringSlicePipeDelimited(r.GetRequestedAudience()),
		GrantedAudience:   StringSlicePipeDelimited(r.GetGrantedAudience()),
		Revoked:           false,
		Form:              r.GetRequestForm().Encode(),
		Session:           sessData,
	}, nil
}

func NewOAuth2BlacklistedJTI(jti string, exp time.Time) (jtiBlacklist *OAuth2BlacklistedJTI) {
	return &OAuth2BlacklistedJTI{
		Signature: fmt.Sprintf("%x", sha256.Sum256([]byte(jti))),
		ExpiresAt: exp,
	}
}

type OAuth2BlacklistedJTI struct {
	ID        int       `db:"id"`
	Signature string    `db:"signature"`
	ExpiresAt time.Time `db:"expires_at"`
}

// OpenIDSession holds OIDC Session information.
type OpenIDSession struct {
	*openid.DefaultSession `json:"idToken"`

	ChallengeID string
	ClientID    string

	Extra map[string]interface{} `json:"extra"`
}

type OAuth2Session struct {
	ID                int                      `db:"id"`
	RequestID         string                   `db:"request_id"`
	ClientID          string                   `db:"client_id"`
	Signature         string                   `db:"signature"`
	RequestedAt       time.Time                `db:"requested_at"`
	Subject           string                   `db:"subject"`
	RequestedScopes   StringSlicePipeDelimited `db:"requested_scopes"`
	GrantedScopes     StringSlicePipeDelimited `db:"granted_scopes"`
	RequestedAudience StringSlicePipeDelimited `db:"requested_scopes"`
	GrantedAudience   StringSlicePipeDelimited `db:"granted_scopes"`
	Revoked           bool                     `db:"revoked"`
	Form              string                   `db:"form_data"`
	Session           []byte                   `db:"session_data"`
}

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
