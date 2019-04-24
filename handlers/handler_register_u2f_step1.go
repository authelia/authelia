package handlers

import (
	"fmt"

	"github.com/clems4ever/authelia/middlewares"
	"github.com/tstranex/u2f"
)

var u2fConfig = &u2f.Config{
	// Chrome 66+ doesn't return the device's attestation
	// certificate by default.
	SkipAttestationVerify: true,
}

// SecondFactorU2FIdentityStart the handler for initiating the identity validation.
var SecondFactorU2FIdentityStart = middlewares.IdentityVerificationStart(middlewares.IdentityVerificationStartArgs{
	MailSubject:           "[Authelia] Register your key",
	MailTitle:             "Register your key",
	MailButtonContent:     "Register",
	TargetEndpoint:        "/security-key-registration",
	ActionClaim:           U2FRegistrationAction,
	IdentityRetrieverFunc: identityRetrieverFromSession,
})

func secondFactorU2FIdentityFinish(ctx *middlewares.AutheliaCtx, username string) {
	appID := fmt.Sprintf("%s://%s", ctx.XForwardedProto(), ctx.XForwardedHost())
	ctx.Logger.Debugf("U2F appID is %s", appID)
	var trustedFacets = []string{appID}

	challenge, err := u2f.NewChallenge(appID, trustedFacets)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to generate new U2F challenge for registration: %s", err), operationFailedMessage)
		return
	}

	// Save the challenge in the user session.
	userSession := ctx.GetSession()
	userSession.U2FChallenge = challenge
	err = ctx.SaveSession(userSession)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to save U2F challenge in session: %s", err), operationFailedMessage)
		return
	}

	request := u2f.NewWebRegisterRequest(challenge, []u2f.Registration{})

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to generate new U2F request for registration: %s", err), operationFailedMessage)
		return
	}

	ctx.SetJSONBody(request)
}

// SecondFactorU2FIdentityFinish the handler for finishing the identity validation
var SecondFactorU2FIdentityFinish = middlewares.IdentityVerificationFinish(
	middlewares.IdentityVerificationFinishArgs{
		ActionClaim:          U2FRegistrationAction,
		IsTokenUserValidFunc: isTokenUserValidFor2FARegistration,
	}, secondFactorU2FIdentityFinish)
