package schema

import (
	"github.com/duo-labs/webauthn/protocol"
)

// WebauthnConfiguration represents the webauthn config.
type WebauthnConfiguration struct {
	Disable     bool   `koanf:"disable"`
	DisplayName string `koanf:"display_name"`

	ConveyancePreference protocol.ConveyancePreference        `koanf:"attestation_conveyance_preference"`
	UserVerification     protocol.UserVerificationRequirement `koanf:"user_verification"`

	Timeout string `koanf:"timeout"`
}

// DefaultWebauthnConfiguration describes the default values for the WebauthnConfiguration.
var DefaultWebauthnConfiguration = WebauthnConfiguration{
	DisplayName: "Authelia",
	Timeout:     "60s",

	ConveyancePreference: protocol.PreferIndirectAttestation,
	UserVerification:     protocol.VerificationPreferred,
}
