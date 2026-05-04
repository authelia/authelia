// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

// RecoveryCodes represents the configuration related to user-managed second-factor recovery codes.
type RecoveryCodes struct {
	Disable bool `koanf:"disable" yaml:"disable" toml:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables the recovery codes feature for all users."`
}

// DefaultRecoveryCodesConfiguration is the default configuration for the recovery codes feature.
var DefaultRecoveryCodesConfiguration = RecoveryCodes{
	Disable: false,
}
