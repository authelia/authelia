package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
)

type userManagementRequestBody struct {
	UserChanges []model.UserInfoChanges `json:"user_changes"`
}

func UserInfoChangePOST(ctx *middlewares.AutheliaCtx) {
	var (
		err               error
		totalRowsAffected int64
		rowsAffected      int64
	)

	var requestBody userManagementRequestBody

	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.Error(err, messageUnableToModifyUser)
		return
	}

	if len(requestBody.UserChanges) == 0 {
		ctx.Logger.Debugf("request body is empty, no users changed")
		return
	}

	for _, userInfo := range requestBody.UserChanges {
		if userInfo.Username == "" {
			ctx.Logger.Errorf("username is required")
			return
		}

		rowsAffected, err = ctx.Providers.StorageProvider.UpdateUserAttributesByUsername(ctx, userInfo.Disabled, userInfo.PasswordChangeRequired, userInfo.LogoutRequired, userInfo.Username)

		totalRowsAffected += rowsAffected
	}

	if err != nil {
		ctx.Logger.WithError(err).Error("Error occurred modifying users")
		ctx.Response.SetStatusCode(500)

		return
	}

	ctx.Response.SetStatusCode(200)
}
