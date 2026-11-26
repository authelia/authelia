// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"

	"authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// OAuth2ConsentGET handles requests to provide consent for OpenID Connect.
func OAuth2ConsentGET(ctx *middlewares.AutheliaCtx) {
	var raw []byte

	if raw = ctx.RequestCtx.QueryArgs().PeekBytes(qryArgFlowID); len(raw) != 0 {
		handleOAuth2ConsentFlowIDGET(ctx, raw, string(ctx.RequestCtx.QueryArgs().PeekBytes(qryArgSubflow)) == flowOpenIDConnectSubFlowNameDeviceAuthorization)
	} else {
		ctx.GetLogger().Error("Error determining the type of consent request to handle")

		ctx.SetJSONError(messageOperationFailed)
	}
}

func handleOAuth2ConsentFlowIDGET(ctx *middlewares.AutheliaCtx, raw []byte, deviceAuthorization bool) {
	var (
		flowID uuid.UUID
		err    error
	)
	if flowID, err = uuid.ParseBytes(raw); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: string(raw)}).
			Error("Error occurred parsing flow ID")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		userSession session.UserSession

		consent *model.OAuth2ConsentSession
		form    url.Values
		client  oidc.Client
		handled bool
	)

	if userSession, _, consent, client, handled = handleOAuth2ConsentGetSessionsAndClient(ctx, flowID); handled {
		return
	}

	if consent.ExpiresAt.Before(ctx.GetClock().Now()) {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldExpiration: consent.ExpiresAt.Unix()}).
			Error("Failed providing consent flow information as the consent session is expired")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if deviceAuthorization {
		if _, err = handleOAuth2ConsentDeviceAuthorizationLoad(ctx, consent); err != nil {
			ctx.GetLogger().
				WithError(err).
				WithFields(map[string]any{logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username}).
				Error("Failed providing consent flow information as the consent session has not claimed a device code session which is pending a decision")

			ctx.SetJSONError(messageOperationFailed)

			return
		}
	}

	if form, err = handleGetFormFromSession(ctx, consent); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldExpiration: consent.ExpiresAt.Unix()}).
			Error("Error occurred getting form from consent session")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(client.GetConsentResponseBody(consent, form, userSession.LastAuthenticatedTime(), deviceAuthorization)); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldExpiration: consent.ExpiresAt.Unix()}).
			Error("Error occurred trying to set JSON body in response")

		ctx.SetJSONError(messageOperationFailed)

		return
	}
}

