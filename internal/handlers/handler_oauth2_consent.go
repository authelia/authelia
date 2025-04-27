package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"path"

	"authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// OAuth2ConsentGET handles requests to provide consent for OpenID Connect.
func OAuth2ConsentGET(ctx *middlewares.AutheliaCtx) {
	var (
		consentID uuid.UUID
		err       error
	)

	if consentID, err = uuid.ParseBytes(ctx.RequestCtx.QueryArgs().PeekBytes(qryArgID)); err != nil {
		ctx.Logger.Errorf("Unable to convert '%s' into a UUID: %+v", ctx.RequestCtx.QueryArgs().PeekBytes(qryArgID), err)
		ctx.ReplyForbidden()

		return
	}

	var (
		consent *model.OAuth2ConsentSession
		form    url.Values
		client  oidc.Client
		handled bool
	)

	if _, consent, client, handled = handleOAuth2ConsentGetSessionsAndClient(ctx, consentID); handled {
		return
	}

	if form, err = handleGetConsentFormFromConsentSession(ctx, consent); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to get form from consent session with id '%s': %+v", consent.ChallengeID, err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(client.GetConsentResponseBody(consent, form)); err != nil {
		ctx.Error(fmt.Errorf("unable to set JSON body: %w", err), "Operation failed")
	}
}

// OAuth2ConsentPOST handles consent responses for OpenID Connect.
//
//nolint:gocyclo
func OAuth2ConsentPOST(ctx *middlewares.AutheliaCtx) {
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
		ctx.Logger.Errorf("Unable to convert '%s' into a UUID: %+v", bodyJSON.ConsentID, err)
		ctx.ReplyForbidden()

		return
	}

	var (
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		client      oidc.Client
		handled     bool
	)

	if userSession, consent, client, handled = handleOAuth2ConsentGetSessionsAndClient(ctx, consentID); handled {
		return
	}

	if consent.ClientID != bodyJSON.ClientID {
		ctx.Logger.Errorf("User '%s' consented to scopes of another client (%s) than expected (%s). Beware this can be a sign of attack",
			userSession.Username, bodyJSON.ClientID, consent.ClientID)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA), authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		ctx.Logger.Errorf("User '%s' can't consent to authorization request for client with id '%s' as they are not sufficiently authenticated",
			userSession.Username, consent.ClientID)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var form url.Values

	if form, err = consent.GetForm(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred getting request form from consent session for user '%s' and client with id '%s'", userSession.Username, consent.ClientID)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !consent.Subject.Valid {
		if consent.Subject.UUID, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			ctx.Error(fmt.Errorf("unable to determine consent subject for client with id '%s' with consent challenge id '%s': %w", client.GetID(), consent.ChallengeID, err), messageAuthenticationFailed)

			return
		}

		consent.Subject.Valid = true
	}

	if bodyJSON.Consent {
		oidc.ConsentGrant(consent, true, bodyJSON.Claims)

		if bodyJSON.PreConfigure {
			if client.GetConsentPolicy().Mode == oidc.ClientConsentModePreConfigured {
				config := model.OAuth2ConsentPreConfig{
					ClientID:      consent.ClientID,
					Subject:       consent.Subject.UUID,
					CreatedAt:     ctx.Clock.Now(),
					ExpiresAt:     sql.NullTime{Time: ctx.Clock.Now().Add(client.GetConsentPolicy().Duration), Valid: true},
					Scopes:        consent.GrantedScopes,
					Audience:      consent.GrantedAudience,
					GrantedClaims: bodyJSON.Claims,
				}

				var (
					requests   *oidc.ClaimsRequests
					actualForm url.Values
				)

				if actualForm, err = handleGetConsentForm(ctx, form); err != nil {
					ctx.Logger.WithError(err).Debug("Error occurred resolving the actual form from the consent form")
				} else if requests, err = oidc.NewClaimRequests(actualForm); err != nil {
					ctx.Logger.WithError(err).Debug("Error occurred parsing claims parameter from request form for claims signature")
				} else if requests == nil {
					config.RequestedClaims = sql.NullString{Valid: false}
					config.SignatureClaims = sql.NullString{Valid: false}
				} else if config.RequestedClaims.String, config.SignatureClaims.String, err = requests.Serialized(); err != nil {
					ctx.Logger.WithError(err).Debug("Error occurred calculating claims signature")
				} else {
					config.RequestedClaims.Valid = true
					config.SignatureClaims.Valid = true
				}

				var id int64

				if id, err = ctx.Providers.StorageProvider.SaveOAuth2ConsentPreConfiguration(ctx, config); err != nil {
					ctx.Logger.Errorf("Failed to save the consent pre-configuration to the database: %+v", err)
					ctx.SetJSONError(messageOperationFailed)

					return
				}

				consent.PreConfiguration = sql.NullInt64{Int64: id, Valid: true}

				ctx.Logger.Debugf("Consent session with id '%s' for user '%s': pre-configured and set to expire at %v", consent.ChallengeID, userSession.Username, config.ExpiresAt.Time)
			} else {
				ctx.Logger.Warnf("Consent session with id '%s' for user '%s': consent pre-configuration was requested and was ignored because it is not permitted on this client", consent.ChallengeID, userSession.Username)
			}
		}
	}

	consent.SetRespondedAt(ctx.Clock.Now(), 0)

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, bodyJSON.Consent); err != nil {
		ctx.Logger.Errorf("Failed to save the consent session response to the database: %+v", err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		redirectURI *url.URL
		query       url.Values
	)

	redirectURI = ctx.RootURL()

	if query, err = url.ParseQuery(consent.Form); err != nil {
		ctx.Logger.Errorf("Failed to parse the consent form values: %+v", err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	query.Set(queryArgConsentID, consent.ChallengeID.String())

	redirectURI.Path = path.Join(redirectURI.Path, oidc.EndpointPathAuthorization)
	redirectURI.RawQuery = query.Encode()

	response := oidc.ConsentPostResponseBody{RedirectURI: redirectURI.String()}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Error(fmt.Errorf("unable to set JSON bodyJSON in response"), "Operation failed")
	}
}

func handleOAuth2ConsentGetSessionsAndClient(ctx *middlewares.AutheliaCtx, consentID uuid.UUID) (userSession session.UserSession, consent *model.OAuth2ConsentSession, client oidc.Client, handled bool) {
	var (
		err error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.Errorf("Unable to load user session for challenge id '%s': %v", consentID, err)
		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, consentID); err != nil {
		ctx.Logger.Errorf("Unable to load consent session with challenge id '%s': %v", consentID, err)
		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID); err != nil {
		ctx.Logger.Errorf("Unable to find related client configuration with name '%s': %v", consent.ClientID, err)
		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	switch client.GetConsentPolicy().Mode {
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
		case !consent.CanGrant(ctx.Clock.Now()):
			ctx.Logger.Errorf("Unable to perform OpenID Connect Consent for user '%s' and client id '%s': the specified consent session cannot be granted", userSession.Username, consent.ClientID)
			ctx.ReplyForbidden()

			return userSession, nil, nil, true
		}
	}

	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA), authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		ctx.Logger.Errorf("Unable to perform OpenID Connect Consent for user '%s' and client id '%s': the user is not sufficiently authenticated", userSession.Username, consent.ClientID)
		ctx.ReplyForbidden()

		return userSession, nil, nil, true
	}

	return userSession, consent, client, false
}

