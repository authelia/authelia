package oidc

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewClaimRequests parses the claims request parameter if set from a http.Request form.
func NewClaimRequests(form url.Values) (requests *ClaimsRequests, err error) {
	var raw string

	if raw = form.Get(FormParameterClaims); len(raw) == 0 {
		return nil, nil
	}

	requests = &ClaimsRequests{}

	if err = json.Unmarshal([]byte(raw), requests); err != nil {
		return nil, oauthelia2.ErrInvalidRequest.WithHint("The OAuth 2.0 client included a malformed 'claims' parameter in the authorization request.").WithWrap(err).WithDebugf("Error occurred attempting to parse the 'claims' parameter: %+v.", err)
	}

	return requests, nil
}

type OrderedClaimsRequests OrderedClaimsRequestsRaw

type OrderedClaimsRequestsRaw struct {
	IDToken  OrderedClaimRequests `json:"id_token,omitempty"`
	UserInfo OrderedClaimRequests `json:"userinfo,omitempty"`
}

func (ocr *OrderedClaimsRequests) MarshalJSON() ([]byte, error) {
	actual := &OrderedClaimsRequestsRaw{}

	if len(ocr.IDToken) > 0 {
		actual.IDToken = make(OrderedClaimRequests, len(ocr.IDToken))

		copy(actual.IDToken, ocr.IDToken)

		sort.SliceStable(actual.IDToken, func(i, j int) bool {
			return actual.IDToken[i].Claim < actual.IDToken[j].Claim
		})
	}

	if len(ocr.UserInfo) > 0 {
		actual.UserInfo = make(OrderedClaimRequests, len(ocr.UserInfo))

		copy(actual.UserInfo, ocr.UserInfo)

		sort.SliceStable(actual.UserInfo, func(i, j int) bool {
			return actual.UserInfo[i].Claim < actual.UserInfo[j].Claim
		})
	}

	return json.Marshal(actual)
}

func (ocr *OrderedClaimsRequests) Signature() (signature string, err error) {
	_, signature, err = ocr.Serialized()

	return
}

func (ocr *OrderedClaimsRequests) Serialized() (serialized, signature string, err error) {
	var data []byte

	if data, err = json.Marshal(ocr); err != nil {
		return "", "", err
	}

	hash := sha256.New()

	hash.Write(data)

	return string(data), fmt.Sprintf("%x", hash.Sum(nil)), nil
}

type OrderedClaimRequests []OrderedClaimRequest

func (ocr OrderedClaimRequests) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("{")

	for i, request := range ocr {
		if i > 0 {
			buf.WriteString(",")
		}

		key, err := json.Marshal(request.Claim)
		if err != nil {
			return nil, err
		}

		val, err := json.Marshal(request.Request)
		if err != nil {
			return nil, err
		}

		buf.Write(key)
		buf.WriteString(":")
		buf.Write(val)
	}

	buf.WriteString("}")

	return buf.Bytes(), nil
}

type OrderedClaimRequest struct {
	Claim   string
	Request *ClaimRequest
}

// ClaimsRequests is a request for a particular set of claims.
type ClaimsRequests struct {
	IDToken  map[string]*ClaimRequest `json:"id_token,omitempty"`
	UserInfo map[string]*ClaimRequest `json:"userinfo,omitempty"`
}

func (r *ClaimsRequests) ToOrdered() *OrderedClaimsRequests {
	requests := &OrderedClaimsRequests{}

	if len(r.IDToken) > 0 {
		requests.IDToken = OrderedClaimRequests{}

		for claim, request := range r.IDToken {
			requests.IDToken = append(requests.IDToken, OrderedClaimRequest{Claim: claim, Request: request})
		}
	}

	if len(r.UserInfo) > 0 {
		requests.UserInfo = OrderedClaimRequests{}

		for claim, request := range r.UserInfo {
			requests.UserInfo = append(requests.UserInfo, OrderedClaimRequest{Claim: claim, Request: request})
		}
	}

	return requests
}

func (r *ClaimsRequests) Signature() (signature string, err error) {
	return r.ToOrdered().Signature()
}

func (r *ClaimsRequests) Serialized() (serialized, signature string, err error) {
	return r.ToOrdered().Serialized()
}

// GetIDTokenRequests returns the IDToken value.
func (r *ClaimsRequests) GetIDTokenRequests() (requests map[string]*ClaimRequest) {
	if r == nil {
		return nil
	}

	return r.IDToken
}

