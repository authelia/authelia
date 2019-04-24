package handlers

import (
	"fmt"

	"github.com/clems4ever/authelia/middlewares"
	"github.com/clems4ever/authelia/storage"
	"github.com/tstranex/u2f"
)

// SecondFactorU2FSignGet handler for initiating a signing request.
func SecondFactorU2FSignGet(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	appID := fmt.Sprintf("%s://%s", ctx.XForwardedProto(), ctx.XForwardedHost())
	var trustedFacets = []string{appID}
	challenge, err := u2f.NewChallenge(appID, trustedFacets)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to create U2F challenge: %s", err), mfaValidationFailedMessage)
		return
	}

	registrationBin, err := ctx.Providers.StorageProvider.LoadU2FDeviceHandle(userSession.Username)

	if err != nil {
		if err == storage.ErrNoU2FDeviceHandle {
			ctx.Error(fmt.Errorf("No device handle found for user %s", userSession.Username), mfaValidationFailedMessage)
			return
		}
		ctx.Error(fmt.Errorf("Unable to retrieve U2F device handle: %s", err), mfaValidationFailedMessage)
		return
	}

	if len(registrationBin) == 0 {
		ctx.Error(fmt.Errorf("Wrong format of device handler for user %s", userSession.Username), mfaValidationFailedMessage)
		return
	}

	var registration u2f.Registration
	err = registration.UnmarshalBinary(registrationBin)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to unmarshal U2F device handle: %s", err), mfaValidationFailedMessage)
		return
	}

	// Save the challenge and registration for use in next request
	userSession.U2FRegistration = &registration
	userSession.U2FChallenge = challenge
	err = ctx.SaveSession(userSession)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to save U2F challenge and registration in session: %s", err), mfaValidationFailedMessage)
		return
	}

	signRequest := challenge.SignRequest([]u2f.Registration{registration})
	err = ctx.SetJSONBody(signRequest)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to set sign request in body: %s", err), mfaValidationFailedMessage)
		return
	}
}
