package handlers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// OpenIDConnectConsentGET handles requests to provide consent for OpenID Connect.
func OpenIDConnectConsentGET(ctx *middlewares.AutheliaCtx) {
	var (
		consentID uuid.UUID
		err       error
	)

	if consentID, err = uuid.Parse(string(ctx.RequestCtx.QueryArgs().PeekBytes(queryArgConsentID))); err != nil {
		ctx.Logger.Errorf("Unable to convert '%s' into a UUID: %+v", ctx.RequestCtx.QueryArgs().PeekBytes(queryArgConsentID), err)
		ctx.ReplyForbidden()

		return
	}

	var (
		consent *model.OAuth2ConsentSession
		client  *oidc.Client
		handled bool
	)

	if _, consent, client, handled = oidcConsentGetSessionsAndClient(ctx, consentID); handled {
		return
	}

	if err = ctx.SetJSONBody(client.GetConsentResponseBody(consent)); err != nil {
		ctx.Error(fmt.Errorf("unable to set JSON body: %v", err), "Operation failed")
	}
}

// OpenIDConnectConsentPOST handles consent responses for OpenID Connect.
func OpenIDConnectConsentPOST(ctx *middlewares.AutheliaCtx) {
	var (
		consentID uuid.UUID
		bodyJSON  oidc.ConsentPostRequestBody
		err       error
	)

	if err = json.Unmarshal(ctx.Request.Body(), &bodyJSON); err != nil {
		ctx.Logger.Errorf("Failed to parse JSON bodyJSON in consent POST: %+v", err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if consentID, err = uuid.Parse(bodyJSON.ConsentID); err != nil {
		ctx.Logger.Errorf("Unable to convert '%s' into a UUID: %+v", ctx.RequestCtx.QueryArgs().PeekBytes(queryArgConsentID), err)
		ctx.ReplyForbidden()

		return
	}

	var (
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		client      *oidc.Client
		handled     bool
	)

	if userSession, consent, client, handled = oidcConsentGetSessionsAndClient(ctx, consentID); handled {
		return
	}

	if consent.ClientID != bodyJSON.ClientID {
		ctx.Logger.Errorf("User '%s' consented to scopes of another client (%s) than expected (%s). Beware this can be a sign of attack",
			userSession.Username, bodyJSON.ClientID, consent.ClientID)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if bodyJSON.Consent {
		if bodyJSON.PreConfigure {
			if client.Consent.Mode == oidc.ClientConsentModePreConfigured {
				ctx.Logger.Warnf("Consent session with id '%s' for user '%s': consent pre-configuration was requested and was ignored because it is not permitted on this client", consent.ChallengeID, userSession.Username)
			} else {
				expiresAt := time.Now().Add(client.Consent.Duration)
				consent.ExpiresAt = &expiresAt

				ctx.Logger.Debugf("Consent session with id '%s' for user '%s': pre-configured and set to expire at %v", consent.ChallengeID, userSession.Username, consent.ExpiresAt)
			}
		}

		consent.Grant()
	}

	var externalRootURL string

	if externalRootURL, err = ctx.ExternalRootURL(); err != nil {
		ctx.Logger.Errorf("Could not determine the external URL during consent session processing with id '%s' for user '%s': %v", consent.ChallengeID, userSession.Username, err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, *consent, bodyJSON.Consent); err != nil {
		ctx.Logger.Errorf("Failed to save the consent session response to the database: %+v", err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		redirectURI *url.URL
		query       url.Values
	)

	if redirectURI, err = url.ParseRequestURI(externalRootURL); err != nil {
		ctx.Logger.Errorf("Failed to parse the consent redirect URL: %+v", err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !strings.HasSuffix(redirectURI.Path, "/") {
		redirectURI.Path += "/"
	}

	if query, err = url.ParseQuery(consent.Form); err != nil {
		ctx.Logger.Errorf("Failed to parse the consent form values: %+v", err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	query.Set(queryArgStrConsentID, consent.ChallengeID.String())

	redirectURI.Path = path.Join(redirectURI.Path, oidc.AuthorizationPath)
	redirectURI.RawQuery = query.Encode()

	response := oidc.ConsentPostResponseBody{RedirectURI: redirectURI.String()}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Error(fmt.Errorf("unable to set JSON bodyJSON in response"), "Operation failed")
	}
}

func oidcConsentGetSessionsAndClient(ctx *middlewares.AutheliaCtx, consentID uuid.UUID) (userSession session.UserSession, consent *model.OAuth2ConsentSession, client *oidc.Client, handled bool) {
	var (
		err error
	)

	userSession = ctx.GetSession()

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, consentID); err != nil {
		ctx.Logger.Errorf("Unable to load consent session with challenge id '%s': %v", consentID, err)
		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	if client, err = ctx.Providers.OpenIDConnect.Store.GetFullClient(consent.ClientID); err != nil {
		ctx.Logger.Errorf("Unable to find related client configuration with name '%s': %v", consent.ClientID, err)
		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	if err = verifyOIDCUserAuthorizedForConsent(ctx, client, userSession, consent, uuid.UUID{}); err != nil {
		ctx.Logger.Errorf("Could not authorize the user user '%s' for the consent session with challenge id '%s' on client with id '%s': %v", userSession.Username, consent.ChallengeID, client.GetID(), err)

		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	switch client.Consent.Mode {
	case oidc.ClientConsentModeImplicit:
		ctx.Logger.Errorf("Unable to perform OpenID Connect Consent for user '%s' and client id '%s': the client is using the implicit consent mode", userSession.Username, consent.ClientID)
		ctx.ReplyForbidden()

		return
	default:
		switch {
		case consent.Responded():
			ctx.Logger.Errorf("Unable to perform OpenID Connect Consent for user '%s' and client id '%s': the client is using the explicit consent mode and this consent session has already been responded to", userSession.Username, consent.ClientID)
			ctx.ReplyForbidden()

			return userSession, nil, nil, true
		case !consent.CanGrant():
			ctx.Logger.Errorf("Unable to perform OpenID Connect Consent for user '%s' and client id '%s': the specified consent session cannot be granted", userSession.Username, consent.ClientID)
			ctx.ReplyForbidden()

			return userSession, nil, nil, true
		}
	}

	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel) {
		ctx.Logger.Errorf("Unable to perform OpenID Connect Consent for user '%s' and client id '%s': the user is not sufficiently authenticated", userSession.Username, consent.ClientID)
		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	return userSession, consent, client, false
}
