package schema

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
)

// WebAuthn represents the webauthn config.
type WebAuthn struct {
	Disable     bool   `koanf:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables the WebAuthn 2FA functionality."`
	DisplayName string `koanf:"display_name" json:"display_name" jsonschema:"default=Authelia,title=Display Name" jsonschema_description:"The display name attribute for the WebAuthn relying party."`

	ConveyancePreference protocol.ConveyancePreference        `koanf:"attestation_conveyance_preference" json:"attestation_conveyance_preference" jsonschema:"default=indirect,enum=none,enum=indirect,enum=direct,title=Conveyance Preference" jsonschema_description:"The default conveyance preference for all WebAuthn credentials."`
	UserVerification     protocol.UserVerificationRequirement `koanf:"user_verification" json:"user_verification" jsonschema:"default=preferred,enum=discouraged,enum=preferred,enum=required,title=User Verification" jsonschema_description:"The default user verification preference for all WebAuthn credentials."`

	Timeout time.Duration `koanf:"timeout" json:"timeout" jsonschema:"default=60 seconds,title=Timeout" jsonschema_description:"The default timeout for all WebAuthn ceremonies."`
}

// DefaultWebAuthnConfiguration describes the default values for the WebAuthn.
var DefaultWebAuthnConfiguration = WebAuthn{
	DisplayName: "Authelia",
	Timeout:     time.Second * 60,

	ConveyancePreference: protocol.PreferIndirectAttestation,
	UserVerification:     protocol.VerificationPreferred,
}
