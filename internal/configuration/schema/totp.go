// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"net/url"
)

// TOTP represents the configuration related to TOTP options.
type TOTP struct {
	Disable          bool   `koanf:"disable" yaml:"disable" toml:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables the TOTP 2FA functionality."`
	Issuer           string `koanf:"issuer" yaml:"issuer,omitempty" toml:"issuer,omitempty" json:"issuer,omitempty" jsonschema:"default=Authelia,title=Issuer" jsonschema_description:"The issuer value for generated TOTP keys."`
	DefaultAlgorithm string `koanf:"algorithm" yaml:"algorithm,omitempty" toml:"algorithm,omitempty" json:"algorithm,omitempty" jsonschema:"default=SHA1,enum=SHA1,enum=SHA256,enum=SHA512,title=Algorithm" jsonschema_description:"The algorithm value for generated TOTP keys."`
	DefaultDigits    int    `koanf:"digits" yaml:"digits" toml:"digits" json:"digits" jsonschema:"default=6,enum=6,enum=8,title=Digits" jsonschema_description:"The digits value for generated TOTP keys."`
	DefaultPeriod    int    `koanf:"period" yaml:"period" toml:"period" json:"period" jsonschema:"default=30,title=Period" jsonschema_description:"The period value for generated TOTP keys."`
	Skew             *int   `koanf:"skew" yaml:"skew,omitempty" toml:"skew,omitempty" json:"skew,omitempty" jsonschema:"default=1,title=Skew" jsonschema_description:"The permitted skew for generated TOTP keys."`
	SecretSize       int    `koanf:"secret_size" yaml:"secret_size" toml:"secret_size" json:"secret_size" jsonschema:"default=32,minimum=20,title=Secret Size" jsonschema_description:"The secret size for generated TOTP keys."`

	AllowedAlgorithms []string `koanf:"allowed_algorithms" yaml:"allowed_algorithms,omitempty" toml:"allowed_algorithms,omitempty" json:"allowed_algorithms,omitempty" jsonschema:"title=Allowed Algorithms,enum=SHA1,enum=SHA256,enum=SHA512,default=SHA1" jsonschema_description:"List of algorithms the user is allowed to select in addition to the default."`
	AllowedDigits     []int    `koanf:"allowed_digits" yaml:"allowed_digits,omitempty" toml:"allowed_digits,omitempty" json:"allowed_digits,omitempty" jsonschema:"title=Allowed Digits,enum=6,enum=8,default=6" jsonschema_description:"List of digits the user is allowed to select in addition to the default."`
	AllowedPeriods    []int    `koanf:"allowed_periods" yaml:"allowed_periods,omitempty" toml:"allowed_periods,omitempty" json:"allowed_periods,omitempty" jsonschema:"title=Allowed Periods,default=30" jsonschema_description:"List of periods the user is allowed to select in addition to the default."`

	DisableReuseSecurityPolicy bool `koanf:"disable_reuse_security_policy" yaml:"disable_reuse_security_policy" toml:"disable_reuse_security_policy" json:"disable_reuse_security_policy" jsonschema:"title=Disable Reuse Security Policy,default=false" jsonschema_description:"Disables the security policy that prevents reuse of a TOTP code."`

	Apps TOTPApps `koanf:"apps" yaml:"apps,omitempty" toml:"apps,omitempty" json:"apps,omitempty" jsonschema:"title=Apps" jsonschema_description:"The authenticator applications suggested to users when they register a TOTP credential."`
}

// TOTPApps represents the store links suggested to users registering a TOTP credential.
type TOTPApps struct {
	AppleStore TOTPAppsStore `koanf:"apple_store" yaml:"apple_store,omitempty" toml:"apple_store,omitempty" json:"apple_store,omitempty" jsonschema:"title=Apple App Store" jsonschema_description:"The Apple App Store listing for the authenticator application suggested to users."`
	GooglePlay TOTPAppsStore `koanf:"google_play" yaml:"google_play,omitempty" toml:"google_play,omitempty" json:"google_play,omitempty" jsonschema:"title=Google Play" jsonschema_description:"The Google Play listing for the authenticator application suggested to users."`
}

// TOTPAppsStore represents an individual store listing suggested to users registering a TOTP credential.
type TOTPAppsStore struct {
	Disable bool    `koanf:"disable" yaml:"disable" toml:"disable" json:"disable" jsonschema:"title=Disable,default=false" jsonschema_description:"Disables the badge for this store."`
	URL     url.URL `koanf:"url" yaml:"url,omitempty" toml:"url,omitempty" json:"url,omitempty" jsonschema:"title=URL" jsonschema_description:"The listing for the authenticator application suggested to users."`
}

var defaultTOTPSkew = 1

// DefaultTOTPApps represents the authenticator application suggested to users when none is configured.
var DefaultTOTPApps = TOTPApps{
	AppleStore: TOTPAppsStore{URL: url.URL{Scheme: "https", Host: "apps.apple.com", Path: "/us/app/google-authenticator/id388497605"}},
	GooglePlay: TOTPAppsStore{URL: url.URL{Scheme: "https", Host: "play.google.com", Path: "/store/apps/details", RawQuery: "id=com.google.android.apps.authenticator2"}},
}

// DefaultTOTPConfiguration represents default configuration parameters for TOTP generation.
var DefaultTOTPConfiguration = TOTP{
	Issuer:            "Authelia",
	DefaultAlgorithm:  TOTPAlgorithmSHA1,
	DefaultDigits:     6,
	DefaultPeriod:     30,
	Skew:              &defaultTOTPSkew,
	SecretSize:        TOTPSecretSizeDefault,
	AllowedAlgorithms: []string{TOTPAlgorithmSHA1},
	AllowedDigits:     []int{6},
	AllowedPeriods:    []int{30},
	Apps:              DefaultTOTPApps,
}
