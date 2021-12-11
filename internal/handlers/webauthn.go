package handlers

import (
	"fmt"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/models"
	"github.com/authelia/authelia/v4/internal/session"
)

func getWebAuthnUser(ctx *middlewares.AutheliaCtx, userSession session.UserSession) (user *models.WebauthnUser, err error) {
	user = &models.WebauthnUser{
		Username:    userSession.Username,
		DisplayName: userSession.DisplayName,
	}

	if user.Devices, err = ctx.Providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, userSession.Username); err != nil {
		return nil, err
	}

	return user, nil
}

func getWebauthn(ctx *middlewares.AutheliaCtx) (w *webauthn.WebAuthn, appid string, err error) {
	var (
		headerProtoV, headerXForwardedHostV []byte
	)

	config := &webauthn.Config{
		RPDisplayName: "Authelia",

		AttestationPreference: ctx.Configuration.Webauthn.AttestationPreference,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			UserVerification: ctx.Configuration.Webauthn.UserVerification,
		},

		Timeout: ctx.Configuration.Webauthn.Timeout,
		Debug:   ctx.Configuration.Webauthn.Debug,
	}

	if ctx.Configuration.Server.ExternalURL.Scheme != "" && ctx.Configuration.Server.Host != "" {
		u := ctx.Configuration.Server.ExternalURL

		config.RPID = u.Hostname()
		config.RPOrigin = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		appid = fmt.Sprintf("%s://%s", u.Scheme, u.Hostname())
	} else {
		if headerProtoV = ctx.XForwardedProto(); headerProtoV == nil {
			return nil, "", errMissingXForwardedProto
		}

		if headerXForwardedHostV = ctx.XForwardedHost(); headerXForwardedHostV == nil {
			return nil, "", errMissingXForwardedHost
		}

		config.RPID = string(headerXForwardedHostV)
		config.RPOrigin = fmt.Sprintf("%s://%s", headerProtoV, headerXForwardedHostV)
		appid = config.RPOrigin
	}

	ctx.Logger.Tracef("Creating new Webauthn RP instance with ID %s AppID %s and Origin %s", config.RPID, appid, config.RPOrigin)

	w, err = webauthn.New(config)

	return w, appid, err
}
