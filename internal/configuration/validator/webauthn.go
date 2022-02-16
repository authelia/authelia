package validator

import (
	"fmt"

	"github.com/duo-labs/webauthn/protocol"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateWebauthn validates and update Webauthn configuration.
func ValidateWebauthn(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.Webauthn.DisplayName == "" {
		configuration.Webauthn.DisplayName = schema.DefaultWebauthnConfiguration.DisplayName
	}

	if configuration.Webauthn.Timeout == 0 {
		configuration.Webauthn.Timeout = schema.DefaultWebauthnConfiguration.Timeout
	}

	switch configuration.Webauthn.ConveyancePreference {
	case protocol.PreferNoAttestation, protocol.PreferIndirectAttestation, protocol.PreferDirectAttestation:
		break
	case "":
		configuration.Webauthn.ConveyancePreference = schema.DefaultWebauthnConfiguration.ConveyancePreference
	default:
		validator.Push(fmt.Errorf(errFmtWebauthnConveyancePreference, configuration.Webauthn.ConveyancePreference))
	}

	switch configuration.Webauthn.UserVerification {
	case protocol.VerificationDiscouraged, protocol.VerificationPreferred, protocol.VerificationRequired:
		break
	case "":
		configuration.Webauthn.UserVerification = schema.DefaultWebauthnConfiguration.UserVerification
	default:
		validator.Push(fmt.Errorf(errFmtWebauthnUserVerification, configuration.Webauthn.UserVerification))
	}
}
