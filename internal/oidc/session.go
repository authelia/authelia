package oidc

import (
	"net/url"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	"authelia.com/provider/oauth2/token/jwt"
	"github.com/google/uuid"
	"github.com/mohae/deepcopy"

	"github.com/authelia/authelia/v4/internal/model"
)

// NewSession creates a new empty OpenIDSession struct with the requested at value being time.Now().
func NewSession() (session *Session) {
	return NewSessionWithRequestedAt(time.Now())
}

// NewSessionWithRequestedAt creates a new empty OpenIDSession struct with a specific requested at value.
func NewSessionWithRequestedAt(requestedAt time.Time) (session *Session) {
	session = &Session{}

	InitializeSessionDefaults(session)

	session.SetRequestedAt(requestedAt.UTC())

	return session
}

// NewSessionWithRequester uses details from a Requester to generate an OpenIDSession.
func NewSessionWithRequester(ctx Context, issuer *url.URL, kid, username string, amr []string, extra map[string]any,
	authTime time.Time, consent *model.OAuth2ConsentSession, requester oauthelia2.Requester, claims *ClaimsRequests) (session *Session) {
	session = NewSessionWithRequestedAt(ctx.GetClock().Now())

	session.SetValuesFromRequester(requester)
	session.SetValuesFromConsentSession(consent)
	session.SetValuesGeneral(ctx, issuer, kid, username, amr, authTime, claims, extra)

	return session
}

// Session holds OpenID Connect 1.0 Session information.
type Session struct {
	*openid.DefaultSession `json:"id_token"`

	ChallengeID           uuid.NullUUID   `json:"challenge_id"`
	KID                   string          `json:"kid"`
	ClientID              string          `json:"client_id"`
	ClientCredentials     bool            `json:"client_credentials"`
	ExcludeNotBeforeClaim bool            `json:"exclude_nbf_claim"`
	AllowedTopLevelClaims []string        `json:"allowed_top_level_claims"`
	ClaimRequests         *ClaimsRequests `json:"claim_requests,omitempty"`
	GrantedClaims         []string        `json:"granted_claims,omitempty"`
	Extra                 map[string]any  `json:"extra"`
}

func (s *Session) SetValuesFromRequester(requester oauthelia2.Requester) {
	s.ClientID = requester.GetClient().GetID()
	s.Claims.AuthorizedParty = requester.GetClient().GetID()
	s.Claims.Nonce = requester.GetRequestForm().Get(FormParameterNonce)
}

func (s *Session) SetValuesFromConsentSession(consent *model.OAuth2ConsentSession) {
	s.SetRequestedAt(consent.RequestedAt)

	s.ChallengeID = model.NullUUID(consent.ChallengeID)
	s.GrantedClaims = consent.GrantedClaims
	s.Subject = consent.Subject.UUID.String()
	s.Claims.Subject = consent.Subject.UUID.String()
}

func (s *Session) SetValuesGeneral(ctx Context, issuer *url.URL, kid string, username string, amr []string, authTime time.Time, claims *ClaimsRequests, extra map[string]any) {
	if issuer != nil {
		s.Claims.Issuer = issuer.String()
	}

	s.Claims.IssuedAt = jwt.NewNumericDate(ctx.GetClock().Now())

	if !authTime.IsZero() {
		s.Claims.AuthTime = jwt.NewNumericDate(authTime)
	}

	if len(kid) != 0 {
		s.Headers.Extra[JWTHeaderKeyIdentifier] = kid
	}

	if len(username) != 0 {
		s.Username = username
	}

	if len(amr) != 0 {
		s.Claims.AuthenticationMethodsReferences = amr
	}

	if claims != nil {
		s.ClaimRequests = claims
	}

	if len(extra) != 0 {
		s.Claims.Extra = extra
	}
}

// GetChallengeID returns the challenge id.
func (s *Session) GetChallengeID() (challenge uuid.NullUUID) {
	return s.ChallengeID
}