// GetUserInfoRequests returns the UserInfo value.
func (r *ClaimsRequests) GetUserInfoRequests() (requests map[string]*ClaimRequest) {
	if r == nil {
		return nil
	}

	return r.UserInfo
}

// MatchesSubject returns true if this *ClaimsRequests matches the subject. i.e. if the claims parameter requires a
// specific subject and that value does not match the current value it returns false, otherwise it returns true as well
// as the subject value.
func (r *ClaimsRequests) MatchesSubject(subject string) (requested string, ok bool) {
	if r == nil {
		return "", true
	}

	return r.stringMatch(subject, ClaimSubject)
}

func (r *ClaimsRequests) MatchesIssuer(issuer *url.URL) (requested string, ok bool) {
	if r == nil {
		return "", true
	}

	return r.stringMatch(issuer.String(), ClaimIssuer)
}

func (r *ClaimsRequests) stringMatch(expected, claim string) (requested string, ok bool) {
	var request *ClaimRequest

	if r.UserInfo != nil {
		if request, ok = r.UserInfo[claim]; ok {
			if request != nil && request.Value != nil {
				if requested, ok = request.Value.(string); !ok {
					return "", false
				}

				if request.Value != expected {
					return requested, false
				}
			}
		}
	}

	if r.IDToken != nil {
		if request, ok = r.IDToken[claim]; ok {
			if request != nil && request.Value != nil {
				if requested, ok = request.Value.(string); !ok {
					return "", false
				}

				if request.Value != nil && request.Value != expected {
					return requested, false
				}
			}
		}
	}

	return requested, true
}

func (r *ClaimsRequests) ToSlice() (claims []string) {
	var essential []string

	claims, essential = r.ToSlices()

	for _, claim := range essential {
		if utils.IsStringInSlice(claim, claims) {
			continue
		}

		claims = append(claims, claim)
	}

	return claims
}

// ToSlices returns the claims in two distinct slices where the first is the requested claims i.e. optional, and the
// second is the essential claims.
func (r *ClaimsRequests) ToSlices() (claims, essential []string) {
	if r == nil {
		return nil, nil
	}

	var (
		ok      bool
		claim   string
		request *ClaimRequest
	)

	for claim, request = range r.IDToken {
		if request != nil && request.Essential {
			essential = append(essential, claim)
		} else if request, ok = r.UserInfo[claim]; ok && request != nil && request.Essential {
			essential = append(essential, claim)
		} else {
			claims = append(claims, claim)
		}
	}

	for claim, request = range r.UserInfo {
		if utils.IsStringInSlice(claim, claims) || utils.IsStringInSlice(claim, essential) {
			continue
		}

		if request != nil && request.Essential {
			essential = append(essential, claim)
		} else {
			claims = append(claims, claim)
		}
	}

	return claims, essential
}

// ClaimRequest is a request for a particular claim.
type ClaimRequest struct {
	Essential bool  `json:"essential,omitempty"`
	Value     any   `json:"value,omitempty"`
	Values    []any `json:"values,omitempty"`
}

func (r *ClaimRequest) String() (value string) {
	if r == nil {
		return ""
	}

	var parts []string

	if r.Value != nil {
		parts = append(parts, fmt.Sprintf("value '%v'", r.Value))
	}

	if r.Values != nil {
		items := make([]string, len(r.Values))

		for i, item := range r.Values {
			items[i] = fmt.Sprintf("%v", item)
		}

		parts = append(parts, fmt.Sprintf("values ['%s']", strings.Join(items, "','")))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("essential '%t'", r.Essential)
	}

	return fmt.Sprintf("%s, essential '%t'", strings.Join(parts, ", "), r.Essential)
}

// Matches is a convenience function which tests if a particular value matches this claims request.
//
//nolint:gocyclo
func (r *ClaimRequest) Matches(value any) (match bool) {
	if r == nil {
		return true
	}

	if r.Value == nil && r.Values == nil {
		return true
	}

	if f, ok := float64As(value); ok {
		return float64Match(f, r.Value, r.Values)
	}

	switch t := value.(type) {
	case bool:
		if r.Value != nil {
			if t == r.Value {
				return true
			}
		}

		if r.Values != nil {
			found := false

			for _, v := range r.Values {
				if t == v {
					found = true

					break
				}
			}

			if found {
				return true
			}
		}
	case string:
		if r.Value != nil {
			if t == r.Value {
				return true
			}
		}

		if r.Values != nil {
			found := false

			for _, v := range r.Values {
				if t == v {
					found = true

					break
				}
			}

			if found {
				return true
			}
		}
	}

	return false
}

