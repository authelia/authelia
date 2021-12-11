package validator

import (
	"github.com/duo-labs/webauthn/protocol"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateWebauthn validates and update Webauthn configuration.
func ValidateWebauthn(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.Webauthn.AuthenticatorSelection == nil {
		configuration.Webauthn.AuthenticatorSelection = &schema.WebAuthnAuthenticatorSelectionConfiguration{
			UserVerification: protocol.VerificationPreferred,
		}
	}
}
