package handlers

import (
	"fmt"

	"github.com/tstranex/u2f"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

var u2fConfig = &u2f.Config{
	// Chrome 66+ doesn't return the device's attestation
	// certificate by default.
	SkipAttestationVerify: true,
}

// SecondFactorU2FIdentityStart the handler for initiating the identity validation.
var SecondFactorU2FIdentityStart = middlewares.IdentityVerificationStart(middlewares.IdentityVerificationStartArgs{
	MailTitle:             "Register your key",
	MailButtonContent:     "Register",
	TargetEndpoint:        "/security-key/register",
	ActionClaim:           ActionU2FRegistration,
	IdentityRetrieverFunc: identityRetrieverFromSession,
})

func secondFactorU2FIdentityFinish(ctx *middlewares.AutheliaCtx, username string) {
	if ctx.XForwardedProto() == nil {
		ctx.Error(errMissingXForwardedProto, messageOperationFailed)
		return
	}

	if ctx.XForwardedHost() == nil {
		ctx.Error(errMissingXForwardedHost, messageOperationFailed)
		return
	}

	appID := fmt.Sprintf("%s://%s", ctx.XForwardedProto(), ctx.XForwardedHost())
	ctx.Logger.Tracef("U2F appID is %s", appID)

	var trustedFacets = []string{appID}

	challenge, err := u2f.NewChallenge(appID, trustedFacets)

	if err != nil {
		ctx.Error(fmt.Errorf("unable to generate new U2F challenge for registration: %s", err), messageOperationFailed)
		return
	}

	// Save the challenge in the user session.
	userSession := ctx.GetSession()
	userSession.U2FChallenge = challenge
	err = ctx.SaveSession(userSession)

	if err != nil {
		ctx.Error(fmt.Errorf("unable to save U2F challenge in session: %s", err), messageOperationFailed)
		return
	}

	err = ctx.SetJSONBody(u2f.NewWebRegisterRequest(challenge, []u2f.Registration{}))
	if err != nil {
		ctx.Logger.Errorf("unable to create request to enrol new token: %s", err)
	}
}

// SecondFactorU2FIdentityFinish the handler for finishing the identity validation.
var SecondFactorU2FIdentityFinish = middlewares.IdentityVerificationFinish(
	middlewares.IdentityVerificationFinishArgs{
		ActionClaim:          ActionU2FRegistration,
		IsTokenUserValidFunc: isTokenUserValidFor2FARegistration,
	}, secondFactorU2FIdentityFinish)