type ClaimResolver func(attribute string) (value any, ok bool)

type ClaimsStrategy interface {
	ValidateClaimsRequests(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, requests *ClaimsRequests) (err error)
	PopulateIDTokenClaims(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, scopes, claims oauthelia2.Arguments, requests map[string]*ClaimRequest, detailer UserDetailer, updated time.Time, original, extra map[string]any) (err error)
	PopulateUserInfoClaims(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, scopes, claims oauthelia2.Arguments, requests map[string]*ClaimRequest, detailer UserDetailer, updated time.Time, original, extra map[string]any) (err error)
	PopulateClientCredentialsUserInfoClaims(ctx Context, client Client, original, extra map[string]any) (err error)
}

func NewDefaultCustomClaimsStrategy() (strategy *CustomClaimsStrategy) {
	return &CustomClaimsStrategy{
		claimsIDToken:     []string{},
		claimsAccessToken: []string{},
		scopes: map[string]map[string]string{
			ScopeProfile: {
				ClaimFullName:          expression.AttributeUserDisplayName,
				ClaimGivenName:         expression.AttributeUserGivenName,
				ClaimFamilyName:        expression.AttributeUserFamilyName,
				ClaimMiddleName:        expression.AttributeUserMiddleName,
				ClaimNickname:          expression.AttributeUserNickname,
				ClaimPreferredUsername: expression.AttributeUserUsername,
				ClaimProfile:           expression.AttributeUserProfile,
				ClaimPicture:           expression.AttributeUserPicture,
				ClaimWebsite:           expression.AttributeUserWebsite,
				ClaimGender:            expression.AttributeUserGender,
				ClaimBirthdate:         expression.AttributeUserBirthdate,
				ClaimZoneinfo:          expression.AttributeUserZoneInfo,
				ClaimLocale:            expression.AttributeUserLocale,
				ClaimUpdatedAt:         expression.AttributeUserUpdatedAt,
			},
			ScopeEmail: {
				ClaimEmail:         expression.AttributeUserEmail,
				ClaimEmailAlts:     expression.AttributeUserEmailsExtra,
				ClaimEmailVerified: expression.AttributeUserEmailVerified,
			},
			ScopePhone: {
				ClaimPhoneNumber:         expression.AttributeUserPhoneNumberRFC3966,
				ClaimPhoneNumberVerified: expression.AttributeUserPhoneNumberVerified,
			},
			ScopeAddress: {
				ClaimAddress: expression.AttributeUserAddress,
			},
			ScopeGroups: {
				ClaimGroups: expression.AttributeUserGroups,
			},
		},
	}
}

func NewCustomClaimsStrategy(client schema.IdentityProvidersOpenIDConnectClient, scopes map[string]schema.IdentityProvidersOpenIDConnectScope, policies map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy) (strategy *CustomClaimsStrategy) {
	strategy = NewDefaultCustomClaimsStrategy()

	if client.ClaimsPolicy == "" {
		return strategy
	}

	var (
		policy  schema.IdentityProvidersOpenIDConnectClaimsPolicy
		mapping schema.IdentityProvidersOpenIDConnectScope
		claim   schema.IdentityProvidersOpenIDConnectCustomClaim

		ok   bool
		name string
	)

	if policy, ok = policies[client.ClaimsPolicy]; !ok {
		return strategy
	}

	if policy.IDToken != nil {
		strategy.claimsIDToken = policy.IDToken
	}

	if policy.AccessToken != nil {
		strategy.claimsAccessToken = policy.AccessToken
	}

	for _, scope := range client.Scopes {
		if mapping, ok = scopes[scope]; !ok {
			continue
		}

		if _, ok = strategy.scopes[scope]; !ok {
			strategy.scopes[scope] = make(map[string]string)
		}

		for _, name = range mapping.Claims {
			switch name {
			case ClaimFullName:
				strategy.scopes[scope][name] = expression.AttributeUserDisplayName
			case ClaimPreferredUsername:
				strategy.scopes[scope][name] = expression.AttributeUserUsername
			case ClaimEmailAlts:
				strategy.scopes[scope][name] = expression.AttributeUserEmailsExtra
			case ClaimPhoneNumber:
				strategy.scopes[scope][name] = expression.AttributeUserPhoneNumberRFC3966
			default:
				claim = policy.CustomClaims[name]

				if claim.Attribute == "" {
					strategy.scopes[scope][name] = name
				} else {
					strategy.scopes[scope][name] = claim.Attribute
				}
			}
		}
	}

	return strategy
}