// OAuth2ConsentPOST handles consent responses for OpenID Connect.
func OAuth2ConsentPOST(ctx *middlewares.AutheliaCtx) {
	var (
		bodyJSON oidc.ConsentPostRequestBody
		err      error
	)
	if err = json.Unmarshal(ctx.Request.Body(), &bodyJSON); err != nil {
		ctx.GetLogger().
			WithError(err).
			Error("Error occurred unmarshaling consent request body")

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

//nolint:gocyclo
func handleOAuth2ConsentFlowIDPOST(ctx *middlewares.AutheliaCtx, bodyJSON oidc.ConsentPostRequestBody) {
	var (
		flowID uuid.UUID
		err    error
	)

	if bodyJSON.FlowID == nil {
		ctx.GetLogger().
			Error("Request is missing the required field 'flow_id' from the JSON body")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if flowID, err = uuid.Parse(*bodyJSON.FlowID); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: *bodyJSON.FlowID}).
			Error("Error occurred parsing flow ID as a UUID")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		userSession session.UserSession
		details     authentication.UserDetails
		consent     *model.OAuth2ConsentSession
		client      oidc.Client
		handled     bool
	)

	if userSession, details, consent, client, handled = handleOAuth2ConsentGetSessionsAndClient(ctx, flowID); handled {
		return
	}

	if consent.ClientID != bodyJSON.ClientID {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, "body_client_id": bodyJSON.ClientID, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("The client id of the form and the client id of the consent session do not match")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	switch _, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByChallengeID(ctx, consent.ChallengeID); {
	case err == nil:
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred performing consent during the Consent Flow stage of the Authorization Flow as the consent session belongs to the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	case !errors.Is(err, sql.ErrNoRows):
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred performing consent during the Consent Flow stage of the Authorization Flow as an error occurred determining if the consent session belongs to the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if !client.IsAuthenticationLevelSufficient(level, authorization.Subject{Username: userSession.Username, Groups: details.Groups, IP: ctx.RemoteIP()}) {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldGroups: details.Groups, logging.FieldAuthenticationLevel: level.String(), logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldAuthorizationPolicy: client.GetAuthorizationPolicy().Name}).
			Error("User is not sufficiently authenticated to provide consent given the client authorization policy")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		query url.Values
		form  url.Values
	)

	if query, err = consent.GetForm(); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred trying to obtain the request form from the consent session")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !consent.Subject.Valid {
		if consent.Subject.UUID, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			ctx.GetLogger().
				WithError(err).
				WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
				Error("Error occurred trying to determine the subject for the consent session")

			ctx.SetJSONError(messageOperationFailed)

			return
		}

		consent.Subject.Valid = true
	}

	if form, err = handleGetConsentForm(ctx, query); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred trying to obtain the actual authorization parameters from the request form")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if oidc.RequestFormRequiresLogin(form, consent.RequestedAt, userSession.LastAuthenticatedTime()) {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("The authorization request requires the user performs a login even prior to providing consent")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if bodyJSON.Consent {
		oidc.ConsentGrant(consent, true, bodyJSON.Claims)

		if bodyJSON.PreConfigure {
			switch {
			case oidc.FormRequiresExplicitConsent(form), !oidc.FormIsClientIdentityAssured(client, form):
				ctx.GetLogger().
					WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
					Warn("Ignored saving pre-configuration as it is not permitted due to constraints within the authorization request form")
			case client.GetConsentPolicy().Mode != oidc.ClientConsentModePreConfigured:
				ctx.GetLogger().
					WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
					Warn("Ignored saving pre-configuration as it is not permitted by the client configuration")
			default:
				handleSavePreConfiguredConsent(ctx, userSession, consent, client, form, bodyJSON.Claims)
			}
		}
	}

	consent.SetRespondedAt(ctx.GetClock().Now(), 0)

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(ctx, consent, bodyJSON.Consent); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred saving the consent session response to the database")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var redirectURI *url.URL

	if redirectURI, err = ctx.IssuerURL(); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred trying to determine the issuer URL")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	redirectURI = handleOAuth2AuthorizationConsentReferenceURL(redirectURI, consent)

	response := oidc.ConsentPostResponseBody{RedirectURI: redirectURI.String()}

	if err = ctx.SetJSONBody(response); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred marshaling JSON response body")

		ctx.SetJSONError(messageOperationFailed)

		return
	}
}

func handleSavePreConfiguredConsent(ctx *middlewares.AutheliaCtx, userSession session.UserSession, consent *model.OAuth2ConsentSession, client oidc.Client, form url.Values, claims []string) {
	var err error

	config := model.OAuth2ConsentPreConfig{
		ClientID:  consent.ClientID,
		Subject:   consent.Subject.UUID,
		CreatedAt: ctx.GetClock().Now(),
		ExpiresAt: sql.NullTime{Time: ctx.GetClock().Now().Add(client.GetConsentPolicy().Duration), Valid: true},
		Scopes:    consent.GrantedScopes,
		Audience:  consent.GrantedAudience,
		Resource:  consent.GrantedResource,
	}

	var (
		requests *oidc.ClaimsRequests
	)

	if requests, err = oidc.NewClaimRequests(form); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred parsing request form claims parameter")
	} else if requests == nil {
		config.RequestedClaims = sql.NullString{Valid: false}
		config.SignatureClaims = sql.NullString{Valid: false}
	} else if config.RequestedClaims.String, config.SignatureClaims.String, err = requests.Serialized(); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred calculating claims signature for consent")
	} else {
		config.RequestedClaims.Valid = true
		config.SignatureClaims.Valid = true
		config.GrantedClaims = claims
	}

	var id int64

	if id, err = ctx.Providers.StorageProvider.SaveOAuth2ConsentPreConfiguration(ctx, config); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred saving consent pre-configuration to the database")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	consent.PreConfiguration = sql.NullInt64{Int64: id, Valid: true}

	ctx.GetLogger().
		WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldExpiration: config.ExpiresAt.Time.Unix(), logging.FieldScope: consent.GrantedScopes, "claims": consent.GrantedClaims}).
		Debug("Saved consent pre-configuration with expiration")
}

