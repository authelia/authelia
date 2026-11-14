// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/appauthenticator"
	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

const (
	appPortalStatePending  = "pending"
	appPortalStateDenied   = "denied"
	appPortalStateRejected = "rejected"

	appConsumeGrace = time.Minute
)

type appChallengePOSTResponse struct {
	Number      string `json:"number"`
	NonceLength int    `json:"nonce_length"`
	ExpiresAt   int64  `json:"expires_at"`
}

type appStatusResponse struct {
	State string `json:"state"`
}

// AppAuthenticatorPOST creates an Authelia app second factor challenge for the session.
func AppAuthenticatorPOST(ctx *middlewares.AutheliaCtx) {
	bodyJSON, ok := appParseSignBody(ctx)
	if !ok {
		return
	}

	userSession, ok := appPortalSession(ctx, "creating the app authenticator challenge", fasthttp.StatusUnauthorized, messageMFAValidationFailed)
	if !ok {
		return
	}

	if !appCheckRegulation(ctx, userSession.Username) {
		return
	}

	devices, err := ctx.Providers.StorageProvider.LoadAppAuthenticatorDevicesByUsername(ctx, userSession.Username)
	if err != nil {
		appPortalError(ctx, err, "Error occurred loading the app authenticator devices", fasthttp.StatusUnauthorized, messageMFAValidationFailed)

		return
	}

	if len(devices) == 0 {
		ctx.Logger.WithField(logging.FieldUsername, userSession.Username).Error("Error occurred creating the app authenticator challenge: the user has no registered app authenticator devices")

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	binding, ok := appPortalBindingEnsure(ctx, &userSession, "creating the app authenticator challenge", fasthttp.StatusUnauthorized, messageMFAValidationFailed)
	if !ok {
		return
	}

	challenge, err := appNewChallenge(ctx, userSession.Username, binding, bodyJSON.TargetURL)
	if err != nil {
		appPortalError(ctx, err, "Error occurred generating the app authenticator challenge", fasthttp.StatusUnauthorized, messageMFAValidationFailed)

		return
	}

	if err = appReplaceChallenge(ctx, binding, challenge); err != nil {
		appPortalError(ctx, err, "Error occurred saving the app authenticator challenge", fasthttp.StatusUnauthorized, messageMFAValidationFailed)

		return
	}

	ctx.Logger.WithFields(map[string]any{
		logging.FieldUsername: userSession.Username,
		"portal_url":          challenge.PortalURL.String,
		"requested_url":       challenge.RequestedURL.String,
	}).Info("App authenticator challenge created")

	appPortalJSON(ctx, appChallengePOSTResponse{Number: challenge.Number, NonceLength: challenge.NonceLength, ExpiresAt: challenge.ExpiresAt.Unix()}, messageMFAValidationFailed)
}

func appReplaceChallenge(ctx *middlewares.AutheliaCtx, binding []byte, challenge model.AppAuthenticatorChallenge) (err error) {
	provider := ctx.Providers.StorageProvider

	tx, err := provider.BeginTX(ctx)
	if err != nil {
		return fmt.Errorf("error beginning the transaction: %w", err)
	}

	if err = provider.VoidAppAuthenticatorChallengesByPortalBinding(tx, binding); err != nil {
		_ = provider.Rollback(tx)

		return fmt.Errorf("error voiding the previous challenges: %w", err)
	}

	if err = provider.SaveAppAuthenticatorChallenge(tx, challenge); err != nil {
		_ = provider.Rollback(tx)

		return fmt.Errorf("error saving the challenge: %w", err)
	}

	if err = provider.Commit(tx); err != nil {
		return fmt.Errorf("error committing the transaction: %w", err)
	}

	return nil
}

// AppAuthenticatorStatusPOST returns the state of the Authelia app challenge of the session, and upgrades the session
// to two factor on the first poll after the challenge was approved.
func AppAuthenticatorStatusPOST(ctx *middlewares.AutheliaCtx) {
	bodyJSON, ok := appParseSignBody(ctx)
	if !ok {
		return
	}

	userSession, ok := appPortalSession(ctx, "retrieving the app authenticator challenge", fasthttp.StatusUnauthorized, messageMFAValidationFailed)
	if !ok {
		return
	}

	log := ctx.Logger.WithField(logging.FieldUsername, userSession.Username)

	binding := appPortalBinding(&userSession)
	if binding == nil {
		log.Error("Error occurred retrieving the app authenticator challenge: the session has no portal binding")

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	challenge, err := ctx.Providers.StorageProvider.LoadAppAuthenticatorChallengeByPortalBinding(ctx, binding)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		log.Error("Error occurred retrieving the app authenticator challenge: the session has no challenge")

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	case err != nil:
		appPortalError(ctx, err, "Error occurred loading the app authenticator challenge", fasthttp.StatusUnauthorized, messageMFAValidationFailed)

		return
	case challenge.Username != userSession.Username:
		log.Error("Error occurred retrieving the app authenticator challenge: the challenge belongs to another user")

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	now := ctx.GetClock().Now()

	var state string

	switch challenge.State {
	case model.AppAuthenticatorChallengeStatePending:
		if now.After(challenge.ExpiresAt) {
			state = appPortalStateExpired
		} else {
			state = appPortalStatePending
		}
	case model.AppAuthenticatorChallengeStateApproved:
		if !challenge.ConsumedAt.Valid && now.After(challenge.ExpiresAt.Add(appConsumeGrace)) {
			state = appPortalStateExpired

			break
		}

		appApprove(ctx, &userSession, challenge, bodyJSON)

		return
	case model.AppAuthenticatorChallengeStateDenied:
		state = appPortalStateDenied
	case model.AppAuthenticatorChallengeStateExpired:
		state = appPortalStateExpired
	default:
		state = appPortalStateRejected
	}

	appPortalJSON(ctx, appStatusResponse{State: state}, messageMFAValidationFailed)
}

func appApprove(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, challenge *model.AppAuthenticatorChallenge, bodyJSON *bodySignDuoRequest) {
	log := ctx.Logger.WithField(logging.FieldUsername, userSession.Username)

	if challenge.ConsumedAt.Valid {
		log.Error("Error occurred upgrading the session with the app authenticator challenge: the challenge was already consumed")

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if !challenge.DecidedDeviceID.Valid {
		log.Error("Error occurred upgrading the session with the app authenticator challenge: the challenge has no deciding device")

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	device, err := ctx.Providers.StorageProvider.LoadAppAuthenticatorDeviceByDeviceID(ctx, challenge.DecidedDeviceID.String)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WithField(appFieldDeviceID, challenge.DecidedDeviceID.String).Error("Error occurred upgrading the session with the app authenticator challenge: the approving device was removed")
		} else {
			log.WithError(err).Error("Error occurred upgrading the session with the app authenticator challenge: error loading the approving device")
		}

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	attested := device.AttestationFormat != "" && device.AttestationFormat != appauthenticator.AttestationFormatNone

	var details *authentication.UserDetails

	if details, err = ctx.GetUserProvider().GetDetails(userSession.Username); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrObtainProfileDetails, regulation.AuthTypeApp, userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	now := ctx.GetClock().Now()

	consumed, err := ctx.Providers.StorageProvider.ConsumeAppAuthenticatorChallenge(ctx, challenge.ID, now, now.Add(-appConsumeGrace))
	if err != nil {
		appPortalError(ctx, err, "Error occurred consuming the app authenticator challenge", fasthttp.StatusUnauthorized, messageMFAValidationFailed)

		return
	}

	if !consumed {
		log.Error("Error occurred consuming the app authenticator challenge: the challenge was already consumed or expired")

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	if err = ctx.RegenerateSession(); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionRegenerate, regulation.AuthTypeApp, userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	userSession.SetTwoFactorApp(now, attested, attested)

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "authentication time", regulation.AuthTypeApp, logFmtActionAuthentication, userSession.Username)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	log.Info("App authenticator challenge consumed and the session was upgraded to two factor")

	if len(bodyJSON.Flow) > 0 {
		handleFlowResponse(ctx, userSession, details, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
	} else {
		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}

func appParseSignBody(ctx *middlewares.AutheliaCtx) (*bodySignDuoRequest, bool) {
	bodyJSON := &bodySignDuoRequest{}

	if len(ctx.PostBody()) == 0 {
		return bodyJSON, true
	}

	if err := json.Unmarshal(ctx.PostBody(), bodyJSON); err != nil {
		ctx.Logger.WithError(err).Errorf(logFmtErrParseRequestBody, regulation.AuthTypeApp)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return nil, false
	}

	return bodyJSON, true
}

func appCheckRegulation(ctx *middlewares.AutheliaCtx, username string) bool {
	ban, value, expires, err := ctx.Providers.Regulator.BanCheck(ctx, username)
	if err == nil {
		return true
	}

	if errors.Is(err, regulation.ErrUserIsBanned) {
		doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(ban, value, expires), regulation.AuthTypeApp, nil)
	} else {
		ctx.Logger.WithError(err).Errorf(logFmtErrRegulationFail, regulation.AuthTypeApp, username)
	}

	respondUnauthorized(ctx, messageMFAValidationFailed)

	return false
}

func appNewChallenge(ctx *middlewares.AutheliaCtx, username string, binding []byte, targetURL string) (challenge model.AppAuthenticatorChallenge, err error) {
	if challenge.ConfirmationID, err = appauthenticator.NewID(); err != nil {
		return challenge, err
	}

	if challenge.Challenge, err = appauthenticator.NewID(); err != nil {
		return challenge, err
	}

	challenge.NonceLength = appNonceLength(ctx)

	if challenge.Number, err = appauthenticator.NewNumber(challenge.NonceLength); err != nil {
		return challenge, err
	}

	now := ctx.GetClock().Now()

	challenge.CreatedAt = now
	challenge.ExpiresAt = now.Add(appChallengeLifespan(ctx))
	challenge.Username = username
	challenge.PortalBinding = binding
	challenge.State = model.AppAuthenticatorChallengeStatePending
	challenge.UserAgent = appNullString(appUserAgentSummary(string(ctx.UserAgent())))

	if ip := ctx.RemoteIP(); ip != nil {
		challenge.RemoteIP = appNullString(ip.String())
	}

	if portalURL, err := appPortalURL(ctx); err == nil {
		challenge.PortalURL = appNullString(portalURL.String())
	} else {
		ctx.Logger.WithError(err).Debug("Unable to determine the portal url for the app authenticator challenge details")
	}

	if targetURL != "" {
		if target, err := url.ParseRequestURI(targetURL); err == nil && ctx.IsSafeRedirectionTargetURI(target) {
			challenge.RequestedURL = appNullString(target.String())
		}
	}

	return challenge, nil
}

var (
	appUserAgentBrowsers = []struct{ token, name string }{
		{"Edg", "Edge"},
		{"OPR/", "Opera"},
		{"Opera", "Opera"},
		{"FxiOS/", "Firefox"},
		{"Firefox/", "Firefox"},
		{"CriOS/", "Chrome"},
		{"Chrome/", "Chrome"},
		{"Safari/", "Safari"},
	}

	appUserAgentSystems = []struct{ token, name string }{
		{"iPhone", "iOS"},
		{"iPod", "iOS"},
		{"iPad", "iPadOS"},
		{"Android", "Android"},
		{"CrOS", "ChromeOS"},
		{"Windows", "Windows"},
		{"Mac OS X", "macOS"},
		{"Macintosh", "macOS"},
		{"Linux", "Linux"},
	}
)

func appUserAgentSummary(userAgent string) string {
	if userAgent == "" {
		return ""
	}

	browser := "Unknown browser"

	for _, b := range appUserAgentBrowsers {
		if strings.Contains(userAgent, b.token) {
			browser = b.name

			break
		}
	}

	for _, s := range appUserAgentSystems {
		if strings.Contains(userAgent, s.token) {
			return browser + " on " + s.name
		}
	}

	return browser
}
