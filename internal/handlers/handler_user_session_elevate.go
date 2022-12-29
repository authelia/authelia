package handlers

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// UserSessionElevationGET handles an elevation request for the user session.
func UserSessionElevationGET(ctx *middlewares.AutheliaCtx) {
	var (
		s        session.UserSession
		identity *session.Identity
		err      error
	)

	s = ctx.GetSession()

	if identity, err = s.Identity(); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	otp := model.NewOneTimePassword(
		identity.Username,
		model.OTPIntentElevateUserSession,
		ctx.Clock.Now(),
		ctx.Clock.Now().Add(ctx.Configuration.IdentityValidation.CredentialRegistration.ElevationExpiration),
		ctx.RemoteIP(),
		ctx.Providers.Random.BytesCustom(ctx.Configuration.IdentityValidation.CredentialRegistration.OTPCharacters, []byte(random.CharSetUnambiguousUpper)),
	)

	if otp.Signature, err = ctx.Providers.StorageProvider.SaveOneTimePassword(ctx, otp); err != nil {
		ctx.Error(err, messageOperationFailed)

		return
	}

	ctx.Logger.Debugf("Sending an email to user '%s' with email address '%s' to confirm their identity for standard session elevation.",
		s.Username, identity.Address())

	data := templates.EmailOneTimePasswordData{
		Title:           "One-Time Password",
		DisplayName:     s.DisplayName,
		RemoteIP:        ctx.RemoteIP().String(),
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

// UserSessionElevationPOST validates a user session elevation response.
func UserSessionElevationPOST(ctx *middlewares.AutheliaCtx) {
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
