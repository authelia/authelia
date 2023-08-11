package oidc

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ory/fosite"
	"gopkg.in/square/go-jose.v2"
)

// IsPushedAuthorizedRequest returns true if the requester has a PushedAuthorizationRequest redirect_uri value.
func IsPushedAuthorizedRequest(r fosite.Requester, prefix string) bool {
	return strings.HasPrefix(r.GetRequestForm().Get(FormParameterRequestURI), prefix)
}

// SortedSigningAlgs is a sorting type which allows the use of sort.Sort to order a list of OAuth 2.0 Signing Algs.
// Sorting occurs in the order of from within the RFC's.
type SortedSigningAlgs []string

func (algs SortedSigningAlgs) Len() int {
	return len(algs)
}

func (algs SortedSigningAlgs) Less(i, j int) bool {
	return isSigningAlgLess(algs[i], algs[j])
}

func (algs SortedSigningAlgs) Swap(i, j int) {
	algs[i], algs[j] = algs[j], algs[i]
}

type SortedJSONWebKey []jose.JSONWebKey

func (jwks SortedJSONWebKey) Len() int {
	return len(jwks)
}

func (jwks SortedJSONWebKey) Less(i, j int) bool {
	if jwks[i].Algorithm == jwks[j].Algorithm {
		return jwks[i].KeyID < jwks[j].KeyID
	}

	return isSigningAlgLess(jwks[i].Algorithm, jwks[j].Algorithm)
}

func (jwks SortedJSONWebKey) Swap(i, j int) {
	jwks[i], jwks[j] = jwks[j], jwks[i]
}

//nolint:gocyclo // Low importance func.
func isSigningAlgLess(i, j string) bool {
	switch {
	case i == j:
		return false
	case i == SigningAlgNone:
		return false
	case j == SigningAlgNone:
		return true
	default:
		var (
			ip, jp string
			it, jt bool
		)

		if len(i) > 2 {
			it = true
			ip = i[:2]
		}

		if len(j) > 2 {
			jt = true
			jp = j[:2]
		}

		switch {
		case it && jt && ip == jp:
			return i < j
		case ip == SigningAlgPrefixHMAC:
			return true
		case jp == SigningAlgPrefixHMAC:
			return false
		case ip == SigningAlgPrefixRSAPSS:
			return false
		case jp == SigningAlgPrefixRSAPSS:
			return true
		case ip == SigningAlgPrefixRSA:
			return true
		case jp == SigningAlgPrefixRSA:
			return false
		case ip == SigningAlgPrefixECDSA:
			return true
		case jp == SigningAlgPrefixECDSA:
			return false
		default:
			return false
		}
	}
}

func JTIFromMapClaims(m jwt.MapClaims) (jti string, err error) {
	var (
		ok  bool
		raw any
	)

	if raw, ok = m[ClaimJWTID]; !ok {
		return "", nil
	}

	if jti, ok = raw.(string); !ok {
		return "", fmt.Errorf("invalid type for claim: jti is invalid")
	}

	return jti, nil
}

func getExpiresIn(r fosite.Requester, key fosite.TokenType, defaultLifespan time.Duration, now time.Time) time.Duration {
	if r.GetSession().GetExpiresAt(key).IsZero() {
		return defaultLifespan
	}

	return time.Duration(r.GetSession().GetExpiresAt(key).UnixNano() - now.UnixNano())
}

// ErrorToDebugRFC6749Error converts the provided error to a *DebugRFC6749Error provided it is not nil and can be
// cast as a *fosite.RFC6749Error.
func ErrorToDebugRFC6749Error(err error) (rfc error) {
	if err == nil {
		return nil
	}

	var e *fosite.RFC6749Error

	if errors.As(err, &e) {
		return &DebugRFC6749Error{e}
	}

	return err
}

// DebugRFC6749Error is a decorator type which makes the underlying *fosite.RFC6749Error expose debug information and
// show the full error description.
type DebugRFC6749Error struct {
	*fosite.RFC6749Error
}

func (err *DebugRFC6749Error) Error() string {
	return err.WithExposeDebug(true).GetDescription()
}

// IntrospectionResponseToMap converts a fosite.IntrospectionResponder into a map[string]any which is used to either
// respond to the introspection request with JSON or a JWT.
func IntrospectionResponseToMap(response fosite.IntrospectionResponder) (aud []string, introspection map[string]any) {
	introspection = map[string]any{
		ClaimActive: false,
	}

	if response.IsActive() {
		introspection[ClaimActive] = true

		var (
			extra fosite.ExtraClaimsSession
			ok    bool
		)

		if extra, ok = response.GetAccessRequester().GetSession().(fosite.ExtraClaimsSession); ok {
			claims := extra.GetExtraClaims()

			for name, value := range claims {
				switch name {
				// We do not allow these to be set through extra claims.
				case ClaimExpirationTime, ClaimClientIdentifier, ClaimScope, ClaimIssuedAt, ClaimSubject, ClaimAudience, ClaimUsername:
					continue
				default:
					introspection[name] = value
				}
			}
		}

		if exp := response.GetAccessRequester().GetSession().GetExpiresAt(fosite.AccessToken); !exp.IsZero() {
			introspection[ClaimExpirationTime] = exp.Unix()
		}

		if id := response.GetAccessRequester().GetClient().GetID(); id != "" {
			introspection[ClaimClientIdentifier] = id
		}

		if scope := response.GetAccessRequester().GetGrantedScopes(); len(scope) > 0 {
			introspection[ClaimScope] = strings.Join(scope, " ")
		}

		if rat := response.GetAccessRequester().GetRequestedAt(); !rat.IsZero() {
			introspection[ClaimIssuedAt] = rat.Unix()
		}

		if sub := response.GetAccessRequester().GetSession().GetSubject(); sub != "" {
			introspection[ClaimSubject] = sub
		}

		if aud = response.GetAccessRequester().GetGrantedAudience(); len(aud) > 0 {
			introspection[ClaimAudience] = aud
		}

		if username := response.GetAccessRequester().GetSession().GetUsername(); username != "" {
			introspection[ClaimUsername] = username
		}
	}

	return aud, introspection
}
