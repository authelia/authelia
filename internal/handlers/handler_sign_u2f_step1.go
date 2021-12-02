package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"github.com/tstranex/u2f"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

// SecondFactorU2FSignGet handler for initiating a signing request.
func SecondFactorU2FSignGet(ctx *middlewares.AutheliaCtx) {
	if ctx.XForwardedProto() == nil {
		ctx.Error(errMissingXForwardedProto, messageMFAValidationFailed)
		return
	}

	if ctx.XForwardedHost() == nil {
		ctx.Error(errMissingXForwardedHost, messageMFAValidationFailed)
		return
	}

	userSession := ctx.GetSession()

	appID := fmt.Sprintf("%s://%s", ctx.XForwardedProto(), ctx.XForwardedHost())

	var trustedFacets = []string{appID}

	challenge, err := u2f.NewChallenge(appID, trustedFacets)
	if err != nil {
		ctx.Logger.Errorf("Unable to create %s challenge for user '%s': %+v", regulation.AuthTypeU2F, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	device, err := ctx.Providers.StorageProvider.LoadU2FDevice(ctx, userSession.Username)
	if err != nil {
		respondUnauthorized(ctx, messageMFAValidationFailed)

		if err == storage.ErrNoU2FDeviceHandle {
			_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypeU2F, fmt.Errorf("no registered U2F device"))
			return
		}

		ctx.Logger.Errorf("Could not load %s devices for user '%s': %+v", regulation.AuthTypeU2F, userSession.Username, err)

		return
	}

	x, y := elliptic.Unmarshal(elliptic.P256(), device.PublicKey)

	registration := u2f.Registration{
		KeyHandle: device.KeyHandle,
		PubKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     x,
			Y:     y,
		},
	}

	// Save the challenge and registration for use in next request
	userSession.U2FRegistration = &session.U2FRegistration{
		KeyHandle: device.KeyHandle,
		PublicKey: device.PublicKey,
	}

	userSession.U2FChallenge = challenge

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf(logFmtErrSessionSave, "challenge and registration", regulation.AuthTypeU2F, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}

	signRequest := challenge.SignRequest([]u2f.Registration{registration})

	if err = ctx.SetJSONBody(signRequest); err != nil {
		ctx.Logger.Errorf(logFmtErrWriteResponseBody, regulation.AuthTypeU2F, userSession.Username, err)

		respondUnauthorized(ctx, messageMFAValidationFailed)

		return
	}
}
