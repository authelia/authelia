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
		flowID uuid.UUID
		err    error
	)

	if flowID, err = uuid.ParseBytes(ctx.RequestCtx.QueryArgs().PeekBytes(qryArgFlowID)); err != nil {
		ctx.Logger.Errorf("Unable to convert '%s' into a UUID: %+v", ctx.RequestCtx.QueryArgs().PeekBytes(qryArgFlowID), err)
		ctx.ReplyForbidden()

		return
	}

	var (
		consent *model.OAuth2ConsentSession
		form    url.Values
		client  oidc.Client
		handled bool
	)

	if _, consent, client, handled = handleOAuth2ConsentGetSessionsAndClient(ctx, flowID); handled {
		return
	}

	if form, err = handleGetFormFromFormSession(ctx, consent); err != nil {
		ctx.Logger.WithError(err).Errorf("Unable to get form from consent session with id '%s': %+v", consent.ChallengeID, err)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(client.GetConsentResponseBody(consent, form)); err != nil {
		ctx.Error(fmt.Errorf("unable to set JSON body: %w", err), "Operation failed")
	}
}

// OAuth2ConsentDeviceAuthorizationGET handles requests to provide consent for OpenID Connect.
func OAuth2ConsentDeviceAuthorizationGET(ctx *middlewares.AutheliaCtx) {
	var (
		err error
	)

	userCode := string(ctx.RequestCtx.QueryArgs().PeekBytes(qryArgUserCode))

	var (
		userSession       session.UserSession
		deviceCodeSession *model.OAuth2DeviceCodeSession
		form              url.Values
		client            oidc.Client
		handled           bool
	)

	if userSession, deviceCodeSession, client, handled = handleOAuth2ConsentDeviceAuthorizationGetSessionsAndClient(ctx, userCode); handled {
		return
	}

	if form, err = handleGetFormFromFormSession(ctx, deviceCodeSession); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"client_id": deviceCodeSession.ClientID, "session_id": deviceCodeSession.ID, "request_id": deviceCodeSession.RequestID, "username": userSession.Username}).
			Error("Failed to get form from device code session")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(client.GetConsentResponseBody(deviceCodeSession, form)); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"client_id": deviceCodeSession.ClientID, "session_id": deviceCodeSession.ID, "request_id": deviceCodeSession.RequestID, "username": userSession.Username}).
			Error("Error occurred trying to set JSON body in response")
		ctx.SetJSONError(messageOperationFailed)

		return
	}
}

// OAuth2ConsentPOST handles consent responses for OpenID Connect.
//
//nolint:gocyclo
func OAuth2ConsentPOST(ctx *middlewares.AutheliaCtx) {
	var (
		bodyJSON oidc.ConsentPostRequestBody
		err      error
	)

	if err = json.Unmarshal(ctx.Request.Body(), &bodyJSON); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred unmarshalling consent request body")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	// TODO: Handle password submission here.
	if bodyJSON.Password != nil {

	}

	switch {
	case bodyJSON.FlowID != nil:
		handleOAuth2ConsentPOSTWithFlowID(ctx, bodyJSON)
	case bodyJSON.UserCode != nil:
		handleOAuth2ConsentPOSTWithUserCode(ctx, bodyJSON)
	default:
		ctx.Logger.Error("Invalid request body")
		ctx.SetJSONError(messageOperationFailed)
	}
}

