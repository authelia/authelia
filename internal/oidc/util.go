package oidc

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	fjwt "authelia.com/provider/oauth2/token/jwt"
	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ory/x/errorsx"
	"golang.org/x/text/language"
)

// IsPushedAuthorizedRequest returns true if the requester has a PushedAuthorizationRequest redirect_uri value.
func IsPushedAuthorizedRequest(r oauthelia2.Requester, prefix string) bool {
	return strings.HasPrefix(r.GetRequestForm().Get(FormParameterRequestURI), prefix)
}

// MatchScopes uses a oauthelia2.ScopeStrategy to check if scopes match.
func MatchScopes(strategy oauthelia2.ScopeStrategy, granted, scopes []string) error {
	for _, scope := range scopes {
		if scope == "" {
			continue
		}

		if !strategy(granted, scope) {
			return errorsx.WithStack(oauthelia2.ErrInvalidScope.WithHintf("The request scope '%s' has not been granted or is not allowed to be requested.", scope))
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

// ErrorToDebugRFC6749Error converts the provided error to a *DebugRFC6749Error provided it is not nil and can be
// cast as a *oauthelia2.RFC6749Error.
func ErrorToDebugRFC6749Error(err error) (rfc error) {
	if err == nil {
		return nil
	}

	var e *oauthelia2.RFC6749Error

	if errors.As(err, &e) {
		return &DebugRFC6749Error{e}
	}

	return err
}

// DebugRFC6749Error is a decorator type which makes the underlying *oauthelia2.RFC6749Error expose debug information and
// show the full error description.
type DebugRFC6749Error struct {
	*oauthelia2.RFC6749Error
}

// Error implements the builtin error interface and shows the error with its debug info and description.
func (err *DebugRFC6749Error) Error() string {
	return err.WithExposeDebug(true).GetDescription()
}

// GetLangFromRequester gets the expected language for a requester.
func GetLangFromRequester(requester oauthelia2.Requester) language.Tag {
	var (
		ctx oauthelia2.G11NContext
		ok  bool
	)

	if ctx, ok = requester.(oauthelia2.G11NContext); ok {
		return ctx.GetLang()
	}

	return language.English
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
	case time.Time:
		return a
	case float64:
		return time.Unix(int64(a), 0).UTC()
	case int:
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

// RFC6750Header turns a *oauthelia2.RFC6749Error into the values for a RFC6750 format WWW-Authenticate Bearer response
// header, excluding the Bearer prefix.
func RFC6750Header(realm, scope string, err *oauthelia2.RFC6749Error) string {
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
func AccessResponderToClearMap(responder oauthelia2.AccessResponder) map[string]any {
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

// PopulateClientCredentialsFlowSessionWithAccessRequest is used to configure a session when performing a client credentials grant.
func PopulateClientCredentialsFlowSessionWithAccessRequest(ctx Context, client oauthelia2.Client, session *Session) (err error) {
	var (
		issuer *url.URL
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		return oauthelia2.ErrServerError.WithWrap(err).WithDebugf("Failed to determine the issuer with error: %s.", err.Error())
	}

	if client == nil {
		return oauthelia2.ErrServerError.WithDebug("Failed to get the client for the request.")
	}

	session.Subject = ""
	session.Claims.Subject = client.GetID()
	session.ClientID = client.GetID()
	session.DefaultSession.Claims.Issuer = issuer.String()
	session.DefaultSession.Claims.IssuedAt = ctx.GetClock().Now().UTC()
	session.DefaultSession.Claims.RequestedAt = ctx.GetClock().Now().UTC()
	session.ClientCredentials = true

	return nil
}

// PopulateClientCredentialsFlowRequester is used to grant the authorized scopes and audiences when performing a client
// credentials grant.
func PopulateClientCredentialsFlowRequester(ctx Context, config oauthelia2.Configurator, client oauthelia2.Client, requester oauthelia2.Requester) (err error) {
	if client == nil || config == nil || requester == nil {
		return oauthelia2.ErrServerError.WithDebug("Failed to get the client, configuration, or requester for the request.")
	}

	scopes := requester.GetRequestedScopes()
	audience := requester.GetRequestedAudience()

	var authz, nauthz bool

	strategy := config.GetScopeStrategy(ctx)

	for _, scope := range scopes {
		switch scope {
		case ScopeOffline, ScopeOfflineAccess:
			break
		case ScopeAutheliaBearerAuthz:
			authz = true
		default:
			nauthz = true
		}

		if strategy(client.GetScopes(), scope) {
			requester.GrantScope(scope)
		} else {
			return oauthelia2.ErrInvalidScope.WithDebugf("The scope '%s' is not authorized on client with id '%s'.", scope, client.GetID())
		}
	}

	if authz && nauthz {
		return oauthelia2.ErrInvalidScope.WithDebugf("The scope '%s' must only be requested by itself or with the '%s' scope, no other scopes are permitted.", ScopeAutheliaBearerAuthz, ScopeOfflineAccess)
	}

	if authz && len(audience) == 0 {
		return oauthelia2.ErrInvalidRequest.WithDebugf("The scope '%s' requires the request also include an audience.", ScopeAutheliaBearerAuthz)
	}

	if err = config.GetAudienceStrategy(ctx)(client.GetAudience(), audience); err != nil {
		return err
	}

	for _, aud := range audience {
		requester.GrantAudience(aud)
	}

	return nil
}
