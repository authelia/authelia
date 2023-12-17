package oidc

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ory/fosite"
	fjwt "github.com/ory/fosite/token/jwt"
	"github.com/ory/x/errorsx"
	"golang.org/x/text/language"
	"gopkg.in/square/go-jose.v2"
)

// IsPushedAuthorizedRequest returns true if the requester has a PushedAuthorizationRequest redirect_uri value.
func IsPushedAuthorizedRequest(r fosite.Requester, prefix string) bool {
	return strings.HasPrefix(r.GetRequestForm().Get(FormParameterRequestURI), prefix)
}

// MatchScopes uses a fosite.ScopeStrategy to check if scopes match.
func MatchScopes(strategy fosite.ScopeStrategy, granted, scopes []string) error {
	for _, scope := range scopes {
		if scope == "" {
			continue
		}

		if !strategy(granted, scope) {
			return errorsx.WithStack(fosite.ErrInvalidScope.WithHintf("The request scope '%s' has not been granted or is not allowed to be requested.", scope))
		}
	}

	return nil
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

// JTIFromMapClaims returns a JTI from a jwt.MapClaims.
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

// Error implements the builtin error interface and shows the error with its debug info and description.
func (err *DebugRFC6749Error) Error() string {
	return err.WithExposeDebug(true).GetDescription()
}

// GetLangFromRequester gets the expected language for a requester.
func GetLangFromRequester(requester fosite.Requester) language.Tag {
	var (
		ctx fosite.G11NContext
		ok  bool
	)

	if ctx, ok = requester.(fosite.G11NContext); ok {
		return ctx.GetLang()
	}

	return language.English
}

// IntrospectionResponseToMap converts a fosite.IntrospectionResponder into a map[string]any which is used to either
// respond to the introspection request with JSON or a JWT.
func IntrospectionResponseToMap(response fosite.IntrospectionResponder) (aud []string, introspection map[string]any) {
	introspection = map[string]any{
		ClaimActive: false,
	}

	if response == nil {
		return nil, introspection
	}

	if response.IsActive() {
		introspection[ClaimActive] = true

		mapIntrospectionAccessRequesterToMap(response.GetAccessRequester(), introspection)
	}

	return sliceIntrospectionResponseToRequesterAudience(response), introspection
}

func mapIntrospectionAccessRequesterToMap(ar fosite.AccessRequester, introspection map[string]any) {
	if ar == nil {
		return
	}

	var (
		ok  bool
		aud fosite.Arguments
	)

	if client := ar.GetClient(); client != nil {
		if id := client.GetID(); id != "" {
			introspection[ClaimClientIdentifier] = id
		}
	}

	if scope := ar.GetGrantedScopes(); len(scope) > 0 {
		introspection[ClaimScope] = strings.Join(scope, " ")
	}

	if _, ok = introspection[ClaimIssuedAt]; !ok {
		if rat := ar.GetRequestedAt(); !rat.IsZero() {
			introspection[ClaimIssuedAt] = rat.Unix()
		}
	}

	if aud = ar.GetGrantedAudience(); len(aud) > 0 {
		introspection[ClaimAudience] = []string(aud)
	}

	mapIntrospectionAccessRequesterSessionToMap(ar, introspection)
}

func mapIntrospectionAccessRequesterSessionToMap(ar fosite.AccessRequester, introspection map[string]any) {
	session := ar.GetSession()

	if session == nil {
		return
	}

	var (
		ok    bool
		extra fosite.ExtraClaimsSession
	)

	if extra, ok = session.(fosite.ExtraClaimsSession); ok {
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

	if exp := session.GetExpiresAt(fosite.AccessToken); !exp.IsZero() {
		introspection[ClaimExpirationTime] = exp.Unix()
	}

	var claimsSession IDTokenClaimsSession

	if sub := session.GetSubject(); sub != "" {
		introspection[ClaimSubject] = sub
	} else if claimsSession, ok = session.(IDTokenClaimsSession); ok {
		claims := claimsSession.GetIDTokenClaims()

		if claims != nil && claims.Subject != "" {
			introspection[ClaimSubject] = claims.Subject
		}
	}

	if username := session.GetUsername(); username != "" {
		introspection[ClaimUsername] = username
	}
}

func sliceIntrospectionResponseToRequesterAudience(response fosite.IntrospectionResponder) (aud []string) {
	if cr, ok := response.(ClientRequesterResponder); ok {
		var client fosite.Client

		if client = cr.GetClient(); client == nil {
			return
		}

		return []string{client.GetID()}
	}

	return nil
}

func mapCopy(src map[string]any) (dst map[string]any) {
	dst = make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}

	return dst
}

func toStringSlice(v any) (result []string) {
	switch s := v.(type) {
	case string:
		return []string{s}
	case []string:
		return s
	case []any:
		for _, sv := range s {
			if ss, ok := sv.(string); ok {
				result = append(result, ss)
			}
		}

		return result
	default:
		return nil
	}
}

func toTime(v any, def time.Time) (t time.Time) {
	switch a := v.(type) {
	case float64:
		return time.Unix(int64(a), 0).UTC()
	case int64:
		return time.Unix(a, 0).UTC()
	default:
		return def
	}
}

// IsJWTProfileAccessToken validates a *jwt.Token is actually a RFC9068 JWT Profile Access Token by checking the
// relevant header as per https://datatracker.ietf.org/doc/html/rfc9068#section-2.1 which explicitly states that
// the header MUST include a typ of 'at+jwt' or 'application/at+jwt' with a preference of 'at+jwt'.
func IsJWTProfileAccessToken(token *fjwt.Token) bool {
	if token == nil || token.Header == nil {
		return false
	}

	var (
		raw any
		typ string
		ok  bool
	)

	if raw, ok = token.Header[JWTHeaderKeyType]; !ok {
		return false
	}

	typ, ok = raw.(string)

	return ok && (typ == JWTHeaderTypeValueAccessTokenJWT)
}

// RFC6750Header turns a *fosite.RFC6749Error into the values for a RFC6750 format WWW-Authenticate Bearer response
// header, excluding the Bearer prefix.
func RFC6750Header(realm, scope string, err *fosite.RFC6749Error) string {
	values := err.ToValues()

	if realm != "" {
		values.Set("realm", realm)
	}

	if scope != "" {
		values.Set("scope", scope)
	}

	//nolint:prealloc
	var (
		keys []string
		key  string
	)

	for key = range values {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		switch keys[i] {
		case fieldRFC6750Realm:
			return true
		case fieldRFC6750Error:
			switch keys[j] {
			case fieldRFC6750ErrorDescription, fieldRFC6750Scope:
				return true
			default:
				return false
			}
		case fieldRFC6750ErrorDescription:
			switch keys[j] {
			case fieldRFC6750Scope:
				return true
			default:
				return false
			}
		case fieldRFC6750Scope:
			switch keys[j] {
			case fieldRFC6750Realm, fieldRFC6750Error, fieldRFC6750ErrorDescription:
				return false
			default:
				return keys[i] < keys[j]
			}
		default:
			return keys[i] < keys[j]
		}
	})

	parts := make([]string, len(keys))

	var i int

	for i, key = range keys {
		parts[i] = fmt.Sprintf(`%s="%s"`, key, values.Get(key))
	}

	return strings.Join(parts, ",")
}

// AccessResponderToClearMap returns a clear friendly map copy of the responder map values.
func AccessResponderToClearMap(responder fosite.AccessResponder) map[string]any {
	m := responder.ToMap()

	data := make(map[string]any, len(m))

	for key, value := range responder.ToMap() {
		switch key {
		case "access_token":
			data[key] = "authelia_at_**************"
		case "refresh_token":
			data[key] = "authelia_rt_**************"
		case "id_token":
			data[key] = "*********.***********.*************"
		default:
			data[key] = value
		}
	}

	return data
}
