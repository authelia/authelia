package handlers

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// UserInfoPOST handles setting up info for users if necessary when they login.
func UserInfoPOST(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		userInfo    model.UserInfo
		err         error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if _, err = ctx.Providers.StorageProvider.LoadPreferred2FAMethod(ctx, userSession.Username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err = ctx.Providers.StorageProvider.SavePreferred2FAMethod(ctx, userSession.Username, ""); err != nil {
				ctx.GetLogger().WithError(err).Error("Error occurred saving the user preferred second factor method while loading the user information")
				ctx.SetJSONError(messageOperationFailed)
			}
		} else {
			ctx.GetLogger().WithError(err).Error("Error occurred looking up the user preferred second factor method while loading the user information")
			ctx.SetJSONError(messageOperationFailed)
		}
	}

	if userInfo, err = ctx.Providers.StorageProvider.LoadUserInfo(ctx, userSession.Username); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred loading the user information")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	var (
		changed bool
	)

	if changed = userInfo.SetDefaultPreferred2FAMethod(ctx.AvailableSecondFactorMethods(), ctx.Configuration.Default2FAMethod); changed {
		if err = ctx.Providers.StorageProvider.SavePreferred2FAMethod(ctx, userSession.Username, userInfo.Method); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred saving the user preferred second factor method")
			ctx.SetJSONError(messageOperationFailed)

			return
		}
	}

	if ctx.Configuration.TOTP.Disable {
		userInfo.HasTOTP = false
	}

	if ctx.Configuration.WebAuthn.Disable {
		userInfo.HasWebAuthn = false
	}

	if ctx.Configuration.DuoAPI.Disable {
		userInfo.HasDuo = false
	}

	userInfo.DisplayName = userSession.DisplayName

	err = ctx.SetJSONBody(userInfo)
	if err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred trying to set user info response in body")
	}
}

// UserInfoGET get the info related to the user identified by the session.
func UserInfoGET(ctx *middlewares.AutheliaCtx) {
	var (
		userSession session.UserSession
		err         error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	userInfo, err := ctx.Providers.StorageProvider.LoadUserInfo(ctx, userSession.Username)
	if err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred loading the user information")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	userInfo.DisplayName = userSession.DisplayName

	// it should be noted that UserInfo only contains info from the database and NOT any info from the authn_backend (email/groups).
	for _, email := range userSession.Emails {
		userInfo.Emails = append(userInfo.Emails, redactEmail(email))
	}

	err = ctx.SetJSONBody(userInfo)
	if err != nil {
		ctx.Logger.Errorf("Unable to set user info response in body: %+v", err)
	}
}

// MethodPreferencePOST update the user preferences regarding 2FA method.
func MethodPreferencePOST(ctx *middlewares.AutheliaCtx) {
	var (
		bodyJSON bodyPreferred2FAMethod

		userSession session.UserSession
		err         error
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred retrieving user session")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred parsing the second factor method preference request body")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !utils.IsStringInSlice(bodyJSON.Method, ctx.AvailableSecondFactorMethods()) {
		ctx.GetLogger().Errorf("Error occurred setting the second factor method preference as the method '%s' is unknown or unavailable, it should be one of %s", bodyJSON.Method, strings.Join(ctx.AvailableSecondFactorMethods(), ", "))
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.Logger.Debugf("Save new preferred 2FA method of user %s to %s", userSession.Username, bodyJSON.Method)

	if err = ctx.Providers.StorageProvider.SavePreferred2FAMethod(ctx, userSession.Username, bodyJSON.Method); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred saving the new preferred second factor method")
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	ctx.ReplyOK()
}