func handleOAuth2ConsentPOSTWithFlowID(ctx *middlewares.AutheliaCtx, bodyJSON oidc.ConsentPostRequestBody) {
	var (
		flowID uuid.UUID
		err    error
	)

	if flowID, err = uuid.Parse(*bodyJSON.FlowID); err != nil {
		ctx.Logger.WithError(err).WithFields(map[string]any{"flow_id": *bodyJSON.FlowID}).Error("Error occurred parsing flow ID as a UUID")
		ctx.ReplyForbidden()

		return
	}

	var (
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		client      oidc.Client
		handled     bool
	)

	if userSession, consent, client, handled = handleOAuth2ConsentGetSessionsAndClient(ctx, flowID); handled {
		return
	}

	if consent.ClientID != bodyJSON.ClientID {
		ctx.Logger.
			WithFields(map[string]any{"username": userSession.Username, "client_id": bodyJSON.ClientID, "consent_client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
			Error("The client id of the form and the client id of the consent session do not match")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if !client.IsAuthenticationLevelSufficient(level, authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		ctx.Logger.
			WithFields(map[string]any{"username": userSession.Username, "groups": userSession.Groups, "authentication_level": level.String(), "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String(), "authorization_policy": client.GetAuthorizationPolicy().Name}).
			Error("User is not sufficiently authenticated to provide consent given the client authorization policy")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		query url.Values
		form  url.Values
	)

	if query, err = consent.GetForm(); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
			Error("Error occurred trying to obtain the request form from the consent session")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !consent.Subject.Valid {
		if consent.Subject.UUID, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			ctx.Logger.
				WithError(err).
				WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
				Error("Error occurred trying to determine the subject for the consent session")

			ctx.SetJSONError(messageOperationFailed)

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
					requests *oidc.ClaimsRequests
				)

				if form, err = handleGetConsentForm(ctx, query); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
						Error("Error occurred trying to obtain the actual authorization parameters from the request form")
				} else if requests, err = oidc.NewClaimRequests(form); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
						Error("Error occurred parsing request form claims parameter")
				} else if requests == nil {
					config.RequestedClaims = sql.NullString{Valid: false}
					config.SignatureClaims = sql.NullString{Valid: false}
				} else if config.RequestedClaims.String, config.SignatureClaims.String, err = requests.Serialized(); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
						Error("Error occurred calculating claims signature for consent")
				} else {
					config.RequestedClaims.Valid = true
					config.SignatureClaims.Valid = true
				}

				var id int64

				if id, err = ctx.Providers.StorageProvider.SaveOAuth2ConsentPreConfiguration(ctx, config); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
						Error("Error occurred saving consent pre-configuration to the database")

					ctx.SetJSONError(messageOperationFailed)

					return
				}

				consent.PreConfiguration = sql.NullInt64{Int64: id, Valid: true}

				ctx.Logger.
					WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String(), "expiration": config.ExpiresAt.Time.Unix()}).
					Debug("Saved consent pre-configuration with expiration")
			} else {
				ctx.Logger.
					WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
					Warn("Ignored saving pre-configuration as it is not permitted by the client configuration")
			}
		}
	}

	consent.SetRespondedAt(ctx.Clock.Now(), 0)

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, bodyJSON.Consent); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
			Error("Error occurred saving the consent session response to the database")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		redirectURI *url.URL
	)

	redirectURI = ctx.RootURL()

	query.Set(queryArgConsentID, consent.ChallengeID.String())

	redirectURI.Path = path.Join(redirectURI.Path, oidc.EndpointPathAuthorization)
	redirectURI.RawQuery = query.Encode()

	response := oidc.ConsentPostResponseBody{RedirectURI: redirectURI.String()}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
			Error("Error occurred marshalling JSON response body")

		ctx.SetJSONError(messageOperationFailed)

		return
	}
}

