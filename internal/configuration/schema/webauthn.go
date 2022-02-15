package schema

import (
	"github.com/duo-labs/webauthn/protocol"
)

// WebauthnConfiguration represents the webauthn config.
type WebauthnConfiguration struct {
	DisplayName string `koanf:"display_name"`
	Timeout     int    `koanf:"timeout"`

	ConveyancePreference protocol.ConveyancePreference        `koanf:"conveyance_preference"`
	UserVerification     protocol.UserVerificationRequirement `koanf:"user_verification"`
}

// WebauthnAuthenticatorSelectionConfiguration represents the authenticator selection.
type WebauthnAuthenticatorSelectionConfiguration struct {
	UserVerification protocol.UserVerificationRequirement `koanf:"user_verification"`
}

// DefaultWebauthnConfiguration describes the default values for the WebauthnConfiguration.
var DefaultWebauthnConfiguration = WebauthnConfiguration{
	DisplayName: "Authelia",
	Timeout:     60000,

	ConveyancePreference: protocol.PreferIndirectAttestation,
	UserVerification:     protocol.VerificationPreferred,
}
