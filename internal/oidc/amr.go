package oidc

func NewAuthenticationMethodsReferencesFromClaim(claim []string) (amr AuthenticationMethodsReferences) {
	for _, ref := range claim {
		switch ref {
		case AMRPasswordBasedAuthentication:
			amr.UsernameAndPassword = true
		case AMROneTimePassword:
			amr.TOTP = true
		case AMRShortMessageService:
			amr.Duo = true
		case AMRHardwareSecuredKey:
			amr.WebAuthn = true
		case AMRUserPresence:
			amr.WebAuthnUserVerified = true
		}
	}

	return amr
}

// AuthenticationMethodsReferences holds AMR information.
type AuthenticationMethodsReferences struct {
	UsernameAndPassword  bool
	TOTP                 bool
	Duo                  bool
	WebAuthn             bool
	WebAuthnUserPresence bool
	WebAuthnUserVerified bool
}

// FactorKnowledge returns true if a "something you know" factor of authentication was used.
func (r AuthenticationMethodsReferences) FactorKnowledge() bool {
	return r.UsernameAndPassword
}

// FactorPossession returns true if a "something you have" factor of authentication was used.
func (r AuthenticationMethodsReferences) FactorPossession() bool {
	return r.TOTP || r.WebAuthn || r.Duo
}

// MultiFactorAuthentication returns true if multiple factors were used.
func (r AuthenticationMethodsReferences) MultiFactorAuthentication() bool {
	return r.FactorKnowledge() && r.FactorPossession()
}

// ChannelBrowser returns true if a browser was used to authenticate.
func (r AuthenticationMethodsReferences) ChannelBrowser() bool {
	return r.UsernameAndPassword || r.TOTP || r.WebAuthn
}

// ChannelService returns true if a non-browser service was used to authenticate.
func (r AuthenticationMethodsReferences) ChannelService() bool {
	return r.Duo
}

// MultiChannelAuthentication returns true if the user used more than one channel to authenticate.
func (r AuthenticationMethodsReferences) MultiChannelAuthentication() bool {
	return r.ChannelBrowser() && r.ChannelService()
}

// MarshalRFC8176 returns the AMR claim slice of strings in the RFC8176 format.
// https://datatracker.ietf.org/doc/html/rfc8176
func (r AuthenticationMethodsReferences) MarshalRFC8176() []string {
	var amr []string

	if r.UsernameAndPassword {
		amr = append(amr, AMRPasswordBasedAuthentication)
	}

	if r.TOTP {
		amr = append(amr, AMROneTimePassword)
	}

	if r.Duo {
		amr = append(amr, AMRShortMessageService)
	}

	if r.WebAuthn {
		amr = append(amr, AMRHardwareSecuredKey)
	}

	if r.WebAuthnUserPresence {
		amr = append(amr, AMRUserPresence)
	}

	if r.WebAuthnUserVerified {
		amr = append(amr, AMRPersonalIdentificationNumber)
	}

	if r.MultiFactorAuthentication() {
		amr = append(amr, AMRMultiFactorAuthentication)
	}

	if r.MultiChannelAuthentication() {
		amr = append(amr, AMRMultiChannelAuthentication)
	}

	return amr
}
