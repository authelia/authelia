package oidc

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/text/language"
)

// IsPushedAuthorizedRequest returns true if the requester has a PushedAuthorizationRequest redirect_uri value.
func IsPushedAuthorizedRequest(r oauthelia2.Requester, prefix string) bool {
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
func IsJWTProfileAccessToken(header map[string]any) bool {
	if header == nil {
		return false
	}

	var (
		raw any
		typ string
		ok  bool
	)

	if raw, ok = header[JWTHeaderKeyType]; !ok {
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

	var authz, nauthz bool

	for _, scope := range scopes {
		switch scope {
		case ScopeOffline, ScopeOfflineAccess:
			break
		case ScopeAutheliaBearerAuthz:
			authz = true
		default:
			nauthz = true
		}
	}

	if authz && nauthz {
		return oauthelia2.ErrInvalidScope.WithDebugf("The scope '%s' must only be requested by itself or with the '%s' scope, no other scopes are permitted.", ScopeAutheliaBearerAuthz, ScopeOfflineAccess)
	}

	return nil
}

func IsAccessToken(ctx Context, value string) (is bool, err error) {
	config := ctx.GetConfiguration()

	if config.IdentityProviders.OIDC == nil || !config.IdentityProviders.OIDC.Discovery.BearerAuthorization {
		return false, nil
	}

	// Opaue Authelia Access Tokens have the 'authelia_at_' prefix and contain a HMAC signature.
	if strings.HasPrefix(value, "authelia_at_") && strings.Count(value, ".") == 1 {
		return true, nil
	}

	if !IsMaybeSignedJWT(value) {
		return false, nil
	}

	var (
		issuer *url.URL
		token  *jwt.Token
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		return false, fmt.Errorf("error occurred determining the issuer: %w", err)
	}

	if token, _, err = jwt.NewParser(jwt.WithoutClaimsValidation()).ParseUnverified(value, &jwt.RegisteredClaims{}); err != nil {
		return false, fmt.Errorf("error occurred parsing bearer token: %w", err)
	}

	if !IsJWTProfileAccessToken(token.Header) {
		return false, fmt.Errorf("error occurred checking the token: the token is not a JWT profile access token")
	}

	var iss string

	if iss, err = token.Claims.GetIssuer(); err != nil {
		return false, fmt.Errorf("error occurred chekcing the token: error getting the token issuer claim: %w", err)
	}

	if strings.EqualFold(iss, issuer.String()) {
		return true, nil
	}

	return false, fmt.Errorf("error occurred checking the token: the token issuer '%s' does not match the expected '%s'", iss, issuer)
}

func IsMaybeSignedJWT(value string) (is bool) {
	return strings.Count(value, ".") == 2
}
