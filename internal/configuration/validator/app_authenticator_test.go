// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newAppAuthenticatorConfig() *schema.Configuration {
	return &schema.Configuration{}
}

func TestValidateAppAuthenticatorShouldSetDefaults(t *testing.T) {
	config := newAppAuthenticatorConfig()
	v := schema.NewStructValidator()

	ValidateAppAuthenticator(config, v)

	assert.Len(t, v.Errors(), 0)
	assert.False(t, config.AppAuthenticator.Disable)
	assert.Equal(t, "Authelia", config.AppAuthenticator.DisplayName)
	assert.Equal(t, 5*time.Minute, config.AppAuthenticator.RegistrationTokenLifespan)
	assert.Equal(t, 2*time.Minute, config.AppAuthenticator.ChallengeLifespan)
	assert.Equal(t, 2, config.AppAuthenticator.NonceLength)
	assert.Equal(t, schema.AppAuthenticatorAttestationModeRequired, config.AppAuthenticator.Attestation.Mode)
	assert.False(t, config.AppAuthenticator.Attestation.HasIdentities())
}

func TestValidateAppAuthenticatorShouldWarnWhenRequiredWithoutIdentities(t *testing.T) {
	config := newAppAuthenticatorConfig()
	v := schema.NewStructValidator()

	ValidateAppAuthenticator(config, v)

	require.Len(t, v.Warnings(), 1)
	assert.Len(t, v.Errors(), 0)
	assert.EqualError(t, v.Warnings()[0], "app_authenticator: attestation: no trusted app identities are configured; app registration is unavailable until 'app_authenticator.attestation.apple' or 'app_authenticator.attestation.android' is set, or 'mode' is set to 'preferred' or 'none'")
}

func TestValidateAppAuthenticatorShouldSkipWhenDisabled(t *testing.T) {
	config := newAppAuthenticatorConfig()
	config.AppAuthenticator.Disable = true
	config.AppAuthenticator.NonceLength = 9
	config.AppAuthenticator.Attestation.Mode = "foo"

	v := schema.NewStructValidator()

	ValidateAppAuthenticator(config, v)

	assert.Len(t, v.Errors(), 0)
	assert.Len(t, v.Warnings(), 0)
	assert.Equal(t, 9, config.AppAuthenticator.NonceLength)
}

func TestValidateAppAuthenticatorNonceLength(t *testing.T) {
	testCases := []struct {
		length int
		err    string
	}{
		{1, "app_authenticator: option 'nonce_length' must be between 2 and 4 but it's configured as '1'"},
		{2, ""},
		{3, ""},
		{4, ""},
		{5, "app_authenticator: option 'nonce_length' must be between 2 and 4 but it's configured as '5'"},
	}

	for _, tc := range testCases {
		t.Run(strconv.Itoa(tc.length), func(t *testing.T) {
			config := newAppAuthenticatorConfig()
			config.AppAuthenticator.NonceLength = tc.length
			v := schema.NewStructValidator()

			ValidateAppAuthenticator(config, v)

			if tc.err == "" {
				assert.Len(t, v.Errors(), 0)
			} else {
				require.Len(t, v.Errors(), 1)
				assert.EqualError(t, v.Errors()[0], tc.err)
			}
		})
	}
}

func TestValidateAppAuthenticatorLifespans(t *testing.T) {
	config := newAppAuthenticatorConfig()
	config.AppAuthenticator.RegistrationTokenLifespan = -time.Minute
	config.AppAuthenticator.ChallengeLifespan = -time.Minute
	v := schema.NewStructValidator()

	ValidateAppAuthenticator(config, v)

	require.Len(t, v.Errors(), 2)
	assert.EqualError(t, v.Errors()[0], "app_authenticator: option 'registration_token_lifespan' must be greater than 0 but it's configured as '-1m0s'")
	assert.EqualError(t, v.Errors()[1], "app_authenticator: option 'challenge_lifespan' must be greater than 0 but it's configured as '-1m0s'")
}

func TestValidateAppAuthenticatorShouldWarnOnLongChallengeLifespan(t *testing.T) {
	config := newAppAuthenticatorConfig()
	config.AppAuthenticator.ChallengeLifespan = 6 * time.Minute
	v := schema.NewStructValidator()

	ValidateAppAuthenticator(config, v)

	assert.Len(t, v.Errors(), 0)

	var found bool

	for _, w := range v.Warnings() {
		if strings.Contains(w.Error(), "challenge_lifespan") {
			found = true
		}
	}

	assert.True(t, found)
}

func TestValidateAppAuthenticatorMode(t *testing.T) {
	config := newAppAuthenticatorConfig()
	config.AppAuthenticator.Attestation.Mode = "foo"
	v := schema.NewStructValidator()

	ValidateAppAuthenticator(config, v)

	require.Len(t, v.Errors(), 1)
	assert.EqualError(t, v.Errors()[0], "app_authenticator: attestation: option 'mode' must be one of 'required', 'preferred', or 'none' but it's configured as 'foo'")

	config = newAppAuthenticatorConfig()
	config.AppAuthenticator.Attestation.Mode = schema.AppAuthenticatorAttestationModeNone
	v = schema.NewStructValidator()

	ValidateAppAuthenticator(config, v)

	assert.Len(t, v.Errors(), 0)
	require.Len(t, v.Warnings(), 1)
	assert.Contains(t, v.Warnings()[0].Error(), "'none'")

	config = newAppAuthenticatorConfig()
	config.AppAuthenticator.Attestation.Mode = schema.AppAuthenticatorAttestationModePreferred
	v = schema.NewStructValidator()

	ValidateAppAuthenticator(config, v)

	assert.Len(t, v.Errors(), 0)
}

