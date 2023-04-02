// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
)

// WebauthnConfiguration represents the webauthn config.
type WebauthnConfiguration struct {
	Disable     bool   `koanf:"disable"`
	DisplayName string `koanf:"display_name"`

	ConveyancePreference protocol.ConveyancePreference        `koanf:"attestation_conveyance_preference"`
	UserVerification     protocol.UserVerificationRequirement `koanf:"user_verification"`

	Timeout time.Duration `koanf:"timeout"`
}

// DefaultWebauthnConfiguration describes the default values for the WebauthnConfiguration.
var DefaultWebauthnConfiguration = WebauthnConfiguration{
	DisplayName: "Authelia",
	Timeout:     time.Second * 60,

	ConveyancePreference: protocol.PreferIndirectAttestation,
	UserVerification:     protocol.VerificationPreferred,
}
