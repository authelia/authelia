package schema

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
)

// WebAuthnConfiguration represents the webauthn config.
type WebAuthnConfiguration struct {
	Disable     bool   `koanf:"disable"`
	DisplayName string `koanf:"display_name"`

	ConveyancePreference protocol.ConveyancePreference        `koanf:"attestation_conveyance_preference"`
	UserVerification     protocol.UserVerificationRequirement `koanf:"user_verification"`

	Timeout time.Duration `koanf:"timeout"`
}

// DefaultWebAuthnConfiguration describes the default values for the WebAuthnConfiguration.
var DefaultWebAuthnConfiguration = WebAuthnConfiguration{
	DisplayName: "Authelia",
	Timeout:     time.Second * 60,

	ConveyancePreference: protocol.PreferIndirectAttestation,
	UserVerification:     protocol.VerificationPreferred,
}
