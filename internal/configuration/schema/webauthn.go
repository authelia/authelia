package schema

import (
	"time"

	"github.com/go-webauthn/webauthn/metadata"
	"github.com/go-webauthn/webauthn/protocol"
)

// WebAuthn represents the webauthn config.
type WebAuthn struct {
	Disable            bool   `koanf:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables the WebAuthn 2FA functionality."`
	DisplayName        string `koanf:"display_name" json:"display_name" jsonschema:"default=Authelia,title=Display Name" jsonschema_description:"The display name attribute for the WebAuthn relying party."`
	EnablePasskeyLogin bool   `koanf:"enable_passkey_login" json:"enable_passkey_login" jsonschema:"default=false,title=Enable Passkey Logins" jsonschema_description:"Allows users to sign in via Passkeys."`

	ConveyancePreference protocol.ConveyancePreference `koanf:"attestation_conveyance_preference" json:"attestation_conveyance_preference" jsonschema:"default=indirect,enum=none,enum=indirect,enum=direct,title=Conveyance Preference" jsonschema_description:"The default conveyance preference for all WebAuthn credentials."`

	Timeout time.Duration `koanf:"timeout" json:"timeout" jsonschema:"default=60 seconds,title=Timeout" jsonschema_description:"The default timeout for all WebAuthn ceremonies."`

	Filtering         WebAuthnFiltering         `koanf:"filtering" json:"filtering" jsonschema:"title=Filtering" jsonschema_description:"WebAuthn Authenticator filtering settings."`
	SelectionCriteria WebAuthnSelectionCriteria `koanf:"selection_criteria" json:"selection_criteria"`
	Metadata          WebAuthnMetadata          `koanf:"metadata" json:"metadata"`
}

type WebAuthnMetadata struct {
	Enabled bool   `koanf:"enabled" json:"enabled"`
	Path    string `koanf:"path" json:"path"`

	ValidateTrustAnchor           bool `koanf:"validate_trust_anchor" json:"validate_trust_anchor"`
	ValidateEntry                 bool `koanf:"validate_entry" json:"validate_entry"`
	ValidateEntryPermitZeroAAGUID bool `koanf:"validate_entry_permit_zero_aaguid" json:"validate_entry_permit_zero_aaguid"`

	ValidateStatus          bool                           `koanf:"validate_status" json:"validate_status"`
	ValidateStatusDesired   []metadata.AuthenticatorStatus `koanf:"validate_status_desired" json:"validate_status_desired" jsonschema:"enum=FIDO_NOT_CERTIFIED,enum=FIDO_CERTIFIED,enum=USER_VERIFICATION_BYPASS,enum=ATTESTATION_KEY_COMPROMISE,enum=USER_KEY_REMOTE_COMPROMISE,enum=USER_KEY_PHYSICAL_COMPROMISE,enum=UPDATE_AVAILABLE,enum=REVOKED,enum=SELF_ASSERTION_SUBMITTED,enum=FIDO_CERTIFIED_L1,enum=FIDO_CERTIFIED_L1plus,enum=FIDO_CERTIFIED_L2,enum=FIDO_CERTIFIED_L2plus,enum=FIDO_CERTIFIED_L3,enum=FIDO_CERTIFIED_L3plus"`
	ValidateStatusUndesired []metadata.AuthenticatorStatus `koanf:"validate_status_undesired" json:"validate_status_undesired" jsonschema:"enum=FIDO_NOT_CERTIFIED,enum=FIDO_CERTIFIED,enum=USER_VERIFICATION_BYPASS,enum=ATTESTATION_KEY_COMPROMISE,enum=USER_KEY_REMOTE_COMPROMISE,enum=USER_KEY_PHYSICAL_COMPROMISE,enum=UPDATE_AVAILABLE,enum=REVOKED,enum=SELF_ASSERTION_SUBMITTED,enum=FIDO_CERTIFIED_L1,enum=FIDO_CERTIFIED_L1plus,enum=FIDO_CERTIFIED_L2,enum=FIDO_CERTIFIED_L2plus,enum=FIDO_CERTIFIED_L3,enum=FIDO_CERTIFIED_L3plus"`
}

type WebAuthnSelectionCriteria struct {
	Attachment       protocol.AuthenticatorAttachment     `koanf:"attachment" json:"attachment" jsonschema:"default=cross-platform,enum=platform,enum=cross-platform"`
	Discoverability  protocol.ResidentKeyRequirement      `koanf:"discoverability" json:"discoverability"  jsonschema:"default=discouraged,enum=discouraged,enum=preferred,enum=required,title=Discoverability Selection" jsonschema_description:"The default discoverable requirement when registering WebAuthn credentials."`
	UserVerification protocol.UserVerificationRequirement `koanf:"user_verification" json:"user_verification" jsonschema:"default=preferred,enum=discouraged,enum=preferred,enum=required,title=User Verification" jsonschema_description:"The default user verification preference for all WebAuthn credentials."`
}

type WebAuthnFiltering struct {
	PermittedAAGUIDs  []string `koanf:"permitted_aaguids" json:"permitted_aaguids"`
	ProhibitedAAGUIDs []string `koanf:"prohibited_aaguids" json:"prohibited_aaguids"`
}

// DefaultWebAuthnConfiguration describes the default values for the WebAuthn.
var DefaultWebAuthnConfiguration = WebAuthn{
	DisplayName: "Authelia",
	Timeout:     time.Second * 60,

	ConveyancePreference: protocol.PreferIndirectAttestation,
	SelectionCriteria: WebAuthnSelectionCriteria{
		Attachment:       protocol.CrossPlatform,
		Discoverability:  protocol.ResidentKeyRequirementDiscouraged,
		UserVerification: protocol.VerificationPreferred,
	},
	Metadata: WebAuthnMetadata{
		Enabled:                       false,
		Path:                          "",
		ValidateTrustAnchor:           false,
		ValidateEntry:                 false,
		ValidateEntryPermitZeroAAGUID: false,
		ValidateStatus:                false,
		ValidateStatusDesired:         []metadata.AuthenticatorStatus{},
		ValidateStatusUndesired: []metadata.AuthenticatorStatus{
			metadata.AttestationKeyCompromise,
			metadata.UserVerificationBypass,
			metadata.UserKeyRemoteCompromise,
			metadata.UserKeyPhysicalCompromise,
			metadata.Revoked,
		},
	},
}
