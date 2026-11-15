// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	oauth2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// OAuth2ConsentUserCodePOST exchanges the user code of the Device Authorization Flow for a consent session bound to the
// user agent, claiming the device code session for it. A user code the user agent already exchanged for a consent
// session which is still pending a response is exchanged for that consent session again.
func OAuth2ConsentUserCodePOST(ctx *middlewares.AutheliaCtx) {
	var (
		bodyJSON    bodyOAuth2ConsentUserCode
		userSession session.UserSession
		signature   string
		device      *model.OAuth2DeviceCodeSession
		client      oidc.Client
		subject     uuid.UUID
		consent     *model.OAuth2ConsentSession
		err         error
	)

	if err = json.Unmarshal(ctx.Request.Body(), &bodyJSON); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred unmarshaling the user code request body during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if n := len(bodyJSON.UserCode); n == 0 || n > 32 {
		ctx.GetLogger().Error("Error occurred exchanging the user code during the Consent Flow stage of the Device Authorization Flow as it is empty or too long")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred fetching user session during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.GetLogger().Error("Error occurred fetching user session during the Consent Flow stage of the Device Authorization Flow as the user is anonymous")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	log := ctx.GetLogger().WithFields(map[string]any{logging.FieldUsername: userSession.Username})

	if signature, err = ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(ctx, bodyJSON.UserCode); err != nil {
		log.WithError(err).Error("Error occurred determining the signature of the user code during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if device, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, signature); err != nil {
		log.WithError(err).Error("Error occurred using the signature of the user code to retrieve the device code session during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	log = log.WithFields(map[string]any{logging.FieldClientID: device.ClientID, logging.FieldSessionID: device.ID})

	if client, subject, err = handleOAuth2ConsentUserCodeSubject(ctx, userSession, device); err != nil {
		log.WithError(err).Error("Error occurred authorizing the user to enter the user code during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if device.ChallengeID.Valid {
		consent, err = handleOAuth2ConsentUserCodeResume(ctx, device, subject)
	} else {
		consent, err = handleOAuth2ConsentUserCodeExchange(ctx, client, device, subject)
	}

	if err != nil {
		log.WithError(err).Error("Error occurred exchanging the user code for a consent session during the Consent Flow stage of the Device Authorization Flow")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(oidc.ConsentPostResponseBody{FlowID: consent.ChallengeID.String()}); err != nil {
		log.WithError(err).Error("Error occurred marshaling JSON response body")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.SetUserValue(middlewares.UserValueRateLimitExempt, true)
}

func handleOAuth2ConsentUserCodeSubject(ctx *middlewares.AutheliaCtx, userSession session.UserSession, device *model.OAuth2DeviceCodeSession) (client oidc.Client, subject uuid.UUID, err error) {
	var details authentication.UserDetails

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, device.ClientID); err != nil {
		return nil, subject, fmt.Errorf("error occurred fetching the client: %w", err)
	}

	if details, err = authentication.MustGetUserDetailsSafe(userSession.Username, ctx.GetUserProvider()); err != nil {
		return nil, subject, fmt.Errorf("error occurred fetching the user details: %w", err)
	}

	level := userSession.AuthenticationLevel(ctx.Configuration.WebAuthn.EnablePasskey2FA)

	if !client.IsAuthenticationLevelSufficient(level, authorization.Subject{Username: userSession.Username, Groups: details.Groups, IP: ctx.RemoteIP()}) {
		return nil, subject, fmt.Errorf("the user is not sufficiently authenticated given the client authorization policy '%s' with the authentication level '%s'", client.GetAuthorizationPolicy().Name, level)
	}

	if subject, err = ctx.Providers.OpenIDConnect.GetSubject(ctx, client.GetSectorIdentifierURI(), userSession.Username); err != nil {
		return nil, subject, fmt.Errorf("error occurred determining the subject: %w", err)
	}

	if device.Subject.Valid && device.Subject.String != subject.String() {
		return nil, subject, errors.New("the subject of the device code session does not match the subject of the user")
	}

	return client, subject, nil
}

func handleOAuth2ConsentUserCodeExchange(ctx *middlewares.AutheliaCtx, client oidc.Client, device *model.OAuth2DeviceCodeSession, subject uuid.UUID) (consent *model.OAuth2ConsentSession, err error) {
	if err = validateOAuth2DeviceCodeSessionEnterable(ctx, device); err != nil {
		return nil, err
	}

	var (
		requester oauth2.DeviceAuthorizeRequester
		expires   time.Time
		binding   string
	)

	if requester, err = device.ToRequest(ctx, oidc.NewSession(), ctx.Providers.OpenIDConnect.Store); err != nil {
		return nil, fmt.Errorf("error occurred restoring the requester: %w", err)
	}

	if expires, err = getOAuth2DeviceCodeSessionUserCodeExpiresAt(device); err != nil {
		return nil, err
	}

	if consent, err = model.NewOAuth2ConsentSession(expires.UTC(), subject, requester); err != nil {
		return nil, fmt.Errorf("error occurred creating the consent session for client '%s': %w", client.GetID(), err)
	}

	if binding, err = handleOAuth2ConsentSessionBind(ctx, consent); err != nil {
		return nil, err
	}

	claim := *device

	claim.ChallengeID = uuid.NullUUID{UUID: consent.ChallengeID, Valid: true}
	claim.Status = model.OAuth2DeviceCodeStatusClaimed

	if err = handleOAuth2ConsentUserCodeClaim(ctx, consent, &claim); err != nil {
		return nil, err
	}

	if err = handleOAuth2ConsentSessionIssueBinding(ctx, binding); err != nil {
		return nil, err
	}

	return consent, nil
}

func handleOAuth2ConsentUserCodeClaim(ctx *middlewares.AutheliaCtx, consent *model.OAuth2ConsentSession, device *model.OAuth2DeviceCodeSession) (err error) {
	var tx context.Context

	if tx, err = ctx.Providers.StorageProvider.BeginTX(ctx); err != nil {
		return fmt.Errorf("error occurred starting the database transaction: %w", err)
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(tx, consent); err != nil {
		err = fmt.Errorf("error occurred saving the consent session: %w", err)
	} else if err = ctx.Providers.StorageProvider.ClaimOAuth2DeviceCodeSession(tx, device); err != nil {
		err = fmt.Errorf("error occurred claiming the device code session as the user code has already been entered or can no longer be used: %w", err)
	} else if err = ctx.Providers.StorageProvider.Commit(tx); err != nil {
		return fmt.Errorf("error occurred committing the database transaction: %w", err)
	} else {
		return nil
	}

	if rerr := ctx.Providers.StorageProvider.Rollback(tx); rerr != nil {
		ctx.GetLogger().WithError(rerr).Error("Error occurred rolling back the database transaction during the Consent Flow stage of the Device Authorization Flow")
	}

	return err
}

func handleOAuth2ConsentUserCodeResume(ctx *middlewares.AutheliaCtx, device *model.OAuth2DeviceCodeSession, subject uuid.UUID) (consent *model.OAuth2ConsentSession, err error) {
	if err = validateOAuth2DeviceCodeSessionClaimed(ctx, device); err != nil {
		return nil, err
	}

	if consent, err = handleOAuth2ConsentSessionLoadBound(ctx, device.ChallengeID.UUID); err != nil {
		return nil, fmt.Errorf("the user code has already been entered by another user agent: %w", err)
	}

	switch {
	case consent.Responded():
		return nil, errors.New("the consent session has already been responded to")
	case !consent.CanGrant(ctx.GetClock().Now()):
		return nil, errors.New("the consent session has already been granted or is expired")
	case !consent.Subject.Valid || consent.Subject.UUID != subject:
		return nil, errors.New("the subject of the consent session does not match the subject of the user")
	}

	var binding string

	if binding, err = handleOAuth2ConsentSessionBind(ctx, consent); err != nil {
		return nil, err
	}

	if err = handleOAuth2ConsentSessionIssueBinding(ctx, binding); err != nil {
		return nil, err
	}

	return consent, nil
}