func handleOAuth2ConsentDeviceAuthorizationPOST(ctx *middlewares.AutheliaCtx, bodyJSON oidc.ConsentPostRequestBody) {
	var (
		flowID uuid.UUID
		err    error
	)

	if bodyJSON.FlowID == nil {
		ctx.GetLogger().
			Error("Request is missing the required field 'flow_id' from the JSON body during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if flowID, err = uuid.Parse(*bodyJSON.FlowID); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: *bodyJSON.FlowID}).
			Error("Error occurred parsing flow ID as a UUID during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		userSession session.UserSession
		consent     *model.OAuth2ConsentSession
		device      *model.OAuth2DeviceCodeSession
		handled     bool
	)

	if userSession, _, consent, _, handled = handleOAuth2ConsentGetSessionsAndClient(ctx, flowID); handled {
		return
	}

	log := ctx.GetLogger().WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID})

	if consent.ClientID != bodyJSON.ClientID {
		log.
			WithFields(map[string]any{"body_client_id": bodyJSON.ClientID}).
			Error("Error occurred during the Consent Flow stage of the Device Authorization Flow as the client id of the form and the client id of the consent session do not match")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if device, err = handleOAuth2ConsentDeviceAuthorizationLoad(ctx, consent); err != nil {
		log.
			WithError(err).
			Error("Error occurred during the Consent Flow stage of the Device Authorization Flow as the consent session has not claimed a device code session which is pending a decision")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !consent.Subject.Valid {
		log.
			Error("Error occurred during the Consent Flow stage of the Device Authorization Flow as the consent session does not have a subject")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !handleOAuth2ConsentDeviceAuthorizationSave(ctx, userSession.Username, bodyJSON, consent, device) {
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(oidc.ConsentPostResponseBody{FlowID: consent.ChallengeID.String()}); err != nil {
		log.
			WithError(err).
			Error("Error occurred marshaling JSON response body")

		ctx.SetJSONError(messageOperationFailed)
	}
}

func handleOAuth2ConsentDeviceAuthorizationLoad(ctx *middlewares.AutheliaCtx, consent *model.OAuth2ConsentSession) (device *model.OAuth2DeviceCodeSession, err error) {
	if device, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByChallengeID(ctx, consent.ChallengeID); err != nil {
		return nil, fmt.Errorf("error occurred loading the device code session: %w", err)
	}

	if err = validateOAuth2DeviceCodeSessionClaimed(ctx, device); err != nil {
		return nil, err
	}

	return device, nil
}

func handleOAuth2ConsentGetSessionsAndClient(ctx *middlewares.AutheliaCtx, flowID uuid.UUID) (userSession session.UserSession, details authentication.UserDetails, consent *model.OAuth2ConsentSession, client oidc.Client, handled bool) {
	var err error
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String()}).
			Error("Error occurred fetching user session during the Consent Flow stage of the Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, details, nil, nil, true
	}

	if consent, err = handleOAuth2ConsentSessionLoadBound(ctx, flowID); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldUsername: userSession.Username}).
			Error("Error occurred fetching consent session during the Consent Flow stage of the Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, details, nil, nil, true
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, consent.ClientID); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
			Error("Error occurred fetching client configuration during the Consent Flow stage of the Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, details, nil, nil, true
	}

	switch {
	case consent.Responded():
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldSubject: consent.Subject.UUID.String(), logging.FieldResponded: consent.RespondedAt.Time.Unix()}).
			Error("Error occurred performing consent during the Consent Flow stage of the Authorization Flow as the consent session has already been responded to")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, details, nil, nil, true
	case !consent.CanGrant(ctx.GetClock().Now()):
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldGranted: consent.Granted, logging.FieldExpiration: consent.ExpiresAt.Unix()}).
			Error("Error occurred performing consent during the Consent Flow stage of the Authorization Flow as the consent session has already been granted or is expired")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, details, nil, nil, true
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if details, err = authentication.MustGetUserDetailsSafe(userSession.Username, ctx.GetUserProvider()); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldGranted: consent.Granted, logging.FieldExpiration: consent.ExpiresAt.Unix()}).
			Error("Error occurred performing consent during the Consent Flow stage of the Authorization Flow as an error occurred retrieving user details")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, details, nil, nil, true
	}

	if !client.IsAuthenticationLevelSufficient(level, authorization.Subject{Username: userSession.Username, Groups: details.Groups, IP: ctx.RemoteIP()}) {
		ctx.GetLogger().
			WithFields(map[string]any{logging.FieldFlowID: flowID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldGroups: details.Groups, logging.FieldAuthenticationLevel: level.String(), logging.FieldAuthorizationPolicy: client.GetAuthorizationPolicy().Name}).
			Error("Error occurred performing consent during the Consent Flow stage of the Authorization Flow as the user is not sufficiently authenticated")

		ctx.SetJSONError(messageOperationFailed)

		return userSession, details, nil, nil, true
	}

	if consent.Subject.Valid {
		var subject uuid.UUID

		if subject, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
			ctx.GetLogger().
				WithError(err).
				WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}).
				Error("Error occurred performing consent during the Consent Flow stage of the Authorization Flow as an error occurred determining the subject of the user")

			ctx.SetJSONError(messageOperationFailed)

			return userSession, details, nil, nil, true
		}

		if consent.Subject.UUID != subject {
			ctx.GetLogger().
				WithFields(map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: userSession.Username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID, logging.FieldSubject: consent.Subject.UUID.String()}).
				Error("Error occurred performing consent during the Consent Flow stage of the Authorization Flow as the subject of the consent session does not match the subject of the user")

			ctx.SetJSONError(messageOperationFailed)

			return userSession, details, nil, nil, true
		}
	}

	return userSession, details, consent, client, false
}

