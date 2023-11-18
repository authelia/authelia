package handlers

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

// UserSessionElevationGET returns the session elevation status.
//
//nolint:gocyclo
func UserSessionElevationGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	response := &bodyGETUserSessionElevate{}

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session elevation state: error occurred retrieving the user session data")

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(fmt.Errorf("user is anonymous")).Error("Error occurred retrieving user session elevation state")

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	switch {
	case userSession.AuthenticationLevel >= authentication.TwoFactor:
		if ctx.Configuration.IdentityValidation.ElevatedSession.SkipSecondFactor {
			response.SkipSecondFactor = true
		}
	case userSession.AuthenticationLevel == authentication.OneFactor:
		var (
			has  bool
			info model.UserInfo
		)

		info, err = ctx.Providers.StorageProvider.LoadUserInfo(ctx, userSession.Username)

		has = info.HasTOTP || info.HasWebAuthn || info.HasDuo

		if ctx.Configuration.IdentityValidation.ElevatedSession.RequireSecondFactor {
			if err != nil || has {
				response.RequireSecondFactor = true
			}
		} else if ctx.Configuration.IdentityValidation.ElevatedSession.SkipSecondFactor && has {
			response.CanSkipSecondFactor = true
		}
	}

	if userSession.Elevations.User != nil {
		var deleted bool

		response.Elevated = true
		response.Expires = int(userSession.Elevations.User.Expires.Sub(ctx.Clock.Now()).Seconds())

		if userSession.Elevations.User.Expires.Before(ctx.Clock.Now()) {
			ctx.Logger.WithFields(map[string]any{"username": userSession.Username, "expired": userSession.Elevations.User.Expires.Unix()}).
				Info("The user session elevation has already expired so it has been destroyed")

			response.Elevated, deleted = false, true
		}

		if !userSession.Elevations.User.RemoteIP.Equal(ctx.RemoteIP()) {
			ctx.Logger.WithFields(map[string]any{"username": userSession.Username, "elevation_ip": userSession.Elevations.User.RemoteIP.String()}).
				Warn("The user session elevation was created from a different remote IP so it has been destroyed")

			response.Expires, response.Elevated, deleted = 0, false, true
		}

		if deleted {
			userSession.Elevations.User = nil

			if err = ctx.SaveSession(userSession); err != nil {
				ctx.Logger.WithError(err).Error("Error occurred retrieving the user session elevation state: error occurred saving the user session data")

				ctx.SetJSONError(messageOperationFailed)
				ctx.SetStatusCode(fasthttp.StatusForbidden)

				return
			}
		}
	}

	if err = ctx.ReplyJSON(middlewares.OKResponse{Status: "OK", Data: response}, fasthttp.StatusOK); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving the user session elevation state: error occurred writing the response body")

		ctx.SetJSONError(messageOperationFailed)

		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}
}

// UserSessionElevationPOST creates a new elevation session to be validated.
func UserSessionElevationPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred creating user session elevation One-Time Code challenge: error occurred retrieving the user session data")

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(fmt.Errorf("user is anonymous")).Error("Error occurred creating user session elevation One-Time Code challenge")

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	var (
		otp *model.OneTimeCode
	)

	if otp, err = model.NewOneTimeCode(ctx, userSession.Username, ctx.Configuration.IdentityValidation.ElevatedSession.Characters, ctx.Configuration.IdentityValidation.ElevatedSession.Expiration); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred creating user session elevation One-Time Code challenge for user '%s': error occurred generating the One-Time Code challenge", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var signature string

	if signature, err = ctx.Providers.StorageProvider.SaveOneTimeCode(ctx, *otp); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred creating user session elevation One-Time Code challenge for user '%s': error occurred saving the challenge to storage", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	deleteID := base64.RawURLEncoding.EncodeToString(otp.PublicID[:])

	linkURL := ctx.RootURL()

	query := linkURL.Query()

	query.Set("id", deleteID)

	linkURL.Path = path.Join(linkURL.Path, "/revoke/one-time-code")
	linkURL.RawQuery = query.Encode()

	identity := userSession.Identity()

	data := templates.EmailIdentityVerificationOTCValues{
		Title:              "Confirm your identity",
		RevocationLinkURL:  linkURL.String(),
		RevocationLinkText: "Revoke",
		DisplayName:        identity.DisplayName,
		RemoteIP:           ctx.RemoteIP().String(),
		OneTimeCode:        string(otp.Code),
	}

	ctx.Logger.WithFields(map[string]any{"signature": signature, "id": otp.PublicID.String(), "username": identity.Username}).
		Debug("Sending an email to user to confirm identity for session elevation")

	if err = ctx.Providers.Notifier.Send(ctx, identity.Address(), data.Title, ctx.Providers.Templates.GetIdentityVerificationOTCEmailTemplate(), data); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred creating user session elevation One-Time Code challenge for user '%s': error occurred sending the user the notification", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(&bodyPOSTUserSessionElevate{
		DeleteID: deleteID,
	}); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred creating user session elevation One-Time Code challenge: error occurred writing the response body")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}
}