func handleOAuth2ConsentPOSTWithUserCode(ctx *middlewares.AutheliaCtx, bodyJSON oidc.ConsentPostRequestBody) {
	var (
		flowID uuid.UUID
		err    error
	)

	if flowID, err = uuid.Parse(*bodyJSON.FlowID); err != nil {
		ctx.Logger.WithError(err).WithFields(map[string]any{"flow_id": *bodyJSON.FlowID}).Error("Error occurred parsing flow ID as a UUID")
		ctx.ReplyForbidden()

		return
	}

	var (
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		client      oidc.Client
		handled     bool
	)

	if userSession, consent, client, handled = handleOAuth2ConsentGetSessionsAndClient(ctx, flowID); handled {
		return
	}

	if consent.ClientID != bodyJSON.ClientID {
		ctx.Logger.
			WithFields(map[string]any{"username": userSession.Username, "client_id": bodyJSON.ClientID, "consent_client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
			Error("The client id of the form and the client id of the consent session do not match")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if !client.IsAuthenticationLevelSufficient(level, authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		ctx.Logger.
			WithFields(map[string]any{"username": userSession.Username, "groups": userSession.Groups, "authentication_level": level.String(), "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String(), "authorization_policy": client.GetAuthorizationPolicy().Name}).
			Error("User is not sufficiently authenticated to provide consent given the client authorization policy")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		query url.Values
		form  url.Values
	)

	if query, err = consent.GetForm(); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
			Error("Error occurred trying to obtain the request form from the consent session")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !consent.Subject.Valid {
		if consent.Subject.UUID, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			ctx.Logger.
				WithError(err).
				WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
				Error("Error occurred trying to determine the subject for the consent session")

			ctx.SetJSONError(messageOperationFailed)

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
					requests *oidc.ClaimsRequests
				)

				if form, err = handleGetConsentForm(ctx, query); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
						Error("Error occurred trying to obtain the actual authorization parameters from the request form")
				} else if requests, err = oidc.NewClaimRequests(form); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
						Error("Error occurred parsing request form claims parameter")
				} else if requests == nil {
					config.RequestedClaims = sql.NullString{Valid: false}
					config.SignatureClaims = sql.NullString{Valid: false}
				} else if config.RequestedClaims.String, config.SignatureClaims.String, err = requests.Serialized(); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
						Error("Error occurred calculating claims signature for consent")
				} else {
					config.RequestedClaims.Valid = true
					config.SignatureClaims.Valid = true
				}

				var id int64

				if id, err = ctx.Providers.StorageProvider.SaveOAuth2ConsentPreConfiguration(ctx, config); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
						Error("Error occurred saving consent pre-configuration to the database")

					ctx.SetJSONError(messageOperationFailed)

					return
				}

				consent.PreConfiguration = sql.NullInt64{Int64: id, Valid: true}

				ctx.Logger.
					WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String(), "expiration": config.ExpiresAt.Time.Unix()}).
					Debug("Saved consent pre-configuration with expiration")
			} else {
				ctx.Logger.
					WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
					Warn("Ignored saving pre-configuration as it is not permitted by the client configuration")
			}
		}
	}

	consent.SetRespondedAt(ctx.Clock.Now(), 0)

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, bodyJSON.Consent); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
			Error("Error occurred saving the consent session response to the database")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		redirectURI *url.URL
	)

	redirectURI = ctx.RootURL()

	query.Set(queryArgConsentID, consent.ChallengeID.String())

	redirectURI.Path = path.Join(redirectURI.Path, oidc.EndpointPathAuthorization)
	redirectURI.RawQuery = query.Encode()

	response := oidc.ConsentPostResponseBody{RedirectURI: redirectURI.String()}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"username": userSession.Username, "client_id": consent.ClientID, "consent_id": consent.ID, "flow_id": flowID.String()}).
			Error("Error occurred marshalling JSON response body")

		ctx.SetJSONError(messageOperationFailed)

		return
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

func handleOAuth2ConsentDeviceAuthorizationGetSessionsAndClient(ctx *middlewares.AutheliaCtx, userCode string) (userSession session.UserSession, deviceCodeSession *model.OAuth2DeviceCodeSession, client oidc.Client, handled bool) {
	var (
		signature string
		err       error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.
			WithError(err).
			Error("Error occurred loading user session during the Consent Flow stage of the Device Authorization Flow")
		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, nil, true
	}

	if signature, err = ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(ctx, userCode); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"username": userSession.Username}).
			Error("Error occurred deriving device code session signature using user code during the Consent Flow stage of the Device Authorization Flow")
		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, nil, true
	}

	if deviceCodeSession, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, signature); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"username": userSession.Username}).
			Error("Error occurred loading device code session using user code signature during the Consent Flow stage of the Device Authorization Flow")
		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, nil, true
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, deviceCodeSession.ClientID); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{"client_id": deviceCodeSession.ClientID, "session_id": deviceCodeSession.ID, "request_id": deviceCodeSession.RequestID, "username": userSession.Username}).
			Error("Error occurred loading registered client using client id during the Consent Flow stage of the Device Authorization Flow")
		ctx.SetJSONError(messageOperationFailed)

		return userSession, deviceCodeSession, nil, true
	}

	if !client.IsAuthenticationLevelSufficient(userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA), authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		ctx.Logger.
			WithFields(map[string]any{"client_id": deviceCodeSession.ClientID, "session_id": deviceCodeSession.ID, "request_id": deviceCodeSession.RequestID, "username": userSession.Username, "groups": userSession.Groups, "ip": ctx.RemoteIP()}).
			Error("Device Authorization Flow failed to retrieve Consent Flow data as the user is not sufficiently authenticated")
		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, nil, true
	}

	return userSession, deviceCodeSession, client, false
}

func handleGetFormFromFormSession(ctx *middlewares.AutheliaCtx, session oidc.FormSession) (form url.Values, err error) {
	if form, err = session.GetForm(); err != nil {
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