func handleGetFormFromSession(ctx *middlewares.AutheliaCtx, session oidc.FormSession) (form url.Values, err error) {
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

	return original, err
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

func handleOAuth2ConsentDeviceAuthorizationSave(ctx *middlewares.AutheliaCtx, username string, bodyJSON oidc.ConsentPostRequestBody, consent *model.OAuth2ConsentSession, device *model.OAuth2DeviceCodeSession) (ok bool) {
	var (
		tx  context.Context
		err error
	)

	fields := func() map[string]any {
		return map[string]any{logging.FieldFlowID: consent.ChallengeID.String(), logging.FieldUsername: username, logging.FieldClientID: consent.ClientID, logging.FieldSessionID: consent.ID}
	}

	if tx, err = ctx.Providers.StorageProvider.BeginTX(ctx); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(fields()).
			Error("Error occurred starting the database transaction during the Consent Flow stage of the Device Authorization Flow")

		return false
	}

	defer func() {
		if ok {
			return
		}

		if err = ctx.Providers.StorageProvider.Rollback(tx); err != nil {
			ctx.GetLogger().
				WithError(err).
				WithFields(fields()).
				Error("Error occurred rolling back the database transaction during the Consent Flow stage of the Device Authorization Flow")
		}
	}()

	if bodyJSON.Consent {
		oidc.ConsentGrant(consent, true, bodyJSON.Claims)
	}

	consent.SetRespondedAt(ctx.GetClock().Now(), 0)

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(tx, consent, bodyJSON.Consent); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(fields()).
			Error("Error occurred saving the consent session response to the database during the Consent Flow stage of the Device Authorization Flow")

		return false
	}

	if !bodyJSON.Consent {
		if err = ctx.Providers.StorageProvider.DenyOAuth2DeviceCodeSession(tx, device.Signature, consent.ChallengeID); err != nil {
			ctx.GetLogger().
				WithError(err).
				WithFields(fields()).
				Error("Error occurred denying the device code session during the Consent Flow stage of the Device Authorization Flow as it is no longer pending a decision")

			return false
		}
	}

	if err = ctx.Providers.StorageProvider.Commit(tx); err != nil {
		ctx.GetLogger().
			WithError(err).
			WithFields(fields()).
			Error("Error occurred committing the database transaction during the Consent Flow stage of the Device Authorization Flow")

		return false
	}

	return true
}