func handleGetConsentFormFromConsentSession(ctx *middlewares.AutheliaCtx, consent *model.OAuth2ConsentSession) (form url.Values, err error) {
	if form, err = consent.GetForm(); err != nil {
		return nil, err
	}

	return handleGetConsentForm(ctx, form)
}

func handleGetConsentForm(ctx *middlewares.AutheliaCtx, original url.Values) (form url.Values, err error) {
	var requester oauth2.AuthorizeRequester

	if requester, err = handleGetConsentFormFromPushedAuthorizeRequestRedirectURI(ctx, original); err != nil {
		return nil, err
	} else if requester != nil {
		return requester.GetRequestForm(), nil
	}

	return form, err
}

func handleGetConsentFormFromPushedAuthorizeRequestRedirectURI(ctx *middlewares.AutheliaCtx, form url.Values) (requester oauth2.AuthorizeRequester, err error) {
	if oidc.IsPushedAuthorizedRequestForm(form, ctx.Providers.OpenIDConnect.GetPushedAuthorizeRequestURIPrefix(ctx)) {
		if requester, err = ctx.Providers.OpenIDConnect.GetPARSession(ctx, form.Get(oidc.FormParameterRequestURI)); err != nil {
			return nil, err
		}

		return requester, nil
	}

	return nil, nil
}
