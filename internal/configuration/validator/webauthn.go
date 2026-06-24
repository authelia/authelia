package validator

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
	"github.com/authelia/authelia/v4/internal/webauthn"
)

// ValidateWebAuthn validates and updates WebAuthn configuration.
//
//nolint:gocyclo
func ValidateWebAuthn(config *schema.Configuration, validator *schema.StructValidator) {
	if config.WebAuthn.DisplayName == "" {
		config.WebAuthn.DisplayName = schema.DefaultWebAuthnConfiguration.DisplayName
	}

	if config.WebAuthn.Timeout <= 0 {
		config.WebAuthn.Timeout = schema.DefaultWebAuthnConfiguration.Timeout
	}

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

	switch {
	case config.WebAuthn.ConveyancePreference == "":
		config.WebAuthn.ConveyancePreference = schema.DefaultWebAuthnConfiguration.ConveyancePreference
	case !utils.IsStringInSlice(string(config.WebAuthn.ConveyancePreference), validWebAuthnConveyancePreferences):
		validator.Push(fmt.Errorf(errFmtWebAuthnConveyancePreference, utils.StringJoinOr(validWebAuthnConveyancePreferences), config.WebAuthn.ConveyancePreference))
	}

	if config.WebAuthn.SelectionCriteria.Attachment != "" && !utils.IsStringInSlice(string(config.WebAuthn.SelectionCriteria.Attachment), validWebAuthnAttachment) {
		validator.Push(fmt.Errorf(errFmtWebAuthnSelectionCriteria, "attachment", utils.StringJoinOr(validWebAuthnAttachment), config.WebAuthn.SelectionCriteria.Attachment))
	}

	if config.WebAuthn.SelectionCriteria.Discoverability != "" && !utils.IsStringInSlice(string(config.WebAuthn.SelectionCriteria.Discoverability), validWebAuthnDiscoverability) {
		validator.Push(fmt.Errorf(errFmtWebAuthnSelectionCriteria, "discoverability", utils.StringJoinOr(validWebAuthnDiscoverability), config.WebAuthn.SelectionCriteria.Discoverability))
	}

	if config.WebAuthn.SelectionCriteria.UserVerification != "" && !utils.IsStringInSlice(string(config.WebAuthn.SelectionCriteria.UserVerification), validWebAuthnUserVerificationRequirement) {
		validator.Push(fmt.Errorf(errFmtWebAuthnSelectionCriteria, "user_verification", utils.StringJoinOr(validWebAuthnUserVerificationRequirement), config.WebAuthn.SelectionCriteria.UserVerification))
	}

	if config.WebAuthn.EnablePasskeyLogin && config.WebAuthn.SelectionCriteria.Discoverability == protocol.ResidentKeyRequirementDiscouraged {
		validator.PushWarning(fmt.Errorf(errFmtWebAuthnPasskeyDiscoverability, protocol.ResidentKeyRequirementPreferred, protocol.ResidentKeyRequirementRequired))
	}

	if len(config.WebAuthn.Filtering.PermittedAAGUIDs) != 0 && len(config.WebAuthn.Filtering.ProhibitedAAGUIDs) != 0 {
		validator.Push(errors.New(errFmtWebAuthnFiltering))
	}

	validateWebAuthnRelatedOrigins(config, validator)
}

func validateWebAuthnRelatedOrigins(config *schema.Configuration, validator *schema.StructValidator) {
	n := len(config.WebAuthn.RelatedOrigins)

	if n == 0 {
		return
	}

	origins := map[string][]string{}

	for relyingPartyID, relatedOrigin := range config.WebAuthn.RelatedOrigins {
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
