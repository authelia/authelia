package authorization

func NewAuthenticationMethodsReferencesFromClaim(claim []string) (amr AuthenticationMethodsReferences) {
	for _, ref := range claim {
		switch ref {
		case AMRKnowledgeBasedAuthentication:
			amr.KnowledgeBasedAuthentication = true
		case AMRPasswordBasedAuthentication:
			amr.UsernameAndPassword = true
		case AMROneTimePassword:
			amr.TOTP = true
		case AMRShortMessageService:
			amr.Duo = true
		case AMRProofOfPossession:
			amr.WebAuthn = true
		case AMRHardwareSecuredKey:
			amr.WebAuthn = true
			amr.WebAuthnHardware = true
		case AMRSoftwareSecuredKey:
			amr.WebAuthn = true
			amr.WebAuthnSoftware = true
		case AMRUserPresence:
			amr.WebAuthnUserVerified = true
		case AMRWindowsIntegratedAuthentication:
			// Kerberos is the only WIA method Authelia supports currently.
			amr.Kerberos = true
		}
	}

	return amr
}

// AuthenticationMethodsReferences holds AMR information.
type AuthenticationMethodsReferences struct {
	KnowledgeBasedAuthentication bool
	UsernameAndPassword          bool
	TOTP                         bool
	Duo                          bool
	WebAuthn                     bool
	WebAuthnHardware             bool
	WebAuthnSoftware             bool
	WebAuthnUserPresence         bool
	WebAuthnUserVerified         bool
	Kerberos                     bool
}

// FactorKnowledge returns true if a "something you know" factor of authentication was used.
func (r AuthenticationMethodsReferences) FactorKnowledge() bool {
	return r.UsernameAndPassword || r.KnowledgeBasedAuthentication
}

// FactorPossession returns true if a "something you have" factor of authentication was used.
func (r AuthenticationMethodsReferences) FactorPossession() bool {
	return r.TOTP || r.Duo || r.WebAuthn || r.WebAuthnHardware || r.WebAuthnSoftware || r.Kerberos
}

// MultiFactorAuthentication returns true if multiple factors were used.
func (r AuthenticationMethodsReferences) MultiFactorAuthentication() bool {
	return r.FactorKnowledge() && r.FactorPossession()
}

// ChannelBrowser returns true if a browser was used to authenticate.
func (r AuthenticationMethodsReferences) ChannelBrowser() bool {
	return r.UsernameAndPassword || r.TOTP || r.WebAuthn || r.WebAuthnHardware || r.WebAuthnSoftware || r.Kerberos
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

	if r.KnowledgeBasedAuthentication {
		amr = append(amr, AMRKnowledgeBasedAuthentication)
	}

	if r.TOTP {
		amr = append(amr, AMROneTimePassword)
	}

	if r.Duo {
		amr = append(amr, AMRShortMessageService)
	}

	if r.Kerberos {
		amr = append(amr, AMRWindowsIntegratedAuthentication)
	}

	if r.WebAuthn || r.WebAuthnHardware || r.WebAuthnSoftware {
		amr = append(amr, AMRProofOfPossession)
	}

	if r.WebAuthnHardware {
		amr = append(amr, AMRHardwareSecuredKey)
	}

	if r.WebAuthnSoftware {
		amr = append(amr, AMRSoftwareSecuredKey)
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
