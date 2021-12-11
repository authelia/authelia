package schema

import (
	"github.com/duo-labs/webauthn/protocol"
)

// WebauthnConfiguration represents the webauthn config.
type WebauthnConfiguration struct {
	Enabled     bool   `koanf:"enabled"`
	DisplayName string `koanf:"display_name"`
	Timeout     int    `koanf:"timeout"`
	Debug       bool   `koanf:"debug"`

	AttestationPreference protocol.ConveyancePreference `koanf:"attestation_preference"`

	AuthenticatorSelection *WebAuthnAuthenticatorSelectionConfiguration `koanf:"authenticator_selection"`
}

// WebAuthnAuthenticatorSelectionConfiguration represents the authenticator selection.
type WebAuthnAuthenticatorSelectionConfiguration struct {
	AuthenticatorAttachment protocol.AuthenticatorAttachment     `koanf:"authenticator_attachment"`
	RequireResidentKey      bool                                 `koanf:"require_resident_key"`
	UserVerification        protocol.UserVerificationRequirement `koanf:"user_verification"`
}
