package handlers

import (
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
)

const (
	webauthnCredentialDescriptionMaxLen = 64
)

func handleGetWebAuthnUserByRPID(ctx *middlewares.AutheliaCtx, username, displayname string, rpid string) (user *model.WebAuthnUser, err error) {
	if user, err = ctx.Providers.StorageProvider.LoadWebAuthnUser(ctx, rpid, username); err != nil {
		return nil, err
	}

	if user == nil {
		user = &model.WebAuthnUser{
			RPID:        rpid,
			Username:    username,
			UserID:      ctx.Providers.Random.StringCustom(64, random.CharSetASCII),
			DisplayName: displayname,
		}

		if err = ctx.Providers.StorageProvider.SaveWebAuthnUser(ctx, *user); err != nil {
			return nil, err
		}
	} else {
		user.DisplayName = displayname
	}

	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}

	if user.Credentials, err = ctx.Providers.StorageProvider.LoadWebAuthnCredentialsByUsername(ctx, rpid, user.Username); err != nil {
		return nil, err
	}

	return user, nil
}

func handlerWebAuthnDiscoverableLogin(ctx *middlewares.AutheliaCtx, rpid string) webauthn.DiscoverableUserHandler {
	return func(rawID, userHandle []byte) (user webauthn.User, err error) {
		var u *model.WebAuthnUser

		if u, err = ctx.Providers.StorageProvider.LoadWebAuthnUserByUserID(ctx, rpid, string(userHandle)); err != nil {
			return nil, err
		}

		if u.Credentials, err = ctx.Providers.StorageProvider.LoadWebAuthnPasskeyCredentialsByUsername(ctx, rpid, u.Username); err != nil {
			return nil, err
		}

		return u, nil
	}
}
