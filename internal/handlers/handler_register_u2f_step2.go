package handlers

import (
	"crypto/elliptic"
	"fmt"

	"github.com/tstranex/u2f"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// SecondFactorU2FRegister handler validating the client has successfully validated the challenge
// to complete the U2F registration.
func SecondFactorU2FRegister(ctx *middlewares.AutheliaCtx) {
	responseBody := u2f.RegisterResponse{}
	err := ctx.ParseBody(&responseBody)

	if err != nil {
		ctx.Error(fmt.Errorf("unable to parse response body: %v", err), messageUnableToRegisterSecurityKey)
	}

	userSession := ctx.GetSession()

	if userSession.U2FChallenge == nil {
		ctx.Error(fmt.Errorf("U2F registration has not been initiated yet"), messageUnableToRegisterSecurityKey)
		return
	}
	// Ensure the challenge is cleared if anything goes wrong.
	defer func() {
		userSession.U2FChallenge = nil

		err := ctx.SaveSession(userSession)
		if err != nil {
			ctx.Logger.Errorf("unable to clear U2F challenge in session for user %s: %s", userSession.Username, err)
		}
	}()

	registration, err := u2f.Register(responseBody, *userSession.U2FChallenge, u2fConfig)

	if err != nil {
		ctx.Error(fmt.Errorf("unable to verify U2F registration: %v", err), messageUnableToRegisterSecurityKey)
		return
	}

	ctx.Logger.Debugf("register U2F device for user %s", userSession.Username)

	publicKey := elliptic.Marshal(elliptic.P256(), registration.PubKey.X, registration.PubKey.Y)
	err = ctx.Providers.StorageProvider.SaveU2FDeviceHandle(userSession.Username, registration.KeyHandle, publicKey)

	if err != nil {
		ctx.Error(fmt.Errorf("unable to register U2F device for user %s: %v", userSession.Username, err), messageUnableToRegisterSecurityKey)
		return
	}

	ctx.ReplyOK()
}
