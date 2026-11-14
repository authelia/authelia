// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

var (
	regexpAppleTeamID = regexp.MustCompile(`^[A-Z0-9]{10}$`)

	validAppAuthenticatorModes = []string{schema.AppAuthenticatorAttestationModeRequired, schema.AppAuthenticatorAttestationModePreferred, schema.AppAuthenticatorAttestationModeNone}
)

// ValidateAppAuthenticator validates and updates the app authenticator configuration.
func ValidateAppAuthenticator(config *schema.Configuration, validator *schema.StructValidator) {
	c := &config.AppAuthenticator

	if c.Disable {
		return
	}

	defaults := schema.DefaultAppAuthenticatorConfiguration

	if c.DisplayName == "" {
		c.DisplayName = defaults.DisplayName
	}

	switch {
	case c.RegistrationTokenLifespan == 0:
		c.RegistrationTokenLifespan = defaults.RegistrationTokenLifespan
	case c.RegistrationTokenLifespan < 0:
		validator.Push(fmt.Errorf(errFmtAppAuthenticatorLifespan, "registration_token_lifespan", c.RegistrationTokenLifespan))
	}

	switch {
	case c.ChallengeLifespan == 0:
		c.ChallengeLifespan = defaults.ChallengeLifespan
	case c.ChallengeLifespan < 0:
		validator.Push(fmt.Errorf(errFmtAppAuthenticatorLifespan, "challenge_lifespan", c.ChallengeLifespan))
	case c.ChallengeLifespan > 5*time.Minute:
		validator.PushWarning(fmt.Errorf(errFmtAppAuthenticatorChallengeLifespanLong, c.ChallengeLifespan))
	}

	switch {
	case c.NonceLength == 0:
		c.NonceLength = defaults.NonceLength
	case c.NonceLength < 2 || c.NonceLength > 4:
		validator.Push(fmt.Errorf(errFmtAppAuthenticatorNonceLength, c.NonceLength))
	}

	validateAppAuthenticatorAttestation(c, validator)
}

func validateAppAuthenticatorAttestation(c *schema.AppAuthenticator, validator *schema.StructValidator) {
	a := &c.Attestation

	if a.Mode == "" {
		a.Mode = schema.DefaultAppAuthenticatorConfiguration.Attestation.Mode
	}

	switch a.Mode {
	case schema.AppAuthenticatorAttestationModeRequired:
		if !a.HasIdentities() {
			validator.PushWarning(errors.New(errAppAuthenticatorNoIdentities))
		}
	case schema.AppAuthenticatorAttestationModePreferred:
	case schema.AppAuthenticatorAttestationModeNone:
		validator.PushWarning(errors.New(errAppAuthenticatorModeNone))
	default:
		validator.Push(fmt.Errorf(errFmtAppAuthenticatorMode, utils.StringJoinOr(validAppAuthenticatorModes), a.Mode))
	}

	for i := range a.Apple {
		validateAppAuthenticatorApple(i+1, &a.Apple[i], validator)
	}

	for i := range a.Android {
		validateAppAuthenticatorAndroid(i+1, &a.Android[i], validator)
	}
}

func validateAppAuthenticatorApple(n int, identity *schema.AppAuthenticatorAppleIdentity, validator *schema.StructValidator) {
	if !regexpAppleTeamID.MatchString(identity.TeamID) {
		validator.Push(fmt.Errorf(errFmtAppAuthenticatorAppleTeamID, n, identity.TeamID))
	}

	if identity.BundleID == "" {
		validator.Push(fmt.Errorf(errFmtAppAuthenticatorIdentityMissing, "apple", n, "bundle_id"))
	}

	switch identity.Environment {
	case "":
		identity.Environment = "production"
	case "production", "development":
	default:
		validator.Push(fmt.Errorf(errFmtAppAuthenticatorAppleEnvironment, n, identity.Environment))
	}
}

func validateAppAuthenticatorAndroid(n int, identity *schema.AppAuthenticatorAndroidIdentity, validator *schema.StructValidator) {
	if identity.PackageName == "" {
		validator.Push(fmt.Errorf(errFmtAppAuthenticatorIdentityMissing, "android", n, "package_name"))
	}

	if len(identity.SigningCertificateDigests) == 0 {
		validator.Push(fmt.Errorf(errFmtAppAuthenticatorAndroidNoDigests, n))

		return
	}

	for i, digest := range identity.SigningCertificateDigests {
		normalized, ok := normalizeAppAuthenticatorDigest(digest)
		if !ok {
			validator.Push(fmt.Errorf(errFmtAppAuthenticatorAndroidDigest, n, i+1))

			continue
		}

		identity.SigningCertificateDigests[i] = normalized
	}
}

func normalizeAppAuthenticatorDigest(digest string) (string, bool) {
	digest = strings.ToLower(strings.TrimSpace(digest))

	if strings.Contains(digest, ":") {
		parts := strings.Split(digest, ":")

		if len(parts) != 32 {
			return "", false
		}

		for _, part := range parts {
			if len(part) != 2 {
				return "", false
			}
		}

		digest = strings.Join(parts, "")
	}

	if len(digest) != 64 {
		return "", false
	}

	if _, err := hex.DecodeString(digest); err != nil {
		return "", false
	}

	return digest, true
}