type CustomClaimsStrategy struct {
	claimsIDToken     []string
	claimsAccessToken []string
	scopes            map[string]map[string]string
}

//nolint:gocyclo
func (s *CustomClaimsStrategy) ValidateClaimsRequests(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, requests *ClaimsRequests) (err error) {
	if requests == nil {
		return nil
	}

	scopes := client.GetScopes()

	claimMatches := map[string][]string{}

	if requests.IDToken != nil {
		for claim := range requests.IDToken {
			for scope, claims := range s.scopes {
				if _, ok := claims[claim]; !ok {
					continue
				}

				if scp, ok := claimMatches[claim]; ok {
					claimMatches[claim] = append(scp, scope)
				} else {
					claimMatches[claim] = []string{scope}
				}
			}
		}
	}

	if requests.UserInfo != nil {
		for claim := range requests.UserInfo {
			for scope, claims := range s.scopes {
				if _, ok := claims[claim]; !ok {
					continue
				}

				if scp, ok := claimMatches[claim]; ok {
					claimMatches[claim] = append(scp, scope)
				} else {
					claimMatches[claim] = []string{scope}
				}
			}
		}
	}

	invalid := map[string][]string{}

claims:
	for claim, possibleScopes := range claimMatches {
		var requiredScopes []string

		for _, scope := range possibleScopes {
			if strategy(scopes, scope) {
				continue claims
			}

			requiredScopes = append(requiredScopes, scope)
		}

		for _, scope := range requiredScopes {
			if invalidClaims, ok := invalid[scope]; ok {
				invalid[scope] = append(invalidClaims, claim)
			} else {
				invalid[scope] = []string{claim}
			}
		}
	}

	if len(invalid) == 0 {
		return nil
	}

	elements := make([]string, 0, len(invalid))

	for scope, claims := range invalid {
		elements = append(elements, fmt.Sprintf("claims %s require the '%s' scope", utils.StringJoinAnd(claims), scope))
	}

	return oauthelia2.ErrInvalidRequest.WithDebugf("The authorization request contained a claims request which is not permitted to make. The %s; but these scopes are absent from the client registration.", strings.Join(elements, ", "))
}

func (s *CustomClaimsStrategy) PopulateIDTokenClaims(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, scopes, claims oauthelia2.Arguments, requests map[string]*ClaimRequest, detailer UserDetailer, updated time.Time, original, extra map[string]any) (err error) {
	resolver := ctx.GetProviderUserAttributeResolver()

	if resolver == nil {
		return oauthelia2.ErrServerError.WithDebug("The claims strategy had an error populating the ID Token Claims. Error occurred obtaining the attribute resolver.")
	}

	resolve := func(claim string) (value any, ok bool) {
		return resolver.Resolve(claim, detailer, updated)
	}

	s.populateClaimsOriginal(original, extra)
	s.populateClaimsAudience(client, original, extra)
	s.populateClaimsScoped(ctx, strategy, client, scopes, resolve, s.claimsIDToken, extra)
	s.populateClaimsRequested(ctx, strategy, client, requests, claims, resolve, extra)

	return nil
}

func (s *CustomClaimsStrategy) PopulateUserInfoClaims(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, scopes, claims oauthelia2.Arguments, requests map[string]*ClaimRequest, detailer UserDetailer, updated time.Time, original, extra map[string]any) (err error) {
	resolver := ctx.GetProviderUserAttributeResolver()

	if resolver == nil {
		return oauthelia2.ErrServerError.WithDebug("The claims strategy had an error populating the ID Token Claims. Error occurred obtaining the attribute resolver.")
	}

	resolve := func(attribute string) (value any, ok bool) {
		return resolver.Resolve(attribute, detailer, updated)
	}

	s.populateClaimsOriginalUserInfo(original, extra)
	s.populateClaimsScoped(ctx, strategy, client, scopes, resolve, nil, extra)
	s.populateClaimsRequested(ctx, strategy, client, requests, claims, resolve, extra)

	return nil
}

func (s *CustomClaimsStrategy) PopulateClientCredentialsUserInfoClaims(ctx Context, client Client, original, extra map[string]any) (err error) {
	s.populateClaimsOriginal(original, extra)
	s.populateClaimsAudience(client, original, extra)

	return nil
}

func (s *CustomClaimsStrategy) isClaimAllowed(claim string, allowed []string) (isAllowed bool) {
	if allowed == nil {
		return true
	}

	return utils.IsStringInSlice(claim, allowed)
}

