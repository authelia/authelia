package handlers

import (
	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

func oidcGrantRequests(strategy oauthelia2.ScopeStrategy, ar oauthelia2.AuthorizeRequester, consent *model.OAuth2ConsentSession, detailer oidc.UserDetailer, claims *oidc.ClaimRequests) (extraClaims map[string]any) {
	extraClaims = map[string]any{}

	var scopes oauthelia2.Arguments

	if ar != nil {
		for _, scope := range consent.GrantedScopes {
			ar.GrantScope(scope)
		}

		for _, audience := range consent.GrantedAudience {
			ar.GrantAudience(audience)
		}

		scopes = ar.GetGrantedScopes()
	}

	if claims == nil || claims.IDToken == nil {
		return extraClaims
	}

	for claim, request := range claims.IDToken {
		switch claim {
		case oidc.ClaimGroups:
			oidcApplyRequestedClaim(strategy, scopes, oidc.ScopeGroups, oidc.ClaimGroups, detailer.GetGroups(), request, extraClaims)
		case oidc.ClaimPreferredUsername:
			oidcApplyRequestedClaim(strategy, scopes, oidc.ScopeProfile, oidc.ClaimPreferredUsername, detailer.GetUsername(), request, extraClaims)
		case oidc.ClaimFullName:
			oidcApplyRequestedClaim(strategy, scopes, oidc.ScopeProfile, oidc.ClaimFullName, detailer.GetDisplayName(), request, extraClaims)
		case oidc.ClaimPreferredEmail:
			emails := detailer.GetEmails()

			if len(emails) == 0 {
				continue
			}

			oidcApplyRequestedClaim(strategy, scopes, oidc.ScopeEmail, oidc.ClaimPreferredEmail, emails[0], request, extraClaims)
		case oidc.ClaimEmailAlts:
			emails := detailer.GetEmails()

			if len(emails) <= 1 {
				continue
			}

			oidcApplyRequestedClaim(strategy, scopes, oidc.ScopeEmail, oidc.ClaimEmailAlts, emails[1:], request, extraClaims)
		case oidc.ClaimEmailVerified:
			if !strategy(scopes, oidc.ScopeEmail) {
				continue
			}

			oidcApplyRequestedClaim(strategy, scopes, oidc.ScopeEmail, oidc.ClaimEmailVerified, true, request, extraClaims)
		}
	}

	return extraClaims
}

func oidcApplyRequestedClaim(strategy oauthelia2.ScopeStrategy, scopes oauthelia2.Arguments, scope string, claim string, value any, request *oidc.ClaimsRequest, extraClaims map[string]any) {
	if !strategy(scopes, scope) {
		return
	}

	if request == nil || request.Value == nil || request.Values == nil {
		extraClaims[claim] = value

		return
	}

	if request.Matches(value) {
		extraClaims[claim] = value
	}
}

func oidcApplyScopeClaims(claims map[string]any, scopes []string, detailer oidc.UserDetailer) {
	for _, scope := range scopes {
		switch scope {
		case oidc.ScopeGroups:
			claims[oidc.ClaimGroups] = detailer.GetGroups()
		case oidc.ScopeProfile:
			claims[oidc.ClaimPreferredUsername] = detailer.GetUsername()
			claims[oidc.ClaimFullName] = detailer.GetDisplayName()
		case oidc.ScopeEmail:
			if emails := detailer.GetEmails(); len(emails) != 0 {
				claims[oidc.ClaimPreferredEmail] = emails[0]
				if len(emails) > 1 {
					claims[oidc.ClaimEmailAlts] = emails[1:]
				}

				// TODO (james-d-elliott): actually verify emails and record that information.
				claims[oidc.ClaimEmailVerified] = true
			}
		}
	}
}

func oidcGetAudience(claims map[string]any) (audience []string, ok bool) {
	var aud any

	if aud, ok = claims[oidc.ClaimAudience]; ok {
		switch v := aud.(type) {
		case string:
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

func oidcApplyUserInfoClaims(clientID string, scopes oauthelia2.Arguments, originalClaims, claims map[string]any, resolver oidcDetailResolver) {
	for claim, value := range originalClaims {
		switch claim {
		case oidc.ClaimJWTID, oidc.ClaimSessionID, oidc.ClaimAccessTokenHash, oidc.ClaimCodeHash, oidc.ClaimExpirationTime, oidc.ClaimNonce, oidc.ClaimStateHash:
			// Skip special OpenID Connect 1.0 Claims.
			continue
		case oidc.ClaimPreferredUsername, oidc.ClaimPreferredEmail, oidc.ClaimEmailVerified, oidc.ClaimEmailAlts, oidc.ClaimGroups, oidc.ClaimFullName:
			continue
		default:
			claims[claim] = value
		}
	}

	audience, ok := oidcGetAudience(originalClaims)

	if !ok || len(audience) == 0 {
		audience = []string{clientID}
	} else if !utils.IsStringInSlice(clientID, audience) {
		audience = append(audience, clientID)
	}

	claims[oidc.ClaimAudience] = audience

	oidcApplyUserInfoDetailsClaims(scopes, claims, resolver)
}

func oidcApplyUserInfoDetailsClaims(scopes oauthelia2.Arguments, claims map[string]any, resolver oidcDetailResolver) {
	var (
		detailer oidc.UserDetailer
		subject  uuid.UUID
		ok       bool
		err      error
	)

	if subject, ok = oidcApplyUserInfoDetailsClaimsGetSubject(scopes, claims); !ok {
		return
	}

	if detailer, err = resolver(subject); err != nil {
		return
	}

	oidcApplyScopeClaims(claims, scopes, detailer)
}

func oidcApplyUserInfoDetailsClaimsGetSubject(scopes oauthelia2.Arguments, claims map[string]any) (subject uuid.UUID, ok bool) {
	if !scopes.HasOneOf(oidc.ScopeProfile, oidc.ScopeEmail, oidc.ScopeGroups) {
		return uuid.UUID{}, false
	}

	var (
		raw   any
		claim string
		err   error
	)

	if raw, ok = claims[oidc.ClaimSubject]; !ok {
		return uuid.UUID{}, false
	}

	if claim, ok = raw.(string); !ok {
		return uuid.UUID{}, false
	}

	if subject, err = uuid.Parse(claim); err != nil {
		return uuid.UUID{}, false
	}

	return subject, true
}

func oidcCtxDetailResolver(ctx *middlewares.AutheliaCtx) oidcDetailResolver {
	return func(subject uuid.UUID) (detailer oidc.UserDetailer, err error) {
		var (
			identifier *model.UserOpaqueIdentifier
			details    *authentication.UserDetails
		)

		if identifier, err = ctx.Providers.StorageProvider.LoadUserOpaqueIdentifier(ctx, subject); err != nil {
			return nil, err
		}

		if details, err = ctx.Providers.UserProvider.GetDetails(identifier.Username); err != nil {
			return nil, err
		}

		return details, nil
	}
}

type oidcDetailResolver func(subject uuid.UUID) (detailer oidc.UserDetailer, err error)
