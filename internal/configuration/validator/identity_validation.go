package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateIdentityValidation validates and updates the IdentityValidation configuration.
func ValidateIdentityValidation(config *schema.Configuration, validator *schema.StructValidator) {
	if config.IdentityValidation.ResetPassword.Expiration <= 0 {
		config.IdentityValidation.ResetPassword.Expiration = schema.DefaultIdentityValidation.ResetPassword.Expiration
	}

	switch {
	case len(config.IdentityValidation.ResetPassword.JWTAlgorithm) == 0:
		config.IdentityValidation.ResetPassword.JWTAlgorithm = schema.DefaultIdentityValidation.ResetPassword.JWTAlgorithm
	case !utils.IsStringInSlice(config.IdentityValidation.ResetPassword.JWTAlgorithm, validIdentityValidationJWTAlgorithms):
		validator.Push(fmt.Errorf(errFmtIdentityValidationResetPasswordJWTAlgorithm, strJoinOr(validIdentityValidationJWTAlgorithms), config.IdentityValidation.ResetPassword.JWTAlgorithm))
	}

	if !config.AuthenticationBackend.PasswordReset.Disable && len(config.IdentityValidation.ResetPassword.JWTSecret) == 0 {
		validator.Push(fmt.Errorf(errFmtIdentityValidationResetPasswordJWTSecret))
	}

	if config.IdentityValidation.ElevatedSession.Expiration <= 0 {
		config.IdentityValidation.ElevatedSession.Expiration = schema.DefaultIdentityValidation.ElevatedSession.Expiration
	}

	if config.IdentityValidation.ElevatedSession.ElevationExpiration <= 0 {
		config.IdentityValidation.ElevatedSession.ElevationExpiration = schema.DefaultIdentityValidation.ElevatedSession.ElevationExpiration
	}

	if config.IdentityValidation.ElevatedSession.Characters <= 0 {
		config.IdentityValidation.ElevatedSession.Characters = schema.DefaultIdentityValidation.ElevatedSession.Characters
	} else if config.IdentityValidation.ElevatedSession.Characters > 20 {
		validator.Push(fmt.Errorf(errFmtIdentityValidationElevatedSessionCharacterLength, config.IdentityValidation.ElevatedSession.Characters))
	}
}
