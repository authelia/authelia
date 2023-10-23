package handlers

import (
	"crypto/subtle"
	"encoding/base64"
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
func UserSessionElevationGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)

	response := &bodyGETUserSessionElevate{}

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Failed to get user session from session provider during user session elevation lookup.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

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

	if userSession.Elevations.User != nil && !userSession.IsAnonymous() {
		var deleted bool

		response.Elevated = true
		response.Expires = int(userSession.Elevations.User.Expires.Sub(ctx.Clock.Now()).Seconds())

		if userSession.Elevations.User.Expires.Before(ctx.Clock.Now()) {
			ctx.Logger.WithFields(map[string]any{"username": userSession.Username, "expired": userSession.Elevations.User.Expires.Unix()}).
				Info("The user session elevation has already expired so it has been deleted.")

			response.Elevated, deleted = false, true
		}

		if !userSession.Elevations.User.RemoteIP.Equal(ctx.RemoteIP()) {
			ctx.Logger.WithFields(map[string]any{"username": userSession.Username, "elevation_ip": userSession.Elevations.User.RemoteIP.String()}).
				Warn("The user session elevation was created from a different remote IP so it has been destroyed.")

			response.Elevated, deleted = false, true
		}

		if deleted {
			userSession.Elevations.User = nil

			if err = ctx.SaveSession(userSession); err != nil {
				ctx.Logger.WithError(err).Error("Failed to save user session.")

				ctx.SetStatusCode(fasthttp.StatusForbidden)
				ctx.SetJSONError(messageOperationFailed)

				return
			}
		}
	}

	if response.Elevated && response.CanSkipSecondFactor {
		response.CanSkipSecondFactor = false
	}

	if err = ctx.ReplyJSON(middlewares.OKResponse{Status: "OK", Data: response}, fasthttp.StatusOK); err != nil {
		ctx.Logger.WithError(err).Error("Failed to write JSON response in elevation lookup.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

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
		ctx.Logger.WithError(err).Error("Failed to get user session from session provider during user session elevation create.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		otp *model.OneTimeCode
	)

	if otp, err = model.NewOneTimeCode(ctx, userSession.Username, ctx.Configuration.IdentityValidation.ElevatedSession.Characters, ctx.Configuration.IdentityValidation.ElevatedSession.Expiration); err != nil {
		ctx.Logger.WithError(err).Error("Failed to generate elevation One-Time Code during user session elevation create.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var signature string

	if signature, err = ctx.Providers.StorageProvider.SaveOneTimeCode(ctx, *otp); err != nil {
		ctx.Logger.WithError(err).Error("Failed to save elevation One-Time Code during user session elevation create.")

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
		Title:       "Confirm your identity",
		LinkURL:     linkURL.String(),
		LinkText:    "Revoke",
		DisplayName: identity.DisplayName,
		RemoteIP:    ctx.RemoteIP().String(),
		OneTimeCode: string(otp.Code),
	}

	ctx.Logger.WithFields(map[string]any{"signature": signature, "id": otp.PublicID.String(), "username": identity.Username}).
		Debug("Sending an email to user to confirm identity for session elevation.")

	if err = ctx.Providers.Notifier.Send(ctx, identity.Address(), data.Title, ctx.Providers.Templates.GetIdentityVerificationOTCEmailTemplate(), data); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.SetJSONBody(&bodyPOSTUserSessionElevate{
		DeleteID: deleteID,
	}); err != nil {
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
		otp         *model.OneTimeCode
		err         error
	)

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Logger.WithError(err).Error("Failed to parse user session elevation body.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if otp, err = ctx.Providers.StorageProvider.LoadOneTimeCode(ctx, userSession.Username, model.OTCIntentUserSessionElevation, bodyJSON.OneTimeCode); err != nil {
		ctx.Logger.WithError(err).WithFields(map[string]any{"username": userSession.Username}).
			Error("Error occurred retrieving user session elevation One-Time Code information from the database. This error should only occur due to database related issues.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	} else if otp == nil {
		ctx.Logger.WithFields(map[string]any{"username": userSession.Username}).
			Error("Error occurred retrieving user session elevation One-Time Code information from the database. The code did not match any recorded codes.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if otp.ExpiresAt.Before(ctx.Clock.Now()) {
		ctx.Logger.Error("Failed to consume the One-Time Code during user session elevation as it's expired.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if otp.RevokedAt.Valid {
		ctx.Logger.Error("Failed to consume the One-Time Code during user session elevation as it's revoked.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if otp.ConsumedAt.Valid {
		ctx.Logger.Error("Failed to consume the One-Time Code during user session elevation as it's already consumed.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if subtle.ConstantTimeCompare(otp.Code, []byte(strings.ToUpper(bodyJSON.OneTimeCode))) != 1 {
		ctx.Logger.Error("Failed to consume the One-Time Code during user session elevation as it's already consumed.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	otp.Consume(ctx)

	if err = ctx.Providers.StorageProvider.ConsumeOneTimeCode(ctx, otp); err != nil {
		ctx.Logger.WithError(err).Error("Failed to consume the One-Time Code during user session elevation due to a database error.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	userSession.Elevations.User = &session.Elevation{
		ID:       otp.ID,
		RemoteIP: ctx.RemoteIP(),
		Expires:  ctx.Clock.Now().Add(ctx.Configuration.IdentityValidation.ElevatedSession.ElevationExpiration),
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.WithError(err).Error("Failed to save the user session elevation to the session.")

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
		id  uuid.UUID
		otp *model.OneTimeCode
		err error
	)

	if _, err = base64.RawURLEncoding.Decode(decoded, []byte(value)); err != nil {
		ctx.Logger.WithError(err).Error("Failed to base64 decode elevation identifier during elevation revocation.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if id, err = uuid.FromBytes(decoded); err != nil {
		ctx.Logger.WithError(err).Error("Failed to parse decoded elevation identifier during elevation revocation.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if otp, err = ctx.Providers.StorageProvider.LoadOneTimeCodeByPublicID(ctx, id); err != nil {
		ctx.Logger.WithError(err).Error("Failed to load the elevation One-Time Code row from the database during elevation revocation.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if otp.RevokedAt.Valid {
		ctx.Logger.Error("Failed to revoke the One-Time Code during elevation revocation as it's already revoked.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if otp.ConsumedAt.Valid {
		ctx.Logger.Error("Failed to revoke the One-Time Code during elevation revocation as it's consumed.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if otp.Intent != model.OTCIntentUserSessionElevation {
		ctx.Logger.Error("Failed to revoke the One-Time Code during elevation revocation as it doesn't have the expected intent.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.RevokeOneTimeCode(ctx, id, model.NewIP(ctx.RemoteIP())); err != nil {
		ctx.Logger.WithError(err).Error("Failed to save the revocation information to the database during elevation revocation.")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}
