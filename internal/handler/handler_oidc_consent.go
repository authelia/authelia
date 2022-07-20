package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/middleware"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// OpenIDConnectConsentGET handles requests to provide consent for OpenID Connect.
func OpenIDConnectConsentGET(ctx *middleware.AutheliaCtx) {
	userSession, consent, client, handled := oidcConsentGetSessionsAndClient(ctx)
	if handled {
		return
	}

	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		ctx.Logger.Errorf("Unable to perform consent without sufficient authentication for user '%s' and client id '%s'", userSession.Username, consent.ClientID)
		ctx.ReplyForbidden()

		return
	}

	if err := ctx.SetJSONBody(client.GetConsentResponseBody(consent)); err != nil {
		ctx.Error(fmt.Errorf("unable to set JSON body: %v", err), "Operation failed")
	}
}

// OpenIDConnectConsentPOST handles consent responses for OpenID Connect.
func OpenIDConnectConsentPOST(ctx *middleware.AutheliaCtx) {
	var (
		body oidc.ConsentPostRequestBody
		err  error
	)

	if err = json.Unmarshal(ctx.Request.Body(), &body); err != nil {
		ctx.Logger.Errorf("Failed to parse JSON body in consent POST: %+v", err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	userSession, consent, client, handled := oidcConsentGetSessionsAndClient(ctx)
	if handled {
		return
	}

	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		ctx.Logger.Debugf("Insufficient permissions to give consent during POST current level: %d, require 2FA: %d", userSession.AuthenticationLevel, client.Policy)
		ctx.ReplyForbidden()

		return
	}

	if consent.ClientID != body.ClientID {
		ctx.Logger.Errorf("User '%s' consented to scopes of another client (%s) than expected (%s). Beware this can be a sign of attack",
			userSession.Username, body.ClientID, consent.ClientID)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		externalRootURL string
		authorized      = true
	)

	switch body.AcceptOrReject {
	case accept:
		if externalRootURL, err = ctx.ExternalRootURL(); err != nil {
			ctx.Logger.Errorf("Could not determine the external URL during consent session processing with challenge id '%s' for user '%s': %v", consent.ChallengeID.String(), userSession.Username, err)
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		if body.PreConfigure {
			if client.PreConfiguredConsentDuration == nil {
				ctx.Logger.Warnf("Consent session with challenge id '%s' for user '%s': consent pre-configuration was requested and was ignored because it is not permitted on this client", consent.ChallengeID.String(), userSession.Username)
			} else {
				expiresAt := time.Now().Add(*client.PreConfiguredConsentDuration)
				consent.ExpiresAt = &expiresAt

				ctx.Logger.Debugf("Consent session with challenge id '%s' for user '%s': pre-configured and set to expire at %v", consent.ChallengeID.String(), userSession.Username, consent.ExpiresAt)
			}
		}

		consent.GrantedScopes = consent.RequestedScopes
		consent.GrantedAudience = consent.RequestedAudience

		if !utils.IsStringInSlice(consent.ClientID, consent.GrantedAudience) {
			consent.GrantedAudience = append(consent.GrantedAudience, consent.ClientID)
		}
	case reject:
		authorized = false
	default:
		ctx.Logger.Warnf("User '%s' tried to reply to consent with an unexpected verb '%s'", userSession.Username, body.AcceptOrReject)
		ctx.ReplyBadRequest()

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, *consent, authorized); err != nil {
		ctx.Logger.Errorf("Failed to save the consent session response to the database: %+v", err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	response := oidc.ConsentPostResponseBody{RedirectURI: fmt.Sprintf("%s%s?%s", externalRootURL, oidc.AuthorizationPath, consent.Form)}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Error(fmt.Errorf("unable to set JSON body in response"), "Operation failed")
	}
}

func oidcConsentGetSessionsAndClient(ctx *middleware.AutheliaCtx) (userSession session.UserSession, consent *model.OAuth2ConsentSession, client *oidc.Client, handled bool) {
	var (
		err error
	)

	userSession = ctx.GetSession()

	if userSession.ConsentChallengeID == nil {
		ctx.Logger.Errorf("Cannot consent for user '%s' when OIDC consent session has not been initiated", userSession.Username)
		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, *userSession.ConsentChallengeID); err != nil {
		ctx.Logger.Errorf("Unable to load consent session with challenge id '%s': %v", userSession.ConsentChallengeID.String(), err)
		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	if client, err = ctx.Providers.OpenIDConnect.Store.GetFullClient(consent.ClientID); err != nil {
		ctx.Logger.Errorf("Unable to find related client configuration with name '%s': %v", consent.ClientID, err)
		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	return userSession, consent, client, false
}
