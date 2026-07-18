package validator

import (
	"fmt"
	"net/url"
	"sort"
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

// validateWebAuthnRelyingPartyBase validates the options common to both the global WebAuthn configuration and each
// individual relying party, defaulting any option which has no value to the equivalent option in defaults. The prefix
// is included in every error to indicate which relying party the error belongs to and is empty for the global
// configuration.
//
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

	switch {
	case base.SelectionCriteria.Attachment == "":
		base.SelectionCriteria.Attachment = defaults.SelectionCriteria.Attachment
	case !utils.IsStringInSlice(string(base.SelectionCriteria.Attachment), validWebAuthnAttachment):
		validator.Push(fmt.Errorf(errFmtWebAuthnSelectionCriteria, prefix, "attachment", utils.StringJoinOr(validWebAuthnAttachment), base.SelectionCriteria.Attachment))
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

	if !base.Filtering.ProhibitBackupEligibility {
		base.Filtering.ProhibitBackupEligibility = defaults.Filtering.ProhibitBackupEligibility
	}

	if len(base.Filtering.PermittedAAGUIDs) == 0 && len(base.Filtering.ProhibitedAAGUIDs) == 0 {
		base.Filtering.PermittedAAGUIDs = defaults.Filtering.PermittedAAGUIDs
		base.Filtering.ProhibitedAAGUIDs = defaults.Filtering.ProhibitedAAGUIDs
	}

	if len(base.Filtering.PermittedAAGUIDs) != 0 && len(base.Filtering.ProhibitedAAGUIDs) != 0 {
		validator.Push(fmt.Errorf(errFmtWebAuthnFiltering, prefix))
	}
}

// validateWebAuthnRelyingParties validates each relying party, defaulting the options they share with the global
// WebAuthn configuration to the value of the global configuration.
func validateWebAuthnRelyingParties(config *schema.Configuration, validator *schema.StructValidator) {
	if len(config.WebAuthn.RelyingParties) == 0 {
		return
	}

	relyingPartyIDs := make([]string, 0, len(config.WebAuthn.RelyingParties))

	for relyingPartyID := range config.WebAuthn.RelyingParties {
		relyingPartyIDs = append(relyingPartyIDs, relyingPartyID)
	}

	sort.Strings(relyingPartyIDs)

	for _, relyingPartyID := range relyingPartyIDs {
		relyingParty := config.WebAuthn.RelyingParties[relyingPartyID]

		prefix := fmt.Sprintf(errFmtWebAuthnRelyingPartyPrefix, relyingPartyID)

		validateWebAuthnRelyingPartyBase(&relyingParty.WebAuthnBase, &config.WebAuthn.WebAuthnBase, prefix, config.WebAuthn.EnablePasskeyLogin, validator)

		validateWebAuthnRelyingPartyOpaqueOrigins(&relyingParty, prefix, validator)

		config.WebAuthn.RelyingParties[relyingPartyID] = relyingParty
	}
}

func validateWebAuthnRelyingPartyOpaqueOrigins(relyingParty *schema.WebAuthnRelyingParty, prefix string, validator *schema.StructValidator) {
	for i, origin := range relyingParty.OpaqueOrigins {
		if origin == "" {
			validator.Push(fmt.Errorf(errFmtWebAuthnRelyingPartyOpaqueOriginEmpty, prefix, i+1))

			continue
		}

		for _, scheme := range invalidWebAuthnOpaqueOriginSchemes {
			if utils.StringHasPrefixFold(origin, scheme) {
				validator.Push(fmt.Errorf(errFmtWebAuthnRelyingPartyOpaqueOriginScheme, prefix, i+1, origin, scheme))

				break
			}
		}
	}
}

func validateWebAuthnRelatedOrigins(config *schema.Configuration, validator *schema.StructValidator) {
	n := len(config.WebAuthn.RelyingParties)

	if n == 0 {
		return
	}

	origins := map[string][]string{}

	for relyingPartyID, relatedOrigin := range config.WebAuthn.RelyingParties {
		if relyingPartyID == "" {
			validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOptionEmpty, relyingPartyID, "relying_party_id"))

			continue
		}

		if relyingPartyID != strings.ToLower(relyingPartyID) {
			validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsRelyingPartyNotLowerCase, relyingPartyID))
		}

		found := false

		for i, origin := range relatedOrigin.Origins {
			if origin == nil {
				validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginEmpty, relyingPartyID, i+1))
				continue
			}

			var (
				values []string
				ok     bool
			)

			strOrigin := origin.String()

			if values, ok = origins[strOrigin]; !ok {
				values = []string{strOrigin}
			} else {
				values = append(values, strOrigin)
			}

			origins[strOrigin] = values

			if origin.Path != "" {
				validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginNotValidPath, relyingPartyID, i+1, origin.String()))
			}

			if !found && strings.EqualFold(origin.Hostname(), relyingPartyID) {
				found = true
			}

			if !originMatchesCookieAutheliaURL(config, origin) {
				validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginNotSessionCookie, relyingPartyID, i+1, origin.String()))
			}
		}

		for origin, values := range origins {
			if len(values) == 1 {
				continue
			}

			validator.Push(fmt.Errorf(errFmtWebAuthnRelatedOriginsOriginDuplicate, origin, utils.StringJoinAnd(values)))
		}

		if !found {
			validator.Push(fmt.Errorf("error rpid %s does not match any origin", relyingPartyID))
		}
	}
}

func originMatchesCookieAutheliaURL(config *schema.Configuration, origin *url.URL) (match bool) {
	for _, domain := range config.Session.Cookies {
		if domain.AutheliaURL == nil {
			continue
		}

		if domain.AutheliaURL.Hostname() == origin.Hostname() {
			return true
		}
	}

	return false
}
