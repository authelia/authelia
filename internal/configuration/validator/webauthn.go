// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

// ValidateWebAuthn validates and updates WebAuthn configuration.
func ValidateWebAuthn(config *schema.Configuration, validator *schema.StructValidator) {
	if config.WebAuthn.EnablePasskeyLogin {
		if config.WebAuthn.Disable {
			validator.Push(fmt.Errorf(errFmtWebAuthnBoolean, "enable_passkey_login", config.WebAuthn.EnablePasskeyLogin, false, "disable", config.WebAuthn.Disable))
		}
	} else {
		if config.WebAuthn.EnablePasskey2FA {
			validator.Push(fmt.Errorf(errFmtWebAuthnBoolean, "experimental_enable_passkey_uv_two_factors", config.WebAuthn.EnablePasskey2FA, false, "enable_passkey_login", config.WebAuthn.EnablePasskeyLogin))
		}

		if config.WebAuthn.EnablePasskeyUpgrade {
			validator.Push(fmt.Errorf(errFmtWebAuthnBoolean, "experimental_enable_passkey_upgrade", config.WebAuthn.EnablePasskeyUpgrade, false, "enable_passkey_login", config.WebAuthn.EnablePasskeyLogin))
		}
	}

	if config.WebAuthn.Metadata.Enabled {
		switch config.WebAuthn.Metadata.CachePolicy {
		case webauthn.CachePolicyStrict, webauthn.CachePolicyRelaxed:
			break
		default:
			validator.Push(fmt.Errorf(errFmtWebAuthnMetadataString, "cache_policy", config.WebAuthn.Metadata.CachePolicy, utils.StringJoinOr([]string{webauthn.CachePolicyStrict, webauthn.CachePolicyRelaxed})))
		}
	}

	validateWebAuthnRelyingPartyBase(&config.WebAuthn.WebAuthnBase, &schema.DefaultWebAuthnConfiguration.WebAuthnBase, "", config.WebAuthn.EnablePasskeyLogin, validator)

	validateWebAuthnRelyingParties(config, validator)

	validateWebAuthnRelatedOrigins(config, validator)
}

