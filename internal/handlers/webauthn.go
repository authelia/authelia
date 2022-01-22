package handlers

import (
	"fmt"
	"strings"

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

	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}

	if user.Devices, err = ctx.Providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, userSession.Username); err != nil {
		return nil, err
	}

	return user, nil
}

func getWebauthn(ctx *middlewares.AutheliaCtx) (w *webauthn.WebAuthn, appid string, err error) {
	config := &webauthn.Config{
		RPDisplayName: ctx.Configuration.Webauthn.DisplayName,

		AttestationPreference: ctx.Configuration.Webauthn.AttestationPreference,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			UserVerification: ctx.Configuration.Webauthn.UserVerification,
		},

		Timeout: ctx.Configuration.Webauthn.Timeout,
	}

	if ctx.Configuration.Server.ExternalURL.Scheme != "" && ctx.Configuration.Server.Host != "" {
		u := ctx.Configuration.Server.ExternalURL

		config.RPID = u.Hostname()
		config.RPOrigin = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		appid = fmt.Sprintf("%s://%s", u.Scheme, u.Hostname())
	} else {
		var (
			headerProtoV, headerXForwardedHostV []byte
		)

		if headerProtoV = ctx.XForwardedProto(); headerProtoV == nil {
			return nil, "", errMissingXForwardedProto
		}

		if headerXForwardedHostV = ctx.XForwardedHost(); headerXForwardedHostV == nil {
			return nil, "", errMissingXForwardedHost
		}

		hostname := strings.Split(string(headerXForwardedHostV), ":")[0]

		config.RPID = hostname
		config.RPOrigin = fmt.Sprintf("%s://%s", headerProtoV, hostname)
		appid = config.RPOrigin
	}

	ctx.Logger.Tracef("Creating new Webauthn RP instance with ID %s AppID %s and Origin %s", config.RPID, appid, config.RPOrigin)

	w, err = webauthn.New(config)

	return w, appid, err
}
