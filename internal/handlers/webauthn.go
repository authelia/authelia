package handlers

import (
	"net/url"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
)

func getWebauthnUserByRPID(ctx *middlewares.AutheliaCtx, username, description string, rpid string) (user *model.WebauthnUser, err error) {
	if user, err = ctx.Providers.StorageProvider.LoadWebauthnUser(ctx, rpid, username); err != nil {
		return nil, err
	}

	if user == nil {
		user = &model.WebauthnUser{
			RPID:        rpid,
			Username:    username,
			UserID:      ctx.Providers.Random.StringCustom(64, random.CharSetASCII),
			DisplayName: description,
		}

		if err = ctx.Providers.StorageProvider.SaveWebauthnUser(ctx, *user); err != nil {
			return nil, err
		}
	}

	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}

	if user.Devices, err = ctx.Providers.StorageProvider.LoadWebauthnDevicesByUsername(ctx, rpid, user.Username); err != nil {
		return nil, err
	}

	return user, nil
}

func newWebauthn(ctx *middlewares.AutheliaCtx) (w *webauthn.WebAuthn, err error) {
	var (
		origin *url.URL
	)

	if origin, err = ctx.GetOrigin(); err != nil {
		return nil, err
	}

	config := &webauthn.Config{
		RPDisplayName: ctx.Configuration.Webauthn.DisplayName,
		RPID:          origin.Hostname(),
		RPOrigins:     []string{origin.String()},
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
