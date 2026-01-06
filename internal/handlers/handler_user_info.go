package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// UserInfoPOST handles setting up info for users if necessary when they log in.
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
				ctx.Error(fmt.Errorf("unable to load user information: error occurred trying to save the users preferred 2FA method: %w", err), messageOperationFailed)
			}
		} else {
			ctx.Error(fmt.Errorf("unable to load user information: error occurred trying to lookup the users preferred 2FA method: %w", err), messageOperationFailed)
		}
	}

	if userInfo, err = ctx.Providers.StorageProvider.LoadUserInfo(ctx, userSession.Username); err != nil {
		ctx.Error(fmt.Errorf("unable to load user information: %w", err), messageOperationFailed)
		return
	}

	var (
		changed bool
	)

	if changed = userInfo.SetDefaultPreferred2FAMethod(ctx.AvailableSecondFactorMethods(), ctx.Configuration.Default2FAMethod); changed {
		if err = ctx.Providers.StorageProvider.SavePreferred2FAMethod(ctx, userSession.Username, userInfo.Method); err != nil {
			ctx.Error(fmt.Errorf("unable to save user two factor method: %w", err), messageOperationFailed)
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
	userInfo.Emails = userSession.Emails

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
		ctx.Error(fmt.Errorf("unable to load user information: %w", err), messageOperationFailed)
		return
	}

	userInfo.Username = userSession.Username
	userInfo.DisplayName = userSession.DisplayName
	userInfo.Groups = userSession.Groups

	// it should be noted that UserInfo only contains info from the database and session and NOT any info from the authn_backend (email/groups).
	for _, email := range userSession.Emails {
		userInfo.Emails = append(userInfo.Emails, redactEmail(email))
	}

	err = ctx.SetJSONBody(userInfo)
	if err != nil {
		ctx.Logger.Errorf("Unable to set user info response in body: %+v", err)
	}
}

func GetUserGET(ctx *middlewares.AutheliaCtx) {
	usernameRaw := ctx.UserValue("username")
	if usernameRaw == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageUsernameRequired)

		return
	}

	username := usernameRaw.(string)

	var (
		err         error
		userInfo    model.UserInfo
		userDetails *authentication.UserDetailsExtended
	)
	if userDetails, err = ctx.Providers.UserProvider.GetDetailsExtended(username); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving users")
		return
	}

	userInfo, err = ctx.Providers.StorageProvider.LoadUserMetadataByUsername(ctx, username)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			ctx.Error(fmt.Errorf("unable to load user attributes: %w", err), messageOperationFailed)
			return
		}
	}

	userDetails = MergeUserDetailsWithInfo(userDetails, userInfo)

	err = ctx.SetJSONBody(userDetails)
	if err != nil {
		ctx.Logger.Errorf("Unable to set user info response in body: %+v", err)
	}
}

// AllUsersInfoGET gets the info related to all users.
func AllUsersInfoGET(ctx *middlewares.AutheliaCtx) {
	var (
		err          error
		userInfo     []model.UserInfo
		usersDetails []authentication.UserDetailsExtended
	)
	if usersDetails, err = ctx.Providers.UserProvider.ListUsers(); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred retrieving users")
		return
	}

	if userInfo, err = ctx.Providers.StorageProvider.LoadAllUserInfoAndMetadata(ctx); err != nil {
		ctx.Error(fmt.Errorf("unable to load user attributes: %w", err), messageOperationFailed)
		return
	}

	usersDetails = MergeUserDetailsWithInfoMany(usersDetails, userInfo)

	err = ctx.SetJSONBody(usersDetails)
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
		ctx.Logger.WithError(err).Error("Error occurred retrieving user session")

		ctx.Error(err, messageOperationFailed)

		return
	}

	if err = ctx.ParseBody(&bodyJSON); err != nil {
		ctx.Error(err, messageOperationFailed)
		return
	}

	if !utils.IsStringInSlice(bodyJSON.Method, ctx.AvailableSecondFactorMethods()) {
		ctx.Error(fmt.Errorf("unknown or unavailable method '%s', it should be one of %s", bodyJSON.Method, strings.Join(ctx.AvailableSecondFactorMethods(), ", ")), messageOperationFailed)
		return
	}

	ctx.Logger.Debugf("Save new preferred 2FA method of user %s to %s", userSession.Username, bodyJSON.Method)

	if err = ctx.Providers.StorageProvider.SavePreferred2FAMethod(ctx, userSession.Username, bodyJSON.Method); err != nil {
		ctx.Error(fmt.Errorf("unable to save new preferred 2FA method: %w", err), messageOperationFailed)
		return
	}

	ctx.ReplyOK()
}
