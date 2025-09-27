package oidc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	fjwt "authelia.com/provider/oauth2/token/jwt"
	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/text/language"

	"github.com/authelia/authelia/v4/internal/utils"
)

// IsPushedAuthorizedRequest returns true if the requester has a PushedAuthorizationRequest redirect_uri value.
func IsPushedAuthorizedRequest(r oauthelia2.Requester, prefix string) (is bool) {
	if r == nil {
		return false
	}

	return IsPushedAuthorizedRequestForm(r.GetRequestForm(), prefix)
}

// IsPushedAuthorizedRequestForm returns true if the provided form is a Pushed Authorization Request Form.
func IsPushedAuthorizedRequestForm(form url.Values, prefix string) (is bool) {
	return strings.HasPrefix(form.Get(FormParameterRequestURI), prefix)
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
	if responder == nil {
		return nil
	}

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

// HydrateClientCredentialsFlowSessionWithAccessRequest is used to configure a session when performing a client credentials grant.
func HydrateClientCredentialsFlowSessionWithAccessRequest(ctx Context, client oauthelia2.Client, session *Session) (err error) {
	var (
		issuer *url.URL
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		return oauthelia2.ErrServerError.WithWrap(err).WithDebugf("Failed to determine the issuer with error: %s.", err.Error())
	}

	if client == nil {
		return oauthelia2.ErrServerError.WithDebug("Failed to get the client for the request.")
	}

	InitializeSessionDefaults(session)

	session.Subject = ""
	session.ClientID = client.GetID()
	session.Claims.Subject = client.GetID()
	session.Claims.Issuer = issuer.String()
	session.Claims.IssuedAt = fjwt.NewNumericDate(ctx.GetClock().Now().UTC())
	session.SetRequestedAt(ctx.GetClock().Now().UTC())
	session.ClientCredentials = true

	return nil
}

// InitializeSessionDefaults ensures a *Session has safe initialized defaults for most purposes.
func InitializeSessionDefaults(session *Session) {
	switch {
	case session.DefaultSession == nil:
		session.DefaultSession = &openid.DefaultSession{
			Headers: &fjwt.Headers{
				Extra: make(map[string]any),
			},
			Claims: &fjwt.IDTokenClaims{
				Extra: make(map[string]any),
			},
		}
	case session.Claims == nil:
		session.Claims = &fjwt.IDTokenClaims{
			Extra: make(map[string]any),
		}
	case session.Claims.Extra == nil:
		session.Claims.Extra = make(map[string]any)
	}

	if session.Extra == nil {
		session.Extra = make(map[string]any)
	}

	if session.Headers == nil {
		session.Headers = &fjwt.Headers{
			Extra: make(map[string]any),
		}
	} else if session.Headers.Extra == nil {
		session.Headers.Extra = make(map[string]any)
	}
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

// IsAccessToken returns true if the provided token is possibly an Authelia OAuth 2.0 Access Token.
func IsAccessToken(ctx Context, value string) (is bool, err error) {
	if ctx == nil {
		return false, fmt.Errorf("error occurred getting configuration: context wasn't provided")
	}

	config := ctx.GetConfiguration()

	if config.IdentityProviders.OIDC == nil || !config.IdentityProviders.OIDC.Discovery.BearerAuthorization {
		return false, nil
	}

	// Opaque Authelia Access Tokens have the 'authelia_at_' prefix and contain a HMAC signature.
	if strings.HasPrefix(value, fmt.Sprintf(fmtAutheliaOpaqueOAuth2Token, "at")) && strings.Count(value, ".") == 1 {
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

// IsMaybeSignedJWT returns true if the provided string has the necessary characteristics to be a Compact Signed JWT.
func IsMaybeSignedJWT(value string) (is bool) {
	return strings.Count(value, ".") == 2
}

// RequesterRequiresLogin returns true if the oauthelia2.Requester requires the user to authenticate again.
func RequesterRequiresLogin(requester oauthelia2.Requester, requested, authenticated time.Time) (required bool) {
	if requester == nil {
		return false
	}

	return RequestFormRequiresLogin(requester.GetRequestForm(), requested, authenticated)
}

// RequestFormRequiresLogin returns true if the form requires the user to authenticate again.
func RequestFormRequiresLogin(form url.Values, requested, authenticated time.Time) (required bool) {
	if form.Has(FormParameterPrompt) {
		if oauthelia2.Arguments(oauthelia2.RemoveEmpty(strings.Split(form.Get(FormParameterPrompt), " "))).Has(PromptLogin) && authenticated.Before(requested) {
			return true
		}
	}

	if form.Has(FormParameterMaximumAge) {
		value := form.Get(FormParameterMaximumAge)

		var (
			age int64
			err error
		)
		if age, err = strconv.ParseInt(value, 10, 64); err != nil {
			age = 0
		}

		return age == 0 || authenticated.IsZero() || requested.IsZero() || authenticated.Add(time.Duration(age)*time.Second).Before(requested)
	}

	return false
}

func ValidateSectorIdentifierURI(ctx ClientContext, cache map[string][]string, sectorURI *url.URL, redirectURIs []string) (err error) {
	var (
		sectorRedirectURIs []string
	)

	if sectorRedirectURIs, err = getSectorIdentifierURICache(ctx, cache, sectorURI); err != nil {
		return err
	}

	var invalidRedirectURIs []string //nolint:prealloc

	for _, rawRedirectURI := range redirectURIs {
		if _, match := oauthelia2.IsMatchingRedirectURI(rawRedirectURI, sectorRedirectURIs); match {
			continue
		}

		invalidRedirectURIs = append(invalidRedirectURIs, rawRedirectURI)
	}

	switch len(invalidRedirectURIs) {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("error checking redirect_uri '%s' against '%s'", invalidRedirectURIs[0], utils.StringJoinAnd(sectorRedirectURIs))
	default:
		return fmt.Errorf("error checking redirect_uris '%s' against '%s'", utils.StringJoinAnd(invalidRedirectURIs), utils.StringJoinAnd(sectorRedirectURIs))
	}
}

func getSectorIdentifierURICache(ctx ClientContext, cache map[string][]string, sectorURI *url.URL) (redirectURIs []string, err error) {
	if cache != nil {
		var ok bool

		if redirectURIs, ok = cache[sectorURI.String()]; ok {
			return redirectURIs, nil
		}
	}

	redirectURIs = make([]string, 0)

	client := ctx.GetHTTPClient()

	var resp *http.Response

	if resp, err = client.Get(sectorURI.String()); err != nil {
		return nil, fmt.Errorf("error occurred making request to '%s' for the sector identifier document: %w", sectorURI, err)
	}

	if err = json.NewDecoder(resp.Body).Decode(&redirectURIs); err != nil {
		return nil, fmt.Errorf("error occurred decoding request from '%s' with the sector identifier document: %w", sectorURI, err)
	}

	if cache != nil {
		cache[sectorURI.String()] = redirectURIs
	}

	return redirectURIs, nil
}

func float64Match(expected float64, value any, values []any) (ok bool) {
	var f float64

	if value != nil {
		if f, ok = float64As(value); ok {
			return expected == f
		}
	}

	for _, v := range values {
		if f, ok = float64As(v); ok && expected == f {
			return true
		}
	}

	return false
}

func float64As(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case int16:
		return float64(v), true
	case int8:
		return float64(v), true
	case int:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint:
		return float64(v), true
	default:
		return 0, false
	}
}

// ParseSpaceDelimitedFromParameter obtains the value of a specific key in a url.Values form returning it as an
// oauth2.Arguments slice using spaces as a delimiter (without empty values).
func ParseSpaceDelimitedFromParameter(form url.Values, key string) oauthelia2.Arguments {
	var value string

	if form.Has(key) {
		value = strings.Join(form[key], " ")
	} else {
		return oauthelia2.Arguments{}
	}

	return oauthelia2.RemoveEmpty(strings.Split(value, " "))
}

// FormRequiresExplicitConsent evaluates form values in the url.Values format for evidence that the form requires
// explicit consent, for example if the client requested explicit consent, or the flow would result in a Refresh Token.
func FormRequiresExplicitConsent(form url.Values) (required bool) {
	prompt := ParseSpaceDelimitedFromParameter(form, FormParameterPrompt)

	if prompt.Has(PromptConsent) {
		return true
	}

	// This is required currently as the user will be presented the consent prompt to enter their password.
	if prompt.Has(PromptLogin) {
		return true
	}

	if FormIsAuthorizeCodeFlow(form) {
		if ParseSpaceDelimitedFromParameter(form, FormParameterScope).HasOneOf(ScopeOffline, ScopeOfflineAccess, ScopeAutheliaBearerAuthz) {
			return true
		}
	}

	return false
}

// RequesterRequiresExplicitConsent evaluates a oauth2.Requester for evidence that the request requires explicit
// consent, for example if the client requested explicit consent, or the flow would result in a Refresh Token.
func RequesterRequiresExplicitConsent(requester oauthelia2.Requester) (required bool) {
	if requester == nil {
		return false
	}

	prompt := ParseSpaceDelimitedFromParameter(requester.GetRequestForm(), FormParameterPrompt)

	if prompt.Has(PromptConsent) {
		return true
	}

	// This is required currently as the user will be presented the consent prompt to enter their password.
	if prompt.Has(PromptLogin) {
		return true
	}

	if RequesterIsAuthorizeCodeFlow(requester) {
		if requester.GetRequestedScopes().HasOneOf(ScopeOffline, ScopeOfflineAccess, ScopeAutheliaBearerAuthz) {
			return true
		}

		if requester.GetGrantedScopes().HasOneOf(ScopeOffline, ScopeOfflineAccess, ScopeAutheliaBearerAuthz) {
			return true
		}
	}

	return false
}

// FormIsAuthorizeCodeFlow evaluates form values in the url.Values format to see if the flow would result in an
// Authorization Code.
func FormIsAuthorizeCodeFlow(form url.Values) (is bool) {
	return ParseSpaceDelimitedFromParameter(form, FormParameterResponseType).Has(ResponseTypeAuthorizationCodeFlow)
}

// RequesterIsAuthorizeCodeFlow evaluates an oauth2.Requester to see if the flow would result in an Authorization Code.
func RequesterIsAuthorizeCodeFlow(requester oauthelia2.Requester) (is bool) {
	if ar, ok := requester.(oauthelia2.AuthorizeRequester); ok && ar.GetResponseTypes().Has(ResponseTypeAuthorizationCodeFlow) {
		return true
	}

	return false
}
