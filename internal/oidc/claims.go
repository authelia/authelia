package oidc

import (
	"encoding/json"
	"fmt"
	"net/url"
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

// ClaimsRequests is a request for a particular set of claims.
type ClaimsRequests struct {
	IDToken  map[string]*ClaimRequest `json:"id_token,omitempty"`
	UserInfo map[string]*ClaimRequest `json:"userinfo,omitempty"`
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
	return r.stringMatch(subject, ClaimSubject)
}

func (r *ClaimsRequests) MatchesIssuer(issuer *url.URL) (requested string, ok bool) {
	return r.stringMatch(issuer.String(), ClaimIssuer)
}

func (r *ClaimsRequests) stringMatch(expected, claim string) (requested string, ok bool) {
	if r == nil {
		return "", true
	}

	var request *ClaimRequest

	if r.UserInfo != nil {
		if request, ok = r.UserInfo[claim]; ok {
			if request.Value != nil {
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
			if request.Value != nil {
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
		} else {
			if request, ok = r.UserInfo[claim]; ok && request != nil && request.Essential {
				essential = append(essential, claim)
			} else {
				claims = append(claims, claim)
			}
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

	switch t := value.(type) {
	case int:
		if r.Value != nil {
			if float64(t) != r.Value && t != r.Value {
				return false
			}
		}
	case int64:
		if r.Value != nil {
			if float64(t) != r.Value && t != r.Value {
				return false
			}
		}

		if r.Values != nil {
			found := false

			for _, v := range r.Values {
				if float64(t) == v || t == v {
					found = true

					break
				}
			}

			if !found {
				return false
			}
		}
	case float64:
		if r.Value != nil {
			if t != r.Value {
				return false
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

			if !found {
				return false
			}
		}
	case string:
		if r.Value != nil {
			if t != r.Value {
				return false
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

			if !found {
				return false
			}
		}
	case []string:
		if r.Value != nil {
			if !utils.IsStringInSlice(fmt.Sprintf("%s", value), t) {
				return false
			}
		}

		if r.Values != nil {
			found := false

		outer:
			for _, v := range r.Values {
				for _, w := range t {
					if v == w {
						found = true

						break outer
					}
				}
			}

			if !found {
				return false
			}
		}
	}

	return true
}

type ClaimResolver func(attribute string) (value any, ok bool)

type ClaimsStrategy interface {
	PopulateIDTokenClaims(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, scopes oauthelia2.Arguments, requests map[string]*ClaimRequest, detailer UserDetailer, updated time.Time, original, extra map[string]any)
	PopulateUserInfoClaims(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, scopes oauthelia2.Arguments, requests map[string]*ClaimRequest, detailer UserDetailer, updated time.Time, original, extra map[string]any)
	PopulateClientCredentialsUserInfoClaims(ctx Context, client Client, original, extra map[string]any)
}

func NewCustomClaimsStrategy(client schema.IdentityProvidersOpenIDConnectClient, scopes map[string]schema.IdentityProvidersOpenIDConnectScope, policies map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy) (strategy *CustomClaimsStrategy) {
	strategy = &CustomClaimsStrategy{
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

		strategy.scopes[scope] = make(map[string]string)

		for _, name = range mapping.Claims {
			if claim, ok = policy.CustomClaims[name]; !ok {
				continue
			}

			if claim.Attribute == "" {
				strategy.scopes[scope][name] = name
			} else {
				strategy.scopes[scope][name] = claim.Attribute
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

func (s *CustomClaimsStrategy) PopulateIDTokenClaims(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, scopes oauthelia2.Arguments, requests map[string]*ClaimRequest, detailer UserDetailer, updated time.Time, original, extra map[string]any) {
	resolver := ctx.GetProviderUserAttributeResolver()

	resolve := func(claim string) (value any, ok bool) {
		return resolver.Resolve(claim, detailer, updated)
	}

	s.populateClaimsOriginal(original, extra)
	s.populateClaimsAudience(client, original, extra)
	s.populateClaimsScoped(ctx, strategy, scopes, resolve, s.claimsIDToken, extra)
	s.populateClaimsRequested(ctx, strategy, client.GetScopes(), requests, resolve, extra)
}

func (s *CustomClaimsStrategy) PopulateUserInfoClaims(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, scopes oauthelia2.Arguments, requests map[string]*ClaimRequest, detailer UserDetailer, updated time.Time, original, extra map[string]any) {
	resolver := ctx.GetProviderUserAttributeResolver()

	resolve := func(attribute string) (value any, ok bool) {
		return resolver.Resolve(attribute, detailer, updated)
	}

	s.populateClaimsOriginalUserInfo(original, extra)
	s.populateClaimsScoped(ctx, strategy, scopes, resolve, nil, extra)
	s.populateClaimsRequested(ctx, strategy, client.GetScopes(), requests, resolve, extra)
}

func (s *CustomClaimsStrategy) PopulateClientCredentialsUserInfoClaims(ctx Context, client Client, original, extra map[string]any) {
	s.populateClaimsOriginal(original, extra)
	s.populateClaimsAudience(client, original, extra)
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

func (s *CustomClaimsStrategy) populateClaimsScoped(ctx Context, strategy oauthelia2.ScopeStrategy, scopes oauthelia2.Arguments, resolve ClaimResolver, allowed []string, extra map[string]any) {
	if resolve == nil {
		return
	}

	for scope, claims := range s.scopes {
		if !strategy(scopes, scope) {
			continue
		}

		for claim, attribute := range claims {
			s.populateClaim(claim, attribute, allowed, resolve, extra, nil)
		}
	}
}

func (s *CustomClaimsStrategy) populateClaimsRequested(ctx Context, strategy oauthelia2.ScopeStrategy, scopes oauthelia2.Arguments, requests map[string]*ClaimRequest, resolve ClaimResolver, extra map[string]any) {
	if requests == nil || resolve == nil {
		return
	}

claim:
	for claim, request := range requests {
		for scope, claims := range s.scopes {
			if !strategy(scopes, scope) {
				continue
			}

			attribute, ok := claims[claim]

			if !ok {
				continue
			}

			s.populateClaim(claim, attribute, nil, resolve, extra, request)

			continue claim
		}

		// TODO: Maybe return error if the claim is not permitted.
	}
}

func (s *CustomClaimsStrategy) populateClaim(claim, attribute string, allowed []string, resolve ClaimResolver, extra map[string]any, request *ClaimRequest) {
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

	if strings.Contains(claim, ".") {
		doClaimResolveApplyMultiLevel(claim, value, extra)

		return
	}

	extra[claim] = value
}

// GrantScopeAudienceConsent grants all scopes and audience values that have received consent.
func GrantScopeAudienceConsent(ar oauthelia2.Requester, consent *model.OAuth2ConsentSession) {
	if ar != nil {
		for _, scope := range consent.GrantedScopes {
			ar.GrantScope(scope)
		}

		for _, audience := range consent.GrantedAudience {
			ar.GrantAudience(audience)
		}
	}
}

/*

// GrantClaimsRequested grants all claims the client has requested provided it's authorized to request them.
//
//nolint:gocyclo
func GrantClaimsRequested(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, requests map[string]*ClaimRequest, detailer UserDetailer, extra map[string]any) {
	if requests == nil {
		return
	}

	resolver := ctx.GetProviderUserAttributeResolver()

	for claim, request := range requests {
		switch claim {
		case ClaimFullName:
			grantRequestedClaimEx(strategy, client, ScopeProfile, ClaimFullName, expression.AttributeUserDisplayName, resolver, detailer, request, extra)
		case ClaimGivenName:
			grantRequestedClaimEx(strategy, client, ScopeProfile, ClaimGivenName, expression.AttributeUserGivenName, resolver, detailer, request, extra)
		case ClaimFamilyName:
			grantRequestedClaimEx(strategy, client, ScopeProfile, ClaimFamilyName, expression.AttributeUserFamilyName, resolver, detailer, request, extra)
		case ClaimMiddleName:
			//grantRequestedClaim(strategy, client, ScopeProfile, ClaimMiddleName, expression.AttributeUserDisplayName, resolver, detailer, request, extra)
		case ClaimNickname:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimNickname, detailer.GetNickname(), request, extra)
		case ClaimPreferredUsername:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimPreferredUsername, detailer.GetUsername(), request, extra)
		case ClaimProfile:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimProfile, detailer.GetProfile(), request, extra)
		case ClaimPicture:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimPicture, detailer.GetPicture(), request, extra)
		case ClaimWebsite:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimWebsite, detailer.GetWebsite(), request, extra)
		case ClaimEmail:
			emails := detailer.GetEmails()

			if len(emails) == 0 {
				continue
			}

			grantRequestedClaim(strategy, client, ScopeEmail, ClaimEmail, emails[0], request, extra)
		case ClaimEmailVerified:
			if !strategy(client.GetScopes(), ScopeEmail) {
				continue
			}

			grantRequestedClaim(strategy, client, ScopeEmail, ClaimEmailVerified, true, request, extra)
		case ClaimGender:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimGender, detailer.GetGender(), request, extra)
		case ClaimBirthdate:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimBirthdate, detailer.GetBirthdate(), request, extra)
		case ClaimZoneinfo:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimZoneinfo, detailer.GetZoneInfo(), request, extra)
		case ClaimLocale:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimLocale, detailer.GetLocale(), request, extra)
		case ClaimPhoneNumber:
			grantRequestedClaim(strategy, client, ScopePhone, ClaimPhoneNumber, detailer.GetPhoneNumberRFC3966(), request, extra)
		case ClaimPhoneNumberVerified:
			grantRequestedClaim(strategy, client, ScopePhone, ClaimPhoneNumberVerified, false, request, extra)
		case ClaimAddress:
			if _, ok := extra[ClaimAddress]; ok {
				continue
			}

			if !strategy(client.GetScopes(), ScopeAddress) {
				continue
			}

			address := ClaimAddressFromDetailer(detailer)

			if address != nil {
				extra[ClaimAddress] = address
			}
		case ClaimUpdatedAt:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimUpdatedAt, time.Now().Unix(), request, extra)
		case ClaimEmailAlts:
			emails := detailer.GetEmails()

			if len(emails) <= 1 {
				continue
			}

			grantRequestedClaim(strategy, client, ScopeEmail, ClaimEmailAlts, emails[1:], request, extra)
		case ClaimGroups:
			grantRequestedClaim(strategy, client, ScopeGroups, ClaimGroups, detailer.GetGroups(), request, extra)
		}
	}
}

func grantRequestedClaim(strategy oauthelia2.ScopeStrategy, client Client, scope string, claim string, value any, request *ClaimRequest, extra map[string]any) {
	if _, ok := extra[claim]; ok {
		return
	}

	// Prevent clients from accessing claims they are NOT entitled to even request.
	if !strategy(client.GetScopes(), scope) {
		return
	}

	if request == nil || request.Value == nil || request.Values == nil {
		extra[claim] = value

		return
	}

	if request.Matches(value) {
		extra[claim] = value
	}
}

func grantRequestedClaimEx(strategy oauthelia2.ScopeStrategy, client Client, scope, claim, attribute string, resolver expression.UserAttributeResolver, detailer UserDetailer, request *ClaimRequest, extra map[string]any) {
	value, ok := resolver.Resolve(attribute, detailer)
	if !ok {
		return
	}

	if _, ok = extra[claim]; ok {
		return
	}

	// Prevent clients from accessing claims they are NOT entitled to even request.
	if !strategy(client.GetScopes(), scope) {
		return
	}

	if request == nil || request.Value == nil || request.Values == nil {
		extra[claim] = value

		return
	}

	if request.Matches(value) {
		extra[claim] = value
	}
}

// GrantClaimsScoped copies the extra claims from the ID Token that may be useful while excluding
// OpenID Connect 1.0 Special Claims, OpenID Connect 1.0 Scope-based Claims which should be granted by GrantClaimsRequested.
//
//nolint:gocyclo
func GrantClaimsScoped(ctx Context, strategy oauthelia2.ScopeStrategy, client Client, scopes oauthelia2.Arguments, detailer UserDetailer, original, claims map[string]any) {
	for claim, value := range original {
		switch claim {
		case ClaimJWTID, ClaimSessionID, ClaimAccessTokenHash, ClaimCodeHash, ClaimExpirationTime, ClaimNonce, ClaimStateHash:
			// Skip special OpenID Connect 1.0 Claims.
			continue
		case ClaimFullName, ClaimGivenName, ClaimFamilyName, ClaimMiddleName, ClaimNickname, ClaimPreferredUsername, ClaimProfile, ClaimPicture, ClaimWebsite, ClaimEmail, ClaimEmailVerified, ClaimGender, ClaimBirthdate, ClaimZoneinfo, ClaimLocale, ClaimPhoneNumber, ClaimPhoneNumberVerified, ClaimAddress:
			// Skip the standard claims.
			continue
		default:
			claims[claim] = value
		}
	}

	if clientID := client.GetID(); clientID != "" {
		audience, ok := GetAudienceFromClaims(original)

		if !ok || len(audience) == 0 {
			audience = []string{clientID}
		} else if !utils.IsStringInSlice(clientID, audience) {
			audience = append(audience, clientID)
		}

		claims[ClaimAudience] = audience
	}

	if detailer == nil {
		return
	}

	resolver := ctx.GetProviderUserAttributeResolver()

	if strategy(scopes, ScopeProfile) {
		doClaimResolveApply(claims, ClaimFullName, expression.AttributeUserDisplayName, resolver, detailer)
		doClaimResolveApply(claims, ClaimGivenName, expression.AttributeUserGivenName, resolver, detailer)
		doClaimResolveApply(claims, ClaimFamilyName, expression.AttributeUserFamilyName, resolver, detailer)
		doClaimResolveApply(claims, ClaimMiddleName, expression.AttributeUserMiddleName, resolver, detailer)
		doClaimResolveApply(claims, ClaimNickname, expression.AttributeUserNickname, resolver, detailer)
		doClaimResolveApply(claims, ClaimPreferredUsername, expression.AttributeUserUsername, resolver, detailer)
		doClaimResolveApply(claims, ClaimProfile, expression.AttributeUserProfile, resolver, detailer)
		doClaimResolveApply(claims, ClaimPicture, expression.AttributeUserPicture, resolver, detailer)
		doClaimResolveApply(claims, ClaimWebsite, expression.AttributeUserWebsite, resolver, detailer)
		doClaimResolveApply(claims, ClaimGender, expression.AttributeUserGender, resolver, detailer)
		doClaimResolveApply(claims, ClaimBirthdate, expression.AttributeUserBirthdate, resolver, detailer)
		doClaimResolveApply(claims, ClaimZoneinfo, expression.AttributeUserZoneInfo, resolver, detailer)
		doClaimResolveApply(claims, ClaimLocale, expression.AttributeUserLocale, resolver, detailer)

		claims[ClaimUpdatedAt] = time.Now().Unix()
	}

	if strategy(scopes, ScopeEmail) {
		switch emails := detailer.GetEmails(); len(emails) {
		case 1:
			claims[ClaimEmail] = emails[0]
			claims[ClaimEmailVerified] = true
		case 0:
			break
		default:
			claims[ClaimEmail] = emails[0]
			claims[ClaimEmailAlts] = emails[1:]
			claims[ClaimEmailVerified] = true
		}
	}

	if strategy(scopes, ScopeAddress) {
		doClaimResolveApply(claims, ClaimAddress+".street_address", expression.AttributeUserStreetAddress, resolver, detailer)
		doClaimResolveApply(claims, ClaimAddress+".locality", expression.AttributeUserLocality, resolver, detailer)
		doClaimResolveApply(claims, ClaimAddress+".region", expression.AttributeUserRegion, resolver, detailer)
		doClaimResolveApply(claims, ClaimAddress+".postal_code", expression.AttributeUserPostalCode, resolver, detailer)
		doClaimResolveApply(claims, ClaimAddress+".country", expression.AttributeUserCountry, resolver, detailer)
	}

	if strategy(scopes, ScopePhone) {
		doClaimResolveApply(claims, ClaimPhoneNumber, expression.AttributeUserPhoneNumberRFC3966, resolver, detailer)
		claims[ClaimPhoneNumberVerified] = false
	}

	if strategy(scopes, ScopeGroups) {
		doClaimResolveApply(claims, ClaimGroups, expression.AttributeUserGroups, resolver, detailer)
	}
}

*/

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

/*

func ClaimAddressFromDetailer(detailer UserDetailer) (claim map[string]any) {
	claim = map[string]any{}

	doClaimsApplyPossibleStringValue(claim, "street_address", detailer.GetStreetAddress())
	doClaimsApplyPossibleStringValue(claim, "locality", detailer.GetLocality())
	doClaimsApplyPossibleStringValue(claim, "region", detailer.GetRegion())
	doClaimsApplyPossibleStringValue(claim, "postal_code", detailer.GetPostalCode())
	doClaimsApplyPossibleStringValue(claim, "country", detailer.GetCountry())

	if len(claim) == 0 {
		return nil
	}

	return
}

func doClaimsApplyPossibleStringValue(claims map[string]any, name, value string) {
	if value != "" {
		claims[name] = value
	}
}


func doClaimResolveApply(claims map[string]any, claim, attribute string, resolver expression.UserAttributeResolver, detailer UserDetailer) {
	value, ok := resolver.Resolve(attribute, detailer)

	if !ok {
		return
	}

	if str, ok := value.(string); ok {
		if str == "" {
			return
		}
	}

	if strings.Contains(claim, ".") {
		doClaimResolveApplyMultiLevel(claims, claim, value)

		return
	}

	claims[claim] = value
}
*/

func doClaimResolveApplyMultiLevel(path string, value any, extra map[string]any) {
	keys := strings.Split(path, ".")
	final := keys[len(keys)-1]
	current := extra

	for _, key := range keys[:len(keys)-1] {
		if _, ok := current[key]; !ok {
			current[key] = make(map[string]any)
		}

		if next, ok := current[key].(map[string]any); ok {
			current = next
		} else {
			return
		}
	}

	if _, exists := current[final]; !exists {
		current[final] = value
	}
}
