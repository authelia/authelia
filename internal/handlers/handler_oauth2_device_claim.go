// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

var (
	errOAuth2DeviceCodeSessionInactive = errors.New("the device code session is inactive or revoked")
	errOAuth2DeviceCodeSessionEntered  = errors.New("the user code has already been entered")
	errOAuth2DeviceCodeSessionExpired  = errors.New("the user code has expired")
)

func validateOAuth2DeviceCodeSessionEnterable(ctx *middlewares.AutheliaCtx, device *model.OAuth2DeviceCodeSession) (err error) {
	switch {
	case !device.Active || device.Revoked:
		return errOAuth2DeviceCodeSessionInactive
	case device.ChallengeID.Valid || device.Status != model.OAuth2DeviceCodeStatusNew:
		return errOAuth2DeviceCodeSessionEntered
	}

	session := oidc.NewSession()

	if err = json.Unmarshal(device.Session, session); err != nil {
		return fmt.Errorf("error occurred restoring the device code session data: %w", err)
	}

	if session.GetExpiresAt(oauthelia2.UserCode).Before(ctx.GetClock().Now()) {
		return errOAuth2DeviceCodeSessionExpired
	}

	return nil
}

func validateOAuth2DeviceAuthorizationConsent(ctx *middlewares.AutheliaCtx, requester oauthelia2.DeviceAuthorizeRequester, consent *model.OAuth2ConsentSession) (err error) {
	switch {
	case consent.ClientID != requester.GetClient().GetID():
		return fmt.Errorf("the consent session belongs to client '%s' rather than client '%s'", consent.ClientID, requester.GetClient().GetID())
	case !consent.Responded() || !consent.IsAuthorized():
		return errors.New("the consent session has not been authorized")
	case !consent.CanGrant(ctx.GetClock().Now()):
		return errors.New("the consent session has already been used or has expired")
	case requester.GetStatus() != oauthelia2.DeviceAuthorizeStatusClaimed:
		return fmt.Errorf("the device code session has the status '%s' rather than 'Claimed'", oauthelia2.DeviceAuthorizeStatusToString(requester.GetStatus()))
	}

	var device *model.OAuth2DeviceCodeSession

	if device, err = ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(ctx, requester.GetUserCodeSignature()); err != nil {
		return fmt.Errorf("error occurred loading the device code session: %w", err)
	}

	if !device.ChallengeID.Valid || device.ChallengeID.UUID != consent.ChallengeID {
		return errors.New("the device code session was claimed by another consent session")
	}

	return nil
}
