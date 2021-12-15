package handlers

import (
	"crypto/elliptic"
	"fmt"

	"github.com/tstranex/u2f"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/models"
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
	TargetEndpoint:        "/u2f/register",
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
		ctx.Logger.Errorf("Unable to create request to enrol new token: %s", err)
	}
}

// SecondFactorU2FIdentityFinish the handler for finishing the identity validation.
var SecondFactorU2FIdentityFinish = middlewares.IdentityVerificationFinish(
	middlewares.IdentityVerificationFinishArgs{
		ActionClaim:          ActionU2FRegistration,
		IsTokenUserValidFunc: isTokenUserValidFor2FARegistration,
	}, secondFactorU2FIdentityFinish)

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
			ctx.Logger.Errorf("Unable to clear U2F challenge in session for user %s: %s", userSession.Username, err)
		}
	}()

	registration, err := u2f.Register(responseBody, *userSession.U2FChallenge, u2fConfig)

	if err != nil {
		ctx.Error(fmt.Errorf("unable to verify U2F registration: %v", err), messageUnableToRegisterSecurityKey)
		return
	}

	ctx.Logger.Debugf("Register U2F device for user %s", userSession.Username)

	publicKey := elliptic.Marshal(elliptic.P256(), registration.PubKey.X, registration.PubKey.Y)

	device := models.WebauthnDevice{
		IP:              models.NewIP(ctx.RemoteIP()),
		Created:         ctx.Clock.Now(),
		Username:        userSession.Username,
		Description:     "Primary",
		KID:             models.NewBase64(registration.KeyHandle),
		PublicKey:       publicKey,
		AttestationType: "fido-u2f",
	}

	if err = ctx.Providers.StorageProvider.SaveWebauthnDevice(ctx, device); err != nil {
		ctx.Error(fmt.Errorf("unable to register U2F device for user %s: %v", userSession.Username, err), messageUnableToRegisterSecurityKey)

		return
	}

	ctx.ReplyOK()
}