//nolint:gocyclo
func validateWebAuthnRelyingPartyBase(base, defaults *schema.WebAuthnBase, prefix string, passkeys bool, validator *schema.StructValidator) {
	if base.DisplayName == "" {
		base.DisplayName = defaults.DisplayName
	}

	if base.Timeout <= 0 {
		base.Timeout = defaults.Timeout
	}

	switch {
	case base.ConveyancePreference == "":
		base.ConveyancePreference = defaults.ConveyancePreference
	case !utils.IsStringInSlice(string(base.ConveyancePreference), validWebAuthnConveyancePreferences):
		validator.Push(fmt.Errorf(errFmtWebAuthnConveyancePreference, prefix, utils.StringJoinOr(validWebAuthnConveyancePreferences), base.ConveyancePreference))
	}

	switch attachment := base.SelectionCriteria.GetAttachment(); {
	case base.SelectionCriteria.Attachment == nil:
		base.SelectionCriteria.Attachment = defaults.SelectionCriteria.Attachment
	case attachment != "" && !utils.IsStringInSlice(string(attachment), validWebAuthnAttachment):
		validator.Push(fmt.Errorf(errFmtWebAuthnSelectionCriteria, prefix, "attachment", utils.StringJoinOr(validWebAuthnAttachment), attachment))
	}

	switch {
	case base.SelectionCriteria.Discoverability == "":
		base.SelectionCriteria.Discoverability = defaults.SelectionCriteria.Discoverability
	case !utils.IsStringInSlice(string(base.SelectionCriteria.Discoverability), validWebAuthnDiscoverability):
		validator.Push(fmt.Errorf(errFmtWebAuthnSelectionCriteria, prefix, "discoverability", utils.StringJoinOr(validWebAuthnDiscoverability), base.SelectionCriteria.Discoverability))
	}

	switch {
	case base.SelectionCriteria.UserVerification == "":
		base.SelectionCriteria.UserVerification = defaults.SelectionCriteria.UserVerification
	case !utils.IsStringInSlice(string(base.SelectionCriteria.UserVerification), validWebAuthnUserVerificationRequirement):
		validator.Push(fmt.Errorf(errFmtWebAuthnSelectionCriteria, prefix, "user_verification", utils.StringJoinOr(validWebAuthnUserVerificationRequirement), base.SelectionCriteria.UserVerification))
	}

	if passkeys && base.SelectionCriteria.Discoverability == protocol.ResidentKeyRequirementDiscouraged {
		validator.PushWarning(fmt.Errorf(errFmtWebAuthnPasskeyDiscoverability, prefix, protocol.ResidentKeyRequirementPreferred, protocol.ResidentKeyRequirementRequired))
	}

	if base.Filtering.ProhibitBackupEligibility == nil {
		base.Filtering.ProhibitBackupEligibility = defaults.Filtering.ProhibitBackupEligibility
	}

	if base.Filtering.PermittedAAGUIDs == nil && base.Filtering.ProhibitedAAGUIDs == nil {
		base.Filtering.PermittedAAGUIDs = defaults.Filtering.PermittedAAGUIDs
		base.Filtering.ProhibitedAAGUIDs = defaults.Filtering.ProhibitedAAGUIDs
	}

	if len(base.Filtering.PermittedAAGUIDs) != 0 && len(base.Filtering.ProhibitedAAGUIDs) != 0 {
		validator.Push(fmt.Errorf(errFmtWebAuthnFiltering, prefix))
	}

	switch {
	case base.ExtensionsUnsolicitedOutputPolicy == "":
		base.ExtensionsUnsolicitedOutputPolicy = defaults.ExtensionsUnsolicitedOutputPolicy
	case !utils.IsStringInSlice(base.ExtensionsUnsolicitedOutputPolicy, validWebAuthnExtensionsUnsolicitedOutputPolicies):
		validator.Push(fmt.Errorf(errFmtWebAuthnUnsolicitedOutputPolicy, prefix, utils.StringJoinOr(validWebAuthnExtensionsUnsolicitedOutputPolicies), base.ExtensionsUnsolicitedOutputPolicy))
	}
}

func validateWebAuthnRelyingParties(config *schema.Configuration, validator *schema.StructValidator) {
	for _, relyingPartyID := range slices.Sorted(maps.Keys(config.WebAuthn.RelyingParties)) {
		relyingParty := config.WebAuthn.RelyingParties[relyingPartyID]

		prefix := fmt.Sprintf(errFmtWebAuthnRelyingPartyPrefix, relyingPartyID)

		validateWebAuthnRelyingPartyBase(&relyingParty.WebAuthnBase, &config.WebAuthn.WebAuthnBase, prefix, config.WebAuthn.EnablePasskeyLogin, validator)

		validateWebAuthnRelyingPartyOpaqueOrigins(&relyingParty, prefix, validator)

		config.WebAuthn.RelyingParties[relyingPartyID] = relyingParty
	}
}

func validateWebAuthnRelyingPartyOpaqueOrigins(relyingParty *schema.WebAuthnRelyingParty, prefix string, validator *schema.StructValidator) {
	for i, origin := range relyingParty.OpaqueOrigins {
		switch {
		case origin == "":
			validator.Push(fmt.Errorf(errFmtWebAuthnRelyingPartyOpaqueOriginEmpty, prefix, i+1))
		case !protocol.IsOpaqueOrigin(origin):
			validator.Push(fmt.Errorf(errFmtWebAuthnRelyingPartyOpaqueOriginNotOpaque, prefix, i+1, origin))
		case !protocol.IsKnownOpaqueOrigin(origin):
			validator.Push(fmt.Errorf(errFmtWebAuthnRelyingPartyOpaqueOriginUnknown, prefix, i+1, origin, utils.StringJoinOr(protocol.OpaqueOriginPrefixes())))
		}
	}
}

