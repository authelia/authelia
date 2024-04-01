package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateIdentityValidation validates and updates the IdentityValidation configuration.
func ValidateIdentityValidation(config *schema.Configuration, validator *schema.StructValidator) {
	if config.IdentityValidation.ResetPassword.JWTExpiration <= 0 {
		config.IdentityValidation.ResetPassword.JWTExpiration = schema.DefaultIdentityValidation.ResetPassword.JWTExpiration
	}

	switch {
	case len(config.IdentityValidation.ResetPassword.JWTAlgorithm) == 0:
		config.IdentityValidation.ResetPassword.JWTAlgorithm = schema.DefaultIdentityValidation.ResetPassword.JWTAlgorithm
	case !utils.IsStringInSlice(config.IdentityValidation.ResetPassword.JWTAlgorithm, validIdentityValidationJWTAlgorithms):
		validator.Push(fmt.Errorf(errFmtIdentityValidationResetPasswordJWTAlgorithm, utils.StringJoinOr(validIdentityValidationJWTAlgorithms), config.IdentityValidation.ResetPassword.JWTAlgorithm))
	}

	if !config.AuthenticationBackend.PasswordReset.Disable && len(config.IdentityValidation.ResetPassword.JWTSecret) == 0 {
		validator.Push(fmt.Errorf(errFmtIdentityValidationResetPasswordJWTSecret))
	}

	if config.IdentityValidation.ElevatedSession.CodeLifespan <= 0 {
		config.IdentityValidation.ElevatedSession.CodeLifespan = schema.DefaultIdentityValidation.ElevatedSession.CodeLifespan
	}

	if config.IdentityValidation.ElevatedSession.ElevationLifespan <= 0 {
		config.IdentityValidation.ElevatedSession.ElevationLifespan = schema.DefaultIdentityValidation.ElevatedSession.ElevationLifespan
	}

	if config.IdentityValidation.ElevatedSession.Characters <= 0 {
		config.IdentityValidation.ElevatedSession.Characters = schema.DefaultIdentityValidation.ElevatedSession.Characters
	} else if config.IdentityValidation.ElevatedSession.Characters > 20 {
		validator.Push(fmt.Errorf(errFmtIdentityValidationElevatedSessionCharacterLength, config.IdentityValidation.ElevatedSession.Characters))
	}
}
