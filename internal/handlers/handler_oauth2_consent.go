package handlers

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"path"
	"time"

	"authelia.com/provider/oauth2"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/logging"
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

	raw := ctx.RequestCtx.QueryArgs().PeekBytes(qryArgFlowID)

	if flowID, err = uuid.ParseBytes(raw); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: string(raw)}).
			Error("Error occurred parsing flow ID")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		form        url.Values
		client      oidc.Client
		handled     bool
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String()}).
			Error("Error occurred fetching user session")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if _, consent, client, handled = handleOAuth2ConsentGetSessionsAndClient(ctx, flowID); handled {
		return
	}

	if consent.ExpiresAt.Before(ctx.Clock.Now()) {
		ctx.Logger.
			WithFields(map[string]any{logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldExpiration: consent.ExpiresAt.Unix()}).
			Error("Failed providing consent flow information as the consent session is expired")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if form, err = handleGetFormFromFormSession(ctx, consent); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldExpiration: consent.ExpiresAt.Unix()}).
			Error("Error occurred getting form from consent session")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(client.GetConsentResponseBody(consent, form, userSession.LastAuthenticatedTime(), false)); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldExpiration: consent.ExpiresAt.Unix()}).
			Error("Error occurred trying to set JSON body in response")

		ctx.SetJSONError(messageOperationFailed)

		return
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
			WithFields(map[string]any{logging.FieldClientID: deviceCodeSession.ClientID, logging.FieldSessionID: deviceCodeSession.ID, logging.FieldRequestID: deviceCodeSession.RequestID, logging.FieldUsername: userSession.Username}).
			Error("Error occurred getting form from device code session")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(client.GetConsentResponseBody(deviceCodeSession, form, userSession.LastAuthenticatedTime(), true)); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldClientID: deviceCodeSession.ClientID, logging.FieldSessionID: deviceCodeSession.ID, logging.FieldRequestID: deviceCodeSession.RequestID, logging.FieldUsername: userSession.Username}).
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
		ctx.Logger.
			WithError(err).
			Error("Error occurred unmarshalling consent request body")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if bodyJSON.SubFlow == nil {
		handleOAuth2ConsentFlowIDPOST(ctx, bodyJSON)

		return
	}

	switch *bodyJSON.SubFlow {
	case flowOpenIDConnectSubFlowNameDeviceAuthorization:
		handleOAuth2ConsentDeviceAuthorizationPOST(ctx, bodyJSON)
	default:
		handleOAuth2ConsentFlowIDPOST(ctx, bodyJSON)
	}
}

func handleOAuth2ConsentFlowIDPOST(ctx *middlewares.AutheliaCtx, bodyJSON oidc.ConsentPostRequestBody) {
	var (
		flowID uuid.UUID
		err    error
	)

	if bodyJSON.FlowID == nil {
		ctx.Logger.
			Error("Request is missing the required field 'flow_id' from the JSON body")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if flowID, err = uuid.Parse(*bodyJSON.FlowID); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: *bodyJSON.FlowID}).
			Error("Error occurred parsing flow ID as a UUID")

		ctx.SetJSONError(messageOperationFailed)

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
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, "body_client_id": bodyJSON.ClientID, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("The client id of the form and the client id of the consent session do not match")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if !client.IsAuthenticationLevelSufficient(level, authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldGroups: userSession.Groups, logging.FieldAuthenticationLevel: level.String(), logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldAuthorizationPolicy: client.GetAuthorizationPolicy().Name}).
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
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred trying to obtain the request form from the consent session")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !consent.Subject.Valid {
		if consent.Subject.UUID, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			ctx.Logger.
				WithError(err).
				WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
				Error("Error occurred trying to determine the subject for the consent session")

			ctx.SetJSONError(messageOperationFailed)

			return
		}

		consent.Subject.Valid = true
	}

	if form, err = handleGetConsentForm(ctx, query); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred trying to obtain the actual authorization parameters from the request form")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if oidc.RequestFormRequiresLogin(form, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("The authorization request requires the user performs a login even prior to providing consent")

		ctx.SetJSONError(messageOperationFailed)

		return
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

				if requests, err = oidc.NewClaimRequests(form); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
						Error("Error occurred parsing request form claims parameter")
				} else if requests == nil {
					config.RequestedClaims = sql.NullString{Valid: false}
					config.SignatureClaims = sql.NullString{Valid: false}
				} else if config.RequestedClaims.String, config.SignatureClaims.String, err = requests.Serialized(); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
						Error("Error occurred calculating claims signature for consent")
				} else {
					config.RequestedClaims.Valid = true
					config.SignatureClaims.Valid = true
				}

				var id int64

				if id, err = ctx.Providers.StorageProvider.SaveOAuth2ConsentPreConfiguration(ctx, config); err != nil {
					ctx.Logger.
						WithError(err).
						WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
						Error("Error occurred saving consent pre-configuration to the database")

					ctx.SetJSONError(messageOperationFailed)

					return
				}

				consent.PreConfiguration = sql.NullInt64{Int64: id, Valid: true}

				ctx.Logger.
					WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldExpiration: config.ExpiresAt.Time.Unix()}).
					Debug("Saved consent pre-configuration with expiration")
			} else {
				ctx.Logger.
					WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
					Warn("Ignored saving pre-configuration as it is not permitted by the client configuration")
			}
		}
	}

	consent.SetRespondedAt(ctx.Clock.Now(), 0)

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, bodyJSON.Consent); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
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
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred marshalling JSON response body")

		ctx.SetJSONError(messageOperationFailed)

		return
	}
}

