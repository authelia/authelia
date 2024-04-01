package oidc

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"

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
	if r == nil {
		return "", true
	}

	var request *ClaimRequest

	if r.UserInfo != nil {
		if request, ok = r.UserInfo[ClaimSubject]; ok {
			requested, _ = request.Value.(string)

			if request.Value != nil && request.Value != subject {
				return requested, false
			}
		}
	}

	if r.IDToken != nil {
		if request, ok = r.IDToken[ClaimSubject]; ok {
			requested, _ = request.Value.(string)

			if request.Value != nil && request.Value != subject {
				return requested, false
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
		return false
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

// GrantScopeAudienceConsent grants all scopes and audience values that have received consent.
func GrantScopeAudienceConsent(ar oauthelia2.AuthorizeRequester, consent *model.OAuth2ConsentSession) {
	if ar != nil {
		for _, scope := range consent.GrantedScopes {
			ar.GrantScope(scope)
		}

		for _, audience := range consent.GrantedAudience {
			ar.GrantAudience(audience)
		}
	}
}

// GrantClaimRequests grants all claims the client has requested provided it's authorized to request them.
//
//nolint:gocyclo
func GrantClaimRequests(strategy oauthelia2.ScopeStrategy, client Client, requests map[string]*ClaimRequest, detailer UserDetailer, extra map[string]any) {
	if requests == nil {
		return
	}

	for claim, request := range requests {
		switch claim {
		case ClaimFullName:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimFullName, detailer.GetDisplayName(), request, extra)
		case ClaimGivenName:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimGivenName, detailer.GetGivenName(), request, extra)
		case ClaimFamilyName:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimFamilyName, detailer.GetFamilyName(), request, extra)
		case ClaimMiddleName:
			grantRequestedClaim(strategy, client, ScopeProfile, ClaimMiddleName, detailer.GetMiddleName(), request, extra)
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
			grantRequestedClaim(strategy, client, ScopePhone, ClaimPhoneNumber, detailer.GetOpenIDConnectPhoneNumber(), request, extra)
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

// GrantScopedClaims copies the extra claims from the ID Token that may be useful while excluding
// OpenID Connect 1.0 Special Claims, OpenID Connect 1.0 Scope-based Claims which should be granted by GrantClaimRequests.
//
//nolint:gocyclo
func GrantScopedClaims(strategy oauthelia2.ScopeStrategy, client Client, scopes oauthelia2.Arguments, detailer UserDetailer, original, claims map[string]any) {
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

	if strategy(scopes, ScopeProfile) {
		doClaimsApplyPossibleStringValue(claims, ClaimFullName, detailer.GetDisplayName())
		doClaimsApplyPossibleStringValue(claims, ClaimGivenName, detailer.GetGivenName())
		doClaimsApplyPossibleStringValue(claims, ClaimFamilyName, detailer.GetFamilyName())
		doClaimsApplyPossibleStringValue(claims, ClaimMiddleName, detailer.GetMiddleName())
		doClaimsApplyPossibleStringValue(claims, ClaimNickname, detailer.GetNickname())
		doClaimsApplyPossibleStringValue(claims, ClaimPreferredUsername, detailer.GetUsername())
		doClaimsApplyPossibleStringValue(claims, ClaimProfile, detailer.GetProfile())
		doClaimsApplyPossibleStringValue(claims, ClaimPicture, detailer.GetPicture())
		doClaimsApplyPossibleStringValue(claims, ClaimWebsite, detailer.GetWebsite())
		doClaimsApplyPossibleStringValue(claims, ClaimGender, detailer.GetGender())
		doClaimsApplyPossibleStringValue(claims, ClaimBirthdate, detailer.GetBirthdate())
		doClaimsApplyPossibleStringValue(claims, ClaimZoneinfo, detailer.GetZoneInfo())
		doClaimsApplyPossibleStringValue(claims, ClaimLocale, detailer.GetLocale())
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
		address := ClaimAddressFromDetailer(detailer)

		if address != nil {
			claims[ClaimAddress] = address
		}
	}

	if strategy(scopes, ScopePhone) {
		doClaimsApplyPossibleStringValue(claims, ClaimPhoneNumber, detailer.GetOpenIDConnectPhoneNumber())
		claims[ClaimPhoneNumberVerified] = false
	}

	if strategy(scopes, ScopeGroups) {
		claims[ClaimGroups] = detailer.GetGroups()
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
