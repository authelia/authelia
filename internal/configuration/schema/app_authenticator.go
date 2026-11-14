// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"time"
)

const (
	// AppAuthenticatorAttestationModeRequired requires a valid attestation from a trusted app identity.
	AppAuthenticatorAttestationModeRequired = "required"

	// AppAuthenticatorAttestationModePreferred verifies attestation when the device provides it.
	AppAuthenticatorAttestationModePreferred = "preferred"

	// AppAuthenticatorAttestationModeNone does not verify attestation.
	AppAuthenticatorAttestationModeNone = "none"
)

// AppAuthenticator represents the configuration related to the Authelia mobile app authenticator.
type AppAuthenticator struct {
	Disable                   bool                        `koanf:"disable" yaml:"disable" toml:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables the app authenticator second factor method."`
	DisplayName               string                      `koanf:"display_name" yaml:"display_name,omitempty" toml:"display_name,omitempty" json:"display_name,omitempty" jsonschema:"default=Authelia,title=Display Name" jsonschema_description:"The name shown in the app for this server."`
	RegistrationTokenLifespan time.Duration               `koanf:"registration_token_lifespan" yaml:"registration_token_lifespan,omitempty" toml:"registration_token_lifespan,omitempty" json:"registration_token_lifespan,omitempty" jsonschema:"default=5 minutes,title=Registration Token Lifespan" jsonschema_description:"How long the registration QR code remains valid."`
	ChallengeLifespan         time.Duration               `koanf:"challenge_lifespan" yaml:"challenge_lifespan,omitempty" toml:"challenge_lifespan,omitempty" json:"challenge_lifespan,omitempty" jsonschema:"default=2 minutes,title=Challenge Lifespan" jsonschema_description:"How long a sign in challenge remains valid."`
	NonceLength               int                         `koanf:"nonce_length" yaml:"nonce_length,omitempty" toml:"nonce_length,omitempty" json:"nonce_length,omitempty" jsonschema:"default=2,minimum=2,maximum=4,title=Nonce Length" jsonschema_description:"The number of digits in the number matching prompt."`
	Attestation               AppAuthenticatorAttestation `koanf:"attestation" yaml:"attestation,omitempty" toml:"attestation,omitempty" json:"attestation,omitempty" jsonschema:"title=Attestation" jsonschema_description:"Device attestation settings."`
}

// AppAuthenticatorAttestation represents the attestation configuration of the app authenticator.
type AppAuthenticatorAttestation struct {
	Mode    string                            `koanf:"mode" yaml:"mode,omitempty" toml:"mode,omitempty" json:"mode,omitempty" jsonschema:"default=required,enum=required,enum=preferred,enum=none,title=Mode" jsonschema_description:"The attestation mode."`
	Apple   []AppAuthenticatorAppleIdentity   `koanf:"apple" yaml:"apple,omitempty" toml:"apple,omitempty" json:"apple,omitempty" jsonschema:"title=Apple" jsonschema_description:"The trusted Apple app identities."`
	Android []AppAuthenticatorAndroidIdentity `koanf:"android" yaml:"android,omitempty" toml:"android,omitempty" json:"android,omitempty" jsonschema:"title=Android" jsonschema_description:"The trusted Android app identities."`
}

// HasIdentities returns true if any trusted app identity is configured.
func (a AppAuthenticatorAttestation) HasIdentities() bool {
	return len(a.Apple) != 0 || len(a.Android) != 0
}

// AppAuthenticatorAppleIdentity represents a trusted Apple app identity.
type AppAuthenticatorAppleIdentity struct {
	TeamID      string `koanf:"team_id" yaml:"team_id" toml:"team_id" json:"team_id" jsonschema:"title=Team ID" jsonschema_description:"The Apple developer team identifier."`
	BundleID    string `koanf:"bundle_id" yaml:"bundle_id" toml:"bundle_id" json:"bundle_id" jsonschema:"title=Bundle ID" jsonschema_description:"The app bundle identifier."`
	Environment string `koanf:"environment" yaml:"environment,omitempty" toml:"environment,omitempty" json:"environment,omitempty" jsonschema:"default=production,enum=production,enum=development,title=Environment" jsonschema_description:"The App Attest environment."`
}

// AppAuthenticatorAndroidIdentity represents a trusted Android app identity.
type AppAuthenticatorAndroidIdentity struct {
	PackageName               string   `koanf:"package_name" yaml:"package_name" toml:"package_name" json:"package_name" jsonschema:"title=Package Name" jsonschema_description:"The Android package name."`
	SigningCertificateDigests []string `koanf:"signing_certificate_digests" yaml:"signing_certificate_digests" toml:"signing_certificate_digests" json:"signing_certificate_digests" jsonschema:"title=Signing Certificate Digests" jsonschema_description:"The SHA-256 digests of the trusted signing certificates."`
}

// DefaultAppAuthenticatorConfiguration describes the default values for the app authenticator.
var DefaultAppAuthenticatorConfiguration = AppAuthenticator{
	DisplayName:               "Authelia",
	RegistrationTokenLifespan: 5 * time.Minute,
	ChallengeLifespan:         2 * time.Minute,
	NonceLength:               2,
	Attestation: AppAuthenticatorAttestation{
		Mode: AppAuthenticatorAttestationModeRequired,
	},
}
