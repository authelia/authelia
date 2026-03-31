package handlers

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

// ConfigurationGET get the configuration accessible to authenticated users.
func ConfigurationGET(ctx *middlewares.AutheliaCtx) {
	body := configurationBody{
		AvailableMethods:       make(MethodList, 0, 3),
		PasswordChangeDisabled: false,
		PasswordResetDisabled:  false,
	}

	if ctx.Providers.Authorizer.IsSecondFactorEnabled() {
		body.AvailableMethods = ctx.AvailableSecondFactorMethods()
	}

	body.PasswordChangeDisabled = ctx.Configuration.AuthenticationBackend.PasswordChange.Disable
	body.PasswordResetDisabled = ctx.Configuration.AuthenticationBackend.PasswordReset.Disable

	ctx.Logger.WithFields(
		map[string]any{
			"available_methods":        body.AvailableMethods,
			"password_change_disabled": body.PasswordChangeDisabled,
			"password_reset_disabled":  body.PasswordResetDisabled,
		}).Trace("Authelia configuration requested")

	if err := ctx.SetJSONBody(body); err != nil {
		ctx.Logger.Errorf("Unable to set configuration response in body: %s", err)
	}
}

type AdminConfigurationResponseBody struct {
	Enabled                bool   `json:"enabled"`
	AdminGroup             string `json:"admin_group"`
	AllowAdminsToAddAdmins bool   `json:"allow_admins_to_add_admins"`
}

// AdminConfigurationGET get the configuration accessible to authenticated administrators.
func AdminConfigurationGET(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		userSession session.UserSession
		adminConfig AdminConfigurationResponseBody
	)
	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving admin config: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred retrieving admin config")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	adminConfig = AdminConfigurationResponseBody{
		Enabled:                ctx.Configuration.Administration.Enabled,
		AdminGroup:             ctx.Configuration.Administration.AdminGroup,
		AllowAdminsToAddAdmins: ctx.Configuration.Administration.AllowAdminsToAddAdmins,
	}

	err = ctx.SetJSONBody(adminConfig)
	if err != nil {
		ctx.Logger.Errorf("Unable to set admin config response in body: %+v", err)
	}
}