func handleOAuth2ConsentDeviceAuthorizationPOST(ctx *middlewares.AutheliaCtx, bodyJSON oidc.ConsentPostRequestBody) {
	var (
		err error
	)

	if bodyJSON.UserCode == nil {
		ctx.Logger.
			Error("Request is missing the required field 'user_code' from the JSON body")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		signature   string
		userSession session.UserSession
		device      *model.OAuth2DeviceCodeSession
		consent     *model.OAuth2ConsentSession
		client      oidc.Client
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.
			WithError(err).
			Error("Error occurred fetching user session during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.
			Error("Error occurred fetching user session during the Consent Flow stage of the Device Authorization Flow as the user is anonymous")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	code := *bodyJSON.UserCode

	if signature, err = ctx.Providers.OpenIDConnect.Config.Strategy.Core.RFC8628UserCodeSignature(ctx, code); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldUsername: userSession.Username, logging.FieldClientID: bodyJSON.ClientID}).
			Error("Error occurred determining the signature of the user code session during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if device, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, signature); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldUsername: userSession.Username, logging.FieldClientID: bodyJSON.ClientID, "signature": signature}).
			Error("Error occurred using the signature of the user code session to retrieve the device code session during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)
	}

	if device.ClientID != bodyJSON.ClientID {
		ctx.Logger.
			WithFields(map[string]any{logging.FieldUsername: userSession.Username, "body_client_id": bodyJSON.ClientID, logging.FieldClientID: device.ClientID, logging.FieldSessionID: device.ID}).
			Error("Error occurred matching the user code to the device code session during the Consent Flow stage of the Device Authorization Flow as the client id of the form and the client id of the consent session do not match")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, device.ClientID); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldUsername: userSession.Username, logging.FieldClientID: device.ClientID, logging.FieldSessionID: device.ID}).
			Error("Error occurred fetching client configuration during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if !client.IsAuthenticationLevelSufficient(level, authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		ctx.Logger.
			WithFields(map[string]any{logging.FieldUsername: userSession.Username, logging.FieldGroups: userSession.Groups, logging.FieldAuthenticationLevel: level.String(), logging.FieldClientID: device.ClientID, logging.FieldSessionID: device.ID, logging.FieldAuthorizationPolicy: client.GetAuthorizationPolicy().Name}).
			Error("User is not sufficiently authenticated to provide consent given the client authorization policy during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		subject uuid.UUID
		r       *oauth2.DeviceAuthorizeRequest
	)

	if subject, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldUsername: userSession.Username, logging.FieldClientID: device.ClientID, logging.FieldSessionID: device.ID}).
			Error("Error occurred trying to determine the subject during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if r, err = device.ToRequest(ctx, oidc.NewSession(), ctx.Providers.OpenIDConnect.Store); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldUsername: userSession.Username, logging.FieldClientID: device.ClientID, logging.FieldSessionID: device.ID}).
			Error("Error occurred trying to restore the requester during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if consent, err = model.NewOAuth2ConsentSession(ctx.Clock.Now().Add(10*time.Second), subject, r); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldUsername: userSession.Username, logging.FieldClientID: device.ClientID, logging.FieldSessionID: device.ID}).
			Error("Error occurred trying to create the consent session during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	device.ChallengeID = uuid.NullUUID{UUID: consent.ChallengeID, Valid: true}

	device.RequestedAudience = device.GrantedAudience
	device.RequestedScopes = device.GrantedScopes

	if bodyJSON.Consent {
		oidc.ConsentGrant(consent, true, bodyJSON.Claims)
	}

	consent.SetRespondedAt(ctx.Clock.Now(), 0)

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(ctx, consent); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred saving the consent session to the database during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, bodyJSON.Consent); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred saving the consent session response to the database during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(ctx, device); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred saving the device code session challenge id to the database during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	response := oidc.ConsentPostResponseBody{FlowID: consent.ChallengeID.String()}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred marshalling JSON response body")

		ctx.SetJSONError(messageOperationFailed)

		return
	}
}

