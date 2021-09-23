package handlers

import (
	"crypto/elliptic"
	"fmt"

	"github.com/tstranex/u2f"

	"github.com/authelia/authelia/v4/internal/middlewares"
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

	appID := fmt.Sprintf("%s://%s", ctx.XForwardedProto(), ctx.XForwardedHost())

	var trustedFacets = []string{appID}
	challenge, err := u2f.NewChallenge(appID, trustedFacets)

	if err != nil {
		handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to create U2F challenge: %s", err), messageMFAValidationFailed)
		return
	}

	userSession := ctx.GetSession()
	keyHandleBytes, publicKeyBytes, err := ctx.Providers.StorageProvider.LoadU2FDeviceHandle(userSession.Username)

	if err != nil {
		if err == storage.ErrNoU2FDeviceHandle {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("no device handle found for user %s", userSession.Username), messageMFAValidationFailed)
			return
		}

		handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to retrieve U2F device handle: %s", err), messageMFAValidationFailed)

		return
	}

	var registration u2f.Registration
	registration.KeyHandle = keyHandleBytes
	x, y := elliptic.Unmarshal(elliptic.P256(), publicKeyBytes)
	registration.PubKey.Curve = elliptic.P256()
	registration.PubKey.X = x
	registration.PubKey.Y = y

	// Save the challenge and registration for use in next request
	userSession.U2FRegistration = &session.U2FRegistration{
		KeyHandle: keyHandleBytes,
		PublicKey: publicKeyBytes,
	}
	userSession.U2FChallenge = challenge
	err = ctx.SaveSession(userSession)

	if err != nil {
		handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to save U2F challenge and registration in session: %s", err), messageMFAValidationFailed)
		return
	}

	signRequest := challenge.SignRequest([]u2f.Registration{registration})
	err = ctx.SetJSONBody(signRequest)

	if err != nil {
		handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to set sign request in body: %s", err), messageMFAValidationFailed)
		return
	}
}
