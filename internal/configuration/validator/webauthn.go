package validator

import (
	"errors"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateWebAuthn validates and update WebAuthn configuration.
func ValidateWebAuthn(config *schema.Configuration, validator *schema.StructValidator) {
	if config.WebAuthn.DisplayName == "" {
		config.WebAuthn.DisplayName = schema.DefaultWebAuthnConfiguration.DisplayName
	}

	if config.WebAuthn.Timeout <= 0 {
		config.WebAuthn.Timeout = schema.DefaultWebAuthnConfiguration.Timeout
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
}
