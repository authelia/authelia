// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"errors"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/appauthenticator"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

const (
	appPortalStateWaiting       = "waiting"
	appPortalStateConfirmNumber = "confirm_number"
	appPortalStateRegistered    = "registered"
	appPortalStateFailed        = "failed"
	appPortalStateExpired       = "expired"
)

type appRegistrationPUTResponse struct {
	QR        string `json:"qr"`
	ExpiresAt int64  `json:"expires_at"`
}

type appRegistrationDevice struct {
	Platform string `json:"platform"`
	Model    string `json:"model"`
	Name     string `json:"name"`
}

type appRegistrationGETResponse struct {
	State       string                 `json:"state"`
	Number      string                 `json:"number,omitempty"`
	NonceLength int                    `json:"nonce_length,omitempty"`
	Device      *appRegistrationDevice `json:"device,omitempty"`
	ID          int                    `json:"id,omitempty"`
	ExpiresAt   int64                  `json:"expires_at,omitempty"`
}

// AppAuthenticatorRegistrationPUT creates an Authelia app registration code for the session and returns its QR text.
func AppAuthenticatorRegistrationPUT(ctx *middlewares.AutheliaCtx) {
	userSession, ok := appPortalSession(ctx, "creating the app authenticator registration", fasthttp.StatusForbidden, messageOperationFailed)
	if !ok {
		return
	}

	log := ctx.Logger.WithField(logging.FieldUsername, userSession.Username)

	if !appRegistrationAvailable(ctx) {
		log.Error("Error occurred creating the app authenticator registration: registration is unavailable as the attestation mode is 'required' and no trusted app identities are configured")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	baseURL, err := appBaseURL(ctx)
	if err != nil {
		log.WithError(err).Error("Error occurred creating the app authenticator registration: the portal url of the session cookie domain is not usable as an app base url, check the session cookie 'authelia_url' configuration")

		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	binding, ok := appPortalBindingEnsure(ctx, &userSession, "creating the app authenticator registration", fasthttp.StatusInternalServerError, messageOperationFailed)
	if !ok {
		return
	}

	if err = ctx.Providers.StorageProvider.CancelAppAuthenticatorRegistrations(ctx, binding); err != nil {
		appPortalError(ctx, err, "Error occurred cancelling the previous app authenticator registrations", fasthttp.StatusInternalServerError, messageOperationFailed)

		return
	}

	random, err := appauthenticator.NewID()
	if err != nil {
		appPortalError(ctx, err, "Error occurred generating the app authenticator registration code", fasthttp.StatusInternalServerError, messageOperationFailed)

		return
	}

	nonceLength := appNonceLength(ctx)

	number, err := appauthenticator.NewNumber(nonceLength)
	if err != nil {
		appPortalError(ctx, err, "Error occurred generating the app authenticator registration number", fasthttp.StatusInternalServerError, messageOperationFailed)

		return
	}

	now := ctx.GetClock().Now()

	registration := model.AppAuthenticatorRegistration{
		CreatedAt:     now,
		ExpiresAt:     now.Add(appRegistrationLifespan(ctx)),
		Code:          appauthenticator.HashID(random),
		Username:      userSession.Username,
		PortalBinding: binding,
		BaseURL:       baseURL,
		Number:        number,
		NonceLength:   nonceLength,
		State:         model.AppAuthenticatorRegistrationStateIssued,
	}

	if err = ctx.Providers.StorageProvider.SaveAppAuthenticatorRegistration(ctx, registration); err != nil {
		appPortalError(ctx, err, "Error occurred saving the app authenticator registration", fasthttp.StatusInternalServerError, messageOperationFailed)

		return
	}

	log.Info("App authenticator registration code issued")

	appPortalJSON(ctx, appRegistrationPUTResponse{QR: appauthenticator.QRText(baseURL, random), ExpiresAt: registration.ExpiresAt.Unix()}, messageOperationFailed)
}

// AppAuthenticatorRegistrationGET returns the state of the Authelia app registration of the session.
func AppAuthenticatorRegistrationGET(ctx *middlewares.AutheliaCtx) {
	userSession, ok := appPortalSession(ctx, "retrieving the app authenticator registration", fasthttp.StatusForbidden, messageOperationFailed)
	if !ok {
		return
	}

	binding := appPortalBinding(&userSession)
	if binding == nil {
		appPortalNotFound(ctx)

		return
	}

	registration, err := ctx.Providers.StorageProvider.LoadAppAuthenticatorRegistrationByPortalBinding(ctx, binding)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		appPortalNotFound(ctx)

		return
	case err != nil:
		appPortalError(ctx, err, "Error occurred loading the app authenticator registration", fasthttp.StatusInternalServerError, messageOperationFailed)

		return
	case registration.Username != userSession.Username, registration.State == model.AppAuthenticatorRegistrationStateCancelled:
		appPortalNotFound(ctx)

		return
	}

	response := appRegistrationState(registration, ctx.GetClock().Now())

	if response.State == appPortalStateRegistered {
		device, err := ctx.Providers.StorageProvider.LoadAppAuthenticatorDeviceByDeviceID(ctx, registration.DeviceID.String)

		switch {
		case err == nil:
			response.ID = device.ID
		case !errors.Is(err, sql.ErrNoRows):
			appPortalError(ctx, err, "Error occurred loading the registered app authenticator device", fasthttp.StatusInternalServerError, messageOperationFailed)

			return
		}
	}

	appPortalJSON(ctx, response, messageOperationFailed)
}

// AppAuthenticatorRegistrationDELETE cancels the unfinished Authelia app registrations of the session.
func AppAuthenticatorRegistrationDELETE(ctx *middlewares.AutheliaCtx) {
	userSession, ok := appPortalSession(ctx, "cancelling the app authenticator registration", fasthttp.StatusForbidden, messageOperationFailed)
	if !ok {
		return
	}

	if binding := appPortalBinding(&userSession); binding != nil {
		if err := ctx.Providers.StorageProvider.CancelAppAuthenticatorRegistrations(ctx, binding); err != nil {
			appPortalError(ctx, err, "Error occurred cancelling the app authenticator registration", fasthttp.StatusInternalServerError, messageOperationFailed)

			return
		}
	}

	ctx.ReplyOK()
}

func appRegistrationState(registration *model.AppAuthenticatorRegistration, now time.Time) appRegistrationGETResponse {
	switch registration.State {
	case model.AppAuthenticatorRegistrationStateIssued:
		if now.After(registration.ExpiresAt) {
			return appRegistrationGETResponse{State: appPortalStateExpired}
		}

		return appRegistrationGETResponse{State: appPortalStateWaiting, ExpiresAt: registration.ExpiresAt.Unix()}
	case model.AppAuthenticatorRegistrationStateBegun:
		if !registration.SessionExpiresAt.Valid || now.After(registration.SessionExpiresAt.Time) {
			return appRegistrationGETResponse{State: appPortalStateExpired}
		}

		return appRegistrationGETResponse{
			State:       appPortalStateConfirmNumber,
			Number:      registration.Number,
			NonceLength: registration.NonceLength,
			Device: &appRegistrationDevice{
				Platform: registration.DevicePlatform.String,
				Model:    registration.DeviceModel.String,
				Name:     registration.DeviceName.String,
			},
			ExpiresAt: registration.SessionExpiresAt.Time.Unix(),
		}
	case model.AppAuthenticatorRegistrationStateCompleted:
		return appRegistrationGETResponse{State: appPortalStateRegistered}
	case model.AppAuthenticatorRegistrationStateExpired:
		return appRegistrationGETResponse{State: appPortalStateExpired}
	default:
		return appRegistrationGETResponse{State: appPortalStateFailed}
	}
}

func appRegistrationAvailable(ctx *middlewares.AutheliaCtx) bool {
	config := ctx.Configuration.AppAuthenticator

	if config.Disable {
		return false
	}

	return appAttestationMode(ctx) != schema.AppAuthenticatorAttestationModeRequired || config.Attestation.HasIdentities()
}

func appRegistrationLifespan(ctx *middlewares.AutheliaCtx) time.Duration {
	if lifespan := ctx.Configuration.AppAuthenticator.RegistrationTokenLifespan; lifespan > 0 {
		return lifespan
	}

	return schema.DefaultAppAuthenticatorConfiguration.RegistrationTokenLifespan
}

func appChallengeLifespan(ctx *middlewares.AutheliaCtx) time.Duration {
	if lifespan := ctx.Configuration.AppAuthenticator.ChallengeLifespan; lifespan > 0 {
		return lifespan
	}

	return schema.DefaultAppAuthenticatorConfiguration.ChallengeLifespan
}

func appNonceLength(ctx *middlewares.AutheliaCtx) int {
	if n := ctx.Configuration.AppAuthenticator.NonceLength; n != 0 {
		return n
	}

	return schema.DefaultAppAuthenticatorConfiguration.NonceLength
}

func appPortalURL(ctx *middlewares.AutheliaCtx) (*url.URL, error) {
	if autheliaURL := ctx.GetSessionConfig().AutheliaURL; autheliaURL != nil {
		return autheliaURL, nil
	}

	return ctx.IssuerURL()
}

func appBaseURL(ctx *middlewares.AutheliaCtx) (string, error) {
	portalURL, err := appPortalURL(ctx)
	if err != nil {
		return "", err
	}

	return appauthenticator.NewBaseURL(portalURL)
}

func appPortalSession(ctx *middlewares.AutheliaCtx, action string, status int, message middlewares.ErrorMessage) (userSession session.UserSession, ok bool) {
	var err error

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred %s: %s", action, errStrUserSessionData)

		ctx.SetStatusCode(status)
		ctx.SetJSONError(message)

		return userSession, false
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Errorf("Error occurred %s", action)

		ctx.SetStatusCode(status)
		ctx.SetJSONError(message)

		return userSession, false
	}

	return userSession, true
}