func TestValidateAppAuthenticatorApple(t *testing.T) {
	testCases := []struct {
		name     string
		identity schema.AppAuthenticatorAppleIdentity
		env      string
		err      string
	}{
		{"ShouldDefaultEnvironment", schema.AppAuthenticatorAppleIdentity{TeamID: "ABCDE12345", BundleID: "com.example.app"}, "production", ""},
		{"ShouldAllowDevelopment", schema.AppAuthenticatorAppleIdentity{TeamID: "ABCDE12345", BundleID: "com.example.app", Environment: "development"}, "development", ""},
		{"ShouldRejectBadTeam", schema.AppAuthenticatorAppleIdentity{TeamID: "abcde12345", BundleID: "com.example.app"}, "", "app_authenticator: attestation: apple: identity 1: option 'team_id' must match the pattern '^[A-Z0-9]{10}$' but it's configured as 'abcde12345'"},
		{"ShouldRejectMissingBundle", schema.AppAuthenticatorAppleIdentity{TeamID: "ABCDE12345"}, "", "app_authenticator: attestation: apple: identity 1: option 'bundle_id' is required"},
		{"ShouldRejectBadEnvironment", schema.AppAuthenticatorAppleIdentity{TeamID: "ABCDE12345", BundleID: "com.example.app", Environment: "staging"}, "", "app_authenticator: attestation: apple: identity 1: option 'environment' must be one of 'production' or 'development' but it's configured as 'staging'"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := newAppAuthenticatorConfig()
			config.AppAuthenticator.Attestation.Apple = []schema.AppAuthenticatorAppleIdentity{tc.identity}
			v := schema.NewStructValidator()

			ValidateAppAuthenticator(config, v)

			if tc.err == "" {
				assert.Len(t, v.Errors(), 0)
				assert.Len(t, v.Warnings(), 0)
				assert.Equal(t, tc.env, config.AppAuthenticator.Attestation.Apple[0].Environment)
			} else {
				require.Len(t, v.Errors(), 1)
				assert.EqualError(t, v.Errors()[0], tc.err)
			}
		})
	}
}

func TestValidateAppAuthenticatorAndroid(t *testing.T) {
	hex64 := strings.Repeat("ab", 32)
	colon := strings.TrimSuffix(strings.Repeat("AB:", 32), ":")

	testCases := []struct {
		name     string
		identity schema.AppAuthenticatorAndroidIdentity
		expected []string
		err      string
	}{
		{"ShouldAcceptHex", schema.AppAuthenticatorAndroidIdentity{PackageName: "com.example.app", SigningCertificateDigests: []string{strings.ToUpper(hex64)}}, []string{hex64}, ""},
		{"ShouldAcceptColon", schema.AppAuthenticatorAndroidIdentity{PackageName: "com.example.app", SigningCertificateDigests: []string{colon}}, []string{hex64}, ""},
		{"ShouldRejectMissingPackage", schema.AppAuthenticatorAndroidIdentity{SigningCertificateDigests: []string{hex64}}, nil, "app_authenticator: attestation: android: identity 1: option 'package_name' is required"},
		{"ShouldRejectNoDigests", schema.AppAuthenticatorAndroidIdentity{PackageName: "com.example.app"}, nil, "app_authenticator: attestation: android: identity 1: option 'signing_certificate_digests' must have at least one value"},
		{"ShouldRejectShortDigest", schema.AppAuthenticatorAndroidIdentity{PackageName: "com.example.app", SigningCertificateDigests: []string{"abcd"}}, nil, "app_authenticator: attestation: android: identity 1: option 'signing_certificate_digests' value 1 must be 32 bytes as 64 hex characters or 32 colon separated hex pairs"},
		{"ShouldRejectNonHex", schema.AppAuthenticatorAndroidIdentity{PackageName: "com.example.app", SigningCertificateDigests: []string{strings.Repeat("zz", 32)}}, nil, "app_authenticator: attestation: android: identity 1: option 'signing_certificate_digests' value 1 must be 32 bytes as 64 hex characters or 32 colon separated hex pairs"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := newAppAuthenticatorConfig()
			config.AppAuthenticator.Attestation.Android = []schema.AppAuthenticatorAndroidIdentity{tc.identity}
			v := schema.NewStructValidator()

			ValidateAppAuthenticator(config, v)

			if tc.err == "" {
				assert.Len(t, v.Errors(), 0)
				assert.Equal(t, tc.expected, config.AppAuthenticator.Attestation.Android[0].SigningCertificateDigests)
			} else {
				require.Len(t, v.Errors(), 1)
				assert.EqualError(t, v.Errors()[0], tc.err)
			}
		})
	}
}