func handleOAuth2ConsentGetSessionsAndClient(ctx *middlewares.AutheliaCtx, flowID uuid.UUID) (userSession session.UserSession, consent *model.OAuth2ConsentSession, client oidc.Client, handled bool) {
	var (
		err error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String()}).
			Error("Error occurred fetching user session during the Consent Flow stage of the Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, nil, true
	}

	if consent, err = ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(ctx, flowID); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldUsername: userSession.Username}).
			Error("Error occurred fetching consent session during the Consent Flow stage of the Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, nil, true
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred fetching client configuration during the Consent Flow stage of the Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, nil, true
	}

	switch client.GetConsentPolicy().Mode {
	case oidc.ClientConsentModeImplicit:
		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred performing consent during the Consent FLow stage of the Authorization Flow as the client is configured to use implicit consent")

		ctx.SetJSONError(messageOperationFailed)

		return
	default:
		switch {
		case consent.Responded():
			ctx.Logger.
				WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldSubject: consent.Subject.UUID.String(), logging.FieldResponded: consent.RespondedAt.Time.Unix()}).
				Error("Error occurred performing consent during the Consent FLow stage of the Authorization Flow as the consent session has already been responded to")

			ctx.SetJSONError(messageOperationFailed)

			return userSession, nil, nil, true
		case !consent.CanGrant(ctx.Clock.Now()):
			ctx.Logger.
				WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldGranted: consent.Granted, logging.FieldExpiration: consent.ExpiresAt.Unix()}).
				Error("Error occurred performing consent during the Consent FLow stage of the Authorization Flow as the consent session has already been granted or is expired")

			ctx.SetJSONError(messageOperationFailed)

			return userSession, nil, nil, true
		}
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if !client.IsAuthenticationLevelSufficient(level, authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		ctx.Logger.
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldGroups: userSession.Groups, logging.FieldAuthenticationLevel: level.String(), logging.FieldAuthorizationPolicy: client.GetAuthorizationPolicy().Name}).
			Error("Error occurred performing consent during the Consent FLow stage of the Authorization Flow as the user is not sufficiently authenticated")

		ctx.SetJSONError(messageOperationFailed)

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
			Error("Error occurred fetching user session during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, nil, true
	}

	if signature, err = ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(ctx, userCode); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldUsername: userSession.Username}).
			Error("Error occurred deriving device code session signature using user code during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, nil, true
	}

	if deviceCodeSession, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, signature); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldUsername: userSession.Username}).
			Error("Error occurred loading device code session using user code signature during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, nil, nil, true
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, deviceCodeSession.ClientID); err != nil {
		ctx.Logger.
			WithError(err).
			WithFields(map[string]any{logging.FieldClientID: deviceCodeSession.ClientID, logging.FieldSessionID: deviceCodeSession.ID, logging.FieldRequestID: deviceCodeSession.RequestID, logging.FieldUsername: userSession.Username}).
			Error("Error occurred loading registered client using client id during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, deviceCodeSession, nil, true
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)
	if !client.IsAuthenticationLevelSufficient(level, authorization.Subject{Username: userSession.Username, Groups: userSession.Groups, IP: ctx.RemoteIP()}) {
		ctx.Logger.
			WithFields(map[string]any{logging.FieldClientID: deviceCodeSession.ClientID, logging.FieldSessionID: deviceCodeSession.ID, logging.FieldRequestID: deviceCodeSession.RequestID, logging.FieldUsername: userSession.Username, logging.FieldGroups: userSession.Groups, logging.FieldAuthenticationLevel: level.String(), logging.FieldAuthorizationPolicy: client.GetAuthorizationPolicy().Name}).
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
