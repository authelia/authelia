package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

func getWebAuthnUser(ctx *middlewares.AutheliaCtx, userSession session.UserSession) (user *model.WebauthnUser, err error) {
	user = &model.WebauthnUser{
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

func newWebauthn(ctx *middlewares.AutheliaCtx) (w *webauthn.WebAuthn, err error) {
	var (
		u *url.URL
	)

	if u, err = ctx.GetXOriginalURLOrXForwardedURL(); err != nil {
		return nil, err
	}

	rpID := u.Hostname()
	origin := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	config := &webauthn.Config{
		RPDisplayName: ctx.Configuration.Webauthn.DisplayName,
		RPID:          rpID,
		RPOrigins:     []string{origin},
		RPIcon:        "",

		AttestationPreference: ctx.Configuration.Webauthn.ConveyancePreference,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.CrossPlatform,
			UserVerification:        ctx.Configuration.Webauthn.UserVerification,
			RequireResidentKey:      protocol.ResidentKeyNotRequired(),
		},

		Timeout: int(ctx.Configuration.Webauthn.Timeout.Milliseconds()),
	}

	ctx.Logger.Tracef("Creating new Webauthn RP instance with ID %s and Origins %s", config.RPID, strings.Join(config.RPOrigins, ", "))

	return webauthn.New(config)
}