// UserSessionElevationPUT validates an elevation session and puts it into effect.
func UserSessionElevationPUT(ctx *middlewares.AutheliaCtx) {
	bodyJSON := bodyPUTUserSessionElevate{}

	var (
		userSession session.UserSession
		code        *model.OneTimeCode
		err         error
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred validating user session elevation One-Time Code challenge: error occurred retrieving the user session data")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(fmt.Errorf("user is anonymous")).Error("Error occurred validating user session elevation One-Time Code challenge")

		ctx.SetJSONError(messageOperationFailed)
		ctx.SetStatusCode(fasthttp.StatusForbidden)

		return
	}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred validating user session elevation One-Time Code challenge: error parsing the request body")

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	bodyJSON.OneTimeCode = strings.ToUpper(bodyJSON.OneTimeCode)

	if code, err = ctx.Providers.StorageProvider.LoadOneTimeCode(ctx, userSession.Username, model.OTCIntentUserSessionElevation, bodyJSON.OneTimeCode); err != nil {
		ctx.Logger.WithError(err).
			Errorf("Error occurred validating user session elevation One-Time Code challenge for user '%s': error occurred retrieving the code challenge from storage", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	} else if code == nil {
		ctx.Logger.WithError(fmt.Errorf("the code didn't match any recorded code challenges")).
			Errorf("Error occurred validating user session elevation One-Time Code challenge for user '%s': error occurred retrieving the code challenge from storage", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if code.ExpiresAt.Before(ctx.Clock.Now()) {
		ctx.Logger.WithError(fmt.Errorf("the code challenge has expired")).Errorf("Error occurred validating user session elevation One-Time Code challenge for user '%s'", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if code.RevokedAt.Valid {
		ctx.Logger.WithError(fmt.Errorf("the code challenge has been revoked")).Errorf("Error occurred validating user session elevation One-Time Code challenge for user '%s'", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if code.ConsumedAt.Valid {
		ctx.Logger.WithError(fmt.Errorf("the code challenge has already been consumed")).Errorf("Error occurred validating user session elevation One-Time Code challenge for user '%s'", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if code.Intent != model.OTCIntentUserSessionElevation {
		ctx.Logger.WithError(fmt.Errorf("the code challenge has the '%s' intent but the '%s' intent is required", code.Intent, model.OTCIntentUserSessionElevation)).Errorf("Error occurred validating user session elevation One-Time Code challenge for user '%s'", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if subtle.ConstantTimeCompare(code.Code, []byte(bodyJSON.OneTimeCode)) != 1 {
		ctx.Logger.WithError(fmt.Errorf("the code does not match the code stored in the challenge")).Errorf("Error occurred validating user session elevation One-Time Code challenge for user '%s'", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	code.Consume(ctx)

	if err = ctx.Providers.StorageProvider.ConsumeOneTimeCode(ctx, code); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred validating user session elevation One-Time Code challenge for user '%s': error occurred saving the consumption of the code to storage", userSession.Username)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	userSession.Elevations.User = &session.Elevation{
		ID:       code.ID,
		RemoteIP: ctx.RemoteIP(),
		Expires:  ctx.Clock.Now().Add(ctx.Configuration.IdentityValidation.ElevatedSession.ElevationExpiration),
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred validating user session elevation One-Time Code challenge: error occurred saving the user session data")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}

// UserSessionElevateDELETE marks a pending elevation session as revoked.
func UserSessionElevateDELETE(ctx *middlewares.AutheliaCtx) {
	value := ctx.UserValue("id").(string)

	decoded := make([]byte, base64.RawURLEncoding.DecodedLen(len(value)))

	var (
		id   uuid.UUID
		code *model.OneTimeCode
		err  error
	)

	if _, err = base64.RawURLEncoding.Decode(decoded, []byte(value)); err != nil {
		ctx.Logger.WithError(err).
			Error("Error occurred revoking user session elevation One-Time Code challenge: error occurred decoding the identifier")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if id, err = uuid.FromBytes(decoded); err != nil {
		ctx.Logger.WithError(err).
			Error("Error occurred revoking user session elevation One-Time Code challenge: error occurred parsing the identifier")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if code, err = ctx.Providers.StorageProvider.LoadOneTimeCodeByPublicID(ctx, id); err != nil {
		ctx.Logger.WithError(err).
			Error("Error occurred revoking user session elevation One-Time Code challenge: error occurred retrieving the code challenge from storage")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if code.RevokedAt.Valid {
		ctx.Logger.WithError(fmt.Errorf("the code challenge has already been revoked")).Errorf("Error occurred validating user session elevation One-Time Code challenge")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if code.ConsumedAt.Valid {
		ctx.Logger.WithError(fmt.Errorf("the code challenge has already been consumed")).Errorf("Error occurred validating user session elevation One-Time Code challenge")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if code.Intent != model.OTCIntentUserSessionElevation {
		ctx.Logger.WithError(fmt.Errorf("the code challenge has the '%s' intent but the '%s' intent is required", code.Intent, model.OTCIntentUserSessionElevation)).Errorf("Error occurred revoking user session elevation One-Time Code challenge for user")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.RevokeOneTimeCode(ctx, id, model.NewIP(ctx.RemoteIP())); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred revoking user session elevation One-Time Code challenge: error occurred saving the revocation of the code being saved to storage")

		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}