func validateWebAuthnRelatedOrigins(config *schema.Configuration, validator *schema.StructValidator) {
	if len(config.WebAuthn.RelyingParties) == 0 {
		return
	}

	origins := map[string][]string{}

	for _, relyingPartyID := range slices.Sorted(maps.Keys(config.WebAuthn.RelyingParties)) {
		relyingParty := config.WebAuthn.RelyingParties[relyingPartyID]

		validateWebAuthnRelatedOriginsRelyingParty(config, relyingPartyID, &relyingParty, origins, validator)
	}

	for _, origin := range slices.Sorted(maps.Keys(origins)) {
		if len(origins[origin]) < 2 {
			continue
		}

		validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginDuplicate, origin, utils.StringJoinAnd(origins[origin])))
	}
}

func validateWebAuthnRelatedOriginsRelyingParty(config *schema.Configuration, relyingPartyID string, relyingParty *schema.WebAuthnRelyingParty, origins map[string][]string, validator *schema.StructValidator) {
	if relyingPartyID == "" {
		validator.Push(errors.New(errFmtWebAuthnRelatedOriginsRelyingPartyIDEmpty))

		return
	}

	prefix := fmt.Sprintf(errFmtWebAuthnRelyingPartyPrefix, relyingPartyID)

	validateWebAuthnRelatedOriginsRelyingPartyID(relyingPartyID, prefix, validator)

	if len(relyingParty.Origins) == 0 {
		validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginsEmpty, prefix))

		return
	}

	var (
		found     bool
		relatable []string
	)

	seen := map[string]bool{}

	for i, origin := range relyingParty.Origins {
		if origin == nil {
			validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginEmpty, prefix, i+1))

			continue
		}

		value := origin.String()

		// Origins are keyed by the same values webauthn.IsOriginEqual compares so equivalent origins which differ only
		// in case or an explicit default port are detected as duplicates.
		key := webauthn.OriginKey(origin)

		if seen[key] {
			validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginDuplicateSelf, prefix, value))

			continue
		}

		seen[key] = true

		origins[key] = append(origins[key], relyingPartyID)

		found = found || strings.EqualFold(origin.Hostname(), relyingPartyID)

		if origin.Path != "" {
			validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginNotValidPath, prefix, i+1, value))

			continue
		}

		if protocol.IsOpaqueOrigin(value) {
			validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginNotRelatable, prefix, i+1, value, utils.StringJoinOr([]string{schemeHTTP, schemeHTTPS})))

			continue
		}

		relatable = append(relatable, value)

		if !originMatchesCookieAutheliaURL(config, origin) {
			validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginNotSessionCookie, prefix, i+1, value))
		}
	}

	if !found {
		validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsRelyingPartyNoOrigin, prefix))
	}

	// The document served at the well known endpoint is built from these origins, so the limits a client applies when
	// it reads that document have to hold here or the endpoint fails at request time instead.
	if len(relatable) != 0 {
		if _, err := protocol.NewRelatedOriginsWithLabeler(webauthn.RelatedOriginLabeler, relatable...); err != nil {
			validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsLabels, prefix, err))
		} else if len(relyingParty.OpaqueOrigins) != 0 {
			// The provider validates the origins with the default labeler when opaque origins are configured, which
			// counts origins under a multi-label public suffix by the second last label, so that count has to hold too.
			if _, err = protocol.NewRelatedOrigins(relatable...); err != nil {
				validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsLabelsOpaque, prefix, err))
			}
		}
	}
}

func validateWebAuthnRelatedOriginsRelyingPartyID(relyingPartyID, prefix string, validator *schema.StructValidator) {
	if relyingPartyID != strings.ToLower(relyingPartyID) {
		validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsRelyingPartyNotLowerCase, prefix))
	}

	if err := protocol.ValidateRPID(relyingPartyID); err != nil {
		validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsRelyingPartyNotValid, prefix, err))
	}
}

func originMatchesCookieAutheliaURL(config *schema.Configuration, origin *url.URL) (match bool) {
	for _, domain := range config.Session.Cookies {
		if domain.AutheliaURL == nil {
			continue
		}

		if webauthn.IsOriginEqual(domain.AutheliaURL, origin) {
			return true
		}
	}

	return false
}
