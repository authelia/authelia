package schema

import (
	"time"
)

// IdentityValidation represents the configuration for identity verification actions/flows.
type IdentityValidation struct {
	ResetPassword        IdentityValidationResetPassword `koanf:"reset_password" json:"reset_password" jsonschema:"title=Reset Password" jsonschema_description:"Identity validation options for the Reset Password flow"`
	CredentialManagement IdentityValidationCredentials   `koanf:"credential_management" json:"credential_management" jsonschema:"title=Credential Management" jsonschema_description:"Identity validation options for the Credential Management flows"`
}

// IdentityValidationResetPassword represents the tunable aspects of the reset password identity verification action/flow.
type IdentityValidationResetPassword struct {
	Expiration   time.Duration `koanf:"expiration" json:"expiration" jsonschema:"title=Expiration" jsonschema_description:"Duration of time the JWT is considered valid"`
	JWTAlgorithm string        `koanf:"jwt_algorithm" json:"jwt_algorithm" jsonschema:"title=JWT Algorithm,default=HS256,enum=HS256,enum=HS384,enum=HS512" jsonschema_description:"The JWT Algorithm (JWA) used to sign the Reset Password flow JWT's'"`
	JWTSecret    string        `koanf:"jwt_secret" json:"jwt_secret" jsonschema:"title=JWT Secret" jsonschema_description:"The JWT secret used to sign the Reset Password flow JWT's"`
}

// IdentityValidationCredentials represents the tunable aspects of the credential control identity verification action/flow.
type IdentityValidationCredentials struct {
	Expiration          time.Duration `koanf:"expiration" json:"expiration" jsonschema:"title=Expiration" jsonschema_description:"Duration of time the OTP code is considered valid"`
	ElevationExpiration time.Duration `koanf:"elevation_expiration" json:"elevation_expiration" jsonschema:"title=Elevation Expiration" jsonschema_description:"Duration of time the elevation can exist for after the user performs the validation"`
	Characters          int           `koanf:"characters" json:"otp_characters" jsonschema:"title=OTP Characters,minimum=6,maximum=12" jsonschema_description:"Number of characters in the generated OTP codes"`
	Skip2FA             bool          `koanf:"skip_2fa" json:"skip_2fa" jsonschema:"title=Skip 2FA" jsonschema_description:"Skips 2FA requirement TODO"`
}

// DefaultIdentityValidation has the default values for the IdentityValidation configuration.
var DefaultIdentityValidation = IdentityValidation{
	ResetPassword: IdentityValidationResetPassword{
		Expiration:   time.Minute * 5,
		JWTAlgorithm: "HS256",
	},
	CredentialManagement: IdentityValidationCredentials{
		Expiration:          time.Minute * 5,
		ElevationExpiration: time.Minute * 5,
		Characters:          8,
	},
}