func (s *CustomClaimsStrategy) populateClaimsOriginalUserInfo(original, extra map[string]any) {
	for claim, value := range original {
		switch claim {
		case ClaimSubject:
			extra[claim] = value
		default:
			continue
		}
	}
}

func (s *CustomClaimsStrategy) populateClaimsOriginal(original, extra map[string]any) {
	for claim, value := range original {
		switch claim {
		case ClaimJWTID, ClaimSessionID, ClaimAccessTokenHash, ClaimCodeHash, ClaimExpirationTime, ClaimNonce, ClaimStateHash:
			// Skip special OpenID Connect 1.0 Claims.
			continue
		case ClaimFullName, ClaimGivenName, ClaimFamilyName, ClaimMiddleName, ClaimNickname, ClaimPreferredUsername, ClaimProfile, ClaimPicture, ClaimWebsite, ClaimEmail, ClaimEmailVerified, ClaimGender, ClaimBirthdate, ClaimZoneinfo, ClaimLocale, ClaimPhoneNumber, ClaimPhoneNumberVerified, ClaimAddress:
			// Skip the standard claims.
			continue
		default:
			extra[claim] = value
		}
	}
}

func (s *CustomClaimsStrategy) populateClaimsAudience(client Client, original, extra map[string]any) {
	if clientID := client.GetID(); clientID != "" {
		audience, ok := GetAudienceFromClaims(original)

		if !ok || len(audience) == 0 {
			audience = []string{clientID}
		} else if !utils.IsStringInSlice(clientID, audience) {
			audience = append(audience, clientID)
		}

		extra[ClaimAudience] = audience
	}
}

func (s *CustomClaimsStrategy) populateClaimsScoped(_ Context, strategy oauthelia2.ScopeStrategy, client Client, scopes oauthelia2.Arguments, resolve ClaimResolver, allowed []string, extra map[string]any) {
	if resolve == nil {
		return
	}

	for scope, claims := range s.scopes {
		if !strategy(scopes, scope) {
			continue
		}

		for claim, attribute := range claims {
			s.populateClaim(client, claim, attribute, allowed, resolve, extra, nil)
		}
	}
}

func (s *CustomClaimsStrategy) populateClaimsRequested(_ Context, strategy oauthelia2.ScopeStrategy, client Client, requests map[string]*ClaimRequest, claims oauthelia2.Arguments, resolve ClaimResolver, extra map[string]any) {
	if requests == nil || resolve == nil {
		return
	}

	scopes := client.GetScopes()

claim:
	for claim, request := range requests {
		for scope, claimSet := range s.scopes {
			if !strategy(scopes, scope) {
				continue
			}

			if (request == nil || !request.Essential) && !claims.Has(claim) {
				continue
			}

			attribute, ok := claimSet[claim]

			if !ok {
				continue
			}

			s.populateClaim(client, claim, attribute, nil, resolve, extra, request)

			continue claim
		}
	}
}

func (s *CustomClaimsStrategy) populateClaim(_ Client, claim, attribute string, allowed []string, resolve ClaimResolver, extra map[string]any, request *ClaimRequest) {
	if !s.isClaimAllowed(claim, allowed) {
		return
	}

	value, ok := resolve(attribute)

	if !ok || value == nil {
		return
	}

	var str string

	if str, ok = value.(string); ok {
		if str == "" {
			return
		}
	}

	if request != nil {
		if !request.Matches(value) {
			return
		}
	}

	extra[claim] = value
}

// GrantScopeAudienceConsent grants all scopes and audience values that have received consent.
func GrantScopeAudienceConsent(ar oauthelia2.Requester, consent *model.OAuth2ConsentSession) {
	if ar == nil || consent == nil {
		return
	}

	for _, scope := range consent.GrantedScopes {
		ar.GrantScope(scope)
	}

	for _, audience := range consent.GrantedAudience {
		ar.GrantAudience(audience)
	}
}

// GetAudienceFromClaims retrieves the various formats of the 'aud' claim and returns them as a []string.
func GetAudienceFromClaims(claims map[string]any) (audience []string, ok bool) {
	var aud any

	if aud, ok = claims[ClaimAudience]; ok {
		switch v := aud.(type) {
		case string:
			if v == "" {
				break
			}

			ok = true

			audience = []string{v}
		case []any:
			var value string

			for _, a := range v {
				if value, ok = a.(string); !ok {
					return nil, false
				}

				audience = append(audience, value)
			}

			ok = true
		case []string:
			ok = true

			audience = v
		}
	}

	return audience, ok
}
