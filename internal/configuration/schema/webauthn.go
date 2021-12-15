package schema

import (
	"github.com/duo-labs/webauthn/protocol"
)

// WebauthnConfiguration represents the webauthn config.
type WebauthnConfiguration struct {
	DisplayName string `koanf:"display_name"`
	Timeout     int    `koanf:"timeout"`

	AttestationPreference protocol.ConveyancePreference        `koanf:"attestation_preference"`
	UserVerification      protocol.UserVerificationRequirement `koanf:"user_verification"`

	EnableU2F bool `koanf:"enable_u2f"`
}

// WebauthnAuthenticatorSelectionConfiguration represents the authenticator selection.
type WebauthnAuthenticatorSelectionConfiguration struct {
	UserVerification protocol.UserVerificationRequirement `koanf:"user_verification"`
}

// DefaultWebauthnConfiguration describes the default values for the WebauthnConfiguration.
var DefaultWebauthnConfiguration = WebauthnConfiguration{
	DisplayName: "Authelia",
	Timeout:     60000,

	AttestationPreference: protocol.PreferIndirectAttestation,
	UserVerification:      protocol.VerificationPreferred,
}
