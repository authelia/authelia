// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"bytes"
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/appauthenticator"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/storage"
)

const (
	appUserValueRandom = "random"
	appUserValueID     = "id"
	appUserValuePullID = "pull_id"
)

type appRegistrationBeginResponse struct {
	Challenge       string `json:"challenge"`
	ConfirmationURL string `json:"confirmation_url"`
	ExpiresAt       int64  `json:"expires_at"`
	Attestation     string `json:"attestation"`
	NonceLength     int    `json:"nonce_length"`
	ServerTime      int64  `json:"server_time"`
}

type appRegistrationConfirmResponse struct {
	BaseURL     string `json:"base_url"`
	DeviceID    string `json:"device_id"`
	PullID      string `json:"pull_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	ServerTime  int64  `json:"server_time"`
}

type appCheckChallenge struct {
	Challenge      string         `json:"challenge"`
	ConfirmationID string         `json:"confirmation_id"`
	ExpiresAt      int64          `json:"expires_at"`
	NonceLength    int            `json:"nonce_length"`
	Details        map[string]any `json:"details"`
}

type appCheckResponse struct {
	PullID     string             `json:"pull_id"`
	Challenge  *appCheckChallenge `json:"challenge"`
	ServerTime int64              `json:"server_time"`
}

type appConfirmResponse struct {
	Result     string `json:"result"`
	ServerTime int64  `json:"server_time"`
}

type appServerTimeResponse struct {
	ServerTime int64 `json:"server_time"`
}

// AppAuthenticatorRegistrationBeginPOST handles the device request which begins a registration from a QR code.
func AppAuthenticatorRegistrationBeginPOST(ctx *middlewares.AutheliaCtx) {
	random := appUserValue(ctx, appUserValueRandom)

	if !appDevicePrecheck(ctx, random) {
		return
	}

	body, ok := appParseObject(ctx)
	if !ok {
		return
	}

	registration, err := ctx.Providers.StorageProvider.LoadAppAuthenticatorRegistrationByCode(ctx, appauthenticator.HashID(random))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			appWriteError(ctx, fasthttp.StatusNotFound, appErrorUnknown)
		} else {
			appWriteInternal(ctx, err, "Error occurred loading the app authenticator registration code")
		}

		return
	}

	now := ctx.GetClock().Now()

	if !appRegistrationCodeUsable(ctx, registration, now) {
		return
	}

	codeChallenge, ok := appCodeChallenge(body)
	if !ok {
		appWriteError(ctx, fasthttp.StatusBadRequest, appErrorInvalidRequest)

		return
	}

	device, ok := body.object("device")
	if !ok {
		appWriteError(ctx, fasthttp.StatusBadRequest, appErrorInvalidRequest)

		return
	}

	sessionID, err := appauthenticator.NewID()
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred generating the app authenticator registration session id")

		return
	}

	challenge, err := appauthenticator.NewID()
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred generating the app authenticator registration challenge")

		return
	}

	platform, _ := device.string("platform")
	deviceModel, _ := device.string("model")
	name, _ := device.string("name")

	registration.State = model.AppAuthenticatorRegistrationStateBegun
	registration.BegunAt = sql.NullTime{Time: now, Valid: true}
	registration.SessionExpiresAt = sql.NullTime{Time: now.Add(appRegistrationLifespan(ctx)), Valid: true}
	registration.ConfirmationID = appauthenticator.HashID(sessionID)
	registration.Challenge = sql.NullString{String: challenge, Valid: true}
	registration.CodeChallenge = sql.NullString{String: codeChallenge, Valid: true}
	registration.DevicePlatform = sql.NullString{String: appTruncate(platform, appDevicePlatformMaxLength), Valid: true}
	registration.DeviceModel = sql.NullString{String: appTruncate(deviceModel, appDeviceInfoMaxLength), Valid: true}
	registration.DeviceName = sql.NullString{String: appTruncate(name, appDeviceInfoMaxLength), Valid: true}

	updated, err := ctx.Providers.StorageProvider.UpdateAppAuthenticatorRegistration(ctx, *registration, model.AppAuthenticatorRegistrationStateIssued)
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred beginning the app authenticator registration")

		return
	}

	if !updated {
		if registration, err = ctx.Providers.StorageProvider.LoadAppAuthenticatorRegistrationByCode(ctx, appauthenticator.HashID(random)); err != nil {
			appWriteInternal(ctx, err, "Error occurred reloading the app authenticator registration code")

			return
		}

		if appRegistrationCodeUsable(ctx, registration, now) {
			appWriteError(ctx, fasthttp.StatusConflict, appErrorAlreadyUsed)
		}

		return
	}

	ctx.Logger.WithFields(map[string]any{
		logging.FieldUsername: registration.Username,
		"platform":            registration.DevicePlatform.String,
		"model":               registration.DeviceModel.String,
		"name":                registration.DeviceName.String,
	}).Info("App authenticator registration begun")

	appWriteJSON(ctx, fasthttp.StatusOK, appRegistrationBeginResponse{
		Challenge:       challenge,
		ConfirmationURL: registration.BaseURL + "/api/register/app/confirm/" + sessionID,
		ExpiresAt:       registration.SessionExpiresAt.Time.Unix(),
		Attestation:     appAttestationMode(ctx),
		NonceLength:     registration.NonceLength,
		ServerTime:      now.Unix(),
	})
}

// AppAuthenticatorRegistrationConfirmPOST handles the device request which completes a registration.
//
//nolint:gocyclo
func AppAuthenticatorRegistrationConfirmPOST(ctx *middlewares.AutheliaCtx) {
	id := appUserValue(ctx, appUserValueID)

	if !appDevicePrecheck(ctx, id) {
		return
	}

	body, ok := appParseObject(ctx)
	if !ok {
		return
	}

	registration, err := ctx.Providers.StorageProvider.LoadAppAuthenticatorRegistrationByConfirmationID(ctx, appauthenticator.HashID(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			appWriteError(ctx, fasthttp.StatusNotFound, appErrorUnknown)
		} else {
			appWriteInternal(ctx, err, "Error occurred loading the app authenticator registration session")
		}

		return
	}

	now := ctx.GetClock().Now()

	if registration.State != model.AppAuthenticatorRegistrationStateBegun {
		appWriteError(ctx, fasthttp.StatusConflict, appErrorAlreadyUsed)

		return
	}

	if !registration.SessionExpiresAt.Valid || now.After(registration.SessionExpiresAt.Time) {
		if appRegistrationSettle(ctx, registration, model.AppAuthenticatorRegistrationStateExpired) {
			ctx.Logger.WithField(logging.FieldUsername, registration.Username).Info("App authenticator registration failed: the registration session expired")

			appWriteError(ctx, fasthttp.StatusGone, appErrorExpired)
		}

		return
	}

	verifier, ok := body.string("verifier")
	if !ok || !appauthenticator.IsValidVerifier(verifier) {
		appWriteError(ctx, fasthttp.StatusBadRequest, appErrorInvalidRequest)

		return
	}

	approvalSPKI, ok1 := body.base64("approval_public_key")
	transportSPKI, ok2 := body.base64("transport_public_key")
	approvalSignature, ok3 := body.string("approval_signature")
	transportSignature, ok4 := body.string("transport_signature")
	attestation, ok5 := body.object("attestation")

	var format string

	if ok5 {
		format, ok5 = attestation.string("format")
	}

	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || !appauthenticator.IsBase64URL(approvalSignature) || !appauthenticator.IsBase64URL(transportSignature) {
		appWriteError(ctx, fasthttp.StatusBadRequest, appErrorInvalidRequest)

		return
	}

	fail := func(reason string) {
		if appRegistrationSettle(ctx, registration, model.AppAuthenticatorRegistrationStateVoided) {
			ctx.Logger.WithField(logging.FieldUsername, registration.Username).Infof("App authenticator registration failed: %s", reason)

			appWriteError(ctx, fasthttp.StatusUnauthorized, appErrorBadSignature)
		}
	}

	if subtle.ConstantTimeCompare([]byte(appauthenticator.CodeChallengeS256(verifier)), []byte(registration.CodeChallenge.String)) != 1 {
		fail("the verifier does not match the code challenge")

		return
	}

	registrationString := appauthenticator.RegistrationString(registration.BaseURL, registration.Challenge.String, registration.Number, verifier,
		appauthenticator.Thumbprint(approvalSPKI), appauthenticator.Thumbprint(transportSPKI))

	keyID, _ := attestation.string("key_id")
	attestationObject, _ := attestation.string("attestation_object")

	if ctx.Providers.AppAuthenticator == nil {
		appWriteInternal(ctx, errors.New("the attestation verifier is not configured"), "Error occurred verifying the app authenticator attestation")

		return
	}

	result, err := ctx.Providers.AppAuthenticator.Verify(ctx, appauthenticator.AttestationInput{
		Format:             format,
		KeyID:              keyID,
		AttestationObject:  attestationObject,
		ApprovalChain:      attestation.strings("approval_certificate_chain"),
		TransportChain:     attestation.strings("transport_certificate_chain"),
		ApprovalSPKI:       approvalSPKI,
		TransportSPKI:      transportSPKI,
		Challenge:          registration.Challenge.String,
		RegistrationString: registrationString,
	})
	if err != nil {
		fail("the attestation was not acceptable: " + err.Error())

		return
	}

	approvalKey, err := appauthenticator.ParsePublicKey(approvalSPKI)
	if err != nil {
		fail("the approval key is not usable: " + err.Error())

		return
	}

	transportKey, err := appauthenticator.ParsePublicKey(transportSPKI)
	if err != nil {
		fail("the transport key is not usable: " + err.Error())

		return
	}

	if bytes.Equal(approvalSPKI, transportSPKI) {
		fail("the approval and transport keys are identical")

		return
	}

	if !appauthenticator.VerifySignature(approvalKey, registrationString, approvalSignature) {
		fail("the approval signature is invalid")

		return
	}

	if !appauthenticator.VerifySignature(transportKey, registrationString, transportSignature) {
		fail("the transport signature is invalid")

		return
	}

	appRegistrationComplete(ctx, registration, result, approvalSPKI, transportSPKI, now)
}

func appRegistrationComplete(ctx *middlewares.AutheliaCtx, registration *model.AppAuthenticatorRegistration, result appauthenticator.AttestationResult, approvalSPKI, transportSPKI []byte, now time.Time) {
	deviceID, err := appauthenticator.NewID()
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred generating the app authenticator device id")

		return
	}

	pullID, err := appauthenticator.NewID()
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred generating the app authenticator pull id")

		return
	}

	device := model.AppAuthenticatorDevice{
		CreatedAt:           now,
		DeviceID:            deviceID,
		Username:            registration.Username,
		BaseURL:             registration.BaseURL,
		Name:                registration.DeviceName.String,
		Model:               registration.DeviceModel.String,
		Platform:            registration.DevicePlatform.String,
		ApprovalPublicKey:   approvalSPKI,
		TransportPublicKey:  transportSPKI,
		AttestationFormat:   result.Format,
		AttestationIdentity: appNullString(appTruncate(result.Identity, appAttestationIdentityMaxLength)),
		AttestationDetail:   appNullString(appTruncate(result.Detail, appAttestationDetailMaxLength)),
		AttestationKeyID:    result.KeyID,
		PullIDCurrent:       appauthenticator.HashID(pullID),
	}

	completed := *registration
	completed.State = model.AppAuthenticatorRegistrationStateCompleted
	completed.DeviceID = sql.NullString{String: deviceID, Valid: true}

	provider := ctx.Providers.StorageProvider

	tx, err := provider.BeginTX(ctx)
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred beginning the app authenticator registration transaction")

		return
	}

	updated, err := provider.UpdateAppAuthenticatorRegistration(tx, completed, model.AppAuthenticatorRegistrationStateBegun)
	if err == nil && updated {
		err = provider.SaveAppAuthenticatorDevice(tx, device)
	}

	if err != nil || !updated {
		_ = provider.Rollback(tx)

		if err != nil {
			appWriteInternal(ctx, err, "Error occurred completing the app authenticator registration")
		} else {
			appWriteError(ctx, fasthttp.StatusConflict, appErrorAlreadyUsed)
		}

		return
	}

	if err = provider.Commit(tx); err != nil {
		appWriteInternal(ctx, err, "Error occurred committing the app authenticator registration")

		return
	}

	ctx.Logger.WithFields(map[string]any{
		logging.FieldUsername: device.Username,
		appFieldDeviceID:      device.DeviceID,
		"attestation_format":  device.AttestationFormat,
		"attestation_id":      device.AttestationIdentity.String,
	}).Info("App authenticator registration completed")

	ctxLogAppAuthenticatorEvent(ctx, &device, true)

	displayName := ctx.Configuration.AppAuthenticator.DisplayName
	if displayName == "" {
		displayName = schema.DefaultAppAuthenticatorConfiguration.DisplayName
	}

	appWriteJSON(ctx, fasthttp.StatusOK, appRegistrationConfirmResponse{
		BaseURL:     device.BaseURL,
		DeviceID:    deviceID,
		PullID:      pullID,
		Username:    device.Username,
		DisplayName: displayName,
		ServerTime:  now.Unix(),
	})
}

// AppAuthenticatorChallengeCheckPOST handles the device request which checks for a pending challenge.
func AppAuthenticatorChallengeCheckPOST(ctx *middlewares.AutheliaCtx) {
	pullID := appUserValue(ctx, appUserValuePullID)

	if !appDevicePrecheck(ctx, pullID) {
		return
	}

	device, ok := appAuthenticate(ctx)
	if !ok {
		return
	}

	body, ok := appParseObject(ctx)
	if !ok {
		return
	}

	codeChallenge, ok := appCodeChallenge(body)
	if !ok {
		appWriteError(ctx, fasthttp.StatusBadRequest, appErrorInvalidRequest)

		return
	}

	now := ctx.GetClock().Now()

	for attempt := 0; attempt < appPullIDAttempts; attempt++ {
		if attempt != 0 {
			var err error

			if device, err = ctx.Providers.StorageProvider.LoadAppAuthenticatorDeviceByDeviceID(ctx, device.DeviceID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					appAuthFailure(ctx, appErrorBadSignature, device.DeviceID, "the device was removed concurrently")
				} else {
					appWriteInternal(ctx, err, "Error occurred reloading the app authenticator device")
				}

				return
			}
		}

		if appCheckAttempt(ctx, device, pullID, codeChallenge, now) {
			return
		}
	}

	ctx.Logger.Warn("App authenticator check lost the pull id rotation race on every attempt")

	appWriteError(ctx, fasthttp.StatusNotFound, appErrorUnknown)
}

func appCheckAttempt(ctx *middlewares.AutheliaCtx, device *model.AppAuthenticatorDevice, pullID, codeChallenge string, now time.Time) (done bool) {
	provider := ctx.Providers.StorageProvider

	tx, err := provider.BeginTX(ctx)
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred beginning the app authenticator check transaction")

		return true
	}

	next, outcome, err := appRotatePullID(tx, provider, device, pullID)

	switch {
	case err != nil:
		_ = provider.Rollback(tx)

		appWriteInternal(ctx, err, "Error occurred rotating the app authenticator pull id")

		return true
	case outcome == appRotationUnknown:
		_ = provider.Rollback(tx)

		appWriteError(ctx, fasthttp.StatusNotFound, appErrorUnknown)

		return true
	case outcome == appRotationLost:
		_ = provider.Rollback(tx)

		return false
	}

	challenge, err := appBindPendingChallenge(tx, provider, device, codeChallenge, now)
	if err != nil {
		_ = provider.Rollback(tx)

		appWriteInternal(ctx, err, "Error occurred binding the pending app authenticator challenge")

		return true
	}

	if err = provider.Commit(tx); err != nil {
		appWriteInternal(ctx, err, "Error occurred committing the app authenticator check")

		return true
	}

	response := appCheckResponse{PullID: next, ServerTime: now.Unix()}

	if challenge != nil {
		response.Challenge = &appCheckChallenge{
			Challenge:      challenge.Challenge,
			ConfirmationID: challenge.ConfirmationID,
			ExpiresAt:      challenge.ExpiresAt.Unix(),
			NonceLength:    challenge.NonceLength,
			Details:        appChallengeDetails(challenge),
		}
	}

	appWriteJSON(ctx, fasthttp.StatusOK, response)

	return true
}

// AppAuthenticatorChallengeConfirmPOST handles the device request which approves or denies a challenge.
func AppAuthenticatorChallengeConfirmPOST(ctx *middlewares.AutheliaCtx) {
	id := appUserValue(ctx, appUserValueID)

	if !appDevicePrecheck(ctx, id) {
		return
	}

	device, ok := appAuthenticate(ctx)
	if !ok {
		return
	}

	body, ok := appParseObject(ctx)
	if !ok {
		return
	}

	challenge, err := ctx.Providers.StorageProvider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		appWriteInternal(ctx, err, "Error occurred loading the app authenticator challenge")

		return
	}

	if err != nil || challenge.Username != device.Username {
		appWriteError(ctx, fasthttp.StatusNotFound, appErrorUnknown)

		return
	}

	decision, _ := body.string("decision")
	if decision != appDecisionApprove && decision != appDecisionDeny {
		appWriteError(ctx, fasthttp.StatusBadRequest, appErrorInvalidRequest)

		return
	}

	now := ctx.GetClock().Now()

	switch {
	case challenge.State == model.AppAuthenticatorChallengeStatePending && now.After(challenge.ExpiresAt):
		expired := *challenge
		expired.State = model.AppAuthenticatorChallengeStateExpired

		if _, err = ctx.Providers.StorageProvider.UpdateAppAuthenticatorChallenge(ctx, expired, model.AppAuthenticatorChallengeStatePending); err != nil {
			appWriteInternal(ctx, err, "Error occurred expiring the app authenticator challenge")

			return
		}

		ctx.Logger.Info("App authenticator challenge rejected: the challenge expired")

		appWriteJSON(ctx, fasthttp.StatusOK, appConfirmResponse{Result: appResultRejected, ServerTime: now.Unix()})
	case challenge.State != model.AppAuthenticatorChallengeStatePending:
		ctx.Logger.Infof("App authenticator challenge rejected: the challenge is already %s", challenge.State)

		appWriteJSON(ctx, fasthttp.StatusOK, appConfirmResponse{Result: appResultRejected, ServerTime: now.Unix()})
	default:
		appChallengeDecide(ctx, device, challenge, body, decision, now)
	}
}

func appChallengeDecide(ctx *middlewares.AutheliaCtx, device *model.AppAuthenticatorDevice, challenge *model.AppAuthenticatorChallenge, body appJSONObject, decision string, now time.Time) {
	state, reason := appChallengeEvaluate(device, challenge, body, decision)

	state, reason, settled, err := appChallengeSettle(ctx, device, challenge, state, reason, now)
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred settling the app authenticator challenge")

		return
	}

	if !settled {
		ctx.Logger.Info("App authenticator challenge rejected: the challenge was settled concurrently")

		appWriteJSON(ctx, fasthttp.StatusOK, appConfirmResponse{Result: appResultRejected, ServerTime: now.Unix()})

		return
	}

	var result string

	switch state {
	case model.AppAuthenticatorChallengeStateApproved:
		result = appResultApproved

		ctx.Logger.Info("App authenticator challenge approved")
	case model.AppAuthenticatorChallengeStateDenied:
		result = appResultDenied

		ctx.Logger.Info("App authenticator challenge denied")
	default:
		result = appResultRejected

		ctx.Logger.WithField(appFieldReason, reason).Info("App authenticator challenge rejected")
	}

	doMarkAuthenticationAttemptWithRequest(ctx, state == model.AppAuthenticatorChallengeStateApproved,
		regulation.NewBan(regulation.BanTypeNone, challenge.Username, nil), regulation.AuthTypeApp, challenge.RequestedURL.String, "", nil)

	appWriteJSON(ctx, fasthttp.StatusOK, appConfirmResponse{Result: result, ServerTime: now.Unix()})
}

func appChallengeSettle(ctx *middlewares.AutheliaCtx, device *model.AppAuthenticatorDevice, challenge *model.AppAuthenticatorChallenge, state, reason string, now time.Time) (string, string, bool, error) {
	provider := ctx.Providers.StorageProvider

	decided := *challenge
	decided.State = state
	decided.DecidedAt = sql.NullTime{Time: now, Valid: true}
	decided.DecidedDeviceID = sql.NullString{String: device.DeviceID, Valid: true}

	updated, err := provider.UpdateAppAuthenticatorChallengeDecision(ctx, decided)
	if err != nil || updated {
		return state, reason, updated, err
	}

	current, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(ctx, challenge.ConfirmationID)
	if err != nil {
		return state, reason, false, err
	}

	if current.State != model.AppAuthenticatorChallengeStatePending {
		return state, reason, false, nil
	}

	voided := *current
	voided.State = model.AppAuthenticatorChallengeStateVoided
	voided.DecidedAt = decided.DecidedAt
	voided.DecidedDeviceID = decided.DecidedDeviceID

	if updated, err = provider.UpdateAppAuthenticatorChallengeDecision(ctx, voided); err != nil {
		return state, reason, false, err
	}

	return model.AppAuthenticatorChallengeStateVoided, "the binding changed concurrently", updated, nil
}

func appChallengeEvaluate(device *model.AppAuthenticatorDevice, challenge *model.AppAuthenticatorChallenge, body appJSONObject, decision string) (state, reason string) {
	verifier, ok := body.string("verifier")

	switch {
	case !ok || !appauthenticator.IsValidVerifier(verifier):
		return model.AppAuthenticatorChallengeStateVoided, "verifier is malformed"
	case !challenge.BoundDeviceID.Valid:
		return model.AppAuthenticatorChallengeStateVoided, "no binding"
	case challenge.BoundDeviceID.String != device.DeviceID:
		return model.AppAuthenticatorChallengeStateVoided, "bound to another device"
	case subtle.ConstantTimeCompare([]byte(appauthenticator.CodeChallengeS256(verifier)), []byte(challenge.BoundCodeChallenge.String)) != 1:
		return model.AppAuthenticatorChallengeStateVoided, "verifier mismatch"
	case decision == appDecisionDeny:
		return model.AppAuthenticatorChallengeStateDenied, ""
	}

	signature, ok := body.string("signature")
	if !ok || !appauthenticator.IsBase64URL(signature) {
		return model.AppAuthenticatorChallengeStateVoided, "signature is malformed"
	}

	approvalKey, err := appauthenticator.ParsePublicKey(device.ApprovalPublicKey)
	if err != nil {
		return model.AppAuthenticatorChallengeStateVoided, "stored approval key is not usable"
	}

	message := appauthenticator.ApprovalString(device.BaseURL, device.DeviceID, challenge.Challenge, challenge.Number, verifier)

	if !appauthenticator.VerifySignature(approvalKey, message, signature) {
		return model.AppAuthenticatorChallengeStateVoided, "bad signature"
	}

	return model.AppAuthenticatorChallengeStateApproved, ""
}

// AppAuthenticatorDeviceDELETE handles the device request which removes the device.
func AppAuthenticatorDeviceDELETE(ctx *middlewares.AutheliaCtx) {
	if !appDevicePrecheck(ctx) {
		return
	}

	device, ok := appAuthenticate(ctx)
	if !ok {
		return
	}

	if err := ctx.Providers.StorageProvider.DeleteAppAuthenticatorDevice(ctx, device.DeviceID); err != nil {
		appWriteInternal(ctx, err, "Error occurred removing the app authenticator device")

		return
	}

	ctx.Logger.Info("App authenticator device removed by the device")

	ctxLogAppAuthenticatorEvent(ctx, device, false)

	appWriteJSON(ctx, fasthttp.StatusOK, appServerTimeResponse{ServerTime: appServerTime(ctx)})
}

func appAttestationMode(ctx *middlewares.AutheliaCtx) string {
	switch mode := ctx.Configuration.AppAuthenticator.Attestation.Mode; mode {
	case schema.AppAuthenticatorAttestationModeNone, schema.AppAuthenticatorAttestationModePreferred:
		return mode
	default:
		return schema.AppAuthenticatorAttestationModeRequired
	}
}

func appRegistrationCodeUsable(ctx *middlewares.AutheliaCtx, registration *model.AppAuthenticatorRegistration, now time.Time) bool {
	if registration.BegunAt.Valid {
		appWriteError(ctx, fasthttp.StatusConflict, appErrorAlreadyUsed)

		return false
	}

	switch registration.State {
	case model.AppAuthenticatorRegistrationStateIssued:
		if now.After(registration.ExpiresAt) {
			appWriteError(ctx, fasthttp.StatusGone, appErrorExpired)

			return false
		}

		return true
	case model.AppAuthenticatorRegistrationStateCancelled:
		appWriteError(ctx, fasthttp.StatusNotFound, appErrorUnknown)
	case model.AppAuthenticatorRegistrationStateExpired:
		appWriteError(ctx, fasthttp.StatusGone, appErrorExpired)
	default:
		appWriteError(ctx, fasthttp.StatusConflict, appErrorAlreadyUsed)
	}

	return false
}

func appRegistrationSettle(ctx *middlewares.AutheliaCtx, registration *model.AppAuthenticatorRegistration, state string) bool {
	settled := *registration
	settled.State = state

	updated, err := ctx.Providers.StorageProvider.UpdateAppAuthenticatorRegistration(ctx, settled, model.AppAuthenticatorRegistrationStateBegun)
	if err != nil {
		appWriteInternal(ctx, err, "Error occurred updating the app authenticator registration session")

		return false
	}

	if !updated {
		appWriteError(ctx, fasthttp.StatusConflict, appErrorAlreadyUsed)

		return false
	}

	return true
}

const (
	appRotationRotated = iota
	appRotationUnknown
	appRotationLost
)

func appRotatePullID(tx context.Context, provider storage.Provider, device *model.AppAuthenticatorDevice, pullID string) (next string, outcome int, err error) {
	hash := appauthenticator.HashID(pullID)

	var previous []byte

	switch {
	case subtle.ConstantTimeCompare(hash, device.PullIDCurrent) == 1:
		previous = device.PullIDCurrent
	case len(device.PullIDPrevious) != 0 && subtle.ConstantTimeCompare(hash, device.PullIDPrevious) == 1:
		previous = device.PullIDPrevious
	default:
		return "", appRotationUnknown, nil
	}

	if next, err = appauthenticator.NewID(); err != nil {
		return "", appRotationLost, err
	}

	updated, err := provider.UpdateAppAuthenticatorDevicePullIDs(tx, device.ID, device.PullIDCurrent, appauthenticator.HashID(next), previous)

	switch {
	case err != nil:
		return "", appRotationLost, err
	case updated:
		return next, appRotationRotated, nil
	default:
		return "", appRotationLost, nil
	}
}

func appBindPendingChallenge(tx context.Context, provider storage.Provider, device *model.AppAuthenticatorDevice, codeChallenge string, now time.Time) (*model.AppAuthenticatorChallenge, error) {
	challenge, err := provider.LoadAppAuthenticatorChallengePending(tx, device.Username, now)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	challenge.BoundDeviceID = sql.NullString{String: device.DeviceID, Valid: true}
	challenge.BoundCodeChallenge = sql.NullString{String: codeChallenge, Valid: true}

	updated, err := provider.UpdateAppAuthenticatorChallenge(tx, *challenge, model.AppAuthenticatorChallengeStatePending)
	if err != nil {
		return nil, err
	}

	if updated {
		return challenge, nil
	}

	current, err := provider.LoadAppAuthenticatorChallengeByConfirmationID(tx, challenge.ConfirmationID)
	if err != nil {
		return nil, err
	}

	if current.State == model.AppAuthenticatorChallengeStatePending && current.BoundDeviceID == challenge.BoundDeviceID && current.BoundCodeChallenge == challenge.BoundCodeChallenge {
		return current, nil
	}

	return nil, nil
}

func appChallengeDetails(challenge *model.AppAuthenticatorChallenge) map[string]any {
	details := map[string]any{"requested_at": challenge.CreatedAt.Unix()}

	for key, value := range map[string]sql.NullString{
		"portal_url":    challenge.PortalURL,
		"requested_url": challenge.RequestedURL,
		"remote_ip":     challenge.RemoteIP,
		"user_agent":    challenge.UserAgent,
	} {
		if value.Valid {
			details[key] = value.String
		}
	}

	return details
}
