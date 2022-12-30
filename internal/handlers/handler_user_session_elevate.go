package handlers

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// UserSessionElevateGET handles an elevation request for the user session.
func UserSessionElevateGET(ctx *middlewares.AutheliaCtx) {
	var (
		s        session.UserSession
		identity *session.Identity
		publicID uuid.UUID
		err      error
	)

	s = ctx.GetSession()

	now := ctx.Clock.Now()

	if s.Elevations.User != nil && s.Elevations.User.Expires.After(now) {
		ctx.SetJSONError("Already Elevated.")
		ctx.Logger.
			WithField("user", s.Username).
			WithField("expires", s.Elevations.User.Expires.String()).
			Warnf("The user is already elevated.")

		return
	}

	if identity, err = s.Identity(); err != nil {
		ctx.SetJSONError(messageOperationFailed)

		ctx.Logger.
			WithField("user", s.Username).
			WithError(err).
			Errorf("Error occurred determining user identity.")

		return
	}

	if publicID, err = uuid.NewRandom(); err != nil {
		ctx.SetJSONError(messageOperationFailed)

		ctx.Logger.
			WithField("user", s.Username).
			WithError(err).
			Errorf("Error occurred generating UUID.")

		return
	}

	otp := model.NewOneTimePassword(
		publicID,
		identity.Username,
		model.OTPIntentElevateUserSession,
		now,
		now.Add(ctx.Configuration.IdentityValidation.CredentialRegistration.ElevationExpiration),
		ctx.RemoteIP(),
		ctx.Providers.Random.BytesCustom(ctx.Configuration.IdentityValidation.CredentialRegistration.OTPCharacters, []byte(random.CharSetUnambiguousUpper)),
	)

	if otp.Signature, err = ctx.Providers.StorageProvider.SaveOneTimePassword(ctx, otp); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	ctx.Logger.Debugf("Sending an email to user '%s' with email address '%s' to confirm their identity for standard session elevation.",
		s.Username, identity.Address())

	revokeURL := ctx.RootURL()

	query := revokeURL.Query()

	query.Set("pid", base64.URLEncoding.EncodeToString(otp.PublicID[:]))

	revokeURL.Path = path.Join(revokeURL.Path, "/revoke/elevate")
	revokeURL.RawQuery = query.Encode()

	data := templates.EmailOneTimePasswordData{
		Title:           "One-Time Password",
		DisplayName:     s.DisplayName,
		RemoteIP:        ctx.RemoteIP().String(),
		RevokeURL:       revokeURL,
		OneTimePassword: string(otp.Password),
	}

	if err = ctx.Providers.Notifier.Send(ctx, identity.Address(), data.Title, ctx.Providers.Templates.GetOneTimePasswordEmailTemplate(), data); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}

type bodySessionElevate struct {
	OTP string `json:"otp"`
}

// UserSessionElevatePOST validates a user session elevation response.
func UserSessionElevatePOST(ctx *middlewares.AutheliaCtx) {
	var (
		bodyJSON = &bodySessionElevate{}
		s        session.UserSession
		otp      *model.OneTimePassword
		err      error
	)

	if err = ctx.ParseBody(bodyJSON); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	s = ctx.GetSession()

	bodyJSON.OTP = utils.StringStripCharSetUnambiguousUpper(bodyJSON.OTP)

	if otp, err = ctx.Providers.StorageProvider.LoadOneTimePassword(ctx, s.Username, model.OTPIntentElevateUserSession, bodyJSON.OTP); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	if !strings.EqualFold(bodyJSON.OTP, string(otp.Password)) {
		ctx.Error(fmt.Errorf("user supplied session elevation one-time password did not match the generated one-time password"), messageOperationFailed)

		return
	}

	if ctx.Clock.Now().After(otp.ExpiresAt) {
		ctx.Error(fmt.Errorf("user supplied session elevation one-time password expired at %s", otp.ExpiresAt), messageOperationFailed)

		return
	}

	if otp.Consumed.Valid {
		ctx.Error(fmt.Errorf("user supplied session elevation one-time password was already consumed at %s", otp.Consumed.Time), messageOperationFailed)

		return
	}

	if otp.Revoked.Valid {
		ctx.Error(fmt.Errorf("user supplied session elevation one-time password was revoked at %s", otp.Revoked.Time), messageOperationFailed)

		return
	}

	otp.Consumed = sql.NullTime{Time: ctx.Clock.Now(), Valid: true}
	otp.ConsumedIP = model.NewNullIP(ctx.RemoteIP())

	if err = ctx.Providers.StorageProvider.ConsumeOneTimePassword(ctx, otp); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	s.Elevations.User = &session.Elevation{
		ID:       otp.ID,
		RemoteIP: ctx.RemoteIP(),
		Expires:  ctx.Clock.Now().Add(ctx.Configuration.IdentityValidation.CredentialRegistration.ElevationExpiration),
	}

	if err = ctx.SaveSession(s); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}

type bodySessionElevateDELETE struct {
	PublicID []byte `json:"public_id"`
}

// UserSessionElevateDELETE allows revoking of an existing elevate request.
func UserSessionElevateDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		bodyJSON = &bodySessionElevateDELETE{}
		err      error
	)

	if err = ctx.ParseBody(bodyJSON); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	publicID, err := uuid.ParseBytes(bodyJSON.PublicID)
	if err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	if err = ctx.Providers.StorageProvider.RevokeOneTimePassword(ctx, publicID, model.NewIP(ctx.RemoteIP())); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}

type bodySessionElevationGET struct {
	Elevated bool       `json:"elevated"`
	Expires  *time.Time `json:"expires,omitempty"`
}

// UserSessionElevationGET retrieves current elevation status.
func UserSessionElevationGET(ctx *middlewares.AutheliaCtx) {
	var (
		s   session.UserSession
		err error
	)

	s = ctx.GetSession()

	resp := &bodySessionElevationGET{}

	if s.Elevations.User == nil || s.Elevations.User.Expires.Before(ctx.Clock.Now()) {
		if err = ctx.SetJSONBody(resp); err != nil {
			ctx.Error(err, messageOperationFailed)

			return
		}

		if s.Elevations.User != nil {
			s.Elevations.User = nil

			if err = ctx.SaveSession(s); err != nil {
				ctx.Error(err, messageOperationFailed)

				return
			}
		}

		return
	}

	resp.Elevated = true
	resp.Expires = &s.Elevations.User.Expires

	if err = ctx.SetJSONBody(resp); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}
}

// UserSessionElevationDELETE allows a user to de-elevate.
func UserSessionElevationDELETE(ctx *middlewares.AutheliaCtx) {
	var (
		s   session.UserSession
		err error
	)

	s = ctx.GetSession()

	if s.Elevations.User == nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	s.Elevations.User = nil

	if err = ctx.SaveSession(s); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}