// GetJWTHeader returns the *jwt.Headers for the OAuth 2.0 JWT Profile Access Token.
func (s *Session) GetJWTHeader() (headers *jwt.Headers) {
	headers = &jwt.Headers{
		Extra: map[string]any{
			JWTHeaderKeyType: JWTHeaderTypeValueAccessTokenJWT,
		},
	}

	if len(s.KID) != 0 {
		headers.Extra[JWTHeaderKeyIdentifier] = s.KID
	}

	return headers
}

// GetJWTClaims returns the jwt.JWTClaimsContainer for the OAuth 2.0 JWT Profile Access Tokens.
func (s *Session) GetJWTClaims() jwt.JWTClaimsContainer {
	//nolint:prealloc
	var (
		allowed []string
		amr     bool
	)

	for _, cl := range s.AllowedTopLevelClaims {
		switch cl {
		case ClaimJWTID, ClaimIssuer, ClaimSubject, ClaimAudience, ClaimExpirationTime, ClaimNotBefore, ClaimIssuedAt, ClaimClientIdentifier, ClaimScopeNonStandard, ClaimExtra:
			continue
		case ClaimAuthenticationMethodsReference:
			amr = true

			continue
		}

		allowed = append(allowed, cl)
	}

	claims := &jwt.JWTClaims{
		Subject:   s.Subject,
		ExpiresAt: s.GetExpiresAt(oauthelia2.AccessToken),
		IssuedAt:  time.Now().UTC(),
		Extra:     map[string]any{},
	}

	if len(s.Extra) > 0 {
		claims.Extra[ClaimExtra] = s.Extra
	}

	if s.DefaultSession != nil && s.Claims != nil {
		for _, allowedClaim := range allowed {
			if cl, ok := s.Claims.Extra[allowedClaim]; ok {
				claims.Extra[allowedClaim] = cl
			}
		}

		claims.Issuer = s.Claims.Issuer

		if amr && len(s.Claims.AuthenticationMethodsReferences) != 0 {
			claims.Extra[ClaimAuthenticationMethodsReference] = s.Claims.AuthenticationMethodsReferences
		}
	}

	if len(s.ClientID) != 0 {
		claims.Extra[ClaimClientIdentifier] = s.ClientID
	}

	return claims
}

// GetIDTokenClaims returns the *jwt.IDTokenClaims for this session.
func (s *Session) GetIDTokenClaims() (claims *jwt.IDTokenClaims) {
	if s.DefaultSession == nil {
		return nil
	}

	return s.Claims
}

// GetExtraClaims returns the Extra/Unregistered claims for this session.
func (s *Session) GetExtraClaims() map[string]any {
	return s.Extra
}

// Clone copies the OpenIDSession to a new oauthelia2.Session.
func (s *Session) Clone() oauthelia2.Session {
	if s == nil {
		return nil
	}

	return deepcopy.Copy(s).(oauthelia2.Session)
}

// ConsentGrantImplicit that handles the implicit consent flow assigning the subject and responded at values then
// allows ConsentGrant to finalize the grant mechanics.
func ConsentGrantImplicit(consent *model.OAuth2ConsentSession, claims []string, subject uuid.UUID, respondedAt time.Time) {
	consent.SetRespondedAt(respondedAt, 0)
	consent.SetSubject(subject)

	ConsentGrant(consent, false, claims)
}

// ConsentGrant is a helper function to perform specific consent granting functionality. In particular in honors the
// requirements around consent like not allowing access to a refresh token unless the user has explicitly consented.
func ConsentGrant(consent *model.OAuth2ConsentSession, explicit bool, claims []string) {
	consent.GrantAudience()
	consent.GrantClaims(claims)

	if explicit {
		consent.GrantScopes()
	} else {
		for _, scope := range consent.RequestedScopes {
			switch scope {
			case ScopeOffline, ScopeOfflineAccess:
				continue
			default:
				consent.GrantScope(scope)
			}
		}
	}
}