func appPortalBinding(userSession *session.UserSession) []byte {
	if userSession.AppAuthenticator == nil || userSession.AppAuthenticator.Binding == "" {
		return nil
	}

	return appauthenticator.HashID(userSession.AppAuthenticator.Binding)
}

func appPortalBindingEnsure(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, action string, status int, message middlewares.ErrorMessage) ([]byte, bool) {
	if binding := appPortalBinding(userSession); binding != nil {
		return binding, true
	}

	token, err := appauthenticator.NewID()
	if err != nil {
		appPortalError(ctx, err, "Error occurred "+action+": error generating the portal binding", status, message)

		return nil, false
	}

	userSession.AppAuthenticator = &session.AppAuthenticator{Binding: token}

	if err = ctx.SaveSession(userSession); err != nil {
		appPortalError(ctx, err, "Error occurred "+action+": "+errStrUserSessionDataSave, status, message)

		return nil, false
	}

	return appauthenticator.HashID(token), true
}

func appPortalError(ctx *middlewares.AutheliaCtx, err error, message string, status int, response middlewares.ErrorMessage) {
	ctx.Logger.WithError(err).Error(message)

	ctx.SetStatusCode(status)
	ctx.SetJSONError(response)
}

func appPortalNotFound(ctx *middlewares.AutheliaCtx) {
	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.SetJSONError(messageOperationFailed)
}

func appPortalJSON(ctx *middlewares.AutheliaCtx, body any, message middlewares.ErrorMessage) {
	if err := ctx.SetJSONBody(body); err != nil {
		appPortalError(ctx, err, "Error occurred encoding the app authenticator response", fasthttp.StatusInternalServerError, message)
	}
}
